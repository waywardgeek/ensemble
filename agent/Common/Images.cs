namespace Ensemble;

// Immutable delivery value. The image contains actual bytes, not a source path.
// Original dimensions are measured after applying the encoded orientation.
public readonly record struct PreparedImage(BlobPart Image, int OriginalWidth,
    int OriginalHeight, int Width, int Height, int SourceBytes, int OutputBytes)
{
    public bool Resized => Width != OriginalWidth || Height != OriginalHeight;
}

public sealed partial class Agent
{
    public const int MaximumImageOutputBytes = 8 * 1024 * 1024;
    public const int MaximumImageSourceBytes = 32 * 1024 * 1024;
    public const long MaximumImagePixels = 24_000_000;
    public const int MaximumImageEdge = 2048;

    private static partial void ApplyImageOrientation(SkiaSharp.SKCanvas canvas,
        SkiaSharp.SKEncodedOrigin origin, int width, int height);

    // Preparation is admission, not provider rendering. Callers retain the
    // resulting bytes in the accepted message; rendering never rereads originals.
    public partial PreparedImage PrepareImage(string path);
    public partial PreparedImage PrepareImage(ReadOnlyMemory<byte> source);
}
