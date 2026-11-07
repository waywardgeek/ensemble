using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class ConfiguredPeerTests
{
    public static async Task Recovery(string temp)
    {
        string container = Path.Combine(temp, "peer-recovery");
        var legacy = new Configuration(Vendor.OpenAI, "old-wire", "http://127.0.0.1:1", "fake-key");
        var app = new Application(TextWriter.Null, dataDirectory: container);
        string registry = Path.Combine(container, "ensemble.json");
        byte[] original;
        try
        {
            Model local = Configure(app);
            var root = new Agent(app, legacy, name: "root");
            var peer = await app.CreatePeerAsync(root, "old-peer");
            await peer.SaveRecallSettingsAsync(new RecallSettings(disabled: true));
            peer.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("Keep this peer's own conversation.")]));
            await root.SetModelAsync(local);
            original = File.ReadAllBytes(registry);
        }
        finally { app.Destroy(); }

        app = new Application(TextWriter.Null, dataDirectory: container);
        try
        {
            Model local = Configure(app);
            var root = new Agent(app, legacy, name: "root");
            Check(ReferenceEquals(root.SelectedModel(), local), "Root restores its switched model.");
            try
            {
                await app.RestorePeersAsync(root);
                throw new Exception("Old unqualified route should require an explicit repair.");
            }
            catch (InvalidDataException) { }
            Check(app.SavedPeers(root).Single().Name == "old-peer" && app.Agents().Length == 1,
                "Failed fleet remains available for human recovery without a phantom peer.");
            string settingsPath = Path.Combine(container, "agents", "old-peer", "settings.json");
            byte[] before = File.ReadAllBytes(settingsPath);
            try
            {
                await app.RestorePeerAsync(root, "old-peer", new ModelIdentity("missing", "missing", "missing"));
                throw new Exception("Missing recovery model accepted.");
            }
            catch (InvalidOperationException) { }
            Check(File.ReadAllBytes(settingsPath).SequenceEqual(before), "Refused recovery changed settings.");
            var peer = await app.RestorePeerAsync(root, "old-peer", local.Identity());
            Check(ReferenceEquals(peer.SelectedModel(), local)
                && JsonNode.Parse(File.ReadAllText(settingsPath))!["recall"]!["disabled"]!.GetValue<bool>(),
                "Explicit repair preserves sibling settings and selects the real Model.");
            Check(JsonCodec.Text(peer.History().View().Dialogue.SelectMany(entry => entry.Parts))
                .Contains("Keep this peer's own conversation."), "Repair discarded saved conversation.");
            Check(File.ReadAllBytes(registry).SequenceEqual(original), "Repair rewrote historical launch choices.");
            Check(peer.History().Events().All(value => value.Kind != EventKind.RequestSent), "Repair performed inference.");
            try
            {
                await app.RestorePeerAsync(root, "old-peer", local.Identity());
                throw new Exception("Live peer was replaced.");
            }
            catch (InvalidOperationException) { }
        }
        finally { app.Destroy(); }

        app = new Application(TextWriter.Null, dataDirectory: container);
        try
        {
            Model local = Configure(app);
            var root = new Agent(app, legacy, name: "root");
            var peer = (await app.RestorePeersAsync(root)).Single();
            Check(ReferenceEquals(peer.SelectedModel(), local), "Repair did not survive the next restart.");
        }
        finally { app.Destroy(); }
    }

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
            var legacy = new Agent(app, policy, name: "legacy");
            Agent explicitPeer = await app.CreatePeerAsync(legacy, "explicit-local",
                cancellation: deadline.Token, offering: model.Identity());
            Check(ReferenceEquals(explicitPeer.SelectedModel(), model)
                && explicitPeer.Configuration().Billing == BillingMode.None
                && ReferenceEquals(legacy.Configuration(), policy),
                "Explicit peer offering changes route/funding without changing the source Agent.");
            await app.DeletePeerAsync(legacy, "explicit-local", deadline.Token);
            string before = File.ReadAllText(Path.Combine(container, "ensemble.json"));
            foreach (bool conflictingName in new[] { false, true })
            {
                bool refused = false;
                try
                {
                    await app.CreatePeerAsync(legacy, "must-not-start",
                        model: conflictingName ? "same-wire-name" : null,
                        cancellation: deadline.Token,
                        offering: conflictingName ? model.Identity()
                            : new ModelIdentity("ollama", "missing", "small"));
                }
                catch (Exception error) when (error is InvalidOperationException or ArgumentException)
                {
                    refused = true;
                }
                Check(refused && app.FindAgent("must-not-start") is null
                    && !Directory.Exists(Path.Combine(container, "agents", "must-not-start"))
                    && File.ReadAllText(Path.Combine(container, "ensemble.json")) == before,
                    "Invalid explicit offering must preserve live membership and saved registry.");
            }
            legacy.Destroy();
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
