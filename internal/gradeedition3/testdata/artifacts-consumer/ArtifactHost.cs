using System.Security.Cryptography;
using System.Text;
using System.Text.Json;
using Ensemble;
using EnsembleGui;

// All fixture state and oracles belong to the grader, outside candidate assemblies.
internal static class ArtifactHost
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
        string skills = Path.Combine(workspace, "skills");
        Directory.CreateDirectory(Path.Combine(skills, "root"));
        Directory.CreateDirectory(Path.Combine(skills, "products"));
        File.WriteAllText(Path.Combine(skills, "root", "SKILL.md"),
            "---\nname: root\ntype: primary\ndescription: Independent artifact grader.\ntools: write_file load_skill wait_for_job\nloadable-skills: products\n---\nUse explicitly granted tools.\n");
        File.WriteAllText(Path.Combine(skills, "products", "SKILL.md"),
            "---\nname: products\ntype: loadable\ndescription: Work products.\ntools: offer_work_product\n---\nOffer the requested work product.\n");
        string nonce = Convert.ToHexString(RandomNumberGenerator.GetBytes(16)).ToLowerInvariant();
        string filename = "product-" + nonce + ".bin";
        string file = Path.Combine(workspace, filename);
        string content = "grader:" + nonce + "\nUnicode: é 界 😀\t\0\n";
        byte[] expected = Encoding.UTF8.GetBytes(content);
        string final = "artifact-complete-" + nonce;
        string local = $"http://127.0.0.1:{port}";
        var config = new Configuration(Vendor.Anthropic, "claude-3-haiku-20240307", local,
            "artifact-grader-only", disableStreaming: true, sandboxRoot: workspace);
        var root = new Ensemble.Ensemble(TextWriter.Null);
        string? viewerName = Environment.GetEnvironmentVariable("ARTIFACT_UNNAMED_VIEWER") == "1" ? null : "viewer";
        var viewer = new Agent(root, config, name: viewerName);
        var owner = new Agent(root, config, name: "candidate", skillSpec: new SkillSpec(skills, "root"));
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(120));
        using var stop = new CancellationTokenSource();
        var grant = await owner.DispatchAsync(new ToolCallPart("grader-load", config.Target(), "load_skill",
            JsonSerializer.SerializeToElement(new { name = "products" })), deadline.Token);
        Require(!grant.IsError, "Candidate failed explicit offer grant");
        var builder = WebApplication.CreateBuilder();
        builder.Logging.ClearProviders();
        builder.WebHost.UseUrls(local);
        await using var host = builder.Build();
        int requests = 0;
        string? failure = null;
        HistorySnapshot? closedHistory = null;
        HistorySnapshot Snapshot() => closedHistory ?? owner.History().Snapshot();
        WorkProductInfo? Offer() => Snapshot().Events.LastOrDefault(fact => fact.WorkProduct is not null)?.WorkProduct;
        object Status() => new
        {
            nonce, filename, final, requests, failure,
            hash = Convert.ToHexString(SHA256.HashData(expected)).ToLowerInvariant(), bytes = expected.Length,
            offer = Offer() is { } offer ? new { id = offer.ID, name = offer.Name, offered_bytes = offer.OfferedBytes } : null,
            returned = Snapshot().Events.Count(fact => fact.Kind == EventKind.ToolReturned),
            offers = Snapshot().Events.Count(fact => fact.WorkProduct is not null)
        };
        host.MapGet("/grade/status", () => Results.Json(Status()));
        host.MapPost("/grade/change-file", () =>
        {
            Require(Offer() is not null, "File changed before successful offer");
            expected = RandomNumberGenerator.GetBytes(32771);
            File.WriteAllBytes(file, expected);
            return Results.Json(Status());
        });
        host.MapPost("/grade/replace-owner", () =>
        {
            closedHistory = owner.History().Snapshot();
            owner.Destroy();
            owner = new Agent(root, config, name: "candidate");
            return Results.Ok();
        });
        host.MapPost("/v1/messages", async (HttpContext context) =>
        {
            try
            {
                using var body = await JsonDocument.ParseAsync(context.Request.Body, cancellationToken: context.RequestAborted);
                Require(context.Request.Headers["x-api-key"] == "artifact-grader-only", "Unexpected provider credentials");
                Require(!body.RootElement.TryGetProperty("stream", out var streaming) || !streaming.GetBoolean(), "Expected buffered fixture request");
                int turn = Interlocked.Increment(ref requests);
                object[] parts;
                string reason;
                if (turn == 1)
                {
                    parts = [new { type = "tool_use", id = "write-" + nonce, name = "write_file", input = new { path = file, content } }];
                    reason = "tool_use";
                }
                else if (turn == 2)
                {
                    Require(File.Exists(file) && File.ReadAllBytes(file).SequenceEqual(expected), "Real write_file effect missing or wrong");
                    Require(Snapshot().Events.Any(fact => fact.Kind == EventKind.ToolReturned && fact.CallID == "write-" + nonce && !fact.IsError), "Write result was not recorded before next request");
                    parts = [new { type = "tool_use", id = "offer-" + nonce, name = "offer_work_product", input = new { path = file } }];
                    reason = "tool_use";
                }
                else
                {
                    Require(turn == 3, "Unexpected extra provider request");
                    var offer = Offer();
                    Require(offer is not null, "Offer metadata missing before final provider request");
                    Require(!body.RootElement.GetRawText().Contains(offer!.ID), "Opaque offer ID leaked into provider request");
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
        var serving = Gui.ServeAsync(host, viewer, new GuiOptions(port, SettingsPath: Path.Combine(workspace, "settings.json"),
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
            Console.WriteLine("STOPPED artifact consumer");
        }
    }
}
