using System.Collections.Immutable;
using System.Text;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class ArchiveTests
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
            history.Parent().Engine().SavedConfiguration(history.View()), []));
    }

    public static void Replay()
    {
        string directory = DirectoryForTest();
        string path = Path.Combine(directory, "save.json");
        string archive = Path.Combine(directory, "archive.jsonl");
        var policy = new PersistenceOptions(path, 0, archive);
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
            Check(history.Events().IsEmpty, "Zero diagnostic retention was not exercised");

            Record(new Event(EventKind.Redacted, redactFrom: response.Seq, redactTo: original[^1].Seq,
                level: Redaction.Result, reason: "archive regression"));
            Request();
            Record(new Event(EventKind.ResponseEnded, parts: [new TextPart("after stubbing")]));
            history.Save(policy);
            Record(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("continue here")]));
            Request();
            Record(new Event(EventKind.ResponseEnded, parts: [new TextPart("after handoff")]));
            history.Save(policy);
            Check(File.ReadAllText(archive) == Log(original), "Archive changed or omitted canonical events");
            Check(!View(history).Contains("ORIGINAL-TOOL-BYTES"), "Working Context was not actually reduced");
            Check(!File.ReadAllText(path + ".journal.jsonl").Contains("ORIGINAL-TOOL-BYTES"),
                "Fixture did not trim the journal past the original payload");

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

            // A zero-log restart has only its recovery suffix in memory. Earlier
            // facts now come from the archive, not invented snapshot messages.
            var restored = Agent(app).History();
            Check(restored.LoadSave(path), "Snapshot-only restart failed");
            restored.StartJournal(path);
            original.Add(restored.Append(Message("after restart")));
            restored.Save(policy);
            Check(File.ReadAllText(archive) == Log(original), "Restart lost the archived prefix or duplicated its suffix");
            replay.History().Load(archive);
            Check(View(replay.History()) == View(restored), "Restart archive replay differs");
        }
        finally
        {
            app.Destroy();
        }
    }

    public static void Refusal()
    {
        string directory = DirectoryForTest();
        string path = Path.Combine(directory, "save.json");
        string archive = Path.Combine(directory, "archive.jsonl");
        var policy = new PersistenceOptions(path, 0, archive);
        var app = new Application(TextWriter.Null);
        try
        {
            var history = Agent(app).History();
            history.StartJournal(path);
            var first = history.Append(Message("first"));
            history.Save(policy);
            var second = history.Append(Message("second"));
            history.Save(policy);
            var third = history.Append(Message("third"));
            string primary = File.ReadAllText(path);
            string backup = File.ReadAllText(path + ".bak");
            string journal = File.ReadAllText(path + ".journal.jsonl");
            string goodArchive = File.ReadAllText(archive);
            string view = View(history);
            var diagnostics = history.Events();
            var conflict = new Event(EventKind.MessageReceived, seq: second.Seq, time: second.Time,
                actor: Actor.Human, parts: [new TextPart("CONFLICTING-PRIVATE-CONTENT")]);
            var future = Message("future").Numbered(third.Seq + 1, DateTimeOffset.UtcNow);
            foreach (string invalid in new[]
            {
                Log([second]), // Missing prefix, even though the journal has this fact.
                Log([first, conflict]), // Same sequence, different canonical bytes.
                Log([first, second, third, future]), // Not this History's current boundary.
                "{\"log_version\":999}\n",
                "{\"log_version\":1}\nnot-json\n"
            })
            {
                File.WriteAllText(archive, invalid);
                Refuses(() => history.Save(policy));
                Check(File.ReadAllText(archive) == invalid, "Rejected archive was overwritten");
                Check(File.ReadAllText(path) == primary && File.ReadAllText(path + ".bak") == backup &&
                    File.ReadAllText(path + ".journal.jsonl") == journal, "Archive refusal changed recovery files");
                Check(View(history) == view && history.Events().SequenceEqual(diagnostics), "Archive refusal trimmed live state");
            }
            byte[] invalidUtf8 = [.. Encoding.UTF8.GetBytes("{\"log_version\":1}\n{\"type\":\""), 0xff];
            File.WriteAllBytes(archive, invalidUtf8);
            Refuses(() => history.Save(policy));
            Check(File.ReadAllBytes(archive).SequenceEqual(invalidUtf8) && File.ReadAllText(path) == primary,
                "Malformed UTF-8 was repaired or published silently");
            File.WriteAllText(archive, goodArchive);
            // The archive is published first. A later snapshot failure may leave
            // it ahead of the primary, but must not trim the journal or live log.
            File.Move(path + ".bak", path + ".bak.held");
            Directory.CreateDirectory(path + ".bak");
            try
            {
                Refuses(() => history.Save(policy));
                Check(File.ReadAllText(archive) == Log([first, second, third]), "Archive did not precede backup rotation");
                Check(File.ReadAllText(path) == primary && File.ReadAllText(path + ".journal.jsonl") == journal,
                    "Later save failure changed primary or trimmed recovery journal");
                Check(history.Events().SequenceEqual(diagnostics), "Later save failure trimmed diagnostics");
            }
            finally
            {
                Directory.Delete(path + ".bak");
                File.Move(path + ".bak.held", path + ".bak");
            }
            history.Save(policy);
            Check(File.ReadAllText(archive) == Log([first, second, third]), "Valid archive could not continue after refusals");

            // Starting a full archive late cannot conjure the prefix removed by
            // both diagnostic and backup-floor journal retention.
            string late = Path.Combine(directory, "late.jsonl");
            Refuses(() => history.Save(policy with { ArchivePath = late }));
            Check(!File.Exists(late), "Incomplete archive was published as full history");

            // Direct snapshot saves have no active journal; this is also explicit
            // refusal rather than a dialogue-only export masquerading as replay.
            string directPath = Path.Combine(directory, "direct.json");
            var direct = Agent(app).History();
            direct.Append(Message("no longer retained"));
            direct.Save(new PersistenceOptions(directPath, 0));
            string directSnapshot = File.ReadAllText(directPath);
            Refuses(() => direct.Save(new PersistenceOptions(directPath, 0, late)));
            Check(File.ReadAllText(directPath) == directSnapshot && !File.Exists(late), "Snapshot-only gap changed files");

            string log = Path.Combine(directory, "active.jsonl");
            var logging = Agent(app).History();
            logging.StartLog(log);
            logging.Append(Message("still writable"));
            string before = File.ReadAllText(log);
            Refuses(() => logging.Save(new PersistenceOptions(log)));
            Refuses(() => logging.Save(new PersistenceOptions(Path.Combine(directory, "log-save.json"), 0, log)));
            Check(File.ReadAllText(log) == before, "Save replaced an active diagnostic log");
            logging.Append(Message("after refusal"));
            Check(File.ReadAllText(log).Contains("after refusal"), "Refusal broke the diagnostic writer");
        }
        finally
        {
            app.Destroy();
        }
    }
}
