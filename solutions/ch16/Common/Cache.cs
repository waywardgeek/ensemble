namespace Ensemble;

// Value-like diagnostic records, not conversation state. Capture owns fresh byte
// arrays; callers may inspect or edit their returned copy without changing Engine. Bytes measure a local
// approximation of provider prompt order; they are never token/billing counts.
public sealed record CachePrediction(long Sequence, int PreviousBytes, int CurrentBytes,
    int SharedBytes, double PredictedHitRate);
public sealed record CacheCapture(byte[] Wire, byte[] Prompt, byte[] Options);

/// <summary>Engine-owned, opt-in private request diagnostics. Files contain prompt
/// content and must not be published. No authorization headers are captured.</summary>
public sealed partial class CacheLens
{
    internal Engine? parent;
    internal bool destroyed;
    private readonly object gate = new();
    private readonly string directory;
    private CacheCapture? previous;
    private long sequence;

    // Register both ends. The caller selects an owned diagnostic directory.
    public partial CacheLens(Engine parent, string directory);
    public partial Engine Parent();
    // Called only at actual transport admission, never during offline Render.
    internal partial CachePrediction? Begin(string body);
    internal partial void Complete(CachePrediction prediction, Usage? usage);
    public partial void Destroy();

    // Pure byte transforms for diagnostic use. These do not canonicalize JSON,
    // normalize text, or remove similarly named fields from user tool arguments.
    public static partial byte[] StripMarkers(byte[] wire);
    public static partial CacheCapture Capture(Vendor vendor, byte[] wire);
}

public sealed partial class Engine
{
    private readonly object lensGate = new();
    private CacheLens? cacheLens;
    public partial CacheLens? Lens();
    internal partial void AddLens(CacheLens child);
    internal partial void RemoveLens(CacheLens child);
}
