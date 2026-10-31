using System.Net.WebSockets;
using System.Text.Json;

namespace Ensemble;

public sealed partial class McpWebSocketTransport
{
    public partial McpWebSocketTransport(WebSocket socket)
    {
        ArgumentNullException.ThrowIfNull(socket);
        if (socket.State != WebSocketState.Open)
        {
            throw new ArgumentException("MCP requires a connected WebSocket", nameof(socket));
        }
        this.socket = socket;
    }

    public partial async Task SendAsync(JsonElement message, CancellationToken cancellation)
    {
        byte[] bytes = JsonSerializer.SerializeToUtf8Bytes(message);
        await writes.WaitAsync(cancellation);
        try
        {
            ObjectDisposedException.ThrowIf(Volatile.Read(ref disposed) != 0, this);
            await socket.SendAsync(bytes.AsMemory(), WebSocketMessageType.Text, true, cancellation);
        }
        catch (WebSocketException)
        {
            throw new IOException("MCP WebSocket send failed.");
        }
        finally
        {
            writes.Release();
        }
    }

    public partial async Task<JsonElement?> ReceiveAsync(CancellationToken cancellation)
    {
        ObjectDisposedException.ThrowIf(Volatile.Read(ref disposed) != 0, this);
        // A WebSocket message, not an individual frame, is one RPC envelope.
        // Keep the same inbound byte bound as the line-framed transport.
        using var message = new MemoryStream();
        byte[] buffer = new byte[8192];
        try
        {
            while (true)
            {
                var received = await socket.ReceiveAsync(buffer.AsMemory(), cancellation);
                if (received.MessageType == WebSocketMessageType.Close)
                {
                    if (message.Length != 0)
                    {
                        throw new InvalidDataException("MCP WebSocket closed during a message.");
                    }
                    return null;
                }
                if (received.MessageType != WebSocketMessageType.Text)
                {
                    throw new InvalidDataException("MCP WebSocket requires text messages.");
                }
                if (message.Length + received.Count > 16 * 1024 * 1024)
                {
                    throw new InvalidDataException("MCP message exceeds the transport bound.");
                }
                message.Write(buffer, 0, received.Count);
                if (received.EndOfMessage)
                {
                    using var document = JsonDocument.Parse(message.GetBuffer().AsMemory(0, (int)message.Length));
                    return document.RootElement.Clone();
                }
            }
        }
        catch (WebSocketException)
        {
            throw new IOException("MCP WebSocket receive failed.");
        }
    }

    public partial ValueTask DisposeAsync()
    {
        if (Interlocked.Exchange(ref disposed, 1) == 0)
        {
            // Teardown is not a peer-controlled close handshake. Abort releases
            // pending reads/writes; the client owns joining its reader task.
            socket.Abort();
            socket.Dispose();
        }
        return ValueTask.CompletedTask;
    }
}
