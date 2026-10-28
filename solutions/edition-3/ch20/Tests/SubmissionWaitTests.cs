using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class SubmissionWaitTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition) throw new Exception(message);
    }

    public static async Task Observers(string temp)
    {
        string work = Path.Combine(temp, "submission-waiters");
        Directory.CreateDirectory(work);
        var app = new Application(TextWriter.Null);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        try
        {
            var config = new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1", "local-only", sandboxRoot: work);
            var parent = new Agent(app, config);
            parent.StartPersistence(new PersistenceOptions(Path.Combine(work, "save.json")));
            Agent child = await parent.CreateChildAsync("worker", cancellation: deadline.Token);
            using var cancelled = new CancellationTokenSource();
            Task<JsonElement> abandoned = child.WaitForSubmissionAsync(cancelled.Token);
            Task<JsonElement> first = child.WaitForSubmissionAsync(deadline.Token);
            Task<JsonElement> second = child.WaitForSubmissionAsync(deadline.Token);
            Check(!first.IsCompleted && !second.IsCompleted, "Wait completed without a submission.");
            cancelled.Cancel();
            try { await abandoned; throw new Exception("Cancelled observer succeeded."); }
            catch (OperationCanceledException) when (cancelled.IsCancellationRequested) { }
            Check(!first.IsCompleted && parent.FindChild("worker") == child,
                "Observer cancellation cancelled shared work or another observer.");

            // An unsuccessful disk replacement cannot publish or satisfy either
            // observer. Directory-at-target gives a real rooted rename refusal.
            string target = Path.Combine(child.DataDirectory(), "submit.md");
            Directory.CreateDirectory(target);
            bool refused = false;
            try { child.PublishSubmission("not durable"); }
            catch (IOException) { refused = true; }
            Check(refused && child.Submission() is null && !first.IsCompleted && !second.IsCompleted,
                "Failed persistence published completion.");
            Directory.Delete(target);
            child.PublishSubmission("durable");
            JsonElement[] results = await Task.WhenAll(first, second).WaitAsync(deadline.Token);
            Check(results.All(value => value.GetProperty("result").GetString() == "durable"),
                "Broadcast observers did not receive the same durable payload.");
            Check((await child.WaitForSubmissionAsync(deadline.Token)).GetProperty("result").GetString() == "durable",
                "Late observer missed published completion.");
            await child.ShutdownAsync().WaitAsync(deadline.Token);
            Check((await child.WaitForSubmissionAsync(deadline.Token)).GetProperty("result").GetString() == "durable",
                "Shutdown erased an already published submission.");

            Agent stopped = await parent.CreateChildAsync("stopped", cancellation: deadline.Token);
            Task<JsonElement> waiting = stopped.WaitForSubmissionAsync(deadline.Token);
            await stopped.ShutdownAsync().WaitAsync(deadline.Token);
            try { await waiting; throw new Exception("Stopped unsubmitted child reported success."); }
            catch (ObjectDisposedException) { }
            try { await parent.WaitForSubmissionAsync(deadline.Token); throw new Exception("Root accepted a parent-submission wait."); }
            catch (InvalidOperationException) { }
        }
        finally { app.Destroy(); }
    }

    public static async Task BottomUp(string temp)
    {
        string work = Path.Combine(temp, "submission-bottom-up");
        Directory.CreateDirectory(work);
        var app = new Application(TextWriter.Null);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        try
        {
            var config = new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1", "local-only", sandboxRoot: work);
            var root = new Agent(app, config);
            root.StartPersistence(new PersistenceOptions(Path.Combine(work, "save.json")));
            Agent parent = await root.CreateChildAsync("parent", cancellation: deadline.Token);
            Agent leaf = await parent.CreateChildAsync("leaf", cancellation: deadline.Token);
            Task<JsonElement> result = parent.WaitForSubmissionAsync(deadline.Token);
            Refuses(parent, "leaf");
            Check(!result.IsCompleted && !File.Exists(Path.Combine(parent.DataDirectory(), "submit.md")),
                "Refused parent submission had a completion effect.");
            await leaf.ShutdownAsync().WaitAsync(deadline.Token);
            Refuses(parent, "leaf"); // Stopped but retained is still unresolved.
            leaf.Destroy();
            Agent submitted = await parent.CreateChildAsync("submitted", cancellation: deadline.Token);
            submitted.PublishSubmission("leaf result");
            parent.PublishSubmission("parent result");
            Check((await result).GetProperty("result").GetString() == "parent result",
                "Bottom-up completion did not wake parent observers.");
            Check(parent.FindChild("submitted") == submitted,
                "Accepting parent submission destroyed a submitted child.");
        }
        finally { app.Destroy(); }
    }

    private static void Refuses(Agent parent, string childID)
    {
        try { parent.PublishSubmission("too early"); }
        catch (InvalidOperationException error)
        {
            Check(error.Message.Contains(childID, StringComparison.Ordinal), "Refusal omitted unresolved child ID.");
            return;
        }
        throw new Exception("Parent submitted while a child remained unresolved.");
    }
}
