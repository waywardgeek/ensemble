using System.Text;
using System.Text.Json;

namespace Ensemble;

public sealed partial class VoiceTranscriber
{
    private static partial VoiceTranscript ParseTranscript(ReadOnlyMemory<byte> body,
        TimeSpan inputDuration, string apiKey)
    {
        using var document = JsonDocument.Parse(body, new JsonDocumentOptions { MaxDepth = 16 });
        JsonElement root = document.RootElement;
        RejectDuplicateProperties(root);
        if (root.ValueKind != JsonValueKind.Object
            || !root.TryGetProperty("text", out var value) || value.ValueKind != JsonValueKind.String)
            throw new InvalidDataException("Voice transcription did not contain text.");
        string text = value.GetString()!.Trim();
        if (text.Length == 0 || text.Contains('\0')
            || new UTF8Encoding(false, true).GetByteCount(text) > MaximumTranscriptBytes)
            throw new InvalidDataException("Voice transcription text is empty or exceeds its limit.");
        if (text.Contains(apiKey, StringComparison.Ordinal))
            throw new InvalidDataException("Voice transcription contained credential material.");

        VoiceTranscriptionUsage? usage = null;
        if (root.TryGetProperty("usage", out var reported) && reported.ValueKind != JsonValueKind.Null)
        {
            if (reported.ValueKind != JsonValueKind.Object
                || !reported.TryGetProperty("type", out var kind) || kind.ValueKind != JsonValueKind.String)
                throw new InvalidDataException("Invalid voice transcription usage.");
            switch (kind.GetString())
            {
                case "tokens":
                    long input = RequiredCount(reported, "input_tokens");
                    long output = RequiredCount(reported, "output_tokens");
                    long total = RequiredCount(reported, "total_tokens");
                    if (input > long.MaxValue - output || input + output != total)
                        throw new InvalidDataException("Inconsistent voice token usage.");
                    long? audio = null;
                    long? textTokens = null;
                    if (reported.TryGetProperty("input_token_details", out var details)
                        && details.ValueKind != JsonValueKind.Null)
                    {
                        if (details.ValueKind != JsonValueKind.Object)
                            throw new InvalidDataException("Invalid voice token details.");
                        audio = OptionalCount(details, "audio_tokens");
                        textTokens = OptionalCount(details, "text_tokens");
                        if (audio > input || textTokens > input
                            || (audio is long a && textTokens is long t && (a > input - t || a + t != input)))
                            throw new InvalidDataException("Inconsistent voice input token details.");
                    }
                    usage = new VoiceTokenUsage(input, output, total, audio, textTokens);
                    break;
                case "duration":
                    if (!reported.TryGetProperty("seconds", out var seconds)
                        || seconds.ValueKind != JsonValueKind.Number || !seconds.TryGetDouble(out double duration)
                        || !double.IsFinite(duration) || duration < 0)
                        throw new InvalidDataException("Invalid reported voice duration.");
                    usage = new VoiceDurationUsage(duration);
                    break;
                default:
                    throw new InvalidDataException("Unsupported voice transcription usage type.");
            }
        }
        return new VoiceTranscript(text, inputDuration, usage);
    }

    private static long RequiredCount(JsonElement value, string name) => OptionalCount(value, name)
        ?? throw new InvalidDataException("Required voice usage count is missing.");

    private static long? OptionalCount(JsonElement value, string name)
    {
        if (!value.TryGetProperty(name, out var count) || count.ValueKind == JsonValueKind.Null) return null;
        if (count.ValueKind != JsonValueKind.Number || !count.TryGetInt64(out long result) || result < 0)
            throw new InvalidDataException("Invalid voice usage count.");
        return result;
    }

    private static void RejectDuplicateProperties(JsonElement value)
    {
        if (value.ValueKind == JsonValueKind.Object)
        {
            var names = new HashSet<string>(StringComparer.Ordinal);
            foreach (var property in value.EnumerateObject())
            {
                if (!names.Add(property.Name))
                    throw new InvalidDataException("Duplicate voice transcription response property.");
                RejectDuplicateProperties(property.Value);
            }
        }
        else if (value.ValueKind == JsonValueKind.Array)
        {
            foreach (var item in value.EnumerateArray()) RejectDuplicateProperties(item);
        }
    }
}
