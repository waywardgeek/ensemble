using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Agent
{
    // Ensemble owns every Agent. A delegated Agent also belongs to one Agent
    // parent, which owns its lifetime. Local IDs address siblings, not the root
    // application's standalone name index. Neither relationship is a service copy.
    private Agent? agentParent;
    private readonly string? childID;
    private readonly List<Agent> children = [];
    private Task? shutdownTask;

    /// <summary>Construct a delegated Agent without admitting a task. Both strong
    /// parents are explicit. The requested policy must not widen its Agent parent's
    /// boundary. Callers configure persistence and skills before submitting work.
    /// Output directory ownership remains with Jobs; destruction preserves files.</summary>
    public partial Agent(Ensemble parent, Agent owner, string childID,
        Configuration configuration, string? outputDirectory = null, SkillSpec? skillSpec = null,
        System.Text.Json.JsonElement? submitSchema = null, string? dataDirectory = null);
    /// <summary>Create a ready child beneath this Agent's data home.
    /// Inherits current configuration and parsed skill templates, not live grants
    /// or clients. Restores only the child's History before initialization. Does
    /// not submit a task. Failed/cancelled creation destroys the child, retaining
    /// its files; callers may retry the same ID after cleanup has completed.</summary>
    public partial Task<Agent> CreateChildAsync(string childID, string? model = null,
        System.Text.Json.JsonElement? submitSchema = null,
        CancellationToken cancellation = default);
    public partial Agent? AgentParent();
    public partial string? ChildID();
    internal static partial void ValidateChildID(string childID);
    public partial ImmutableArray<Agent> Children();
    public partial Agent? FindChild(string id);
    internal partial void AddChild(Agent child);
    internal partial void RemoveChild(Agent child);
    private partial void ConstructOwners(Ensemble root, Agent? owner,
        string? outputDirectory, SkillSpec? skillSpec, string? dataDirectory);
    private partial void DestroyChildren();
}
