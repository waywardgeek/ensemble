using System.Net;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    private static string PressureReply(int input, string? tool = null, object? args = null, int cached = 0)
    {
        var body = JsonNode.Parse(AccountReply("gpt-4o-mini", "default", input, cached, 1))!;
        if (tool is not null)
        {
            body["choices"]![0]!["message"]!["tool_calls"] = JsonSerializer.SerializeToNode(new[]
            {
                new { id = Guid.NewGuid().ToString("N"), type = "function", function = new
                { name = tool, arguments = JsonSerializer.Serialize(args ?? new { }) } }
            });
            body["choices"]![0]!["finish_reason"] = "tool_calls";
        }
        return body.ToJsonString();
    }

    private static async Task<string> PressureWire(HttpListenerContext request)
    {
        using var reader = new StreamReader(request.Request.InputStream);
        return await reader.ReadToEndAsync();
    }

    private static string[] PressureTools(Agent agent)
    {
        var wire = agent.Engine().Render(agent.History().View());
        return wire["tools"]?.AsArray().Select(tool => tool!["function"]!["name"]!.GetValue<string>()).ToArray() ?? [];
    }

    private static async Task PressureTurn(Agent agent, HttpListener listener, string reply)
    {
        var turn = agent.SubmitAsync([new TextPart("Continue")]);
        await Reply(await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10)), reply);
        await turn.WaitAsync(TimeSpan.FromSeconds(10));
    }

    public static Task MemoryPressureThresholds() => WorkingContextPressure();

    public static async Task MemoryPressureUnknown(string directory)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "unknown-window",
                listener.Prefixes.Single(), disableStreaming: true));
            await agent.SetWorkingContextLimitAsync(new WorkingContextLimit(Vendor.OpenAI, "unknown-window", 250_000));
            await PressureTurn(agent, listener, PressureReply(200_000));
            Check(agent.WorkingContextStatus().Percent == 80m && agent.WorkingContextStatus().CapacityTokens is null,
                "Unknown capacity disabled the configured working limit");
            var direct = agent.Engine().SendAsync("{}");
            await Reply(await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10)), PressureReply(int.MaxValue));
            await direct;
            Check(agent.WorkingContextStatus().Percent == 80m, "Direct Engine send replaced actor measurement");
            var omitted = JsonNode.Parse(PressureReply(1))!;
            omitted.AsObject().Remove("usage");
            await PressureTurn(agent, listener, omitted.ToJsonString());
            Check(agent.WorkingContextStatus().InputTokens is null && agent.WorkingContextStatus().Percent is null,
                "Missing usage became zero or retained a stale percentage");
            await PressureTurn(agent, listener, PressureReply(200_000));
            await agent.SetModelAsync("gpt-4o-mini");
            Check(agent.WorkingContextStatus().InputTokens is null, "Model switch retained another model's usage");
            await agent.SetModelAsync("unknown-window");
            Check(agent.WorkingContextStatus().WorkingTokens == 250_000, "Model switch lost configured limit");

            var parent = await RuntimeAgent(app, Path.Combine(directory, "pressure-aux"), listener.Prefixes.Single());
            await parent.SetModelAsync("gpt-4o-mini");
            await RuntimeCheckpoint(parent);
            var worker = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            await Reply(worker, AccountReply("gpt-4o-mini", "default", 127_999, 100_000, 1, "COMPACT"));
            await RuntimeUntil(parent, () => parent.History().View().CompressionOutcome is not null);
            Check(parent.WorkingContextStatus().InputTokens is null, "Compressor usage leaked into main pressure");
        }
        finally { app.Destroy(); }
    }

    public static async Task MemoryPressureDeferred()
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "claude-3-haiku-20240307",
                listener.Prefixes.Single(), disableStreaming: true, deferredTools: true));
            _ = new Tool(agent, ToolKind.AgentStatus);
            var result = new ToolResultPart("lookup", [new TextPart("Discovery data")], toolReferences: ["agent_status"]);
            var history = new Context(dialogue:
                [new(1, Actor.Agent, [new ToolCallPart("lookup", agent.Configuration().Target(), "tool_search", JsonSerializer.SerializeToElement(new { }))]), new(2, Actor.Tool, [result])]);
            await PressureTurn(agent, listener, """
                {"model":"claude-3-haiku-20240307","content":[{"type":"text","text":"done"}],"stop_reason":"end_turn",
                 "usage":{"input_tokens":10000,"cache_creation_input_tokens":90000,"cache_read_input_tokens":90000,"output_tokens":1}}
                """);
            Check(agent.WorkingContextStatus().Percent == 95m, "Anthropic disjoint cache buckets counted incorrectly");
            string wire = agent.Engine().Render(history).ToJsonString();
            Check(wire.Contains("tool_reference") && wire.Contains("agent_status"), "Pressure changed existing tool authority");
            Check(result.ToolReferences.SequenceEqual(["agent_status"]), "Pressure rewrote canonical discovery");
            var allowed = await agent.DispatchAsync(new ToolCallPart("status", agent.Configuration().Target(), "agent_status", JsonSerializer.SerializeToElement(new { })));
            Check(!allowed.IsError, "Pressure incorrectly blocked an authorized control");
        }
        finally { app.Destroy(); }
    }
}
