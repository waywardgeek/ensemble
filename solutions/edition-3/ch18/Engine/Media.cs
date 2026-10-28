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
        foreach (var row in modelCatalog)
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
    private static void CheckMedia(IEnumerable<Part> parts, Configuration config, bool inToolResult = false)
    {
        foreach (var part in parts)
        {
            if (part is BlobPart blob)
            {
                Media kind = MediaKind(blob.Mime);
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
                if (inToolResult && config.Vendor != Vendor.Anthropic)
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

    private static JsonObject AnthropicBlob(BlobPart blob)
    {
        string encoded = Convert.ToBase64String(blob.Ref.Data());
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

    private static JsonObject OpenAIBlob(BlobPart blob)
    {
        string encoded = Convert.ToBase64String(blob.Ref.Data());
        string uri = "data:" + blob.Mime + ";base64," + encoded;
        return new JsonObject
        {
            ["type"] = "image_url",
            ["image_url"] = new JsonObject { ["url"] = uri }
        };
    }

    private static JsonObject GeminiBlob(BlobPart blob)
    {
        string encoded = Convert.ToBase64String(blob.Ref.Data());
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
