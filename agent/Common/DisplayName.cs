namespace Ensemble;

public sealed partial class Agent
{
    private string? displayName;
    /// <summary>Optional human-readable label, never an address or storage identity.</summary>
    public partial string? DisplayName();
    /// <summary>Persist before publishing. Null or blank restores identity-based labels.
    /// Cancellation after admission detaches the observer, not the update.</summary>
    public partial Task SetDisplayNameAsync(string? name, CancellationToken cancellation = default);
    internal static partial string? NormalizeDisplayName(string? name);
    private partial void ApplyDisplayName(Mail message);
}
