using System.Collections.Immutable;

namespace Ensemble;

// Value-like capability facts. These flags describe attachments independently:
// image support never implies audio, video, or document support.
[Flags]
public enum Media
{
    None = 0,
    Image = 1,
    Audio = 2,
    Video = 4,
    Document = 8
}

// Value-like renderer capabilities and Ensemble policy, not live discovery or
// a claim that an account can access the model.
// Default arrays mean no verified reasoning options. Exact rows avoid prefix guesses.
public readonly record struct ModelFeatures(Media Media, bool Stream = false,
    ImmutableArray<string> ReasoningEfforts = default,
    ImmutableArray<string> ReasoningSummaries = default,
    bool RequireToolExplanation = false, bool ResponsesOnly = false,
    long? ContextWindow = null);

public sealed partial class Engine
{
    // One immutable, value-like capability table. This is application knowledge,
    // not account discovery, pricing, or a guarantee of current availability.
    private static readonly ImmutableArray<(Vendor Vendor, string Name, ModelFeatures Features)> modelCatalog =
    [
        // Historical launch window, not current availability or a beta limit.
        // https://www.anthropic.com/news/claude-3-family
        (Vendor.Anthropic, "claude-3-haiku-20240307", new(Media.Image, Stream: true, ContextWindow: 200_000)),
        // https://developers.openai.com/api/docs/models/gpt-4o
        (Vendor.OpenAI, "gpt-4o", new(Media.Image, Stream: true, ContextWindow: 128_000)),
        // Documented 128,000-token window (not availability or entitlement):
        // https://developers.openai.com/api/docs/models/gpt-4o-mini
        (Vendor.OpenAI, "gpt-4o-mini", new(Media.Image, Stream: true, ContextWindow: 128_000)),
        // Ordinary pre-tool explanation is Ensemble policy. Human image input
        // follows the documented Responses contract; tool media is separate.
        // https://developers.openai.com/api/docs/guides/images-vision
        (Vendor.OpenAI, "gpt-6-astra", new(Media.Image, Stream: true,
            ReasoningEfforts: ["low", "medium", "high", "xhigh", "max"],
            ReasoningSummaries: ["auto"], RequireToolExplanation: true, ResponsesOnly: true)),
        // Documented input-token limit, not a combined input/output allowance.
        // https://ai.google.dev/gemini-api/docs/models/gemini-2.5-flash
        (Vendor.Gemini, "gemini-2.5-flash",
            new(Media.Image | Media.Audio | Media.Video | Media.Document,
                Stream: true, ContextWindow: 1_048_576))
    ];

    // Human-curated recommendations, not media capability, pricing, entitlement,
    // or automatic vendor discovery. Order is deliberate; Astra remains primary.
    public static ImmutableArray<(Vendor Vendor, string Name, string Label)> RecommendedModels { get; } =
    [
        (Vendor.OpenAI, "gpt-6-astra", "GPT-6 Astra"),
        (Vendor.OpenAI, "gpt-5.6-sol", "GPT-5.6 Sol"),
        (Vendor.Anthropic, "claude-opus-5", "Claude Opus 5"),
        (Vendor.Anthropic, "claude-sonnet-5", "Claude Sonnet 5"),
        (Vendor.Anthropic, "claude-opus-4-6", "Claude Opus 4.6"),
        (Vendor.Gemini, "gemini-3.8-flash", "Gemini 3.8 Flash"),
        (Vendor.Gemini, "gemini-4-argon", "Gemini 4 Argon")
    ];

    // Preserve current selection (including unknown IDs), then offer compatible,
    // same-vendor recommendations. Never silently change a running configuration.
    public static partial ImmutableArray<string> ModelChoices(Configuration configuration);

    // Request-time bytes use the actual Agent boundary. BlobRef stays a
    // value-like locator; it neither owns nor carries a filesystem service.
    private partial byte[] ReadMedia(BlobRef reference);

    // Exact selected-model lookup. There is no default media row; a miss still
    // permits text. Vendor is part of the key, so names cannot cross dialects.
    public static partial bool LookupModel(Vendor vendor, string model, out ModelFeatures features);
}

public sealed partial class BlobRef
{
    // Explicit standalone resolution has ordinary host filesystem access.
    // Engine must instead use its Agent's Sandbox for path attachments.
    // History keeps this locator, not a second copy of its external bytes.
    // Chapter 11 also accepts inline base64 data URIs. Remote URI fetching and
    // job-handle resolution remain unsupported; this method does no network I/O.
    public partial byte[] Data();
}
