using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    public static async Task ChildGracefulDrain(string temp)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        var firstEntered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var secondEntered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var releaseFirst = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var releaseSecond = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        Task server = Task.Run(async () =>
        {
            for (int index = 0; index < 2; index++)
            {
                using var client = await listener.AcceptTcpClientAsync(deadline.Token);
                using var stream = client.GetStream();
                await ReadRequest(stream, deadline.Token);
                (index == 0 ? firstEntered : secondEntered).TrySetResult();
                await (index == 0 ? releaseFirst : releaseSecond).Task.WaitAsync(deadline.Token);
                await Reply(stream, index == 0 ? "first completed" : "second completed", deadline.Token);
            }
        });
        var app = new Application(TextWriter.Null);
        try
        {
            string work = Path.Combine(temp, "graceful-drain");
            var config = new Configuration(Vendor.Anthropic, "fixture",
                "http://127.0.0.1:" + ((IPEndPoint)listener.LocalEndpoint).Port);
            var parent = new Agent(app, config, dataDirectory: work);
            parent.InstallTools();
            Agent child = await parent.CreateChildAsync("worker", cancellation: deadline.Token);
            Agent sibling = await parent.CreateChildAsync("sibling", cancellation: deadline.Token);
            string home = child.DataDirectory();
            Task<string> first = child.SubmitAsync([new TextPart("first")]);
            await firstEntered.Task.WaitAsync(deadline.Token);
            Task<string> second = child.SubmitAsync([new TextPart("queued")]);
            await child.WorkStatusAsync(deadline.Token); // Admission barrier, not a delay.
            ToolResultPart early = await ChildCall(parent, "shutdown_agent",
                new { agent_id = "worker", ai_callback_delay = 0 }, deadline.Token);
            Check(early.Job?.Status == "running", "Graceful shutdown did not wait for active work.");
            Job shutdown = parent.Jobs().Find(early.Job!.Handle);
            Check(!first.IsCompleted && parent.FindChild("worker") == child,
                "Graceful shutdown cancelled the active turn or detached the child.");
            releaseFirst.TrySetResult();
            Check(await first.WaitAsync(deadline.Token) == "first completed", "Active turn did not finish normally.");
            await secondEntered.Task.WaitAsync(deadline.Token);
            // A durable submission is not enough: the second transport still has
            // to finish and publish its response before graceful teardown starts.
            child.PublishSubmission("queued task submitted");
            Check(shutdown.Snapshot().Status == "running" && !second.IsCompleted,
                "Graceful shutdown confused submission with drained queued work.");
            releaseSecond.TrySetResult();
            Check(await second.WaitAsync(deadline.Token) == "second completed", "Queued turn was cancelled.");
            ToolResultPart final = await shutdown.CompletionAsync(deadline.Token);
            Check(!final.IsError && parent.FindChild("worker") is null && child.AgentParent() is null
                && parent.FindChild("sibling") == sibling, "Graceful shutdown did not join exactly its subtree.");
            Check(File.Exists(Path.Combine(home, "save.json")) && File.Exists(Path.Combine(home, "submit.md")),
                "Graceful shutdown lost retained work.");
            await server.WaitAsync(deadline.Token);
        }
        finally
        {
            releaseFirst.TrySetResult();
            releaseSecond.TrySetResult();
            app.Destroy();
            listener.Stop();
            try { await server; }
            catch (Exception) when (server.IsFaulted || server.IsCanceled) { }
        }
    }

    public static async Task ChildGracefulCancellation(string temp)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var app = new Application(TextWriter.Null);
        try
        {
            var parent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1"),
                dataDirectory: Path.Combine(temp, "graceful-cancel"));
            parent.InstallTools();
            Agent child = await parent.CreateChildAsync("worker", cancellation: deadline.Token);
            await child.PostAsync(new AgentInput(InputKind.Pause));
            Task<string> queued = child.SubmitAsync([new TextPart("paused task")]);
            Check((await child.WorkStatusAsync(deadline.Token)).Paused, "Pause was not admitted.");
            ToolResultPart early = await ChildCall(parent, "shutdown_agent",
                new { agent_id = "worker", ai_callback_delay = 0 }, deadline.Token);
            Job observer = parent.Jobs().Find(early.Job!.Handle);
            await observer.KillAsync().WaitAsync(deadline.Token);
            AgentWorkStatus status = await child.WorkStatusAsync(deadline.Token);
            Check(status.Paused && status.Activity == AgentActivity.Busy && !queued.IsCompleted
                && parent.FindChild("worker") == child,
                "Cancelling graceful observation stopped, resumed, or detached paused work.");
            // Kill is the escape hatch for an intentionally paused/stuck drain.
            ToolResultPart killed = await ChildCall(parent, "kill_agent", new { agent_id = "worker" }, deadline.Token);
            if (killed.Job?.Status == "running")
            {
                killed = await parent.Jobs().Find(killed.Job.Handle).CompletionAsync(deadline.Token);
            }
            Check(!killed.IsError && parent.FindChild("worker") is null, "Kill did not release paused work.");
            try
            {
                await queued.WaitAsync(deadline.Token);
                throw new Exception("Killed queued task returned success.");
            }
            catch (ObjectDisposedException) { }
            catch (OperationCanceledException) when (!deadline.IsCancellationRequested) { }
        }
        finally { app.Destroy(); }
    }

    public static async Task ChildKillLifetime(string temp,
        [System.Runtime.CompilerServices.CallerFilePath] string sourceFile = "")
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        var modelEntered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var closeServer = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        Task server = Task.Run(async () =>
        {
            using var client = await listener.AcceptTcpClientAsync(deadline.Token);
            using var stream = client.GetStream();
            await ReadRequest(stream, deadline.Token);
            modelEntered.TrySetResult();
            // No response is ever supplied: kill must not require provider progress.
            await closeServer.Task.WaitAsync(deadline.Token);
        });
        var app = new Application(TextWriter.Null);
        try
        {
            string skills = Path.GetFullPath(Path.Combine(Path.GetDirectoryName(sourceFile)!, "..", "Skills"));
            var config = new Configuration(Vendor.Anthropic, "fixture",
                "http://127.0.0.1:" + ((IPEndPoint)listener.LocalEndpoint).Port);
            var parent = new Agent(app, config, dataDirectory: Path.Combine(temp, "kill-lifetime"),
                skillSpec: new SkillSpec(skills, "ensemble"));
            await parent.SkillManager()!.InitializeAsync(deadline.Token);
            Check(parent.ToolEnabled("kill_agent") && parent.Engine().Render(parent.History().View())
                .ToJsonString().Contains("kill_agent", StringComparison.Ordinal), "Shipped skill omitted kill_agent.");
            var child = new Agent(app, parent, "worker", config);
            child.InstallTools();
            child.StartPersistence(new PersistenceOptions(Path.Combine(child.DataDirectory(), "save.json")));
            var nested = new Agent(app, child, "nested", config);
            Agent sibling = await parent.CreateChildAsync("sibling", cancellation: deadline.Token);
            var foreign = new Agent(app, config, name: "foreign");
            _ = new Agent(app, foreign, "elsewhere", config);
            foreach (string tool in new[] { "shutdown_agent", "kill_agent" })
            {
                ToolResultPart unknown = await ChildCall(parent, tool, new { agent_id = "missing" }, deadline.Token);
                foreach (string id in new[] { "nested", "foreign", "elsewhere" })
                {
                    ToolResultPart refusal = await ChildCall(parent, tool, new { agent_id = id }, deadline.Token);
                    Check(unknown.IsError && refusal.IsError && JsonCodec.Text(unknown.Parts) == JsonCodec.Text(refusal.Parts),
                        "Lifecycle control exposed or stopped a non-owned Agent.");
                }
            }
            // A direct tool dispatch returns a running handle, leaving this Job
            // independent of the later model turn. Teardown must join its finally.
            var workEntered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            var workEnded = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            _ = new Tool(child, "held_lifetime", "Wait until lifetime cancellation",
                JsonSerializer.SerializeToElement(new { type = "object", properties = new { } }),
                async (job, args, cancellation) =>
                {
                    workEntered.TrySetResult();
                    try { await Task.Delay(Timeout.InfiniteTimeSpan, cancellation); }
                    finally { workEnded.TrySetResult(); }
                    return "unexpected completion";
                });
            ToolResultPart running = await ChildCall(child, "held_lifetime", new { ai_callback_delay = 0 }, deadline.Token);
            await workEntered.Task.WaitAsync(deadline.Token);
            Check(running.Job?.Status == "running", "Independent work did not start.");
            Task<string> turn = child.SubmitAsync([new TextPart("held model request")]);
            await modelEntered.Task.WaitAsync(deadline.Token);
            ToolResultPart result = await ChildCall(parent, "kill_agent", new { agent_id = "worker" }, deadline.Token);
            if (result.Job?.Status == "running")
            {
                result = await parent.Jobs().Find(result.Job.Handle).CompletionAsync(deadline.Token);
            }
            Check(!result.IsError && workEnded.Task.IsCompleted && parent.FindChild("worker") is null
                && child.AgentParent() is null && nested.AgentParent() is null && parent.FindChild("sibling") == sibling,
                "Kill returned before joined work/detachment, or removed a sibling.");
            try
            {
                await turn.WaitAsync(deadline.Token);
                throw new Exception("Killed model turn returned success.");
            }
            catch (ObjectDisposedException) { }
            catch (OperationCanceledException) when (!deadline.IsCancellationRequested) { }
        }
        finally
        {
            closeServer.TrySetResult();
            app.Destroy();
            listener.Stop();
            try { await server; }
            catch (Exception) when (server.IsFaulted || server.IsCanceled) { }
        }
    }
}
