using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

// A separate assembly exercises only public APIs. The tiny HTTP peer replaces
// the provider, not the Agent, renderer, History, lens, or meter. Its reported
// cache counts demonstrate accounting; they do not prove real cache behavior.
if (args.Length != 1 || Directory.Exists(args[0]) || File.Exists(args[0]))
{
    Console.Error.WriteLine("Supply one new owned exercise directory.");
    return 2;
}
string directory = Path.GetFullPath(args[0]);
Directory.CreateDirectory(directory);
string captures = Path.Combine(directory, "private-captures");
string save = Path.Combine(directory, "save.json");

// Reserve a free loopback port. No existing application or shared port is used.
var reservation = new TcpListener(IPAddress.Loopback, 0);
reservation.Start();
int port = ((IPEndPoint)reservation.LocalEndpoint).Port;
reservation.Stop();
using var listener = new HttpListener();
listener.Prefixes.Add($"http://127.0.0.1:{port}/");
listener.Start();
var app = new Application(Console.Error);
try
{
    var configuration = new Configuration(Vendor.OpenAI, "gpt-4o-mini",
        listener.Prefixes.Single(), "dummy-local", disableStreaming: true,
        promptCacheKey: "public-consumer");
    var agent = new Agent(app, configuration, name: "first");
    var engine = agent.Engine();
    var lens = new CacheLens(engine, captures);
    Require(ReferenceEquals(lens.Parent(), engine) && ReferenceEquals(engine.Lens(), lens),
        "Lens attachment did not establish both ends.");

    // Rendering is pure preparation, not a send and not a billable observation.
    _ = engine.Render(agent.History().View());
    Require(!Directory.Exists(captures) && engine.Meter().Latest is null,
        "Offline rendering advanced diagnostics or metering.");
    byte[] firstWire = await Turn("first", "default");
    var first = engine.Meter();
    Require(first.Latest!.Counts.Input == 60 && first.Latest.Counts.CacheRead == 40
        && first.Totals.AllInput() == 100 && first.HitRate() == 0.4
        && first.EstimatedUsd() == 0.000018m, "Incorrect first response accounting.");
    byte[] secondWire = await Turn("second", "unpriced-fixture-tier");
    var second = engine.Meter();
    Require(second.Totals.AllInput() == 200 && second.Totals.Output == 20
        && second.EstimatedUsd() is null && first.Totals.AllInput() == 100,
        "Unknown price lost counts or mutated an earlier snapshot.");
    Require(File.ReadAllBytes(Path.Combine(captures, "request.prev.wire.json")).SequenceEqual(firstWire)
        && File.ReadAllBytes(Path.Combine(captures, "request.wire.json")).SequenceEqual(secondWire),
        "Lens captures differ from the actual sent bytes.");
    string prediction = File.ReadAllText(Path.Combine(captures, "request.prediction.json"));
    _ = engine.Render(agent.History().View());
    Require(File.ReadAllText(Path.Combine(captures, "request.prediction.json")) == prediction,
        "Offline rendering rotated actual-send captures.");

    // History preserves the conversation; it does not seed a new process meter.
    agent.History().Save(new PersistenceOptions(save, RetainLog: 0));
    var restored = new Agent(app, configuration, name: "restored");
    Require(restored.History().LoadSave(save), "Snapshot-only restore failed.");
    Require(restored.Engine().Render(restored.History().View()).ToJsonString().Contains("ANSWER_second")
        && restored.Engine().Meter().Latest is null
        && restored.Engine().Meter().EstimatedUsd() == 0m, "Restore billed old work again.");

    lens.Destroy();
    Require(engine.Lens() is null, "Lens destruction left a parent link.");
    lens.Destroy(); // Explicit logical destruction is idempotent.
    await agent.ShutdownAsync();
    await restored.ShutdownAsync();
    Console.WriteLine("PASS public cache consumer: two real turns, exact captures, identity-aware estimates, immutable snapshots, restart isolation and lens detachment.");

    async Task<byte[]> Turn(string prompt, string tier)
    {
        Task<string> answer = agent.SendAsync(prompt);
        var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
        using var received = new MemoryStream();
        await request.Request.InputStream.CopyToAsync(received);
        byte[] wire = received.ToArray();
        using var body = JsonDocument.Parse(wire);
        Require(body.RootElement.GetProperty("prompt_cache_key").GetString() == "public-consumer",
            "The configured routing key was absent from the real request.");
        byte[] reply = Encoding.UTF8.GetBytes(JsonSerializer.Serialize(new
        {
            model = "gpt-4o-mini-2024-07-18", service_tier = tier,
            choices = new[] { new { message = new { role = "assistant", content = "ANSWER_" + prompt }, finish_reason = "stop" } },
            usage = new { prompt_tokens = 100, prompt_tokens_details = new { cached_tokens = 40 }, completion_tokens = 10 }
        }));
        request.Response.ContentType = "application/json";
        request.Response.ContentLength64 = reply.Length;
        await request.Response.OutputStream.WriteAsync(reply);
        request.Response.Close();
        Require((await answer.WaitAsync(TimeSpan.FromSeconds(10))).Contains("ANSWER_" + prompt),
            "The real Agent turn did not return its answer.");
        return wire;
    }
}
finally
{
    listener.Stop();
    app.Destroy();
}
return 0;

static void Require(bool condition, string message)
{
    if (!condition)
    {
        throw new InvalidOperationException(message);
    }
}
