using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

public static partial class JsonCodec
{
    private static string Label<T>(T value) where T : struct, Enum => value switch
    {
        Vendor.OpenAI => "openai",
        Redaction.Result => "redact_result",
        Redaction.Tool => "redact_tool",
        Redaction.Dialogue => "redact_dialogue",
        Redaction.Summary => "redact_summary",
        _ => JsonNamingPolicy.SnakeCaseLower.ConvertName(value.ToString())
    };
    private static T EnumValue<T>(string? text) where T : struct, Enum
    {
        foreach (var value in Enum.GetValues<T>())
        {
            if (Label(value) == text)
            {
                return value;
            }
        }
        throw new FormatException($"Unknown {typeof(T).Name}: {text}");
    }

    private static string String(JsonElement o, string key) => o.GetProperty(key).GetString() ?? throw new FormatException($"Null {key}");
    private static string OptionalString(JsonElement o, string key) => o.TryGetProperty(key, out var v) ? v.GetString() ?? "" : "";
    private static IEnumerable<Part> Parts(JsonElement o) => o.TryGetProperty("parts", out var p) ? p.EnumerateArray().Select(ReadPart).ToArray() : [];
    private static JsonArray Parts(IEnumerable<Part> parts) => new(parts.Select(p => (JsonNode)WritePart(p)).ToArray());
    private static bool Flag(JsonElement o, string key) => o.TryGetProperty(key, out var v) && v.GetBoolean();
    public static partial JsonObject WriteProvenance(Provenance value) => new()
    {
        ["vendor"] = Label(value.Vendor),
        ["model"] = value.Model,
        ["surface"] = Label(value.Surface)
    };
    public static partial Provenance ReadProvenance(JsonElement value)
    {
        var vendor = EnumValue<Vendor>(String(value, "vendor"));
        var model = String(value, "model");
        var surface = String(value, "surface");
        // Both published spellings identify the same Gemini surface.
        if (surface == "generatecontent")
        {
            surface = "generate_content";
        }
        return new Provenance(vendor, model, EnumValue<Surface>(surface));
    }

    private static JsonObject WriteRef(BlobRef value) => new()
    {
        ["kind"] = (int)value.Kind,
        ["locator"] = value.Locator
    };
    private static BlobRef ReadRef(JsonElement value) => new((RefKind)value.GetProperty("kind").GetInt32(), String(value, "locator"));
    public static partial JsonObject WriteUsage(Usage value) => new()
    {
        ["input"] = value.Input,
        ["cache_write"] = value.CacheWrite,
        ["cache_read"] = value.CacheRead,
        ["output"] = value.Output
    };
    private static Usage ReadUsage(JsonElement value)
    {
        long? Count(string key)
        {
            if (!value.TryGetProperty(key, out var count))
            {
                return 0;
            }
            if (count.ValueKind == JsonValueKind.Null)
            {
                return null;
            }
            return count.GetInt64();
        }
        return new(Count("input"), Count("cache_write"), Count("cache_read"), Count("output"));
    }

    public static partial JsonNode? Canonical(JsonElement value)
    {
        // Sorting object keys is deterministic without ever converting numbers to double.
        if (value.ValueKind == JsonValueKind.Object)
        {
            var o = new JsonObject();
            foreach (var p in value.EnumerateObject().OrderBy(p => p.Name, StringComparer.Ordinal))
            {
                o.Add(p.Name, Canonical(p.Value));
            }

            return o;
        }
        if (value.ValueKind == JsonValueKind.Array)
        {
            return new JsonArray(value.EnumerateArray().Select(Canonical).ToArray());
        }

        return JsonNode.Parse(value.GetRawText());
    }

    public static partial JsonObject WritePart(Part part) => part switch
    {
        HandoffPart p => new() { ["type"] = "handoff", ["text"] = p.Text },
        SkillPart p => new()
        {
            ["type"] = "skill", ["name"] = p.Name, ["body"] = p.Body, ["primary"] = p.Primary, ["pending"] = p.Pending,
            ["tools"] = new JsonArray(p.Tools.Select(name => (JsonNode?)JsonValue.Create(name)).ToArray())
        },
        TextPart p => Text(p),
        BlobPart p => new()
        {
            ["type"] = "blob",
            ["mime"] = p.Mime,
            ["ref"] = WriteRef(p.Ref)
        },
        OpaquePart p => new()
        {
            ["type"] = "opaque",
            ["from"] = WriteProvenance(p.From),
            ["data"] = Canonical(p.Data)
        },
        ToolCallPart p => ToolCall(p),
        ToolResultPart p => new()
        {
            ["type"] = "tool_result",
            ["call_id"] = p.CallID,
            ["parts"] = Parts(p.Parts),
            ["tool_references"] = new JsonArray(p.ToolReferences.Select(name => (JsonNode?)JsonValue.Create(name)).ToArray()),
            ["is_error"] = p.IsError
        },
        RedactedPart p => Redacted(p),
        _ => throw new FormatException("Unknown part type")
    };
    private static JsonObject Text(TextPart part)
    {
        var value = new JsonObject { ["type"] = "text", ["text"] = part.Text };
        if (part.Opaque is JsonElement opaque && part.From is not null)
        {
            value["from"] = WriteProvenance(part.From);
            value["opaque"] = Canonical(opaque);
        }
        return value;
    }

    private static JsonObject ToolCall(ToolCallPart p)
    {
        var o = new JsonObject
        {
            ["type"] = "tool_call",
            ["call_id"] = p.CallID,
            ["from"] = WriteProvenance(p.From),
            ["name"] = p.Name,
            ["args"] = Canonical(p.Args)
        };
        if (p.Opaque is JsonElement opaque)
        {
            o["opaque"] = Canonical(opaque);
        }

        return o;
    }

    private static JsonObject Redacted(RedactedPart p)
    {
        var o = new JsonObject
        {
            ["type"] = "redacted",
            ["kind"] = p.Kind,
            ["bytes"] = p.Bytes,
            ["reason"] = p.Reason
        };
        if (p.Ref is not null)
        {
            o["ref"] = WriteRef(p.Ref);
        }

        return o;
    }

    public static partial Part ReadPart(JsonElement value) => String(value, "type") switch
    {
            "handoff" => new HandoffPart(String(value, "text")),
            "skill" => new SkillPart(String(value, "name"), String(value, "body"),
                value.GetProperty("tools").EnumerateArray().Select(item => item.GetString() ?? throw new FormatException("Null tool name")), Flag(value, "primary"), Flag(value, "pending")),
        "text" => new TextPart(String(value, "text"),
            value.TryGetProperty("from", out var origin) ? ReadProvenance(origin) : null,
            value.TryGetProperty("opaque", out var textOpaque) ? textOpaque : null),
        "blob" => new BlobPart(String(value, "mime"), ReadRef(value.GetProperty("ref"))),
        "opaque" => new OpaquePart(ReadProvenance(value.GetProperty("from")), value.GetProperty("data")),
        "tool_call" => new ToolCallPart(String(value, "call_id"), ReadProvenance(value.GetProperty("from")), String(value, "name"), value.GetProperty("args"), value.TryGetProperty("opaque", out var op) ? op : null),
        "tool_result" => new ToolResultPart(String(value, "call_id"), Parts(value), Flag(value, "is_error"), toolReferences: value.TryGetProperty("tool_references", out var references) ? references.EnumerateArray().Select(item => item.GetString()!) : null),
        "redacted" => new RedactedPart(String(value, "kind"), value.GetProperty("bytes").GetInt64(), String(value, "reason"), value.TryGetProperty("ref", out var r) ? ReadRef(r) : null),
        var t => throw new FormatException($"Unknown part type: {t}")
    };
    public static partial string WriteEvent(Event value)
    {
        var o = new JsonObject
        {
            ["seq"] = value.Seq,
            ["type"] = Label(value.Kind)
        };
        if (value.TaskID.Length != 0)
        {
            o["task_id"] = value.TaskID;
        }
        if (value.Time is DateTimeOffset time)
        {
            o["time"] = time.ToString("O");
        }

        switch (value.Kind)
        {
            case EventKind.HintReceived:
            case EventKind.MessageReceived:
                o["message"] = new JsonObject
                {
                    ["actor"] = value.Actor is Actor actor ? Label(actor) : null,
                    ["parts"] = Parts(value.Parts)
                };
                break;
            case EventKind.RequestSent:
                o["request"] = new JsonObject { ["to"] = value.From is null ? null : WriteProvenance(value.From) };
                break;
            case EventKind.ResponseStarted:
            case EventKind.ResponseEnded:
                var response = new JsonObject { ["parts"] = Parts(value.Parts) };
                if (value.From is not null)
                {
                    response["from"] = WriteProvenance(value.From);
                }

                if (value.Usage is not null)
                {
                    response["usage"] = WriteUsage(value.Usage);
                }

                if (value.ResponseID != 0)
                {
                    response["response_id"] = value.ResponseID;
                    response["part_ids"] = new JsonArray(value.PartIDs.Select(id => (JsonNode?)JsonValue.Create(id)).ToArray());
                }
                o["response"] = response;
                break;
            case EventKind.ToolCalled:
            case EventKind.ToolStarted:
            case EventKind.ToolReturned:
            case EventKind.ToolObserved:
                var tool = new JsonObject { ["call_id"] = value.CallID };
                if (value.Kind is EventKind.ToolCalled or EventKind.ToolStarted)
                {
                    tool["name"] = value.Name;
                    tool["args"] = value.Args is JsonElement a ? Canonical(a) : null;
                }
                else
                {
                    tool["parts"] = Parts(value.Parts);
                    tool["is_error"] = value.IsError;
                    if (!value.ToolReferences.IsEmpty)
                    {
                        tool["tool_references"] = new JsonArray(value.ToolReferences.Select(name => (JsonNode?)JsonValue.Create(name)).ToArray());
                    }
                }
                if (value.Job is not null)
                {
                    tool["job"] = new JsonObject
                    {
                        ["handle"] = value.Job.Handle,
                        ["status"] = value.Job.Status,
                        ["output"] = new JsonObject { ["kind"] = "handle", ["locator"] = value.Job.OutputPath },
                        ["bytes"] = value.Job.Bytes,
                        ["truncated"] = value.Job.Truncated,
                        ["exit_code"] = value.Job.ExitCode,
                        ["cwd"] = value.Job.Cwd
                    };
                }
                o["tool"] = tool;
                break;
            case EventKind.CompressionLaunched:
                o["compression"] = value.Compression is null ? null : WriteCompressionLaunch(value.Compression);
                break;
            case EventKind.CompressionFinished:
                o["outcome"] = value.Outcome is null ? null : WriteCompressionOutcome(value.Outcome);
                break;
            case EventKind.MemoryPopulated:
                o["memory"] = value.Memory is null ? null : WriteMemory(value.Memory);
                break;
            case EventKind.MemoryRetired:
                o["retirement"] = value.Retirement is null ? null : WriteMemoryRetirement(value.Retirement);
                break;
            case EventKind.HandoffRecorded:
            case EventKind.RuntimeConfigured:
            case EventKind.SkillLoaded:
                o["parts"] = Parts(value.Parts);
                break;
            case EventKind.SkillUnloadPending:
                o["name"] = value.Name;
                break;
            case EventKind.AgentPaused:
            case EventKind.AgentResumed:
            case EventKind.AgentShutdown:
            case EventKind.TurnInterrupted:
            case EventKind.TurnEnded:
                o["reason"] = value.Reason;
                break;
            case EventKind.JobKilled:
                o["job"] = new JsonObject
                {
                    ["handle"] = value.Job!.Handle,
                    ["status"] = value.Job.Status,
                    ["reason"] = value.Reason,
                    ["output"] = new JsonObject { ["kind"] = "handle", ["locator"] = value.Job.OutputPath },
                    ["bytes"] = value.Job.Bytes
                };
                break;
            case EventKind.Redacted:
                o["redact"] = new JsonObject
                {
                    ["from"] = value.RedactFrom,
                    ["to"] = value.RedactTo,
                    ["level"] = value.Level is Redaction level ? Label(level) : null,
                    ["reason"] = value.Reason,
                    ["replacement"] = Parts(value.Parts)
                };
                break;
            case EventKind.ErrorOccurred:
                o["error"] = new JsonObject { ["message"] = value.Error };
                break;
        }
        return o.ToJsonString();
    }

    public static partial Event ReadEvent(string line)
    {
        using var doc = JsonDocument.Parse(line);
        var root = doc.RootElement;
        var seq = root.GetProperty("seq").GetInt64();
        if (seq <= 0)
        {
            throw new FormatException("Sequence must be positive");
        }

        DateTimeOffset? time = root.TryGetProperty("time", out var stamp) ? stamp.GetDateTimeOffset() : null;
        var kind = EnumValue<EventKind>(String(root, "type"));
        string taskID = "";
        if (root.TryGetProperty("task_id", out var identity))
        {
            // Presence is explicit correlation, never the legacy empty default.
            // A damaged ID must not turn a held removal into an immediate cut.
            if (identity.ValueKind != JsonValueKind.String || string.IsNullOrWhiteSpace(identity.GetString())
                || kind is not (EventKind.Redacted or EventKind.MemoryPopulated or EventKind.MemoryRetired))
            {
                throw new FormatException("Invalid compression task identity");
            }
            taskID = identity.GetString()!;
        }
        switch (kind)
        {
            case EventKind.HintReceived:
            case EventKind.MessageReceived:
                if (!root.TryGetProperty("message", out var m))
                {
                    return new(kind, seq, time);
                }
                Actor? actor = m.TryGetProperty("actor", out var actorValue) && actorValue.ValueKind != JsonValueKind.Null
                    ? EnumValue<Actor>(actorValue.GetString()!) : null;
                return new(kind, seq, time, actor, Parts(m));
            case EventKind.RequestSent:
                var req = root.GetProperty("request");
                return new(kind, seq, time, from: ReadProvenance(req.GetProperty("to")));
            case EventKind.ResponseStarted:
            case EventKind.ResponseEnded:
                var response = root.GetProperty("response");
                return new(kind, seq, time, parts: Parts(response), from: response.TryGetProperty("from", out var origin) ? ReadProvenance(origin) : null,
                    usage: response.TryGetProperty("usage", out var u) ? ReadUsage(u) : null,
                    responseID: response.TryGetProperty("response_id", out var rid) ? rid.GetInt64() : 0,
                    partIDs: response.TryGetProperty("part_ids", out var ids)
                        ? ids.EnumerateArray().Select(id => id.GetInt32()) : null);
            case EventKind.ToolCalled:
            case EventKind.ToolStarted:
            case EventKind.ToolReturned:
            case EventKind.ToolObserved:
                var tool = root.GetProperty("tool");
                return new(kind, seq, time, parts: Parts(tool), callID: String(tool, "call_id"), name: OptionalString(tool, "name"),
                    args: tool.TryGetProperty("args", out var args) ? args : null, isError: Flag(tool, "is_error"),
                    toolReferences: tool.TryGetProperty("tool_references", out var toolRefs) ? toolRefs.EnumerateArray().Select(item => item.GetString()!) : null,
                    job: tool.TryGetProperty("job", out var job) ? new JobInfo(
                        job.GetProperty("handle").GetInt32(), String(job, "status"),
                        String(job.GetProperty("output"), "locator"), job.GetProperty("bytes").GetInt64(),
                        Flag(job, "truncated"), job.TryGetProperty("exit_code", out var exit) && exit.ValueKind == JsonValueKind.Number ? exit.GetInt32() : null, OptionalString(job, "cwd")) : null);
            case EventKind.CompressionLaunched:
                return new(kind, seq, time, compression: root.TryGetProperty("compression", out var launch) && launch.ValueKind != JsonValueKind.Null
                    ? ReadCompressionLaunch(launch) : null);
            case EventKind.CompressionFinished:
                return new(kind, seq, time, outcome: root.TryGetProperty("outcome", out var outcome) && outcome.ValueKind != JsonValueKind.Null
                    ? ReadCompressionOutcome(outcome) : null);
            case EventKind.MemoryPopulated:
                return new(kind, seq, time, memory:
                    root.TryGetProperty("memory", out var memory) && memory.ValueKind != JsonValueKind.Null
                        ? ReadMemory(memory) : null, taskID: taskID);
            case EventKind.MemoryRetired:
                return new(kind, seq, time, retirement:
                    root.TryGetProperty("retirement", out var retirement) && retirement.ValueKind != JsonValueKind.Null
                        ? ReadMemoryRetirement(retirement) : null, taskID: taskID);
            case EventKind.HandoffRecorded:
            case EventKind.RuntimeConfigured:
            case EventKind.SkillLoaded:
                return new Event(kind, seq, time, parts: Parts(root));
            case EventKind.SkillUnloadPending:
                return new Event(kind, seq, time, name: OptionalString(root, "name"));
            case EventKind.AgentPaused:
            case EventKind.AgentResumed:
            case EventKind.AgentShutdown:
            case EventKind.TurnInterrupted:
            case EventKind.TurnEnded:
                return new(kind, seq, time, reason: String(root, "reason"));
            case EventKind.JobKilled:
                var killed = root.GetProperty("job");
                return new(kind, seq, time, reason: String(killed, "reason"), job: new JobInfo(
                    killed.GetProperty("handle").GetInt32(), String(killed, "status"),
                    String(killed.GetProperty("output"), "locator"), killed.GetProperty("bytes").GetInt64()));
            case EventKind.Redacted:
                if (!root.TryGetProperty("redact", out var r))
                {
                    return new(kind, seq, time, taskID: taskID);
                }
                Redaction? level = r.TryGetProperty("level", out var levelValue) && levelValue.ValueKind != JsonValueKind.Null
                    ? EnumValue<Redaction>(levelValue.GetString()!) : null;
                return new(kind, seq, time,
                    parts: r.TryGetProperty("replacement", out var replacement) ? replacement.EnumerateArray().Select(ReadPart) : [],
                    redactFrom: r.TryGetProperty("from", out var first) ? first.GetInt64() : 0,
                    redactTo: r.TryGetProperty("to", out var last) ? last.GetInt64() : 0,
                    level: level, reason: OptionalString(r, "reason"), taskID: taskID);
            default:
                var err = root.GetProperty("error");
                return new(kind, seq, time, error: err.ValueKind == JsonValueKind.String ? err.GetString()! : String(err, "message"));
        }
    }

    public static partial string Text(IEnumerable<Part> parts) => string.Concat(parts.Select(p => p switch
    {
        TextPart t => t.Text,
        HandoffPart h => h.Text,
        RedactedPart r => $"[redacted {r.Kind}: {r.Bytes} bytes; {r.Reason}]",
        ToolResultPart r => Text(r.Parts),
        _ => ""
    }));
}
