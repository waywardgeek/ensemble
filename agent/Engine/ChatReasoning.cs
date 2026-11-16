using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class Engine
{
    // One native message-level envelope, not a Chat content block. Display text
    // is derived; the provider fields remain the authority for compatible replay.
    private static JsonObject? ChatReasoning(JsonObject message)
    {
        var native = new JsonObject { ["type"] = "chat_reasoning" };
        foreach (string field in ChatReasoningFields)
        {
            if (message[field] is not JsonNode value) continue;
            if (field == "reasoning_details")
            {
                foreach (var item in value.AsArray())
                {
                    var detail = item?.AsObject() ?? throw new FormatException("Missing reasoning detail");
                    foreach (string payload in new[] { "text", "summary", "data" })
                    {
                        if (detail[payload] is JsonNode content) _ = content.GetValue<string>();
                    }
                }
            }
            else _ = value.GetValue<string>();
            native[field] = value.DeepClone();
        }
        if (native.Count == 1) return null;
        native["text"] = ChatReasoningText(native);
        return native;
    }

    private static string ChatReasoningText(JsonObject native)
    {
        foreach (string field in new[] { "reasoning_content", "reasoning" })
        {
            if (native[field] is JsonValue text && text.GetValue<string>().Length > 0)
                return text.GetValue<string>();
        }
        if (native["reasoning_details"] is not JsonArray details) return "";
        return string.Concat(details.Select(item =>
        {
            var detail = item!.AsObject();
            return String(detail, "type") switch
            {
                "reasoning.text" => String(detail, "text"),
                "reasoning.summary" => String(detail, "summary"),
                _ => "" // Encrypted data and unknown extensions are replay-only.
            };
        }));
    }

    private static bool ReplayChatReasoning(OpaquePart part, JsonObject message)
    {
        if (part.Data.ValueKind != JsonValueKind.Object) return false;
        string kind = GetString(part.Data, "type");
        if (kind == "reasoning_content")
        {
            // Compatibility with Chat reasoning recorded before the envelope.
            message["reasoning_content"] = part.Data.GetProperty("text").GetString();
            return true;
        }
        if (kind != "chat_reasoning") return false;
        foreach (string field in ChatReasoningFields)
        {
            if (part.Data.TryGetProperty(field, out var value)) message[field] = JsonCodec.Canonical(value);
        }
        return true;
    }

    private static void AccumulateChatReasoning(StreamAssembly state, JsonObject delta, Action<Chunk> emit)
    {
        if (!ChatReasoningFields.Any(field => delta[field] is not null)) return;
        if (!state.ToolIDs.TryGetValue(-2, out int id))
        {
            id = state.NextID++;
            state.ToolIDs.Add(-2, id);
            state.Blocks.Add(id, new JsonObject { ["type"] = "chat_reasoning" });
        }
        var block = state.Blocks[id];
        foreach (string field in new[] { "reasoning_content", "reasoning" })
        {
            if (delta[field] is JsonNode value) block[field] = String(block, field) + value.GetValue<string>();
        }
        if (delta["reasoning_details"] is JsonNode supplied)
        {
            var fragments = supplied.AsArray();
            var details = block["reasoning_details"] as JsonArray;
            if (details is null) block["reasoning_details"] = details = new JsonArray();
            foreach (var value in fragments)
            {
                var fragment = value?.AsObject() ?? throw new FormatException("Missing reasoning fragment");
                JsonObject? target = null;
                if (fragment["index"] is JsonNode index)
                {
                    int ordinal = index.GetValue<int>();
                    if (ordinal < 0) throw new FormatException("Negative reasoning index");
                    target = details.OfType<JsonObject>().SingleOrDefault(item => item["index"]?.GetValue<int>() == ordinal);
                }
                else if (fragment["id"] is JsonNode identity)
                {
                    string key = identity.GetValue<string>();
                    target = details.OfType<JsonObject>().SingleOrDefault(item => item["id"]?.GetValue<string>() == key);
                }
                else
                {
                    // A single anonymous stream can be assembled unambiguously.
                    // Multiple anonymous items cannot safely be correlated.
                    if (fragments.Count != 1 || details.Count > 1)
                        throw new FormatException("Ambiguous reasoning fragments");
                    target = details.FirstOrDefault()?.AsObject();
                }
                if (target is null)
                {
                    target = new JsonObject();
                    details.Add(target);
                }
                foreach (var property in fragment)
                {
                    if (property.Key is "text" or "summary" or "data")
                    {
                        if (property.Value is not null)
                            target[property.Key] = String(target, property.Key) + property.Value.GetValue<string>();
                    }
                    else if (!target.ContainsKey(property.Key)) target[property.Key] = property.Value?.DeepClone();
                    else if (!JsonNode.DeepEquals(target[property.Key], property.Value))
                        throw new FormatException("Conflicting reasoning metadata");
                }
            }
        }
        // Gateways can send the same text in an alias and in structured details.
        // Choose one display source for this response; retain all native fields.
        if (state.ChatReasoningDisplayField is null)
        {
            state.ChatReasoningDisplayField = ChatReasoningFields.FirstOrDefault(field =>
                block[field] is JsonNode value && (field == "reasoning_details"
                    ? ChatReasoningText(new JsonObject { [field] = value.DeepClone() }).Length > 0
                    : value.GetValue<string>().Length > 0));
        }
        string? selected = state.ChatReasoningDisplayField;
        if (selected is not null && block[selected] is JsonNode display)
        {
            string text = selected == "reasoning_details"
                ? ChatReasoningText(new JsonObject { [selected] = display.DeepClone() })
                : display.GetValue<string>();
            if (!text.StartsWith(state.ChatReasoningDisplayed, StringComparison.Ordinal))
                throw new FormatException("Reasoning display changed its prefix");
            string suffix = text[state.ChatReasoningDisplayed.Length..];
            state.ChatReasoningDisplayed = text;
            if (suffix.Length > 0) emit(new Chunk(id, ChunkKind.Thinking, Text: suffix));
        }
    }
}
