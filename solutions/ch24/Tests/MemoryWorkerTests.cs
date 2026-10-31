using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    private static MemoryCapture WorkerCapture(Agent parent, bool curated = false)
    {
        MemoryBand band = curated ? MemoryBand.Archive64 : MemoryBand.Session;
        string corpus = Path.Combine(parent.DataDirectory(), curated ? "bucket-1" : "memory");
        Directory.CreateDirectory(corpus);
        string first = curated ? "2026-01-01-1--2026-01-01-2.md" : "2026-01-01-1.md";
        string second = curated ? "2026-01-01-3--2026-01-01-4.md" : "2026-01-01-2.md";
        File.WriteAllText(Path.Combine(corpus, first), "SOURCE_ONE");
        File.WriteAllText(Path.Combine(corpus, second), "SOURCE_TWO");
        if (curated)
        {
            File.WriteAllText(Path.Combine(parent.DataDirectory(), "MEMORY.md"), "EXISTING_CURATED_REASON");
            foreach (var entry in parent.Memory().Scan(parent.History().View(), MemoryBand.Memory))
            {
                parent.History().Append(new Event(EventKind.MemoryPopulated, memory: entry));
            }
        }
        foreach (var entry in parent.Memory().Scan(parent.History().View(), band))
        {
            parent.History().Append(new Event(EventKind.MemoryPopulated, memory: entry));
        }
        var capture = parent.Memory().CaptureGraduation(parent.History().View(), band,
            2, 1, DateTimeOffset.UtcNow) ?? throw new Exception("Missing worker input");
        parent.History().Append(new Event(EventKind.CompressionLaunched, compression: capture.Launch));
        return capture;
    }

    private static string WorkerChat(string arguments, string finish = "tool_calls", int calls = 1)
    {
        return JsonSerializer.Serialize(new
        {
            model = "fixture",
            choices = new[] { new { message = new { content = "I will preserve the important facts.",
                tool_calls = Enumerable.Range(0, calls).Select(i => new { id = "submit-" + i,
                    type = "function", function = new { name = "submit", arguments } }).ToArray() },
                finish_reason = finish } },
            usage = new { prompt_tokens = 11, completion_tokens = 7 }
        });
    }

    public static async Task MemoryWorkerSuccess(string directory)
    {
        foreach (bool curated in new[] { false, true })
        {
            using var listener = Listen();
            var app = new Application(TextWriter.Null);
            try
            {
                var parent = MemoryAgent(app, Path.Combine(directory, "worker-success-" + curated), listener.Prefixes.Single());
                _ = new Tool(parent, ToolKind.ReadFile);
                var capture = WorkerCapture(parent, curated);
                int before = parent.History().Events().Length;
                var work = parent.RunMemoryCompressorAsync(capture, new MemoryBandSettings(budget: 1500), default);
                var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
                using var reader = new StreamReader(request.Request.InputStream);
                var wire = JsonNode.Parse(await reader.ReadToEndAsync())!;
                var tools = wire["tools"]!.AsArray();
                Check(tools.Count == 1 && tools[0]!["function"]!["name"]!.GetValue<string>() == "submit",
                    "Worker inherited tools or omitted submit");
                var expectedSchema = JsonNode.Parse("""
                    {"type":"object","properties":{"memory":{"type":"string"}},
                     "required":["memory"],"additionalProperties":false}
                    """);
                Check(JsonNode.DeepEquals(tools[0]!["function"]!["parameters"], expectedSchema),
                    "Worker did not advertise the chapter's exact memory schema");
                Check(!wire.AsObject().ContainsKey("max_completion_tokens")
                    && wire.ToJsonString().Contains("SOURCE_ONE") && wire.ToJsonString().Contains("SOURCE_TWO"),
                    "Worker cap or captured input was wrong");
                string prompt = string.Join("\n", wire["messages"]!.AsArray()
                    .Select(message => message!["content"]?.ToString()));
                Check(prompt.Contains("first person", StringComparison.Ordinal)
                    && prompt.Contains("attribute other speakers", StringComparison.Ordinal)
                    && prompt.Contains("Do not invent", StringComparison.Ordinal),
                    "Worker prompt omitted continuation voice, attribution, or uncertainty");
                if (curated)
                {
                    Check(prompt.Split("EXISTING_CURATED_REASON", StringSplitOptions.None).Length == 2,
                        "Terminal worker repeated or omitted existing curated memory");
                }
                var child = parent.Children().Single();
                Check(child.IsMemoryCompressor() && child.SkillManager() is null
                    && child.MemorySettings().Conversation.Disabled, "Worker role or recursive memory enabled");
                string home = child.DataDirectory();
                await Reply(request, WorkerChat("{\"memory\":\"KEEP_THIS_MEMORY\"}"));
                var result = await work.WaitAsync(TimeSpan.FromSeconds(10));
                Check(result.TaskID == capture.Launch.TaskID && result.Summary == "KEEP_THIS_MEMORY"
                    && result.Error is null && !result.Cancelled && result.Meter.Totals.Output == 7,
                    "Valid worker result or usage lost");
                Check(parent.Children().Length == 0 && app.Agents().Length == 1,
                    "Joined worker retained ownership");
                Check(parent.History().Events().Length == before && parent.Memory().StillCurrent(capture, parent.History().View()),
                    "Worker changed parent replay or disk source");
                Check(File.Exists(Path.Combine(parent.DataDirectory(), capture.Launch.TracePath))
                    && File.Exists(Path.Combine(home, "submit.md")) && File.Exists(Path.Combine(home, "save.json")),
                    "Worker trace/submission not retained");
                var saved = File.ReadAllText(Path.Combine(home, "save.json"));
                Check(saved.Contains("KEEP_THIS_MEMORY") && saved.Contains("SOURCE_ONE"), "Trace lost input/output");
            }
            finally { app.Destroy(); }
        }
    }

    public static async Task MemoryWorkerBounds(string directory)
    {
        foreach (string mode in new[] { "invalid", "blank", "prose", "multiple", "forbidden" })
        {
            using var listener = Listen();
            var app = new Application(TextWriter.Null);
            try
            {
                var parent = MemoryAgent(app, Path.Combine(directory, "worker-" + mode), listener.Prefixes.Single());
                var capture = WorkerCapture(parent);
                var work = parent.RunMemoryCompressorAsync(capture, new MemoryBandSettings(), default);
                int expected = mode is "invalid" or "forbidden" ? 3 : mode == "multiple" ? 2 : 1;
                string? home = null;
                for (int i = 0; i < expected; i++)
                {
                    var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
                    var child = parent.Children().Single();
                    home = child.DataDirectory();
                    Check(child.Submission() is null, "Invalid batch published a submission");
                    string response = mode == "prose"
                        ? "{\"choices\":[{\"message\":{\"content\":\"No submit\"},\"finish_reason\":\"stop\"}]}"
                        : WorkerChat(mode == "invalid" ? "{\"memory\":7}"
                            : mode == "blank" ? "{\"memory\":\" \"}" : "{\"memory\":\"valid\"}",
                            calls: mode == "multiple" && i == 0 ? 2 : 1);
                    if (mode == "forbidden")
                    {
                        response = response.Replace("\"name\":\"submit\"", "\"name\":\"read_file\"", StringComparison.Ordinal);
                    }
                    await Reply(request, response);
                }
                var result = await work.WaitAsync(TimeSpan.FromSeconds(10));
                Check((result.Summary is not null) == (mode == "multiple"), "Worker bound accepted invalid result");
                Check(parent.Children().Length == 0, "Failed worker leaked child");
                var save = File.ReadAllText(Path.Combine(home!, "save.json"));
                Check(save.Contains("SOURCE_TWO"), "Failed worker lost trace");
                if (mode is "invalid" or "forbidden")
                {
                    Check(result.Meter.Totals.Output == 21 && !File.Exists(Path.Combine(home!, "submit.md")),
                        "Three-send bound or failed usage wrong");
                }
                if (mode == "blank")
                {
                    Check(result.Error is not null && result.Meter.Totals.Output == 7,
                        "Blank memory was retried or lost its measured usage");
                }
                Check(parent.Memory().StillCurrent(capture, parent.History().View()), "Failed attempt consumed source");
            }
            finally { app.Destroy(); }
        }
    }

    public static async Task MemoryWorkerTerminal(string directory)
    {
        foreach (Vendor vendor in Enum.GetValues<Vendor>())
        {
            foreach (var (limited, streaming) in new[] { (false, false), (true, false), (false, true), (true, true) })
            {
                using var listener = Listen();
                var app = new Application(TextWriter.Null);
                try
                {
                    string home = Path.Combine(directory, "worker-terminal-" + vendor + limited + streaming);
                    Directory.CreateDirectory(home);
                    string model = vendor switch
                    {
                        Vendor.OpenAI => "gpt-4o",
                        Vendor.Anthropic => "claude-3-haiku-20240307",
                        _ => "gemini-2.5-flash"
                    };
                    var parent = new Agent(app, new Configuration(vendor, model, listener.Prefixes.Single(),
                        disableStreaming: !streaming, sandboxRoot: home), dataDirectory: home);
                    var work = parent.RunMemoryCompressorAsync(WorkerCapture(parent), new MemoryBandSettings(), default);
                    var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
                    using var reader = new StreamReader(request.Request.InputStream);
                    var wire = JsonNode.Parse(await reader.ReadToEndAsync())!;
                    if (vendor == Vendor.Gemini)
                    {
                        Check(request.Request.Url!.AbsolutePath.Contains(streaming ? "streamGenerateContent" : "generateContent"),
                            "Gemini worker did not request the selected transport");
                    }
                    else
                    {
                        Check((wire["stream"]?.GetValue<bool>() ?? false) == streaming,
                            "Worker did not request the selected transport");
                    }
                    string reason = limited ? "max_tokens" : "tool_use";
                    string response;
                    if (vendor == Vendor.Anthropic)
                    {
                        Check(wire["max_tokens"]!.GetValue<int>() == 16384, "Anthropic compressor cap too small");
                        response = "{\"content\":[{\"type\":\"text\",\"text\":\"I will retain the facts.\"},"
                            + "{\"type\":\"tool_use\",\"id\":\"s\",\"name\":\"submit\",\"input\":{\"memory\":\"valid\"}}],"
                            + "\"stop_reason\":\"" + reason + "\",\"usage\":{\"input_tokens\":11,\"output_tokens\":7}}";
                    }
                    else if (vendor == Vendor.Gemini)
                    {
                        response = "{\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"I will retain the facts.\"},"
                            + "{\"functionCall\":{\"name\":\"submit\",\"args\":{\"memory\":\"valid\"}}}]},"
                            + "\"finishReason\":\"" + (limited ? "MAX_TOKENS" : "STOP")
                            + "\"}],\"usageMetadata\":{\"promptTokenCount\":11,\"candidatesTokenCount\":7}}";
                    }
                    else
                    {
                        response = WorkerChat("{\"memory\":\"valid\"}", limited ? "length" : "tool_calls");
                    }
                    if (streaming)
                    {
                        await WorkerStreamReply(request, vendor, response);
                    }
                    else
                    {
                        await Reply(request, response);
                    }
                    var result = await work.WaitAsync(TimeSpan.FromSeconds(10));
                    Check((result.Summary is null) == limited && result.Meter.Totals.Output == 7,
                        "Terminal acceptance or rejected-response accounting wrong for " + vendor);
                    Check(parent.Children().Length == 0, "Terminal failure leaked worker");
                }
                finally { app.Destroy(); }
            }
        }
    }

    private static async Task WorkerStreamReply(System.Net.HttpListenerContext request, Vendor vendor, string response)
    {
        var root = JsonNode.Parse(response)!;
        var frames = new List<JsonNode>();
        if (vendor == Vendor.OpenAI)
        {
            var choice = root["choices"]![0]!;
            var message = choice["message"]!.DeepClone();
            foreach (var call in message["tool_calls"]!.AsArray())
            {
                call!["index"] = 0;
            }
            frames.Add(new JsonObject
            {
                ["choices"] = new JsonArray(new JsonObject
                {
                    ["index"] = 0, ["delta"] = message,
                    ["finish_reason"] = choice["finish_reason"]!.DeepClone()
                }),
                ["usage"] = root["usage"]!.DeepClone()
            });
        }
        else if (vendor == Vendor.Anthropic)
        {
            frames.Add(new JsonObject { ["type"] = "message_start",
                ["message"] = new JsonObject { ["usage"] = new JsonObject { ["input_tokens"] = 11 } } });
            int index = 0;
            foreach (var part in root["content"]!.AsArray())
            {
                frames.Add(new JsonObject { ["type"] = "content_block_start", ["index"] = index,
                    ["content_block"] = part!.DeepClone() });
                frames.Add(new JsonObject { ["type"] = "content_block_stop", ["index"] = index });
                index++;
            }
            frames.Add(new JsonObject { ["type"] = "message_delta", ["delta"] = new JsonObject
                { ["stop_reason"] = root["stop_reason"]!.DeepClone() },
                ["usage"] = new JsonObject { ["output_tokens"] = 7 } });
            frames.Add(new JsonObject { ["type"] = "message_stop" });
        }
        else
        {
            frames.Add(root);
        }
        string body = string.Concat(frames.Select(frame => "data: " + frame.ToJsonString() + "\n\n"));
        if (vendor == Vendor.OpenAI)
        {
            body += "data: [DONE]\n\n";
        }
        byte[] bytes = System.Text.Encoding.UTF8.GetBytes(body);
        request.Response.StatusCode = 200;
        request.Response.ContentType = "text/event-stream";
        request.Response.ContentLength64 = bytes.Length;
        await request.Response.OutputStream.WriteAsync(bytes);
        request.Response.Close();
    }

    public static async Task MemoryWorkerCancellation(string directory)
    {
        using var listener = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var parent = MemoryAgent(app, Path.Combine(directory, "worker-cancellation"), listener.Prefixes.Single());
            var capture = WorkerCapture(parent);
            using var cancellation = new CancellationTokenSource();
            var work = parent.RunMemoryCompressorAsync(capture, new MemoryBandSettings(), cancellation.Token);
            var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            string home = parent.Children().Single().DataDirectory();
            cancellation.Cancel();
            var result = await work.WaitAsync(TimeSpan.FromSeconds(10));
            request.Response.Abort();
            Check(result.Cancelled && result.Summary is null && parent.Children().Length == 0,
                "Cancellation did not join/detach worker");
            Check(File.Exists(Path.Combine(home, "save.json"))
                && parent.Memory().StillCurrent(capture, parent.History().View()), "Cancellation lost trace/source");
        }
        finally { app.Destroy(); }
    }
}
