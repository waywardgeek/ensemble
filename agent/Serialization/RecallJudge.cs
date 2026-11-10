namespace Ensemble;

public sealed partial class Configuration
{
    internal partial Configuration ForRecallJudge(string model, string instructions)
    {
        if (Offering is not null)
        {
            if (!string.IsNullOrWhiteSpace(model) && model != Model)
                throw new ArgumentException("Recall override requires the configured offering's exact model.");
            return new Configuration(this, instructions, false);
        }
        return new Configuration(Vendor, string.IsNullOrWhiteSpace(model) ? Model : model,
            BaseUrl, ApiKey, primaryInstructions: instructions,
            disableStreaming: DisableStreaming, cacheBreakpoints: false,
            surface: Surface, billing: Billing, account: Account,
            safeMode: true, enableWebSearch: false, sandboxRoot: SandboxRoot, processReadPaths: ProcessReadPaths);
    }
}
