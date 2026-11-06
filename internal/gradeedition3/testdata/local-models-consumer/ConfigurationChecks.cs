using System.Text;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class Probe
{
    private static async Task ConfigurationChecks()
    {
        string home = Path.Combine(Environment.CurrentDirectory, "restore-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(home);
        string wire = "wire-" + Guid.NewGuid().ToString("N");
        string conversation = "retained-" + Guid.NewGuid().ToString("N");
        string secret = "synthetic-" + Guid.NewGuid().ToString("N");
        var identity = new ModelIdentity("local", "chat", "small");
        string document = $$"""
            {"version":1,"vendors":[{"id":"local","label":"Local","endpoints":[
              {"id":"chat","address":"http://127.0.0.1:1","surface":"chat_completions",
               "authentication":"none","models":[{"id":"small","name":"{{wire}}",
                "label":"Small","features":{"stream":true,"context_window":262144},
                "serving_context_tokens":4096}]}]}]}
            """;
        var policy = new Configuration(Vendor.Anthropic, "legacy", "http://127.0.0.1:2",
            secret, sandboxRoot: home, disableStreaming: true);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        var app = new Application(TextWriter.Null);
        try
        {
            var binding = new ModelAuthentication(app, "none");
            var parent = new Agent(app, policy, dataDirectory: home);
            app.LoadModelConfiguration(Encoding.UTF8.GetBytes(document));
            Model model = app.ResolveModel(identity);
            Require(ReferenceEquals(parent.Configuration(), policy) && parent.SelectedModel() is null,
                "configuration-load-preserves-launch");
            Require(ReferenceEquals(model.Parent().Authentication(), binding)
                && binding.Endpoints().Length == 1 && model.ServingContextTokens == 4096
                && model.Features.ContextWindow == 262144, "configuration-real-binding-and-capacity");

            // A valid earlier vendor must be rolled back when a later binding fails.
            var invalid = JsonNode.Parse(document)!.AsObject();
            var first = invalid["vendors"]![0]!;
            first["id"] = "temporary";
            var second = first.DeepClone();
            second["id"] = "invalid";
            second["endpoints"]![0]!["authentication"] = "absent";
            invalid["vendors"]!.AsArray().Add(second);
            bool refused = false;
            try { app.LoadModelConfiguration(Encoding.UTF8.GetBytes(invalid.ToJsonString())); }
            catch (Exception error) when (error is InvalidOperationException or InvalidDataException or ArgumentException) { refused = true; }
            Require(refused && app.ModelVendors().Length == 1 && binding.Endpoints().Length == 1
                && ReferenceEquals(app.ResolveModel(identity), model)
                && ReferenceEquals(parent.Configuration(), policy), "configuration-late-refusal-rolls-back");

            await parent.SetModelAsync(model);
            Agent child = await parent.CreateChildAsync("child", cancellation: deadline.Token);
            await child.PostAsync(new AgentInput(InputKind.Pause));
            child.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart(conversation)]));
            Require(ReferenceEquals(child.SelectedModel(), model) && model.Agents().Contains(child)
                && child.Configuration().Billing == BillingMode.None, "child-inherits-real-offering");
        }
        finally { app.Destroy(); }

        string registration = Path.Combine(home, "children", "child", "child.json");
        byte[] saved = File.ReadAllBytes(registration);
        var record = JsonNode.Parse(saved)!.AsObject();
        Require(record["Offering"]?["VendorID"]?.GetValue<string>() == "local"
            && record["Offering"]?["EndpointID"]?.GetValue<string>() == "chat"
            && record["Offering"]?["ModelID"]?.GetValue<string>() == "small"
            && !Encoding.UTF8.GetString(saved).Contains(secret, StringComparison.Ordinal),
            "child-registration-identity-not-secret");

        app = new Application(TextWriter.Null);
        try
        {
            _ = new ModelAuthentication(app, "none");
            app.LoadModelConfiguration(Encoding.UTF8.GetBytes(document));
            var parent = new Agent(app, policy, dataDirectory: home);
            await parent.RestoreChildrenAsync(deadline.Token);
            Agent child = parent.FindChild("child") ?? throw new InvalidOperationException("restored child absent");
            Model model = app.ResolveModel(identity);
            Require(ReferenceEquals(child.SelectedModel(), model) && model.Agents().Contains(child)
                && child.Configuration().Offering == identity && child.Configuration().Billing == BillingMode.None
                && child.Configuration().BaseUrl == "http://127.0.0.1:1", "child-restores-exact-offering");
            Require((await child.WorkStatusAsync(deadline.Token)).Paused
                && JsonCodec.Text(child.History().View().Dialogue.SelectMany(entry => entry.Parts)).Contains(conversation)
                && !child.History().Events().Any(fact => fact.Kind == EventKind.RequestSent),
                "child-restores-conversation-without-request");
            model.Destroy();
            bool refused = false;
            try { _ = child.Configuration(); }
            catch (InvalidOperationException) { refused = true; }
            Require(refused && child.SelectedModel() is null && parent.Children().Contains(child)
                && JsonCodec.Text(child.History().View().Dialogue.SelectMany(entry => entry.Parts)).Contains(conversation),
                "model-loss-refuses-without-agent-destruction");
        }
        finally { app.Destroy(); }

        // Same wire name under a different endpoint is not a replacement identity.
        byte[] beforeRefusal = File.ReadAllBytes(registration);
        app = new Application(TextWriter.Null);
        try
        {
            _ = new ModelAuthentication(app, "none");
            app.LoadModelConfiguration(Encoding.UTF8.GetBytes(document.Replace("\"id\":\"chat\"", "\"id\":\"decoy\"")));
            var parent = new Agent(app, policy, dataDirectory: home);
            bool refused = false;
            try { await parent.RestoreChildrenAsync(deadline.Token); }
            catch (Exception error) when (error is InvalidOperationException or InvalidDataException) { refused = true; }
            Require(refused && parent.Children().Length == 0, "child-missing-route-no-wire-name-fallback");
            Require(File.ReadAllBytes(registration).SequenceEqual(beforeRefusal)
                && !parent.History().Events().Any(fact => fact.Kind == EventKind.RequestSent),
                "child-refusal-preserves-registration-no-request");
        }
        finally { app.Destroy(); }
    }
}
