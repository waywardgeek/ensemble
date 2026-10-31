using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    public static async Task HandoffBatch()
    {
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        var started = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "model-a",
                $"http://127.0.0.1:{((IPEndPoint)listener.LocalEndpoint).Port}"));
            agent.InstallTools();
            _ = new Tool(agent, "held_companion", "Hold a real companion until released",
                JsonSerializer.SerializeToElement(new { type = "object", properties = new { }, additionalProperties = false }),
                async (job, args, cancellation) =>
                {
                    started.TrySetResult();
                    await release.Task.WaitAsync(cancellation);
                    return "companion completed";
                }, waitForCompletion: true);
            // Retained facts are not another memory store. The old note exists in
            // canonical history but must disappear from active Context later.
            agent.History().Append(new Event(EventKind.HintReceived, actor: Actor.System,
                parts: [new TextPart("ordinary system instructions")]));
            agent.History().Append(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("old note")]));
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
                            content = new object[]
                            {
                                new { type = "text", text = "I will preserve progress and finish the companion." },
                                new { type = "tool_use", id = "note-1", name = "micro_handoff", input = new { text = "first replacement" } },
                                new { type = "tool_use", id = "companion", name = "held_companion", input = new { } }
                            },
                            usage = new { input_tokens = 1, output_tokens = 1 }
                        }, deadline.Token);
                    }
                    else
                    {
                        string expected = round == 1 ? "first replacement" : "newest replacement";
                        Check(request.Contains(expected) && !request.Contains("old note"), "Wrong active note on next request");
                        using var document = JsonDocument.Parse(request);
                        string messages = document.RootElement.GetProperty("messages").GetRawText();
                        Check(!messages.Contains("\"tool_use\"") && !messages.Contains("\"tool_result\""), "Handoff retained tool traffic");
                        if (round == 1)
                        {
                            await ReplyJson(stream, new
                            {
                                content = new object[] { new { type = "tool_use", id = "note-2", name = "micro_handoff", input = new { text = "newest replacement" } } },
                                usage = new { input_tokens = 1, output_tokens = 1 }
                            }, deadline.Token);
                        }
                        else
                        {
                            Check(!request.Contains("first replacement"), "Older active note survived replacement");
                            await Reply(stream, "handoff complete", deadline.Token);
                        }
                    }
                }
            });
            var turn = agent.SendAsync("continue the work");
            await started.Task.WaitAsync(deadline.Token);
            var held = agent.History().View();
            Check(held.Dialogue.SelectMany(entry => entry.Parts).OfType<HandoffPart>().Single().Text == "old note",
                "Handoff reduced before companion result");
            Check(agent.History().Events().Any(fact => fact.Kind == EventKind.ToolReturned && fact.CallID == "note-1"),
                "Successful handoff result not recorded before holding companion");
            release.TrySetResult();
            Check(await turn.WaitAsync(deadline.Token) == "handoff complete", "Turn did not finish");
            await server.WaitAsync(deadline.Token);
            var events = agent.History().Events();
            long firstNote = events.Single(fact => fact.Kind == EventKind.HandoffRecorded
                && fact.Parts.OfType<HandoffPart>().Any(note => note.Text == "first replacement")).Seq;
            Check(events.Single(fact => fact.Kind == EventKind.ToolReturned && fact.CallID == "companion").Seq < firstNote,
                "Companion result recorded after reduction");
            var context = agent.History().View();
            Check(context.Dialogue.SelectMany(entry => entry.Parts).OfType<HandoffPart>().Single().Text == "newest replacement",
                "Newest note not sole active handoff");
            Check(JsonCodec.Text(context.Dialogue.SelectMany(entry => entry.Parts)).Contains("ordinary system instructions"),
                "Ordinary system content lost");
            Check(events.Count(fact => fact.Kind == EventKind.HandoffRecorded) == 3, "Canonical history lost older notes");

            var invalid = new ToolCallPart("invalid-note", agent.Configuration().Target(), "micro_handoff",
                JsonSerializer.SerializeToElement(new { text = "   " }));
            Check((await agent.DispatchAsync(invalid)).IsError, "Blank handoff accepted");
            Check(agent.History().Events().Count(fact => fact.Kind == EventKind.HandoffRecorded) == 3,
                "Failed handoff caused reduction");
            var direct = new ToolCallPart("direct-note", agent.Configuration().Target(), "micro_handoff",
                JsonSerializer.SerializeToElement(new { text = "direct replacement" }));
            Check(!(await agent.DispatchAsync(direct)).IsError, "Direct handoff refused");
            await agent.FlushAsync(deadline.Token);
            Check(agent.History().View().Dialogue.SelectMany(entry => entry.Parts)
                .OfType<HandoffPart>().Single().Text == "direct replacement", "Direct handoff not applied");
        }
        finally
        {
            release.TrySetResult();
            app.Destroy();
            listener.Stop();
        }
    }
}

internal static class HandoffTests
{
    public static void Replay()
    {
        string directory = Path.Combine(Path.GetTempPath(), "handoff-replay-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(directory);
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "model-a", "http://127.0.0.1:1"));
            var history = agent.History();
            history.Append(new Event(EventKind.SkillLoaded, parts: [new SkillPart("manual", "retained manual", ["write_file"], pending: true)]));
            history.Append(new Event(EventKind.MessageReceived, actor: Actor.Human, parts: [new TextPart("dialogue")]));
            history.Append(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("old")]));
            history.Append(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("new")]));
            string save = Path.Combine(directory, "save.json");
            history.Save(new PersistenceOptions(save, RetainLog: 0));
            var restored = new Agent(app, agent.Configuration());
            if (!restored.History().LoadSave(save)) throw new Exception("Handoff save absent");
            var context = restored.History().View();
            var parts = context.Dialogue.SelectMany(entry => entry.Parts).ToArray();
            if (parts.OfType<HandoffPart>().Single().Text != "new"
                || parts.OfType<TextPart>().Single().Text != "dialogue"
                || !parts.OfType<SkillPart>().Single().Pending
                || restored.ToolEnabled("write_file"))
            {
                throw new Exception("Snapshot changed handoff/dialogue/manual authority");
            }
        }
        finally
        {
            app.Destroy();
            Directory.Delete(directory, true);
        }
    }
}

internal static partial class ActorTests
{
    public static async Task InterruptedHandoff()
    {
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var started = new TaskCompletionSource<Job>(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var app = new Application(TextWriter.Null);
        bool tailRan = false;
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "model-a",
                $"http://127.0.0.1:{((IPEndPoint)listener.LocalEndpoint).Port}"));
            agent.InstallTools();
            var schema = JsonSerializer.SerializeToElement(new { type = "object", properties = new { }, additionalProperties = false });
            _ = new Tool(agent, "held_companion", "Held job", schema, async (job, args, cancellation) =>
            {
                started.TrySetResult(job);
                await release.Task.WaitAsync(cancellation);
                return "independent job completed";
            }, waitForCompletion: true);
            _ = new Tool(agent, "tail_effect", "Must not run after interruption", schema, (job, args, cancellation) =>
            {
                tailRan = true;
                return Task.FromResult("unexpected effect");
            });
            var server = Task.Run(async () =>
            {
                using var client = await listener.AcceptTcpClientAsync(deadline.Token);
                using var stream = client.GetStream();
                _ = await ReadRequest(stream, deadline.Token);
                await ReplyJson(stream, new
                {
                    content = new object[]
                    {
                        new { type = "text", text = "I will record progress before the remaining work." },
                        new { type = "tool_use", id = "note", name = "micro_handoff", input = new { text = "interrupted continuation" } },
                        new { type = "tool_use", id = "held", name = "held_companion", input = new { } },
                        new { type = "tool_use", id = "tail", name = "tail_effect", input = new { } }
                    },
                    usage = new { input_tokens = 1, output_tokens = 1 }
                }, deadline.Token);
            });
            var turn = agent.SendAsync("preserve progress");
            var job = await started.Task.WaitAsync(deadline.Token);
            await agent.PostAsync(new AgentInput(InputKind.InterruptNow)).WaitAsync(deadline.Token);
            try
            {
                await turn.WaitAsync(deadline.Token);
                throw new Exception("Interrupted turn reported success");
            }
            catch (OperationCanceledException) { }
            await agent.FlushAsync(deadline.Token);
            var events = agent.History().Events();
            long cut = events.Single(fact => fact.Kind == EventKind.HandoffRecorded).Seq;
            Check(events.Single(fact => fact.Kind == EventKind.ToolReturned && fact.CallID == "tail").Seq < cut
                && events.Single(fact => fact.Kind == EventKind.ToolReturned && fact.CallID == "held").Seq < cut,
                "Interrupted companion/tail was not paired before handoff");
            Check(!tailRan, "Undispatched tail produced an effect");
            Check(agent.History().View().Dialogue.SelectMany(entry => entry.Parts).OfType<HandoffPart>().Single().Text
                == "interrupted continuation", "Interrupted batch lost its accepted note");
            release.TrySetResult();
            var completion = await job.CompletionAsync(deadline.Token);
            Check(!completion.IsError && JsonCodec.Text(completion.Parts) == "independent job completed",
                "Interrupt killed the independent companion Job");
            await server.WaitAsync(deadline.Token);
        }
        finally
        {
            release.TrySetResult();
            app.Destroy();
            listener.Stop();
        }
    }
}
