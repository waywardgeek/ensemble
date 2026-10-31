using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    public static async Task ChildStatus(string directory,
        [System.Runtime.CompilerServices.CallerFilePath] string sourceFile = "")
    {
        using var listener = Listen();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        var app = new Application(TextWriter.Null);
        var config = new Configuration(Vendor.Anthropic, "selected-alias", listener.Prefixes.Single());
        string output = Path.Combine(directory, "child-status-meter");
        string skills = Path.GetFullPath(Path.Combine(Path.GetDirectoryName(sourceFile)!, "..", "Skills"));
        var root = new Agent(app, config, outputDirectory: output, skillSpec: new SkillSpec(skills, "ensemble"));
        var peer = new Agent(app, config, outputDirectory: output, name: "peer");
        var child = new Agent(app, root, "worker", config, outputDirectory: output);
        _ = new Agent(app, peer, "foreign", config, outputDirectory: output);
        _ = new Agent(app, child, "nested", config, outputDirectory: output);
        try
        {
            await root.SkillManager()!.InitializeAsync(deadline.Token);
            Check(root.ToolEnabled("get_sub_agent_status"), "Shipped ensemble skill does not grant child status.");
            Check(root.Engine().Render(root.History().View()).ToJsonString().Contains("get_sub_agent_status", StringComparison.Ordinal),
                "Shipped child status grant is absent from provider projection.");
            async Task<ToolResultPart> Call(string id)
            {
                return await root.DispatchAsync(new ToolCallPart(Guid.NewGuid().ToString("N"), config.Target(),
                    "get_sub_agent_status", JsonSerializer.SerializeToElement(new { agent_id = id })), deadline.Token);
            }
            static string Text(ToolResultPart result)
            {
                string text = JsonCodec.Text(result.Parts);
                return text.StartsWith("job ", StringComparison.Ordinal)
                    ? text[(text.IndexOf('\n') + 1)..] : text;
            }
            async Task<JsonElement> Status()
            {
                ToolResultPart reply = await Call("worker");
                Check(!reply.IsError, "Own-child status was refused: " + Text(reply));
                return JsonSerializer.Deserialize<JsonElement>(Text(reply));
            }
            JsonElement before = await Status();
            Check(before.GetProperty("activity").GetString() == "idle"
                && before.GetProperty("work_phase").GetString() == "idle"
                && !before.GetProperty("submitted").GetBoolean()
                && before.GetProperty("last_response_usage").ValueKind == JsonValueKind.Null
                && before.GetProperty("estimated_cost_usd").GetDecimal() == 0,
                "Unused child status fabricated activity, submission, or usage.");
            ToolResultPart unknown = await Call("missing");
            foreach (string outside in new[] { "foreign", "nested", "peer" })
            {
                ToolResultPart refused = await Call(outside);
                Check(unknown.IsError && refused.IsError && Text(unknown) == Text(refused),
                    "Status exposed a peer, foreign child, or non-direct descendant.");
            }

            // Hold a real HTTP send: the actor's counted direct-send guard must
            // report thinking without inventing a delegated task generation.
            var send = child.Engine().SendAsync("{}", deadline.Token);
            var request = await listener.GetContextAsync().WaitAsync(deadline.Token);
            JsonElement during = await Status();
            Check(during.GetProperty("activity").GetString() == "busy"
                && during.GetProperty("work_phase").GetString() == "thinking",
                "Detailed status missed an in-flight Engine request.");
            await Reply(request, """
                {"model":"measured-model","content":[],"usage":{"input_tokens":13,"output_tokens":5,"cache_read_input_tokens":7}}
                """);
            await send;
            JsonElement after = await Status();
            JsonElement counts = after.GetProperty("session_usage");
            Check(after.GetProperty("selected_model").GetString() == "selected-alias"
                && after.GetProperty("activity").GetString() == "idle"
                && counts.GetProperty("input_tokens").GetInt64() == 13
                && counts.GetProperty("output_tokens").GetInt64() == 5
                && counts.GetProperty("cache_read_tokens").GetInt64() == 7
                && after.GetProperty("measured_models")[0].GetProperty("model").GetString() == "measured-model",
                "Selected model or measured response counts were replaced with guesses.");
            Check(after.GetProperty("estimated_cost_usd").ValueKind == JsonValueKind.Null
                && after.GetProperty("cost_is_estimate").GetBoolean()
                && Math.Abs(after.GetProperty("cache_hit_rate").GetDouble() - 0.35) < 0.000001,
                "Unknown model was priced or cache-hit ratio was misreported.");

            var missing = child.Engine().SendAsync("{}", deadline.Token);
            await Reply(await listener.GetContextAsync().WaitAsync(deadline.Token),
                """{"model":"measured-model","content":[],"usage":{"output_tokens":2}}""");
            await missing;
            JsonElement partial = await Status();
            Check(partial.GetProperty("session_usage").GetProperty("input_tokens").ValueKind == JsonValueKind.Null
                && partial.GetProperty("session_usage").GetProperty("output_tokens").GetInt64() == 7
                && partial.GetProperty("cache_hit_rate").ValueKind == JsonValueKind.Null,
                "Missing usage was turned into zero or known counts were discarded.");
            Check(root.Engine().Meter().Latest is null && peer.Engine().Meter().Latest is null,
                "Status charged child usage to an observing or unrelated Agent.");
            await child.ShutdownAsync();
            JsonElement stopped = await Status();
            Check(stopped.GetProperty("activity").GetString() == "stopped"
                && stopped.GetProperty("session_usage").GetProperty("output_tokens").GetInt64() == 7,
                "Stopped retained child lost status or current-process usage.");
            child.Destroy();
            Check((await Call("worker")).IsError, "Detached child remained in status lookup.");
        }
        finally
        {
            app.Destroy();
        }
    }
}
