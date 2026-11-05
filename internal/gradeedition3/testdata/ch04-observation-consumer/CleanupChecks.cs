using System.Text;
using System.Text.Json;
using Ensemble;

internal static class CleanupChecks
{
    internal static async Task<int> Run()
    {
        var checks = new List<string>();
        void Check(bool condition, string name)
        {
            if (!condition) throw new InvalidOperationException(name);
            checks.Add(name);
        }
        try
        {
            using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
            using var emergencyStop = new CancellationTokenSource();
            var writer = new FailingWriter();
            var root = new Ensemble.Ensemble(writer);
            var entered = new TaskCompletionSource<Job>(TaskCreationOptions.RunContinuationsAsynchronously);
            var cancelled = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            var exited = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            Job? job = null;
            Task<Exception?>? destruction = null;
            try
            {
                var configuration = new Configuration(Vendor.Anthropic, "local", "http://127.0.0.1:1", "synthetic-unused");
                var owner = new Agent(root, configuration);
                var peer = new Agent(root, configuration);
                var jobs = owner.Jobs();
                // This is explicit synthetic History setup, not a provider receipt.
                // Unknown usage causes the early History implementation to emit its
                // real diagnostic while appending subsequent lifecycle facts.
                owner.History().Append(new Event(EventKind.ResponseEnded,
                    parts: [new TextPart("synthetic prior response")], usage: new Usage(null, null, null, null)));
                _ = new Tool(owner, "cleanup_barrier", "Cooperative callback with externally held cleanup",
                    JsonSerializer.SerializeToElement(new { type = "object", properties = new { } }),
                    async (actual, _, token) =>
                    {
                        entered.TrySetResult(actual);
                        using var stop = CancellationTokenSource.CreateLinkedTokenSource(token, emergencyStop.Token);
                        try { await Task.Delay(Timeout.Infinite, stop.Token); }
                        finally
                        {
                            if (token.IsCancellationRequested) cancelled.TrySetResult();
                            await release.Task.WaitAsync(deadline.Token);
                            exited.TrySetResult();
                        }
                        return "unreachable";
                    });
                var result = await owner.DispatchAsync(new ToolCallPart("cleanup-" + Guid.NewGuid().ToString("N"),
                    configuration.Target(), "cleanup_barrier",
                    JsonSerializer.SerializeToElement(new { ai_callback_delay = 0 })), deadline.Token);
                job = await entered.Task.WaitAsync(deadline.Token);
                Check(result.Job?.Status == "running", "cleanup-real-running-observation");
                Check(writer.Writes > 0, "cleanup-real-diagnostic-path");
                writer.Fail = true;
                destruction = Task.Run<Exception?>(() =>
                {
                    try { owner.Destroy(); return null; }
                    catch (Exception error) { return error; }
                });
                await Task.WhenAny(cancelled.Task, destruction).WaitAsync(deadline.Token);
                Check(cancelled.Task.IsCompleted, "cleanup-cancellation-despite-diagnostic-failure");
                Check(!destruction.IsCompleted && !exited.Task.IsCompleted, "cleanup-destruction-joins-held-callback");
                Check(writer.Failures > 0, "cleanup-diagnostic-failure-actually-exercised");
                release.TrySetResult();
                Exception? error = await destruction.WaitAsync(deadline.Token);
                Check(exited.Task.IsCompleted, "cleanup-callback-exited-before-destruction");
                Check(error is not null, "cleanup-recording-failure-remains-visible");
                Check(jobs.Children().Length == 0, "cleanup-owned-jobs-detached-after-failure");
                Check(!root.Agents().Contains(owner) && root.Agents().Contains(peer), "cleanup-owner-detached-peer-preserved");
                Check(!writer.Disposed, "cleanup-borrowed-writer-not-disposed");
                writer.Fail = false;
            }
            finally
            {
                // Release even a rejected candidate. Destruction errors are evidence,
                // not a reason for this grader to leave its own barrier blocked.
                writer.Fail = false;
                release.TrySetResult();
                emergencyStop.Cancel();
                if (job is not null) await exited.Task.WaitAsync(deadline.Token);
                if (destruction is not null) await destruction.WaitAsync(deadline.Token);
                if (job is not null) job.Destroy();
                root.Destroy();
            }
            Console.WriteLine(JsonSerializer.Serialize(new { passed = true, checks }));
            return 0;
        }
        catch (Exception error)
        {
            Console.WriteLine(JsonSerializer.Serialize(new { passed = false, checks, failure = error.Message }));
            return 1;
        }
    }

    private sealed class FailingWriter : TextWriter
    {
        public override Encoding Encoding => Encoding.UTF8;
        internal bool Fail;
        internal int Writes;
        internal int Failures;
        internal bool Disposed;
        public override void WriteLine(string? value)
        {
            if (Fail)
            {
                Interlocked.Increment(ref Failures);
                throw new IOException("synthetic borrowed diagnostic failure");
            }
            Interlocked.Increment(ref Writes);
        }
        protected override void Dispose(bool disposing) { Disposed = true; base.Dispose(disposing); }
    }
}
