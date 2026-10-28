using System.Collections.Immutable;
using System.Text.Json;

namespace Ensemble;

/// <summary>Value-like join progress. A non-null Submissions dictionary is the
/// final result (possibly empty). Otherwise Questions contains newly observed
/// questions, not consumed replies. Typed payloads retain their JSON values.</summary>
public sealed record ChildJoinUpdate(ImmutableArray<QuestionInfo> Questions,
    ImmutableDictionary<string, JsonElement>? Submissions);

public sealed partial class Agent
{
    /// <summary>Observe a fixed set of direct children, selected when enumeration
    /// starts. Null/empty IDs selects all current children. Yield each pending
    /// question once, then finish with explicit current submissions. Failure or
    /// shutdown without submission fails the join. Disposing/cancelling this
    /// observer never cancels children or questions. This is not a quiescence
    /// barrier: await ShutdownAsync before destroying an owned child.</summary>
    public partial IAsyncEnumerable<ChildJoinUpdate> JoinChildrenAsync(
        IEnumerable<string>? childIDs = null, CancellationToken cancellation = default);
}
