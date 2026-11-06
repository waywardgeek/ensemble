namespace Ensemble;

public sealed partial class Configuration
{
    internal partial Configuration(Configuration policy, Model model)
    {
        ArgumentNullException.ThrowIfNull(policy);
        ArgumentNullException.ThrowIfNull(model);
        // Capture all route facts under the actual graph's gate. This projection
        // neither registers a selection nor owns a credential-renewal lifetime.
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
        // Application instructions and security boundaries belong to the Agent.
        // Provider-specific cache/deferred settings do not travel to a new route.
        PrimaryInstructions = policy.PrimaryInstructions;
        ToolRoundLimit = policy.ToolRoundLimit;
        SafeMode = policy.SafeMode;
        EnableWebSearch = policy.EnableWebSearch;
        SandboxRoot = policy.SandboxRoot;
        DisableStreaming = policy.DisableStreaming;
        RequireToolExplanation = policy.RequireToolExplanation || Features.RequireToolExplanation;
        DeferredTools = false;
        StubsToolResults = false;
        CacheBreakpoints = false;
        PromptCacheKey = "";
        ReasoningEffort = "";
        ReasoningSummary = "";
        if (Surface == Surface.Responses && !Features.ReasoningEfforts.IsDefaultOrEmpty)
        {
            if (!Features.ReasoningEfforts.Contains("medium")
                || Features.ReasoningSummaries.IsDefaultOrEmpty
                || !Features.ReasoningSummaries.Contains("auto"))
            {
                throw new ArgumentException("Configured model does not support default reasoning options.");
            }
            ReasoningEffort = "medium";
            ReasoningSummary = "auto";
        }
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
