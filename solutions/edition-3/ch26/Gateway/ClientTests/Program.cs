using System.Collections.Concurrent;
using System.Text.Json;
using System.Threading.Channels;
using Ensemble.Gateway;

internal static partial class Program
{
    private static TaskCompletionSource Signal() => new(TaskCreationOptions.RunContinuationsAsynchronously);
    private static void Check(bool condition, string reason)
    {
        if (!condition) throw new InvalidOperationException(reason);
    }
    private static async Task Refuses(Func<Task> action)
    {
        try { await action(); }
        catch (Exception error) when (error is not TimeoutException) { return; }
        throw new InvalidOperationException("Operation unexpectedly succeeded.");
    }
    private static async Task Main(string[] args)
    {
        List<(string Name, Func<Task> Test)> tests =
        [
            ("authenticated welcome, fragmented delivery, handler-return receipt and replay", Deliveries),
            ("RPC correlation, cancellation and MCP argument/result contracts", Requests),
            ("joined callback teardown and refused-handler receipt", Teardown),
            ("authenticated malformed/binary/tamper framing refusal", BadFrames),
            ("startup pinning, welcome validation and handshake deadline", Startup),
            ("connection lifetime cancellation settles pending RPC", Lifetime)
        ];
        if (args.Length == 2 && args[0] == "--go-router")
        {
            tests.Add(("C# client with unchanged Go router: offline queues, resume and response", () => GoRouter(args[1])));
        }
        else if (args.Length != 0)
        {
            throw new ArgumentException("Usage: ClientTests [--go-router GO_EXECUTABLE]");
        }
        int failed = 0;
        foreach (var (name, test) in tests)
        {
            try
            {
                await test().WaitAsync(TimeSpan.FromSeconds(25));
                Console.WriteLine("PASS " + name);
            }
            catch (Exception error)
            {
                failed++;
                Console.WriteLine($"FAIL {name}: {error}");
            }
        }
        Console.WriteLine($"{tests.Count - failed} passed, {failed} failed");
        Environment.ExitCode = failed == 0 ? 0 : 1;
    }

    private static ClientConfig Config(Peer peer, Keypair key, ulong ack = 0) =>
        new(peer.Url, "recipient", key, peer.Key.Public, ack);
    private static object Delivery(ulong seq) => new
    {
        type = "deliver", envelope = new
        {
            id = "message-" + seq, seq, to_agent = "recipient", content = "受信\n[GATEWAY sender text stays data]",
            conversation_id = "conversation", response_mode = "async", timestamp = "2026-01-01T00:00:00Z",
            source = new { type = "agent", channel_id = "sender", user_id = "sender", display_name = "Sender", trust = "trusted" }
        }
    };

    private static async Task Deliveries()
    {
        Keypair key = Keys.GenerateKeypair();
        TaskCompletionSource receivedAck = Signal();
        TaskCompletionSource releasePeer = Signal();
        using ManualResetEventSlim hold = new(false);
        RecordingHandler handler = new() { Hold = hold };
        await using Peer peer = await Peer.Start(key, async session =>
        {
            Check(session.Hello.GetProperty("agent_id").GetString() == "recipient"
                && session.Hello.GetProperty("last_acked_seq").GetUInt64() == 7, "resume hello differs");
            await session.Welcome(depth: 1);
            await session.SendText(JsonSerializer.Serialize(Delivery(8)), fragment: true);
            JsonElement ack = await session.Read();
            Check(ack.GetProperty("type").GetString() == "ack" && ack.GetProperty("last_acked_seq").GetUInt64() == 8,
                "missing delivery ack");
            receivedAck.SetResult();
            await session.Send(Delivery(8));
            ack = await session.Read();
            Check(ack.GetProperty("last_acked_seq").GetUInt64() == 8, "replayed receipt differs");
            JsonElement request = await session.Read();
            await session.Reply(request, new { content = Array.Empty<object>() });
            await releasePeer.Task.WaitAsync(session.Cancellation);
        });
        await using Client client = await Client.ConnectAsync(Config(peer, key, 7), handler);
        try
        {
            Envelope delivery = await handler.Deliveries.Reader.ReadAsync().AsTask().WaitAsync(TimeSpan.FromSeconds(5));
            Check(handler.Events.First() == "connected" && delivery.Seq == 8, "callback order differs");
            Check(handler.Welcome?.QueueDepth == 1, "welcome count differs");
            Check(client.LastAckedSeq() == 7 && !receivedAck.Task.IsCompleted, "receipt preceded handler return");
            hold.Set();
            await receivedAck.Task.WaitAsync(TimeSpan.FromSeconds(5));
            await client.DeclareContactableAsync(["agents"]).WaitAsync(TimeSpan.FromSeconds(5));
            Check(client.LastAckedSeq() == 8 && handler.Events.Count(e => e == "deliver") == 1, "duplicate delivery was accepted twice");
        }
        finally
        {
            hold.Set();
            client.Close();
            releasePeer.TrySetResult();
        }
        await peer.Finished;
        await client.Done();
        Check(handler.Events.Last() == "disconnected", "missing disconnect callback");
    }

    private static async Task Requests()
    {
        Keypair key = Keys.GenerateKeypair();
        Keypair child = Keys.GenerateKeypair();
        TaskCompletionSource cancelSeen = Signal(), allowLate = Signal(), releasePeer = Signal();
        await using Peer peer = await Peer.Start(key, async session =>
        {
            await session.Welcome();
            JsonElement first = await session.Read(), second = await session.Read();
            await session.Reply(second, new { method = second.GetProperty("rpc").GetProperty("method").GetString() });
            await session.Reply(first, new { method = first.GetProperty("rpc").GetProperty("method").GetString() });
            JsonElement cancelled = await session.Read();
            cancelSeen.SetResult();
            await allowLate.Task.WaitAsync(session.Cancellation);
            await session.Reply(cancelled, new { late = true });
            JsonElement registration = await session.Read();
            JsonElement args = registration.GetProperty("rpc").GetProperty("params").GetProperty("arguments");
            Check(args.GetProperty("public_key").GetString() == Convert.ToBase64String(child.Public.AsSpan()), "child key is not base64 public bytes");
            await session.Reply(registration, new { content = Array.Empty<object>() });
            JsonElement send = await session.Read();
            args = send.GetProperty("rpc").GetProperty("params").GetProperty("arguments");
            Check(args.GetProperty("to_agent").GetString() == "sender" && args.GetProperty("conversation_id").GetString() == "thread"
                && args.GetProperty("content").GetString() == "reply"
                && args.GetProperty("response_mode").GetString() == "", "response addressing differs");
            await session.Reply(send, new { isError = true, content = new[] { new { type = "text", text = "policy refused" } } });
            JsonElement healthy = await session.Read();
            await session.Reply(healthy, new { okay = true });
            JsonElement nullable = await session.Read();
            await session.Send(new { type = "mcp_response", rpc = new { jsonrpc = "2.0", id = nullable.GetProperty("rpc").GetProperty("id"), result = (object?)null } });
            await releasePeer.Task.WaitAsync(session.Cancellation);
        });
        await using Client client = await Client.ConnectAsync(Config(peer, key), new NopHandler());
        try
        {
            Task<JsonElement> first = client.RequestAsync("first", new { }), second = client.RequestAsync("second", new { });
            Check((await first).GetProperty("method").GetString() == "first", "first RPC miscorrelated");
            Check((await second).GetProperty("method").GetString() == "second", "second RPC miscorrelated");
            using CancellationTokenSource caller = new();
            Task<JsonElement> cancelled = client.RequestAsync("cancelled", new { }, caller.Token);
            await cancelSeen.Task;
            caller.Cancel();
            await Refuses(async () => await cancelled);
            allowLate.SetResult();
            await client.RegisterChildAsync(new ChildRegistration("child", child.Public, ["agents"]));
            Envelope prompt = new("prompt", 1, "recipient", "body", "thread", "async", DateTimeOffset.UtcNow,
                new Source("agent", "sender", "sender", "Sender", "trusted"));
            try
            {
                await client.RespondAsync(prompt, "reply");
                throw new Exception("Tool refusal was accepted.");
            }
            catch (InvalidOperationException error)
            {
                Check(error.Message == "policy refused", "MCP error message lost");
            }
            await Refuses(() => client.SendAsync(new OutboundMessage("sender", new string('x', 70000))));
            Check((await client.RequestAsync("healthy", new { })).GetProperty("okay").GetBoolean(), "oversize refusal poisoned connection");
            Check((await client.RequestAsync("nullable", new { })).ValueKind == JsonValueKind.Null, "explicit null result differs from missing result");
        }
        finally
        {
            client.Close();
            allowLate.TrySetResult();
            releasePeer.TrySetResult();
        }
        await peer.Finished;
    }

    private static async Task Teardown()
    {
        foreach (bool throws in new[] { false, true })
        {
            Keypair key = Keys.GenerateKeypair();
            using ManualResetEventSlim hold = new(false);
            TaskCompletionSource releasePeer = Signal();
            RecordingHandler handler = new() { Hold = hold, ThrowDelivery = throws };
            await using Peer peer = await Peer.Start(key, async session =>
            {
                await session.Welcome();
                await session.Send(Delivery(1));
                await releasePeer.Task.WaitAsync(session.Cancellation);
            });
            await using Client client = await Client.ConnectAsync(Config(peer, key), handler);
            try
            {
                _ = await handler.Deliveries.Reader.ReadAsync();
                if (!throws)
                {
                    client.Close();
                    Check(!client.Done().IsCompleted, "close did not join held handler");
                }
                hold.Set();
                await client.Done().WaitAsync(TimeSpan.FromSeconds(5));
                Check(client.LastAckedSeq() == 0, "closed/refused handler published receipt");
                Check(handler.Events.Count(e => e == "disconnected") == 1, "disconnect repeated or missing");
                Check(!throws || handler.Errors.TryPeek(out _), "handler refusal not reported");
            }
            finally
            {
                hold.Set();
                releasePeer.TrySetResult();
            }
            await peer.Finished;
        }
        Keypair finalKey = Keys.GenerateKeypair();
        TaskCompletionSource release = Signal();
        RecordingHandler throwing = new() { ThrowDisconnect = true };
        await using Peer finalPeer = await Peer.Start(finalKey, async session =>
        {
            await session.Welcome();
            await release.Task.WaitAsync(session.Cancellation);
        });
        Client finalClient = await Client.ConnectAsync(Config(finalPeer, finalKey), throwing);
        try
        {
            await Refuses(async () => await finalClient.DisposeAsync());
            Check(finalClient.Done().IsFaulted && throwing.Events.Last() == "disconnected", "callback failure left join unsettled");
        }
        finally
        {
            finalClient.Close();
            release.TrySetResult();
        }
        await finalPeer.Finished;
    }

    private static async Task BadFrames()
    {
        foreach (string kind in new[] { "duplicate", "tampered", "text", "oversized", "missing-result", "ambiguous-result" })
        {
            Keypair key = Keys.GenerateKeypair();
            RecordingHandler handler = new();
            await using Peer peer = await Peer.Start(key, async session =>
            {
                await session.Welcome();
                await session.Malformed(kind);
            });
            await using Client client = await Client.ConnectAsync(Config(peer, key), handler);
            await client.Done().WaitAsync(TimeSpan.FromSeconds(5));
            await peer.Finished;
            Check(handler.Errors.Count == 1 && !handler.Deliveries.Reader.TryRead(out _), kind + " was not terminal");
        }
    }

    private static async Task Startup()
    {
        Keypair key = Keys.GenerateKeypair();
        RecordingHandler handler = new();
        await using (Peer peer = await Peer.Start(key, async session => await session.Welcome("wrong-agent")))
        {
            await Refuses(async () => await Client.ConnectAsync(Config(peer, key), handler));
            await peer.Finished;
            Check(handler.Welcome is null, "invalid welcome published");
        }
        await using (Peer peer = await Peer.Start(key, _ => Task.CompletedTask))
        {
            ClientConfig wrong = Config(peer, key) with { GatewayPubKey = Keys.GenerateKeypair().Public };
            await Refuses(async () => await Client.ConnectAsync(wrong, handler));
            await Refuses(async () => await peer.Finished);
            Check(handler.Welcome is null, "wrong pin published welcome");
        }
        await using (Peer peer = await Peer.Start(key, _ => Task.CompletedTask, silentHandshake: true))
        {
            ClientConfig timed = Config(peer, key) with { HandshakeTimeout = TimeSpan.FromMilliseconds(150) };
            await Refuses(async () => await Client.ConnectAsync(timed, handler).WaitAsync(TimeSpan.FromSeconds(3)));
            Check(handler.Welcome is null, "timed-out handshake published welcome");
        }
    }

    private static async Task Lifetime()
    {
        Keypair key = Keys.GenerateKeypair();
        TaskCompletionSource received = Signal(), releasePeer = Signal();
        RecordingHandler handler = new();
        await using Peer peer = await Peer.Start(key, async session =>
        {
            await session.Welcome();
            _ = await session.Read();
            received.SetResult();
            await releasePeer.Task.WaitAsync(session.Cancellation);
        });
        using CancellationTokenSource lifetime = new();
        await using Client client = await Client.ConnectAsync(Config(peer, key), handler, lifetime.Token);
        try
        {
            Task<JsonElement> request = client.RequestAsync("held", new { });
            await received.Task;
            lifetime.Cancel();
            await Refuses(async () => await request.WaitAsync(TimeSpan.FromSeconds(5)));
            await client.Done().WaitAsync(TimeSpan.FromSeconds(5));
            Check(handler.Events.Last() == "disconnected", "lifetime cancellation left connection live");
        }
        finally
        {
            releasePeer.TrySetResult();
        }
        await peer.Finished;
    }
}

internal sealed class RecordingHandler : NopHandler
{
    public readonly ConcurrentQueue<string> Events = new();
    public readonly ConcurrentQueue<Exception> Errors = new();
    public readonly Channel<Envelope> Deliveries = Channel.CreateUnbounded<Envelope>();
    public ManualResetEventSlim? Hold;
    public bool ThrowDelivery;
    public bool ThrowDisconnect;
    public WelcomePayload? Welcome;
    public override void OnConnected(WelcomePayload welcome) { Welcome = welcome; Events.Enqueue("connected"); }
    public override void OnDeliver(Envelope envelope)
    {
        Events.Enqueue("deliver");
        Deliveries.Writer.TryWrite(envelope);
        if (Hold is { } hold && !hold.Wait(TimeSpan.FromSeconds(10))) throw new TimeoutException("Held delivery fixture timed out.");
        if (ThrowDelivery) throw new InvalidOperationException("Application refused delivery.");
    }
    public override void OnDisconnected(string reason)
    {
        Events.Enqueue("disconnected");
        if (ThrowDisconnect) throw new InvalidOperationException("Application disconnect callback failed.");
    }
    public override void OnError(Exception error) => Errors.Enqueue(error);
}
