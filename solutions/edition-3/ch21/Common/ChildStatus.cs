namespace Ensemble;

// Value-like observation, not a mutable lifecycle owner or completion signal.
// Submission and failure remain separate from activity: a submitted child can
// still be draining, and an idle child has not necessarily submitted anything.
public enum AgentActivity { Idle, Busy, Stopping, Stopped }
public sealed record AgentWorkStatus(AgentActivity Activity, bool Paused, bool Submitted, string? Failure,
    WorkActivity? Work = null);

// Value-like progress, independent of lifecycle, submission and History's
// request frontier. Waiting is an executing observation, not completed work.
public enum WorkPhase { Idle, Thinking, ToolCall, Waiting }
public sealed record WorkActivity(WorkPhase Phase, string Tool = "", string Detail = "");

public sealed partial class Agent
{
    /// <summary>Query live scheduling state on the actor. Closed lifetimes use
    /// synchronized shutdown state instead. Cancellation detaches only this
    /// observer. This snapshot is not a completion or shutdown barrier.</summary>
    public partial Task<AgentWorkStatus> WorkStatusAsync(CancellationToken cancellation = default);
    // This is only the last emitted projection for duplicate suppression.
    // Open status derives current activity from actor state. Once admission
    // closes, readers use this final/closing projection without racing that state.
    private WorkActivity lastEmittedActivity = new(WorkPhase.Idle);
    private partial WorkActivity ReadWorkActivity();
    private static partial WorkActivity ToolActivity(ToolCallPart call);
    private partial void PublishWorkActivity();
    private partial void ReadWorkStatus(Mail message);
    private partial AgentWorkStatus CaptureWorkStatus(bool busy, bool isPaused);
}
