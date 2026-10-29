using System.Collections.Immutable;
using System.Net.WebSockets;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace Ensemble.Gateway;

// Wire records are immutable values, not live relationships. The router stamps
// Source; content and display names remain data and never confer authority.
public sealed record Source(
    [property: JsonPropertyName("type")] string Type,
    [property: JsonPropertyName("channel_id")] string ChannelID,
    [property: JsonPropertyName("user_id")] string UserID,
    [property: JsonPropertyName("display_name")] string DisplayName,
    [property: JsonPropertyName("trust")] string Trust);

public sealed record Envelope(
    [property: JsonPropertyName("id")] string ID,
    [property: JsonPropertyName("seq")] ulong Seq,
    [property: JsonPropertyName("to_agent")] string ToAgent,
    [property: JsonPropertyName("content")] string Content,
    [property: JsonPropertyName("conversation_id")] string ConversationID,
    [property: JsonPropertyName("response_mode")] string ResponseMode,
    [property: JsonPropertyName("timestamp")] DateTimeOffset Timestamp,
    [property: JsonPropertyName("source")] Source Source);

public sealed record WelcomePayload(
    [property: JsonPropertyName("agent_id")] string AgentID,
    // Unlike optional Frame fields, the Go welcome always includes zero depth.
    [property: JsonPropertyName("queue_depth"), JsonIgnore(Condition = JsonIgnoreCondition.Never)] int QueueDepth,
    [property: JsonPropertyName("server_version")] string ServerVersion);

public sealed record OutboundMessage(
    [property: JsonPropertyName("to_agent")] string ToAgent,
    [property: JsonPropertyName("content")] string Content,
    [property: JsonPropertyName("conversation_id")] string ConversationID = "",
    [property: JsonPropertyName("response_mode")] string ResponseMode = "");

/// <summary>Value-like registration. Public bytes are encoded as base64 at the
/// protocol boundary, never as a JSON integer array or a private key.</summary>
public sealed record ChildRegistration(string AgentID, ImmutableArray<byte> PublicKey,
    ImmutableArray<string> Groups);

public static partial class Protocol
{
    public const string ResponseAsync = "async";
    /// <summary>Format only the router-authored header; the router adds the blank
    /// line and sender text separately. Formatting grants no routing authority.</summary>
    public static partial string ProvenanceHeader(Source source);
}

// Internal frame/RPC values match the existing Go JSON names. One authenticated
// WebSocket binary message carries exactly one frame; unknown fields are allowed
// for compatibility, but duplicate JSON names and malformed supplied values refuse.
internal sealed record RpcError(
    [property: JsonPropertyName("code")] int Code,
    [property: JsonPropertyName("message")] string Message);
internal sealed record RpcMessage(
    [property: JsonPropertyName("jsonrpc")] string Jsonrpc = "2.0",
    [property: JsonPropertyName("id")] JsonElement? ID = null,
    [property: JsonPropertyName("method")] string? Method = null,
    [property: JsonPropertyName("params")] JsonElement? Params = null,
    [property: JsonPropertyName("result")] JsonElement? Result = null,
    [property: JsonPropertyName("error")] RpcError? Error = null);
internal sealed record Frame(
    [property: JsonPropertyName("type")] string Type,
    [property: JsonPropertyName("version")] int Version = 0,
    [property: JsonPropertyName("agent_id")] string? AgentID = null,
    [property: JsonPropertyName("last_acked_seq")] ulong Seq = 0,
    [property: JsonPropertyName("welcome")] WelcomePayload? Welcome = null,
    [property: JsonPropertyName("envelope")] Envelope? Envelope = null,
    [property: JsonPropertyName("message")] OutboundMessage? Message = null,
    [property: JsonPropertyName("rpc")] RpcMessage? Rpc = null,
    [property: JsonPropertyName("error")] string? Error = null);

/// <summary>Pure wire codec and bounded WebSocket framing shared by client and
/// router. Authentication is performed by the owning connection before Decode.</summary>
internal static partial class Wire
{
    internal const int MaxCiphertext = 65535;
    internal const int MaxPlaintext = MaxCiphertext - 16;
    private static readonly JsonSerializerOptions Options = new()
    {
        DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingDefault,
        MaxDepth = 32
    };
    internal static partial byte[] Encode(Frame frame);
    internal static partial Frame Decode(ReadOnlyMemory<byte> bytes);
    // Go omits empty optional strings in both wire and saved envelopes.
    internal static partial Envelope NormalizeEnvelope(Envelope value);
    internal static partial void UniqueNames(JsonElement value);
    internal static partial Task<byte[]> ReadBinaryAsync(WebSocket socket, CancellationToken cancellation);
}
