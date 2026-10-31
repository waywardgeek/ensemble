using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

// Real Agent/Jobs/question/persistence integration. There is no model boundary
// in these tests: the later executable workflow must prove model-led tool use.
internal static class ChildJoinTests
{
    private static void Check(bool value, string why)
    {
        if (!value) throw new Exception(why);
    }

    private static Agent Parent(Application app, string temp, string name)
    {
        string work = Path.Combine(temp, name);
        Directory.CreateDirectory(work);
        var parent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture",
            "http://127.0.0.1:1", "local-only", sandboxRoot: work), dataDirectory: work);
        parent.InstallTools();
        parent.StartPersistence(new PersistenceOptions(Path.Combine(work, "save.json")));
        return parent;
    }

    private static JsonElement J(string text) => JsonSerializer.Deserialize<JsonElement>(text);
    private static Task<ToolResultPart> Call(Agent parent, string id, string tool, object args,
        CancellationToken cancellation) => parent.DispatchAsync(new ToolCallPart(id,
            parent.Configuration().Target(), tool, JsonSerializer.SerializeToElement(args)), cancellation);

    public static async Task Selection(string temp)
    {
        var app = new Application(TextWriter.Null);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        try
        {
            Agent parent = Parent(app, temp, "join-selection");
            await using (var empty = parent.JoinChildrenAsync(cancellation: deadline.Token).GetAsyncEnumerator())
            {
                Check(await empty.MoveNextAsync() && empty.Current.Submissions?.Count == 0,
                    "No children did not yield an empty completed result.");
                Check(!await empty.MoveNextAsync(), "Completed empty join continued.");
            }
            Agent one = await parent.CreateChildAsync("one", submitSchema: J("""
                {"type":"object","properties":{"count":{"type":"integer"}},"required":["count"],"additionalProperties":false}
                """), cancellation: deadline.Token);
            Agent two = await parent.CreateChildAsync("two", cancellation: deadline.Token);
            await using var all = parent.JoinChildrenAsync(cancellation: deadline.Token).GetAsyncEnumerator();
            Task<bool> waiting = all.MoveNextAsync().AsTask();
            Check(!waiting.IsCompleted, "Idle children counted as submitted.");
            Agent later = await parent.CreateChildAsync("later", cancellation: deadline.Token);
            one.PublishSubmission(J("{\"count\":9007199254740993}"));
            Check(!waiting.IsCompleted, "Partial submission completed fan-in.");
            two.PublishSubmission("second");
            Check(await waiting, "Join omitted its result.");
            var payloads = all.Current.Submissions!;
            Check(payloads.Count == 2 && !payloads.ContainsKey("later")
                && payloads["one"].GetProperty("count").GetRawText() == "9007199254740993"
                && payloads["two"].GetProperty("result").GetString() == "second",
                "Join widened selection or lost typed payloads.");
            Check(later.Submission() is null && !await all.MoveNextAsync(), "Join affected later child.");
            await using var subset = parent.JoinChildrenAsync(["two"], deadline.Token).GetAsyncEnumerator();
            Check(await subset.MoveNextAsync() && subset.Current.Submissions?.Count == 1,
                "Explicit subset included unresolved unrelated child.");
        }
        finally { app.Destroy(); }
    }

    public static async Task Questions(string temp)
    {
        var app = new Application(TextWriter.Null);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        try
        {
            Agent parent = Parent(app, temp, "join-questions");
            Agent child = await parent.CreateChildAsync("worker", cancellation: deadline.Token);
            Task<QuestionReply> answer = child.SendMessageToParentAsync("Which file?", deadline.Token);
            parent.Jobs().SetLimits(new JobLimits(120, "will-not-match"));
            ToolResultPart first = await Call(parent, "join", "join_agents", new { }, deadline.Token);
            Check(!first.IsError && first.Job?.Status == "running", "Question ended join instead of yielding a handle.");
            string text = JsonCodec.Text(first.Parts);
            Check(text.Contains("pending_question") && text.Contains("worker") && text.Contains("Which file?"),
                "Running report omitted actionable question.");
            int handle = first.Job!.Handle;
            QuestionInfo question = parent.PendingQuestions().Single();
            Check(text.Contains("\"message_id\":" + question.MessageID), "Question ID missing from report.");
            parent.RespondToQuestion("note.txt", "worker", question.MessageID);
            Check((await answer).Response == "note.txt", "Question reply lost.");
            ToolResultPart silent = await parent.Jobs().Find(handle).ObserveAsync("silent", 0,
                cancellation: deadline.Token);
            Check(silent.Job?.Status == "running" && !JsonCodec.Text(silent.Parts).Contains("pending_question"),
                "Same question was repeated after its output was consumed.");

            Task<QuestionReply> another = child.SendMessageToParentAsync("Which format?", deadline.Token);
            parent.Jobs().SetLimits(new JobLimits(120, "will-not-match"));
            ToolResultPart second = await Call(parent, "wait-question", "wait_for_job", new { handle }, deadline.Token);
            Check(second.Job?.Handle == handle && second.Job.Status == "running"
                && JsonCodec.Text(second.Parts).Contains("Which format?"), "Second question did not wake same join.");
            parent.RespondToQuestion("text", "worker");
            Check((await another).Response == "text", "Second answer lost.");
            child.PublishSubmission("finished");
            ToolResultPart done = await Call(parent, "wait-done", "wait_for_job", new { handle }, deadline.Token);
            Check(done.Job?.Status == "done" && !done.IsError
                && JsonCodec.Text(done.Parts).Contains("\"worker\":{\"result\":\"finished\"}"),
                "Join did not finish with typed child payload.");
            Check(parent.FindChild("worker") == child, "Join destroyed child.");
        }
        finally { app.Destroy(); }
    }

    public static async Task Cancellation(string temp)
    {
        var app = new Application(TextWriter.Null);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        try
        {
            Agent parent = Parent(app, temp, "join-cancellation");
            Agent child = await parent.CreateChildAsync("worker", cancellation: deadline.Token);
            using var observer = new CancellationTokenSource();
            await using (var stream = parent.JoinChildrenAsync(cancellation: observer.Token).GetAsyncEnumerator())
            {
                Task<bool> wait = stream.MoveNextAsync().AsTask();
                observer.Cancel();
                try { await wait; throw new Exception("Cancelled join observer succeeded."); }
                catch (OperationCanceledException) when (observer.IsCancellationRequested) { }
            }
            Task<QuestionReply> question = child.SendMessageToParentAsync("Still here?", deadline.Token);
            ToolResultPart report = await Call(parent, "join-cancel", "join_agents", new { }, deadline.Token);
            Check(report.Job?.Status == "running", "Expected running join.");
            await parent.Jobs().Find(report.Job!.Handle).KillAsync();
            Check(parent.PendingQuestions().Length == 1 && !question.IsCompleted
                && parent.FindChild("worker") == child, "Killing join cancelled child/question.");
            parent.RespondToQuestion("yes", "worker");
            Check((await question).Response == "yes", "Question did not survive join cancellation.");
            child.PublishSubmission("survived");
            await using var fresh = parent.JoinChildrenAsync(cancellation: deadline.Token).GetAsyncEnumerator();
            Check(await fresh.MoveNextAsync() && fresh.Current.Submissions?.Count == 1,
                "Fresh observer could not see surviving child completion.");
        }
        finally { app.Destroy(); }
    }

    public static async Task Failures(string temp)
    {
        var app = new Application(TextWriter.Null);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        try
        {
            Agent parent = Parent(app, temp, "join-failures");
            Agent child = await parent.CreateChildAsync("worker", cancellation: deadline.Token);
            foreach (object args in new object[] { new { agent_ids = new[] { "missing" } },
                new { agent_ids = new[] { "worker", "worker" } }, new { agent_ids = new[] { 7 } } })
            {
                ToolResultPart refused = await Call(parent, "invalid", "join_agents", args, deadline.Token);
                Check(refused.IsError, "Invalid join selection did not return a tool error.");
            }
            parent.Jobs().SetLimits(new JobLimits(0));
            ToolResultPart started = await Call(parent, "start-stop", "join_agents", new { }, deadline.Token);
            Check(started.Job?.Status == "running", "Unsubmitted child completed join.");
            await child.ShutdownAsync().WaitAsync(deadline.Token);
            ToolResultPart failed = await Call(parent, "wait-fail", "wait_for_job",
                new { handle = started.Job!.Handle }, deadline.Token);
            Check(failed.IsError && failed.Job?.Status == "done"
                && JsonCodec.Text(failed.Parts).Contains("worker")
                && JsonCodec.Text(failed.Parts).Contains("without submission"),
                "Stopped child did not produce actionable terminal tool failure.");
            child.Destroy();
            Agent live = await parent.CreateChildAsync("live", cancellation: deadline.Token);
            await using var stream = parent.JoinChildrenAsync(cancellation: deadline.Token).GetAsyncEnumerator();
            Task<bool> wait = stream.MoveNextAsync().AsTask();
            await parent.ShutdownAsync().WaitAsync(deadline.Token);
            try { await wait; throw new Exception("Parent shutdown returned successful join."); }
            catch (ObjectDisposedException) { }
            Check(live.Submission() is null, "Shutdown invented submission.");
        }
        finally { app.Destroy(); }
    }
}
