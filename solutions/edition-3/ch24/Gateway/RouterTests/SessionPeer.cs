using System.Collections.Immutable;
using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using System.Threading.Channels;
using Ensemble.Gateway;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.Logging;

// This is a real loopback host for the real Router. The optional write hold is
// transport backpressure, not a mocked policy call or a manufactured receipt.
internal sealed class RouterHost : IAsyncDisposable
{
    private readonly WebApplication app;
    private bool disposed;
    public readonly CancellationTokenSource Deadline = new(TimeSpan.FromSeconds(30));
    public readonly Channel<Connection> Admitted = Channel.CreateUnbounded<Connection>();
    public readonly Keypair Key;
    public readonly Keypair Alice;
    public readonly Keypair Bob;
    public readonly string Home;
    public readonly Router Router;
    public string Url = "";
    public sealed record Connection(HeldSocket Socket, Task<Exception?> End);

    private RouterHost(Keypair? key, Keypair? alice, Keypair? bob, string? home)
    {
        Key = key ?? Keys.GenerateKeypair();
        Alice = alice ?? Keys.GenerateKeypair();
        Bob = bob ?? Keys.GenerateKeypair();
        Home = home ?? Path.Combine(Path.GetTempPath(), "gateway-sessions-" + Guid.NewGuid().ToString("N"));
        RouterConfig config = new(new("127.0.0.1:0", Home),
            new Dictionary<string, GroupDefinition> { ["agents"] = new(["send-message", "declare-contactable"], "trusted") }.ToImmutableDictionary(),
            new Dictionary<string, IdentityDefinition>
            {
                ["alice"] = new(Keys.EncodePublicKey(Alice.Public.AsSpan()), ["agents"]),
                ["bob"] = new(Keys.EncodePublicKey(Bob.Public.AsSpan()), ["agents"])
            }.ToImmutableDictionary());
        Router = new(config, Key);
        WebApplicationBuilder builder = WebApplication.CreateSlimBuilder(new WebApplicationOptions
        {
            Args = [], ApplicationName = typeof(RouterHost).Assembly.FullName
        });
        builder.Logging.ClearProviders();
        builder.WebHost.UseUrls("http://127.0.0.1:0");
        app = builder.Build();
        app.UseWebSockets();
        app.Map("/v1/ws", async context =>
        {
            string id = context.Request.Query["agent_id"].ToString();
            if (!Router.ContainsIdentity(id)) { context.Response.StatusCode = 403; return; }
            if (!context.WebSockets.IsWebSocketRequest) { context.Response.StatusCode = 400; return; }
            HeldSocket socket = new(await context.WebSockets.AcceptWebSocketAsync());
            Task<Exception?> ending = Observe(Router.AcceptAsync(id, socket, Deadline.Token));
            Admitted.Writer.TryWrite(new(socket, ending));
            await ending;
        });
    }

    public static async Task<RouterHost> Start(Keypair? key = null, Keypair? alice = null, Keypair? bob = null, string? home = null)
    {
        RouterHost host = new(key, alice, bob, home);
        try
        {
            await host.app.StartAsync(host.Deadline.Token);
            host.Url = host.app.Urls.Single().Replace("http:", "ws:") + "/v1/ws";
            return host;
        }
        catch { await host.DisposeAsync(); throw; }
    }

    private static async Task<Exception?> Observe(Task work)
    {
        try { await work; return null; }
        catch (Exception error) { return error; }
    }

    public async Task<Connection> Next() => await Admitted.Reader.ReadAsync(Deadline.Token);
    public ClientConfig Config(string id) => new(Url, id, id == "alice" ? Alice : Bob, Key.Public);
    public void Contact(string id) => Router.Identity(id).ReplaceGroups(GroupLink.Contact, [Router.Group("agents")]);

    public async ValueTask DisposeAsync()
    {
        if (disposed) return;
        disposed = true;
        try { await Router.DisposeAsync(); }
        finally
        {
            Deadline.Cancel();
            await app.StopAsync();
            await app.DisposeAsync();
            Deadline.Dispose();
        }
    }
}

internal sealed class HeldSocket(WebSocket inner) : WebSocket
{
    private readonly object gate = new();
    private TaskCompletionSource? resume;
    private readonly TaskCompletionSource blocked = new(TaskCreationOptions.RunContinuationsAsynchronously);
    public bool Disposed;
    public bool FailDispose;
    public Task Blocked => blocked.Task;
    public void HoldWrites() { lock (gate) { resume = new(TaskCreationOptions.RunContinuationsAsynchronously); } }
    public override WebSocketCloseStatus? CloseStatus => inner.CloseStatus;
    public override string? CloseStatusDescription => inner.CloseStatusDescription;
    public override string? SubProtocol => inner.SubProtocol;
    public override WebSocketState State => inner.State;
    public override void Abort() => inner.Abort();
    public override void Dispose()
    {
        Disposed = true;
        inner.Dispose();
        if (FailDispose) throw new IOException("Owned fixture disposal failure.");
    }
    public override Task CloseAsync(WebSocketCloseStatus status, string? description, CancellationToken cancellation) => inner.CloseAsync(status, description, cancellation);
    public override Task CloseOutputAsync(WebSocketCloseStatus status, string? description, CancellationToken cancellation) => inner.CloseOutputAsync(status, description, cancellation);
    public override Task<WebSocketReceiveResult> ReceiveAsync(ArraySegment<byte> bytes, CancellationToken cancellation) => inner.ReceiveAsync(bytes, cancellation);
    public override async Task SendAsync(ArraySegment<byte> bytes, WebSocketMessageType type, bool end, CancellationToken cancellation)
    {
        Task? release;
        lock (gate) { release = resume?.Task; }
        if (release is not null)
        {
            blocked.TrySetResult();
            await release.WaitAsync(cancellation);
        }
        await inner.SendAsync(bytes, type, end, cancellation);
    }
}

// Independent JSON names intentionally do not use Frame or Wire. The public
// Noise wrapper supplies actual cryptography; receipt spelling is asserted here.
internal sealed class RawGatewayPeer : IDisposable
{
    public readonly ClientWebSocket Socket = new();
    public readonly NoiseKK Noise;
    private readonly CancellationToken cancellation;
    private readonly string id;
    public RouterHost.Connection Connection = null!;

    private RawGatewayPeer(RouterHost host, string id, Keypair key)
    {
        this.id = id;
        cancellation = host.Deadline.Token;
        Noise = new(NoiseRole.Initiator, key, host.Key.Public.AsSpan(), Encoding.UTF8.GetBytes("ensemble-gateway-v1:" + id));
    }

    public static async Task<RawGatewayPeer> Open(RouterHost host, string id, Keypair? key = null)
    {
        RawGatewayPeer peer = new(host, id, key ?? (id == "alice" ? host.Alice : host.Bob));
        try
        {
            await peer.Socket.ConnectAsync(new Uri(host.Url + "?agent_id=" + id), host.Deadline.Token);
            peer.Connection = await host.Next();
            return peer;
        }
        catch { peer.Dispose(); throw; }
    }

    public async Task Handshake(string? helloID = null, ulong receipt = 0)
    {
        await Socket.SendAsync(Noise.WriteHandshakeMessage([]).AsMemory(), WebSocketMessageType.Binary, true, cancellation);
        Noise.ReadHandshakeMessage(await RawRead());
        await Send(new { type = "hello", version = 1, agent_id = helloID ?? id, last_acked_seq = receipt });
    }

    public Task Send(object value) => SendText(JsonSerializer.Serialize(value));
    public async Task SendText(string text, bool fragmented = false)
    {
        byte[] bytes = Noise.Encrypt(Encoding.UTF8.GetBytes(text), []);
        if (fragmented)
        {
            await Socket.SendAsync(bytes.AsMemory(0, 3), WebSocketMessageType.Binary, false, cancellation);
            await Socket.SendAsync(bytes.AsMemory(3), WebSocketMessageType.Binary, true, cancellation);
        }
        else await Socket.SendAsync(bytes.AsMemory(), WebSocketMessageType.Binary, true, cancellation);
    }
    public async Task<JsonElement> Read()
    {
        using JsonDocument document = JsonDocument.Parse(Noise.Decrypt(await RawRead(), []));
        return document.RootElement.Clone();
    }
    public async Task<byte[]> RawRead()
    {
        using MemoryStream output = new();
        byte[] buffer = new byte[4096];
        while (true)
        {
            ValueWebSocketReceiveResult part = await Socket.ReceiveAsync(buffer.AsMemory(), cancellation);
            if (part.MessageType == WebSocketMessageType.Close) throw new EndOfStreamException();
            if (part.MessageType != WebSocketMessageType.Binary) throw new InvalidDataException("Expected binary frame.");
            output.Write(buffer, 0, part.Count);
            if (output.Length > 65535) throw new InvalidDataException("Oversized frame.");
            if (part.EndOfMessage) return output.ToArray();
        }
    }
    public async Task Ping()
    {
        await SendText("{\"type\":\"ping\"}", fragmented: true);
        if ((await Read()).GetProperty("type").GetString() != "pong") throw new Exception("Missing pong barrier.");
    }
    public void Dispose() { Socket.Abort(); Socket.Dispose(); Noise.Dispose(); }
}
