using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class FileEvidenceTests
{
    private static void Check(bool value, string message)
    {
        if (!value) throw new Exception(message);
    }
    private static void Refuses(Action action)
    {
        try { action(); }
        catch (ArgumentException) { return; }
        throw new Exception("Invalid file evidence accepted");
    }
    internal static async Task Run(string root)
    {
        var excerpt = FileExcerpt.Capture(new string('x', 32767) + "😀tail");
        Check(excerpt.Text.Length == 32767 && !excerpt.Complete && excerpt.TotalBytes == 32775,
            "UTF-8 display bound split a character or lost total");
        Refuses(() => new FileExcerpt("x", 0));
        Refuses(() => new FileExcerpt(new string('x', 32769), 32769));
        Refuses(() => new FileEvidence(FileOperation.Read, "file", 0, 1, excerpt));
        Refuses(() => new FileEvidence(FileOperation.Edit, "file", 1, 1, excerpt));
        var data = new FileEvidence(FileOperation.Append, "file.cs", 8, 4, excerpt);
        Refuses(() => new ToolResultPart("bad", [], true, fileEvidence: data));
        Refuses(() => new Event(EventKind.MessageReceived, fileEvidence: data));
        Check(new ToolResultPart("legacy", []).FileEvidence is null, "legacy invented evidence");
        var part = new ToolResultPart("file", [new TextPart("model report")], fileEvidence: data);
        using var json = JsonDocument.Parse(JsonCodec.WritePart(part).ToJsonString());
        var restored = (ToolResultPart)JsonCodec.ReadPart(json.RootElement);
        Check(restored.FileEvidence!.StartColumn == 4 && restored.FileEvidence.Content.TotalBytes == 32775,
            "part roundtrip lost evidence");
        Check(JsonCodec.Text(restored.Parts) == "model report", "display evidence changed model text");
        var fact = new Event(EventKind.ToolReturned, callID: "file", parts: part.Parts, fileEvidence: data)
            .Numbered(1, DateTimeOffset.UtcNow);
        Check(JsonCodec.ReadEvent(JsonCodec.WriteEvent(fact)).FileEvidence!.StartLine == 8,
            "event copies or codec lost evidence");

        string directory = Path.Combine(root, "file-evidence");
        Directory.CreateDirectory(directory);
        string path = Path.Combine(directory, "source.cs");
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "local", "http://127.0.0.1:1"));
            agent.InstallTools();
            int sequence = 0;
            Task<ToolResultPart> Dispatch(string name, object args) => agent.DispatchAsync(new ToolCallPart(
                "file-evidence-" + ++sequence, agent.Configuration().Target(), name, JsonSerializer.SerializeToElement(args)));
            async Task<ToolResultPart> Call(string name, object args)
            {
                var result = await Dispatch(name, args);
                for (int attempt = 0; result.Job?.Status == "running" && attempt < 5; attempt++)
                    result = await Dispatch("wait_for_job", new { handle = result.Job.Handle, ai_callback_delay = 3, max_output_bytes = 64 });
                Check(result.Job?.Status == "done", "file operation did not complete");
                return result;
            }
            var written = await Call("write_file", new { path, content = "head\n😀tail", ai_callback_delay = 3 });
            Check(written.FileEvidence is { Operation: FileOperation.Write, StartLine: 1, StartColumn: 1 }, "write evidence missing");
            var appended = await Call("write_file", new { path, content = "!\nnext", append = true, ai_callback_delay = 3 });
            Check(appended.FileEvidence is { Operation: FileOperation.Append, StartLine: 2, StartColumn: 7 }, "append position incorrect");
            Check(JsonCodec.Text(appended.Parts).Contains("Lines 2–3"), "model lost written range");
            var read = await Call("read_file", new { path, start_line = 2, max_bytes = 5, ai_callback_delay = 3 });
            Check(read.FileEvidence is { Operation: FileOperation.Read, StartLine: 2, StartColumn: 1 }
                && read.FileEvidence.Content.Text == "😀t" && !read.FileEvidence.Content.Complete,
                "read cap/offset evidence incorrect");
            var edited = await Call("edit_file", new { path, old_text = "tail!", new_text = "new\nvalue", ai_callback_delay = 3 });
            Check(edited.FileEvidence is { Operation: FileOperation.Edit, StartLine: 2, StartColumn: 3 }
                && edited.FileEvidence.Removed!.Text == "tail!" && edited.FileEvidence.Content.Text == "new\nvalue",
                "applied replacement evidence incorrect");
            var failed = await Call("edit_file", new { path, old_text = "absent", new_text = "no", ai_callback_delay = 3 });
            Check(failed.IsError && failed.FileEvidence is null, "failed edit claimed operation evidence");
            File.WriteAllText(path, new string('z', 40000));
            var capped = await Call("read_file", new { path, max_output_bytes = 64, ai_callback_delay = 3 });
            Check(capped.Job!.Truncated && capped.FileEvidence!.Content.Text.Length == 32768
                && !capped.FileEvidence.Content.Complete, "display evidence bypassed observation or display cap");
            var saved = agent.History().Events().Single(e => e.CallID == edited.CallID && e.Kind is EventKind.ToolReturned or EventKind.ToolObserved);
            Check(JsonCodec.ReadEvent(JsonCodec.WriteEvent(saved)).FileEvidence!.Removed!.Text == "tail!",
                "recorded operation changed when file changed");
            var again = await Dispatch("wait_for_job", new { handle = capped.Job.Handle, ai_callback_delay = 0 });
            Check(again.FileEvidence is null, "presentation evidence delivered twice");
        }
        finally { app.Destroy(); }
    }
}
