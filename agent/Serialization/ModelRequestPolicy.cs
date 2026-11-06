namespace Ensemble;

public sealed partial record ModelRequestPolicy
{
    public partial ModelRequestPolicy(bool disableStreaming, bool requireToolExplanation,
        bool deferredTools, bool stubsToolResults, bool cacheBreakpoints,
        string promptCacheKey, string reasoningEffort, string reasoningSummary)
    {
        ArgumentNullException.ThrowIfNull(promptCacheKey);
        ArgumentNullException.ThrowIfNull(reasoningEffort);
        ArgumentNullException.ThrowIfNull(reasoningSummary);
        DisableStreaming = disableStreaming;
        RequireToolExplanation = requireToolExplanation;
        DeferredTools = deferredTools;
        StubsToolResults = stubsToolResults;
        CacheBreakpoints = cacheBreakpoints;
        PromptCacheKey = promptCacheKey;
        ReasoningEffort = reasoningEffort;
        ReasoningSummary = reasoningSummary;
    }

    internal partial ModelRequestPolicy ForModel(Surface surface, ModelFeatures features)
    {
        if (!Enum.IsDefined(surface)) throw new ArgumentException("Unknown wire surface.");
        if ((DeferredTools || CacheBreakpoints) && surface != Surface.Messages)
        {
            throw new ArgumentException("Deferred tools and block caching require Messages.");
        }
        if (PromptCacheKey.Length != 0 && surface is not (Surface.ChatCompletions or Surface.Responses))
        {
            throw new ArgumentException("Prompt cache keys require an OpenAI-compatible surface.");
        }
        bool supportsReasoning = surface == Surface.Responses
            && !features.ReasoningEfforts.IsDefaultOrEmpty;
        if (!supportsReasoning && (ReasoningEffort.Length != 0 || ReasoningSummary.Length != 0))
        {
            throw new ArgumentException("Reasoning options require verified model and surface support.");
        }
        string effort = supportsReasoning && ReasoningEffort.Length == 0 ? "medium" : ReasoningEffort;
        string summary = supportsReasoning && ReasoningSummary.Length == 0 ? "auto" : ReasoningSummary;
        if (supportsReasoning && !features.ReasoningEfforts.Contains(effort))
        {
            throw new ArgumentException("Unsupported reasoning effort for configured model.");
        }
        if (supportsReasoning && summary != "off"
            && (features.ReasoningSummaries.IsDefaultOrEmpty || !features.ReasoningSummaries.Contains(summary)))
        {
            throw new ArgumentException("Unsupported reasoning summary for configured model.");
        }
        return new ModelRequestPolicy(DisableStreaming,
            RequireToolExplanation || features.RequireToolExplanation, DeferredTools,
            StubsToolResults, CacheBreakpoints, PromptCacheKey, effort, summary);
    }
}
