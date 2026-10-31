using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class Tool
{
    public partial Tool(Agent parent, ToolKind kind)
    {
        if (kind == ToolKind.KeepToolResults && !parent.Configuration().StubsToolResults)
        {
            throw new ArgumentException("keep_tool_results requires explicit selected-model stubbing policy", nameof(kind));
        }
        this.kind = kind;
        (name, description, schema) = Define(kind);
        if (kind == ToolKind.Submit)
        {
            schema = parent.SubmissionSchema();
        }
        parent.AddTool(this);
    }

    public partial Tool(Agent parent, string name, string description, JsonElement schema,
        Func<Job, JsonElement, CancellationToken, Task<string>> handler, bool waitForCompletion)
    {
        if (string.IsNullOrWhiteSpace(name) || schema.ValueKind != JsonValueKind.Object)
        {
            throw new ArgumentException("Custom tool needs a name and object schema");
        }
        this.kind = ToolKind.Custom;
        this.waitForCompletion = waitForCompletion;
        this.name = name;
        this.description = description;
        this.schema = schema.Clone();
        this.handler = handler ?? throw new ArgumentNullException(nameof(handler));
        parent.AddTool(this);
    }

    public partial bool IsControl() => kind is ToolKind.WaitForJob or ToolKind.SendInput or ToolKind.KillJob or ToolKind.ToolLimits or ToolKind.LoadSkill or ToolKind.UnloadSkill or ToolKind.MicroHandoff or ToolKind.KeepToolResults;
    internal partial bool IsHandoff() => kind == ToolKind.MicroHandoff;
    internal partial bool IsKeepResults() => kind == ToolKind.KeepToolResults;

    internal partial Tool(Agent parent, McpClient client, McpToolDefinition definition)
    {
        kind = ToolKind.Remote;
        name = definition.Name;
        description = definition.Description;
        schema = definition.InputSchema;
        parent.AddTool(this);
        try
        {
            client.AddTool(this);
        }
        catch
        {
            parent.RemoveTool(this);
            throw;
        }
    }

    public partial McpClient? Remote() => remote;
    internal partial bool HasStructuredResult() => kind == ToolKind.Remote || waitForCompletion;
    internal partial bool WaitsForCompletion() => waitForCompletion;

    public partial Agent Parent() => parent ?? throw new InvalidOperationException("Tool is detached");
    public partial string Name() => name;

    public partial void Destroy()
    {
        // Either strong parent may initiate deletion. Serialize the complete
        // unlink, not only the flag, so a second caller joins logical cleanup.
        lock (destructionGate)
        {
            if (destroyed)
            {
                return;
            }
            destroyed = true;
            remote?.RemoveTool(this);
            parent?.RemoveTool(this);
        }
    }

    private static (string, string, JsonElement) Define(ToolKind kind)
    {
        var properties = new JsonObject();
        var required = new JsonArray();
        void Field(string name, string type, string explanation, bool mandatory = false)
        {
            properties[name] = new JsonObject
            {
                ["type"] = type,
                ["description"] = explanation
            };
            if (mandatory)
            {
                required.Add(name);
            }
        }
        string name;
        string description;
        switch (kind)
        {
            case ToolKind.ReadFile:
                name = "read_file";
                description = "Read a UTF-8 text file, optionally selecting a one-based inclusive line range.";
                Field("path", "string", "File path.", true);
                Field("start_line", "integer", "First line, default or zero means 1.");
                Field("end_line", "integer", "Last line, default or zero means end of file.");
                Field("max_bytes", "integer", "UTF-8 byte cap after line selection, default or zero means 32768; truncation is reported.");
                break;
            case ToolKind.ListDirectory:
                name = "list_directory";
                description = "List immediate directory entries sorted by name; directories end in /.";
                Field("path", "string", "Directory path.", true);
                break;
            case ToolKind.SearchFiles:
                name = "search_files";
                description = "Search text files recursively using a .NET regular expression; return grep-style lines and merged context windows.";
                Field("path", "string", "File or directory path.", true);
                Field("pattern", "string", "Regular expression.", true);
                Field("context_lines", "integer", "Nonnegative context radius, default 0.");
                break;
            case ToolKind.RunCommand:
                name = "run_command";
                description = "Run a fresh PTY shell. Merged output and exit_code stay on disk. A running report is a live job, not failure. No implicit timeout.";
                Field("command", "string", "Shell command.", true);
                Field("cwd", "string", "Working directory for this call only.");
                break;
            case ToolKind.WriteFile:
                name = "write_file";
                description = "Create or write a UTF-8 file, creating parent directories. Existing files require overwrite:true or append:true.";
                Field("path", "string", "File path.", true);
                Field("content", "string", "Exact text to write.", true);
                Field("overwrite", "boolean", "Allow replacing an existing file, default false.");
                Field("append", "boolean", "Append instead of replacing, default false.");
                break;
            case ToolKind.EditFile:
                name = "edit_file";
                description = "Replace one exact unique nonempty old_text anchor. Zero or multiple sites refuse without changing bytes; read and narrow the anchor.";
                Field("path", "string", "File path.", true);
                Field("old_text", "string", "Exact nonempty anchor.", true);
                Field("new_text", "string", "Replacement text, possibly empty.", true);
                break;
            case ToolKind.Submit:
                name = "submit";
                description = "Validate and durably submit this delegated Agent's whole argument object. Ordinary final text is not submission. Resolve or shut down unresolved children first; standalone Agents finish in chat.";
                // The constructor replaces this placeholder with the owning
                // Agent's exact immutable contract, not a result-wrapper schema.
                break;
            case ToolKind.SendMessageToParent:
                name = "send_message_to_parent";
                description = "Ask the owning Agent, or the human for a root Agent, and wait for its reply. A running Job is still waiting, not submitted. Use wait_for_job on its handle; observation does not cancel the question.";
                Field("content", "string", "Question for the parent or human.", true);
                break;
            case ToolKind.RespondToParentMessage:
                name = "respond_to_parent_message";
                description = "Answer an owned child's pending question. Omit message_id only when that child has exactly one pending question. Does not submit or destroy the child.";
                Field("agent_id", "string", "Direct child ID.", true);
                Field("response", "string", "Answer text.", true);
                Field("message_id", "integer", "Receiver-scoped pending question ID.");
                break;
            case ToolKind.InterruptAgent:
                name = "interrupt_agent";
                description = "Interrupt a direct child's current turn without destroying the Agent or its independent Jobs. now interrupts the current turn immediately; after_tool stops at the next tool-report boundary. Omitted mode means now. Delivery is acknowledged, not a drain barrier.";
                Field("agent_id", "string", "Direct owned child ID.", true);
                Field("mode", "string", "now (default) or after_tool.");
                properties["mode"]!["enum"] = new JsonArray("now", "after_tool");
                break;
            case ToolKind.SendMessage:
            case ToolKind.AskAgent:
            case ToolKind.WaitForAgent:
                name = kind switch
                {
                    ToolKind.SendMessage => "send_message",
                    ToolKind.AskAgent => "ask_agent",
                    _ => "wait_for_agent"
                };
                description = kind switch
                {
                    ToolKind.SendMessage => "Send live steering to a direct child: a hint when busy, a new turn when idle. Return finalized output since cursor without waiting for the turn to finish.",
                    ToolKind.AskAgent => "Queue a separate question to a direct child and wait for that turn's reliable reply. Cancellation stops observation, not the admitted turn.",
                    _ => "Wait for finalized output after a child-local History cursor, or return when the child is no longer busy. This does not wait for submission and never consumes shared GUI observations."
                };
                Field("agent_id", "string", "Direct owned child ID.", true);
                if (kind != ToolKind.AskAgent)
                {
                    Field("cursor", "integer", "Nonnegative child-local History sequence, no greater than current History; defaults to zero.");
                    properties["cursor"]!["minimum"] = 0;
                }
                if (kind != ToolKind.WaitForAgent)
                {
                    Field("content", "string", "Nonempty message for this child.", true);
                }
                break;
            case ToolKind.WaitForAgentChange:
                name = "wait_for_agent_change";
                description = "Observe fixed direct-child membership for status, finalized History or transient stream progress, removal, questions, pattern match, or lack of progress. Omitted/empty IDs means all current children. Cancellation only stops observation.";
                Field("agent_ids", "array", "Direct child IDs, captured at admission.");
                properties["agent_ids"]!["items"] = new JsonObject { ["type"] = "string" };
                Field("pattern", "string", "Optional regex on finalized output; matching is time bounded.");
                Field("timeout_seconds", "number", "Overall observation deadline; default 300, zero disables.");
                Field("stuck_seconds", "number", "Busy without observed progress; default 15, zero disables.");
                properties["timeout_seconds"]!["minimum"] = 0;
                properties["stuck_seconds"]!["minimum"] = 0;
                break;
            case ToolKind.GetSubmission:
                name = "get_submission";
                description = "Read a direct child's retained typed submission freshly from its owned storage, even after shutdown. Historical data is not current-task completion; this never respawns the child or satisfies a join. Missing or invalid artifacts refuse.";
                Field("agent_id", "string", "Direct owned child ID, not a qualified tree address.", true);
                break;
            case ToolKind.GetSubAgentStatus:
                name = "get_sub_agent_status";
                description = "Read one direct owned child’s actor status, work phase and current-process usage. Selected and measured models are distinct; costs are estimates and unknown usage remains null. This observes without stopping or joining the child.";
                Field("agent_id", "string", "Direct owned child ID.", true);
                break;
            case ToolKind.AgentsStatus:
                name = "agents_status";
                description = "List direct owned children with actor-observed activity, pause state, explicit submission and retained failure. Idle is not submitted; submitted can still be draining. Snapshots do not join or stop children.";
                break;
            case ToolKind.SpawnSubAgent:
                name = "spawn_sub_agent";
                description = "Create an owned child, restore its conversation, and admit task. A running Job reports the child ID while retaining its task reply. Canceling that Job only stops observing; use shutdown_agent to drain or kill_agent to cancel the child. Completion requires explicit submit, not ordinary final text.";
                Field("agent_id", "string", "Unique direct child ID.", true);
                Field("task", "string", "Delegated task.", true);
                Field("model", "string", "Optional compatible same-provider model.");
                Field("submit_schema", "object", "Optional complete bounded submission schema; validated before child creation.");
                break;
            case ToolKind.ShutdownAgent:
                name = "shutdown_agent";
                description = "Wait for a direct child's current and queued turns and foreground controls, then stop and remove its subtree. Paused work is not resumed; later input may extend waiting. Canceling the waiting Job leaves the child alive. Independent Jobs and descendants are cancelled during final teardown; saved files remain. Use kill_agent to cancel without draining.";
                Field("agent_id", "string", "Direct child ID.", true);
                break;
            case ToolKind.KillAgent:
                name = "kill_agent";
                description = "Cancel one direct child's lifetime immediately, join owned work and descendants, and remove the subtree. Does not wait for turns to finish normally. Saved conversations and submissions remain; siblings survive. Once teardown starts it is joined even if observation is cancelled.";
                Field("agent_id", "string", "Direct owned child ID.", true);
                break;
            case ToolKind.JoinAgents:
                name = "join_agents";
                description = "Wait for selected direct children to submit. Omitted or empty agent_ids selects all current children. Pending questions yield a running Job; answer them and wait_for_job on that same handle. Submission is not a shutdown barrier.";
                Field("agent_ids", "array", "Direct child IDs, captured when the join starts.");
                properties["agent_ids"]!["items"] = new JsonObject { ["type"] = "string" };
                break;
            case ToolKind.AgentStatus:
                name = "agent_status";
                description = "Report selected model, last completed response and session token counts, session cache hit rate, and estimated USD cost. Unknown counts or prices remain unknown; this is not a billing statement.";
                break;
            case ToolKind.KeepToolResults:
                name = "keep_tool_results";
                description = "Keep the previous tool batch whole instead of once-only result stubbing. Call in the consuming response; takes no arguments. Later budget/explicit clearing still applies.";
                break;
            case ToolKind.MicroHandoff:
                name = "micro_handoff";
                description = "Record a continuation note and clear tool traffic after this batch finishes. Replaces older active handoff notes, not dialogue or skill instructions.";
                Field("text", "string", "Nonempty continuation note: decisions, evidence limits, and next work.", true);
                break;
            case ToolKind.LoadSkill:
            case ToolKind.UnloadSkill:
                name = kind == ToolKind.LoadSkill ? "load_skill" : "unload_skill";
                Field("name", "string", "Exact skill name.", true);
                description = kind == ToolKind.LoadSkill
                    ? "Activate an offered skill and its dependencies."
                    : "Revoke a skill’s tools now while retaining its recorded instruction manual.";
                break;
            case ToolKind.WaitForJob:
                name = "wait_for_job";
                description = "Observe a job without cancelling it. Keep its handle while running.";
                Field("handle", "integer", "Existing job handle.", true);
                break;
            case ToolKind.SendInput:
                name = "send_input";
                description = "Submit a line to a live PTY, adding newline when missing, and observe fresh output.";
                Field("handle", "integer", "Existing job handle.", true);
                Field("input", "string", "One terminal input line.", true);
                break;
            case ToolKind.KillJob:
                name = "kill_job";
                description = "Explicitly cancel and drain a job and its process group.";
                Field("handle", "integer", "Existing job handle.", true);
                break;
            case ToolKind.ToolLimits:
                name = "tool_limits";
                description = "Set observation options for the very next call, including controls. Does not cancel work.";
                break;
            default:
                throw new ArgumentOutOfRangeException(nameof(kind));
        }
        var schema = new JsonObject
        {
            ["type"] = "object",
            ["properties"] = properties,
            ["required"] = required,
            ["additionalProperties"] = false
        };
        return (name, description, JsonSerializer.SerializeToElement(schema));
    }

    public partial JsonObject Declaration(Vendor vendor)
    {
        var parameters = JsonCodec.Canonical(schema);
        // Observer options are projected, not added to the execution schema.
        // Jobs strips them before calling any built-in or custom body.
        // JSON Schema permits an object without a properties member. Create
        // the projection slot without changing the authoritative tool schema.
        if (parameters!["properties"] is null)
        {
            parameters["properties"] = new JsonObject();
        }
        var properties = parameters["properties"]!.AsObject();
        properties["ai_callback_delay"] = new JsonObject { ["type"] = "number", ["description"] = "Seconds before waking; default 3, zero immediate. Does not cancel." };
        properties["ai_callback_pattern"] = new JsonObject { ["type"] = "string", ["description"] = "Regex matching only fresh output." };
        properties["max_output_bytes"] = new JsonObject { ["type"] = "integer", ["description"] = "Positive inline report cap; default 16384. Full output stays on disk." };
        if (vendor == Vendor.Gemini)
        {
            parameters = GeminiSchema(JsonSerializer.SerializeToElement(parameters));
        }
        var definition = new JsonObject
        {
            ["name"] = name,
            ["description"] = description,
            [vendor == Vendor.Anthropic ? "input_schema" : "parameters"] = parameters
        };
        if (vendor == Vendor.OpenAI)
        {
            return new JsonObject
            {
                ["type"] = "function",
                ["function"] = definition
            };
        }
        return definition;
    }

    // Apply the allow-list to schema nodes, not to property names. A user's
    // property may legitimately be named "additionalProperties" or "format".
    private static JsonObject GeminiSchema(JsonElement schema)
    {
        // The provider's reduced declaration format has no boolean schema
        // nodes. Omit that constraint here, just as we omit unsupported object
        // keywords below. The owning submission contract still validates the
        // original schema before effects; projection never widens admission.
        if (schema.ValueKind is JsonValueKind.True or JsonValueKind.False)
        {
            return new JsonObject();
        }
        var projected = new JsonObject();
        foreach (var property in schema.EnumerateObject())
        {
            switch (property.Name)
            {
                case "type":
                case "description":
                case "enum":
                case "required":
                    projected[property.Name] = JsonCodec.Canonical(property.Value);
                    break;
                case "items":
                    projected["items"] = GeminiSchema(property.Value);
                    break;
                case "properties":
                    var properties = new JsonObject();
                    foreach (var child in property.Value.EnumerateObject())
                    {
                        properties[child.Name] = GeminiSchema(child.Value);
                    }
                    projected["properties"] = properties;
                    break;
            }
        }
        return projected;
    }

    internal partial void Validate(JsonElement args)
    {
        if (kind == ToolKind.Submit)
        {
            // Structured submission permits nested/boolean schemas, exact
            // integers and additional properties beyond the shallow tool subset.
            Parent().ValidateSubmission(args);
            return;
        }
        if (args.ValueKind != JsonValueKind.Object)
        {
            throw new ArgumentException("Arguments must be a JSON object.");
        }
        // A no-argument object schema may omit properties entirely. Our
        // declared-argument validator still refuses any undeclared input.
        schema.TryGetProperty("properties", out var properties);
        var names = new HashSet<string>(StringComparer.Ordinal);
        foreach (var arg in args.EnumerateObject())
        {
            if (!names.Add(arg.Name) || properties.ValueKind != JsonValueKind.Object
                || !properties.TryGetProperty(arg.Name, out var field))
            {
                throw new ArgumentException("Unknown or duplicate argument: " + arg.Name);
            }
            var valid = field.GetProperty("type").GetString() switch
            {
                "string" => arg.Value.ValueKind == JsonValueKind.String,
                "boolean" => arg.Value.ValueKind is JsonValueKind.True or JsonValueKind.False,
                "integer" => arg.Value.ValueKind == JsonValueKind.Number && arg.Value.TryGetInt64(out _),
                "number" => arg.Value.ValueKind == JsonValueKind.Number,
                "object" => arg.Value.ValueKind == JsonValueKind.Object,
                "array" => arg.Value.ValueKind == JsonValueKind.Array,
                _ => false
            };
            if (!valid)
            {
                throw new ArgumentException("Wrong argument type: " + arg.Name);
            }
        }
        // In JSON Schema, omitting required means every property is optional.
        if (schema.TryGetProperty("required", out var requiredNames))
        {
            foreach (var required in requiredNames.EnumerateArray())
            {
                if (!args.TryGetProperty(required.GetString()!, out _))
                {
                    throw new ArgumentException("Missing argument: " + required.GetString());
                }
            }
        }
    }

    private static string String(JsonElement args, string name, string fallback = "")
    {
        return args.TryGetProperty(name, out var value) ? value.GetString()! : fallback;
    }

    private static int Integer(JsonElement args, string name, int fallback)
    {
        return args.TryGetProperty(name, out var value) ? value.GetInt32() : fallback;
    }

    private static bool Boolean(JsonElement args, string name)
    {
        return args.TryGetProperty(name, out var value) && value.GetBoolean();
    }

    public partial Task<ToolResultPart> ExecuteAsync(ToolCallPart call, CancellationToken cancellation)
    {
        return Parent().DispatchAsync(call, cancellation);
    }

    internal partial async Task<ToolResultPart> ExecuteJobAsync(Job job, ToolCallPart call, CancellationToken cancellation)
    {
        _ = Parent();
        try
        {
            cancellation.ThrowIfCancellationRequested();
            if (kind == ToolKind.Remote)
            {
                var client = remote ?? throw new IOException("MCP tool is disconnected.");
                return await client.CallAsync(call, cancellation);
            }
            var text = kind switch
            {
                ToolKind.AgentStatus => Status(job),
                ToolKind.GetSubAgentStatus => await DetailedChildStatusAsync(job, call.Args, cancellation),
                ToolKind.JoinAgents => await JoinAgentsAsync(job, call.Args, cancellation),
                ToolKind.Submit or ToolKind.SendMessageToParent or ToolKind.RespondToParentMessage
                    or ToolKind.ShutdownAgent or ToolKind.KillAgent or ToolKind.SpawnSubAgent or ToolKind.AgentsStatus or ToolKind.InterruptAgent or ToolKind.GetSubmission or ToolKind.SendMessage or ToolKind.AskAgent
                    or ToolKind.WaitForAgent or ToolKind.WaitForAgentChange => await ChildOperationAsync(job, call.Args, cancellation),
                ToolKind.ReadFile => ReadFile(call.Args),
                ToolKind.ListDirectory => ListDirectory(call.Args),
                ToolKind.SearchFiles => SearchFiles(call.Args),
                ToolKind.RunCommand => await job.RunCommandAsync(call.Args, cancellation),
                ToolKind.Custom => await handler!(job, call.Args, cancellation),
                ToolKind.WriteFile => WriteFile(call.Args),
                ToolKind.EditFile => EditFile(call.Args),
                _ => throw new InvalidOperationException("Unknown tool kind")
            };
            return new ToolResultPart(call.CallID, [new TextPart(text)]);
        }
        catch (Exception error) when (error is ArgumentException or IOException or UnauthorizedAccessException or System.ComponentModel.Win32Exception)
        {
            // These failures are useful model feedback, unlike a process crash.
            // Provider credentials are never included in this diagnostic path.
            return new ToolResultPart(call.CallID, [new TextPart(error.Message)], isError: true);
        }
    }
}
