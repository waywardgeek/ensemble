using System.Diagnostics;
using System.Text;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class JournalTests
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
        catch (Exception error) when (error is IOException or InvalidDataException or UnauthorizedAccessException or FormatException or ArgumentException)
        {
            return;
        }
        throw new Exception("Expected persistence refusal");
    }

    private static string PathForTest()
    {
        string directory = Path.Combine(Path.GetTempPath(), "journal-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(directory);
        return Path.Combine(directory, "save.json");
    }

    private static Agent Agent(Application app)
    {
        return new Agent(app, new Configuration(Vendor.OpenAI, "offline", "http://127.0.0.1:1"));
    }

    private static Event Message(string text)
    {
        return new Event(EventKind.MessageReceived, actor: Actor.Human, parts: [new TextPart(text)]);
    }

    private static string Text(History history)
    {
        return string.Join("|", history.View().Dialogue.SelectMany(entry => entry.Parts).OfType<TextPart>().Select(part => part.Text));
    }

    public static void Recovery()
    {
        string path = PathForTest();
        string journal = path + ".journal.jsonl";
        var app = new Application(TextWriter.Null);
        try
        {
            var source = Agent(app).History();
            source.StartJournal(path);
            source.Append(Message("one"));
            source.Save(new PersistenceOptions(path, 0));
            source.Append(Message("two"));
            source.Save(new PersistenceOptions(path, 0));
            source.Append(Message("three"));
            Check(source.Events().Length == 1, "Diagnostic trimming fixture failed");
            var facts = File.ReadAllLines(journal).Select(JsonCodec.ReadEvent).ToArray();
            Check(facts.Select(fact => fact.Seq).SequenceEqual([2L, 3L]), "Journal cut at primary instead of backup");
            Check(JsonCodec.ReadSave(File.ReadAllText(path + ".bak")).Sequence == 1, "Wrong backup boundary");
            var primary = Agent(app).History();
            Check(primary.LoadSave(path) && Text(primary) == "one|two|three", "Primary replay omitted or doubled journal");
            Check(primary.Snapshot().Sequence == 3, "Wrong recovered primary sequence");
            File.WriteAllText(path, "damaged primary");
            var backup = Agent(app).History();
            Check(backup.LoadSave(path) && Text(backup) == "one|two|three", "Backup cannot recover through journal");
            backup.StartJournal(path);
            Check(backup.Append(Message("four")).Seq == 4, "Recovered append reused sequence");
            backup.Save(new PersistenceOptions(path, 0));
            Check(JsonCodec.ReadSave(File.ReadAllText(path + ".bak")).Sequence == 1, "Damaged primary overwrote backup");
            var replay = Agent(app).History();
            Check(replay.LoadSave(path) && Text(replay) == "one|two|three|four", "Repaired primary duplicated journal");
            Refuses(() => backup.Save(new PersistenceOptions(path, 0, journal)));
            Refuses(() => backup.StartLog(journal));
        }
        finally
        {
            app.Destroy();
        }
    }

    public static void TornAndInvalid()
    {
        string path = PathForTest();
        string journal = path + ".journal.jsonl";
        var app = new Application(TextWriter.Null);
        try
        {
            var source = Agent(app).History();
            source.StartJournal(path);
            source.Append(Message("before first snapshot"));
            byte[] complete = File.ReadAllBytes(journal);
            // A process can die mid-UTF8 codepoint as well as mid-JSON record.
            using (var stream = new FileStream(journal, FileMode.Append))
            {
                stream.Write([0x7b, 0x22, 0xf0, 0x9f]);
            }
            byte[] torn = File.ReadAllBytes(journal);
            var restored = Agent(app).History();
            Check(restored.LoadSave(path) && Text(restored) == "before first snapshot", "Journal-only startup failed");
            Check(File.ReadAllBytes(journal).SequenceEqual(torn), "Read-only recovery repaired disk prematurely");
            restored.StartJournal(path);
            Check(File.ReadAllBytes(journal).SequenceEqual(complete), "Writer did not remove only the torn suffix");
            restored.Append(Message("after recovery"));
            var verify = Agent(app).History();
            Check(verify.LoadSave(path) && Text(verify) == "before first snapshot|after recovery", "Repaired journal cannot replay");
            string live = Text(verify);
            string good = File.ReadAllText(journal);
            foreach (string invalid in new[]
            {
                good + "not-json\n",
                good + JsonCodec.WriteEvent(Message("duplicate").Numbered(2, DateTimeOffset.UtcNow)) + "\n",
                JsonCodec.WriteEvent(Message("gap").Numbered(2, DateTimeOffset.UtcNow)) + "\n"
            })
            {
                File.WriteAllText(journal, invalid);
                Refuses(() => verify.LoadSave(path));
                Check(Text(verify) == live && File.ReadAllText(journal) == invalid, "Rejected recovery mutated live state or file");
            }
        }
        finally
        {
            app.Destroy();
        }
    }

    public static void FailedAppend()
    {
        string path = PathForTest();
        string journal = path + ".journal.jsonl";
        var app = new Application(TextWriter.Null);
        try
        {
            var history = Agent(app).History();
            history.StartJournal(path);
            history.Append(Message("committed"));
            File.Move(journal, journal + ".held");
            Directory.CreateDirectory(journal);
            Refuses(() => history.Append(Message("must not publish")));
            Check(history.Snapshot().Sequence == 1 && Text(history) == "committed", "Failed append published state");
            Directory.Delete(journal);
            File.Move(journal + ".held", journal);
            Refuses(() => history.Append(Message("unsafe retry")));
            Refuses(() => history.Save(new PersistenceOptions(path)));
            Check(!File.Exists(path), "Faulted writer published snapshot");
        }
        finally
        {
            try
            {
                app.Destroy();
                throw new Exception("Faulted History shutdown was incorrectly reported as successful");
            }
            catch (AggregateException error)
            {
                Check(error.Flatten().InnerExceptions.All(failure => failure is IOException),
                    "Cleanup fault was not the original recording failure");
                Check(app.Agents().Length == 0, "Faulted cleanup retained an owned Agent");
            }
        }
    }

    public static async Task WriterSerialization()
    {
        // A real FIFO supplies a latched file-read boundary; History itself is not
        // mocked. Once the writer peer opens, Save is blocked reading the primary.
        if (OperatingSystem.IsWindows())
        {
            throw new PlatformNotSupportedException("This file-I/O latch requires a POSIX FIFO");
        }
        string path = PathForTest();
        var app = new Application(TextWriter.Null);
        var opened = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        Task? peer = null;
        try
        {
            var history = Agent(app).History();
            history.StartJournal(path);
            history.Append(Message("one"));
            history.Save(new PersistenceOptions(path));
            byte[] previous = File.ReadAllBytes(path);
            history.Append(Message("two"));
            File.Delete(path);
            var start = new ProcessStartInfo("/usr/bin/mkfifo") { UseShellExecute = false };
            start.ArgumentList.Add(path);
            using var command = Process.Start(start)!;
            await command.WaitForExitAsync();
            Check(command.ExitCode == 0, "FIFO setup failed");
            peer = Task.Run(async () =>
            {
                using var output = new FileStream(path, FileMode.Open, FileAccess.Write);
                opened.SetResult();
                await release.Task;
                output.Write(previous);
            });
            Task save = Task.Run(() => history.Save(new PersistenceOptions(path, 0)));
            await opened.Task.WaitAsync(TimeSpan.FromSeconds(5));
            var snapshot = await Task.Run(history.Snapshot).WaitAsync(TimeSpan.FromSeconds(2));
            Check(snapshot.Sequence == 2 && Text(history) == "one|two", "Committed reader blocked or changed during save");
            Task<Event> append = Task.Run(() => history.Append(Message("three")));
            Check(await Task.WhenAny(append, Task.Delay(100)) != append, "Append raced the held snapshot writer");
            release.SetResult();
            await peer.WaitAsync(TimeSpan.FromSeconds(5));
            await save.WaitAsync(TimeSpan.FromSeconds(5));
            Check((await append.WaitAsync(TimeSpan.FromSeconds(5))).Seq == 3, "Serialized append has wrong identity");
            var restored = Agent(app).History();
            Check(restored.LoadSave(path) && Text(restored) == "one|two|three", "Save/append race lost or duplicated a fact");
        }
        finally
        {
            release.TrySetResult();
            if (peer is not null)
            {
                await peer.WaitAsync(TimeSpan.FromSeconds(5));
            }
            app.Destroy();
        }
    }
}
