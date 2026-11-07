using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class SearchEvidenceTests
{
    private static void Check(bool value, string message) { if (!value) throw new Exception(message); }
    private static void Refuses(Action action)
    {
        try { action(); } catch (ArgumentException) { return; }
        throw new Exception("Invalid search evidence accepted");
    }
    internal static async Task Run(string root)
    {
        var line = new SearchLine("file:with\nseparator", 2, "😀 hit hit", [new(3, 3), new(7, 3)]);
        var evidence = new SearchEvidence([line], 1);
        Refuses(() => new SearchLine("file", 0, "x", []));
        Refuses(() => new SearchLine("file", 1, "x", [new(-1, 1)]));
        Refuses(() => new SearchLine("file", 1, "x", [new(0, 2)]));
        Refuses(() => new SearchLine("file", 1, "xx", [new(0, 2), new(1, 1)]));
        Refuses(() => new SearchEvidence([line], 0));
        Refuses(() => new SearchEvidence(Enumerable.Repeat(line, 257), 257));
        Refuses(() => new ToolResultPart("bad", [], true, searchEvidence: evidence));
        Refuses(() => new Event(EventKind.MessageReceived, searchEvidence: evidence));
        var result = new ToolResultPart("codec", [new TextPart("raw")], searchEvidence: evidence);
        using var encoded = JsonDocument.Parse(JsonCodec.WritePart(result).ToJsonString());
        Check(((ToolResultPart)JsonCodec.ReadPart(encoded.RootElement)).SearchEvidence!.Lines[0].Matches[1].Start == 7,
            "part codec lost UTF-16 spans");
        var fact = new Event(EventKind.ToolReturned, callID: "codec", parts: result.Parts, searchEvidence: evidence).Numbered(1, DateTimeOffset.UtcNow);
        Check(JsonCodec.ReadEvent(JsonCodec.WriteEvent(fact)).SearchEvidence!.Lines[0].Path == line.Path, "event codec lost path");
        var directory = Path.Combine(root, "search-evidence"); Directory.CreateDirectory(directory);
        var path = Path.Combine(directory, "source:odd\nname.txt");
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "local", "http://127.0.0.1:1"));
            agent.InstallTools(); int sequence = 0;
            Task<ToolResultPart> Dispatch(string name, object args) => agent.DispatchAsync(new ToolCallPart(
                "search-evidence-" + ++sequence, agent.Configuration().Target(), name, JsonSerializer.SerializeToElement(args)));
            async Task<ToolResultPart> Search(string pattern, int radius = 0)
            {
                var response = await Dispatch("search_files", new { path, pattern, context_lines = radius, max_output_bytes = 64, ai_callback_delay = 3 });
                for (int attempt = 0; response.Job?.Status == "running" && attempt < 5; attempt++)
                    response = await Dispatch("wait_for_job", new { handle = response.Job.Handle, max_output_bytes = 64, ai_callback_delay = 3 });
                Check(response.Job?.Status == "done", "search did not finish"); return response;
            }
            File.WriteAllText(path, "before\n😀 hit hit\nafter\ngap\nlast hit\n");
            var searched = await Search("hit", 1);
            var data = searched.SearchEvidence!;
            Check(searched.Job!.Truncated && data.Complete && data.TotalLines == 5, "context or observation cap changed");
            Check(data.Lines[1].Matches.SequenceEqual(line.Matches) && data.Lines[0].Matches.IsEmpty, "match/context spans wrong");
            Check(data.Lines.Select(x => x.Number).SequenceEqual(new long[] { 1, 2, 3, 4, 5 }), "merged windows duplicated lines");
            File.WriteAllText(path, "changed");
            var saved = agent.History().Events().Single(e => e.CallID == searched.CallID && e.Kind is EventKind.ToolReturned or EventKind.ToolObserved);
            Check(JsonCodec.ReadEvent(JsonCodec.WriteEvent(saved)).SearchEvidence!.Lines[1].Text == "😀 hit hit", "replay lost captured text");
            Check((await Search("absent")).SearchEvidence is { Complete: true, TotalLines: 0 }, "empty success missing");
            var zero = (await Search("^")).SearchEvidence!;
            Check(zero.Lines[0].Matches.Single() == new SearchSpan(0, 0), "zero-width match lost");
            var bad = await Search("["); Check(bad.IsError && bad.SearchEvidence is null, "failed regex claimed evidence");
            File.WriteAllText(path, string.Concat(Enumerable.Repeat("hit\n", 300)));
            var many = (await Search("hit")).SearchEvidence!;
            Check(!many.Complete && many.Lines.Length == 256 && many.TotalLines == 300, "line bound lost omissions");
            File.WriteAllText(path, new string('x', 40000));
            var longLine = (await Search("x")).SearchEvidence!;
            Check(longLine.Lines.IsEmpty && longLine.TotalLines == 1 && !longLine.Complete, "oversize line not bounded");
            File.WriteAllText(path, new string('x', 257));
            Check((await Search("x")).SearchEvidence!.Lines.IsEmpty, "span bound not enforced");
        }
        finally { app.Destroy(); }
    }
}
