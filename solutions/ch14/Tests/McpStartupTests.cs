using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

static partial class McpTests
{
    public static async Task PrimaryStartup()
    {
        string directory = SkillFixture();
        Manual(directory, "primary", "type: primary\ndepends: support\ntools: primary_echo\n"
            + Server(directory, "primary_echo", "startup"), "PRIMARY_STARTUP $LABEL\n$TOOLS");
        Manual(directory, "support", "type: dependency\ntools: support_echo\n"
            + Server(directory, "support_echo", "normal"), "SUPPORT_STARTUP");
        var root = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.OpenAI, "fake", "http://127.0.0.1:1"),
                skillSpec: new SkillSpec(directory, "primary", new Dictionary<string, string> { ["LABEL"] = "before" }));
            var skills = agent.SkillManager()!;
            Check(!skills.RequiresBrowserStartup(), "Stdio primary incorrectly requires a browser");
            Check(!skills.Ready() && skills.Active().IsEmpty && !agent.ToolEnabled("write_file"),
                "Pending primary exposed provisional grants");
            skills.RegisterVariable("LABEL", "after");
            var startup = skills.InitializeAsync();
            var other = skills.InitializeAsync();
            await WaitFile(Path.Combine(directory, "primary_echo.ready"));
            Check(!File.Exists(Path.Combine(directory, "reverse.txt")), "Startup reverse call gained provisional authority");
            Check(skills.Clients().Length == 2 && agent.Tools().All(tool => tool.Name() != "support_echo"),
                "Prepared dependency was published before primary readiness");
            using var cancelled = new CancellationTokenSource();
            var observer = skills.InitializeAsync(cancelled.Token);
            cancelled.Cancel();
            try
            {
                await observer;
                throw new Exception("Cancelled startup observer returned success");
            }
            catch (OperationCanceledException)
            {
            }
            try
            {
                await agent.SendAsync("Must not reach HTTP").WaitAsync(TimeSpan.FromSeconds(5));
                throw new Exception("Pending primary admitted model work");
            }
            catch (InvalidOperationException)
            {
            }
            await agent.PostAsync(new AgentInput(InputKind.Pause)).WaitAsync(TimeSpan.FromSeconds(5));
            File.WriteAllText(Path.Combine(directory, "primary_echo.release"), "release");
            await Task.WhenAll(startup, other).WaitAsync(TimeSpan.FromSeconds(10));
            await skills.InitializeAsync();
            var facts = agent.History().Events().Where(fact => fact.Kind == EventKind.SkillLoaded).ToArray();
            Check(facts.Length == 1, "Repeated initialization published more than one activation");
            var manuals = facts[0].Parts.OfType<SkillPart>().ToArray();
            Check(manuals.Select(part => part.Name).SequenceEqual(new[] { "support", "primary" })
                && !manuals[0].Primary && manuals[1].Primary, "Initial dependency closure lost its primary identity");
            Check(manuals[1].Body.Contains("before") && !manuals[1].Body.Contains("after"),
                "Primary variables were not frozen at construction");
            Check(skills.Ready() && agent.ToolEnabled("load_skill") && agent.ToolEnabled("unload_skill"),
                "Initial activation omitted its universal controls");
            Check(agent.History().Events().Where(fact => fact.Kind == EventKind.ToolObserved)
                .All(fact => fact.CallID != ""), "Startup fabricated a host tool result");
            string wire = agent.Engine().Render(agent.History().View()).ToJsonString();
            using var request = JsonDocument.Parse(wire);
            var instruction = request.RootElement.GetProperty("messages")[0];
            Check(instruction.GetProperty("role").GetString() == "system"
                && instruction.GetProperty("content").GetString()!.Contains("PRIMARY_STARTUP")
                && wire.Split("PRIMARY_STARTUP").Length == 2, "Primary startup lost instruction placement or duplicated its body");
            await agent.PostAsync(new AgentInput(InputKind.Resume));
            var call = new ToolCallPart("startup-echo", agent.Configuration().Target(), "primary_echo",
                JsonSerializer.SerializeToElement(new { }));
            var result = await agent.DispatchAsync(call).WaitAsync(TimeSpan.FromSeconds(5));
            Check(!result.IsError && result.Parts.OfType<TextPart>().Any(part => part.Text.Contains("remote:primary_echo")), "Started primary tool was not callable");
            var pids = new[] { "support_echo", "primary_echo" }
                .Select(name => int.Parse(File.ReadAllText(Path.Combine(directory, name + ".pid")))).ToArray();
            var manager = agent.Mcp()!;
            await agent.ShutdownAsync().WaitAsync(TimeSpan.FromSeconds(10));
            Check(skills.Clients().IsEmpty && manager.Clients().IsEmpty && agent.Mcp() is null && pids.All(ProcessStopped),
                "Primary shutdown retained a process or relationship");
        }
        finally
        {
            root.Destroy();
            Directory.Delete(directory, true);
        }
    }

    public static async Task PrimaryFailure()
    {
        string directory = SkillFixture();
        Manual(directory, "primary", "type: primary\ndepends: support\ntools: primary_echo\n"
            + Server(directory, "primary_echo", "closed"), "Must not publish");
        Manual(directory, "support", "type: dependency\ntools: support_echo\n"
            + Server(directory, "support_echo", "normal"), "Must roll back");
        var root = new Application(TextWriter.Null);
        try
        {
            string save = Path.Combine(directory, "session.json");
            var old = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"));
            old.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human, parts: [new TextPart("Retain me")]));
            old.History().Save(new PersistenceOptions(save));
            old.Destroy();
            byte[] before = File.ReadAllBytes(save);
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"),
                skillSpec: new SkillSpec(directory, "primary"));
            Check(agent.StartPersistence(new PersistenceOptions(save)), "Startup did not restore prior conversation");
            var skills = agent.SkillManager()!;
            try
            {
                await skills.InitializeAsync().WaitAsync(TimeSpan.FromSeconds(10));
                throw new Exception("Closed primary initialized successfully");
            }
            catch (InvalidDataException)
            {
            }
            Check(!skills.Ready() && skills.Active().IsEmpty && skills.Clients().IsEmpty
                && agent.Tools().All(tool => tool.Name() != "support_echo"), "Failed primary retained partial activation");
            int pid = int.Parse(File.ReadAllText(Path.Combine(directory, "support_echo.pid")));
            Check(ProcessStopped(pid), "Failed primary retained its prepared dependency process");
            await agent.ShutdownAsync().WaitAsync(TimeSpan.FromSeconds(10));
            Check(before.SequenceEqual(File.ReadAllBytes(save)), "Failed startup overwrote prior saved conversation");
        }
        finally
        {
            root.Destroy();
            Directory.Delete(directory, true);
        }
    }

    public static async Task PrimaryShutdown()
    {
        string directory = SkillFixture();
        Manual(directory, "primary", "type: primary\ntools: primary_echo\n"
            + Server(directory, "primary_echo", "startup"), "Held primary");
        var root = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"),
                skillSpec: new SkillSpec(directory, "primary"));
            var startup = agent.SkillManager()!.InitializeAsync();
            await WaitFile(Path.Combine(directory, "primary_echo.ready"));
            int pid = int.Parse(File.ReadAllText(Path.Combine(directory, "primary_echo.pid")));
            await agent.ShutdownAsync().WaitAsync(TimeSpan.FromSeconds(10));
            try
            {
                await startup.WaitAsync(TimeSpan.FromSeconds(5));
                throw new Exception("Shutdown published a successful primary");
            }
            catch (Exception error) when (error is OperationCanceledException or InvalidDataException or ObjectDisposedException)
            {
            }
            Check(!agent.SkillManager()!.Ready() && agent.SkillManager()!.Clients().IsEmpty && ProcessStopped(pid),
                "Shutdown left primary startup work alive");
        }
        finally
        {
            root.Destroy();
            Directory.Delete(directory, true);
        }
    }
}
