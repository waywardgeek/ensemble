namespace Ensemble;

public sealed partial class Configuration
{
    internal partial Configuration ForRecallJudge(string model, string instructions)
    {
        return new Configuration(Vendor, string.IsNullOrWhiteSpace(model) ? Model : model,
            BaseUrl, ApiKey, primaryInstructions: instructions,
            disableStreaming: DisableStreaming, cacheBreakpoints: false,
            surface: Surface, billing: Billing, account: Account,
            safeMode: true, enableWebSearch: false, sandboxRoot: SandboxRoot);
    }
}
