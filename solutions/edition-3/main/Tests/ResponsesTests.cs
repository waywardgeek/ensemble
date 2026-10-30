using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class ResponsesTests
{
    private static void Check(bool condition, string reason)
    {
        if (!condition) throw new Exception(reason);
    }
    private static JsonElement J(string text) => JsonSerializer.Deserialize<JsonElement>(text);
    private static string Frame(JsonObject value) => "data: " + value.ToJsonString() + "\n\n";
    private static JsonObject Text(string text) => new()
    {
        ["type"] = "message", ["role"] = "assistant", ["status"] = "completed",
        ["content"] = new JsonArray(new JsonObject { ["type"] = "output_text", ["text"] = text })
    };
    private static JsonObject Call(string args = "{}") => new()
    {
        ["type"] = "function_call", ["id"] = "fc-private", ["call_id"] = "call-0",
        ["name"] = "effect", ["arguments"] = args, ["status"] = "completed"
    };
    private static JsonObject Summary(string text) => new()
    {
        ["type"] = "reasoning", ["id"] = "rs-private", ["encrypted_content"] = "opaque-replay",
        ["summary"] = new JsonArray(new JsonObject { ["type"] = "summary_text", ["text"] = text })
    };
    private static JsonObject Body(params JsonObject[] items) => new()
    {
        ["status"] = "completed", ["model"] = "gpt-4o-mini", ["service_tier"] = "default",
        ["output"] = new JsonArray(items.Select(item => (JsonNode)item.DeepClone()).ToArray()),
        ["usage"] = new JsonObject
        {
            ["input_tokens"] = 100, ["input_tokens_details"] = new JsonObject { ["cached_tokens"] = 40 },
            ["output_tokens"] = 20
        }
    };
    private static string Terminal(JsonObject body) => Frame(new JsonObject
    {
        ["type"] = "response.completed", ["response"] = body.DeepClone()
    });
    private static string Item(string kind, int index, JsonObject item) => Frame(new JsonObject
    {
        ["type"] = "response.output_item." + kind, ["output_index"] = index, ["item"] = item.DeepClone()
    });
    private static HttpResponseMessage Response(string body, bool stream = true) => new(HttpStatusCode.OK)
    {
        Content = new StringContent(body, Encoding.UTF8, stream ? "text/event-stream" : "application/json")
    };
    private static Configuration Config(string url = "http://127.0.0.1:1", bool stream = true,
        BillingMode billing = BillingMode.ApiKey, bool explain = false) => new(Vendor.OpenAI,
            "gpt-4o-mini", url, billing == BillingMode.ApiKey ? "dummy-api-key" : "",
            primaryInstructions: "Primary once", requireToolExplanation: explain, disableStreaming: !stream,
            surface: Surface.Responses, billing: billing, account: billing == BillingMode.ApiKey ? "" : "local-account");
    private static void Refuses(Action action)
    {
        try { action(); }
        catch (ArgumentException) { return; }
        throw new Exception("Expected configuration refusal");
    }

    public static void Projection()
    {
        Refuses(() => new Configuration(Vendor.Anthropic, "x", "http://127.0.0.1", surface: Surface.Responses));
        Refuses(() => new Configuration(Vendor.OpenAI, "x", "http://127.0.0.1", billing: BillingMode.ChatGptPlan, account: "a"));
        Refuses(() => new Configuration(Vendor.OpenAI, "x", "http://127.0.0.1", "key", surface: Surface.Responses,
            billing: BillingMode.ChatGptPlan, account: "a"));
        Refuses(() => Config(stream: false, billing: BillingMode.ChatGptPlan));
        Refuses(() => Config("http://example.invalid", billing: BillingMode.ChatGptPlan));
        Refuses(() => Config("https://user@example.invalid", billing: BillingMode.ChatGptPlan));
        Refuses(() => Config("https://example.invalid?secret=x", billing: BillingMode.ChatGptPlan));
        Refuses(() => Config("https://example.invalid#fragment", billing: BillingMode.ChatGptPlan));
        Refuses(() => new Configuration(Vendor.OpenAI, "x", "http://127.0.0.1", account: "a"));
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, Config());
            _ = new Tool(agent, "effect", "A tool", J("{\"type\":\"object\"}"),
                (job, args, token) => Task.FromResult("ok"));
            var origin = agent.Configuration().Target();
            var context = new Context([
                new Entry(1, Actor.Human, [new TextPart("question")]),
                new Entry(2, Actor.Agent, [new OpaquePart(origin, J(Summary("summary").ToJsonString())),
                    new ToolCallPart("call-0", origin, "effect", J("{\"n\":9007199254740993}"), J(Call("{\"n\":9007199254740993}").ToJsonString()))]),
                new Entry(3, Actor.Tool, [new ToolResultPart("call-0", [new TextPart("result")])])]);
            var wire = agent.Engine().Render(context);
            string json = wire.ToJsonString();
            Check(wire["messages"] is null && wire["instructions"]!.GetValue<string>().Split("Primary once").Length == 2, "Responses instructions route");
            Check(wire["store"]!.GetValue<bool>() == false
                && wire["include"]![0]!.GetValue<string>() == "reasoning.encrypted_content", "stateless replay support");
            Check(wire["stream"]!.GetValue<bool>() && wire["stream_options"] is null, "Responses streaming shape");
            Check(wire["tools"]![0]!["type"]!.GetValue<string>() == "function", "flat API functions");
            Check(json.Contains("9007199254740993") && json.Contains("function_call_output") && json.Contains("opaque-replay"), "exact parts and replay");
            var old = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-4o-mini", "http://127.0.0.1"));
            Check(!old.Engine().Render(context).ToJsonString().Contains("opaque-replay"), "Responses opaque crossed surface");
            var plan = new Agent(app, Config(billing: BillingMode.ChatGptPlan));
            _ = new Tool(plan, "effect", "A tool", J("{\"type\":\"object\"}"), (job, args, token) => Task.FromResult("ok"));
            var planWire = plan.Engine().Render(context);
            Check(planWire["tools"]![0]!["type"]!.GetValue<string>() == "namespace", "plan namespace");
            Check(planWire["input"]![2]!["namespace"]!.GetValue<string>() == "ensemble", "plan call namespace");
            string saved = plan.Engine().SavedConfiguration(new Context()).GetRawText();
            Check(!saved.Contains("local-account") && !saved.Contains("dummy-api-key") && saved.Contains("ChatGptPlan"), "saved config is credential-free");
            var capture = CacheLens.Capture(Vendor.OpenAI, Encoding.UTF8.GetBytes(planWire.ToJsonString()));
            string prompt = Encoding.UTF8.GetString(capture.Prompt);
            Check(prompt.IndexOf("tools", StringComparison.Ordinal) < prompt.IndexOf("instructions", StringComparison.Ordinal)
                && prompt.Contains("question"), "lens uses Responses prompt order");
        }
        finally { app.Destroy(); }
    }

    public static void RedactedResults()
    {
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, Config());
            var call = new Entry(1, Actor.Agent,
                [new ToolCallPart("redacted-call", agent.Configuration().Target(), "effect", J("{}"))]);
            var context = new Context([call,
                new Entry(2, Actor.Tool, [new ToolResultPart("redacted-call", [new TextPart("result")])])]);

            // Use the same reducer as the automatic tool-result budget. Its marker
            // is readable text evidence, not a request to send unsupported media.
            var redacted = History.Reduce(context, new Event(EventKind.Redacted,
                redactFrom: 2, redactTo: 2, level: Redaction.Result, reason: "tool result budget"));
            var wire = agent.Engine().Render(redacted);
            var output = wire["input"]!.AsArray().Single(item =>
                item!["type"]!.GetValue<string>() == "function_call_output")!;
            Check(output["call_id"]!.GetValue<string>() == "redacted-call", "redacted result retains call pairing");
            Check(output["output"]!.GetValue<string>() == "[redacted text: 6 bytes; tool result budget]",
                "Responses renders the actual redaction marker");
            Check(((ToolResultPart)redacted.Dialogue[1].Parts[0]).Parts[0] is RedactedPart,
                "wire projection does not mutate the redacted context");
            Check(((ToolResultPart)context.Dialogue[1].Parts[0]).Parts[0] is TextPart { Text: "result" },
                "redaction leaves its original input context intact");

            // Job-backed markers carry a private locator for later retrieval. The
            // normal text formatter must not leak that locator into provider input.
            var marker = new RedactedPart("tool_result", 12, "budget",
                new BlobRef(RefKind.Path, "/not-opened/private-job-output"));
            var mixed = new Context([call, new Entry(2, Actor.Tool,
                [new ToolResultPart("redacted-call", [new TextPart("left"), marker, new TextPart("right")])])]);
            string mixedJson = agent.Engine().Render(mixed).ToJsonString();
            Check(mixedJson.Contains("left[redacted tool_result: 12 bytes; budget]right")
                && !mixedJson.Contains("private-job-output"), "mixed text preserves ordering without exposing locator");

            var empty = new Context([call, new Entry(2, Actor.Tool, [new ToolResultPart("redacted-call", [])])]);
            var emptyWire = agent.Engine().Render(empty);
            Check(emptyWire["input"]!.AsArray().Single(item =>
                item!["type"]!.GetValue<string>() == "function_call_output")!["output"]!.GetValue<string>() == "",
                "a genuinely empty result remains empty");

            // This repair must not silently discard actual media on a surface
            // whose tool-media projection has not been implemented.
            var image = new BlobPart("image/png", new BlobRef(RefKind.Uri,
                "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+lA3sAAAAASUVORK5CYII="));
            var media = new Context([call, new Entry(2, Actor.Tool, [new ToolResultPart("redacted-call", [image])])]);
            bool refused = false;
            try
            {
                agent.Engine().Render(media);
            }
            catch (NotSupportedException)
            {
                refused = true;
            }
            Check(refused, "real Responses tool media remains explicitly unsupported");
        }
        finally
        {
            app.Destroy();
        }
    }

    public static void Plain()
    {
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, Config());
            var facts = agent.Engine().Parse(200, Body(Summary("visible summary"), Text(""), Call("{\"n\":9007199254740993}")).ToJsonString());
            var final = facts.Single(f => f.Kind == EventKind.ResponseEnded);
            Check(final.Parts.Length == 3 && final.Parts[1] is TextPart { Text: "" }, "empty text retained");
            Check(final.Usage!.Input == 60 && final.Usage.CacheRead == 40 && final.Usage.Output == 20, "disjoint usage");
            var call = (ToolCallPart)final.Parts[2];
            Check(call.CallID == "call-0" && call.Args.GetRawText().Contains("9007199254740993") && call.Opaque is not null, "call precision and replay");
            string encoded = JsonCodec.WriteEvent(new Event(final.Kind, seq: 1, actor: final.Actor,
                parts: final.Parts, from: final.From, usage: final.Usage));
            // Event codec round trip must retain the new surface and native item.
            var decoded = JsonCodec.ReadEvent(encoded);
            Check(JsonCodec.WriteEvent(decoded) == encoded, "Responses provenance round trip");
            foreach (var bad in new[]
            {
                Body(Call("[]")), Body(Call("{")), Body(Call(), Call()),
                new JsonObject { ["status"] = "incomplete", ["output"] = new JsonArray(Call()) },
                new JsonObject { ["status"] = "failed", ["output"] = new JsonArray(Text("bad")) }
            })
            {
                Check(agent.Engine().Parse(200, bad.ToJsonString()).Single().Kind == EventKind.ErrorOccurred, "invalid plain response finalized");
            }
            var unknown = Body(Text("ok"));
            unknown["usage"]!["input_tokens_details"] = "malformed";
            var usage = agent.Engine().Parse(200, unknown.ToJsonString()).Last().Usage!;
            Check(usage.Input is null && usage.CacheRead is null && usage.Output == 20, "malformed cache operand became zero");
            Check(agent.Engine().Meter().Latest is null && agent.History().Events().IsEmpty, "offline decode mutated owners");
        }
        finally { app.Destroy(); }
    }

    public static async Task Streaming()
    {
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, Config());
            string delta = Frame(new JsonObject { ["type"] = "response.reasoning_summary_text.delta", ["output_index"] = 0,
                ["summary_index"] = 0, ["delta"] = "first" });
            string stream = Item("added", 0, Summary("")) + delta + Item("done", 0, Summary("first suffix"))
                + Item("done", 1, Text("answer")) + Item("done", 2, Call()) + Terminal(Body());
            var plainChunks = new List<Chunk>();
            using (var plain = Response(Body(Summary("plain summary"), Text("plain answer")).ToJsonString(), false))
            {
                var decoded = await agent.Engine().ParseAsync(plain, plainChunks.Add);
                var parts = decoded.Events.Last().Parts;
                Check(plainChunks.Count == 0 && parts[0] is OpaquePart
                    && parts[1] is TextPart { Text: "plain answer" },
                    "plain responses must retain separate finals, not fabricate deltas");
            }
            var chunks = new List<Chunk>();
            using var response = Response(stream);
            var parsed = await agent.Engine().ParseAsync(response, chunks.Add);
            var final = parsed.Events.Single(f => f.Kind == EventKind.ResponseEnded);
            Check(final.Parts.Length == 3 && final.Parts[2] is ToolCallPart, "empty terminal lost completed items");
            Check(chunks.Where(c => c.Kind == ChunkKind.ReasoningSummary).Select(c => c.Text).Aggregate("", (a,b) => a+b) == "first suffix", "summary duplicated or final suffix lost");
            Check(chunks.Where(c => c.Kind == ChunkKind.ReasoningSummary).Select(c => c.PartID).Distinct().Count() == 1, "summary identity changed");
            Check(parsed.PartIDs.Distinct().Count() == 3, "response-local identities collide");
            Check(!final.Parts.OfType<TextPart>().Any(p => p.Text.Contains("suffix")), "summary became answer");
            foreach (string? mediaType in new string?[] { null, "text/plain" })
            {
                using var unlabelled = Response(stream);
                unlabelled.Content.Headers.ContentType = mediaType is null ? null
                    : new System.Net.Http.Headers.MediaTypeHeaderValue(mediaType);
                var delivered = new List<Chunk>();
                var decoded = await agent.Engine().ParseAsync(unlabelled, delivered.Add);
                Check(decoded.Events.Last().Kind == EventKind.ResponseEnded
                    && decoded.Events.Last().Parts.Length == 3
                    && delivered.Any(chunk => chunk.Kind == ChunkKind.ReasoningSummary),
                    "headerless/text-plain Responses lost incremental parsing or completed items");
                using var failedStream = Response(Item("done", 0, Call()));
                failedStream.Content.Headers.ContentType = unlabelled.Content.Headers.ContentType;
                var refused = await agent.Engine().ParseAsync(failedStream);
                Check(refused.Events.Single().Kind == EventKind.ErrorOccurred,
                    "headerless Responses bypassed terminal-success gate");
            }
            var strict = new Agent(app, Config(explain: true));
            using (var explained = Response(Item("done", 0, Text("I will inspect it."))
                + Item("done", 1, Call()) + Terminal(Body())))
            {
                var ordered = await strict.Engine().ParseAsync(explained);
                Check(!ordered.ToolBeforeText && ordered.Events.Last().Kind == EventKind.ResponseEnded,
                    "completed visible explanation was ignored");
            }
            using (var late = Response(Item("done", 1, Call())
                + Item("done", 0, Text("too late")) + Terminal(Body())))
            {
                var ordered = await strict.Engine().ParseAsync(late);
                Check(ordered.ToolBeforeText, "final array order concealed an early tool fragment");
            }
            using var authoritative = Response(Item("done", 0, Text("old")) + Terminal(Body(Text("new"))));
            var replacement = await agent.Engine().ParseAsync(authoritative);
            Check(((TextPart)replacement.Events.Last().Parts.Single()).Text == "new", "nonempty terminal was not authoritative");
            foreach (string bad in new[]
            {
                Item("done", 0, Call()),
                Item("added", 0, Call()) + Terminal(Body()),
                Item("done", 0, Call("{")) + Terminal(Body()),
                Item("done", 0, Call()) + Frame(new JsonObject { ["type"] = "response.failed", ["secret"] = "private-provider-error" }),
                Terminal(Body()) + delta,
                "data: [DONE]\n\n",
                "data: {bad}\n\n"
            })
            {
                using var invalid = Response(bad);
                var failed = await agent.Engine().ParseAsync(invalid);
                Check(failed.Events.Single().Kind == EventKind.ErrorOccurred, "failed stream finalized");
                Check(!JsonCodec.WriteEvent(failed.Events.Single()).Contains("private-provider-error"), "provider detail leaked");
            }
        }
        finally { app.Destroy(); }
    }

    private static HttpListener Listen()
    {
        var reservation = new TcpListener(IPAddress.Loopback, 0);
        reservation.Start();
        int port = ((IPEndPoint)reservation.LocalEndpoint).Port;
        reservation.Stop();
        var listener = new HttpListener();
        listener.Prefixes.Add($"http://127.0.0.1:{port}/");
        listener.Start();
        return listener;
    }
    private static async Task Reply(HttpListenerContext request, string body, bool stream)
    {
        byte[] bytes = Encoding.UTF8.GetBytes(body);
        request.Response.ContentType = stream ? "text/event-stream" : "application/json";
        request.Response.ContentLength64 = bytes.Length;
        await request.Response.OutputStream.WriteAsync(bytes);
        request.Response.Close();
    }

    public static async Task Transport(string root)
    {
        using var timeout = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        using var server = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, Config(server.Prefixes.Single()));
            string capture = Path.Combine(root, "responses-lens");
            _ = new CacheLens(agent.Engine(), capture);
            string wire = agent.Engine().Render(new Context([new Entry(1, Actor.Human, [new TextPart("hello")])])).ToJsonString();
            for (int i = 0; i < 2; i++)
            {
                var sending = agent.Engine().SendAsync(wire, timeout.Token);
                var request = await server.GetContextAsync().WaitAsync(timeout.Token);
                using var reader = new StreamReader(request.Request.InputStream);
                Check(await reader.ReadToEndAsync(timeout.Token) == wire, "wire changed during send");
                Check(request.Request.Url!.AbsolutePath == "/v1/responses" && request.Request.Headers["Authorization"] == "Bearer dummy-api-key", "route or API authorization");
                await Reply(request, i == 0 ? Body(Text("plain")).ToJsonString() : Terminal(Body(Text("stream"))), i != 0);
                var result = await sending;
                Check(result.Events.Any(f => f.Kind == EventKind.ResponseEnded), "real response missing");
            }
            var meter = agent.Engine().Meter();
            Check(meter.Totals.Input == 120 && meter.Totals.CacheRead == 80 && meter.Totals.Output == 40 && meter.EstimatedUsd() is > 0, "actual completions counted once");
            Check(File.ReadAllText(Path.Combine(capture, "request.wire.json")) == wire, "lens wire mismatch");
            Check(!Directory.GetFiles(capture).Any(p => File.ReadAllText(p).Contains("dummy-api-key")), "lens captured bearer");
        }
        finally { app.Destroy(); }
    }

    public static async Task Plan(string root)
    {
        using var timeout = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        using var server = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, Config(server.Prefixes.Single(), billing: BillingMode.ChatGptPlan));
            bool refused = false;
            try { await agent.Engine().SendAsync("{}", timeout.Token); }
            catch (InvalidOperationException) { refused = true; }
            Check(refused, "missing credentials fell back");
            var store = new Credentials(app, Path.Combine(root, "responses-auth"));
            await store.StoreAsync(new AuthRegistration("local-account", "https://auth.openai.com", "subject", "issued-client", null,
                "dummy-plan-access", "dummy-plan-refresh", "dummy-id-token", DateTimeOffset.UtcNow.AddHours(1), "openid chatgpt.tokens.use.direct"));
            var sending = agent.Engine().SendAsync("{}", timeout.Token);
            var request = await server.GetContextAsync().WaitAsync(timeout.Token);
            Check(request.Request.Headers["Authorization"] == "Bearer dummy-plan-access", "plan token not obtained through owner");
            await Reply(request, Terminal(Body(Text("plan"))), true);
            await sending;
            Check(agent.Engine().Meter().Totals.Output == 20 && agent.Engine().Meter().EstimatedUsd() is null, "plan usage received API dollars");
            Check(agent.Engine().Meter().Latest!.Identity.Billing == BillingMode.ChatGptPlan, "meter lost billing identity");
            Check(agent.Engine().SavedConfiguration(new Context()).GetRawText().Contains("ChatGptPlan") && agent.History().Events().IsEmpty, "private token entered History");
            sending = agent.Engine().SendAsync("{}", timeout.Token);
            request = await server.GetContextAsync().WaitAsync(timeout.Token);
            request.Response.StatusCode = 307;
            request.Response.RedirectLocation = server.Prefixes.Single() + "must-not-follow";
            request.Response.ContentLength64 = 0;
            request.Response.Close();
            var redirected = await sending;
            Check(redirected.Events.Single().Kind == EventKind.ErrorOccurred
                && agent.Engine().Meter().Totals.Output == 20, "plan followed redirect or metered failure");
            await store.LogoutAsync("local-account");
            refused = false;
            try { await agent.Engine().SendAsync("{}", timeout.Token); }
            catch (InvalidOperationException) { refused = true; }
            Check(refused && agent.Engine().Meter().Totals.Output == 20, "logout fallback or failed-send accounting");
            using var cancelled = new CancellationTokenSource();
            cancelled.Cancel();
            bool cancellationPreserved = false;
            try { await agent.Engine().SendAsync("{}", cancelled.Token); }
            catch (OperationCanceledException) { cancellationPreserved = true; }
            Check(cancellationPreserved && agent.History().Events().IsEmpty,
                "cancelled credential lookup became an auth fact");

            bool actorRefused = false;
            try { await agent.SendAsync("signed-out request", timeout.Token); }
            catch (PlanAuthorizationException) { actorRefused = true; }
            var facts = agent.History().Events();
            var errors = facts.Where(fact => fact.Kind == EventKind.ErrorOccurred).ToArray();
            Check(actorRefused && errors.Length == 1
                && errors[0].Error!.Contains("No API-key fallback")
                && facts.Count(fact => fact.Kind == EventKind.TurnEnded) == 1
                && !facts.Any(fact => fact.Kind == EventKind.ResponseEnded),
                "auth refusal did not record exactly one failed turn and durable error");
            Check(agent.Engine().Meter().Totals.Output == 20, "refusal advanced usage");
        }
        finally { app.Destroy(); }
    }

    public static void ReasoningPolicy()
    {
        var app = new Application(TextWriter.Null);
        try
        {
            var selected = new Configuration(Vendor.OpenAI, "gpt-6-astra", "http://127.0.0.1:1");
            Check(selected.Surface == Surface.Responses && selected.RequireToolExplanation,
                "selected model did not choose Responses and explanation policy");
            var agent = new Agent(app, selected);
            _ = new Tool(agent, "effect", "A real tool declaration", J("{\"type\":\"object\"}"),
                (job, args, token) => Task.FromResult("ok"));
            var wire = agent.Engine().Render(new Context());
            Check(wire["reasoning"]?["effort"]?.GetValue<string>() == "medium"
                && wire["reasoning"]?["summary"]?.GetValue<string>() == "auto", "model defaults missing");
            Check(wire["instructions"]!.GetValue<string>().Contains("Before each tool batch"), "explanation rule not model-facing");
            var saved = agent.Engine().SavedConfiguration(new Context());
            Check(saved.GetProperty("reasoning_effort").GetString() == "medium"
                && saved.GetProperty("reasoning_summary").GetString() == "auto"
                && saved.GetProperty("require_tool_explanation").GetBoolean(), "saved policy mismatch");
            foreach (string effort in new[] { "low", "medium", "high", "xhigh", "max" })
            {
                var other = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-6-astra", "http://127.0.0.1:1",
                    reasoningEffort: effort, reasoningSummary: "off"));
                var request = other.Engine().Render(new Context());
                Check(request["reasoning"]?["effort"]?.GetValue<string>() == effort
                    && request["reasoning"]?["summary"] is null, "explicit option not respected");
                other.Destroy();
            }
            foreach (string model in new[] { "unknown-reasoning-model", "gpt-6-astra-other", "gpt-4o-mini" })
            {
                var other = new Agent(app, new Configuration(Vendor.OpenAI, model, "http://127.0.0.1:1", surface: Surface.Responses));
                Check(other.Engine().Render(new Context())["reasoning"] is null && !other.Configuration().RequireToolExplanation,
                    "another Agent inherited reasoning policy");
                other.Destroy();
                Refuses(() => new Configuration(Vendor.OpenAI, model, "http://127.0.0.1:1",
                    surface: Surface.Responses, reasoningEffort: "medium"));
                Refuses(() => new Configuration(Vendor.OpenAI, model, "http://127.0.0.1:1",
                    surface: Surface.Responses, reasoningSummary: "auto"));
            }
            Refuses(() => new Configuration(Vendor.OpenAI, "gpt-6-astra", "http://127.0.0.1:1", reasoningEffort: "none"));
            Refuses(() => new Configuration(Vendor.OpenAI, "gpt-6-astra", "http://127.0.0.1:1", reasoningSummary: "detailed"));
            Refuses(() => new Configuration(Vendor.OpenAI, "gpt-6-astra", "http://127.0.0.1:1", surface: Surface.ChatCompletions));
            Check(agent.Engine().Render(new Context()).ToJsonString() == wire.ToJsonString(), "other selections mutated first Agent");
        }
        finally { app.Destroy(); }
    }

    public static async Task ReasoningDispatch()
    {
        using var timeout = new CancellationTokenSource(TimeSpan.FromSeconds(30));
        // Each case uses a real actor/tool and a real HTTP boundary. No mock
        // dispatch expectations substitute for the observable tool effect.
        foreach (bool stream in new[] { false, true })
        foreach (string order in new[] { "summary-only", "late-text", "explanation-first" })
        {
            using var server = Listen();
            var app = new Application(TextWriter.Null);
            try
            {
                var agent = new Agent(app, new Configuration(Vendor.OpenAI, "gpt-6-astra", server.Prefixes.Single(),
                    apiKey: "dummy", disableStreaming: !stream));
                int effects = 0;
                _ = new Tool(agent, "effect", "Observe dispatch", J("{\"type\":\"object\"}"),
                    (job, args, token) => { effects++; return Task.FromResult("worked"); });
                var turn = agent.SendAsync("Use the tool", timeout.Token);
                var request = await server.GetContextAsync().WaitAsync(timeout.Token);
                using (var reader = new StreamReader(request.Request.InputStream))
                {
                    var body = JsonNode.Parse(await reader.ReadToEndAsync(timeout.Token))!;
                    Check(body["reasoning"]?["summary"]?.GetValue<string>() == "auto", "real wire lost reasoning policy");
                }
                JsonObject[] items = order switch
                {
                    "summary-only" => [Summary("I considered the tool."), Call()],
                    "late-text" => [Call(), Text("This explanation arrived too late.")],
                    _ => [Text("I will use the tool to check the result."), Call()]
                };
                var terminal = Body(items);
                terminal["model"] = "gpt-6-astra";
                string payload = terminal.ToJsonString();
                if (stream)
                {
                    // Even a provider-reordered terminal cannot erase the fact
                    // that the call arrived before ordinary assistant text.
                    if (order == "late-text") terminal = Body(Text("Later"), Call());
                    terminal["model"] = "gpt-6-astra";
                    payload = string.Concat(items.Select((item, i) => Item("done", i, item))) + Terminal(terminal);
                }
                await Reply(request, payload, stream);
                var next = await server.GetContextAsync().WaitAsync(timeout.Token);
                using (var reader = new StreamReader(next.Request.InputStream))
                {
                    string body = await reader.ReadToEndAsync(timeout.Token);
                    Check(body.Contains("function_call_output"), "paired tool result missing");
                    if (order == "explanation-first")
                    {
                        Check(body.Contains("worked") && effects == 1, "explained tool did not run");
                    }
                    else
                    {
                        Check(body.Contains("Tool not executed") && effects == 0,
                            "unexplained tool was not paired with a refusal");
                        Check(!agent.History().Events().Any(e => e.Kind == EventKind.ToolStarted),
                            "unexplained tool was admitted");
                    }
                }
                var final = Body(Text("finished")); final["model"] = "gpt-6-astra";
                await Reply(next, stream ? Terminal(final) : final.ToJsonString(), stream);
                Check(await turn == "finished", "paired refusal or successful call did not complete turn");
                Check(agent.Engine().Meter().EstimatedUsd() is null, "unpriced model acquired a price");
            }
            finally { app.Destroy(); }
        }
    }

    public static async Task Dispatch()
    {
        using var timeout = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        using var server = Listen();
        var app = new Application(TextWriter.Null);
        try
        {
            int effects = 0;
            var agent = new Agent(app, Config(server.Prefixes.Single()));
            _ = new Tool(agent, "effect", "Count a real dispatch", J("{\"type\":\"object\"}"),
                (job, args, token) => { Interlocked.Increment(ref effects); return Task.FromResult("worked"); });
            var turn = agent.SendAsync("use the tool", timeout.Token);
            var first = await server.GetContextAsync().WaitAsync(timeout.Token);
            using (var reader = new StreamReader(first.Request.InputStream)) { await reader.ReadToEndAsync(timeout.Token); }
            // End with a failed terminal after a complete call. No effects allowed.
            await Reply(first, Item("done", 0, Call()) + Frame(new JsonObject { ["type"] = "response.failed" }), true);
            try { await turn; } catch (InvalidOperationException) { }
            Check(effects == 0, "failed terminal dispatched a call");
            turn = agent.SendAsync("try again", timeout.Token);
            first = await server.GetContextAsync().WaitAsync(timeout.Token);
            using (var reader = new StreamReader(first.Request.InputStream)) { await reader.ReadToEndAsync(timeout.Token); }
            // Publish a complete tool item, but hold terminal success. Observe
            // the actual stream delivery before asserting that no effect ran.
            long requestID = agent.History().Events().Last(f => f.Kind == EventKind.RequestSent).Seq;
            var observed = agent.WaitAsync(notice => notice.ResponseID == requestID
                && notice.Delta?.Kind == ChunkKind.ToolCall, timeout.Token);
            first.Response.ContentType = "text/event-stream";
            first.Response.SendChunked = true;
            await first.Response.OutputStream.WriteAsync(Encoding.UTF8.GetBytes(Item("done", 0, Call())), timeout.Token);
            await first.Response.OutputStream.FlushAsync(timeout.Token);
            await observed;
            Check(effects == 0, "item-done dispatched before terminal success");
            await first.Response.OutputStream.WriteAsync(Encoding.UTF8.GetBytes(Terminal(Body())), timeout.Token);
            first.Response.Close();
            var second = await server.GetContextAsync().WaitAsync(timeout.Token);
            using (var reader = new StreamReader(second.Request.InputStream))
            {
                string input = await reader.ReadToEndAsync(timeout.Token);
                Check(input.Contains("function_call_output") && input.Contains("worked") && effects == 1, "tool result was not correlated before next send");
            }
            await Reply(second, Terminal(Body(Text("finished"))), true);
            Check(await turn == "finished", "actor answer mismatch");
            Check(effects == 1, "tool dispatched twice");
        }
        finally { app.Destroy(); }
    }
}
