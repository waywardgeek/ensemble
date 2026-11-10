using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    public static async Task HandoffContinuation()
    {
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        var app = new Application(TextWriter.Null);
        Task? server = null;
        int requests = 0;
        string directory = Path.Combine(Path.GetTempPath(), "handoff-checkpoint-" + Guid.NewGuid().ToString("N"));
        string savePath = Path.Combine(directory, "save.json");
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "model-a",
                $"http://127.0.0.1:{((IPEndPoint)listener.LocalEndpoint).Port}"));
            agent.InstallTools();
            agent.StartPersistence(new PersistenceOptions(savePath));
            server = Task.Run(async () =>
            {
                // A bounded adversarial peer models the ambiguity: a future-tense
                // promise plus an unlabelled note looks like unfinished work.
                for (int round = 0; round < 3; round++)
                {
                    using var client = await listener.AcceptTcpClientAsync(deadline.Token);
                    using var stream = client.GetStream();
                    string request = await ReadRequest(stream, deadline.Token);
                    requests++;
                    bool completed = request.Contains("[Completed micro_handoff checkpoint]");
                    if (round > 0 && completed)
                    {
                        Check(request.Contains("Next: inspect the replay test."), "Checkpoint content lost");
                        await Reply(stream, "Resuming the replay test.", deadline.Token);
                        return;
                    }
                    if (round == 2)
                    {
                        await Reply(stream, "Repeated handoff without progress.", deadline.Token);
                        return;
                    }
                    await ReplyJson(stream, new
                    {
                        content = new object[]
                        {
                            new { type = "text", text = "I will save the checkpoint now." },
                            new { type = "tool_use", id = "checkpoint-" + round, name = "micro_handoff",
                                input = new { text = "Next: inspect the replay test." } }
                        },
                        usage = new { input_tokens = 1, output_tokens = 1 }
                    }, deadline.Token);
                }
            });
            string answer = await agent.SendAsync("Checkpoint once, then continue.").WaitAsync(deadline.Token);
            await server.WaitAsync(deadline.Token);
            Check(answer == "Resuming the replay test." && requests == 2,
                "Handoff continuation repeated the checkpoint instead of resuming work");
            Check(agent.History().Events().Count(fact => fact.Kind == EventKind.HandoffRecorded) == 1,
                "Continuation recorded duplicate handoffs");
            var checkpoint = JsonCodec.ReadSave(File.ReadAllText(savePath));
            long handoffSequence = agent.History().Events().Single(fact => fact.Kind == EventKind.HandoffRecorded).Seq;
            Check(checkpoint.Sequence == handoffSequence && checkpoint.LatestHandoff == handoffSequence,
                "Completed handoff did not checkpoint before continuation/shutdown");
            Check(checkpoint.Context!.Dialogue.SelectMany(entry => entry.Parts).OfType<HandoffPart>()
                .Single().Text == "Next: inspect the replay test.", "Checkpoint did not persist reduced Context");
        }
        finally
        {
            deadline.Cancel();
            listener.Stop();
            app.Destroy();
            if (Directory.Exists(directory)) Directory.Delete(directory, recursive: true);
            if (server is not null)
            {
                try { await server; }
                catch (OperationCanceledException) when (deadline.IsCancellationRequested) { }
            }
        }
    }

    public static async Task HandoffBatch()
    {
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        var started = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var app = new Application(TextWriter.Null);
        Task? server = null;
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
            // A successful replacement retires older active notes and bounds working events.
            // Until companion results are durable, the old note remains active.
            agent.History().Append(new Event(EventKind.HintReceived, actor: Actor.System,
                parts: [new TextPart("ordinary system instructions")]));
            agent.History().Append(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("old note")]));
            server = Task.Run(async () =>
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
                        Check(request.Contains(expected) && !request.Contains("old note"), "Older handoff survived on the wire");
                        Check(request.Split("[Completed micro_handoff checkpoint]").Length == 2,
                            "Provider request did not contain exactly one checkpoint");
                        using var document = JsonDocument.Parse(request);
                        string messages = document.RootElement.GetProperty("messages").GetRawText();
                        Check(!messages.Contains("\"tool_use\"") && !messages.Contains("\"tool_result\""), "Handoff retained tool traffic");
                        if (round == 1)
                        {
                            var paired = agent.History().Events();
                            long firstNote = paired.Single(fact => fact.Kind == EventKind.HandoffRecorded
                                && fact.Parts.OfType<HandoffPart>().Any(note => note.Text == "first replacement")).Seq;
                            Check(paired.Single(fact => fact.Kind == EventKind.ToolReturned && fact.CallID == "companion").Seq < firstNote,
                                "Companion result recorded after reduction");
                            await ReplyJson(stream, new
                            {
                                content = new object[] { new { type = "tool_use", id = "note-2", name = "micro_handoff", input = new { text = "newest replacement" } } },
                                usage = new { input_tokens = 1, output_tokens = 1 }
                            }, deadline.Token);
                        }
                        else
                        {
                            Check(!request.Contains("first replacement"), "Superseded checkpoint survived on the wire");
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
            await Task.WhenAll(server, turn).WaitAsync(deadline.Token);
            Check(await turn == "handoff complete", "Turn did not finish");
            var events = agent.History().Events();
            Check(events.First().Kind == EventKind.HandoffRecorded
                && events.First().Parts.OfType<HandoffPart>().Single().Text == "first replacement",
                "Working history did not begin at preceding handoff");
            var context = agent.History().View();
            Check(context.Dialogue.SelectMany(entry => entry.Parts).OfType<HandoffPart>().Select(note => note.Text)
                .SequenceEqual(["newest replacement"]),
                "Active context did not retain only the newest checkpoint");
            Check(JsonCodec.Text(context.Dialogue.SelectMany(entry => entry.Parts)).Contains("ordinary system instructions"),
                "Ordinary system content lost");
            Check(events.Count(fact => fact.Kind == EventKind.HandoffRecorded) == 2, "Working history did not retain exactly two handoffs");

            var invalid = new ToolCallPart("invalid-note", agent.Configuration().Target(), "micro_handoff",
                JsonSerializer.SerializeToElement(new { text = "   " }));
            Check((await agent.DispatchAsync(invalid)).IsError, "Blank handoff accepted");
            Check(agent.History().Events().Count(fact => fact.Kind == EventKind.HandoffRecorded) == 2,
                "Failed handoff caused reduction");
            var direct = new ToolCallPart("direct-note", agent.Configuration().Target(), "micro_handoff",
                JsonSerializer.SerializeToElement(new { text = "direct replacement" }));
            Check(!(await agent.DispatchAsync(direct)).IsError, "Direct handoff refused");
            await agent.FlushAsync(deadline.Token);
            Check(agent.History().View().Dialogue.SelectMany(entry => entry.Parts)
                .OfType<HandoffPart>().Select(note => note.Text)
                .SequenceEqual(["direct replacement"]),
                "Direct handoff did not replace all older notes");
        }
        finally
        {
            release.TrySetResult();
            deadline.Cancel();
            listener.Stop();
            try { app.Destroy(); }
            finally
            {
                if (server is not null)
                {
                    try { await server; }
                    catch (OperationCanceledException) when (deadline.IsCancellationRequested) { }
                }
            }
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
            string save = Path.Combine(directory, "save.json");
            agent.StartPersistence(new PersistenceOptions(save, Audit: true));
            history.Append(new Event(EventKind.SkillLoaded, parts: [new SkillPart("manual", "retained manual", ["write_file"], pending: true)]));
            history.Append(new Event(EventKind.MessageReceived, actor: Actor.Human, parts: [new TextPart("dialogue")]));
            string[] mediaTypes = ["image/png", "audio/wav", "video/mp4", "application/pdf"];
            foreach (string mime in mediaTypes)
            {
                var media = new BlobPart(mime, new BlobRef(RefKind.Path, "/absent-handoff-media/" + mime));
                history.Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                    parts: [new TextPart("caption " + mime), media]));
                history.Append(new Event(EventKind.MessageReceived, actor: Actor.Human, parts: [media]));
            }
            history.Append(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("RETIRED_CHECKPOINT_MARKER")]));
            history.Append(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("CURRENT_CHECKPOINT_MARKER")]));
            // Raw media is retained on disk only when audit is requested.
            // The resident window begins at the preceding handoff.
            if (history.Events().SelectMany(fact => fact.Parts).OfType<BlobPart>().Any())
                throw new Exception("Working history retained media before its handoff boundary");
            if (history.Events().Count(fact => fact.Kind == EventKind.HandoffRecorded) != 2)
                throw new Exception("Working history lost recent handoffs");
            var diskFacts = File.ReadLines(save + ".journal.jsonl").Select(JsonCodec.ReadEvent).ToArray();
            if (diskFacts.SelectMany(fact => fact.Parts).OfType<BlobPart>().Count() != 8)
                throw new Exception("Audit journal lost original media");
            var replayed = new Agent(app, agent.Configuration());
            replayed.History().Load(save + ".journal.jsonl");
            history.Save(new PersistenceOptions(save, Audit: true));
            var restored = new Agent(app, agent.Configuration());
            if (!restored.History().LoadSave(save)) throw new Exception("Handoff save absent");
            var context = restored.History().View();
            var parts = context.Dialogue.SelectMany(entry => entry.Parts).ToArray();
            if (!parts.OfType<HandoffPart>().Select(note => note.Text).SequenceEqual(["CURRENT_CHECKPOINT_MARKER"])
                || !parts.OfType<TextPart>().Select(part => part.Text).SequenceEqual(
                    new[] { "dialogue" }.Concat(mediaTypes.Select(mime => "caption " + mime)))
                || !parts.OfType<SkillPart>().Single().Pending
                || restored.ToolEnabled("write_file"))
            {
                throw new Exception("Snapshot changed handoff/dialogue/manual authority");
            }
            foreach (var owner in new[] { agent, restored, replayed })
            {
                var reduced = owner.History().View();
                if (reduced.Dialogue.Any(entry => entry.Parts.IsEmpty)
                    || reduced.Dialogue.SelectMany(entry => entry.Parts).OfType<BlobPart>().Any())
                    throw new Exception("Handoff retained active media or empty messages");
                var projections = Enum.GetValues<Vendor>()
                    .Select(vendor => new Configuration(vendor, "model-a", "http://127.0.0.1:1"))
                    .Append(new Configuration(Vendor.OpenAI, "gpt-6-astra", "http://127.0.0.1:1"));
                foreach (var projection in projections)
                {
                    var renderer = new Agent(app, projection);
                    string wire = renderer.Engine().Render(reduced).ToJsonString();
                    if (wire.Split("[Completed micro_handoff checkpoint]").Length != 2)
                        throw new Exception("Provider projection must contain exactly one completed checkpoint after replay");
                    if (wire.Contains("RETIRED_CHECKPOINT_MARKER")
                        || wire.Split("CURRENT_CHECKPOINT_MARKER").Length != 2)
                        throw new Exception("Provider projection retained old notes or duplicated the newest");
                    if (wire.Contains("absent-handoff-media") || !wire.Contains("caption image/png"))
                        throw new Exception("Provider projection lost caption or retained media locator");
                }
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
