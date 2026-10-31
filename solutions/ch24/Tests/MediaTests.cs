using System.Net;
using System.Net.Sockets;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    public static async Task MediaProjection()
    {
        string directory = Path.Combine(Path.GetTempPath(), "ensemble-media-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(directory);
        string path = Path.Combine(directory, "pixel.png");
        byte[] bytes = Convert.FromBase64String("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+aZ1sAAAAASUVORK5CYII=");
        File.WriteAllBytes(path, bytes);
        var app = new Application(TextWriter.Null);
        try
        {
            foreach (var sample in new[]
            {
                (Vendor.Anthropic, "claude-3-haiku-20240307"),
                (Vendor.OpenAI, "gpt-4o"),
                (Vendor.Gemini, "gemini-2.5-flash")
            })
            {
                var agent = new Agent(app, new Configuration(sample.Item1, sample.Item2, "http://127.0.0.1:1"));
                var blob = new BlobPart("image/png", new BlobRef(RefKind.Path, path));
                agent.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                    parts: [new TextPart("before"), blob, new TextPart("after")]));
                var fact = agent.History().Events()[0];
                string original = JsonCodec.WriteEvent(fact);
                JsonObject request = agent.Engine().Render(agent.History().View());
                JsonArray parts;
                string encoded;
                if (sample.Item1 == Vendor.Anthropic)
                {
                    parts = request["messages"]![0]!["content"]!.AsArray();
                    Check(parts[1]!["type"]!.GetValue<string>() == "image", "Missing Anthropic image type.");
                    Check(parts[1]!["source"]!["media_type"]!.GetValue<string>() == "image/png", "Missing MIME.");
                    encoded = parts[1]!["source"]!["data"]!.GetValue<string>();
                }
                else if (sample.Item1 == Vendor.OpenAI)
                {
                    parts = request["messages"]![1]!["content"]!.AsArray();
                    string uri = parts[1]!["image_url"]!["url"]!.GetValue<string>();
                    Check(uri.StartsWith("data:image/png;base64,"), "Missing image data URI.");
                    encoded = uri[(uri.IndexOf(',') + 1)..];
                }
                else
                {
                    parts = request["contents"]![0]!["parts"]!.AsArray();
                    Check(parts[1]!["inlineData"]!["mimeType"]!.GetValue<string>() == "image/png", "Missing Gemini MIME.");
                    encoded = parts[1]!["inlineData"]!["data"]!.GetValue<string>();
                }
                Check(parts.Count == 3 && parts[0]!["text"]!.GetValue<string>() == "before"
                    && parts[2]!["text"]!.GetValue<string>() == "after", "Attachment changed part order.");
                Check(Convert.FromBase64String(encoded).SequenceEqual(bytes), "Attachment did not contain file bytes.");
                Check(JsonCodec.WriteEvent(agent.History().Events()[0]) == original, "Rendering changed canonical facts.");
                var restored = (BlobPart)JsonCodec.ReadEvent(original).Parts[1];
                Check(restored.Ref.Kind == RefKind.Path && restored.Ref.Locator == path
                    && !original.Contains(encoded), "Log did not retain locator-only media.");
                agent.Destroy();
            }

            // One silent PCM sample is a complete WAV. The PDF is a small
            // document; the MP4 fixture is a container header, not a playable
            // movie. These checks test projection, not media decoding.
            using var wavBuffer = new MemoryStream();
            using (var wav = new BinaryWriter(wavBuffer, System.Text.Encoding.ASCII, leaveOpen: true))
            {
                wav.Write("RIFF"u8);
                wav.Write(38);
                wav.Write("WAVEfmt "u8);
                wav.Write(16);
                wav.Write((short)1);
                wav.Write((short)1);
                wav.Write(8000);
                wav.Write(16000);
                wav.Write((short)2);
                wav.Write((short)16);
                wav.Write("data"u8);
                wav.Write(2);
                wav.Write((short)0);
            }
            var fixtures = new[]
            {
                ("audio/wav", wavBuffer.ToArray()),
                ("application/pdf", System.Text.Encoding.ASCII.GetBytes(
                    "%PDF-1.4\n1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj\n"
                    + "2 0 obj<</Type/Pages/Count 0/Kids[]>>endobj\n"
                    + "trailer<</Root 1 0 R>>\n%%EOF\n")),
                ("video/mp4", Convert.FromHexString("000000186674797069736F6D0000020069736F6D6D703431"))
            };
            foreach (var (mime, data) in fixtures)
            {
                string mediaPath = Path.Combine(directory, "attachment");
                File.WriteAllBytes(mediaPath, data);
                var mediaAgent = new Agent(app, new Configuration(Vendor.Gemini,
                    "gemini-2.5-flash", "http://127.0.0.1:1"));
                mediaAgent.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                    parts: [new TextPart("before"), new BlobPart(mime, new BlobRef(RefKind.Path, mediaPath)),
                        new TextPart("after")]));
                string before = JsonCodec.WriteEvent(mediaAgent.History().Events()[0]);
                var projected = mediaAgent.Engine().Render(mediaAgent.History().View());
                var content = projected["contents"]![0]!["parts"]!.AsArray();
                var inline = content[1]!["inlineData"]!;
                Check(inline["mimeType"]!.GetValue<string>() == mime, "Wrong attachment MIME.");
                Check(Convert.FromBase64String(inline["data"]!.GetValue<string>()).SequenceEqual(data),
                    "Non-image attachment bytes changed.");
                Check(content.Count == 3 && content[0]!["text"]!.GetValue<string>() == "before"
                    && content[2]!["text"]!.GetValue<string>() == "after", "Non-image order changed.");
                Check(before == JsonCodec.WriteEvent(mediaAgent.History().Events()[0]),
                    "Non-image projection changed History.");
                mediaAgent.Destroy();
            }

            var unknown = new Agent(app, new Configuration(Vendor.Anthropic, "unknown-model", "http://127.0.0.1:1"));
            unknown.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human, parts: [new TextPart("text") ]));
            unknown.Engine().Render(unknown.History().View());
            unknown.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new BlobPart("image/png", new BlobRef(RefKind.Path, path))]));
            Refused(unknown, "unknown-model", "image");

            var mismatch = new Agent(app, new Configuration(Vendor.OpenAI, "claude-3-haiku-20240307", "http://127.0.0.1:1"));
            mismatch.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new BlobPart("image/png", new BlobRef(RefKind.Path, path))]));
            Refused(mismatch, "claude-3-haiku-20240307", "image");

            var uriAgent = new Agent(app, new Configuration(Vendor.Anthropic, "claude-3-haiku-20240307", "http://127.0.0.1:1"));
            uriAgent.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new BlobPart("image/png", new BlobRef(RefKind.Uri, "https://example.invalid/image.png"))]));
            Refused(uriAgent, "claude-3-haiku-20240307", "local paths");

            // Capability refusal happens before reading the deliberately absent
            // file, recording RequestSent, or contacting even a local provider.
            var listener = new TcpListener(IPAddress.Loopback, 0);
            listener.Start();
            try
            {
                int port = ((IPEndPoint)listener.LocalEndpoint).Port;
                var audio = new Agent(app, new Configuration(Vendor.Anthropic, "claude-3-haiku-20240307",
                    "http://127.0.0.1:" + port, "dummy"));
                using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(5));
                try
                {
                    await audio.SubmitAsync([new BlobPart("audio/wav", new BlobRef(RefKind.Path, path + ".absent"))])
                        .WaitAsync(deadline.Token);
                    throw new Exception("Unsupported audio was accepted.");
                }
                catch (NotSupportedException error)
                {
                    Check(error.Message.Contains("claude-3-haiku-20240307") && error.Message.Contains("audio"),
                        "Media refusal did not identify model and category.");
                }
                Check(!listener.Pending() && !audio.History().Events().Any(e => e.Kind == EventKind.RequestSent),
                    "Unsupported media reached transport.");
            }
            finally
            {
                listener.Stop();
            }

            var missing = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4o", "http://127.0.0.1:1"));
            missing.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new BlobPart("image/png", new BlobRef(RefKind.Path, path + ".absent"))]));
            try
            {
                missing.Engine().Render(missing.History().View());
                throw new Exception("Missing file was silently dropped.");
            }
            catch (FileNotFoundException) { }
        }
        finally
        {
            app.Destroy();
            Directory.Delete(directory, recursive: true);
        }

        static void Refused(Agent agent, string model, string category)
        {
            try
            {
                agent.Engine().Render(agent.History().View());
                throw new Exception("Expected media refusal.");
            }
            catch (NotSupportedException error)
            {
                Check(error.Message.Contains(model) && error.Message.Contains(category), "Incomplete refusal diagnostic.");
            }
        }
    }
}
