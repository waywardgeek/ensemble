using System.Text.Json;
using System.Threading.Channels;

namespace EnsembleGui;

public sealed partial class McpTunnel
{
    public partial McpTunnel(ChannelWriter<string> output, string agentID)
    {
        this.output = output;
        this.agentID = agentID;
        input = Channel.CreateBounded<JsonElement>(new BoundedChannelOptions(256)
        {
            SingleReader = true,
            // Socket replies and local capture refusals are separate producers.
            SingleWriter = false,
            FullMode = BoundedChannelFullMode.Wait
        });
    }

    public partial void Accept(JsonElement message)
    {
        if (message.ValueKind != JsonValueKind.Object)
        {
            throw new InvalidDataException("MCP data must be a JSON-RPC object.");
        }
        if (AcceptCaptureReply(message)) return;
        if (Volatile.Read(ref disposed) != 0 || !input.Writer.TryWrite(message.Clone()))
        {
            throw new IOException("MCP browser channel is closed or full.");
        }
    }

    public partial async Task SendAsync(JsonElement message, CancellationToken cancellation)
    {
        ObjectDisposedException.ThrowIf(Volatile.Read(ref disposed) != 0, this);
        cancellation.ThrowIfCancellationRequested();
        if (!AdmitCaptureRequest(message)) return;
        using var wait = CancellationTokenSource.CreateLinkedTokenSource(cancellation, closed.Token);
        string envelope = JsonSerializer.Serialize(new { type = "mcp", agent_id = agentID, data = message });
        bool written = false;
        try
        {
            // The host's single socket writer serializes RPC with display
            // frames. Backpressure never silently drops a reliable RPC reply.
            await output.WriteAsync(envelope, wait.Token);
            written = true;
        }
        catch (ChannelClosedException)
        {
            throw new IOException("GUI output closed during MCP send.");
        }
        finally
        {
            // A request that never reached the browser cannot keep a capture
            // reservation. Other MCP requests have distinct RPC identities.
            if (!written)
            {
                long? id = ReplyID(message);
                lock (captureGate)
                    if (id is not null && captureID == id) captureID = null;
            }
        }
    }

    public partial async Task<JsonElement?> ReceiveAsync(CancellationToken cancellation)
    {
        try
        {
            return await input.Reader.ReadAsync(cancellation);
        }
        catch (ChannelClosedException)
        {
            return null;
        }
    }

    public partial ValueTask DisposeAsync()
    {
        if (Interlocked.Exchange(ref disposed, 1) == 0)
        {
            lock (captureGate) captureID = null;
            // The host's queue is unbounded while connected. A completed queue
            // means socket teardown already owns browser-side cancellation.
            output.TryWrite(JsonSerializer.Serialize(new
            {
                type = "mcp", agent_id = agentID,
                data = new { jsonrpc = "2.0", method = "notifications/ensemble/browser_closed" }
            }));
            closed.Cancel();
            input.Writer.TryComplete();
            closed.Dispose();
            // Do not complete the shared output queue or close the socket:
            // unloading browser tools must not destroy the conversation UI.
        }
        return ValueTask.CompletedTask;
    }
}
