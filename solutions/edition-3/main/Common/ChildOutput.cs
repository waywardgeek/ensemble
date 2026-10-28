using System.Collections.Immutable;

namespace Ensemble;

/// <summary>Value-like finalized output. Cursor is this child's durable History
/// sequence, not a byte offset or global identity. Status is a separate current
/// actor snapshot; idle does not imply submission.</summary>
public sealed record ChildOutput(string AgentID, long Cursor, string Output, AgentWorkStatus Status);

/// <summary>Value-like observation of fixed direct-child membership. A reason
/// describes why observation returned, never permission to cancel the children.</summary>
public sealed record ChildChange(string Reason, ImmutableArray<ChildOutput> Agents,
    ImmutableArray<QuestionInfo> Questions);

public sealed partial class Agent
{
    // Only transient delta publication advances this counter. History already
    // owns durable progress. Polling status must not manufacture progress.
    private long streamProgress;

    /// <summary>Read finalized assistant text and tool-result text since cursor.
    /// Waiting ends on new text or a nonbusy status. Cancellation detaches only
    /// this observer; the lossy shared observation queue is never consumed.</summary>
    public partial Task<ChildOutput> ReadOutputAsync(long cursor = 0, bool wait = false,
        CancellationToken cancellation = default);

    /// <summary>Observe a fixed set of direct children using bounded snapshot
    /// polling. Defaults are 300 seconds overall and 15 seconds without progress.
    /// Zero disables the respective deadline. Pattern matching has a bounded
    /// regex execution time. Pending questions belong only to selected children.</summary>
    public partial Task<ChildChange> WaitForChildChangeAsync(IEnumerable<string>? childIDs = null,
        string? pattern = null, double timeoutSeconds = 300, double stuckSeconds = 15,
        CancellationToken cancellation = default);
    private partial ChildOutput CaptureOutput(ImmutableArray<Event> events, long cursor, AgentWorkStatus status);
}
