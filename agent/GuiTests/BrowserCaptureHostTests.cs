using System.Net;
using System.Net.Sockets;
using System.Net.WebSockets;
using System.Text.Json;
using Ensemble;
using EnsembleGui;
using SkiaSharp;

static class BrowserCaptureHostTests
{
    static void Check(bool value, string message)
    {
        if (!value) throw new InvalidOperationException(message);
    }
    public static async Task RunAsync()
    {
        string work = Path.Combine(Path.GetTempPath(), "gui-capture-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(Path.Combine(work, "skills", "plain"));
        File.WriteAllText(Path.Combine(work, "skills", "plain", "SKILL.md"),
            "---\nname: plain\ndescription: Capture fixture\ntype: primary\ntools: gui_capture\nmcp_servers:\n  - name: page\n    transport: websocket\n---\nUse the synthetic capture peer.\n");
        Console.WriteLine("Capture host workspace: " + work);
        var root = new Ensemble.Ensemble(TextWriter.Null, dataDirectory: Path.Combine(work, "fleet"));
        var agent = new Agent(root, new Configuration(Vendor.OpenAI, "gpt-6-astra", "http://127.0.0.1:1", surface: Surface.Responses),
            name: "capture", skillSpec: new SkillSpec(Path.Combine(work, "skills"), "plain"));
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(90));
        var token = deadline.Token;
        using var probe = new TcpListener(IPAddress.Loopback, 0);
        probe.Start();
        int port = ((IPEndPoint)probe.LocalEndpoint).Port;
        probe.Stop();
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
        string target = "";
        bool discovered = false;
        var requests = new List<Task>();
        Task Send(object value) => socket.SendAsync(JsonSerializer.SerializeToUtf8Bytes(value), WebSocketMessageType.Text, true, token);
        async Task<JsonElement> Frame(string wanted)
        {
            byte[] buffer = new byte[8192];
            while (true)
            {
                using var bytes = new MemoryStream();
                WebSocketReceiveResult received;
                do
                {
                    received = await socket.ReceiveAsync(buffer, token);
                    Check(received.MessageType == WebSocketMessageType.Text, "Unexpected socket closure.");
                    bytes.Write(buffer, 0, received.Count);
                } while (!received.EndOfMessage);
                using var document = JsonDocument.Parse(bytes.ToArray());
                var frame = document.RootElement;
                string type = frame.GetProperty("type").GetString()!;
                if (type == "mcp")
                {
                    var rpc = frame.GetProperty("data");
                    if (rpc.TryGetProperty("method", out var method))
                    {
                        string name = method.GetString()!;
                        if (name == "initialize" || name == "tools/list")
                        {
                            object result = name == "initialize"
                                ? new { protocolVersion = "2024-11-05", capabilities = new { tools = new { } }, serverInfo = new { name = "capture-fixture", version = "1" } }
                                : new { tools = new[] { new { name = "gui_capture", description = "Synthetic capture fixture", inputSchema = new { type = "object", properties = new { } } } } };
                            await Send(new { type = "mcp", agent_id = frame.GetProperty("agent_id").GetString(),
                                data = new { jsonrpc = "2.0", id = rpc.GetProperty("id").GetInt64(), result } });
                            if (name == "tools/list")
                            {
                                discovered = true;
                                if (wanted == "capture_discovered") return frame.Clone();
                            }
                            continue;
                        }
                        if (name.StartsWith("notifications/", StringComparison.Ordinal)) continue;
                    }
                }
                if (type == wanted) return frame.Clone();
            }
        }
        async Task<(long ID, Task<JsonElement> Reply)> Begin()
        {
            var client = agent.Mcp()!.Browser()!.Client()!;
            Task<JsonElement> reply = client.RequestAsync("tools/call", JsonSerializer.SerializeToElement(new { name = "gui_capture", arguments = new { } }), token);
            requests.Add(reply);
            var frame = await Frame("mcp");
            var rpc = frame.GetProperty("data");
            Check(rpc.GetProperty("params").GetProperty("name").GetString() == "gui_capture", "Wrong RPC dispatched.");
            return (rpc.GetProperty("id").GetInt64(), reply);
        }
        try
        {
            await host.StartAsync(token);
            await socket.ConnectAsync(new Uri($"ws://127.0.0.1:{port}/ws"), token);
            target = (await Frame("replay")).GetProperty("agent_id").GetString()!;
            Task initialization = agent.SkillManager()!.InitializeAsync(token);
            requests.Add(initialization);
            if (!discovered) await Frame("capture_discovered");
            await initialization;
            Check(agent.ToolEnabled("gui_capture") && agent.Mcp()?.Browser()?.Client() is not null,
                "The explicit fixture skill did not claim and grant its browser tool.");
            await agent.PostAsync(new AgentInput(InputKind.Pause));
            using var bitmap = new SKBitmap(new SKImageInfo(1024, 384, SKColorType.Rgba8888, SKAlphaType.Unpremul));
            new Random(191).NextBytes(bitmap.GetPixelSpan());
            for (int i = 3; i < bitmap.ByteCount; i += 4) bitmap.GetPixelSpan()[i] = 255;
            using var image = SKImage.FromBitmap(bitmap);
            using var png = image.Encode(SKEncodedImageFormat.Png, 100);
            string source = Convert.ToBase64String(png.ToArray());
            Check(source.Length > 1024 * 1024, "Fixture must exceed ordinary control limit.");
            object Envelope(long id, string address, object block) => new { type = "mcp", agent_id = address,
                data = new { jsonrpc = "2.0", id, result = new { content = new[] { block } } } };
            object pixels = new { type = "image", mimeType = "image/png", data = source };
            await Send(Envelope(9999, target, pixels));
            await Frame("error");
            var first = await Begin();
            await Send(Envelope(first.ID, "wrong-target", pixels));
            await Frame("error");
            Check(!first.Reply.IsCompleted, "Wrong-target reply consumed the real request.");
            await Send(Envelope(first.ID, target, pixels));
            var result = await first.Reply.WaitAsync(token);
            var decoded = McpContent.Decode("capture-wire", result);
            Check(!decoded.IsError, "Matching large reply was refused.");
            using var prepared = SKBitmap.Decode(decoded.Parts.OfType<BlobPart>().Single().Ref.Data());
            Check(prepared.Width == 1024 && prepared.Height == 384
                && prepared.GetPixel(51, 77) == bitmap.GetPixel(51, 77), "Real prepared pixels changed.");
            await Send(Envelope(first.ID, target, pixels));
            await Frame("error");
            Console.WriteLine("PASS capture host: matching large pixels, wrong-target and duplicate/stale refusal");

            var switched = await Begin();
            await Send(new { type = "select_agent", agent_id = target });
            await Frame("replay");
            Check((await switched.Reply.WaitAsync(token)).GetProperty("isError").GetBoolean(), "Selection did not invalidate capture.");
            await Send(Envelope(switched.ID, target, pixels));
            await Frame("error");
            var malformed = await Begin();
            await Send(Envelope(malformed.ID, target, new { type = "image", data = source }));
            Check((await malformed.Reply.WaitAsync(token)).GetProperty("isError").GetBoolean(), "Malformed capture escaped paired refusal.");
            await Send(new { type = "prompt", text = new string('q', 1024 * 1024 + 1) });
            await Frame("error");
            string trace = await File.ReadAllTextAsync(tracePath, token);
            Check(trace.Contains("media_redacted", StringComparison.Ordinal), "Media redaction was not exercised.");
            Check(!trace.Contains(source.Substring(256, 256), StringComparison.Ordinal), "MCP image bytes leaked into trace.");
            Console.WriteLine("PASS capture host: selection invalidation, ordinary size ceiling and malformed-media trace privacy");
        }
        finally
        {
            socket.Abort();
            stop.Cancel();
            try
            {
                await host.StopAsync();
                await serving;
                try { await Task.WhenAll(requests); } catch { /* Original request errors remain on their tasks. */ }
            }
            finally { root.Destroy(); }
        }
    }
}
