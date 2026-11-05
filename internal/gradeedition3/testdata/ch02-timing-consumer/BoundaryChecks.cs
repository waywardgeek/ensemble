// Early public APIs only. Writer and reader are separate process invocations.
using System.Globalization;
using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;

static class BoundaryChecks
{
    sealed record Expected(string CallID, string Payload, string Time, long Ticks);

    public static int Run(string[] args)
    {
        var checks = new List<string>();
        void Require(bool condition, string name)
        {
            if (!condition) throw new InvalidOperationException(name);
            checks.Add(name);
        }
        using var diagnostics = new StringWriter();
        var app = new Ensemble.Ensemble(diagnostics);
        string mode = args[0];
        try
        {
            Require(args.Length == 1 && mode is "log-write" or "log-read" or "render", "valid-mode");
            var config = new Configuration(Vendor.Anthropic, "claude-3-haiku-20240307",
                "http://127.0.0.1:1", "synthetic-unused");
            var owner = new Agent(app, config);
            var history = owner.History();
            if (mode == "log-write")
            {
                Require(!File.Exists("events.jsonl") && !File.Exists("expected.json") &&
                    !File.Exists("legacy.jsonl") && !File.Exists("dump.jsonl"), "fresh-writer-workspace");
                history.StartLog(Path.GetFullPath("events.jsonl"));
                var expected = new List<Expected>();
                foreach (long ticks in new long[] { 12345678, 0 })
                {
                    string id = Guid.NewGuid().ToString("N");
                    string payload = "retained-" + Guid.NewGuid().ToString("N");
                    DateTimeOffset before = DateTimeOffset.UtcNow;
                    var fact = history.Append(new Event(EventKind.ToolReturned, callID: id,
                        parts: [new TextPart(payload)], observationDuration: TimeSpan.FromTicks(ticks)));
                    DateTimeOffset after = DateTimeOffset.UtcNow;
                    Require(fact.Time is { } time && time.Offset == TimeSpan.Zero && time >= before && time <= after &&
                        fact.ObservationDuration?.Ticks == ticks, "append-stamps-utc-and-retains-" + ticks);
                    expected.Add(new Expected(id, payload, fact.Time!.Value.ToString("O"), ticks));
                }
                using (var dump = new StreamWriter("dump.jsonl")) history.Dump(dump);
                File.WriteAllText("expected.json", JsonSerializer.Serialize(expected));
                // Independent headerless legacy fixture: Append correctly stamps new facts,
                // so it cannot create a historical fact whose time was never recorded.
                var legacy = new Event(EventKind.ToolReturned, seq: 1, callID: "legacy-unknown",
                    parts: [new TextPart("legacy payload")]);
                File.WriteAllText("legacy.jsonl", JsonCodec.WriteEvent(legacy) + "\n");
                Require(File.ReadAllText("events.jsonl") == File.ReadAllText("dump.jsonl"), "dump-matches-written-log");
                owner.Destroy(); // May append lifecycle facts; close before reader admission.
            }
            else if (mode == "log-read")
            {
                var expected = JsonSerializer.Deserialize<Expected[]>(File.ReadAllText("expected.json"))!;
                Require(expected.Length == 2 && expected[0].Ticks == 12345678 && expected[1].Ticks == 0 &&
                    expected.Select(item => item.CallID).Distinct().Count() == 2, "writer-manifest-shape");
                foreach (string path in new[] { "events.jsonl", "dump.jsonl" })
                {
                    history.Load(Path.GetFullPath(path));
                    foreach (var item in expected)
                    {
                        var entry = history.View().Dialogue.Single(value => value.Parts.OfType<ToolResultPart>()
                            .Any(part => part.CallID == item.CallID));
                        var exact = DateTimeOffset.Parse(item.Time, CultureInfo.InvariantCulture);
                        Require(entry.Time?.EqualsExact(exact) == true && entry.ObservationDuration?.Ticks == item.Ticks &&
                            JsonCodec.Text(entry.Parts).Contains(item.Payload, StringComparison.Ordinal),
                            path + "-retains-" + item.Ticks);
                    }
                }
                history.Load(Path.GetFullPath("legacy.jsonl"));
                var unknown = history.View().Dialogue.Single();
                Require(unknown.Time is null && unknown.ObservationDuration is null &&
                    JsonCodec.Text(unknown.Parts).Contains("legacy payload", StringComparison.Ordinal), "legacy-log-remains-unknown");
                using var dumped = new StringWriter();
                history.Dump(dumped);
                var roundtrip = JsonCodec.ReadEvent(dumped.ToString().Split('\n', StringSplitOptions.RemoveEmptyEntries).Last());
                Require(roundtrip.Time is null && roundtrip.ObservationDuration is null, "legacy-dump-remains-unknown");
            }
            else
            {
                foreach (var (vendor, model, surface) in new[] {
                    (Vendor.Anthropic, "claude-3-haiku-20240307", Surface.Messages),
                    (Vendor.OpenAI, "gpt-4o-mini", Surface.ChatCompletions),
                    (Vendor.Gemini, "gemini-2.0-flash", Surface.GenerateContent) })
                {
                    var route = new Configuration(vendor, model, "http://127.0.0.1:1", "synthetic-unused");
                    Require(route.Target().Surface == surface, "early-route-" + surface);
                    var renderer = new Agent(app, route);
                    string nonce = Guid.NewGuid().ToString("N");
                    var call = new ToolCallPart("call-" + nonce, route.Target(), "timing_probe",
                        JsonSerializer.SerializeToElement(new { value = nonce }));
                    var result = new ToolResultPart(call.CallID, [new TextPart("result-" + nonce)]);
                    var time = DateTimeOffset.Parse("2004-03-02T01:02:03.4567890+05:30", CultureInfo.InvariantCulture);
                    var plain = new Context([new Entry(1, Actor.Agent, [call]), new Entry(2, Actor.Tool, [result])]);
                    var timed = new Context([new Entry(1, Actor.Agent, [call], time: time),
                        new Entry(2, Actor.Tool, [result], time: time.AddSeconds(1), observationDuration: TimeSpan.FromTicks(12345678))]);
                    var baseline = renderer.Engine().Render(plain);
                    var measured = renderer.Engine().Render(timed);
                    Require(measured.ToJsonString().Contains("result-" + nonce, StringComparison.Ordinal) &&
                        measured.ToJsonString().Contains("timing_probe", StringComparison.Ordinal), "render-payload-" + surface);
                    Require(JsonNode.DeepEquals(baseline, measured), "render-independent-" + surface);
                    renderer.Destroy();
                }
            }
            Console.WriteLine(JsonSerializer.Serialize(new { contract = "chapter2-timing-boundaries-v1", mode, passed = true, checks }));
            return 0;
        }
        catch (Exception error)
        {
            Console.WriteLine(JsonSerializer.Serialize(new { contract = "chapter2-timing-boundaries-v1", mode, passed = false, checks, failure = error.Message }));
            return 1;
        }
        finally { app.Destroy(); }
    }
}
