using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Engine
{
    internal partial void ImportUsage(MeterSnapshot usage)
    {
        ArgumentNullException.ThrowIfNull(usage);
        lock (meterGate)
        {
            foreach (var row in usage.Rows)
            {
                meterRows[row.Identity] = meterRows.TryGetValue(row.Identity, out var previous)
                    ? previous.Add(row.Counts) : row.Counts;
            }
        }
    }

    public partial MeterSnapshot Meter()
    {
        lock (meterGate)
        {
            var rows = meterRows.Select(pair => new MeterEntry(pair.Key, pair.Value)).ToImmutableArray();
            return new MeterSnapshot(meterLatest, rows);
        }
    }

    private partial void RecordUsage(UsageIdentity identity, Usage usage)
    {
        lock (meterGate)
        {
            meterRows.TryGetValue(identity, out var previous);
            meterRows[identity] = previous is null ? usage : previous.Add(usage);
            // For concurrent direct sends, latest means latest completion. The
            // normal Agent serializes requests, so no second scheduler is needed.
            meterLatest = new MeterEntry(identity, usage);
        }
    }
}

public sealed partial class MeterSnapshot
{
    public partial MeterSnapshot(MeterEntry? latest, ImmutableArray<MeterEntry> rows)
    {
        if (rows.IsDefault)
        {
            throw new ArgumentException("Meter rows must be initialized.", nameof(rows));
        }
        Latest = latest;
        Rows = rows;
        var totals = new Usage();
        foreach (var row in rows)
        {
            totals = totals.Add(row.Counts);
        }
        Totals = totals;
    }

    public partial double? HitRate()
    {
        long? all = Totals.AllInput();
        if (all is not long count || count == 0 || Totals.CacheRead is not long read)
        {
            return null;
        }
        return (double)read / count;
    }

    public partial decimal? EstimatedUsd()
    {
        decimal total = 0;
        foreach (var row in Rows)
        {
            decimal? cost = TokenPrices.EstimateUsd(row.Identity, row.Counts);
            if (cost is null)
            {
                return null;
            }
            total += cost.Value;
        }
        return total;
    }
}
