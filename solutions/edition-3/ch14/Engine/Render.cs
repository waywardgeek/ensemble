using System.Text.Json;
using System.Text.Json.Nodes;
namespace Ensemble;

public sealed partial class Engine
{
    public partial JsonObject Render(Context context, string requestInstruction)
    {
        var config = Parent().Configuration();
        var entries = Project(context, config);
        var request = config.Vendor switch
        {
            Vendor.Anthropic => RenderAnthropic(entries, config),
            Vendor.OpenAI => RenderOpenAI(entries, config),
            Vendor.Gemini => RenderGemini(entries, config),
            _ => throw new ArgumentException("Unknown vendor")
        };
        if (!config.DisableStreaming && config.Vendor != Vendor.Gemini
            && LookupModel(config.Vendor, config.Model, out var features) && features.Stream)
        {
            request["stream"] = true;
            if (config.Vendor == Vendor.OpenAI)
            {
                request["stream_options"] = new JsonObject { ["include_usage"] = true };
            }
        }
        var skills = context.Dialogue.SelectMany(entry => entry.Parts).OfType<SkillPart>().ToArray();
        var enabled = context.EnabledTools();
        var eager = context.Dialogue.FirstOrDefault(entry => entry.Parts.Any(part => part is SkillPart))?
            .Parts.OfType<SkillPart>().Where(skill => !skill.Pending)
            .SelectMany(skill => skill.Tools).ToHashSet(StringComparer.Ordinal) ?? enabled.ToHashSet(StringComparer.Ordinal);
        bool deferred = config.DeferredTools && skills.Length > 0 && Parent().SkillManager() is not null;
        IReadOnlySet<string> declared = deferred
            ? Parent().SkillManager()!.CatalogTools(context).ToHashSet(StringComparer.Ordinal)
            : enabled;
        var tools = Parent().Tools().Where(tool => !Parent().IsAutomaticTool(tool.Name())
            && ((skills.Length == 0 && Parent().SkillManager() is null) || declared.Contains(tool.Name()))).ToArray();
        if (tools.Length > 0)
        {
            var declarations = new JsonArray();
            foreach (var tool in tools)
            {
                var declaration = tool.Declaration(config.Vendor);
                if (deferred && !eager.Contains(tool.Name()))
                {
                    declaration["defer_loading"] = true;
                }
                declarations.Add(declaration);
            }
            string instruction = config.PrimaryInstructions
                + "\n\nBefore each tool batch, give a concise user-visible explanation of what you will do and why. "
                + "Use the declared tools to inspect and change the workspace. "
                + "Report tool failures honestly; do not invent results. "
                + "Shell state does not persist between calls. "
                + "Every ordinary tool returns a job report. "
                + "A running job is not failure: retain its handle and use wait_for_job. "
                + "Polling delay never cancels work. "
                + "send_input submits a line to a PTY and adds newline if missing. "
                + "Only kill_job or shutdown cancels work. "
                + "tool_limits sets observation options for the very next call, including controls, which reports consumption. "
                + "Full output is retained at output_path; inline truncation does not stop the job.";
            if (config.Vendor == Vendor.Anthropic)
            {
                request["system"] = instruction;
            }
            else if (config.Vendor == Vendor.OpenAI)
            {
                request["messages"]![0]!["content"] = instruction;
            }
            else
            {
                request["systemInstruction"] = new JsonObject
                {
                    ["parts"] = new JsonArray(TextBlock(instruction, false))
                };
            }
            request["tools"] = config.Vendor == Vendor.Gemini
                ? new JsonArray(new JsonObject { ["functionDeclarations"] = declarations })
                : declarations;
        }
        if (skills.Length > 0)
        {
            string instructions = string.Join("\n\n", skills.OrderByDescending(skill => skill.Primary)
                .Select(skill => skill.Body));
            if (config.Vendor == Vendor.Anthropic)
            {
                request["system"] = instructions;
            }
            else if (config.Vendor == Vendor.OpenAI)
            {
                request["messages"]![0]!["content"] = instructions;
            }
            else
            {
                request["systemInstruction"] = new JsonObject
                {
                    ["parts"] = new JsonArray(TextBlock(instructions, false))
                };
            }
        }
        if (requestInstruction.Length > 0)
        {
            // Append to the already-selected primary/tool guidance. No entry is
            // inserted into dialogue, so signed/opaque blocks remain untouched.
            if (config.Vendor == Vendor.Anthropic)
            {
                request["system"] = request["system"]!.GetValue<string>() + "\n\n" + requestInstruction;
            }
            else if (config.Vendor == Vendor.OpenAI)
            {
                var primary = request["messages"]![0]!;
                primary["content"] = primary["content"]!.GetValue<string>() + "\n\n" + requestInstruction;
            }
            else
            {
                request["systemInstruction"]!["parts"]!.AsArray().Add(TextBlock(requestInstruction, false));
            }
        }
        return request;
    }

    // Correlate occurrences, not globally unique IDs: vendors may reuse an ID in a
    // later completed exchange. Refuse ambiguous/foreign results instead of guessing.
    // This disposable projection moves complete result groups ahead of interleaved
    // human messages. Neither the entries nor their immutable parts are changed.
    private static List<Entry> Project(Context context, Configuration config)
    {
        var entries = context.Dialogue.Where(entry => !entry.Parts.Any(part => part is SkillPart)).ToArray();
        var revoked = context.RevokedTools();
        var pending = new Dictionary<string, (int Index, ToolCallPart Call)>(StringComparer.Ordinal);
        var results = new Dictionary<(int Index, string ID), ToolResultPart>();
        for (int i = 0; i < entries.Length; i++)
        {
            CheckMedia(entries[i].Parts, config);
            foreach (var part in entries[i].Parts)
            {
                if (part is ToolCallPart call)
                {
                    if (entries[i].Actor != Actor.Agent || !pending.TryAdd(call.CallID, (i, call)))
                    {
                        throw new FormatException("Ambiguous tool call identity");
                    }
                }
                if (part is ToolResultPart result)
                {
                    if (!pending.Remove(result.CallID, out var match))
                    {
                        throw new FormatException("Tool result has no unambiguous preceding call");
                    }

                    // A historical discovery result is evidence, not a permanent
                    // capability. Filter only the disposable request projection.
                    var projectedResult = new ToolResultPart(result.CallID, result.Parts,
                        result.IsError, result.Job, result.ToolReferences.Where(name => !revoked.Contains(name)));
                    results.Add((match.Index, result.CallID), projectedResult);
                }
            }
        }
        var projected = new List<Entry>();
        for (int i = 0; i < entries.Length; i++)
        {
            var entry = entries[i];
            var retained = entry.Parts.Where(p => p is not ToolResultPart)
                .Select(p => p is HandoffPart note ? new TextPart(note.Text) : p).ToArray();
            if (retained.Length > 0)
            {
                projected.Add(new Entry(entry.Seq, entry.Actor, retained));
            }

            var group = new List<Part>();
            foreach (var call in entry.Parts.OfType<ToolCallPart>())
            {
                if (!results.TryGetValue((i, call.CallID), out var result))
                {
                    result = new ToolResultPart(call.CallID,
                        [new TextPart("No result was recorded; execution outcome unknown.")],
                        isError: true);
                }
                group.Add(result);
            }

            if (group.Count > 0)
            {
                projected.Add(new Entry(entry.Seq, Actor.Tool, group));
            }
        }
        CheckMedia(context.Ephemeral, config);
        if (context.Ephemeral.Length > 0)
        {
            projected.Add(new Entry(0, Actor.Human, context.Ephemeral));
        }

        return projected;
    }

    // Anthropic uses typed blocks; Gemini uses an object with only text.
    private static JsonObject TextBlock(string text, bool anthropic)
    {
        var block = new JsonObject();
        if (anthropic)
        {
            block["type"] = "text";
        }
        block["text"] = text;
        return block;
    }

    private JsonArray AnthropicParts(IEnumerable<Part> parts, Provenance target, bool deferred = false)
    {
        var blocks = new JsonArray();
        foreach (var p in parts)
        {
            switch (p)
            {
                case TextPart t:
                    blocks.Add(TextBlock(t.Text, true));
                    break;
                case BlobPart blob:
                    blocks.Add(AnthropicBlob(blob));
                    break;
                case RedactedPart r:
                    blocks.Add(TextBlock(JsonCodec.Text([r]), true));
                    break;
                case OpaquePart o when o.From.Matches(target):
                    blocks.Add(JsonCodec.Canonical(o.Data));
                    break;
                case ToolCallPart t:
                    var call = new JsonObject
                    {
                        ["type"] = "tool_use",
                        ["id"] = t.CallID,
                        ["name"] = t.Name,
                        ["input"] = JsonCodec.Canonical(t.Args)
                    };
                    CopyOpaque(t, target, call);
                    blocks.Add(call);
                    break;
                case ToolResultPart t:
                    // Canonical results remain intact. Even old/replayed discovery
                    // metadata cannot advertise automatic tools to the model.
                    var references = t.ToolReferences.Where(name => !Parent().IsAutomaticTool(name)).ToArray();
                    blocks.Add(new JsonObject
                    {
                        ["type"] = "tool_result",
                        ["tool_use_id"] = t.CallID,
                        ["content"] = deferred && references.Length > 0
                            ? new JsonArray(references.Select(name => (JsonNode)new JsonObject
                            {
                                ["type"] = "tool_reference",
                                ["tool_name"] = name
                            }).ToArray())
                            : AnthropicParts(t.Parts, target),
                        ["is_error"] = t.IsError
                    });
                    break;
            }
        }
        return blocks;
    }

    // Extra provider fields may accompany a matching call, but cannot override
    // its semantic ID, name, or arguments. Gemini's scalar signature is handled
    // separately because it belongs beside functionCall, not inside it.
    private static void CopyOpaque(ToolCallPart call, Provenance target, JsonObject into)
    {
        if (!call.From.Matches(target) || call.Opaque is not JsonElement opaque)
        {
            return;
        }

        if (opaque.ValueKind != JsonValueKind.Object)
        {
            throw new FormatException("Per-call opaque material must be an object");
        }

        foreach (var field in opaque.EnumerateObject())
        {
            if (into.ContainsKey(field.Name))
            {
                throw new FormatException("Opaque material conflicts with semantic call fields");
            }

            into[field.Name] = JsonCodec.Canonical(field.Value);
        }
    }

    // Anthropic places tool results in user messages. Merge adjacent same-role
    // entries while preserving the required result-before-text order.
    private JsonObject RenderAnthropic(List<Entry> entries, Configuration config)
    {
        var messages = new JsonArray();
        var target = config.Target();
        foreach (var entry in entries)
        {
            var blocks = AnthropicParts(entry.Parts, target, config.DeferredTools);
            if (blocks.Count == 0)
            {
                continue;
            }

            var role = entry.Actor == Actor.Agent ? "assistant" : "user";
            if (messages.LastOrDefault() is JsonObject previous && previous["role"]!.GetValue<string>() == role)
            {
                var content = (JsonArray)previous["content"]!;
                foreach (var b in blocks)
                {
                    content.Add(b!.DeepClone());
                }

                if (role == "user")
                {
                    // Anthropic demands all results before ordinary text in a merged user turn.
                    var sorted = content.OrderBy(p => p?["type"]?.GetValue<string>() == "tool_result" ? 0 : 1).Select(p => p!.DeepClone()).ToArray();
                    previous["content"] = new JsonArray(sorted);
                }
            }
            else
            {
                messages.Add(new JsonObject
                {
                    ["role"] = role,
                    ["content"] = blocks
                });
            }
        }
        return new JsonObject
        {
            ["model"] = config.Model,
            ["max_tokens"] = 1024,
            ["system"] = config.PrimaryInstructions,
            ["messages"] = messages
        };
    }

    // Chat Completions uses a separate tool role and JSON arguments encoded as
    // a string. Canonical preserves number tokens before that string is produced.
    private static JsonObject RenderOpenAI(List<Entry> entries, Configuration config)
    {
        var messages = new JsonArray { new JsonObject { ["role"] = "system",
["content"] = config.PrimaryInstructions } };
        var target = config.Target();
        foreach (var entry in entries)
        {
            foreach (var result in entry.Parts.OfType<ToolResultPart>())
            {
                messages.Add(new JsonObject
                {
                    ["role"] = "tool",
                    ["tool_call_id"] = result.CallID,
                    ["content"] = (result.IsError ? "Tool error: " : "") + JsonCodec.Text(result.Parts)
                });
            }

            var calls = new JsonArray();
            foreach (var call in entry.Parts.OfType<ToolCallPart>())
            {
                var value = new JsonObject
                {
                    ["id"] = call.CallID,
                    ["type"] = "function",
                    ["function"] = new JsonObject
                    {
                        ["name"] = call.Name,
                        ["arguments"] = JsonCodec.Canonical(call.Args)?.ToJsonString() ?? "null"
                    }
                };
                CopyOpaque(call, target, value);
                calls.Add(value);
            }
            var text = JsonCodec.Text(entry.Parts.Where(p => p is not ToolResultPart));
            var opaque = entry.Parts.OfType<OpaquePart>().Where(p => p.From.Matches(target)).ToArray();
            bool hasMedia = entry.Parts.Any(p => p is BlobPart);
            if (text.Length == 0 && calls.Count == 0 && opaque.Length == 0 && !hasMedia && !entry.Parts.Any(p => p is TextPart))
            {
                continue;
            }

            var message = new JsonObject
            {
                ["role"] = entry.Actor == Actor.Agent ? "assistant" : entry.Actor == Actor.System ? "system" : "user",
                ["content"] = text
            };
            if (calls.Count > 0)
            {
                message["tool_calls"] = calls;
            }
            // Non-text content blocks retain their original shape only on their source model.
            if (opaque.Length > 0 || hasMedia)
            {
                var content = new JsonArray();
                foreach (var p in entry.Parts)
                {
                    if (p is TextPart t)
                    {
                        content.Add(new JsonObject
                        {
                            ["type"] = "text",
                            ["text"] = t.Text
                        });
                    }
                    else if (p is RedactedPart redacted)
                    {
                        content.Add(new JsonObject
                        {
                            ["type"] = "text",
                            ["text"] = JsonCodec.Text([redacted])
                        });
                    }
                    else if (p is BlobPart blob)
                    {
                        content.Add(OpenAIBlob(blob));
                    }
                    else if (p is OpaquePart o && o.From.Matches(target))
                    {
                        content.Add(JsonCodec.Canonical(o.Data));
                    }
                }

                message["content"] = content;
            }
            messages.Add(message);
        }
        return new JsonObject
        {
            ["model"] = config.Model,
            ["max_completion_tokens"] = 1024,
            ["messages"] = messages
        };
    }

    // Preserve an object result structurally. Failure wraps rather than edits
    // that object, so an existing field named "error" cannot be overwritten.
    private static JsonObject GeminiResult(ToolResultPart result)
    {
        var text = JsonCodec.Text(result.Parts);
        JsonObject? obj = null;
        try
        {
            obj = JsonNode.Parse(text) as JsonObject;
        }
        catch (JsonException)
        {
            // Plain prose is valid tool output; represent it as a text member.
        }
        if (result.IsError)
        {
            return new JsonObject { ["error"] = obj ?? (JsonNode)new JsonObject { ["text"] = text } };
        }

        return obj ?? new JsonObject { ["text"] = text };
    }

    // Gemini requires the function name on a response. Recover it from the
    // correlated call; the call ID itself is not a substitute for that name.
    private static JsonObject RenderGemini(List<Entry> entries, Configuration config)
    {
        var contents = new JsonArray();
        var target = config.Target();
        var names = new Dictionary<string, string>(StringComparer.Ordinal);
        foreach (var entry in entries)
        {
            var parts = new JsonArray();
            foreach (var part in entry.Parts)
            {
                switch (part)
                {
                    case TextPart t:
                        var textPart = TextBlock(t.Text, false);
                        if (t.From is not null && t.From.Matches(target) && t.Opaque is JsonElement textSignature)
                        {
                            textPart["thoughtSignature"] = JsonCodec.Canonical(textSignature);
                        }
                        parts.Add(textPart);
                        break;
                    case BlobPart blob:
                        parts.Add(GeminiBlob(blob));
                        break;
                    case RedactedPart r:
                        parts.Add(TextBlock(JsonCodec.Text([r]), false));
                        break;
                    case OpaquePart o when o.From.Matches(target):
                        parts.Add(JsonCodec.Canonical(o.Data));
                        break;
                    case ToolCallPart call:
                        names[call.CallID] = call.Name;
                        var p = new JsonObject
                        {
                            ["functionCall"] = new JsonObject
                            {
                                ["name"] = call.Name,
                                ["id"] = call.CallID,
                                ["args"] = JsonCodec.Canonical(call.Args)
                            }
                        };
                        if (call.From.Matches(target) && call.Opaque is JsonElement signature)
                        {
                            p["thoughtSignature"] = JsonCodec.Canonical(signature);
                        }

                        parts.Add(p);
                        break;
                    case ToolResultPart result:
                        if (!names.TryGetValue(result.CallID, out var name))
                        {
                            throw new FormatException("No name for Gemini function result");
                        }

                        parts.Add(new JsonObject
                        {
                            ["functionResponse"] = new JsonObject
                            {
                                ["name"] = name,
                                ["id"] = result.CallID,
                                ["response"] = GeminiResult(result)
                            }
                        });
                        break;
                }
            }
            if (parts.Count == 0)
            {
                continue;
            }

            var role = entry.Actor == Actor.Agent ? "model" : "user";
            if (contents.LastOrDefault() is JsonObject last && last["role"]!.GetValue<string>() == role)
            {
                foreach (var part in parts)
                {
                    ((JsonArray)last["parts"]!).Add(part!.DeepClone());
                }
            }
            else
            {
                contents.Add(new JsonObject
                {
                    ["role"] = role,
                    ["parts"] = parts
                });
            }
        }
        return new JsonObject
        {
            ["systemInstruction"] = new JsonObject { ["parts"] = new JsonArray(TextBlock(config.PrimaryInstructions, false)) },
            ["contents"] = contents
        };
    }
}
