using System.Collections.Immutable;
using System.Text;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class JournalRetentionTests
{
    private static void Check(bool value, string message)
    {
        if (!value)
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
        catch (Exception error) when (error is IOException or InvalidDataException or FormatException or ArgumentException)
        {
            return;
        }
        throw new Exception("Expected archive refusal");
    }

    private static string DirectoryForTest()
    {
        string path = Path.Combine(Path.GetTempPath(), "archive-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(path);
        return path;
    }

    private static Agent Agent(Application app)
    {
        return new Agent(app, new Configuration(Vendor.OpenAI, "offline", "http://127.0.0.1:1"));
    }

    private static Event Message(string text)
    {
        return new Event(EventKind.MessageReceived, actor: Actor.Human, parts: [new TextPart(text)]);
    }

    private static string Log(IEnumerable<Event> facts)
    {
        return "{\"log_version\":1}\n" + string.Concat(facts.Select(value => JsonCodec.WriteEvent(value) + "\n"));
    }

    private static string View(History history)
    {
        return JsonCodec.WriteSave(new SaveDocument(history.View(), history.Snapshot().Sequence,
            history.Parent().Engine().SavedConfiguration(history.View())));
    }

    public static void Replay()
    {
        string directory = DirectoryForTest();
        string path = Path.Combine(directory, "save.json");
        string archive = path + ".journal.jsonl";
        var policy = new PersistenceOptions(path, Audit: true);
        var app = new Application(TextWriter.Null);
        try
        {
            var source = Agent(app);
            var history = source.History();
            history.StartJournal(path);
            var original = new List<Event>();
            var requests = new List<string>();
            Event Record(Event value)
            {
                var recorded = history.Append(value);
                original.Add(recorded);
                return recorded;
            }
            void Request()
            {
                requests.Add(source.Engine().Render(history.View()).ToJsonString());
                Record(new Event(EventKind.RequestSent, from: source.Configuration().Target()));
            }

            Record(Message("retain the conversation"));
            Request();
            var call = new ToolCallPart("read-1", source.Configuration().Target(), "read_file",
                JsonSerializer.SerializeToElement(new { path = "owned.txt" }));
            var response = Record(new Event(EventKind.ResponseEnded, parts: [new TextPart("reading"), call]));
            Record(new Event(EventKind.ToolCalled, callID: call.CallID, name: call.Name, args: call.Args));
            Record(new Event(EventKind.ToolReturned,
                callID: call.CallID, parts: [new TextPart("ORIGINAL-TOOL-BYTES" + new string('x', 4096))]));
            Request();
            Record(new Event(EventKind.ResponseEnded, parts: [new TextPart("finished")]));
            history.Save(policy);
            Check(history.Events().Length == original.Count, "History trimmed before a handoff boundary");

            Record(new Event(EventKind.Redacted, redactFrom: response.Seq, redactTo: original[^1].Seq,
                level: Redaction.Result, reason: "archive regression"));
            Request();
            Record(new Event(EventKind.ResponseEnded, parts: [new TextPart("after stubbing")]));
            history.Save(policy);
            Record(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("continue here")]));
            Request();
            Record(new Event(EventKind.ResponseEnded, parts: [new TextPart("after handoff")]));
            history.Save(policy);
            Check(File.ReadAllText(archive) == string.Concat(original.Select(fact => JsonCodec.WriteEvent(fact) + "\n")), "Archive changed or omitted canonical events");
            Check(!View(history).Contains("ORIGINAL-TOOL-BYTES"), "Working Context was not actually reduced");
            Check(File.ReadAllText(archive).Contains("ORIGINAL-TOOL-BYTES"), "Audit lost original payload");

            var replay = Agent(app);
            replay.History().Load(archive);
            Check(View(replay.History()) == View(history), "Full archive replay differs from saved working Context");
            var context = new Context();
            var reconstructed = new List<string>();
            foreach (var value in replay.History().Events())
            {
                if (value.Kind == EventKind.RequestSent)
                {
                    reconstructed.Add(replay.Engine().Render(context).ToJsonString());
                }
                context = History.Reduce(context, value);
            }
            Check(reconstructed.SequenceEqual(requests), "Archived request prefixes do not reproduce original projections");
            Check(reconstructed[1].Contains("ORIGINAL-TOOL-BYTES") && !reconstructed[2].Contains("ORIGINAL-TOOL-BYTES"),
                "Replay did not distinguish full and stubbed request windows");

            // Audit restart reuses the same journal, never a second archive.
            var restored = Agent(app).History();
            Check(restored.LoadSave(path), "Snapshot-only restart failed");
            restored.StartJournal(path);
            original.Add(restored.Append(Message("after restart")));
            restored.Save(policy);
            Check(File.ReadAllText(archive) == string.Concat(original.Select(fact => JsonCodec.WriteEvent(fact) + "\n")), "Restart lost the archived prefix or duplicated its suffix");
            replay.History().Load(archive);
            Check(View(replay.History()) == View(restored), "Restart archive replay differs");
        }
        finally
        {
            app.Destroy();
            Directory.Delete(directory, recursive: true);
        }
    }

    public static void BackupBoundary()
    {
        foreach (bool audit in new[] { false, true })
        {
            string directory = DirectoryForTest();
            string path = Path.Combine(directory, "save.json");
            string journal = path + ".journal.jsonl";
            var app = new Application(TextWriter.Null);
            try
            {
                var history = Agent(app).History();
                var policy = new PersistenceOptions(path, Audit: audit);
                history.StartJournal(path);
                history.Append(Message("original"));
                history.Save(policy); // Backup will deliberately lag the handoff window.
                history.Append(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("first")]));
                history.Append(Message("between"));
                var preceding = history.Append(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("second")]));
                history.Append(Message("recent"));
                history.Append(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("third")]));
                history.Append(Message("latest"));
                Check(history.Events()[0].Seq == preceding.Seq, "Live events exceed the two-handoff window");
                history.Save(policy);
                var lines = File.ReadAllLines(journal).Select(JsonCodec.ReadEvent).ToArray();
                Check(lines[0].Seq == (audit ? 1 : 2), "Journal lost the lagging backup suffix");
                string expected = View(history);
                File.WriteAllText(path, "damaged primary");
                var recovered = Agent(app).History();
                Check(recovered.LoadSave(path) && View(recovered) == expected, "Backup recovery lost state");
                Check(recovered.Events()[0].Seq == preceding.Seq, "Restart loaded audit prefix into working History");
                recovered.StartJournal(path);
                recovered.Save(policy); // Invalid primary must not displace the old backup.
                Check(JsonCodec.ReadSave(File.ReadAllText(path + ".bak")).Sequence == 1, "Invalid primary replaced backup");
                recovered.Save(policy); // Now the backup can advance and permit the working cutoff.
                lines = File.ReadAllLines(journal).Select(JsonCodec.ReadEvent).ToArray();
                Check(lines[0].Seq == (audit ? 1 : preceding.Seq), "Wrong handoff/audit journal cutoff");
                foreach (string snapshot in new[] { path, path + ".bak" })
                {
                    using var json = JsonDocument.Parse(File.ReadAllText(snapshot));
                    Check(!json.RootElement.TryGetProperty("log", out _), "Snapshot duplicates event data");
                }
                Check(!File.Exists(Path.Combine(directory, "events.jsonl"))
                    && !Directory.Exists(Path.Combine(directory, "archive")), "Duplicate event output exists");
                // Enabling audit late preserves what remains; it cannot invent a deleted prefix.
                string before = File.ReadAllText(journal);
                recovered.Save(policy with { Audit = true });
                Check(File.ReadAllText(journal) == before, "Audit rewrote retained journal");
            }
            finally
            {
                app.Destroy();
                Directory.Delete(directory, recursive: true);
            }
        }
    }
}
