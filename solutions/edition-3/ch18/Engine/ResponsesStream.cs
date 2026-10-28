using System.Collections.Immutable;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class Engine
{
    private async Task<ParsedResponse> ParseResponsesStreamAsync(HttpResponseMessage response,
        Action<Chunk>? callback, CancellationToken cancellation)
    {
        var state = new ResponsesAssembly();
        try
        {
            await using var stream = await response.Content.ReadAsStreamAsync(cancellation);
            using var reader = new StreamReader(stream, new UTF8Encoding(false, true));
            var data = new StringBuilder();
            while (await reader.ReadLineAsync(cancellation) is string line)
            {
                if (line.Length == 0)
                {
                    ConsumeResponsesFrame(data, state, callback);
                }
                else if (line.StartsWith("data:", StringComparison.Ordinal))
                {
                    if (data.Length > 0)
                    {
                        data.Append('\n');
                    }
                    data.Append(line.AsSpan(line.Length > 5 && line[5] == ' ' ? 6 : 5));
                }
            }
            ConsumeResponsesFrame(data, state, callback);
            if (state.Terminal is not JsonObject terminal)
            {
                throw new InvalidDataException("Responses stream ended without terminal success");
            }
            var output = terminal["output"] as JsonArray
                ?? throw new InvalidDataException("Missing terminal output");
            int[] indices;
            if (output.Count == 0)
            {
                // Item-done is the only fallback source. Added/delta material may
                // contain incomplete encrypted content or function arguments.
                if (!state.SeenItems.SetEquals(state.Completed.Keys))
                {
                    throw new InvalidDataException("Responses terminal left unfinished items");
                }
                indices = state.Completed.Keys.ToArray();
                output = new JsonArray(state.Completed.Values.Select(item => (JsonNode)item.DeepClone()).ToArray());
                terminal["output"] = output;
            }
            else
            {
                // A nonempty terminal payload wins; never concatenate it with
                // completed items or manufacture duplicate calls and text.
                indices = Enumerable.Range(0, output.Count).ToArray();
            }
            var facts = ParseResponse((int)response.StatusCode, terminal.ToJsonString(), out var billing);
            if (facts.Any(fact => fact.Kind == EventKind.ErrorOccurred))
            {
                return new ParsedResponse(facts, []);
            }
            var parts = facts.Single(fact => fact.Kind == EventKind.ResponseEnded).Parts;
            var ids = ImmutableArray.CreateBuilder<int>();
            int cursor = 0;
            var config = Parent().Configuration();
            var origin = new Provenance(config.Vendor, terminal["model"]?.GetValue<string>() ?? config.ResolvedModel, Surface.Responses);
            for (int itemIndex = 0; itemIndex < output.Count; itemIndex++)
            {
                using var document = JsonDocument.Parse(output[itemIndex]!.ToJsonString());
                var itemParts = ResponsesItem(document.RootElement, origin);
                for (int partIndex = 0; partIndex < itemParts.Count; partIndex++)
                {
                    int id = ResponsesID(state, indices[itemIndex], partIndex);
                    ids.Add(id);
                    var part = parts[cursor++];
                    // Final-only summaries and final suffixes use the same stable
                    // ID as deltas. Already-delivered text is never spoken twice.
                    string? text = part switch
                    {
                        TextPart value => value.Text,
                        OpaquePart opaque => ResponsesSummary(opaque.Data),
                        _ => null
                    };
                    if (text is not null)
                    {
                        string shown = state.Displayed.GetValueOrDefault(id, "");
                        if (text.StartsWith(shown, StringComparison.Ordinal) && text.Length > shown.Length)
                        {
                            var kind = part is OpaquePart ? ChunkKind.ReasoningSummary : ChunkKind.Reply;
                            callback?.Invoke(new Chunk(id, kind, text[shown.Length..]));
                        }
                    }
                }
            }
            return new ParsedResponse(facts, ids.ToImmutable(), state.ToolBeforeText, billing);
        }
        catch (Exception error) when (error is JsonException or InvalidOperationException
            or FormatException or KeyNotFoundException or ArgumentException or InvalidDataException)
        {
            return new ParsedResponse([new Event(EventKind.ErrorOccurred,
                error: "Invalid Responses stream: " + error.GetType().Name)], []);
        }
    }

    private void ConsumeResponsesFrame(StringBuilder data, ResponsesAssembly state, Action<Chunk>? callback)
    {
        if (data.Length == 0)
        {
            return;
        }
        string payload = data.ToString();
        data.Clear();
        Parent().Parent().Trace(Parent(), "frame", payload);
        if (payload == "[DONE]")
        {
            if (state.Terminal is null)
            {
                throw new InvalidDataException("DONE is not Responses terminal success");
            }
            return;
        }
        if (state.Terminal is not null)
        {
            throw new InvalidDataException("Responses data followed terminal success");
        }
        var frame = JsonNode.Parse(payload)?.AsObject()
            ?? throw new InvalidDataException("Responses event must be an object");
        string type = frame["type"]?.GetValue<string>()
            ?? throw new InvalidDataException("Responses event type missing");
        switch (type)
        {
            case "error":
            case "response.failed":
            case "response.incomplete":
                throw new InvalidDataException("Responses stream did not succeed");
            case "response.completed":
                var terminal = frame["response"]?.DeepClone().AsObject()
                    ?? throw new InvalidDataException("Missing terminal response");
                if (terminal["status"]?.GetValue<string>() != "completed")
                {
                    throw new InvalidDataException("Invalid completed status");
                }
                state.Terminal = terminal;
                break;
            case "response.output_item.added":
            case "response.output_item.done":
                int index = ResponsesIndex(frame, "output_index");
                state.SeenItems.Add(index);
                var item = frame["item"]?.DeepClone().AsObject()
                    ?? throw new InvalidDataException("Missing output item");
                // A completed visible message is an explanation even when the
                // provider omitted deltas. An added skeleton or reasoning summary
                // is not a user-visible explanation of a later tool action.
                if (type == "response.output_item.done" && item["type"]?.GetValue<string>() == "message"
                    && item["content"] is JsonArray content)
                {
                    foreach (var block in content.OfType<JsonObject>())
                    {
                        string? kind = block["type"]?.GetValue<string>();
                        string? visible = kind switch
                        {
                            "output_text" => block["text"]?.GetValue<string>(),
                            "refusal" => block["refusal"]?.GetValue<string>(),
                            _ => null
                        };
                        if (!string.IsNullOrWhiteSpace(visible))
                        {
                            state.SeenText = true;
                        }
                    }
                }
                if (item["type"]?.GetValue<string>() == "function_call")
                {
                    if (!state.SeenText)
                    {
                        state.ToolBeforeText = true;
                    }
                    int id = ResponsesID(state, index, 0);
                    callback?.Invoke(new Chunk(id, ChunkKind.ToolCall, ToolName: item["name"]?.GetValue<string>() ?? ""));
                }
                if (type == "response.output_item.done" && !state.Completed.TryAdd(index, item))
                {
                    throw new InvalidDataException("Duplicate completed Responses item");
                }
                break;
            case "response.output_text.delta":
            case "response.refusal.delta":
            case "response.reasoning_summary_text.delta":
                int outputIndex = ResponsesIndex(frame, "output_index");
                bool summary = type == "response.reasoning_summary_text.delta";
                int contentIndex = summary ? 0 : ResponsesIndex(frame, "content_index");
                if (summary)
                {
                    _ = ResponsesIndex(frame, "summary_index");
                }
                state.SeenItems.Add(outputIndex);
                int textID = ResponsesID(state, outputIndex, contentIndex);
                string text = frame["delta"]?.GetValue<string>()
                    ?? throw new InvalidDataException("Missing text delta");
                state.Displayed[textID] = state.Displayed.GetValueOrDefault(textID, "") + text;
                if (!summary && !string.IsNullOrWhiteSpace(text))
                {
                    state.SeenText = true;
                }
                callback?.Invoke(new Chunk(textID, summary ? ChunkKind.ReasoningSummary : ChunkKind.Reply, text));
                break;
            case "response.function_call_arguments.delta":
                int callIndex = ResponsesIndex(frame, "output_index");
                state.SeenItems.Add(callIndex);
                if (!state.SeenText)
                {
                    state.ToolBeforeText = true;
                }
                callback?.Invoke(new Chunk(ResponsesID(state, callIndex, 0), ChunkKind.ToolCall,
                    ArgsText: frame["delta"]?.GetValue<string>() ?? throw new InvalidDataException("Missing arguments delta")));
                break;
            // Lifecycle and done notifications carry no additional authority.
            // The completed output item and final response are the facts.
        }
    }

    private static int ResponsesIndex(JsonObject frame, string field)
    {
        int index = frame[field]?.GetValue<int>() ?? throw new InvalidDataException("Missing Responses index");
        if (index < 0)
        {
            throw new InvalidDataException("Negative Responses index");
        }
        return index;
    }

    private static int ResponsesID(ResponsesAssembly state, int item, int part)
    {
        var key = (item, part);
        if (!state.IDs.TryGetValue(key, out int id))
        {
            id = state.IDs.Count;
            state.IDs.Add(key, id);
        }
        return id;
    }
}
