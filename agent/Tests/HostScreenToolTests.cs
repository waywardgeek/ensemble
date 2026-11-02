using System.Text.Json;
using Ensemble;
using SkiaSharp;
using Application = Ensemble.Ensemble;

internal static partial class HostScreenTests
{
    public static void Arguments()
    {
        Check(Tool.ParseScreenRequest(JsonSerializer.SerializeToElement(new { })) == 0, "Omitted display did not select primary.");
        Check(Tool.ParseScreenRequest(JsonSerializer.SerializeToElement(new { display = 0 })) == 0, "Explicit primary changed.");
        Check(Tool.ParseScreenRequest(JsonSerializer.SerializeToElement(new { display = 1 })) == 1, "Secondary display changed.");
        Check(Tool.ParseScreenRequest(JsonSerializer.SerializeToElement(new { ai_callback_delay = 0, max_output_bytes = 1000 })) == 0,
            "Observation controls changed default display.");
        foreach (string invalid in new[] {
            "null", "[]", "{\"display\":null}", "{\"display\":-1}", "{\"display\":0.5}", "{\"display\":2147483648}",
            "{\"x\":0}", "{\"x\":0,\"y\":0,\"width\":0,\"height\":1}",
            "{\"x\":0,\"y\":0,\"width\":100000,\"height\":100000}",
            "{\"x\":2147483647,\"y\":0,\"width\":2,\"height\":1}",
            "{\"display\":0,\"x\":0,\"y\":0,\"width\":1,\"height\":1}",
            "{\"display\":0,\"display\":1}", "{\"width\":null}", "{\"window\":1}", "{\"display\":\"0\"}" })
        {
            using var doc = JsonDocument.Parse(invalid);
            Refuses(() => Tool.ParseScreenRequest(doc.RootElement));
        }
    }

    public static async Task AuthorityAndPixels(string temp)
    {
        string workspace = Path.Combine(temp, "host-screen-authority");
        string skills = Path.Combine(workspace, "skills");
        Directory.CreateDirectory(Path.Combine(skills, "root"));
        Directory.CreateDirectory(Path.Combine(skills, "host-screen"));
        File.WriteAllText(Path.Combine(skills, "root", "SKILL.md"),
            "---\nname: root\ntype: primary\ndescription: Isolated screen tests.\ntools: load_skill unload_skill wait_for_job\nloadable-skills: host-screen\n---\nTest authority only.\n");
        File.Copy(Path.Combine(Environment.CurrentDirectory, "agent", "skills", "host-screen", "SKILL.md"),
            Path.Combine(skills, "host-screen", "SKILL.md"), true);
        var app = new Application(TextWriter.Null);
        try
        {
            var config = new Configuration(Vendor.OpenAI, "gpt-6-astra", "http://127.0.0.1:1", sandboxRoot: workspace);
            var legacy = new Agent(app, config, Path.Combine(workspace, "legacy-jobs"));
            legacy.InstallTools();
            Check(!legacy.ToolEnabled("screenshot") && !legacy.ToolEnabled("list_displays"), "Legacy registry implicitly granted screens.");
            string legacyWire = legacy.Engine().Render(new Context([])).ToJsonString();
            Check(!legacyWire.Contains("\"name\":\"screenshot\"") && !legacyWire.Contains("\"name\":\"list_displays\""),
                "Legacy registry advertised host-screen declarations.");
            int sequence = 0;
            async Task<ToolResultPart> Call(Agent owner, string name, object args)
            {
                var result = await owner.DispatchAsync(new ToolCallPart("screen-" + ++sequence,
                    owner.Configuration().Target(), name, JsonSerializer.SerializeToElement(args)));
                while (result.Job?.Status == "running")
                {
                    result = await owner.DispatchAsync(new ToolCallPart("screen-" + ++sequence,
                        owner.Configuration().Target(), "wait_for_job",
                        JsonSerializer.SerializeToElement(new { handle = result.Job.Handle, ai_callback_delay = 1 })));
                }
                return result;
            }
            foreach (string name in new[] { "list_displays", "screenshot" })
                Check((await Call(legacy, name, new { })).IsError, "Unskilled screen operation executed.");

            var agent = new Agent(app, config, Path.Combine(workspace, "jobs"), skillSpec: new SkillSpec(skills, "root"));
            Check((await Call(agent, "screenshot", new { })).IsError, "Unloaded skill executed capture.");
            Check(!(await Call(agent, "load_skill", new { name = "host-screen" })).IsError, "Screen skill failed to load.");
            Check(agent.ToolEnabled("list_displays") && agent.ToolEnabled("screenshot"), "Explicit grant absent.");
            using (var rendered = JsonDocument.Parse(agent.Engine().Render(agent.History().View()).ToJsonString()))
            {
                var declaration = rendered.RootElement.GetProperty("tools").EnumerateArray()
                    .Single(tool => tool.TryGetProperty("name", out var name) && name.GetString() == "screenshot");
                var schema = declaration.GetProperty("parameters");
                var properties = schema.GetProperty("properties");
                Check(properties.TryGetProperty("display", out var display), "Display selector missing from model schema.");
                Check(display.GetProperty("type").GetString() == "integer"
                    && display.GetProperty("default").GetInt32() == 0, "Integer display default is not advertised.");
                foreach (string removed in new[] { "x", "y", "width", "height" })
                    Check(!properties.TryGetProperty(removed, out _), "Removed rectangle field remains in model schema.");
                if (schema.TryGetProperty("required", out var required))
                    Check(!required.EnumerateArray().Any(field => field.GetString() == "display"), "Display is not optional.");
            }
            foreach (object invalid in new object[] { new { x = 0 }, new { display = -1 }, new { x = 0, y = 0, width = 0, height = 1 } })
                Check((await Call(agent, "screenshot", invalid)).IsError, "Malformed selection reached native work.");

            var screenshot = agent.Tools().Single(tool => tool.Name() == "screenshot");
            Check(screenshot.HasStructuredResult(), "Screenshot lost structured Job results.");
            using var bitmap = new SKBitmap(4, 2);
            bitmap.Erase(SKColors.Red);
            using var image = SKImage.FromBitmap(bitmap);
            using var encoded = image.Encode(SKEncodedImageFormat.Png, 100);
            byte[] png = encoded.ToArray();
            var call = new ToolCallPart("synthetic-screen", config.Target(), "screenshot", JsonSerializer.SerializeToElement(new { }));
            var selection = new ScreenSelection(new ScreenRectangle(0, 0, 4, 2), 1);
            var pixels = screenshot.PrepareScreenshotResult(call, selection, png, CancellationToken.None);
            Check(!pixels.IsError && pixels.Parts.OfType<BlobPart>().Count() == 1, "Synthetic screenshot did not return pixels.");
            var context = new Context([new Entry(1, Actor.Agent, [call]), new Entry(2, Actor.Tool, [pixels])]);
            string wire = agent.Engine().Render(context).ToJsonString();
            using var wireJson = JsonDocument.Parse(wire);
            string delivered = wireJson.RootElement.GetRawText();
            Check(delivered.Contains("input_image") && delivered.Contains("data:image/png;base64,"), "Screenshot pixels absent from model wire.");
            try
            {
                screenshot.PrepareScreenshotResult(call, new ScreenSelection(new(0, 0, 3, 2), 1), png, CancellationToken.None);
                throw new Exception("Mislabeled capture dimensions accepted.");
            }
            catch (InvalidDataException) { }
            using var cancelled = new CancellationTokenSource();
            cancelled.Cancel();
            try { screenshot.PrepareScreenshotResult(call, selection, png, cancelled.Token); throw new Exception("Cancelled result published pixels."); }
            catch (OperationCanceledException) { }
            Check(!(await Call(agent, "unload_skill", new { name = "host-screen" })).IsError, "Screen skill unload failed.");
            Check(!agent.ToolEnabled("screenshot") && (await Call(agent, "screenshot", new { })).IsError, "Unload retained capture authority.");
            try { screenshot.PrepareScreenshotResult(call, selection, png, CancellationToken.None); throw new Exception("Revoked result published pixels."); }
            catch (UnauthorizedAccessException) { }

            var unsupported = new Agent(app, new Configuration(Vendor.OpenAI, "unknown-image-model", "http://127.0.0.1:1", sandboxRoot: workspace), Path.Combine(workspace, "unsupported-jobs"),
                skillSpec: new SkillSpec(skills, "root"));
            Check(!(await Call(unsupported, "load_skill", new { name = "host-screen" })).IsError, "Unsupported fixture grant failed.");
            var refusal = await Call(unsupported, "screenshot", new { ai_callback_delay = 0 });
            string reason = JsonCodec.Text(refusal.Parts);
            Check(refusal.IsError && (reason.Contains("does not support") || reason.Contains("cannot render")),
                "Model capability did not refuse before native work: " + reason);

            var safe = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-6-astra", "http://127.0.0.1:1", sandboxRoot: workspace, safeMode: true), Path.Combine(workspace, "safe-jobs"), skillSpec: new SkillSpec(skills, "root"));
            Check((await Call(safe, "load_skill", new { name = "host-screen" })).IsError, "Safe mode loaded host-screen authority.");
            foreach (string name in new[] { "list_displays", "screenshot" })
                Check((await Call(safe, name, new { })).IsError, "Safe mode executed a screen operation.");
        }
        finally { app.Destroy(); }
    }
}
