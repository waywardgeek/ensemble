using System.Collections.Immutable;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace Ensemble;

// A value-like startup record in the child's own directory, not a second fleet
// registry or transcript. Current host policy and credentials are never saved here.
internal sealed record ChildRestartState(int Version, string ID, string Model,
    JsonElement Schema, long Generation, bool Submitted, string? SubmissionHash, bool Paused,
    GoalConfiguration Goal,
    [property: JsonIgnore(Condition = JsonIgnoreCondition.WhenWritingNull)] ModelIdentity? Offering = null);

public sealed partial class Agent
{
    private bool childRestartEnabled;
    private bool childRestoring;
    /// <summary>Restore registered ordinary descendants under current runtime
    /// authority. Restore History, not interrupted calls or process Jobs. Enabled
    /// goals restart with a fresh idle interval; paused/submitted children stay idle.
    /// Existing live children are left alone. Temporary compressors are excluded.</summary>
    public partial Task RestoreChildrenAsync(CancellationToken cancellation = default, bool waitForBrowser = false);
    internal partial void RetireChildRestart();
    private partial void SaveChildRestart(GoalConfiguration? goal = null,
        bool? pause = null, long? generation = null, bool? submitted = null,
        Configuration? selection = null, JsonElement? completion = null);
    private partial ChildRestartState? ReadChildRestart(string id);
    private partial Task<Agent> CreateChildCoreAsync(string childID, string? model,
        JsonElement? submitSchema, ChildRestartState? restart, CancellationToken cancellation, bool waitForBrowser = false);
    private partial Task RestoreChildGoalAsync(ChildRestartState restart, CancellationToken cancellation);
}
