using Ensemble;
using EnsembleGui;
using SkiaSharp;
using System.Text.Json;
using System.Threading.Channels;

static class BrowserCaptureTests
{
    static void Check(bool condition, string message)
    {
        if (!condition) throw new InvalidOperationException(message);
    }
    static JsonElement Json(object value) => JsonSerializer.SerializeToElement(value);
    static JsonElement Request(long id) => Json(new { jsonrpc = "2.0", id,
        method = "tools/call", @params = new { name = "gui_capture", arguments = new { } } });
    static JsonElement Reply(long id, object block) => Json(new { jsonrpc = "2.0", id,
        result = new { content = new[] { block } } });
    static void Refuses(Action action, string message)
    {
        try { action(); }
        catch (InvalidDataException) { return; }
        throw new InvalidOperationException(message);
    }
    public static async Task RunAsync()
    {
        string work = Path.Combine(Path.GetTempPath(), "capture-tunnel-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(work);
        var root = new Ensemble.Ensemble(TextWriter.Null, dataDirectory: work);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(30));
        var token = deadline.Token;
        try
        {
            var target = new Agent(root, new Configuration(Vendor.OpenAI, "gpt-6-astra",
                "http://127.0.0.1:1", surface: Surface.Responses), name: "capture");
            var output = Channel.CreateUnbounded<string>();
            await using var tunnel = new McpTunnel(output.Writer, "capture", target);
            async Task<JsonElement> Result()
                => await tunnel.ReceiveAsync(token) ?? throw new InvalidOperationException("Tunnel closed unexpectedly.");
            async Task Begin(long id)
            {
                await tunnel.SendAsync(Request(id), token);
                using var sent = JsonDocument.Parse(await output.Reader.ReadAsync(token));
                Check(sent.RootElement.GetProperty("data").GetProperty("id").GetInt64() == id,
                    "Capture RPC identity changed.");
            }
            using var bitmap = new SKBitmap(4096, 64);
            using (var canvas = new SKCanvas(bitmap))
            using (var red = new SKPaint { Color = SKColors.Red })
            using (var blue = new SKPaint { Color = SKColors.Blue })
            {
                canvas.DrawRect(0, 0, 2048, 64, red);
                canvas.DrawRect(2048, 0, 2048, 64, blue);
            }
            using var image = SKImage.FromBitmap(bitmap);
            using var encoded = image.Encode(SKEncodedImageFormat.Png, 100);
            object pixels = new { type = "image", mimeType = "image/png", data = Convert.ToBase64String(encoded.ToArray()) };
            await Begin(1);
            Check(tunnel.IsCaptureReply(Reply(1, pixels)), "Pending reply not recognized.");
            tunnel.Accept(Reply(1, pixels));
            var result = await Result();
            var decoded = McpContent.Decode("call-1", result.GetProperty("result"));
            Check(!decoded.IsError && decoded.Parts.Length == 2, "Prepared result lacks text and pixels.");
            var blob = decoded.Parts.OfType<BlobPart>().Single();
            using var actual = SKBitmap.Decode(blob.Ref.Data());
            Check(actual.Width == 2048 && actual.Height == 32, "Capture was not resized by the Agent.");
            Check(actual.GetPixel(100, 16).Red > 240 && actual.GetPixel(1900, 16).Blue > 240,
                "Capture pixels were lost or changed sides.");
            target.Engine().ValidateMedia([decoded]);
            Refuses(() => tunnel.Accept(Reply(1, pixels)), "Duplicate capture image was admitted.");
            Console.WriteLine("PASS capture preparation: actual resized PNG and duplicate refusal");

            object[] invalid = [
                new { type = "text", text = "screenshot.png 4096x64" },
                new { type = "image", mimeType = "image/png" },
                new { type = "image", mimeType = "image/jpeg", data = "AAAA" },
                new { type = "image", mimeType = "image/png", data = "AAAA" },
                new { type = "image", mimeType = "image/png", data = "not base64" }
            ];
            long sequence = 2;
            foreach (object invalidBlock in invalid)
            {
                await Begin(sequence);
                tunnel.Accept(Reply(sequence, invalidBlock));
                Check((await Result()).GetProperty("result").GetProperty("isError").GetBoolean(),
                    "Invalid capture must be an ordinary paired error.");
                sequence++;
            }
            Console.WriteLine("PASS capture refusal: metadata, missing bytes, MIME, signature and base64");

            await Begin(20);
            await tunnel.SendAsync(Request(21), token);
            Check((await Result()).GetProperty("id").GetInt64() == 21, "Concurrent refusal lost identity.");
            Check(!output.Reader.TryRead(out _), "Concurrent capture reached browser.");
            var cancel = Json(new { jsonrpc = "2.0", method = "$/cancelRequest", @params = new { id = 20 } });
            await tunnel.SendAsync(cancel, token);
            await output.Reader.ReadAsync(token);
            Check(!tunnel.IsCaptureReply(Reply(20, pixels)), "Cancelled capture remains pending.");
            Refuses(() => tunnel.Accept(Reply(20, pixels)), "Late cancelled pixels were admitted.");
            await Begin(22);
            tunnel.Accept(Json(new { jsonrpc = "2.0", id = 22, result = new { isError = true,
                content = new[] { new { type = "text", text = "User declined capture." } } } }));
            Check((await Result()).GetProperty("result").GetProperty("isError").GetBoolean(),
                "User refusal was not retained.");
            Console.WriteLine("PASS capture cancellation, concurrency and subsequent recovery");

            var ordinary = Json(new { jsonrpc = "2.0", id = "ordinary-string-id", result = new { text = "ok" } });
            tunnel.Accept(ordinary);
            Check((await Result()).GetProperty("id").GetString() == "ordinary-string-id",
                "Ordinary MCP traffic changed.");
            var unsupported = new Agent(root, new Configuration(Vendor.OpenAI, "unverified-capture-model",
                "http://127.0.0.1:1", surface: Surface.Responses), name: "unsupported");
            var refusedOutput = Channel.CreateUnbounded<string>();
            await using var refused = new McpTunnel(refusedOutput.Writer, "unsupported", unsupported);
            await refused.SendAsync(Request(1), token);
            var unsupportedReply = await refused.ReceiveAsync(token);
            Check(unsupportedReply!.Value.GetProperty("result").GetProperty("isError").GetBoolean(),
                "Unsupported model was not refused.");
            Check(!refusedOutput.Reader.TryRead(out _), "Unsupported model caused a browser capture request.");
            var blockedOutput = Channel.CreateBounded<string>(1);
            Check(blockedOutput.Writer.TryWrite("occupied"), "Could not establish output backpressure.");
            await using (var blocked = new McpTunnel(blockedOutput.Writer, "capture", target))
            {
                using var stopSend = CancellationTokenSource.CreateLinkedTokenSource(token);
                Task sending = blocked.SendAsync(Request(30), stopSend.Token);
                Check(!sending.IsCompleted, "Capture did not wait for real backpressure.");
                stopSend.Cancel();
                try { await sending; throw new InvalidOperationException("Blocked send ignored cancellation."); }
                catch (OperationCanceledException) { }
                Check(!blocked.IsCaptureReply(Reply(30, pixels)), "Failed send retained capture authority.");
                await blockedOutput.Reader.ReadAsync(token);
                await blocked.SendAsync(Request(31), token);
                using var forwarded = JsonDocument.Parse(await blockedOutput.Reader.ReadAsync(token));
                Check(forwarded.RootElement.GetProperty("data").GetProperty("id").GetInt64() == 31,
                    "Cancelled send prevented a later capture.");
            }
            Console.WriteLine("PASS cancelled outbound send releases reservation under backpressure");
            await Begin(23);
            tunnel.SetCaptureSelection(unsupported);
            Check((await Result()).GetProperty("result").GetProperty("isError").GetBoolean(),
                "Selection change did not refuse the pending capture.");
            Refuses(() => tunnel.Accept(Reply(23, pixels)), "Selection change allowed late pixels.");
            await tunnel.SendAsync(Request(24), token);
            Check((await Result()).GetProperty("result").GetProperty("isError").GetBoolean(),
                "Capture was permitted while viewing another Agent.");
            Check(!output.Reader.TryRead(out _), "Unselected capture reached browser.");
            tunnel.SetCaptureSelection(target);
            Refuses(() => tunnel.Accept(Reply(23, pixels)), "Selecting back revived an old capture.");
            await Begin(25);
            await tunnel.DisposeAsync();
            Refuses(() => tunnel.Accept(Reply(25, pixels)), "Disposed capture admitted pixels.");
            Console.WriteLine("PASS capability-before-dispatch, ordinary RPC compatibility and disposal");
        }
        finally { root.Destroy(); }
    }
}
