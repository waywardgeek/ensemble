using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class ResultStubbingTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition) throw new Exception(message);
    }

    public static void Selection()
    {
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "model-a", "http://127.0.0.1:1", stubsToolResults: true));
            _ = new Tool(agent, ToolKind.KeepToolResults);
            agent.SetContextByteBudget(1);
            Check(agent.ToolCompactionBudgets() == (1, 1), "Positive target disabled an integer-rounded band");
            agent.SetContextByteBudget(0);
            Check(agent.ToolCompactionBudgets() == (0, 0), "Zero target retained an automatic band");
            agent.SetContextByteBudget(400_000);
            JsonElement args = JsonSerializer.SerializeToElement(new { });
            var big = new ToolResultPart("a", [new TextPart(new string('x', 2000))]);
            var small = new ToolResultPart("b", [new TextPart("small")]);
            Entry response = new(1, Actor.Agent, [new ToolCallPart("a", agent.Configuration().Target(), "large", args), new ToolCallPart("b", agent.Configuration().Target(), "small", args)]);
            Entry first = new(2, Actor.Tool, [big]);
            Entry second = new(3, Actor.Tool, [small]);
            var context = new Context([response, first, second, new Entry(4, Actor.Human, [new TextPart("continue")])]);
            var cut = agent.SelectImmediateCuts(context, 5, 512);
            Check(cut.Length == 1 && cut[0].RedactFrom == 2 && cut[0].RedactTo == 2, "Wrong per-result cut");
            int exact = Encoding.UTF8.GetByteCount(JsonCodec.WritePart(big).ToJsonString());
            Check(agent.SelectImmediateCuts(context, 5, exact).IsEmpty, "Equality must retain the whole result");
            Check(agent.SelectImmediateCuts(context, 3, 0).IsEmpty, "Incomplete presented batch selected");
            Check(agent.SelectImmediateCuts(new Context([response, first]), 5, 0).IsEmpty, "Missing companion accepted");
            Check(agent.SelectImmediateCuts(new Context([response, first, second, second]), 5, 0).IsEmpty, "Duplicate result accepted");
            var consumed = new Context(context.Dialogue.Add(new Entry(6, Actor.Agent, [])));
            Check(agent.SelectImmediateCuts(consumed, 7, 0).IsEmpty, "Empty consuming response did not close the old batch");
            var next = new Context(consumed.Dialogue.Add(new Entry(8, Actor.Agent, [new ToolCallPart("a", agent.Configuration().Target(), "large", args)]))
                .Add(new Entry(9, Actor.Tool, [big])));
            Check(agent.SelectImmediateCuts(next, 10, 512).Single().RedactFrom == 9, "Repeated ID selected an older batch");

            var keep = new Context([new Entry(1, Actor.Agent, [new ToolCallPart("a", agent.Configuration().Target(), "keep_tool_results", args)]), first]);
            Check(agent.SelectImmediateCuts(keep, 3, 0).IsEmpty, "Keep control result selected for stubbing");
            var custom = new Agent(app, agent.Configuration());
            _ = new Tool(custom, "keep_tool_results", "Not the control", args, (job, values, cancellation) => Task.FromResult("custom"));
            Check(custom.SelectImmediateCuts(keep, 3, 0).Length == 1, "Same-named custom tool gained control semantics");
            var ordinary = new Agent(app, new Configuration(Vendor.Anthropic, "model-a", "http://127.0.0.1:1"));
            ordinary.InstallTools();
            Check(ordinary.Tools().All(tool => tool.Name() != "keep_tool_results"), "Opt-out installed keep control");
            try
            {
                _ = new Tool(ordinary, ToolKind.KeepToolResults);
                throw new Exception("Opt-out constructed keep control");
            }
            catch (ArgumentException) { }
            string directory = Path.Combine(Path.GetTempPath(), "keep-skills-" + Guid.NewGuid().ToString("N"));
            Directory.CreateDirectory(Path.Combine(directory, "primary"));
            try
            {
                File.WriteAllText(Path.Combine(directory, "primary", "SKILL.md"),
                    "---\nname: primary\ndescription: Policy fixture\ntype: primary\n---\nTOOLS=$TOOLS\n");
                foreach (var vendor in new[] { Vendor.Anthropic, Vendor.OpenAI, Vendor.Gemini })
                {
                    foreach (bool enabled in new[] { false, true })
                    {
                        var configured = new Agent(app, new Configuration(vendor, "model-a", "http://127.0.0.1:1",
                            deferredTools: vendor == Vendor.Anthropic, stubsToolResults: enabled),
                            skillSpec: new SkillSpec(directory, "primary"));
                        Check(configured.ToolEnabled("keep_tool_results") == enabled, "Startup keep grant differs from selected policy");
                        string wire = configured.Engine().Render(configured.History().View()).ToJsonString();
                        Check(wire.Contains("\"name\":\"keep_tool_results\"") == enabled, "Provider declaration differs from startup grant");
                        Check(configured.SkillManager()!.CatalogTools().Contains("keep_tool_results") == enabled,
                            "Deferred catalog differs from startup grant");
                        JsonElement saved = configured.Engine().SavedConfiguration(configured.History().View());
                        Check(saved.GetProperty("stubs_tool_results").GetBoolean() == enabled
                            && saved.GetProperty("context_budget_bytes").GetInt32() == 400_000, "Saved policy is not documentary current state");
                    }
                }
            }
            finally
            {
                Directory.Delete(directory, true);
            }
        }
        finally
        {
            app.Destroy();
        }
    }
}

internal static partial class ActorTests
{
    public static async Task OnceOnlyResultStubbing()
    {
        foreach (string scenario in new[] { "ordinary", "stub", "keep", "invalid", "custom", "disabled", "keep-large" })
        {
            var listener = new TcpListener(IPAddress.Loopback, 0);
            listener.Start();
            using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(30));
            var held = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            string directory = Path.Combine(Path.GetTempPath(), "once-only-" + Guid.NewGuid().ToString("N"));
            Directory.CreateDirectory(directory);
            var app = new Application(TextWriter.Null);
            string firstPayload = "FIRST-WHOLE-" + new string('a', scenario == "keep-large" ? 120000 : 8000);
            string nextPayload = "NEXT-WHOLE-" + new string('b', 8000);
            bool optIn = scenario != "ordinary";
            bool keepFirst = scenario is "ordinary" or "keep" or "disabled";
            bool stubPolicy = optIn && scenario != "disabled";
            try
            {
                var config = new Configuration(Vendor.Anthropic, "model-a",
                    $"http://127.0.0.1:{((IPEndPoint)listener.LocalEndpoint).Port}", stubsToolResults: optIn);
                Agent Create()
                {
                    var owner = new Agent(app, config);
                    var schema = JsonSerializer.SerializeToElement(new { type = "object", properties = new { }, additionalProperties = false });
                    _ = new Tool(owner, "large_result", "First real output", schema,
                        (job, args, cancellation) => Task.FromResult(firstPayload), waitForCompletion: true);
                    _ = new Tool(owner, "small_result", "Small real output", schema,
                        (job, args, cancellation) => Task.FromResult("SMALL-RETAINED"), waitForCompletion: true);
                    _ = new Tool(owner, "next_result", "Held companion output", schema,
                        async (job, args, cancellation) =>
                        {
                            held.TrySetResult();
                            await release.Task.WaitAsync(cancellation);
                            return nextPayload;
                        }, waitForCompletion: true);
                    if (scenario == "custom")
                    {
                        _ = new Tool(owner, "keep_tool_results", "Custom, no policy authority", schema,
                            (job, args, cancellation) => Task.FromResult("Not the keep control"), waitForCompletion: true);
                    }
                    else if (optIn)
                    {
                        _ = new Tool(owner, ToolKind.KeepToolResults);
                    }
                    if (scenario == "disabled")
                    {
                        owner.SetContextByteBudget(0);
                    }
                    owner.StartPersistence(new PersistenceOptions(Path.Combine(directory, "save.json"), RetainLog: 0));
                    return owner;
                }
                var agent = Create();
                var server = Task.Run(async () =>
                {
                    for (int round = 0; round < 5; round++)
                    {
                        using var client = await listener.AcceptTcpClientAsync(deadline.Token);
                        using var stream = client.GetStream();
                        string request = await ReadRequest(stream, deadline.Token);
                        var content = new List<object>();
                        if (round == 0)
                        {
                            content.Add(new { type = "tool_use", id = "a", name = "large_result", input = new { } });
                            content.Add(new { type = "tool_use", id = "s", name = "small_result", input = new { } });
                        }
                        else if (round == 1)
                        {
                            Check(request.Contains(firstPayload) && request.Contains("SMALL-RETAINED"), scenario + ": first consuming request was stubbed early");
                            if (scenario is "keep" or "invalid" or "custom" or "keep-large")
                            {
                                object input = scenario == "invalid" ? new { handle = 7 } : new { };
                                content.Add(new { type = "tool_use", id = "k", name = "keep_tool_results", input });
                            }
                            content.Add(new { type = "tool_use", id = "b", name = "next_result", input = new { } });
                        }
                        else
                        {
                            Check(request.Contains(firstPayload) == keepFirst, scenario + ": old batch keep/stub decision changed");
                            Check(request.Contains("SMALL-RETAINED") == (scenario != "keep-large"), scenario + ": small result did not follow its whole-batch ladder boundary");
                            Check(request.Contains(nextPayload) == (round == 2 || !stubPolicy), scenario + ": next batch not independently decided");
                            // An empty successful response is still the once-only
                            // boundary and must survive zero-log persistence.
                            if (round != 3)
                            {
                                content.Add(new { type = "text", text = "complete" });
                            }
                        }
                        await ReplyJson(stream, new { content, usage = new { input_tokens = 1, output_tokens = 1 } }, deadline.Token);
                    }
                });
                try
                {
                    var turn = agent.SendAsync("first");
                    await held.Task.WaitAsync(deadline.Token);
                    Check(!agent.History().Events().Any(fact => fact.Kind == EventKind.Redacted), scenario + ": cut before held companion completed");
                    release.TrySetResult();
                    Check(await turn.WaitAsync(deadline.Token) == "complete", scenario + ": initial turn failed");
                    var events = agent.History().Events();
                    Check(events.Any(fact => fact.Kind == EventKind.ToolReturned && JsonCodec.Text(fact.Parts).Contains(firstPayload)), "Canonical payload lost");
                    if (scenario == "keep-large")
                    {
                        long oldResult = events.Single(fact => fact.Kind == EventKind.ToolReturned && fact.CallID == "a").Seq;
                        Check(events.Any(fact => fact.Kind == EventKind.Redacted && fact.Reason == "tool result budget"
                            && fact.RedactFrom <= oldResult && fact.RedactTo >= oldResult), "Keep bypassed the ordinary byte ladder");
                        Check(!events.Any(fact => fact.Kind == EventKind.Redacted && fact.Reason == "once-only tool result stubbing"
                            && fact.RedactFrom <= oldResult && fact.RedactTo >= oldResult), "Keep did not veto immediate policy");
                    }
                    if (scenario == "invalid")
                    {
                        Check(events.Any(fact => fact.Kind == EventKind.ToolReturned && fact.CallID == "k" && fact.IsError), "Bad keep arguments succeeded");
                    }
                    await agent.SendAsync("text-only later").WaitAsync(deadline.Token);
                    await agent.ShutdownAsync().WaitAsync(deadline.Token);
                    agent.Destroy();
                    var restored = Create();
                    Check(await restored.SendAsync("after restart").WaitAsync(deadline.Token) == "complete", scenario + ": restart failed");
                    await server.WaitAsync(deadline.Token);
                    await restored.ShutdownAsync().WaitAsync(deadline.Token);
                }
                catch
                {
                    // Surface a boundary-fake assertion, not its secondary HTTP
                    // disconnect, when the fixture itself rejected a request.
                    if (server.IsFaulted)
                    {
                        await server;
                    }
                    throw;
                }
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

internal static partial class ActorTests
{
    public static async Task InterruptedKeep()
    {
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        var started = new TaskCompletionSource<Job>(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "model-a",
                $"http://127.0.0.1:{((IPEndPoint)listener.LocalEndpoint).Port}", stubsToolResults: true));
            _ = new Tool(agent, ToolKind.KeepToolResults);
            var schema = JsonSerializer.SerializeToElement(new { type = "object", properties = new { }, additionalProperties = false });
            string payload = new string('x', 8000);
            _ = new Tool(agent, "large_result", "First batch", schema,
                (job, args, cancellation) => Task.FromResult(payload), waitForCompletion: true);
            _ = new Tool(agent, "held", "Held consuming-response companion", schema, async (job, args, cancellation) =>
            {
                started.TrySetResult(job);
                await release.Task.WaitAsync(cancellation);
                return "independent completion";
            }, waitForCompletion: true);
            var server = Task.Run(async () =>
            {
                for (int round = 0; round < 2; round++)
                {
                    using var client = await listener.AcceptTcpClientAsync(deadline.Token);
                    using var stream = client.GetStream();
                    string request = await ReadRequest(stream, deadline.Token);
                    object[] content;
                    if (round == 0)
                    {
                        content = [new { type = "tool_use", id = "a", name = "large_result", input = new { } }];
                    }
                    else
                    {
                        Check(request.Contains(payload), "First consuming request did not contain whole result");
                        content = [new { type = "tool_use", id = "held", name = "held", input = new { } },
                            new { type = "tool_use", id = "keep", name = "keep_tool_results", input = new { } }];
                    }
                    await ReplyJson(stream, new { content, usage = new { input_tokens = 1, output_tokens = 1 } }, deadline.Token);
                }
            });
            var turn = agent.SendAsync("interrupt before keep");
            var job = await started.Task.WaitAsync(deadline.Token);
            Check(!agent.History().Events().Any(fact => fact.Kind == EventKind.Redacted), "Cut preceded companion completion");
            await agent.PostAsync(new AgentInput(InputKind.InterruptNow)).WaitAsync(deadline.Token);
            try
            {
                await turn.WaitAsync(deadline.Token);
                throw new Exception("Interrupted turn succeeded");
            }
            catch (OperationCanceledException) { }
            await agent.FlushAsync(deadline.Token);
            var events = agent.History().Events();
            long cut = events.Single(fact => fact.Kind == EventKind.Redacted).Seq;
            Check(events.Single(fact => fact.Kind == EventKind.ToolReturned && fact.CallID == "keep") is var refused
                && refused.IsError && refused.Seq < cut, "Undispatched keep vetoed policy or was not paired before cut");
            Check(events.Single(fact => fact.Kind == EventKind.ToolReturned && fact.CallID == "held").Seq < cut,
                "Started companion outcome was not recorded before cut");
            release.TrySetResult();
            var completion = await job.CompletionAsync(deadline.Token);
            Check(!completion.IsError && JsonCodec.Text(completion.Parts) == "independent completion", "Interrupt killed independent Job");
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
