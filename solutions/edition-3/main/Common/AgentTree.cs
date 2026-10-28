using System.Collections.Immutable;

namespace Ensemble;

/// <summary>Value-like observation of ownership, not a second owner or registry.
/// Agent is the actual identity, as in AgentObservation; the snapshot grants no
/// task-control authority and does not keep that Agent logically alive. Names
/// and child IDs remain on Agent. Two branches may have the same local child ID.
/// Children captures membership, not live scheduling or submission state.</summary>
public sealed record AgentTreeNode(Agent Agent, ImmutableArray<AgentTreeNode> Children, string? Address = null);

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
    // Trusted host lookup only, never model-facing child-tool authority. Each
    // slash-separated local name is canonically percent-escaped. No default
    // root is inferred; lookup does not acquire ownership or a lifetime.
    public partial Agent? FindAgentAddress(string address);
    private static partial AgentTreeNode CaptureTree(Agent agent);
}

public sealed partial class Agent
{
    // Derived through real owning parents, never stored as a second identity.
    // Unnamed transient roots and their descendants have no global address.
    // Detached Agents also return null. This is an address, not a filesystem
    // path or a guarantee that the same object will occupy it after recreation.
    public partial string? Address();
}
