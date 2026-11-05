using System.Globalization;
using System.Text;
using System.Text.Json;
using Ensemble;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.Logging;

internal static class Program
{
    private static readonly List<string> checks = [];

    private static void Check(bool condition, string name)
    {
        if (!condition) throw new InvalidOperationException(name);
        lock (checks) checks.Add(name);
    }

    public static async Task<int> Main()
    {
        try
        {
            await Run();
            Console.WriteLine(JsonSerializer.Serialize(new { contract = "chapter4-model-consumer-v1", passed = true, checks }));
            return 0;
        }
        catch (Exception error)
        {
            Console.WriteLine(JsonSerializer.Serialize(new { contract = "chapter4-model-consumer-v1", passed = false, checks, failure = error.Message }));
            return 1;
        }
    }

    private static async Task Run()
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(60));
        using var diagnostics = new StringWriter();
        Ensemble.Ensemble? root = null;
        var builder = WebApplication.CreateBuilder(Array.Empty<string>());
        builder.Logging.ClearProviders();
        builder.WebHost.UseUrls("http://127.0.0.1:0");
        await using var host = builder.Build();
        Agent? agent = null;
        int requests = 0;
        int executions = 0;
        string? providerFailure = null;
        string nonce = Guid.NewGuid().ToString("N");
        string key = "synthetic-" + nonce;
        string shoutInput = "hello framework " + nonce;
        string shoutOutput = shoutInput.ToUpperInvariant();
        string fileOutput = "file-" + nonce + " 界\n";
        string shellOutput = "shell-" + nonce;
        string answer = "verified-" + nonce;
        string logMessage = "shout called " + nonce;
        string file = Path.GetFullPath("input-" + nonce + ".txt");
        string[] ids = ["shout-" + nonce, "read-" + nonce, "command-" + nonce];
        string[] names = ["shout", "read_file", "run_command"];
        string[] expected = [shoutOutput, fileOutput, shellOutput];
        var began = DateTimeOffset.UtcNow;
        File.WriteAllText(file, fileOutput, new UTF8Encoding(false));

        host.MapPost("/v1/messages", async (HttpContext context) =>
        {
            try
            {
                using var document = await JsonDocument.ParseAsync(context.Request.Body, cancellationToken: deadline.Token);
                var request = document.RootElement;
                Check(context.Request.Headers["x-api-key"] == key, "provider-synthetic-key-" + (requests + 1));
                Check(!request.TryGetProperty("stream", out var stream) || !stream.GetBoolean(),
                    "provider-buffered-request-" + (requests + 1));
                int turn = Interlocked.Increment(ref requests);
                object[] content;
                if (turn == 1)
                {
                    var tools = request.GetProperty("tools").EnumerateArray().ToArray();
                    foreach (string name in new[] { "shout", "read_file", "list_directory", "search_files", "run_command", "write_file", "edit_file" })
                        Check(tools.Count(tool => tool.GetProperty("name").GetString() == name) == 1,
                            "declared-on-real-request-" + name);
                    Check(agent!.History().Events().Any(fact => fact.Kind == EventKind.RequestSent),
                        "request-recorded-before-provider");
                    content = [
                        new { type = "text", text = "I will invoke the custom tool and two preserved built-ins." },
                        new { type = "tool_use", id = ids[0], name = names[0], input = (object)new { text = shoutInput, ai_callback_delay = 10 } },
                        new { type = "tool_use", id = ids[1], name = names[1], input = (object)new { path = file, ai_callback_delay = 10 } },
                        new { type = "tool_use", id = ids[2], name = names[2], input = (object)new { command = "printf '%s\\n' '" + shellOutput + "'", ai_callback_delay = 10 } }
                    ];
                }
                else
                {
                    Check(turn == 2, "exactly-two-provider-requests");
                    var facts = agent!.History().Events();
                    var returned = facts.Where(fact => fact.Kind == EventKind.ToolReturned).ToArray();
                    Check(returned.Length == 3, "three-real-recorded-results-before-answer");
                    var wireResults = request.GetProperty("messages").EnumerateArray()
                        .SelectMany(message => message.GetProperty("content").ValueKind == JsonValueKind.Array
                            ? message.GetProperty("content").EnumerateArray().ToArray() : [])
                        .Where(part => part.GetProperty("type").GetString() == "tool_result").ToArray();
                    Check(wireResults.Length == 3, "three-correlated-provider-results");
                    var handles = new HashSet<int>();
                    for (int index = 0; index < ids.Length; index++)
                    {
                        var called = facts.Single(fact => fact.Kind == EventKind.ToolCalled && fact.CallID == ids[index]);
                        var result = returned.Single(fact => fact.CallID == ids[index]);
                        Check(called.Name == names[index] && called.Seq < result.Seq
                            && !result.IsError && result.Job is { Status: "done" }, "successful-paired-result-" + names[index]);
                        Check(JsonCodec.Text(result.Parts).Contains(expected[index], StringComparison.Ordinal),
                            "canonical-real-output-" + names[index]);
                        var job = result.Job!;
                        Check(handles.Add(job.Handle) && called.Job?.Handle == job.Handle,
                            "distinct-owned-job-" + names[index]);
                        byte[] bytes = File.ReadAllBytes(job.OutputPath);
                        Check(bytes.LongLength == job.Bytes && Encoding.UTF8.GetString(bytes).Contains(expected[index], StringComparison.Ordinal),
                            "durable-output-before-answer-" + names[index]);
                        var wire = wireResults.Single(part => part.GetProperty("tool_use_id").GetString() == ids[index]);
                        Check((!wire.TryGetProperty("is_error", out var error) || !error.GetBoolean())
                            && WireText(wire.GetProperty("content")).Contains(expected[index], StringComparison.Ordinal),
                            "provider-real-output-" + names[index]);
                    }
                    Check(returned.Single(fact => fact.CallID == ids[2]).Job!.ExitCode == 0,
                        "shell-success-exit-code");
                    Check(executions == 1, "custom-executed-once-before-answer");
                    var line = diagnostics.ToString().Split('\n', StringSplitOptions.RemoveEmptyEntries)
                        .Single(value => value.TrimEnd('\r').EndsWith(" " + logMessage, StringComparison.Ordinal));
                    int space = line.IndexOf(' ');
                    Check(space > 0 && DateTimeOffset.TryParse(line[..space], CultureInfo.InvariantCulture,
                        DateTimeStyles.RoundtripKind, out var recorded) && recorded >= began && recorded <= DateTimeOffset.UtcNow,
                        "real-root-timestamped-diagnostic-before-answer");
                    content = [new { type = "text", text = answer }];
                }
                await context.Response.WriteAsJsonAsync(new
                {
                    id = "msg-" + nonce + "-" + turn, type = "message", role = "assistant",
                    model = "claude-3-haiku-20240307", content,
                    stop_reason = turn == 1 ? "tool_use" : "end_turn",
                    usage = new { input_tokens = 40, output_tokens = 20 }
                }, cancellationToken: deadline.Token);
            }
            catch (Exception error)
            {
                providerFailure = error.Message;
                context.Response.StatusCode = 500;
                await context.Response.WriteAsync("Owned fake rejected the candidate request.", CancellationToken.None);
            }
        });

        try
        {
            await host.StartAsync(deadline.Token);
            var endpoint = new Uri(host.Urls.Single());
            Check(endpoint.IsLoopback && endpoint.Port is not (8088 or 8089 or 8090), "owned-isolated-loopback-provider");
            root = new Ensemble.Ensemble(diagnostics);
            agent = new Agent(root, CandidateEntry.Configuration(endpoint.ToString(), key));
            agent.InstallTools();
            int builtins = agent.Tools().Length;
            var tool = new Tool(agent, "shout", "Return text in uppercase.", JsonSerializer.SerializeToElement(new
            {
                type = "object", properties = new { text = new { type = "string" } },
                required = new[] { "text" }, additionalProperties = false
            }), (job, arguments, cancellation) =>
            {
                cancellation.ThrowIfCancellationRequested();
                Check(ReferenceEquals(job.Parent().Parent(), agent)
                    && ReferenceEquals(job.Parent().Parent().Parent(), root), "executing-custom-real-parent-chain");
                Check(agent.History().Events().Any(fact => fact.Kind == EventKind.ToolCalled
                    && fact.CallID == ids[0] && fact.Job?.Handle == job.Handle), "custom-call-recorded-before-effect");
                Interlocked.Increment(ref executions);
                job.Parent().Parent().Parent().Log(logMessage);
                return Task.FromResult(arguments.GetProperty("text").GetString()!.ToUpperInvariant());
            });
            Check(ReferenceEquals(tool.Parent(), agent) && agent.Tools().Length == builtins + 1,
                "custom-registration-preserves-builtins");
            string reply = await CandidateEntry.Run(agent, "Use the registered tools to demonstrate the public library.", deadline.Token);
            if (providerFailure is not null) throw new InvalidOperationException(providerFailure);
            Check(reply == answer && requests == 2 && executions == 1, "final-answer-after-verified-real-effects");
        }
        catch (Exception error) when (providerFailure is not null)
        {
            throw new InvalidOperationException(providerFailure, error);
        }
        finally
        {
            deadline.Cancel();
            try { root?.Destroy(); }
            finally { await host.StopAsync(CancellationToken.None); }
        }
    }

    private static string WireText(JsonElement content)
    {
        if (content.ValueKind == JsonValueKind.String) return content.GetString() ?? "";
        if (content.ValueKind != JsonValueKind.Array) return "";
        return string.Concat(content.EnumerateArray()
            .Where(part => part.TryGetProperty("type", out var type) && type.GetString() == "text")
            .Select(part => part.GetProperty("text").GetString()));
    }
}
