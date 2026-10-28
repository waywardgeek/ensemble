using System.Net;
using System.Net.Sockets;
using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

static partial class McpTests
{
    // Real RFC 6455 framing over owned loopback sockets. HTTP upgrade is a
    // separate GUI-host responsibility and is not claimed by this fixture.
    static async Task<(McpWebSocketTransport Client, WebSocket Peer)> WebSocketPair()
    {
        using var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        // Transfer socket ownership to NetworkStream after connection; a
        // separate TcpClient wrapper must not dispose it underneath WebSocket.
        var client = new TcpClient();
        Task connecting = client.ConnectAsync(IPAddress.Loopback, ((IPEndPoint)listener.LocalEndpoint).Port);
        var peer = await listener.AcceptTcpClientAsync();
        await connecting;
        var clientSocket = WebSocket.CreateFromStream(new NetworkStream(client.Client, ownsSocket: true),
            isServer: false, subProtocol: null, keepAliveInterval: Timeout.InfiniteTimeSpan);
        var peerSocket = WebSocket.CreateFromStream(new NetworkStream(peer.Client, ownsSocket: true),
            isServer: true, subProtocol: null, keepAliveInterval: Timeout.InfiniteTimeSpan);
        return (new McpWebSocketTransport(clientSocket), peerSocket);
    }

    static async Task FragmentedReply(WebSocket peer, JsonElement request, object result)
    {
        byte[] bytes = JsonSerializer.SerializeToUtf8Bytes(new { jsonrpc = "2.0", id = request.GetProperty("id"), result });
        int middle = bytes.Length / 2;
        await peer.SendAsync(bytes.AsMemory(0, middle), WebSocketMessageType.Text, false, CancellationToken.None);
        await peer.SendAsync(bytes.AsMemory(middle), WebSocketMessageType.Text, true, CancellationToken.None);
    }

    public static async Task WebSocketBoundary()
    {
        var pair = await WebSocketPair();
        await using var peer = new McpWebSocketTransport(pair.Peer);
        var root = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"));
            var manager = new Mcp(agent);
            var client = new McpClient(manager, "web-peer", pair.Client);
            Task connecting = client.ConnectAsync();
            var initialize = await Read(peer);
            Check(initialize.GetProperty("method").GetString() == "initialize", "WebSocket initialize order");
            await FragmentedReply(pair.Peer, initialize, new { protocolVersion = "2024-11-05", capabilities = new { } });
            var initialized = await Read(peer);
            Check(initialized.GetProperty("method").GetString() == "notifications/initialized", "WebSocket notification order");
            var list = await Read(peer);
            await FragmentedReply(pair.Peer, list, new { tools = Array.Empty<object>() });
            await connecting.WaitAsync(TimeSpan.FromSeconds(8));
            Task<JsonElement> first = client.RequestAsync("first", J("{}"));
            Task<JsonElement> second = client.RequestAsync("second", J("{}"));
            var one = await Read(peer);
            var two = await Read(peer);
            await FragmentedReply(pair.Peer, two, new { value = "second" });
            await FragmentedReply(pair.Peer, one, new { value = "first" });
            Check((await first.WaitAsync(TimeSpan.FromSeconds(8))).GetProperty("value").GetString() == "first"
                && (await second.WaitAsync(TimeSpan.FromSeconds(8))).GetProperty("value").GetString() == "second",
                "Fragmented WebSocket replies lost correlation");
            Task held = client.RequestAsync("hold", J("{}"));
            await Read(peer);
            await pair.Peer.CloseOutputAsync(WebSocketCloseStatus.NormalClosure, "done", CancellationToken.None);
            await Fails(held);
            await client.DestroyAsync().WaitAsync(TimeSpan.FromSeconds(8));
            Check(manager.Clients().IsEmpty, "WebSocket client remained attached after close");
        }
        finally
        {
            root.Destroy();
        }

        var binaryPair = await WebSocketPair();
        await using var binaryClient = binaryPair.Client;
        using var binaryPeer = binaryPair.Peer;
        await binaryPeer.SendAsync(Encoding.UTF8.GetBytes("{}"), WebSocketMessageType.Binary, true, CancellationToken.None);
        await Fails(binaryClient.ReceiveAsync());

        var partialPair = await WebSocketPair();
        await using var partialClient = partialPair.Client;
        using var partialPeer = partialPair.Peer;
        Task partialRead = partialClient.ReceiveAsync();
        await partialPeer.SendAsync(Encoding.UTF8.GetBytes("{\"id\":"),
            WebSocketMessageType.Text, false, CancellationToken.None);
        await partialPeer.CloseOutputAsync(WebSocketCloseStatus.NormalClosure, "stopped", CancellationToken.None);
        await Fails(partialRead);

        var abortPair = await WebSocketPair();
        using var abortPeer = abortPair.Peer;
        Task receiving = abortPair.Client.ReceiveAsync();
        await abortPair.Client.DisposeAsync();
        await Fails(receiving);
        await abortPair.Client.DisposeAsync();
    }
}
