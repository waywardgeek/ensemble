using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class Probe
{
    private static async Task LifecycleChecks()
    {
        var reservation = new TcpListener(IPAddress.Loopback, 0);
        reservation.Start();
        int port = ((IPEndPoint)reservation.LocalEndpoint).Port;
        reservation.Stop();
        using var listener = new HttpListener();
        string address = $"http://127.0.0.1:{port}/";
        listener.Prefixes.Add(address);
        listener.Start();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        using var emergencyStop = new CancellationTokenSource();
        var started = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var cancelled = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var releaseCleanup = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var exited = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var callbackWork = new TaskCompletionSource<Task<string>>(TaskCreationOptions.RunContinuationsAsynchronously);
        var app = new Application(TextWriter.Null);
        Task? server = null;
        Task<string>? turn = null;
        Task? destruction = null;
        string callId = "held-" + Guid.NewGuid().ToString("N");
        string wire = "local-" + Guid.NewGuid().ToString("N");
        int effects = 0;
        try
        {
            var none = new ModelAuthentication(app, "none");
            var vendor = new ModelVendor(app, "local", "Local");
            var endpoint = new ModelEndpoint(vendor, "chat", address, Surface.ChatCompletions, none);
            var model = new Model(endpoint, "small", wire, "Small", new ModelFeatures(Media.None), 4096);
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "legacy", address, "synthetic"));
            var sibling = new Agent(app, new Configuration(Vendor.OpenAI, "sibling", address, "synthetic"));
            await agent.SetModelAsync(model);
            using var schema = JsonDocument.Parse("{\"type\":\"object\",\"properties\":{}}");
            _ = new Tool(agent, "held", "Wait for cancellation", schema.RootElement,
                (job, arguments, token) =>
                {
                    async Task<string> RunCallback()
                    {
                        Interlocked.Increment(ref effects);
                        started.TrySetResult();
                        using var stop = CancellationTokenSource.CreateLinkedTokenSource(token, emergencyStop.Token);
                        try { await Task.Delay(Timeout.Infinite, stop.Token); }
                        finally
                        {
                            if (token.IsCancellationRequested) cancelled.TrySetResult();
                            // The fixture releases this after observing destruction pending.
                            await releaseCleanup.Task;
                            exited.TrySetResult();
                        }
                        return "unreachable";
                    }
                    Task<string> work = RunCallback();
                    callbackWork.TrySetResult(work);
                    return work;
                });
            server = Task.Run(async () =>
            {
                var incoming = await listener.GetContextAsync().WaitAsync(deadline.Token);
                try
                {
                    using var reader = new StreamReader(incoming.Request.InputStream);
                    using var request = JsonDocument.Parse(await reader.ReadToEndAsync(deadline.Token));
                    Require(incoming.Request.Headers["Authorization"] is null
                        && request.RootElement.GetProperty("model").GetString() == wire,
                        "lifecycle-real-configured-request");
                    byte[] reply = Encoding.UTF8.GetBytes(JsonSerializer.Serialize(new
                    {
                        model = wire,
                        choices = new[] { new { message = new { content = "", tool_calls = new[] {
                            new { id = callId, type = "function", function = new { name = "held", arguments = "{}" } }
                        } } } }
                    }));
                    incoming.Response.ContentType = "application/json";
                    incoming.Response.ContentLength64 = reply.Length;
                    await incoming.Response.OutputStream.WriteAsync(reply, deadline.Token);
                }
                finally { incoming.Response.Close(); }
            });
            turn = agent.SendAsync("Use the held tool.", deadline.Token);
            await started.Task.WaitAsync(deadline.Token);
            await server.WaitAsync(deadline.Token);
            Require(effects == 1 && agent.History().Events().Count(fact => fact.Kind == EventKind.ToolStarted) == 1,
                "lifecycle-real-tool-started");
            destruction = Task.Run(agent.Destroy);
            await cancelled.Task.WaitAsync(deadline.Token);
            Require(!exited.Task.IsCompleted, "lifecycle-cancellation-reaches-executing-tool");
            Task observed = await Task.WhenAny(destruction, Task.Delay(150, deadline.Token));
            Require(!ReferenceEquals(observed, destruction), "lifecycle-destruction-joins-held-cleanup");
            releaseCleanup.TrySetResult();
            await destruction.WaitAsync(deadline.Token);
            Require(exited.Task.IsCompleted && effects == 1, "lifecycle-worker-exited-before-destruction-return");
            Require(!app.Agents().Contains(agent) && app.Agents().Contains(sibling)
                && model.Agents().Length == 0, "lifecycle-detaches-without-destroying-sibling");
        }
        finally
        {
            // Emergency release cannot satisfy the earlier pending-destruction check.
            releaseCleanup.TrySetResult();
            emergencyStop.Cancel();
            deadline.Cancel();
            listener.Stop();
            try
            {
                try { await Task.WhenAll(new Task?[] { server, turn, destruction }.OfType<Task>()); }
                catch (Exception error) when (error is OperationCanceledException or HttpListenerException or ObjectDisposedException) { }
            }
            finally
            {
                if (started.Task.IsCompleted)
                {
                    Task<string> callback = await callbackWork.Task;
                    try { await callback; }
                    catch (OperationCanceledException) { }
                }
                app.Destroy();
            }
        }
    }
}
