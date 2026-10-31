using System.Diagnostics;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

// Application composition only: the library knows nothing about these counters.
// One consumer reads the bounded progress queue. Its metrics describe delivered
// observations, not lossless billing or an alternative completion mechanism.
string address = Environment.GetEnvironmentVariable("LLM_BASE_URL") ?? "";
if (!Uri.TryCreate(address, UriKind.Absolute, out var endpoint) || !endpoint.IsLoopback)
{
    Console.Error.WriteLine("This offline exercise requires an owned loopback LLM_BASE_URL.");
    return 2;
}
if (!Enum.TryParse<Vendor>(Environment.GetEnvironmentVariable("LLM_VENDOR") ?? "Anthropic", true, out var vendor))
{
    Console.Error.WriteLine("Unsupported LLM_VENDOR");
    return 2;
}
string model = Environment.GetEnvironmentVariable("LLM_MODEL") ?? "local";
string key = Environment.GetEnvironmentVariable("LLM_API_KEY") ?? "dummy";
bool disabled = Environment.GetEnvironmentVariable("CH07_NO_STREAM") == "1";
string prompt = args.Length == 0 ? "what does the configuration say?" : string.Join(' ', args);
using var log = new StreamWriter(Environment.GetEnvironmentVariable("CH07_LOG") ?? "stream.jsonl")
{
    AutoFlush = true
};
var app = new Application(Console.Error);
Task? observing = null;
var clock = new Stopwatch();
double? firstDelta = null;
var deltas = new Dictionary<string, int> { ["text"] = 0, ["thinking"] = 0, ["tool_call"] = 0 };
var bytes = new Dictionary<string, int> { ["text"] = 0, ["thinking"] = 0, ["tool_call"] = 0 };
var parts = new Dictionary<string, int> { ["text"] = 0, ["thinking"] = 0, ["tool_call"] = 0 };
var identities = new HashSet<(long Response, int Part)>();
int transitions = 0;
int steps = 0;
try
{
    var config = new Configuration(vendor, model, endpoint.ToString(), key, disableStreaming: disabled);
    var agent = new Agent(app, config);
    agent.InstallTools();
    string? historyPath = Environment.GetEnvironmentVariable("CH07_AGENT_LOG");
    if (!string.IsNullOrEmpty(historyPath))
    {
        agent.History().StartLog(historyPath);
    }
    clock.Start();
    observing = Observe(agent);
    string answer = await agent.SendAsync(prompt);
    double elapsed = clock.Elapsed.TotalMilliseconds;
    var usage = JsonCodec.WriteUsage(agent.History().View().Usage);
    // Reliable completion governs lifetime. Destruction closes observations;
    // the observer drains queued values before the aggregate is read here.
    app.Destroy();
    await observing;
    bool streaming = Engine.LookupModel(vendor, model, out var features) && features.Stream && !disabled;
    var report = new JsonObject
    {
        ["observation"] = "stream_stats",
        ["deltas"] = JsonSerializer.SerializeToNode(deltas),
        ["bytes"] = JsonSerializer.SerializeToNode(bytes),
        ["parts"] = JsonSerializer.SerializeToNode(parts),
        ["state_transitions"] = transitions,
        ["steps"] = steps,
        // Zero with has_delta:false means no transient chunk was delivered,
        // not an instantaneous first byte from a plain response.
        ["first_delta_ms"] = firstDelta ?? 0,
        ["has_delta"] = firstDelta.HasValue,
        ["elapsed_ms"] = elapsed,
        ["streaming"] = streaming,
        ["assistant"] = answer,
        ["usage"] = usage
    };
    string line = report.ToJsonString();
    log.WriteLine(line);
    Console.WriteLine(line);
    Console.Out.Flush();
    return 0;
}
catch (Exception error) when (error is ArgumentException or InvalidOperationException or IOException
    or JsonException or NotSupportedException)
{
    Console.Error.WriteLine(error.Message);
    return 1;
}
finally
{
    try
    {
        app.Destroy();
    }
    finally
    {
        if (observing is not null)
        {
            await observing;
        }
    }
}

async Task Observe(Agent agent)
{
    try
    {
        while (true)
        {
            var value = await agent.WaitAsync(_ => true);
            if (value.Delta is Chunk chunk)
            {
                firstDelta ??= clock.Elapsed.TotalMilliseconds;
                string kind = chunk.Kind switch
                {
                    ChunkKind.Reply => "text",
                    ChunkKind.Thinking => "thinking",
                    _ => "tool_call"
                };
                string display = chunk.Kind == ChunkKind.ToolCall
                    ? chunk.ToolName + chunk.ArgsText : chunk.Text;
                deltas[kind]++;
                bytes[kind] += Encoding.UTF8.GetByteCount(display);
            }
            else if (value.Kind == ObservationKind.PartFinal && value.PartID is int id)
            {
                // Part ID zero is valid and may recur in a later response.
                // Exclude human/job parts, which have no parser-supplied ID.
                if (!identities.Add((value.ResponseID, id)))
                {
                    throw new InvalidDataException("A model part was finalized twice.");
                }
                string kind = value.Part switch
                {
                    TextPart => "text",
                    ToolCallPart => "tool_call",
                    _ => "thinking"
                };
                parts[kind]++;
            }
            else if (value.Kind == ObservationKind.StateChanged)
            {
                transitions++;
                if (value.To == Phase.InFlight)
                {
                    steps++;
                }
            }
        }
    }
    catch (ObjectDisposedException)
    {
        // End of the drained queue, not a failed request.
    }
}
