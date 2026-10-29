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

    public static async Task MemoryPressureThresholds()
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4o-mini",
                listener.Prefixes.Single(), disableStreaming: true));
            _ = new Tool(agent, ToolKind.MicroHandoff);
            int effects = 0;
            _ = new Tool(agent, "effect", "Count actual dispatch", JsonSerializer.SerializeToElement(new
                { type = "object", properties = new { }, additionalProperties = false }),
                (job, args, cancellation) => { Interlocked.Increment(ref effects); return Task.FromResult("effect"); },
                waitForCompletion: true);

            await PressureTurn(agent, listener, PressureReply(115_200, cached: 100_000)); // exactly 90%, not 168%
            Check(PressureTools(agent).Contains("effect"), "90% incorrectly restricted tools or counted cached input twice");
            var turn = agent.SubmitAsync([new TextPart("Continue past warning")]);
            var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            string warning = await PressureWire(request);
            Check(warning.Contains("90.0%") && warning.Contains("Call micro_handoff soon"), "Missing request-only 90% warning");
            await Reply(request, PressureReply(121_599)); // one token below 95%
            await turn.WaitAsync(TimeSpan.FromSeconds(10));
            Check(PressureTools(agent).Contains("effect"), "Rounded 94.999% up to forced threshold");

            turn = agent.SubmitAsync([new TextPart("Force at exact threshold")]);
            request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            // All three normalized input categories count once. The raw Chat
            // total already includes both cache categories.
            var threeBuckets = JsonNode.Parse(PressureReply(121_600, "effect", cached: 60_000))!;
            threeBuckets["usage"]!["prompt_tokens_details"]!["cache_write_tokens"] = 60_000;
            await Reply(request, threeBuckets.ToJsonString());
            request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            string forced = await PressureWire(request);
            Check(PressureTools(agent).SequenceEqual(["micro_handoff"]) && effects == 0,
                "95% failed to restrict declarations or admitted side effects");
            Check(forced.Contains("95.0%") && forced.Contains("Tool is not enabled: effect"), "Missing correlated dispatch refusal");
            // A subsequent short response must not silently release the latch
            // or leave its still-enforced restriction unexplained.
            await Reply(request, PressureReply(1, "effect"));
            request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            string stillForced = await PressureWire(request);
            Check(effects == 0 && PressureTools(agent).SequenceEqual(["micro_handoff"])
                && stillForced.Contains("Only micro_handoff is currently enabled"), "Outstanding checkpoint lost its instruction");
            await Reply(request, PressureReply(121_600, "micro_handoff", new { text = "Resume with the measured state." }));
            request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            string after = await PressureWire(request);
            Check(PressureTools(agent).Contains("effect") && !after.Contains("Context window warning:"),
                "Successful handoff did not release restriction or remove transient warning");
            await Reply(request, PressureReply(1, "effect"));
            request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            Check(effects == 1, "Tool did not become executable after handoff");
            await Reply(request, PressureReply(1));
            await turn.WaitAsync(TimeSpan.FromSeconds(10));
            Check(!agent.History().Events().Any(fact => JsonCodec.WriteEvent(fact).Contains("Context window warning:")),
                "Request-only warning became canonical context");
        }
        finally { app.Destroy(); }
    }

    public static async Task MemoryPressureUnknown(string directory)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "unknown-window",
                listener.Prefixes.Single(), disableStreaming: true));
            _ = new Tool(agent, ToolKind.MicroHandoff);
            _ = new Tool(agent, ToolKind.AgentStatus);
            await PressureTurn(agent, listener, PressureReply(int.MaxValue));
            Check(PressureTools(agent).Contains("agent_status"), "Unknown window acquired guessed forcing");
            await agent.SetModelAsync("gpt-4o-mini");
            var direct = agent.Engine().SendAsync("{}");
            await Reply(await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10)), PressureReply(int.MaxValue));
            await direct;
            Check(PressureTools(agent).Contains("agent_status"), "Direct library send forced actor memory");
            var omitted = JsonNode.Parse(PressureReply(1))!;
            omitted.AsObject().Remove("usage");
            await PressureTurn(agent, listener, omitted.ToJsonString());
            Check(PressureTools(agent).Contains("agent_status"), "Missing usage invented occupancy");
            await PressureTurn(agent, listener, PressureReply(121_600));
            Check(!PressureTools(agent).Contains("agent_status"), "Positive forcing control missing");
            await agent.SetModelAsync("unknown-window");
            Check(PressureTools(agent).Contains("agent_status"), "Model switch retained another model's force latch");

            var parent = await RuntimeAgent(app, Path.Combine(directory, "pressure-aux"), listener.Prefixes.Single());
            await parent.SetModelAsync("gpt-4o-mini");
            _ = new Tool(parent, ToolKind.AgentStatus);
            await RuntimeCheckpoint(parent);
            var worker = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            await Reply(worker, AccountReply("gpt-4o-mini", "default", 127_999, 100_000, 1, "COMPACT"));
            await RuntimeUntil(parent, () => parent.History().View().CompressionOutcome is not null);
            Check(PressureTools(parent).Contains("agent_status") && parent.Engine().Meter().Latest is null,
                "Compressor occupancy leaked into main forcing");
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
            _ = new Tool(agent, ToolKind.MicroHandoff);
            _ = new Tool(agent, ToolKind.AgentStatus);
            var result = new ToolResultPart("lookup", [new TextPart("Discovery data")], toolReferences: ["agent_status"]);
            var history = new Context(dialogue:
                [new(1, Actor.Agent, [new ToolCallPart("lookup", agent.Configuration().Target(), "tool_search", JsonSerializer.SerializeToElement(new { }))]), new(2, Actor.Tool, [result])]);
            Check(agent.Engine().Render(history).ToJsonString().Contains("tool_reference"), "Missing deferred-reference positive control");
            await PressureTurn(agent, listener, """
                {"model":"claude-3-haiku-20240307","content":[{"type":"text","text":"done"}],"stop_reason":"end_turn",
                 "usage":{"input_tokens":10000,"cache_creation_input_tokens":90000,"cache_read_input_tokens":90000,"output_tokens":1}}
                """);
            string forced = agent.Engine().Render(history).ToJsonString();
            Check(!forced.Contains("tool_reference") && !forced.Contains("agent_status") && forced.Contains("micro_handoff"),
                "Deferred references or declarations bypassed 95% restriction");
            Check(result.ToolReferences.SequenceEqual(["agent_status"]), "Restriction rewrote canonical discovery");
            var denied = await agent.DispatchAsync(new ToolCallPart("denied", agent.Configuration().Target(), "agent_status", JsonSerializer.SerializeToElement(new { })));
            Check(denied.IsError, "Direct/control admission bypassed restriction");
            var note = await agent.DispatchAsync(new ToolCallPart("checkpoint", agent.Configuration().Target(), "micro_handoff", JsonSerializer.SerializeToElement(new { text = "Keep measured progress" })));
            await agent.FlushAsync().WaitAsync(TimeSpan.FromSeconds(10));
            Check(!note.IsError && agent.Engine().Render(history).ToJsonString().Contains("tool_reference"),
                "Direct successful checkpoint failed to restore deferred projection");
        }
        finally { app.Destroy(); }
    }

    private static async Task PressureWaiting(Agent agent)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(10));
        while ((await agent.WorkStatusAsync(deadline.Token)).Work?.Detail != "Waiting for memory headroom")
        {
            await Task.Delay(5, deadline.Token);
        }
    }

    public static async Task MemoryPressureHeadroom(string directory)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        var agent = await RuntimeAgent(app, Path.Combine(directory, "pressure-headroom"), listener.Prefixes.Single());
        try
        {
            await agent.SetModelAsync("gpt-4o-mini");
            await RuntimeCheckpoint(agent);
            var worker = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            var turn = agent.SubmitAsync([new TextPart("MAIN_REQUEST")]);
            var main = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            await Reply(main, PressureReply(121_600, "micro_handoff", new { text = "Keep source until commit." }));
            await PressureWaiting(agent);
            Check(agent.History().Events().Count(e => e.Kind == EventKind.RequestSent) == 1 && !turn.IsCompleted,
                "Wait rendered/sent another request");
            await agent.PostAsync(new AgentInput(InputKind.Pause)).WaitAsync(TimeSpan.FromSeconds(10));
            Check((await agent.WorkStatusAsync()).Paused, "Pause blocked behind headroom");
            await agent.SetMemorySettingsAsync(RuntimePolicy(soulOff: false)).WaitAsync(TimeSpan.FromSeconds(10));
            Check(agent.Children().Length == 1, "Unrelated policy canceled headroom worker");
            await agent.PostAsync(new AgentInput(InputKind.Resume)).WaitAsync(TimeSpan.FromSeconds(10));
            await Reply(worker, AccountReply("gpt-4o-mini", "default", 10, 0, 1, "REPLACEMENT"));
            main = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            string wire = await PressureWire(main);
            Check(wire.Contains("REPLACEMENT") && !wire.Contains("SOURCE_ONE") && !wire.Contains("SOURCE_TWO"),
                "Resumed render did not use committed memory");
            await Reply(main, PressureReply(2));
            await turn.WaitAsync(TimeSpan.FromSeconds(10));
            Check(agent.Children().Length == 0, "Completed headroom worker survived teardown");
        }
        finally { app.Destroy(); }
    }

    private static async Task PressureCanceled(Task<string> turn)
    {
        bool canceled = false;
        try { await turn.WaitAsync(TimeSpan.FromSeconds(10)); }
        catch (OperationCanceledException) { canceled = true; }
        catch (ObjectDisposedException) { canceled = true; }
        Check(canceled, "Stopped foreground turn incorrectly reported success");
    }

    public static async Task MemoryPressureHeadroomExit(string directory)
    {
        foreach (string mode in new[] { "failed", "invalidated", "interrupted", "shutdown" })
        {
            using var listener = Listen();
            var app = new Application(TextWriter.Null);
            var agent = await RuntimeAgent(app, Path.Combine(directory, "pressure-" + mode), listener.Prefixes.Single());
            try
            {
                await agent.SetModelAsync("gpt-4o-mini");
                await RuntimeCheckpoint(agent);
                var worker = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
                var turn = agent.SubmitAsync([new TextPart("Wait for headroom")]);
                var main = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
                await Reply(main, PressureReply(121_600, "micro_handoff", new { text = "Preserve current state." }));
                await PressureWaiting(agent);
                if (mode == "failed")
                {
                    await Reply(worker, PressureReply(1)); // ordinary prose is not a compressor submission
                }
                else if (mode == "invalidated")
                {
                    await agent.SetMemorySettingsAsync(RuntimePolicy(sessionOff: true)).WaitAsync(TimeSpan.FromSeconds(10));
                }
                else if (mode == "interrupted")
                {
                    await agent.PostAsync(new AgentInput(InputKind.InterruptNow)).WaitAsync(TimeSpan.FromSeconds(10));
                    await PressureCanceled(turn);
                    Check(agent.Children().Length == 1, "Interrupt canceled independent compressor");
                    await Reply(worker, AccountReply("gpt-4o-mini", "default", 10, 0, 1, "AFTER_INTERRUPT"));
                    await RuntimeUntil(agent, () => agent.History().View().CompressionOutcome is not null);
                }
                else
                {
                    await agent.ShutdownAsync().WaitAsync(TimeSpan.FromSeconds(10));
                }
                if (mode is "failed" or "invalidated")
                {
                    bool refused = false;
                    try { await turn.WaitAsync(TimeSpan.FromSeconds(10)); }
                    catch (InvalidOperationException error) when (error.Message.Contains("headroom")) { refused = true; }
                    Check(refused, "Failed wait did not surface explicit headroom error");
                }
                else { await PressureCanceled(turn); }
                Check(agent.History().Events().Count(e => e.Kind == EventKind.RequestSent) == 1,
                    "Exit sent a main request despite unresolved headroom");
                Check(agent.Children().Length == 0, "Exit leaked worker");
            }
            finally { app.Destroy(); }
        }
    }
}
