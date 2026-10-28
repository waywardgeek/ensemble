using System.Collections.Immutable;

namespace Ensemble;

/// <summary>Value-like observation of ownership, not a second owner or registry.
/// Agent is the actual identity, as in AgentObservation; the snapshot grants no
/// task-control authority and does not keep that Agent logically alive. Names
/// and child IDs remain on Agent. Two branches may have the same local child ID.
/// Children captures membership, not live scheduling or submission state.</summary>
public sealed record AgentTreeNode(Agent Agent, ImmutableArray<AgentTreeNode> Children);

public sealed partial class Ensemble
{
    /// <summary>Capture standalone Agents and their actual delegated descendants.
    /// The Ensemble itself is the common owner, not a fabricated Agent. Each
    /// membership list is captured under its own lock, released before descent.
    /// This is not a globally atomic lifecycle barrier: concurrent additions may
    /// appear on the next refresh and captured Agents may subsequently detach.
    /// Consumers query WorkStatusAsync and Engine.Meter outside this traversal,
    /// rather than holding ownership locks while awaiting an actor.</summary>
    public partial ImmutableArray<AgentTreeNode> AgentTree();
    private static partial AgentTreeNode CaptureTree(Agent agent);
}
