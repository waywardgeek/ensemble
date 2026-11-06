using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class ConfiguredPeerTests
{
    private static void Check(bool condition, string reason)
    {
        if (!condition) throw new Exception(reason);
    }

    private static Model Configure(Application app)
    {
        var authentication = new ModelAuthentication(app, "local");
        var vendor = new ModelVendor(app, "ollama", "Ollama");
        var endpoint = new ModelEndpoint(vendor, "test", "http://127.0.0.1:1",
            Surface.ChatCompletions, authentication);
        return new Model(endpoint, "small", "same-wire-name", "Small local model",
            new ModelFeatures(Media.None), servingContextTokens: 4096);
    }

    public static async Task Persistence(string temp)
    {
        string work = Path.Combine(temp, "configured-peer");
        string container = Path.Combine(work, "ensemble");
        Directory.CreateDirectory(work);
        var policy = new Configuration(Vendor.Anthropic, "legacy-host", "http://127.0.0.1:2",
            "synthetic-old-key", sandboxRoot: work, disableStreaming: true);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(30));
        var app = new Application(TextWriter.Null, dataDirectory: container);
        try
        {
            Model model = Configure(app);
            var root = new Agent(app, new Configuration(policy, model), name: "root");
            Agent peer = await app.CreatePeerAsync(root, "local", cancellation: deadline.Token);
            Check(ReferenceEquals(peer.SelectedModel(), model) && model.Agents().Contains(peer),
                "Peer inherits the actual configured Model relationship.");
            peer.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("Retain peer conversation without inference.")]));
        }
        finally { app.Destroy(); }

        string registry = Path.Combine(container, "ensemble.json");
        string original = File.ReadAllText(registry);
        var rows = JsonNode.Parse(original)!.AsArray();
        Check(rows[0]!["offering"] is JsonObject && !original.Contains("synthetic-old-key")
            && !original.Contains("http:"), "Registry stores offering identity, not credentials or endpoint copies.");

        app = new Application(TextWriter.Null, dataDirectory: container);
        try
        {
            Model model = Configure(app);
            var root = new Agent(app, policy, name: "root");
            Agent peer = (await app.RestorePeersAsync(root, deadline.Token)).Single();
            Configuration restored = peer.Configuration();
            Check(ReferenceEquals(peer.SelectedModel(), model) && restored.Offering == model.Identity()
                && restored.Billing == BillingMode.None && restored.ApiKey.Length == 0
                && restored.BaseUrl == "http://127.0.0.1:1",
                "Restore uses the saved offering instead of the host's legacy route and funding.");
            Check(JsonCodec.Text(peer.History().View().Dialogue.SelectMany(entry => entry.Parts))
                .Contains("Retain peer conversation"), "Peer conversation survives restoration.");
            Check(File.ReadAllText(registry) == original, "Restoration does not rewrite launch choices.");
        }
        finally { app.Destroy(); }

        foreach (string variant in new[] { "missing-route", "wrong-wire", "null", "missing-field",
            "duplicate-field", "legacy-under-configured-host", "later-missing-route" })
        {
            rows = JsonNode.Parse(original)!.AsArray();
            if (variant == "wrong-wire") rows[0]!["model"] = "other-name";
            if (variant == "null") rows[0]!["offering"] = null;
            if (variant == "missing-field") rows[0]!["offering"]!.AsObject().Remove("endpoint");
            if (variant == "legacy-under-configured-host") rows[0]!.AsObject().Remove("offering");
            if (variant == "later-missing-route")
            {
                rows[0]!["name"] = "never-first";
                JsonNode later = rows[0]!.DeepClone();
                later["name"] = "never-started";
                later["offering"]!["endpoint"] = "absent";
                rows.Add(later);
            }
            string damaged = rows.ToJsonString();
            if (variant == "duplicate-field")
            {
                damaged = damaged.Replace("\"vendor\":\"ollama\"", "\"vendor\":\"ollama\",\"vendor\":\"ollama\"");
                Check(damaged != rows.ToJsonString(), "Duplicate-field fixture must actually change JSON.");
            }
            File.WriteAllText(registry, damaged);
            app = new Application(TextWriter.Null, dataDirectory: container);
            try
            {
                Model? model = variant == "missing-route" ? null : Configure(app);
                Configuration host = variant == "legacy-under-configured-host"
                    ? new Configuration(policy, model!) : policy;
                var root = new Agent(app, host, name: "root");
                bool refused = false;
                try { await app.RestorePeersAsync(root, deadline.Token); }
                catch (Exception error) when (error is InvalidDataException or InvalidOperationException or ArgumentException)
                {
                    refused = true;
                }
                Check(refused && app.Agents().Length == 1,
                    "Invalid launch set must refuse before constructing peers: " + variant);
                Check(!Directory.Exists(Path.Combine(container, "agents", "never-started"))
                    && !Directory.Exists(Path.Combine(container, "agents", "never-first"))
                    && File.ReadAllText(registry) == damaged,
                    "Refusal must preserve registry and avoid later peer startup effects: " + variant);
            }
            finally { app.Destroy(); }
        }
        File.WriteAllText(registry, original);
    }
}
