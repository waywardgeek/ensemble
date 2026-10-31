using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    private static async Task<ToolResultPart> InterruptChild(Agent parent, object arguments,
        CancellationToken cancellation)
    {
        ToolResultPart result = await parent.DispatchAsync(new ToolCallPart(Guid.NewGuid().ToString("N"),
            parent.Configuration().Target(), "interrupt_agent", JsonSerializer.SerializeToElement(arguments)), cancellation);
        if (result.Job?.Status == "running")
        {
            result = await parent.Jobs().Find(result.Job.Handle).CompletionAsync(cancellation);
        }
        return result;
    }

    public static async Task ChildInterruptOwnership(string temp,
        [System.Runtime.CompilerServices.CallerFilePath] string sourceFile = "")
    {
        var app = new Application(TextWriter.Null);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        string output = Path.Combine(temp, "interrupt-ownership");
        string skills = Path.GetFullPath(Path.Combine(Path.GetDirectoryName(sourceFile)!, "..", "Skills"));
        var config = new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1");
        var root = new Agent(app, config, outputDirectory: output, skillSpec: new SkillSpec(skills, "ensemble"));
        var peer = new Agent(app, config, outputDirectory: output, name: "peer");
        var child = new Agent(app, root, "worker", config, outputDirectory: output);
        _ = new Agent(app, peer, "foreign", config, outputDirectory: output);
        _ = new Agent(app, child, "nested", config, outputDirectory: output);
        try
        {
            await root.SkillManager()!.InitializeAsync(deadline.Token);
            Check(root.ToolEnabled("interrupt_agent")
                && root.Engine().Render(root.History().View()).ToJsonString().Contains("interrupt_agent", StringComparison.Ordinal),
                "Shipped skill does not grant and advertise interrupt_agent.");
            ToolResultPart unknown = await InterruptChild(root, new { agent_id = "missing" }, deadline.Token);
            foreach (string id in new[] { "foreign", "nested", "peer" })
            {
                ToolResultPart refused = await InterruptChild(root, new { agent_id = id }, deadline.Token);
                Check(unknown.IsError && refused.IsError && JsonCodec.Text(unknown.Parts) == JsonCodec.Text(refused.Parts),
                    "Interrupt exposed or controlled an Agent outside direct ownership.");
            }
            foreach (object invalid in new object[] {
                new { agent_id = "worker", mode = "later" },
                new { agent_id = "worker", mode = "" },
                new { agent_id = "worker", mode = 1 } })
            {
                Check((await InterruptChild(root, invalid, deadline.Token)).IsError, "Invalid mode was accepted.");
            }
            ToolResultPart delivered = await InterruptChild(root, new { agent_id = "worker" }, deadline.Token);
            Check(!delivered.IsError && JsonCodec.Text(delivered.Parts).Contains("\"mode\":\"now\"", StringComparison.Ordinal),
                "Omitted mode did not default to now.");
            Check(root.FindChild("worker") == child && (await child.WorkStatusAsync(deadline.Token)).Activity == AgentActivity.Idle,
                "An idle interrupt destroyed or started work on the child.");
            await child.ShutdownAsync();
            Check((await InterruptChild(root, new { agent_id = "worker" }, deadline.Token)).IsError,
                "Stopped child falsely acknowledged delivery.");
        }
        finally
        {
            app.Destroy();
        }
    }

    public static async Task ChildInterruptBoundary(string temp, bool afterTool)
    {
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        var entered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        int forbiddenEffects = 0;
        var server = Task.Run(async () =>
        {
            // First turn has held work and an undispatched tail. The later turn
            // proves that the same delegated Agent can still submit and finish.
            for (int request = 0; request < 3; request++)
            {
                using var client = await listener.AcceptTcpClientAsync(deadline.Token);
                using var stream = client.GetStream();
                _ = await ReadRequest(stream, deadline.Token);
                if (request == 0)
                {
                    await ReplyJson(stream, new
                    {
                        model = "fixture", content = new object[] {
                            new { type = "tool_use", id = "held", name = "held_work", input = new { ai_callback_delay = 120 } },
                            new { type = "tool_use", id = "tail", name = "forbidden_work", input = new { } }
                        }, usage = new { input_tokens = 1, output_tokens = 1 }
                    }, deadline.Token);
                }
                else if (request == 1)
                {
                    await ReplyJson(stream, new
                    {
                        model = "fixture", content = new object[] {
                            new { type = "tool_use", id = "finish", name = "submit", input = new { result = "recovered" } }
                        }, usage = new { input_tokens = 1, output_tokens = 1 }
                    }, deadline.Token);
                }
                else
                {
                    await Reply(stream, "continued", deadline.Token);
                }
            }
        });
        var app = new Application(TextWriter.Null);
        try
        {
            int port = ((IPEndPoint)listener.LocalEndpoint).Port;
            var config = new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:" + port);
            string output = Path.Combine(temp, afterTool ? "interrupt-boundary" : "interrupt-now");
            var parent = new Agent(app, config, outputDirectory: output);
            parent.InstallTools();
            var child = new Agent(app, parent, "worker", config, outputDirectory: output,
                dataDirectory: Path.Combine(output, "child"));
            child.StartPersistence(new PersistenceOptions(Path.Combine(output, "child", "save.json")));
            child.InstallTools();
            var schema = JsonSerializer.SerializeToElement(new { type = "object", properties = new { } });
            _ = new Tool(child, "held_work", "Work held at an explicit test gate", schema, async (job, args, cancellation) =>
            {
                entered.TrySetResult();
                await release.Task.WaitAsync(cancellation);
                return "actual worker result";
            });
            _ = new Tool(child, "forbidden_work", "Must remain undispatched", schema, (job, args, cancellation) =>
            {
                Interlocked.Increment(ref forbiddenEffects);
                return Task.FromResult("incorrect effect");
            });
            Task<string> turn = child.SubmitAsync([new TextPart("start")]);
            await entered.Task.WaitAsync(deadline.Token);
            ToolResultPart delivered = await InterruptChild(parent, new {
                agent_id = "worker", mode = afterTool ? "after_tool" : "now"
            }, deadline.Token);
            Check(!delivered.IsError && JsonCodec.Text(delivered.Parts).Contains("\"interrupt_delivered\":true", StringComparison.Ordinal),
                "Owned interrupt was not acknowledged.");
            Job work = child.Jobs().Children().Single();
            Check(work.Snapshot().Status == "running", "Interrupt killed independent work.");
            if (afterTool)
            {
                // The delivery acknowledgment follows the child's actor commit;
                // no sleep or external scheduling assumption establishes this check.
                Check(!turn.IsCompleted, "after_tool stopped before its report boundary.");
                release.TrySetResult();
            }
            try
            {
                await turn.WaitAsync(deadline.Token);
                throw new Exception("Interrupted turn returned success.");
            }
            catch (OperationCanceledException) when (!deadline.IsCancellationRequested) { }
            Check(forbiddenEffects == 0 && parent.FindChild("worker") == child,
                "Interrupt dispatched a tail or detached the child.");
            if (!afterTool)
            {
                Check(work.Snapshot().Status == "running", "now joined or killed the independent Job.");
            }
            string continuation;
            try
            {
                continuation = await child.SubmitAsync([new TextPart("continue")]).WaitAsync(deadline.Token);
            }
            catch
            {
                foreach (Event fact in child.History().Events().Where(fact => fact.Kind == EventKind.ToolReturned))
                {
                    Console.WriteLine($"Continuation diagnostic: {fact.CallID}: {JsonCodec.Text(fact.Parts)}");
                }
                throw;
            }
            Check(continuation == "continued"
                && child.Submission()!.Value.GetProperty("result").GetString() == "recovered",
                "The interrupted child could not complete later delegated work.");
            release.TrySetResult();
            ToolResultPart completed = await work.CompletionAsync(deadline.Token);
            Check(!completed.IsError && JsonCodec.Text(completed.Parts).Contains("actual worker result", StringComparison.Ordinal),
                "Surviving Job lost its actual result.");
            Event skipped = child.History().Events().Single(fact => fact.Kind == EventKind.ToolReturned && fact.CallID == "tail");
            Check(skipped.IsError && JsonCodec.Text(skipped.Parts).Contains("not executed", StringComparison.Ordinal),
                "Undispatched tail was not honestly paired.");
            await server.WaitAsync(deadline.Token);
        }
        finally
        {
            release.TrySetResult();
            app.Destroy();
            listener.Stop();
            // Cleanup also joins the fixture if an assertion failed mid-protocol.
            try { await server; }
            catch (Exception) when (server.IsFaulted || server.IsCanceled) { }
        }
    }
}
