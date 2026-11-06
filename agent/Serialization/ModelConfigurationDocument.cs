using System.Collections.Immutable;
using System.Text.Json;
using System.Text;

namespace Ensemble;

internal static partial class ModelConfigurationDocument
{
    internal static partial ImmutableArray<ModelVendorDefinition> Read(ReadOnlyMemory<byte> utf8)
    {
        if (utf8.Length == 0 || utf8.Length > MaximumBytes)
            throw new InvalidDataException("Model configuration must contain at most 1 MiB of JSON.");
        try
        {
            _ = new UTF8Encoding(false, true).GetCharCount(utf8.Span);
            using var document = JsonDocument.Parse(utf8, new JsonDocumentOptions { MaxDepth = 16 });
            JsonElement root = document.RootElement;
            Fields(root, ["version", "vendors"]);
            if (!root.GetProperty("version").TryGetInt32(out int version) || version != 1)
                throw new InvalidDataException("Unsupported model configuration version.");
            var vendors = ImmutableArray.CreateBuilder<ModelVendorDefinition>();
            int endpointCount = 0;
            int modelCount = 0;
            foreach (var vendor in Rows(root.GetProperty("vendors"), 32))
            {
                Fields(vendor, ["id", "label", "endpoints"]);
                var endpoints = ImmutableArray.CreateBuilder<ModelEndpointDefinition>();
                foreach (var endpoint in Rows(vendor.GetProperty("endpoints"), 64))
                {
                    if (++endpointCount > 64) throw new InvalidDataException("Too many configured endpoints.");
                    Fields(endpoint, ["id", "address", "surface", "authentication", "models"]);
                    Surface surface = Text(endpoint, "surface") switch
                    {
                        "messages" => Surface.Messages,
                        "chat_completions" => Surface.ChatCompletions,
                        "responses" => Surface.Responses,
                        "generate_content" => Surface.GenerateContent,
                        _ => throw new InvalidDataException("Unknown configured wire surface.")
                    };
                    var models = ImmutableArray.CreateBuilder<ModelDefinition>();
                    foreach (var model in Rows(endpoint.GetProperty("models"), 256))
                    {
                        if (++modelCount > 256) throw new InvalidDataException("Too many configured models.");
                        Fields(model, ["id", "name", "label", "features"], ["serving_context_tokens", "request_policy"]);
                        models.Add(new ModelDefinition(Text(model, "id"), Text(model, "name"),
                            Text(model, "label"), Features(model.GetProperty("features")),
                            Tokens(model, "serving_context_tokens"),
                            model.TryGetProperty("request_policy", out var requestPolicy) ? Policy(requestPolicy) : null));
                    }
                    endpoints.Add(new ModelEndpointDefinition(Text(endpoint, "id"), Text(endpoint, "address"),
                        surface, Text(endpoint, "authentication"), models.ToImmutable()));
                }
                vendors.Add(new ModelVendorDefinition(Text(vendor, "id"), Text(vendor, "label"),
                    endpoints.ToImmutable()));
            }
            return vendors.ToImmutable();
        }
        catch (Exception error) when (error is JsonException or InvalidOperationException or FormatException or DecoderFallbackException)
        {
            // Do not echo potentially sensitive invalid document values or parser snippets.
            throw new InvalidDataException("Invalid model configuration document.");
        }

        static void Fields(JsonElement value, string[] required, string[]? optional = null)
        {
            if (value.ValueKind != JsonValueKind.Object)
                throw new InvalidDataException("Expected a model configuration object.");
            var names = new HashSet<string>(StringComparer.Ordinal);
            foreach (var property in value.EnumerateObject())
            {
                if (!names.Add(property.Name) || (!required.Contains(property.Name)
                    && !(optional?.Contains(property.Name) ?? false)))
                    throw new InvalidDataException("Duplicate or unknown model configuration field.");
            }
            if (required.Any(name => !names.Contains(name)))
                throw new InvalidDataException("Missing model configuration field.");
        }

        static JsonElement.ArrayEnumerator Rows(JsonElement value, int maximum)
        {
            if (value.ValueKind != JsonValueKind.Array || value.GetArrayLength() > maximum)
                throw new InvalidDataException("Invalid model configuration array.");
            return value.EnumerateArray();
        }

        static string Text(JsonElement value, string name)
        {
            JsonElement field = value.GetProperty(name);
            if (field.ValueKind != JsonValueKind.String || string.IsNullOrWhiteSpace(field.GetString()))
                throw new InvalidDataException("Expected nonempty model configuration text.");
            return field.GetString()!;
        }

        static bool Flag(JsonElement value, string name)
        {
            if (!value.TryGetProperty(name, out var field)) return false;
            if (field.ValueKind is not (JsonValueKind.True or JsonValueKind.False))
                throw new InvalidDataException("Expected a model capability boolean.");
            return field.GetBoolean();
        }

        static long? Tokens(JsonElement value, string name)
        {
            if (!value.TryGetProperty(name, out var field) || field.ValueKind == JsonValueKind.Null) return null;
            if (field.ValueKind != JsonValueKind.Number || !field.TryGetInt64(out long count) || count <= 0)
                throw new InvalidDataException("Context capacity must be positive whole tokens or null.");
            return count;
        }

        static ImmutableArray<string> Strings(JsonElement value, string name)
        {
            if (!value.TryGetProperty(name, out var field)) return [];
            var values = ImmutableArray.CreateBuilder<string>();
            var unique = new HashSet<string>(StringComparer.Ordinal);
            foreach (var item in Rows(field, 16))
            {
                if (item.ValueKind != JsonValueKind.String) throw new InvalidDataException("Invalid capability name.");
                string text = item.GetString()!;
                if (string.IsNullOrWhiteSpace(text) || text.Length > 64 || text.Any(char.IsControl) || !unique.Add(text))
                    throw new InvalidDataException("Invalid or duplicate capability name.");
                values.Add(text);
            }
            return values.ToImmutable();
        }

        static ModelRequestPolicy Policy(JsonElement value)
        {
            Fields(value, [], ["disable_streaming", "require_tool_explanation", "deferred_tools",
                "stub_tool_results", "cache_breakpoints", "prompt_cache_key", "reasoning_effort", "reasoning_summary"]);
            return new ModelRequestPolicy(Flag(value, "disable_streaming"), Flag(value, "require_tool_explanation"),
                Flag(value, "deferred_tools"), Flag(value, "stub_tool_results"), Flag(value, "cache_breakpoints"),
                Option(value, "prompt_cache_key"), Option(value, "reasoning_effort"), Option(value, "reasoning_summary"));
        }

        static string Option(JsonElement value, string name)
        {
            if (!value.TryGetProperty(name, out var field)) return "";
            if (field.ValueKind != JsonValueKind.String)
                throw new InvalidDataException("Expected request policy text.");
            return field.GetString()!;
        }

        static ModelFeatures Features(JsonElement value)
        {
            Fields(value, ["stream"], ["media", "reasoning_efforts", "reasoning_summaries",
                "require_tool_explanation", "responses_only", "context_window", "function_response_images"]);
            Media media = Media.None;
            foreach (string item in Strings(value, "media"))
            {
                media |= item switch
                {
                    "image" => Media.Image,
                    "audio" => Media.Audio,
                    "video" => Media.Video,
                    "document" => Media.Document,
                    _ => throw new InvalidDataException("Unknown configured media capability.")
                };
            }
            return new ModelFeatures(media, Flag(value, "stream"), Strings(value, "reasoning_efforts"),
                Strings(value, "reasoning_summaries"), Flag(value, "require_tool_explanation"),
                Flag(value, "responses_only"), Tokens(value, "context_window"), Flag(value, "function_response_images"));
        }
    }
}
