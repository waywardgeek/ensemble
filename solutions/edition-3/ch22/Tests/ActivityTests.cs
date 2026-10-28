using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    // Drain the ordinary lossy progress channel, retaining every phase seen on
    // the way to the requested boundary. Reliable state is checked separately.
    private static async Task<WorkActivity> NextWork(Agent agent, WorkPhase phase,
        List<WorkActivity> seen, CancellationToken cancellation)
    {
        Observation observation = await agent.WaitAsync(value =>
        {
            if (value.Kind != ObservationKind.ActivityChanged)
            {
                return false;
            }
            Check(value.Activity is not null, "Activity observation omitted its value.");
            seen.Add(value.Activity!);
            return value.Activity!.Phase == phase;
        }, cancellation);
        return observation.Activity!;
    }

    public static async Task ActivityModelCycle(string temp)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        var firstReceived = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var firstRelease = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var toolEntered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var toolRelease = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var secondReceived = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var secondRelease = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        Task server = Task.Run(async () =>
        {
            using (var client = await listener.AcceptTcpClientAsync(deadline.Token))
            using (var stream = client.GetStream())
            {
                _ = await ReadRequest(stream, deadline.Token);
                firstReceived.SetResult();
                await firstRelease.Task.WaitAsync(deadline.Token);
                await ReplyJson(stream, new
                {
                    model = "fixture", content = new object[]
                    {
                        new { type = "text", text = "I will run the held operation." },
                        new { type = "tool_use", id = "activity-tool", name = "activity_held",
                            input = new { ai_callback_delay = 120, ai_callback_pattern = "never-match" } }
                    }, usage = new { input_tokens = 1, output_tokens = 1 }
                }, deadline.Token);
            }
            using (var client = await listener.AcceptTcpClientAsync(deadline.Token))
            using (var stream = client.GetStream())
            {
                string body = await ReadRequest(stream, deadline.Token);
                Check(!body.Contains("Waiting for the model response") && !body.Contains("Executing a tool"),
                    "Activity projection entered model context.");
                secondReceived.SetResult();
                await secondRelease.Task.WaitAsync(deadline.Token);
                await Reply(stream, "finished", deadline.Token);
            }
        });
        var app = new Application(TextWriter.Null);
        try
        {
            int port = ((IPEndPoint)listener.LocalEndpoint).Port;
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:" + port),
                outputDirectory: Path.Combine(temp, "activity-cycle"));
            _ = new Tool(agent, "activity_held", "Held external operation",
                JsonSerializer.SerializeToElement(new { type = "object" }), async (job, args, cancellation) =>
                {
                    toolEntered.TrySetResult();
                    await toolRelease.Task.WaitAsync(cancellation);
                    return "operation complete";
                });
            var seen = new List<WorkActivity>();
            Task<string> turn = agent.SubmitAsync([new TextPart("exercise activity")]);
            await firstReceived.Task.WaitAsync(deadline.Token);
            await NextWork(agent, WorkPhase.Thinking, seen, deadline.Token);
            Check((await agent.WorkStatusAsync(deadline.Token)).Work?.Phase == WorkPhase.Thinking,
                "Reliable status missed the held model request.");
            // These mailbox messages do not change work and must not duplicate
            // the currently published thinking phase.
            await agent.PostAsync(new AgentInput(InputKind.Pause));
            await agent.PostAsync(new AgentInput(InputKind.Resume));
            firstRelease.TrySetResult();
            await toolEntered.Task.WaitAsync(deadline.Token);
            WorkActivity tool = await NextWork(agent, WorkPhase.ToolCall, seen, deadline.Token);
            Check(tool.Tool == "activity_held" && tool.Detail == "Executing a tool", "Tool identity was lost.");
            Check((await agent.WorkStatusAsync(deadline.Token)).Work == tool, "Tool status differs from observation.");
            toolRelease.TrySetResult();
            await secondReceived.Task.WaitAsync(deadline.Token);
            await NextWork(agent, WorkPhase.Thinking, seen, deadline.Token);
            secondRelease.TrySetResult();
            Check(await turn.WaitAsync(deadline.Token) == "finished", "Activity changed reliable task output.");
            await NextWork(agent, WorkPhase.Idle, seen, deadline.Token);
            await agent.FlushAsync(deadline.Token);
            Check((await agent.WorkStatusAsync(deadline.Token)).Work?.Phase == WorkPhase.Idle, "Turn stayed busy.");
            Check(seen.Count(value => value.Phase == WorkPhase.Thinking) == 2, "Thinking was duplicated or missed.");
            Check(!seen.Zip(seen.Skip(1)).Any(pair => pair.First == pair.Second), "Duplicate activity publication.");
            Check(agent.History().Events().Count(fact => fact.Kind == EventKind.RequestSent) == 2,
                "Activity changed request recording.");
            Check(!agent.History().Events().Any(fact => JsonCodec.WriteEvent(fact).Contains("ActivityChanged")),
                "Progress was persisted as dialogue.");
            await server;
        }
        finally
        {
            firstRelease.TrySetResult(); toolRelease.TrySetResult(); secondRelease.TrySetResult();
            app.Destroy(); listener.Stop();
            await server.ConfigureAwait(ConfigureAwaitOptions.SuppressThrowing);
        }
    }

    public static async Task ActivityControlWait(string temp)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var app = new Application(TextWriter.Null);
        try
        {
            var owner = new Agent(app, new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1"));
            owner.InstallTools();
            var agent = new Agent(app, owner, "worker", owner.Configuration(),
                outputDirectory: Path.Combine(temp, "activity-control"));
            Job job = await HeldMessageJob(agent, release, deadline.Token);
            var seen = new List<WorkActivity>();
            await NextWork(agent, WorkPhase.Idle, seen, deadline.Token);
            Task<ToolResultPart> wait = ChildCall(agent, "wait_for_job", new
            {
                handle = job.Handle, ai_callback_delay = 120, ai_callback_pattern = "never-match"
            }, deadline.Token);
            WorkActivity waiting = await NextWork(agent, WorkPhase.Waiting, seen, deadline.Token);
            AgentWorkStatus state = await agent.WorkStatusAsync(deadline.Token);
            Check(waiting.Tool == "wait_for_job" && state.Work == waiting && state.Activity == AgentActivity.Busy,
                "Control-only wait was reported as idle or ordinary tool execution.");
            Check(!wait.IsCompleted && job.Snapshot().Status == "running", "Observation ended its independent work.");
            ToolResultPart listed = await ChildCall(owner, "agents_status", new { }, deadline.Token);
            using (var json = JsonDocument.Parse(string.Concat(listed.Parts.OfType<TextPart>().Select(part => part.Text))))
            {
                JsonElement row = json.RootElement.GetProperty("agents")[0];
                Check(row.GetProperty("work_phase").GetString() == "waiting"
                    && row.GetProperty("tool").GetString() == "wait_for_job"
                    && row.GetProperty("activity").GetString() == "busy", "Status wire fields lost work or lifecycle.");
            }
            release.TrySetResult();
            Check((await wait.WaitAsync(deadline.Token)).Job?.Status == "done", "Underlying Job did not finish.");
            await NextWork(agent, WorkPhase.Idle, seen, deadline.Token);
            await agent.ShutdownAsync();
            state = await agent.WorkStatusAsync(deadline.Token);
            Check(state.Activity == AgentActivity.Stopped && state.Work?.Phase == WorkPhase.Idle,
                "Closed status retained a stale phase.");
        }
        finally { release.TrySetResult(); app.Destroy(); }
    }

    public static async Task ActivityInterruptAndFailure(string temp)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1"),
                outputDirectory: Path.Combine(temp, "activity-interrupt"));
            Job job = await HeldMessageJob(agent, release, deadline.Token);
            var seen = new List<WorkActivity>();
            await NextWork(agent, WorkPhase.Idle, seen, deadline.Token);
            // A real foreground tool observation can be interrupted without
            // cancelling its independently owned execution.
            Task<ToolResultPart> held = ChildCall(agent, "held_message", new
            {
                ai_callback_delay = 120, ai_callback_pattern = "never-match"
            }, deadline.Token);
            await NextWork(agent, WorkPhase.ToolCall, seen, deadline.Token);
            await agent.PostAsync(new AgentInput(InputKind.InterruptNow));
            try { await held.WaitAsync(deadline.Token); }
            catch (OperationCanceledException) { }
            await NextWork(agent, WorkPhase.Idle, seen, deadline.Token);
            Check(job.Snapshot().Status == "running", "Interrupt killed independent execution.");
            _ = new Tool(agent, "activity_failure", "Failing external boundary",
                JsonSerializer.SerializeToElement(new { type = "object" }), (running, args, cancellation) =>
                    Task.FromException<string>(new IOException("expected boundary failure")));
            ToolResultPart failed = await ChildCall(agent, "activity_failure", new { ai_callback_delay = 120 }, deadline.Token);
            Check(failed.IsError, "Boundary failure became success.");
            await agent.FlushAsync(deadline.Token);
            Check((await agent.WorkStatusAsync(deadline.Token)).Work?.Phase == WorkPhase.Idle,
                "Failed tool left stale activity.");
            release.TrySetResult();
        }
        finally { release.TrySetResult(); app.Destroy(); }
    }

    public static async Task ActivityTransportEnds()
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        var received = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        Task server = Task.Run(async () =>
        {
            using (var client = await listener.AcceptTcpClientAsync(deadline.Token))
            using (var stream = client.GetStream())
            {
                _ = await ReadRequest(stream, deadline.Token);
                await ReplyJson(stream, new { malformed_provider_reply = true }, deadline.Token);
            }
            using (var client = await listener.AcceptTcpClientAsync(deadline.Token))
            using (var stream = client.GetStream())
            {
                _ = await ReadRequest(stream, deadline.Token);
                received.TrySetResult();
                await release.Task.WaitAsync(deadline.Token);
            }
        });
        var app = new Application(TextWriter.Null);
        try
        {
            int port = ((IPEndPoint)listener.LocalEndpoint).Port;
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:" + port));
            var seen = new List<WorkActivity>();
            Task<string> failed = agent.SubmitAsync([new TextPart("malformed transport")]);
            await NextWork(agent, WorkPhase.Thinking, seen, deadline.Token);
            // Agent reports transport failure through its existing task contract;
            // the progress stream must settle independently of that reply shape.
            try { await failed.WaitAsync(deadline.Token); }
            catch (Exception) when (failed.IsFaulted) { }
            await NextWork(agent, WorkPhase.Idle, seen, deadline.Token);
            Check(agent.History().Events().Any(fact => fact.Kind == EventKind.ErrorOccurred), "Malformed peer did not fail.");
            Task<string> pending = agent.SubmitAsync([new TextPart("shutdown transport")]);
            await received.Task.WaitAsync(deadline.Token);
            await NextWork(agent, WorkPhase.Thinking, seen, deadline.Token);
            await agent.ShutdownAsync().WaitAsync(deadline.Token);
            await NextWork(agent, WorkPhase.Idle, seen, deadline.Token);
            Check(pending.IsCompleted, "Shutdown stranded the reliable task reply.");
            await ((Task)pending).ConfigureAwait(ConfigureAwaitOptions.SuppressThrowing);
            AgentWorkStatus status = await agent.WorkStatusAsync(deadline.Token);
            Check(status.Activity == AgentActivity.Stopped && status.Work?.Phase == WorkPhase.Idle,
                "Shutdown left thinking in its closed snapshot.");
            release.TrySetResult();
            await server;
        }
        finally
        {
            release.TrySetResult(); app.Destroy(); listener.Stop();
            await server.ConfigureAwait(ConfigureAwaitOptions.SuppressThrowing);
        }
    }

}
