using System.Text.Json;
using System.Threading.Channels;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class ToolSchemaPolicyTests
{
    private static JsonElement Json(string text) => JsonSerializer.Deserialize<JsonElement>(text);
    private static void Check(bool condition, string message)
    {
        if (!condition) throw new InvalidOperationException(message);
    }

    private sealed class Peer(JsonElement schema) : IMcpTransport
    {
        private readonly Channel<JsonElement> replies = Channel.CreateUnbounded<JsonElement>();
        public int Calls { get; private set; }
        public Task SendAsync(JsonElement message, CancellationToken cancellation = default)
        {
            if (!message.TryGetProperty("id", out var id)) return Task.CompletedTask;
            object result;
            switch (message.GetProperty("method").GetString())
            {
                case "initialize":
                    result = new { protocolVersion = "2024-11-05", capabilities = new { tools = new { } } };
                    break;
                case "tools/list":
                    result = new { tools = new[] { new { name = "remote_domain", description = "Synthetic", inputSchema = schema } } };
                    break;
                case "tools/call":
                    Calls++;
                    result = new { content = new[] { new { type = "text", text = "accepted" } } };
                    break;
                default: throw new InvalidOperationException("Unexpected fixture RPC");
            }
            if (!replies.Writer.TryWrite(JsonSerializer.SerializeToElement(new { jsonrpc = "2.0", id, result })))
                throw new IOException("Fixture closed");
            return Task.CompletedTask;
        }
        public async Task<JsonElement?> ReceiveAsync(CancellationToken cancellation = default)
        {
            try { return await replies.Reader.ReadAsync(cancellation); }
            catch (ChannelClosedException) { return null; }
        }
        public ValueTask DisposeAsync() { replies.Writer.TryComplete(); return ValueTask.CompletedTask; }
    }

    public static async Task RunAsync()
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "offline", "http://127.0.0.1:1"));
            var vendor = new ModelVendor(app, "fixture", "Fixture");
            var endpoint = new ModelEndpoint(vendor, "local", "http://127.0.0.1:1", Surface.Messages,
                new ModelAuthentication(app, "none"));
            var adapted = new Model(endpoint, "adapted", "adapted", "Adapted", new ModelFeatures(Media.None),
                requestPolicy: new ModelRequestPolicy(toolSchemas: ToolSchemaPolicy.PatternLengthCompatibility));
            var normal = new Model(endpoint, "normal", "normal", "Normal", new ModelFeatures(Media.None));
            var schema = Json("""
                {"type":"object","properties":{"domains":{"type":"array","items":{
                  "type":"string","pattern":"^[a-z.]+$","minLength":3,"maxLength":12}},
                  "pattern":{"type":"string"}},"required":["domains"],"additionalProperties":false}
                """);
            int effects = 0;
            var tool = new Tool(agent, "local_domain", "Synthetic", schema, (job, args, cancellation) =>
            {
                effects++;
                return Task.FromResult("accepted");
            });
            var peer = new Peer(schema);
            var client = new McpClient(agent.EnsureMcp(), "schema-fixture", peer);
            await client.InstallToolsAsync(deadline.Token);
            await agent.SetModelAsync(adapted);
            var rendered = agent.Engine().Render(new Context());
            var wireItem = rendered["tools"]![0]!["input_schema"]!["properties"]!["domains"]!["items"]!;
            Check(wireItem["maxLength"] is null && wireItem["pattern"]!.GetValue<string>() == "^[a-z.]+$"
                && wireItem["minLength"]!.GetValue<int>() == 3, "Adaptation changed more than maximum length.");
            Check(tool.Declaration(Vendor.Anthropic)["input_schema"]!["properties"]!["domains"]!["items"]!["maxLength"]!.GetValue<int>() == 12,
                "Projection mutated the Tool's original schema.");
            Check(rendered["tools"]![0]!["input_schema"]!["properties"]!["pattern"] is not null,
                "A property named pattern was treated as a schema keyword.");
            await agent.SetModelAsync(normal);
            Check(agent.Engine().Render(new Context())["tools"]![0]!["input_schema"]!["properties"]!["domains"]!["items"]!["maxLength"] is not null,
                "Schema policy leaked to another offering.");
            // Validate even after selection changes; the contract belongs to the Tool.
            foreach (string name in new[] { "local_domain", "remote_domain" })
            {
                int index = 0;
                foreach (string bad in new[] { "{\"domains\":[\"abcdefghijklmnop.com\"]}",
                    "{\"domains\":[\"BAD.COM\"]}", "{\"domains\":[\"a\"]}", "{\"domains\":[42]}" })
                {
                    string id = name + "-bad-" + index++;
                    var refused = await agent.DispatchAsync(new ToolCallPart(id, agent.Configuration().Target(), name, Json(bad)), deadline.Token);
                    Check(refused.IsError && effects == 0 && peer.Calls == 0
                        && !agent.History().Snapshot().Events.Any(fact => fact.Kind == EventKind.ToolStarted && fact.CallID == id),
                        "Invalid nested input reached tool admission or execution.");
                }
            }
            foreach (string name in new[] { "local_domain", "remote_domain" })
            {
                var accepted = await agent.DispatchAsync(new ToolCallPart(name + "-good", agent.Configuration().Target(), name,
                    Json("{\"domains\":[\"example.com\"]}")), deadline.Token);
                Check(!accepted.IsError, "Valid input was rejected.");
            }
            Check(effects == 1 && peer.Calls == 1, "Valid local/remote effects must occur exactly once.");
            var unicode = new Tool(agent, "unicode", "Synthetic", Json("""
                {"type":"object","properties":{"text":{"type":"string","pattern":"^.*$","maxLength":1}},"required":["text"]}
                """), (job, args, cancellation) => Task.FromResult("accepted"));
            var scalar = await unicode.ExecuteAsync(new ToolCallPart("unicode-good", agent.Configuration().Target(), "unicode",
                Json("{\"text\":\"😀\"}")), deadline.Token);
            Check(!scalar.IsError, "maxLength counts Unicode scalars, not UTF-16 code units.");
            var nested = new Tool(agent, "nested", "Synthetic", Json("""
                {"type":"object","properties":{"box":{"type":"object","properties":{
                  "text":{"type":"string","pattern":"^a$","maxLength":1}}}}}
                """), (job, args, cancellation) => throw new InvalidOperationException("Must not execute"));
            bool duplicateRefused = false;
            try
            {
                var duplicate = await nested.ExecuteAsync(new ToolCallPart("duplicate", agent.Configuration().Target(), "nested",
                    Json("{\"box\":{\"text\":\"too long\",\"text\":\"a\"}}")), deadline.Token);
                duplicateRefused = duplicate.IsError;
            }
            catch (ArgumentException)
            {
                // Canonical History also refuses duplicate keys before recording
                // an ordinary paired result. Neither path may start the tool.
                duplicateRefused = true;
            }
            Check(duplicateRefused && !agent.History().Snapshot().Events.Any(
                fact => fact.Kind == EventKind.ToolStarted && fact.CallID == "duplicate"),
                "Duplicate nested fields must not split validator and peer interpretation.");
            var conditional = new Tool(agent, "conditional", "Synthetic", Json("""
                {"type":"object","properties":{"choice":{"type":"string","anyOf":[
                  {"type":"string","pattern":"^a$","maxLength":1},{"const":"other"}]}}}
                """), (job, args, cancellation) => Task.FromResult("accepted"));
            Check(conditional.Declaration(Vendor.Anthropic, ToolSchemaPolicy.PatternLengthCompatibility)
                ["input_schema"]!["properties"]!["choice"]!["anyOf"]![0]!["maxLength"] is not null,
                "Conditional schema was rewritten without supported validation semantics.");
        }
        finally { app.Destroy(); }
    }
}
