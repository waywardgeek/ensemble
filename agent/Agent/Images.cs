using SkiaSharp;

namespace Ensemble;

public sealed partial class Agent
{
    public partial PreparedImage PrepareImage(string path)
    {
        using var stream = Sandbox().OpenRegularRead(path);
        long length = stream.Length;
        if (length <= 0 || length > MaximumImageSourceBytes)
        {
            throw new InvalidDataException("Image source must contain between 1 byte and 64 MiB.");
        }
        byte[] bytes = new byte[(int)length];
        stream.ReadExactly(bytes);
        if (stream.ReadByte() != -1)
        {
            throw new IOException("Image source grew while being read.");
        }
        return PrepareImage(bytes.AsMemory());
    }

    public partial PreparedImage PrepareImage(ReadOnlyMemory<byte> source)
    {
        if (source.Length == 0 || source.Length > MaximumImageSourceBytes)
        {
            throw new InvalidDataException("Image source must contain between 1 byte and 64 MiB.");
        }

        // SKData copies the bounded input. Native objects are scoped to this call;
        // no decoder or mutable bitmap survives in History or Agent state.
        using var data = SKData.CreateCopy(source.Span);
        using var codec = SKCodec.Create(data)
            ?? throw new InvalidDataException("Image format is not recognized.");
        if (codec.EncodedFormat is not (SKEncodedImageFormat.Png or SKEncodedImageFormat.Jpeg
            or SKEncodedImageFormat.Webp or SKEncodedImageFormat.Gif))
        {
            throw new InvalidDataException("Only PNG, JPEG, WebP and single-frame GIF images are supported.");
        }
        var info = codec.Info;
        if (info.Width <= 0 || info.Height <= 0
            || (long)info.Width * info.Height > MaximumImagePixels
            || info.Width > 32768 || info.Height > 32768)
        {
            throw new InvalidDataException("Image dimensions exceed the 24-million-pixel or 32768-edge source limit.");
        }
        // Skia reports zero frames for static non-animated formats.
        if (codec.FrameCount > 1 || codec.RepetitionCount != 0)
        {
            throw new InvalidDataException("Animated or multi-frame images are not supported.");
        }

        using var decoded = new SKBitmap(new SKImageInfo(info.Width, info.Height,
            SKColorType.Rgba8888, SKAlphaType.Premul));
        var result = codec.GetPixels(decoded.Info, decoded.GetPixels());
        if (result != SKCodecResult.Success)
        {
            throw new InvalidDataException($"Image decoding did not complete: {result}.");
        }

        bool swap = codec.EncodedOrigin is SKEncodedOrigin.LeftTop or SKEncodedOrigin.RightTop
            or SKEncodedOrigin.RightBottom or SKEncodedOrigin.LeftBottom;
        int originalWidth = swap ? info.Height : info.Width;
        int originalHeight = swap ? info.Width : info.Height;
        double scale = Math.Min(1.0, (double)MaximumImageEdge / Math.Max(originalWidth, originalHeight));
        int width = Math.Max(1, (int)Math.Floor(originalWidth * scale));
        int height = Math.Max(1, (int)Math.Floor(originalHeight * scale));

        // A finite reduction policy; do not retry providers or rewrite originals.
        // Each attempt is rendered from the original decoded pixels, not a
        // repeatedly resampled derivative. Output metadata is deliberately absent.
        for (int attempt = 0; attempt < 8; attempt++)
        {
            using var bitmap = new SKBitmap(new SKImageInfo(width, height,
                SKColorType.Rgba8888, SKAlphaType.Premul));
            using (var canvas = new SKCanvas(bitmap))
            {
                canvas.Clear(SKColors.Transparent);
                canvas.Scale((float)width / originalWidth, (float)height / originalHeight);
                ApplyImageOrientation(canvas, codec.EncodedOrigin, info.Width, info.Height);
                canvas.DrawBitmap(decoded, 0, 0, new SKSamplingOptions(SKFilterMode.Linear));
                canvas.Flush();
            }
            using var image = SKImage.FromBitmap(bitmap);
            using var encoded = image.Encode(SKEncodedImageFormat.Png, 100)
                ?? throw new InvalidDataException("Image encoding failed.");
            if (encoded.Size > 0 && encoded.Size <= MaximumImageOutputBytes)
            {
                byte[] bytes = encoded.ToArray();
                var blob = new BlobPart("image/png", new BlobRef(RefKind.Uri,
                    "data:image/png;base64," + Convert.ToBase64String(bytes)));
                return new PreparedImage(blob, originalWidth, originalHeight, width, height,
                    source.Length, bytes.Length);
            }
            if (width == 1 && height == 1)
            {
                break;
            }
            scale *= 0.75;
            width = Math.Max(1, (int)Math.Floor(originalWidth * scale));
            height = Math.Max(1, (int)Math.Floor(originalHeight * scale));
        }
        throw new InvalidDataException("A bounded image could not be encoded after eight size reductions.");
    }

    private static partial void ApplyImageOrientation(SKCanvas canvas, SKEncodedOrigin origin,
        int width, int height)
    {
        // Transforms map encoded coordinates into upright coordinates. Canvas
        // concatenation retains the output-size scaling already applied above.
        var matrix = origin switch
        {
            SKEncodedOrigin.TopLeft => SKMatrix.Identity,
            SKEncodedOrigin.TopRight => new SKMatrix(-1, 0, width, 0, 1, 0, 0, 0, 1),
            SKEncodedOrigin.BottomRight => new SKMatrix(-1, 0, width, 0, -1, height, 0, 0, 1),
            SKEncodedOrigin.BottomLeft => new SKMatrix(1, 0, 0, 0, -1, height, 0, 0, 1),
            SKEncodedOrigin.LeftTop => new SKMatrix(0, 1, 0, 1, 0, 0, 0, 0, 1),
            SKEncodedOrigin.RightTop => new SKMatrix(0, -1, height, 1, 0, 0, 0, 0, 1),
            SKEncodedOrigin.RightBottom => new SKMatrix(0, -1, height, -1, 0, width, 0, 0, 1),
            SKEncodedOrigin.LeftBottom => new SKMatrix(0, 1, 0, -1, 0, width, 0, 0, 1),
            _ => throw new InvalidDataException("Unsupported image orientation.")
        };
        canvas.Concat(matrix);
    }
}
