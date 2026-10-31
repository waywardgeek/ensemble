using System.Collections.Immutable;
using System.Text.Json;
using System.Text.Json.Nodes;
namespace Ensemble;

public sealed partial class Engine
{
    public partial ImmutableArray<Event> Parse(int status, string body)
    {
        return ParseResponse(status, body, out _);
    }

    private partial ImmutableArray<Event> ParseResponse(int status, string body, out UsageIdentity? billing)
    {
        billing = null;
        if (status < 200 || status >= 300)
        {
            return [new Event(EventKind.ErrorOccurred, error: $"Provider HTTP status {status}")];
        }

        try
        {
            using var document = JsonDocument.Parse(body);
            var root = document.RootElement;
            var config = Parent().Configuration();
            // A reply's model identity is more precise than a routing alias.
            var modelField = config.Vendor == Vendor.Gemini ? "modelVersion" : "model";
            var model = GetString(root, modelField, config.ResolvedModel);
            var from = new Provenance(config.Vendor, model, config.Target().Surface);
            var parts = config.Vendor switch
            {
                Vendor.Anthropic => ParseAnthropic(root, from),
                Vendor.OpenAI => ParseOpenAI(root, from),
                _ => ParseGemini(root, from)
            };
            // Reject the whole response before dispatch if identities collide.
            // IDs are scoped to this response; later rounds may reuse them.
            var ids = new HashSet<string>(StringComparer.Ordinal);
            foreach (var call in parts.OfType<ToolCallPart>())
            {
                if (!ids.Add(call.CallID))
                {
                    throw new FormatException("Duplicate tool-call ID");
                }
            }
            var usage = ParseUsage(root, config.Vendor);
            if (usage.Input is null || usage.CacheRead is null || usage.CacheWrite is null || usage.Output is null)
            {
                Parent().Parent().Log("Provider usage is incomplete or invalid; affected counts remain unknown.");
            }

            // Billing must not silently substitute the requested routing alias
            // for a missing actual model or an unknown service tier. Provenance
            // keeps its established fallback independently of this estimate.
            string? tier = null;
            if (root.TryGetProperty("service_tier", out var suppliedTier)
                && suppliedTier.ValueKind == JsonValueKind.String)
            {
                tier = suppliedTier.GetString();
            }
            billing = new UsageIdentity(config.Vendor, GetString(root, modelField, ""), tier);
            return
            [
                new Event(EventKind.ResponseStarted, from: from),
                new Event(EventKind.ResponseEnded, parts: parts, from: from, usage: usage)
            ];
        }
        catch (Exception ex) when (ex is JsonException or FormatException or ArgumentException or InvalidOperationException or KeyNotFoundException or OverflowException)
        {
            return [new Event(EventKind.ErrorOccurred, error: "Invalid provider response: " + ex.GetType().Name)];
        }
    }

    // Use provider IDs when supplied, including numeric zero. Missing identities
    // are local to this response, stable through logging and later projection.
    private static string CallID(JsonElement call, string field, int ordinal, string prefix)
    {
        if (call.TryGetProperty(field, out var value))
        {
            if (value.ValueKind == JsonValueKind.Number)
            {
                return value.GetRawText();
            }
            if (value.ValueKind == JsonValueKind.String && value.GetString()!.Length > 0)
            {
                return value.GetString()!;
            }
        }
        return prefix + ordinal;
    }

    private static string GetString(JsonElement value, string key, string fallback = "")
    {
        if (value.TryGetProperty(key, out var field) && field.ValueKind == JsonValueKind.String)
        {
            return field.GetString()!;
        }
        return fallback;
    }

    // Retain unrecognized call fields as provider-owned material, not new shared
    // properties. They can travel back only with matching provenance.
    private static JsonElement? Remaining(JsonElement value, params string[] known)
    {
        var o = new JsonObject();
        foreach (var field in value.EnumerateObject())
        {
            if (!known.Contains(field.Name, StringComparer.Ordinal))
            {
                o[field.Name] = JsonCodec.Canonical(field.Value);
            }
        }

        return o.Count == 0 ? null : JsonSerializer.SerializeToElement(o);
    }

    private static List<Part> ParseAnthropic(JsonElement root, Provenance from)
    {
        var parts = new List<Part>();
        foreach (var block in root.GetProperty("content").EnumerateArray())
        {
            switch (GetString(block, "type"))
            {
                case "text":
                    parts.Add(new TextPart(block.GetProperty("text").GetString()!));
                    break;
                case "tool_use":
                    var extra = Remaining(block, "type", "id", "name", "input");
                    parts.Add(new ToolCallPart(CallID(block, "id", parts.Count, "anthropic-call-"), from,
                        GetString(block, "name"), block.GetProperty("input"), extra));
                    break;
                default:
                    parts.Add(new OpaquePart(from, block));
                    break;
            }
        }
        return parts;
    }

    private static List<Part> ParseOpenAI(JsonElement root, Provenance from)
    {
        var message = root.GetProperty("choices")[0].GetProperty("message");
        var parts = new List<Part>();
        if (message.TryGetProperty("content", out var content))
        {
            if (content.ValueKind == JsonValueKind.String)
            {
                parts.Add(new TextPart(content.GetString()!));
            }
            else if (content.ValueKind == JsonValueKind.Array)
            {
                foreach (var p in content.EnumerateArray())
                {
                    if (GetString(p, "type") == "text")
                    {
                        parts.Add(new TextPart(p.GetProperty("text").GetString()!));
                    }
                    else
                    {
                        parts.Add(new OpaquePart(from, p));
                    }
                }
            }
            else if (content.ValueKind != JsonValueKind.Null)
            {
                throw new FormatException("Invalid OpenAI content");
            }
        }
        // Compatible Chat Completions endpoints may expose readable reasoning.
        // Keep its native field opaque and provenance-bound in canonical history.
        if (message.TryGetProperty("reasoning_content", out var reasoning)
            && reasoning.ValueKind != JsonValueKind.Null)
        {
            var native = new JsonObject
            {
                ["type"] = "reasoning_content", ["text"] = reasoning.GetString()
            };
            parts.Add(new OpaquePart(from, JsonSerializer.SerializeToElement(native)));
        }
        if (message.TryGetProperty("tool_calls", out var calls))
        {
            foreach (var call in calls.EnumerateArray())
            {
                var function = call.GetProperty("function");
                using var args = JsonDocument.Parse(function.GetProperty("arguments").GetString()!);
                var extra = Remaining(call, "id", "type", "function");
                parts.Add(new ToolCallPart(CallID(call, "id", parts.Count, "openai-call-"), from,
                    GetString(function, "name"), args.RootElement, extra));
            }
        }

        return parts;
    }

    private static List<Part> ParseGemini(JsonElement root, Provenance from)
    {
        var parts = new List<Part>();
        int ordinal = 0;
        var candidates = root.GetProperty("candidates");
        if (candidates.GetArrayLength() == 0)
        {
            throw new FormatException("No Gemini candidate");
        }

        foreach (var p in candidates[0].GetProperty("content").GetProperty("parts").EnumerateArray())
        {
            if (p.TryGetProperty("functionCall", out var call))
            {
                // A response-local ordinal is deterministic; do not hash semantic
                // arguments, which would collapse two identical calls in one reply.
                var id = CallID(call, "id", ordinal, "gemini-call-");
                ordinal++;
                var args = call.TryGetProperty("args", out var a) ? a : JsonSerializer.SerializeToElement(new JsonObject());
                // Keep the signature's original JSON value on this exact call.
                JsonElement? capsule = null;
                if (p.TryGetProperty("thoughtSignature", out var signature))
                {
                    capsule = signature;
                }
                parts.Add(new ToolCallPart(id, from, GetString(call, "name"), args, capsule));
            }
            else if (p.TryGetProperty("text", out var text) && !(p.TryGetProperty("thought", out var thought) && thought.ValueKind == JsonValueKind.True))
            {
                parts.Add(new TextPart(text.GetString()!, from,
                    p.TryGetProperty("thoughtSignature", out var textSignature) ? textSignature : null));
            }
            else
            {
                parts.Add(new OpaquePart(from, p));
            }
        }
        return parts;
    }

    // Absence of an optional count means zero; a supplied invalid count is
    // unknown. Never accept negatives, fractional numbers, or overflowing counts.
    private static long? Count(JsonElement value, string name, bool optional = false)
    {
        if (value.ValueKind != JsonValueKind.Object || !value.TryGetProperty(name, out var count))
        {
            return optional ? 0 : null;
        }

        if (count.ValueKind == JsonValueKind.Number && count.TryGetInt64(out var n) && n >= 0)
        {
            return n;
        }
        return null;
    }

    // Vendors may report inclusive totals; the shared buckets are disjoint.
    // An invalid subtraction is unknown, never a silently clamped zero.
    private static long? Subtract(long? total, params long?[] subsets)
    {
        if (total is not long remaining)
        {
            return null;
        }

        foreach (var subset in subsets)
        {
            if (subset is not long n || n > remaining)
            {
                return null;
            }

            remaining -= n;
        }
        return remaining;
    }

    private static Usage ParseUsage(JsonElement root, Vendor vendor)
    {
        if (vendor == Vendor.Gemini)
        {
            root.TryGetProperty("usageMetadata", out var u);
            var cached = Count(u, "cachedContentTokenCount", true);
            var candidates = new Usage(output: Count(u, "candidatesTokenCount"));
            var thoughts = new Usage(output: Count(u, "thoughtsTokenCount", true));
            var output = candidates.Add(thoughts).Output;
            var input = Subtract(Count(u, "promptTokenCount"), cached);
            return new Usage(input, 0, cached, output);
        }
        root.TryGetProperty("usage", out var usage);
        if (vendor == Vendor.Anthropic)
        {
            return new Usage(
                Count(usage, "input_tokens"),
                Count(usage, "cache_creation_input_tokens", true),
                Count(usage, "cache_read_input_tokens", true),
                Count(usage, "output_tokens"));
        }

        JsonElement details = default;
        if (usage.ValueKind == JsonValueKind.Object)
        {
            usage.TryGetProperty("prompt_tokens_details", out details);
        }
        // Missing detail objects mean no optional cache count was supplied. An
        // explicitly malformed object is not evidence that those counts were zero.
        bool malformed = usage.ValueKind == JsonValueKind.Object && usage.TryGetProperty("prompt_tokens_details", out var supplied) && supplied.ValueKind != JsonValueKind.Object;
        var read = malformed ? null : Count(details, "cached_tokens", true);
        var write = malformed ? null : Count(details, "cache_write_tokens", true);
        return new(Subtract(Count(usage, "prompt_tokens"), write, read), write, read, Count(usage, "completion_tokens"));
    }
}
