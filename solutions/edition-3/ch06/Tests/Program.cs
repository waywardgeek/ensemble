using System.Collections.Immutable;
using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

// A small dependency-free executable suite. Assertions inspect real model state
// and real wire bodies; the only fake is an actual loopback HTTP provider.
var failures = 0;
var passed = 0;
var temp = Path.Combine(Environment.GetEnvironmentVariable("CH02_TEST_DIR") ?? Path.GetTempPath(), "ch02-" + Guid.NewGuid().ToString("N"));
Directory.CreateDirectory(temp);
void CheckCase(string name, Action action)
{
    try
    {
        action();
        passed++;
        Console.WriteLine("PASS " + name);
    }
    catch (Exception ex) { failures++; Console.WriteLine("FAIL " + name + ": " + ex); }
}
async Task AsyncCase(string name, Func<Task> action)
{
    try
    {
        await action();
        passed++;
        Console.WriteLine("PASS " + name);
    }
    catch (Exception ex) { failures++; Console.WriteLine("FAIL " + name + ": " + ex); }
}
static void Check(bool yes, string why)
{
    if (!yes)
    {
        throw new Exception(why);
    }
}
static void Refuses(Action operation)
{
    try
    {
        operation();
    }
    catch (Exception e) when (e is ArgumentException or FormatException or InvalidOperationException or JsonException or NotSupportedException or KeyNotFoundException) { return; }
    throw new Exception("Expected refusal");
}
static JsonElement J(string text) { using var d = JsonDocument.Parse(text); return d.RootElement.Clone(); }
static Provenance Origin(Vendor vendor = Vendor.Anthropic, string model = "model-a") => new Configuration(vendor, model, "http://127.0.0.1").Target();
static ToolCallPart Call(string id = "a", Vendor vendor = Vendor.Anthropic) => new(id, Origin(vendor), "read_file", J("{\"n\":9007199254740993,\"path\":\"x\"}"));
static Agent NewAgent(Application app, Vendor vendor = Vendor.Anthropic, string url = "http://127.0.0.1", string resolved = "model-a") => new(app, new Configuration(vendor, "model-a", url, "dummy-local-only", resolved));
static string Wire(Vendor vendor, Context context, string model = "model-a")
{
    var app = new Application(TextWriter.Null);
    try
    {
        return NewAgent(app, vendor, resolved: model).Engine().Render(context).ToJsonString();
    }
    finally { app.Destroy(); }
}
try
{
    await AsyncCase("Chapter 3 tools and loop", () => ToolTests.Run(temp));
    await AsyncCase("Chapter 4 jobs and PTY", () => JobTests.Run(temp));
    await AsyncCase("Chapter 5 transport isolation", ActorTests.TransportIsolation);
    await AsyncCase("Chapter 5 mailbox controls", ActorTests.MailboxControls);
    await AsyncCase("Chapter 5 busy hint", ActorTests.BusyHint);
    await AsyncCase("Chapter 5 immediate interrupt", ActorTests.ImmediateInterrupt);
    await AsyncCase("Chapter 5 boundary interrupt", ActorTests.BoundaryInterrupt);
    await AsyncCase("Chapter 5 shutdown joins", ActorTests.ShutdownJoins);
    await AsyncCase("Chapter 5 HTTP interrupt boundaries", ActorTests.HttpInterrupts);
    await AsyncCase("Chapter 5 direct tool hint", ActorTests.DirectHint);
    await AsyncCase("Chapter 5 input admission outcomes", ActorTests.InputAdmission);
    await AsyncCase("Chapter 5 reliable flush", ActorTests.FlushTurns);
    await AsyncCase("Chapter 5 unlimited tool batches", ActorTests.UnlimitedBatches);
    await AsyncCase("Chapter 5 local media projection and refusal", ActorTests.MediaProjection);
    await AsyncCase("Chapter 5 named Agents and caller roles", ActorTests.NamedAgents);
    await AsyncCase("Chapter 6 HTTP parser and stream identity", StreamingTests.Parser);
    await AsyncCase("Chapter 6 incremental actor publication", ActorTests.StreamingDelivery);
    await AsyncCase("Chapter 6 streaming tool policy and interrupt boundaries", ActorTests.StreamingToolBoundary);
    CheckCase("construction and complete idempotent destruction", () =>
    {
        var app = new Application(TextWriter.Null);
        var a = NewAgent(app);
        var b = NewAgent(app);
        var history = a.History();
        var engine = a.Engine();
        Check(app.Agents().Length == 2 && ReferenceEquals(a.Parent(), app), "both root links");
        Check(ReferenceEquals(history.Parent(), a) && ReferenceEquals(engine.Parent(), a), "both child links");
        engine.Destroy();
        Refuses(() => a.Engine());
        Refuses(() => engine.Parent());
        _ = new Engine(a);
        a.Destroy();
        a.Destroy();
        Check(app.Agents().Length == 1 && ReferenceEquals(app.Agents()[0], b), "sibling survives removal");
        Refuses(() => history.Parent());
        Refuses(() => a.Parent());
        app.Destroy();
        app.Destroy();
        Check(app.Agents().Length == 0, "root cascades");
        Refuses(() => b.Parent());
        Refuses(() => NewAgent(app));
    });
    CheckCase("each relationship operation completes both ends or changes neither", () =>
    {
        var root = new Application(TextWriter.Null);
        var foreign = new Application(TextWriter.Null);
        var a = NewAgent(root);
        var other = NewAgent(root);
        var h = a.History();
        var e = a.Engine();
        try
        {
            Refuses(() => foreign.RemoveAgent(a));
            Refuses(() => root.AddAgent(a));
            Check(a.Parent() == root && root.Agents().Contains(a), "refused root change leaves both ends");
            root.RemoveAgent(a);
            Check(!root.Agents().Contains(a) && a.parent is null && !a.destroyed, "remove detaches without destroying");
            root.AddAgent(a);
            Check(a.Parent() == root && root.Agents().Contains(a), "add establishes both ends");
            Refuses(() => other.RemoveHistory(h));
            Refuses(() => other.AddHistory(h));
            Check(h.Parent() == a && a.History() == h, "refused history change leaves both ends");
            a.RemoveHistory(h);
            Check(h.parent is null && !h.destroyed, "parent remove clears History back-reference");
            Refuses(() => a.History());
            a.AddHistory(h);
            Check(a.History() == h && h.Parent() == a, "History reattached at both ends");
            Refuses(() => other.RemoveEngine(e));
            Refuses(() => other.AddEngine(e));
            Check(e.Parent() == a && a.Engine() == e, "refused Engine change leaves both ends");
            a.RemoveEngine(e);
            Check(e.parent is null && !e.destroyed, "parent remove clears Engine back-reference");
            Refuses(() => a.Engine());
            a.AddEngine(e);
            Check(a.Engine() == e && e.Parent() == a, "Engine reattached at both ends");
            a.Destroy();
            Check(a.parent is null && h.parent is null && e.parent is null, "Destroy relies on complete removes");
            Refuses(() => root.AddAgent(a));
            Refuses(() => other.AddHistory(h));
            Refuses(() => other.AddEngine(e));
            Check(!root.Agents().Contains(a), "destroyed objects cannot be reattached");
        }
        finally { e.Destroy(); h.Destroy(); a.Destroy(); root.Destroy(); foreign.Destroy(); }
    });
    CheckCase("immutable exact-number values and JSON round trip", () =>
    {
        var call = Call();
        var copy = (ToolCallPart)JsonCodec.ReadPart(J(JsonCodec.WritePart(call).ToJsonString()));
        Check(copy.Args.GetProperty("n").GetInt64() == 9007199254740993, "no double rounding");
        var source = new List<Part> { new TextPart("original") };
        var entry = new Entry(1, Actor.Human, source);
        source.Clear();
        Check(entry.Parts.Length == 1, "immutable sequence copies input");
        Refuses(() => new BlobRef(0, "bad"));
        Refuses(() => new Provenance(0, "x", Surface.Messages));
        Refuses(() => new Event((EventKind)999));
        Refuses(() => JsonCodec.ReadPart(J("{\"type\":\"blob\",\"mime\":\"x\",\"path\":\"old\"}")));
        Refuses(() => JsonCodec.ReadPart(J("{\"type\":\"blob\",\"mime\":\"x\",\"ref\":{\"kind\":\"\",\"locator\":\"x\"}}")));
    });
    CheckCase("all 32 phase/event pairs follow the published transition table", () =>
    {
        Event[] values = [new(EventKind.MessageReceived, actor: Actor.Human), new(EventKind.RequestSent, from: Origin()),
            new(EventKind.ResponseStarted, from: Origin()), new(EventKind.ResponseEnded, from: Origin()),
            new(EventKind.ToolCalled, callID: "a", name: "read_file", args: J("{}")), new(EventKind.ToolReturned, callID: "a"),
            new(EventKind.Redacted, redactFrom: 1, redactTo: 1, level: Redaction.Result), new(EventKind.ErrorOccurred, error: "bad")];
        foreach (var phase in Enum.GetValues<Phase>())
        {
            foreach (var value in values)
            {
                var before = new Context(phase: phase);
                var expected = (phase, value.Kind) switch
                {
                    (Phase.Idle, EventKind.MessageReceived) => Phase.InputPending,
                    (Phase.InputPending, EventKind.RequestSent) => Phase.InFlight,
                    (Phase.InFlight, EventKind.ResponseEnded or EventKind.ErrorOccurred) => Phase.Idle,
                    _ => phase
                };
                Check(History.Reduce(before, value).Phase == expected, phase + "/" + value.Kind);
            }
        }

        var pending = History.Reduce(new Context(phase: Phase.InFlight), new Event(EventKind.ResponseEnded, parts: [Call("a"), Call("b")]));
        Check(pending.Phase == Phase.ToolsPending, "calls enter ToolsPending");
        pending = History.Reduce(pending, new Event(EventKind.ToolReturned, callID: "b"));
        Check(pending.Phase == Phase.ToolsPending, "not all results");
        pending = History.Reduce(pending, new Event(EventKind.ToolReturned, callID: "a"));
        Check(pending.Phase == Phase.InputPending, "last result releases");
    });
    CheckCase("published numeric Ref, legacy surface alias, and optional timestamps", () =>
    {
        for (int kind = 1; kind <= 3; kind++)
        {
            var part = (BlobPart)JsonCodec.ReadPart(J("""{"type":"blob","mime":"image/png","ref":{"kind":KIND,"locator":"kept"}}""".Replace("KIND", kind.ToString())));
            Check((int)part.Ref.Kind == kind && JsonCodec.WritePart(part)["ref"]!["kind"]!.GetValue<int>() == kind, "numeric Ref preserved");
        }
        Refuses(() => JsonCodec.ReadPart(J("""{"type":"blob","mime":"image/png","ref":{"kind":0,"locator":"bad"}}""")));
        Check(JsonCodec.ReadProvenance(J("""{"vendor":"gemini","model":"exact","surface":"generatecontent"}""")).Surface == Surface.GenerateContent, "documented alias reads");
        var e = JsonCodec.ReadEvent("""{"seq":1,"type":"message_received","message":{"actor":"human","parts":[{"type":"text","text":"hi"}]}}""");
        Check(e.Time is null && !JsonCodec.WriteEvent(e).Contains("time"), "absent time stays absent");
    });
    CheckCase("ephemera replaces, clears, stays out of dialogue, and consumes on send", () =>
    {
        var app = new Application(TextWriter.Null);
        try
        {
            var a = NewAgent(app);
            a.SetEphemeral("old");
            a.SetEphemeral("new");
            Check(a.History().View().Dialogue.Length == 0 && JsonCodec.Text(a.History().View().Ephemeral) == "new", "one pending slot");
            a.SetEphemeral("");
            Check(a.History().View().Ephemeral.IsEmpty, "empty clears");
            a.SetEphemeral("once");
            var first = a.Engine().Render(a.History().View()).ToJsonString();
            var second = a.Engine().Render(a.History().View()).ToJsonString();
            Check(first == second && first.Contains("once"), "render does not consume");
            a.History().Append(new Event(EventKind.RequestSent, from: Origin()));
            a.History().Append(new Event(EventKind.ErrorOccurred, error: "failed"));
            Check(!a.Engine().Render(a.History().View()).ToJsonString().Contains("once"), "attempt consumes even on error");
        }
        finally { app.Destroy(); }
    });
    CheckCase("log round trip, deterministic replay, and failed load isolation", () =>
    {
        var app = new Application(TextWriter.Null);
        var path = Path.Combine(temp, "roundtrip.log");
        try
        {
            var a = NewAgent(app);
            a.History().StartLog(path);
            a.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human, parts: [new TextPart("question")]));
            a.History().Append(new Event(EventKind.RequestSent, from: Origin()));
            a.History().Append(new Event(EventKind.ResponseEnded, from: Origin(), parts: [new TextPart("answer")], usage: new Usage(1, 2, 3, 4)));
            var b = NewAgent(app);
            b.History().Load(path);
            Check(a.Engine().Render(a.History().View()).ToJsonString() == b.Engine().Render(b.History().View()).ToJsonString(), "live equals replay");
            using var dumped = new StringWriter();
            b.History().Dump(dumped);
            var dump = dumped.ToString();
            File.WriteAllText(Path.Combine(temp, "dump.log"), dump);
            b.History().Load(Path.Combine(temp, "dump.log"));
            using var again = new StringWriter();
            b.History().Dump(again);
            Check(dump == again.ToString(), "dump stable");
            File.WriteAllText(Path.Combine(temp, "bad.log"), "{\"log_version\":1}\n" + JsonCodec.WriteEvent(new Event(EventKind.MessageReceived, 1, actor: Actor.Human)) + "\n{\"seq\":2,\"type\":\"future\"}\n");
            Refuses(() => b.History().Load(Path.Combine(temp, "bad.log")));
            Check(b.History().Events().Length == 3, "failed load preserved old events");
            var lines = File.ReadAllLines(path);
            File.WriteAllLines(Path.Combine(temp, "headerless.log"), lines.Skip(1));
            b.History().Load(Path.Combine(temp, "headerless.log"));
            Check(b.History().Events().Length == 3, "headerless first event preserved");
            File.WriteAllText(Path.Combine(temp, "future-version.log"), "{\"log_version\":2}\n");
            Refuses(() => b.History().Load(Path.Combine(temp, "future-version.log")));
            Check(b.History().Events().Length == 3, "unknown explicit version refused");
            File.WriteAllLines(Path.Combine(temp, "duplicate.log"), lines.Concat([lines[^1]]));
            Refuses(() => b.History().Load(Path.Combine(temp, "duplicate.log")));
            Check(b.History().Events().Length == 3, "duplicate sequence rejected transactionally");
        }
        finally { app.Destroy(); }
    });
    CheckCase("redaction retains raw facts and exact blob locator without I/O", () =>
    {
        foreach (var kind in Enum.GetValues<RefKind>())
        {
            var reference = new BlobRef(kind, "never-fetch-this-sensitive-location");
            var original = new ToolResultPart("a", [new TextPart("SECRET"), new BlobPart("image/png", reference)], true);
            var view = new Context([new Entry(1, Actor.Agent, [Call()]), new Entry(2, Actor.Tool, [original])]);
            var redacted = History.Reduce(view, new Event(EventKind.Redacted, redactFrom: 2, redactTo: 2, level: Redaction.Result, reason: "test"));
            var result = (ToolResultPart)redacted.Dialogue[1].Parts[0];
            var stub = (RedactedPart)result.Parts[1];
            Check(stub.Ref!.Kind == kind && stub.Ref.Locator == reference.Locator && result.IsError, "ref and error survived reducer");
            Check(original.Parts[0] is TextPart { Text: "SECRET" }, "original untouched");
            foreach (var vendor in Enum.GetValues<Vendor>())
            {
                var wire = Wire(vendor, redacted);
                Check(!wire.Contains("SECRET") && !wire.Contains(reference.Locator) && wire.Contains("read_file"), "only stub renders");
                Refuses(() => Wire(vendor, view));
            }
        }
        var dialogue = new Context([new Entry(1, Actor.Human, [new TextPart("SECRET")]), new Entry(2, Actor.Agent, [Call()])]);
        var textGone = History.Reduce(dialogue, new Event(EventKind.Redacted, redactFrom: 1, redactTo: 2, level: Redaction.Dialogue));
        Check(textGone.Dialogue[1].Parts[0] is ToolCallPart && textGone.Dialogue[0].Parts.IsEmpty, "dialogue keeps tool exchange");
        var toolsGone = History.Reduce(dialogue, new Event(EventKind.Redacted, redactFrom: 2, redactTo: 2, level: Redaction.Tool));
        Check(toolsGone.Dialogue[1].Parts.IsEmpty, "tool removal filter");
        var summary = History.Reduce(dialogue, new Event(EventKind.Redacted, redactFrom: 1, redactTo: 2, level: Redaction.Summary));
        Check(summary.Dialogue.Length == 1 && summary.Dialogue[0].Actor == Actor.System, "range collapses once");
        var replacement = new Event(EventKind.Redacted, seq: 3, parts: [new TextPart("compressed explanation")], redactFrom: 1, redactTo: 2, level: Redaction.Summary);
        var restored = JsonCodec.ReadEvent(JsonCodec.WriteEvent(replacement));
        Check(JsonCodec.Text(History.Reduce(dialogue, restored).Dialogue[0].Parts) == "compressed explanation", "supplied summary survives log and replay");
    });
    CheckCase("interleaved results reordered only in wire and missing result is honest", () =>
    {
        var entries = new[] { new Entry(1, Actor.Agent, [Call("a"), Call("b")]),
            new Entry(2, Actor.Tool, [new ToolResultPart("a", [new TextPart("FIRST")])]),
            new Entry(3, Actor.Human, [new TextPart("HUMAN")]),
            new Entry(4, Actor.Tool, [new ToolResultPart("b", [new TextPart("SECOND")])]) };
        var view = new Context(entries);
        foreach (var vendor in Enum.GetValues<Vendor>())
        {
            var wire = Wire(vendor, view);
            Check(wire.IndexOf("FIRST", StringComparison.Ordinal) < wire.IndexOf("SECOND", StringComparison.Ordinal) && wire.IndexOf("SECOND", StringComparison.Ordinal) < wire.IndexOf("HUMAN", StringComparison.Ordinal), "group results before human");
            var missing = Wire(vendor, new Context(entries.Take(3)));
            Check(missing.Contains("execution outcome unknown") && !missing.Contains("SECOND"), "honest missing-result repair");
            Check(wire.Contains("9007199254740993"), "wire exact number");
        }
        Check(ReferenceEquals(view.Dialogue[2], entries[2]), "canonical order unmodified");
        var repeated = new Context(entries.Concat([new Entry(5, Actor.Agent, [Call("a")]), new Entry(6, Actor.Tool, [new ToolResultPart("a", [new TextPart("LATER")])])]));
        Check(Wire(Vendor.Anthropic, repeated).Contains("LATER"), "IDs can be reused after completion");
        Refuses(() => Wire(Vendor.Anthropic, new Context([new Entry(1, Actor.Agent, [Call(), Call()])])));
        Refuses(() => Wire(Vendor.Anthropic, new Context([new Entry(1, Actor.Tool, [new ToolResultPart("foreign", [])])])));
    });
    CheckCase("opaque exact-model capture/replay and Gemini error objects", () =>
    {
        var app = new Application(TextWriter.Null);
        try
        {
            var a = NewAgent(app, Vendor.Gemini);
            var events = a.Engine().Parse(200, """{"modelVersion":"returned-model","candidates":[{"content":{"parts":[{"functionCall":{"name":"read_file","args":{"n":9007199254740993}},"thoughtSignature":"SIG"},{"functionCall":{"name":"read_file","args":{}}}]}}],"usageMetadata":{"promptTokenCount":4,"candidatesTokenCount":2}}""");
            var calls = events[^1].Parts.OfType<ToolCallPart>().ToArray();
            Check(calls.Length == 2 && calls[0].CallID != calls[1].CallID, "distinct response-local IDs");
            Check(calls[0].From.Model == "returned-model" && calls[0].Opaque!.Value.GetString() == "SIG", "capture returned identity and per-call opaque");
            var view = new Context([new Entry(1, Actor.Agent, [calls[0]]), new Entry(2, Actor.Tool,
                [new ToolResultPart(calls[0].CallID, [new TextPart("{\"message\":\"bad\",\"error\":\"original detail\"}")], true)])]);
            var exact = Wire(Vendor.Gemini, view, "returned-model");
            Check(exact.Contains("SIG") && exact.Contains("\"error\"") && exact.Contains("read_file"), "signature plus JSON error marker and actual name");
            var original = JsonNode.Parse(exact)!["contents"]![1]!["parts"]![0]!["functionResponse"]!["response"]!["error"]!;
            Check(original["message"]!.GetValue<string>() == "bad" && original["error"]!.GetValue<string>() == "original detail", "failure wrapper preserves every original object field");
            Check(!Wire(Vendor.Gemini, view, "different-model").Contains("SIG") && !Wire(Vendor.OpenAI, view).Contains("SIG"), "opaque gated by exact provenance");
            Check(a.Engine().Parse(403, "not json")[0].Kind == EventKind.ErrorOccurred, "status before decode");
            Check(a.Engine().Parse(200, "not json")[0].Kind == EventKind.ErrorOccurred, "bad response is an error event");
        }
        finally { app.Destroy(); }
    });
    CheckCase("empty assistant text is retained, not null or absent", () =>
    {
        var view = new Context([new Entry(1, Actor.Human, [new TextPart("question")]), new Entry(2, Actor.Agent, [new TextPart("")])]);
        var wire = JsonNode.Parse(Wire(Vendor.OpenAI, view))!;
        Check(wire["messages"]!.AsArray().Count == 3 && wire["messages"]![2]!["content"]!.GetValue<string>() == "", "empty assistant survives projection");
    });
    CheckCase("usage normalization, invalid operands, overflow, and known output", () =>
    {
        var app = new Application(new StringWriter());
        try
        {
            var a = NewAgent(app, Vendor.OpenAI);
            Usage Parse(string usage) => a.Engine().Parse(200, "{\"model\":\"m\",\"choices\":[{\"message\":{\"content\":\"ok\"}}],\"usage\":" + usage + "}")[^1].Usage!;
            var valid = Parse("{\"prompt_tokens\":100,\"completion_tokens\":7,\"prompt_tokens_details\":{\"cached_tokens\":20,\"cache_write_tokens\":30}}");
            Check(valid.Input == 50 && valid.CacheRead == 20 && valid.CacheWrite == 30 && valid.Output == 7, "disjoint categories");
            foreach (var bad in new[] { "-1", "1.5", "\"3\"", "null", "true", "9223372036854775808" })
            {
                var invalid = Parse("{\"prompt_tokens\":100,\"completion_tokens\":7,\"prompt_tokens_details\":{\"cached_tokens\":" + bad + "}}");
                Check(invalid.Input is null && invalid.CacheRead is null && invalid.Output == 7, "invalid count must not become zero: " + bad);
            }
            var inconsistent = Parse("{\"prompt_tokens\":1,\"completion_tokens\":7,\"prompt_tokens_details\":{\"cached_tokens\":2}}");
            Check(inconsistent.Input is null && inconsistent.Output == 7, "invalid subset does not erase output");
            Check(Parse("{\"prompt_tokens\":5,\"completion_tokens\":2}").Input == 5, "missing optional counts are zero");
            Check(new Usage(long.MaxValue).Add(new Usage(1)).Input is null, "overflow unknown");
        }
        finally { app.Destroy(); }
    });
    foreach (var vendor in Enum.GetValues<Vendor>())
    {
        await AsyncCase("real Agent HTTP exchange and failure consumption: " + vendor, async () =>
        {
            var listener = new TcpListener(IPAddress.Loopback, 0);
            listener.Start();
            using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
            var requests = new List<string>();
            var port = ((IPEndPoint)listener.LocalEndpoint).Port;
            var server = Task.Run(async () =>
            {
                for (int turn = 0; turn < 2; turn++)
                {
                    using var client = await listener.AcceptTcpClientAsync(deadline.Token);
                    using var stream = client.GetStream();
                    requests.Add(await ReadRequest(stream, deadline.Token));
                    var body = vendor switch
                    {
                        Vendor.Anthropic => """{"model":"model-a","content":[{"type":"text","text":"local reply"}],"usage":{"input_tokens":3,"output_tokens":2}}""",
                        Vendor.OpenAI => """{"model":"model-a","choices":[{"message":{"content":"local reply"}}],"usage":{"prompt_tokens":3,"completion_tokens":2}}""",
                        _ => """{"modelVersion":"model-a","candidates":[{"content":{"parts":[{"text":"local reply"}]}}],"usageMetadata":{"promptTokenCount":3,"candidatesTokenCount":2}}"""
                    };
                    if (turn == 0)
                    {
                        body = "{\"error\":\"temporary local refusal\"}";
                    }

                    var bytes = Encoding.UTF8.GetBytes(body);
                    var header = Encoding.ASCII.GetBytes($"HTTP/1.1 {(turn == 0 ? "503 Service Unavailable" : "200 OK")}\r\nContent-Type: application/json\r\nContent-Length: {bytes.Length}\r\nConnection: close\r\n\r\n");
                    await stream.WriteAsync(header, deadline.Token);
                    await stream.WriteAsync(bytes, deadline.Token);
                }
            }, deadline.Token);
            var app = new Application(TextWriter.Null);
            try
            {
                var a = NewAgent(app, vendor, "http://127.0.0.1:" + port);
                a.SetEphemeral("ONCE_ONLY");
                try
                {
                    await a.SendAsync("first", deadline.Token);
                    throw new Exception("Expected provider refusal");
                }
                catch (InvalidOperationException) { }
                Check(a.History().View().Phase == Phase.Idle, "failure returns idle");
                var text = await a.SendAsync("second", deadline.Token);
                await server;
                Check(text == "local reply" && requests.Count == 2, "actual response and no retry");
                Check(requests[0].Contains("ONCE_ONLY") && !requests[1].Contains("ONCE_ONLY"), "ephemera consumed across real failure");
                Check(requests[1].Contains("first") && requests[1].Contains("second"), "recorded conversation retained");
                Check(a.History().View().Usage.Input == 3 && a.History().View().Usage.Output == 2, "real usage derived from events");
                Check(a.History().Events().All(e => !JsonCodec.WriteEvent(e).Contains("dummy-local-only")), "credentials not recorded");
            }
            finally { app.Destroy(); listener.Stop(); deadline.Cancel(); try { await server; } catch (OperationCanceledException) { } }
        });
    }
}
finally { Directory.Delete(temp, true); }
Console.WriteLine($"{passed} passed; {failures} failed");
return failures == 0 ? 0 : 1;

static async Task<string> ReadRequest(NetworkStream stream, CancellationToken cancellation)
{
    using var header = new MemoryStream();
    var one = new byte[1];
    while (true)
    {
        if (await stream.ReadAsync(one, cancellation) != 1)
        {
            throw new IOException("EOF in headers");
        }

        header.WriteByte(one[0]);
        if (header.Length > 65536)
        {
            throw new IOException("oversized headers");
        }

        var bytes = header.GetBuffer();
        var n = (int)header.Length;
        if (n >= 4 && bytes[n - 4] == 13 && bytes[n - 3] == 10 && bytes[n - 2] == 13 && bytes[n - 1] == 10)
        {
            break;
        }
    }
    var headers = Encoding.ASCII.GetString(header.ToArray());
    var lengthLine = headers.Split("\r\n").Single(x => x.StartsWith("Content-Length:", StringComparison.OrdinalIgnoreCase));
    var length = int.Parse(lengthLine.Split(':', 2)[1].Trim());
    if (length is < 0 or > 1048576)
    {
        throw new IOException("oversized body");
    }

    var body = new byte[length];
    await stream.ReadExactlyAsync(body, cancellation);
    using var parsed = JsonDocument.Parse(body); // A real protocol fake rejects malformed JSON.
    return Encoding.UTF8.GetString(body);
}
