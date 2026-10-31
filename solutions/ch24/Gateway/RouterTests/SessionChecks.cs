using System.Net.WebSockets;
using System.Security.Cryptography;
using System.Text.Json;
using System.Threading.Channels;
using Ensemble.Gateway;

internal static class SessionChecks
{
    public static IEnumerable<(string Name, Func<Task> Run)> Cases =>
    [
        ("real C# client/server delivery, RPC concurrency and reply provenance", PublicClients),
        ("offline replay, lost receipt, drained restart and continuing sequence", Replay),
        ("failed authentication/hello cannot replace an active session", Replacement),
        ("blocked delivery does not hold policy gate; future receipt refuses", Backpressure),
        ("encrypted frame validation, protocol errors and handshake deadline", BadFrames),
        ("shutdown joins pending handshake and active sessions before saving", Shutdown)
    ];

    private static void Check(bool condition, string message)
    {
        if (!condition) throw new Exception(message);
    }

    private static async Task Rejected<T>(RawGatewayPeer peer, RouterHost host) where T : Exception
    {
        Exception? error = await peer.Connection.End.WaitAsync(host.Deadline.Token);
        Check(error is T, "Expected " + typeof(T).Name + ", observed " + (error?.ToString() ?? "successful end"));
        Check(peer.Connection.Socket.Disposed, "Refused socket was not disposed.");
    }

    private static async Task Welcome(RawGatewayPeer peer, string id, int depth)
    {
        JsonElement frame = await peer.Read();
        Check(frame.GetProperty("type").GetString() == "welcome", "Missing welcome.");
        JsonElement welcome = frame.GetProperty("welcome");
        Check(welcome.GetProperty("agent_id").GetString() == id && welcome.GetProperty("queue_depth").GetInt32() == depth,
            "Welcome identity/depth mismatch.");
    }

    private sealed class Deliveries : NopHandler
    {
        public readonly Channel<Envelope> Received = Channel.CreateUnbounded<Envelope>();
        public override void OnDeliver(Envelope envelope) => Received.Writer.TryWrite(envelope);
    }

    private static async Task PublicClients()
    {
        await using RouterHost host = await RouterHost.Start();
        Deliveries aliceMessages = new();
        Deliveries bobMessages = new();
        await using Client alice = await Client.ConnectAsync(host.Config("alice"), aliceMessages, host.Deadline.Token);
        _ = await host.Next();
        await using Client bob = await Client.ConnectAsync(host.Config("bob"), bobMessages, host.Deadline.Token);
        _ = await host.Next();
        await alice.DeclareContactableAsync(["agents"], host.Deadline.Token);
        await bob.DeclareContactableAsync(["agents"], host.Deadline.Token);
        Task[] sends = Enumerable.Range(0, 16).Select(index => alice.SendAsync(
            new("bob", "message " + index, "conversation", "async"), host.Deadline.Token)).ToArray();
        await Task.WhenAll(sends);
        List<Envelope> received = [];
        for (int i = 0; i < sends.Length; i++) received.Add(await bobMessages.Received.Reader.ReadAsync(host.Deadline.Token));
        Check(received.Select(message => message.Seq).SequenceEqual(Enumerable.Range(1, 16).Select(number => (ulong)number)), "Wire delivery order.");
        Check(received.All(message => message.Source.UserID == "alice" && message.Source.ChannelID == "agents" &&
            message.Source.Trust == "trusted" && message.Content.StartsWith("[GATEWAY source=agents user=alice trust=trusted]\n\n", StringComparison.Ordinal)), "Accepted provenance.");
        Check(received.Select(message => message.Content).Distinct().Count() == 16, "Concurrent RPC lost or duplicated a message.");
        // Observe the client's post-ack receipt, then use a later RPC as a server
        // processing barrier. Polling observes state; elapsed time is not success.
        while (bob.LastAckedSeq() != 16) await Task.Delay(5, host.Deadline.Token);
        _ = await bob.RequestAsync("tools/list", new { }, host.Deadline.Token);
        Check(host.Router.Identity("bob").Queue().Snapshot().Ack == 16, "Server did not process client receipts.");
        await bob.RespondAsync(received[0], "answer", host.Deadline.Token);
        Envelope response = await aliceMessages.Received.Reader.ReadAsync(host.Deadline.Token);
        Check(response.Source.UserID == "bob" && response.ConversationID == "conversation" && response.ResponseMode == "", "Response provenance/conversation or recursive response mode.");
    }

    private static async Task Replay()
    {
        await using RouterHost host = await RouterHost.Start();
        host.Contact("bob");
        host.Router.Route(host.Router.Identity("alice"), new("bob", "one"));
        host.Router.Route(host.Router.Identity("alice"), new("bob", "two"));
        using (RawGatewayPeer first = await RawGatewayPeer.Open(host, "bob"))
        {
            await first.Handshake();
            await Welcome(first, "bob", 2);
            Check((await first.Read()).GetProperty("envelope").GetProperty("seq").GetUInt64() == 1, "First offline delivery.");
            Check((await first.Read()).GetProperty("envelope").GetProperty("seq").GetUInt64() == 2, "Second offline delivery.");
            first.Socket.Abort();
            _ = await first.Connection.End.WaitAsync(host.Deadline.Token);
        }
        Check(host.Router.Identity("bob").Queue().Snapshot().Ack == 0, "Sending alone acknowledged a delivery.");
        using (RawGatewayPeer resumed = await RawGatewayPeer.Open(host, "bob"))
        {
            await resumed.Handshake(receipt: 1);
            await Welcome(resumed, "bob", 1);
            Check((await resumed.Read()).GetProperty("envelope").GetProperty("seq").GetUInt64() == 2, "Lost receipt/replay negotiation.");
            await resumed.Send(new { type = "ack", last_acked_seq = 2 });
            await resumed.Ping();
            Check(host.Router.Identity("bob").Queue().Snapshot().Ack == 2, "Independent JSON receipt spelling.");
        }
        await host.DisposeAsync();
        await using RouterHost restart = await RouterHost.Start(host.Key, host.Alice, host.Bob, host.Home);
        Check(!host.Router.ContainsIdentity("bob"), "Disposed router still admits connections.");
        restart.Router.Route(restart.Router.Identity("alice"), new("bob", "three"));
        using RawGatewayPeer last = await RawGatewayPeer.Open(restart, "bob");
        await last.Handshake(receipt: 2);
        await Welcome(last, "bob", 1);
        Check((await last.Read()).GetProperty("envelope").GetProperty("seq").GetUInt64() == 3, "Drained sequence reset across restart.");
        await last.Send(new { type = "ack", last_acked_seq = 3 });
        await last.Ping();
        await restart.Router.DisposeAsync();
        using JsonDocument disk = JsonDocument.Parse(File.ReadAllBytes(Path.Combine(host.Home, "queues", "bob.json")));
        Check(disk.RootElement.GetProperty("next").GetUInt64() == 3 && disk.RootElement.GetProperty("ack").GetUInt64() == 3,
            "Joined final save did not retain receipt three.");
    }

    private static async Task Replacement()
    {
        await using RouterHost host = await RouterHost.Start();
        using RawGatewayPeer current = await RawGatewayPeer.Open(host, "alice");
        await current.Handshake();
        await Welcome(current, "alice", 0);
        using (RawGatewayPeer wrongKey = await RawGatewayPeer.Open(host, "alice", Keys.GenerateKeypair()))
        {
            await wrongKey.Socket.SendAsync(wrongKey.Noise.WriteHandshakeMessage([]).AsMemory(), WebSocketMessageType.Binary, true, host.Deadline.Token);
            await Rejected<CryptographicException>(wrongKey, host);
        }
        await current.Ping();
        using (RawGatewayPeer wrongHello = await RawGatewayPeer.Open(host, "alice"))
        {
            await wrongHello.Handshake(helloID: "bob");
            await Rejected<InvalidDataException>(wrongHello, host);
        }
        await current.Ping();
        using (RawGatewayPeer future = await RawGatewayPeer.Open(host, "alice"))
        {
            await future.Handshake(receipt: 1);
            await Rejected<InvalidDataException>(future, host);
        }
        await current.Ping();
        using RawGatewayPeer replacement = await RawGatewayPeer.Open(host, "alice");
        await replacement.Handshake();
        await Welcome(replacement, "alice", 0);
        _ = await current.Connection.End.WaitAsync(host.Deadline.Token);
        Check(current.Connection.Socket.Disposed && host.Router.Identity("alice").Sessions().Length == 1, "Replacement did not join/detach old session.");
        await replacement.Ping();
    }

    private static async Task Backpressure()
    {
        await using RouterHost host = await RouterHost.Start();
        host.Contact("bob");
        using RawGatewayPeer bob = await RawGatewayPeer.Open(host, "bob");
        await bob.Handshake();
        await Welcome(bob, "bob", 0);
        bob.Connection.Socket.HoldWrites();
        host.Router.Route(host.Router.Identity("alice"), new("bob", "held first delivery"));
        await bob.Connection.Socket.Blocked.WaitAsync(host.Deadline.Token);
        host.Router.Route(host.Router.Identity("alice"), new("bob", "not yet sent"));
        await using Client alice = await Client.ConnectAsync(host.Config("alice"), new NopHandler(), host.Deadline.Token);
        _ = await host.Next();
        _ = await alice.RequestAsync("tools/list", new { }, host.Deadline.Token);
        await Task.Run(host.Router.Save).WaitAsync(host.Deadline.Token);
        // next is two, but this connection has only begun sending sequence one.
        await bob.Send(new { type = "ack", last_acked_seq = 2 });
        await Rejected<InvalidDataException>(bob, host);
        QueueState queue = host.Router.Identity("bob").Queue().Snapshot();
        Check(queue.Next == 2 && queue.Ack == 0 && queue.Messages!.Value.Length == 2, "Future receipt removed unsent data.");
    }

    private static async Task BadFrames()
    {
        await using RouterHost host = await RouterHost.Start();
        foreach (string kind in new[] { "duplicate", "text", "tamper", "oversize" })
        {
            using RawGatewayPeer peer = await RawGatewayPeer.Open(host, "bob");
            await peer.Handshake();
            await Welcome(peer, "bob", 0);
            if (kind == "duplicate") await peer.SendText("{\"type\":\"ack\",\"last_acked_seq\":0,\"last_acked_seq\":0}");
            else
            {
                byte[] bytes = kind == "oversize" ? new byte[65536] : kind == "text" ? "text"u8.ToArray() : peer.Noise.Encrypt("{\"type\":\"ping\"}"u8, []);
                if (kind == "tamper") bytes[^1] ^= 1;
                try { await peer.Socket.SendAsync(bytes.AsMemory(), kind == "text" ? WebSocketMessageType.Text : WebSocketMessageType.Binary, true, host.Deadline.Token); }
                catch (WebSocketException) { /* A rejecting server may abort an in-flight oversized write. */ }
            }
            if (kind == "tamper") await Rejected<CryptographicException>(peer, host);
            else if (kind == "duplicate") await Rejected<JsonException>(peer, host);
            else await Rejected<InvalidDataException>(peer, host);
        }
        using RawGatewayPeer valid = await RawGatewayPeer.Open(host, "bob");
        await valid.Handshake();
        await Welcome(valid, "bob", 0);
        await valid.Send(new { type = "unknown" });
        Check((await valid.Read()).GetProperty("type").GetString() == "error", "Unknown frame did not report an error.");
        await valid.Send(new { type = "response", message = new { to = "unknown", content = "not deliverable" } });
        Check((await valid.Read()).GetProperty("type").GetString() == "error", "Denied response did not report an error.");
        await valid.Ping();
        using RawGatewayPeer silent = await RawGatewayPeer.Open(host, "alice");
        await Rejected<OperationCanceledException>(silent, host);
        Check(!host.Deadline.IsCancellationRequested, "Fixture timeout was mistaken for a product handshake deadline.");
        await valid.Ping();
    }

    private static async Task Shutdown()
    {
        await using RouterHost host = await RouterHost.Start();
        using RawGatewayPeer pending = await RawGatewayPeer.Open(host, "alice");
        Ensemble.Gateway.Session session = host.Router.Identity("alice").Sessions().Single();
        Check(session.HasNoise(), "Pending handshake did not own Noise.");
        await using Client active = await Client.ConnectAsync(host.Config("bob"), new NopHandler(), host.Deadline.Token);
        RouterHost.Connection activeConnection = await host.Next();
        Task first = host.Router.DisposeAsync().AsTask();
        Task second = host.Router.DisposeAsync().AsTask();
        Check(ReferenceEquals(first, second), "Concurrent shutdown did not share a joined result.");
        await first.WaitAsync(host.Deadline.Token);
        _ = await pending.Connection.End.WaitAsync(host.Deadline.Token);
        _ = await activeConnection.End.WaitAsync(host.Deadline.Token);
        await active.Done().WaitAsync(host.Deadline.Token);
        Check(session.parent is null && !session.HasNoise() && pending.Connection.Socket.Disposed && activeConnection.Socket.Disposed,
            "Shutdown retained session/Noise/socket ownership.");
        Check(File.Exists(Path.Combine(host.Home, "state.json")), "Shutdown did not save state.");

        RouterHost faulty = await RouterHost.Start();
        bool injected = false;
        try
        {
            using RawGatewayPeer peer = await RawGatewayPeer.Open(faulty, "bob");
            await peer.Handshake();
            await Welcome(peer, "bob", 0);
            Ensemble.Gateway.Session child = faulty.Router.Identity("bob").Sessions().Single();
            peer.Connection.Socket.FailDispose = injected = true;
            try
            {
                await faulty.Router.DisposeAsync().AsTask().WaitAsync(faulty.Deadline.Token);
                throw new Exception("Shutdown hid its transport cleanup failure.");
            }
            catch (AggregateException error)
            {
                Check(error.Flatten().InnerExceptions.Any(item => item is IOException), "Wrong cleanup failure.");
            }
            Check(child.parent is null && !child.HasNoise() && peer.Connection.Socket.Disposed, "Cleanup failure retained ownership.");
            Check(File.Exists(Path.Combine(faulty.Home, "state.json")), "Cleanup failure skipped final save.");
            Check(await peer.Connection.End.WaitAsync(faulty.Deadline.Token) is AggregateException, "HTTP caller missed cleanup failure.");
        }
        finally
        {
            try { await faulty.DisposeAsync(); }
            catch (AggregateException error) when (injected && error.Flatten().InnerExceptions.Any(item => item is IOException))
            { /* The same failed root join remains visible to this observer too. */ }
        }
    }
}
