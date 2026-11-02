using Ensemble;
using System.Collections.Immutable;
using System.Text;
using System.Text.Json;

namespace EnsembleGui;

public static partial class Gui
{
    private static partial (AgentInput Input, ImmutableArray<PreparedImage> Images) PrepareImageInput(Agent selected, JsonElement command, CancellationToken cancellation)
    {
        string text = command.TryGetProperty("text", out var textValue) ? textValue.GetString() ?? "" : "";
        if (Encoding.UTF8.GetByteCount(text) > 32 * 1024)
            throw new InvalidDataException("Image message text exceeds 32 KiB.");
        var images = command.GetProperty("images");
        if (images.ValueKind != JsonValueKind.Array || images.GetArrayLength() is < 1 or > Agent.MaximumInputImages)
            throw new InvalidDataException("An image message requires one to four images.");
        var sources = new List<string>();
        var parts = new List<Part>();
        long total = 0;
        foreach (var image in images.EnumerateArray())
        {
            string mime = image.GetProperty("mime").GetString() ?? "";
            if (mime is not ("image/png" or "image/jpeg" or "image/webp" or "image/gif"))
                throw new InvalidDataException("Unsupported source image format.");
            string data = image.GetProperty("data").GetString() ?? "";
            if (data.Length == 0 || data.Length % 4 != 0)
                throw new InvalidDataException("Image payload must be unspaced base64.");
            long bytes = (long)data.Length / 4 * 3;
            if (data.EndsWith("==", StringComparison.Ordinal))
            {
                bytes -= 2;
            }
            else if (data.EndsWith('='))
            {
                bytes--;
            }
            total += bytes;
            if (total > Agent.MaximumImageSourceBytes)
                throw new InvalidDataException("Combined image sources exceed 64 MiB.");
            sources.Add(data);
            parts.Add(new BlobPart(mime, new BlobRef(RefKind.Uri, $"data:{mime};base64,{data}")));
        }
        // Refuse provider capabilities before raster decoding. The same
        // renderer validator remains authoritative for admitted image parts.
        selected.Engine().ValidateMedia(parts, Actor.Human);
        var prepared = ImmutableArray.CreateBuilder<PreparedImage>();
        int outputBytes = 0;
        foreach (string source in sources)
        {
            cancellation.ThrowIfCancellationRequested();
            if (source.Any(char.IsWhiteSpace))
                throw new InvalidDataException("Image payload must be unspaced base64.");
            byte[] bytes = Convert.FromBase64String(source);
            var image = selected.PrepareImage(bytes);
            outputBytes += image.OutputBytes;
            if (outputBytes > Agent.MaximumInputImageBytes)
                throw new InvalidDataException("Prepared images exceed the combined 16 MiB input limit.");
            cancellation.ThrowIfCancellationRequested();
            prepared.Add(image);
        }
        var imagesForInput = prepared.Select(image => image.Image).ToImmutableArray();
        return (new AgentInput(InputKind.Message, text, imagesForInput), prepared.ToImmutable());
    }

    private static partial string TraceProjection(string json)
    {
        using var document = JsonDocument.Parse(json);
        bool ContainsImage(JsonElement value)
        {
            if (value.ValueKind == JsonValueKind.Object)
            {
                foreach (var field in value.EnumerateObject())
                {
                    bool imageMime = field.Name is "mime" or "mimeType" && field.Value.ValueKind == JsonValueKind.String
                        && (field.Value.GetString() ?? "").StartsWith("image/", StringComparison.OrdinalIgnoreCase);
                    // MCP payload privacy must not depend on valid MIME/type
                    // metadata: malformed image replies are traced before refusal.
                    bool opaqueData = field.Name == "data" && field.Value.ValueKind == JsonValueKind.String;
                    if (field.Name is "images" or "audio" || imageMime || opaqueData)
                    {
                        return true;
                    }
                    if (ContainsImage(field.Value))
                    {
                        return true;
                    }
                }
            }
            else if (value.ValueKind == JsonValueKind.Array)
            {
                foreach (var item in value.EnumerateArray())
                {
                    if (ContainsImage(item))
                    {
                        return true;
                    }
                }
            }
            else if (value.ValueKind == JsonValueKind.String
                && (value.GetString() ?? "").Contains("data:image/", StringComparison.OrdinalIgnoreCase))
            {
                return true;
            }
            return false;
        }
        if (!ContainsImage(document.RootElement)) return json;
        string type = document.RootElement.TryGetProperty("type", out var value) && value.ValueKind == JsonValueKind.String
            ? value.GetString() ?? "" : "";
        // Do not persist any of the frame, including nested History/tool media.
        return JsonSerializer.Serialize(new { type, media_redacted = true, utf8_bytes = Encoding.UTF8.GetByteCount(json) });
    }
}
