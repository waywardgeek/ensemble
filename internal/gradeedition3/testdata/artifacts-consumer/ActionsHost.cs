using System.Security.Cryptography;
using System.Text.Json;
using Ensemble;
using EnsembleGui;

// Independent Actions fixture: the only controlled boundary is a public custom
// tool callback. It proves ordinary Job observations, not delayed offer metadata.
internal static class ActionsHost
{
    private static void Require(bool condition, string message)
    {
        if (!condition) throw new InvalidOperationException(message);
    }

    internal static async Task RunAsync()
    {
        int port = int.Parse(Environment.GetEnvironmentVariable("ARTIFACT_GRADE_PORT") ?? "0");
        Require(port > 0 && port is not (8088 or 8089 or 8090), "Explicit isolated port required");
        string work = Directory.GetCurrentDirectory();
        string workspace = Path.Combine(work, "workspace");
        Directory.CreateDirectory(workspace);
        string nonce = Convert.ToHexString(RandomNumberGenerator.GetBytes(16)).ToLowerInvariant();
        string final = "actions-complete-" + nonce;
        string output = Path.Combine(workspace, "report-" + nonce + ".txt");
        File.WriteAllText(output, "A pathname in tool output does not grant download authority.\n");
        string local = $"http://127.0.0.1:{port}";
        var config = new Configuration(Vendor.Anthropic, "claude-3-haiku-20240307", local,
            "artifact-grader-only", disableStreaming: true, sandboxRoot: workspace);
        var root = new Ensemble.Ensemble(TextWriter.Null);
        var owner = new Agent(root, config, name: "candidate");
        // Load retained facts through the public persistence boundary before mounting.
        var legacy = new Agent(root, config, name: "legacy-timing");
        var legacyTime = DateTimeOffset.Parse("2004-03-02T01:02:03.4567890+05:30",
            System.Globalization.CultureInfo.InvariantCulture);
        string legacyPath = Path.Combine(workspace, "legacy-timing.json");
        var legacyConfig = JsonSerializer.SerializeToElement(new
        {
            vendor = "anthropic", model = "offline", system_prompt = "", tools = Array.Empty<object>()
        });
        File.WriteAllText(legacyPath, JsonCodec.WriteSave(new SaveDocument(new Context([
            new Entry(1, Actor.Tool, [new ToolResultPart("unknown", [new TextPart("legacy-unknown-" + nonce)])]),
            new Entry(2, Actor.Tool, [new ToolResultPart("zero", [new TextPart("legacy-zero-" + nonce)])],
                time: legacyTime, observationDuration: TimeSpan.Zero)
        ]), 2, legacyConfig, [])));
        Require(legacy.History().LoadSave(legacyPath), "legacy browser snapshot actually loaded");
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(120));
        using var stop = new CancellationTokenSource();
        var entered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        int executions = 0;
        _ = new Tool(owner, "held_action", "Return a report pathname after the external operation completes.",
            JsonSerializer.SerializeToElement(new { type = "object", properties = new { }, additionalProperties = false }),
            async (_, _, cancellation) =>
            {
                Interlocked.Increment(ref executions);
                entered.TrySetResult();
                await release.Task.WaitAsync(cancellation);
                return output;
            });
        _ = new Tool(owner, ToolKind.WaitForJob);
        // This public observation setting affects the first real model tool call.
        owner.Jobs().SetLimits(new JobLimits(Delay: 0));
        var builder = WebApplication.CreateBuilder();
        builder.Logging.ClearProviders();
        builder.WebHost.UseUrls(local);
        await using var host = builder.Build();
        int requests = 0;
        string? failure = null;
        object Status() => new
        {
            nonce, final, output, requests, failure, executions,
            observations = owner.History().Snapshot().Events
                .Where(fact => fact.Kind == EventKind.ToolReturned)
                .Select(fact => new { seq = fact.Seq, state = fact.Job?.Status,
                    recorded = fact.Time?.ToString("O"), milliseconds = fact.ObservationDuration?.TotalMilliseconds }).ToArray(),
            entered = entered.Task.IsCompleted, released = release.Task.IsCompleted,
            returned = owner.History().Snapshot().Events.Count(fact => fact.Kind == EventKind.ToolReturned),
            running = owner.History().Snapshot().Events.Count(fact => fact.Kind == EventKind.ToolReturned && fact.Job?.Status == "running"),
            done = owner.History().Snapshot().Events.Count(fact => fact.Kind == EventKind.ToolReturned && fact.Job?.Status == "done")
        };
        host.MapGet("/grade/status", () => Results.Json(Status()));
        host.MapPost("/grade/release-action", () =>
        {
            Require(entered.Task.IsCompleted, "Release before callback entered");
            release.TrySetResult();
            return Results.Ok();
        });
        host.MapPost("/v1/messages", async (HttpContext context) =>
        {
            try
            {
                using var body = await JsonDocument.ParseAsync(context.Request.Body, cancellationToken: context.RequestAborted);
                Require(context.Request.Headers["x-api-key"] == "artifact-grader-only", "Unexpected provider credentials");
                Require(!body.RootElement.TryGetProperty("stream", out var streaming) || !streaming.GetBoolean(), "Expected buffered request");
                int turn = Interlocked.Increment(ref requests);
                object[] parts;
                string reason;
                if (turn == 1)
                {
                    parts = [new { type = "tool_use", id = "held-" + nonce, name = "held_action", input = new { } }];
                    reason = "tool_use";
                }
                else if (turn == 2)
                {
                    var observed = owner.History().Snapshot().Events.Single(fact => fact.Kind == EventKind.ToolReturned && fact.CallID == "held-" + nonce);
                    Require(observed.Job is { Status: "running" } && observed.WorkProduct is null, "real-running-observation-recorded");
                    await entered.Task.WaitAsync(deadline.Token);
                    await release.Task.WaitAsync(deadline.Token);
                    // Observation may wake on output before terminal publication.
                    // Wait on the public snapshot without consuming result bytes.
                    while (owner.Jobs().Find(observed.Job!.Handle).Snapshot().Status != "done")
                        await Task.Delay(10, deadline.Token);
                    parts = [new { type = "tool_use", id = "wait-" + nonce, name = "wait_for_job", input = new { handle = observed.Job!.Handle, ai_callback_delay = 10 } }];
                    reason = "tool_use";
                }
                else
                {
                    Require(turn == 3, "Unexpected extra provider request");
                    var observed = owner.History().Snapshot().Events.Single(fact => fact.Kind == EventKind.ToolReturned && fact.CallID == "wait-" + nonce);
                    Require(observed.Job is { Status: "done" } && !observed.IsError && observed.WorkProduct is null,
                        $"real-terminal-observation-without-offer: status={observed.Job?.Status ?? "missing"}, error={observed.IsError}");
                    Require(executions == 1, "Tool must execute exactly once");
                    Require(body.RootElement.GetRawText().Contains(Path.GetFileName(output)), "Report pathname missing from actual result request");
                    parts = [new { type = "text", text = final }];
                    reason = "end_turn";
                }
                await context.Response.WriteAsJsonAsync(new { id = "response-" + turn, type = "message", role = "assistant",
                    model = config.Model, content = parts, stop_reason = reason, usage = new { input_tokens = 1, output_tokens = 1 } });
            }
            catch (Exception error)
            {
                failure = error.Message;
                context.Response.StatusCode = 500;
                await context.Response.WriteAsJsonAsync(new { error = "Grader provider oracle failed" });
            }
        });
        var serving = Gui.ServeAsync(host, owner, new GuiOptions(port, SettingsPath: Path.Combine(workspace, "settings.json"),
            BasePath: "/assistant"), stop.Token);
        try
        {
            await host.StartAsync(deadline.Token);
            if (serving.IsCompleted) await serving;
            Console.WriteLine("READY " + local + "/assistant/");
            while (!File.Exists(Path.Combine(work, "stop")))
            {
                if (serving.IsCompleted) { await serving; throw new InvalidOperationException("GUI ended before stop"); }
                await Task.Delay(25, deadline.Token);
            }
        }
        finally
        {
            release.TrySetResult();
            stop.Cancel();
            try { await serving; }
            finally
            {
                try { await host.StopAsync(CancellationToken.None); }
                finally
                {
                    try { File.WriteAllText(Path.Combine(work, "result.json"), JsonSerializer.Serialize(Status())); }
                    finally { root.Destroy(); }
                }
            }
            Console.WriteLine("STOPPED Actions consumer");
        }
    }
}
