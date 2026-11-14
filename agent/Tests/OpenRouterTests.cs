using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    // The fake speaks real SSE and examines the next wire request. This checks
    // continuation through the Agent, not a parser-only reconstruction.
    public static async Task OpenRouterMessagesRoundTrip()
    {
        using var listener = Listen();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var app = new Application(TextWriter.Null);
        Task<string>? turn = null;
        Agent? agent = null;
        int effects = 0;
        try
        {
            var vendor = new ModelVendor(app, "openrouter", "OpenRouter");
            var key = new ModelAuthentication(app, "router-key", "synthetic-router-key", ApiKeyHeader.Bearer);
            var endpoint = new ModelEndpoint(vendor, "api", listener.Prefixes.Single(), Surface.Messages, key);
            var model = new Model(endpoint, "flint-v1_0-free", "recursive/flint-v1.0:free", "Flint",
                new ModelFeatures(Media.None, Stream: true, ContextWindow: 262144));
            agent = new Agent(app, new Configuration(Vendor.OpenAI, "legacy", "http://127.0.0.1:1"));
            await agent.SetModelAsync(model);
            using var schema = JsonDocument.Parse("""
                {"type":"object","properties":{},"additionalProperties":false}
                """);
            _ = new Tool(agent, "read_value", "Read the fixture value", schema.RootElement,
                (job, arguments, cancellation) =>
                {
                    effects++;
                    return Task.FromResult("value-42");
                });
            turn = agent.SendAsync("Read the fixture value using its tool.");
            for (int index = 0; index < 2; index++)
            {
                var incoming = await listener.GetContextAsync().WaitAsync(deadline.Token);
                try
                {
                    using var reader = new StreamReader(incoming.Request.InputStream);
                    using var body = JsonDocument.Parse(await reader.ReadToEndAsync(deadline.Token));
                    Check(body.RootElement.GetProperty("model").GetString() == "recursive/flint-v1.0:free"
                        && body.RootElement.GetProperty("stream").GetBoolean()
                        && incoming.Request.Headers["Authorization"] == "Bearer synthetic-router-key"
                        && incoming.Request.Headers["x-api-key"] is null,
                        "Exact free route, streaming and independent Bearer credential must reach the wire.");
                    if (index == 1)
                    {
                        var blocks = body.RootElement.GetProperty("messages").EnumerateArray()
                            .SelectMany(message => message.GetProperty("content").EnumerateArray()).ToArray();
                        var thinking = blocks.Single(block => block.GetProperty("type").GetString() == "thinking");
                        var result = blocks.Single(block => block.GetProperty("type").GetString() == "tool_result");
                        Check(thinking.GetProperty("thinking").GetString() == "Inspect the fixture."
                            && thinking.GetProperty("signature").GetString() == "fixture-signature",
                            "Thinking and signature must survive streaming, History and rendering unchanged.");
                        Check(result.GetProperty("tool_use_id").GetString() == "fixture-call"
                            && result.GetRawText().Contains("value-42", StringComparison.Ordinal)
                            && effects == 1,
                            "Exactly one tool effect must precede its correlated result.");
                    }
                    string start = """
                        {"type":"message_start","message":{"id":"fixture-message","type":"message","role":"assistant","model":"recursive/flint-v1.0:free","content":[],"usage":{"input_tokens":20,"output_tokens":0}}}
                        """;
                    string[] blocksToSend = index == 0 ?
                    [
                        """{"type":"content_block_start","index":0,"content_block":{"type":"thinking","thinking":"","signature":""}}""",
                        """{"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"Inspect the fixture."}}""",
                        """{"type":"content_block_delta","index":0,"delta":{"type":"signature_delta","signature":"fixture-signature"}}""",
                        """{"type":"content_block_stop","index":0}""",
                        """{"type":"content_block_start","index":1,"content_block":{"type":"tool_use","id":"fixture-call","name":"read_value","input":{}}}""",
                        """{"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"{}"}}""",
                        """{"type":"content_block_stop","index":1}""",
                        """{"type":"message_delta","delta":{"stop_reason":"tool_use"},"usage":{"output_tokens":8}}"""
                    ] :
                    [
                        """{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}""",
                        """{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"The value is 42."}}""",
                        """{"type":"content_block_stop","index":0}""",
                        """{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":5}}"""
                    ];
                    string[] frames = [start, .. blocksToSend, """{"type":"message_stop"}"""];
                    string stream = string.Concat(frames.Select(frame => "data: " + frame + "\n\n"));
                    await Reply(incoming, stream, type: "text/event-stream");
                }
                finally
                {
                    incoming.Response.Abort();
                }
            }
            Check(await turn.WaitAsync(deadline.Token) == "The value is 42." && effects == 1,
                "Messages tool round trip completes once with the final streamed answer.");
        }
        finally
        {
            deadline.Cancel();
            listener.Close();
            try
            {
                if (turn is not null && !turn.IsCompleted)
                {
                    // Interrupt the Agent before joining an unfinished fixture turn.
                    await agent!.PostAsync(new AgentInput(InputKind.InterruptNow));
                    try
                    {
                        await turn;
                    }
                    catch (Exception error) when (error is OperationCanceledException or HttpRequestException)
                    {
                    }
                }
            }
            finally
            {
                app.Destroy();
            }
        }
    }

    // Exercise the real sender against a working HTTP boundary. The synthetic
    // gateway uses Messages, but must not inherit that codec's native key header.
    public static async Task MessagesAuthentication()
    {
        foreach (ApiKeyHeader header in new[] { ApiKeyHeader.ProtocolDefault, ApiKeyHeader.Bearer })
        {
            using var listener = Listen();
            using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(10));
            var app = new Application(TextWriter.Null);
            Task? send = null;
            try
            {
                var binding = new ModelAuthentication(app, "test-key", "synthetic-router-key", header);
                var vendor = new ModelVendor(app, "openrouter", "OpenRouter");
                var endpoint = new ModelEndpoint(vendor, "api", listener.Prefixes.Single(),
                    Surface.Messages, binding);
                var model = new Model(endpoint, "flint-v1_0-free", "recursive/flint-v1.0:free", "Flint",
                    new ModelFeatures(Media.None, Stream: true, ContextWindow: 262144));
                var agent = new Agent(app, new Configuration(Vendor.OpenAI, "legacy",
                    "http://127.0.0.1:1", "must-not-leak"));
                await agent.SetModelAsync(model);
                Check(agent.SelectedModel() == model && model.Vendor().ID == "openrouter"
                    && model.ServingContextTokens is null && model.Features.ContextWindow == 262144,
                    "Gateway identity and advertised versus serving context remain distinct.");
                send = agent.Engine().SendAsync("{}", deadline.Token);
                var incoming = await listener.GetContextAsync().WaitAsync(deadline.Token);
                try
                {
                    Check(incoming.Request.Url!.AbsolutePath == "/v1/messages",
                        "Messages endpoint suffix is preserved.");
                    Check(incoming.Request.Headers["anthropic-version"] == "2023-06-01",
                        "Messages protocol version is retained.");
                    Check(incoming.Request.Headers["Authorization"] ==
                        (header == ApiKeyHeader.Bearer ? "Bearer synthetic-router-key" : null),
                        "Only explicit Bearer authentication sets Authorization.");
                    Check(incoming.Request.Headers["x-api-key"] ==
                        (header == ApiKeyHeader.ProtocolDefault ? "synthetic-router-key" : null),
                        "Native Messages key header remains the default, never a second gateway header.");
                    await Reply(incoming, """
                        {"model":"recursive/flint-v1.0:free","content":[{"type":"text","text":"ok"}],
                         "usage":{"input_tokens":1,"output_tokens":1}}
                        """);
                    await send.WaitAsync(deadline.Token);
                }
                finally
                {
                    incoming.Response.Abort();
                }
            }
            finally
            {
                deadline.Cancel();
                listener.Close();
                try
                {
                    if (send is not null)
                    {
                        try
                        {
                            await send;
                        }
                        catch (OperationCanceledException)
                        {
                            // Fixture shutdown joins its own cancelled request.
                        }
                    }
                }
                finally
                {
                    app.Destroy();
                }
            }
        }
    }
}
