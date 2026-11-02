using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    public static async Task ImageAdmissionRecovery(string temp)
    {
        string home = Path.Combine(temp, "image-admission");
        string skills = Path.Combine(home, "skills");
        Directory.CreateDirectory(Path.Combine(skills, "root"));
        File.WriteAllText(Path.Combine(skills, "root", "SKILL.md"),
            "---\nname: root\ndescription: fixture\ntype: primary\ntools: inspect_image micro_handoff\n---\nFixture");
        using var bitmap = new SkiaSharp.SKBitmap(8, 8);
        bitmap.Erase(SkiaSharp.SKColors.Blue);
        using var image = SkiaSharp.SKImage.FromBitmap(bitmap);
        using var encoded = image.Encode(SkiaSharp.SKEncodedImageFormat.Png, 100);
        byte[] bytes = encoded.ToArray();
        File.WriteAllBytes(Path.Combine(home, "tiny.png"), bytes);
        var tiny = new BlobPart("image/png", new BlobRef(RefKind.Uri,
            "data:image/png;base64," + Convert.ToBase64String(bytes)));
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(30));
        var app = new Application(TextWriter.Null);
        Task? server = null;
        try
        {
            var config = new Configuration(Vendor.Anthropic, "claude-3-haiku-20240307",
                $"http://127.0.0.1:{((IPEndPoint)listener.LocalEndpoint).Port}", "fake-only",
                sandboxRoot: home, disableStreaming: true);
            var agent = new Agent(app, config, Path.Combine(home, "jobs"),
                skillSpec: new SkillSpec(skills, "root"));
            agent.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: Enumerable.Repeat<Part>(tiny, 32)));
            server = Task.Run(async () =>
            {
                for (int round = 0; round < 4; round++)
                {
                    using var client = await listener.AcceptTcpClientAsync(deadline.Token);
                    using var stream = client.GetStream();
                    string body = await ReadRequest(stream, deadline.Token);
                    using var json = JsonDocument.Parse(body);
                    int Count(JsonElement value)
                    {
                        if (value.ValueKind == JsonValueKind.Array) return value.EnumerateArray().Sum(Count);
                        if (value.ValueKind != JsonValueKind.Object) return 0;
                        if (value.TryGetProperty("type", out var kind) && kind.GetString() == "image") return 1;
                        return value.EnumerateObject().Sum(property => Count(property.Value));
                    }
                    Check(Count(json.RootElement) == (round < 2 ? 32 : round == 2 ? 0 : 1),
                        "Recovery request " + round + " has " + Count(json.RootElement) + " images.");
                    if (round == 1)
                        Check(body.Contains("Image delivery refused; no new images were delivered."),
                            "Model was not told its image delivery was refused: " + string.Join(" | ", json.RootElement.GetProperty("messages").EnumerateArray().SelectMany(message => message.GetProperty("content").EnumerateArray()).Where(part => part.GetProperty("type").GetString() == "tool_result").Select(part => part.ToString())));
                    if (round == 2)
                        Check(body.Contains("[Completed micro_handoff checkpoint]"), "Checkpoint was not applied.");
                    if (round == 3)
                    {
                        await Reply(stream, "Recovered and inspected the image.", deadline.Token);
                        continue;
                    }
                    string name = round == 1 ? "micro_handoff" : "inspect_image";
                    object input = round == 1 ? new { text = "Retry tiny.png after this checkpoint." }
                        : new { path = "tiny.png" };
                    await ReplyJson(stream, new
                    {
                        content = new object[]
                        {
                            new { type = "text", text = "I will checkpoint or inspect the fixture as appropriate." },
                            new { type = "tool_use", id = "image-recovery-" + round, name, input }
                        },
                        usage = new { input_tokens = 1, output_tokens = 1 }
                    }, deadline.Token);
                }
            });
            string answer;
            try
            {
                answer = await agent.SendAsync("Inspect tiny.png; checkpoint if needed.", deadline.Token)
                    .WaitAsync(deadline.Token);
            }
            catch when (server.IsFaulted)
            {
                await server;
                throw;
            }
            await server.WaitAsync(deadline.Token);
            Check(answer == "Recovered and inspected the image.", "Recovery did not finish.");
            var facts = agent.History().Events();
            Check(facts.Count(fact => fact.Kind == EventKind.HandoffRecorded) == 1, "Wrong handoff count.");
            Check(facts.Any(fact => fact.Kind == EventKind.ToolReturned && fact.CallID == "image-recovery-0"
                && fact.IsError && !fact.Parts.OfType<BlobPart>().Any()), "Refused pixels entered History.");
            Check(facts.Any(fact => fact.Kind == EventKind.MessageReceived && fact.Parts.OfType<BlobPart>().Count() == 32),
                "Checkpoint deleted canonical image evidence.");
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
