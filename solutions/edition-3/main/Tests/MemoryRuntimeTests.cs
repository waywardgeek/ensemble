using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    private static MemorySettings RuntimePolicy(bool sessionOff = false, bool soulOff = true) => new(
        conversation: new MemoryBandSettings(disabled: true),
        soul: new MemoryBandSettings(disabled: soulOff),
        session: new MemoryBandSettings(1, 2, 1, sessionOff));

    private static async Task<Agent> RuntimeAgent(Application app, string home, string endpoint)
    {
        var agent = MemoryAgent(app, home, endpoint);
        Directory.CreateDirectory(Path.Combine(home, "memory"));
        File.WriteAllText(Path.Combine(home, "memory/2026-01-01-1.md"), "SOURCE_ONE");
        File.WriteAllText(Path.Combine(home, "memory/2026-01-01-2.md"), "SOURCE_TWO");
        agent.StartPersistence(new PersistenceOptions(Path.Combine(home, "save.json")));
        _ = new Tool(agent, ToolKind.MicroHandoff);
        await agent.SetMemorySettingsAsync(RuntimePolicy()).WaitAsync(TimeSpan.FromSeconds(10));
        return agent;
    }

    private static async Task RuntimeCheckpoint(Agent agent)
    {
        var call = new ToolCallPart(Guid.NewGuid().ToString("N"), agent.Configuration().Target(), "micro_handoff",
            JsonSerializer.SerializeToElement(new { text = "Continue the existing work carefully." }));
        Check(!(await agent.DispatchAsync(call).WaitAsync(TimeSpan.FromSeconds(10))).IsError, "Checkpoint refused");
        await agent.FlushAsync().WaitAsync(TimeSpan.FromSeconds(10));
    }

    // Poll authoritative state with an actor barrier and a bounded deadline;
    // elapsed time never stands in for a lifecycle or commit assertion.
    private static async Task RuntimeUntil(Agent agent, Func<bool> ready)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(10));
        while (!ready())
        {
            await agent.WorkStatusAsync(deadline.Token);
            await Task.Delay(5, deadline.Token);
        }
    }

    public static async Task MemoryRuntimeCommit(string directory)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        string home = Path.Combine(directory, "runtime-commit");
        var agent = await RuntimeAgent(app, home, listener.Prefixes.Single());
        try
        {
            await RuntimeCheckpoint(agent);
            var compression = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            string taskID = agent.History().View().Compression!.Launch.TaskID;
            Check(agent.Children().Length == 1, "Worker was not owned while waiting");

            // A parent turn proceeds independently. Its transport deliberately
            // stays pending after the compressor has finished and detached.
            var turn = agent.SubmitAsync([new TextPart("NEWER_UNRELATED_WORK")]);
            var main = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            using (var reader = new StreamReader(main.Request.InputStream))
            {
                string wire = await reader.ReadToEndAsync();
                Check(wire.Contains("SOURCE_ONE") && wire.Contains("NEWER_UNRELATED_WORK"), "Source disappeared while worker ran");
            }
            await Reply(compression, WorkerChat("{\"memory\":\"RETAINED_SUMMARY\"}"));
            await RuntimeUntil(agent, () => agent.Children().Length == 0);
            Check(agent.History().View().Compression?.Committed is null
                && agent.History().View().Memories.Count(entry => entry.Band == MemoryBand.Session) == 2,
                "Completion committed through an unfinished main request");
            await Reply(main, "{\"choices\":[{\"message\":{\"content\":\"main finished\"},\"finish_reason\":\"stop\"}]}");
            await turn.WaitAsync(TimeSpan.FromSeconds(10));
            await RuntimeUntil(agent, () => agent.History().View().CompressionOutcome?.TaskID == taskID);
            var context = agent.History().View();
            Check(context.Compression is null && context.CompressionOutcome!.State == CompressionState.Committed,
                "Compression did not finish");
            Check(context.Memories.Single(entry => entry.Band == MemoryBand.Archive8).Text == "RETAINED_SUMMARY"
                && !context.Memories.Any(entry => entry.Band == MemoryBand.Session), "Replacement was not atomic");
            Check(context.Dialogue.SelectMany(entry => entry.Parts).OfType<TextPart>()
                .Any(part => part.Text.Contains("NEWER_UNRELATED_WORK")), "Later dialogue was removed");
            Check(File.Exists(Path.Combine(home, "memory/2026-01-01-1.md.retired"))
                && !File.Exists(Path.Combine(home, ".memory-stage.json")), "Disk finalization incomplete");
            var facts = agent.History().Events().Where(fact => fact.TaskID == taskID).ToArray();
            Check(facts.Length == 2 && facts[0].Kind == EventKind.MemoryRetired
                && facts[1].Kind == EventKind.MemoryPopulated, "Correlated removal/population order wrong");
        }
        finally { app.Destroy(); }
    }

    public static async Task MemoryRuntimePolicy(string directory)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        string home = Path.Combine(directory, "runtime-policy");
        var agent = await RuntimeAgent(app, home, listener.Prefixes.Single());
        try
        {
            await RuntimeCheckpoint(agent);
            var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            string taskID = agent.History().View().Compression!.Launch.TaskID;
            File.WriteAllText(Path.Combine(home, "SOUL.md"), "FRESH_UNRELATED_BAND");
            string settingsPath = Path.Combine(home, "settings.json");
            Directory.CreateDirectory(settingsPath);
            await SettingsRefuses(() => agent.SaveMemorySettingsAsync(RuntimePolicy(sessionOff: true)));
            Check(agent.History().View().Compression?.Launch.TaskID == taskID && agent.Children().Length == 1
                && !agent.MemorySettings().Session.Disabled,
                "Failed settings save canceled live compression");
            Directory.Delete(settingsPath);
            await agent.SaveMemorySettingsAsync(RuntimePolicy(soulOff: false)).WaitAsync(TimeSpan.FromSeconds(10));
            Check(agent.History().View().Compression?.Launch.TaskID == taskID && agent.Children().Length == 1,
                "Unrelated re-enable invalidated live work");
            Check(agent.History().View().Memories.Any(entry => entry.Text == "FRESH_UNRELATED_BAND"),
                "Unrelated re-enable did not scan fresh bytes");
            await agent.SaveMemorySettingsAsync(RuntimePolicy(sessionOff: true, soulOff: false))
                .WaitAsync(TimeSpan.FromSeconds(10));
            Check(agent.Children().Length == 0 && agent.History().View().Compression is null,
                "Policy acknowledged before joined cancellation");
            Check(agent.History().View().CompressionOutcome?.State == CompressionState.Invalidated
                && !agent.History().View().Memories.Any(entry => entry.Band == MemoryBand.Session),
                "Invalidation/disable was not applied");
            Check(File.ReadAllText(Path.Combine(home, "memory/2026-01-01-1.md")) == "SOURCE_ONE",
                "Invalidation removed source file");
            request.Response.Abort();
            await agent.SetMemorySettingsAsync(RuntimePolicy(soulOff: false));
            await agent.WorkStatusAsync();
            Check(agent.Children().Length == 0 && agent.History().Events().Count(e => e.Kind == EventKind.CompressionLaunched) == 1,
                "Re-enable retried without a checkpoint");

            // Public History can contain a pending launch without this actor
            // owning its worker. Settings must not mutate that transaction.
            var orphan = agent.Memory().CaptureGraduation(agent.History().View(), MemoryBand.Session,
                2, 1, DateTimeOffset.UtcNow)!;
            agent.History().Append(new Event(EventKind.CompressionLaunched, compression: orphan.Launch));
            string savedPolicy = File.ReadAllText(settingsPath);
            bool refused = false;
            try
            {
                await agent.SaveMemorySettingsAsync(RuntimePolicy(sessionOff: true, soulOff: false))
                    .WaitAsync(TimeSpan.FromSeconds(10));
            }
            catch (InvalidOperationException) { refused = true; }
            Check(refused && File.ReadAllText(settingsPath) == savedPolicy && !agent.MemorySettings().Session.Disabled
                && agent.History().View().Compression?.Launch.TaskID == orphan.Launch.TaskID,
                "Settings mutated a pending replay transaction with no live capture");
        }
        finally { app.Destroy(); }
    }

    public static async Task MemoryRuntimeStale(string directory)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        string home = Path.Combine(directory, "runtime-stale");
        var agent = await RuntimeAgent(app, home, listener.Prefixes.Single());
        try
        {
            await RuntimeCheckpoint(agent);
            var first = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            File.WriteAllText(Path.Combine(home, "memory/2026-01-01-1.md"), "HUMAN_EDIT");
            await Reply(first, WorkerChat("{\"memory\":\"STALE_SUMMARY\"}"));
            await RuntimeUntil(agent, () => agent.History().View().CompressionOutcome is not null);
            Check(agent.History().View().CompressionOutcome!.State == CompressionState.Invalidated,
                "Stale source committed");
            Check(!agent.History().View().Memories.Any(entry => entry.Text == "STALE_SUMMARY")
                && agent.History().View().Memories.Count(entry => entry.Band == MemoryBand.Session) == 2,
                "Stale completion cut replay source");
            await agent.WorkStatusAsync();
            Check(agent.History().Events().Count(e => e.Kind == EventKind.CompressionLaunched) == 1,
                "Failure retried itself");
            // Restore selected bytes; a later checkpoint, not a completion loop,
            // authorizes a fresh attempt with its own task and child identity.
            File.WriteAllText(Path.Combine(home, "memory/2026-01-01-1.md"), "SOURCE_ONE");
            await RuntimeCheckpoint(agent);
            var second = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            await Reply(second, WorkerChat("{\"memory\":\"FRESH_SUMMARY\"}"));
            await RuntimeUntil(agent, () => agent.History().View().CompressionOutcome?.State == CompressionState.Committed);
            Check(agent.History().Events().Count(e => e.Kind == EventKind.CompressionLaunched) == 2,
                "Fresh checkpoint did not permit exactly one retry");
        }
        finally { app.Destroy(); }
    }

    public static async Task MemoryRuntimeShutdown(string directory)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        string home = Path.Combine(directory, "runtime-shutdown");
        var agent = await RuntimeAgent(app, home, listener.Prefixes.Single());
        try
        {
            await RuntimeCheckpoint(agent);
            var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            await agent.ShutdownAsync().WaitAsync(TimeSpan.FromSeconds(10));
            Check(agent.Children().Length == 0 && agent.History().View().Compression is null,
                "Shutdown left a live worker or pending transaction");
            Check(agent.History().View().CompressionOutcome?.State == CompressionState.Canceled,
                "Shutdown did not record cancellation");
            Check(File.Exists(Path.Combine(home, "memory/2026-01-01-1.md")), "Shutdown removed source");
            request.Response.Abort();
        }
        finally { app.Destroy(); }
    }
    public static async Task MemoryRuntimeConversation(string directory)
    {
        foreach (bool blockArchive in new[] { false, true })
        {
            using var listener = Listen();
            var app = new Application(TextWriter.Null);
            try
            {
                string home = Path.Combine(directory, "runtime-conversation-" + blockArchive);
                var agent = MemoryAgent(app, home, listener.Prefixes.Single());
                agent.StartPersistence(new PersistenceOptions(Path.Combine(home, "save.json")));
                _ = new Tool(agent, ToolKind.MicroHandoff);
                await agent.SetMemorySettingsAsync(new MemorySettings(conversation: new MemoryBandSettings(1)));
                agent.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                    parts: [new TextPart("ORIGINAL_CONVERSATION_FACT")]));
                agent.History().Append(new Event(EventKind.ResponseEnded, actor: Actor.Agent,
                    parts: [new TextPart("visible answer", from: agent.Configuration().Target(),
                        opaque: JsonSerializer.SerializeToElement(new { replay = "CANONICAL_OPAQUE_ENVELOPE" }))]));
                if (blockArchive)
                {
                    File.WriteAllText(Path.Combine(home, "archive"), "Fixture blocks directory creation");
                }
                await RuntimeCheckpoint(agent);
                var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
                Check(agent.History().View().Dialogue.SelectMany(entry => entry.Parts).OfType<TextPart>()
                    .Any(text => text.Text.Contains("ORIGINAL_CONVERSATION_FACT")), "Checkpoint eagerly cut source");
                await Reply(request, WorkerChat("{\"memory\":\"CONVERSATION_MEMORY\"}"));
                await RuntimeUntil(agent, () => agent.History().View().CompressionOutcome is not null);
                var context = agent.History().View();
                bool sourceRetained = context.Dialogue.SelectMany(entry => entry.Parts).OfType<TextPart>()
                    .Any(text => text.Text.Contains("ORIGINAL_CONVERSATION_FACT"));
                if (blockArchive)
                {
                    Check(context.CompressionOutcome!.State == CompressionState.Failed && sourceRetained,
                        "Failed archive retired conversation source");
                    Check(!context.Memories.Any(entry => entry.Text == "CONVERSATION_MEMORY")
                        && !File.Exists(Path.Combine(home, ".memory-stage.json")), "Failed preparation exposed a candidate");
                    await agent.WorkStatusAsync();
                    Check(agent.History().Events().Count(e => e.Kind == EventKind.CompressionLaunched) == 1,
                        "Archive failure retried before another checkpoint");
                }
                else
                {
                    Check(context.CompressionOutcome!.State == CompressionState.Committed && !sourceRetained,
                        "Conversation replacement did not commit");
                    var memory = context.Memories.Single(entry => entry.Band == MemoryBand.Session);
                    Check(memory.Text == "CONVERSATION_MEMORY"
                        && File.ReadAllText(Path.Combine(home, "memory", memory.ID + ".md")) == memory.Text,
                        "Conversation memory not promoted");
                    string archive = File.ReadAllText(Path.Combine(home, "archive/events.jsonl"));
                    Check(archive.Contains("ORIGINAL_CONVERSATION_FACT") && archive.Contains("CANONICAL_OPAQUE_ENVELOPE")
                        && archive.Contains("micro_handoff"), "Archive substituted reduced compressor input for canonical events");
                }
            }
            finally { app.Destroy(); }
        }
    }

    public static async Task MemoryRuntimeCascade(string directory)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            string home = Path.Combine(directory, "runtime-cascade");
            Directory.CreateDirectory(Path.Combine(home, "bucket-0"));
            File.WriteAllText(Path.Combine(home, "bucket-0/2025-01-01-1--2025-01-01-2.md"), "OLDER_ARCHIVE");
            var agent = await RuntimeAgent(app, home, listener.Prefixes.Single());
            await agent.SetMemorySettingsAsync(new MemorySettings(
                conversation: new MemoryBandSettings(disabled: true),
                session: new MemoryBandSettings(1, 2, 1), archive8: new MemoryBandSettings(1, 2, 1)));
            await RuntimeCheckpoint(agent);
            var first = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            await Reply(first, WorkerChat("{\"memory\":\"SESSION_SUMMARY\"}"));
            var second = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            Check(agent.History().View().Compression!.Launch.Source == MemoryBand.Archive8
                && agent.Children().Length == 1, "Successful commit did not start the next eligible band alone");
            using (var reader = new StreamReader(second.Request.InputStream))
            {
                string wire = await reader.ReadToEndAsync();
                Check(wire.Contains("OLDER_ARCHIVE") && wire.Contains("SESSION_SUMMARY"), "Cascade captured the wrong active files");
            }
            await Reply(second, WorkerChat("{\"memory\":\"CASCADE_SUMMARY\"}"));
            await RuntimeUntil(agent, () => agent.History().View().Compression is null);
            Check(agent.History().View().Memories.Single(entry => entry.Band == MemoryBand.Archive64).Text == "CASCADE_SUMMARY"
                && !agent.History().View().Memories.Any(entry => entry.Band is MemoryBand.Session or MemoryBand.Archive8),
                "Cascade left duplicate active source");
            Check(agent.History().Events().Count(e => e.Kind == EventKind.CompressionLaunched) == 2
                && agent.Children().Length == 0, "Cascade launched extra work or retained a child");
            var receipts = agent.History().Events().Where(e => e.Kind == EventKind.CompressionFinished)
                .SelectMany(e => e.Outcome!.Usage).ToArray();
            Check(receipts.Length == 2 && agent.Engine().Meter().Totals.AllInput()
                == receipts.Sum(row => row.Counts.AllInput()) && agent.Engine().Meter().Latest is null,
                "Waterfall reset lost or duplicated a worker's usage");
        }
        finally { app.Destroy(); }
    }
}
