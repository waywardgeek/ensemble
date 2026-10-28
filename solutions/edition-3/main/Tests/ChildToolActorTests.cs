using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    public static async Task ChildToolSubmission(string temp)
    {
        string work = Path.Combine(temp, "model-child-submit");
        Directory.CreateDirectory(work);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        int requests = 0;
        var server = Task.Run(async () =>
        {
            using (var socket = await listener.AcceptTcpClientAsync(deadline.Token))
            using (var stream = socket.GetStream())
            {
                string request = await ReadRequest(stream, deadline.Token);
                using var json = JsonDocument.Parse(request);
                JsonElement declaration = json.RootElement.GetProperty("tools").EnumerateArray()
                    .Single(tool => tool.GetProperty("name").GetString() == "submit");
                Check(declaration.GetProperty("input_schema").GetProperty("properties").TryGetProperty("count", out _),
                    "Provider did not receive the child's structured submit schema.");
                requests++;
                await ReplyJson(stream, new
                {
                    model = "fixture",
                    content = new object[] {
                        new { type = "text", text = "I will submit the checked count." },
                        new { type = "tool_use", id = "typed-submit", name = "submit", input = new { count = 7 } }
                    },
                    usage = new { input_tokens = 1, output_tokens = 1 }
                }, deadline.Token);
            }
            using (var socket = await listener.AcceptTcpClientAsync(deadline.Token))
            using (var stream = socket.GetStream())
            {
                string request = await ReadRequest(stream, deadline.Token);
                using var json = JsonDocument.Parse(request);
                JsonElement result = json.RootElement.GetProperty("messages").EnumerateArray()
                    .SelectMany(message => message.GetProperty("content").EnumerateArray())
                    .Single(part => part.GetProperty("type").GetString() == "tool_result"
                        && part.GetProperty("tool_use_id").GetString() == "typed-submit");
                Check(!result.TryGetProperty("is_error", out JsonElement error) || !error.GetBoolean(),
                    "Model-led submit returned a tool error.");
                requests++;
                await Reply(stream, "Submitted.", deadline.Token);
            }
        });
        var app = new Application(TextWriter.Null);
        try
        {
            var config = new Configuration(Vendor.Anthropic, "fixture",
                "http://127.0.0.1:" + ((IPEndPoint)listener.LocalEndpoint).Port, "local-only", sandboxRoot: work);
            var parent = new Agent(app, config, dataDirectory: work);
            parent.StartPersistence(new PersistenceOptions(Path.Combine(work, "save.json")));
            JsonElement schema = JsonSerializer.Deserialize<JsonElement>("""
                {"type":"object","properties":{"count":{"type":"integer"}},"required":["count"],"additionalProperties":false}
                """);
            Agent child = await parent.CreateChildAsync("worker", submitSchema: schema, cancellation: deadline.Token);
            Task<string> turn = child.SubmitAsync([new TextPart("Count and submit.")]);
            JsonElement payload = await child.WaitForSubmissionAsync(deadline.Token);
            Check(payload.GetProperty("count").GetInt32() == 7, "Tool did not publish typed completion.");
            await turn.WaitAsync(deadline.Token);
            await server;
            Check(requests == 2 && child.ChildFailure() is null, "Submit tool triggered corrections or failed draining.");
            Check(child.History().Events().Any(e => e.Kind == EventKind.ToolReturned && e.CallID == "typed-submit"),
                "Submission bypassed canonical tool-result recording.");
        }
        finally
        {
            deadline.Cancel();
            listener.Stop();
            app.Destroy();
            try { await server; }
            catch (Exception error) when (error is OperationCanceledException or IOException or SocketException) { }
        }
    }
}
