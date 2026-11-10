using System.Diagnostics;
using System.Text;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    // Verify transport of the complete read, not model comprehension or quality.
    // The Python budget is passed unchanged to the actual native tool, then the
    // next real provider-wire request is checked for every numbered source line.
    public static async Task FictionCompleteReads(string directory)
    {
        string home = Path.Combine(directory, "fiction-complete-reads");
        Directory.CreateDirectory(home);
        using var listener = Listen();
        var root = new Application(TextWriter.Null);
        try
        {
            var parent = MemoryAgent(root, home, listener.Prefixes.Single());
            parent.InstallTools();
            foreach (string kind in new[] { "chapter", "manuscript" })
            {
                string childHome = Path.Combine(home, "children", kind);
                Directory.CreateDirectory(childHome);
                File.WriteAllText(Path.Combine(childHome, "settings.json"), JsonSerializer.Serialize(new
                {
                    version = 1,
                    memory = JsonCodec.WriteMemorySettings(new MemorySettings()),
                    recall = JsonCodec.WriteRecallSettings(new RecallSettings(disabled: true))
                }));
                string[] lines = Enumerable.Range(1, kind == "chapter" ? 900 : 1800)
                    .Select(index => kind + " passage " + index + " " + new string('x', 40)).ToArray();
                string source = string.Join("\n", lines) + "\nFINAL_REVIEW_CANARY_" + kind;
                string file = Path.Combine(home, kind + ".md");
                File.WriteAllText(file, source, new UTF8Encoding(false));
                JsonElement read = await FictionReadArguments(home, kind + ".md");
                Check(read.GetProperty("max_bytes").GetInt32() > 32768, "Fixture did not exceed default read cap");
                var child = await parent.CreateChildAsync(kind, cancellation: CancellationToken.None);
                var turn = child.SubmitAsync([new TextPart("Read the complete assigned text once.")]);
                var first = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
                using (var reader = new StreamReader(first.Request.InputStream))
                {
                    await reader.ReadToEndAsync();
                }
                await Reply(first, JsonSerializer.Serialize(new
                {
                    choices = new[] { new { message = new
                    {
                        content = "I will read the full assigned file in one call.",
                        tool_calls = new[] { new { id = "complete-read", type = "function", function = new
                        {
                            name = "read_file", arguments = read.GetRawText()
                        } } }
                    }, finish_reason = "tool_calls" } }
                }));
                var second = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(10));
                using var body = new StreamReader(second.Request.InputStream);
                using var wire = JsonDocument.Parse(await body.ReadToEndAsync());
                var results = wire.RootElement.GetProperty("messages").EnumerateArray()
                    .Where(message => message.GetProperty("role").GetString() == "tool").ToArray();
                Check(results.Length == 1, "Review did not produce exactly one read result");
                string received = results[0].GetProperty("content").GetString()!;
                string expected = string.Join("\n", source.Split('\n')
                    .Select((line, index) => (index + 1) + ": " + line));
                Check(received == expected, "Complete numbered source did not reach reviewer request");
                Check(child.History().Events().Count(fact => fact.Kind == EventKind.ToolCalled) == 1,
                    "Reviewer made extra calls rather than one complete read");
                child.PublishSubmission("complete read verified by fixture");
                await Reply(second, JsonSerializer.Serialize(new
                {
                    choices = new[] { new { message = new { content = "Read checked" }, finish_reason = "stop" } }
                }));
                await turn.WaitAsync(TimeSpan.FromSeconds(10));
                child.Destroy();
            }
        }
        finally
        {
            root.Destroy();
        }
    }

    private static async Task<JsonElement> FictionReadArguments(string home, string source)
    {
        var start = new ProcessStartInfo("/usr/bin/python3")
        {
            UseShellExecute = false, RedirectStandardOutput = true, RedirectStandardError = true
        };
        start.ArgumentList.Add("-c");
        start.ArgumentList.Add("import sys,json; sys.path.insert(0,sys.argv[1]); "
            + "from run_state import RunStore; from review_turn import source_info; "
            + "s=RunStore(sys.argv[2]); "
            + "print(json.dumps(source_info(s,sys.argv[3],1000000,1100000)['read']))");
        start.ArgumentList.Add(Path.GetFullPath("agent/skills/fiction-workflow"));
        start.ArgumentList.Add(home);
        start.ArgumentList.Add(source);
        using var process = Process.Start(start) ?? throw new IOException("Cannot start read-budget fixture");
        var output = process.StandardOutput.ReadToEndAsync();
        var errors = process.StandardError.ReadToEndAsync();
        try
        {
            await process.WaitForExitAsync().WaitAsync(TimeSpan.FromSeconds(10));
            Check(process.ExitCode == 0, "Read budget failed: " + await errors);
            using var result = JsonDocument.Parse(await output);
            return result.RootElement.Clone();
        }
        finally
        {
            if (!process.HasExited)
            {
                process.Kill(entireProcessTree: true);
            }
            await process.WaitForExitAsync();
            await output;
            await errors;
        }
    }
}
