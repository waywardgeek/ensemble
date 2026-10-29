using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    private static Agent MemoryAgent(Application app, string home, string endpoint)
    {
        Directory.CreateDirectory(home);
        return new Agent(app, new Configuration(Vendor.OpenAI, "fixture", endpoint,
            "local-fixture", disableStreaming: true, sandboxRoot: home), dataDirectory: home);
    }

    // The peer implements the real Chat Completions wire boundary. No model or
    // provider is contacted, and assertions examine the request actually sent.
    private static async Task MemoryTurn(Agent agent, System.Net.HttpListener listener,
        string expected, string? absent = null)
    {
        var turn = agent.SubmitAsync([new TextPart("Read the current memory")]);
        var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
        using var reader = new StreamReader(request.Request.InputStream);
        string wire = await reader.ReadToEndAsync();
        Check(wire.Contains(expected, StringComparison.Ordinal), "Current disk memory missing from wire");
        Check(absent is null || !wire.Contains(absent, StringComparison.Ordinal), "Stale memory reached wire");
        await Reply(request, JsonSerializer.Serialize(new
        {
            choices = new[] { new { message = new { content = "memory observed" }, finish_reason = "stop" } }
        }));
        await turn.WaitAsync(TimeSpan.FromSeconds(10));
        await agent.FlushAsync().WaitAsync(TimeSpan.FromSeconds(10));
    }

    public static async Task MemoryStartupAndPolicy(string directory)
    {
        string home = Path.Combine(directory, "memory-startup");
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = MemoryAgent(app, home, listener.Prefixes.Single());
            var history = agent.History();
            history.Append(new Event(EventKind.MemoryPopulated,
                memory: new MemoryEntry(MemoryBand.Soul, "SOUL", DateTimeOffset.UnixEpoch, "STALE_SOUL")));
            history.Append(new Event(EventKind.MemoryPopulated,
                memory: new MemoryEntry(MemoryBand.Session, "2026-01-01-1", DateTimeOffset.UnixEpoch, "DELETED_MEMORY")));
            File.WriteAllText(Path.Combine(home, "SOUL.md"), "FRESH_SOUL");
            await MemoryTurn(agent, listener, "FRESH_SOUL", "STALE_SOUL");
            Check(!history.View().Memories.Any(entry => entry.Text == "DELETED_MEMORY"), "Missing file survived startup");

            int requests = history.Events().Count(fact => fact.Kind == EventKind.RequestSent);
            var off = new MemorySettings(soul: new MemoryBandSettings(disabled: true));
            await agent.SetMemorySettingsAsync(off).WaitAsync(TimeSpan.FromSeconds(10));
            Check(ReferenceEquals(off, agent.MemorySettings()) && history.View().Memories.Length == 0,
                "Idle disable was not applied before acknowledgment");
            Check(File.ReadAllText(Path.Combine(home, "SOUL.md")) == "FRESH_SOUL", "Disable deleted corpus");
            File.WriteAllText(Path.Combine(home, "SOUL.md"), "EDITED_WHILE_OFF");
            await agent.SetMemorySettingsAsync(new MemorySettings()).WaitAsync(TimeSpan.FromSeconds(10));
            Check(history.View().Memories.Single().Text == "EDITED_WHILE_OFF", "Re-enable reused stale replay bytes");
            long sequence = history.Events().Length;
            await agent.SetMemorySettingsAsync(new MemorySettings()).WaitAsync(TimeSpan.FromSeconds(10));
            Check(sequence == history.Events().Length, "Unchanged policy generated redundant facts");
            Check(history.Events().Count(fact => fact.Kind == EventKind.RequestSent) == requests,
                "Idle policy update started inference");
            await MemoryTurn(agent, listener, "EDITED_WHILE_OFF", "FRESH_SOUL");
        }
        finally
        {
            app.Destroy();
        }
    }

    public static async Task MemoryPolicyRefusal(string directory)
    {
        string home = Path.Combine(directory, "memory-policy-refusal");
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = MemoryAgent(app, home, "http://127.0.0.1:1");
            File.WriteAllBytes(Path.Combine(home, "SOUL.md"), [0xff]);
            // The first admitted settings update must govern startup scans too.
            var off = new MemorySettings(soul: new MemoryBandSettings(disabled: true),
                memory: new MemoryBandSettings(disabled: true));
            await agent.SetMemorySettingsAsync(off).WaitAsync(TimeSpan.FromSeconds(10));
            File.WriteAllText(Path.Combine(home, "MEMORY.md"), "valid other band");
            long sequence = agent.History().Events().Length;
            bool refused = false;
            try
            {
                await agent.SetMemorySettingsAsync(new MemorySettings()).WaitAsync(TimeSpan.FromSeconds(10));
            }
            catch (System.Text.DecoderFallbackException)
            {
                refused = true;
            }
            Check(refused && ReferenceEquals(off, agent.MemorySettings()) && sequence == agent.History().Events().Length,
                "Failed multi-band scan partially applied policy or History");
            File.WriteAllText(Path.Combine(home, "SOUL.md"), "repaired");
            await agent.SetMemorySettingsAsync(new MemorySettings()).WaitAsync(TimeSpan.FromSeconds(10));
            Check(agent.History().View().Memories.Length == 2, "Refused update damaged actor reuse");
            using var cancelled = new CancellationTokenSource();
            cancelled.Cancel();
            try
            {
                await agent.SetMemorySettingsAsync(off, cancelled.Token);
                throw new Exception("Pre-cancelled settings accepted");
            }
            catch (OperationCanceledException) { }
            Check(!agent.MemorySettings().Soul.Disabled, "Pre-cancelled update changed policy");
            Check(new MemorySettings().Conversation.Budget == 3000
                && new MemorySettings().Session.HighWatermark == 3000, "Wrong byte defaults");
            bool invalid = false;
            try { _ = new MemorySettings(session: new MemoryBandSettings(10, 5, 6)); }
            catch (ArgumentException) { invalid = true; }
            Check(invalid, "Invalid effective watermarks accepted");
        }
        finally { app.Destroy(); }
    }

    private static MemoryCapture PrepareMemoryRecovery(Agent agent, string home, bool commit)
    {
        string corpus = Path.Combine(home, "memory");
        Directory.CreateDirectory(corpus);
        File.WriteAllText(Path.Combine(corpus, "2026-01-01-1.md"), "SOURCE_ONE");
        File.WriteAllText(Path.Combine(corpus, "2026-01-01-2.md"), "SOURCE_TWO");
        foreach (var entry in agent.Memory().Scan(agent.History().View(), MemoryBand.Session))
        {
            agent.History().Append(new Event(EventKind.MemoryPopulated, memory: entry));
        }
        var capture = agent.Memory().CaptureGraduation(agent.History().View(), MemoryBand.Session, 2, 1,
            DateTimeOffset.UtcNow) ?? throw new Exception("Missing selection");
        agent.History().Append(new Event(EventKind.CompressionLaunched, compression: capture.Launch));
        agent.Memory().Stage(capture, "COMMITTED_MEMORY");
        if (commit)
        {
            agent.History().Append(new Event(EventKind.MemoryRetired,
                retirement: new MemoryRetirement(MemoryBand.Session, capture.Launch.SourceIDs), taskID: capture.Launch.TaskID));
            agent.History().Append(new Event(EventKind.MemoryPopulated,
                memory: new MemoryEntry(capture.Launch.Destination, capture.Launch.ID, capture.Launch.Timestamp,
                    "COMMITTED_MEMORY"), taskID: capture.Launch.TaskID));
        }
        return capture;
    }

    public static async Task MemoryActorRecovery(string directory)
    {
        foreach (bool commit in new[] { false, true })
        {
            string home = Path.Combine(directory, "memory-actor-recovery-" + commit);
            using var listener = Listen();
            var app = new Application(TextWriter.Null);
            var first = MemoryAgent(app, home, listener.Prefixes.Single());
            var capture = PrepareMemoryRecovery(first, home, commit);
            string path = Path.Combine(home, "save.json");
            first.History().Save(new PersistenceOptions(path, RetainLog: 0));
            app.Destroy();
            app = new Application(TextWriter.Null);
            try
            {
                var agent = MemoryAgent(app, home, listener.Prefixes.Single());
                agent.History().LoadSave(path);
                agent.StartPersistence(new PersistenceOptions(path));
                Check(agent.History().View().Compression is not null, "Offline replay unexpectedly recovered disk");
                await MemoryTurn(agent, listener, commit ? "COMMITTED_MEMORY" : "SOURCE_ONE",
                    commit ? "SOURCE_ONE" : "COMMITTED_MEMORY");
                var context = agent.History().View();
                Check(context.Compression is null && context.CompressionOutcome?.TaskID == capture.Launch.TaskID
                    && context.CompressionOutcome.State == (commit ? CompressionState.Committed : CompressionState.Abandoned),
                    "Actor did not journal recovery outcome");
                Check(!File.Exists(Path.Combine(home, ".memory-stage.json")), "Terminal stage not released");
                Check(File.Exists(Path.Combine(home, "memory/2026-01-01-1.md.retired")) == commit,
                    "Wrong source retirement after startup recovery");
                Check(agent.Children().Length == 0, "Recovery respawned a worker");
            }
            finally { app.Destroy(); }
        }
    }

    public static Task MemoryCanonicalArchive(string directory)
    {
        string home = Path.Combine(directory, "memory-canonical-archive");
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = MemoryAgent(app, home, "http://127.0.0.1:1");
            var history = agent.History();
            history.Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("original source")]));
            history.Append(new Event(EventKind.ResponseEnded, actor: Actor.Agent,
                parts: [new TextPart("visible answer", from: agent.Configuration().Target(),
                    opaque: JsonSerializer.SerializeToElement(new { private_replay = "EXACT_PRIVATE_BYTES" }))]));
            history.Append(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("checkpoint")]));
            Check(history.View().Dialogue.SelectMany(entry => entry.Parts).OfType<TextPart>()
                .All(part => part.Opaque is null), "Fixture handoff did not strip opaque material");
            history.ArchiveMemory();
            string archive = Path.Combine(home, "archive/events.jsonl");
            var lines = File.ReadAllLines(archive).Skip(1).ToArray();
            Check(lines.SequenceEqual(history.Events().Select(JsonCodec.WriteEvent)), "Archive replaced canonical events with reduced text");
            Check(File.ReadAllText(archive).Contains("EXACT_PRIVATE_BYTES", StringComparison.Ordinal),
                "Archive lost private canonical source");
            history.Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("later source")]));
            history.ArchiveMemory();
            string complete = File.ReadAllText(archive);
            Check(File.ReadAllLines(archive).Length == history.Events().Length + 1, "Archive append duplicated a prefix");
            string save = Path.Combine(home, "save.json");
            history.Save(new PersistenceOptions(save, RetainLog: 0));
            app.Destroy();
            app = new Application(TextWriter.Null);
            agent = MemoryAgent(app, home, "http://127.0.0.1:1");
            agent.History().LoadSave(save);
            agent.History().ArchiveMemory();
            Check(File.ReadAllText(archive) == complete, "Archive failed to bridge a retained snapshot");
            File.Delete(archive);
            bool refused = false;
            try { agent.History().ArchiveMemory(); }
            catch (InvalidDataException) { refused = true; }
            Check(refused && !File.Exists(archive), "Missing canonical history was silently reconstructed");
            agent.StartPersistence(new PersistenceOptions(save));
            string before = File.ReadAllText(save);
            File.CreateSymbolicLink(archive, save);
            bool aliasRefused = false;
            try { agent.History().ArchiveMemory(); }
            catch (ArgumentException) { aliasRefused = true; }
            Check(aliasRefused && File.ReadAllText(save) == before, "Archive alias overwrote the active primary");
        }
        finally { app.Destroy(); }
        return Task.CompletedTask;
    }

    public static async Task MemoryRecoveryFailure(string directory)
    {
        string home = Path.Combine(directory, "memory-recovery-failure");
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = MemoryAgent(app, home, "http://127.0.0.1:1");
            PrepareMemoryRecovery(agent, home, commit: true);
            File.Delete(Path.Combine(home, ".memory-stage.json"));
            var task = agent.SubmitAsync([new TextPart("must never send")]);
            bool refused = false;
            try { await task.WaitAsync(TimeSpan.FromSeconds(10)); }
            catch (IOException) { refused = true; }
            Check(refused, "Recovery failure stranded the triggering caller");
            try
            {
                await agent.ShutdownAsync().WaitAsync(TimeSpan.FromSeconds(10));
                throw new Exception("Faulted actor shutdown concealed recovery failure");
            }
            catch (IOException) { }
            Check(!agent.History().Events().Any(fact => fact.Kind == EventKind.RequestSent), "Failed recovery sent inference");
            Check(File.Exists(Path.Combine(home, "memory/2026-01-01-1.md")), "Failed recovery removed source");
        }
        finally
        {
            bool retained = false;
            try { app.Destroy(); }
            catch (AggregateException error)
            {
                retained = error.Flatten().InnerExceptions.All(item => item is IOException);
                if (!retained) throw;
            }
            Check(retained && app.Agents().Length == 0, "Faulted cleanup lost error or ownership detach");
        }
    }
}
