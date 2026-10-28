using System.Collections.Immutable;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class Engine
{
    public partial async Task<ParsedResponse> ParseAsync(HttpResponseMessage response,
        Action<Chunk>? onChunk, CancellationToken cancellation)
    {
        int status = (int)response.StatusCode;
        if (!response.IsSuccessStatusCode)
        {
            // Do not decode or retain a remote error body as a successful reply.
            return new ParsedResponse(Parse(status, ""), []);
        }
        bool streaming = string.Equals(response.Content.Headers.ContentType?.MediaType,
            "text/event-stream", StringComparison.OrdinalIgnoreCase);
        if (!streaming)
        {
            string body = await response.Content.ReadAsStringAsync(cancellation);
            Parent().Parent().Trace(Parent(), "response", body);
            var events = ParseResponse(status, body, out var billing);
            var ended = events.FirstOrDefault(fact => fact.Kind == EventKind.ResponseEnded);
            if (ended is null)
            {
                return new ParsedResponse(events, []);
            }
            var ids = ImmutableArray.CreateBuilder<int>();
            bool seenText = false;
            bool earlyTool = false;
            foreach (var part in ended.Parts)
            {
                int id = ids.Count;
                ids.Add(id);
                Chunk? chunk = Display(part, id);
                if (chunk is not null)
                {
                    // Plain responses have authoritative finals only; ordering
                    // analysis must not fabricate transient streaming chunks.
                    TrackArrival(chunk, ref seenText, ref earlyTool);
                }
            }
            return new ParsedResponse(events, ids.ToImmutable(), earlyTool, billing);
        }

        var state = new StreamAssembly();
        var vendor = Parent().Configuration().Vendor;
        void Emit(Chunk chunk)
        {
            TrackArrival(chunk, ref state.SeenText, ref state.ToolBeforeText);
            onChunk?.Invoke(chunk);
        }
        try
        {
            await using var stream = await response.Content.ReadAsStreamAsync(cancellation);
            using var reader = new StreamReader(stream, Encoding.UTF8, true, 4096, leaveOpen: true);
            var data = new StringBuilder();
            bool hasData = false;
            // ReadLine accepts LF, CRLF and CR. SSE joins data fields with LF;
            // comments, event names, retry and id fields are not JSON payloads.
            while (true)
            {
                string? line = await reader.ReadLineAsync(cancellation);
                if (line is null || line.Length == 0)
                {
                    if (hasData)
                    {
                        string payload = data.ToString();
                        Parent().Parent().Trace(Parent(), "frame", payload);
                        data.Clear();
                        hasData = false;
                        if (payload == "[DONE]")
                        {
                            if (vendor != Vendor.OpenAI || !state.FinishSeen)
                            {
                                throw new FormatException("Unexpected stream terminator");
                            }
                            state.Terminal = true;
                        }
                        else
                        {
                            if (state.Terminal)
                            {
                                throw new FormatException("Data after stream terminator");
                            }
                            var frame = JsonNode.Parse(payload)?.AsObject()
                                ?? throw new FormatException("Empty stream frame");
                            switch (vendor)
                            {
                                case Vendor.Anthropic:
                                    AccumulateAnthropic(state, frame, Emit);
                                    break;
                                case Vendor.OpenAI:
                                    AccumulateOpenAI(state, frame, Emit);
                                    break;
                                case Vendor.Gemini:
                                    AccumulateGemini(state, frame, Emit);
                                    break;
                            }
                        }
                    }
                    if (line is null)
                    {
                        break;
                    }
                    continue;
                }
                if (line.StartsWith(':'))
                {
                    continue;
                }
                int colon = line.IndexOf(':');
                string field = colon < 0 ? line : line[..colon];
                if (field != "data")
                {
                    continue;
                }
                string value = colon < 0 ? "" : line[(colon + 1)..];
                if (value.StartsWith(' '))
                {
                    value = value[1..];
                }
                if (hasData)
                {
                    data.Append('\n');
                }
                data.Append(value);
                hasData = true;
            }
            if (!state.Terminal || (vendor == Vendor.Anthropic && state.OpenBlocks.Count != 0))
            {
                throw new FormatException("Stream ended without completion");
            }
            JsonObject completed = CompleteNativeResponse(state, vendor);
            var events = ParseResponse(status, completed.ToJsonString(), out var billing);
            var final = events.FirstOrDefault(fact => fact.Kind == EventKind.ResponseEnded);
            if (final is null)
            {
                return new ParsedResponse(events, []);
            }
            if (final.Parts.Length != state.Order.Count)
            {
                throw new FormatException("Final part identity mismatch");
            }
            return new ParsedResponse(events, state.Order.ToImmutableArray(), state.ToolBeforeText, billing);
        }
        catch (Exception error) when (error is JsonException or FormatException or InvalidOperationException
            or ArgumentException or KeyNotFoundException or OverflowException)
        {
            // A malformed stream can have displayed chunks, but cannot authorize
            // a final response or a partially assembled tool call.
            return new ParsedResponse([new Event(EventKind.ErrorOccurred,
                error: "Invalid provider stream: " + error.GetType().Name)], []);
        }
    }

    private static void TrackArrival(Chunk chunk, ref bool seenText, ref bool earlyTool)
    {
        if (chunk.Kind == ChunkKind.Reply && !string.IsNullOrWhiteSpace(chunk.Text))
        {
            seenText = true;
        }
        if (chunk.Kind == ChunkKind.ToolCall && !seenText)
        {
            earlyTool = true;
        }
    }

    private static Chunk? Display(Part part, int id)
    {
        return part switch
        {
            TextPart text => new Chunk(id, ChunkKind.Reply, Text: text.Text),
            ToolCallPart call => new Chunk(id, ChunkKind.ToolCall,
                ToolName: call.Name, ArgsText: call.Args.GetRawText()),
            OpaquePart opaque when opaque.Data.TryGetProperty("thinking", out var thinking) =>
                new Chunk(id, ChunkKind.Thinking, Text: thinking.GetString() ?? ""),
            OpaquePart opaque when opaque.Data.TryGetProperty("text", out var text) =>
                new Chunk(id, ChunkKind.Thinking, Text: text.GetString() ?? ""),
            _ => null
        };
    }

    // Missing members are malformed provider input, never object-lifetime
    // failures. The outer parser translates these into fixed safe errors.
    private static JsonNode Required(JsonObject source, string field)
    {
        return source[field] ?? throw new FormatException("Missing " + field);
    }

    private static string String(JsonObject obj, string field)
    {
        return obj[field]?.GetValue<string>() ?? "";
    }

    private static void CopyFields(JsonObject source, JsonObject destination)
    {
        foreach (var field in source)
        {
            destination[field.Key] = field.Value?.DeepClone();
        }
    }

    private static void AccumulateAnthropic(StreamAssembly state, JsonObject frame, Action<Chunk> emit)
    {
        string type = String(frame, "type");
        if (type == "error")
        {
            throw new FormatException("Provider stream error");
        }
        if (type == "message_start")
        {
            CopyFields(Required(frame, "message").AsObject(), state.Root);
            return;
        }
        if (type == "message_delta")
        {
            if (frame["usage"] is JsonObject usage)
            {
                var accumulated = state.Root["usage"] as JsonObject ?? new JsonObject();
                CopyFields(usage, accumulated);
                if (accumulated.Parent is null)
                {
                    state.Root["usage"] = accumulated;
                }
            }
            return;
        }
        if (type == "message_stop")
        {
            state.Terminal = true;
            return;
        }
        if (type is not ("content_block_start" or "content_block_delta" or "content_block_stop"))
        {
            return; // ping and other non-content protocol bookkeeping.
        }
        int id = Required(frame, "index").GetValue<int>();
        if (id < 0)
        {
            throw new FormatException("Negative part index");
        }
        if (type == "content_block_start")
        {
            var block = Required(frame, "content_block").DeepClone().AsObject();
            state.Blocks.Add(id, block);
            state.OpenBlocks.Add(id);
            switch (String(block, "type"))
            {
                case "text":
                    emit(new Chunk(id, ChunkKind.Reply, Text: String(block, "text")));
                    break;
                case "thinking":
                    emit(new Chunk(id, ChunkKind.Thinking, Text: String(block, "thinking")));
                    break;
                case "tool_use":
                    emit(new Chunk(id, ChunkKind.ToolCall, ToolName: String(block, "name")));
                    break;
            }
            return;
        }
        if (!state.Blocks.TryGetValue(id, out JsonObject? current))
        {
            throw new FormatException("Delta or stop without a matching block start");
        }
        if (!state.OpenBlocks.Contains(id))
        {
            throw new FormatException("Content after block stop");
        }
        if (type == "content_block_stop")
        {
            state.OpenBlocks.Remove(id);
            if (state.Arguments.TryGetValue(id, out string? args))
            {
                current["input"] = JsonNode.Parse(args);
            }
            return;
        }
        JsonObject delta = Required(frame, "delta").AsObject();
        switch (String(delta, "type"))
        {
            case "text_delta":
                string text = String(delta, "text");
                current["text"] = String(current, "text") + text;
                emit(new Chunk(id, ChunkKind.Reply, Text: text));
                break;
            case "thinking_delta":
                string thinking = String(delta, "thinking");
                current["thinking"] = String(current, "thinking") + thinking;
                emit(new Chunk(id, ChunkKind.Thinking, Text: thinking));
                break;
            case "signature_delta":
                current["signature"] = String(current, "signature") + String(delta, "signature");
                break;
            case "input_json_delta":
                string fragment = String(delta, "partial_json");
                state.Arguments.TryGetValue(id, out string? previous);
                state.Arguments[id] = previous + fragment;
                emit(new Chunk(id, ChunkKind.ToolCall, ArgsText: fragment));
                break;
        }
    }

    private static void AccumulateOpenAI(StreamAssembly state, JsonObject frame, Action<Chunk> emit)
    {
        if (frame.ContainsKey("error"))
        {
            throw new FormatException("Provider stream error");
        }
        if (frame["model"] is JsonNode model)
        {
            state.Root["model"] = model.DeepClone();
        }
        if (frame.ContainsKey("service_tier"))
        {
            // Preserve the final supplied tier, including null/invalid values;
            // the decoder will treat those as unpriced, never as standard.
            state.Root["service_tier"] = frame["service_tier"]?.DeepClone();
        }
        if (frame["usage"] is JsonObject usage)
        {
            state.Root["usage"] = usage.DeepClone();
        }
        foreach (var value in Required(frame, "choices").AsArray())
        {
            var choice = (value?.AsObject() ?? throw new FormatException("Missing choice"));
            if (choice["index"]?.GetValue<int>() is int choiceIndex && choiceIndex != 0)
            {
                continue;
            }
            if (state.FinishSeen)
            {
                throw new FormatException("Content after OpenAI finish reason");
            }
            if (choice["finish_reason"] is not null)
            {
                state.FinishSeen = true;
            }
            var delta = Required(choice, "delta").AsObject();
            foreach (string field in new[] { "content", "reasoning_content" })
            {
                if (delta[field] is not JsonNode textNode)
                {
                    continue;
                }
                string text = textNode.GetValue<string>();
                int key = field == "content" ? -1 : -2;
                if (!state.ToolIDs.TryGetValue(key, out int id))
                {
                    id = state.NextID++;
                    state.ToolIDs.Add(key, id);
                    state.Blocks.Add(id, new JsonObject { ["type"] = field, ["text"] = "" });
                }
                var block = state.Blocks[id];
                block["text"] = String(block, "text") + text;
                emit(new Chunk(id, field == "content" ? ChunkKind.Reply : ChunkKind.Thinking, Text: text));
            }
            if (delta["tool_calls"] is not JsonArray tools)
            {
                continue;
            }
            foreach (var toolValue in tools)
            {
                var tool = (toolValue?.AsObject() ?? throw new FormatException("Missing tool delta"));
                int index = Required(tool, "index").GetValue<int>();
                if (index < 0)
                {
                    throw new FormatException("Negative tool index");
                }
                if (!state.ToolIDs.TryGetValue(index, out int id))
                {
                    id = state.NextID++;
                    state.ToolIDs.Add(index, id);
                    state.Blocks.Add(id, new JsonObject { ["type"] = "function", ["id"] = "",
                        ["function"] = new JsonObject { ["name"] = "", ["arguments"] = "" } });
                }
                var call = state.Blocks[id];
                if (tool["id"] is not null)
                {
                    call["id"] = String(call, "id") + String(tool, "id");
                }
                var function = call["function"]!.AsObject();
                var fragment = tool["function"] as JsonObject ?? new JsonObject();
                string name = String(fragment, "name");
                string args = String(fragment, "arguments");
                function["name"] = String(function, "name") + name;
                function["arguments"] = String(function, "arguments") + args;
                emit(new Chunk(id, ChunkKind.ToolCall, ToolName: name, ArgsText: args));
            }
        }
    }

    private static void AccumulateGemini(StreamAssembly state, JsonObject frame, Action<Chunk> emit)
    {
        if (frame.ContainsKey("error"))
        {
            throw new FormatException("Provider stream error");
        }
        foreach (string field in new[] { "modelVersion", "usageMetadata" })
        {
            if (frame[field] is JsonNode value)
            {
                state.Root[field] = value.DeepClone();
            }
        }
        if (frame["candidates"] is not JsonArray candidates || candidates.Count == 0)
        {
            return;
        }
        var candidate = (candidates[0]?.AsObject() ?? throw new FormatException("Missing candidate"));
        if (candidate["finishReason"] is not null)
        {
            string reason = String(candidate, "finishReason");
            if (reason is not ("STOP" or "MAX_TOKENS"))
            {
                throw new FormatException("Unsuccessful Gemini completion");
            }
            state.Terminal = true;
        }
        if (candidate["content"]?["parts"] is not JsonArray parts)
        {
            return;
        }
        for (int position = 0; position < parts.Count; position++)
        {
            var part = (parts[position]?.AsObject() ?? throw new FormatException("Missing part"));
            bool signatureOnly = part.Count == 1 && part.ContainsKey("thoughtSignature");
            int? previousID = state.GeminiCurrent;
            JsonObject? previous = previousID is int prior ? state.Blocks[prior] : null;
            bool unsigned = previous is not null && !previous.ContainsKey("thoughtSignature");
            bool textContinuation = position == 0 && unsigned && previous!.ContainsKey("text")
                && part.ContainsKey("text")
                && (previous["thought"]?.GetValue<bool>() ?? false) == (part["thought"]?.GetValue<bool>() ?? false);
            int id;
            if (unsigned && signatureOnly)
            {
                id = previousID!.Value;
                previous!["thoughtSignature"] = Required(part, "thoughtSignature").DeepClone();
            }
            else if (textContinuation)
            {
                id = previousID!.Value;
                previous!["text"] = String(previous, "text") + String(part, "text");
                if (part["thoughtSignature"] is JsonNode signature)
                {
                    previous["thoughtSignature"] = signature.DeepClone();
                }
            }
            else
            {
                id = state.NextID++;
                state.Blocks.Add(id, part.DeepClone().AsObject());
                state.GeminiCurrent = id;
            }
            if (part["text"] is JsonNode text)
            {
                var kind = part["thought"]?.GetValue<bool>() == true ? ChunkKind.Thinking : ChunkKind.Reply;
                emit(new Chunk(id, kind, Text: text.GetValue<string>()));
            }
            if (part["functionCall"] is JsonObject call)
            {
                string args = call["args"]?.ToJsonString() ?? "{}";
                emit(new Chunk(id, ChunkKind.ToolCall, ToolName: String(call, "name"), ArgsText: args));
            }
        }
    }

    private static JsonObject CompleteNativeResponse(StreamAssembly state, Vendor vendor)
    {
        var blocks = new JsonArray();
        if (vendor is Vendor.Anthropic or Vendor.Gemini)
        {
            foreach (var (id, block) in state.Blocks)
            {
                blocks.Add(block.DeepClone());
                state.Order.Add(id);
            }
            if (vendor == Vendor.Anthropic)
            {
                state.Root["content"] = blocks;
            }
            else
            {
                state.Root["candidates"] = new JsonArray(new JsonObject
                {
                    ["content"] = new JsonObject { ["parts"] = blocks }
                });
            }
            return state.Root;
        }

        var message = new JsonObject();
        // This is native OpenAI's completed-message order, not arrival order.
        // The independent ID list preserves the mapping across that projection.
        foreach (string field in new[] { "content", "reasoning_content" })
        {
            foreach (var (id, block) in state.Blocks.Where(pair => String(pair.Value, "type") == field))
            {
                message[field] = block["text"]!.DeepClone();
                state.Order.Add(id);
            }
        }
        foreach (var (id, block) in state.Blocks.Where(pair => String(pair.Value, "type") == "function"))
        {
            blocks.Add(block.DeepClone());
            state.Order.Add(id);
        }
        if (blocks.Count > 0)
        {
            message["tool_calls"] = blocks;
        }
        state.Root["choices"] = new JsonArray(new JsonObject { ["message"] = message });
        return state.Root;
    }
}
