using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

static partial class McpTests
{
    static JsonElement J(string text) => JsonSerializer.Deserialize<JsonElement>(text);

    static void Check(bool condition, string message)
    {
        if (!condition)
        {
            throw new Exception(message);
        }
    }

    // The peer speaks the actual line-framed protocol. Only the external server
    // is replaced; Agent dispatch, file tools, History, and RPC correlation run.
    // Working failure boundary: bytes really enter the output buffer before
    // an I/O failure. The client must not reuse a potentially partial frame.
    sealed class FailingOutput : MemoryStream
    {
        public bool Closed { get; private set; }
        public override ValueTask WriteAsync(ReadOnlyMemory<byte> buffer, CancellationToken cancellationToken = default)
        {
            Write(buffer.Span[..Math.Min(4, buffer.Length)]);
            return ValueTask.FromException(new IOException("Output boundary failed mid-frame"));
        }
        protected override void Dispose(bool disposing)
        {
            Closed = true;
            base.Dispose(disposing);
        }
    }

    static async Task<(McpStreamTransport Client, McpStreamTransport Peer)> Pair(Stream? output = null)
    {
        using var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        var client = new TcpClient();
        Task connect = client.ConnectAsync(IPAddress.Loopback, ((IPEndPoint)listener.LocalEndpoint).Port);
        var peer = await listener.AcceptTcpClientAsync();
        await connect;
        // NetworkStream owns each socket when constructed this way. Dispose the
        // transports, not a second independent socket owner in the test.
        var clientStream = new NetworkStream(client.Client, ownsSocket: true);
        var peerStream = new NetworkStream(peer.Client, ownsSocket: true);
        return (new McpStreamTransport(clientStream, output ?? clientStream),
            new McpStreamTransport(peerStream, peerStream));
    }

    static async Task<JsonElement> Read(IMcpTransport peer) =>
        await peer.ReceiveAsync().WaitAsync(TimeSpan.FromSeconds(8))
            ?? throw new Exception("Unexpected MCP EOF");

    static Task Reply(IMcpTransport peer, JsonElement request, object result) =>
        peer.SendAsync(JsonSerializer.SerializeToElement(new
        {
            jsonrpc = "2.0", id = request.GetProperty("id"), result
        }));

    static async Task Fails(Task task)
    {
        try
        {
            await task.WaitAsync(TimeSpan.FromSeconds(8));
        }
        catch (Exception error) when (error is IOException or InvalidDataException
            or OperationCanceledException or ObjectDisposedException)
        {
            return;
        }
        throw new Exception("Expected a closed/cancelled/invalid RPC refusal");
    }

    public static async Task Correlation()
    {
        var streams = await Pair();
        await using var peer = streams.Peer;
        var root = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1"));
            var manager = new Mcp(agent);
            var client = new McpClient(manager, "test", streams.Client);
            Task connected = client.ConnectAsync();
            var initialize = await Read(peer);
            Check(initialize.GetProperty("method").GetString() == "initialize", "initialize must come first");
            await Reply(peer, initialize, new { protocolVersion = "2024-11-05", capabilities = new { } });
            var notice = await Read(peer);
            Check(notice.GetProperty("method").GetString() == "notifications/initialized"
                && !notice.TryGetProperty("id", out _), "initialized is a notification");
            var list = await Read(peer);
            Check(list.GetProperty("method").GetString() == "tools/list", "discovery follows initialized");
            await Reply(peer, list, new
            {
                tools = new[] { new { name = "remote", description = "A peer tool",
                    inputSchema = new { type = "object" }, ephemeral = "round" } }
            });
            await connected.WaitAsync(TimeSpan.FromSeconds(8));
            Check(client.Catalog().Single().Ephemeral == "round", "discovery preserves metadata");
            Check(!agent.Tools().Any(tool => tool.Name() == "remote"), "discovery alone must not grant execution");

            // Two independent requests finish in reverse order. Result values
            // must follow IDs, never the order in which the reader sees them.
            Task<JsonElement> first = client.RequestAsync("first", J("{}"));
            Task<JsonElement> second = client.RequestAsync("second", J("{}"));
            var requests = new[] { await Read(peer), await Read(peer) };
            await Reply(peer, requests[1], new { value = "second" });
            await Reply(peer, requests[0], new { value = "first" });
            Check((await first).GetProperty("value").GetString() == "first", "first reply correlation");
            Check((await second).GetProperty("value").GetString() == "second", "second reply correlation");

            using var cancelled = new CancellationTokenSource();
            var abandoned = client.RequestAsync("cancelled", J("{}"), cancelled.Token);
            var late = await Read(peer);
            cancelled.Cancel();
            await Fails(abandoned);
            await Reply(peer, late, new { ignored = true });
            var live = client.RequestAsync("still-live", J("{}"));
            await Reply(peer, await Read(peer), new { alive = true });
            Check((await live).GetProperty("alive").GetBoolean(), "late reply cannot poison another request");

            // Concurrent sends must remain whole newline-delimited envelopes.
            var many = Enumerable.Range(0, 20).Select(index => client.RequestAsync("parallel", J($"{{\"index\":{index}}}"))).ToArray();
            for (int i = 0; i < many.Length; i++)
            {
                var request = await Read(peer);
                await Reply(peer, request, request.GetProperty("params"));
            }
            var results = await Task.WhenAll(many).WaitAsync(TimeSpan.FromSeconds(8));
            Check(results.Select(result => result.GetProperty("index").GetInt32()).SequenceEqual(Enumerable.Range(0, 20)),
                "concurrent writes preserve framing and identities");
            Task destruction = client.DestroyAsync();
            // Admission closes at the call, not whenever asynchronous disposal
            // happens to run. RequestAsync must already have refused this call.
            var refused = client.RequestAsync("after-destroy", J("{}"));
            Check(refused.IsCompleted, "destroy request immediately closes RPC admission");
            await Fails(refused);
            try
            {
                _ = client.ServeAsync();
                throw new Exception("Destroyed client admitted service startup");
            }
            catch (ObjectDisposedException)
            {
                // Service startup has the same synchronous admission boundary.
            }
            await Task.WhenAll(destruction, client.DestroyAsync());
            Check(manager.Clients().Length == 0, "client destruction removes manager membership");
            await manager.DestroyAsync();
            Check(agent.Mcp() is null, "manager destruction detaches both ends");
        }
        finally
        {
            root.Destroy();
        }
    }

    public static async Task ReverseAndClosure(string temp)
    {
        var streams = await Pair();
        await using var peer = streams.Peer;
        var root = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1"));
            agent.InstallTools();
            var started = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            _ = new Tool(agent, "held_local", "Hold one real local tool turn", J("{\"type\":\"object\"}"),
                async (job, args, cancellation) =>
                {
                    started.TrySetResult();
                    await release.Task.WaitAsync(cancellation);
                    return "released";
                });
            var heldTurn = agent.DispatchAsync(new ToolCallPart("held-local", agent.Configuration().Target(),
                "held_local", J("{\"ai_callback_delay\":30}")));
            await started.Task.WaitAsync(TimeSpan.FromSeconds(8));
            var manager = new Mcp(agent);
            var client = new McpClient(manager, "reverse", streams.Client);
            var outbound = client.RequestAsync("held", J("{}"));
            var held = await Read(peer);
            string path = Path.Combine(temp, "mcp-reverse.txt");
            await peer.SendAsync(JsonSerializer.SerializeToElement(new
            {
                jsonrpc = "2.0", id = "reverse-1", method = "tools/call",
                @params = new { name = "write_file", arguments = new { path, content = "actual reverse effect" } }
            }));
            var reversed = await Read(peer);
            Check(reversed.GetProperty("id").GetString() == "reverse-1", "reverse reply preserves string ID");
            Check(!reversed.GetProperty("result").GetProperty("isError").GetBoolean(), "reverse local tool succeeds");
            Check(File.ReadAllText(path) == "actual reverse effect", "reverse dispatch performed real effect");
            Check(!outbound.IsCompleted, "reverse request does not fabricate outbound completion");
            Check(!heldTurn.IsCompleted, "reverse request completed before the held turn released");
            Check(agent.History().Events().Any(fact => fact.Kind == EventKind.ToolObserved && fact.CallID == "mcp:\"reverse-1\""),
                "reverse result is a host observation, not an orphan model result");
            release.TrySetResult();
            await heldTurn.WaitAsync(TimeSpan.FromSeconds(8));
            await Reply(peer, held, new { finished = true });
            Check((await outbound).GetProperty("finished").GetBoolean(), "reader continued after reverse call");

            await peer.SendAsync(J("{\"jsonrpc\":\"2.0\",\"id\":9,\"method\":\"tools/list\",\"params\":{}}"));
            var tools = await Read(peer);
            Check(tools.GetProperty("result").GetProperty("tools").EnumerateArray()
                .Any(tool => tool.GetProperty("name").GetString() == "read_file"), "reverse discovery maps local schemas");
            await peer.SendAsync(J("{\"jsonrpc\":\"2.0\",\"id\":10,\"method\":\"unknown\",\"params\":{}}"));
            Check((await Read(peer)).GetProperty("error").GetProperty("code").GetInt32() == -32601,
                "unknown reverse method gets an error, not a silent hang");

            var pending = client.RequestAsync("never-replied", J("{}"));
            await Read(peer);
            await peer.DisposeAsync();
            await Fails(pending);
            await client.DestroyAsync().WaitAsync(TimeSpan.FromSeconds(8));
            Check(manager.Clients().IsEmpty, "disconnected client detaches on destruction");
        }
        finally
        {
            root.Destroy();
        }
    }

    public static async Task RefusalAndShutdown()
    {
        var streams = await Pair();
        await using var peer = streams.Peer;
        var root = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1"));
            var manager = new Mcp(agent);
            var client = new McpClient(manager, "bad", streams.Client);
            var connected = client.ConnectAsync();
            await Reply(peer, await Read(peer), new { protocolVersion = "2024-11-05" });
            await Read(peer);
            var list = await Read(peer);
            var definition = new { name = "duplicate", inputSchema = new { type = "object" } };
            await Reply(peer, list, new { tools = new[] { definition, definition } });
            await Fails(connected);
            Check(client.Catalog().IsEmpty && manager.Clients().IsEmpty, "failed discovery publishes nothing and detaches");
            await manager.DestroyAsync();
            Check(agent.Mcp() is null, "failed startup manager remains destroyable");

            var blockedStreams = await Pair();
            await using var blockedPeer = blockedStreams.Peer;
            var next = new Mcp(agent);
            var blocked = new McpClient(next, "blocked", blockedStreams.Client);
            var waiting = blocked.RequestAsync("held", J("{}"));
            await Read(blockedPeer);
            await agent.ShutdownAsync().WaitAsync(TimeSpan.FromSeconds(8));
            await Fails(waiting);
            Check(next.Clients().IsEmpty && agent.Mcp() is null, "close wakes pending replies and removes both relationships");

            var failingOutput = new FailingOutput();
            var brokenStreams = await Pair(failingOutput);
            await using var brokenPeer = brokenStreams.Peer;
            var other = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1"));
            var brokenManager = new Mcp(other);
            var broken = new McpClient(brokenManager, "partial-write", brokenStreams.Client);
            await Fails(broken.RequestAsync("will-fail", J("{}")));
            Check(failingOutput.Closed && brokenManager.Clients().IsEmpty,
                "failed send closes owned boundary and detaches the client");
            await Fails(broken.RequestAsync("cannot-reuse", J("{}")));
        }
        finally
        {
            root.Destroy();
        }
    }
}
