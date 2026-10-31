using System.Collections.Immutable;
using System.Net;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    private static JsonObject JudgeReply(Surface surface, string text, string model, bool complete = true)
    {
        var root = surface switch
        {
            Surface.Messages => new JsonObject
            {
                ["model"] = model, ["stop_reason"] = complete ? "end_turn" : "max_tokens",
                ["content"] = new JsonArray(new JsonObject { ["type"] = "text", ["text"] = text }),
                ["usage"] = new JsonObject { ["input_tokens"] = 13, ["output_tokens"] = 2 }
            },
            Surface.GenerateContent => new JsonObject
            {
                ["modelVersion"] = model,
                ["candidates"] = new JsonArray(new JsonObject
                {
                    ["content"] = new JsonObject { ["parts"] = new JsonArray(new JsonObject { ["text"] = text }) },
                    ["finishReason"] = complete ? "STOP" : "MAX_TOKENS"
                }),
                ["usageMetadata"] = new JsonObject { ["promptTokenCount"] = 13, ["candidatesTokenCount"] = 2 }
            },
            Surface.Responses => new JsonObject
            {
                ["model"] = model, ["status"] = complete ? "completed" : "incomplete",
                ["output"] = new JsonArray(new JsonObject
                {
                    ["type"] = "message", ["role"] = "assistant", ["status"] = "completed",
                    ["content"] = new JsonArray(new JsonObject { ["type"] = "output_text", ["text"] = text })
                }),
                ["usage"] = new JsonObject { ["input_tokens"] = 13, ["output_tokens"] = 2 }
            },
            _ => new JsonObject
            {
                ["model"] = model,
                ["choices"] = new JsonArray(new JsonObject
                {
                    ["message"] = new JsonObject { ["role"] = "assistant", ["content"] = text },
                    ["finish_reason"] = complete ? "stop" : "length"
                }),
                ["usage"] = new JsonObject { ["prompt_tokens"] = 13, ["completion_tokens"] = 2 }
            }
        };
        root["service_tier"] = "default";
        return root;
    }

    private static async Task SendJudgeReply(HttpListenerContext request, Surface surface,
        string text, string model, bool stream = false, bool complete = true)
    {
        var body = JudgeReply(surface, text, model, complete);
        if (!stream)
        {
            await Reply(request, body.ToJsonString());
            return;
        }
        var frames = new List<JsonObject>();
        if (surface == Surface.Responses)
        {
            frames.Add(new JsonObject { ["type"] = complete ? "response.completed" : "response.incomplete", ["response"] = body });
        }
        else if (surface == Surface.Messages)
        {
            frames.Add(new JsonObject
            {
                ["type"] = "message_start", ["message"] = new JsonObject
                {
                    ["model"] = model, ["content"] = new JsonArray(), ["usage"] = body["usage"]!.DeepClone()
                }
            });
            frames.Add(new JsonObject { ["type"] = "content_block_start", ["index"] = 0, ["content_block"] = body["content"]![0]!.DeepClone() });
            frames.Add(new JsonObject { ["type"] = "content_block_stop", ["index"] = 0 });
            frames.Add(new JsonObject
            {
                ["type"] = "message_delta", ["delta"] = new JsonObject { ["stop_reason"] = body["stop_reason"]!.DeepClone() },
                ["usage"] = new JsonObject { ["output_tokens"] = 2 }
            });
            frames.Add(new JsonObject { ["type"] = "message_stop" });
        }
        else if (surface == Surface.ChatCompletions)
        {
            var choice = body["choices"]![0]!.AsObject();
            choice["delta"] = choice["message"]!.DeepClone();
            choice.Remove("message");
            frames.Add(body);
        }
        else
        {
            frames.Add(body);
        }
        string wire = string.Join("", frames.Select(frame => "data: " + frame.ToJsonString() + "\n\n"));
        if (surface == Surface.ChatCompletions)
        {
            wire += "data: [DONE]\n\n";
        }
        await Reply(request, wire, type: "text/event-stream");
    }

    private static ImmutableArray<RecallHit> JudgeHits() =>
    [
        new(new RecallSnippet(RecallSource.Docs, "docs/first.md", "First", "FIRST_DOC"), 10),
        new(new RecallSnippet(RecallSource.Docs, "docs/second.md", "Second", "SECOND_DOC"), 9),
        new(new RecallSnippet(RecallSource.Memory, "memory/old.md", "Memory", "MEMORY_FACT"), 1)
    ];

    public static async Task RecallJudgeTransport(string directory)
    {
        foreach (Surface surface in Enum.GetValues<Surface>())
        foreach (bool stream in new[] { false, true })
        {
            using var listener = Listen();
            var app = new Application(TextWriter.Null);
            try
            {
                Vendor vendor = surface == Surface.Messages ? Vendor.Anthropic
                    : surface == Surface.GenerateContent ? Vendor.Gemini : Vendor.OpenAI;
                string model = vendor == Vendor.Anthropic ? "claude-3-haiku-20240307"
                    : vendor == Vendor.Gemini ? "gemini-2.5-flash" : "gpt-4o-mini";
                string judge = vendor == Vendor.OpenAI ? "gpt-4o" : model;
                var config = new Configuration(vendor, model, listener.Prefixes.Single(), "fixture-key",
                    primaryInstructions: "MAIN_INSTRUCTIONS_SECRET", disableStreaming: !stream,
                    cacheBreakpoints: vendor == Vendor.Anthropic, surface: surface);
                var agent = new Agent(app, config);
                var engine = agent.Engine();
                string lensPath = Path.Combine(directory, "judge-lens-" + surface + stream);
                _ = new CacheLens(engine, lensPath);

                var main = engine.SendAsync(engine.Render(new Context()).ToJsonString());
                var mainRequest = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
                await SendJudgeReply(mainRequest, surface, "ordinary", "actual-main");
                await main;
                var before = engine.Meter();
                string[] lensBefore = Directory.GetFiles(lensPath, "*", SearchOption.AllDirectories)
                    .Order().Select(path => path + ":" + Convert.ToHexString(System.Security.Cryptography.SHA256.HashData(File.ReadAllBytes(path)))).ToArray();
                long sequence = agent.History().Snapshot().Sequence;
                var send = engine.JudgeRecallAsync(judge, "JUDGE_INSTRUCTION", "UNTRUSTED_DATA", default);
                var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
                using var reader = new StreamReader(request.Request.InputStream);
                string wireText = await reader.ReadToEndAsync();
                var wire = JsonNode.Parse(wireText)!;
                Check(!wireText.Contains("MAIN_INSTRUCTIONS_SECRET") && !wireText.Contains("cache_control")
                    && wire["tools"] is null && wireText.Contains("UNTRUSTED_DATA") && wireText.Contains("JUDGE_INSTRUCTION"),
                    "Judge inherited main instructions/tools/cache state: " + surface);
                if (surface != Surface.GenerateContent)
                {
                    Check(wire["model"]!.GetValue<string>() == judge, "Judge override not rendered");
                    Check((wire["stream"]?.GetValue<bool>() == true) == stream, "Judge streaming request policy");
                }
                else
                {
                    Check(request.Request.Url!.AbsolutePath.Contains(judge), "Gemini judge route lost model");
                }
                await SendJudgeReply(request, surface, "[0]", "actual-judge", stream);
                var parsed = await send;
                Check(parsed.Events.Any(fact => fact.Kind == EventKind.ResponseEnded), "Native judge response failed: " + surface + "/" + stream + ": " + string.Join("; ", parsed.Events.Select(fact => fact.Error)));
                var after = engine.Meter();
                Check(after.Totals.Output == before.Totals.Output + 2
                    && after.Rows.Any(row => row.Identity.Model == "actual-judge" && row.Counts.Output == 2)
                    && Equals(after.Latest, before.Latest), "Auxiliary meter identity/latest isolation");
                Check(ReferenceEquals(agent.Configuration(), config) && agent.History().Snapshot().Sequence == sequence,
                    "Judge changed running configuration or History");
                string[] lensAfter = Directory.GetFiles(lensPath, "*", SearchOption.AllDirectories)
                    .Order().Select(path => path + ":" + Convert.ToHexString(System.Security.Cryptography.SHA256.HashData(File.ReadAllBytes(path)))).ToArray();
                Check(lensBefore.SequenceEqual(lensAfter), "Judge changed main cache lens");
                var unnamed = engine.JudgeRecallAsync(judge, "judge", "data", default);
                var unnamedRequest = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
                var unnamedBody = JudgeReply(surface, "[]", "omitted");
                unnamedBody.Remove(surface == Surface.GenerateContent ? "modelVersion" : "model");
                await Reply(unnamedRequest, unnamedBody.ToJsonString());
                var unnamedResponse = await unnamed;
                Check(unnamedResponse.Billing?.Model == ""
                    && unnamedResponse.Events.Single(fact => fact.Kind == EventKind.ResponseEnded).From?.Model == judge,
                    "Unknown billing identity or frozen judge provenance was lost");

            }
            finally { app.Destroy(); }
        }
    }

    public static async Task RecallJudgeSelection()
    {
        foreach (string verdict in new[] { "[]", "[1]", "broken", "truncated", "http-error", "tool" })
        {
            using var listener = Listen();
            var app = new Application(TextWriter.Null);
            try
            {
                var agent = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4o-mini", listener.Prefixes.Single(), disableStreaming: true));
                var context = new Context([
                    new Entry(1, Actor.System, [new TextPart("SYSTEM_SECRET")]),
                    new Entry(2, Actor.Human, [new TextPart("OLD_VISIBLE")]),
                    new Entry(3, Actor.Tool, [new TextPart("TOOL_SECRET")]),
                    new Entry(4, Actor.Human, [new RecallPart(new RecallAttachment("RECALL_SECRET", [new string('a', 64)]))]),
                    new Entry(5, Actor.Agent, [new HandoffPart("HANDOFF_SECRET")]),
                    new Entry(6, Actor.Agent, [new TextPart("RECENT_VISIBLE")])]);
                var work = agent.Recaller().SelectAsync(JudgeHits(), "latest query", context,
                    new RecallSettings(minQueryLength: 1, candidates: 2, contextMessages: 1, maxSnippets: 2));
                var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
                using var reader = new StreamReader(request.Request.InputStream);
                string wire = await reader.ReadToEndAsync();
                Check(wire.Contains("RECENT_VISIBLE") && wire.Contains("latest query")
                    && !new[] { "SYSTEM_SECRET", "OLD_VISIBLE", "TOOL_SECRET", "RECALL_SECRET", "HANDOFF_SECRET", "MEMORY_FACT" }.Any(wire.Contains),
                    "Judge context selection or shortlist isolation");
                if (verdict == "http-error") await Reply(request, "{}", 503);
                else if (verdict == "tool")
                {
                    var body = JudgeReply(Surface.ChatCompletions, "[]", "judge");
                    body["choices"]![0]!["message"]!["tool_calls"] = new JsonArray(new JsonObject
                    {
                        ["id"] = "not-authorized", ["type"] = "function", ["function"] = new JsonObject { ["name"] = "read_file", ["arguments"] = "{}" }
                    });
                    await Reply(request, body.ToJsonString());
                }
                else await SendJudgeReply(request, Surface.ChatCompletions, verdict == "truncated" ? "[]" : verdict,
                    "judge", complete: verdict != "truncated");
                var result = await work.WaitAsync(TimeSpan.FromSeconds(5));
                if (verdict == "[]") Check(result is null, "Empty verdict refilled");
                else if (verdict == "[1]") Check(result is not null && result.Hashes.Length == 1 && result.Text.Contains("SECOND_DOC")
                    && !result.Text.Contains("MEMORY_FACT"), "Valid verdict incorrectly used quotas");
                else Check(result is not null && result.Text.Contains("MEMORY_FACT") && result.Text.Contains("FIRST_DOC"),
                    "Failure lost full-pool quota fallback: " + verdict);
                Check(agent.History().Snapshot().Sequence == 0 && agent.Children().Length == 0, "Judge created History or child");
                Check(agent.Engine().Meter().Latest is null, "First auxiliary send became latest");
            }
            finally { app.Destroy(); }
        }
    }

    public static async Task RecallJudgeCancellation()
    {
        foreach (bool callerCancellation in new[] { true, false })
        {
            using var listener = Listen();
            using var cancel = new CancellationTokenSource();
            var app = new Application(TextWriter.Null);
            HttpListenerResponse? held = null;
            try
            {
                var agent = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4o-mini", listener.Prefixes.Single()));
                var work = agent.Recaller().SelectAsync(JudgeHits(), "latest query", new Context(),
                    new RecallSettings(minQueryLength: 1, candidates: 2, maxSnippets: 2), cancel.Token);
                var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
                using var reader = new StreamReader(request.Request.InputStream);
                await reader.ReadToEndAsync();
                held = request.Response;
                held.ContentType = "text/event-stream";
                held.SendChunked = true;
                await held.OutputStream.WriteAsync(Encoding.UTF8.GetBytes("data: {"));
                await held.OutputStream.FlushAsync();
                if (callerCancellation) cancel.Cancel();
                try
                {
                    var result = await work.WaitAsync(TimeSpan.FromSeconds(25));
                    Check(!callerCancellation && result is not null && result.Text.Contains("MEMORY_FACT"),
                        "Deadline did not produce full-pool fallback");
                }
                catch (OperationCanceledException) when (callerCancellation) { }
                // The peer remains open until here: completion depended on the
                // HTTP read's cancellation, not EOF or disposing the fake server.
                Check(work.IsCompleted && agent.Engine().Meter().Rows.Length == 0,
                    "Canceled partial response was billed or HTTP work remained live");
                Check(agent.History().Snapshot().Sequence == 0, "Cancellation published judge facts");
            }
            finally
            {
                held?.Abort();
                app.Destroy();
            }
        }
    }

    public static async Task RecallJudgeTerminal()
    {
        foreach (Surface surface in Enum.GetValues<Surface>())
        foreach (bool stream in new[] { false, true })
        {
            using var listener = Listen();
            var app = new Application(TextWriter.Null);
            try
            {
                Vendor vendor = surface == Surface.Messages ? Vendor.Anthropic
                    : surface == Surface.GenerateContent ? Vendor.Gemini : Vendor.OpenAI;
                var agent = new Agent(app, new Configuration(vendor, "judge-route", listener.Prefixes.Single(),
                    disableStreaming: !stream, surface: surface));
                var task = agent.Engine().JudgeRecallAsync("", "judge", "data", default);
                var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(5));
                await SendJudgeReply(request, surface, "[]", "actual-judge", stream, complete: false);
                var result = await task;
                Check(result.Events.Any(fact => fact.Kind == EventKind.ErrorOccurred)
                    && !result.Events.Any(fact => fact.Kind == EventKind.ResponseEnded),
                    "Truncated valid-looking array became success: " + surface + "/" + stream);
                var meter = agent.Engine().Meter();
                // Responses failed/incomplete terminal events have never been
                // successful responses. The other protocols report completed
                // usage even when their finish reason is an output-length cap.
                Check(meter.Totals.Output == (surface == Surface.Responses ? 0 : 2)
                    && meter.Latest is null, "Truncation accounting changed: " + surface);
            }
            finally { app.Destroy(); }
        }
    }

    public static async Task RecallJudgeMechanical()
    {
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.OpenAI, "unknown", "http://127.0.0.1:1"));
            var result = await agent.Recaller().SelectAsync(JudgeHits(), "query", new Context(),
                new RecallSettings(disableJudge: true, minQueryLength: 1, minScore: 0.5, maxSnippets: 2));
            Check(result is not null && result.Text.Contains("MEMORY_FACT"), "Disabled judge lost mechanical quota");
            foreach (var settings in new[] { new RecallSettings(disabled: true), new RecallSettings(), new RecallSettings(minQueryLength: 1) })
            {
                string query = settings.MinQueryLength == 1 ? "the and" : "short";
                Check(await agent.Recaller().SelectAsync(JudgeHits(), query, new Context(), settings) is null, "Admission guard searched");
            }
            Check(agent.Engine().Meter().Rows.Length == 0, "Mechanical selection sent request");
        }
        finally { app.Destroy(); }
    }
}
