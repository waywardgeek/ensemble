using System.Collections.Immutable;

namespace Ensemble;

// Value-like keys and measurements, never parents or mutable pricing accounts.
// Model/tier are reported by the response. Empty model or absent tier means
// unknown identity, not permission to guess from the selected routing alias.
public sealed record UsageIdentity(Vendor Vendor, string Model, string? ServiceTier, BillingMode Billing = BillingMode.ApiKey);
public sealed record MeterEntry(UsageIdentity Identity, Usage Counts);

/// <summary>Value-like, coherent view of one Engine's current-process usage.
/// Latest is absent before the first completed response. Rows retain the actual
/// model/tier under which counts accrued; totals are derived from those rows.
/// Neither dollar amounts nor restored History seed the meter.</summary>
public sealed partial class MeterSnapshot
{
    public readonly MeterEntry? Latest;
    public readonly ImmutableArray<MeterEntry> Rows;
    public readonly Usage Totals;
    public partial MeterSnapshot(MeterEntry? latest, ImmutableArray<MeterEntry> rows);
    // A ratio, not a percentage or byte-prefix prediction. No input is unknown,
    // not a fabricated 0% cache hit. Output uncertainty does not affect this rate.
    public partial double? HitRate();
    // Standard token-price estimate, not an invoice or account balance. If any
    // contributing row cannot be priced, the session estimate stays unknown.
    public partial decimal? EstimatedUsd();
}

public sealed partial class Usage
{
    // Provider parsers normalize disjoint buckets before this sum. Overflow or
    // any unknown input bucket makes ALL unknown; output is independent.
    public partial long? AllInput();
}

public static partial class TokenPrices
{
    // Exact verified provider/model/tier rows only. No synthetic test models,
    // prefix matching, alias guessing, or implication of model availability.
    public static partial decimal? EstimateUsd(UsageIdentity identity, Usage counts);
}

public sealed partial class Engine
{
    private readonly object meterGate = new();
    private readonly Dictionary<UsageIdentity, Usage> meterRows = new();
    private MeterEntry? meterLatest;

    // Capture latest and rows together. No disk or History access occurs here;
    // callers can hold the immutable snapshot after the next send.
    public partial MeterSnapshot Meter();
    // Only completed SendAsync calls record usage. Parse/Render/replay do not.
    private partial void RecordUsage(UsageIdentity identity, Usage usage);
    // Same decoder for plain and completed streaming responses. Public Parse
    // discards transient billing identity and never mutates the process meter.
    private partial ImmutableArray<Event> ParseResponse(int status, string body, out UsageIdentity? billing);
}
