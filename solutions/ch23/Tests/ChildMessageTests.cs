using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using System.Text;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    private static async Task<ToolResultPart> ChildMessageResult(Agent parent, string name,
        object args, CancellationToken cancellation)
    {
        ToolResultPart result = await ChildCall(parent, name, args, cancellation);
        return result.Job?.Status == "running"
            ? await parent.Jobs().Find(result.Job.Handle).CompletionAsync(cancellation) : result;
    }

    public static async Task ChildOutputCursors(string temp)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var app = new Application(TextWriter.Null);
        try
        {
            var parent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1"),
                dataDirectory: Path.Combine(temp, "output-cursors"));
            parent.InstallTools();
            Agent child = await parent.CreateChildAsync("worker", cancellation: deadline.Token);
            child.History().Append(new Event(EventKind.HintReceived, actor: Actor.Human,
                parts: [new TextPart("not finalized output")]));
            child.History().Append(new Event(EventKind.ToolObserved, callID: "recorded-result",
                parts: [new TextPart("exact 9007199254740993")]));
            ChildOutput initial = await child.ReadOutputAsync(cancellation: deadline.Token);
            Check(initial.Output == "exact 9007199254740993\n" && initial.Cursor > 0,
                "Output exposed hints or lost exact finalized text.");
            Check((await child.ReadOutputAsync(initial.Cursor, true, deadline.Token)).Output == "",
                "Idle read replayed consumed output.");
            await child.PostAsync(new AgentInput(InputKind.Pause));
            Task<string> queued = child.SubmitAsync([new TextPart("queued")]);
            Check((await child.WorkStatusAsync(deadline.Token)).Activity == AgentActivity.Busy,
                "Queued paused work was not busy.");
            using var cancelled = new CancellationTokenSource();
            Task<ChildOutput> first = child.ReadOutputAsync(initial.Cursor, true, deadline.Token);
            Task<ChildOutput> second = child.ReadOutputAsync(initial.Cursor, true, deadline.Token);
            Task<ChildOutput> abandoned = child.ReadOutputAsync(initial.Cursor, true, cancelled.Token);
            cancelled.Cancel();
            try { await abandoned; throw new Exception("Cancelled observer completed normally."); }
            catch (OperationCanceledException) { }
            child.History().Append(new Event(EventKind.ToolObserved, callID: "next-result",
                parts: [new TextPart("fresh")]));
            ChildOutput a = await first;
            ChildOutput b = await second;
            Check(a.Output == "fresh\n" && a == b && a.Cursor > initial.Cursor && !queued.IsCompleted,
                "Independent readers consumed output or cancelled work.");
            ToolResultPart read = await ChildMessageResult(parent, "wait_for_agent",
                new { agent_id = "worker", cursor = initial.Cursor }, deadline.Token);
            using var row = JsonDocument.Parse(JsonCodec.Text(read.Parts));
            Check(!read.IsError && row.RootElement.GetProperty("output").GetString() == "fresh\n",
                "Output tool did not preserve finalized cursor output.");
            // Closure must settle an output waiter even after History releases
            // its owned event list. It must not report a spurious cursor error.
            Task<ChildOutput> closing = child.ReadOutputAsync(a.Cursor, true, deadline.Token);
            child.Destroy();
            Check((await closing).Status.Activity is AgentActivity.Stopped or AgentActivity.Stopping,
                "Destroyed child left an output observer waiting.");
            await ((Task)queued).ConfigureAwait(ConfigureAwaitOptions.SuppressThrowing);
        }
        finally { app.Destroy(); }
    }

    public static async Task ChildLiveMessages(string temp,
        [System.Runtime.CompilerServices.CallerFilePath] string sourceFile = "")
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var app = new Application(TextWriter.Null);
        try
        {
            string skills = Path.GetFullPath(Path.Combine(Path.GetDirectoryName(sourceFile)!, "..", "Skills"));
            var config = new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1");
            var parent = new Agent(app, config, dataDirectory: Path.Combine(temp, "message-authority"),
                skillSpec: new SkillSpec(skills, "ensemble"));
            await parent.SkillManager()!.InitializeAsync(deadline.Token);
            Agent child = await parent.CreateChildAsync("worker", cancellation: deadline.Token);
            var peer = new Agent(app, config, dataDirectory: Path.Combine(temp, "message-peer"));
            _ = new Agent(app, peer, "foreign", config);
            _ = new Agent(app, child, "nested", config);
            foreach (string tool in new[] { "send_message", "ask_agent", "wait_for_agent", "wait_for_agent_change" })
            {
                Check(parent.ToolEnabled(tool) && parent.Engine().Render(parent.History().View())
                    .ToJsonString().Contains(tool, StringComparison.Ordinal), "Shipped skill omitted " + tool);
                object Missing(string id) => tool == "wait_for_agent_change"
                    ? new { agent_ids = new[] { id } } : (object)new { agent_id = id, content = "test" };
                ToolResultPart missing = await ChildMessageResult(parent, tool, Missing("missing"), deadline.Token);
                foreach (string id in new[] { "foreign", "nested" })
                {
                    ToolResultPart refused = await ChildMessageResult(parent, tool, Missing(id), deadline.Token);
                    Check(refused.IsError && missing.IsError && JsonCodec.Text(refused.Parts) == JsonCodec.Text(missing.Parts),
                        "Message operation exposed another owner's child.");
                }
            }
            await child.PostAsync(new AgentInput(InputKind.Pause));
            // Idle Message admits work without resuming pause; later Message is
            // steering of that same accepted task, not a separate question.
            Check(!(await ChildMessageResult(parent, "send_message",
                new { agent_id = "worker", content = "start idle" }, deadline.Token)).IsError, "Idle send refused.");
            Check(!(await ChildMessageResult(parent, "send_message",
                new { agent_id = "worker", content = "steer busy" }, deadline.Token)).IsError, "Busy send refused.");
            AgentWorkStatus status = await child.WorkStatusAsync(deadline.Token);
            Check(status.Paused && status.Activity == AgentActivity.Busy
                && child.History().Events().Any(e => e.Kind == EventKind.HintReceived
                    && JsonCodec.Text(e.Parts) == "steer busy"), "Live steering did not use the hint path.");
            int hints = child.History().Events().Count(e => e.Kind == EventKind.HintReceived);
            foreach (object invalid in new object[] {
                new { agent_id = "worker", content = "bad cursor", cursor = -1 },
                new { agent_id = "worker", content = "bad cursor", cursor = long.MaxValue },
                new { agent_id = "worker", content = " " } })
            {
                Check((await ChildMessageResult(parent, "send_message", invalid, deadline.Token)).IsError,
                    "Invalid send admitted work.");
            }
            Check(child.History().Events().Count(e => e.Kind == EventKind.HintReceived) == hints,
                "Refused send recorded a hint.");
            ToolResultPart waiting = await ChildCall(parent, "ask_agent",
                new { agent_id = "worker", content = "queued separately", ai_callback_delay = 0 }, deadline.Token);
            Check(waiting.Job?.Status == "running", "Paused ask did not retain a running observer.");
            await parent.Jobs().Find(waiting.Job!.Handle).KillAsync();
            Check((await child.WorkStatusAsync(deadline.Token)).Paused && parent.FindChild("worker") == child,
                "Cancelling ask changed child lifetime or pause.");
        }
        finally { app.Destroy(); }
    }

    public static async Task ChildQueuedQuestion(string temp)
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
                await Reply(stream, index == 0 ? "older final" : "separate answer", deadline.Token);
            }
        });
        var app = new Application(TextWriter.Null);
        try
        {
            var parent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture",
                "http://127.0.0.1:" + ((IPEndPoint)listener.LocalEndpoint).Port),
                dataDirectory: Path.Combine(temp, "queued-question"));
            parent.InstallTools();
            Agent child = await parent.CreateChildAsync("worker", cancellation: deadline.Token);
            Task<string> old = child.SubmitAsync([new TextPart("old task")]);
            await firstEntered.Task.WaitAsync(deadline.Token);
            // Consume the original exact admission before admitting the second.
            // Production output readers never consume this shared queue.
            await child.WaitAsync(value => value.Kind == ObservationKind.Input
                && value.Text == "Submit" && value.Accepted == true, deadline.Token);
            ToolResultPart pending = await ChildCall(parent, "ask_agent",
                new { agent_id = "worker", content = "distinct question", ai_callback_delay = 0 }, deadline.Token);
            Job ask = parent.Jobs().Find(pending.Job!.Handle);
            await child.WaitAsync(value => value.Kind == ObservationKind.Input && value.Text == "Submit" && value.Accepted == true,
                deadline.Token);
            releaseFirst.TrySetResult();
            Check(await old.WaitAsync(deadline.Token) == "older final", "Original reply was replaced.");
            await secondEntered.Task.WaitAsync(deadline.Token);
            Check(!ask.CompletionAsync(deadline.Token).IsCompleted, "Ask used the older final.");
            child.PublishSubmission(JsonSerializer.SerializeToElement(new { result = "accepted" }));
            releaseSecond.TrySetResult();
            ToolResultPart result = await ask.CompletionAsync(deadline.Token);
            using var parsed = JsonDocument.Parse(JsonCodec.Text(result.Parts));
            Check(!result.IsError && parsed.RootElement.GetProperty("response").GetString() == "separate answer",
                "Ask did not return its reliable turn reply.");
            ChildOutput output = await child.ReadOutputAsync(cancellation: deadline.Token);
            Check(output.Output.Contains("older final") && output.Output.Contains("separate answer")
                && !output.Output.Contains("distinct question"), "Finalized output included input or lost a final.");
            await server;
        }
        finally
        {
            releaseFirst.TrySetResult(); releaseSecond.TrySetResult();
            app.Destroy(); listener.Stop();
        }
    }
    public static async Task ChildChangeWaits(string temp)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var app = new Application(TextWriter.Null);
        try
        {
            var parent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1"),
                dataDirectory: Path.Combine(temp, "change-waits"));
            parent.InstallTools();
            Agent child = await parent.CreateChildAsync("worker", cancellation: deadline.Token);
            Agent other = await parent.CreateChildAsync("other", cancellation: deadline.Token);
            Task<QuestionReply> question = other.SendMessageToParentAsync("other question", deadline.Token);
            Check((await parent.WaitForChildChangeAsync(["worker"], timeoutSeconds: .15,
                cancellation: deadline.Token)).Reason == "timeout", "Unselected question ended a wait.");
            ChildChange questions = await parent.WaitForChildChangeAsync(["other"], cancellation: deadline.Token);
            Check(questions.Reason == "question" && questions.Questions.Single().AgentID == "other",
                "Selected question was not reported.");
            parent.RespondToQuestion("answered", childID: "other");
            Check((await question).Response == "answered", "Observation consumed the question.");
            child.History().Append(new Event(EventKind.ToolObserved, callID: "match",
                parts: [new TextPart("needle") ]));
            Check((await parent.WaitForChildChangeAsync(["worker"], pattern: "needle",
                cancellation: deadline.Token)).Reason == "pattern", "Pattern did not short-circuit.");
            foreach (object invalid in new object[] {
                new { agent_ids = new[] { "worker", "worker" } },
                new { agent_ids = new[] { "worker" }, pattern = "[" },
                new { timeout_seconds = -1 }, new { stuck_seconds = -1 } })
            {
                Check((await ChildMessageResult(parent, "wait_for_agent_change", invalid, deadline.Token)).IsError,
                    "Invalid change wait did not refuse.");
            }
            // The first async call captures membership and queues a status query
            // before it returns. This explicit actor barrier precedes mutation.
            Task<ChildChange> changed = parent.WaitForChildChangeAsync(["worker"], cancellation: deadline.Token);
            await child.WorkStatusAsync(deadline.Token);
            child.History().Append(new Event(EventKind.ToolObserved, callID: "change",
                parts: [new TextPart("later") ]));
            Check((await changed).Reason == "changed", "History progress did not wake the observer.");
            Task<ChildChange> fixedSet = parent.WaitForChildChangeAsync(["worker"], timeoutSeconds: .2,
                cancellation: deadline.Token);
            await child.WorkStatusAsync(deadline.Token);
            _ = await parent.CreateChildAsync("later", cancellation: deadline.Token);
            Check((await fixedSet).Reason == "timeout", "A later child entered fixed membership.");
            await child.PostAsync(new AgentInput(InputKind.Pause));
            Task<string> paused = child.SubmitAsync([new TextPart("paused")]);
            await child.WorkStatusAsync(deadline.Token);
            ChildChange stuck = await parent.WaitForChildChangeAsync(["worker"], stuckSeconds: .15,
                cancellation: deadline.Token);
            Check(stuck.Reason == "stuck" && stuck.Agents.Single().Status.Paused && !paused.IsCompleted,
                "Stuck wait changed paused work or misreported progress.");
            using var cancelled = new CancellationTokenSource();
            Task<ChildChange> abandoned = parent.WaitForChildChangeAsync(["worker"], timeoutSeconds: 0,
                stuckSeconds: 0, cancellation: cancelled.Token);
            cancelled.Cancel();
            try { await abandoned; throw new Exception("Cancelled change waiter completed normally."); }
            catch (OperationCanceledException) { }
            Check(parent.FindChild("worker") == child && !paused.IsCompleted,
                "Wait cancellation cancelled the child.");
            Task<ChildChange> removed = parent.WaitForChildChangeAsync(["worker"], cancellation: deadline.Token);
            await child.WorkStatusAsync(deadline.Token);
            child.Destroy();
            Check((await removed).Reason is "removed" or "stopped", "Removed child wait did not finish.");
            await ((Task)paused).ConfigureAwait(ConfigureAwaitOptions.SuppressThrowing);
            ToolResultPart timeout = await ChildMessageResult(parent, "wait_for_agent_change",
                new { agent_ids = new[] { "other" }, timeout_seconds = .1 }, deadline.Token);
            using var result = JsonDocument.Parse(JsonCodec.Text(timeout.Parts));
            Check(!timeout.IsError && result.RootElement.GetProperty("reason").GetString() == "timeout",
                "Tool timeout became an execution error.");
        }
        finally { app.Destroy(); }
    }

    public static async Task ChildStreamingProgress(string temp)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        var sendDelta = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var sendFinal = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        string first = "data: {\"type\":\"message_start\",\"message\":{\"model\":\"claude-3-haiku-20240307\",\"usage\":{\"input_tokens\":3}}}\n\n"
            + "data: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n"
            + "data: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"first \"}}\n\n";
        string delta = "data: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"second\"}}\n\n";
        string final = "data: {\"type\":\"content_block_stop\",\"index\":0}\n\n"
            + "data: {\"type\":\"message_delta\",\"usage\":{\"output_tokens\":2}}\n\n"
            + "data: {\"type\":\"message_stop\"}\n\n";
        Task server = Task.Run(async () =>
        {
            using var client = await listener.AcceptTcpClientAsync(deadline.Token);
            using var stream = client.GetStream();
            await ReadRequest(stream, deadline.Token);
            string header = "HTTP/1.1 200 OK\r\nContent-Type: text/event-stream\r\nContent-Length: "
                + Encoding.UTF8.GetByteCount(first + delta + final) + "\r\nConnection: close\r\n\r\n";
            await stream.WriteAsync(Encoding.UTF8.GetBytes(header + first), deadline.Token);
            await stream.FlushAsync(deadline.Token);
            await sendDelta.Task.WaitAsync(deadline.Token);
            await stream.WriteAsync(Encoding.UTF8.GetBytes(delta), deadline.Token);
            await stream.FlushAsync(deadline.Token);
            await sendFinal.Task.WaitAsync(deadline.Token);
            await stream.WriteAsync(Encoding.UTF8.GetBytes(final), deadline.Token);
        });
        var app = new Application(TextWriter.Null);
        try
        {
            var parent = new Agent(app, new Configuration(Vendor.Anthropic, "claude-3-haiku-20240307",
                "http://127.0.0.1:" + ((IPEndPoint)listener.LocalEndpoint).Port),
                dataDirectory: Path.Combine(temp, "child-stream-progress"));
            Agent child = await parent.CreateChildAsync("worker", cancellation: deadline.Token);
            Task<string> turn = child.SubmitAsync([new TextPart("stream")]);
            await child.WaitAsync(value => value.Delta?.Text == "first ", deadline.Token);
            ChildOutput before = await child.ReadOutputAsync(cancellation: deadline.Token);
            Check(before.Output == "", "Partial output became finalized text.");
            Task<ChildChange> change = parent.WaitForChildChangeAsync(["worker"], cancellation: deadline.Token);
            await child.WorkStatusAsync(deadline.Token);
            sendDelta.TrySetResult();
            ChildChange observed = await change;
            Check(observed.Reason == "changed" && observed.Agents.Single().Cursor == before.Cursor
                && observed.Agents.Single().Output == "" && !turn.IsCompleted,
                "Delta progress was lost, finalized prematurely, or changed durable History.");
            // The progress observer must not steal the GUI's delta observation.
            await child.WaitAsync(value => value.Delta?.Text == "second", deadline.Token);
            child.PublishSubmission(JsonSerializer.SerializeToElement(new { result = "done" }));
            sendFinal.TrySetResult();
            Check(await turn.WaitAsync(deadline.Token) == "first second", "Stream final changed.");
            Check((await child.ReadOutputAsync(before.Cursor, cancellation: deadline.Token)).Output == "first second\n",
                "Finalized stream text was not cursor-readable exactly once.");
            await server;
        }
        finally
        {
            sendDelta.TrySetResult(); sendFinal.TrySetResult();
            app.Destroy(); listener.Stop();
        }
    }

}
