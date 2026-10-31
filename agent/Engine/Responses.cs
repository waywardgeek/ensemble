using System.Collections.Immutable;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

// Responses is another Engine wire surface, not another actor or History.
// All projection here is disposable. Opaque items retain exact-model provenance.
public sealed partial class Engine
{
    private JsonObject RenderResponses(List<Entry> entries, Configuration config, ImageRequestBudget budget)
    {
        var input = new JsonArray();
        foreach (var entry in entries)
        {
            string role = entry.Actor switch
            {
                Actor.Agent => "assistant",
                Actor.System => "developer",
                _ => "user"
            };
            JsonArray? content = null;
            // Adjacent text/images describe one message. A tool or native opaque
            // item ends that content group without reordering canonical parts.
            void AppendContent(JsonObject block)
            {
                if (content is null)
                {
                    content = new JsonArray();
                    input.Add(new JsonObject
                    {
                        ["type"] = "message", ["role"] = role, ["content"] = content
                    });
                }
                content.Add(block);
            }
            foreach (var part in entry.Parts)
            {
                if (part is not TextPart and not BlobPart) content = null;
                switch (part)
                {
                    case TextPart text:
                        string type = entry.Actor == Actor.Agent ? "output_text" : "input_text";
                        AppendContent(new JsonObject { ["type"] = type, ["text"] = text.Text });
                        break;
                    case ToolCallPart call:
                        var item = new JsonObject
                        {
                            ["type"] = "function_call", ["call_id"] = call.CallID,
                            ["name"] = call.Name, ["arguments"] = call.Args.GetRawText()
                        };
                        if (call.From.Matches(config.Target()) && call.Opaque is JsonElement native)
                        {
                            item = JsonNode.Parse(native.GetRawText())!.AsObject();
                        }
                        // A billing switch keeps the same Responses provenance,
                        // but the declared namespace changes with the route profile.
                        if (config.Billing is BillingMode.ChatGptPlan or BillingMode.Codex)
                        {
                            item["namespace"] = "ensemble";
                        }
                        else
                        {
                            item.Remove("namespace");
                        }
                        input.Add(item);
                        break;
                    case ToolResultPart result:
                        input.Add(new JsonObject
                        {
                            ["type"] = "function_call_output", ["call_id"] = result.CallID,
                            ["output"] = ResponsesResult(result, budget)
                        });
                        break;
                    case OpaquePart opaque:
                        if (opaque.From.Matches(config.Target()))
                        {
                            input.Add(JsonNode.Parse(opaque.Data.GetRawText()));
                        }
                        break;
                    case BlobPart blob:
                        // Project has validated modality and placement. Reuse the
                        // Agent-owned Sandbox read; never expose its file locator.
                        AppendContent(ResponsesImage(blob, budget));
                        break;
                    default:
                        throw new InvalidOperationException("Unexpected Responses projection part");
                }
            }
        }
        var request = new JsonObject
        {
            ["model"] = config.Model,
            ["instructions"] = config.PrimaryInstructions,
            ["input"] = input,
            // Stateless turns must carry their own provider-signed reasoning.
            ["include"] = new JsonArray("reasoning.encrypted_content"),
            ["store"] = false
        };
        if (config.ReasoningEffort.Length != 0 || config.ReasoningSummary.Length != 0)
        {
            var reasoning = new JsonObject();
            if (config.ReasoningEffort.Length != 0)
            {
                reasoning["effort"] = config.ReasoningEffort;
            }
            if (config.ReasoningSummary.Length != 0)
            {
                reasoning["summary"] = config.ReasoningSummary;
            }
            request["reasoning"] = reasoning;
        }
        return request;
    }

    private JsonObject ResponsesImage(BlobPart blob, ImageRequestBudget budget)
    {
        string encoded = Convert.ToBase64String(ReadMedia(blob, budget));
        return new JsonObject
        {
            ["type"] = "input_image",
            ["image_url"] = $"data:{blob.Mime};base64,{encoded}",
            ["detail"] = "auto"
        };
    }

    private JsonNode ResponsesResult(ToolResultPart result, ImageRequestBudget budget)
    {
        if (result.Parts.Any(part => part is not TextPart and not RedactedPart and not BlobPart))
        {
            throw new NotSupportedException("Unsupported Responses tool-result part");
        }
        // Preserve the existing string contract for ordinary results, including
        // readable redaction markers without opening their private locators.
        if (!result.Parts.Any(part => part is BlobPart))
        {
            return JsonValue.Create(JsonCodec.Text(result.Parts))!;
        }
        var content = new JsonArray();
        foreach (var part in result.Parts)
        {
            if (part is BlobPart blob)
            {
                content.Add(ResponsesImage(blob, budget));
            }
            else
            {
                content.Add(new JsonObject
                {
                    ["type"] = "input_text", ["text"] = JsonCodec.Text([part])
                });
            }
        }
        return content;
    }

    private static List<Part> ParseResponses(JsonElement root, Provenance from)
    {
        if (GetString(root, "status") != "completed"
            || (root.TryGetProperty("error", out var error) && error.ValueKind != JsonValueKind.Null))
        {
            throw new InvalidDataException("Responses did not complete successfully");
        }
        if (!root.TryGetProperty("output", out var output) || output.ValueKind != JsonValueKind.Array)
        {
            throw new InvalidDataException("Responses output must be an array");
        }
        var parts = new List<Part>();
        foreach (var item in output.EnumerateArray())
        {
            parts.AddRange(ResponsesItem(item, from));
        }
        return parts;
    }

    private static List<Part> ResponsesItem(JsonElement item, Provenance from)
    {
        var parts = new List<Part>();
        if (item.TryGetProperty("status", out var status) && status.GetString() != "completed")
        {
            throw new InvalidDataException("Responses item is incomplete");
        }
        switch (GetString(item, "type"))
        {
            case "message":
                if (GetString(item, "role") != "assistant")
                {
                    throw new InvalidDataException("Responses output message must be assistant content");
                }
                foreach (var content in item.GetProperty("content").EnumerateArray())
                {
                    string text = GetString(content, "type") switch
                    {
                        "output_text" => content.GetProperty("text").GetString()
                            ?? throw new InvalidDataException("Missing output text"),
                        "refusal" => content.GetProperty("refusal").GetString()
                            ?? throw new InvalidDataException("Missing refusal text"),
                        _ => throw new InvalidDataException("Unsupported Responses message content")
                    };
                    parts.Add(new TextPart(text));
                }
                break;
            case "function_call":
                string? space = item.TryGetProperty("namespace", out var ns) && ns.ValueKind != JsonValueKind.Null
                    ? ns.GetString() : null;
                if (space is not null && space != "ensemble")
                {
                    throw new InvalidDataException("Unknown Responses tool namespace");
                }
                string arguments = item.GetProperty("arguments").GetString()
                    ?? throw new InvalidDataException("Missing function arguments");
                using (var parsed = JsonDocument.Parse(arguments))
                {
                    if (parsed.RootElement.ValueKind != JsonValueKind.Object)
                    {
                        throw new InvalidDataException("Function arguments must be an object");
                    }
                    parts.Add(new ToolCallPart(GetString(item, "call_id"), from, GetString(item, "name"), parsed.RootElement, item));
                }
                break;
            case "reasoning":
                // Validate summaries separately from opaque encrypted replay bytes.
                _ = ResponsesSummary(item);
                parts.Add(new OpaquePart(from, item));
                break;
            default:
                throw new InvalidDataException("Unsupported Responses output item");
        }
        return parts;
    }

    private static string ResponsesSummary(JsonElement item)
    {
        if (!item.TryGetProperty("summary", out var summary))
        {
            return "";
        }
        if (summary.ValueKind != JsonValueKind.Array)
        {
            throw new InvalidDataException("Reasoning summary must be an array");
        }
        var texts = new List<string>();
        foreach (var part in summary.EnumerateArray())
        {
            if (GetString(part, "type") != "summary_text")
            {
                throw new InvalidDataException("Unsupported reasoning summary content");
            }
            texts.Add(part.GetProperty("text").GetString()
                ?? throw new InvalidDataException("Missing reasoning summary text"));
        }
        return string.Concat(texts);
    }

    private static Usage ResponsesUsage(JsonElement root)
    {
        if (!root.TryGetProperty("usage", out var usage) || usage.ValueKind == JsonValueKind.Null)
        {
            return new Usage(input: null, output: null);
        }
        JsonElement details = default;
        bool malformed = usage.ValueKind == JsonValueKind.Object
            && usage.TryGetProperty("input_tokens_details", out details)
            && details.ValueKind != JsonValueKind.Object;
        long? read = malformed ? null : Count(details, "cached_tokens", true);
        return new Usage(Subtract(Count(usage, "input_tokens"), read), 0, read, Count(usage, "output_tokens"));
    }
}
