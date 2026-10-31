namespace Ensemble;

// Value-like observation, not a mutable lifecycle owner or completion signal.
// Submission and failure remain separate from activity: a submitted child can
// still be draining, and an idle child has not necessarily submitted anything.
public enum AgentActivity { Idle, Busy, Stopping, Stopped }
public sealed record AgentWorkStatus(AgentActivity Activity, bool Paused, bool Submitted, string? Failure);

public sealed partial class Agent
{
    /// <summary>Query live scheduling state on the actor. Closed lifetimes use
    /// synchronized shutdown state instead. Cancellation detaches only this
    /// observer. This snapshot is not a completion or shutdown barrier.</summary>
    public partial Task<AgentWorkStatus> WorkStatusAsync(CancellationToken cancellation = default);
    private partial void ReadWorkStatus(Mail message);
    private partial AgentWorkStatus CaptureWorkStatus(bool busy, bool isPaused);
}
