using Ensemble;
using System.Text.Json.Nodes;
using Application = Ensemble.Ensemble;

internal static class GlobalSettingsTests
{
    private static void Check(bool value, string reason)
    {
        if (!value) throw new Exception(reason);
    }

    private static void Refuses(Action action)
    {
        try { action(); }
        catch (Exception error) when (error is ArgumentException or InvalidOperationException
            or System.Text.Json.JsonException or IOException or InvalidDataException or UnauthorizedAccessException) { return; }
        throw new Exception("Expected global settings refusal.");
    }

    public static async Task Persistence(string temp)
    {
        string path = Path.Combine(temp, "global-settings", "settings.json");
        Directory.CreateDirectory(Path.GetDirectoryName(path)!);
        File.WriteAllText(path, """
            {"openai_api_key":"synthetic-openai-key","anthropic_api_key":"synthetic-anthropic-key",
             "gemini_api_key":"synthetic-gemini-key","openrouter_api_key":"synthetic-router-key","future":{"retain":[1,"two",true]}}
            """);
        var root = new Application(TextWriter.Null);
        GlobalSettings settings;
        try
        {
            settings = new GlobalSettings(root, path);
            Check(ReferenceEquals(root.GlobalSettings(), settings) && ReferenceEquals(settings.Parent(), root),
                "Global settings establish both ownership ends.");
            Refuses(() => new GlobalSettings(root, path));
            Check(ReferenceEquals(root.GlobalSettings(), settings), "Duplicate owner refusal preserves original.");
            var api = settings.BindApiKey(Vendor.OpenAI, "openai-api")!;
            var anthropic = settings.BindApiKey(Vendor.Anthropic, "anthropic-api")!;
            var gemini = settings.BindApiKey(Vendor.Gemini, "gemini-api")!;
            Check(root.ModelAuthentications().Length == 3 && api.Kind == ModelAuthenticationKind.ApiKey
                && anthropic.Kind == ModelAuthenticationKind.ApiKey && gemini.Kind == ModelAuthenticationKind.ApiKey
                && api.DiagnosticKey() == "synthetic-openai-key"
                && anthropic.DiagnosticKey() == "synthetic-anthropic-key"
                && gemini.DiagnosticKey() == "synthetic-gemini-key",
                "Each global key creates its own real independent credential binding.");
            var router = settings.BindOpenRouterApiKey("openrouter-api")!;
            Check(router.DiagnosticKey() == "synthetic-router-key"
                && router.KeyHeader == ApiKeyHeader.Bearer
                && api.KeyHeader == ApiKeyHeader.ProtocolDefault,
                "Gateway key and header convention remain independent of OpenAI credentials.");
            var vendor = new ModelVendor(root, "openai", "OpenAI");
            var firstEndpoint = new ModelEndpoint(vendor, "api", "http://127.0.0.1:1", Surface.ChatCompletions, api);
            var otherBinding = new ModelAuthentication(root, "other-api", "synthetic-other-key");
            var secondEndpoint = new ModelEndpoint(vendor, "other", "http://127.0.0.1:2", Surface.ChatCompletions, otherBinding);
            var first = new Model(firstEndpoint, "same", "same-wire", "Same", new ModelFeatures(Media.None));
            var second = new Model(secondEndpoint, "same", "same-wire", "Same", new ModelFeatures(Media.None));
            Refuses(() => settings.ResolveModel(vendor, "same-wire"));
            var agent = new Agent(root, new Configuration(Vendor.OpenAI, "legacy", "http://127.0.0.1:1"));
            await agent.SetModelAsync(first);
            settings.RememberFunding(second);
            Check(settings.FundingDefaults()["openai"] == otherBinding.ID
                && ReferenceEquals(settings.ResolveModel(vendor, "same-wire"), second)
                && ReferenceEquals(agent.SelectedModel(), first),
                "Vendor default guides future resolution without retargeting an existing Agent.");
            var external = JsonNode.Parse(File.ReadAllText(path))!;
            external["future"]!["new"] = 42;
            File.WriteAllText(path, external.ToJsonString());
            settings.RememberFunding(first);
            var saved = JsonNode.Parse(File.ReadAllText(path))!;
            Check(JsonNode.DeepEquals(saved["future"], external["future"])
                && saved["openai_api_key"]!.GetValue<string>() == "synthetic-openai-key"
                && saved["anthropic_api_key"]!.GetValue<string>() == "synthetic-anthropic-key"
                && saved["gemini_api_key"]!.GetValue<string>() == "synthetic-gemini-key"
                && saved["openrouter_api_key"]!.GetValue<string>() == "synthetic-router-key",
                "Read-modify-replace preserves current unrelated fields and all credential values.");
            string original = File.ReadAllText(path);
            using (var held = new FileStream(path + ".lock", FileMode.Open, FileAccess.ReadWrite, FileShare.None))
            {
                Refuses(() => settings.RememberFunding(second));
                Check(File.ReadAllText(path) == original && ReferenceEquals(agent.SelectedModel(), first),
                    "Lock contention cannot change defaults or Agent selection.");
            }
            string backup = path + ".backup";
            File.Move(path, backup);
            Directory.CreateDirectory(path);
            try
            {
                Refuses(() => settings.RememberFunding(second));
                Check(File.ReadAllText(backup) == original, "I/O refusal preserves the existing settings bytes.");
            }
            finally
            {
                Directory.Delete(path);
                File.Move(backup, path);
            }
            api.Destroy();
            Refuses(() => settings.ResolveModel(vendor, "same-wire"));
            Check(File.ReadAllText(path) == original, "Unavailable saved funding refuses without another-key fallback.");
            settings.RememberFunding(second);
            Check(ReferenceEquals(settings.ResolveModel(vendor, "same-wire"), second)
                && ReferenceEquals(agent.SelectedModel(), first),
                "An unrelated missing binding does not hide explicitly selected valid funding or retarget the Agent.");
            settings.Destroy();
            Check(root.GlobalSettings() is null && root.ModelAuthentications().Contains(otherBinding),
                "Settings destruction detaches without destroying separately owned bindings.");
            settings = new GlobalSettings(root, path);
            Check(settings.FundingDefaults()["openai"] == "other-api", "Saved vendor default survives a new owner.");
        }
        finally { root.Destroy(); }
        Check(root.GlobalSettings() is null, "Application destruction tears down global settings.");
    }

    public static void Refusal(string temp)
    {
        string directory = Path.Combine(temp, "global-settings-refusal");
        string path = Path.Combine(directory, "settings.json");
        var root = new Application(TextWriter.Null);
        try
        {
            var absent = new GlobalSettings(root, path);
            Check(!Directory.Exists(directory) && absent.FundingDefaults().IsEmpty
                && absent.BindApiKey(Vendor.Gemini, "missing") is null
                && absent.BindOpenRouterApiKey("missing-router") is null,
                "Missing file reads do not create storage or borrow another credential.");
            absent.Destroy();
            Directory.CreateDirectory(directory);
            foreach (string bad in new[] { "null", "{} {}", "{\"openrouter_api_key\":42}", "{\"openai_api_key\":null}",
                "{\"openai_api_key\":\"a\",\"openai_api_key\":\"b\"}",
                "{\"model_funding_defaults\":null}", "{\"model_funding_defaults\":{\"openai\":null}}",
                "{\"model_funding_defaults\":{\"openai\":\"bad/reference\"}}", new string(' ', 1024 * 1024 + 1) })
            {
                File.WriteAllText(path, bad);
                Refuses(() => new GlobalSettings(root, path));
                Check(root.GlobalSettings() is null && File.ReadAllText(path) == bad,
                    "Malformed settings refuse without ownership or repair.");
            }
            foreach (string valid in new[]
            {
                """{"openai_api_key":"not-a-router-key"}""",
                """{"openai_api_key":"not-a-router-key","openrouter_api_key":"  "}"""
            })
            {
                File.WriteAllText(path, valid);
                var settings = new GlobalSettings(root, path);
                Check(settings.BindOpenRouterApiKey("absent-router") is null
                    && root.ModelAuthentications().IsEmpty,
                    "Absent or blank OpenRouter keys never borrow an available OpenAI key.");
                settings.Destroy();
            }
            File.WriteAllBytes(path, [0xff]);
            Refuses(() => new GlobalSettings(root, path));
        }
        finally { root.Destroy(); }
    }
}
