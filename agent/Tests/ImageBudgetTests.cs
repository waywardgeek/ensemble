using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class ImageBudgetTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition) throw new Exception(message);
    }

    private static BlobPart Inline(byte[] bytes) => new("image/png",
        new BlobRef(RefKind.Uri, "data:image/png;base64," + Convert.ToBase64String(bytes)));

    private static int Images(JsonNode? node)
    {
        if (node is JsonArray array) return array.Sum(Images);
        if (node is not JsonObject value) return 0;
        string? type = value["type"]?.GetValue<string>();
        if (type is "image" or "input_image" or "image_url" || value["inlineData"] is not null)
            return 1;
        return value.Sum(property => Images(property.Value));
    }

    private static void Refuse(Action action)
    {
        try { action(); }
        catch (ImageBudgetExceededException)
        {
            return;
        }
        throw new Exception("Expected cumulative image-budget refusal.");
    }

    public static async Task Admission()
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        var app = new Application(TextWriter.Null);
        var tiny = Inline(Convert.FromBase64String(
            "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII="));
        var config = new Configuration(Vendor.Anthropic, "claude-3-haiku-20240307",
            "http://127.0.0.1:1", "fake-only");
        try
        {
            var agent = new Agent(app, config);
            await agent.PostAsync(new AgentInput(InputKind.Pause)).WaitAsync(deadline.Token);
            for (int i = 0; i < Engine.MaximumRequestImages; i++)
                await agent.PostAsync(new AgentInput(InputKind.Message, "accepted", [tiny])).WaitAsync(deadline.Token);
            int before = agent.History().Events().Length;
            try
            {
                await agent.PostAsync(new AgentInput(InputKind.Message, "REFUSED", [tiny])).WaitAsync(deadline.Token);
                throw new Exception("Overflow input was admitted.");
            }
            catch (ImageBudgetExceededException) { }
            Check(agent.History().Events().Length == before, "Refused input changed History.");
            Check(Images(agent.Engine().Render(agent.History().View())) == 32, "Prior images were lost.");
            await agent.PostAsync(new AgentInput(InputKind.Message, "Text remains admissible")).WaitAsync(deadline.Token);

            var queued = new Agent(app, config);
            await queued.PostAsync(new AgentInput(InputKind.Pause)).WaitAsync(deadline.Token);
            var pending = queued.SubmitAsync(Enumerable.Repeat<Part>(tiny, 32));
            try
            {
                await queued.PostAsync(new AgentInput(InputKind.Hint, "queued overflow", [tiny])).WaitAsync(deadline.Token);
                throw new Exception("Queued image reservation was ignored.");
            }
            catch (ImageBudgetExceededException) { }
            await queued.ShutdownAsync().WaitAsync(deadline.Token);
            try { await pending; } catch (ObjectDisposedException) { }

            var legacy = new Agent(app, config);
            legacy.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: Enumerable.Repeat<Part>(tiny, 33)));
            string original = JsonCodec.WriteEvent(legacy.History().Events()[0]);
            try
            {
                await legacy.SendAsync("hello", deadline.Token).WaitAsync(deadline.Token);
                throw new Exception("Over-limit restored context was sent.");
            }
            catch (ImageBudgetExceededException) { }
            var facts = legacy.History().Events();
            Check(!facts.Any(fact => fact.Kind == EventKind.RequestSent), "Refusal reached transport.");
            Check(facts.Count(fact => fact.Kind == EventKind.ErrorOccurred) == 1,
                "Local refusal did not publish exactly one narratable error.");
            Check(JsonCodec.WriteEvent(facts[0]) == original, "Recovery rewrote canonical images.");
        }
        finally { app.Destroy(); }
    }

    public static void RequestLimits(string temp)
    {
        string home = Path.Combine(temp, "request-image-budget-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(home);
        // Tiny real PNG for count tests. Large raw payloads below deliberately
        // test byte accounting only; raster validation belongs to PrepareImage.
        var tiny = Inline(Convert.FromBase64String(
            "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII="));
        var large = Inline(new byte[Engine.MaximumImageBytes]);
        string path = Path.Combine(home, "bounded-image-bytes.bin");
        File.WriteAllBytes(path, new byte[Engine.MaximumImageBytes]);
        var file = new BlobPart("image/png", new BlobRef(RefKind.Path, path));
        var app = new Application(TextWriter.Null);
        try
        {
            foreach (var sample in new[]
            {
                (Vendor.OpenAI, "gpt-6-astra", Surface.Responses),
                (Vendor.OpenAI, "gpt-4o", Surface.ChatCompletions),
                (Vendor.Anthropic, "claude-3-haiku-20240307", Surface.Messages),
                (Vendor.Gemini, "gemini-2.5-flash", Surface.GenerateContent),
                (Vendor.Gemini, "gemini-3-flash-preview", Surface.GenerateContent)
            })
            {
                var config = new Configuration(sample.Item1, sample.Item2,
                    "http://127.0.0.1:1", surface: sample.Item3);
                var agent = new Agent(app, config, Path.Combine(home, sample.Item2));
                try
                {
                    var entries = Enumerable.Range(1, Engine.MaximumRequestImages)
                        .Select(index => new Entry(index, Actor.Human, [tiny])).ToArray();
                    var exactCount = new Context(entries);
                    Check(Images(agent.Engine().Render(exactCount)) == Engine.MaximumRequestImages,
                        "Exact count boundary lost images.");
                    Refuse(() => agent.Engine().Render(new Context(entries, ephemeral: [tiny])));
                    agent.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                        parts: Enumerable.Repeat<Part>(tiny, Engine.MaximumRequestImages + 1)));
                    string history = JsonCodec.WriteEvent(agent.History().Events()[0]);
                    Refuse(() => agent.Engine().Render(agent.History().View()));
                    Check(JsonCodec.WriteEvent(agent.History().Events()[0]) == history,
                        "Budget refusal rewrote canonical History.");

                    int fullImages = checked((int)(Engine.MaximumRequestImageBytes / Engine.MaximumImageBytes));
                    var byteEntries = Enumerable.Range(1, fullImages)
                        .Select(index => new Entry(index, Actor.Human, [index % 2 == 0 ? file : large])).ToArray();
                    Check(Images(agent.Engine().Render(new Context(byteEntries))) == fullImages,
                        "Exact byte boundary lost file or inline images.");
                    Refuse(() => agent.Engine().Render(new Context(byteEntries, ephemeral: [tiny])));

                    bool toolImages = sample.Item1 == Vendor.Anthropic || sample.Item3 == Surface.Responses
                        || sample.Item2 == "gemini-3-flash-preview";
                    if (toolImages)
                    {
                        using var empty = JsonDocument.Parse("{}");
                        var call = new ToolCallPart("image-budget-call", config.Target(),
                            "inspect_image", empty.RootElement.Clone());
                        var result = new ToolResultPart("image-budget-call", [tiny, tiny]);
                        var mixed = entries.Take(Engine.MaximumRequestImages - 1).Concat([
                            new Entry(100, Actor.Agent, [call]), new Entry(101, Actor.Tool, [result])]);
                        Refuse(() => agent.Engine().Render(new Context(mixed)));
                    }

                    // Refusals and parallel renders share no mutable budget.
                    Parallel.For(0, 4, _ => Check(
                        Images(agent.Engine().Render(exactCount)) == Engine.MaximumRequestImages,
                        "A prior/concurrent render consumed this request's allowance."));
                }
                finally { agent.Destroy(); }
            }
        }
        finally
        {
            app.Destroy();
            Directory.Delete(home, recursive: true);
        }
    }
}
