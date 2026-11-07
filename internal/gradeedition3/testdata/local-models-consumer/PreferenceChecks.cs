using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class Probe
{
    private static async Task PreferenceChecks()
    {
        string directory = Path.Combine(Environment.CurrentDirectory, "preferences-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(directory);
        string home = Path.Combine(directory, "agent");
        string globalPath = Path.Combine(directory, "global.json");
        string secret = "synthetic-" + Guid.NewGuid().ToString("N");
        File.WriteAllText(globalPath, new JsonObject
        {
            ["openai_api_key"] = secret, ["unrelated"] = "keep-" + Guid.NewGuid().ToString("N")
        }.ToJsonString());
        string originalGlobal = File.ReadAllText(globalPath);
        string wire = "wire-" + Guid.NewGuid().ToString("N");
        var policy = new Configuration(Vendor.Anthropic, "legacy", "http://127.0.0.1:1",
            "synthetic-launch", sandboxRoot: directory);
        Model Add(ModelVendor vendor, string id) => new(new ModelEndpoint(vendor, id,
            "http://127.0.0.1:2/" + id, Surface.ChatCompletions,
            new ModelAuthentication(vendor.Parent(), "funding-" + id)), "model", wire, id,
            new ModelFeatures(Media.None));
        bool Refuses(Action action)
        {
            try { action(); return false; }
            catch (Exception error) when (error is InvalidOperationException or ArgumentException
                or IOException or InvalidDataException) { return true; }
        }
        var app = new Application(TextWriter.Null);
        try
        {
            var global = new GlobalSettings(app, globalPath);
            var vendor = new ModelVendor(app, "local", "Local");
            Model a = Add(vendor, "a"), b = Add(vendor, "b");
            var agent = new Agent(app, policy, dataDirectory: home);
            Require(Refuses(() => global.ResolveModel(vendor, wire)), "preference-ambiguous-funding-refuses");
            await agent.SetModelAsync(a);
            string settingsPath = Path.Combine(home, "settings.json");
            var first = JsonNode.Parse(File.ReadAllText(settingsPath))!.AsObject();
            Require(first["selected_offering"]?["endpoint"]?.GetValue<string>() == "a"
                && !first.ToJsonString().Contains(secret, StringComparison.Ordinal), "preference-agent-saves-reference-not-secret");
            global.RememberFunding(b);
            Require(ReferenceEquals(agent.SelectedModel(), a), "preference-default-does-not-retarget-agent");
            Require(ReferenceEquals(global.ResolveModel(vendor, wire), b), "preference-resolves-remembered-binding");
            var updatedGlobal = JsonNode.Parse(File.ReadAllText(globalPath))!;
            var original = JsonNode.Parse(originalGlobal)!;
            Require(JsonNode.DeepEquals(updatedGlobal["unrelated"], original["unrelated"])
                && updatedGlobal["openai_api_key"]!.GetValue<string>() == secret,
                "preference-global-preserves-unrelated-and-key");
            byte[] savedGlobal = File.ReadAllBytes(globalPath);
            using (var held = new FileStream(globalPath + ".lock", FileMode.OpenOrCreate,
                FileAccess.ReadWrite, FileShare.None))
                Require(Refuses(() => global.RememberFunding(a))
                    && File.ReadAllBytes(globalPath).SequenceEqual(savedGlobal)
                    && ReferenceEquals(global.ResolveModel(vendor, wire), b), "preference-global-lock-refusal-preserves-default");
            await agent.SetModelAsync(b);
            var second = JsonNode.Parse(File.ReadAllText(settingsPath))!.AsObject();
            first.Remove("selected_offering"); second.Remove("selected_offering");
            Require(first.Count > 0 && JsonNode.DeepEquals(first, second), "preference-selection-preserves-sibling-settings");
            string backup = settingsPath + ".held";
            File.Move(settingsPath, backup);
            Directory.CreateDirectory(settingsPath);
            bool refused = false;
            try
            {
                try { await agent.SetModelAsync(a); }
                catch (Exception error) when (error is InvalidOperationException or IOException
                    or InvalidDataException or UnauthorizedAccessException) { refused = true; }
                Require(refused && ReferenceEquals(agent.SelectedModel(), b)
                    && b.Agents().Contains(agent) && !a.Agents().Contains(agent),
                    "preference-agent-save-refusal-preserves-selection");
            }
            finally { Directory.Delete(settingsPath); File.Move(backup, settingsPath); }
            global.RememberFunding(a);
            Require(ReferenceEquals(agent.SelectedModel(), b), "preference-new-default-keeps-explicit-selection");
            File.WriteAllText(globalPath, "{\"model_funding_defaults\":{\"local\":\"missing\"}}");
            Require(Refuses(() => global.ResolveModel(vendor, wire))
                && ReferenceEquals(agent.SelectedModel(), b), "preference-missing-funding-no-fallback");
            global.RememberFunding(a);
            Require(!agent.History().Events().Any(value => value.Kind == EventKind.RequestSent),
                "preference-operations-no-inference");
        }
        finally { app.Destroy(); }

        app = new Application(TextWriter.Null);
        try
        {
            var global = new GlobalSettings(app, globalPath);
            var vendor = new ModelVendor(app, "local", "Local");
            Model a = Add(vendor, "a"), b = Add(vendor, "b");
            var restored = new Agent(app, policy, dataDirectory: home);
            Require(ReferenceEquals(global.ResolveModel(vendor, wire), a)
                && ReferenceEquals(restored.SelectedModel(), b), "preference-restores-agent-ahead-of-global-default");
            Require(!restored.History().Events().Any(value => value.Kind == EventKind.RequestSent),
                "preference-restore-no-replayed-inference");
        }
        finally { app.Destroy(); }

        byte[] saved = File.ReadAllBytes(Path.Combine(home, "settings.json"));
        app = new Application(TextWriter.Null);
        try
        {
            var global = new GlobalSettings(app, globalPath);
            var vendor = new ModelVendor(app, "local", "Local");
            Model a = Add(vendor, "a");
            Require(ReferenceEquals(global.ResolveModel(vendor, wire), a), "preference-refusal-has-usable-global-decoy");
            Require(Refuses(() => _ = new Agent(app, policy, dataDirectory: home))
                && app.Agents().Length == 0
                && File.ReadAllBytes(Path.Combine(home, "settings.json")).SequenceEqual(saved),
                "preference-unavailable-agent-choice-refuses-global-fallback");
        }
        finally { app.Destroy(); }
    }
}
