using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class Probe
{
    private sealed record RestartManifest(int WriterPid, string Wire, string Conversation, string Document);

    private static async Task RestartChecks(bool write)
    {
        string home = Path.Combine(Environment.CurrentDirectory, "restart-home");
        string manifestPath = Path.Combine(Environment.CurrentDirectory, "restart-manifest.json");
        string registration = Path.Combine(home, "children", "child", "child.json");
        var identity = new ModelIdentity("local", "selected", "small");
        string globalPath = Path.Combine(Environment.CurrentDirectory, "restart-global.json");
        RestartManifest manifest;
        if (write)
        {
            Require(!Directory.Exists(home) && !File.Exists(manifestPath), "restart-write-fresh-state");
            string wire = "wire-" + Guid.NewGuid().ToString("N");
            string document = $$"""
                {"version":1,"vendors":[{"id":"local","label":"Local","endpoints":[
                {"id":"chat","address":"http://127.0.0.1:1","surface":"chat_completions",
                "authentication":"none","models":[{"id":"small","name":"{{wire}}",
                "label":"Small","features":{"stream":true,"context_window":262144},
                "serving_context_tokens":4096}]}]}]}
                """;
            var graph = JsonNode.Parse(document)!;
            var endpoints = graph["vendors"]![0]!["endpoints"]!.AsArray();
            foreach (string id in new[] { "selected", "default" })
            {
                var endpoint = endpoints[0]!.DeepClone();
                endpoint["id"] = id;
                endpoint["address"] = "http://127.0.0.1:1/" + id;
                if (id == "default") endpoint["authentication"] = "default";
                endpoints.Add(endpoint);
            }
            manifest = new(Environment.ProcessId, wire, "retained-" + Guid.NewGuid().ToString("N"), graph.ToJsonString());
            Directory.CreateDirectory(home);
        }
        else
        {
            // No creation or seed fallback: only the successfully exited writer supplies state.
            manifest = JsonSerializer.Deserialize<RestartManifest>(File.ReadAllBytes(manifestPath))
                ?? throw new InvalidDataException("Missing restart manifest.");
            Require(manifest.WriterPid > 0 && manifest.WriterPid != Environment.ProcessId
                && manifest.Wire.StartsWith("wire-", StringComparison.Ordinal)
                && manifest.Conversation.StartsWith("retained-", StringComparison.Ordinal),
                "restart-read-separate-process-manifest");
            Require(File.Exists(registration), "restart-read-registration-exists");
        }

        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        var app = new Application(TextWriter.Null);
        byte[]? initialLaunch = null;
        try
        {
            _ = new ModelAuthentication(app, "none");
            _ = new ModelAuthentication(app, "default");
            var document = JsonNode.Parse(manifest.Document)!;
            if (!write) document["vendors"]![0]!["endpoints"]!.AsArray().RemoveAt(0);
            app.LoadModelConfiguration(Encoding.UTF8.GetBytes(document.ToJsonString()));
            Model model = app.ResolveModel(identity);
            Model globalChoice = app.ResolveModel(new ModelIdentity("local", "default", "small"));
            var global = new GlobalSettings(app, globalPath);
            // The reader host deliberately starts on a different legacy route and payer.
            var policy = new Configuration(Vendor.Anthropic, "legacy", "http://127.0.0.1:2",
                "synthetic-restart-key", sandboxRoot: home, disableStreaming: true);
            var parent = new Agent(app, policy, dataDirectory: home);
            if (write)
            {
                await parent.SetModelAsync(app.ResolveModel(new ModelIdentity("local", "chat", "small")));
                Agent child = await parent.CreateChildAsync("child", cancellation: deadline.Token);
                await child.PostAsync(new AgentInput(InputKind.Pause));
                initialLaunch = File.ReadAllBytes(registration);
                await child.SetModelAsync(model);
                await parent.SetModelAsync(model);
                global.RememberFunding(globalChoice);
                child.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                    parts: [new TextPart(manifest.Conversation)]));
                Require(ReferenceEquals(child.SelectedModel(), model)
                    && !child.History().Events().Any(fact => fact.Kind == EventKind.RequestSent),
                    "restart-write-real-child-no-request");
            }
            else
            {
                await parent.RestoreChildrenAsync(deadline.Token);
                Agent child = parent.FindChild("child") ?? throw new InvalidDataException("Restored child absent.");
                Require(ReferenceEquals(child.SelectedModel(), model) && model.Agents().Contains(child)
                    && child.Configuration().Offering == identity && child.Configuration().Model == manifest.Wire
                    && child.Configuration().Billing == BillingMode.None
                    && child.Configuration().BaseUrl == "http://127.0.0.1:1/selected",
                    "restart-read-exact-offering-no-host-fallback");
                Require(ReferenceEquals(parent.SelectedModel(), model)
                    && ReferenceEquals(global.ResolveModel(model.Vendor(), manifest.Wire), globalChoice)
                    && model.Vendor().Endpoints().All(endpoint => endpoint.ID != "chat"),
                    "restart-read-agent-preference-over-launch-and-global");
                Require((await child.WorkStatusAsync(deadline.Token)).Paused
                    && JsonCodec.Text(child.History().View().Dialogue.SelectMany(entry => entry.Parts))
                        .Contains(manifest.Conversation, StringComparison.Ordinal),
                    "restart-read-paused-conversation");
                Require(!child.History().Events().Any(fact => fact.Kind == EventKind.RequestSent)
                    && !parent.History().Events().Any(fact => fact.Kind == EventKind.RequestSent),
                    "restart-read-no-replayed-request");
            }
        }
        finally { app.Destroy(); }

        if (write)
        {
            // Other persistent state changes may refresh child.json. Deliberately
            // restore the grader-owned original launch to exercise precedence,
            // without altering the Agent's independently saved preference.
            File.WriteAllBytes(registration, initialLaunch ?? throw new InvalidDataException("Missing initial launch."));
            var record = JsonNode.Parse(File.ReadAllBytes(registration))!.AsObject();
            Require(record["Offering"]?["VendorID"]?.GetValue<string>() == "local"
                && record["Offering"]?["EndpointID"]?.GetValue<string>() == "chat"
                && record["Offering"]?["ModelID"]?.GetValue<string>() == "small"
                && !File.ReadAllText(registration).Contains("synthetic-restart-key", StringComparison.Ordinal),
                "restart-write-saved-identity-no-secret");
            var preference = JsonNode.Parse(File.ReadAllBytes(Path.Combine(home, "children", "child", "settings.json")))!;
            Require(preference["selected_offering"]?["endpoint"]?.GetValue<string>() == "selected"
                && JsonNode.Parse(File.ReadAllBytes(globalPath))!["model_funding_defaults"]?["local"]?.GetValue<string>() == "default",
                "restart-write-distinct-agent-and-global-preferences");
            File.WriteAllBytes(manifestPath, JsonSerializer.SerializeToUtf8Bytes(manifest));
        }
    }
}
