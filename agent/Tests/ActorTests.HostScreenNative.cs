using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using SkiaSharp;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    // Explicit native acceptance only; never register in the automated suite.
    internal static async Task<int> HostScreenNativeTurn(string displayText, string evidenceDirectory)
    {
        if (!int.TryParse(displayText, out int displayIndex) || displayIndex < 0)
            throw new ArgumentException("Native display acceptance requires an explicit nonnegative display index.");
        // The human must cover this entire display with the agreed synthetic
        // pattern before opting in. Never reuse the retired rectangle command.
        var selection = HostScreens.Select(HostScreens.ListDisplays(), displayIndex, null);
        Directory.CreateDirectory(evidenceDirectory);
        string skills = Path.Combine(evidenceDirectory, "skills");
        Directory.CreateDirectory(Path.Combine(skills, "root"));
        Directory.CreateDirectory(Path.Combine(skills, "host-screen"));
        File.WriteAllText(Path.Combine(skills, "root", "SKILL.md"),
            "---\nname: root\ntype: primary\ndescription: Native acceptance fixture.\ntools: load_skill wait_for_job\nloadable-skills: host-screen\n---\nOnly the explicitly approved synthetic display may be captured.\n");
        File.Copy(Path.Combine(Environment.CurrentDirectory, "agent/skills/host-screen/SKILL.md"),
            Path.Combine(skills, "host-screen/SKILL.md"));
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(60));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        var app = new Application(TextWriter.Null);
        Task? server = null;
        int requests = 0;
        bool observedRunning = false;
        byte[]? delivered = null;
        try
        {
            int port = ((IPEndPoint)listener.LocalEndpoint).Port;
            var configuration = new Configuration(Vendor.OpenAI, "gpt-6-astra", "http://127.0.0.1:" + port,
                "dummy-local-only", disableStreaming: true, requireToolExplanation: true, sandboxRoot: evidenceDirectory);
            var agent = new Agent(app, configuration, Path.Combine(evidenceDirectory, "jobs"),
                skillSpec: new SkillSpec(skills, "root"));
            var loaded = await agent.DispatchAsync(new ToolCallPart("setup", configuration.Target(), "load_skill",
                JsonSerializer.SerializeToElement(new { name = "host-screen" }))).WaitAsync(deadline.Token);
            Check(!loaded.IsError, "Native fixture could not load host-screen.");

            JsonObject Reply(string text, string? name = null, object? arguments = null)
            {
                var output = new JsonArray(new JsonObject
                {
                    ["type"] = "message", ["role"] = "assistant", ["status"] = "completed",
                    ["content"] = new JsonArray(new JsonObject { ["type"] = "output_text", ["text"] = text })
                });
                if (name is not null) output.Add(new JsonObject
                {
                    ["type"] = "function_call", ["id"] = "fc-native-" + requests,
                    ["call_id"] = "native-" + requests, ["name"] = name,
                    ["arguments"] = JsonSerializer.Serialize(arguments), ["status"] = "completed"
                });
                return new JsonObject
                {
                    ["status"] = "completed", ["model"] = "gpt-6-astra", ["output"] = output,
                    ["usage"] = new JsonObject { ["input_tokens"] = 10, ["output_tokens"] = 10 }
                };
            }
            server = Task.Run(async () =>
            {
                bool replay = false;
                while (requests < 10)
                {
                    using var client = await listener.AcceptTcpClientAsync(deadline.Token);
                    using var stream = client.GetStream();
                    string body = await ReadRequest(stream, deadline.Token);
                    requests++;
                    if (requests == 1)
                    {
                        await ReplyJson(stream, Reply("I will capture the whole approved synthetic display.", "screenshot",
                            new { display = displayIndex, ai_callback_delay = 0 }), deadline.Token);
                        continue;
                    }
                    var results = agent.History().View().Dialogue.SelectMany(entry => entry.Parts)
                        .OfType<ToolResultPart>().ToArray();
                    var imageResults = results.Where(result => result.Parts.OfType<BlobPart>().Any()).ToArray();
                    if (imageResults.Length == 0)
                    {
                        var latest = results.Last();
                        Check(!latest.IsError, "Native tool failed: " + JsonCodec.Text(latest.Parts));
                        Check(latest.Job?.Status == "running", "Expected a running native capture Job.");
                        observedRunning = true;
                        await ReplyJson(stream, Reply("I will collect the same capture Job, without starting another capture.",
                            "wait_for_job", new { handle = latest.Job!.Handle, ai_callback_delay = 1 }), deadline.Token);
                        continue;
                    }
                    Check(imageResults.Length == 1, "Native pixels duplicated in History.");
                    var blob = imageResults[0].Parts.OfType<BlobPart>().Single();
                    byte[] bytes = blob.Ref.Data();
                    using var wire = JsonDocument.Parse(body);
                    var images = wire.RootElement.GetProperty("input").EnumerateArray()
                        .Where(item => item.TryGetProperty("type", out var type) && type.GetString() == "function_call_output")
                        .Where(item => item.GetProperty("output").ValueKind == JsonValueKind.Array)
                        .SelectMany(item => item.GetProperty("output").EnumerateArray())
                        .Where(part => part.GetProperty("type").GetString() == "input_image").ToArray();
                    Check(images.Length == 1 && images[0].GetProperty("image_url").GetString() ==
                        "data:image/png;base64," + Convert.ToBase64String(bytes), "Actual native pixels missing from provider request.");
                    using var bitmap = SKBitmap.Decode(bytes);
                    Check(bitmap is not null && bitmap.Width > 0 && bitmap.Height > 0
                        && bitmap.Width <= 2048 && bitmap.Height <= 2048
                        && (long)bitmap.Width * selection.Bounds.Height == (long)bitmap.Height * selection.Bounds.Width,
                        "Native prepared display dimensions changed.");
                    var left = bitmap!.GetPixel(bitmap.Width / 4, bitmap.Height / 2);
                    var right = bitmap.GetPixel(3 * bitmap.Width / 4, bitmap.Height / 2);
                    Check(left.Red > 230 && left.Green < 25 && left.Blue < 25 && right.Blue > 230
                        && right.Red < 25 && right.Green < 25, "Native synthetic pixels are incorrect.");
                    if (replay)
                    {
                        Check(delivered!.SequenceEqual(bytes), "Native replay bytes changed.");
                        await ReplyJson(stream, Reply("Native screenshot replay verified."), deadline.Token);
                        return;
                    }
                    delivered = bytes;
                    replay = true;
                    await ReplyJson(stream, Reply("Native screenshot delivered."), deadline.Token);
                }
                throw new Exception("Native acceptance exceeded its bounded provider exchanges.");
            });
            try
            {
                string first = await agent.SendAsync("Capture the whole approved synthetic display.", deadline.Token).WaitAsync(deadline.Token);
                Check(first.Contains("Native screenshot delivered."), "Native model turn did not finish.");
                string second = await agent.SendAsync("Verify the retained screenshot; do not capture again.", deadline.Token).WaitAsync(deadline.Token);
                Check(second.Contains("replay verified"), "Native replay turn did not finish.");
            }
            catch when (server.IsFaulted) { await server; throw; }
            await server.WaitAsync(deadline.Token);
            Check(observedRunning, "Zero-delay capture did not exercise delayed Job delivery.");
            int captures = agent.History().View().Dialogue.SelectMany(entry => entry.Parts)
                .OfType<ToolCallPart>().Count(call => call.Name == "screenshot");
            Check(captures == 1, "Native model path captured more than once.");
            File.WriteAllBytes(Path.Combine(evidenceDirectory, "delivered.png"), delivered!);
            string receipt = JsonSerializer.Serialize(new { requests, captures, observedRunning,
                delivered_bytes = delivered!.Length, replay_verified = true });
            File.WriteAllText(Path.Combine(evidenceDirectory, "result.json"), receipt);
            Console.WriteLine(receipt);
            return 0;
        }
        finally
        {
            deadline.Cancel();
            listener.Stop();
            app.Destroy();
            if (server is not null) { try { await server; } catch { } }
        }
    }
}
