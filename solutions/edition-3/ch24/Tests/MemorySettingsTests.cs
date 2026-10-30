using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    public static async Task MemorySettingsPersistence(string directory)
    {
        string home = Path.Combine(directory, "memory-settings-persist");
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = MemoryAgent(app, home, "http://127.0.0.1:1");
            File.WriteAllText(Path.Combine(home, "SOUL.md"), "first soul");
            await agent.WorkStatusAsync().WaitAsync(TimeSpan.FromSeconds(10));
            var disabled = new MemoryBandSettings(1700, 3500, 1800, disabled: true);
            var desired = new MemorySettings(conversation: new MemoryBandSettings(4000), soul: disabled,
                memory: disabled, archive64: disabled, archive8: disabled, session: disabled);
            await agent.SaveMemorySettingsAsync(desired).WaitAsync(TimeSpan.FromSeconds(10));
            Check(ReferenceEquals(agent.MemorySettings(), desired), "Save acknowledged before publication");
            Check(agent.History().View().Memories.IsEmpty, "Disabled corpus remained active");
            string saved = File.ReadAllText(Path.Combine(home, "settings.json"));
            Check(saved.Contains("3500", StringComparison.Ordinal), "Custom thresholds not persisted");
            app.Destroy();
            File.WriteAllText(Path.Combine(home, "SOUL.md"), "edited while stopped");
            app = new Application(TextWriter.Null);
            agent = MemoryAgent(app, home, "http://127.0.0.1:1");
            await agent.WorkStatusAsync().WaitAsync(TimeSpan.FromSeconds(10));
            Check(agent.MemorySettings().Soul.Disabled && agent.MemorySettings().Session.HighWatermark == 3500
                && agent.MemorySettings().Conversation.Budget == 4000, "Restart lost saved policy");
            Check(agent.History().View().Memories.IsEmpty, "Startup enabled saved-disabled corpus");
            await agent.SaveMemorySettingsAsync(new MemorySettings()).WaitAsync(TimeSpan.FromSeconds(10));
            Check(agent.History().View().Memories.Single().Text == "edited while stopped", "Re-enable did not read fresh bytes");
            Check(!agent.History().Events().Any(e => e.Kind == EventKind.RequestSent), "Settings started inference");
        }
        finally { app.Destroy(); }
    }

    public static async Task MemorySettingsRefusal(string directory)
    {
        string home = Path.Combine(directory, "memory-settings-refusal");
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = MemoryAgent(app, home, "http://127.0.0.1:1");
            File.WriteAllText(Path.Combine(home, "SOUL.md"), "keep this soul");
            await agent.WorkStatusAsync().WaitAsync(TimeSpan.FromSeconds(10));
            var previous = agent.MemorySettings();
            int count = agent.History().Events().Length;
            string path = Path.Combine(home, "settings.json");
            Directory.CreateDirectory(path);
            await SettingsRefuses(() => agent.SaveMemorySettingsAsync(
                new MemorySettings(soul: new MemoryBandSettings(disabled: true))));
            Check(ReferenceEquals(previous, agent.MemorySettings()) && agent.History().Events().Length == count,
                "Save failure changed runtime policy or replay");
            Directory.Delete(path);
            var off = new MemorySettings(soul: new MemoryBandSettings(disabled: true));
            await agent.SaveMemorySettingsAsync(off);
            string saved = File.ReadAllText(path);
            File.WriteAllBytes(Path.Combine(home, "SOUL.md"), [0xff]);
            await SettingsRefuses(() => agent.SaveMemorySettingsAsync(new MemorySettings()));
            Check(ReferenceEquals(off, agent.MemorySettings()) && File.ReadAllText(path) == saved,
                "Scan refusal replaced durable or live settings");
            Check((await agent.WorkStatusAsync()).Activity == AgentActivity.Idle, "Save refusal killed actor");
        }
        finally { app.Destroy(); }
    }

    public static async Task MemorySettingsCodec(string directory)
    {
        var app = new Application(TextWriter.Null);
        bool startupRefused = false;
        try
        {
            string home = Path.Combine(directory, "memory-settings-codec");
            var agent = MemoryAgent(app, home, "http://127.0.0.1:1");
            string path = Path.Combine(home, "settings.json");
            string policy = JsonCodec.WriteMemorySettings(new MemorySettings()).ToJsonString();
            string valid = new JsonObject { ["version"] = 1, ["memory"] = JsonNode.Parse(policy) }.ToJsonString();
            foreach (string text in new[]
            {
                "{", valid.Replace("\"version\":1", "\"version\":2"),
                valid.Replace("\"version\":1", "\"version\":1,\"version\":1"),
                valid.Replace("\"budget\":3000", "\"budget\":3000,\"budget\":3000"),
                valid.Replace("\"budget\":3000", "\"budget\":-1"),
                valid.Replace("\"disabled\":false", "\"disabled\":null"), new string(' ', 16385)
            })
            {
                File.WriteAllText(path, text);
                bool refused = false;
                try { agent.Memory().ReadSettings(); }
                catch (Exception e) when (e is IOException or InvalidDataException or JsonException or FormatException
                    or ArgumentException or InvalidOperationException) { refused = true; }
                Check(refused, "Malformed settings accepted");
            }
            File.WriteAllText(path, valid);
            Check(agent.Memory().ReadSettings()!.Memory.Conversation.Budget == 3000, "Valid document refused");
            File.Delete(path);
            Check(agent.Memory().ReadSettings() is null, "Absent settings did not select defaults");
            File.WriteAllText(path, valid.Replace("\"version\":1", "\"version\":2"));
            try
            {
                await agent.WorkStatusAsync().WaitAsync(TimeSpan.FromSeconds(10));
            }
            catch (InvalidDataException)
            {
                startupRefused = true;
            }
            Check(startupRefused && !agent.History().Events().Any(e => e.Kind == EventKind.RequestSent),
                "Invalid saved policy did not fail startup before inference");
        }
        finally
        {
            try { app.Destroy(); }
            catch (Exception error) when (startupRefused && error.GetBaseException() is InvalidDataException)
            {
                // Startup faults remain observable during joined destruction.
            }
        }
    }

    private static async Task SettingsRefuses(Func<Task> action)
    {
        try { await action().WaitAsync(TimeSpan.FromSeconds(10)); }
        catch (Exception e) when (e is IOException or InvalidDataException or UnauthorizedAccessException
            or ArgumentException or InvalidOperationException) { return; }
        throw new Exception("Expected settings refusal");
    }
}
