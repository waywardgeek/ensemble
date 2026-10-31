using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    private static async Task<string> CompletionFailure(Task task, CancellationToken deadline)
    {
        try { await task.WaitAsync(deadline); }
        catch (Exception error) when (error is not OperationCanceledException && !deadline.IsCancellationRequested)
        {
            return error.Message;
        }
        throw new Exception("Expected an explicit task failure, not success or deadline expiry.");
    }

    public static async Task ChildCompletionExhaustion(string temp)
    {
        string work = Path.Combine(temp, "completion-exhaustion");
        Directory.CreateDirectory(work);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        int requests = 0;
        var recovering = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var releaseRecovery = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var server = Task.Run(async () =>
        {
            for (int index = 0; index < 4; index++)
            {
                using var socket = await listener.AcceptTcpClientAsync(deadline.Token);
                using var stream = socket.GetStream();
                string request = await ReadRequest(stream, deadline.Token);
                requests++;
                if (index > 0)
                {
                    Check(request.Contains($"Framework correction {index}/3"), "Correction missing from next request.");
                }
                await Reply(stream, "I am done (but no submission)", deadline.Token);
            }
            using var recoverySocket = await listener.AcceptTcpClientAsync(deadline.Token);
            using var recoveryStream = recoverySocket.GetStream();
            await ReadRequest(recoveryStream, deadline.Token);
            recovering.TrySetResult();
            await releaseRecovery.Task.WaitAsync(deadline.Token);
            await Reply(recoveryStream, "recovered", deadline.Token);
        });
        var app = new Application(TextWriter.Null);
        try
        {
            var config = new Configuration(Vendor.Anthropic, "fixture",
                "http://127.0.0.1:" + ((IPEndPoint)listener.LocalEndpoint).Port, "local-only", sandboxRoot: work);
            var parent = new Agent(app, config);
            var child = new Agent(app, parent, "child", config, dataDirectory: Path.Combine(work, "child"));
            child.StartPersistence(new PersistenceOptions(Path.Combine(work, "child", "save.json")));
            Task<JsonElement> waiting = child.WaitForSubmissionAsync(deadline.Token);
            string failure = await CompletionFailure(child.SubmitAsync([new TextPart("do the task")]), deadline.Token);
            Check(failure.Contains("three framework corrections"), "Missing explicit exhaustion failure.");
            Check(await CompletionFailure(waiting, deadline.Token) == failure, "Submission observer saw a different outcome.");
            Check(requests == 4 && child.ChildFailure() == failure, "Unbounded retries or lost failure.");
            Check(child.Submission() is null && !File.Exists(Path.Combine(child.DataDirectory(), "submit.md")),
                "Final prose manufactured completion.");
            var events = child.History().Events();
            Check(events.SelectMany(e => e.Parts).OfType<TextPart>()
                .Count(p => p.Text.StartsWith("Framework correction ")) == 3, "Incorrect reminder count.");
            Check(events.Any(e => e.Kind == EventKind.ErrorOccurred && e.Error == failure), "Failure absent from History.");
            Task<string> recovery = child.SubmitAsync([new TextPart("recover explicitly")]);
            await recovering.Task.WaitAsync(deadline.Token);
            Check(child.ChildFailure() is null && child.Submission() is null,
                "New task inherited old failure or completion.");
            Task<JsonElement> recovered = child.WaitForSubmissionAsync(deadline.Token);
            Check(!recovered.IsCompleted, "New task's observer completed prematurely.");
            child.PublishSubmission("recovered");
            Check((await recovered).GetProperty("result").GetString() == "recovered", "Recovery lost submission.");
            releaseRecovery.TrySetResult();
            await recovery.WaitAsync(deadline.Token);
            await server;
        }
        finally
        {
            deadline.Cancel();
            listener.Stop();
            app.Destroy();
            try { await server; }
            catch (Exception error) when (error is OperationCanceledException or IOException or SocketException) { }
        }
    }

    public static async Task ChildCompletionResetAndDrain(string temp)
    {
        string work = Path.Combine(temp, "completion-reset");
        Directory.CreateDirectory(work);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        var reached = Enumerable.Range(0, 2).Select(_ => new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously)).ToArray();
        var release = Enumerable.Range(0, 2).Select(_ => new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously)).ToArray();
        var server = Task.Run(async () =>
        {
            for (int task = 0; task < 2; task++)
            {
                for (int index = 0; index < 4; index++)
                {
                    using var socket = await listener.AcceptTcpClientAsync(deadline.Token);
                    using var stream = socket.GetStream();
                    string request = await ReadRequest(stream, deadline.Token);
                    if (index > 0 && task == 0)
                    {
                        Check(request.Contains("Join or shut down unresolved children") && request.Contains("unfinished"),
                            "Reminder omitted unresolved child ID.");
                    }
                    if (index == 3)
                    {
                        reached[task].TrySetResult();
                        await release[task].Task.WaitAsync(deadline.Token);
                    }
                    await Reply(stream, "ordinary final text", deadline.Token);
                }
            }
        });
        var app = new Application(TextWriter.Null);
        try
        {
            var config = new Configuration(Vendor.Anthropic, "fixture",
                "http://127.0.0.1:" + ((IPEndPoint)listener.LocalEndpoint).Port, "local-only", sandboxRoot: work);
            var parent = new Agent(app, config);
            var child = new Agent(app, parent, "child", config, dataDirectory: Path.Combine(work, "child"));
            child.StartPersistence(new PersistenceOptions(Path.Combine(work, "child", "save.json")));
            var descendant = new Agent(app, child, "unfinished", config);
            for (int task = 0; task < 2; task++)
            {
                Task<string> turn = child.SubmitAsync([new TextPart("task " + task)]);
                await reached[task].Task.WaitAsync(deadline.Token);
                if (task == 0) descendant.Destroy();
                child.PublishSubmission("completed " + task);
                JsonElement payload = await child.WaitForSubmissionAsync(deadline.Token);
                Check(payload.GetProperty("result").GetString() == "completed " + task, "Wrong submission generation.");
                Check(!turn.IsCompleted, "Submission claimed model/save draining was complete.");
                release[task].TrySetResult();
                await turn.WaitAsync(deadline.Token);
                Check(child.ChildFailure() is null, "Valid submission retained failure.");
            }
            await server;
            Check(child.History().Events().SelectMany(e => e.Parts).OfType<TextPart>()
                .Count(p => p.Text.StartsWith("Framework correction ")) == 6, "Successful submit did not reset correction budget.");
        }
        finally
        {
            foreach (var signal in release) signal.TrySetResult();
            deadline.Cancel();
            listener.Stop();
            app.Destroy();
            try { await server; }
            catch (Exception error) when (error is OperationCanceledException or IOException or SocketException) { }
        }
    }

    public static async Task ChildCompletionErrorsAndRoot(string temp)
    {
        string work = Path.Combine(temp, "completion-error");
        Directory.CreateDirectory(work);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        int requests = 0;
        var server = Task.Run(async () =>
        {
            using (var socket = await listener.AcceptTcpClientAsync(deadline.Token))
            using (var stream = socket.GetStream())
            {
                await ReadRequest(stream, deadline.Token);
                requests++;
                await ReplyJson(stream, new { invalid = "not a provider response" }, deadline.Token);
            }
            using (var socket = await listener.AcceptTcpClientAsync(deadline.Token))
            using (var stream = socket.GetStream())
            {
                string request = await ReadRequest(stream, deadline.Token);
                Check(request.Contains("root task"), "Broken child response triggered a retry.");
                requests++;
                await Reply(stream, "root answer", deadline.Token);
            }
        });
        var app = new Application(TextWriter.Null);
        try
        {
            var config = new Configuration(Vendor.Anthropic, "fixture",
                "http://127.0.0.1:" + ((IPEndPoint)listener.LocalEndpoint).Port, "local-only", sandboxRoot: work);
            var root = new Agent(app, config);
            var child = new Agent(app, root, "child", config, dataDirectory: Path.Combine(work, "child"));
            child.StartPersistence(new PersistenceOptions(Path.Combine(work, "child", "save.json")));
            string failure = await CompletionFailure(child.SubmitAsync([new TextPart("child task")]), deadline.Token);
            Check(child.ChildFailure() == failure, "Real execution failure was not retained.");
            Check(await CompletionFailure(child.WaitForSubmissionAsync(deadline.Token), deadline.Token) == failure,
                "Waiter hung or lost actual execution failure.");
            Check(await root.SubmitAsync([new TextPart("root task")]).WaitAsync(deadline.Token) == "root answer",
                "Root chat incorrectly required submit.");
            await server;
            Check(requests == 2 && root.ChildFailure() is null, "Unexpected retries or root failure state.");
            await child.ShutdownAsync();
            Check(child.ChildFailure() == failure, "Shutdown erased the actual task failure.");
        }
        finally
        {
            deadline.Cancel();
            listener.Stop();
            app.Destroy();
            try { await server; }
            catch (Exception error) when (error is OperationCanceledException or IOException or SocketException) { }
        }
    }
}
