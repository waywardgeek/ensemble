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

// A renderer declaration, not a live discovery result or promise of availability.
public readonly record struct ModelFeatures(Media Media);

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
    // URI and job-handle attachments remain deferred in Chapter 5.
    public partial byte[] Data();
}
