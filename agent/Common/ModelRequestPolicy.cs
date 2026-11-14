namespace Ensemble;

// Named declaration adaptation, never an arbitrary request mutation callback.
public enum ToolSchemaPolicy { Unchanged, PatternLengthCompatibility }

/// <summary>Credential-free request options for one configured offering. This
/// value carries neither a destination nor permission to contact one.</summary>
public sealed partial record ModelRequestPolicy
{
    public readonly ToolSchemaPolicy ToolSchemas;
    public readonly bool DisableStreaming;
    public readonly bool RequireToolExplanation;
    public readonly bool DeferredTools;
    public readonly bool StubsToolResults;
    public readonly bool CacheBreakpoints;
    public readonly string PromptCacheKey;
    public readonly string ReasoningEffort;
    // "off" explicitly suppresses a supported summary; empty selects the default.
    public readonly string ReasoningSummary;

    public partial ModelRequestPolicy(bool disableStreaming = false,
        bool requireToolExplanation = false, bool deferredTools = false,
        bool stubsToolResults = false, bool cacheBreakpoints = false,
        string promptCacheKey = "", string reasoningEffort = "", string reasoningSummary = "",
        ToolSchemaPolicy toolSchemas = ToolSchemaPolicy.Unchanged);
    internal partial ModelRequestPolicy ForModel(Surface surface, ModelFeatures features);
}
