using System.Diagnostics;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    public static async Task GoalsIdleBackoff(string temp)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = GoalAgent(app, GoalSkills(temp), listener.Prefixes.Single());
            await LoadGoals(agent);
            string file = Path.Combine(temp, "goal-idle-backoff.md");
            File.WriteAllText(file, "Inspect retained results before continuing. No blind effect replay.");
            await agent.PostAsync(new AgentInput(InputKind.Pause));
            // The legacy interval must not delay the new idle-transition policy.
            await agent.GoalControlAsync(new GoalUpdate(file, true, 1000));
            var incoming = listener.GetContextAsync();
            await Task.Delay(100);
            Check(!incoming.IsCompleted, "Paused goal started a request");
            await agent.PostAsync(new AgentInput(InputKind.Resume));
            var request = await incoming.WaitAsync(TimeSpan.FromSeconds(3));
            Check((await ReadRecallRequest(request)).ToJsonString().Contains("Automatic goal reminder"),
                "Idle wake was not a labelled runtime continuation");

            async Task<GoalStatus> RetryStatus(int count)
            {
                using var timeout = new CancellationTokenSource(TimeSpan.FromSeconds(5));
                while (true)
                {
                    var status = await agent.GoalStatusAsync(timeout.Token);
                    if (status.Phase == "retry_wait" && status.FailureCount == count) return status;
                    await Task.Delay(10, timeout.Token);
                }
            }

            for (int count = 1; count <= 5; count++)
            {
                incoming = listener.GetContextAsync();
                var elapsed = Stopwatch.StartNew();
                await Reply(request, "{\"error\":{\"message\":\"owned retry fixture\"}}", 503);
                var status = await RetryStatus(count);
                int seconds = count == 1 ? 1 : count == 2 ? 10 : count == 3 ? 100 : 1000;
                double remaining = (status.DueUtc!.Value - DateTimeOffset.UtcNow).TotalSeconds;
                Check(remaining > seconds - 1 && remaining <= seconds,
                    "Wrong failure-backoff stage: " + count);
                Check(status.RetryError.Length > 0 && status.LastError.Length == 0,
                    "Transient request failure became a goal-file error");
                Check((await agent.GoalStatusAsync()).DueUtc == status.DueUtc,
                    "Status read reset retry deadline");
                if (count > 1)
                {
                    Check(!incoming.IsCompleted, "Backoff did not hold the next request");
                    await agent.GoalControlAsync(new GoalUpdate(RetryNow: true));
                }
                request = await incoming.WaitAsync(TimeSpan.FromSeconds(4));
                if (count == 1) Check(elapsed.Elapsed >= TimeSpan.FromMilliseconds(850),
                    "First failure retried without one-second backoff");
                Check((await agent.GoalStatusAsync()).FailureCount == count,
                    "Retry-now erased the failure streak");
            }

            // A successful turn resets the streak and immediately continues.
            incoming = listener.GetContextAsync();
            await Reply(request, "{\"choices\":[{\"message\":{\"role\":\"assistant\",\"content\":\"checked\"},\"finish_reason\":\"stop\"}]}");
            request = await incoming.WaitAsync(TimeSpan.FromSeconds(3));
            Check((await agent.GoalStatusAsync()).FailureCount == 0, "Success did not reset backoff");
            await agent.GoalControlAsync(new GoalUpdate(Enabled: false));
            incoming = listener.GetContextAsync();
            await Reply(request, "{\"choices\":[{\"message\":{\"role\":\"assistant\",\"content\":\"finished\"},\"finish_reason\":\"stop\"}]}");
            await agent.FlushAsync().WaitAsync(TimeSpan.FromSeconds(5));
            await Task.Delay(100);
            Check(!incoming.IsCompleted && !(await agent.GoalStatusAsync()).Enabled,
                "Off allowed another automatic continuation");
            await agent.ShutdownAsync();
        }
        finally
        {
            // Closing the owned listener also releases an unanswered fake request
            // if an assertion fails. Agent destruction joins its transport.
            listener.Stop();
            app.Destroy();
        }
    }
}
