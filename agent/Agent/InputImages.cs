using System.Collections.Immutable;

namespace Ensemble;

public sealed partial record AgentInput
{
    internal partial ImmutableArray<Part> Parts()
    {
        var parts = ImmutableArray.CreateBuilder<Part>();
        parts.Add(new TextPart(Text));
        if (!Images.IsDefaultOrEmpty) parts.AddRange(Images);
        return parts.ToImmutable();
    }
}

public sealed partial class Agent
{
    // Hosts prepare images once before this admission boundary. Only immutable
    // inline derivatives may enter live inputs: never a late file lookup.
    private partial void ValidateInputImages(AgentInput input)
    {
        if (input.Images.IsDefaultOrEmpty) return;
        if (input.Kind is not (InputKind.Message or InputKind.Hint or InputKind.Chat))
        {
            throw new ArgumentException("Images belong only to human messages or hints.");
        }
        if (input.Images.Length > MaximumInputImages)
        {
            throw new ArgumentException($"An input accepts at most {MaximumInputImages} images.");
        }
        const string prefix = "data:image/png;base64,";
        long total = 0;
        foreach (BlobPart image in input.Images)
        {
            if (image is null || image.Mime != "image/png" || image.Ref.Kind != RefKind.Uri
                || !image.Ref.Locator.StartsWith(prefix, StringComparison.Ordinal))
            {
                throw new ArgumentException("Interactive images must be prepared inline PNGs.");
            }
            string locator = image.Ref.Locator;
            int length = locator.Length - prefix.Length;
            int maximumEncoded = ((MaximumImageOutputBytes + 2) / 3) * 4;
            if (length == 0 || length > maximumEncoded || length % 4 != 0)
            {
                throw new ArgumentException("Interactive image payload is empty, invalid or too large.");
            }
            int padding = locator.EndsWith("==", StringComparison.Ordinal) ? 2
                : locator.EndsWith('=') ? 1 : 0;
            int bytes = length / 4 * 3 - padding;
            total += bytes;
            if (bytes > MaximumImageOutputBytes || total > MaximumInputImageBytes)
            {
                throw new ArgumentException("Interactive images exceed the per-image or combined byte limit.");
            }
            // Decode only bounded base64 here, not the raster again. Pixel and
            // format validation belongs to PrepareImage at the host boundary.
            try
            {
                byte[] data = Convert.FromBase64String(locator[prefix.Length..]);
                if (data.Length == 0 || data.Length != bytes)
                {
                    throw new ArgumentException("Interactive image base64 must be canonical and nonempty.");
                }
            }
            catch (FormatException error)
            {
                throw new ArgumentException("Invalid interactive image base64.", error);
            }
        }
        try
        {
            Engine().ValidateMedia(input.Parts(), Actor.Human);
        }
        catch (NotSupportedException error)
        {
            throw new ArgumentException(error.Message, error);
        }
    }
}
