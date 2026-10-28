using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

// This assembly sees only the public library. No InternalsVisibleTo grant,
// copied implementation, new dispatch loop, or reflection bridges the boundary.
var address = Environment.GetEnvironmentVariable("LLM_BASE_URL");
if (!Uri.TryCreate(address, UriKind.Absolute, out var endpoint) || !endpoint.IsLoopback)
{
    Console.Error.WriteLine("Set LLM_BASE_URL to the owned loopback fake for this exercise.");
    return 2;
}
if (!Enum.TryParse<Vendor>(Environment.GetEnvironmentVariable("LLM_VENDOR") ?? "Anthropic", true, out var vendor))
{
    Console.Error.WriteLine("Unsupported LLM_VENDOR");
    return 2;
}
var model = Environment.GetEnvironmentVariable("LLM_MODEL") ?? "local";
var key = Environment.GetEnvironmentVariable("LLM_API_KEY") ?? "dummy";
var app = new Application(Console.Error);
try
{
    var agent = new Agent(app, new Configuration(vendor, model, endpoint.ToString(), key));
    agent.InstallTools();
    var schema = JsonSerializer.Deserialize<JsonElement>("""
        {"type":"object","properties":{"text":{"type":"string"}},"required":["text"],"additionalProperties":false}
        """);
    _ = new Tool(agent, "shout", "Return the supplied text in uppercase.", schema,
        (job, arguments, cancellation) =>
        {
            cancellation.ThrowIfCancellationRequested();
            // One real ownership path reaches the Ensemble-owned logger.
            job.Parent().Parent().Parent().Log("shout called");
            return Task.FromResult(arguments.GetProperty("text").GetString()!.ToUpperInvariant());
        });
    var reply = await agent.SendAsync("Use the registered tools to demonstrate the public library.");
    Console.WriteLine(reply);
    var returned = agent.History().Events().Where(e => e.Kind == EventKind.ToolReturned).ToArray();
    if (returned.Length != 3 || returned.Any(e => e.IsError || e.Job is null))
    {
        throw new InvalidOperationException("Expected three successful supervised tool results");
    }
    Console.WriteLine("Three public-library jobs completed with durable output.");
    return 0;
}
finally
{
    app.Destroy();
}
