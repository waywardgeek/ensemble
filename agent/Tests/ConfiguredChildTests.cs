using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class ConfiguredChildTests
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

    // No inference is admitted: this checks real creation, storage, skills, and
    // confinement without substituting provider availability for policy evidence.
    public static async Task Launch(string temp)
    {
        string home = Path.Combine(temp, "selected-child-launch");
        string manuals = Path.Combine(temp, "selected-child-manuals");
        string imported = Path.Combine(temp, "selected-child-import");
        Directory.CreateDirectory(imported);
        McpTests.Manual(manuals, "primary", "type: primary\ntools: read_file\nloadable-skills: extra\n", "Parent primary");
        McpTests.Manual(manuals, "extra", "type: loadable\ntools: list_directory\n", "Parent startup");
        McpTests.Manual(manuals, "worker", "type: primary\ntools: read_file wait_for_job\n", "Selected worker");
        var policy = new Configuration(Vendor.Anthropic, "parent-model", "http://127.0.0.1:2");
        var app = new Application(TextWriter.Null);
        try
        {
            Model model = Configure(app);
            var parent = new Agent(app, policy, dataDirectory: home,
                skillSpec: new SkillSpec(manuals, "primary", StartupSkills: ["extra"]));
            await parent.SkillManager()!.InitializeAsync();
            Agent child = await parent.CreateChildAsync("trial", launch:
                new ChildLaunchOptions(model.Identity(), "worker", true, [imported]));
            string root = child.Sandbox().Root();
            Check(root == child.DataDirectory() && root != parent.Sandbox().Root(),
                "Confined child must own its complete writable workspace.");
            Check(child.Configuration().Offering == model.Identity()
                && child.Configuration().ProcessReadPaths!.Value.SequenceEqual([imported]),
                "Exact route and canonical imports were not selected.");
            Check(child.ToolEnabled("wait_for_job") && !child.ToolEnabled("list_directory"),
                "Explicit primary inherited unrelated startup grants.");
            bool refused = false;
            try
            {
                child.Sandbox().ReadDirectory(manuals);
            }
            catch (IOException)
            {
                refused = true;
            }
            Check(refused, "Built-in file access escaped the child's workspace.");
            child.Destroy();
            child = await parent.CreateChildAsync("trial");
            Check(child.Sandbox().Root() == root && child.ToolEnabled("wait_for_job")
                && !child.ToolEnabled("list_directory") && child.Submission() is null,
                "Reused child ID lost launch policy or revived old completion.");
            child.Destroy();
            refused = false;
            try
            {
                await parent.CreateChildAsync("trial", launch: new ChildLaunchOptions());
            }
            catch (ArgumentException)
            {
                refused = true;
            }
            Check(refused && parent.Children().IsEmpty,
                "Reused child ID accepted replacement launch authority.");
            await parent.RestoreChildrenAsync();
            child = parent.FindChild("trial")!;
            Check(child.Sandbox().Root() == root && child.ToolEnabled("wait_for_job")
                && child.Configuration().ProcessReadPaths!.Value.SequenceEqual([imported]),
                "Restart lost confinement, primary, or imports.");
        }
        finally
        {
            app.Destroy();
        }
    }

    public static async Task Restart(string temp)
    {
        string home = Path.Combine(temp, "configured-child-restart");
        Directory.CreateDirectory(home);
        var policy = new Configuration(Vendor.Anthropic, "legacy-parent", "http://127.0.0.1:2",
            "synthetic-old-key", sandboxRoot: home, disableStreaming: true);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(30));
        var app = new Application(TextWriter.Null);
        try
        {
            Model model = Configure(app);
            var parent = new Agent(app, new Configuration(policy, model), dataDirectory: home);
            bool refused = false;
            try
            {
                await parent.CreateChildAsync("too-large", workingContextTokens: 700_000);
            }
            catch (ArgumentOutOfRangeException)
            {
                refused = true;
            }
            Check(refused && parent.Children().IsEmpty,
                "Serving capacity must refuse oversized child policy before ownership changes.");
            Agent child = await parent.CreateChildAsync("local", cancellation: deadline.Token,
                workingContextTokens: 3000);
            await child.PostAsync(new AgentInput(InputKind.Pause));
            child.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("Keep this conversation without replaying an inference.")]));
        }
        finally { app.Destroy(); }

        string path = Path.Combine(home, "children", "local", "child.json");
        string original = File.ReadAllText(path);
        JsonObject record = JsonNode.Parse(original)!.AsObject();
        Check(record["Version"]!.GetValue<int>() == 3
            && record["Offering"]!["VendorID"]!.GetValue<string>() == "ollama"
            && !original.Contains("synthetic-old-key"),
            "Configured child saves qualified identity, not inherited credentials.");

        app = new Application(TextWriter.Null);
        try
        {
            Model model = Configure(app);
            // Restoration must use the saved offering, not this parent's route.
            var parent = new Agent(app, policy, dataDirectory: home);
            await parent.RestoreChildrenAsync(deadline.Token);
            Agent child = parent.FindChild("local")!;
            await child.SetMemorySettingsAsync(child.MemorySettings());
            Configuration restored = child.Configuration();
            Check(child.WorkingContextStatus().WorkingTokens == 3000
                && child.WorkingContextStatus().CapacityTokens == 4096
                && child.WorkingContextLimits().Single().Offering == model.Identity(),
                "Child restart lost its exact-offering working policy.");
            Check(restored.Offering == model.Identity() && restored.BaseUrl == "http://127.0.0.1:1"
                && restored.Billing == BillingMode.None && restored.ApiKey.Length == 0,
                "Restoration preserves the configured route and explicit unauthenticated funding.");
            Check((await child.WorkStatusAsync(deadline.Token)).Paused
                && JsonCodec.Text(child.History().View().Dialogue.SelectMany(entry => entry.Parts))
                    .Contains("Keep this conversation"),
                "Restoration preserves pause and conversation without a provider request.");
            await child.SetModelAsync(model.Name);
            string settings = Path.Combine(child.DataDirectory(), "settings.json");
            Check(JsonNode.Parse(File.ReadAllText(settings))!["selected_offering"]!["endpoint"]!.GetValue<string>() == "test",
                "Explicit selection saves its offering in Agent settings.");
        }
        finally { app.Destroy(); }

        // These are legacy-record refusal checks: no explicit preference may
        // mask their route-name assertions. Preference precedence has its own test.
        string settingsPath = Path.Combine(home, "children", "local", "settings.json");
        var settingsDocument = JsonNode.Parse(File.ReadAllText(settingsPath))!.AsObject();
        settingsDocument.Remove("selected_offering");
        File.WriteAllText(settingsPath, settingsDocument.ToJsonString());

        // Version two predates persisted sandbox/primary choices. Its existing
        // exact offering must still restore under the parent's inherited policy.
        JsonObject legacy = JsonNode.Parse(original)!.AsObject();
        legacy["Version"] = 2;
        legacy.Remove("SandboxRoot");
        legacy.Remove("ProcessReadPaths");
        legacy.Remove("PrimarySkill");
        File.WriteAllText(path, legacy.ToJsonString());
        app = new Application(TextWriter.Null);
        try
        {
            Model model = Configure(app);
            var parent = new Agent(app, policy, dataDirectory: home);
            await parent.RestoreChildrenAsync(deadline.Token);
            Check(parent.FindChild("local")!.Configuration().Offering == model.Identity(),
                "Valid version-two registration lost its exact offering.");
        }
        finally
        {
            app.Destroy();
        }

        // Each damaged registration gets a new root and no existing child that
        // could hide a refusal. No mutation is repaired by the reader.
        foreach (string variant in new[] { "missing-configuration", "wrong-wire-name", "null-identity",
            "missing-identity-field", "duplicate-identity-field", "legacy-under-configured-parent",
            "missing-sandbox-root", "missing-process-read-paths" })
        {
            record = JsonNode.Parse(original)!.AsObject();
            if (variant == "missing-sandbox-root") record.Remove("SandboxRoot");
            if (variant == "missing-process-read-paths") record.Remove("ProcessReadPaths");
            if (variant == "wrong-wire-name") record["Model"] = "another-wire-name";
            if (variant == "null-identity") record["Offering"] = null;
            if (variant == "missing-identity-field") record["Offering"]!.AsObject().Remove("EndpointID");
            if (variant == "legacy-under-configured-parent")
            {
                record["Version"] = 1;
                record.Remove("Offering");
                record.Remove("SandboxRoot");
                record.Remove("ProcessReadPaths");
                record.Remove("PrimarySkill");
            }
            string damaged = record.ToJsonString();
            if (variant == "duplicate-identity-field")
            {
                damaged = damaged.Replace("\"VendorID\":\"ollama\"", "\"VendorID\":\"ollama\",\"VendorID\":\"ollama\"");
                Check(damaged != record.ToJsonString(), "Duplicate-field fixture must actually differ.");
            }
            File.WriteAllText(path, damaged);
            app = new Application(TextWriter.Null);
            try
            {
                Model? model = variant == "missing-configuration" ? null : Configure(app);
                Configuration parentPolicy = variant == "legacy-under-configured-parent"
                    ? new Configuration(policy, model!) : policy;
                var parent = new Agent(app, parentPolicy, dataDirectory: home);
                bool refused = false;
                try { await parent.RestoreChildrenAsync(deadline.Token); }
                catch (Exception error) when (error is InvalidDataException or InvalidOperationException) { refused = true; }
                Check(refused && parent.Children().Length == 0, "Configured restore must refuse " + variant);
                Check(File.ReadAllText(path) == damaged, "Refusal must not rewrite the saved identity.");
            }
            finally { app.Destroy(); }
        }
        File.WriteAllText(path, original);
    }
}
