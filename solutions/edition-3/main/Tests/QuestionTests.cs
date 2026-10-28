using Ensemble;
using Application = Ensemble.Ensemble;

internal static class QuestionTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition) throw new Exception(message);
    }

    private static void Refuses(Action action)
    {
        try { action(); }
        catch (InvalidOperationException) { return; }
        throw new Exception("Expected question refusal.");
    }

    public static async Task Routing(string temp)
    {
        string work = Path.Combine(temp, "question-routing");
        Directory.CreateDirectory(work);
        var app = new Application(TextWriter.Null);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        try
        {
            var config = new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1", "local-only", sandboxRoot: work);
            var root = new Agent(app, config);
            var a = new Agent(app, root, "a", config);
            var b = new Agent(app, root, "b", config);
            var leaf = new Agent(app, a, "leaf", config);
            var observed = root.WaitForQuestionsAsync(cancellation: deadline.Token);
            var humanObserved = root.WaitForQuestionsAsync(humanOnly: true, cancellation: deadline.Token);
            Task<string> first = a.SendMessageToParentAsync("first", deadline.Token);
            Check((await observed).Single().Content == "first" && !humanObserved.IsCompleted,
                "Child question did not wake parent, or reached the human-only sink.");
            Task<string> second = a.SendMessageToParentAsync("second", deadline.Token);
            Task<string> other = b.SendMessageToParentAsync("other", deadline.Token);
            QuestionInfo[] pending = root.PendingQuestions().ToArray();
            Check(pending.Length == 3 && pending.Select(question => question.MessageID).Distinct().Count() == 3,
                "Receiver did not allocate unique question IDs.");
            Refuses(() => root.RespondToQuestion("ambiguous", "a"));
            long otherID = pending.Single(question => question.AgentID == "b").MessageID;
            Refuses(() => root.RespondToQuestion("wrong sender", "a", otherID));
            Check(root.PendingQuestions().Length == 3 && !first.IsCompleted && !second.IsCompleted,
                "Refused reply consumed a pending question.");
            long firstID = pending.Single(question => question.Content == "first").MessageID;
            root.RespondToQuestion("answer one", "a", firstID);
            root.RespondToQuestion("answer two", "a");
            root.RespondToQuestion("answer b", "b");
            Check(await first == "answer one" && await second == "answer two" && await other == "answer b",
                "Reply was delivered to the wrong question.");
            Refuses(() => root.RespondToQuestion("duplicate", "a", firstID));

            Task<string> nested = leaf.SendMessageToParentAsync("nested", deadline.Token);
            Check(root.PendingQuestions().IsEmpty && a.PendingQuestions().Single().AgentID == "leaf",
                "Question skipped its actual Agent parent.");
            a.RespondToQuestion("nested reply", "leaf");
            Check(await nested == "nested reply", "Nested reply lost.");

            Task<string> human = root.SendMessageToParentAsync("human please", deadline.Token);
            QuestionInfo humanInfo = (await humanObserved).Single();
            Check(humanInfo.ForHuman && humanInfo.AgentID is null && humanInfo.MessageID > otherID,
                "Human inbox does not share the receiver's stable ID sequence.");
            root.RespondToQuestion("human answer");
            Check(await human == "human answer" && root.PendingQuestions().IsEmpty,
                "Human reply failed to detach its self-edge.");
        }
        finally { app.Destroy(); }
    }

    public static async Task CancellationAndOwners(string temp)
    {
        string work = Path.Combine(temp, "question-cancellation");
        Directory.CreateDirectory(work);
        var app = new Application(TextWriter.Null);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        try
        {
            var config = new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1", "local-only", sandboxRoot: work);
            var root = new Agent(app, config);
            var child = new Agent(app, root, "child", config);
            using var stopObserver = new CancellationTokenSource();
            var cancelledObserver = root.WaitForQuestionsAsync(cancellation: stopObserver.Token);
            var observer1 = root.WaitForQuestionsAsync(cancellation: deadline.Token);
            var observer2 = root.WaitForQuestionsAsync(cancellation: deadline.Token);
            stopObserver.Cancel();
            await Cancelled(cancelledObserver, deadline.Token);
            using var stopCaller = new CancellationTokenSource();
            Task<string> caller = child.SendMessageToParentAsync("retain this question", stopCaller.Token);
            Check((await observer1).Single().MessageID == (await observer2).Single().MessageID,
                "One observer consumed the other's notification.");
            stopCaller.Cancel();
            await Cancelled(caller, deadline.Token);
            Check(root.PendingQuestions().Single().Content == "retain this question",
                "Caller cancellation discarded an admitted question.");
            root.RespondToQuestion("late answer", "child");
            Check(root.PendingQuestions().IsEmpty, "Late reply did not detach the cancelled caller's question.");

            Task<string> outgoing = child.SendMessageToParentAsync("child stops", deadline.Token);
            await child.ShutdownAsync().WaitAsync(deadline.Token);
            await Cancelled(outgoing, deadline.Token);
            Check(root.PendingQuestions().IsEmpty, "Sender shutdown left its recipient link.");
            child.Destroy();
            var replacement = new Agent(app, root, "child", config);
            Task<string> pending = replacement.SendMessageToParentAsync("parent stops", deadline.Token);
            Task<string> human = root.SendMessageToParentAsync("human stops", deadline.Token);
            root.RespondToQuestion("human done");
            Check(await human == "human done", "Human reply selected a child's question.");
            var emptyHumanInbox = root.WaitForQuestionsAsync(humanOnly: true, cancellation: deadline.Token);
            await root.ShutdownAsync().WaitAsync(deadline.Token);
            await Cancelled(pending, deadline.Token);
            try { await emptyHumanInbox; throw new Exception("Closed inbox reported a question."); }
            catch (ObjectDisposedException) { }
            Check(root.PendingQuestions().IsEmpty && replacement.PendingQuestions().IsEmpty,
                "Receiver shutdown left a question relationship.");
        }
        finally { app.Destroy(); }
    }

    public static async Task ConcurrentShutdown(string temp)
    {
        string work = Path.Combine(temp, "question-race");
        Directory.CreateDirectory(work);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        for (int index = 0; index < 30; index++)
        {
            var app = new Application(TextWriter.Null);
            try
            {
                var config = new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1", "local-only", sandboxRoot: work);
                var root = new Agent(app, config);
                var child = new Agent(app, root, "child", config);
                var start = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
                Task asking = Task.Run(async () =>
                {
                    await start.Task;
                    try { await child.SendMessageToParentAsync("racing question", deadline.Token); }
                    catch (ObjectDisposedException) { return; }
                    catch (OperationCanceledException) when (!deadline.IsCancellationRequested) { return; }
                    throw new Exception("Unanswered question unexpectedly succeeded.");
                });
                Task stopping = Task.Run(async () => { await start.Task; await root.ShutdownAsync(); });
                start.SetResult();
                await Task.WhenAll(asking, stopping).WaitAsync(deadline.Token);
                Check(root.PendingQuestions().IsEmpty, "Question admitted after shutdown drain.");
            }
            finally { app.Destroy(); }
        }
    }

    private static async Task Cancelled(Task task, CancellationToken deadline)
    {
        try { await task.WaitAsync(deadline); }
        catch (OperationCanceledException) when (!deadline.IsCancellationRequested) { return; }
        throw new Exception("Expected unresolved question cancellation.");
    }
}
