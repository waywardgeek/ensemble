using System.Diagnostics;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

static partial class McpTests
{
    internal static string SkillFixture()
    {
        string directory = Path.Combine(Path.GetTempPath(), "mcp-skills-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(directory);
        File.WriteAllText(Path.Combine(directory, "peer.py"), """
            import json, os, sys, time
            mode, folder, name = sys.argv[1:]
            def send(value):
                print(json.dumps(value), flush=True)
            def receive():
                return json.loads(sys.stdin.readline())
            with open(os.path.join(folder, name + '.pid'), 'w') as f:
                f.write(str(os.getpid()))
            init = receive()
            if mode == 'closed':
                sys.exit(0)
            if mode in ('hold', 'startup'):
                # The client reader must remain alive and the Agent actor must
                # dispatch this call while load_skill is awaiting initialize.
                send({'jsonrpc':'2.0', 'id':'reverse-init', 'method':'tools/call',
                      'params':{'name':'write_file', 'arguments':{
                          'path':os.path.join(folder, 'reverse.txt'), 'content':'reverse worked'}}})
                answer = receive()
                assert answer['id'] == 'reverse-init'
                assert answer['result']['isError'] == (mode == 'startup')
                open(os.path.join(folder, name + '.ready'), 'w').close()
                while not os.path.exists(os.path.join(folder, name + '.release')):
                    time.sleep(0.01)
            send({'jsonrpc':'2.0', 'id':init['id'], 'result':{'protocolVersion':'2024-11-05',
                  'capabilities':{}, 'serverInfo':{'name':'skill-peer', 'version':'1'}}})
            assert receive()['method'] == 'notifications/initialized'
            listing = receive()
            tool = 'read_file' if mode == 'collision' else name
            schema = {'type':'object', 'properties':{}, 'additionalProperties':False}
            tools = [{'name':tool, 'description':'selected', 'inputSchema':schema},
                     {'name':'unrequested', 'description':'must stay hidden', 'inputSchema':schema}]
            send({'jsonrpc':'2.0', 'id':listing['id'], 'result':{'tools':tools}})
            for line in sys.stdin:
                request = json.loads(line)
                if request.get('method') == 'tools/call':
                    send({'jsonrpc':'2.0', 'id':request['id'], 'result':{
                        'content':[{'type':'text', 'text':'remote:' + name}], 'isError':False}})
            """);
        return directory;
    }

    internal static void Manual(string directory, string name, string fields, string body)
    {
        string folder = Path.Combine(directory, name);
        Directory.CreateDirectory(folder);
        File.WriteAllText(Path.Combine(folder, "SKILL.md"),
            "---\nname: " + name + "\n" + fields + "---\n" + body);
    }

    internal static string Server(string directory, string name, string mode)
    {
        string args = JsonSerializer.Serialize(new[] { Path.Combine(directory, "peer.py"), mode, directory, name });
        return "mcp_servers:\n  - name: " + name + "\n    transport: stdio\n"
            + "    command: /usr/bin/python3\n    args: " + args + "\n";
    }

    private static ToolCallPart SkillCall(Agent agent, string tool, string name)
    {
        return new ToolCallPart(Guid.NewGuid().ToString("N"), agent.Configuration().Target(), tool,
            JsonSerializer.SerializeToElement(new { name }));
    }

    internal static async Task WaitFile(string path)
    {
        using var bound = new CancellationTokenSource(TimeSpan.FromSeconds(10));
        while (!File.Exists(path))
        {
            await Task.Delay(10, bound.Token);
        }
    }

    internal static bool ProcessStopped(int pid)
    {
        try
        {
            using var process = Process.GetProcessById(pid);
            return process.HasExited;
        }
        catch (ArgumentException)
        {
            return true;
        }
    }

    public static async Task SkillActivation()
    {
        string directory = SkillFixture();
        Manual(directory, "primary", "type: primary\ntools: write_file read_file\nloadable-skills: remote\n", "Frozen primary $LABEL");
        Manual(directory, "remote", "type: loadable\ntools: remote_echo\n" + Server(directory, "remote_echo", "hold"),
            "Remote $LABEL\n$TOOLS");
        var root = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"),
                skillSpec: new SkillSpec(directory, "primary", new Dictionary<string, string> { ["LABEL"] = "before" }));
            var skills = agent.SkillManager()!;
            var primary = skills.Active().Single();
            var load = agent.DispatchAsync(SkillCall(agent, "load_skill", "remote"));
            await WaitFile(Path.Combine(directory, "remote_echo.ready"));
            Check(File.ReadAllText(Path.Combine(directory, "reverse.txt")) == "reverse worked",
                "Initialization reverse call did not complete through the actor");
            Check(!agent.ToolEnabled("remote_echo") && skills.Active().Length == 1,
                "Prepared connection published grants early");
            skills.RegisterVariable("LABEL", "after");
            await agent.PostAsync(new AgentInput(InputKind.Pause)).WaitAsync(TimeSpan.FromSeconds(5));
            File.WriteAllText(Path.Combine(directory, "remote_echo.release"), "");
            var loaded = await load.WaitAsync(TimeSpan.FromSeconds(10));
            Check(!loaded.IsError, "Remote activation returned an error");
            Check(skills.Active().Single(skill => skill.Name == "remote").Body.Contains("Remote before"),
                "Manual did not use the activation-time variable snapshot");
            Check(skills.Active().Single(skill => skill.Primary).Body == primary.Body, "Primary manual changed");
            Check(agent.ToolEnabled("remote_echo") && !agent.Tools().Any(tool => tool.Name() == "unrequested"),
                "Discovered tools bypassed explicit skill grants");
            var client = skills.Clients().Single();
            Check(client.SkillOwner() == skills && client.Parent() == agent.Mcp()
                && client.Tools().Single().Remote() == client, "Connection ownership is incomplete");
            skills.RemoveClient(client);
            Check(client.SkillOwner() is null && skills.Clients().IsEmpty, "Remove did not update both ends");
            skills.AddClient(client);
            await agent.PostAsync(new AgentInput(InputKind.Resume));
            var call = new ToolCallPart("remote-use", agent.Configuration().Target(), "remote_echo", J("{}"));
            var result = await agent.DispatchAsync(call).WaitAsync(TimeSpan.FromSeconds(10));
            Check(!result.IsError && result.Parts.OfType<TextPart>().Any(part => part.Text == "remote:remote_echo"),
                "Activated tool did not execute through its real Job");
            int pid = int.Parse(File.ReadAllText(Path.Combine(directory, "remote_echo.pid")));
            var unloaded = await agent.DispatchAsync(SkillCall(agent, "unload_skill", "remote"));
            Check(!unloaded.IsError && skills.Pending("remote"), "Unload did not retain the pending manual");
            Check(skills.Clients().IsEmpty && agent.Mcp()!.Clients().IsEmpty && ProcessStopped(pid),
                "Unload did not join the process and detach both parents");
            Check(!agent.Tools().Any(tool => tool.Name() == "remote_echo"), "Unload left a dead remote Tool attached");
        }
        finally
        {
            root.Destroy();
            Directory.Delete(directory, true);
        }
    }

    public static async Task SkillDependencies()
    {
        string directory = SkillFixture();
        Manual(directory, "primary", "type: primary\nloadable-skills: remote dependency\n", "Primary");
        string dependencyServer = Server(directory, "dep_tool", "ok").Replace("name: dep_tool", "name: local");
        string remoteServer = Server(directory, "remote_tool", "ok").Replace("name: remote_tool", "name: local");
        Manual(directory, "dependency", "type: loadable\ntools: dep_tool\n" + dependencyServer, "Dependency");
        Manual(directory, "remote", "type: loadable\ndepends: dependency\ntools: remote_tool\n" + remoteServer, "Remote");
        var root = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"),
                skillSpec: new SkillSpec(directory, "primary"));
            var result = await agent.DispatchAsync(SkillCall(agent, "load_skill", "remote"))
                .WaitAsync(TimeSpan.FromSeconds(10));
            Check(!result.IsError, "Dependency closure failed to load");
            var skills = agent.SkillManager()!;
            Check(skills.Clients().Length == 2 && skills.Clients().All(client => client.Name == "local"),
                "Skill-local connection names were treated as global identity");
            var activation = agent.History().Events().Last(fact => fact.Kind == EventKind.SkillLoaded);
            Check(activation.Parts.OfType<SkillPart>().Select(part => part.Name)
                .SequenceEqual(new[] { "dependency", "remote" }), "Closure was not one dependency-first publication");
            await agent.DispatchAsync(SkillCall(agent, "unload_skill", "remote"));
            Check(skills.Clients().Single().SkillName == "dependency", "Unload destroyed another skill's server");
            var call = new ToolCallPart("dependency-use", agent.Configuration().Target(), "dep_tool", J("{}"));
            var reply = await agent.DispatchAsync(call).WaitAsync(TimeSpan.FromSeconds(10));
            Check(!reply.IsError, "Surviving dependency lost its remote capability");
            await agent.DispatchAsync(SkillCall(agent, "unload_skill", "dependency"));
            Check(skills.Clients().IsEmpty && agent.Mcp()!.Clients().IsEmpty, "Final unload retained ownership");
        }
        finally
        {
            root.Destroy();
            Directory.Delete(directory, true);
        }
    }

    public static async Task SkillRollback()
    {
        foreach (string mode in new[] { "missing", "collision", "closed" })
        {
            string directory = SkillFixture();
            Manual(directory, "primary", "type: primary\ntools: read_file\nloadable-skills: remote\n", "Primary");
            Manual(directory, "dependency", "type: dependency\ntools: dep_tool\n" + Server(directory, "dep_tool", "ok"), "Dependency");
            string requested = mode == "collision" ? "read_file" : "absent_tool";
            Manual(directory, "remote", "type: loadable\ndepends: dependency\ntools: " + requested + "\n"
                + Server(directory, "remote_tool", mode), "Remote");
            var root = new Application(TextWriter.Null);
            try
            {
                var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"),
                    skillSpec: new SkillSpec(directory, "primary"));
                var before = agent.SkillManager()!.Active();
                var result = await agent.DispatchAsync(SkillCall(agent, "load_skill", "remote"))
                    .WaitAsync(TimeSpan.FromSeconds(10));
                Check(result.IsError && agent.SkillManager()!.Active().SequenceEqual(before),
                    "Failed dependency closure published manuals");
                Check(agent.Mcp()!.Clients().IsEmpty && agent.SkillManager()!.Clients().IsEmpty
                    && !agent.Tools().Any(tool => tool.Remote() is not null), "Rollback retained remote ownership");
                foreach (string pidFile in Directory.GetFiles(directory, "*.pid"))
                {
                    Check(ProcessStopped(int.Parse(File.ReadAllText(pidFile))), "Rollback left a process alive");
                }
                Check(agent.Tools().Any(tool => tool.Name() == "read_file"), "Rollback destroyed a preexisting local tool");
            }
            finally
            {
                root.Destroy();
                Directory.Delete(directory, true);
            }
        }
    }

    public static async Task SkillShutdown()
    {
        await CancelSkill();
        await ActorTests.InterruptSkillPreparation();
    }

    private static async Task CancelSkill()
    {
        string directory = SkillFixture();
        Manual(directory, "primary", "type: primary\ntools: write_file\nloadable-skills: remote\n", "Primary");
        Manual(directory, "remote", "type: loadable\ntools: remote_echo\n" + Server(directory, "remote_echo", "hold"), "Remote");
        var root = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"),
                skillSpec: new SkillSpec(directory, "primary"));
            var load = agent.DispatchAsync(SkillCall(agent, "load_skill", "remote"));
            await WaitFile(Path.Combine(directory, "remote_echo.ready"));
            int pid = int.Parse(File.ReadAllText(Path.Combine(directory, "remote_echo.pid")));
            await agent.ShutdownAsync().WaitAsync(TimeSpan.FromSeconds(10));
            try
            {
                await load.WaitAsync(TimeSpan.FromSeconds(5));
            }
            catch (Exception error) when (error is OperationCanceledException or ObjectDisposedException)
            {
            }
            Check(agent.SkillManager()!.Active().Length == 1 && agent.SkillManager()!.Clients().IsEmpty,
                "Cancellation allowed a late activation or retained its client");
            Check(ProcessStopped(pid), "Cancellation left the preparing process alive");
        }
        finally
        {
            root.Destroy();
            Directory.Delete(directory, true);
        }
    }
}

internal static partial class ActorTests
{
    public static async Task InterruptSkillPreparation()
    {
        string directory = McpTests.SkillFixture();
        McpTests.Manual(directory, "primary", "type: primary\ntools: write_file\nloadable-skills: remote\n", "Primary");
        McpTests.Manual(directory, "remote", "type: loadable\ntools: remote_echo\n"
            + McpTests.Server(directory, "remote_echo", "hold"), "Remote");
        var listener = new System.Net.Sockets.TcpListener(System.Net.IPAddress.Loopback, 0);
        listener.Start();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var root = new Application(TextWriter.Null);
        var server = Task.Run(async () =>
        {
            using var connection = await listener.AcceptTcpClientAsync(deadline.Token);
            using var stream = connection.GetStream();
            await ReadRequest(stream, deadline.Token);
            await ReplyJson(stream, new
            {
                content = new object[]
                {
                    new { type = "text", text = "I will load the remote skill." },
                    new { type = "tool_use", id = "held-load", name = "load_skill", input = new { name = "remote" } }
                },
                usage = new { input_tokens = 1, output_tokens = 1 }
            }, deadline.Token);
        });
        try
        {
            int port = ((System.Net.IPEndPoint)listener.LocalEndpoint).Port;
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", $"http://127.0.0.1:{port}"),
                skillSpec: new SkillSpec(directory, "primary"));
            var turn = agent.SendAsync("Load remote");
            await McpTests.WaitFile(Path.Combine(directory, "remote_echo.ready"));
            int pid = int.Parse(File.ReadAllText(Path.Combine(directory, "remote_echo.pid")));
            await agent.PostAsync(new AgentInput(InputKind.InterruptNow)).WaitAsync(deadline.Token);
            try
            {
                await turn.WaitAsync(deadline.Token);
            }
            catch (OperationCanceledException) when (!deadline.IsCancellationRequested)
            {
            }
            await agent.FlushAsync().WaitAsync(deadline.Token);
            await agent.PostAsync(new AgentInput(InputKind.Pause)).WaitAsync(deadline.Token);
            Check(agent.SkillManager()!.Active().Length == 1 && agent.SkillManager()!.Clients().IsEmpty,
                "Interrupted turn published a late activation or retained its connection");
            Check(McpTests.ProcessStopped(pid), "Interrupted preparation left a process alive");
            Check(agent.History().Events().Any(fact => fact.Kind == EventKind.ToolReturned && fact.CallID == "held-load"),
                "Interrupted load lost its exact-ID terminal result");
            await server.WaitAsync(deadline.Token);
        }
        finally
        {
            deadline.Cancel();
            listener.Stop();
            root.Destroy();
            try
            {
                await server;
            }
            catch (OperationCanceledException)
            {
            }
            Directory.Delete(directory, true);
        }
    }
}
