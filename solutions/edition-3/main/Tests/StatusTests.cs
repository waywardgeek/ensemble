using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    public static async Task Status(string directory)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4o",
                listener.Prefixes.Single()), Path.Combine(directory, "status-jobs"));
            agent.InstallTools();
            var other = new Agent(app, new Configuration(Vendor.OpenAI, "other", listener.Prefixes.Single()));
            _ = new Tool(other, ToolKind.AgentStatus);

            async Task<string> Read(Agent owner)
            {
                using var args = JsonDocument.Parse("{}");
                var call = new ToolCallPart("status", owner.Configuration().Target(), "agent_status", args.RootElement);
                var result = await owner.DispatchAsync(call).WaitAsync(TimeSpan.FromSeconds(10));
                Check(!result.IsError, "status is a successful real tool dispatch");
                return JsonCodec.Text(result.Parts);
            }

            string empty = await Read(agent);
            Check(empty == "model: gpt-4o\n"
                + "last response: unknown in, unknown out, unknown cache read, unknown cache write\n"
                + "session total: 0 in, 0 out, 0 cache read, 0 cache write\n"
                + "cache hit rate: unknown\ncost: $0 (estimate)", "empty session uses the specified labels and unknown latest values");
            Check(agent.Engine().Meter().Latest is null, "status performs no inference and adds no usage");

            async Task Send(string model)
            {
                var send = agent.Engine().SendAsync("{}");
                var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
                await Reply(request, JsonSerializer.Serialize(new
                {
                    model, service_tier = "default",
                    choices = new[] { new { message = new { content = "ok" } } },
                    usage = new { prompt_tokens = 120, prompt_tokens_details = new { cached_tokens = 20 }, completion_tokens = 10 }
                }));
                await send;
            }

            await Send("gpt-4o-mini-2024-07-18");
            await Send("gpt-4o-2024-08-06");
            MeterSnapshot before = agent.Engine().Meter();
            string status = await Read(agent);
            Check(status.Contains("last response: 100 in, 10 out, 20 cache read, 0 cache write")
                && status.Contains("session total: 200 in, 20 out, 40 cache read, 0 cache write"),
                "latest and session disjoint counts come from completed requests");
            Check(status.Contains("cache hit rate: 16.7%")
                && status.EndsWith("cost: $0.0003975 (estimate)")
                && before.EstimatedUsd() == 0.0003975m, "derive mixed-model estimate and session hit rate");
            Check(ReferenceEquals(before.Latest, agent.Engine().Meter().Latest), "reading status never increments the meter");
            Check((await Read(other)).Contains("session total: 0 in, 0 out, 0 cache read, 0 cache write"), "Agent meters remain isolated");

            await Send("unpriced-provider-model");
            string unknown = await Read(agent);
            Check(unknown.Contains("cost: unknown")
                && unknown.Contains("session total: 300 in, 30 out, 60 cache read, 0 cache write"),
                "unknown price preserves measured usage");
            using var invalid = JsonDocument.Parse("{\"model\":\"do-not-switch\"}");
            var refusal = await agent.DispatchAsync(new ToolCallPart("invalid-status", agent.Configuration().Target(),
                "agent_status", invalid.RootElement));
            Check(refusal.IsError && agent.Configuration().Model == "gpt-4o", "status accepts no model-setting argument");
        }
        finally
        {
            app.Destroy();
        }
    }
    public static async Task StatusTurn(string directory)
    {
        using var listener = Listen();
        string skills = Path.Combine(directory, "status-skills");
        string primary = Path.Combine(skills, "primary");
        Directory.CreateDirectory(primary);
        File.WriteAllText(Path.Combine(primary, "SKILL.md"),
            "---\nname: primary\ndescription: Status only\ntype: primary\ntools: agent_status\n---\nReport measured status, not invented usage.");
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4o-mini",
                listener.Prefixes.Single(), disableStreaming: true),
                Path.Combine(directory, "status-turn-jobs"), skillSpec: new SkillSpec(skills, "primary"));
            await agent.SkillManager()!.InitializeAsync();
            Check(agent.ToolEnabled("agent_status") && !agent.ToolEnabled("read_file"), "status obeys primary grants");
            var turn = agent.SubmitAsync([new TextPart("Report your status.")]);
            var first = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            await Reply(first, """
                {"model":"gpt-4o-mini-2024-07-18","service_tier":"default",
                 "choices":[{"message":{"content":"I will read the measured status.","tool_calls":[
                   {"id":"status-1","type":"function","function":{"name":"agent_status","arguments":"{}"}}]}}],
                 "usage":{"prompt_tokens":120,"prompt_tokens_details":{"cached_tokens":20},"completion_tokens":10}}
                """);
            var second = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            using var reader = new StreamReader(second.Request.InputStream);
            using var request = JsonDocument.Parse(await reader.ReadToEndAsync());
            string report = request.RootElement.GetProperty("messages").EnumerateArray()
                .Single(message => message.GetProperty("role").GetString() == "tool")
                .GetProperty("content").GetString()!;
            Check(report.Contains("last response: 100 in, 10 out, 20 cache read, 0 cache write")
                && report.Contains("session total: 100 in, 10 out, 20 cache read, 0 cache write"),
                "tool reads its requesting response before the final answer");
            await Reply(second, """
                {"model":"gpt-4o-mini-2024-07-18","service_tier":"default",
                 "choices":[{"message":{"content":"Status delivered."}}],
                 "usage":{"prompt_tokens":50,"completion_tokens":5}}
                """);
            await turn.WaitAsync(TimeSpan.FromSeconds(10));
            Check(agent.Engine().Meter().Totals.Input == 150 && agent.Engine().Meter().Latest!.Counts.Input == 50,
                "final answer adds usage after the status result");
            Check(agent.History().Events().Any(value => value.Kind == EventKind.ToolReturned
                && value.CallID == "status-1" && JsonCodec.Text(value.Parts) == report),
                "canonical correlated status result is not rewritten by later usage");
        }
        finally
        {
            app.Destroy();
        }
    }

}
