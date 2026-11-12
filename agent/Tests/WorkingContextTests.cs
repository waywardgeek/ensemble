using System.Collections.Immutable;

using Ensemble;
using System.Text.Json;

internal static class WorkingContextTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition) throw new Exception(message);
    }

    private static void Refuses(Action action)
    {
        try { action(); }
        catch (Exception error) when (error is ArgumentException or InvalidDataException or JsonException or FormatException)
        {
            return;
        }
        throw new Exception("Expected invalid token policy refusal");
    }

    public static async Task Persistence(string temp)
    {
        string home = Path.Combine(temp, "working-context-home");
        Directory.CreateDirectory(home);
        var first = new WorkingContextLimit(Vendor.OpenAI, "gpt-6-astra", 250_000, 1_000_000);
        var second = new WorkingContextLimit(Vendor.OpenAI, "review", 750_000, 1_000_000);
        var identity = new ModelIdentity("local", "et", "small");
        var qualified = new WorkingContextLimit(first.Vendor, first.Model, 2048, 4096, identity);
        var app = new Ensemble.Ensemble(TextWriter.Null);
        Agent Create() => new Agent(app, new Configuration(Vendor.OpenAI, "fixture",
            "http://127.0.0.1:1", "fixture", sandboxRoot: home), dataDirectory: home);
        try
        {
            var agent = Create();
            await agent.SetWorkingContextLimitAsync(first, persist: true);
            await agent.SetWorkingContextLimitAsync(second, persist: true);
            await agent.SetWorkingContextLimitAsync(qualified, persist: true);
            await agent.SaveMemorySettingsAsync(new MemorySettings());
            await agent.SaveRecallSettingsAsync(new RecallSettings(disabled: true));
            Check(agent.WorkingContextLimits().Length == 3, "Both model policies remain live");
            // Restore prevalidation reads through the existing Agent's Sandbox,
            // without constructing the destination Agent or another parser.
            var saved = AgentSettings.Read(agent.Sandbox(), Path.Combine(home, "settings.json"));
            Check(saved is not null && saved.Recall.Disabled && saved.WorkingContext.Length == 3,
                "Shared settings reader preserves the complete saved policy");
            string probe = Path.Combine(home, "settings-probe.json");
            Check(AgentSettings.Read(agent.Sandbox(), probe) is null, "Only missing settings supply defaults");
            foreach (string invalid in new[] { "null", "{}", "{\"version\":1,\"version\":1}", new string(' ', 16 * 1024 + 1) })
            {
                agent.Sandbox().ReplaceFile(probe, System.Text.Encoding.UTF8.GetBytes(invalid));
                Refuses(() => AgentSettings.Read(agent.Sandbox(), probe));
            }
        }
        finally { app.Destroy(); }
        app = new Ensemble.Ensemble(TextWriter.Null);
        try
        {
            var restored = Create();
            await restored.SaveMemorySettingsAsync(new MemorySettings());
            var limits = restored.WorkingContextLimits();
            Check(limits.Length == 3 && limits.Single(x => x.Model == first.Model && x.Offering is null).WorkingTokens == 250_000
                && limits.Single(x => x.Model == second.Model).WorkingTokens == 750_000,
                "Restart and first-operation sibling save preserve per-model limits");
            Check(limits.Single(x => x.Offering == identity).WorkingTokens == 2048,
                "Qualified policy survives restart without merging with the legacy wire name");
            Check(restored.RecallSettings().Disabled, "Context settings preserve recall policy");
            var noHome = new Agent(app, new Configuration(Vendor.OpenAI, "fixture", "http://127.0.0.1:1", "fixture"));
            await noHome.SetWorkingContextLimitAsync(first);
            bool refused = false;
            try { await noHome.SetWorkingContextLimitAsync(second, persist: true); }
            catch (InvalidOperationException) { refused = true; }
            Check(refused && noHome.WorkingContextLimits().Length == 1,
                "Failed durable save leaves live policy unchanged");
        }
        finally { app.Destroy(); }
        await OfferingIsolation();
    }

    private static async Task OfferingIsolation()
    {
        var root = new Ensemble.Ensemble(TextWriter.Null);
        try
        {
            var auth = new ModelAuthentication(root, "none");
            var vendor = new ModelVendor(root, "local", "Local");
            var firstEndpoint = new ModelEndpoint(vendor, "first", "http://127.0.0.1:1", Surface.ChatCompletions, auth);
            var secondEndpoint = new ModelEndpoint(vendor, "second", "http://127.0.0.1:2", Surface.ChatCompletions, auth);
            var first = new Model(firstEndpoint, "small", "same-wire", "First", new ModelFeatures(Media.None), 4096);
            var second = new Model(secondEndpoint, "small", "same-wire", "Second", new ModelFeatures(Media.None), 8192);
            var policy = new Configuration(Vendor.OpenAI, "legacy", "http://127.0.0.1:3", "synthetic");
            var agent = new Agent(root, new Configuration(policy, first));
            await agent.SetWorkingContextLimitAsync(new WorkingContextLimit(Vendor.OpenAI, "same-wire", 999));
            Check(agent.WorkingContextStatus().WorkingTokens == 4096,
                "Legacy wire policy does not override a configured offering's serving default");
            await agent.SetWorkingContextLimitAsync(new WorkingContextLimit(Vendor.OpenAI, "same-wire", 2000, 4096, first.Identity()));
            await agent.SetWorkingContextLimitAsync(new WorkingContextLimit(Vendor.OpenAI, "same-wire", 6000, 8192, second.Identity()));
            Check(agent.WorkingContextLimits().Length == 3 && agent.WorkingContextStatus().WorkingTokens == 2000
                && agent.WorkingContextStatus().Offering == first.Identity(),
                "Same wire name retains separate legacy and configured policies and reports selected identity");
            await agent.SetWorkingContextLimitAsync(new WorkingContextLimit(Vendor.OpenAI, "same-wire", 2100, 4096, first.Identity()));
            Check(agent.WorkingContextLimits().Length == 3 && agent.WorkingContextStatus().WorkingTokens == 2100,
                "Updating one offering replaces only that row");
            var sibling = new Agent(root, new Configuration(policy, second));
            foreach (var limit in agent.WorkingContextLimits()) await sibling.SetWorkingContextLimitAsync(limit);
            Check(sibling.WorkingContextStatus().WorkingTokens == 6000, "Second offering selects its own policy");
        }
        finally { root.Destroy(); }
    }

    public static void Values()
    {
        Check(Engine.LookupModel(Vendor.OpenAI, "gpt-6-astra", out var astraFeatures)
            && astraFeatures.ContextWindow == 1_050_000, "Astra advertised context window");
        var astra = new WorkingContextLimit(Vendor.OpenAI, "gpt-6-astra", 250_000, 1_000_000);
        var small = new WorkingContextLimit(Vendor.OpenAI, "small", 250);
        var review = new WorkingContextLimit(Vendor.OpenAI, "review", 750_000, 1_000_000);
        using var document = JsonDocument.Parse(JsonCodec.WriteWorkingContextLimits([astra, small, review]).ToJsonString());
        var restored = JsonCodec.ReadWorkingContextLimits(document.RootElement);
        Check(restored.Length == 3 && restored[0].WorkingTokens == 250_000
            && restored[0].CapacityTokens == 1_000_000, "Exact policy roundtrip");
        Check(restored[1].WorkingTokens == 250 && restored[1].CapacityTokens is null,
            "Tokens are literal; unknown capacity does not disable working policy");
        Check(restored[2].WorkingTokens == 750_000, "Deliberate large review is permitted");
        var offering = new ModelIdentity("local", "et", "small");
        var configured = new WorkingContextLimit(small.Vendor, small.Model, 2000, 4096, offering);
        using var qualified = JsonDocument.Parse(JsonCodec.WriteWorkingContextLimits([small, configured]).ToJsonString());
        var separate = JsonCodec.ReadWorkingContextLimits(qualified.RootElement);
        Check(separate.Length == 2 && separate[0].Offering is null && separate[1].Offering == offering,
            "Legacy and configured identities roundtrip separately");
        using var repeated = JsonDocument.Parse(JsonCodec.WriteWorkingContextLimits([configured, configured]).ToJsonString());
        Refuses(() => JsonCodec.ReadWorkingContextLimits(repeated.RootElement));
        foreach (string malformed in new[] { "null", "{}", "{\"vendor\":\"local\",\"endpoint\":\"et\",\"model\":\"small\",\"model\":\"other\"}" })
        {
            string row = "[{\"vendor\":\"OpenAI\",\"model\":\"small\",\"working_tokens\":1,\"capacity_tokens\":null,\"offering\":" + malformed + "}]";
            using var invalidOffering = JsonDocument.Parse(row);
            Refuses(() => JsonCodec.ReadWorkingContextLimits(invalidOffering.RootElement));
        }
        Refuses(() => new WorkingContextLimit(Vendor.OpenAI, "m", 0));
        Refuses(() => new WorkingContextLimit(Vendor.OpenAI, "m", -1));
        Refuses(() => new WorkingContextLimit(Vendor.OpenAI, "m", 250_000, 200_000));
        Refuses(() => new WorkingContextLimit(Vendor.OpenAI, "m", 1, 0));
        Refuses(() => new WorkingContextLimit(Vendor.OpenAI, "\n", 1));
        Refuses(() => new WorkingContextLimit((Vendor)999, "m", 1));
        foreach (string invalid in new[]
        {
            "null", "{}",
            "[{\"vendor\":\"OpenAI\",\"model\":\"m\",\"working_tokens\":1}]",
            "[{\"vendor\":\"OpenAI\",\"model\":\"m\",\"working_tokens\":1.5,\"capacity_tokens\":null}]",
            "[{\"vendor\":\"OpenAI\",\"model\":\"m\",\"working_tokens\":\"250000\",\"capacity_tokens\":null}]",
            "[{\"vendor\":\"1\",\"model\":\"m\",\"working_tokens\":1,\"capacity_tokens\":null}]",
            "[{\"vendor\":\"OpenAI\",\"model\":\"m\",\"working_tokens\":1,\"capacity_tokens\":null,\"skip\":true}]"
        })
        {
            using var bad = JsonDocument.Parse(invalid);
            Refuses(() => JsonCodec.ReadWorkingContextLimits(bad.RootElement));
        }
        using var duplicate = JsonDocument.Parse(JsonCodec.WriteWorkingContextLimits([astra, astra]).ToJsonString());
        Refuses(() => JsonCodec.ReadWorkingContextLimits(duplicate.RootElement));
        using var tooMany = JsonDocument.Parse(JsonCodec.WriteWorkingContextLimits(
            Enumerable.Range(0, 65).Select(i => new WorkingContextLimit(Vendor.OpenAI, "m" + i, 1)).ToImmutableArray()).ToJsonString());
        Refuses(() => JsonCodec.ReadWorkingContextLimits(tooMany.RootElement));
    }
}
