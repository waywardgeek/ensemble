using System.Net;
using System.Text;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    public static Task ConfiguredStreamingTools() => ConfiguredStreamingCase("success");

    public static async Task ConfiguredStreamingFailures()
    {
        await ConfiguredStreamingCase("interrupt");
        await ConfiguredStreamingCase("provider-error");
    }

    private static async Task ConfiguredStreamingCase(string mode)
    {
        using var listener = Listen();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var app = new Application(TextWriter.Null);
        Task? server = null;
        Task<string>? turn = null;
        int effects = 0;
        int requests = 0;
        const string wireName = "uncatalogued-local-stream";
        try
        {
            var vendor = new ModelVendor(app, "local-stream", "Owned local fixture");
            var authentication = new ModelAuthentication(app, "none");
            var endpoint = new ModelEndpoint(vendor, "chat", listener.Prefixes.Single(),
                Surface.ChatCompletions, authentication);
            var model = new Model(endpoint, "small", wireName, "Small",
                new ModelFeatures(Media.None, Stream: true, ContextWindow: 262144), 4096);
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "legacy",
                "http://127.0.0.1:1", "old-key-must-not-leak", disableStreaming: true));
            await agent.SetModelAsync(model);
            using var schema = JsonDocument.Parse("{\"type\":\"object\",\"properties\":{\"value\":{\"type\":\"integer\"}},\"required\":[\"value\"]}");
            _ = new Tool(agent, "record_value", "Record a value", schema.RootElement,
                (job, arguments, cancellation) =>
                {
                    Check(arguments.GetProperty("value").GetInt32() == 42, "Final streamed arguments changed.");
                    Interlocked.Increment(ref effects);
                    return Task.FromResult("recorded-42");
                });

            async Task Write(HttpListenerResponse response, string data)
            {
                await response.OutputStream.WriteAsync(Encoding.UTF8.GetBytes(data), deadline.Token);
                await response.OutputStream.FlushAsync(deadline.Token);
            }
            server = Task.Run(async () =>
            {
                for (int index = 0; index < 2; index++)
                {
                    var incoming = await listener.GetContextAsync().WaitAsync(deadline.Token);
                    Interlocked.Increment(ref requests);
                    try
                    {
                        Check(incoming.Request.Headers["Authorization"] is null
                            && incoming.Request.Headers["x-api-key"] is null,
                            "Local stream inherited credentials.");
                        using var reader = new StreamReader(incoming.Request.InputStream);
                        using var body = JsonDocument.Parse(await reader.ReadToEndAsync(deadline.Token));
                        Check(body.RootElement.GetProperty("model").GetString() == wireName
                            && body.RootElement.GetProperty("stream").GetBoolean(),
                            "Configured unknown model did not request its declared stream.");
                        if (index == 1 && mode == "success")
                        {
                            var messages = body.RootElement.GetProperty("messages").EnumerateArray().ToArray();
                            var result = messages.Single(message => message.GetProperty("role").GetString() == "tool");
                            Check(result.GetProperty("tool_call_id").GetString() == "local-call"
                                && result.GetProperty("content").GetString() == "recorded-42"
                                && effects == 1, "Follow-up request lost the real correlated tool result.");
                            Check(agent.History().Events().Count(fact => fact.Kind == EventKind.ToolReturned) == 1,
                                "Follow-up request preceded canonical tool-result recording.");
                        }
                        if (index == 1 && mode != "success")
                        {
                            Check(!body.RootElement.GetProperty("messages").EnumerateArray().Any(message =>
                                message.GetProperty("role").GetString() == "tool" || message.TryGetProperty("tool_calls", out _)),
                                "Fresh request replayed an orphan call or result from the partial stream.");
                        }
                        incoming.Response.ContentType = "text/event-stream";
                        incoming.Response.SendChunked = true;
                        if (index == 0)
                        {
                            await Write(incoming.Response, "data: {\"model\":\"" + wireName + "\",\"choices\":[{\"index\":0,\"delta\":{\"tool_calls\":[{\"index\":0,\"id\":\"local-call\",\"function\":{\"name\":\"record_value\",\"arguments\":\"{\\\"value\\\":\"}}]}}]}\n\n");
                            await release.Task.WaitAsync(deadline.Token);
                            try
                            {
                                if (mode == "provider-error")
                                {
                                    await Write(incoming.Response, "data: {\"error\":{\"message\":\"synthetic stream failure\"}}\n\ndata: [DONE]\n\n");
                                }
                                else
                                {
                                    await Write(incoming.Response, "data: {\"choices\":[{\"index\":0,\"delta\":{\"tool_calls\":[{\"index\":0,\"function\":{\"arguments\":\"42}\"}}]},\"finish_reason\":\"tool_calls\"}]}\n\ndata: [DONE]\n\n");
                                }
                            }
                            catch (Exception error) when (mode == "interrupt"
                                && error is IOException or HttpListenerException) { }
                        }
                        else
                        {
                            await Write(incoming.Response, "data: {\"model\":\"" + wireName + "\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Recorded 42.\"},\"finish_reason\":\"stop\"}]}\n\n"
                                + "data: {\"choices\":[],\"usage\":{\"prompt_tokens\":31,\"completion_tokens\":4,\"prompt_tokens_details\":{\"cached_tokens\":7}}}\n\ndata: [DONE]\n\n");
                        }
                        incoming.Response.Close();
                    }
                    finally { incoming.Response.Abort(); }
                }
            });
            turn = agent.SendAsync("Record 42 using the tool.");
            await agent.WaitAsync(observation => observation.Delta?.Kind == ChunkKind.ToolCall, deadline.Token);
            Check(effects == 0 && !turn.IsCompleted
                && !agent.History().Events().Any(fact => fact.Kind is EventKind.ToolCalled or EventKind.ResponseEnded),
                "An incomplete configured stream authorized an effect or canonical response.");
            if (mode == "interrupt")
                await agent.PostAsync(new AgentInput(InputKind.InterruptNow)).WaitAsync(deadline.Token);
            release.TrySetResult();
            if (mode != "success")
            {
                Exception? failure = null;
                try { await turn.WaitAsync(deadline.Token); }
                catch (Exception error) { failure = error; }
                Check(failure is not null && failure is not TimeoutException
                    && !deadline.IsCancellationRequested,
                    "Interrupted or rejected stream must fail before the fixture deadline.");
                if (mode == "interrupt")
                    Check(failure is OperationCanceledException, "Immediate interruption lost cancellation semantics.");
                if (mode == "provider-error")
                {
                    Check(failure is InvalidOperationException
                        && agent.History().Events().Count(fact => fact.Kind == EventKind.ErrorOccurred
                            && fact.Error == "Invalid provider stream: FormatException") == 1,
                        "Provider error did not produce one controlled canonical stream failure.");
                }
                Check(effects == 0 && !agent.History().Events().Any(fact =>
                    fact.Kind is EventKind.ToolCalled or EventKind.ResponseEnded),
                    "An interrupted or failed partial stream became an executable response.");
                turn = agent.SendAsync("Fresh turn after the failed stream.");
            }
            Check(await turn.WaitAsync(deadline.Token) == "Recorded 42.", "Streamed final answer changed.");
            await server.WaitAsync(deadline.Token);
            var responses = agent.History().Events().Where(fact => fact.Kind == EventKind.ResponseEnded).ToArray();
            Check(responses.Length == (mode == "success" ? 2 : 1) && responses.All(fact => fact.From!.Matches(agent.Configuration().Target())),
                "Streamed responses lost qualified offering provenance.");
            Check(effects == (mode == "success" ? 1 : 0) && requests == 2
                && agent.WorkingContextStatus().InputTokens == 31,
                "Configured stream duplicated effects/requests or counted cached input twice.");
        }
        finally
        {
            release.TrySetResult();
            deadline.Cancel();
            listener.Stop();
            try
            {
                try { await Task.WhenAll(new Task?[] { server, turn }.OfType<Task>()); }
                catch (Exception error) when (error is OperationCanceledException or HttpListenerException
                    or ObjectDisposedException) { }
            }
            finally { app.Destroy(); }
        }
    }
}
