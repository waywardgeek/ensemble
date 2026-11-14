using System.Globalization;

namespace Ensemble;

public sealed partial class Tool
{
    private static partial string Status(Job job)
    {
        // Reach the same Engine that sent the tool-requesting response through
        // real parents. The tool owns neither counters nor a copied service.
        Agent agent = job.Parent().Parent();
        MeterSnapshot meter = agent.Engine().Meter();
        WorkingContextStatus context = agent.WorkingContextStatus();
        Usage latest = meter.Latest?.Counts ?? new Usage(null, null, null, null);
        decimal? cost = meter.EstimatedUsd();
        double? hitRate = meter.HitRate();

        // Null means unknown, not zero. Invariant formatting keeps the report
        // legible and its units unambiguous on hosts with different locales.
        static string Count(long? count) => count?.ToString(CultureInfo.InvariantCulture) ?? "unknown";
        static string Counts(Usage usage) => $"{Count(usage.Input)} in, {Count(usage.Output)} out, "
            + $"{Count(usage.CacheRead)} cache read, {Count(usage.CacheWrite)} cache write";
        string hit = hitRate.HasValue
            ? (hitRate.Value * 100).ToString("0.0", CultureInfo.InvariantCulture) + "%"
            : "unknown";
        string price = cost.HasValue
            ? "$" + cost.Value.ToString("0.##########", CultureInfo.InvariantCulture) + " (estimate)"
            : "unknown (unpriced model/tier or incomplete usage; not a billing statement)";

        string occupancy = context.Percent is decimal percent
            ? percent.ToString("0.0", CultureInfo.InvariantCulture) + "%" : "unknown";

        // One immutable meter snapshot supplies latest and accumulated billing values.
        // Reading it neither records a model response nor restores old usage.
        return $"model: {agent.Configuration().Model}\n"
            + $"working context limit: {Count(context.WorkingTokens)} tokens\n"
            + $"configured context capacity: {Count(context.CapacityTokens)} tokens (not measured capacity)\n"
            + $"last main-request input: {Count(context.InputTokens)} tokens\n"
            + $"working context occupancy: {occupancy}\n"
            + $"last response: {Counts(latest)}\n"
            + $"session total: {Counts(meter.Totals)}\n"
            + $"cache hit rate: {hit}\n"
            + $"cost: {price}";
    }
}
