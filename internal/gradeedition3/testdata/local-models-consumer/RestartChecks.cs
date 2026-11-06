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
        var identity = new ModelIdentity("local", "chat", "small");
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
            manifest = new(Environment.ProcessId, wire, "retained-" + Guid.NewGuid().ToString("N"), document);
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
        try
        {
            _ = new ModelAuthentication(app, "none");
            app.LoadModelConfiguration(Encoding.UTF8.GetBytes(manifest.Document));
            Model model = app.ResolveModel(identity);
            // The reader host deliberately starts on a different legacy route and payer.
            var policy = new Configuration(Vendor.Anthropic, "legacy", "http://127.0.0.1:2",
                "synthetic-restart-key", sandboxRoot: home, disableStreaming: true);
            var parent = new Agent(app, policy, dataDirectory: home);
            if (write)
            {
                await parent.SetModelAsync(model);
                Agent child = await parent.CreateChildAsync("child", cancellation: deadline.Token);
                await child.PostAsync(new AgentInput(InputKind.Pause));
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
                    && child.Configuration().BaseUrl == "http://127.0.0.1:1",
                    "restart-read-exact-offering-no-host-fallback");
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
            var record = JsonNode.Parse(File.ReadAllBytes(registration))!.AsObject();
            Require(record["Offering"]?["VendorID"]?.GetValue<string>() == "local"
                && record["Offering"]?["EndpointID"]?.GetValue<string>() == "chat"
                && record["Offering"]?["ModelID"]?.GetValue<string>() == "small"
                && !File.ReadAllText(registration).Contains("synthetic-restart-key", StringComparison.Ordinal),
                "restart-write-saved-identity-no-secret");
            File.WriteAllBytes(manifestPath, JsonSerializer.SerializeToUtf8Bytes(manifest));
        }
    }
}
