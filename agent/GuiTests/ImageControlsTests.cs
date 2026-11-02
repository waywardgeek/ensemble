using System.Net;
using System.Net.Sockets;
using System.Net.WebSockets;
using System.Text.Json;
using Ensemble;
using EnsembleGui;
using SkiaSharp;

static class ImageControlsTests
{
    static void Check(bool condition, string message)
    {
        if (!condition) throw new InvalidOperationException(message);
    }

    public static async Task RunAsync()
    {
        string work = Path.Combine(Path.GetTempPath(), "gui-images-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(Path.Combine(work, "skills", "plain"));
        File.WriteAllText(Path.Combine(work, "skills", "plain", "SKILL.md"),
            "---\nname: plain\ndescription: Plain\ntype: primary\n---\nPlain\n");
        Console.WriteLine("Image controls workspace: " + work);
        var app = new Ensemble.Ensemble(TextWriter.Null, dataDirectory: Path.Combine(work, "fleet"));
        var agent = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-6-astra", "http://127.0.0.1:1", surface: Surface.Responses),
            name: "images", skillSpec: new SkillSpec(Path.Combine(work, "skills"), "plain"));
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(90));
        var token = deadline.Token;
        await agent.SkillManager()!.InitializeAsync(token);
        await agent.PostAsync(new AgentInput(InputKind.Pause));
        using var portProbe = new TcpListener(IPAddress.Loopback, 0);
        portProbe.Start();
        int port = ((IPEndPoint)portProbe.LocalEndpoint).Port;
        portProbe.Stop();
        string url = $"http://127.0.0.1:{port}";
        var builder = WebApplication.CreateBuilder();
        builder.Logging.ClearProviders();
        builder.WebHost.UseUrls(url);
        await using var host = builder.Build();
        using var stop = new CancellationTokenSource();
        string tracePath = Path.Combine(work, "gui.log");
        Task serving = Gui.ServeAsync(host, agent, new GuiOptions(port,
            SettingsPath: Path.Combine(work, "settings.json"), TracePath: tracePath), stop.Token);
        using var socket = new ClientWebSocket();
        socket.Options.SetRequestHeader("Origin", url);
        async Task<JsonElement> Frame(string type)
        {
            byte[] buffer = new byte[8192];
            while (true)
            {
                using var bytes = new MemoryStream();
                WebSocketReceiveResult received;
                do
                {
                    received = await socket.ReceiveAsync(buffer, token);
                    Check(received.MessageType == WebSocketMessageType.Text, "Socket closed unexpectedly.");
                    bytes.Write(buffer, 0, received.Count);
                } while (!received.EndOfMessage);
                using var json = JsonDocument.Parse(bytes.ToArray());
                if (json.RootElement.GetProperty("type").GetString() == type) return json.RootElement.Clone();
            }
        }
        Task Send(object value) => socket.SendAsync(JsonSerializer.SerializeToUtf8Bytes(value), WebSocketMessageType.Text, true, token);
        int Images() => agent.History().Events().SelectMany(e => e.Parts).OfType<BlobPart>().Count();
        try
        {
            await host.StartAsync(token);
            await socket.ConnectAsync(new Uri($"ws://127.0.0.1:{port}/ws"), token);
            var replay = await Frame("replay");
            string target = replay.GetProperty("agent_id").GetString()!;
            using var bitmap = new SKBitmap(new SKImageInfo(1024, 1024, SKColorType.Rgba8888, SKAlphaType.Unpremul));
            new Random(71).NextBytes(bitmap.GetPixelSpan());
            for (int i = 3; i < bitmap.ByteCount; i += 4) bitmap.GetPixelSpan()[i] = 255;
            using var image = SKImage.FromBitmap(bitmap);
            using var encoded = image.Encode(SKEncodedImageFormat.Png, 100);
            string source = Convert.ToBase64String(encoded.ToArray());
            Check(source.Length > 1024 * 1024, "Fixture did not exceed ordinary control limit.");
            object Envelope(long sequence, string address, string data) => new
            {
                type = "image_prompt", sequence, agent_id = address, text = "WITH_IMAGE_TEXT",
                images = new[] { new { mime = "image/png", data } }
            };
            await Send(Envelope(1, target, source));
            var accepted = await Frame("image_accepted");
            Check(accepted.GetProperty("sequence").GetInt64() == 1 && Images() == 1, "Image was not admitted exactly once.");
            var dimensions = accepted.GetProperty("images")[0];
            Check(dimensions.GetProperty("original_width").GetInt32() == 1024
                && dimensions.GetProperty("width").GetInt32() == 1024
                && !dimensions.GetProperty("resized").GetBoolean(), "Acknowledgment omitted preparation metadata.");
            var retained = agent.History().Events().SelectMany(e => e.Parts).OfType<BlobPart>().Single();
            using var decoded = SKBitmap.Decode(Convert.FromBase64String(retained.Ref.Locator.Split(',')[1]));
            Check(decoded.GetPixel(51, 77) == bitmap.GetPixel(51, 77), "Actual retained pixels differ.");
            Check(agent.History().Events().SelectMany(e => e.Parts).OfType<TextPart>().Any(p => p.Text == "WITH_IMAGE_TEXT"),
                "Accompanying text was lost.");
            async Task Refuse(object value)
            {
                await Send(value);
                await Frame("error");
                Check(Images() == 1, "Refused input added an image to History.");
            }
            await Refuse(Envelope(1, target, source));
            await Refuse(Envelope(2, "wrong-agent", source));
            await Refuse(Envelope(3, target, ""));
            await Refuse(Envelope(4, target, "AAAA"));
            await Refuse(new
            {
                type = "image_prompt", sequence = 5, agent_id = target,
                images = new[] { new { mime = "image/png", data = source }, new { mime = "image/png", data = "AAAA" } }
            });
            await Refuse(new { type = "image_prompt", sequence = 6, images = new[] { new { mime = "image/png", data = source } } });
            await Refuse(new { type = "prompt", agent_id = target, images = new[] { new { mime = "image/png", data = source } } });
            await Refuse(new { type = "prompt", agent_id = target, text = new string('x', 1024 * 1024 + 1) });
            await Send(Envelope(7, target, source));
            await Frame("image_accepted");
            Check(Images() == 2, "Refusals disrupted subsequent valid admission.");
            // Fill the remaining count allowance with tiny prepared images;
            // the next WebSocket image must be refused with its sequence intact.
            using var smallBitmap = new SKBitmap(4, 4);
            smallBitmap.Erase(SKColors.Blue);
            using var smallImage = SKImage.FromBitmap(smallBitmap);
            using var smallPng = smallImage.Encode(SKEncodedImageFormat.Png, 100);
            var small = agent.PrepareImage(smallPng.ToArray()).Image;
            for (int i = 0; i < 30; i++)
                await agent.PostAsync(new AgentInput(InputKind.Message, "fill", [small])).WaitAsync(token);
            int beforeRefusal = agent.History().Events().Length;
            await Send(Envelope(8, target, source));
            var overflow = await Frame("error");
            Check(overflow.GetProperty("image_sequence").GetInt64() == 8
                && overflow.GetProperty("message").GetString()!.Contains("micro_handoff"),
                "Overflow refusal was not correlated or actionable.");
            Check(Images() == 32 && agent.History().Events().Length == beforeRefusal,
                "Overflow admitted an image or caption.");
            // Inbound envelopes and outbound History frames both pass through
            // the same trace projection. No pixel payload belongs in this log.
            string trace = File.ReadAllText(tracePath);
            Check(trace.Contains("media_redacted") && !trace.Contains(source[..128]), "Trace retained source image bytes.");
            Check(!trace.Contains(retained.Ref.Locator.Split(',')[1][..128]), "Trace retained prepared image bytes.");
        }
        finally
        {
            socket.Abort();
            stop.Cancel();
            try { await serving.WaitAsync(TimeSpan.FromSeconds(15)); }
            finally
            {
                await host.StopAsync(CancellationToken.None);
                app.Destroy();
            }
        }
        Console.WriteLine("PASS GUI images: payloads, source preparation, explicit routing, duplicate refusal, control limits and trace privacy");
    }
}
