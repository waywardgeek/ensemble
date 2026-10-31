using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

// MCP content crosses the boundary once, into the same values used by local
// History and provider renderers. Never flatten an image into diagnostic text.
public static partial class McpContent
{
    public static partial ToolResultPart Decode(string callID, JsonElement result)
    {
        if (result.ValueKind != JsonValueKind.Object
            || !result.TryGetProperty("content", out var content)
            || content.ValueKind != JsonValueKind.Array)
        {
            throw new InvalidDataException("MCP result needs a content array.");
        }
        bool isError = false;
        if (result.TryGetProperty("isError", out var error))
        {
            if (error.ValueKind is not (JsonValueKind.True or JsonValueKind.False))
            {
                throw new InvalidDataException("MCP isError must be Boolean.");
            }
            isError = error.GetBoolean();
        }
        var parts = new List<Part>();
        foreach (var item in content.EnumerateArray())
        {
            string type = RequiredString(item, "type");
            if (type == "text")
            {
                parts.Add(new TextPart(RequiredString(item, "text")));
                continue;
            }
            if (type is "image" or "audio")
            {
                string mime = RequiredString(item, "mimeType");
                string data = RequiredString(item, "data");
                ValidateMedia(type, mime, data);
                parts.Add(new BlobPart(mime, new BlobRef(RefKind.Uri, $"data:{mime};base64,{data}")));
                continue;
            }
            throw new InvalidDataException("Unsupported MCP content block type.");
        }
        return new ToolResultPart(callID, parts, isError);
    }

    public static partial JsonElement Encode(ToolResultPart result)
    {
        var content = new JsonArray();
        foreach (var part in result.Parts)
        {
            if (part is TextPart text)
            {
                content.Add(new JsonObject { ["type"] = "text", ["text"] = text.Text });
                continue;
            }
            if (part is BlobPart blob && (blob.Mime.StartsWith("image/", StringComparison.Ordinal)
                || blob.Mime.StartsWith("audio/", StringComparison.Ordinal)))
            {
                // Reverse calls may return already resolved bytes, not an
                // invitation for the transport to read arbitrary local paths.
                string type = blob.Mime.StartsWith("image/", StringComparison.Ordinal) ? "image" : "audio";
                string prefix = $"data:{blob.Mime};base64,";
                if (blob.Ref.Kind != RefKind.Uri || !blob.Ref.Locator.StartsWith(prefix, StringComparison.Ordinal))
                {
                    throw new InvalidDataException("MCP media requires resolved base64 data.");
                }
                string data = blob.Ref.Locator[prefix.Length..];
                ValidateMedia(type, blob.Mime, data);
                content.Add(new JsonObject
                {
                    ["type"] = type, ["mimeType"] = blob.Mime, ["data"] = data
                });
                continue;
            }
            throw new InvalidDataException("Unsupported MCP result part.");
        }
        return JsonSerializer.SerializeToElement(new JsonObject
        {
            ["content"] = content, ["isError"] = result.IsError
        });
    }

    private static string RequiredString(JsonElement value, string name)
    {
        if (value.ValueKind != JsonValueKind.Object
            || !value.TryGetProperty(name, out var field)
            || field.ValueKind != JsonValueKind.String)
        {
            throw new InvalidDataException("Invalid MCP content field.");
        }
        return field.GetString()!;
    }

    private static void ValidateMedia(string type, string mime, string data)
    {
        if (!mime.StartsWith(type + "/", StringComparison.Ordinal)
            || mime.Contains(';') || mime.Contains(',') || mime.Any(char.IsWhiteSpace))
        {
            throw new InvalidDataException("Invalid MCP media type.");
        }
        try
        {
            Convert.FromBase64String(data);
        }
        catch (FormatException error)
        {
            throw new InvalidDataException("Invalid MCP base64 data.", error);
        }
    }
}
