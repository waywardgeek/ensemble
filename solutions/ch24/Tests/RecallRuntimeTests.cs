using System.Net;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    private const string RecallQuery = "  Please explain the quartz migration decision, its rationale and remaining deployment uncertainty.\r\n";
    private const string RecallFact = "quartz migration uses explicit approval; deployment remains unverified";

    private static Agent RecallActor(Application app, HttpListener listener, string home,
        bool mechanical = false)
    {
        Directory.CreateDirectory(Path.Combine(home, "docs"));
        File.WriteAllText(Path.Combine(home, "docs", "decision.md"), RecallFact);
        File.WriteAllText(Path.Combine(home, "docs", "other.md"), "unrelated garden botany");
        var agent = new Agent(app, new Configuration(Vendor.OpenAI, "main-route",
            listener.Prefixes.Single(), disableStreaming: true), dataDirectory: home);
        agent.SetRecallSettings(new RecallSettings(disableJudge: mechanical,
            judgeModel: "judge-route", minScore: 0.01));
        return agent;
    }

    private static async Task<JsonObject> ReadRecallRequest(HttpListenerContext request)
    {
        using var reader = new StreamReader(request.Request.InputStream);
        return JsonNode.Parse(await reader.ReadToEndAsync())!.AsObject();
    }

    private static async Task RecallCanceled(Task<string> turn)
    {
        try
        {
            await turn.WaitAsync(TimeSpan.FromSeconds(3));
        }
        catch (OperationCanceledException)
        {
            return;
        }
        throw new Exception("interrupted recall did not cancel its turn reply");
    }

    public static async Task RecallAdmission(string temp)
    {
        foreach (bool interactive in new[] { false, true })
        {
            using var listener = Listen();
            var app = new Application(TextWriter.Null);
            try
            {
                var agent = RecallActor(app, listener, Path.Combine(temp, "recall-admit-" + interactive));
                Task<string> turn = interactive
                    ? agent.PostAsync(new AgentInput(InputKind.Chat, RecallQuery))
                    : agent.SubmitAsync([new TextPart(RecallQuery)]);
                var judge = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
                var judgeBody = await ReadRecallRequest(judge);
                Check(judgeBody["model"]!.GetValue<string>() == "judge-route", "main request preceded judge");
                Check(!judgeBody.ContainsKey("tools"), "judge gained tools");
                var data = JsonNode.Parse(judgeBody["messages"]!.AsArray().Last()!["content"]!.GetValue<string>())!;
                Check(data["query"]!.GetValue<string>() == RecallQuery, "query changed original human text");
                Check(data["candidates"]!.AsArray().Count == 1, "wrong rooted candidates");
                await SendJudgeReply(judge, Surface.ChatCompletions, "[0]", "judge-route");
                var main = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
                var body = await ReadRecallRequest(main);
                Check(body["model"]!.GetValue<string>() == "main-route", "judge changed main model");
                var users = body["messages"]!.AsArray().Where(value => value!["role"]!.GetValue<string>() == "user").ToArray();
                Check(users.Length == 2 && users[0]!["content"]!.GetValue<string>() == RecallQuery
                    && users[1]!["content"]!.GetValue<string>().Contains(RecallFact, StringComparison.Ordinal),
                    "recall not separate user data after exact human bytes");
                var facts = agent.History().Events();
                int human = Array.FindIndex(facts.ToArray(), value => value.Parts.OfType<TextPart>().Any(part => part.Text == RecallQuery));
                int recall = Array.FindIndex(facts.ToArray(), value => value.Parts.OfType<RecallPart>().Any());
                Check(human >= 0 && recall > human, "actor did not durably order recall after input");
                await SendJudgeReply(main, Surface.ChatCompletions, "answer", "main-route");
                Check(await turn.WaitAsync(TimeSpan.FromSeconds(5)) == "answer", "main answer lost");
                Check(agent.History().View().Dialogue.SelectMany(entry => entry.Parts).OfType<RecallPart>().Count() == 1,
                    "wrong recall entry count");
                // Retained hashes suppress another judge, not just another attachment.
                var again = agent.SubmitAsync([new TextPart(RecallQuery)]);
                var second = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
                Check((await ReadRecallRequest(second))["model"]!.GetValue<string>() == "main-route",
                    "retained recall triggered judge again");
                await SendJudgeReply(second, Surface.ChatCompletions, "again", "main-route");
                await again.WaitAsync(TimeSpan.FromSeconds(5));
            }
            finally { app.Destroy(); }
        }
    }

    public static async Task RecallResponsive(string temp)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = RecallActor(app, listener, Path.Combine(temp, "recall-pause"));
            var turn = agent.SubmitAsync([new TextPart(RecallQuery)]);
            var judge = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
            await ReadRecallRequest(judge);
            await agent.PostAsync(new AgentInput(InputKind.Pause)).WaitAsync(TimeSpan.FromSeconds(2));
            await agent.PostAsync(new AgentInput(InputKind.Hint, "keep uncertainty explicit")).WaitAsync(TimeSpan.FromSeconds(2));
            // Settings affect future turns, not the already captured selection.
            agent.SetRecallSettings(new RecallSettings(disabled: true));
            using var timeout = new CancellationTokenSource(TimeSpan.FromSeconds(5));
            var recorded = agent.WaitAsync(value => value.Fact?.Parts.OfType<RecallPart>().Any() == true, timeout.Token);
            var nextHttp = listener.GetContextAsync();
            await SendJudgeReply(judge, Surface.ChatCompletions, "[0]", "judge-route");
            await recorded;
            await agent.WorkStatusAsync();
            Check(!nextHttp.IsCompleted && !turn.IsCompleted, "pause allowed main request or turn completion");
            await agent.PostAsync(new AgentInput(InputKind.Resume));
            var main = await nextHttp.WaitAsync(TimeSpan.FromSeconds(5));
            Check((await ReadRecallRequest(main)).ToJsonString().Contains("keep uncertainty explicit", StringComparison.Ordinal),
                "hint during recall lost");
            await SendJudgeReply(main, Surface.ChatCompletions, "answer", "main-route");
            await turn.WaitAsync(TimeSpan.FromSeconds(5));
        }
        finally { app.Destroy(); }

        using var other = Listen();
        var secondApp = new Application(TextWriter.Null);
        try
        {
            var agent = RecallActor(secondApp, other, Path.Combine(temp, "recall-late-duplicate"));
            var turn = agent.SubmitAsync([new TextPart(RecallQuery)]);
            var judge = await other.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
            await ReadRecallRequest(judge);
            await agent.PostAsync(new AgentInput(InputKind.Hint, RecallFact));
            await SendJudgeReply(judge, Surface.ChatCompletions, "[0]", "judge-route");
            var main = await other.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
            await ReadRecallRequest(main);
            Check(!agent.History().Events().Any(value => value.Parts.OfType<RecallPart>().Any()),
                "late exact hint failed publication-time duplicate check");
            await SendJudgeReply(main, Surface.ChatCompletions, "answer", "main-route");
            await turn.WaitAsync(TimeSpan.FromSeconds(5));
        }
        finally { secondApp.Destroy(); }
    }

    public static async Task RecallInterruption(string temp)
    {
        foreach (var kind in new[] { InputKind.InterruptNow, InputKind.InterruptAfterTool })
        {
            using var listener = Listen();
            var app = new Application(TextWriter.Null);
            try
            {
                var agent = RecallActor(app, listener, Path.Combine(temp, "recall-cancel-" + kind));
                var first = agent.SubmitAsync([new TextPart(RecallQuery)]);
                var stalled = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
                await ReadRecallRequest(stalled);
                var second = agent.SubmitAsync([new TextPart("next turn")]);
                await agent.PostAsync(new AgentInput(kind)).WaitAsync(TimeSpan.FromSeconds(2));
                await RecallCanceled(first);
                var main = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
                Check((await ReadRecallRequest(main))["model"]!.GetValue<string>() == "main-route", "canceled judge leaked into next turn");
                Check(!agent.History().Events().Any(value => value.Parts.OfType<RecallPart>().Any()), "canceled recall was published");
                stalled.Response.Close();
                await SendJudgeReply(main, Surface.ChatCompletions, "next", "main-route");
                Check(await second.WaitAsync(TimeSpan.FromSeconds(5)) == "next", "queued turn lost");
            }
            finally { app.Destroy(); }
        }
        using var shutdownPeer = Listen();
        var shutdownApp = new Application(TextWriter.Null);
        try
        {
            var agent = RecallActor(shutdownApp, shutdownPeer, Path.Combine(temp, "recall-shutdown"));
            var turn = agent.SubmitAsync([new TextPart(RecallQuery)]);
            var stalled = await shutdownPeer.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
            await ReadRecallRequest(stalled);
            await agent.ShutdownAsync().WaitAsync(TimeSpan.FromSeconds(3));
            try
            {
                await turn.WaitAsync(TimeSpan.FromSeconds(3));
                throw new Exception("shutdown did not reject its active turn reply");
            }
            catch (ObjectDisposedException) { }
            Check(!agent.History().Events().Any(value => value.Parts.OfType<RecallPart>().Any()), "shutdown published recall");
            stalled.Response.Close();
        }
        finally { shutdownApp.Destroy(); }
    }

    public static async Task RecallGuards(string temp)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = RecallActor(app, listener, Path.Combine(temp, "recall-mechanical"), mechanical: true);
            var turn = agent.SubmitAsync([new TextPart(RecallQuery)]);
            var main = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
            var body = await ReadRecallRequest(main);
            Check(body["model"]!.GetValue<string>() == "main-route" && body.ToJsonString().Contains(RecallFact),
                "mechanical path called judge or lost recall");
            await SendJudgeReply(main, Surface.ChatCompletions, "answer", "main-route");
            await turn.WaitAsync(TimeSpan.FromSeconds(5));
        }
        finally { app.Destroy(); }

        var failureApp = new Application(TextWriter.Null);
        try
        {
            var agent = RecallActor(failureApp, listener, Path.Combine(temp, "recall-read-failure"));
            File.WriteAllBytes(Path.Combine(agent.DataDirectory(), "docs", "decision.md"), [0xff]);
            try
            {
                await agent.SubmitAsync([new TextPart(RecallQuery)]).WaitAsync(TimeSpan.FromSeconds(5));
                throw new Exception("invalid corpus was silently treated as empty");
            }
            catch (System.Text.DecoderFallbackException) { }
            Check(!agent.History().Events().Any(value => value.Parts.OfType<RecallPart>().Any()), "failed scan published recall");
        }
        finally { failureApp.Destroy(); }
    }

    public static void RecallInstructionProjection()
    {
        foreach (Surface surface in Enum.GetValues<Surface>())
        {
            var app = new Application(TextWriter.Null);
            try
            {
                Vendor vendor = surface == Surface.Messages ? Vendor.Anthropic
                    : surface == Surface.GenerateContent ? Vendor.Gemini : Vendor.OpenAI;
                var agent = new Agent(app, new Configuration(vendor, "test", "http://127.0.0.1:1",
                    primaryInstructions: "", surface: surface));
                foreach (bool tools in new[] { false, true })
                {
                    if (tools) _ = new Tool(agent, ToolKind.ReadFile);
                    foreach (bool manuals in new[] { false, true })
                    {
                        var context = manuals ? new Context([
                            new Entry(1, Actor.System, [new SkillPart("manual", RecallFact, ["read_file"])])]) : new Context();
                        string projected = agent.Engine().ProjectInstructions(context);
                        var request = agent.Engine().Render(context);
                        string actual = surface switch
                        {
                            Surface.Messages => request["system"]!.GetValue<string>(),
                            Surface.Responses => request["instructions"]!.GetValue<string>(),
                            Surface.GenerateContent => request["systemInstruction"]!["parts"]![0]!["text"]!.GetValue<string>(),
                            _ => request["messages"]![0]!["content"]!.GetValue<string>()
                        };
                        Check(projected == actual, "instruction projection drift: " + surface);
                        if (manuals)
                        {
                            var hit = new RecallHit(new RecallSnippet(RecallSource.Docs, "docs/a.md", "", RecallFact), 1);
                            Check(RecallText.Available([hit], context, projected).IsEmpty, "manual failed duplicate suppression");
                        }
                    }
                }
            }
            finally { app.Destroy(); }
        }
    }
}
