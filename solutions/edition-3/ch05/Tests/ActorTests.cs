using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

// Chapter 5 tests drive the real Agent and transport. This tiny HTTP server is
// the only fake: it accepts complete requests and returns valid provider JSON.
internal static partial class ActorTests
{
    public static async Task TransportIsolation()
    {
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(10));
        var server = Task.Run(async () =>
        {
            using var client = await listener.AcceptTcpClientAsync(deadline.Token);
            using var stream = client.GetStream();
            var request = await ReadRequest(stream, deadline.Token);
            Check(request.Contains("FROZEN_INPUT"), "The admitted input is missing.");
            Check(!request.Contains("LATER_HINT"), "A later hint changed the frozen request.");
            await Reply(stream, "transport only", deadline.Token);
        });
        var app = new Application(TextWriter.Null);
        try
        {
            var port = ((IPEndPoint)listener.LocalEndpoint).Port;
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "model-a",
                "http://127.0.0.1:" + port, "dummy-local-only"));
            agent.History().Append(new Event(EventKind.MessageReceived,
                actor: Actor.Human, parts: [new TextPart("FROZEN_INPUT")]));
            var request = agent.Engine().Render(agent.History().View()).ToJsonString();
            agent.SetEphemeral("LATER_HINT");
            var before = agent.History().Events();
            var received = await agent.Engine().SendAsync(request, deadline.Token);
            await server;
            Check(received.Any(e => e.Kind == EventKind.ResponseEnded), "No response fact was returned.");
            Check(agent.History().Events().SequenceEqual(before), "Transport mutated canonical History.");
        }
        finally
        {
            app.Destroy();
            listener.Stop();
        }
    }

    public static async Task MailboxControls()
    {
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var received = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        int requestCount = 0;
        var server = Task.Run(async () =>
        {
            for (int index = 0; index < 3; index++)
            {
                using var client = await listener.AcceptTcpClientAsync(deadline.Token);
                using var stream = client.GetStream();
                string body = await ReadRequest(stream, deadline.Token);
                Interlocked.Increment(ref requestCount);
                if (index == 0)
                {
                    Check(body.Contains("FIRST"), "First request missing.");
                    received.SetResult();
                    await release.Task.WaitAsync(deadline.Token);
                    await Reply(stream, "initial response", deadline.Token);
                }
                else if (index == 1)
                {
                    Check(body.Contains("LIVE_HINT"), "Hint missing from continuation.");
                    Check(!body.Contains("SECOND"), "Queued request leaked into active turn.");
                    await Reply(stream, "hint addressed", deadline.Token);
                }
                else
                {
                    Check(body.Contains("SECOND"), "Second request never ran.");
                    await Reply(stream, "second answer", deadline.Token);
                }
            }
        });
        var app = new Application(TextWriter.Null);
        try
        {
            var port = ((IPEndPoint)listener.LocalEndpoint).Port;
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "model-a",
                "http://127.0.0.1:" + port, "dummy-local-only"));
            using var detached = new CancellationTokenSource();
            var firstWaiter = agent.SendAsync("FIRST", detached.Token);
            await received.Task.WaitAsync(deadline.Token);
            detached.Cancel();
            try
            {
                await firstWaiter;
                throw new InvalidOperationException("Cancelled waiter unexpectedly completed.");
            }
            catch (OperationCanceledException) { }
            await agent.PostAsync(new AgentInput(InputKind.Hint, "LIVE_HINT")).WaitAsync(deadline.Token);
            await agent.PostAsync(new AgentInput(InputKind.Pause)).WaitAsync(deadline.Token);
            var second = agent.SubmitAsync([new TextPart("SECOND")]);
            release.SetResult();
            await agent.WaitAsync(o => o.Part is TextPart { Text: "initial response" }, deadline.Token);
            // The acknowledged pause precedes the response. No timing sleep is
            // needed: completion publication occurs before Advance can schedule.
            Check(Volatile.Read(ref requestCount) == 1, "Pause allowed another HTTP request.");
            Check(!second.IsCompleted, "Queued request completed while paused.");
            await agent.PostAsync(new AgentInput(InputKind.Resume));
            Check(await second.WaitAsync(deadline.Token) == "second answer", "Wrong reliable reply.");
            await server;
            var events = agent.History().Events();
            Check(events.Count(e => e.Kind == EventKind.TurnEnded) == 2, "Caller cancellation ended a turn.");
            Check(events.Any(e => e.Kind == EventKind.HintReceived), "Hint was not recorded.");
            var observation = await app.WaitAnyAsync(o => o.Observation.Kind == ObservationKind.TurnEnded,
                deadline.Token);
            Check(ReferenceEquals(observation.Agent, agent), "Application observation lost identity.");
            // Replay contains the same terminal state, including the new facts.
            var replay = new Application(TextWriter.Null);
            try
            {
                var restored = new Agent(replay, agent.Configuration());
                foreach (var fact in events)
                {
                    restored.History().Append(JsonCodec.ReadEvent(JsonCodec.WriteEvent(fact)));
                }
                Check(restored.History().View().Phase == Phase.Idle, "Replay did not finish idle.");
            }
            finally
            {
                replay.Destroy();
            }
        }
        finally
        {
            release.TrySetResult();
            app.Destroy();
            listener.Stop();
        }
    }

    public static Task BusyHint() => BusyToolControl(InputKind.Hint);
    public static Task ImmediateInterrupt() => BusyToolControl(InputKind.InterruptNow);
    public static Task BoundaryInterrupt() => BusyToolControl(InputKind.InterruptAfterTool);

    private static async Task BusyToolControl(InputKind kind)
    {
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var entered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        int forbiddenEffects = 0;
        string continuation = "";
        var server = Task.Run(async () =>
        {
            using (var client = await listener.AcceptTcpClientAsync(deadline.Token))
            using (var stream = client.GetStream())
            {
                _ = await ReadRequest(stream, deadline.Token);
                await ReplyJson(stream, new
                {
                    model = "model-a",
                    content = new object[]
                    {
                        new { type = "tool_use", id = "slow-id", name = "slow", input = new { ai_callback_delay = 10 } },
                        new { type = "tool_use", id = "tail-id", name = "forbidden", input = new { } }
                    },
                    usage = new { input_tokens = 1, output_tokens = 1 }
                }, deadline.Token);
            }
            using (var client = await listener.AcceptTcpClientAsync(deadline.Token))
            using (var stream = client.GetStream())
            {
                continuation = await ReadRequest(stream, deadline.Token);
                await Reply(stream, "hint handled", deadline.Token);
            }
        });
        var app = new Application(TextWriter.Null);
        try
        {
            int port = ((IPEndPoint)listener.LocalEndpoint).Port;
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "model-a", "http://127.0.0.1:" + port));
            var schema = JsonSerializer.SerializeToElement(new { type = "object", properties = new { } });
            _ = new Tool(agent, "slow", "Held external work", schema, async (job, args, cancellation) =>
            {
                entered.TrySetResult();
                await release.Task.WaitAsync(cancellation);
                return "released";
            });
            _ = new Tool(agent, "forbidden", "Must never execute", schema, (job, args, cancellation) =>
            {
                Interlocked.Increment(ref forbiddenEffects);
                return Task.FromResult("bad effect");
            });
            var turn = agent.SubmitAsync([new TextPart("do work")]);
            await entered.Task.WaitAsync(deadline.Token);
            await agent.PostAsync(new AgentInput(kind, "BUSY_HINT")).WaitAsync(deadline.Token);
            if (kind == InputKind.InterruptAfterTool)
            {
                Check(!turn.IsCompleted, "Boundary interrupt stopped before the current tool settled.");
                release.TrySetResult();
            }
            if (kind != InputKind.Hint)
            {
                try
                {
                    await turn.WaitAsync(deadline.Token);
                    throw new InvalidOperationException("Interrupted request returned success.");
                }
                catch (OperationCanceledException) when (!deadline.IsCancellationRequested) { }
                turn = agent.SubmitAsync([new TextPart("continue after interruption")]);
            }
            Check(await turn.WaitAsync(deadline.Token) == "hint handled", "Actor did not continue after control.");
            await server;
            foreach (var fact in agent.History().Events().Where(e => e.Kind is
                EventKind.HintReceived or EventKind.TurnInterrupted or EventKind.TurnEnded))
            {
                var restored = JsonCodec.ReadEvent(JsonCodec.WriteEvent(fact));
                Check(restored.Kind == fact.Kind && restored.Seq == fact.Seq,
                    "Actor control fact did not round-trip through the public log codec.");
            }

            Check(forbiddenEffects == 0, "Undispatched batch tail ran.");
            var job = agent.Jobs().Children().Single();
            Check(job.Snapshot().Status != "killed", "Control killed independent work.");
            if (kind != InputKind.InterruptAfterTool)
            {
                Check(job.Snapshot().Status == "running", "Foreground cancellation waited for the worker.");
            }
            var facts = agent.History().Events();
            var result = facts.Single(e => e.Kind == EventKind.ToolReturned && e.CallID == "slow-id");
            Check(result.Job?.Handle == job.Handle && result.Job.Status != "killed", "Missing real job snapshot.");
            var skipped = facts.Single(e => e.Kind == EventKind.ToolReturned && e.CallID == "tail-id");
            Check(skipped.IsError && JsonCodec.Text(skipped.Parts).Contains("not executed"), "Tail was not paired honestly.");
            Check((kind != InputKind.Hint || continuation.Contains("BUSY_HINT"))
                && continuation.Contains("slow-id") && continuation.Contains("tail-id"),
                "Continuation omitted hint or correlated results.");
            release.TrySetResult();
        }
        finally
        {
            release.TrySetResult();
            app.Destroy();
            listener.Stop();
        }
    }

    public static async Task HttpInterrupts()
    {
        foreach (var kind in new[] { InputKind.InterruptNow, InputKind.InterruptAfterTool })
        {
            using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(10));
            var listener = new TcpListener(IPAddress.Loopback, 0);
            listener.Start();
            var received = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            var server = Task.Run(async () =>
            {
                using (var client = await listener.AcceptTcpClientAsync(deadline.Token))
                using (var stream = client.GetStream())
                {
                    _ = await ReadRequest(stream, deadline.Token);
                    received.TrySetResult();
                    await release.Task.WaitAsync(deadline.Token);
                    try
                    {
                        await ReplyJson(stream, new
                        {
                            model = "model-a",
                            content = new[] { new { type = "tool_use", id = "late", name = "effect", input = new { } } },
                            usage = new { input_tokens = 1, output_tokens = 1 }
                        }, deadline.Token);
                    }
                    catch (IOException) when (kind == InputKind.InterruptNow)
                    {
                        // The real client may have already closed the socket.
                    }
                }
                using (var client = await listener.AcceptTcpClientAsync(deadline.Token))
                using (var stream = client.GetStream())
                {
                    _ = await ReadRequest(stream, deadline.Token);
                    await Reply(stream, "fresh reply", deadline.Token);
                }
            });
            var app = new Application(TextWriter.Null);
            try
            {
                int port = ((IPEndPoint)listener.LocalEndpoint).Port;
                var agent = new Agent(app, new Configuration(Vendor.Anthropic, "model-a", "http://127.0.0.1:" + port));
                int effects = 0;
                _ = new Tool(agent, "effect", "Must not run", JsonSerializer.SerializeToElement(new { type = "object" }),
                    (job, args, token) =>
                    {
                        Interlocked.Increment(ref effects);
                        return Task.FromResult("bad effect");
                    });
                var turn = agent.SubmitAsync([new TextPart("first")]);
                var firstOutcome = await Task.WhenAny(received.Task, turn).WaitAsync(deadline.Token);
                if (ReferenceEquals(firstOutcome, turn))
                {
                    await turn;
                    throw new InvalidOperationException("Turn ended before HTTP admission.");
                }
                await received.Task.WaitAsync(deadline.Token);
                await agent.PostAsync(new AgentInput(kind)).WaitAsync(deadline.Token);
                release.TrySetResult();
                try
                {
                    await turn.WaitAsync(deadline.Token);
                    throw new InvalidOperationException("HTTP interrupt returned success.");
                }
                catch (OperationCanceledException) when (!deadline.IsCancellationRequested) { }
                Check(await agent.SubmitAsync([new TextPart("fresh")]).WaitAsync(deadline.Token) == "fresh reply",
                    "Late HTTP completion resurrected or blocked the old turn.");
                await server;
                Check(effects == 0 && !agent.History().Events().Any(e => e.Kind == EventKind.ToolCalled),
                    "Interrupted response caused a late effect.");
                if (kind == InputKind.InterruptAfterTool)
                {
                    Check(agent.History().Events().Any(e => e.Kind == EventKind.ToolReturned && e.CallID == "late" && e.IsError),
                        "Retained boundary response lost its skipped-call result.");
                }
                Check(agent.History().View().Phase == Phase.Idle, "Interrupt left reducer in flight.");
            }
            finally
            {
                release.TrySetResult();
                app.Destroy();
                listener.Stop();
            }
        }
    }

    public static async Task DirectHint()
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(10));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        var server = Task.Run(async () =>
        {
            using var client = await listener.AcceptTcpClientAsync(deadline.Token);
            using var stream = client.GetStream();
            string request = await ReadRequest(stream, deadline.Token);
            Check(request.Contains("direct hint"), "Direct-call hint did not reach the next model request.");
            await Reply(stream, "hint answered", deadline.Token);
        });
        var started = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var app = new Application(TextWriter.Null);
        try
        {
            int port = ((IPEndPoint)listener.LocalEndpoint).Port;
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "model-a", "http://127.0.0.1:" + port));
            _ = new Tool(agent, "held", "Held direct job", JsonSerializer.SerializeToElement(new { type = "object" }),
                async (job, args, token) =>
                {
                    started.TrySetResult();
                    await release.Task.WaitAsync(token);
                    return "job settled";
                });
            var call = new ToolCallPart("direct", agent.Configuration().Target(), "held",
                JsonSerializer.SerializeToElement(new { ai_callback_delay = 10 }));
            var result = agent.DispatchAsync(call);
            await started.Task.WaitAsync(deadline.Token);
            await agent.PostAsync(new AgentInput(InputKind.Hint, "direct hint")).WaitAsync(deadline.Token);
            Check((await result.WaitAsync(deadline.Token)).Job?.Status == "running",
                "Direct hint stopped the independent job.");
            var ended = await agent.WaitAsync(e => e.Kind == ObservationKind.TurnEnded && e.Text == "hint answered",
                deadline.Token);
            Check(ended.Text == "hint answered", "Hint trigger was lost at direct reply completion.");
            var fact = agent.History().Events().Single(e => e.Kind == EventKind.ToolObserved);
            var restored = JsonCodec.ReadEvent(JsonCodec.WriteEvent(fact));
            Check(restored.CallID == call.CallID && restored.Job?.Handle == fact.Job?.Handle,
                "Host-only outcome lost its audit identity or locator.");
            Check(!agent.History().View().Dialogue.Any(entry => entry.Parts.OfType<ToolResultPart>().Any()),
                "Host-only result became an orphan model result.");
            await server;
        }
        finally
        {
            release.TrySetResult();
            app.Destroy();
            listener.Stop();
        }
    }

    public static async Task ShutdownJoins()
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(10));
        var cancelled = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var started = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        bool exited = false;
        var app = new Application(TextWriter.Null);
        var agent = new Agent(app, new Configuration(Vendor.Anthropic, "model-a", "http://127.0.0.1"));
        try
        {
            agent.InstallTools();
            var schema = JsonSerializer.SerializeToElement(new { type = "object", properties = new { } });
            _ = new Tool(agent, "held", "Cooperative shutdown probe", schema, async (job, args, token) =>
            {
                started.TrySetResult();
                try
                {
                    await Task.Delay(Timeout.Infinite, token);
                }
                catch (OperationCanceledException) when (token.IsCancellationRequested)
                {
                    cancelled.TrySetResult();
                    await release.Task;
                }
                Check(ReferenceEquals(job.Parent().Parent(), agent), "Detached while callback was alive.");
                exited = true;
                return "settled";
            });
            var call = new ToolCallPart("shutdown-probe", agent.Configuration().Target(), "held",
                JsonSerializer.SerializeToElement(new { ai_callback_delay = 10 }));
            var result = agent.DispatchAsync(call);
            await started.Task.WaitAsync(deadline.Token);
            var jobs = agent.Jobs();
            var child = jobs.Children().Single();
            await agent.PostAsync(new AgentInput(InputKind.Pause));
            var limits = new ToolCallPart("live-control", agent.Configuration().Target(), "tool_limits",
                JsonSerializer.SerializeToElement(new { ai_callback_delay = 0 }));
            var controlResult = await agent.DispatchAsync(limits).WaitAsync(deadline.Token);
            Check(!controlResult.IsError && jobs.Children().Length == 1,
                "Paused foreground turn blocked a handle-free control or allocated its job.");
            Task shutdown = agent.ShutdownAsync();
            await cancelled.Task.WaitAsync(deadline.Token);
            Check(!shutdown.IsCompleted, "Shutdown did not join the managed worker.");
            release.TrySetResult();
            await shutdown.WaitAsync(deadline.Token);
            Check(exited && child.parent is null && jobs.parent is null, "Shutdown left worker or ownership alive.");
            Check(agent.History().Events().Any(e => e.Kind == EventKind.JobKilled && e.Reason == "shutdown"),
                "Shutdown reason was lost before History teardown.");
            // A direct call may return its real running snapshot before shutdown
            // kills the independent job. In either case it must settle once.
            await result.WaitAsync(deadline.Token);
            bool closed = false;
            try { _ = agent.SubmitAsync([new TextPart("too late")]); }
            catch (ObjectDisposedException) { closed = true; }
            Check(closed, "Shutdown reopened admission.");
            await agent.ShutdownAsync();
        }
        finally
        {
            release.TrySetResult();
            app.Destroy();
        }
    }

    private static void Check(bool condition, string message)
    {
        if (!condition)
        {
            throw new InvalidOperationException(message);
        }
    }

    private static async Task<string> ReadRequest(NetworkStream stream, CancellationToken token)
    {
        using var headers = new MemoryStream();
        var one = new byte[1];
        while (true)
        {
            await stream.ReadExactlyAsync(one, token);
            headers.WriteByte(one[0]);
            if (headers.Length > 65536)
            {
                throw new InvalidDataException("Oversized HTTP header.");
            }
            var bytes = headers.GetBuffer();
            var count = (int)headers.Length;
            if (count >= 4 && bytes[count - 4] == 13 && bytes[count - 3] == 10
                && bytes[count - 2] == 13 && bytes[count - 1] == 10)
            {
                break;
            }
        }
        var lines = Encoding.ASCII.GetString(headers.ToArray()).Split("\r\n");
        var length = int.Parse(lines.Single(line => line.StartsWith("Content-Length:",
            StringComparison.OrdinalIgnoreCase)).Split(':', 2)[1].Trim());
        if (length < 0 || length > 1048576)
        {
            throw new InvalidDataException("Oversized HTTP body.");
        }
        var body = new byte[length];
        await stream.ReadExactlyAsync(body, token);
        using var parsed = JsonDocument.Parse(body);
        return Encoding.UTF8.GetString(body);
    }

    private static async Task Reply(NetworkStream stream, string text, CancellationToken token)
    {
        await ReplyJson(stream, new
        {
            model = "model-a",
            content = new[] { new { type = "text", text } },
            usage = new { input_tokens = 1, output_tokens = 1 }
        }, token);
    }

    private static async Task ReplyJson(NetworkStream stream, object response, CancellationToken token)
    {
        var body = JsonSerializer.SerializeToUtf8Bytes(response);
        var header = Encoding.ASCII.GetBytes($"HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: {body.Length}\r\nConnection: close\r\n\r\n");
        await stream.WriteAsync(header, token);
        await stream.WriteAsync(body, token);
    }
}
