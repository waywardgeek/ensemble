using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

// These tests use the real skill parser, reducer, renderers, actor and process peer.
// Only the external MCP server is replaced by a working local implementation.
static partial class McpTests
{
    public static async Task StartupSelection()
    {
        string directory = SkillTests.Fixture();
        var root = new Application(TextWriter.Null);
        try
        {
            foreach (var vendor in Enum.GetValues<Vendor>())
            {
                var selected = new List<string> { "editing", "extra", "editing" };
                var agent = new Agent(root, new Configuration(vendor, "model-a", "http://127.0.0.1:1"),
                    skillSpec: new SkillSpec(directory, "root", new Dictionary<string, string> { ["WHO"] = "startup" }, selected));
                selected.Clear();
                var skills = agent.SkillManager()!;
                await skills.InitializeAsync();
                Check(skills.Ready() && !skills.RequiresBrowserStartup(), "Local startup was not ready");
                Check(skills.Active().Select(skill => skill.Name).SequenceEqual(new[] { "root", "helper", "editing", "extra" }),
                    "Startup order, dependencies or duplicate elimination failed");
                Check(agent.History().Events().Count(fact => fact.Kind == EventKind.SkillLoaded) == 1,
                    "Startup did not publish one activation");
                var body = agent.Engine().Render(agent.History().View());
                string instructions = vendor switch
                {
                    Vendor.Anthropic => body["system"]!.GetValue<string>(),
                    Vendor.OpenAI => body["messages"]![0]!["content"]!.GetValue<string>(),
                    _ => body["systemInstruction"]!["parts"]![0]!["text"]!.GetValue<string>()
                };
                Check(instructions.Contains("PRIMARY-startup") && instructions.Contains("EDIT-startup")
                    && instructions.Contains("EXTRA-startup") && body.ToJsonString().Split("EDIT-startup").Length == 2,
                    "Startup instructions absent or duplicated: " + vendor);
                Check(agent.ToolEnabled("write_file") && agent.ToolEnabled("search_files"), "Startup grants missing");
                Check(!agent.History().Events().Any(fact => fact.Kind is EventKind.ToolCalled or EventKind.ToolObserved),
                    "Startup invented a tool exchange");
                agent.Destroy();
            }
            foreach (var selection in new[] { new[] { "extra", "editing" }, new[] { "editing", "missing" },
                new[] { "broken" }, new[] { "cycle" }, new[] { "helper" }, new[] { "root" }, new[] { "" } })
            {
                var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"));
                agent.InstallTools();
                try
                {
                    _ = new Skills(agent, new SkillSpec(directory, "root", StartupSkills: selection));
                    throw new Exception("Invalid startup accepted: " + string.Join(',', selection));
                }
                catch (InvalidDataException)
                {
                }
                Check(agent.SkillManager() is null && !agent.History().Events().Any(fact => fact.Kind == EventKind.SkillLoaded),
                    "Rejected selection left attachment or partial instructions");
                agent.Destroy();
            }
        }
        finally
        {
            root.Destroy();
            Directory.Delete(directory, true);
        }
    }

    public static async Task OptionalServerStartup()
    {
        string directory = SkillFixture();
        Manual(directory, "primary", "type: primary\ntools: read_file\nloadable-skills: debug\n", "PRIMARY_OPTIONAL $LABEL");
        Manual(directory, "debug", "type: loadable\ntools: debug_echo\n" + Server(directory, "debug_echo", "startup"),
            "DEBUG_OPTIONAL $LABEL\n$TOOLS");
        var root = new Application(TextWriter.Null);
        try
        {
            var selected = new List<string> { "debug" };
            var agent = new Agent(root, new Configuration(Vendor.OpenAI, "fake", "http://127.0.0.1:1"),
                skillSpec: new SkillSpec(directory, "primary", new Dictionary<string, string> { ["LABEL"] = "before" }, selected));
            selected[0] = "missing";
            var skills = agent.SkillManager()!;
            skills.RegisterVariable("LABEL", "after");
            var startup = skills.InitializeAsync();
            await WaitFile(Path.Combine(directory, "debug_echo.ready"));
            Check(!skills.Ready() && skills.Active().IsEmpty && !agent.ToolEnabled("read_file"),
                "Optional server published a partial primary");
            try
            {
                await agent.SendAsync("no early HTTP").WaitAsync(TimeSpan.FromSeconds(5));
                throw new Exception("Pending optional startup admitted a turn");
            }
            catch (InvalidOperationException)
            {
            }
            File.WriteAllText(Path.Combine(directory, "debug_echo.release"), "release");
            await startup.WaitAsync(TimeSpan.FromSeconds(10));
            Check(skills.Ready() && skills.Active().All(skill => skill.Body.Contains("before") && !skill.Body.Contains("after")),
                "Optional startup lost its construction snapshot");
            Check(agent.History().Events().Count(fact => fact.Kind == EventKind.SkillLoaded) == 1 && agent.ToolEnabled("debug_echo"),
                "Optional server did not commit once");
            var result = await agent.DispatchAsync(new ToolCallPart("optional-call", agent.Configuration().Target(), "debug_echo",
                JsonSerializer.SerializeToElement(new { })));
            Check(!result.IsError, "Startup tool unavailable");
            int pid = int.Parse(File.ReadAllText(Path.Combine(directory, "debug_echo.pid")));
            await agent.ShutdownAsync();
            Check(skills.Clients().IsEmpty && ProcessStopped(pid), "Optional server survived shutdown");
        }
        finally
        {
            root.Destroy();
            Directory.Delete(directory, true);
        }
    }
    public static async Task OptionalStartupFailure()
    {
        string directory = SkillFixture();
        Manual(directory, "primary", "type: primary\ntools: read_file\nloadable-skills: debug\n", "NEW_PRIMARY");
        Manual(directory, "debug", "type: loadable\ntools: failed_echo\n" + Server(directory, "failed_echo", "closed"), "NEW_DEBUG");
        var root = new Application(TextWriter.Null);
        try
        {
            string save = Path.Combine(directory, "save.json");
            var old = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"));
            old.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human, parts: [new TextPart("prior dialogue")]));
            old.History().Save(new PersistenceOptions(save));
            old.Destroy();
            byte[] before = File.ReadAllBytes(save);
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"),
                skillSpec: new SkillSpec(directory, "primary", StartupSkills: ["debug"]));
            agent.StartPersistence(new PersistenceOptions(save));
            try
            {
                await agent.SkillManager()!.InitializeAsync().WaitAsync(TimeSpan.FromSeconds(10));
                throw new Exception("Failed optional server accepted");
            }
            catch (InvalidDataException)
            {
            }
            Check(agent.SkillManager()!.Active().IsEmpty && agent.SkillManager()!.Clients().IsEmpty
                && !agent.ToolEnabled("read_file"), "Failed optional startup leaked state");
            await agent.ShutdownAsync();
            Check(before.SequenceEqual(File.ReadAllBytes(save)), "Failed optional startup overwrote prior save");
            Check(ProcessStopped(int.Parse(File.ReadAllText(Path.Combine(directory, "failed_echo.pid")))),
                "Failed optional peer survived cleanup");
        }
        finally
        {
            root.Destroy();
            Directory.Delete(directory, true);
        }
    }

}
