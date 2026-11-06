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
        catch (Exception error) when (error is ArgumentException or InvalidDataException or JsonException)
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
        var app = new Ensemble.Ensemble(TextWriter.Null);
        Agent Create() => new Agent(app, new Configuration(Vendor.OpenAI, "fixture",
            "http://127.0.0.1:1", "fixture", sandboxRoot: home), dataDirectory: home);
        try
        {
            var agent = Create();
            await agent.SetWorkingContextLimitAsync(first, persist: true);
            await agent.SetWorkingContextLimitAsync(second, persist: true);
            await agent.SaveMemorySettingsAsync(new MemorySettings());
            await agent.SaveRecallSettingsAsync(new RecallSettings(disabled: true));
            Check(agent.WorkingContextLimits().Length == 2, "Both model policies remain live");
        }
        finally { app.Destroy(); }
        app = new Ensemble.Ensemble(TextWriter.Null);
        try
        {
            var restored = Create();
            await restored.SaveMemorySettingsAsync(new MemorySettings());
            var limits = restored.WorkingContextLimits();
            Check(limits.Length == 2 && limits.Single(x => x.Model == first.Model).WorkingTokens == 250_000
                && limits.Single(x => x.Model == second.Model).WorkingTokens == 750_000,
                "Restart and first-operation sibling save preserve per-model limits");
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
    }

    public static void Values()
    {
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
