using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class Engine
{
    public static partial bool LookupModel(Vendor vendor, string model, out ModelFeatures features)
    {
        foreach (var row in modelCatalog)
        {
            if (row.Vendor == vendor && row.Name == model)
            {
                features = row.Features;
                return true;
            }
        }
        features = new ModelFeatures(Media.None);
        return false;
    }

    public static partial System.Collections.Immutable.ImmutableArray<string> ModelChoices(Configuration configuration)
    {
        var choices = System.Collections.Immutable.ImmutableArray.CreateBuilder<string>();
        choices.Add(configuration.Model);
        foreach (var row in RecommendedModels)
        {
            if (row.Vendor != configuration.Vendor || row.Name == configuration.Model)
            {
                continue;
            }
            try
            {
                // Use the same validator as SetModelAsync. The candidate is a
                // disposable value; constructing it cannot change the Agent.
                _ = configuration.WithModel(row.Name);
                choices.Add(row.Name);
            }
            catch (ArgumentException)
            {
                // A known row can be incompatible with this fixed wire surface.
            }
        }
        return choices.ToImmutable();
    }

    private static Media MediaKind(string mime)
    {
        if (mime.StartsWith("image/", StringComparison.Ordinal))
        {
            return Media.Image;
        }
        if (mime.StartsWith("audio/", StringComparison.Ordinal))
        {
            return Media.Audio;
        }
        if (mime.StartsWith("video/", StringComparison.Ordinal))
        {
            return Media.Video;
        }
        return mime == "application/pdf" ? Media.Document : Media.None;
    }

    // Walk nested tool results too. A supported model does not make an API
    // surface support every placement; refusing is safer than silently dropping.
    private static void CheckMedia(IEnumerable<Part> parts, Configuration config, bool inToolResult = false,
        Actor actor = Actor.Human)
    {
        foreach (var part in parts)
        {
            if (part is BlobPart blob)
            {
                Media kind = MediaKind(blob.Mime);
                // Responses accepts human images and typed image tool outputs,
                // not arbitrary modalities or assistant/system images.
                if (config.Surface == Surface.Responses
                    && (kind != Media.Image || (!inToolResult && actor != Actor.Human)))
                {
                    throw new NotSupportedException("Responses supports images in human entries and tool results only");
                }
                string category = kind == Media.None ? blob.Mime : kind.ToString().ToLowerInvariant();
                bool known = LookupModel(config.Vendor, config.Model, out var features);
                if (!known || kind == Media.None || (features.Media & kind) == 0)
                {
                    throw new NotSupportedException($"Model '{config.Model}' does not support {category} media.");
                }
                bool inline = blob.Ref.Kind == RefKind.Uri
                    && blob.Ref.Locator.StartsWith($"data:{blob.Mime};base64,", StringComparison.Ordinal);
                if (blob.Ref.Kind != RefKind.Path && !inline)
                {
                    throw new NotSupportedException($"Model '{config.Model}' requires local paths or matching base64 data URIs for {category} media from {blob.Ref.Kind}.");
                }
                if (inToolResult && config.Vendor != Vendor.Anthropic && config.Surface != Surface.Responses)
                {
                    throw new NotSupportedException($"Model '{config.Model}' cannot render {category} media inside this adapter's tool-result content.");
                }
            }
            else if (part is ToolResultPart result)
            {
                CheckMedia(result.Parts, config, inToolResult: true);
            }
        }
    }

    private partial byte[] ReadMedia(BlobPart blob)
    {
        // Project refuses unsupported capabilities/references before any reads.
        // Non-image modalities retain their existing contract.
        var reference = blob.Ref;
        if (MediaKind(blob.Mime) != Media.Image)
        {
            return reference.Kind == RefKind.Path
                ? Parent().Sandbox().ReadFile(reference.Locator) : reference.Data();
        }
        byte[] bytes;
        if (reference.Kind == RefKind.Path)
        {
            using var input = Parent().Sandbox().OpenRegularRead(reference.Locator);
            long length = input.Length;
            if (length <= 0 || length > MaximumImageBytes)
            {
                throw new InvalidDataException("Image must contain 1 to 8388608 bytes.");
            }
            bytes = new byte[(int)length];
            input.ReadExactly(bytes);
            // A growing file cannot bypass the allocation bound. A truncated
            // file fails ReadExactly; either failure disposes the owned stream.
            if (input.ReadByte() != -1)
            {
                throw new InvalidDataException("Image changed size while being read.");
            }
        }
        else
        {
            string prefix = $"data:{blob.Mime};base64,";
            if (reference.Kind != RefKind.Uri || !reference.Locator.StartsWith(prefix, StringComparison.Ordinal))
            {
                throw new NotSupportedException("Images require local paths or matching base64 data URIs.");
            }
            int encodedLength = reference.Locator.Length - prefix.Length;
            int maximumEncodedLength = ((MaximumImageBytes + 2) / 3) * 4;
            if (encodedLength <= 0 || encodedLength > maximumEncodedLength)
            {
                throw new InvalidDataException("Encoded image exceeds the image byte limit or is empty.");
            }
            int padding = reference.Locator.EndsWith("==", StringComparison.Ordinal) ? 2
                : reference.Locator.EndsWith('=') ? 1 : 0;
            int decodedCeiling = ((encodedLength + 3) / 4) * 3 - padding;
            if (decodedCeiling > MaximumImageBytes)
            {
                throw new InvalidDataException("Decoded image exceeds the image byte limit.");
            }
            bytes = Convert.FromBase64String(reference.Locator[prefix.Length..]);
            if (bytes.Length == 0 || bytes.Length > MaximumImageBytes)
            {
                throw new InvalidDataException("Image must contain 1 to 8388608 bytes.");
            }
        }
        return bytes;
    }

    private JsonObject AnthropicBlob(BlobPart blob)
    {
        string encoded = Convert.ToBase64String(ReadMedia(blob));
        return new JsonObject
        {
            ["type"] = "image",
            ["source"] = new JsonObject
            {
                ["type"] = "base64",
                ["media_type"] = blob.Mime,
                ["data"] = encoded
            }
        };
    }

    private JsonObject OpenAIBlob(BlobPart blob)
    {
        string encoded = Convert.ToBase64String(ReadMedia(blob));
        string uri = "data:" + blob.Mime + ";base64," + encoded;
        return new JsonObject
        {
            ["type"] = "image_url",
            ["image_url"] = new JsonObject { ["url"] = uri }
        };
    }

    private JsonObject GeminiBlob(BlobPart blob)
    {
        string encoded = Convert.ToBase64String(ReadMedia(blob));
        return new JsonObject
        {
            ["inlineData"] = new JsonObject
            {
                ["mimeType"] = blob.Mime,
                ["data"] = encoded
            }
        };
    }
}
