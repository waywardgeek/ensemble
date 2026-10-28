using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    // The fake holds HTTP, not the actor: changing policy while that response is
    // held must affect only the next turn. A real registered tool counts effects.
    public static async Task RuntimeBudget()
    {
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var entered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var server = Task.Run(async () =>
        {
            for (int request = 0; request < 5; request++)
            {
                using var client = await listener.AcceptTcpClientAsync(deadline.Token);
                using var stream = client.GetStream();
                string body = await ReadRequest(stream, deadline.Token);
                using var json = JsonDocument.Parse(body);
                string system = json.RootElement.GetProperty("system").GetString()!;
                Check(system.Contains("Tool batch limit reached") == (request is 1 or 4),
                    "Wrong request-local budget instruction.");
                if (request == 0)
                {
                    entered.SetResult();
                    await release.Task.WaitAsync(deadline.Token);
                }
                if (request == 4)
                {
                    await Reply(stream, "finished with two batches", deadline.Token);
                }
                else
                {
                    await ReplyJson(stream, new
                    {
                        model = "model-a",
                        content = new[] { new { type = "tool_use", id = "counter", name = "effect", input = new { } } },
                        usage = new { input_tokens = 1, output_tokens = 1 }
                    }, deadline.Token);
                }
            }
        });
        var app = new Application(TextWriter.Null);
        try
        {
            int port = ((IPEndPoint)listener.LocalEndpoint).Port;
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "model-a", "http://127.0.0.1:" + port));
            int effects = 0;
            _ = new Tool(agent, "effect", "Count real executions",
                JsonSerializer.SerializeToElement(new { type = "object" }), (job, args, token) =>
                {
                    Interlocked.Increment(ref effects);
                    return Task.FromResult("effect completed");
                });
            agent.SetMaxToolRounds(1);
            var first = agent.SendAsync("one batch", deadline.Token);
            await entered.Task.WaitAsync(deadline.Token);
            agent.SetMaxToolRounds(2);
            release.SetResult();
            try
            {
                await first;
                throw new Exception("Expected the captured limit to refuse the second batch.");
            }
            catch (InvalidOperationException error) when (error.Message.Contains("batch limit")) { }
            Check(effects == 1, "An in-flight turn adopted a later settings change.");
            string answer = await agent.SendAsync("two batches", deadline.Token);
            await server;
            Check(effects == 3 && answer == "finished with two batches", "Next turn did not capture the new policy.");
            Check(agent.History().Events().All(e => !JsonCodec.WriteEvent(e).Contains("Do not call more tools")),
                "Request-only instruction leaked into History.");
            agent.SetMaxToolRounds(0);
            Check(agent.MaxToolRounds() == 0, "Auto was normalized in the public snapshot.");
            try
            {
                agent.SetMaxToolRounds(-1);
                throw new Exception("Negative runtime update accepted.");
            }
            catch (ArgumentOutOfRangeException) { }
            Check(agent.MaxToolRounds() == 0, "Refused update changed the policy.");

            foreach (var vendor in Enum.GetValues<Vendor>())
            {
                var projectionAgent = new Agent(app, new Configuration(vendor, "local", "http://127.0.0.1:1"));
                projectionAgent.History().Append(new Event(EventKind.MessageReceived,
                    actor: Actor.Human, parts: [new TextPart("hello")]));
                string before = projectionAgent.Engine().Render(projectionAgent.History().View()).ToJsonString();
                string withNote = projectionAgent.Engine().Render(projectionAgent.History().View(), "BUDGET_NOTE").ToJsonString();
                Check(withNote.Contains("BUDGET_NOTE"), "Renderer lost the request instruction.");
                Check(projectionAgent.Engine().Render(projectionAgent.History().View()).ToJsonString() == before,
                    "Rendering retained request-local instructions.");
            }
        }
        finally
        {
            release.TrySetResult();
            deadline.Cancel();
            listener.Stop();
            app.Destroy();
        }
    }
}
