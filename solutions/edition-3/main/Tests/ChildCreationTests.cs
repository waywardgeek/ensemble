using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class ChildCreationTests
{
    public static async Task StorageAndTemplates(string directory)
    {
        string home = Path.Combine(directory, "child-creation");
        string manuals = Path.Combine(home, "manuals");
        McpTests.Manual(manuals, "primary", "type: primary\ntools: read_file\nloadable-skills: extra\n", "Primary $ROLE");
        McpTests.Manual(manuals, "extra", "type: loadable\ntools: list_directory\n", "Extra $ROLE");
        var app = new Application(TextWriter.Null);
        try
        {
            var config = new Configuration(Vendor.Anthropic, "model-a", "http://127.0.0.1:1",
                safeMode: true, sandboxRoot: home);
            var parent = new Agent(app, config, dataDirectory: Path.Combine(home, "state"), skillSpec: new SkillSpec(manuals, "primary",
                new Dictionary<string, string> { ["ROLE"] = "original" }, ["extra"]));
            parent.StartPersistence(new PersistenceOptions(Path.Combine(home, "state", "save.json")));
            await parent.SkillManager()!.InitializeAsync();
            parent.SkillManager()!.RegisterVariable("ROLE", "inherited");
            parent.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("parent-only dialogue")]));
            // Templates are already parsed; children must not reread this directory.
            Directory.Delete(manuals, recursive: true);
            var child = await parent.CreateChildAsync("writer", model: "model-b");
            string childHome = Path.Combine(parent.DataDirectory(), "children", "writer");
            Check(child.DataDirectory() == childHome && child.Jobs().OutputDirectory == Path.Combine(childHome, "jobs"),
                "Child state/output paths do not follow parent ownership");
            Check(child.Configuration().Model == "model-b" && parent.Configuration().Model == "model-a"
                && child.Configuration().SafeMode && child.Sandbox().Root() == parent.Sandbox().Root(),
                "Configuration inheritance/override changed the parent or widened policy");
            Check(child.SkillManager() != parent.SkillManager() && child.SkillManager()!.Ready()
                && child.SkillManager()!.Active().Any(s => s.Name == "extra" && s.Body.Contains("inherited"))
                && child.ToolEnabled("list_directory") && !child.ToolEnabled("run_command"),
                "Parsed startup templates or current variable values were not independently inherited");
            Check(!Text(child).Contains("parent-only dialogue"), "Parent conversation copied to child");
            parent.SkillManager()!.RegisterVariable("ROLE", "later");
            child.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("child-owned dialogue")]));
            child.PublishSubmission("old completion");
            var grandchild = await child.CreateChildAsync("writer");
            Check(grandchild.SkillManager()!.Active().Any(s => s.Body.Contains("inherited")),
                "Child variables share parent mutation");
            Check(grandchild.DataDirectory() == Path.Combine(childHome, "children", "writer")
                && grandchild.AgentParent() == child, "Recursive storage or local child ID failed");
            child.Destroy();
            Check(parent.FindChild("writer") is null && File.Exists(Path.Combine(childHome, "save.json")),
                "Destruction failed to detach or retain conversation");
            var restored = await parent.CreateChildAsync("writer");
            Check(Text(restored).Contains("child-owned dialogue") && restored.Submission() is null,
                "Own History was not restored, or an old submission became new completion");
            Check(restored.SkillManager()!.Active().Any(s => s.Body.Contains("later")),
                "Saved authority replaced current bootstrap values");
        }
        finally
        {
            app.Destroy();
        }
    }

    public static async Task Refusals(string directory)
    {
        string home = Path.Combine(directory, "child-creation-refusals");
        Directory.CreateDirectory(home);
        var app = new Application(TextWriter.Null);
        try
        {
            var parent = new Agent(app, new Configuration(Vendor.Anthropic, "model-a", "http://127.0.0.1:1",
                sandboxRoot: home));
            await Refuses(() => parent.CreateChildAsync("no-storage"));
            parent.StartPersistence(new PersistenceOptions(Path.Combine(home, "save.json")));
            await Refuses(() => parent.CreateChildAsync("still-no-home"));
            Configuration configuration = parent.Configuration();
            parent.Destroy();
            parent = new Agent(app, configuration, dataDirectory: home);
            parent.StartPersistence(new PersistenceOptions(Path.Combine(home, "save.json")));
            using var badSchema = JsonDocument.Parse("{\"type\":\"object\",\"unsupported\":true}");
            await Refuses(() => parent.CreateChildAsync("bad-schema", submitSchema: badSchema.RootElement));
            await Refuses(() => parent.CreateChildAsync("../escape"));
            using var cancelled = new CancellationTokenSource();
            cancelled.Cancel();
            await Refuses(() => parent.CreateChildAsync("cancelled", cancellation: cancelled.Token));
            Check(parent.Children().IsEmpty && app.Agents().Length == 1
                && !Directory.Exists(Path.Combine(home, "children")), "Preflight refusal created child state");
            var child = await parent.CreateChildAsync("one");
            await Refuses(() => parent.CreateChildAsync("one"));
            Check(parent.FindChild("one") == child && parent.Children().Length == 1,
                "Duplicate creation replaced or detached existing child");
            string damaged = Path.Combine(home, "children", "damaged");
            Directory.CreateDirectory(damaged);
            string save = Path.Combine(damaged, "save.json");
            File.WriteAllText(save, "malformed preserved save");
            await Refuses(() => parent.CreateChildAsync("damaged"));
            Check(parent.FindChild("damaged") is null && app.Agents().Length == 2
                && File.ReadAllText(save) == "malformed preserved save", "Failed restore leaked child or rewrote save");
            // Removing only the fixture's bad input allows a later clean attempt.
            File.Delete(save);
            var retry = await parent.CreateChildAsync("damaged");
            Check(retry.AgentParent() == parent, "Failed creation reserved the ID permanently");
            await parent.ShutdownAsync();
            await Refuses(() => parent.CreateChildAsync("closed"));
        }
        finally
        {
            app.Destroy();
        }
    }

    private static string Text(Agent agent) => string.Join("\n", agent.History().Events()
        .SelectMany(e => e.Parts).OfType<TextPart>().Select(p => p.Text));
    private static async Task Refuses(Func<Task<Agent>> action)
    {
        try { await action(); }
        catch (Exception error) when (error is ArgumentException or InvalidOperationException
            or IOException or InvalidDataException or OperationCanceledException or FormatException) { return; }
        throw new Exception("Expected child creation refusal");
    }
    private static void Check(bool condition, string message)
    {
        if (!condition) { throw new Exception(message); }
    }
}
