using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    private static string AutomaticFixture(bool dynamic = false, string mode = "normal")
    {
        string folder = McpTests.SkillFixture();
        File.WriteAllText(Path.Combine(folder, "peer.py"), """
            import json, os, sys, time
            mode, folder, name = sys.argv[1:]
            def send(id, result):
                print(json.dumps({'jsonrpc':'2.0','id':id,'result':result}), flush=True)
            counts = {}
            for line in sys.stdin:
                request = json.loads(line)
                method = request.get('method')
                id = request.get('id')
                if method == 'initialize':
                    send(id, {'protocolVersion':request['params']['protocolVersion'],
                              'capabilities':{},'serverInfo':{'name':'snapshots','version':'1'}})
                elif method == 'tools/list':
                    schema = {'type':'object','properties':{},'additionalProperties':False}
                    send(id, {'tools':[{'name':n,'inputSchema':schema} for n in
                        ['per_turn_status','per_round_view','ordinary_echo','per_round_ungranted']]})
                elif method == 'tools/call':
                    tool = request['params']['name']
                    counts[tool] = counts.get(tool, 0) + 1
                    with open(os.path.join(folder, 'calls'), 'a') as f:
                        f.write(tool + '\n')
                    if mode == 'hold' and tool == 'per_round_view':
                        open(os.path.join(folder, 'ready'), 'w').close()
                        while not os.path.exists(os.path.join(folder, 'release')):
                            time.sleep(0.01)
                    text = tool + '-VALUE-' + str(counts[tool])
                    send(id, {'content':[{'type':'text','text':text}], 'isError':mode == 'error'})
            """);
        string grants = "tools: per_turn_status per_round_view ordinary_echo\n";
        if (dynamic)
        {
            McpTests.Manual(folder, "primary", "type: primary\nloadable-skills: snapshots\ntools: read_file\n", "PRIMARY $TOOLS");
            McpTests.Manual(folder, "snapshots", "type: loadable\n" + grants + McpTests.Server(folder, "snapshots", mode), "SNAPSHOTS $TOOLS");
        }
        else
        {
            McpTests.Manual(folder, "primary", "type: primary\n" + grants.Replace("ordinary_echo", "ordinary_echo read_file")
                + McpTests.Server(folder, "snapshots", mode), "PRIMARY $TOOLS");
        }
        return folder;
    }

    public static async Task AutomaticSnapshots()
    {
        string folder = AutomaticFixture();
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        var root = new Application(TextWriter.Null);
        try
        {
            var captured = new List<string>();
            var server = Task.Run(async () =>
            {
                for (int index = 0; index < 3; index++)
                {
                    using var connection = await listener.AcceptTcpClientAsync(deadline.Token);
                    using var stream = connection.GetStream();
                    string body = await ReadRequest(stream, deadline.Token);
                    captured.Add(body);
                    using var document = JsonDocument.Parse(body);
                    var request = document.RootElement;
                    Check(!request.GetProperty("system").GetString()!.Contains("per_"), "$TOOLS leaked automatic names");
                    Check(request.GetProperty("tools").EnumerateArray().All(tool =>
                        !tool.GetProperty("name").GetString()!.StartsWith("per_")), "Automatic schema reached provider");
                    int turn = index == 2 ? 2 : 1;
                    Check(body.Contains($"per_turn_status-VALUE-{turn}") && body.Contains($"per_round_view-VALUE-{index + 1}"),
                        "Snapshot frequency or request injection was wrong");
                    Check(!body.Contains("per_round_ungranted"), "Catalog-only snapshot gained authority");
                    if (index > 0)
                    {
                        Check(!body.Contains($"per_round_view-VALUE-{index}"), "Previous round snapshot became dialogue");
                    }
                    if (index == 0)
                    {
                        await ReplyJson(stream, new { content = new object[] {
                            new { type = "tool_use", id = "read", name = "read_file", input = new { path = Path.Combine(folder, "calls") } }
                        } }, deadline.Token);
                    }
                    else
                    {
                        await Reply(stream, "complete", deadline.Token);
                    }
                }
            });
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "model-a",
                $"http://127.0.0.1:{((IPEndPoint)listener.LocalEndpoint).Port}", deferredTools: true),
                skillSpec: new SkillSpec(folder, "primary"));
            await agent.SkillManager()!.InitializeAsync();
            Check(!agent.ToolEnabled("per_turn_status") && agent.Tools().All(tool => !tool.Name().StartsWith("per_")),
                "Automatic definition became a callable tool");
            await agent.PostAsync(new AgentInput(InputKind.Ephemeral, "MANUAL-ONCE"));
            Check(await agent.SendAsync("first").WaitAsync(deadline.Token) == "complete", "First turn did not complete");
            Check(await agent.SendAsync("second").WaitAsync(deadline.Token) == "complete", "Second turn did not complete");
            await server.WaitAsync(deadline.Token);
            Check(captured[0].Contains("MANUAL-ONCE") && captured.Skip(1).All(body => !body.Contains("MANUAL-ONCE")),
                "Manual ephemera was lost or repeated");
            string[] calls = File.ReadAllLines(Path.Combine(folder, "calls"));
            Check(calls.Count(name => name == "per_turn_status") == 2 && calls.Count(name => name == "per_round_view") == 3,
                "Automatic jobs ran at the wrong frequency");
            var history = agent.History();
            Check(history.View().Ephemeral.IsEmpty && history.View().Dialogue.All(entry =>
                !JsonCodec.Text(entry.Parts).Contains("-VALUE-")), "Snapshot became durable conversation");
            // Replay each request's exact input from the event prefix. No live
            // server, collector, or synthetic result is needed during replay.
            var replay = new Agent(root, new Configuration(Vendor.Anthropic, "model-a", "http://127.0.0.1:1", deferredTools: true));
            replay.InstallTools();
            int requestIndex = 0;
            foreach (var fact in history.Events())
            {
                if (fact.Kind == EventKind.RequestSent)
                {
                    string wire = replay.Engine().Render(replay.History().View()).ToJsonString();
                    Check(wire.Contains($"per_round_view-VALUE-{++requestIndex}"), "Replay lost recorded ephemera");
                }
                replay.History().Append(fact);
            }
            Check(requestIndex == 3, "Replay missed a request");
            foreach (Vendor vendor in Enum.GetValues<Vendor>())
            {
                var renderer = new Agent(root, new Configuration(vendor, "model-a", "http://127.0.0.1:1"));
                renderer.InstallTools();
                string wire = renderer.Engine().Render(history.View()).ToJsonString();
                using var body = JsonDocument.Parse(wire);
                string declarations = body.RootElement.TryGetProperty("tools", out var tools) ? tools.GetRawText() : "";
                Check(!declarations.Contains("per_turn_status") && !declarations.Contains("per_round_view"),
                    "Automatic names leaked through another renderer");
            }
        }
        finally
        {
            deadline.Cancel();
            listener.Stop();
            root.Destroy();
            Directory.Delete(folder, true);
        }
    }

    public static async Task AutomaticInterrupt()
    {
        foreach (InputKind interrupt in new[] { InputKind.InterruptNow, InputKind.InterruptAfterTool })
        {
            string folder = AutomaticFixture(mode: "hold");
            var listener = new TcpListener(IPAddress.Loopback, 0);
            listener.Start();
            var root = new Application(TextWriter.Null);
            try
            {
                var agent = new Agent(root, new Configuration(Vendor.Anthropic, "model-a",
                    $"http://127.0.0.1:{((IPEndPoint)listener.LocalEndpoint).Port}"), skillSpec: new SkillSpec(folder, "primary"));
                await agent.SkillManager()!.InitializeAsync();
                var turn = agent.SendAsync("held snapshot");
                await McpTests.WaitFile(Path.Combine(folder, "ready"));
                Check(!listener.Pending(), "HTTP started before automatic completion");
                await agent.PostAsync(new AgentInput(InputKind.Hint, "Hint must not cancel collection")).WaitAsync(TimeSpan.FromSeconds(5));
                await agent.PostAsync(new AgentInput(InputKind.Pause)).WaitAsync(TimeSpan.FromSeconds(5));
                Check(!turn.IsCompleted && !listener.Pending(), "Hint or pause abandoned the collector");
                await agent.PostAsync(new AgentInput(interrupt)).WaitAsync(TimeSpan.FromSeconds(5));
                try
                {
                    await turn.WaitAsync(TimeSpan.FromSeconds(5));
                    throw new Exception("Interrupted turn reported success");
                }
                catch (OperationCanceledException)
                {
                }
                var running = agent.Jobs().Children().Single(job => job.Snapshot().Status == "running");
                Check(!listener.Pending(), "Interrupted collector sent a request");
                File.WriteAllText(Path.Combine(folder, "release"), "release");
                ToolResultPart result;
                do
                {
                    result = await running.ObserveAsync("inspect", 1, "", 10000);
                } while (result.Job!.Status == "running");
                Check(!result.IsError && JsonCodec.Text(result.Parts).Contains("per_round_view-VALUE-1"),
                    "Interrupt killed or discarded independent automatic output");
                Check(agent.History().View().Ephemeral.IsEmpty, "Interrupted output contaminated the next request slot");
                await agent.ShutdownAsync().WaitAsync(TimeSpan.FromSeconds(10));
            }
            catch (Exception error)
            {
                Console.WriteLine("Automatic interrupt original failure: " + error);
                throw;
            }
            finally
            {
                listener.Stop();
                root.Destroy();
                Directory.Delete(folder, true);
            }
        }
    }

    public static async Task AutomaticDiscovery()
    {
        string folder = AutomaticFixture(dynamic: true);
        var root = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "model-a", "http://127.0.0.1:1", deferredTools: true),
                skillSpec: new SkillSpec(folder, "primary"));
            ToolCallPart Call(string name, object args) => new(Guid.NewGuid().ToString("N"),
                agent.Configuration().Target(), name, JsonSerializer.SerializeToElement(args));
            var loaded = await agent.DispatchAsync(Call("load_skill", new { name = "snapshots" }));
            Check(!loaded.IsError && loaded.ToolReferences.SequenceEqual(new[] { "ordinary_echo" }),
                "Automatic discovery leaked references or rejected valid grants");
            Check(!agent.ToolEnabled("per_turn_status") && !agent.ToolEnabled("per_round_view"), "Enabled names leaked automatic tools");
            Check(agent.SkillManager()!.Active().Single(skill => skill.Name == "snapshots").Tools.Contains("per_turn_status"),
                "Automatic scheduling grant was lost");
            var discoveryCall = Call("load_skill", new { name = "snapshots" });
            agent.History().Append(new Event(EventKind.ResponseEnded, actor: Actor.Agent, parts: [discoveryCall]));
            agent.History().Append(new Event(EventKind.ToolReturned, callID: discoveryCall.CallID,
                parts: [new TextPart("discovery")], toolReferences: ["ordinary_echo", "per_turn_status", "per_round_view"]));
            string replayWire = agent.Engine().Render(agent.History().View()).ToJsonString();
            Check(replayWire.Contains("tool_reference") && replayWire.Contains("ordinary_echo")
                && !replayWire.Contains("per_turn_status") && !replayWire.Contains("per_round_view"),
                "Recorded automatic reference leaked during replay");
            Check(agent.History().Events().Last().ToolReferences.Contains("per_turn_status"),
                "Request filtering rewrote canonical discovery metadata");
            var refused = await agent.DispatchAsync(Call("per_round_view", new { }));
            Check(refused.IsError && !File.Exists(Path.Combine(folder, "calls")), "Guessed automatic name ran through ordinary dispatch");
            await agent.DispatchAsync(Call("unload_skill", new { name = "snapshots" }));
            Check(agent.SkillManager()!.Clients().IsEmpty && agent.Tools().All(tool => tool.Name() != "ordinary_echo"),
                "Unloading automatic skill retained its connection or tools");
        }
        finally
        {
            root.Destroy();
            Directory.Delete(folder, true);
        }
    }

    public static async Task AutomaticPauseAndError()
    {
        foreach (string mode in new[] { "hold", "error" })
        {
            string folder = AutomaticFixture(mode: mode);
            var listener = new TcpListener(IPAddress.Loopback, 0);
            listener.Start();
            using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
            var root = new Application(TextWriter.Null);
            try
            {
                var agent = new Agent(root, new Configuration(Vendor.Anthropic, "model-a",
                    $"http://127.0.0.1:{((IPEndPoint)listener.LocalEndpoint).Port}"), skillSpec: new SkillSpec(folder, "primary"));
                await agent.SkillManager()!.InitializeAsync();
                var turn = agent.SendAsync("snapshot");
                if (mode == "hold")
                {
                    await McpTests.WaitFile(Path.Combine(folder, "ready"));
                    await agent.PostAsync(new AgentInput(InputKind.Pause));
                    File.WriteAllText(Path.Combine(folder, "release"), "release");
                    foreach (var job in agent.Jobs().Children())
                    {
                        // Observing the real completion must not steal the
                        // structured result from the actor's collector.
                        ToolResultPart result;
                        do
                        {
                            result = await job.ObserveAsync("host", 1, "", 10000, deadline.Token);
                        } while (result.Job!.Status == "running");
                        Check(!result.IsError, "Held automatic job did not finish");
                    }
                    Check(!listener.Pending() && !turn.IsCompleted, "Pause allowed HTTP after snapshot completion");
                    await agent.PostAsync(new AgentInput(InputKind.Resume));
                }
                using var connection = await listener.AcceptTcpClientAsync(deadline.Token);
                using var stream = connection.GetStream();
                string body = await ReadRequest(stream, deadline.Token);
                if (mode == "error")
                {
                    Check(body.Contains("unavailable (error, job") && !body.Contains("-VALUE-"),
                        "Server failure became snapshot data rather than a brief diagnostic");
                    foreach (var job in agent.Jobs().Children())
                    {
                        var result = await job.ObserveAsync("host", 1, "", 10000, deadline.Token);
                        Check(result.IsError && result.Parts.OfType<TextPart>().Any(part => part.Text.Contains("-VALUE-1")),
                            "Failed automatic bytes were lost instead of retained on the Job");
                    }
                }
                else
                {
                    Check(body.Contains("per_turn_status-VALUE-1") && body.Contains("per_round_view-VALUE-1"),
                        "Host observation consumed automatic snapshot data");
                }
                await Reply(stream, "complete", deadline.Token);
                Check(await turn.WaitAsync(deadline.Token) == "complete", "Snapshot turn did not finish");
                Check(agent.History().View().Ephemeral.IsEmpty, "Completed request retained snapshot slot");
                await agent.ShutdownAsync().WaitAsync(deadline.Token);
            }
            finally
            {
                deadline.Cancel();
                listener.Stop();
                root.Destroy();
                Directory.Delete(folder, true);
            }
        }
    }

    public static async Task AutomaticFailedRequest()
    {
        string folder = AutomaticFixture();
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var root = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "model-a",
                $"http://127.0.0.1:{((IPEndPoint)listener.LocalEndpoint).Port}"), skillSpec: new SkillSpec(folder, "primary"));
            await agent.SkillManager()!.InitializeAsync();
            var turn = agent.SendAsync("failed transport");
            using (var connection = await listener.AcceptTcpClientAsync(deadline.Token))
            {
                string body = await ReadRequest(connection.GetStream(), deadline.Token);
                Check(body.Contains("per_round_view-VALUE-1"), "Snapshot missing before failed response");
                // Close without an HTTP response: a real transport failure.
            }
            try
            {
                await turn.WaitAsync(deadline.Token);
                throw new Exception("Failed HTTP transport reported success");
            }
            catch (InvalidOperationException error) when (error.Message.Contains("HTTP exchange failed"))
            {
            }
            Check(agent.History().View().Ephemeral.IsEmpty, "Failed send retained automatic ephemera");
            Check(agent.History().Events().Count(fact => fact.Kind == EventKind.RequestSent) == 1,
                "Failed send lost or repeated its request attempt");
            try
            {
                await agent.SubmitAsync([new BlobPart("image/png", new BlobRef(RefKind.Path,
                    Path.Combine(folder, "deliberately-absent.png")))]).WaitAsync(deadline.Token);
                throw new Exception("Unsupported image reported success");
            }
            catch (NotSupportedException)
            {
            }
            Check(agent.History().View().Ephemeral.IsEmpty && !listener.Pending()
                && agent.History().Events().Count(fact => fact.Kind == EventKind.RequestSent) == 1,
                "Rendering refusal retained ephemera or reached the transport boundary");
            await agent.ShutdownAsync().WaitAsync(deadline.Token);
        }
        finally
        {
            deadline.Cancel();
            listener.Stop();
            root.Destroy();
            Directory.Delete(folder, true);
        }
    }
}
