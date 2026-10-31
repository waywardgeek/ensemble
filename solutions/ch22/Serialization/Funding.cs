namespace Ensemble;

public sealed partial class Configuration
{
    public partial string FundingDescription()
    {
        // A plan refusal never authorizes use of an API key. This description
        // reports immutable launch policy, not remaining quota or a price quote.
        string funding = Billing == BillingMode.ChatGptPlan
            ? $"ChatGPT plan (account: {Account}); no API-key fallback."
            : "API key (metered usage).";
        return $"Model: {Model}. Funding: {funding}";
    }
}
