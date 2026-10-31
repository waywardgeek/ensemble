using System.Net;
using System.Net.Sockets;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    public static async Task InputAdmission()
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(10));
        var app = new Application(TextWriter.Null);
        var agent = new Agent(app, new Configuration(Vendor.Anthropic, "local", "http://127.0.0.1:1"));
        try
        {
            await agent.PostAsync(new AgentInput(InputKind.Pause));
            await Outcome(true);
            await Refused(() => agent.PostAsync(new AgentInput((InputKind)999)), false);
            await Refused(() => agent.PostAsync(null!), false);
            await Refused(() => agent.PostAsync(new AgentInput(InputKind.Hint, null!)), false);
            await Refused(() => agent.SubmitAsync(null!), false);
            await Refused(() => agent.SubmitAsync(new Part[] { null! }), false);
            await Refused(() => agent.SubmitAsync([new TextPart(null!)]), false);
            await Refused(() => agent.DispatchAsync(null!), false);
            Check(agent.History().Events().Length == 0, "Rejected input changed History.");

            // Admission and completion are different facts. A paused request is
            // accepted now, then shutdown settles its reliable reply as refused work.
            var pending = agent.SubmitAsync([new TextPart("queued")]);
            await Outcome(true);
            Check(!pending.IsCompleted, "Pause failed to hold the request.");
            await agent.ShutdownAsync().WaitAsync(deadline.Token);
            try
            {
                await pending;
                throw new Exception("Shutdown reported queued work as successful.");
            }
            catch (ObjectDisposedException) { }
            await Refused(() => agent.PostAsync(new AgentInput(InputKind.Resume)), true);
            await Refused(() => agent.SubmitAsync([new TextPart("too late")]), true);

            using var noMore = new CancellationTokenSource(TimeSpan.FromMilliseconds(50));
            try
            {
                await agent.WaitAsync(o => o.Kind == ObservationKind.Input, noMore.Token);
                throw new Exception("Duplicate input outcome.");
            }
            catch (OperationCanceledException) when (noMore.IsCancellationRequested) { }
        }
        finally
        {
            app.Destroy();
        }

        async Task Outcome(bool accepted)
        {
            var local = await agent.WaitAsync(o => o.Kind == ObservationKind.Input, deadline.Token);
            var merged = await app.WaitAnyAsync(o => o.Observation.Kind == ObservationKind.Input, deadline.Token);
            Check(local.Accepted == accepted, "Incorrect admission outcome.");
            Check(ReferenceEquals(merged.Agent, agent) && ReferenceEquals(merged.Observation, local),
                "Root observation lost source identity or copied its payload.");
        }

        async Task Refused(Func<Task> action, bool disposed)
        {
            try
            {
                await action().WaitAsync(deadline.Token);
                throw new Exception("Expected input refusal.");
            }
            catch (ArgumentException) when (!disposed) { }
            catch (ObjectDisposedException) when (disposed) { }
            await Outcome(false);
        }
    }

    public static async Task UnlimitedBatches()
    {
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var server = Task.Run(async () =>
        {
            for (int index = 0; index < 5; index++)
            {
                using var client = await listener.AcceptTcpClientAsync(deadline.Token);
                using var stream = client.GetStream();
                await ReadRequest(stream, deadline.Token);
                if (index == 4)
                {
                    await Reply(stream, "finished by model", deadline.Token);
                }
                else
                {
                    await ReplyJson(stream, new
                    {
                        model = "model-a",
                        content = new[] { new { type = "tool_use", id = "repeat", name = "absent", input = new { } } },
                        usage = new { input_tokens = 1, output_tokens = 1 }
                    }, deadline.Token);
                }
            }
        });
        var app = new Application(TextWriter.Null);
        try
        {
            int port = ((IPEndPoint)listener.LocalEndpoint).Port;
            var configuration = new Configuration(Vendor.Anthropic, "model-a",
                "http://127.0.0.1:" + port, "dummy", toolRoundLimit: -1);
            var agent = new Agent(app, configuration);
            string answer = await agent.SendAsync("continue", deadline.Token);
            await server;
            Check(answer == "finished by model", "Negative limit refused work.");
            Check(agent.History().Events().Count(e => e.Kind == EventKind.ToolReturned) == 4,
                "Missing results from unlimited batches.");
            Check(new Configuration(Vendor.Anthropic, "local", "http://127.0.0.1", toolRoundLimit: 0)
                .ToolRoundLimit == 200, "Zero no longer selects the default.");
        }
        finally
        {
            app.Destroy();
            listener.Stop();
            deadline.Cancel();
        }
    }
}
