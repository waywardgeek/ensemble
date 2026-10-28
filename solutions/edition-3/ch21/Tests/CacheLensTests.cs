using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class CacheLensTests
{
    private static void Check(bool condition, string reason)
    {
        if (!condition)
        {
            throw new InvalidOperationException(reason);
        }
    }

    private static byte[] Bytes(string text) => Encoding.UTF8.GetBytes(text);
    private static string Text(byte[] bytes) => Encoding.UTF8.GetString(bytes);

    public static void Confined(string workspace)
    {
        string root = Path.Combine(workspace, "lens-root");
        string outside = Path.Combine(workspace, "lens-outside");
        Directory.CreateDirectory(root);
        Directory.CreateDirectory(outside);
        var log = new StringWriter();
        var app = new Application(log);
        var agent = new Agent(app, new Configuration(Vendor.Anthropic, "unknown", "http://127.0.0.1:1", "dummy", sandboxRoot: root));
        try
        {
            try
            {
                _ = new CacheLens(agent.Engine(), outside);
                throw new InvalidOperationException("Outside cache capture accepted");
            }
            catch (Exception error) when (error is IOException or UnauthorizedAccessException) { }
            Check(agent.Engine().Lens() is null, "refused lens must not attach");
            var lens = new CacheLens(agent.Engine(), "capture");
            lens.Begin("{\"messages\":[]}");
            string directory = Path.Combine(root, "capture");
            Check(File.Exists(Path.Combine(directory, "request.wire.json")), "relative capture inside root");
            Directory.Move(directory, directory + "-held");
            Directory.CreateSymbolicLink(directory, outside);
            lens.Begin("{\"messages\":[1]}");
            Check(!Directory.EnumerateFileSystemEntries(outside).Any(), "swapped capture must not write outside");
            Check(log.ToString().Contains("Cache lens capture failed"), "capture failure remains nonintrusive");
        }
        finally
        {
            app.Destroy();
        }
    }

    public static void ByteViews()
    {
        // Only recognized marker locations are stripped. Both user JSON and
        // text containing the same spelling remain byte-for-byte meaningful.
        const string original = """
        { "model":"unknown", "tools":[{"cache_control":{"type":"ephemeral"},"name":"read","input_schema":{"properties":{"cache_control":{"type":"string"}}}}], "system":[{"type":"text","text":"A\u0042" , "cache_control": {"type":"ephemeral"}}], "messages":[{"role":"assistant","content":[{"type":"tool_use","input":{"cache_control":1.00,"x":9007199254740993},"cache_control":{}},{"cache_control":{},"cache_control":{},"type":"text","text":"cache_control / 雪"}]}], "max_tokens":1024 }
        """;
        const string expected = """
        { "model":"unknown", "tools":[{"name":"read","input_schema":{"properties":{"cache_control":{"type":"string"}}}}], "system":[{"type":"text","text":"A\u0042"}], "messages":[{"role":"assistant","content":[{"type":"tool_use","input":{"cache_control":1.00,"x":9007199254740993}},{"type":"text","text":"cache_control / 雪"}]}], "max_tokens":1024 }
        """;
        byte[] wire = Bytes(original);
        Check(Text(CacheLens.StripMarkers(wire)) == expected, "only marker spans and necessary commas removed");
        Check(Text(wire) == original, "input bytes untouched");
        foreach (string body in new[] { "{}", "{\"cache_control\":{}}", "{\"cache_control\":{},\"cache_control\":{}}" })
        {
            using var parsed = JsonDocument.Parse(CacheLens.StripMarkers(Bytes(body)));
            Check(!parsed.RootElement.EnumerateObject().Any(), "empty/all-marker objects valid");
        }
        const string middle = "{\"a\":1, \"cache_control\":{}, \"b\":2}";
        Check(Text(CacheLens.StripMarkers(Bytes(middle))) == "{\"a\":1, \"b\":2}", "middle marker comma handling");
        const string nested = "{\"messages\":[{\"content\":[{\"content\":[{\"text\":\"x\",\"cache_control\":{}}],\"type\":\"tool_result\"}]}]}";
        Check(!Text(CacheLens.StripMarkers(Bytes(nested))).Contains("cache_control"), "nested result marker removed");
        var capture = CacheLens.Capture(Vendor.Anthropic, wire);
        string prompt = Text(capture.Prompt);
        Check(prompt.StartsWith("{\"tools\":"), "Anthropic tools first");
        Check(prompt.IndexOf("\"system\"", StringComparison.Ordinal) < prompt.IndexOf("\"messages\"", StringComparison.Ordinal), "system before dialogue");
        Check(prompt.Contains("A\\u0042") && prompt.Contains("1.00") && prompt.Contains("9007199254740993"), "raw lexemes preserved");
        Check(Text(capture.Options) == "{\"model\":\"unknown\",\"max_tokens\":1024}", "non-prompt options separated");
        Check(Text(capture.Wire) == original, "exact wire retained including markers");
        var openai = CacheLens.Capture(Vendor.OpenAI, Bytes("{\"messages\":[],\"temperature\":0.00,\"tools\":[]}"));
        Check(Text(openai.Prompt) == "{\"tools\":[],\"messages\":[]}" && Text(openai.Options).Contains("0.00"), "OpenAI order and options");
        var gemini = CacheLens.Capture(Vendor.Gemini, Bytes("{\"contents\":[],\"tools\":[],\"systemInstruction\":{\"parts\":[]},\"generationConfig\":{}}"));
        Check(Text(gemini.Prompt) == "{\"systemInstruction\":{\"parts\":[]},\"tools\":[],\"contents\":[]}", "Gemini prompt order");
    }

    private static HttpListener Listen()
    {
        var port = new TcpListener(IPAddress.Loopback, 0);
        port.Start();
        int number = ((IPEndPoint)port.LocalEndpoint).Port;
        port.Stop();
        var listener = new HttpListener();
        listener.Prefixes.Add($"http://127.0.0.1:{number}/");
        listener.Start();
        return listener;
    }

    private static async Task Reply(HttpListenerContext context, long fresh, long cached)
    {
        string body = JsonSerializer.Serialize(new
        {
            model = "unknown", content = new[] { new { type = "text", text = "done" } },
            usage = new { input_tokens = fresh, cache_read_input_tokens = cached, output_tokens = 2 }
        });
        byte[] data = Bytes(body);
        context.Response.ContentType = "application/json";
        context.Response.ContentLength64 = data.Length;
        await context.Response.OutputStream.WriteAsync(data);
        context.Response.Close();
    }

    private static JsonElement[] Reports(StringWriter log) => log.ToString().Split('\n')
        .Where(line => line.Contains("\"cache_lens\"", StringComparison.Ordinal))
        .Select(line => JsonDocument.Parse(line[line.IndexOf('{')..]).RootElement.Clone()).ToArray();

    public static async Task Transport(string root)
    {
        string directory = Path.Combine(root, "lens-transport");
        var log = new StringWriter();
        var app = new Application(log);
        using var listener = Listen();
        var agent = new Agent(app, new Configuration(Vendor.Anthropic, "unknown", listener.Prefixes.Single(), "local-dummy"), Path.Combine(root, "lens-jobs"));
        var engine = agent.Engine();
        var lens = new CacheLens(engine, directory);
        try
        {
            Check(ReferenceEquals(engine.Lens(), lens) && ReferenceEquals(lens.Parent(), engine), "two-way ownership");
            engine.Render(agent.History().Snapshot().Context);
            Check(!Directory.Exists(directory), "offline rendering does not capture or create files");
            const string a = "{\"model\":\"unknown\",\"system\":\"steady\",\"messages\":[{\"role\":\"user\",\"content\":\"one\"}]}";
            const string b = "{\"model\":\"unknown\",\"system\":\"steady\",\"messages\":[{\"role\":\"user\",\"content\":\"one plus\"}]}";
            var first = engine.SendAsync(a);
            var firstRequest = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            var second = engine.SendAsync(b);
            var secondRequest = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            using (var reader = new StreamReader(firstRequest.Request.InputStream))
            {
                Check(await reader.ReadToEndAsync() == a, "first outgoing body unchanged");
            }
            using (var reader = new StreamReader(secondRequest.Request.InputStream))
            {
                Check(await reader.ReadToEndAsync() == b, "second outgoing body unchanged");
            }
            Check(File.ReadAllText(Path.Combine(directory, "request.prev.wire.json")) == a, "prior wire rotates");
            Check(File.ReadAllText(Path.Combine(directory, "request.wire.json")) == b, "current exact wire");
            Check(File.ReadAllBytes(Path.Combine(directory, "request.prev.body")).SequenceEqual(CacheLens.Capture(Vendor.Anthropic, Bytes(a)).Prompt), "prior prompt rotates");
            // Finish in reverse order. A shared mutable prediction would make
            // the cold first response incorrectly inherit the second prefix.
            await Reply(secondRequest, 1, 9);
            await second.WaitAsync(TimeSpan.FromSeconds(10));
            await Reply(firstRequest, 10, 0);
            await first.WaitAsync(TimeSpan.FromSeconds(10));
            var reports = Reports(log);
            Check(reports.Length == 2 && reports[0].GetProperty("cache_lens").GetInt64() == 2
                && reports[1].GetProperty("cache_lens").GetInt64() == 1, "response/prediction correlation");
            Check(reports[1].GetProperty("shared_bytes").GetInt32() == 0, "cold prediction not overwritten");
            Check(reports[0].GetProperty("shared_bytes").GetInt32() > 0
                && reports[0].GetProperty("actual_hit_rate").GetDouble() == 0.9, "warm byte prediction and provider ratio");
            using var cancel = new CancellationTokenSource();
            var cancelled = engine.SendAsync(b, cancel.Token);
            var held = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            cancel.Cancel();
            try
            {
                await cancelled;
                throw new InvalidOperationException("Cancellation was ignored");
            }
            catch (OperationCanceledException) { }
            held.Response.Abort();
            Check(Reports(log).Last().GetProperty("actual_hit_rate").ValueKind == JsonValueKind.Null, "cancelled response has unknown usage, not a hit");
        }
        finally
        {
            await agent.ShutdownAsync();
            app.Destroy();
        }
        Check(engine.Lens() is null, "Engine destruction removes child");
        lens.Destroy();
        try
        {
            lens.Parent();
            throw new InvalidOperationException("Lens retained destroyed parent");
        }
        catch (ObjectDisposedException) { }
    }

    public static async Task Failure(string root)
    {
        string blocked = Path.Combine(root, "lens-blocked");
        File.WriteAllText(blocked, "owned obstacle");
        var log = new StringWriter();
        var app = new Application(log);
        using var listener = Listen();
        var agent = new Agent(app, new Configuration(Vendor.Anthropic, "unknown", listener.Prefixes.Single(), "dummy"), Path.Combine(root, "lens-failure-jobs"));
        var engine = agent.Engine();
        var lens = new CacheLens(engine, blocked);
        try
        {
            try
            {
                _ = new CacheLens(engine, blocked);
                throw new InvalidOperationException("Duplicate lens accepted");
            }
            catch (InvalidOperationException ex) when (ex.Message == "Invalid cache lens registration") { }
            Check(ReferenceEquals(engine.Lens(), lens), "refusal preserves owner");
            var send = engine.SendAsync("{\"messages\":[]}");
            var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            await Reply(request, 4, 0);
            var result = await send.WaitAsync(TimeSpan.FromSeconds(10));
            Check(result.Events.Any(fact => fact.Kind == EventKind.ResponseEnded), "capture failure does not block inference");
            Check(log.ToString().Contains("Cache lens capture failed (IOException)"), "safe capture failure diagnostic");
            Check(File.ReadAllText(blocked) == "owned obstacle", "blocked path preserved");
            using (var cancel = new CancellationTokenSource())
            {
                cancel.Cancel();
                int reports = Reports(log).Length;
                try
                {
                    await engine.SendAsync("{}", cancel.Token);
                    throw new InvalidOperationException("Pre-cancelled send admitted");
                }
                catch (OperationCanceledException) { }
                Check(Reports(log).Length == reports, "pre-cancellation creates no lens report");
            }
            // Deliberately malformed caller input goes unchanged to the boundary;
            // a diagnostic parser must not secretly replace transport admission.
            var malformed = engine.SendAsync("{bad-json");
            var malformedRequest = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            using (var reader = new StreamReader(malformedRequest.Request.InputStream))
            {
                Check(await reader.ReadToEndAsync() == "{bad-json", "invalid caller body unchanged");
            }
            malformedRequest.Response.StatusCode = 400;
            malformedRequest.Response.Close();
            await malformed.WaitAsync(TimeSpan.FromSeconds(10));
            Check(log.ToString().Contains("Cache lens capture skipped (invalid JSON)."), "invalid JSON diagnostic is safe");
            lens.Destroy();
            Check(engine.Lens() is null, "explicit child destruction detaches parent");
            var replacement = new CacheLens(engine, Path.Combine(root, "replacement-lens"));
            replacement.Destroy();
        }
        finally
        {
            await agent.ShutdownAsync();
            app.Destroy();
        }
    }
}
