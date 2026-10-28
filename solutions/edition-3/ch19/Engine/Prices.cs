namespace Ensemble;

public static partial class TokenPrices
{
    public static partial decimal? EstimateUsd(UsageIdentity identity, Usage counts)
    {
        ArgumentNullException.ThrowIfNull(identity);
        ArgumentNullException.ThrowIfNull(counts);
        // Verified primary model pages, USD per million text tokens:
        // https://developers.openai.com/api/docs/models/gpt-4o-mini
        // https://developers.openai.com/api/docs/models/gpt-4o
        // These pages list deprecated models too; pricing is not availability.
        // Only the documented default alias/snapshot pairs are included here.
        // Missing service_tier is intentionally unpriced: project-level routing
        // may select a tier other than standard even when the caller omitted it.
        if (identity.Billing != BillingMode.ApiKey || identity.Vendor != Vendor.OpenAI || identity.ServiceTier != "default")
        {
            return null;
        }
        (decimal Input, decimal Read, decimal Output)? rates = identity.Model switch
        {
            "gpt-4o-mini" or "gpt-4o-mini-2024-07-18" => (0.15m, 0.075m, 0.60m),
            "gpt-4o" or "gpt-4o-2024-08-06" => (2.50m, 1.25m, 10m),
            _ => null
        };
        if (rates is null)
        {
            return null;
        }
        // This API has no separately verified cache-write price. Zero writes
        // add zero; a nonzero or unknown write count must not be priced as free.
        if (counts.Input is not long input || counts.CacheRead is not long read
            || counts.Output is not long output || counts.CacheWrite != 0)
        {
            return null;
        }
        var price = rates.Value;
        return (input * price.Input + read * price.Read + output * price.Output) / 1_000_000m;
    }
}
