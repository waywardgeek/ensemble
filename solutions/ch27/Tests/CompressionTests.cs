using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class CompressionTests
{
    private static readonly DateTimeOffset Epoch = DateTimeOffset.Parse("2026-01-01T00:00:00Z");
    private static void Check(bool condition, string message)
    {
        if (!condition)
        {
            throw new Exception(message);
        }
    }

    private static void Refuses(Action action)
    {
        try
        {
            action();
        }
        catch (ArgumentException) { return; }
        catch (FormatException) { return; }
        throw new Exception("Invalid compression fact accepted");
    }

    private static CompressionLaunch Conversation(string task = "task") => new(task, null, MemoryBand.Session,
        "session-one", Epoch, new string('a', 64), 1, 3, tracePath: "memory/compactor-task.jsonl");
    private static Event Cut(string task = "task", long to = 3) => new(EventKind.Redacted,
        redactFrom: 1, redactTo: to, level: Redaction.Summary, reason: "memory compression", taskID: task);
    private static MemoryEntry Replacement(CompressionLaunch launch, string text = "summary") =>
        new(launch.Destination, launch.ID, launch.Timestamp, text);
    private static Event Population(CompressionLaunch launch) => new(EventKind.MemoryPopulated,
        memory: Replacement(launch), taskID: launch.TaskID);
    private static Event Finished(CompressionLaunch launch, CompressionState state) =>
        new(EventKind.CompressionFinished, outcome: new CompressionOutcome(launch.TaskID, state));
    private static Context Source() => new([
        new Entry(1, Actor.Human, [new TextPart("source dialogue")]),
        new Entry(2, Actor.System, [new SkillPart("keep", "authoritative manual", ["read_file"])]),
        new Entry(3, Actor.System, [new HandoffPart("source reasons")]),
        new Entry(4, Actor.Human, [new TextPart("unfinished tail")])]);

    public static void ConversationReplay()
    {
        var source = Source();
        var launch = Conversation();
        var launched = History.Reduce(source, new Event(EventKind.CompressionLaunched, compression: launch));
        var held = History.Reduce(launched, Cut().Numbered(6, Epoch));
        Check(held.Dialogue.SequenceEqual(source.Dialogue) && held.Compression?.Removal is not null
            && held.Memories.IsEmpty, "Held cut removed active source");
        var withAppend = History.Reduce(held, new Event(EventKind.MessageReceived, seq: 7,
            actor: Actor.Human, parts: [new TextPart("unrelated later work")]));
        var committed = History.Reduce(withAppend, Population(launch));
        Check(committed.Dialogue.Select(entry => entry.Seq).SequenceEqual([2L, 4L, 7L])
            && committed.EnabledTools().Contains("read_file"), "Commit removed instruction or unfinished/appended work");
        Check(committed.Memories.Single().Text == "summary" && committed.Compression?.Committed?.Text == "summary",
            "Commit failed to publish replacement and retained receipt together");
        Check(held.Dialogue.Length == 4 && held.Memories.IsEmpty, "Commit mutated an older snapshot");
        var handoff = History.Reduce(committed, new Event(EventKind.HandoffRecorded, seq: 8, parts: [new HandoffPart("later checkpoint")]));
        Check(handoff.Compression?.Committed?.Text == "summary", "Handoff discarded recovery receipt");
        var final = History.Reduce(handoff, Finished(launch, CompressionState.Committed));
        Check(final.Compression is null && final.CompressionOutcome?.State == CompressionState.Committed
            && final.Memories.Single().Text == "summary", "Finalization changed active memory or retained live task");
        Check(History.Reduce(final, Finished(launch, CompressionState.Committed)).Memories.Length == 1,
            "Repeated finalization duplicated content");
        var failed = History.Reduce(held, Finished(launch, CompressionState.Failed));
        Check(failed.Dialogue.SequenceEqual(source.Dialogue) && failed.Compression is null
            && failed.Memories.IsEmpty, "Failed compression lost source or retained a live task");
        var legacy = History.Reduce(source, new Event(EventKind.MemoryPopulated, memory: Replacement(launch)));
        Check(legacy.Dialogue.SequenceEqual(source.Dialogue), "Independent disk population borrowed a conversation cut");
    }

    public static void GraduationReplay()
    {
        var source = new Context(memories: [
            new MemoryEntry(MemoryBand.Session, "one", Epoch, "old one"),
            new MemoryEntry(MemoryBand.Session, "two", Epoch.AddSeconds(1), "old two"),
            new MemoryEntry(MemoryBand.Session, "three", Epoch.AddSeconds(2), "newer")]);
        var launch = new CompressionLaunch("graduate", MemoryBand.Session, MemoryBand.Archive8,
            "archive-one", Epoch, new string('b', 64), sourceIDs: ["one", "two"], tracePath: "memory/trace.jsonl");
        var launched = History.Reduce(source, new Event(EventKind.CompressionLaunched, compression: launch));
        var removal = new Event(EventKind.MemoryRetired, taskID: "graduate",
            retirement: new MemoryRetirement(MemoryBand.Session, ["one", "two"]));
        var held = History.Reduce(launched, removal);
        Check(held.Memories.SequenceEqual(source.Memories), "Graduation retired before population");
        var committed = History.Reduce(held, Population(launch));
        Check(committed.Memories.Select(item => item.ID).SequenceEqual(["archive-one", "three"])
            && committed.Compression?.Committed?.Band == MemoryBand.Archive8, "Graduation was not atomic or removed newer memory");
        var canceled = History.Reduce(held, Finished(launch, CompressionState.Canceled));
        Check(canceled.Memories.SequenceEqual(source.Memories), "Canceled graduation lost source");
        Refuses(() => History.Reduce(launched, new Event(EventKind.MemoryRetired, taskID: "graduate",
            retirement: new MemoryRetirement(MemoryBand.Session))));
        Refuses(() => History.Reduce(launched, new Event(EventKind.MemoryRetired, taskID: "graduate",
            retirement: new MemoryRetirement(MemoryBand.Session, ["two", "one"]))));
    }

    public static void Correlation()
    {
        var source = Source();
        var launch = Conversation();
        var start = new Event(EventKind.CompressionLaunched, compression: launch);
        var launched = History.Reduce(source, start);
        var held = History.Reduce(launched, Cut());
        var committed = History.Reduce(held, Population(launch));
        Refuses(() => History.Reduce(source, Cut()));
        Refuses(() => History.Reduce(source, Population(launch)));
        Refuses(() => History.Reduce(launched, Population(launch)));
        Refuses(() => History.Reduce(launched, Cut("wrong")));
        Refuses(() => History.Reduce(launched, Cut(to: 4)));
        Refuses(() => History.Reduce(launched, Cut(task: "")));
        Refuses(() => History.Reduce(launched, start));
        Refuses(() => History.Reduce(held, Cut()));
        Refuses(() => History.Reduce(held, new Event(EventKind.MemoryPopulated, taskID: launch.TaskID,
            memory: Replacement(launch, "   "))));
        Refuses(() => History.Reduce(held, new Event(EventKind.MemoryPopulated, taskID: launch.TaskID,
            memory: new MemoryEntry(MemoryBand.Memory, launch.ID, launch.Timestamp, "wrong band"))));
        Refuses(() => History.Reduce(held, Finished(launch, CompressionState.Committed)));
        Refuses(() => History.Reduce(committed, Finished(launch, CompressionState.Invalidated)));
        Refuses(() => History.Reduce(committed, Population(launch)));
        Refuses(() => History.Reduce(held, Finished(Conversation("other"), CompressionState.Failed)));
        var done = History.Reduce(committed, Finished(launch, CompressionState.Committed));
        Refuses(() => History.Reduce(done, start));
        Refuses(() => new PendingCompression(launch, committed: Replacement(launch)));
        Refuses(() => new CompressionLaunch("task", null, MemoryBand.Session, "one", Epoch, "not a digest", 1, 3, tracePath: "trace"));
        Refuses(() => new CompressionLaunch("task", MemoryBand.Memory, MemoryBand.Soul, "one", Epoch,
            new string('a', 64), sourceIDs: ["one", "two"], tracePath: "trace"));
        var malformed = JsonNode.Parse(JsonCodec.WriteEvent(Cut().Numbered(12, Epoch)))!;
        malformed["task_id"] = null;
        Refuses(() => JsonCodec.ReadEvent(malformed.ToJsonString()));
        malformed["task_id"] = "";
        Refuses(() => JsonCodec.ReadEvent(malformed.ToJsonString()));
        malformed["task_id"] = 123;
        Refuses(() => JsonCodec.ReadEvent(malformed.ToJsonString()));
        Refuses(() => new Context(compression: held.Compression,
            compressionOutcome: new CompressionOutcome(launch.TaskID, CompressionState.Failed)));
        foreach (var value in new[] { start, Cut(), Population(launch), Finished(launch, CompressionState.Failed) })
        {
            var original = value.Numbered(12, Epoch).WithDisplay(0, []);
            var encoded = JsonCodec.WriteEvent(original);
            var decoded = JsonCodec.ReadEvent(encoded);
            Check(JsonCodec.WriteEvent(decoded) == encoded, "Event copy/codec dropped transaction identity");
        }
    }

    public static void Persistence()
    {
        string directory = Path.Combine(Path.GetTempPath(), "compression-replay-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(directory);
        var app = new Application(TextWriter.Null);
        try
        {
            var config = new Configuration(Vendor.OpenAI, "test", "http://127.0.0.1:1", "unused");
            var agent = new Agent(app, config);
            string path = Path.Combine(directory, "source.json");
            agent.StartPersistence(new PersistenceOptions(path));
            var history = agent.History();
            history.Append(new Event(EventKind.MessageReceived, actor: Actor.Human, parts: [new TextPart("source dialogue")]));
            history.Append(new Event(EventKind.SkillLoaded, parts: [new SkillPart("keep", "manual", ["read_file"])]));
            history.Append(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("source reasons")]));
            var launch = Conversation();
            history.Append(new Event(EventKind.CompressionLaunched, compression: launch));
            long sequence = history.Snapshot().Sequence;
            string journalBefore = File.ReadAllText(path + ".journal.jsonl");
            Refuses(() => history.Append(Cut("wrong")));
            Check(history.Snapshot().Sequence == sequence
                && File.ReadAllText(path + ".journal.jsonl") == journalBefore,
                "Rejected fact advanced sequence or wrote journal bytes");
            history.Append(Cut());
            string replay = Path.Combine(directory, "journal-only.json");
            File.Copy(path + ".journal.jsonl", replay + ".journal.jsonl");
            var restored = new Agent(app, config);
            Check(restored.History().LoadSave(replay) && restored.History().View().Compression?.Removal is not null
                && restored.History().View().Dialogue.Any(entry => entry.Parts.Any(part => part is HandoffPart)),
                "Journal-only held cut lost source/pending identity");
            history.Save(new PersistenceOptions(path, RetainLog: 0));
            var heldSave = JsonCodec.ReadSave(File.ReadAllText(path));
            Check(heldSave.Context!.Compression?.Removal?.TaskID == "task" && heldSave.Events.IsEmpty,
                "Trimmed snapshot lost held removal");
            history.Append(Population(launch));
            history.Save(new PersistenceOptions(path, RetainLog: 0));
            var saved = JsonCodec.ReadSave(File.ReadAllText(path));
            Check(saved.Context!.Compression?.Committed?.Text == "summary" && saved.Context.EnabledTools().Contains("read_file"),
                "Snapshot lost committed receipt or authority");
            Check(restored.History().LoadSave(path) && restored.History().View().Compression?.Committed?.Text == "summary",
                "Restore did not preserve committed-but-unfinalized receipt");
            // A retained outcome survives trimming; loading does not run recovery.
            history.Append(Finished(launch, CompressionState.Committed));
            history.Save(new PersistenceOptions(path, RetainLog: 0));
            Check(restored.History().LoadSave(path) && restored.History().View().Compression is null
                && restored.History().View().CompressionOutcome?.State == CompressionState.Committed,
                "Final receipt/outcome did not survive save/load");
            var bad = JsonNode.Parse(JsonCodec.WriteSave(saved))!;
            bad["context"]!["compression"]!["removal"]!["task_id"] = "wrong";
            Refuses(() => JsonCodec.ReadSave(bad.ToJsonString()));
            var old = JsonNode.Parse(JsonCodec.WriteSave(saved))!;
            old["context"]!.AsObject().Remove("compression");
            old["context"]!.AsObject().Remove("compression_outcome");
            Check(JsonCodec.ReadSave(old.ToJsonString()).Context!.Compression is null, "Older snapshot requires new transaction fields");
        }
        finally
        {
            app.Destroy();
            Directory.Delete(directory, true);
        }
    }
}
