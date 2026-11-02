using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using EnsembleGui;
using SkiaSharp;

// Real Agent/History/MCP/GUI; only the model HTTP boundary is fake.
internal static class CaptureModelFixture
{
    internal static async Task RunAsync()
    {
        string repository = Environment.GetEnvironmentVariable("CAPTURE_REPOSITORY")
            ?? throw new InvalidOperationException("CAPTURE_REPOSITORY is required.");
        string home = Directory.GetCurrentDirectory();
        string skills = Path.Combine(home, "skills");
        Directory.CreateDirectory(Path.Combine(skills, "root"));
        Directory.CreateDirectory(Path.Combine(skills, "browser"));
        File.WriteAllText(Path.Combine(skills, "root", "SKILL.md"),
            "---\nname: root\ndescription: isolated capture model fixture\ntype: primary\ntools: load_skill tool_limits\nloadable-skills: browser\n---\nUse the connected browser only for this synthetic test.");
        File.Copy(Path.Combine(repository, "agent/skills/browser/SKILL.md"),
            Path.Combine(skills, "browser", "SKILL.md"));
        using var stop = new CancellationTokenSource(TimeSpan.FromMinutes(2));
        var root = new Ensemble.Ensemble(TextWriter.Null, dataDirectory: Path.Combine(home, "data"));
        var providerBuilder = WebApplication.CreateBuilder(); providerBuilder.Logging.ClearProviders();
        await using var provider = providerBuilder.Build();
        provider.Urls.Add("http://127.0.0.1:18203");
        var hostBuilder = WebApplication.CreateBuilder(); hostBuilder.Logging.ClearProviders();
        await using var host = hostBuilder.Build(); host.Urls.Add("http://127.0.0.1:18202");
        var config = new Configuration(Vendor.OpenAI, "gpt-6-astra", "http://127.0.0.1:18203",
            "synthetic-capture-fixture", surface: Surface.Responses, disableStreaming: true,
            requireToolExplanation: true, sandboxRoot: home);
        var agent = new Agent(root, config, Path.Combine(home, "jobs"), skillSpec: new SkillSpec(skills, "root"));
        int requests = 0;
        string? failure = null;
        byte[]? retained = null;
        object statusGate = new();
        int statusStage = 0;
        void Status(int stage, string? error = null)
        {
            lock (statusGate)
            {
                statusStage = Math.Max(statusStage, stage);
                File.WriteAllText(Path.Combine(home, "status.next"), JsonSerializer.Serialize(new
                    { stage = statusStage, requests, failure = error, imageBytes = retained?.Length }));
                File.Move(Path.Combine(home, "status.next"), Path.Combine(home, "status.json"), true);
            }
        }
        void Check(bool condition, string message)
        {
            if (!condition) throw new InvalidOperationException(message);
        }
        provider.MapPost("/v1/responses", async (HttpContext context) =>
        {
            int index = Interlocked.Increment(ref requests) - 1;
            try
            {
                Check(index < 4, "Unexpected extra model request.");
                Check(context.Request.Headers.Authorization == "Bearer synthetic-capture-fixture", "Wrong fake provider credential.");
                using var document = await JsonDocument.ParseAsync(context.Request.Body, cancellationToken: context.RequestAborted);
                var wire = document.RootElement;
                var names = wire.GetProperty("tools").EnumerateArray().Where(t => t.TryGetProperty("name", out _))
                    .Select(t => t.GetProperty("name").GetString()).ToHashSet();
                if (index == 0)
                {
                    Check(names.Contains("load_skill") && names.Contains("tool_limits"), "Missing fixture setup authority.");
                    Check(!names.Contains("gui_capture"), "Capture appeared before browser skill activation.");
                }
                else Check(names.Contains("gui_capture"), "Real browser skill/catalogue did not expose gui_capture.");
                if (index >= 2)
                {
                    var result = wire.GetProperty("input").EnumerateArray().Single(p =>
                        p.GetProperty("type").GetString() == "function_call_output"
                        && p.GetProperty("call_id").GetString() == "capture-call");
                    Check(result.GetProperty("output").ValueKind == JsonValueKind.Array, "Capture result contains no typed media.");
                    var image = result.GetProperty("output").EnumerateArray().Single(p => p.GetProperty("type").GetString() == "input_image");
                    string locator = image.GetProperty("image_url").GetString()!;
                    const string prefix = "data:image/png;base64,";
                    Check(locator.StartsWith(prefix, StringComparison.Ordinal), "Capture image is not inline PNG.");
                    byte[] pixels = Convert.FromBase64String(locator[prefix.Length..]);
                    using var bitmap = SKBitmap.Decode(pixels);
                    Check(bitmap is not null && bitmap.Width == 960 && bitmap.Height == 640, "Wrong native capture dimensions.");
                    var left = bitmap!.GetPixel(16, 88); var right = bitmap.GetPixel(944, 88);
                    Check(left.Red >= 250 && left.Blue <= 4 && right.Blue >= 250 && right.Red <= 4,
                        "Native synthetic tab markers missing from model pixels.");
                    if (retained is null) retained = pixels;
                    else Check(retained.SequenceEqual(pixels), "Replay changed the capture bytes.");
                    var parts = agent.History().View().Dialogue.SelectMany(e => e.Parts).ToArray();
                    Check(parts.OfType<ToolCallPart>().Count(p => p.CallID == "capture-call") == 1, "Capture call missing or duplicated in History.");
                    var stored = parts.OfType<ToolResultPart>().Single(p => p.CallID == "capture-call");
                    Check(!stored.IsError && stored.Parts.OfType<BlobPart>().Single().Ref.Data().SequenceEqual(pixels),
                        "Canonical History does not retain the actual capture derivative.");
                }
                var output = new JsonArray(new JsonObject
                {
                    ["type"] = "message", ["role"] = "assistant", ["status"] = "completed",
                    ["content"] = new JsonArray(new JsonObject { ["type"] = "output_text", ["text"] = index switch
                    {
                        0 => "I will load the browser skill so I can request this page's pixels.",
                        1 => "I will allow time for explicit human approval, then request the screenshot.",
                        2 => "Capture verified.",
                        _ => "Capture replay verified."
                    } })
                });
                void Call(string id, string name, object arguments) => output.Add(new JsonObject
                {
                    ["type"] = "function_call", ["id"] = "fc-" + id, ["call_id"] = id,
                    ["name"] = name, ["arguments"] = JsonSerializer.Serialize(arguments), ["status"] = "completed"
                });
                if (index == 0) Call("load-browser", "load_skill", new { name = "browser" });
                if (index == 1)
                {
                    Call("capture-limits", "tool_limits", new { ai_callback_delay = 30 });
                    Call("capture-call", "gui_capture", new { });
                }
                await context.Response.WriteAsJsonAsync(new JsonObject
                {
                    ["status"] = "completed", ["model"] = "gpt-6-astra", ["output"] = output,
                    ["usage"] = new JsonObject { ["input_tokens"] = 10, ["output_tokens"] = 10 }
                }, context.RequestAborted);
                Status(index + 1);
            }
            catch (Exception error)
            {
                failure = error.Message; Status(index + 1, failure);
                context.Response.StatusCode = 500;
            }
        });
        Task? serving = null;
        try
        {
            Status(0);
            await provider.StartAsync(stop.Token);
            // Use the normal target: retained screenshots must survive replay
            // without a fixture-only budget increase or an implicit size cut.
            serving = Gui.ServeAsync(host, agent, new GuiOptions(18202, SettingsPath: Path.Combine(home, "appearance.json"),
                TracePath: Path.Combine(home, "gui.log"), BasePath: "/assistant"), stop.Token);
            await host.StartAsync(stop.Token);
            Console.WriteLine("READY capture model fixture http://127.0.0.1:18202/assistant/");
            while (!File.Exists(Path.Combine(home, "stop"))) await Task.Delay(100, stop.Token);
            Check(failure is null, failure ?? "Provider fixture failed.");
            Check(requests == 4 && retained is not null, "Model capture/replay sequence did not finish.");
            Console.WriteLine("PASS real skill discovery, Agent-requested capture, model pixels, canonical History and replay; four loopback requests");
        }
        finally
        {
            stop.Cancel();
            try
            {
                await host.StopAsync();
                if (serving is not null) await serving;
            }
            finally
            {
                await provider.StopAsync();
                root.Destroy();
            }
            Console.WriteLine("STOPPED capture model fixture, provider and Agents");
        }
    }
}
