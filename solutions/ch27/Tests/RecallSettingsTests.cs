using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    public static async Task RecallSettingsPersistence(string temp)
    {
        string home = Path.Combine(temp, "recall-settings-persistence");
        var memory = new MemorySettings(soul: new MemoryBandSettings(disabled: true));
        var recall = new RecallSettings(disabled: true, disableJudge: true, judgeModel: "saved-model",
            minScore: 1.25, maxSnippets: 7, minQueryLength: 17, candidates: 11,
            contextMessages: 4, maxRecallBytes: 2300);
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = MemoryAgent(app, home, "http://127.0.0.1:1");
            // A durable recall command may be the very first actor operation.
            await agent.SaveRecallSettingsAsync(recall);
            await agent.SaveMemorySettingsAsync(memory);
            // Transient changes must not leak into the sibling's next save.
            agent.SetRecallSettings(new RecallSettings(judgeModel: "transient"));
            await agent.SaveMemorySettingsAsync(memory);
            await agent.SetMemorySettingsAsync(new MemorySettings());
            await agent.SaveRecallSettingsAsync(recall);
            var disk = agent.Memory().ReadSettings()!;
            Check(disk.Memory.Soul.Disabled && disk.Recall.JudgeModel == "saved-model",
                "Sibling save persisted transient policy");
            Check(JsonCodec.WriteRecallSettings(disk.Recall).ToJsonString()
                == JsonCodec.WriteRecallSettings(recall).ToJsonString(), "Recall fields lost on disk");
        }
        finally { app.Destroy(); }
        app = new Application(TextWriter.Null);
        try
        {
            var agent = MemoryAgent(app, home, "http://127.0.0.1:1");
            // Memory updates have their own startup path. It must still restore
            // saved recall, without applying the old memory policy first.
            await agent.SaveMemorySettingsAsync(new MemorySettings());
            Check(agent.RecallSettings().Disabled && agent.RecallSettings().JudgeModel == "saved-model",
                "First memory save erased stored recall");
            Check(!agent.MemorySettings().Soul.Disabled, "First memory save ignored requested policy");
        }
        finally { app.Destroy(); }
        app = new Application(TextWriter.Null);
        try
        {
            var agent = MemoryAgent(app, home, "http://127.0.0.1:1");
            await agent.WorkStatusAsync();
            Check(agent.RecallSettings().MaxRecallBytes == 2300 && agent.RecallSettings().Disabled,
                "Restart did not restore recall before admission");
            var noHome = new Agent(app, new Configuration(Vendor.OpenAI, "offline", "http://127.0.0.1:1"));
            var before = noHome.RecallSettings();
            await SettingsRefuses(() => noHome.SaveRecallSettingsAsync(recall));
            Check(ReferenceEquals(before, noHome.RecallSettings()), "No-home save published policy");
        }
        finally { app.Destroy(); }
    }

    public static async Task RecallSettingsCodec(string temp)
    {
        string home = Path.Combine(temp, "recall-settings-codec");
        var app = new Application(TextWriter.Null);
        bool startupRefused = false;
        try
        {
            var agent = MemoryAgent(app, home, "http://127.0.0.1:1");
            string path = Path.Combine(home, "settings.json");
            var root = new JsonObject { ["version"] = 1,
                ["memory"] = JsonCodec.WriteMemorySettings(new MemorySettings()) };
            File.WriteAllText(path, root.ToJsonString());
            var legacy = agent.Memory().ReadSettings()!;
            Check(!legacy.Recall.Disabled && legacy.Recall.MinQueryLength == 80,
                "Memory-only document did not supply default recall");
            root["recall"] = JsonCodec.WriteRecallSettings(new RecallSettings());
            string valid = root.ToJsonString();
            foreach (string text in new[] {
                valid.Replace("\"recall\":{", "\"recall\":null,\"recall\":{"),
                valid.Replace("\"judge_model\":\"\"", "\"judge_model\":null"),
                valid.Replace("\"min_score\":3", "\"min_score\":1e999"),
                valid.Replace("\"max_snippets\":3", "\"max_snippets\":-1"),
                valid.Replace("\"max_snippets\":3", "\"max_snippets\":1.5"),
                valid.Replace("\"candidates\":20,", ""),
                valid.Replace("\"context_messages\":3", "\"context_messages\":3,\"extra\":true"),
                valid.Replace("\"disable_judge\":false", "\"disable_judge\":false,\"disable_judge\":true") })
            {
                File.WriteAllText(path, text);
                bool refused = false;
                try { agent.Memory().ReadSettings(); }
                catch (Exception e) when (e is InvalidDataException or JsonException or FormatException
                    or ArgumentException or InvalidOperationException or OverflowException) { refused = true; }
                Check(refused, "Malformed recall document accepted: " + text);
            }
            // Genuine absence and explicit zeros are defaults, but malformed
            // supplied policy must fail the first actor barrier, before a send.
            var zero = JsonCodec.WriteRecallSettings(new RecallSettings());
            foreach (string name in new[] { "min_score", "max_snippets", "min_query_length", "candidates",
                "context_messages", "max_recall_bytes" }) zero[name] = 0;
            using var doc = JsonDocument.Parse(zero.ToJsonString());
            Check(JsonCodec.ReadRecallSettings(doc.RootElement).Candidates == 20, "Zero defaults lost");
            root["recall"] = null;
            File.WriteAllText(path, root.ToJsonString());
            try { await agent.WorkStatusAsync().WaitAsync(TimeSpan.FromSeconds(5)); }
            catch (FormatException) { startupRefused = true; }
            Check(startupRefused && !agent.History().Events().Any(e => e.Kind == EventKind.RequestSent),
                "Malformed recall reached inference");
        }
        finally
        {
            try { app.Destroy(); }
            catch (Exception e) when (startupRefused && e.GetBaseException() is FormatException) { }
        }
    }

    public static async Task RecallSettingsDamagedSibling(string temp)
    {
        var app = new Application(TextWriter.Null);
        try
        {
            string home = Path.Combine(temp, "recall-settings-damaged");
            var agent = MemoryAgent(app, home, "http://127.0.0.1:1");
            await agent.WorkStatusAsync();
            var before = agent.MemorySettings();
            string path = Path.Combine(home, "settings.json");
            foreach (string damaged in new[] { "{", "{\"version\":1,\"memory\":null}" })
            {
                File.WriteAllText(path, damaged);
                bool refused = false;
                try { await agent.SaveMemorySettingsAsync(new MemorySettings(soul: new MemoryBandSettings(disabled: true))); }
                catch (Exception e) when (e is JsonException or FormatException) { refused = true; }
                Check(refused && ReferenceEquals(before, agent.MemorySettings())
                    && File.ReadAllText(path) == damaged, "Damaged sibling save changed policy");
                // A live save refusal is not a fatal startup failure.
                await agent.WorkStatusAsync().WaitAsync(TimeSpan.FromSeconds(5));
            }
            File.Delete(path);
            await agent.SaveRecallSettingsAsync(new RecallSettings(disabled: true));
        }
        finally
        {
            try { app.Destroy(); }
            catch (Exception e) when (e.GetBaseException() is JsonException or FormatException) { }
        }
    }

    public static async Task RecallSettingsDuringTurn(string temp)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            string home = Path.Combine(temp, "recall-settings-live");
            var agent = RecallActor(app, listener, home);
            var turn = agent.SubmitAsync([new TextPart(RecallQuery)]);
            var judge = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
            await ReadRecallRequest(judge);
            var previous = agent.RecallSettings();
            string path = Path.Combine(home, "settings.json");
            Directory.CreateDirectory(path);
            await SettingsRefuses(() => agent.SaveRecallSettingsAsync(new RecallSettings(disabled: true)));
            Check(ReferenceEquals(previous, agent.RecallSettings()), "Failed save changed live policy");
            Directory.Delete(path);
            await agent.SaveRecallSettingsAsync(new RecallSettings(disabled: true)).WaitAsync(TimeSpan.FromSeconds(3));
            Check(agent.RecallSettings().Disabled && agent.Memory().ReadSettings()!.Recall.Disabled,
                "Durable acknowledgment preceded save or publication");
            string saved = File.ReadAllText(path);
            await SettingsRefuses(() => agent.SaveRecallSettingsAsync(new RecallSettings(judgeModel: new string('x', 20000))));
            Check(File.ReadAllText(path) == saved && agent.RecallSettings().Disabled, "Oversized save altered policy");
            await SendJudgeReply(judge, Surface.ChatCompletions, "[0]", "judge-route");
            var main = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
            var request = await ReadRecallRequest(main);
            Check(request.ToJsonString().Contains(RecallFact, StringComparison.Ordinal),
                "Settings update changed captured in-flight policy");
            await SendJudgeReply(main, Surface.ChatCompletions, "done", "main-route");
            await turn.WaitAsync(TimeSpan.FromSeconds(5));
            // A fresh fact proves the next long turn bypasses search, rather
            // than being skipped merely because the first fact is retained.
            File.WriteAllText(Path.Combine(home, "docs", "decision.md"), "quartz migration NEW_DISABLED_FACT");
            var next = agent.SubmitAsync([new TextPart(RecallQuery)]);
            var nextRequest = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
            var body = await ReadRecallRequest(nextRequest);
            Check(body["model"]!.GetValue<string>() == "main-route"
                && !body.ToJsonString().Contains("NEW_DISABLED_FACT", StringComparison.Ordinal),
                "Next turn ignored disabled recall");
            await SendJudgeReply(nextRequest, Surface.ChatCompletions, "next", "main-route");
            await next.WaitAsync(TimeSpan.FromSeconds(5));
        }
        finally { app.Destroy(); }
    }
}
