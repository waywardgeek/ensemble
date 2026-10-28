using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Text.Json;
using Ensemble;

// These tests use real owners and a working HTTP server at the external boundary.
// No internal method calls are mocked or scripted.
int passed = 0;
await Check("complete two-way relationship operations", () =>
{
    var root = new Ensemble.Ensemble(TextWriter.Null);
    var other = new Ensemble.Ensemble(TextWriter.Null);
    var agent = new Agent(root, new Configuration("http://127.0.0.1", "dummy", "fake"));
    Engine engine = agent.Engine();
    Throws(() => other.RemoveAgent(agent));
    Require(agent.Parent() == root && root.Agents().Length == 1, "refusal changed ownership");
    root.RemoveAgent(agent);
    Require(agent.parent is null && root.Agents().IsEmpty, "remove left reverse link");
    other.AddAgent(agent);
    Require(agent.Parent() == other && other.Agents().Single() == agent, "add incomplete");
    agent.RemoveEngine(engine);
    Require(engine.parent is null, "engine back-reference remains");
    Throws(() => agent.Engine());
    agent.AddEngine(engine);
    Require(engine.Parent() == agent && agent.Engine() == engine, "engine add incomplete");
    Throws(() => new Engine(agent));
    Require(agent.Engine() == engine, "failed construction replaced engine");
    other.Destroy();
    other.Destroy();
    Require(agent.parent is null && engine.parent is null, "destroy left ownership");
    Require(agent.destroyed && engine.destroyed && other.Agents().IsEmpty, "cascade incomplete");
    Throws(() => root.AddAgent(agent));
    root.Destroy();
    return Task.CompletedTask;
});

await Check("successful pairs, all text blocks and cumulative receipts", async () =>
{
    using var fake = new Provider();
    var diagnostics = new StringWriter();
    var root = new Ensemble.Ensemble(diagnostics);
    try
    {
        var agent = new Agent(root, new Configuration(fake.Url, "test-key", "test-model"));
        Task first = fake.Reply(Response("hello", 7, 3), request =>
        {
            Require(request.Headers["x-api-key"] == "test-key", "wrong key header");
            Require(request.Headers["anthropic-version"] == "2023-06-01", "wrong version");
        });
        Require(await agent.AskAsync("one") == "hello!", "lost text block");
        await first;
        Require(agent.Messages().Length == 2, "first pair missing");
        Task second = fake.Reply(Response("again", 8, 4));
        await agent.AskAsync("two");
        await second;
        using JsonDocument sent = JsonDocument.Parse(fake.Bodies[1]);
        JsonElement body = sent.RootElement;
        Require(body.GetProperty("messages").GetArrayLength() == 3, "history not resent");
        Require(body.GetProperty("model").GetString() == "test-model", "wrong model");
        Require(body.GetProperty("max_tokens").GetInt32() == 1024, "wrong maximum");
        Require(agent.Engine().Usage() == new Usage(15, 7), "wrong cumulative usage");
        Require(diagnostics.ToString() == "", "unexpected diagnostics");
    }
    finally
    {
        root.Destroy();
    }
});

await Check("failures do not publish pairs or leak remote data", async () =>
{
    using var fake = new Provider();
    var diagnostics = new StringWriter();
    var root = new Ensemble.Ensemble(diagnostics);
    try
    {
        var agent = new Agent(root, new Configuration(fake.Url, "secret-key", "test-model"));
        string[] badBodies =
        [
            "secret-key remote body",
            "{\"content\":[{\"type\":\"text\",\"text\":\"unaccounted\"}]}",
            "{\"usage\":{\"input_tokens\":-1,\"output_tokens\":2},\"content\":[]}",
            "{\"usage\":{\"input_tokens\":1.5,\"output_tokens\":2},\"content\":[]}"
        ];
        foreach (string body in badBodies)
        {
            Task reply = fake.Reply(body);
            await ThrowsAsync(() => agent.AskAsync("not committed"));
            await reply;
            Require(agent.Messages().IsEmpty, "failed request published history");
            Require(agent.Engine().Usage() == new Usage(0, 0), "invalid usage counted");
        }
        Task refused = fake.Reply("secret-key private body", status: 401);
        await ThrowsAsync(() => agent.AskAsync("refused"));
        await refused;
        Require(fake.Bodies.Count == 5, "request retried");
        Require(!diagnostics.ToString().Contains("secret-key"), "secret leaked");
        Task empty = fake.Reply("{\"usage\":{\"input_tokens\":9,\"output_tokens\":1},\"content\":[]}");
        await ThrowsAsync(() => agent.AskAsync("empty"));
        await empty;
        Require(agent.Engine().Usage() == new Usage(9, 1), "lost measured failed-answer usage");
        Require(agent.Messages().IsEmpty, "empty answer published");
        Task recovery = fake.Reply(Response("recovered", 1, 2));
        await agent.AskAsync("new question");
        await recovery;
        using JsonDocument sent = JsonDocument.Parse(fake.Bodies[^1]);
        Require(sent.RootElement.GetProperty("messages").GetArrayLength() == 1, "failed turns resent");
        int before = fake.Bodies.Count;
        await ThrowsAsync(() => agent.AskAsync("  "));
        Require(fake.Bodies.Count == before, "blank question sent");
    }
    finally
    {
        root.Destroy();
    }
});

await Check("caller can cancel an admitted request without publishing a pair", async () =>
{
    using var fake = new Provider();
    var root = new Ensemble.Ensemble(TextWriter.Null);
    try
    {
        var agent = new Agent(root, new Configuration(fake.Url, "dummy", "fake"));
        using var cancellation = new CancellationTokenSource();
        Task<string> exchange = agent.AskAsync("wait", cancellation.Token);
        // The request must really reach the server before cancellation is tested.
        HttpListenerContext request = await fake.Listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
        cancellation.Cancel();
        bool cancelled = false;
        try
        {
            await exchange;
        }
        catch (OperationCanceledException)
        {
            cancelled = true;
        }
        request.Response.Abort();
        Require(cancelled && agent.Messages().IsEmpty, "cancellation not preserved");
    }
    finally
    {
        root.Destroy();
    }
});
Console.WriteLine($"{passed} test groups passed.");

async Task Check(string name, Func<Task> run)
{
    await run();
    passed++;
    Console.WriteLine($"PASS {name}");
}

static void Require(bool condition, string message)
{
    if (!condition)
    {
        throw new Exception(message);
    }
}

static void Throws(Action action)
{
    try
    {
        action();
    }
    catch (InvalidOperationException)
    {
        return;
    }
    throw new Exception("Expected relationship refusal.");
}

static async Task ThrowsAsync(Func<Task> action)
{
    try
    {
        await action();
    }
    catch (Exception error) when (error is InvalidOperationException or ArgumentException)
    {
        Require(!error.ToString().Contains("secret-key"), "exception leaked secret");
        return;
    }
    throw new Exception("Expected failed exchange.");
}

static string Response(string text, int input, int output)
{
    return JsonSerializer.Serialize(new
    {
        content = new object[]
        {
            new { type = "other" },
            new { type = "text", text },
            new { type = "text", text = "!" }
        },
        usage = new { input_tokens = input, output_tokens = output }
    });
}

// A small protocol fake: it receives and records real requests and returns
// caller-supplied wire responses, including errors. It does not replace Engine.
sealed class Provider : IDisposable
{
    public readonly HttpListener Listener = new();
    public readonly List<string> Bodies = new();
    public readonly string Url;

    public Provider()
    {
        var reservation = new TcpListener(IPAddress.Loopback, 0);
        reservation.Start();
        int port = ((IPEndPoint)reservation.LocalEndpoint).Port;
        reservation.Stop();
        Url = $"http://127.0.0.1:{port}";
        Listener.Prefixes.Add(Url + "/");
        Listener.Start();
    }

    public async Task Reply(string body, Action<HttpListenerRequest>? inspect = null, int status = 200)
    {
        HttpListenerContext context = await Listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
        using var reader = new StreamReader(context.Request.InputStream);
        Bodies.Add(await reader.ReadToEndAsync());
        if (context.Request.HttpMethod != "POST" || context.Request.RawUrl != "/v1/messages")
        {
            throw new Exception("Wrong provider route.");
        }
        inspect?.Invoke(context.Request);
        byte[] bytes = Encoding.UTF8.GetBytes(body);
        context.Response.StatusCode = status;
        context.Response.ContentType = "application/json";
        context.Response.ContentLength64 = bytes.Length;
        await context.Response.OutputStream.WriteAsync(bytes);
        context.Response.Close();
    }

    public void Dispose()
    {
        Listener.Close();
    }
}
