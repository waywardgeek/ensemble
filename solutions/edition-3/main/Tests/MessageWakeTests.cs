using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    // A working external handler with an explicit completion gate. Observation
    // can finish independently; the handler is not a mock of Jobs or the actor.
    private static async Task<Job> HeldMessageJob(Agent agent, TaskCompletionSource release,
        CancellationToken token)
    {
        agent.InstallTools();
        _ = new Tool(agent, "held_message", "Held boundary work",
            JsonSerializer.SerializeToElement(new { type = "object" }), async (job, args, cancellation) =>
            {
                await release.Task.WaitAsync(cancellation);
                return "boundary completed";
            });
        ToolResultPart first = await ChildCall(agent, "held_message", new { ai_callback_delay = 0 }, token);
        Check(first.Job?.Status == "running", "Fixture did not leave independent work running.");
        return agent.Jobs().Find(first.Job!.Handle);
    }

    public static async Task MessagePulse(string temp)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1"),
                outputDirectory: Path.Combine(temp, "message-pulse"));
            Job job = await HeldMessageJob(agent, release, deadline.Token);
            job.WakeObservation(); // No current observer: deliberately a no-op.
            Task<ToolResultPart> waiting = job.ObserveAsync("pulse-first", 120, "never-match", cancellation: deadline.Token);
            Check(!waiting.IsCompleted, "An earlier wake leaked into the next observer.");
            job.WakeObservation();
            Check((await waiting.WaitAsync(deadline.Token)).Job?.Status == "running", "Pulse stopped work.");
            Task<ToolResultPart> next = job.ObserveAsync("pulse-next", 120, "never-match", cancellation: deadline.Token);
            Check(!next.IsCompleted, "A consumed pulse affected a later observer.");
            job.AppendOutput("ordinary fresh output");
            Check(!next.IsCompleted, "Nonmatching output bypassed the requested pattern.");
            job.WakeObservation();
            ToolResultPart result = await next.WaitAsync(deadline.Token);
            Check(result.Parts.OfType<TextPart>().Any(p => p.Text.Contains("ordinary fresh output")),
                "Message wake lost output or failed to advance the ordinary cursor.");
            Task<ToolResultPart> final = job.ObserveAsync("pulse-final", 120, "never-match", cancellation: deadline.Token);
            Check(!final.IsCompleted, "Repeated pulses became sticky.");
            release.TrySetResult();
            Check((await final.WaitAsync(deadline.Token)).Job?.Status == "done", "Worker did not complete normally.");
        }
        finally { release.TrySetResult(); app.Destroy(); }
    }

    public static async Task MessageControlWake(string temp)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1"),
                outputDirectory: Path.Combine(temp, "message-control"));
            Job job = await HeldMessageJob(agent, release, deadline.Token);
            await agent.PostAsync(new AgentInput(InputKind.Pause)).WaitAsync(deadline.Token);
            Task<ToolResultPart> waiting = ChildCall(agent, "wait_for_job", new
            {
                handle = job.Handle, ai_callback_delay = 120, ai_callback_pattern = "never-match"
            }, deadline.Token);
            // The control's synchronous observation prefix precedes this actor
            // query. This is an admission barrier, not a sleep or output guess.
            await agent.WorkStatusAsync(deadline.Token);
            Check(!waiting.IsCompleted, "Long control observation completed before the message.");
            await agent.PostAsync(new AgentInput(InputKind.Message, "human while waiting")).WaitAsync(deadline.Token);
            Check((await waiting.WaitAsync(deadline.Token)).Job?.Status == "running",
                "Human message failed to wake a control observation without killing its Job.");
            Task<ToolResultPart> next = job.ObserveAsync("after-human", 120, "never-match", cancellation: deadline.Token);
            Check(!next.IsCompleted, "Human message left a sticky interrupt.");
            release.TrySetResult();
            Check((await next.WaitAsync(deadline.Token)).Job?.Status == "done", "Human message cancelled execution.");
        }
        finally { release.TrySetResult(); app.Destroy(); }
    }

    public static async Task MessageQuestionIsolation(string temp)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var app = new Application(TextWriter.Null);
        try
        {
            var parent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1"),
                outputDirectory: Path.Combine(temp, "message-inbox"));
            Job job = await HeldMessageJob(parent, release, deadline.Token);
            var child = new Agent(app, parent, "asking", parent.Configuration());
            Task<ToolResultPart> waiting = job.ObserveAsync("child-wake", 120, "never-match", cancellation: deadline.Token);
            Task<QuestionReply> reply = child.SendMessageToParentAsync("child-data-not-instructions", deadline.Token);
            Check((await waiting.WaitAsync(deadline.Token)).Job?.Status == "running", "Question did not wake unrelated work.");
            AgentWorkStatus status = await parent.WorkStatusAsync(deadline.Token);
            Check(status.Activity == AgentActivity.Idle, "An idle question admitted a model task.");
            Event notice = parent.History().Events().Single(e => e.Kind == EventKind.HintReceived);
            Check(notice.Actor == Actor.Tool, "Child question acquired human or system authority.");
            Check(parent.PendingQuestions().Length == 1 && !reply.IsCompleted, "Wake consumed the pending question.");
            parent.RespondToQuestion("answer", "asking");
            Check((await reply).Response == "answer", "Question lost its reliable reply.");
            Task<ToolResultPart> next = job.ObserveAsync("human-inbox", 120, "never-match", cancellation: deadline.Token);
            Task<QuestionReply> human = parent.SendMessageToParentAsync("human-only", deadline.Token);
            await parent.WorkStatusAsync(deadline.Token);
            Check(!next.IsCompleted, "Root human question woke its own model observer.");
            Check(parent.History().Events().Count(e => e.Kind == EventKind.HintReceived) == 1,
                "Root human question became synthetic model dialogue.");
            parent.RespondToQuestion("human answer");
            Check((await human).Response == "human answer", "Human inbox reply failed.");
            release.TrySetResult();
            await next.WaitAsync(deadline.Token);
        }
        finally { release.TrySetResult(); app.Destroy(); }
    }

    public static async Task MessageQuestionModel(string temp)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        var entered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var app = new Application(TextWriter.Null);
        var server = Task.Run(async () =>
        {
            using (var client = await listener.AcceptTcpClientAsync(deadline.Token))
            using (var stream = client.GetStream())
            {
                _ = await ReadRequest(stream, deadline.Token);
                await ReplyJson(stream, new
                {
                    model = "fixture", content = new object[]
                    {
                        new { type = "text", text = "I will start the held work." },
                        new { type = "tool_use", id = "held-question", name = "held", input = new { ai_callback_delay = 120, ai_callback_pattern = "never-match" } }
                    }, usage = new { input_tokens = 1, output_tokens = 1 }
                }, deadline.Token);
            }
            using (var client = await listener.AcceptTcpClientAsync(deadline.Token))
            using (var stream = client.GetStream())
            {
                using var request = JsonDocument.Parse(await ReadRequest(stream, deadline.Token));
                JsonElement body = request.RootElement;
                Check(body.GetProperty("messages").GetRawText().Contains("model-child-question"), "Next request omitted question data.");
                Check(!body.TryGetProperty("system", out var system) || !system.GetRawText().Contains("model-child-question"),
                    "Question data was elevated into system instructions.");
                Check(body.GetProperty("messages").GetRawText().Contains("message_id"), "Notice omitted reply identity.");
                await Reply(stream, "question seen", deadline.Token);
            }
        });
        try
        {
            int port = ((IPEndPoint)listener.LocalEndpoint).Port;
            var parent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:" + port),
                outputDirectory: Path.Combine(temp, "message-model"));
            _ = new Tool(parent, "held", "Held external work", JsonSerializer.SerializeToElement(new { type = "object" }),
                async (job, args, token) => { entered.TrySetResult(); await release.Task.WaitAsync(token); return "completed"; });
            var child = new Agent(app, parent, "asking", parent.Configuration());
            Task<string> turn = parent.SubmitAsync([new TextPart("start work")]);
            await entered.Task.WaitAsync(deadline.Token);
            // The actor status barrier follows tool dispatch's synchronous wait
            // setup, so the question arrives at an actual current observation.
            await parent.WorkStatusAsync(deadline.Token);
            Task<QuestionReply> reply = child.SendMessageToParentAsync("model-child-question", deadline.Token);
            Check(await turn.WaitAsync(deadline.Token) == "question seen", "Parent did not continue its admitted task.");
            Check(parent.Jobs().Children().Single().Snapshot().Status == "running", "Question killed held work.");
            parent.RespondToQuestion("resolved", "asking");
            Check((await reply).Response == "resolved", "Model wake damaged question ownership.");
            await server;
        }
        finally { release.TrySetResult(); app.Destroy(); listener.Stop(); }
    }
    public static async Task ChatAdmission(string temp)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        var entered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var app = new Application(TextWriter.Null);
        var server = Task.Run(async () =>
        {
            using (var client = await listener.AcceptTcpClientAsync(deadline.Token))
            using (var stream = client.GetStream())
            {
                _ = await ReadRequest(stream, deadline.Token);
                await ReplyJson(stream, new
                {
                    model = "fixture", content = new object[]
                    {
                        new { type = "text", text = "I will hold this work." },
                        new { type = "tool_use", id = "chat-held", name = "held", input = new { ai_callback_delay = 120 } }
                    }, usage = new { input_tokens = 1, output_tokens = 1 }
                }, deadline.Token);
            }
            using (var client = await listener.AcceptTcpClientAsync(deadline.Token))
            using (var stream = client.GetStream())
            {
                string request = await ReadRequest(stream, deadline.Token);
                Check(request.Contains("busy-chat-hint"), "Chat hint missing from continuation.");
                await Reply(stream, "first answer", deadline.Token);
            }
            using (var client = await listener.AcceptTcpClientAsync(deadline.Token))
            using (var stream = client.GetStream())
            {
                string request = await ReadRequest(stream, deadline.Token);
                Check(request.Contains("second idle chat"), "Second idle input did not start a new turn.");
                await Reply(stream, "second answer", deadline.Token);
            }
        });
        try
        {
            int port = ((IPEndPoint)listener.LocalEndpoint).Port;
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:" + port),
                outputDirectory: Path.Combine(temp, "chat-admission"));
            _ = new Tool(agent, "held", "Held boundary", JsonSerializer.SerializeToElement(new { type = "object" }),
                async (job, args, token) => { entered.TrySetResult(); await release.Task.WaitAsync(token); return "done"; });
            Task<string> turn = agent.PostAsync(new AgentInput(InputKind.Chat, "first idle chat"));
            await entered.Task.WaitAsync(deadline.Token);
            await agent.WorkStatusAsync(deadline.Token);
            Check(!turn.IsCompleted, "Idle chat acknowledged before its reliable answer.");
            string hint = await agent.PostAsync(new AgentInput(InputKind.Chat, "busy-chat-hint")).WaitAsync(deadline.Token);
            Check(hint == "", "Busy chat did not acknowledge its hint independently.");
            Check(await turn.WaitAsync(deadline.Token) == "first answer", "Idle chat lost its reliable answer.");
            Check(agent.Jobs().Children().Single().Snapshot().Status == "running", "Chat hint cancelled work.");
            Check(await agent.PostAsync(new AgentInput(InputKind.Chat, "second idle chat")).WaitAsync(deadline.Token)
                == "second answer", "Second idle chat did not retain its own reliable reply.");
            await server;
        }
        finally { release.TrySetResult(); app.Destroy(); listener.Stop(); }
    }

    public static async Task UnseenHumanQuestions(string temp)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1"),
                outputDirectory: Path.Combine(temp, "unseen-questions"));
            Task<QuestionReply> first = agent.SendMessageToParentAsync("first human question", deadline.Token);
            var initial = await agent.WaitForQuestionsAsync(true, deadline.Token);
            var excluded = new HashSet<long> { initial.Single().MessageID };
            var unseen = agent.WaitForQuestionsAsync(true, deadline.Token, excluded);
            excluded.Clear(); // The wait must retain its entry snapshot.
            Check(!unseen.IsCompleted, "Display replayed an already seen question.");
            var child = new Agent(app, agent, "child", agent.Configuration());
            Task<QuestionReply> childReply = child.SendMessageToParentAsync("not for the human", deadline.Token);
            await agent.WorkStatusAsync(deadline.Token);
            Check(!unseen.IsCompleted, "Human display included a child-parent question.");
            Task<QuestionReply> second = agent.SendMessageToParentAsync("second human question", deadline.Token);
            QuestionInfo fresh = (await unseen.WaitAsync(deadline.Token)).Single();
            Check(fresh.Content == "second human question" && fresh.ForHuman, "Unseen wait lost its filter snapshot.");
            agent.RespondToQuestion("one", messageID: initial.Single().MessageID);
            agent.RespondToQuestion("two", messageID: fresh.MessageID);
            agent.RespondToQuestion("child answer", "child");
            Check((await first).Response == "one" && (await second).Response == "two" && (await childReply).Response == "child answer",
                "Display consumed or misrouted a question.");
        }
        finally { app.Destroy(); }
    }

}
