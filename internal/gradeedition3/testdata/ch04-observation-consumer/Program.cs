using System.Text.Json;
using Ensemble;

internal static class Program
{
    private static readonly List<string> checks = [];

    private static void Check(bool condition, string name)
    {
        if (!condition) throw new InvalidOperationException(name);
        checks.Add(name);
    }

    private static Event ResultFact(Agent agent, string callID)
    {
        // Early Chapter 4 records ToolReturned. Later direct dispatch separates
        // ToolObserved from model dialogue. Neither requires a second transcript.
        return agent.History().Events().Single(fact => fact.CallID == callID
            && (fact.Kind == EventKind.ToolReturned || fact.Kind.ToString() == "ToolObserved"));
    }

    public static async Task<int> Main(string[] args)
    {
        if (args.Length == 1 && args[0] == "cleanup") return await CleanupChecks.Run();
        if (args.Length == 1 && args[0] == "receiver") return await ReceiverChecks.Run();
        if (args.Length == 1 && args[0] == "log-read") return LogReplay.Read();
        if (args.Length > 1 || (args.Length == 1 && args[0] != "log-write"))
        {
            Console.Error.WriteLine("Expected no arguments, log-write, log-read, receiver, or cleanup.");
            return 2;
        }
        try
        {
            await Run(args.Length == 1);
            Console.WriteLine(JsonSerializer.Serialize(new { passed = true, checks }));
            return 0;
        }
        catch (Exception error)
        {
            Console.WriteLine(JsonSerializer.Serialize(new { passed = false, checks, failure = error.Message }));
            return 1;
        }
    }

    private static async Task Run(bool persist)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(30));
        using var diagnostics = new StringWriter();
        var application = new Ensemble.Ensemble(diagnostics);
        var agent = new Agent(application, new Configuration(Vendor.Anthropic, "local",
            "http://127.0.0.1:1", "synthetic-unused"));
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var entered = new TaskCompletionSource<Job>(TaskCreationOptions.RunContinuationsAsynchronously);
        string nonce = Guid.NewGuid().ToString("N");
        string output = "finished-" + nonce + "\n";
        int executions = 0;
        try
        {
            if (persist) LogReplay.Begin(agent.History(), Check);
            _ = new Tool(agent, ToolKind.WaitForJob);
            _ = new Tool(agent, "held_observation", "External barrier operation",
                JsonSerializer.SerializeToElement(new { type = "object", properties = new { }, additionalProperties = false }),
                async (job, _, cancellation) =>
                {
                    Interlocked.Increment(ref executions);
                    entered.TrySetResult(job);
                    await release.Task.WaitAsync(cancellation);
                    return output;
                });

            Task<ToolResultPart> Call(string id, string name, object args) => agent.DispatchAsync(
                new ToolCallPart(id, agent.Configuration().Target(), name, JsonSerializer.SerializeToElement(args)),
                deadline.Token);

            var first = await Call("first-" + nonce, "held_observation", new { ai_callback_delay = 0 });
            var job = await entered.Task.WaitAsync(deadline.Token);
            Check(!first.IsError && first.Job is { Status: "running" }, "initial-running-observation");
            Check(first.Job!.Handle == job.Handle, "initial-observation-real-handle");
            Check(ReferenceEquals(job.Parent().Parent(), agent)
                && ReferenceEquals(job.Parent().Parent().Parent(), application), "callback-real-parent-chain");
            Check(job.Parent().Children().Length == 1, "one-owned-job");
            var firstFact = ResultFact(agent, first.CallID);
            Check(firstFact.Job == first.Job, "initial-canonical-job-snapshot");
            Check(firstFact.ObservationDuration is { Ticks: > 0 }, "initial-recorded-observation-duration");
            var originalDuration = firstFact.ObservationDuration;
            Check(agent.History().Events().Any(fact => fact.CallID == first.CallID
                && fact.Kind == EventKind.ToolStarted && fact.Seq < firstFact.Seq), "dispatch-before-observation");
            Check(!release.Task.IsCompleted && job.Snapshot().Status == "running", "duration-before-worker-completion");

            var running = await Call("wait-running-" + nonce, "wait_for_job",
                new { handle = job.Handle, ai_callback_delay = 0 });
            Check(!running.IsError && running.Job is { Status: "running" }
                && running.Job.Handle == job.Handle, "control-running-existing-job");
            var runningFact = ResultFact(agent, running.CallID);
            Check(runningFact.Job == running.Job, "control-running-canonical-snapshot");
            Check(runningFact.ObservationDuration is { Ticks: > 0 }, "control-running-recorded-duration");
            Check(agent.History().Events().Single(fact => fact.Kind == EventKind.ToolCalled
                && fact.CallID == running.CallID).Job is null, "control-call-no-wrapper-snapshot");
            Check(job.Parent().Children().Length == 1, "control-no-wrapper-job");

            release.SetResult();
            // Wait on the public terminal snapshot, not a guessed sleep before
            // asserting terminal output. This does not consume the output cursor.
            while (job.Snapshot().Status == "running")
                await Task.Delay(10, deadline.Token);
            var completed = await Call("wait-done-" + nonce, "wait_for_job",
                new { handle = job.Handle, ai_callback_delay = 0 });
            Check(!completed.IsError && completed.Job is { Status: "done" }
                && completed.Job.Handle == job.Handle, "control-terminal-existing-job");
            Check(JsonCodec.Text(completed.Parts).Contains(output, StringComparison.Ordinal), "terminal-real-handler-output");
            var completedFact = ResultFact(agent, completed.CallID);
            Check(completedFact.Job == completed.Job, "control-terminal-canonical-snapshot");
            Check(completedFact.ObservationDuration is { Ticks: > 0 }, "control-terminal-recorded-duration");
            Check(ResultFact(agent, first.CallID).ObservationDuration == originalDuration
                && ResultFact(agent, first.CallID).Job?.Status == "running", "initial-observation-remains-fixed");
            Check(File.ReadAllText(completed.Job!.OutputPath) == output, "owned-output-file-exact");
            Check(executions == 1 && job.Parent().Children().Length == 1, "one-execution-no-control-jobs");

            var invalid = await Call("refused-" + nonce, "held_observation",
                new { unexpected = true, ai_callback_delay = 0 });
            var invalidFact = ResultFact(agent, invalid.CallID);
            Check(invalidFact.ObservationDuration is null, "schema-refusal-duration-unknown");
            Check(!agent.History().Events().Any(fact => fact.CallID == invalid.CallID
                && fact.Kind == EventKind.ToolStarted), "schema-refusal-not-dispatched");
            Check(executions == 1, "schema-refusal-no-handler-effect");
            if (persist)
                LogReplay.Write(agent.History(), [firstFact, runningFact, completedFact, invalidFact], Check);
        }
        finally
        {
            // Release before joined destruction even when an assertion fails.
            // A failing observer must not strand the external test callback.
            release.TrySetResult();
            agent.Destroy();
            application.Destroy();
        }
    }
}
