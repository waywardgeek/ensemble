using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using SkiaSharp;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    public static async Task ImageToolTurns(string temp)
    {
        var directory = new DirectoryInfo(Directory.GetCurrentDirectory());
        while (directory is not null && !File.Exists(Path.Combine(directory.FullName,
            "agent/skills/code-tools/SKILL.md"))) directory = directory.Parent;
        Check(directory is not null, "Run image-tool tests from the repository.");
        string root = Path.Combine(temp, "image-model-turn-" + Guid.NewGuid().ToString("N"));
        string skills = Path.Combine(root, "skills");
        Directory.CreateDirectory(Path.Combine(skills, "root"));
        Directory.CreateDirectory(Path.Combine(skills, "code-tools"));
        File.WriteAllText(Path.Combine(skills, "root", "SKILL.md"),
            "---\nname: root\ndescription: fixture\ntype: primary\ntools: load_skill read_file\nloadable-skills: code-tools\n---\nFixture");
        File.Copy(Path.Combine(directory!.FullName, "agent/skills/code-tools/SKILL.md"),
            Path.Combine(skills, "code-tools", "SKILL.md"));
        foreach (var sample in new[]
        {
            (Vendor.OpenAI, "gpt-6-astra", Surface.Responses),
            (Vendor.Anthropic, "claude-3-haiku-20240307", Surface.Messages),
            (Vendor.Gemini, "gemini-3-flash-preview", Surface.GenerateContent)
        })
        {
            string home = Path.Combine(root, sample.Item3.ToString());
            Directory.CreateDirectory(home);
            using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(45));
            var listener = new TcpListener(IPAddress.Loopback, 0);
            listener.Start();
            var app = new Application(TextWriter.Null);
            Task? server = null;
            try
            {
                int port = ((IPEndPoint)listener.LocalEndpoint).Port;
                var config = new Configuration(sample.Item1, sample.Item2,
                    "http://127.0.0.1:" + port, "dummy-local-only", surface: sample.Item3,
                    disableStreaming: true, requireToolExplanation: true, sandboxRoot: home);
                var agent = new Agent(app, config, Path.Combine(home, "jobs"),
                    skillSpec: new SkillSpec(skills, "root"));
                var loaded = await agent.DispatchAsync(new ToolCallPart("setup-skill", config.Target(),
                    "load_skill", JsonSerializer.SerializeToElement(new { name = "code-tools" })))
                    .WaitAsync(deadline.Token);
                Check(!loaded.IsError, "Fixture could not activate code-tools.");
                using var sourceBitmap = new SKBitmap(4096, 32);
                sourceBitmap.Erase(SKColors.Magenta);
                using var sourceImage = SKImage.FromBitmap(sourceBitmap);
                using var sourceData = sourceImage.Encode(SKEncodedImageFormat.Png, 100);
                byte[] sourceBytes = sourceData.ToArray();
                string sourcePath = Path.Combine(home, "source.png");
                File.WriteAllBytes(sourcePath, sourceBytes);
                string nonce = "follow-up-" + Guid.NewGuid().ToString("N");
                File.WriteAllText(Path.Combine(home, "caption.txt"), nonce);
                // Expected derivative is prepared before the turn; inspect_image
                // itself is requested only by the fake model below.
                byte[] expected = agent.PrepareImage(sourceBytes).Image.Ref.Data();
                string encoded = Convert.ToBase64String(expected);
                server = Task.Run(async () =>
                {
                    for (int request = 0; request < 3; request++)
                    {
                        using var client = await listener.AcceptTcpClientAsync(deadline.Token);
                        using var stream = client.GetStream();
                        string body = await ReadRequest(stream, deadline.Token);
                        using var json = JsonDocument.Parse(body);
                        if (request == 0)
                        {
                            var tools = json.RootElement.GetProperty("tools").EnumerateArray().ToArray();
                            var declarations = sample.Item3 == Surface.GenerateContent
                                ? tools.SelectMany(tool => tool.GetProperty("functionDeclarations").EnumerateArray())
                                : tools.Where(tool => tool.TryGetProperty("name", out _));
                            var names = declarations.Select(tool => tool.GetProperty("name").GetString()).ToHashSet();
                            Check(names.Contains("inspect_image") && names.Contains("read_file"),
                                "Fixture must advertise both inspection and follow-up read authority.");
                        }
                        if (request > 0)
                        {
                            AssertModelImage(json.RootElement, sample.Item3, encoded);
                            if (request == 1)
                            {
                                Check(File.ReadAllBytes(sourcePath).SequenceEqual(sourceBytes),
                                    "Model-issued inspection rewrote its source.");
                                File.Delete(sourcePath);
                            }
                            else
                            {
                                Check(!File.Exists(sourcePath), "Source unexpectedly reappeared.");
                                Check(body.Contains(nonce, StringComparison.Ordinal),
                                    "Follow-up tool did not execute before the third request.");
                            }
                        }
                        await ReplyJson(stream, ImageTurnReply(sample.Item3, sample.Item2, request), deadline.Token);
                    }
                });
                string answer;
                try
                {
                    answer = await agent.SendAsync("Inspect source.png, then read caption.txt.", deadline.Token)
                        .WaitAsync(deadline.Token);
                }
                catch when (server.IsFaulted)
                {
                    // Surface the fake provider's precise wire assertion rather
                    // than the secondary HTTP failure caused by its closed socket.
                    await server;
                    throw;
                }
                await server.WaitAsync(deadline.Token);
                Check(answer.Contains("Image inspected", StringComparison.Ordinal), "Model turn did not finish.");
                var parts = agent.History().View().Dialogue.SelectMany(entry => entry.Parts).ToArray();
                Check(parts.OfType<ToolCallPart>().Count(call => call.CallID == "image-call") == 1,
                    "Inspection call missing or duplicated in canonical History.");
                var result = parts.OfType<ToolResultPart>().Single(value => value.CallID == "image-call");
                Check(!result.IsError, "Model-issued image inspection failed.");
                var retained = result.Parts.OfType<BlobPart>().Single();
                Check(retained.Ref.Kind == RefKind.Uri && retained.Ref.Data().SequenceEqual(expected),
                    "Canonical History did not retain the actual prepared derivative.");
                using var pixels = SKBitmap.Decode(retained.Ref.Data());
                Check(pixels is not null && pixels.Width == 2048 && pixels.Height == 16
                    && pixels.GetPixel(12, 8) == SKColors.Magenta, "Retained image pixels are incorrect.");
                Check(JsonCodec.Text(result.Parts).Contains("4096 × 32 to 2048 × 16", StringComparison.Ordinal),
                    "Model-turn result lost resize reporting.");
            }
            finally
            {
                deadline.Cancel();
                listener.Stop();
                app.Destroy();
                if (server is not null)
                {
                    // Normal-path assertions were awaited above. Always join
                    // the fake, preserving any earlier product/test exception.
                    try { await server; } catch { }
                }
            }
        }
    }

    private static void AssertModelImage(JsonElement wire, Surface surface, string encoded)
    {
        if (surface == Surface.GenerateContent)
        {
            var parts = wire.GetProperty("contents").EnumerateArray()
                .SelectMany(entry => entry.GetProperty("parts").EnumerateArray()).ToArray();
            var call = parts.Single(part => part.TryGetProperty("functionCall", out var value)
                && value.GetProperty("id").GetString() == "image-call");
            Check(call.GetProperty("thoughtSignature").GetString() == "fixture-signature-image-call",
                "Gemini replay lost the call's original thought signature.");
            var result = parts.Single(part => part.TryGetProperty("functionResponse", out var value)
                && value.GetProperty("id").GetString() == "image-call").GetProperty("functionResponse");
            Check(result.GetProperty("name").GetString() == "inspect_image", "Gemini lost function-name pairing.");
            var image = result.GetProperty("parts").EnumerateArray().Single().GetProperty("inlineData");
            Check(image.GetProperty("mimeType").GetString() == "image/png"
                && image.GetProperty("data").GetString() == encoded, "Gemini lost actual derivative bytes.");
            var content = result.GetProperty("response").GetProperty("content").EnumerateArray().ToArray();
            Check(content.Any(part => part.TryGetProperty("text", out _)), "Gemini lost accompanying text.");
            Check(content.Single(part => part.TryGetProperty("$ref", out _)).GetProperty("$ref").GetString()
                == image.GetProperty("displayName").GetString(), "Gemini image reference is not paired.");
        }
        else if (surface == Surface.Responses)
        {
            var result = wire.GetProperty("input").EnumerateArray().Single(part =>
                part.GetProperty("type").GetString() == "function_call_output"
                && part.GetProperty("call_id").GetString() == "image-call");
            var content = result.GetProperty("output").EnumerateArray().ToArray();
            Check(content.Count(part => part.GetProperty("type").GetString() == "input_image") == 1,
                "Responses result lost or duplicated its image block.");
            var image = content.Single(part => part.GetProperty("type").GetString() == "input_image");
            Check(image.GetProperty("image_url").GetString() == "data:image/png;base64," + encoded,
                "Responses received metadata or incorrect pixels instead of the derivative.");
            Check(content.Any(part => part.GetProperty("type").GetString() == "input_text"),
                "Responses result lost its accompanying text.");
        }
        else
        {
            var result = wire.GetProperty("messages").EnumerateArray()
                .SelectMany(message => message.GetProperty("content").EnumerateArray())
                .Single(part => part.GetProperty("type").GetString() == "tool_result"
                    && part.GetProperty("tool_use_id").GetString() == "image-call");
            var content = result.GetProperty("content").EnumerateArray().ToArray();
            Check(content.Count(part => part.GetProperty("type").GetString() == "image") == 1,
                "Anthropic result lost or duplicated its image block.");
            var image = content.Single(part => part.GetProperty("type").GetString() == "image");
            Check(image.GetProperty("source").GetProperty("data").GetString() == encoded,
                "Anthropic received metadata or incorrect pixels instead of the derivative.");
            Check(content.Any(part => part.GetProperty("type").GetString() == "text"),
                "Anthropic result lost its accompanying text.");
        }
    }

    private static JsonObject ImageTurnReply(Surface surface, string model, int request)
    {
        string text = request switch
        {
            0 => "I will inspect the synthetic image and report its dimensions.",
            1 => "I will read the separate caption file to verify the follow-up.",
            _ => "Image inspected."
        };
        string id = request == 0 ? "image-call" : "caption-call";
        string name = request == 0 ? "inspect_image" : "read_file";
        var args = request == 0
            ? JsonSerializer.SerializeToNode(new { path = "source.png", ai_callback_delay = 3, max_output_bytes = 1 })!
            : JsonSerializer.SerializeToNode(new { path = "caption.txt" })!;
        if (surface == Surface.GenerateContent)
        {
            var parts = new JsonArray(new JsonObject { ["text"] = text });
            if (request < 2)
            {
                parts.Add(new JsonObject
                {
                    ["functionCall"] = new JsonObject { ["id"] = id, ["name"] = name, ["args"] = args },
                    ["thoughtSignature"] = "fixture-signature-" + id
                });
            }
            return new JsonObject
            {
                ["modelVersion"] = model,
                ["candidates"] = new JsonArray(new JsonObject
                {
                    ["content"] = new JsonObject { ["role"] = "model", ["parts"] = parts },
                    ["finishReason"] = "STOP"
                }),
                ["usageMetadata"] = new JsonObject
                {
                    ["promptTokenCount"] = 10, ["candidatesTokenCount"] = 10, ["totalTokenCount"] = 20
                }
            };
        }
        if (surface == Surface.Responses)
        {
            var output = new JsonArray(new JsonObject
            {
                ["type"] = "message", ["role"] = "assistant", ["status"] = "completed",
                ["content"] = new JsonArray(new JsonObject { ["type"] = "output_text", ["text"] = text })
            });
            if (request < 2) output.Add(new JsonObject
            {
                ["type"] = "function_call", ["id"] = "fc-" + id, ["call_id"] = id,
                ["name"] = name, ["arguments"] = args.ToJsonString(), ["status"] = "completed"
            });
            return new JsonObject
            {
                ["status"] = "completed", ["model"] = model, ["output"] = output,
                ["usage"] = new JsonObject { ["input_tokens"] = 10, ["output_tokens"] = 10 }
            };
        }
        var content = new JsonArray(new JsonObject { ["type"] = "text", ["text"] = text });
        if (request < 2) content.Add(new JsonObject
        {
            ["type"] = "tool_use", ["id"] = id, ["name"] = name, ["input"] = args
        });
        return new JsonObject
        {
            ["model"] = model, ["content"] = content,
            ["stop_reason"] = request < 2 ? "tool_use" : "end_turn",
            ["usage"] = new JsonObject { ["input_tokens"] = 10, ["output_tokens"] = 10 }
        };
    }
}
