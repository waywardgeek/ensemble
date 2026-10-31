using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    public static async Task RootedMedia()
    {
        string workspace = Path.Combine(Path.GetTempPath(), "ensemble-media-root-" + Guid.NewGuid().ToString("N"));
        string inside = Path.Combine(workspace, "inside");
        string outside = Path.Combine(workspace, "outside.png");
        Directory.CreateDirectory(inside);
        byte[] bytes = [1, 2, 3, 4, 5];
        File.WriteAllBytes(Path.Combine(inside, "image.png"), bytes);
        File.WriteAllBytes(outside, [9, 8, 7]);
        File.CreateSymbolicLink(Path.Combine(inside, "inward.png"), "image.png");
        File.CreateSymbolicLink(Path.Combine(inside, "outward.png"), outside);
        var app = new Application(TextWriter.Null);
        try
        {
            foreach (var (vendor, model) in new[]
            {
                (Vendor.Anthropic, "claude-3-haiku-20240307"),
                (Vendor.OpenAI, "gpt-4o"),
                (Vendor.Gemini, "gemini-2.5-flash")
            })
            {
                var agent = new Agent(app, new Configuration(vendor, model, "http://127.0.0.1:1",
                    sandboxRoot: inside), Path.Combine(inside, "jobs"));
                foreach (string path in new[] { outside, "../outside.png", "outward.png" })
                {
                    RefusePath(() => Render(new BlobRef(RefKind.Path, path)), vendor + " outside attachment");
                    RefusePath(() => Render(new BlobRef(RefKind.Path, path), ephemeral: true),
                        vendor + " ephemeral attachment");
                }
                string encoded = Convert.ToBase64String(bytes);
                foreach (string path in new[] { "image.png", "inward.png", Path.Combine(inside, "image.png") })
                {
                    Check(Render(new BlobRef(RefKind.Path, path)).Contains(encoded),
                        vendor + " lost rooted attachment bytes");
                }
                Check(Render(new BlobRef(RefKind.Uri, "data:image/png;base64," + encoded)).Contains(encoded),
                    vendor + " lost inline attachment bytes");

                if (vendor == Vendor.Anthropic)
                {
                    // This adapter supports media inside correlated tool results.
                    // It must use the same boundary as ordinary and ephemeral parts.
                    RefusePath(() => Render(new BlobRef(RefKind.Path, outside), nested: true),
                        "Nested tool-result attachment");
                    Check(Render(new BlobRef(RefKind.Path, "inward.png"), nested: true).Contains(encoded),
                        "Nested rooted attachment was lost");
                }
                agent.Destroy();

                string Render(BlobRef reference, bool ephemeral = false, bool nested = false)
                {
                    var blob = new BlobPart("image/png", reference);
                    Context context;
                    if (nested)
                    {
                        using var arguments = JsonDocument.Parse("{}");
                        context = new Context(dialogue:
                        [
                            new Entry(1, Actor.Agent, [new ToolCallPart("media", agent.Configuration().Target(),
                                "capture", arguments.RootElement.Clone())]),
                            new Entry(2, Actor.Tool, [new ToolResultPart("media", [blob])])
                        ]);
                    }
                    else if (ephemeral)
                    {
                        context = new Context(ephemeral: [blob]);
                    }
                    else
                    {
                        context = new Context(dialogue: [new Entry(1, Actor.Human, [blob])]);
                    }
                    string locator = reference.Locator;
                    string rendered = agent.Engine().Render(context).ToJsonString();
                    Check(reference.Locator == locator, "Rendering rewrote the media locator");
                    return rendered;
                }
            }

            // Exercise actor admission as well as direct rendering. Neither a
            // sandbox refusal nor capability refusal may reach the provider.
            var listener = new TcpListener(IPAddress.Loopback, 0);
            listener.Start();
            try
            {
                string url = "http://127.0.0.1:" + ((IPEndPoint)listener.LocalEndpoint).Port;
                foreach (string model in new[] { "gpt-4o", "unknown-model" })
                {
                    var agent = new Agent(app, new Configuration(Vendor.OpenAI, model, url,
                        sandboxRoot: inside), Path.Combine(inside, "jobs"));
                    using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(5));
                    try
                    {
                        await agent.SubmitAsync([new BlobPart("image/png", new BlobRef(RefKind.Path, outside))])
                            .WaitAsync(deadline.Token);
                        throw new Exception("Outside attachment was admitted to a request");
                    }
                    catch (IOException) when (model == "gpt-4o") { }
                    catch (NotSupportedException) when (model == "unknown-model") { }
                    Check(!listener.Pending() && !agent.History().Events().Any(e => e.Kind == EventKind.RequestSent),
                        "Refused attachment reached transport or RequestSent");
                    var received = agent.History().Events().Single(e => e.Kind == EventKind.MessageReceived);
                    Check(((BlobPart)received.Parts.Single()).Ref.Locator == outside,
                        "Refusal rewrote canonical media history");
                    agent.Destroy();
                }
            }
            finally
            {
                listener.Stop();
            }
        }
        finally
        {
            app.Destroy();
            Directory.Delete(workspace, recursive: true);
        }

        static void RefusePath(Action action, string description)
        {
            try
            {
                action();
            }
            catch (IOException)
            {
                return;
            }
            throw new Exception(description + " escaped the Sandbox");
        }
    }
}
