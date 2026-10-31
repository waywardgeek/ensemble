using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    public static async Task StreamingDelivery()
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        const string key = "dummy-trace-secret";
        string first = "data: {\"type\":\"message_start\",\"message\":{\"model\":\"claude-3-haiku-20240307\",\"usage\":{\"input_tokens\":3}}}\n\n"
            + "data: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n"
            + "data: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"hello \"}}\n\n";
        string last = "data: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"" + key + "\"}}\n\n"
            + "data: {\"type\":\"content_block_stop\",\"index\":0}\n\n"
            + "data: {\"type\":\"message_delta\",\"usage\":{\"output_tokens\":2}}\n\n"
            + "data: {\"type\":\"message_stop\"}\n\n";
        var server = Task.Run(async () =>
        {
            using var client = await listener.AcceptTcpClientAsync(deadline.Token);
            using var stream = client.GetStream();
            using var request = JsonDocument.Parse(await ReadRequest(stream, deadline.Token));
            Check(request.RootElement.GetProperty("stream").GetBoolean(), "Known model did not request SSE.");
            int size = Encoding.UTF8.GetByteCount(first + last);
            string header = "HTTP/1.1 200 OK\r\nContent-Type: text/event-stream\r\nContent-Length: "
                + size + "\r\nConnection: close\r\n\r\n";
            await stream.WriteAsync(Encoding.UTF8.GetBytes(header + first), deadline.Token);
            await stream.FlushAsync(deadline.Token);
            await release.Task.WaitAsync(deadline.Token);
            await stream.WriteAsync(Encoding.UTF8.GetBytes(last), deadline.Token);
        });
        var trace = new StringWriter();
        var app = new Application(TextWriter.Null, trace);
        try
        {
            int port = ((IPEndPoint)listener.LocalEndpoint).Port;
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "claude-3-haiku-20240307",
                "http://127.0.0.1:" + port, key));
            var reply = agent.SendAsync("show " + key);
            var delta = await agent.WaitAsync(o => o.Delta?.Text == "hello ", deadline.Token);
            Check(delta.Kind == ObservationKind.PartDelta && delta.PartID == 0 && delta.ResponseID > 0,
                "Delta identity missing or zero part ID refused.");
            Check(!reply.IsCompleted, "Reply completed before held terminal frames.");
            Check(!agent.History().Events().Any(e => e.Kind == EventKind.ResponseEnded),
                "Partial stream entered canonical History.");
            release.TrySetResult();
            var final = await agent.WaitAsync(o => o.Kind == ObservationKind.PartFinal, deadline.Token);
            Check(final.PartID == delta.PartID && final.ResponseID == delta.ResponseID, "Final identity changed.");
            Check(agent.History().Events().Any(e => e.Seq == final.Seq && e.Kind == EventKind.ResponseEnded),
                "Final published before recording.");
            Check(await reply.WaitAsync(deadline.Token) == "hello " + key, "Canonical provider text was redacted or lost.");
            await server;
            Check(!trace.ToString().Contains(key), "Configured credential leaked to trace.");
            Check(trace.ToString().Contains("[REDACTED]") && trace.ToString().Contains("message_stop"),
                "Trace omitted sanitized request or received frames.");
            Check(agent.History().Events().Count(e => e.Kind == EventKind.ResponseEnded) == 1,
                "Deltas duplicated canonical response.");

            var disabled = new Agent(app, new Configuration(Vendor.Anthropic, "claude-3-haiku-20240307",
                "http://127.0.0.1", disableStreaming: true));
            Check(!disabled.Engine().Render(disabled.History().View()).ContainsKey("stream"),
                "Explicit no-stream override ignored.");
            var unknown = new Agent(app, new Configuration(Vendor.OpenAI, "unknown", "http://127.0.0.1"));
            Check(!unknown.Engine().Render(unknown.History().View()).ContainsKey("stream"),
                "Unknown model assumed streaming support.");
        }
        finally
        {
            release.TrySetResult();
            app.Destroy();
            listener.Stop();
        }
    }
    public static async Task StreamingToolBoundary()
    {
        // Each case uses actual HTTP, actor admission and a real custom Tool.
        // Policy refusal and both interrupts must leave its effect count zero.
        foreach (string mode in new[] { "policy", "now", "after_tool" })
        {
            using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
            var listener = new TcpListener(IPAddress.Loopback, 0);
            listener.Start();
            var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            string first = "data: {\"choices\":[{\"index\":0,\"delta\":{\"tool_calls\":[{\"index\":0,\"id\":\"call\",\"function\":{\"name\":\"effect\",\"arguments\":\"{\"}}]}}]}\n\n";
            string last = "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"late explanation\",\"tool_calls\":[{\"index\":0,\"function\":{\"arguments\":\"}\"}}]}}]}\n\n"
                + "data: {\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"tool_calls\"}]}\n\ndata: [DONE]\n\n";
            var server = Task.Run(async () =>
            {
                using (var client = await listener.AcceptTcpClientAsync(deadline.Token))
                using (var stream = client.GetStream())
                {
                    _ = await ReadRequest(stream, deadline.Token);
                    string header = "HTTP/1.1 200 OK\r\nContent-Type: text/event-stream\r\nContent-Length: "
                        + Encoding.UTF8.GetByteCount(first + last) + "\r\nConnection: close\r\n\r\n";
                    await stream.WriteAsync(Encoding.UTF8.GetBytes(header + first), deadline.Token);
                    await stream.FlushAsync(deadline.Token);
                    await release.Task.WaitAsync(deadline.Token);
                    try
                    {
                        await stream.WriteAsync(Encoding.UTF8.GetBytes(last), deadline.Token);
                    }
                    catch (System.IO.IOException) when (mode == "now") { }
                }
                using var next = await listener.AcceptTcpClientAsync(deadline.Token);
                using var nextStream = next.GetStream();
                string request = await ReadRequest(nextStream, deadline.Token);
                if (mode == "policy")
                {
                    Check(request.Contains("first explain"), "Refused tool result missing from next request.");
                }
                string body = "{\"choices\":[{\"message\":{\"content\":\"safe answer\"}}],\"usage\":{\"prompt_tokens\":1,\"completion_tokens\":1}}";
                string reply = "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: "
                    + Encoding.UTF8.GetByteCount(body) + "\r\nConnection: close\r\n\r\n" + body;
                await nextStream.WriteAsync(Encoding.UTF8.GetBytes(reply), deadline.Token);
            });
            var app = new Application(TextWriter.Null);
            int effects = 0;
            try
            {
                int port = ((IPEndPoint)listener.LocalEndpoint).Port;
                var agent = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4o",
                    "http://127.0.0.1:" + port, requireToolExplanation: mode == "policy"));
                using var schema = JsonDocument.Parse("{\"type\":\"object\"}");
                _ = new Tool(agent, "effect", "Record one effect", schema.RootElement,
                    (job, args, token) =>
                    {
                        Interlocked.Increment(ref effects);
                        return Task.FromResult("effect");
                    });
                var turn = agent.SendAsync("start");
                var delta = await agent.WaitAsync(o => o.Delta?.Kind == ChunkKind.ToolCall, deadline.Token);
                Check(effects == 0 && !agent.History().Events().Any(e => e.Kind == EventKind.ToolCalled),
                    "Partial tool call dispatched.");
                if (mode != "policy")
                {
                    var interrupt = mode == "now" ? InputKind.InterruptNow : InputKind.InterruptAfterTool;
                    await agent.PostAsync(new AgentInput(interrupt)).WaitAsync(deadline.Token);
                }
                release.TrySetResult();
                if (mode == "policy")
                {
                    Check(await turn.WaitAsync(deadline.Token) == "safe answer", "Policy recovery failed.");
                }
                else
                {
                    try
                    {
                        await turn.WaitAsync(deadline.Token);
                        throw new InvalidOperationException("Interrupted turn reported success.");
                    }
                    catch (OperationCanceledException) { }
                    Check(await agent.SendAsync("fresh").WaitAsync(deadline.Token) == "safe answer",
                        "Interrupted streaming actor could not be reused.");
                }
                await server;
                Check(effects == 0, "Refused or interrupted stream dispatched a tool.");
                if (mode == "after_tool")
                {
                    Check(agent.History().Events().Any(e => e.Kind == EventKind.ToolReturned && e.IsError),
                        "Retained after-tool boundary has no paired skipped result.");
                }
            }
            finally
            {
                release.TrySetResult();
                app.Destroy();
                listener.Stop();
            }
        }
    }

}
