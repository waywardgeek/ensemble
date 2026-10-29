using Ensemble;
using Ensemble.Gateway;

namespace Ensemble.GatewayAdapter;

/// <summary>One application-boundary invocation, not another Agent owner or
/// scheduler. RunAsync borrows a caller-owned Agent, just as a mounted GUI does.
/// The caller must keep that Agent alive until RunAsync returns, then decide
/// whether to drain or shut it down. The invocation owns its Client lifetime;
/// Client borrows this handler through the public IHandler boundary.</summary>
public sealed partial class Bridge : IHandler
{
    private readonly object gate = new();
    private readonly List<Task> replies = [];
    private readonly TaskCompletionSource connected = new(TaskCreationOptions.RunContinuationsAsynchronously);
    private readonly TaskCompletionSource ready = new(TaskCreationOptions.RunContinuationsAsynchronously);
    private Agent? agent;
    private Client? client;
    private CancellationTokenSource? lifetime;
    private bool started;
    private bool accepting;

    public partial Bridge();
    /// <summary>Connect, declare contactability, and wait for disconnect or caller
    /// cancellation. Single-use. Exit joins the Client and all reply observers,
    /// but never cancels admitted Agent work. Unexpected disconnect is a failure.
    /// There is no reconnect, durable receipt, or exactly-once execution claim.</summary>
    public partial Task RunAsync(Agent agent, ClientConfig configuration,
        IEnumerable<string> groups, CancellationToken cancellation = default);
    /// <summary>Settles after authenticated setup and contactability declaration.
    /// Not evidence of provider availability or completed inference.</summary>
    public partial Task Ready();
    public partial void OnConnected(WelcomePayload welcome);
    public partial void OnDeliver(Envelope envelope);
    public partial void OnDisconnected(string reason);
    public partial void OnError(Exception error);
    private partial Task ReplyAsync(Agent actor, Envelope envelope, Task<string> result,
        CancellationToken cancellation);
}
