using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class CacheMarkerTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition)
        {
            throw new InvalidOperationException(message);
        }
    }

    private static JsonElement Json(string value) => JsonSerializer.Deserialize<JsonElement>(value);
    private static bool Marked(JsonNode? block) => block?["cache_control"] is not null;
    private static JsonObject[] Blocks(JsonObject request) => request["messages"]!.AsArray()
        .SelectMany(message => message!["content"]!.AsArray()).Cast<JsonObject>().ToArray();

    private static JsonObject WithoutMarkers(JsonObject request)
    {
        var clean = JsonNode.Parse(CacheLens.StripMarkers(Encoding.UTF8.GetBytes(request.ToJsonString())))!.AsObject();
        // Opt-in Anthropic SYSTEM uses a text-block array. Normalize only this
        // required envelope to compare the exact ordinary instruction string.
        if (clean["system"] is JsonArray system)
        {
            clean["system"] = string.Concat(system.Select(block => block!["text"]!.GetValue<string>()));
        }
        return clean;
    }

    public static void Boundaries()
    {
        var app = new Application(TextWriter.Null);
        try
        {
            var marked = new Agent(app, new Configuration(Vendor.Anthropic, "caller-asserted", "http://127.0.0.1:1",
                primaryInstructions: "PRIMARY", cacheBreakpoints: true));
            var ordinary = new Agent(app, new Configuration(Vendor.Anthropic, "caller-asserted", "http://127.0.0.1:1",
                primaryInstructions: "PRIMARY"));
            marked.InstallTools();
            ordinary.InstallTools();
            var call = new ToolCallPart("reused", marked.Configuration().Target(), "read_file",
                Json("{\"path\":\"x\",\"cache_control\":\"user-data\",\"number\":1.00}"));
            var context = new Context([
                new Entry(1, Actor.System, [new HandoffPart("CONTINUE")]),
                new Entry(2, Actor.Human, [new TextPart("first")]),
                new Entry(3, Actor.Agent, [call]),
                new Entry(4, Actor.Human, [new TextPart("interleaved")]),
                new Entry(5, Actor.Tool, [new ToolResultPart(call.CallID, [new TextPart("RESULT")])]),
                new Entry(6, Actor.Agent, [new TextPart("done")]),
                new Entry(7, Actor.Agent, [call]),
                new Entry(8, Actor.Tool, [new ToolResultPart(call.CallID, [new TextPart("SECOND")])])
            ], [new TextPart("TRANSIENT")]);
            var request = marked.Engine().Render(context);
            var blocks = Blocks(request);
            Check(blocks.Count(Marked) == 2, "handoff and final stable result boundaries");
            Check(Marked(blocks.Single(block => block["text"]?.GetValue<string>() == "CONTINUE")), "handoff boundary retained");
            Check(Marked(blocks.Last(block => block["type"]!.GetValue<string>() == "tool_result")), "last result before ephemera marked");
            Check(!Marked(blocks.Last()) && blocks.Last()["text"]!.GetValue<string>() == "TRANSIENT", "ephemera excluded after merge");
            var tools = request["tools"]!.AsArray();
            Check(tools.Count(Marked) == 1 && Marked(tools.Last()), "last eager tool marked");
            Check(Marked(request["system"]![0]), "SYSTEM block marked");
            Check(JsonNode.DeepEquals(WithoutMarkers(request), ordinary.Engine().Render(context)), "non-marker provider content preserved");
            Check(call.Args.GetProperty("number").GetRawText() == "1.00" && call.Args.GetProperty("cache_control").GetString() == "user-data",
                "canonical argument lexemes and user property preserved");
            Check(context.Dialogue.First().Parts.Single() is HandoffPart && context.Ephemeral.Length == 1, "canonical boundary values unchanged");
            Check(marked.Engine().SavedConfiguration(context).GetProperty("system_prompt").GetString()!.Contains("PRIMARY"), "block SYSTEM saved as instruction text");

            var empty = new Agent(app, new Configuration(Vendor.Anthropic, "caller-asserted", "http://127.0.0.1:1", primaryInstructions: "", cacheBreakpoints: true));
            var thinking = new OpaquePart(empty.Configuration().Target(), Json("{\"type\":\"thinking\",\"thinking\":\"private\",\"signature\":\"signed\"}"));
            var tail = new Context([new Entry(1, Actor.Human, [new TextPart("stable")]),
                new Entry(2, Actor.Agent, [thinking, new TextPart("")])], [new TextPart("later")]);
            var tailWire = empty.Engine().Render(tail);
            Check(tailWire["system"]!.GetValue<string>() == "", "empty SYSTEM not marked");
            Check(tailWire["tools"] is null && Blocks(tailWire).Count(Marked) == 1, "no-tool and empty/thinking cases do not manufacture markers");
            Check(Blocks(tailWire).Where(block => block["type"]!.GetValue<string>() == "thinking").All(block => !Marked(block)), "thinking never explicitly marked");
            var noteOnly = empty.Engine().Render(new Context([new Entry(1, Actor.System, [new HandoffPart("one")])]));
            Check(Blocks(noteOnly).Count(Marked) == 1, "coincident handoff/tail marker deduplicated");
            var transientOnly = empty.Engine().Render(new Context(ephemeral: [new TextPart("only transient")]));
            Check(!Blocks(transientOnly).Any(Marked), "transient-only request has no conversation marker");
        }
        finally
        {
            app.Destroy();
        }
    }

    public static async Task Deferred()
    {
        string directory = Path.Combine(Path.GetTempPath(), "cache-skills-" + Guid.NewGuid().ToString("N"));
        var app = new Application(TextWriter.Null);
        try
        {
            Directory.CreateDirectory(Path.Combine(directory, "root"));
            Directory.CreateDirectory(Path.Combine(directory, "edit"));
            File.WriteAllText(Path.Combine(directory, "root", "SKILL.md"),
                "---\nname: root\ntype: primary\ntools: read_file\nloadable-skills: edit\n---\nPRIMARY_MANUAL");
            File.WriteAllText(Path.Combine(directory, "edit", "SKILL.md"),
                "---\nname: edit\ntype: loadable\ntools: write_file\n---\nEDIT_MANUAL");
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "caller-asserted", "http://127.0.0.1:1",
                deferredTools: true, cacheBreakpoints: true), skillSpec: new SkillSpec(directory, "root"));
            void AssertCatalog(Context context)
            {
                var wire = agent.Engine().Render(context);
                var tools = wire["tools"]!.AsArray().Cast<JsonObject>().ToArray();
                Check(tools.Any(tool => tool["defer_loading"]?.GetValue<bool>() == true), "fixture includes deferred tool");
                Check(tools.Where(tool => tool["defer_loading"]?.GetValue<bool>() == true).All(tool => !Marked(tool)), "deferred tools never marked");
                Check(Marked(tools.Last(tool => tool["defer_loading"]?.GetValue<bool>() != true)), "last non-deferred tool marked");
            }
            AssertCatalog(agent.History().View());
            var load = new ToolCallPart("load", agent.Configuration().Target(), "load_skill", Json("{\"name\":\"edit\"}"));
            Check(!(await agent.DispatchAsync(load)).IsError, "real skill activated");
            AssertCatalog(agent.History().View());
            var active = agent.Engine().Render(agent.History().View());
            string instructions = active["system"]![0]!["text"]!.GetValue<string>();
            Check(instructions.Split("PRIMARY_MANUAL").Length == 2 && instructions.Split("EDIT_MANUAL").Length == 2,
                "manuals occur once in SYSTEM");
            // A frozen replay with no initial grants must omit the tool marker,
            // rather than putting it on the last deferred catalog declaration.
            var noEager = new Context([new Entry(1, Actor.System, [new SkillPart("root", "PRIMARY_MANUAL", [], primary: true)])]);
            var allDeferred = agent.Engine().Render(noEager)["tools"]!.AsArray();
            Check(allDeferred.Count > 0 && allDeferred.All(tool => tool!["defer_loading"]!.GetValue<bool>() && !Marked(tool)), "all-deferred catalog omits marker");
        }
        finally
        {
            app.Destroy();
            Directory.Delete(directory, true);
        }
    }

    public static async Task TransportAndOptions()
    {
        var reservation = new TcpListener(IPAddress.Loopback, 0);
        reservation.Start();
        int port = ((IPEndPoint)reservation.LocalEndpoint).Port;
        reservation.Stop();
        using var listener = new HttpListener();
        listener.Prefixes.Add($"http://127.0.0.1:{port}/");
        listener.Start();
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "caller-asserted", $"http://127.0.0.1:{port}", cacheBreakpoints: true));
            var context = new Context([new Entry(1, Actor.Human, [new TextPart("stable")])], [new TextPart("temporary")]);
            string body = agent.Engine().Render(context).ToJsonString();
            var send = agent.Engine().SendAsync(body);
            var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
            using var reader = new StreamReader(request.Request.InputStream);
            string received = await reader.ReadToEndAsync();
            byte[] response = Encoding.UTF8.GetBytes("{\"content\":[{\"type\":\"text\",\"text\":\"ok\"}],\"usage\":{\"input_tokens\":10,\"output_tokens\":1}}");
            request.Response.ContentType = "application/json";
            request.Response.ContentLength64 = response.Length;
            await request.Response.OutputStream.WriteAsync(response);
            request.Response.Close();
            var events = await send;
            Check(received == body && events.Events.Any(fact => fact.Kind == EventKind.ResponseEnded), "actual transport preserves marked request bytes and completes");
            foreach (var vendor in Enum.GetValues<Vendor>())
            {
                var ordinary = new Agent(app, new Configuration(vendor, "unknown", "http://127.0.0.1:1"));
                Check(!ordinary.Engine().Render(context).ToJsonString().Contains("cache_control"), "ordinary models remain marker-free");
            }
            var openai = new Agent(app, new Configuration(Vendor.OpenAI, "unknown", "http://127.0.0.1:1", promptCacheKey: "owned-test-namespace"));
            var openaiWire = openai.Engine().Render(context);
            Check(openaiWire["prompt_cache_key"]!.GetValue<string>() == "owned-test-namespace" && !openaiWire.ToJsonString().Contains("cache_control"), "OpenAI routing key without invented block markers");
            Check(openai.Engine().SavedConfiguration(context).GetProperty("prompt_cache_key").GetString() == "owned-test-namespace", "documentary routing setting retained");
            bool wrongVendor = false;
            try { _ = new Configuration(Vendor.Gemini, "unknown", "http://127.0.0.1:1", cacheBreakpoints: true); }
            catch (ArgumentException) { wrongVendor = true; }
            Check(wrongVendor, "wrong-vendor marker configuration refused");
        }
        finally
        {
            app.Destroy();
            listener.Stop();
        }
    }
}
