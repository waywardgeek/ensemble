using System.Net;
using System.Net.Sockets;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    public static async Task FlushTurns()
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        var entered = new TaskCompletionSource();
        var release = new TaskCompletionSource();
        var server = Task.Run(async () =>
        {
            for (int i = 0; i < 2; i++)
            {
                using var client = await listener.AcceptTcpClientAsync(deadline.Token);
                using var stream = client.GetStream();
                await ReadRequest(stream, deadline.Token);
                if (i == 0)
                {
                    entered.SetResult();
                    await release.Task.WaitAsync(deadline.Token);
                }
                await Reply(stream, "drained", deadline.Token);
            }
        });
        var app = new Application(TextWriter.Null);
        try
        {
            int port = ((IPEndPoint)listener.LocalEndpoint).Port;
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "model-a",
                "http://127.0.0.1:" + port, "dummy"));
            await agent.PostAsync(new AgentInput(InputKind.Pause));
            var answer = agent.SubmitAsync([new TextPart("queued")]);
            var flushed = agent.FlushAsync(deadline.Token);
            // This acknowledgment proves the earlier flush entered the mailbox.
            await agent.PostAsync(new AgentInput(InputKind.Pause));
            Check(!flushed.IsCompleted, "Flush ignored paused queued work.");
            await agent.PostAsync(new AgentInput(InputKind.Resume));
            await entered.Task.WaitAsync(deadline.Token);
            Check(!flushed.IsCompleted, "Flush ignored held HTTP.");
            // Neither progress stream has a reader. Exceed both bounded queues
            // while HTTP is held; reliable replies must remain independent.
            for (int i = 0; i < 400; i++)
            {
                await agent.PostAsync(new AgentInput(InputKind.Ephemeral, "status " + i));
            }
            using var detached = new CancellationTokenSource();
            var cancelledWait = agent.FlushAsync(detached.Token);
            detached.Cancel();
            try
            {
                await cancelledWait;
                throw new Exception("Cancelled flush waiter returned success.");
            }
            catch (OperationCanceledException) { }
            Check(!answer.IsCompleted, "Cancelling flush cancelled the request.");
            release.SetResult();
            await flushed;
            Check(await answer == "drained", "Flush lost the reliable reply.");

            // EOF after a live prompt must wait for that prompt too. There is
            // deliberately no Submit reply available to the caller here.
            await agent.PostAsync(new AgentInput(InputKind.Message, "live"));
            await agent.FlushAsync(deadline.Token);
            await server;
            Check(agent.History().Events().Count(e => e.Kind == EventKind.ResponseEnded) == 2,
                "Flush returned before the live response was recorded.");
            await agent.PostAsync(new AgentInput(InputKind.Pause));
            var discarded = agent.SubmitAsync([new TextPart("never run")]);
            var stoppedFlush = agent.FlushAsync();
            await agent.ShutdownAsync();
            await Refused(discarded);
            await Refused(stoppedFlush);
        }
        finally
        {
            release.TrySetResult();
            app.Destroy();
            listener.Stop();
            deadline.Cancel();
        }

        static async Task Refused(Task task)
        {
            try
            {
                await task;
                throw new Exception("Shutdown credited discarded work.");
            }
            catch (ObjectDisposedException) { }
        }
    }
}
