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
            SingleWriter = true,
            FullMode = BoundedChannelFullMode.Wait
        });
    }

    public partial void Accept(JsonElement message)
    {
        if (message.ValueKind != JsonValueKind.Object)
        {
            throw new InvalidDataException("MCP data must be a JSON-RPC object.");
        }
        if (Volatile.Read(ref disposed) != 0 || !input.Writer.TryWrite(message.Clone()))
        {
            throw new IOException("MCP browser channel is closed or full.");
        }
    }

    public partial async Task SendAsync(JsonElement message, CancellationToken cancellation)
    {
        ObjectDisposedException.ThrowIf(Volatile.Read(ref disposed) != 0, this);
        using var wait = CancellationTokenSource.CreateLinkedTokenSource(cancellation, closed.Token);
        string envelope = JsonSerializer.Serialize(new { type = "mcp", agent_id = agentID, data = message });
        try
        {
            // The host's single socket writer serializes RPC with display
            // frames. Backpressure never silently drops a reliable RPC reply.
            await output.WriteAsync(envelope, wait.Token);
        }
        catch (ChannelClosedException)
        {
            throw new IOException("GUI output closed during MCP send.");
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
            closed.Cancel();
            input.Writer.TryComplete();
            closed.Dispose();
            // Do not complete the shared output queue or close the socket:
            // unloading browser tools must not destroy the conversation UI.
        }
        return ValueTask.CompletedTask;
    }
}
