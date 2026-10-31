using Ensemble;
using System.Collections.Immutable;
using System.Text.Json;

namespace EnsembleGui;

public static partial class Gui
{
    private const int MaximumOrdinaryControlBytes = 1024 * 1024;
    private const int MaximumImageControlBytes = ((Agent.MaximumImageSourceBytes + 2) / 3) * 4 + 64 * 1024;
    private static partial (AgentInput Input, ImmutableArray<PreparedImage> Images) PrepareImageInput(Agent selected, JsonElement command, CancellationToken cancellation);
    private static partial string TraceProjection(string json);
}
