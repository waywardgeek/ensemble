using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

// Replay and request projection use real History/Engine components. There is no
// filesystem scanner or compressor in this foundation slice to fake or credit.
internal static class MemoryBandTests
{
    private static readonly DateTimeOffset Epoch = new(2026, 10, 28, 0, 0, 0, TimeSpan.Zero);

    private static void Check(bool value, string message)
    {
        if (!value) throw new Exception(message);
    }

    private static void Refuses(Action action)
    {
        try { action(); }
        catch (Exception error) when (error is ArgumentException or FormatException or JsonException or InvalidOperationException)
        {
            return;
        }
        throw new Exception("Expected invalid memory to refuse");
    }

    private static MemoryEntry Item(MemoryBand band, string id, int day, string? text = null) =>
        new(band, id, Epoch.AddDays(day), text ?? $"payload-{band}-{id}");

    private static Event Populate(MemoryEntry entry) => new(EventKind.MemoryPopulated, memory: entry);

    public static void Replay()
    {
        var dialogue = new Entry(1, Actor.Human, [new TextPart("current task")]);
        var before = new Context([dialogue], [new TextPart("ephemeral")], Phase.ToolsPending, new Usage(12, 2, 3, 4));
        var events = new[]
        {
            Populate(Item(MemoryBand.Session, "later", 3)),
            Populate(Item(MemoryBand.Archive8, "eight", 2)),
            Populate(Item(MemoryBand.Soul, "human", 9)),
            Populate(Item(MemoryBand.Archive64, "sixtyfour", 1)),
            Populate(Item(MemoryBand.Memory, "human", 9)),
            Populate(Item(MemoryBand.Session, "earlier", 0))
        };
        Context folded = before;
        foreach (var fact in events)
        {
            folded = History.Reduce(folded, JsonCodec.ReadEvent(JsonCodec.WriteEvent(fact.Numbered(2, Epoch))));
        }
        Check(folded.Memories.Select(item => item.Band).SequenceEqual(new[]
        {
            MemoryBand.Soul, MemoryBand.Memory, MemoryBand.Archive64, MemoryBand.Archive8,
            MemoryBand.Session, MemoryBand.Session
        }), "Band order follows completion order");
        Check(folded.Memories[^2].ID == "earlier" && folded.Memories[^1].ID == "later", "Within-band chronology lost");
        Check(before.Memories.IsEmpty && ReferenceEquals(folded.Dialogue[0], dialogue), "Reduction mutated prior data");
        Check(folded.Ephemeral.SequenceEqual(before.Ephemeral) && folded.Phase == before.Phase
            && ReferenceEquals(folded.Usage, before.Usage), "Memory changed actor/request state");
        Check(folded.EnabledTools().IsEmpty, "Memory granted tools");

        var updated = Item(MemoryBand.Session, "later", 3, "edited\r\nexact\n");
        folded = History.Reduce(folded, Populate(updated));
        folded = History.Reduce(folded, Populate(updated));
        Check(folded.Memories.Length == 6 && folded.Memories[^1].Text == updated.Text, "Edited/replayed identity duplicated or normalized");
        folded = History.Reduce(folded, new Event(EventKind.MemoryRetired,
            retirement: new MemoryRetirement(MemoryBand.Session, ["earlier", "missing"])));
        Check(folded.Memories.Length == 5 && folded.Memories[^1].ID == "later", "Targeted retirement removed wrong entries");
        folded = History.Reduce(folded, new Event(EventKind.MemoryRetired,
            retirement: new MemoryRetirement(MemoryBand.Session)));
        Check(folded.Memories.Length == 4 && folded.Dialogue.Length == 1, "Band retirement cut dialogue");
        Check(ReferenceEquals(folded, History.Reduce(folded, new Event(EventKind.MemoryRetired,
            retirement: new MemoryRetirement(MemoryBand.Session)))), "Inapplicable retirement was not a diagnosed no-op");
        Check(ReferenceEquals(folded, History.Reduce(folded, new Event(EventKind.MemoryPopulated))), "Missing body was not a no-op");
        var handed = History.Reduce(folded, new Event(EventKind.HandoffRecorded, seq: 99,
            parts: [new HandoffPart("continue") ]));
        Check(handed.Memories.SequenceEqual(folded.Memories), "Manual handoff erased separate memory bands");
    }

    public static void ValuesAndCodec()
    {
        Refuses(() => Item((MemoryBand)0, "x", 0));
        Refuses(() => Item(MemoryBand.Session, " ", 0));
        Refuses(() => new MemoryRetirement(MemoryBand.Session, ["x", "x"]));
        Refuses(() => new MemoryRetirement((MemoryBand)100));
        var item = Item(MemoryBand.Session, "one", 0);
        Refuses(() => new Context(memories: [item, item]));
        Refuses(() => new Context(memories: [null!]));
        Refuses(() => new Event(EventKind.TurnEnded, memory: item));
        Refuses(() => new Event(EventKind.MemoryPopulated, retirement: new MemoryRetirement(MemoryBand.Session)));
        var ids = new[] { "one", "two" };
        var retirement = new MemoryRetirement(MemoryBand.Session, ids);
        ids[0] = "mutated";
        Check(retirement.IDs[0] == "one", "Retirement borrowed mutable input");
        var recorded = new Event(EventKind.MemoryRetired, retirement: retirement).Numbered(8, Epoch);
        var decoded = JsonCodec.ReadEvent(JsonCodec.WriteEvent(recorded));
        Check(decoded.Retirement!.Band == MemoryBand.Session && decoded.Retirement.IDs.SequenceEqual(retirement.IDs), "Retirement codec lost identities");
        var bad = JsonNode.Parse(JsonCodec.WriteEvent(Populate(item).Numbered(3, Epoch)))!;
        bad["memory"]!["band"] = "unknown";
        Refuses(() => JsonCodec.ReadEvent(bad.ToJsonString()));
        bad["memory"]!["band"] = "session";
        bad["memory"]!["timestamp"] = "not a date";
        Refuses(() => JsonCodec.ReadEvent(bad.ToJsonString()));
        // Equal instants normalize identically regardless of source timezone.
        var offset = new MemoryEntry(MemoryBand.Soul, "one", Epoch.ToOffset(TimeSpan.FromHours(3)), "text");
        Check(offset.Timestamp.Offset == TimeSpan.Zero && offset.Timestamp == Epoch, "Timestamp normalization is nondeterministic");
    }

    public static void Persistence()
    {
        string directory = Path.Combine(Path.GetTempPath(), "memory-replay-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(directory);
        var app = new Application(TextWriter.Null);
        try
        {
            var config = new Configuration(Vendor.OpenAI, "test", "http://127.0.0.1:1", "unused");
            var agent = new Agent(app, config);
            var history = agent.History();
            history.Append(Populate(Item(MemoryBand.Session, "first", 0, "private memory bytes")));
            history.Append(new Event(EventKind.MessageReceived, actor: Actor.Human, parts: [new TextPart("task")]));
            string path = Path.Combine(directory, "save.json");
            history.Save(new PersistenceOptions(path, RetainLog: 0));
            var saved = JsonCodec.ReadSave(File.ReadAllText(path));
            Check(saved.Events.IsEmpty && saved.Context!.Memories.Length == 1, "Trimmed save dropped memory projection");
            Check(!saved.Configuration.GetRawText().Contains("private memory bytes"), "Memory entered configuration summary");
            var restored = new Agent(app, config);
            Check(restored.History().LoadSave(path), "Actual snapshot did not restore");
            Check(restored.History().View().Memories.Single().Text == "private memory bytes", "Snapshot memory changed");
            var newer = Populate(Item(MemoryBand.Session, "first", 0, "new text")).Numbered(saved.Sequence + 1, Epoch);
            File.WriteAllText(path, JsonCodec.WriteSave(saved with { Events = [newer] }));
            Check(restored.History().LoadSave(path) && restored.History().View().Memories.Single().Text == "new text", "Tail update lost or duplicated");
            var old = JsonNode.Parse(JsonCodec.WriteSave(saved))!;
            old["context"]!.AsObject().Remove("memories");
            Check(JsonCodec.ReadSave(old.ToJsonString()).Context!.Memories.IsEmpty, "Pre-memory snapshot refused");
            var malformed = JsonNode.Parse(JsonCodec.WriteSave(saved))!;
            malformed["context"]!["memories"] = new JsonArray(
                malformed["context"]!["memories"]![0]!.DeepClone(),
                malformed["context"]!["memories"]![0]!.DeepClone());
            Refuses(() => JsonCodec.ReadSave(malformed.ToJsonString()));
        }
        finally
        {
            app.Destroy();
            Directory.Delete(directory, true);
        }
    }

    public static void Journal()
    {
        string directory = Path.Combine(Path.GetTempPath(), "memory-journal-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(directory);
        var app = new Application(TextWriter.Null);
        try
        {
            var config = new Configuration(Vendor.OpenAI, "test", "http://127.0.0.1:1", "unused");
            var source = new Agent(app, config);
            string path = Path.Combine(directory, "source.json");
            source.StartPersistence(new PersistenceOptions(path));
            source.History().Append(Populate(Item(MemoryBand.Session, "old", 0)));
            source.History().Append(Populate(Item(MemoryBand.Session, "retained", 1)));
            source.History().Append(new Event(EventKind.MemoryRetired,
                retirement: new MemoryRetirement(MemoryBand.Session, ["old"])));
            // Copy only committed bytes to an independent reader. There is no
            // shutdown snapshot and no edit to the source journal under its writer.
            string replay = Path.Combine(directory, "recovered.json");
            File.Copy(path + ".journal.jsonl", replay + ".journal.jsonl");
            var restored = new Agent(app, config);
            Check(restored.History().LoadSave(replay), "Journal-only memory recovery failed");
            var memories = restored.History().View().Memories;
            Check(memories.Length == 1 && memories[0].ID == "retained", "Journal retirement did not replay");
            long sequence = restored.History().Snapshot().Sequence;
            Check(restored.History().LoadSave(replay)
                && restored.History().Snapshot().Sequence == sequence
                && restored.History().View().Memories.Length == 1, "Repeated recovery duplicated facts");
        }
        finally
        {
            app.Destroy();
            Directory.Delete(directory, true);
        }
    }

    public static void Wire()
    {
        var entries = new[]
        {
            Item(MemoryBand.Session, "one", 0, "unique-session"),
            Item(MemoryBand.Archive8, "one", 0, "unique-eight"),
            Item(MemoryBand.Archive64, "one", 0, "unique-sixtyfour"),
            Item(MemoryBand.Memory, "one", 0, "unique-memory"),
            Item(MemoryBand.Soul, "one", 0, "unique-soul")
        };
        var context = new Context([new Entry(1, Actor.Human, [new TextPart("unique-conversation")])], memories: entries);
        var app = new Application(TextWriter.Null);
        try
        {
            foreach (var (vendor, surface) in new[]
            {
                (Vendor.OpenAI, Surface.ChatCompletions), (Vendor.OpenAI, Surface.Responses),
                (Vendor.Anthropic, Surface.Messages), (Vendor.Gemini, Surface.GenerateContent)
            })
            {
                var agent = new Agent(app, new Configuration(vendor, "test", "http://127.0.0.1:1", "unused",
                    primaryInstructions: "trusted-primary", surface: surface));
                string wire = agent.Engine().Render(context).ToJsonString();
                using var parsed = JsonDocument.Parse(wire);
                var root = parsed.RootElement;
                string instruction = surface == Surface.Responses ? root.GetProperty("instructions").GetRawText()
                    : vendor == Vendor.Anthropic ? root.GetProperty("system").GetRawText()
                    : vendor == Vendor.Gemini ? root.GetProperty("systemInstruction").GetRawText()
                    : root.GetProperty("messages")[0].GetRawText();
                Check(!instruction.Contains("unique-") && instruction.Contains("trusted-primary"), "Memory entered trusted instructions");
                int previous = -1;
                foreach (string token in new[] { "unique-soul", "unique-memory", "unique-sixtyfour", "unique-eight", "unique-session", "unique-conversation" })
                {
                    int at = wire.IndexOf(token, StringComparison.Ordinal);
                    Check(at > previous && wire.LastIndexOf(token, StringComparison.Ordinal) == at, $"Memory order/duplication on {surface}");
                    previous = at;
                }
                var messages = root.GetProperty(vendor == Vendor.Gemini ? "contents" : surface == Surface.Responses ? "input" : "messages");
                foreach (var message in messages.EnumerateArray())
                {
                    if (message.GetRawText().Contains("unique-"))
                    {
                        Check(message.GetProperty("role").GetString() == "user", "Memory acquired privileged role");
                    }
                }
                agent.Destroy();
            }
        }
        finally { app.Destroy(); }
    }
}
