using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class DirectoryListingTests
{
    private static void Check(bool value, string message)
    {
        if (!value) throw new Exception(message);
    }
    private static void Refuses(Action action)
    {
        try { action(); }
        catch (ArgumentException) { return; }
        throw new Exception("Invalid directory evidence accepted");
    }
    internal static async Task Run(string root)
    {
        var listing = new DirectoryListing([new("one\ntwo", DirectoryEntryKind.Unknown)], 2);
        Check(!listing.Complete, "partial evidence marked complete");
        var original = new ToolResultPart("directory", [new TextPart("bounded text")], directoryListing: listing);
        using var json = JsonDocument.Parse(JsonCodec.WritePart(original).ToJsonString());
        var restored = (ToolResultPart)JsonCodec.ReadPart(json.RootElement);
        Check(restored.DirectoryListing!.Entries.Single().Name == "one\ntwo", "part codec split filename");
        Check(JsonCodec.Text(restored.Parts) == "bounded text", "presentation leaked into text projection");
        var fact = new Event(EventKind.ToolReturned, callID: "directory", parts: original.Parts, directoryListing: listing);
        var replay = JsonCodec.ReadEvent(JsonCodec.WriteEvent(fact.Numbered(1, DateTimeOffset.UtcNow)));
        Check(replay.DirectoryListing!.TotalEntries == 2 && !replay.DirectoryListing.Complete, "event replay lost subset evidence");
        Check(new ToolResultPart("legacy", []).DirectoryListing is null, "legacy invented metadata");
        Refuses(() => new DirectoryListing([new("x/y", DirectoryEntryKind.File)], 1));
        Refuses(() => new DirectoryListing([new("x", (DirectoryEntryKind)0)], 1));
        Refuses(() => new DirectoryListing([new("x", DirectoryEntryKind.File)], 0));
        Refuses(() => new DirectoryListing(Enumerable.Repeat(new DirectoryEntry("x", DirectoryEntryKind.File), 257), 257));
        Refuses(() => new DirectoryListing([new(new string('x', 32769), DirectoryEntryKind.File)], 1));
        Refuses(() => new ToolResultPart("bad", [], true, directoryListing: listing));
        Refuses(() => new Event(EventKind.MessageReceived, directoryListing: listing));

        string directory = Path.Combine(root, "directory-evidence");
        Directory.CreateDirectory(directory);
        File.WriteAllText(Path.Combine(directory, "one\ntwo"), "");
        Directory.CreateDirectory(Path.Combine(directory, "folder"));
        for (int i = 0; i < 300; i++) File.WriteAllText(Path.Combine(directory, $"z{i:D3}"), "");
        var app = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "local", "http://127.0.0.1:1"));
            agent.InstallTools();
            int sequence = 0;
            Task<ToolResultPart> Call(string name, object args) => agent.DispatchAsync(new ToolCallPart(
                "directory-" + ++sequence, agent.Configuration().Target(), name, JsonSerializer.SerializeToElement(args)));
            async Task<ToolResultPart> List(string path)
            {
                var result = await Call("list_directory", new { path, ai_callback_delay = 3, max_output_bytes = 64 });
                for (int attempt = 0; result.Job?.Status == "running" && attempt < 5; attempt++)
                    result = await Call("wait_for_job", new { handle = result.Job.Handle, ai_callback_delay = 3, max_output_bytes = 64 });
                Check(result.Job?.Status == "done" && !result.IsError, "directory operation did not complete");
                return result;
            }
            var result = await List(directory);
            var data = result.DirectoryListing ?? throw new Exception("observation lost directory evidence");
            Check(result.Job!.Truncated && JsonCodec.Text(result.Parts).Contains("output truncated"), "metadata bypassed model output cap");
            Check(data.TotalEntries == 302 && data.Entries.Length == 256 && !data.Complete, "display cap lost explicit omitted count");
            Check(data.Entries.Single(e => e.Name == "one\ntwo").Kind == DirectoryEntryKind.Unknown, "newline filename not one entry");
            Check(data.Entries.Single(e => e.Name == "folder").Kind == DirectoryEntryKind.Directory, "directory classification lost");
            var saved = agent.History().Events().Single(e => e.CallID == result.CallID && e.Kind is EventKind.ToolReturned or EventKind.ToolObserved);
            Check(JsonCodec.ReadEvent(JsonCodec.WriteEvent(saved)).DirectoryListing!.TotalEntries == 302, "recorded observation lost metadata");
            var again = await Call("wait_for_job", new { handle = result.Job.Handle, ai_callback_delay = 0 });
            Check(again.DirectoryListing is null, "already observed listing delivered twice");
            var empty = await List(Path.Combine(directory, "folder"));
            Check(empty.DirectoryListing is { TotalEntries: 0, Complete: true }, "empty listing lost completion evidence");
        }
        finally { app.Destroy(); }
    }
}
