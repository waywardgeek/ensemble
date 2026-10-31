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
    bool RequireToolExplanation = false, bool ResponsesOnly = false);

public sealed partial class Engine
{
    // Exact selected-model lookup. There is no default media row; a miss still
    // permits text. Vendor is part of the key, so names cannot cross dialects.
    public static partial bool LookupModel(Vendor vendor, string model, out ModelFeatures features);
}

public sealed partial class BlobRef
{
    // Resolve a local attachment only when constructing a request. History keeps
    // this locator, not a second authoritative copy of its external bytes.
    // Chapter 11 also accepts inline base64 data URIs. Remote URI fetching and
    // job-handle resolution remain unsupported; this method does no network I/O.
    public partial byte[] Data();
}
