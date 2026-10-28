using System.Globalization;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Writing;
using Application = Ensemble.Ensemble;

// This is application composition, not another agent loop. The three Agents
// receive their prompts and toolsets from outside the library assembly.
string address = Environment.GetEnvironmentVariable("LLM_BASE_URL") ?? "";
if (!Uri.TryCreate(address, UriKind.Absolute, out var endpoint) || !endpoint.IsLoopback)
{
    Console.Error.WriteLine("This offline exercise requires an owned loopback LLM_BASE_URL.");
    return 2;
}
if (!Enum.TryParse<Vendor>(Environment.GetEnvironmentVariable("LLM_VENDOR") ?? "Anthropic", true, out var vendor))
{
    Console.Error.WriteLine("Unsupported LLM_VENDOR");
    return 2;
}
string model = Environment.GetEnvironmentVariable("LLM_MODEL") ?? "local";
string key = Environment.GetEnvironmentVariable("LLM_API_KEY") ?? "dummy";
string logPath = Environment.GetEnvironmentVariable("CH06_LOG") ?? "workflow.jsonl";
using var log = new StreamWriter(logPath) { AutoFlush = true };
var outputGate = new object();
var app = new Application(Console.Error);
var draft = new Draft();
Task? observations = null;
Task? workflow = null;
try
{
    var author = Create("author", "You are the author. Write the requested draft with write_draft; use word_count to check it.");
    var editor = Create("editor", "You are the editor. Read the draft with read_draft and improve it with edit_draft.");
    var reviewer = Create("reviewer", "You are the reviewer. Inspect the supplied draft and call review with accept or reject and notes.");
    var textSchema = Schema("""
        {"type":"object","properties":{"text":{"type":"string"}},"required":["text"],"additionalProperties":false}
        """);
    var emptySchema = Schema("""
        {"type":"object","properties":{},"additionalProperties":false}
        """);
    _ = new Tool(author, "write_draft", "Replace the manuscript text.", textSchema,
        (job, args, cancellation) =>
        {
            cancellation.ThrowIfCancellationRequested();
            job.Parent().Parent().Parent().Log("author writes manuscript");
            return Task.FromResult(draft.Replace(args.GetProperty("text").GetString()!));
        });
    _ = new Tool(author, "word_count", "Count whitespace-separated manuscript words.", emptySchema,
        (job, args, cancellation) =>
        {
            cancellation.ThrowIfCancellationRequested();
            return Task.FromResult(draft.WordCount().ToString(CultureInfo.InvariantCulture));
        });
    _ = new Tool(author, "think", "Wait the numeric seconds, then return the thought.", Schema("""
        {"type":"object","properties":{"seconds":{"type":"number","minimum":0},"thought":{"type":"string"}},"required":["seconds","thought"],"additionalProperties":false}
        """), async (job, args, cancellation) =>
        {
            double seconds = args.GetProperty("seconds").GetDouble();
            await Task.Delay(TimeSpan.FromSeconds(seconds), cancellation);
            return args.GetProperty("thought").GetString()!;
        });
    // A hint can end the foreground observation of think, not the Job. The
    // author may subsequently observe that same Job through the ordinary tool.
    _ = new Tool(author, ToolKind.WaitForJob);
    _ = new Tool(editor, "read_draft", "Read the current manuscript.", emptySchema,
        (job, args, cancellation) =>
        {
            cancellation.ThrowIfCancellationRequested();
            return Task.FromResult(draft.Read());
        });
    _ = new Tool(editor, "edit_draft", "Replace the manuscript with edited text.", textSchema,
        (job, args, cancellation) =>
        {
            cancellation.ThrowIfCancellationRequested();
            job.Parent().Parent().Parent().Log("editor revises manuscript");
            return Task.FromResult(draft.Replace(args.GetProperty("text").GetString()!));
        });
    _ = new Tool(reviewer, "review", "Record an accept/reject decision with notes.", Schema("""
        {"type":"object","properties":{"decision":{"type":"string","enum":["accept","reject"]},"notes":{"type":"string"}},"required":["decision","notes"],"additionalProperties":false}
        """), (job, args, cancellation) =>
        {
            cancellation.ThrowIfCancellationRequested();
            // The immutable result becomes the review record in this Agent's
            // History. No second mutable review ledger is needed.
            return Task.FromResult(args.GetRawText());
        });

    observations = Observe();
    while (await Console.In.ReadLineAsync() is string line)
    {
        if (string.IsNullOrWhiteSpace(line))
        {
            continue;
        }
        using var input = JsonDocument.Parse(line);
        var value = input.RootElement;
        string kind = value.GetProperty("kind").GetString() ?? "";
        string text = value.GetProperty("text").GetString()
            ?? throw new FormatException("text must be a string");
        if (kind == "prompt")
        {
            if (workflow is not null)
            {
                throw new FormatException("This example accepts one workflow per run.");
            }
            workflow = RunWorkflow(text, author, editor, reviewer);
        }
        else if (kind == "hint")
        {
            string name = value.GetProperty("agent").GetString() ?? "";
            var target = app.FindAgent(name) ?? throw new FormatException("Unknown hint target: " + name);
            await target.PostAsync(new AgentInput(InputKind.Hint, text));
            Emit(new JsonObject { ["observation"] = "hint_accepted", ["agent"] = target.Name });
        }
        else
        {
            throw new FormatException("Expected prompt or addressed hint.");
        }
    }
    if (workflow is not null)
    {
        // EOF joins the reliable pipeline. A slow/lossy observer cannot decide
        // that an author, editor, or reviewer has completed.
        await workflow;
    }
    return 0;
}
catch (Exception error) when (error is ArgumentException or InvalidOperationException or IOException
    or JsonException or FormatException or KeyNotFoundException or NotSupportedException)
{
    Console.Error.WriteLine(error.Message);
    return 1;
}
finally
{
    try
    {
        foreach (var agent in app.Agents())
        {
            await agent.ShutdownAsync();
        }
        // A read/parse failure may leave the pipeline awaiting an Agent reply.
        // Collect that outcome too; it must not become detached background work.
        if (workflow is not null)
        {
            try
            {
                await workflow;
            }
            catch (Exception) when (workflow.IsFaulted || workflow.IsCanceled)
            {
                // Its original outcome was already reported above, or shutdown
                // ended it after an input error. This join must not replace that error.
            }
        }
    }
    finally
    {
        // Destruction completes the root queue only after its Agents stop.
        // Drain already-published finals instead of cancelling their reader.
        try
        {
            app.Destroy();
        }
        finally
        {
            if (observations is not null)
            {
                await observations;
            }
        }
    }
}

Agent Create(string name, string primary)
{
    var config = new Configuration(vendor, model, endpoint.ToString(), key, primaryInstructions: primary,
        disableStreaming: Environment.GetEnvironmentVariable("EN_DISABLE_STREAMING") == "1");
    var agent = new Agent(app, config, name: name);
    agent.History().StartLog(logPath + "." + name + ".jsonl");
    return agent;
}

async Task RunWorkflow(string topic, Agent author, Agent editor, Agent reviewer)
{
    await Complete(author, topic);
    await Complete(editor, "Edit this manuscript:\n" + draft.Read());
    await Complete(reviewer, "Review this manuscript:\n" + draft.Read());
}

async Task Complete(Agent agent, string prompt)
{
    await agent.SubmitAsync([new TextPart(prompt)]);
    Emit(new JsonObject { ["observation"] = "turn_ended", ["agent"] = agent.Name });
    Emit(new JsonObject { ["pipeline"] = agent.Name + "_done" });
}

async Task Observe()
{
    try
    {
        while (true)
        {
            var tagged = await app.WaitAnyAsync(_ => true);
            var value = tagged.Observation;
            if (value.Kind == ObservationKind.StateChanged)
            {
                Emit(new JsonObject
                {
                    ["observation"] = "state_changed", ["agent"] = tagged.Agent.Name,
                    ["from"] = State(value.From), ["to"] = State(value.To)
                });
            }
            else if (value.Delta is Chunk chunk)
            {
                string kind = chunk.Kind switch
                {
                    ChunkKind.Reply => "text",
                    ChunkKind.Thinking => "thinking",
                    _ => "tool_call"
                };
                string text = chunk.Kind == ChunkKind.ToolCall
                    ? chunk.ToolName + chunk.ArgsText : chunk.Text;
                Emit(new JsonObject
                {
                    ["observation"] = "part_delta", ["agent"] = tagged.Agent.Name,
                    ["response_id"] = value.ResponseID, ["part_id"] = chunk.PartID,
                    ["kind"] = kind, ["chunk"] = text
                });
            }
            else if (value.Kind == ObservationKind.PartFinal && value.PartID is int partId)
            {
                // Only model finals carry parser IDs. Human input and Job
                // results must not masquerade as parts of a streamed response.
                Emit(new JsonObject
                {
                    ["observation"] = "part_final", ["agent"] = tagged.Agent.Name,
                    ["seq"] = value.Seq, ["response_id"] = value.ResponseID,
                    ["part_id"] = partId, ["part"] = JsonCodec.WritePart(value.Part!)
                });
            }
        }
    }
    catch (ObjectDisposedException)
    {
        // Normal queue completion after explicit root destruction.
    }
}

void Emit(JsonObject record)
{
    string line = record.ToJsonString();
    lock (outputGate)
    {
        log.WriteLine(line);
        Console.WriteLine(line);
        Console.Out.Flush();
    }
}

static JsonElement Schema(string json)
{
    using var document = JsonDocument.Parse(json);
    return document.RootElement.Clone();
}

static string? State(Phase? phase) => phase switch
{
    Phase.Idle => "idle",
    Phase.InputPending => "input_pending",
    Phase.InFlight => "in_flight",
    Phase.ToolsPending => "tools_pending",
    _ => null
};
