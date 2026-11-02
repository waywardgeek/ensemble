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
    // Borrow the actual requesting Agent. The socket scope still owns this
    // transport; it does not own or reconstruct the Agent's lifetime.
    private Agent? captureTarget;
    private Agent? captureSelection;
    private readonly object captureGate = new();
    private long? captureID;
    private bool captureReplyStarted;


    public partial McpTunnel(ChannelWriter<string> output, string agentID);
    public partial McpTunnel(ChannelWriter<string> output, string agentID, Agent captureTarget);
    public partial bool IsCaptureReply(JsonElement message);
    public partial void SetCaptureSelection(Agent selected);
    private partial bool AdmitCaptureRequest(JsonElement message);
    private partial bool AcceptCaptureReply(JsonElement message);
    private partial void RequireCaptureCapability();
    private partial JsonElement PrepareCaptureReply(long id, JsonElement message);
    private static partial JsonElement CaptureError(long id, string message);
    private static partial bool HasImageResult(JsonElement message);
    private static partial long? ReplyID(JsonElement message);

    // Host routes only this socket's matching-agent envelopes here.
    public partial void Accept(JsonElement message);
    public partial Task SendAsync(JsonElement message, CancellationToken cancellation = default);
    public partial Task<JsonElement?> ReceiveAsync(CancellationToken cancellation = default);
    public partial ValueTask DisposeAsync();
}
