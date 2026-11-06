using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

return await Probe.Run();

internal static partial class Probe
{
    private static readonly List<string> checks = [];

    private static void Require(bool condition, string name)
    {
        if (!condition) throw new InvalidOperationException(name);
        lock (checks) checks.Add(name);
    }

    public static async Task<int> Run()
    {
        string? failure = null;
        try
        {
            foreach (string mode in new[] { "buffered", "streamed", "prose" })
                await Conversation(mode);
            await ConfigurationChecks();
            await LifecycleChecks();
        }
        catch (Exception error) { failure = error.Message; }
        Console.WriteLine(JsonSerializer.Serialize(new { passed = failure is null, checks, failure }));
        return failure is null ? 0 : 1;
    }

    private static async Task Conversation(string mode)
    {
        // The candidate receives only ordinary public configuration and HTTP replies.
        // Random payloads are generated here, never supplied by candidate tests.
        string nonce = Guid.NewGuid().ToString("N");
        string callId = "call-" + nonce;
        string resultText = "result-" + nonce;
        string finalText = "answer-" + nonce;
        string modelName = "local-" + nonce;
        var portReservation = new TcpListener(IPAddress.Loopback, 0);
        portReservation.Start();
        int port = ((IPEndPoint)portReservation.LocalEndpoint).Port;
        portReservation.Stop();
        using var listener = new HttpListener();
        string address = $"http://127.0.0.1:{port}/";
        listener.Prefixes.Add(address);
        listener.Start();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var app = new Application(TextWriter.Null);
        Task? server = null;
        Task<string>? turn = null;
        int effects = 0;
        int requests = 0;
        try
        {
            var auth = new ModelAuthentication(app, "none");
            var vendor = new ModelVendor(app, "local", "Grader-owned endpoint");
            var endpoint = new ModelEndpoint(vendor, "chat", address + "local",
                Surface.ChatCompletions, auth);
            var model = new Model(endpoint, "small", modelName, "Small",
                new ModelFeatures(Media.None, Stream: mode == "streamed", ContextWindow: 262144), 4096);
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "legacy",
                address + "forbidden", "synthetic-old-key", disableStreaming: true));
            await agent.SetModelAsync(model);
            Require(ReferenceEquals(agent.SelectedModel(), model) && model.Agents().Contains(agent), mode + "-selection-membership");
            Require(agent.WorkingContextStatus().WorkingTokens == 4096, mode + "-serving-not-advertised-context");
            using var schema = JsonDocument.Parse("{\"type\":\"object\",\"properties\":{\"nonce\":{\"type\":\"string\"}},\"required\":[\"nonce\"]}");
            _ = new Tool(agent, "echo_nonce", "Return the supplied nonce", schema.RootElement,
                (job, arguments, cancellation) =>
                {
                    if (arguments.GetProperty("nonce").GetString() != nonce)
                        throw new InvalidOperationException("Callback arguments differ from the supplied nonce.");
                    Interlocked.Increment(ref effects);
                    return Task.FromResult(resultText);
                });
            int expectedRequests = mode == "prose" ? 1 : 2;
            server = Task.Run(async () =>
            {
                for (int index = 0; index < expectedRequests; index++)
                {
                    var incoming = await listener.GetContextAsync().WaitAsync(deadline.Token);
                    Interlocked.Increment(ref requests);
                    try
                    {
                        Require(incoming.Request.Url!.AbsolutePath == "/local/v1/chat/completions", mode + "-route-" + index);
                        Require(incoming.Request.Headers["Authorization"] is null
                            && incoming.Request.Headers["x-api-key"] is null
                            && incoming.Request.Headers["x-goog-api-key"] is null, mode + "-no-borrowed-credentials-" + index);
                        using var reader = new StreamReader(incoming.Request.InputStream);
                        using var request = JsonDocument.Parse(await reader.ReadToEndAsync(deadline.Token));
                        var root = request.RootElement;
                        Require(root.GetProperty("model").GetString() == modelName, mode + "-wire-name-" + index);
                        bool streams = root.TryGetProperty("stream", out var stream) && stream.GetBoolean();
                        Require(streams == (mode == "streamed"), mode + "-declared-stream-policy-" + index);
                        if (index == 1)
                        {
                            var messages = root.GetProperty("messages").EnumerateArray().ToArray();
                            var result = messages.Single(message => message.GetProperty("role").GetString() == "tool");
                            Require(effects == 1 && result.GetProperty("tool_call_id").GetString() == callId
                                && result.GetProperty("content").GetString() == resultText, mode + "-real-correlated-result");
                            Require(agent.History().Events().Count(fact => fact.Kind == EventKind.ToolReturned) == 1,
                                mode + "-result-recorded-before-next-request");
                        }
                        if (mode == "streamed")
                        {
                            incoming.Response.ContentType = "text/event-stream";
                            incoming.Response.SendChunked = true;
                            async Task Frame(object frame)
                            {
                                byte[] bytes = Encoding.UTF8.GetBytes("data: " + JsonSerializer.Serialize(frame) + "\n\n");
                                await incoming.Response.OutputStream.WriteAsync(bytes, deadline.Token);
                                await incoming.Response.OutputStream.FlushAsync(deadline.Token);
                            }
                            if (index == 0)
                            {
                                await Frame(new { model = modelName, choices = new[] { new { index = 0, delta = new {
                                    tool_calls = new[] { new { index = 0, id = callId, function = new { name = "echo_nonce", arguments = "{" } } }
                                } } } });
                                await release.Task.WaitAsync(deadline.Token);
                                string arguments = JsonSerializer.Serialize(new { nonce });
                                await Frame(new { choices = new[] { new { index = 0, delta = new {
                                    tool_calls = new[] { new { index = 0, function = new { arguments = arguments[1..] } } }
                                }, finish_reason = "tool_calls" } } });
                            }
                            else
                            {
                                await Frame(new { model = modelName, choices = new[] { new { index = 0,
                                    delta = new { content = finalText }, finish_reason = "stop" } } });
                                await Frame(new { choices = Array.Empty<object>(), usage = new { prompt_tokens = 37,
                                    completion_tokens = 3, prompt_tokens_details = new { cached_tokens = 11 } } });
                            }
                            await incoming.Response.OutputStream.WriteAsync(Encoding.UTF8.GetBytes("data: [DONE]\n\n"), deadline.Token);
                        }
                        else
                        {
                            JsonNode message;
                            if (mode == "prose")
                            {
                                // A tool-shaped string is still ordinary assistant content.
                                message = new JsonObject { ["content"] = JsonSerializer.Serialize(new {
                                    name = "echo_nonce", arguments = new { nonce } }) };
                            }
                            else if (index == 0)
                            {
                                message = JsonSerializer.SerializeToNode(new { content = "", tool_calls = new[] {
                                    new { id = callId, type = "function", function = new { name = "echo_nonce",
                                        arguments = JsonSerializer.Serialize(new { nonce }) } } } })!;
                            }
                            else message = new JsonObject { ["content"] = finalText };
                            var response = new JsonObject { ["model"] = modelName,
                                ["choices"] = new JsonArray(new JsonObject { ["message"] = message }),
                                ["usage"] = JsonSerializer.SerializeToNode(new { prompt_tokens = 37, completion_tokens = 3,
                                    prompt_tokens_details = new { cached_tokens = 11 } }) };
                            byte[] bytes = Encoding.UTF8.GetBytes(response.ToJsonString());
                            incoming.Response.ContentType = "application/json";
                            incoming.Response.ContentLength64 = bytes.Length;
                            await incoming.Response.OutputStream.WriteAsync(bytes, deadline.Token);
                        }
                        incoming.Response.Close();
                    }
                    finally { incoming.Response.Abort(); }
                }
            });
            turn = agent.SendAsync("Use echo_nonce for " + nonce);
            if (mode == "streamed")
            {
                await agent.WaitAsync(observation => observation.Delta?.Kind == ChunkKind.ToolCall, deadline.Token);
                Require(effects == 0 && !turn.IsCompleted
                    && !agent.History().Events().Any(fact => fact.Kind is EventKind.ToolCalled or EventKind.ResponseEnded),
                    "streamed-partial-has-no-execution-authority");
                release.TrySetResult();
            }
            string answer = await turn.WaitAsync(deadline.Token);
            await server.WaitAsync(deadline.Token);
            Require(mode == "prose" ? answer.Contains(nonce, StringComparison.Ordinal) : answer == finalText, mode + "-final-answer");
            Require(effects == (mode == "prose" ? 0 : 1) && requests == expectedRequests, mode + "-exact-effects-and-requests");
            var responses = agent.History().Events().Where(fact => fact.Kind == EventKind.ResponseEnded).ToArray();
            Require(responses.Length == expectedRequests && responses.All(fact => fact.From!.Offering == model.Identity()), mode + "-qualified-response-provenance");
            Require(agent.WorkingContextStatus().InputTokens == 37, mode + "-cached-input-counted-once");
            var meter = agent.Engine().Meter();
            Require(meter.Latest is not null && meter.Latest.Identity.Model == modelName
                && meter.Latest.Identity.Billing == BillingMode.None
                && meter.Latest.Counts.AllInput() == 37 && meter.EstimatedUsd() is null,
                mode + "-known-usage-unknown-cost");
        }
        finally
        {
            release.TrySetResult();
            deadline.Cancel();
            listener.Stop();
            try
            {
                try { await Task.WhenAll(new Task?[] { server, turn }.OfType<Task>()); }
                catch (Exception error) when (error is OperationCanceledException or HttpListenerException or ObjectDisposedException) { }
            }
            finally { app.Destroy(); }
        }
    }
}
