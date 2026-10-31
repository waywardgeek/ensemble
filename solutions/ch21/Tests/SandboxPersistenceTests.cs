using Ensemble;
using Application = Ensemble.Ensemble;

internal static class SandboxPersistenceTests
{
    public static void Replay()
    {
        InWorkspace((app, inside, outside) =>
        {
            var source = Agent(app, inside).History();
            // This first assertion is the regression baseline: an owned History
            // must not truncate an outside diagnostic destination.
            Refuses(() => source.StartLog(Path.Combine(outside, "diagnostic.jsonl")));
            Check(!File.Exists(Path.Combine(outside, "diagnostic.jsonl")), "Outside log was created");
            Directory.CreateDirectory(Path.Combine(inside, "state"));
            Directory.CreateSymbolicLink(Path.Combine(inside, "alias"), Path.Combine(inside, "state"));
            source.StartJournal("alias/save.json");
            source.StartLog("alias/diagnostic.jsonl");
            source.Append(Message("one"));
            source.Save(new PersistenceOptions("alias/save.json", 0, "alias/archive.jsonl"));
            source.Append(Message("two"));
            source.Save(new PersistenceOptions("alias/save.json", 0, "alias/archive.jsonl"));
            source.Append(Message("three"));
            string path = Path.Combine(inside, "state", "save.json");
            Check(JsonCodec.ReadSave(File.ReadAllText(path + ".bak")).Sequence == 1, "Wrong rooted backup");
            var recovered = Agent(app, inside).History();
            Check(recovered.LoadSave("alias/save.json") && Text(recovered) == "one|two|three",
                "Rooted primary/journal recovery failed");
            File.WriteAllText(path, "damaged primary");
            var backup = Agent(app, inside).History();
            Check(backup.LoadSave("state/save.json") && Text(backup) == "one|two|three",
                "Rooted backup/journal recovery failed");
            var archive = Agent(app, inside).History();
            archive.Load("alias/archive.jsonl");
            Check(Text(archive) == "one|two", "Rooted archive replay failed");
            var log = Agent(app, inside).History();
            log.Load("state/diagnostic.jsonl");
            Check(Text(log) == "one|two|three", "Rooted diagnostic replay failed");
            Refuses(() => source.Save(new PersistenceOptions("state/save.json", 0, "alias/save.json.bak")));
            Check(!Directory.EnumerateFiles(Path.Combine(inside, "state"), "*.tmp").Any(),
                "Successful replacement retained temporary files");

            // Exercise startup normalization through Agent, not just History.
            var lifecycle = Agent(app, inside);
            Check(!lifecycle.StartPersistence(new PersistenceOptions("new/session.json")), "New state unexpectedly existed");
            lifecycle.History().Append(Message("lifecycle"));
            lifecycle.Destroy();
            var next = Agent(app, inside).History();
            Check(next.LoadSave("new/session.json") && Text(next) == "lifecycle", "Rooted lifecycle save failed");
        });
    }

    public static void Refusals()
    {
        InWorkspace((app, inside, outside) =>
        {
            string sentinel = Path.Combine(outside, "sentinel");
            File.WriteAllText(sentinel, "outside-original");
            Directory.CreateSymbolicLink(Path.Combine(inside, "escape"), outside);
            var owner = Agent(app, inside);
            var history = owner.History();
            history.Append(Message("committed"));
            foreach (string path in new[] { sentinel, "../outside/sentinel", "escape/sentinel" })
            {
                Refuses(() => history.Save(new PersistenceOptions(path)));
                Refuses(() => history.LoadSave(path));
                Refuses(() => history.StartLog(path));
                Refuses(() => history.Load(path));
                Refuses(() => history.StartJournal(path));
                Refuses(() => history.Save(new PersistenceOptions("safe.json", 0, path)));
                Check(Text(history) == "committed", "Refusal changed committed History");
            }
            Check(!File.Exists(Path.Combine(inside, "safe.json")), "Rejected archive changed recovery file");

            // Every sibling is independently opened through the boundary. A
            // previously validated primary must not authorize a swapped backup.
            history.Save(new PersistenceOptions("safe.json"));
            string primary = Path.Combine(inside, "safe.json");
            string original = File.ReadAllText(primary);
            File.CreateSymbolicLink(primary + ".bak", sentinel);
            Refuses(() => history.Save(new PersistenceOptions("safe.json")));
            Check(File.ReadAllText(primary) == original, "Failed backup replaced primary");
            File.Delete(primary + ".bak");
            File.CreateSymbolicLink(primary + ".journal.jsonl", sentinel);
            Refuses(() => Agent(app, inside).History().LoadSave("safe.json"));
            File.Delete(primary + ".journal.jsonl");

            history.StartJournal("journal/state.json");
            long sequence = history.Snapshot().Sequence;
            string directory = Path.Combine(inside, "journal");
            Directory.Move(directory, directory + ".saved");
            Directory.CreateSymbolicLink(directory, outside);
            Refuses(() => history.Append(Message("must not publish")));
            Check(history.Snapshot().Sequence == sequence && Text(history) == "committed",
                "Failed rooted append published a fact");
            Directory.Delete(directory);
            Directory.Move(directory + ".saved", directory);
            Refuses(() => history.Append(Message("poisoned")));
            Refuses(() => history.Save(new PersistenceOptions("journal/state.json")));
            Check(File.ReadAllText(sentinel) == "outside-original"
                && Directory.GetFileSystemEntries(outside).Length == 1, "Persistence changed outside files");
            // Poisoning also forbids the shutdown fact. Destruction must report
            // that error while still releasing every owner relationship.
            Refuses(owner.Destroy);
            Check(!app.Agents().Contains(owner), "Poisoned History prevented Agent detach");
        });
    }

    private static Agent Agent(Application app, string root) => new(app,
        new Configuration(Vendor.OpenAI, "offline", "http://127.0.0.1:1", sandboxRoot: root));

    private static Event Message(string text) => new(EventKind.MessageReceived,
        actor: Actor.Human, parts: [new TextPart(text)]);

    private static string Text(History history) => string.Join("|", history.View().Dialogue
        .SelectMany(entry => entry.Parts).OfType<TextPart>().Select(part => part.Text));

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
        catch (IOException)
        {
            return;
        }
        catch (ArgumentException)
        {
            return;
        }
        throw new Exception("Expected rooted persistence refusal");
    }

    private static void InWorkspace(Action<Application, string, string> test)
    {
        string workspace = Path.Combine(Path.GetTempPath(), "rooted-state-" + Guid.NewGuid().ToString("N"));
        string inside = Path.Combine(workspace, "inside");
        string outside = Path.Combine(workspace, "outside");
        Directory.CreateDirectory(inside);
        Directory.CreateDirectory(outside);
        var app = new Application(TextWriter.Null);
        try
        {
            test(app, inside, outside);
        }
        finally
        {
            app.Destroy();
            Directory.Delete(workspace, recursive: true);
        }
    }
}
