using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    public static async Task OpenRouterChatRoundTrip()
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
            var endpoint = new ModelEndpoint(vendor, "api", listener.Prefixes.Single(), Surface.ChatCompletions, key);
            var model = new Model(endpoint, "flint-v1_0-free", "recursive/flint-v1.0:free", "Flint",
                new ModelFeatures(Media.None, Stream: true, ContextWindow: 262144),
                requestPolicy: new ModelRequestPolicy(maxOutputTokens: 32768));
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
                    Check(body.RootElement.GetProperty("max_completion_tokens").GetInt32() == 32768,
                        "Flint must allow its recommended 32768 tokens for thinking and completion.");
                    Check(body.RootElement.GetProperty("model").GetString() == "recursive/flint-v1.0:free"
                        && body.RootElement.GetProperty("stream").GetBoolean()
                        && incoming.Request.Headers["Authorization"] == "Bearer synthetic-router-key"
                        && incoming.Request.Headers["x-api-key"] is null,
                        "Exact free route, streaming and independent Bearer credential must reach the wire.");
                    if (index == 1)
                    {
                        var messages = body.RootElement.GetProperty("messages").EnumerateArray().ToArray();
                        var assistant = messages.Single(message => message.TryGetProperty("tool_calls", out _));
                        Check(assistant.GetProperty("content").ValueKind == JsonValueKind.String
                            && assistant.GetProperty("reasoning").GetString() == "Inspect the fixture."
                            && assistant.GetProperty("reasoning_content").GetString() == "Inspect the fixture.",
                            "Reasoning aliases belong on the assistant, not in its content array.");
                        var details = assistant.GetProperty("reasoning_details");
                        using var expected = JsonDocument.Parse("""
                            [{"index":0,"id":"r0","type":"reasoning.text","format":"unknown-format","text":"Inspect the fixture.","extension":{"kept":true}},
                             {"index":1,"id":"r1","type":"reasoning.encrypted","format":"opaque-format","data":"secret-a-secret-b"}]
                            """);
                        Check(JsonElement.DeepEquals(details, expected.RootElement),
                            "Structured details, opaque data, IDs and extensions must survive unchanged.");
                        var result = messages.Single(message => message.GetProperty("role").GetString() == "tool");
                        Check(result.GetProperty("tool_call_id").GetString() == "fixture-call"
                            && result.GetProperty("content").GetString() == "value-42" && effects == 1,
                            "Exactly one real tool result must reach the follow-up request.");
                    }
                    string[] frames = index == 0 ?
                    [
                        """{"model":"recursive/flint-v1.0:free","choices":[{"index":0,"delta":{"reasoning":"Inspect ","reasoning_content":"Inspect ","reasoning_details":[{"index":0,"id":"r0","type":"reasoning.text","format":"unknown-format","text":"Inspect ","extension":{"kept":true}}]}}]}""",
                        """{"choices":[{"index":0,"delta":{"reasoning":"the fixture.","reasoning_content":"the fixture.","reasoning_details":[{"index":0,"text":"the fixture."},{"index":1,"id":"r1","type":"reasoning.encrypted","format":"opaque-format","data":"secret-a-"}]}}]}""",
                        """{"choices":[{"index":0,"delta":{"reasoning_details":[{"index":1,"data":"secret-b"}],"tool_calls":[{"index":0,"id":"fixture-call","type":"function","function":{"name":"read_value","arguments":"{}"}}]},"finish_reason":"tool_calls"}]}"""
                    ] :
                    [
                        """{"model":"recursive/flint-v1.0:free","choices":[{"index":0,"delta":{"content":"The value is 42."},"finish_reason":"stop"}]}"""
                    ];
                    string stream = string.Concat(frames.Select(frame => "data: " + frame + "\n\n")) + "data: [DONE]\n\n";
                    await Reply(incoming, stream, type: "text/event-stream");
                }
                finally
                {
                    incoming.Response.Abort();
                }
            }
            Check(await turn.WaitAsync(deadline.Token) == "The value is 42." && effects == 1,
                "Chat tool round trip completes once with the final streamed answer.");
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

    public static async Task ChatReasoningCodec()
    {
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "model-a", "http://127.0.0.1:1"));
            foreach (bool aliases in new[] { false, true })
            {
                string alias1 = aliases ? "\"reasoning\":\"Plan \",\"reasoning_content\":\"Plan \"," : "";
                string alias2 = aliases ? "\"reasoning\":\"once.\",\"reasoning_content\":\"once.\"," : "";
                string stream = "data: {\"model\":\"model-a\",\"choices\":[{\"index\":0,\"delta\":{" + alias1 +
                    "\"reasoning_details\":[{\"index\":0,\"id\":\"r0\",\"type\":\"reasoning.text\",\"text\":\"Plan \"}]}}]}\n\n" +
                    "data: {\"choices\":[{\"index\":0,\"delta\":{" + alias2 +
                    "\"reasoning_details\":[{\"index\":0,\"text\":\"once.\"},{\"index\":1,\"type\":\"reasoning.encrypted\",\"data\":\"NEVER-DISPLAY\"}]},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n";
                using var response = new HttpResponseMessage(System.Net.HttpStatusCode.OK)
                {
                    Content = new StringContent(stream, System.Text.Encoding.UTF8, "text/event-stream")
                };
                var chunks = new List<Chunk>();
                var parsed = await agent.Engine().ParseAsync(response, chunks.Add);
                var ended = parsed.Events.Single(fact => fact.Kind == EventKind.ResponseEnded);
                var opaque = (OpaquePart)ended.Parts.Single();
                Check(parsed.PartIDs.Length == 1 && chunks.All(chunk => chunk.PartID == parsed.PartIDs[0])
                    && string.Concat(chunks.Select(chunk => chunk.Text)) == "Plan once."
                    && opaque.Data.GetProperty("text").GetString() == "Plan once.",
                    "Aliases and details display once, with stable identity and no encrypted payload.");
                using var saved = JsonDocument.Parse(JsonCodec.WritePart(opaque).ToJsonString());
                Part restored = JsonCodec.ReadPart(saved.RootElement);
                var context = new Context([new Entry(1, Actor.Agent, [restored])]);
                var message = agent.Engine().Render(context)["messages"]!.AsArray().Last()!.AsObject();
                Check(message["reasoning_details"]!.AsArray()[1]!["data"]!.GetValue<string>() == "NEVER-DISPLAY"
                    && message["content"]!.GetValue<string>() == "",
                    "Save/reload retains encrypted details for native replay, outside content.");
                var plain = new System.Text.Json.Nodes.JsonObject
                {
                    ["model"] = "model-a",
                    ["choices"] = new System.Text.Json.Nodes.JsonArray(new System.Text.Json.Nodes.JsonObject
                    {
                        ["message"] = message.DeepClone(), ["finish_reason"] = "stop"
                    })
                };
                var reparsed = agent.Engine().Parse(200, plain.ToJsonString())
                    .Single(fact => fact.Kind == EventKind.ResponseEnded).Parts.OfType<OpaquePart>().Single();
                Check(JsonElement.DeepEquals(opaque.Data, reparsed.Data), "Stream and plain reasoning disagree.");
                var other = new Agent(app, new Configuration(Vendor.OpenAI, "model-b", "http://127.0.0.1:1"));
                Check(!other.Engine().Render(context).ToJsonString().Contains("NEVER-DISPLAY", StringComparison.Ordinal),
                    "Reasoning must not leak to an incompatible model.");

                using var broken = new HttpResponseMessage(System.Net.HttpStatusCode.OK)
                {
                    Content = new StringContent(stream.Replace("\"index\":0,\"text\":\"once.\"",
                        "\"index\":0,\"id\":\"different\",\"text\":\"once.\""),
                        System.Text.Encoding.UTF8, "text/event-stream")
                };
                Check((await agent.Engine().ParseAsync(broken)).Events.Single().Kind == EventKind.ErrorOccurred,
                    "Contradictory detail identities must refuse the response.");
            }
            using var legacyData = JsonDocument.Parse("""{"type":"reasoning_content","text":"Legacy plan."}""");
            var legacy = new OpaquePart(agent.Configuration().Target(), legacyData.RootElement);
            var legacyWire = agent.Engine().Render(new Context([new Entry(1, Actor.Agent, [legacy])]));
            Check(legacyWire["messages"]!.AsArray().Last()!["reasoning_content"]!.GetValue<string>() == "Legacy plan.",
                "Previously recorded Chat reasoning must replay without History migration.");
            foreach (string malformed in new[] { "42", "{}", "[null]", "[{\"type\":\"reasoning.text\",\"text\":42}]" })
            {
                var failed = agent.Engine().Parse(200,
                    "{\"choices\":[{\"message\":{\"reasoning_details\":" + malformed + "}}]}");
                Check(failed.Single().Kind == EventKind.ErrorOccurred, "Malformed reasoning cannot become a completed response.");
            }
            foreach (string field in new[] { "reasoning", "reasoning_content" })
            {
                string reply = "{\"model\":\"model-a\",\"choices\":[{\"message\":{\"content\":null,\"" + field + "\":\"Only this.\"}}]}";
                var parts = agent.Engine().Parse(200, reply).Single(fact => fact.Kind == EventKind.ResponseEnded).Parts;
                var rendered = agent.Engine().Render(new Context([new Entry(1, Actor.Agent, parts)]));
                Check(rendered["messages"]!.AsArray().Last()![field]!.GetValue<string>() == "Only this.",
                    "Scalar reasoning aliases must roundtrip as message-level fields.");
            }
        }
        finally { app.Destroy(); }
    }
}
