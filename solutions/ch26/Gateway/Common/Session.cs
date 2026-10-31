using System.Net.WebSockets;
using System.Threading.Channels;

namespace Ensemble.Gateway;

public sealed partial class Router
{
    // One shared close result joins concurrent disposers. Sessions remain owned
    // by identities until both their transport loops and cleanup have finished.
    private TaskCompletionSource? closing;
    /// <summary>HTTP preflight only, not authentication or a reservation. Accept
    /// rechecks admission after upgrade, including shutdown racing the upgrade.</summary>
    public partial bool ContainsIdentity(string id);
    /// <summary>Take ownership of an upgraded socket, authenticate its claimed
    /// identity, and serve until teardown is joined. Even refused sockets are
    /// disposed. The caller observes connection failures and owns HTTP hosting.</summary>
    public partial Task AcceptAsync(string id, WebSocket socket, CancellationToken cancellation = default);
    internal partial Keypair StaticKey();
    private partial Task FinishCloseAsync(Session[] sessions, TaskCompletionSource completion);
}

internal sealed partial class RouterIdentity
{
    // Pending handshakes are owned too. Current is a weak selection among these
    // children, not another session owner or a copied permission snapshot.
    private readonly HashSet<Session> sessions = [];
    private Session? current;
    internal partial void AddSession(Session child);
    internal partial void RemoveSession(Session child);
    internal partial Session[] Sessions();
    internal partial bool IsCurrent(Session child);
    internal partial (Session? Previous, WelcomePayload Welcome, ulong Receipt) Activate(Session child, ulong receipt);
    internal partial void Notify();
}

/// <summary>An identity-owned connection, including its unauthenticated handshake.
/// Session owns socket, cancellation, write serialization and Noise. Policy stays
/// with the Router; sent is only this connection's delivery frontier, not an ack.
/// A completed Done means both loops, resources and parent detachment are joined.</summary>
internal sealed partial class Session
{
    internal RouterIdentity? parent;
    private readonly WebSocket socket;
    private readonly CancellationTokenSource lifetime;
    private readonly CancellationTokenRegistration abortRegistration;
    private readonly SemaphoreSlim writeGate = new(1, 1);
    private readonly Channel<byte> wake = Channel.CreateBounded<byte>(new BoundedChannelOptions(1)
    {
        SingleReader = true,
        AllowSynchronousContinuations = false,
        FullMode = BoundedChannelFullMode.DropWrite
    });
    private readonly TaskCompletionSource done = new(TaskCreationOptions.RunContinuationsAsynchronously);
    private NoiseKK? noise;
    private Task? delivery;
    private ulong sent;

    internal partial Session(RouterIdentity parent, WebSocket socket, CancellationToken cancellation);
    internal partial RouterIdentity Parent();
    internal partial void AddNoise(NoiseKK child);
    internal partial void RemoveNoise(NoiseKK child);
    internal partial bool HasNoise();
    internal partial Task RunAsync();
    internal partial Task Done();
    internal partial void Notify();
    internal partial void Close();
    private partial void RequireCurrent();
    private partial Task HandshakeAsync();
    private partial Task<Frame> ReadAsync(CancellationToken cancellation);
    private partial Task WriteAsync(Frame frame, CancellationToken cancellation);
    private partial Task SendLockedAsync(byte[] plain, CancellationToken cancellation);
    private partial Task DeliverAsync();
    private partial Task ReceiveAsync();
}
