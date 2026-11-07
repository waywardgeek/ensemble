using Ensemble;
using System.Text.Json.Nodes;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    public static async Task ConfiguredPreferenceRestart(string temp)
    {
        string container = Path.Combine(temp, "model-preference-restart");
        Directory.CreateDirectory(container);
        var policy = new Configuration(Vendor.Anthropic, "legacy", "http://127.0.0.1:1",
            "synthetic-not-saved", sandboxRoot: container);
        Model Configure(Application root, string id)
        {
            var vendor = root.ModelVendors().SingleOrDefault();
            if (vendor is null)
            {
                vendor = new ModelVendor(root, "local", "Local");
                _ = new ModelEndpoint(vendor, "test", "http://127.0.0.1:2", Surface.ChatCompletions,
                    new ModelAuthentication(root, "none"));
            }
            return new Model(vendor.Endpoints().Single(), id, "wire-" + id, id, new ModelFeatures(Media.None));
        }
        string childLaunch;
        string peerLaunch;
        string childHome;
        string peerHome;
        string rootHome;
        var app = new Application(TextWriter.Null, dataDirectory: container);
        try
        {
            Model old = Configure(app, "old");
            Model selected = Configure(app, "selected");
            var root = new Agent(app, policy, name: "root");
            rootHome = root.DataDirectory();
            await root.SetModelAsync(old);
            Agent child = await root.CreateChildAsync("child");
            Agent peer = await app.CreatePeerAsync(root, "peer");
            childHome = child.DataDirectory();
            peerHome = peer.DataDirectory();
            childLaunch = File.ReadAllText(Path.Combine(childHome, "child.json"));
            peerLaunch = File.ReadAllText(Path.Combine(container, "ensemble.json"));
            foreach (Agent agent in new[] { root, child, peer })
            {
                await agent.SetModelAsync(selected);
                await agent.SaveMemorySettingsAsync(new MemorySettings());
                await agent.SaveRecallSettingsAsync(new RecallSettings(disabled: true));
                await agent.SetWorkingContextLimitAsync(new WorkingContextLimit(Vendor.OpenAI,
                    selected.Name, 1234, offering: selected.Identity()), persist: true);
                var saved = AgentSettings.Read(agent.Sandbox(), Path.Combine(agent.DataDirectory(), "settings.json"));
                Check(saved?.SelectedOffering == selected.Identity() && saved.Recall.Disabled
                    && saved.WorkingContext.Single().WorkingTokens == 1234,
                    "Every sibling policy save preserves the explicit selected offering.");
                Check(!File.ReadAllText(Path.Combine(agent.DataDirectory(), "settings.json"))
                    .Contains("synthetic-not-saved"), "Agent preference must not contain a credential.");
            }
            Check(File.ReadAllText(Path.Combine(childHome, "child.json")) == childLaunch
                && File.ReadAllText(Path.Combine(container, "ensemble.json")) == peerLaunch,
                "Selection has one settings commit point, not a second launch-record write.");
        }
        finally { app.Destroy(); }

        // Preserve deliberately stale launch records. The old offering will not
        // even exist in the new root, so resolving it first cannot pass this test.
        File.WriteAllText(Path.Combine(childHome, "child.json"), childLaunch);
        File.WriteAllText(Path.Combine(container, "ensemble.json"), peerLaunch);
        app = new Application(TextWriter.Null, dataDirectory: container);
        try
        {
            Model selected = Configure(app, "selected");
            var root = new Agent(app, policy, name: "root");
            Check(ReferenceEquals(root.SelectedModel(), selected), "Top-level startup restores Agent selection.");
            await root.RestoreChildrenAsync();
            Agent peer = (await app.RestorePeersAsync(root)).Single();
            Check(ReferenceEquals(root.FindChild("child")!.SelectedModel(), selected)
                && ReferenceEquals(peer.SelectedModel(), selected),
                "Agent settings precede unavailable old child and peer launch routes.");
            foreach (Agent agent in new[] { root, root.FindChild("child")!, peer })
            {
                Check(!agent.History().Events().Any(e => e.Kind == EventKind.RequestSent),
                    "Restoring preferences must not replay inference.");
            }
        }
        finally { app.Destroy(); }

        foreach (bool peerCase in new[] { false, true })
        {
            string path = Path.Combine(peerCase ? peerHome : childHome, "settings.json");
            string saved = File.ReadAllText(path);
            var document = JsonNode.Parse(saved)!;
            document["selected_offering"]!["model"] = "absent";
            string damaged = document.ToJsonString();
            File.WriteAllText(path, damaged);
            // A valid first entry with a fresh home proves prevalidation, not
            // merely rollback of a peer that already performed startup effects.
            if (peerCase)
            {
                var entries = JsonNode.Parse(peerLaunch)!.AsArray();
                var first = entries[0]!.DeepClone();
                first["name"] = "never-created";
                first["model"] = "wire-selected";
                first["offering"]!["model"] = "selected";
                entries.Insert(0, first);
                File.WriteAllText(Path.Combine(container, "ensemble.json"), entries.ToJsonString());
            }
            app = new Application(TextWriter.Null, dataDirectory: container);
            try
            {
                Configure(app, "old"); // A valid legacy fallback must not be used.
                Configure(app, "selected");
                var root = new Agent(app, policy, name: "root");
                bool refused = false;
                try
                {
                    if (peerCase) await app.RestorePeersAsync(root);
                    else await root.RestoreChildrenAsync();
                }
                catch (InvalidOperationException) { refused = true; }
                Check(refused && app.Agents().Length == 1 && root.Children().IsEmpty
                    && !Directory.Exists(Path.Combine(container, "agents", "never-created"))
                    && File.ReadAllText(path) == damaged,
                    "Unavailable saved preference refuses before fleet startup, without legacy fallback.");
            }
            finally
            {
                app.Destroy();
                File.WriteAllText(path, saved);
                File.WriteAllText(Path.Combine(container, "ensemble.json"), peerLaunch);
            }
        }

        string rootSettings = Path.Combine(rootHome, "settings.json");
        string original = File.ReadAllText(rootSettings);
        foreach (string variant in new[] { "unavailable", "null", "duplicate" })
        {
            var document = JsonNode.Parse(original)!;
            if (variant == "unavailable") document["selected_offering"]!["model"] = "absent";
            if (variant == "null") document["selected_offering"] = null;
            string damaged = document.ToJsonString();
            if (variant == "duplicate")
                damaged = damaged.Replace("\"selected_offering\":", "\"selected_offering\":{},\"selected_offering\":");
            File.WriteAllText(rootSettings, damaged);
            app = new Application(TextWriter.Null, dataDirectory: container);
            try
            {
                Model selected = Configure(app, "selected");
                bool refused = false;
                try { _ = new Agent(app, policy, name: "root"); }
                catch (Exception error) when (error is InvalidOperationException or InvalidDataException
                    or FormatException or System.Text.Json.JsonException) { refused = true; }
                Check(refused && app.Agents().IsEmpty && selected.Agents().IsEmpty
                    && File.ReadAllText(rootSettings) == damaged,
                    "Invalid explicit preference refuses without fallback, membership or repair: " + variant);
            }
            finally { app.Destroy(); }
        }
        File.WriteAllText(rootSettings, original);
    }
}
