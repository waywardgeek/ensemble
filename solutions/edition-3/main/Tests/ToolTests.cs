using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

// These tests execute real files and shells. Only the remote model is replaced.
internal static class ToolTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition)
        {
            throw new Exception(message);
        }
    }

    private static JsonElement Args(object value) => JsonSerializer.SerializeToElement(value);

    internal static async Task Run(string directory)
    {
        var app = new Application(TextWriter.Null);
        var agent = new Agent(app, new Configuration(Vendor.Anthropic, "local", "http://127.0.0.1"));
        try
        {
            Check(agent.Tools().Length == 0, "new library Agent must have an empty registry");
            Check(!agent.Engine().Render(new Context()).ContainsKey("tools"), "omit empty tools field");
            agent.InstallTools();
            string[] expectedTools = ["read_file", "write_file", "edit_file", "run_command", "list_directory", "search_files",
                "wait_for_job", "send_input", "kill_job", "tool_limits", "micro_handoff", "agent_status", "join_agents", "submit", "send_message_to_parent",
                "respond_to_parent_message", "shutdown_agent", "spawn_sub_agent", "agents_status", "get_sub_agent_status", "interrupt_agent", "get_submission", "kill_agent", "send_message", "ask_agent", "wait_for_agent", "wait_for_agent_change"];
            Check(agent.Tools().Select(value => value.Name()).Order().SequenceEqual(expectedTools.Order()),
                "install file/process tools, job controls and read-only Agent status");
            var other = new Agent(app, agent.Configuration());
            Check(other.Tools().Length == 0, "registries are Agent-local");
            var tool = agent.Tools()[0];
            agent.RemoveTool(tool);
            Check(tool.parent is null && !agent.Tools().Contains(tool), "remove completes both ends");
            agent.AddTool(tool);
            Check(tool.Parent() == agent && agent.Tools().Contains(tool), "add completes both ends");

            async Task<ToolResultPart> Call(string name, object input)
            {
                return await agent.DispatchAsync(new ToolCallPart("test", agent.Configuration().Target(), name, Args(input)));
            }
            static string Text(ToolResultPart result)
            {
                var report = JsonCodec.Text(result.Parts);
                return report.StartsWith("job ", StringComparison.Ordinal) ? report[(report.IndexOf('\n') + 1)..] : report;
            }
            var file = Path.Combine(directory, "tool.txt");
            Check(!(await Call("write_file", new
            {
                path = file,
                content = "original\n"
            })).IsError, "create");
            Check((await Call("write_file", new
            {
                path = file,
                content = "bad"
            })).IsError, "overwrite refused");
            Check(File.ReadAllText(file) == "original\n", "refusal preserves bytes");
            Check(!(await Call("write_file", new
            {
                path = file,
                content = "tail\n",
                append = true
            })).IsError, "append");
            Check(File.ReadAllText(file) == "original\ntail\n", "append preserves bytes");
            Check(Text(await Call("read_file", new
            {
                path = file,
                start_line = 2,
                end_line = 2
            })) == "tail\n", "inclusive range");
            foreach (var bad in new object[] { new { path = 42 }, new { path = file, start_line = 1.5 }, new { path = file, start_line = "2" } })
            {
                Check((await Call("read_file", bad)).IsError, "strict argument type");
            }
            Check((await Call("write_file", new
            {
                path = file,
                content = "bad",
                overwrite = "true"
            })).IsError, "strict boolean");
            Check((await Call("unknown", new
            {
            })).IsError, "unknown tool is a result");
            File.WriteAllText(file, "aaa");
            Check((await Call("edit_file", new
            {
                path = file,
                old_text = "aa",
                new_text = "b"
            })).IsError, "overlapping sites refused");
            Check(File.ReadAllText(file) == "aaa", "ambiguous edit preserved");
            Check((await Call("edit_file", new
            {
                path = file,
                old_text = "",
                new_text = "b"
            })).IsError, "empty anchor refused");
            Check((await Call("edit_file", new
            {
                path = file,
                old_text = "missing",
                new_text = "b"
            })).IsError, "zero sites refused");
            Check(!(await Call("edit_file", new
            {
                path = file,
                old_text = "aaa",
                new_text = "done"
            })).IsError, "unique edit succeeds");
            Check(File.ReadAllText(file) == "done", "actual replacement");

            File.WriteAllText(file, "one\nneedle\nthree\nfour\nneedle\nsix\nseven\neight\nneedle\nten\n");
            var search = Text(await Call("search_files", new
            {
                path = file,
                pattern = "needle",
                context_lines = 1
            }));
            var expected = $"{file}-1-one\n{file}:2:needle\n{file}-3-three\n{file}-4-four\n{file}:5:needle\n{file}-6-six\n--\n{file}-8-eight\n{file}:9:needle\n{file}-10-ten\n";
            Check(search == expected, "merged grep context windows");
            Check((await Call("search_files", new
            {
                path = file,
                pattern = "["
            })).IsError, "bad regex is a result");
            Check(!(await Call("search_files", new
            {
                path = file,
                pattern = "absent"
            })).IsError, "no matches is success");
            Directory.CreateDirectory(Path.Combine(directory, "zzz"));
            var listing = Text(await Call("list_directory", new
            {
                path = directory
            }));
            Check(listing.Contains("zzz/"), "directory marker");

            var originalKey = Environment.GetEnvironmentVariable("ANTHROPIC_API_KEY");
            try
            {
                Environment.SetEnvironmentVariable("ANTHROPIC_API_KEY", "never-in-child");
                var capped = await Call("read_file", new
                {
                    path = file,
                    start_line = 0,
                    end_line = 0,
                    max_bytes = 3
                });
                Check(!capped.IsError && Text(capped).StartsWith("one") && JsonCodec.Text(capped.Parts).Contains("truncated"), "read cap follows line selection and reports truncation");
                var command = await Call("run_command", new
                {
                    command = "printf '%s' \"${ANTHROPIC_API_KEY-unset}\"; printf err >&2; exit 7",
                    cwd = directory
                });
                Check(!command.IsError, "nonzero exit is ordinary data");
                Check(Text(command).Contains("unset") && Text(command).Contains("err") && Text(command).Contains("7"), "stdout stderr exit and key stripping");
                Check(!Text(command).Contains("never-in-child"), "key not inherited");
                await Call("run_command", new
                {
                    command = "export ONLY_THIS_SHELL=value; cd /",
                    cwd = directory
                });
                var isolated = Text(await Call("run_command", new
                {
                    command = "printf '%s\\n' \"${ONLY_THIS_SHELL-unset}\"; pwd",
                    cwd = directory
                }));
                Check(isolated.Contains("unset") && isolated.Contains(directory), "fresh environment and cwd");
            }
            finally
            {
                Environment.SetEnvironmentVariable("ANTHROPIC_API_KEY", originalKey);
            }
            using (var cancellation = new CancellationTokenSource(TimeSpan.FromMilliseconds(150)))
            {
                var interrupted = false;
                try
                {
                    var call = new ToolCallPart("cancel", agent.Configuration().Target(), "run_command",
                        Args(new { command = "sleep 30", cwd = directory }));
                    var unexpected = await agent.DispatchAsync(call, cancellation.Token).WaitAsync(TimeSpan.FromSeconds(5));
                    throw new Exception("Expected cancelled observation; returned: " + JsonCodec.Text(unexpected.Parts));
                }
                catch (OperationCanceledException)
                {
                    interrupted = true;
                }
                Check(interrupted, "caller cancellation ends observation, not the job");
            }
            int attached = agent.Tools().Length;
            tool.Destroy();
            Check(tool.parent is null && agent.Tools().Length == attached - 1, "destroy detaches tool");
            agent.Destroy();
            Check(agent.Tools().Length == 0, "Agent cascades through tools");
        }
        finally
        {
            app.Destroy();
        }
        Console.WriteLine("PASS real file/shell tools, strict arguments, registry and ownership");
        ProjectionTests();
        await Loop(directory, false);
        await Loop(directory, true);
        await RoundLimit();
    }

    private static void ProjectionTests()
    {
        var app = new Application(TextWriter.Null);
        try
        {
            foreach (var vendor in Enum.GetValues<Vendor>())
            {
                var agent = new Agent(app, new Configuration(vendor, "local", "http://127.0.0.1"));
                agent.InstallTools();
                var request = agent.Engine().Render(new Context());
                Check(request.ContainsKey("tools"), "each vendor declares tools");
                foreach (var tool in agent.Tools())
                {
                    var original = tool.Declaration(Vendor.Anthropic).ToJsonString();
                    var gemini = tool.Declaration(Vendor.Gemini).ToJsonString();
                    Check(!gemini.Contains("additionalProperties"), "Gemini strips unsupported schema fields");
                    Check(tool.Declaration(Vendor.Anthropic).ToJsonString() == original, "projection cannot mutate definition");
                    if (tool.Name() == "read_file")
                    {
                        Check(original.Contains("additionalProperties"), "projection stripping needs a positive control");
                    }
                }
                agent.Destroy();
            }
            var geminiAgent = new Agent(app, new Configuration(Vendor.Gemini, "local", "http://127.0.0.1"));
            const string reply = """{"modelVersion":"local","candidates":[{"content":{"parts":[{"text":"Visible","thoughtSignature":"text-sig"},{"functionCall":{"id":0,"name":"read_file","args":{"path":"a"}},"thoughtSignature":"call-sig"},{"functionCall":{"name":"read_file","args":{"path":"b"}}}]}}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1}}""";
            var events = geminiAgent.Engine().Parse(200, reply);
            Check(events.All(e => e.Kind != EventKind.ErrorOccurred), "numeric and absent IDs parse");
            var calls = events.Last().Parts.OfType<ToolCallPart>().ToArray();
            Check(calls[0].CallID == "0" && calls[0].CallID != calls[1].CallID, "numeric zero preserved and missing ID distinct");
            foreach (var parsed in events)
            {
                geminiAgent.History().Append(parsed);
            }
            var recorded = geminiAgent.History().Events().Last();
            var restored = JsonCodec.ReadEvent(JsonCodec.WriteEvent(recorded));
            var context = new Context([new Entry(1, Actor.Agent, restored.Parts)]);
            var rendered = geminiAgent.Engine().Render(context).ToJsonString();
            Check(rendered.Contains("text-sig") && rendered.Contains("call-sig"), "both bound signatures round trip");
            var foreign = new Agent(app, new Configuration(Vendor.Gemini, "other", "http://127.0.0.1"));
            var foreignRequest = foreign.Engine().Render(context).ToJsonString();
            Check(!foreignRequest.Contains("text-sig") && !foreignRequest.Contains("call-sig") && foreignRequest.Contains("Visible"), "foreign model loses opaque material, not visible text");
            var anthropic = new Agent(app, new Configuration(Vendor.Anthropic, "local", "http://127.0.0.1"));
            var duplicate = anthropic.Engine().Parse(200, """{"model":"local","content":[{"type":"tool_use","id":"same","name":"write_file","input":{}},{"type":"tool_use","id":"same","name":"read_file","input":{}}],"usage":{"input_tokens":1,"output_tokens":1}}""");
            Check(duplicate.Length == 1 && duplicate[0].Kind == EventKind.ErrorOccurred, "duplicate IDs reject entire response");
        }
        finally
        {
            app.Destroy();
        }
        Console.WriteLine("PASS provider declarations, identity normalization and bound signatures");
    }

    private static async Task RoundLimit()
    {
        var socket = new System.Net.Sockets.TcpListener(IPAddress.Loopback, 0);
        socket.Start();
        int port = ((IPEndPoint)socket.LocalEndpoint).Port;
        socket.Stop();
        using var listener = new HttpListener();
        listener.Prefixes.Add($"http://127.0.0.1:{port}/");
        listener.Start();
        var server = Task.Run(async () =>
        {
            for (int round = 0; round < 4; round++)
            {
                var context = await listener.GetContextAsync();
                using var reader = new StreamReader(context.Request.InputStream);
                await reader.ReadToEndAsync();
                // Reusing a call ID in another response is valid. The unknown
                // tool yields a result and lets this fake repeat indefinitely.
                var bytes = Encoding.UTF8.GetBytes("""{"model":"local","content":[{"type":"tool_use","id":"again","name":"absent","input":{}}],"usage":{"input_tokens":1,"output_tokens":1}}""");
                context.Response.ContentType = "application/json";
                context.Response.ContentLength64 = bytes.Length;
                await context.Response.OutputStream.WriteAsync(bytes);
                context.Response.Close();
            }
        });
        var app = new Application(TextWriter.Null);
        var agent = new Agent(app, new Configuration(Vendor.Anthropic, "local", $"http://127.0.0.1:{port}", "dummy", toolRoundLimit: 3));
        try
        {
            var refused = false;
            try
            {
                await agent.SendAsync("Repeat").WaitAsync(TimeSpan.FromSeconds(10));
            }
            catch (InvalidOperationException ex) when (ex.Message.Contains("batch limit"))
            {
                refused = true;
            }
            await server.WaitAsync(TimeSpan.FromSeconds(5));
            Check(refused && agent.History().Events().Count(e => e.Kind == EventKind.ToolReturned) == 4,
                "three dispatched batches and one explicitly refused batch settle the limit");
        }
        finally
        {
            app.Destroy();
        }
        Console.WriteLine("PASS bounded actor tool batches with a final result-consuming exchange");
    }

    private static async Task Loop(string directory, bool strict)
    {
        var probe = new TcpListener(IPAddress.Loopback, 0);
        probe.Start();
        var port = ((IPEndPoint)probe.LocalEndpoint).Port;
        probe.Stop();
        using var listener = new HttpListener();
        listener.Prefixes.Add($"http://127.0.0.1:{port}/");
        listener.Start();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var path = Path.Combine(directory, strict ? "strict-loop.txt" : "from-loop.txt");
        var rounds = strict ? 3 : 2;
        var requests = new List<JsonObject>();
        var server = Task.Run(async () =>
        {
            for (int round = 0; round < rounds; round++)
            {
                var request = await listener.GetContextAsync().WaitAsync(deadline.Token);
                using var reader = new StreamReader(request.Request.InputStream);
                requests.Add(JsonNode.Parse(await reader.ReadToEndAsync(deadline.Token))!.AsObject());
                var parts = new JsonArray();
                if (strict && round == 0)
                {
                    parts.Add(new JsonObject { ["type"] = "tool_use", ["id"] = "refused", ["name"] = "write_file", ["input"] = new JsonObject { ["path"] = path, ["content"] = "must not write" } });
                }
                else if (round == rounds - 2)
                {
                    if (strict)
                    {
                        Check(!File.Exists(path), "missing explanation prevents the effect");
                        Check(requests[^1].ToJsonString().Contains("first explain"), "missing explanation returns a repairable error");
                    }
                    parts.Add(new JsonObject { ["type"] = "text", ["text"] = "I will write and read the file." });
                    parts.Add(new JsonObject { ["type"] = "tool_use", ["id"] = "write", ["name"] = "write_file", ["input"] = new JsonObject { ["path"] = path, ["content"] = "actual bytes" } });
                    parts.Add(new JsonObject { ["type"] = "tool_use", ["id"] = "missing", ["name"] = "read_file", ["input"] = new JsonObject { ["path"] = path + ".absent" } });
                    parts.Add(new JsonObject { ["type"] = "tool_use", ["id"] = "read", ["name"] = "read_file", ["input"] = new JsonObject { ["path"] = path } });
                }
                else
                {
                    parts.Add(new JsonObject { ["type"] = "text", ["text"] = "Finished." });
                }
                var reply = new JsonObject { ["model"] = "local", ["content"] = parts, ["usage"] = new JsonObject { ["input_tokens"] = 2, ["output_tokens"] = 1 } };
                var bytes = Encoding.UTF8.GetBytes(reply.ToJsonString());
                request.Response.ContentType = "application/json";
                request.Response.ContentLength64 = bytes.Length;
                await request.Response.OutputStream.WriteAsync(bytes, deadline.Token);
                request.Response.Close();
            }
        });
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "local", $"http://127.0.0.1:{port}", "dummy", requireToolExplanation: strict));
            agent.InstallTools();
            var answer = await agent.SendAsync("Exercise the tools", deadline.Token);
            await server;
            Check(answer.Contains("Finished."), "loop reaches final answer");
            Check(File.ReadAllText(path) == "actual bytes", "tool effect is real");
            var installedNames = agent.Tools().Select(tool => tool.Name()).Order().ToArray();
            Check(requests.All(request => request["tools"]!.AsArray()
                .Select(tool => tool!["name"]!.GetValue<string>()).Order().SequenceEqual(installedNames)),
                "installed tools declared every round");
            Check(requests[^1].ToJsonString().Contains("actual bytes") && requests[^1].ToJsonString().Contains("is_error"), "real result reaches provider");
            var events = agent.History().Events();
            Check(events.Count(e => e.Kind == EventKind.ToolCalled) == rounds + 1 && events.Count(e => e.Kind == EventKind.ToolReturned) == rounds + 1, "paired dispatch facts");
            foreach (var id in new[] { "write", "missing", "read" })
            {
                Check(events.Single(e => e.Kind == EventKind.ToolCalled && e.CallID == id).Seq < events.Single(e => e.Kind == EventKind.ToolReturned && e.CallID == id).Seq, "record dispatch before result");
            }
            Check(agent.History().View().Usage.Input == 2 * rounds && agent.History().View().Phase == Phase.Idle, "all rounds count and settle");
        }
        finally
        {
            app.Destroy();
            listener.Stop();
        }
        Console.WriteLine("PASS real Agent HTTP multi-call loop and result publication");
    }
}
