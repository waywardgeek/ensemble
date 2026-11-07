using Ensemble;
using System.Text.Json;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    public static async Task WorkingContextIndependentJob()
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        var entered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var exited = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        CancellationToken workerToken = default;
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4o-mini",
                listener.Prefixes.Single(), disableStreaming: true));
            await agent.SetWorkingContextLimitAsync(new WorkingContextLimit(Vendor.OpenAI, "gpt-4o-mini", 100_000, 128_000));
            _ = new Tool(agent, "held", "Held independent work", JsonSerializer.SerializeToElement(new
                { type = "object", properties = new { }, additionalProperties = false }),
                async (job, args, cancellation) =>
                {
                    workerToken = cancellation;
                    entered.TrySetResult();
                    try { await release.Task.WaitAsync(cancellation); return "completed"; }
                    finally { exited.TrySetResult(); }
                });
            var observed = await agent.DispatchAsync(new ToolCallPart("held-call", agent.Configuration().Target(), "held", JsonSerializer.SerializeToElement(new { })));
            await entered.Task.WaitAsync(TimeSpan.FromSeconds(10));
            Check(observed.Job is { Status: "running" }, "Expected actual independent running Job");
            await PressureTurn(agent, listener, PressureReply(100_000));
            Check(agent.History().Events().Any(fact => fact.Kind == EventKind.HandoffRecorded)
                && !workerToken.IsCancellationRequested && !exited.Task.IsCompleted,
                "Automatic checkpoint canceled independent work");
            release.TrySetResult();
            await exited.Task.WaitAsync(TimeSpan.FromSeconds(10));
        }
        finally { release.TrySetResult(); app.Destroy(); }
    }

    public static async Task WorkingContextPressure()
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4o-mini",
                listener.Prefixes.Single(), disableStreaming: true));
            await agent.SetWorkingContextLimitAsync(new WorkingContextLimit(Vendor.OpenAI,
                "gpt-4o-mini", 100_000, 128_000));
            int effects = 0;
            _ = new Tool(agent, "effect", "Count a real effect", JsonSerializer.SerializeToElement(new
                { type = "object", properties = new { }, additionalProperties = false }),
                (job, args, cancellation) => { Interlocked.Increment(ref effects); return Task.FromResult("effect completed"); },
                waitForCompletion: true);
            // No micro_handoff grant: automatic maintenance is not a fake tool call.
            await PressureTurn(agent, listener, PressureReply(79_999));
            var belowTurn = agent.SubmitAsync([new TextPart("One token below warning")]);
            var belowRequest = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            Check(!(await PressureWire(belowRequest)).Contains("Ensemble context reminder:"), "Rounded up below 80% boundary");
            await Reply(belowRequest, PressureReply(80_000, cached: 70_000));
            await belowTurn.WaitAsync(TimeSpan.FromSeconds(10));
            Check(agent.WorkingContextStatus().Percent == 80m, "Cached input counted more than once");

            async Task<string> Turn(int tokens)
            {
                var task = agent.SubmitAsync([new TextPart("Continue")]);
                var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
                string wire = await PressureWire(request);
                await Reply(request, PressureReply(tokens));
                await task.WaitAsync(TimeSpan.FromSeconds(10));
                return wire;
            }
            string first = await Turn(80_000);
            Check(first.Contains("80.0%") && first.Contains("next good handoff point")
                && !first.Contains("Call micro_handoff now"), "80% gentle reminder missing or escalated early");
            Check(!(await Turn(90_000)).Contains("Ensemble context reminder:"), "Gentle reminder repeated");
            foreach (int tokens in new[] { 90_000, 99_999 })
            {
                string urgent = await Turn(tokens);
                Check(urgent.Contains("90.0%") && urgent.Contains("Call micro_handoff now")
                    && urgent.Contains("Do not begin another investigation, edit, or test")
                    && urgent.Contains("finish recording its outcomes and hand off immediately afterward")
                    && !urgent.Contains("next good handoff point"),
                    "90% immediate handoff instruction missing from repeated request");
            }
            Check(!agent.History().Events().Any(fact => fact.Kind == EventKind.HandoffRecorded),
                "Checkpoint occurred below exact 100% threshold");

            var turn = agent.SubmitAsync([new TextPart("Reach working limit")]);
            var next = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            var batch = System.Text.Json.Nodes.JsonNode.Parse(PressureReply(100_000, "effect"))!;
            var batchCalls = batch["choices"]![0]!["message"]!["tool_calls"]!.AsArray();
            var companion = batchCalls[0]!.DeepClone();
            companion["id"] = "companion-call";
            batchCalls.Add(companion);
            await Reply(next, batch.ToJsonString());
            next = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            string after = await PressureWire(next);
            var facts = agent.History().Events();
            var checkpoint = facts.Single(fact => fact.Kind == EventKind.HandoffRecorded);
            var call = facts.Last(fact => fact.Kind == EventKind.ToolCalled && fact.Name == "effect");
            var result = facts.Single(fact => fact.Kind == EventKind.ToolReturned && fact.CallID == call.CallID);
            Check(effects == 2 && result.Seq < checkpoint.Seq
                && facts.Count(fact => fact.Kind == EventKind.ToolReturned && fact.Seq < checkpoint.Seq) == 2,
                "Automatic handoff ran before all companion results");
            Check(checkpoint.Parts.OfType<HandoffPart>().Single().Text.Contains("No model-authored continuation note"), "Fallback note is not honest");
            Check(!facts.Any(fact => fact.Kind == EventKind.ToolCalled && fact.Name == "micro_handoff"),
                "Runtime fabricated a model tool call");
            Check(agent.WorkingContextStatus().InputTokens is null && !after.Contains("Ensemble context reminder:"),
                "Old usage retriggered pressure after checkpoint");
            Check(after.Contains("Completed micro_handoff checkpoint"), "Completed checkpoint marker missing");
            await Reply(next, PressureReply(1));
            await turn.WaitAsync(TimeSpan.FromSeconds(10));
            Check(agent.History().Events().Count(fact => fact.Kind == EventKind.HandoffRecorded) == 1,
                "Automatic checkpoint repeated without a new high measurement");
            await agent.SetWorkingContextLimitAsync(new WorkingContextLimit(Vendor.OpenAI,
                "gpt-4o-mini", 1, 128_000));
            // Actor processes an idle settings update before deciding on maintenance.
            await agent.FlushAsync();
            Check(agent.History().Events().Count(fact => fact.Kind == EventKind.HandoffRecorded) == 2,
                "Lowered working limit did not apply to latest reported usage");
        }
        finally { app.Destroy(); }
    }
}
