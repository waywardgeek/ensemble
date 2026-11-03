using System.Security.Cryptography;
using System.Text.Json;
using System.Text.Json.Nodes;
using System.Text.RegularExpressions;

namespace Ensemble;

internal static partial class DiagnosticMedia
{
    internal static partial string Project(string text)
    {
        try
        {
            string trimmed = text.TrimStart();
            if (trimmed == "[DONE]") return text;
            if (trimmed.StartsWith('{') || trimmed.StartsWith('[') || trimmed.StartsWith('"'))
            {
                // Parse a new tree. The request string and its media are never
                // modified, and invalid structured data is not dumped raw.
                JsonNode? node = JsonNode.Parse(text);
                if (Text(node) is string rootText) return JsonSerializer.Serialize(ProjectUrls(rootText));
                Visit(node, null);
                return node?.ToJsonString() ?? text;
            }
            return ProjectUrls(text);
        }
        catch (Exception error) when (error is JsonException or ArgumentException or InvalidOperationException)
        {
            return "[diagnostic structured payload omitted: invalid JSON]";
        }
        catch (RegexMatchTimeoutException)
        {
            return "[diagnostic payload omitted: media projection limit]";
        }
    }

    private static partial string? Text(JsonNode? node) =>
        node is JsonValue value && value.TryGetValue<string>(out var text) ? text : null;

    private static partial void Visit(JsonNode? node, string? name)
    {
        if (node is JsonObject obj)
        {
            string? mime = Text(obj["media_type"]) ?? Text(obj["mimeType"]) ?? Text(obj["mime_type"]) ?? Text(obj["mime"]);
            if (mime is null && (name is "input_audio" or "audio")) mime = "audio/" + (Text(obj["format"]) ?? "unknown");
            bool media = mime is not null && (mime.StartsWith("image/", StringComparison.OrdinalIgnoreCase)
                || mime.StartsWith("audio/", StringComparison.OrdinalIgnoreCase));
            if (media && obj.TryGetPropertyValue("data", out var data))
            {
                obj["data"] = Text(data) is string encoded ? Summary(mime!, encoded)
                    : "[media omitted; non-string payload]";
            }
            if (Text(obj["type"]) is "response.audio.delta" or "response.output_audio.delta"
                && Text(obj["delta"]) is string audio)
                obj["delta"] = Summary("audio/unknown", audio);
            if (Text(obj["type"]) == "image_generation_call" && Text(obj["result"]) is string image)
                obj["result"] = Summary("image/unknown", image);
            foreach (var property in obj.ToArray())
            {
                if (Text(property.Value) is string text)
                {
                    if (text.Contains("data:", StringComparison.OrdinalIgnoreCase))
                        obj[property.Key] = ProjectUrls(text);
                }
                else Visit(property.Value, property.Key);
            }
        }
        else if (node is JsonArray array)
        {
            for (int index = 0; index < array.Count; index++)
            {
                if (Text(array[index]) is string text)
                {
                    if (text.Contains("data:", StringComparison.OrdinalIgnoreCase)) array[index] = ProjectUrls(text);
                }
                else Visit(array[index], name);
            }
        }
    }

    private static partial string ProjectUrls(string text)
    {
        if (!text.Contains("data:", StringComparison.OrdinalIgnoreCase)) return text;
        if (text.StartsWith("data:image/", StringComparison.OrdinalIgnoreCase)
            || text.StartsWith("data:audio/", StringComparison.OrdinalIgnoreCase))
        {
            int separator = text.IndexOf(";base64,", StringComparison.OrdinalIgnoreCase);
            if (separator >= 0) return Summary(text[5..separator], text[(separator + 8)..]);
        }
        return DataUrls.Replace(text, match => Summary(match.Groups["mime"].Value, match.Groups["data"].Value));
    }

    private static partial string Summary(string mime, string encoded)
    {
        if (mime.Length > 128) mime = "unrecognized-media-type";
        string prefix = $"[media omitted; mime={mime}; base64_chars={encoded.Length}";
        using var hash = IncrementalHash.CreateHash(HashAlgorithmName.SHA256);
        Span<byte> bytes = stackalloc byte[3072];
        long decodedBytes = 0;
        for (int offset = 0; offset < encoded.Length; offset += 4096)
        {
            int length = Math.Min(4096, encoded.Length - offset);
            var chunk = encoded.AsSpan(offset, length);
            if ((offset + length < encoded.Length && chunk.Contains('='))
                || !Convert.TryFromBase64Chars(chunk, bytes, out int written))
            {
                return prefix + "; decoded_bytes=unavailable; sha256=unavailable]";
            }
            decodedBytes += written;
            hash.AppendData(bytes[..written]);
        }
        return prefix + $"; decoded_bytes={decodedBytes}; sha256={Convert.ToHexStringLower(hash.GetHashAndReset())}]";
    }
}
