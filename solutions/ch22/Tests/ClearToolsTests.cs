using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    public static async Task ManualToolClear()
    {
        foreach (bool interrupted in new[] { false, true })
        {
            var listener = new TcpListener(IPAddress.Loopback, 0);
            listener.Start();
            using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
            var started = new TaskCompletionSource<Job>(TaskCreationOptions.RunContinuationsAsynchronously);
            var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            string directory = Path.Combine(Path.GetTempPath(), "manual-clear-" + Guid.NewGuid().ToString("N"));
            Directory.CreateDirectory(directory);
            var app = new Application(TextWriter.Null);
            try
            {
                var agent = new Agent(app, new Configuration(Vendor.Anthropic, "model-a",
                    $"http://127.0.0.1:{((IPEndPoint)listener.LocalEndpoint).Port}"));
                var history = agent.History();
                // Pending manuals are instruction snapshots, not disposable
                // tool traffic. A clear must not retire or reactivate them.
                history.Append(new Event(EventKind.SkillLoaded,
                    parts: [new SkillPart("manual", "retained manual", ["held"], pending: true)]));
                history.Append(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("retained note")]));
                history.Append(new Event(EventKind.HintReceived, actor: Actor.System,
                    parts: [new TextPart("ordinary system instruction")]));
                // A separate active grant makes this tool usable without
                // changing the pending manual's authority or stored state.
                history.Append(new Event(EventKind.SkillLoaded,
                    parts: [new SkillPart("active", "active manual", ["held", "tail"])]));
                var schema = JsonSerializer.SerializeToElement(new { type = "object", properties = new { }, additionalProperties = false });
                _ = new Tool(agent, "held", "Held real job", schema, async (job, args, cancellation) =>
                {
                    started.TrySetResult(job);
                    await release.Task.WaitAsync(cancellation);
                    return "original job output";
                }, waitForCompletion: true);
                bool tailRan = false;
                _ = new Tool(agent, "tail", "Tail must not execute after interrupt", schema, (job, args, cancellation) =>
                {
                    tailRan = true;
                    return Task.FromResult("tail");
                });
                var server = Task.Run(async () =>
                {
                    using (var client = await listener.AcceptTcpClientAsync(deadline.Token))
                    using (var stream = client.GetStream())
                    {
                        _ = await ReadRequest(stream, deadline.Token);
                        var content = new List<object>
                        {
                            new { type = "text", text = "visible explanation" },
                            new { type = "tool_use", id = "held-id", name = "held", input = new { } }
                        };
                        if (interrupted)
                        {
                            content.Add(new { type = "tool_use", id = "tail-id", name = "tail", input = new { } });
                        }
                        await ReplyJson(stream, new { content, usage = new { input_tokens = 1, output_tokens = 1 } }, deadline.Token);
                    }
                    if (!interrupted)
                    {
                        using var client = await listener.AcceptTcpClientAsync(deadline.Token);
                        using var stream = client.GetStream();
                        string request = await ReadRequest(stream, deadline.Token);
                        using var parsed = JsonDocument.Parse(request);
                        string messages = parsed.RootElement.GetProperty("messages").GetRawText();
                        Check(!messages.Contains("tool_use") && !messages.Contains("tool_result"), "Manual clear left unpaired request traffic");
                        Check(request.Contains("visible explanation") && request.Contains("retained note")
                            && request.Contains("retained manual"), "Manual clear removed retained context");
                        await Reply(stream, "done", deadline.Token);
                    }
                });
                var turn = agent.SendAsync("keep the dialogue");
                var job = await started.Task.WaitAsync(deadline.Token);
                await agent.PostAsync(new AgentInput(InputKind.Pause));
                await agent.PostAsync(new AgentInput(InputKind.ClearTools));
                Check(!history.Events().Any(fact => fact.Reason == "manual tool clear"), "Clear ran before held outcome");
                if (interrupted)
                {
                    await agent.PostAsync(new AgentInput(InputKind.InterruptNow));
                    try
                    {
                        await turn.WaitAsync(deadline.Token);
                        throw new Exception("Interrupted turn reported success");
                    }
                    catch (OperationCanceledException) { }
                }
                else
                {
                    release.TrySetResult();
                }
                while (!history.Events().Any(fact => fact.Reason == "manual tool clear"))
                {
                    await Task.Delay(10, deadline.Token);
                }
                var cut = history.Events().Single(fact => fact.Reason == "manual tool clear");
                Check(cut.Level == Redaction.Tool && history.Events().Where(fact => fact.Kind == EventKind.ToolReturned)
                    .All(fact => fact.Seq < cut.Seq), "Cut preceded batch outcomes");
                Check(history.Events().Count(fact => fact.Kind == EventKind.ToolReturned) == (interrupted ? 2 : 1), "Missing paired outcomes");
                Check(!tailRan, "Interrupted tail executed");
                Check(!history.View().Dialogue.SelectMany(entry => entry.Parts).Any(part => part is ToolCallPart or ToolResultPart), "Tool traffic survived clear");
                Check(history.View().Dialogue.SelectMany(entry => entry.Parts).OfType<SkillPart>().Single(part => part.Name == "manual").Pending,
                    "Clear retired or reactivated pending manual");
                release.TrySetResult();
                var completion = await job.CompletionAsync(deadline.Token);
                Check(!completion.IsError && JsonCodec.Text(completion.Parts) == "original job output", "Clear/interrupt killed independent Job");
                await agent.PostAsync(new AgentInput(InputKind.Resume));
                if (!interrupted)
                {
                    await turn.WaitAsync(deadline.Token);
                }
                await agent.FlushAsync(deadline.Token);
                await agent.PostAsync(new AgentInput(InputKind.ClearTools));
                await agent.FlushAsync(deadline.Token);
                Check(history.Events().Count(fact => fact.Reason == "manual tool clear") == 1, "Empty repeat manufactured another cut");
                string save = Path.Combine(directory, "save.json");
                history.Save(new PersistenceOptions(save, RetainLog: 0));
                var restored = new Agent(app, agent.Configuration());
                Check(restored.History().LoadSave(save), "Clear snapshot absent");
                var parts = restored.History().View().Dialogue.SelectMany(entry => entry.Parts).ToArray();
                Check(!parts.Any(part => part is ToolCallPart or ToolResultPart)
                    && parts.OfType<HandoffPart>().Single().Text == "retained note"
                    && parts.OfType<TextPart>().Any(part => part.Text == "keep the dialogue")
                    && parts.OfType<TextPart>().Any(part => part.Text == "ordinary system instruction"), "Snapshot replay changed manual clear semantics");
                await server.WaitAsync(deadline.Token);
            }
            finally
            {
                release.TrySetResult();
                app.Destroy();
                listener.Stop();
                Directory.Delete(directory, true);
            }
        }
    }
}
