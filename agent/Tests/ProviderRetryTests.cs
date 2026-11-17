using System.Diagnostics;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    private sealed class RetryBrokenStream : MemoryStream
    {
        public override ValueTask<int> ReadAsync(Memory<byte> buffer, CancellationToken token = default)
            => ValueTask.FromException<int>(new IOException("synthetic network break"));
    }

    private sealed class RetryTransport(bool broken) : HttpMessageHandler
    {
        protected override Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken token)
        {
            HttpContent content = broken ? new StreamContent(new RetryBrokenStream())
                : new StringContent("{\"choices\":[]}");
            content.Headers.ContentType = new System.Net.Http.Headers.MediaTypeHeaderValue(
                broken ? "text/event-stream" : "application/json");
            return Task.FromResult(new HttpResponseMessage(System.Net.HttpStatusCode.OK) { Content = content });
        }
    }

    private sealed class RetryBrokenWriter : StringWriter
    {
        private int writes;
        public override void WriteLine(string? value)
        {
            if (++writes == 2) throw new IOException("synthetic local diagnostic failure");
            base.WriteLine(value);
        }
    }

    public static async Task ProviderRetryKeepsTask(string temp)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "retry-test",
                listener.Prefixes.Single(), surface: Surface.ChatCompletions, disableStreaming: true));
            Task<string> task = agent.SubmitAsync([new TextPart("Keep this original task.")]);
            var first = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
            var elapsed = Stopwatch.StartNew();
            await Reply(first, "{}", 503);
            var second = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(4));
            Check(!task.IsCompleted, "Provider failure prematurely completed the original task");
            Check(elapsed.Elapsed.TotalSeconds >= .85, "Provider retry bypassed backoff");
            using (var reader = new StreamReader(second.Request.InputStream))
            {
                string wire = await reader.ReadToEndAsync();
                Check(wire.Contains("Provider failure (runtime data)") && wire.Contains("503"),
                    "Recovery request omitted the recorded provider error");
            }
            await Reply(second, "{}", 503);
            using (var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(3)))
            {
                while (agent.History().Events().Count(e => e.Kind == EventKind.ErrorOccurred) < 2)
                    await Task.Delay(10, deadline.Token);
            }
            await agent.GoalStatusAsync();
            // A status query is not progress and must not release the ten-second wait.
            var afterHint = listener.GetContextAsync();
            await Task.Delay(150);
            Check(!afterHint.IsCompleted, "Recovery bookkeeping reset its own backoff");
            await agent.PostAsync(new AgentInput(InputKind.Message, "Network is available again."));
            second = await afterHint.WaitAsync(TimeSpan.FromSeconds(3));
            Check(!task.IsCompleted, "Human input replaced the original task during recovery");
            using (var reader = new StreamReader(second.Request.InputStream))
            {
                string wire = await reader.ReadToEndAsync();
                Check(wire.Contains("Network is available again."), "Recovery omitted new human input");
            }
            // Human progress resets the streak, not merely the current timer.
            await Reply(second, "{}", 503);
            second = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(4));
            await Reply(second, "{\"choices\":[{\"message\":{\"role\":\"assistant\",\"content\":\"recovered\"},\"finish_reason\":\"stop\"}]}");
            Check(await task.WaitAsync(TimeSpan.FromSeconds(5)) == "recovered", "Retry lost task reply");
            Check(agent.History().Events().Count(e => e.Kind == EventKind.MessageReceived && e.Actor == Actor.Human) == 1,
                "Retry duplicated original input");
            Check(!(await agent.GoalStatusAsync()).Enabled, "Provider retry enabled goal reminders");

            // A recovered exchange resets the streak. Pause holds the next
            // retry without completing its admission; resume uses the deadline.
            var resumed = agent.SubmitAsync([new TextPart("Recover again after pause.")]);
            var failed = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
            await agent.PostAsync(new AgentInput(InputKind.Pause));
            await Reply(failed, "{}", 503);
            var pending = listener.GetContextAsync();
            await Task.Delay(1250);
            Check(!pending.IsCompleted && !resumed.IsCompleted, "Pause dispatched or finished retry");
            await agent.PostAsync(new AgentInput(InputKind.Resume));
            var recovery = await pending.WaitAsync(TimeSpan.FromSeconds(3));
            await Reply(recovery, "{\"choices\":[{\"message\":{\"content\":\"resumed\"},\"finish_reason\":\"stop\"}]}");
            Check(await resumed.WaitAsync(TimeSpan.FromSeconds(5)) == "resumed",
                "Resume lost original reply or previous success failed to reset backoff");

            int effects = 0;
            using var schema = JsonDocument.Parse("{\"type\":\"object\",\"properties\":{}}");
            _ = new Tool(agent, "effect", "Count one real effect", schema.RootElement,
                (job, arguments, cancellation) =>
                {
                    effects++;
                    return Task.FromResult("recorded-result");
                });
            var toolTask = agent.SubmitAsync([new TextPart("Perform one effect.")]);
            var toolRequest = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
            await Reply(toolRequest, """
                {"choices":[{"message":{"tool_calls":[{"id":"once","type":"function","function":{"name":"effect","arguments":"{}"}}]},"finish_reason":"tool_calls"}]}
                """);
            // Both the failed continuation and its retry must contain the same
            // completed result. Neither request may execute the tool again.
            for (int attempt = 0; attempt < 2; attempt++)
            {
                var continuation = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
                using var reader = new StreamReader(continuation.Request.InputStream);
                using var body = JsonDocument.Parse(await reader.ReadToEndAsync());
                var results = body.RootElement.GetProperty("messages").EnumerateArray()
                    .Where(message => message.GetProperty("role").GetString() == "tool").ToArray();
                Check(results.Length == 1 && results[0].GetProperty("content").GetString() == "recorded-result"
                    && effects == 1, "Retry lost result or repeated completed effect");
                if (attempt == 0)
                {
                    await Reply(continuation, "{}", 503);
                }
                else
                {
                    await Reply(continuation, "{\"choices\":[{\"message\":{\"content\":\"done\"},\"finish_reason\":\"stop\"}]}");
                }
            }
            Check(await toolTask.WaitAsync(TimeSpan.FromSeconds(5)) == "done" && effects == 1,
                "Tool continuation did not finish exactly once");
        }
        finally { listener.Stop(); app.Destroy(); }
    }
    public static async Task ProviderRetryAdvice(string temp)
    {
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Gemini, "advice-test", "http://localhost"));
            var marked = new Event(EventKind.ErrorOccurred, seq: 1, error: "HTTP 503",
                reason: "provider_failure", parts: [new TextPart("Provider failure (runtime data): HTTP 503")]);
            var replayed = History.Reduce(agent.History().View(),
                JsonCodec.ReadEvent(JsonCodec.WriteEvent(marked)));
            Check(replayed.Dialogue.Length == 1 && replayed.Dialogue[0].Actor == Actor.Tool,
                "Replayed provider error lost its data role or duplicated its notice");
            var localError = new Event(EventKind.ErrorOccurred, seq: 2, error: "local failure");
            Check(History.Reduce(replayed, JsonCodec.ReadEvent(JsonCodec.WriteEvent(localError)))
                .Dialogue.Length == 1, "Local error unexpectedly entered model dialogue");
            async Task<ParsedResponse> Parse(string body, int code = 503, string? header = null,
                string media = "application/json")
            {
                using var response = new HttpResponseMessage((System.Net.HttpStatusCode)code)
                { Content = new StringContent(body, System.Text.Encoding.UTF8, media) };
                if (header is not null) response.Headers.TryAddWithoutValidation("Retry-After", header);
                return await agent.Engine().ParseAsync(response);
            }
            string advice = """{"error":{"code":503,"details":[{"@type":"type.googleapis.com/google.rpc.RetryInfo","retryDelay":"2.125s"}],"message":"private payload"}}""";
            var parsed = await Parse(advice);
            Check(parsed.ProviderFailure && parsed.RetryAfter == TimeSpan.FromSeconds(2.125), "Gemini delay lost");
            Check(!parsed.Events.Single().Error.Contains("private"), "Remote payload leaked");
            Check((await Parse(advice, header: "4")).RetryAfter == TimeSpan.FromSeconds(4), "Longer header ignored");
            Check((await Parse(advice, 200)).ProviderFailure, "HTTP200 error not retryable");
            Check((await Parse(advice, 200, "4")).RetryAfter == TimeSpan.FromSeconds(4), "HTTP200 header ignored");
            Check(!(await Parse("{\"error\":null}", 200)).ProviderFailure, "Null error became retryable failure");
            Check((await Parse("{\"type\":\"error\",\"error\":null}", 200)).ProviderFailure,
                "Explicit null-detail error was treated as success");
            Check((await Parse(advice.Replace("2.125s", "NaNs"))).RetryAfter is null, "Invalid duration accepted");
            Check((await Parse("{", 200)).ProviderFailure == false, "Malformed response became provider retry");
            var stream = await Parse("data: " + advice + "\n\n", 200, media: "text/event-stream");
            Check(stream.ProviderFailure && stream.RetryAfter == TimeSpan.FromSeconds(2.125), "Stream advice lost");
            Check((await Parse("data: " + advice + "\n\n", 200, "4", "text/event-stream")).RetryAfter
                == TimeSpan.FromSeconds(4), "Stream header advice lost");
            var responsesAgent = new Agent(app, new Configuration(Vendor.OpenAI, "responses-retry",
                "http://localhost", surface: Surface.Responses));
            foreach (string kind in new[] { "error", "response.failed", "response.incomplete" })
            {
                using var response = new HttpResponseMessage(System.Net.HttpStatusCode.OK)
                {
                    Content = new StringContent("data: {\"type\":\"" + kind + "\"}\n\n",
                        System.Text.Encoding.UTF8, "text/event-stream")
                };
                response.Headers.TryAddWithoutValidation("Retry-After", "4");
                var result = await responsesAgent.Engine().ParseAsync(response);
                bool remoteFailure = kind != "response.incomplete";
                Check(result.ProviderFailure == remoteFailure && result.PartIDs.Length == 0,
                    "Responses failure classification or discarded parts incorrect");
                if (remoteFailure)
                {
                    Check(result.RetryAfter == TimeSpan.FromSeconds(4), "Responses header advice lost");
                }
            }
            agent.Engine().Destroy();
            var brokenEngine = new Engine(agent, new RetryTransport(true));
            Check((await brokenEngine.SendAsync("{}")).ProviderFailure, "Network read failure was not retryable");
            brokenEngine.Destroy();
            var localEngine = new Engine(agent, new RetryTransport(false));
            agent.AttachDiagnostics(TextWriter.Null, new RetryBrokenWriter());
            try
            {
                await localEngine.SendAsync("{}");
                throw new Exception("Local diagnostic failure became a response");
            }
            catch (IOException error) when (error.Message == "synthetic local diagnostic failure") { }
            var delay = typeof(Agent).GetMethod("FailureRetryDelay", System.Reflection.BindingFlags.NonPublic
                | System.Reflection.BindingFlags.Static)!;
            foreach (var (count, seconds) in new[] { (1, 1), (2, 10), (3, 100), (4, 1000), (99, 1000) })
                Check((TimeSpan)delay.Invoke(null, [count])! == TimeSpan.FromSeconds(seconds), "Wrong shared retry schedule");
        }
        finally { app.Destroy(); }
    }

    public static async Task ProviderRetryPauseAndStop(string temp)
    {
        foreach (bool shutdown in new[] { false, true })
        {
            using var listener = Listen();
            var app = new Application(TextWriter.Null);
            try
            {
                var agent = new Agent(app, new Configuration(Vendor.OpenAI, "retry-stop",
                    listener.Prefixes.Single(), surface: Surface.ChatCompletions, disableStreaming: true));
                var task = agent.SubmitAsync([new TextPart("pause recovery")]);
                var first = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
                await agent.PostAsync(new AgentInput(InputKind.Pause));
                await Reply(first, "{}", 503);
                await agent.PostAsync(new AgentInput(InputKind.Message, "Remain paused despite new input."));
                var next = listener.GetContextAsync();
                await Task.Delay(1250);
                Check(!next.IsCompleted && !task.IsCompleted, "Paused retry made progress");
                if (shutdown) await agent.ShutdownAsync();
                else await agent.PostAsync(new AgentInput(InputKind.InterruptNow));
                try { await task.WaitAsync(TimeSpan.FromSeconds(3)); throw new Exception("Stopped task succeeded"); }
                catch (Exception error) when (error is OperationCanceledException or ObjectDisposedException) { }
                if (!shutdown) await agent.PostAsync(new AgentInput(InputKind.Resume));
                await Task.Delay(150);
                Check(!next.IsCompleted, "Stopped retry timer sent another request");
            }
            finally { listener.Stop(); app.Destroy(); }
        }
    }

}
