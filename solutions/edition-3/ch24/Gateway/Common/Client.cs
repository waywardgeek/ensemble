using System.Collections.Immutable;
using System.Net.WebSockets;
using System.Text.Json;

namespace Ensemble.Gateway;

/// <summary>Immutable connection inputs. Zero timeout selects ten seconds.
/// Connect's cancellation token owns the entire connection, not just startup.</summary>
public sealed record ClientConfig(string GatewayURL, string AgentID, Keypair StaticKey,
    ImmutableArray<byte> GatewayPubKey, ulong LastAckedSeq = 0, TimeSpan HandshakeTimeout = default);

/// <summary>Borrowed application boundary, not a second connection owner.
/// Callbacks are sequential. Returning from OnDeliver accepts that delivery;
/// throwing refuses it and closes the connection. Do not synchronously wait for
/// client RPCs or joined disposal in callbacks: they need this receive loop.</summary>
public interface IHandler
{
    void OnConnected(WelcomePayload welcome);
    void OnDeliver(Envelope envelope);
    void OnDisconnected(string reason);
    void OnError(Exception error);
}

/// <summary>Optional callback base, not containment inheritance.</summary>
public partial class NopHandler : IHandler
{
    public virtual partial void OnConnected(WelcomePayload welcome);
    public virtual partial void OnDeliver(Envelope envelope);
    public virtual partial void OnDisconnected(string reason);
    public virtual partial void OnError(Exception error);
}

/// <summary>Standalone root owning one authenticated connection and its Noise
/// child. No Ensemble actor, retry scheduler or durable receipt store is hidden
/// here. Close interrupts I/O; Done/DisposeAsync join the callbacks and cleanup.</summary>
public sealed partial class Client : IAsyncDisposable
{
    private readonly object gate = new();
    private readonly SemaphoreSlim writeGate = new(1, 1);
    private readonly ClientWebSocket socket = new();
    private readonly CancellationTokenSource lifetime;
    private readonly CancellationTokenRegistration abortRegistration;
    private readonly ClientConfig config;
    private readonly IHandler handler;
    private readonly TaskCompletionSource done = new(TaskCreationOptions.RunContinuationsAsynchronously);
    private readonly Dictionary<ulong, TaskCompletionSource<RpcMessage>> pending = new();
    private NoiseKK? noise;
    private ulong nextID;
    private ulong lastAck;
    private bool closed;

    private partial Client(ClientConfig config, IHandler handler, CancellationToken cancellation);
    public static partial Task<Client> ConnectAsync(ClientConfig config, IHandler handler,
        CancellationToken cancellation = default);
    public partial Task SendAsync(OutboundMessage message, CancellationToken cancellation = default);
    public partial Task RespondAsync(Envelope prompt, string content, CancellationToken cancellation = default);
    public partial Task DeclareContactableAsync(IEnumerable<string> groups, CancellationToken cancellation = default);
    public partial Task RegisterChildAsync(ChildRegistration child, CancellationToken cancellation = default);
    public partial Task<JsonElement> RequestAsync(string method, object? parameters,
        CancellationToken cancellation = default);
    public partial Task<JsonElement> CallToolAsync(string name, object? arguments,
        CancellationToken cancellation = default);
    public partial ulong LastAckedSeq();
    public partial void Close();
    public partial Task Done();
    public partial ValueTask DisposeAsync();

    // Constructor registration and destruction maintain both relationship ends.
    // Noise is never disposed while holding gate, avoiding lock-order inversion.
    internal partial void AddNoise(NoiseKK child);
    internal partial void RemoveNoise(NoiseKK child);
    private partial Task HandshakeAsync(CancellationToken cancellation);
    private partial Task RunAsync();
    private partial Task WriteAsync(Frame frame, CancellationToken cancellation);
    private partial Task<Frame> ReadAsync(CancellationToken cancellation);
    private partial Task ReleaseAsync();
    private partial void ReceiveReply(RpcMessage reply);
    private partial Task DeliverAsync(Envelope envelope);
}
