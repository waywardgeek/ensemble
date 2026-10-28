using Ensemble;
using System.Text.Json;
using Application = Ensemble.Ensemble;

// These are core prerequisites for a reconnecting client, not browser tests.
internal static class GuiCoreTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition)
        {
            throw new Exception(message);
        }
    }

    public static async Task SnapshotsAndIdentity()
    {
        var app = new Application(TextWriter.Null);
        var agent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1"));
        try
        {
            var history = agent.History();
            history.Append(new Event(EventKind.MessageReceived, actor: Actor.Human, parts: [new TextPart("prompt")]));
            var request = history.Append(new Event(EventKind.RequestSent, from: agent.Configuration().Target()));
            var final = history.Append(new Event(EventKind.ResponseEnded,
                parts: [new TextPart("zero"), new TextPart("seven")], responseID: request.Seq, partIDs: [0, 7]));
            var restored = JsonCodec.ReadEvent(JsonCodec.WriteEvent(final));
            Check(restored.ResponseID == request.Seq && restored.PartIDs.SequenceEqual([0, 7]),
                "Stable response-local identities did not survive serialization");
            Check(JsonCodec.Text(restored.Parts) == JsonCodec.Text(final.Parts), "Display metadata changed content");
            var legacy = JsonCodec.ReadEvent("{\"seq\":1,\"type\":\"response_ended\",\"response\":{\"parts\":[]}}");
            Check(legacy.ResponseID == 0 && legacy.PartIDs.IsEmpty, "Historical logs need no invented parser identity");
            foreach (var ids in new[] { new[] { 0 }, new[] { 0, 0 }, new[] { -1, 2 } })
            {
                bool refused = false;
                try
                {
                    _ = final.WithDisplay(request.Seq, ids);
                }
                catch (ArgumentException)
                {
                    refused = true;
                }
                Check(refused, "Invalid identity metadata was accepted");
            }

            var frozen = history.Snapshot();
            var writer = Task.Run(() =>
            {
                for (int i = 0; i < 250; i++)
                {
                    history.Append(new Event(EventKind.MessageReceived, actor: Actor.Human, parts: [new TextPart(i.ToString())]));
                }
            });
            do
            {
                var snapshot = history.Snapshot();
                var messages = snapshot.Events.Count(e => e.Kind is EventKind.MessageReceived or EventKind.ResponseEnded);
                Check(snapshot.Context.Dialogue.Length == messages, "Snapshot mixes different append instants");
                Check(snapshot.Events.Select(e => e.Seq).SequenceEqual(Enumerable.Range(1, snapshot.Events.Length).Select(n => (long)n)),
                    "Snapshot lost sequence ordering");
            }
            while (!writer.IsCompleted);
            await writer;
            Check(frozen.Events.Length == 3 && frozen.Context.Dialogue.Length == 2, "An old snapshot mutated");
        }
        finally
        {
            app.Destroy();
        }
    }

    public static async Task LifecycleFacts()
    {
        var app = new Application(TextWriter.Null);
        var agent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1"));
        int effects = 0;
        _ = new Tool(agent, "probe", "Admission probe", JsonSerializer.SerializeToElement(new
        {
            type = "object", properties = new { }, additionalProperties = false
        }), (job, args, cancellation) =>
        {
            // Admission must already be durable when the real handler begins.
            Check(job.Parent().Parent().History().Events().Any(e => e.Kind == EventKind.ToolStarted && e.CallID == "ok"),
                "Handler preceded its started fact");
            Interlocked.Increment(ref effects);
            return Task.FromResult("done");
        });
        try
        {
            await agent.PostAsync(new AgentInput(InputKind.Pause));
            await agent.PostAsync(new AgentInput(InputKind.Resume));
            var origin = agent.Configuration().Target();
            var bad = new ToolCallPart("bad", origin, "probe", JsonSerializer.SerializeToElement(new { extra = true }));
            Check((await agent.DispatchAsync(bad)).IsError, "Invalid schema must refuse");
            var good = new ToolCallPart("ok", origin, "probe", JsonSerializer.SerializeToElement(new { }));
            Check(!(await agent.DispatchAsync(good)).IsError && effects == 1, "Valid tool did not execute exactly once");
            var facts = agent.History().Snapshot().Events;
            Check(!facts.Any(e => e.Kind == EventKind.ToolStarted && e.CallID == "bad"), "Refusal falsely claimed dispatch");
            Check(facts.Any(e => e.Kind == EventKind.AgentPaused) && facts.Any(e => e.Kind == EventKind.AgentResumed),
                "Pause and resume are not replayable");
            foreach (var fact in facts)
            {
                Check(JsonCodec.ReadEvent(JsonCodec.WriteEvent(fact)).Kind == fact.Kind, "Lifecycle codec lost a fact");
            }
            await agent.ShutdownAsync();
            await agent.ShutdownAsync();
            var shutdown = agent.History().Events().Where(e => e.Kind == EventKind.AgentShutdown).ToArray();
            Check(shutdown.Length == 1 && agent.History().Events().Last() == shutdown[0],
                "Joined shutdown must record exactly one terminal fact");
            Check(JsonCodec.ReadEvent(JsonCodec.WriteEvent(shutdown[0])).Kind == EventKind.AgentShutdown,
                "Shutdown fact must survive replay");
        }
        finally
        {
            app.Destroy();
        }
    }
}
