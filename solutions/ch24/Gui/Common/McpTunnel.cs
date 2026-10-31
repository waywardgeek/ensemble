using Ensemble;
using System.Text.Json;
using System.Threading.Channels;

namespace EnsembleGui;

// Socket-scoped I/O boundary, not an Agent relationship or transcript. The
// host owns the shared outbound writer; this resource owns only MCP's inbound
// queue and lifetime. The core client sees ordinary JSON-RPC messages.
public sealed partial class McpTunnel : IMcpTransport
{
    private readonly ChannelWriter<string> output;
    private readonly string agentID;
    private readonly Channel<JsonElement> input;
    private readonly CancellationTokenSource closed = new();
    private int disposed;

    public partial McpTunnel(ChannelWriter<string> output, string agentID);
    // Host routes only this socket's matching-agent envelopes here.
    public partial void Accept(JsonElement message);
    public partial Task SendAsync(JsonElement message, CancellationToken cancellation = default);
    public partial Task<JsonElement?> ReceiveAsync(CancellationToken cancellation = default);
    public partial ValueTask DisposeAsync();
}
