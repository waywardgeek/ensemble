using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using Ensemble.Gateway;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.Logging;

// A small working protocol peer, not a mocked Client. It uses its own JSON codec
// and the real public Noise API. Each test supplies a finite server conversation.
internal sealed class Peer : IAsyncDisposable
{
    private readonly WebApplication app;
    private readonly CancellationTokenSource lifetime = new(TimeSpan.FromSeconds(15));
    private readonly TaskCompletionSource finished = new(TaskCreationOptions.RunContinuationsAsynchronously);
    private WebSocket? socket;
    public readonly Keypair Key = Keys.GenerateKeypair();
    public string Url = "";
    public Task Finished => finished.Task;

    private Peer()
    {
        WebApplicationBuilder builder = WebApplication.CreateSlimBuilder(new WebApplicationOptions
        {
            Args = [], ApplicationName = typeof(Peer).Assembly.FullName
        });
        builder.Logging.ClearProviders();
        builder.WebHost.UseUrls("http://127.0.0.1:0");
        app = builder.Build();
        app.UseWebSockets();
    }

    public static async Task<Peer> Start(Keypair clientKey, Func<Session, Task> script,
        bool silentHandshake = false)
    {
        Peer peer = new();
        peer.app.Map("/v1/ws", async context =>
        {
            try
            {
                using WebSocket socket = await context.WebSockets.AcceptWebSocketAsync();
                peer.socket = socket;
                using CancellationTokenRegistration abort = peer.lifetime.Token.Register(socket.Abort);
                if (silentHandshake)
                {
                    await Task.Delay(Timeout.Infinite, peer.lifetime.Token);
                    return;
                }
                string id = context.Request.Query["agent_id"].ToString();
                using NoiseKK noise = new(NoiseRole.Responder, peer.Key, clientKey.Public.AsSpan(),
                    Encoding.UTF8.GetBytes("ensemble-gateway-v1:" + id));
                noise.ReadHandshakeMessage(await Session.ReadRaw(socket, peer.lifetime.Token));
                await socket.SendAsync(noise.WriteHandshakeMessage([]).AsMemory(), WebSocketMessageType.Binary,
                    true, peer.lifetime.Token);
                Session session = new(socket, noise, id, peer.lifetime.Token);
                session.Hello = await session.Read();
                await script(session);
                peer.finished.TrySetResult();
            }
            catch (Exception error)
            {
                if (peer.lifetime.IsCancellationRequested)
                {
                    peer.finished.TrySetResult();
                }
                else
                {
                    peer.finished.TrySetException(error);
                }
            }
        });
        await peer.app.StartAsync();
        peer.Url = peer.app.Urls.Single().Replace("http:", "ws:") + "/v1/ws";
        return peer;
    }

    public async ValueTask DisposeAsync()
    {
        lifetime.Cancel();
        socket?.Abort();
        await app.StopAsync();
        await app.DisposeAsync();
        lifetime.Dispose();
    }
}

internal sealed class Session(WebSocket socket, NoiseKK noise, string id, CancellationToken cancellation)
{
    public JsonElement Hello;
    public CancellationToken Cancellation => cancellation;
    public Task Welcome(string? agent = null, int depth = 0) => Send(new
    {
        type = "welcome", welcome = new { agent_id = agent ?? id, queue_depth = depth, server_version = "1" }
    });
    public Task Send(object frame) => SendText(JsonSerializer.Serialize(frame));
    public async Task SendText(string json, bool fragment = false)
    {
        byte[] encrypted = noise.Encrypt(Encoding.UTF8.GetBytes(json), []);
        if (fragment)
        {
            await socket.SendAsync(encrypted.AsMemory(0, 7), WebSocketMessageType.Binary, false, cancellation);
            await socket.SendAsync(encrypted.AsMemory(7), WebSocketMessageType.Binary, true, cancellation);
        }
        else
        {
            await socket.SendAsync(encrypted.AsMemory(), WebSocketMessageType.Binary, true, cancellation);
        }
    }
    public async Task<JsonElement> Read()
    {
        byte[] bytes = noise.Decrypt(await ReadRaw(socket, cancellation), []);
        using JsonDocument json = JsonDocument.Parse(bytes);
        return json.RootElement.Clone();
    }
    public Task Reply(JsonElement request, object result) => Send(new
    {
        type = "mcp_response", rpc = new { jsonrpc = "2.0", id = request.GetProperty("rpc").GetProperty("id"), result }
    });
    public async Task Malformed(string kind)
    {
        if (kind == "duplicate")
        {
            await SendText("{\"type\":\"pong\",\"type\":\"deliver\"}");
        }
        else if (kind == "missing-result")
        {
            await SendText("{\"type\":\"mcp_response\",\"rpc\":{\"jsonrpc\":\"2.0\",\"id\":1}}");
        }
        else if (kind == "ambiguous-result")
        {
            await SendText("{\"type\":\"mcp_response\",\"rpc\":{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":null,\"error\":null}}");
        }
        else if (kind == "tampered")
        {
            byte[] bytes = noise.Encrypt("{\"type\":\"pong\"}"u8, []);
            bytes[^1] ^= 1;
            await socket.SendAsync(bytes.AsMemory(), WebSocketMessageType.Binary, true, cancellation);
        }
        else
        {
            byte[] bytes = kind == "oversized" ? new byte[65536] : "not binary"u8.ToArray();
            await socket.SendAsync(bytes.AsMemory(), kind == "oversized" ? WebSocketMessageType.Binary : WebSocketMessageType.Text,
                true, cancellation);
        }
    }
    public static async Task<byte[]> ReadRaw(WebSocket socket, CancellationToken cancellation)
    {
        using MemoryStream stream = new();
        byte[] buffer = new byte[2048];
        while (true)
        {
            ValueWebSocketReceiveResult part = await socket.ReceiveAsync(buffer.AsMemory(), cancellation);
            if (part.MessageType == WebSocketMessageType.Close)
            {
                throw new EndOfStreamException();
            }
            if (part.MessageType != WebSocketMessageType.Binary)
            {
                throw new InvalidDataException("Fixture expected binary message.");
            }
            stream.Write(buffer, 0, part.Count);
            if (stream.Length > 65535)
            {
                throw new InvalidDataException("Fixture frame too large.");
            }
            if (part.EndOfMessage)
            {
                return stream.ToArray();
            }
        }
    }
}
