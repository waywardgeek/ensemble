using System.Threading.Channels;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

// The executable is composition and input/output only. It does not own a second
// transcript, replay reducer, transport, or token counter.
return await Run(args);

static string Env(params string[] names) => names.Select(Environment.GetEnvironmentVariable).FirstOrDefault(s => !string.IsNullOrEmpty(s)) ?? "";
static Configuration Configuration(bool disableStreaming = false)
{
    var vendor = Env("LLM_VENDOR") switch
    {
        "" or "anthropic" => Vendor.Anthropic,
        "openai" => Vendor.OpenAI,
        "gemini" => Vendor.Gemini,
        _ => throw new ArgumentException("Unknown LLM_VENDOR")
    };
    var prefix = vendor switch
    {
        Vendor.Anthropic => "ANTHROPIC",
        Vendor.OpenAI => "OPENAI",
        _ => "GEMINI"
    };
    var model = Env("LLM_MODEL", prefix + "_MODEL");
    var url = Env("LLM_BASE_URL", prefix + "_BASE_URL", prefix + "_API_URL");
    if (url == "")
    {
        url = vendor switch
        {
            Vendor.Anthropic => "https://api.anthropic.com",
            Vendor.OpenAI => "https://api.openai.com",
            _ => "https://generativelanguage.googleapis.com"
        };
    }

    var requireExplanation = Env("LLM_REQUIRE_TOOL_EXPLANATION") switch
    {
        "" or "false" => false,
        "true" => true,
        _ => throw new ArgumentException("LLM_REQUIRE_TOOL_EXPLANATION must be true or false")
    };
    return new(vendor, model == "" ? "offline" : model, url, Env("LLM_API_KEY", prefix + "_API_KEY"), Env("LLM_RESOLVED_MODEL"), requireExplanation,
        disableStreaming: disableStreaming || Env("EN_DISABLE_STREAMING") == "1");
}
static void Print(JsonObject value)
{
    Console.WriteLine(value.ToJsonString());
    Console.Out.Flush();
}
static async Task<int> Run(string[] args)
{
    Application? app = null;
    StreamWriter? apiTrace = null;
    try
    {
        bool disableStreaming = args.Contains("-no-stream");
        args = args.Where(value => value != "-no-stream").ToArray();
        if (args.Length == 0 || args[0] == "chat")
        {
            string tracePath = Env("CH06_API_LOG");
            apiTrace = new StreamWriter(tracePath.Length == 0 ? "api.log" : tracePath, append: true);
        }
        app = new Application(Console.Error, apiTrace);
        if (args.Length > 2)
        {
            throw new ArgumentException("Usage: ch06 [-no-stream] [chat | dump [logfile] | render logfile]");
        }

        var config = Configuration(disableStreaming);
        var agent = new Agent(app, config, Path.Combine(Environment.CurrentDirectory, "cr", "io"));
        agent.InstallTools();
        var log = Env("CH02_LOG");
        if (log == "")
        {
            log = "ch02.jsonl";
        }

        var command = args.FirstOrDefault() ?? "";
        if (command is "render" or "dump")
        {
            if (command == "render" && args.Length != 2)
            {
                throw new ArgumentException("render requires a log path");
            }

            agent.History().Load(args.Length == 2 ? args[1] : log);
            if (command == "dump")
            {
                agent.History().Dump(Console.Out);
            }
            else
            {
                Print(agent.Engine().Render(agent.History().View()));
            }

            return 0;
        }
        if (command != "" && command != "chat")
        {
            throw new ArgumentException("Unknown command");
        }

        if (args.Length > 1)
        {
            throw new ArgumentException("Unexpected command argument");
        }

        agent.History().StartLog(log);
        JsonObject usage;
        if (command == "chat")
        {
            // Progress is lossy by contract, so stderr is a preview. Stdout
            // receives the reliable completed answer exactly once per turn.
            Task preview = WriteChatProgress(agent);
            try
            {
                while (true)
                {
                    Console.Error.Write("You: ");
                    Console.Error.Flush();
                    var line = await Console.In.ReadLineAsync();
                    if (line is null or "/quit" or "quit" or "exit")
                    {
                        break;
                    }
                    if (string.IsNullOrWhiteSpace(line))
                    {
                        continue;
                    }
                    RequireKey(config);
                    string answer = await agent.SendAsync(line);
                    Console.WriteLine(answer);
                    Console.Out.Flush();
                }
                usage = JsonCodec.WriteUsage(agent.History().View().Usage);
            }
            finally
            {
                try
                {
                    agent.Destroy();
                }
                finally
                {
                    await preview;
                }
            }
        }
        else
        {
            usage = await RunJson(agent, config);
        }
        if (command == "chat")
        {
            Console.Error.WriteLine("Usage: " + usage.ToJsonString());
        }
        else
        {
            Print(new JsonObject { ["usage"] = usage });
        }

        return 0;
    }
    catch (Exception ex) when (ex is ArgumentException or FormatException or InvalidOperationException or JsonException or IOException or NotSupportedException or KeyNotFoundException)
    {
        Console.Error.WriteLine(ex.Message);
        return 1;
    }
    finally
    {
        // Cascade to History's writer and Engine's HTTP client on success or failure.
        try
        {
            app?.Destroy();
        }
        finally
        {
            apiTrace?.Dispose();
        }
    }
}
// Read input independently of request completion. This queue owns only pending
// output Tasks; the Agent owns scheduling, History, and every reliable reply.
// Display only. Neither streamed argument fragments nor display text may
// trigger tools or supply the reliable response returned by SendAsync.
static async Task WriteChatProgress(Agent agent)
{
    bool color = !Console.IsErrorRedirected;
    try
    {
        while (true)
        {
            var observation = await agent.WaitAsync(_ => true);
            if (observation.Delta is Chunk chunk)
            {
                string style = chunk.Kind switch
                {
                    ChunkKind.Thinking => "\u001b[2m",
                    ChunkKind.ToolCall => "\u001b[33m",
                    _ => ""
                };
                string text = chunk.Kind == ChunkKind.ToolCall
                    ? chunk.ToolName + chunk.ArgsText : chunk.Text;
                if (color && style.Length != 0)
                {
                    Console.Error.Write(style);
                }
                Console.Error.Write(text);
                if (color && style.Length != 0)
                {
                    Console.Error.Write("\u001b[0m");
                }
                Console.Error.Flush();
            }
            else if (observation.Kind == ObservationKind.PartFinal && observation.PartID is not null)
            {
                Console.Error.WriteLine();
            }
        }
    }
    catch (ObjectDisposedException)
    {
        // Destruction closes the queue; queued previews have now been drained.
    }
}

static async Task<JsonObject> RunJson(Agent agent, Configuration config)
{
    var replies = Channel.CreateUnbounded<Task<JsonObject>>(new UnboundedChannelOptions
    {
        SingleReader = true,
        SingleWriter = true
    });
    var writer = WriteReplies(replies.Reader);
    using var observing = new CancellationTokenSource();
    Task? progress = null;
    bool shutdown = false;
    JsonObject? usage = null;
    try
    {
        while (await Console.In.ReadLineAsync() is string line)
        {
            if (string.IsNullOrWhiteSpace(line))
            {
                continue;
            }
            using var doc = JsonDocument.Parse(line);
            var input = doc.RootElement;
            if (input.ValueKind != JsonValueKind.Object)
            {
                throw new FormatException("Expected a JSON input object");
            }
            if (!input.TryGetProperty("kind", out var kindValue))
            {
                if (input.TryGetProperty("user", out var user))
                {
                    Fields(input, "user", "parts");
                    RequireKey(config);
                    // Structured parts take precedence, including an empty list.
                    // The public codec is shared with replay; no CLI media model.
                    Part[] parts = input.TryGetProperty("parts", out var values)
                        ? values.EnumerateArray().Select(JsonCodec.ReadPart).ToArray()
                        : [new TextPart(user.GetString() ?? throw new FormatException("user must be a string"))];
                    replies.Writer.TryWrite(AssistantReply(agent.SubmitAsync(parts)));
                }
                else
                {
                    Fields(input, "ephemeral");
                    string text = RequiredText(input, "ephemeral");
                    await agent.PostAsync(new AgentInput(InputKind.Ephemeral, text));
                    Print(new JsonObject { ["ack"] = "ephemeral" });
                }
                continue;
            }

            string kind = kindValue.GetString() ?? throw new FormatException("kind must be a string");
            // Only explicit live mode enables progress output. Legacy users
            // still receive exactly one assistant record per request.
            progress ??= WriteProgress(agent, observing.Token);
            switch (kind)
            {
                case "prompt":
                case "hint":
                    Fields(input, "kind", "text");
                    RequireKey(config);
                    var messageKind = kind == "prompt" ? InputKind.Message : InputKind.Hint;
                    await agent.PostAsync(new AgentInput(messageKind, RequiredText(input, "text")));
                    break;
                case "pause":
                case "resume":
                    Fields(input, "kind");
                    await agent.PostAsync(new AgentInput(kind == "pause" ? InputKind.Pause : InputKind.Resume));
                    break;
                case "interrupt":
                    Fields(input, "kind", "mode");
                    string mode = input.TryGetProperty("mode", out var modeValue)
                        ? modeValue.GetString() ?? "" : "now";
                    var interrupt = mode switch
                    {
                        "now" => InputKind.InterruptNow,
                        "after_tool" => InputKind.InterruptAfterTool,
                        _ => throw new FormatException("interrupt mode must be now or after_tool")
                    };
                    await agent.PostAsync(new AgentInput(interrupt));
                    break;
                case "tool":
                    Fields(input, "kind", "id", "name", "arguments");
                    var call = new ToolCallPart(RequiredText(input, "id"), config.Target(),
                        RequiredText(input, "name"), input.GetProperty("arguments"));
                    replies.Writer.TryWrite(ToolReply(agent.DispatchAsync(call)));
                    break;
                case "shutdown":
                    Fields(input, "kind");
                    await agent.ShutdownAsync();
                    shutdown = true;
                    break;
                default:
                    throw new FormatException("Unknown input kind: " + kind);
            }
            Print(new JsonObject { ["ack"] = kind });
            if (shutdown)
            {
                break;
            }
        }
        if (!shutdown)
        {
            // EOF drains turns; it is not an implicit interrupt. Paused work
            // stays paused, so send resume or shutdown before ending such input.
            await agent.FlushAsync();
        }
    }
    finally
    {
        // On malformed input or read failure, shutdown unblocks pending replies
        // before we join the output writer. Independent Jobs are stopped here.
        try
        {
            await agent.ShutdownAsync();
        }
        finally
        {
            replies.Writer.TryComplete();
            usage = JsonCodec.WriteUsage(agent.History().View().Usage);
            // Closing the writer lets display drain queued observations. Reliable
            // completion alone does not prove that a slower reader caught up.
            try
            {
                agent.Destroy();
            }
            finally
            {
                if (progress is not null)
                {
                    await progress;
                }
                await writer;
            }
        }
    }
    return usage!;
}

static void Fields(JsonElement input, params string[] allowed)
{
    var seen = new HashSet<string>(StringComparer.Ordinal);
    foreach (var field in input.EnumerateObject())
    {
        if (!allowed.Contains(field.Name) || !seen.Add(field.Name))
        {
            throw new FormatException("Unexpected or duplicate input field: " + field.Name);
        }
    }
}

static string RequiredText(JsonElement input, string name)
{
    return input.GetProperty(name).GetString()
        ?? throw new FormatException(name + " must be a string");
}

static async Task<JsonObject> AssistantReply(Task<string> reply)
{
    return new JsonObject { ["assistant"] = await reply };
}

static async Task<JsonObject> ToolReply(Task<ToolResultPart> reply)
{
    return new JsonObject { ["tool_result"] = JsonCodec.WritePart(await reply) };
}

static async Task WriteReplies(ChannelReader<Task<JsonObject>> replies)
{
    bool failed = false;
    await foreach (var reply in replies.ReadAllAsync())
    {
        try
        {
            Print(await reply);
        }
        catch (Exception error) when (error is OperationCanceledException or ObjectDisposedException)
        {
            // Interrupt/shutdown is an explicit terminal outcome, not success.
            Print(new JsonObject { ["error"] = error is OperationCanceledException
                ? "Turn interrupted." : "Agent shut down before completion." });
        }
        catch (Exception error) when (error is ArgumentException or InvalidOperationException
            or IOException or JsonException or NotSupportedException)
        {
            failed = true;
            Console.Error.WriteLine(error.Message);
            Print(new JsonObject { ["error"] = error.Message });
        }
    }
    if (failed)
    {
        throw new InvalidOperationException("One or more requests failed.");
    }
}

static async Task WriteProgress(Agent agent, CancellationToken cancellation)
{
    try
    {
        while (true)
        {
            var value = await agent.WaitAsync(_ => true, cancellation);
            string kind = JsonNamingPolicy.SnakeCaseLower.ConvertName(value.Kind.ToString());
            if (value.Kind == ObservationKind.PartFinal && value.PartID is null)
            {
                kind = value.Part is ToolResultPart ? "tool_result" : "input_part";
            }
            var record = new JsonObject
            {
                ["observation"] = kind,
                ["seq"] = value.Seq,
                ["text"] = value.Text,
                ["accepted"] = value.Accepted,
                ["from"] = value.From is Phase from ? JsonNamingPolicy.SnakeCaseLower.ConvertName(from.ToString()) : null,
                ["to"] = value.To is Phase to ? JsonNamingPolicy.SnakeCaseLower.ConvertName(to.ToString()) : null,
                ["part"] = value.Part is null ? null : JsonCodec.WritePart(value.Part)
            };
            if (value.PartID is int id)
            {
                record["part_id"] = id;
                record["response_id"] = value.ResponseID;
            }
            if (value.Delta is Chunk delta)
            {
                record["kind"] = delta.Kind switch
                {
                    ChunkKind.Reply => "text",
                    ChunkKind.Thinking => "thinking",
                    _ => "tool_call"
                };
                record["chunk"] = delta.Kind == ChunkKind.ToolCall
                    ? delta.ToolName + delta.ArgsText : delta.Text;
            }
            if (value.Succeeded is bool succeeded)
            {
                record["succeeded"] = succeeded;
                if (succeeded && !value.HasReply)
                {
                    record["assistant"] = value.Text;
                }
            }
            Print(record);
        }
    }
    catch (OperationCanceledException) when (cancellation.IsCancellationRequested) { }
    catch (ObjectDisposedException)
    {
        // Destroy closes the stream after shutdown; buffered records were read.
    }
}

static void RequireKey(Configuration config)
{
    if (config.ApiKey.Length == 0)
    {
        throw new InvalidOperationException("No API key configured; render and dump remain available offline");
    }

    if (config.Model == "offline")
    {
        throw new InvalidOperationException("Set LLM_MODEL before sending a live request");
    }
}
