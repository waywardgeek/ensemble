using System.Collections.Immutable;
using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using SkiaSharp;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    public static async Task ImageInputs()
    {
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(30));
        var received = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var app = new Application(TextWriter.Null);
        Task? server = null;
        try
        {
            int port = ((IPEndPoint)listener.LocalEndpoint).Port;
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "claude-3-haiku-20240307",
                "http://127.0.0.1:" + port, "dummy-local-only"));
            using var bitmap = new SKBitmap(12, 8);
            bitmap.Erase(SKColors.Magenta);
            using var image = SKImage.FromBitmap(bitmap);
            using var encoded = image.Encode(SKEncodedImageFormat.Png, 100);
            BlobPart prepared = agent.PrepareImage(encoded.ToArray()).Image;
            string pixels = prepared.Ref.Locator.Split(',', 2)[1];
            server = Task.Run(async () =>
            {
                for (int request = 0; request < 2; request++)
                {
                    using var client = await listener.AcceptTcpClientAsync(deadline.Token);
                    using var stream = client.GetStream();
                    string body = await ReadRequest(stream, deadline.Token);
                    using var json = JsonDocument.Parse(body);
                    var content = json.RootElement.GetProperty("messages").EnumerateArray()
                        .SelectMany(message => message.GetProperty("content").EnumerateArray())
                        .Where(part => part.GetProperty("type").GetString() == "image").ToArray();
                    Check(content.Length == request + 1, "Wrong number of images in frozen request.");
                    Check(content.All(part => part.GetProperty("source").GetProperty("data").GetString() == pixels),
                        "Provider did not receive the admitted derivative bytes.");
                    if (request == 0)
                    {
                        Check(!body.Contains("LATER_IMAGE"), "Hint leaked into frozen request.");
                        received.SetResult();
                        await release.Task.WaitAsync(deadline.Token);
                    }
                    else Check(body.Contains("LATER_IMAGE"), "Image hint lost accompanying text.");
                    await Reply(stream, request == 0 ? "initial" : "images addressed", deadline.Token);
                }
            });
            await agent.PostAsync(new AgentInput(InputKind.Pause)).WaitAsync(deadline.Token);
            async Task Refuse(AgentInput input)
            {
                int before = agent.History().Events().Count(fact => fact.Kind == EventKind.HintReceived);
                try
                {
                    await agent.PostAsync(input).WaitAsync(deadline.Token);
                    throw new InvalidOperationException("Invalid attachment input was admitted.");
                }
                catch (ArgumentException) { }
                Check(agent.History().Events().Count(fact => fact.Kind == EventKind.HintReceived) == before,
                    "Refused images entered History.");
            }
            await Refuse(new AgentInput(InputKind.Resume, Images: [prepared]));
            await Refuse(new AgentInput(InputKind.Ephemeral, Images: [prepared]));
            await Refuse(new AgentInput(InputKind.Message, Images: [prepared, prepared, prepared, prepared, prepared]));
            await Refuse(new AgentInput(InputKind.Message, Images:
                [new BlobPart("image/png", new BlobRef(RefKind.Path, "/not-an-admitted-image"))]));
            await Refuse(new AgentInput(InputKind.Message, Images:
                [new BlobPart("image/png", new BlobRef(RefKind.Uri, "data:image/png;base64,!!!!"))]));
            var overImageLimit = new BlobPart("image/png", new BlobRef(RefKind.Uri,
                "data:image/png;base64," + new string('A', ((Agent.MaximumImageOutputBytes + 2) / 3) * 4)));
            await Refuse(new AgentInput(InputKind.Message, Images: [overImageLimit]));
            // Three six-MiB payloads exceed the aggregate admission limit even
            // though each fits individually. These need not be raster-decoded.
            var sixMiB = new BlobPart("image/png", new BlobRef(RefKind.Uri,
                "data:image/png;base64," + new string('A', 8 * 1024 * 1024)));
            await Refuse(new AgentInput(InputKind.Message, Images: [sixMiB, sixMiB, sixMiB]));
            await agent.PostAsync(new AgentInput(InputKind.Message, "FIRST_IMAGE", [prepared])).WaitAsync(deadline.Token);
            Check(agent.History().Events().SelectMany(fact => fact.Parts).OfType<BlobPart>().Count() == 1,
                "Idle image input was not recorded once.");
            await agent.PostAsync(new AgentInput(InputKind.Resume)).WaitAsync(deadline.Token);
            await received.Task.WaitAsync(deadline.Token);
            await agent.PostAsync(new AgentInput(InputKind.Hint, "LATER_IMAGE", [prepared])).WaitAsync(deadline.Token);
            release.TrySetResult();
            await agent.FlushAsync(deadline.Token);
            await server;
            Check(agent.History().Events().SelectMany(fact => fact.Parts).OfType<BlobPart>().Count() == 2,
                "Busy hint lost or duplicated its image in canonical History.");
            var unsupported = new Agent(app, new Configuration(Vendor.Anthropic, "unknown-image-model",
                "http://127.0.0.1:1", "dummy-local-only"));
            await unsupported.PostAsync(new AgentInput(InputKind.Pause)).WaitAsync(deadline.Token);
            try
            {
                await unsupported.PostAsync(new AgentInput(InputKind.Message, Images: [prepared])).WaitAsync(deadline.Token);
                throw new InvalidOperationException("Unsupported model accepted an image.");
            }
            catch (ArgumentException) { }
            await unsupported.PostAsync(new AgentInput(InputKind.Message, "text still works")).WaitAsync(deadline.Token);
        }
        finally
        {
            release.TrySetResult();
            deadline.Cancel();
            listener.Stop();
            app.Destroy();
            if (server is not null)
            {
                // Normal-path server assertions were awaited above. Cleanup
                // joins the fake even when an earlier assertion already failed.
                try { await server; } catch { }
            }
        }
    }
}
