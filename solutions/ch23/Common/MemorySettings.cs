using System.Collections.Immutable;

namespace Ensemble;

/// <summary>Value-like byte policy. Zero sizes select defaults, not token limits.
/// Disabling hides this band without deleting files. The containing settings
/// validate effective low/high watermarks using the appropriate band default.</summary>
public sealed partial class MemoryBandSettings
{
    public readonly long Budget;
    public readonly long HighWatermark;
    public readonly long LowWatermark;
    public readonly bool Disabled;
    public partial MemoryBandSettings(long budget = 0, long highWatermark = 0,
        long lowWatermark = 0, bool disabled = false);
    internal partial MemoryBandSettings Effective(long defaultBudget);
    internal partial bool Same(MemoryBandSettings other);
}

/// <summary>Immutable per-Agent memory policy, independent of the existing
/// 400,000-byte tool-history target. Conversation defaults to 3,000 bytes;
/// corpus bands default to 1,500. Effective watermarks default to twice budget
/// and budget. SOUL never graduates and MEMORY is terminal.</summary>
public sealed partial class MemorySettings
{
    public readonly MemoryBandSettings Conversation;
    public readonly MemoryBandSettings Soul;
    public readonly MemoryBandSettings Memory;
    public readonly MemoryBandSettings Archive64;
    public readonly MemoryBandSettings Archive8;
    public readonly MemoryBandSettings Session;
    public partial MemorySettings(MemoryBandSettings? conversation = null,
        MemoryBandSettings? soul = null, MemoryBandSettings? memory = null,
        MemoryBandSettings? archive64 = null, MemoryBandSettings? archive8 = null,
        MemoryBandSettings? session = null);
    // Null denotes dialogue, which is not a corpus band.
    internal partial MemoryBandSettings For(MemoryBand? band);
}

public sealed partial class Agent
{
    // The actor applies policy; ingress protects the public immutable snapshot.
    // Readiness is actor-only and never copied into the disk component.
    private MemorySettings memorySettings = new();
    private bool memoryReady;
    public partial MemorySettings MemorySettings();
    /// <summary>Apply in-memory policy and acknowledge actual corpus reconciliation,
    /// even idle. This does not save a restart preference. Hosts normally use
    /// SaveMemorySettingsAsync instead. Cancellation detaches only the observer
    /// after admission; it does not retract the policy update.</summary>
    public partial Task SetMemorySettingsAsync(MemorySettings settings,
        CancellationToken cancellation = default);
    // Durable human settings. Admission is not canceled by a canceled observer.
    // Save failure leaves policy and live compression unchanged. A later apply
    // failure is reported, not acknowledged as success; restart reads the save.
    public partial Task SaveMemorySettingsAsync(MemorySettings settings,
        CancellationToken cancellation = default);
    private partial bool SaveMemorySettingsBeforeApply(Mail message);
    private partial void EnsureMemoryReady();
    private partial void ApplyMemorySettings(Mail message);
    private partial void RecoverMemory();
    private partial ImmutableArray<Event> ReadMemoryChanges(MemorySettings desired, bool startup);
    private partial void PublishMemorySettings(MemorySettings desired, ImmutableArray<Event> changes);
}

public static partial class JsonCodec
{
    // Exact six-band value document shared by persistence and trusted GUI wire.
    // It contains policy only: no paths, credentials, corpus, or runtime state.
    public static partial System.Text.Json.Nodes.JsonObject WriteMemorySettings(MemorySettings settings);
    public static partial MemorySettings ReadMemorySettings(System.Text.Json.JsonElement value);
    private static partial System.Text.Json.Nodes.JsonObject WriteMemoryBandSettings(MemoryBandSettings settings);
    private static partial MemoryBandSettings ReadMemoryBandSettings(System.Text.Json.JsonElement value);
}
