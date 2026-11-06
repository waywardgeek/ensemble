using System.Collections.Immutable;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

/// <summary>Human-selected working size, distinct from provider capacity and
/// byte-based result stubbing. All quantities are whole input tokens.</summary>
public sealed partial class WorkingContextLimit
{
    public readonly Vendor Vendor;
    public readonly string Model;
    public readonly ModelIdentity? Offering;
    public readonly long WorkingTokens;
    public readonly long? CapacityTokens;
    public partial WorkingContextLimit(Vendor vendor, string model, long workingTokens,
        long? capacityTokens = null, ModelIdentity? offering = null);
}

/// <summary>Latest actor-accepted request, not a next-request prediction.
/// A completed checkpoint invalidates the old input measurement.</summary>
public sealed record WorkingContextStatus(Vendor Vendor, string Model, long? InputTokens,
    long? WorkingTokens, long? CapacityTokens, decimal? Percent, ModelIdentity? Offering = null);

public static partial class JsonCodec
{
    public static partial JsonArray WriteWorkingContextLimits(ImmutableArray<WorkingContextLimit> limits);
    public static partial ImmutableArray<WorkingContextLimit> ReadWorkingContextLimits(JsonElement value);
}

public sealed partial class Agent
{
    private ImmutableArray<WorkingContextLimit> workingContextLimits = [];
    public partial ImmutableArray<WorkingContextLimit> WorkingContextLimits();
    public partial Task SetWorkingContextLimitAsync(WorkingContextLimit limit,
        bool persist = false, CancellationToken cancellation = default);
    private partial void ApplyWorkingContextLimit(Mail message);
    private partial void RestoreWorkingContextLimits(ImmutableArray<WorkingContextLimit> limits);
}
