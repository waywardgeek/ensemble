namespace Ensemble;

public sealed partial class Configuration
{
    internal partial Configuration(Configuration policy, Model model)
    {
        ArgumentNullException.ThrowIfNull(policy);
        ArgumentNullException.ThrowIfNull(model);
        // Capture all route facts under the actual graph's gate. This projection
        // neither registers a selection nor owns a credential-renewal lifetime.
        ModelRequestPolicy requestPolicy;
        lock (model.graphGate)
        {
            ModelEndpoint endpoint = model.Parent();
            ModelAuthentication binding = endpoint.Authentication();
            if (!ReferenceEquals(endpoint.Parent().Parent(), binding.Parent()))
            {
                throw new InvalidOperationException("Configured model and authentication must share their root.");
            }
            binding.ValidateEndpoint(endpoint.Address, endpoint.Surface);
            Offering = model.Identity();
            Authentication = binding;
            Features = model.Features;
            requestPolicy = model.RequestPolicy;
            ServingContextTokens = model.ServingContextTokens;
            Surface = endpoint.Surface;
            Vendor = Surface switch
            {
                Surface.Messages => Vendor.Anthropic,
                Surface.ChatCompletions or Surface.Responses => Vendor.OpenAI,
                Surface.GenerateContent => Vendor.Gemini,
                _ => throw new ArgumentException("Unknown configured wire surface.")
            };
            Model = model.Name;
            ResolvedModel = model.Name;
            BaseUrl = endpoint.Address.AbsoluteUri.TrimEnd('/');
            ApiKey = "";
            Billing = binding.Kind switch
            {
                ModelAuthenticationKind.None => BillingMode.None,
                ModelAuthenticationKind.ApiKey => BillingMode.ApiKey,
                ModelAuthenticationKind.ChatGptPlan => BillingMode.ChatGptPlan,
                ModelAuthenticationKind.Codex => BillingMode.Codex,
                _ => throw new ArgumentException("Unknown configured authentication.")
            };
            Account = binding.Account;
        }
        // Instructions and security remain Agent policy. Request options come
        // from this offering, never from the route that happened to precede it.
        PrimaryInstructions = policy.PrimaryInstructions;
        ToolRoundLimit = policy.ToolRoundLimit;
        SafeMode = policy.SafeMode;
        EnableWebSearch = policy.EnableWebSearch;
        SandboxRoot = policy.SandboxRoot;
        DisableStreaming = requestPolicy.DisableStreaming;
        RequireToolExplanation = requestPolicy.RequireToolExplanation;
        DeferredTools = requestPolicy.DeferredTools;
        StubsToolResults = requestPolicy.StubsToolResults;
        CacheBreakpoints = requestPolicy.CacheBreakpoints;
        PromptCacheKey = requestPolicy.PromptCacheKey;
        ReasoningEffort = requestPolicy.ReasoningEffort;
        ReasoningSummary = requestPolicy.ReasoningSummary == "off" ? "" : requestPolicy.ReasoningSummary;
        if (Billing is BillingMode.ChatGptPlan or BillingMode.Codex
            && (DisableStreaming || !Features.Stream))
        {
            throw new ArgumentException("Configured subscription funding requires streaming capability.");
        }
    }

    private partial Configuration(Configuration source, string instructions, bool compressor)
        : this(source, true, false, source.SandboxRoot)
    {
        PrimaryInstructions = instructions;
        RequireToolExplanation = compressor;
        ToolRoundLimit = compressor ? 3 : 200;
        if (!compressor)
        {
            CacheBreakpoints = false;
            PromptCacheKey = "";
            ReasoningEffort = "";
            ReasoningSummary = "";
        }
    }
}
