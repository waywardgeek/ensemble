using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    private static Task<ToolResultPart> ChildCall(Agent owner, string name, object args, CancellationToken token)
        => owner.DispatchAsync(new ToolCallPart(Guid.NewGuid().ToString("N"), owner.Configuration().Target(),
            name, JsonSerializer.SerializeToElement(args)), token);

    public static async Task ChildSpawnLifecycle(string temp)
    {
        string work = Path.Combine(temp, "spawn-lifecycle");
        Directory.CreateDirectory(work);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        var entered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var releaseSubmit = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var draining = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var releaseFinal = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var server = Task.Run(async () =>
        {
            using (var socket = await listener.AcceptTcpClientAsync(deadline.Token))
            using (var stream = socket.GetStream())
            {
                using var request = JsonDocument.Parse(await ReadRequest(stream, deadline.Token));
                Check(request.RootElement.GetProperty("model").GetString() == "child-model", "Spawn ignored model override.");
                var submit = request.RootElement.GetProperty("tools").EnumerateArray()
                    .Single(tool => tool.GetProperty("name").GetString() == "submit");
                Check(submit.GetProperty("input_schema").GetProperty("properties").TryGetProperty("count", out _),
                    "Spawn lost its typed submission contract.");
                entered.SetResult();
                await releaseSubmit.Task.WaitAsync(deadline.Token);
                await ReplyJson(stream, new
                {
                    model = "child-model", content = new object[] {
                        new { type = "text", text = "I will submit the checked result." },
                        new { type = "tool_use", id = "spawn-submit", name = "submit", input = new { count = 9 } }
                    }, usage = new { input_tokens = 1, output_tokens = 1 }
                }, deadline.Token);
            }
            using (var socket = await listener.AcceptTcpClientAsync(deadline.Token))
            using (var stream = socket.GetStream())
            {
                string request = await ReadRequest(stream, deadline.Token);
                Check(request.Contains("spawn-submit") && request.Contains("tool_result"), "Spawned task lost its tool result.");
                draining.SetResult();
                await releaseFinal.Task.WaitAsync(deadline.Token);
                await Reply(stream, "Submitted.", deadline.Token);
            }
        });
        var app = new Application(TextWriter.Null);
        try
        {
            var parent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture",
                "http://127.0.0.1:" + ((IPEndPoint)listener.LocalEndpoint).Port, "local", sandboxRoot: work), dataDirectory: work);
            parent.InstallTools();
            parent.StartPersistence(new PersistenceOptions(Path.Combine(work, "save.json")));
            var schema = JsonSerializer.Deserialize<JsonElement>("""
                {"type":"object","properties":{"count":{"type":"integer"}},"required":["count"],"additionalProperties":false}
                """);
            ToolResultPart early = await ChildCall(parent, "spawn_sub_agent", new {
                agent_id = "worker", task = "CHILD_ONLY_TASK", model = "child-model", submit_schema = schema,
                ai_callback_delay = 120, ai_callback_pattern = "never-match-this"
            }, deadline.Token);
            Check(!early.IsError && early.Job?.Status == "running" && JsonCodec.Text(early.Parts).Contains("worker"),
                "Spawn did not urgently yield its child ID while retaining the task.");
            Agent child = parent.FindChild("worker")!;
            Job spawn = parent.Jobs().Find(early.Job!.Handle);
            await entered.Task.WaitAsync(deadline.Token);
            AgentWorkStatus busy = await child.WorkStatusAsync(deadline.Token);
            Check(busy.Activity == AgentActivity.Busy && !busy.Submitted && busy.Failure is null,
                "Running child status invented completion.");
            ToolResultPart listed = await ChildCall(parent, "agents_status", new { }, deadline.Token);
            Check(!listed.IsError && JsonCodec.Text(listed.Parts).Contains("worker")
                && JsonCodec.Text(listed.Parts).Contains("busy"), "Ordinary status omitted its live child.");
            releaseSubmit.SetResult();
            Check((await child.WaitForSubmissionAsync(deadline.Token)).GetProperty("count").GetInt32() == 9,
                "Spawned child lost typed submission.");
            await draining.Task.WaitAsync(deadline.Token);
            AgentWorkStatus submitted = await child.WorkStatusAsync(deadline.Token);
            Check(submitted.Activity == AgentActivity.Busy && submitted.Submitted,
                "Submission was mistaken for quiescence.");
            Task<ToolResultPart> completed = spawn.CompletionAsync(deadline.Token);
            Check(!completed.IsCompleted, "Spawn released its reliable task reply before draining.");
            releaseFinal.SetResult();
            ToolResultPart final = await completed;
            await server;
            Check(!final.IsError && JsonCodec.Text(final.Parts).Contains("\"task_completed\":true")
                && JsonCodec.Text(final.Parts).Contains("\"count\":9"), "Spawn final report lost completion/payload.");
            AgentWorkStatus idle = await child.WorkStatusAsync(deadline.Token);
            Check(idle.Activity == AgentActivity.Idle && idle.Submitted, "Drained child state was not idle/submitted.");
            Check(child.DataDirectory() != parent.DataDirectory()
                && File.ReadAllText(Path.Combine(child.DataDirectory(), "save.json.journal.jsonl")).Contains("CHILD_ONLY_TASK"),
                "Spawned task did not reach its distinct durable conversation journal.");
            await ChildCall(parent, "shutdown_agent", new { agent_id = "worker" }, deadline.Token);
            Check((await child.WorkStatusAsync(deadline.Token)).Activity == AgentActivity.Stopped,
                "Destroyed child did not report stopped lifetime.");
            ToolResultPart after = await ChildCall(parent, "agents_status", new { }, deadline.Token);
            Check(!JsonCodec.Text(after.Parts).Contains("worker"), "Status retained a removed child.");
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

    public static async Task ChildSpawnCancellation(string temp)
    {
        string work = Path.Combine(temp, "spawn-cancellation");
        Directory.CreateDirectory(work);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        var entered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        int requests = 0;
        var server = Task.Run(async () =>
        {
            using var socket = await listener.AcceptTcpClientAsync(deadline.Token);
            using var stream = socket.GetStream();
            _ = await ReadRequest(stream, deadline.Token);
            requests++;
            entered.SetResult();
            await release.Task.WaitAsync(deadline.Token);
            await Reply(stream, "The separately submitted task is finished.", deadline.Token);
        });
        var app = new Application(TextWriter.Null);
        try
        {
            var parent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture",
                "http://127.0.0.1:" + ((IPEndPoint)listener.LocalEndpoint).Port, "local", sandboxRoot: work), dataDirectory: work);
            parent.InstallTools();
            parent.StartPersistence(new PersistenceOptions(Path.Combine(work, "save.json")));
            foreach (object invalid in new object[] {
                new { agent_id = "bad-task", task = " " },
                new { agent_id = "bad-schema", task = "work", submit_schema = new { type = "object", unsupported = true } },
                new { agent_id = "../escape", task = "work" }
            })
            {
                Check((await ChildCall(parent, "spawn_sub_agent", invalid, deadline.Token)).IsError,
                    "Invalid spawn arguments were accepted.");
                Check(parent.Children().IsEmpty, "Refused spawn attached a child.");
            }
            ToolResultPart early = await ChildCall(parent, "spawn_sub_agent", new {
                agent_id = "worker", task = "Keep working independently", ai_callback_delay = 120
            }, deadline.Token);
            await entered.Task.WaitAsync(deadline.Token);
            Agent child = parent.FindChild("worker")!;
            Check((await ChildCall(parent, "spawn_sub_agent", new { agent_id = "worker", task = "duplicate" }, deadline.Token)).IsError
                && parent.Children().Length == 1, "Duplicate spawn replaced its live sibling.");
            await parent.Jobs().Find(early.Job!.Handle).KillAsync();
            Check(parent.FindChild("worker") == child && (await child.WorkStatusAsync(deadline.Token)).Activity == AgentActivity.Busy,
                "Killing the spawn observer cancelled or destroyed its child.");
            using (var cancelled = new CancellationTokenSource())
            {
                cancelled.Cancel();
                bool refused = false;
                try { await child.WorkStatusAsync(cancelled.Token); }
                catch (OperationCanceledException) when (!deadline.IsCancellationRequested) { refused = true; }
                Check(refused, "Status ignored observer cancellation.");
            }
            child.PublishSubmission("still owned");
            release.SetResult();
            await child.FlushAsync().WaitAsync(deadline.Token);
            await server;
            Check(requests == 1 && child.Submission()!.Value.GetProperty("result").GetString() == "still owned",
                "Cancelled spawn observer lost or retried the child's work.");
            Task<AgentWorkStatus>[] queries = Enumerable.Range(0, 100).Select(_ => child.WorkStatusAsync(deadline.Token)).ToArray();
            await child.ShutdownAsync().WaitAsync(deadline.Token);
            await Task.WhenAll(queries).WaitAsync(deadline.Token);
            Check((await child.WorkStatusAsync(deadline.Token)).Activity == AgentActivity.Stopped,
                "Shutdown stranded status queries or failed to close lifetime.");
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

    public static async Task ChildSpawnFailure(string temp)
    {
        string work = Path.Combine(temp, "spawn-failure");
        Directory.CreateDirectory(work);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        int requests = 0;
        var server = Task.Run(async () =>
        {
            for (int i = 0; i < 4; i++)
            {
                using var socket = await listener.AcceptTcpClientAsync(deadline.Token);
                using var stream = socket.GetStream();
                _ = await ReadRequest(stream, deadline.Token);
                Interlocked.Increment(ref requests);
                await release.Task.WaitAsync(deadline.Token);
                await Reply(stream, "Finished without submitting.", deadline.Token);
            }
        });
        var app = new Application(TextWriter.Null);
        try
        {
            var parent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture",
                "http://127.0.0.1:" + ((IPEndPoint)listener.LocalEndpoint).Port, "local", sandboxRoot: work), dataDirectory: work);
            parent.InstallTools();
            parent.StartPersistence(new PersistenceOptions(Path.Combine(work, "save.json")));
            Agent paused = await parent.CreateChildAsync("paused", cancellation: deadline.Token);
            await paused.PostAsync(new AgentInput(InputKind.Pause));
            Task<string> queued = paused.SubmitAsync([new TextPart("Must not execute while paused")]);
            AgentWorkStatus waiting = await paused.WorkStatusAsync(deadline.Token);
            Check(waiting.Activity == AgentActivity.Busy && waiting.Paused && !waiting.Submitted && requests == 0,
                "Paused queued work was reported idle or executed.");
            await paused.ShutdownAsync().WaitAsync(deadline.Token);
            bool stopped = false;
            try { await queued; }
            catch (ObjectDisposedException) { stopped = true; }
            Check(stopped && (await paused.WorkStatusAsync(deadline.Token)).Activity == AgentActivity.Stopped,
                "Queued task was not resolved at shutdown.");
            paused.Destroy();
            ToolResultPart early = await ChildCall(parent, "spawn_sub_agent", new {
                agent_id = "drifter", task = "Explicitly submit the result", ai_callback_delay = 120
            }, deadline.Token);
            Check(early.Job?.Status == "running", "Failure fixture did not retain a running spawn Job.");
            release.SetResult();
            ToolResultPart failure = await parent.Jobs().Find(early.Job!.Handle).CompletionAsync(deadline.Token);
            await server;
            AgentWorkStatus state = await parent.FindChild("drifter")!.WorkStatusAsync(deadline.Token);
            Check(failure.IsError && requests == 4 && state.Activity == AgentActivity.Idle
                && !state.Submitted && state.Failure is not null,
                "Spawn concealed exhausted correction failure or reported idle as submission.");
            ToolResultPart listed = await ChildCall(parent, "agents_status", new { }, deadline.Token);
            Check(!listed.IsError && JsonCodec.Text(listed.Parts).Contains("drifter")
                && JsonCodec.Text(listed.Parts).Contains(state.Failure!), "Status lost retained child failure.");
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

    public static async Task ChildStatusDirectSend(string temp)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        var entered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var server = Task.Run(async () =>
        {
            using var socket = await listener.AcceptTcpClientAsync(deadline.Token);
            using var stream = socket.GetStream();
            _ = await ReadRequest(stream, deadline.Token);
            entered.SetResult();
            await release.Task.WaitAsync(deadline.Token);
            await Reply(stream, "Direct response.", deadline.Token);
        });
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture",
                "http://127.0.0.1:" + ((IPEndPoint)listener.LocalEndpoint).Port, "local"));
            Check((await agent.WorkStatusAsync(deadline.Token)).Activity == AgentActivity.Idle,
                "Empty Agent was not idle.");
            var send = agent.Engine().SendAsync("{}", deadline.Token);
            await entered.Task.WaitAsync(deadline.Token);
            Check((await agent.WorkStatusAsync(deadline.Token)).Activity == AgentActivity.Busy,
                "Status missed an Engine send outside the actor.");
            release.SetResult();
            await send;
            await server;
            Check((await agent.WorkStatusAsync(deadline.Token)).Activity == AgentActivity.Idle,
                "Direct send left a stale busy state.");
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
