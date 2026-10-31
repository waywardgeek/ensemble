using System.Diagnostics;
using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    public static async Task HintObservationCaps(string temp)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(25));
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1"),
                outputDirectory: Path.Combine(temp, "hint-timing"));
            Job job = await HeldMessageJob(agent, release, deadline.Token);
            var elapsed = Stopwatch.StartNew();
            Task<ToolResultPart> waiting = job.ObserveAsync("early", 120, "never-match", cancellation: deadline.Token);
            job.WakeObservation();
            await Task.Delay(1800, deadline.Token);
            job.WakeObservation(); // A second hint must not restart the three seconds.
            ToolResultPart early = await waiting.WaitAsync(deadline.Token);
            Check(elapsed.Elapsed.TotalSeconds is >= 2.7 and < 4.5, "Early/repeated hint did not cap from observation start.");
            Check(early.CallID == "early" && early.Job?.Status == "running", "Early hint misreported work.");

            waiting = job.ObserveAsync("aged", 120, "never-match", cancellation: deadline.Token);
            await Task.Delay(3200, deadline.Token); // Deliberately make this observation older than the cap.
            elapsed.Restart();
            job.WakeObservation();
            ToolResultPart aged = await waiting.WaitAsync(deadline.Token);
            Check(elapsed.Elapsed.TotalSeconds < 1, "Aged observation waited another three seconds.");
            Check(aged.CallID == "aged" && aged.Job?.Handle == job.Handle, "Aged hint lost correlation/handle.");

            elapsed.Restart();
            waiting = job.ObserveAsync("short", .15, "never-match", cancellation: deadline.Token);
            job.WakeObservation();
            await waiting.WaitAsync(deadline.Token);
            Check(elapsed.Elapsed.TotalSeconds < 1, "Hint lengthened an explicitly short delay.");

            elapsed.Restart();
            waiting = job.ObserveAsync("fast", 120, "never-match", cancellation: deadline.Token);
            job.WakeObservation();
            release.TrySetResult();
            ToolResultPart fast = await waiting.WaitAsync(deadline.Token);
            Check(elapsed.Elapsed.TotalSeconds < 1 && fast.Job?.Status == "done", "Fast completion waited for the cap.");
            Check(JsonCodec.Text(fast.Parts).Contains("boundary completed"), "Fast completion lost its actual result.");
        }
        finally { release.TrySetResult(); app.Destroy(); }
    }

    public static async Task HintBatchCaps(string temp)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(25));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        var entered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var app = new Application(TextWriter.Null);
        var effects = new int[4];
        var times = new long[4];
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
                        new { type = "tool_use", id = "cap-0", name = "cap_work", input = new { index = 0, ai_callback_delay = 120.0 } },
                        new { type = "tool_use", id = "cap-1", name = "cap_work", input = new { index = 1, ai_callback_delay = 120.0 } },
                        new { type = "tool_use", id = "cap-2", name = "cap_work", input = new { index = 2, ai_callback_delay = 120.0 } },
                        new { type = "tool_use", id = "cap-3", name = "cap_work", input = new { index = 3, ai_callback_delay = .15 } }
                    }, usage = new { input_tokens = 1, output_tokens = 1 }
                }, deadline.Token);
            }
            using (var client = await listener.AcceptTcpClientAsync(deadline.Token))
            using (var stream = client.GetStream())
            {
                string request = await ReadRequest(stream, deadline.Token);
                Check(request.Contains("CAP_HINT") && request.Contains("fast-result"), "Next inference omitted hint or fast result: " + request);
                for (int i = 0; i < 4; i++)
                {
                    Check(request.Contains("cap-" + i), "Next inference omitted a correlated call.");
                }
                await Reply(stream, "batch seen", deadline.Token);
            }
        });
        try
        {
            int port = ((IPEndPoint)listener.LocalEndpoint).Port;
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:" + port),
                outputDirectory: Path.Combine(temp, "hint-batch"));
            _ = new Tool(agent, "cap_work", "Gated effects", JsonSerializer.SerializeToElement(new { type = "object", properties = new { index = new { type = "integer" } }, required = new[] { "index" } }),
                async (job, args, token) =>
                {
                    int index = args.GetProperty("index").GetInt32();
                    Interlocked.Increment(ref effects[index]);
                    times[index] = Stopwatch.GetTimestamp();
                    if (index == 0) entered.TrySetResult();
                    if (index != 2) await release.Task.WaitAsync(token);
                    return index == 2 ? "fast-result" : "released-" + index;
                });
            Task<string> turn = agent.SubmitAsync([new TextPart("Run the batch")]);
            Task first = await Task.WhenAny(entered.Task, turn, server).WaitAsync(deadline.Token);
            if (first != entered.Task)
            {
                await first;
                throw new Exception("Fixture ended before the first handler entered.");
            }
            await agent.WorkStatusAsync(deadline.Token); // Admission barrier before the elapsed-time scenario.
            await Task.Delay(3200, deadline.Token);
            var elapsed = Stopwatch.StartNew();
            await agent.PostAsync(new AgentInput(InputKind.Hint, "CAP_HINT")).WaitAsync(deadline.Token);
            Check(await turn.WaitAsync(deadline.Token) == "batch seen", "Batch did not continue normally.");
            await server.WaitAsync(deadline.Token);
            Check(elapsed.Elapsed.TotalSeconds is >= 2.7 and < 5.5, "Pending batch did not use per-observation caps.");
            Check(Stopwatch.GetElapsedTime(times[1], times[2]).TotalSeconds is >= 2.7 and < 4.5,
                "Second call failed to receive its own three-second cap.");
            Check(Stopwatch.GetElapsedTime(times[2], times[3]).TotalSeconds < 1, "Fast pending call waited three seconds.");
            Check(effects.All(count => count == 1), "A finalized call was lost or executed twice.");
            Event[] results = agent.History().Events().Where(e => e.Kind == EventKind.ToolReturned).ToArray();
            Check(results.Length == 4, "Batch did not record exactly four results.");
            for (int i = 0; i < 4; i++)
            {
                Event result = results.Single(e => e.CallID == "cap-" + i);
                Check(!result.IsError && result.Job is not null, "Hint fabricated a skipped/error result.");
                Check(result.Job!.Status == (i == 2 ? "done" : "running"), "Batch result misreported completion.");
                Job job = agent.Jobs().Find(result.Job.Handle);
                Check(job.Snapshot().Status == result.Job.Status, "Reported handle did not identify the actual Job.");
            }
            release.TrySetResult();
            foreach (Job job in agent.Jobs().Children())
            {
                Check((await job.ObserveAsync("drain-" + job.Handle, 2, cancellation: deadline.Token)).Job?.Status == "done",
                    "Hint cancelled or lost underlying work.");
            }
        }
        finally { release.TrySetResult(); app.Destroy(); listener.Stop(); }
    }
}
