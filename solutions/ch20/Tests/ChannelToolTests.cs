using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class ChannelToolTests
{
    // A latch is the working external boundary. The actual Tool, actor, Job,
    // output file and observation cursor are not replaced by test doubles.
    internal static async Task Run()
    {
        var directory = Path.Combine(Path.GetTempPath(), "atomic-tool-" + Guid.NewGuid().ToString("N"));
        var app = new Application(TextWriter.Null);
        var agent = new Agent(app, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1"), directory);
        var entered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var stopped = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var schema = JsonSerializer.SerializeToElement(new { type = "object", properties = new { } });
        string fullValue = new string('x', 20000);
        _ = new Tool(agent, "atomic", "Held channel", schema, async (job, args, cancellation) =>
        {
            entered.TrySetResult();
            try
            {
                await release.Task.WaitAsync(cancellation);
                return fullValue;
            }
            finally
            {
                stopped.TrySetResult();
            }
        }, waitForCompletion: true);
        ToolCallPart Call(string id) => new(id, agent.Configuration().Target(), "atomic",
            JsonSerializer.SerializeToElement(new { ai_callback_delay = 0, max_output_bytes = 3 }));
        try
        {
            var invocation = agent.DispatchAsync(Call("first"));
            await entered.Task.WaitAsync(TimeSpan.FromSeconds(5));
            await Task.Delay(75);
            if (invocation.IsCompleted) throw new Exception("Atomic tool returned a timed running report");
            release.SetResult();
            var result = await invocation.WaitAsync(TimeSpan.FromSeconds(5));
            if (result.IsError || result.Job?.Status != "done" || result.Parts.OfType<TextPart>().Single().Text != fullValue)
            {
                throw new Exception("Atomic completion was truncated or not terminal");
            }
            var again = await agent.Jobs().Find(result.Job.Handle).ObserveAsync("again", 0);
            if (again.Parts.OfType<TextPart>().Any(part => part.Text.Contains(fullValue)))
            {
                throw new Exception("Atomic completion replayed through the observation cursor");
            }

            entered = new(TaskCreationOptions.RunContinuationsAsynchronously);
            release = new(TaskCreationOptions.RunContinuationsAsynchronously);
            stopped = new(TaskCreationOptions.RunContinuationsAsynchronously);
            var pending = agent.DispatchAsync(Call("second"));
            await entered.Task.WaitAsync(TimeSpan.FromSeconds(5));
            await agent.ShutdownAsync().WaitAsync(TimeSpan.FromSeconds(5));
            await stopped.Task.WaitAsync(TimeSpan.FromSeconds(5));
            await pending.WaitAsync(TimeSpan.FromSeconds(5));
        }
        finally
        {
            app.Destroy();
            Directory.Delete(directory, true);
        }
    }
}
