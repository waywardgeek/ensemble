// Separate processes exercise the early JSON-lines API, not later snapshots or grants.
using System.Text.Json;
using Ensemble;

static class LogChecks
{
    sealed record Expected(string CallID, string Payload, string ID, string Name, long Bytes);

    public static int Run(string mode)
    {
        var checks = new List<string>();
        void Require(bool condition, string name)
        {
            if (!condition) throw new InvalidOperationException(name);
            checks.Add(name);
        }
        bool Matches(WorkProductInfo? info, Expected expected) => info is not null &&
            info.ID == expected.ID && info.Name == expected.Name && info.OfferedBytes == expected.Bytes;

        using var diagnostics = new StringWriter();
        var app = new Ensemble.Ensemble(diagnostics);
        try
        {
            var owner = new Agent(app, new Configuration(Vendor.Anthropic,
                "claude-3-haiku-20240307", "http://127.0.0.1:1", "synthetic-unused"));
            var history = owner.History();
            if (mode == "log-write")
            {
                Require(new[] { "events.jsonl", "dump.jsonl", "expected.json" }.All(path => !File.Exists(path)),
                    "fresh-writer-workspace");
                var expected = new Expected("call-" + Guid.NewGuid().ToString("N"),
                    "payload-" + Guid.NewGuid().ToString("N"), Guid.NewGuid().ToString("N"),
                    "report-" + Guid.NewGuid().ToString("N") + ".txt", 12345);
                history.StartLog(Path.GetFullPath("events.jsonl"));
                var fact = history.Append(new Event(EventKind.ToolReturned, callID: expected.CallID,
                    parts: [new TextPart(expected.Payload)],
                    workProduct: new WorkProductInfo(expected.ID, expected.Name, expected.Bytes)));
                Require(Matches(fact.WorkProduct, expected), "append-retains-description");
                var result = history.View().Dialogue.SelectMany(entry => entry.Parts)
                    .OfType<ToolResultPart>().Single(part => part.CallID == expected.CallID);
                Require(Matches(result.WorkProduct, expected) && JsonCodec.Text(result.Parts) == expected.Payload,
                    "writer-context-retains-description");
                using (var dump = new StreamWriter("dump.jsonl")) history.Dump(dump);
                File.WriteAllText("expected.json", JsonSerializer.Serialize(expected));
                // Compare at the same lifecycle boundary; destruction may append a shutdown fact.
                Require(File.ReadAllText("events.jsonl") == File.ReadAllText("dump.jsonl"), "dump-matches-written-log");
                owner.Destroy();
            }
            else
            {
                // No seed or repair path: a separate successful writer must supply these files.
                var expected = JsonSerializer.Deserialize<Expected>(File.ReadAllText("expected.json"))!;
                Require(expected.Bytes == 12345 && expected.ID.Length == 32 && expected.CallID.StartsWith("call-", StringComparison.Ordinal)
                    && expected.Payload.StartsWith("payload-", StringComparison.Ordinal), "writer-manifest-shape");
                foreach (string path in new[] { "events.jsonl", "dump.jsonl" })
                {
                    var recorded = File.ReadLines(path).Where(line => !string.IsNullOrWhiteSpace(line))
                        .Select(line => JsonDocument.Parse(line)).ToArray();
                    try
                    {
                        var raw = recorded.Single(document => document.RootElement.TryGetProperty("tool", out var tool) &&
                            tool.TryGetProperty("call_id", out var call) && call.GetString() == expected.CallID);
                        var fact = JsonCodec.ReadEvent(raw.RootElement.GetRawText());
                        Require(Matches(fact.WorkProduct, expected) && JsonCodec.Text(fact.Parts) == expected.Payload,
                            path + "-canonical-description");
                    }
                    finally { foreach (var document in recorded) document.Dispose(); }
                    history.Load(Path.GetFullPath(path));
                    var result = history.View().Dialogue.SelectMany(entry => entry.Parts)
                        .OfType<ToolResultPart>().Single(part => part.CallID == expected.CallID);
                    Require(Matches(result.WorkProduct, expected) && JsonCodec.Text(result.Parts) == expected.Payload,
                        path + "-restored-description");
                    using var dump = new StringWriter();
                    history.Dump(dump);
                    var events = dump.ToString().Split('\n', StringSplitOptions.RemoveEmptyEntries)
                        .Select(line => JsonDocument.Parse(line)).ToArray();
                    try
                    {
                        var raw = events.Single(document => document.RootElement.TryGetProperty("tool", out var tool) &&
                            tool.TryGetProperty("call_id", out var call) && call.GetString() == expected.CallID);
                        Require(Matches(JsonCodec.ReadEvent(raw.RootElement.GetRawText()).WorkProduct, expected),
                            path + "-redump-description");
                    }
                    finally { foreach (var document in events) document.Dispose(); }
                }
            }
            Console.WriteLine(JsonSerializer.Serialize(new { contract = "chapter3-metadata-log-v1", mode, passed = true, checks }));
            return 0;
        }
        catch (Exception error)
        {
            Console.WriteLine(JsonSerializer.Serialize(new { contract = "chapter3-metadata-log-v1", mode, passed = false, checks, failure = error.Message }));
            return 1;
        }
        finally { app.Destroy(); }
    }
}
