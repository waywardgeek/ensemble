using Ensemble;
using Application = Ensemble.Ensemble;

internal static class StorageTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition) throw new Exception(message);
    }

    private static void Refuses(Action action)
    {
        try
        {
            action();
        }
        catch (Exception error) when (error is ArgumentException or InvalidOperationException
            or UnauthorizedAccessException or IOException)
        {
            return;
        }
        throw new Exception("Expected storage refusal.");
    }

    public static async Task NamedHomes(string directory)
    {
        string work = Path.Combine(directory, "storage-named");
        Directory.CreateDirectory(work);
        string container = Path.Combine(work, "ensemble");
        var app = new Application(TextWriter.Null, dataDirectory: container);
        var config = new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1",
            safeMode: true, sandboxRoot: work);
        try
        {
            var first = new Agent(app, config, name: "cadence");
            var peer = new Agent(app, config, name: "Flóra");
            string firstHome = Path.Combine(container, "agents", "cadence");
            string peerHome = Path.Combine(container, "agents", "Flóra");
            Check(app.DataDirectory() == container && first.DataDirectory() == firstHome
                && peer.DataDirectory() == peerHome, "Top-level homes are not siblings beneath Ensemble.");
            Check(first.AgentParent() is null && peer.AgentParent() is null
                && first.Parent() == app && peer.Parent() == app,
                "A top-level peer acquired an Agent parent.");
            Check(first.Jobs().OutputDirectory == Path.Combine(firstHome, "jobs"),
                "Default Job storage did not follow the actor home.");

            // The save path is deliberately deeper than the data home. Neither
            // submission nor descendant creation may use its containing directory.
            string save = Path.Combine(firstHome, "session", "save.json");
            first.StartPersistence(new PersistenceOptions(save));
            first.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("named actor dialogue") ]));
            var child = await first.CreateChildAsync("worker");
            string childHome = Path.Combine(firstHome, "children", "worker");
            Check(child.DataDirectory() == childHome && child.AgentParent() == first,
                "Child storage followed the conversation path rather than its owning actor.");
            child.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("child dialogue") ]));
            child.PublishSubmission("durable child result");
            var nested = await child.CreateChildAsync("worker");
            Check(nested.DataDirectory() == Path.Combine(childHome, "children", "worker"),
                "Recursive child storage lost its actual owner.");
            first.Destroy();
            Check(app.FindAgent("cadence") is null && app.FindAgent("Flóra") == peer
                && child.AgentParent() is null && nested.AgentParent() is null,
                "Destroying one actor damaged its peer or retained child links.");
            Check(File.Exists(save) && File.Exists(Path.Combine(childHome, "submit.md")),
                "Destruction erased durable actor work.");

            var restored = new Agent(app, config, name: "cadence");
            Check(restored.StartPersistence(new PersistenceOptions(save)), "Actor History did not restore.");
            var restoredChild = await restored.CreateChildAsync("worker");
            Check(restored.DataDirectory() == firstHome && restoredChild.DataDirectory() == childHome
                && restoredChild.History().Events().SelectMany(e => e.Parts).OfType<TextPart>()
                    .Any(p => p.Text == "child dialogue") && restoredChild.Submission() is null,
                "Restoration moved storage, lost child History, or revived old completion.");
        }
        finally
        {
            app.Destroy();
        }
    }

    public static void Admission(string directory)
    {
        string work = Path.Combine(directory, "storage-admission");
        Directory.CreateDirectory(work);
        var app = new Application(TextWriter.Null, dataDirectory: Path.Combine(work, "ensemble"));
        var config = new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1",
            safeMode: true, sandboxRoot: work);
        try
        {
            foreach (string name in new[] { "", " ", ".", "..", "../escape", "a/b", "a\\b" })
            {
                Refuses(() => new Agent(app, config, name: name));
                Check(app.Agents().IsEmpty, "Invalid named storage published an Agent.");
            }
            Refuses(() => new Agent(app, config, name: "elsewhere",
                dataDirectory: Path.Combine(work, "outside-container")));
            Check(app.Agents().IsEmpty && !Directory.Exists(Path.Combine(work, "outside-container")),
                "An explicit home bypassed Ensemble's named-actor assignment.");
            var named = new Agent(app, config, name: "one");
            Refuses(() => new Agent(app, config, name: "one"));
            Check(app.Agents().Length == 1 && app.FindAgent("one") == named,
                "Duplicate named storage replaced the existing owner.");
            Refuses(() => new Agent(app, config, dataDirectory: Path.Combine(directory, "outside-home")));
            Check(app.Agents().Length == 1, "Sandbox refusal left a partial Agent registered.");

            var transient = new Agent(app, config, outputDirectory: Path.Combine(work, "transient-jobs"));
            transient.StartPersistence(new PersistenceOptions(Path.Combine(work, "session", "save.json")));
            Refuses(() => transient.DataDirectory());
            Check(transient.Name is null, "An unnamed library Agent was assigned a fabricated identity.");
            var explicitHome = new Agent(app, config, dataDirectory: Path.Combine(work, "explicit"));
            explicitHome.StartPersistence(new PersistenceOptions(Path.Combine(work, "another-session", "save.json")));
            Check(explicitHome.DataDirectory() == Path.Combine(work, "explicit"),
                "Persistence replaced the explicit actor home.");
        }
        finally
        {
            app.Destroy();
        }
    }
}
