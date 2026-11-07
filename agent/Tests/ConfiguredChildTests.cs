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
            Agent child = await parent.CreateChildAsync("local", cancellation: deadline.Token);
            await child.PostAsync(new AgentInput(InputKind.Pause));
            child.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("Keep this conversation without replaying an inference.")]));
        }
        finally { app.Destroy(); }

        string path = Path.Combine(home, "children", "local", "child.json");
        string original = File.ReadAllText(path);
        JsonObject record = JsonNode.Parse(original)!.AsObject();
        Check(record["Version"]!.GetValue<int>() == 2
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
            Configuration restored = child.Configuration();
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

        // Each damaged registration gets a new root and no existing child that
        // could hide a refusal. No mutation is repaired by the reader.
        foreach (string variant in new[] { "missing-configuration", "wrong-wire-name", "null-identity",
            "missing-identity-field", "duplicate-identity-field", "legacy-under-configured-parent" })
        {
            record = JsonNode.Parse(original)!.AsObject();
            if (variant == "wrong-wire-name") record["Model"] = "another-wire-name";
            if (variant == "null-identity") record["Offering"] = null;
            if (variant == "missing-identity-field") record["Offering"]!.AsObject().Remove("EndpointID");
            if (variant == "legacy-under-configured-parent")
            {
                record["Version"] = 1;
                record.Remove("Offering");
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
