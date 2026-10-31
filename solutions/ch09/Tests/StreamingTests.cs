using System.Net;
using System.Text;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class StreamingTests
{
    private static void Check(bool condition, string reason)
    {
        if (!condition)
        {
            throw new InvalidOperationException(reason);
        }
    }

    private static string Frame(string json) => "data: " + json + "\r\n\r\n";

    private static HttpResponseMessage Response(string body, bool stream = true, HttpStatusCode status = HttpStatusCode.OK)
    {
        return new HttpResponseMessage(status)
        {
            Content = new StringContent(body, Encoding.UTF8, stream ? "text/event-stream" : "application/json")
        };
    }

    public static async Task Parser()
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var app = new Application(TextWriter.Null);
        try
        {
            var anthropic = new Agent(app, new Configuration(Vendor.Anthropic, "model-a", "http://127.0.0.1"));
            string start = """
                {"type":"message_start","message":{"model":"model-a","usage":{"input_tokens":8,"output_tokens":0}}}
                """;
            string stream = ": heartbeat\r\nevent: message_start\r\n" + Frame(start)
                + Frame("""{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}""")
                + "data: {\r\ndata: \"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"I will inspect.\"}}\r\n\r\n"
                + Frame("""{"type":"content_block_stop","index":0}""")
                + Frame("""{"type":"content_block_start","index":2,"content_block":{"type":"thinking","thinking":"","signature":""}}""")
                + Frame("""{"type":"content_block_delta","index":2,"delta":{"type":"thinking_delta","thinking":"plan"}}""")
                + Frame("""{"type":"content_block_delta","index":2,"delta":{"type":"signature_delta","signature":"sig"}}""")
                + Frame("""{"type":"content_block_stop","index":2}""")
                + Frame("""{"type":"content_block_start","index":7,"content_block":{"type":"tool_use","id":"call","name":"inspect","input":{}}}""")
                + Frame("""{"type":"content_block_delta","index":7,"delta":{"type":"input_json_delta","partial_json":"{\"n\":9007199"}}""")
                + Frame("""{"type":"content_block_delta","index":7,"delta":{"type":"input_json_delta","partial_json":"254740993}"}}""")
                + Frame("""{"type":"content_block_stop","index":7}""")
                + Frame("""{"type":"message_delta","usage":{"output_tokens":5}}""")
                + Frame("""{"type":"message_stop"}""");
            string plain = """
                {"model":"model-a","content":[{"type":"text","text":"I will inspect."},
                {"type":"thinking","thinking":"plan","signature":"sig"},
                {"type":"tool_use","id":"call","name":"inspect","input":{"n":9007199254740993}}],
                "usage":{"input_tokens":8,"output_tokens":5}}
                """;
            var chunks = new List<Chunk>();
            using var response = Response(stream);
            var parsed = await anthropic.Engine().ParseAsync(response, chunks.Add, deadline.Token);
            var final = parsed.Events.Single(fact => fact.Kind == EventKind.ResponseEnded);
            Check(parsed.PartIDs.SequenceEqual(new[] { 0, 2, 7 }), "Parser renumbered native part IDs.");
            Check(!parsed.ToolBeforeText, "Pre-tool explanation was lost.");
            Check(chunks.Where(c => c.PartID == 7).Select(c => c.ArgsText).Aggregate("", (a, b) => a + b)
                == "{\"n\":9007199254740993}", "Argument fragments changed precision or order.");
            Check(chunks.Where(c => c.PartID == 2).All(c => c.Kind == ChunkKind.Thinking), "Thinking became reply text.");
            using var plainResponse = Response(plain, false);
            var plainChunks = new List<Chunk>();
            var plainParsed = await anthropic.Engine().ParseAsync(plainResponse, plainChunks.Add, deadline.Token);
            var plainFinal = plainParsed.Events.Single(fact => fact.Kind == EventKind.ResponseEnded);
            Check(JsonCodec.WriteEvent(final) == JsonCodec.WriteEvent(plainFinal), "Stream/plain canonical content differs.");
            Check(plainChunks.Count == 0, "Plain mode fabricated streaming deltas.");
            Check(anthropic.History().Events().IsEmpty, "Parser changed History.");

            foreach (string broken in new[]
            {
                stream.Replace(Frame("""{"type":"message_stop"}"""), ""),
                stream.Replace(Frame("""{"type":"content_block_stop","index":7}"""), ""),
                stream + Frame("""{"type":"error","error":{"message":"secret remote detail"}}"""),
                Frame("{malformed")
            })
            {
                using var badResponse = Response(broken);
                var bad = await anthropic.Engine().ParseAsync(badResponse, cancellation: deadline.Token);
                Check(bad.Events.All(e => e.Kind != EventKind.ResponseEnded), "Broken stream became successful.");
                Check(!JsonCodec.WriteEvent(bad.Events.Single()).Contains("secret remote"), "Remote error leaked.");
            }
            using var errorResponse = Response(plain, false, HttpStatusCode.BadRequest);
            Check((await anthropic.Engine().ParseAsync(errorResponse)).Events.Single().Kind == EventKind.ErrorOccurred,
                "HTTP error decoded as success.");

            await OpenAI(app, deadline.Token);
            await Gemini(app, deadline.Token);

            // Missing/null protocol members are untrusted data errors, not
            // unexpected object-lifetime faults escaping the parser.
            foreach (var (vendor, broken) in new[]
            {
                (Vendor.Anthropic, "{\"type\":\"message_start\"}"),
                (Vendor.Anthropic, "{\"type\":\"content_block_start\"}"),
                (Vendor.Anthropic, "{\"type\":\"content_block_start\",\"index\":0}"),
                (Vendor.Anthropic, "{\"type\":\"content_block_delta\",\"index\":0}"),
                (Vendor.OpenAI, "{}"),
                (Vendor.OpenAI, "{\"choices\":[null]}"),
                (Vendor.OpenAI, "{\"choices\":[{}]}"),
                (Vendor.OpenAI, "{\"choices\":[{\"delta\":{\"tool_calls\":[null]}}]}"),
                (Vendor.OpenAI, "{\"choices\":[{\"delta\":{\"tool_calls\":[{}]}}]}"),
                (Vendor.Gemini, "{\"candidates\":[null]}"),
                (Vendor.Gemini, "{\"candidates\":[{\"content\":{\"parts\":[null]}}]}")
            })
            {
                var malformedAgent = new Agent(app, new Configuration(vendor, "model-a", "http://127.0.0.1"));
                using var malformedResponse = Response(Frame(broken));
                var malformed = await malformedAgent.Engine().ParseAsync(malformedResponse, cancellation: deadline.Token);
                Check(malformed.Events.Single().Kind == EventKind.ErrorOccurred,
                    "Malformed " + vendor + " frame escaped safe error handling.");
            }
        }
        finally
        {
            app.Destroy();
        }
    }

    private static async Task OpenAI(Application app, CancellationToken cancellation)
    {
        var agent = new Agent(app, new Configuration(Vendor.OpenAI, "model-a", "http://127.0.0.1"));
        string stream = Frame("""{"model":"model-a","choices":[{"index":0,"delta":{"tool_calls":[{"index":4,"id":"a","function":{"name":"fi","arguments":"{\"a\":"}}]}}]}""")
            + Frame("""{"choices":[{"index":0,"delta":{"content":"late explanation","tool_calls":[{"index":1,"id":"b","function":{"name":"second","arguments":"{}"}},{"index":4,"function":{"name":"rst","arguments":"1}"}}]}}]}""")
            + Frame("""{"choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}""")
            + Frame("""{"choices":[],"usage":{"prompt_tokens":12,"completion_tokens":4}}""")
            + Frame("[DONE]");
        var chunks = new List<Chunk>();
        using var response = Response(stream);
        var parsed = await agent.Engine().ParseAsync(response, chunks.Add, cancellation);
        var final = parsed.Events.Single(e => e.Kind == EventKind.ResponseEnded);
        Check(parsed.ToolBeforeText, "Late text retroactively explained a tool call.");
        Check(parsed.PartIDs.SequenceEqual(new[] { 1, 0, 2 }), "Final IDs inferred from sorted positions.");
        var tools = final.Parts.OfType<ToolCallPart>().ToArray();
        Check(tools.Length == 2 && tools[0].Name == "first" && tools[1].Name == "second", "Interleaved calls merged.");
        Check(tools[0].Args.GetProperty("a").GetInt32() == 1, "Split arguments not assembled.");
        Check(final.Usage!.Input == 12 && final.Usage.Output == 4, "Usage-only frame lost.");
        Check(string.Concat(chunks.Where(c => c.PartID == 0).Select(c => c.ToolName)) == "first", "Name fragments lost.");
        using var incomplete = Response(stream.Replace(Frame("[DONE]"), ""));
        Check((await agent.Engine().ParseAsync(incomplete)).Events.Single().Kind == EventKind.ErrorOccurred,
            "OpenAI EOF accepted without terminator.");
    }

    private static async Task Gemini(Application app, CancellationToken cancellation)
    {
        var agent = new Agent(app, new Configuration(Vendor.Gemini, "model-a", "http://127.0.0.1"));
        string stream = Frame("""{"modelVersion":"model-a","candidates":[{"content":{"parts":[{"text":"hel"}]}}]}""")
            + Frame("""{"candidates":[{"content":{"parts":[{"text":"lo","thoughtSignature":"s1"}]}}]}""")
            + Frame("""{"candidates":[{"content":{"parts":[{"text":"","thoughtSignature":"s2"},{"text":"plan","thought":true}]}}]}""")
            + Frame("""{"candidates":[{"content":{"parts":[{"thoughtSignature":"s3"}]}}]}""")
            + Frame("""{"candidates":[{"content":{"parts":[{"functionCall":{"id":"c","name":"inspect","args":{}},"thoughtSignature":"s4"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":9,"candidatesTokenCount":2}}""");
        var chunks = new List<Chunk>();
        using var response = Response(stream);
        var parsed = await agent.Engine().ParseAsync(response, chunks.Add, cancellation);
        var final = parsed.Events.Single(e => e.Kind == EventKind.ResponseEnded);
        Check(parsed.PartIDs.SequenceEqual(new[] { 0, 1, 2, 3 }), "Gemini part boundaries changed.");
        Check(final.Parts[0] is TextPart { Text: "hello" } first
            && first.Opaque!.Value.GetString() == "s1", "Split text/signature lost.");
        Check(final.Parts[1] is TextPart { Text: "" } empty
            && empty.Opaque!.Value.GetString() == "s2", "Empty signed text lost.");
        Check(final.Parts[2] is OpaquePart thought
            && thought.Data.GetProperty("thoughtSignature").GetString() == "s3", "Signature-only frame misbound.");
        Check(chunks.Any(c => c.PartID == 2 && c.Kind == ChunkKind.Thinking && c.Text == "plan"), "Thought not displayed.");
        Check(final.Parts[3] is ToolCallPart call
            && call.Opaque!.Value.GetString() == "s4", "Call signature lost.");
    }
}
