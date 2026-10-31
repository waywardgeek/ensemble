using System.Collections.Immutable;

namespace Ensemble;

/// <summary>Value-like durable launch choices. These are not the running Agent's
/// mutable model selection, credentials, authority, or conversation.</summary>
public sealed record PeerLaunch(string Name, string Model, string Skill, DateTimeOffset Created);

public sealed partial class Ensemble
{
    // Serialize registry read/modify/replace and restoration. This is not another
    // live registry: Agents() remains the sole Ensemble ownership collection.
    private readonly SemaphoreSlim peerOperations = new(1, 1);
    private readonly CancellationTokenSource peerLifetime = new();

    /// <summary>Create a persistent sibling using a current top-level Agent's
    /// configuration and parsed skill templates. Null/empty model and skill inherit
    /// that source. No source relationship is retained. Success means startup,
    /// initial History save and atomic ensemble.json replacement have completed.
    /// A host that can accept an independent browser may opt into waiting for it.
    /// Headless/default callers retain immediate missing-browser refusal. Cancelling
    /// an opted-in wait joins cleanup and removes the pending peer.</summary>
    public partial Task<Agent> CreatePeerAsync(Agent source, string name, string? model = null,
        string? skill = null, CancellationToken cancellation = default, bool waitForBrowser = false);

    /// <summary>Restore all registry peers, not delegated children. The host's
    /// initial Agent is already constructed and is not a registry entry. Validate
    /// the entire registry before construction; roll back new live membership on
    /// failure. Ordinary application destruction preserves registry and data.</summary>
    public partial Task<ImmutableArray<Agent>> RestorePeersAsync(Agent source,
        CancellationToken cancellation = default, bool waitForBrowser = false);

    /// <summary>Explicitly stop and remove a registered live peer. Refuse retained
    /// unfinished children before closing admission. Once shutdown starts, join to
    /// a definite result even if the observer cancels. Remove the registry entry
    /// before detaching the Agent; preserve its saved files. A failed registry
    /// write leaves the stopped peer registered for a later retry.</summary>
    public partial Task DeletePeerAsync(Agent source, string name, CancellationToken cancellation = default);

    private partial CancellationTokenSource BeginPeerOperation(Agent source, CancellationToken cancellation);
    private partial ImmutableArray<PeerLaunch> ReadPeerLaunches(Agent source);
    private partial void SavePeerLaunches(Agent source, ImmutableArray<PeerLaunch> entries);
    private partial Task<Agent> ConstructPeerAsync(PeerLaunch entry, Configuration configuration,
        SkillBootstrap? bootstrap, CancellationToken cancellation, bool waitForBrowser);
}

public sealed partial class Agent
{
    /// <summary>Under parent admission, check the existing bottom-up completion
    /// rule and begin ordinary joined shutdown. This is deletion's guard, not a
    /// claim that ShutdownAsync is graceful or equivalent to a future kill API.</summary>
    internal partial Task ShutdownForPeerDeletionAsync();
}
