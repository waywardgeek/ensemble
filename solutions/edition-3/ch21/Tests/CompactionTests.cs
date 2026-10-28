using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class CompactionTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition) throw new Exception(message);
    }

    public static void HandoffPrivacy()
    {
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Gemini, "model-a", "http://127.0.0.1:1"));
            var origin = agent.Configuration().Target();
            var signature = JsonSerializer.SerializeToElement(new { thoughtSignature = "PRIVATE_SIGNATURE" });
            var reasoning = new OpaquePart(origin, JsonSerializer.SerializeToElement(new { thought = true, text = "PRIVATE_REASONING" }));
            agent.History().Append(new Event(EventKind.ResponseEnded, actor: Actor.Agent,
                parts: [reasoning, new TextPart("visible explanation", origin, signature)], responseID: 7, partIDs: [1, 2]));
            var original = agent.History().View();
            agent.History().Append(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("resume here")]));
            var current = agent.History().View();
            Check(!current.Dialogue.SelectMany(entry => entry.Parts).Any(part => part is OpaquePart), "Handoff retained opaque reasoning");
            var visible = current.Dialogue.First();
            var text = visible.Parts.OfType<TextPart>().Single();
            Check(text.Text == "visible explanation" && text.From is null && text.Opaque is null, "Handoff lost text or retained its private signature");
            Check(visible.ResponseID == 7 && visible.PartIDs.SequenceEqual(new[] { 2 }), "Handoff changed surviving display identity");
            Check(original.Dialogue.First().Parts.OfType<TextPart>().Single().Opaque is not null, "Handoff mutated prior Context");
            Check(agent.History().Events().First().Parts.OfType<OpaquePart>().Single() == reasoning, "Handoff erased canonical reasoning");
            string request = agent.Engine().Render(current).ToJsonString();
            Check(!request.Contains("PRIVATE_") && request.Contains("visible explanation"), "Handoff request leaked replay data or lost prose");
        }
        finally
        {
            app.Destroy();
        }
    }

    public static void Selection()
    {
        var origin = new Configuration(Vendor.Anthropic, "model-a", "http://127.0.0.1:1").Target();
        var args = JsonSerializer.SerializeToElement(new { });
        // Two complete parallel-call batches deliberately reuse the same IDs.
        // The second batch must never be confused with the first one's results.
        var entries = new List<Entry>();
        for (int batch = 0; batch < 2; batch++)
        {
            int first = batch * 3 + 1;
            entries.Add(new Entry(first, Actor.Agent,
                [new TextPart("retain explanation"), new ToolCallPart("a", origin, "test", args), new ToolCallPart("b", origin, "test", args)]));
            entries.Add(new Entry(first + 1, Actor.Tool, [new ToolResultPart("a", [new TextPart(new string('x', 400))])]));
            entries.Add(new Entry(first + 2, Actor.Tool, [new ToolResultPart("b", [new TextPart(new string('y', 400))])]));
        }
        var context = new Context(entries);
        long total = entries.Sum(entry => entry.ToolBytes);
        long independentlyCounted = entries.SelectMany(entry => entry.Parts)
            .Where(part => part is ToolCallPart or ToolResultPart)
            .Sum(part => (long)Encoding.UTF8.GetByteCount(JsonCodec.WritePart(part).ToJsonString()));
        Check(total == independentlyCounted, "Cached retained bytes differ from canonical tool serialization");
        Check(Agent.SelectToolCut(context, 6, (int)total, Redaction.Result) is null, "Cut below high water");
        Check(Agent.SelectToolCut(context, 6, 0, Redaction.Result) is null, "Disabled band cut");
        Check(Agent.SelectToolCut(context, 0, 1, Redaction.Result) is null, "Unseen history cut");
        Check(Agent.SelectToolCut(context, 2, 1, Redaction.Result) is null, "Partial parallel batch cut");
        int oneBatchBytes = (int)entries.Take(3).Sum(entry => entry.ToolBytes);
        Check(Agent.SelectToolCut(context, 6, oneBatchBytes, Redaction.Result) is null, "Cut at exact high-water boundary");
        var third = entries.Take(3).Select(entry => new Entry(entry.Seq + 6, entry.Actor, entry.Parts));
        var threeBatches = new Context(entries.Concat(third));
        var hysteresis = Agent.SelectToolCut(threeBatches, 9, oneBatchBytes, Redaction.Result)!;
        Check(hysteresis.RedactTo == 6, "Hysteresis did not leave newest batch at target budget");
        var cut = Agent.SelectToolCut(context, 3, 1, Redaction.Result)!;
        Check(cut.RedactFrom == 1 && cut.RedactTo == 3, "Cut split batch or crossed presentation watermark");
        var reduced = History.Reduce(context, cut);
        Check(reduced.Dialogue[1].Parts.OfType<ToolResultPart>().Single().Parts.All(part => part is RedactedPart), "Old results not stubbed");
        Check(reduced.Dialogue[4].Parts.OfType<ToolResultPart>().Single().Parts[0] is TextPart, "Unseen result stubbed");
        Check(reduced.Dialogue[1].ToolBytes < context.Dialogue[1].ToolBytes, "Stub charged original payload bytes");
        var restubbed = History.Reduce(reduced, cut);
        Check(JsonCodec.WritePart(restubbed.Dialogue[1].Parts[0]).ToJsonString()
            == JsonCodec.WritePart(reduced.Dialogue[1].Parts[0]).ToJsonString(), "Repeated result cut changed stub metadata");
        var toolCut = Agent.SelectToolCut(reduced, 6, 1, Redaction.Tool)!;
        Check(toolCut.RedactTo == 3, "Tool cut crossed into whole band");
        var removed = History.Reduce(reduced, toolCut);
        Check(removed.Dialogue[0].Parts.Single() is TextPart, "Tool removal lost accompanying explanation");
        Check(removed.Dialogue.Skip(3).SelectMany(entry => entry.Parts).OfType<ToolCallPart>().Count() == 2, "New calls removed");
        var replayed = History.Reduce(History.Reduce(context, cut), toolCut);
        Check(replayed.Dialogue.Sum(entry => entry.ToolBytes) == removed.Dialogue.Sum(entry => entry.ToolBytes), "Recorded replay differs");
        Check(context.Dialogue[1].Parts.OfType<ToolResultPart>().Single().Parts[0] is TextPart, "Selection mutated old Context");

        // An orphan cannot justify cutting a later apparently complete exchange.
        var orphan = new Entry(1, Actor.Tool, [new ToolResultPart("missing", [new TextPart("orphan")])]);
        Check(Agent.SelectToolCut(new Context(new[] { orphan }.Concat(entries.Skip(3))), 99, 1, Redaction.Result) is null,
            "Selection guessed through orphan traffic");
    }
}

internal static partial class ActorTests
{
    public static async Task ToolCompaction()
    {
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        var app = new Application(TextWriter.Null);
        string directory = Path.Combine(Path.GetTempPath(), "compaction-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(directory);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "model-a",
                $"http://127.0.0.1:{((IPEndPoint)listener.LocalEndpoint).Port}"));
            Check(agent.ContextByteBudget() == 400_000 && agent.ToolCompactionBudgets() == (50_000, 25_000),
                "Wrong target-derived tool budgets");
            agent.SetContextByteBudget(8000);
            try
            {
                agent.SetContextByteBudget(-1);
                throw new Exception("Negative budget accepted");
            }
            catch (ArgumentOutOfRangeException) { }
            Check(agent.ContextByteBudget() == 8000 && agent.ToolCompactionBudgets() == (1000, 500), "Invalid update partially changed policy");
            string payload = "unique-result-" + new string('z', 2000);
            _ = new Tool(agent, "large_result", "Produce real retained output",
                JsonSerializer.SerializeToElement(new { type = "object", properties = new { } }),
                (job, args, cancellation) => Task.FromResult(payload), waitForCompletion: true);
            var server = Task.Run(async () =>
            {
                for (int round = 0; round < 3; round++)
                {
                    using var client = await listener.AcceptTcpClientAsync(deadline.Token);
                    using var stream = client.GetStream();
                    string request = await ReadRequest(stream, deadline.Token);
                    if (round == 0)
                    {
                        await ReplyJson(stream, new
                        {
                            content = new object[] { new { type = "tool_use", id = "large", name = "large_result", input = new { } } },
                            usage = new { input_tokens = 1, output_tokens = 1 }
                        }, deadline.Token);
                    }
                    else
                    {
                        Check(request.Contains(payload) == (round == 1), "Result not presented whole once before compaction");
                        Check(request.Contains("tool_result") && request.Contains("tool_use"), "Compaction broke call/result adjacency");
                        await Reply(stream, "complete", deadline.Token);
                    }
                }
            });
            Check(await agent.SendAsync("first").WaitAsync(deadline.Token) == "complete", "First turn failed");
            Check(!agent.History().Events().Any(fact => fact.Kind == EventKind.Redacted), "Result cut before successful consumption");
            Check(await agent.SendAsync("second").WaitAsync(deadline.Token) == "complete", "Second turn failed");
            await server.WaitAsync(deadline.Token);
            Check(agent.History().Events().Count(fact => fact.Kind == EventKind.Redacted) == 1, "Wrong number of durable cuts");
            Check(agent.History().Events().Any(fact => fact.Kind == EventKind.ToolReturned && JsonCodec.Text(fact.Parts).Contains(payload)),
                "Canonical result lost");
            string save = Path.Combine(directory, "save.json");
            agent.History().Save(new PersistenceOptions(save, RetainLog: 0));
            var restored = new Agent(app, agent.Configuration());
            Check(restored.History().LoadSave(save), "Compacted snapshot not restored");
            Check(restored.History().View().Dialogue.Sum(entry => entry.ToolBytes) == agent.History().View().Dialogue.Sum(entry => entry.ToolBytes),
                "Snapshot did not rebuild derived byte caches");
            Check(!restored.Engine().Render(restored.History().View()).ToJsonString().Contains(payload), "Snapshot resurrected reduced payload");
        }
        finally
        {
            app.Destroy();
            listener.Stop();
            Directory.Delete(directory, true);
        }
    }
}
