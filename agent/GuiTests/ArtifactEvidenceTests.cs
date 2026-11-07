using System.Text.Json;
using Ensemble;
using EnsembleGui;

internal static class ArtifactEvidenceTests
{
    public static async Task RunAsync()
    {
        string work = Path.Combine(Path.GetTempPath(), "artifact-evidence-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(work);
        var root = new Ensemble.Ensemble(TextWriter.Null);
        try
        {
            string path = Path.Combine(work, "main.go");
            const string source = "package main\nimport \"fmt\"\nfunc main() { fmt.Println(\"hello\") }\n";
            File.WriteAllText(path, source);
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "offline", "http://127.0.0.1:1", sandboxRoot: work));
            agent.InstallTools();
            async Task Check(string tool, object args, string field)
            {
                var call = new ToolCallPart(tool, agent.Configuration().Target(), tool, JsonSerializer.SerializeToElement(args));
                var result = await agent.DispatchAsync(call);
                if (result.IsError || result.Job?.Status != "done") throw new Exception("Fixture operation did not finish: " + tool);
                var evidence = JsonCodec.WritePart(result)[field];
                if (evidence is null) throw new Exception("Real tool omitted evidence: " + tool);
                if (tool == "read_file" && result.FileEvidence!.Content.Text != source)
                    throw new Exception("Source evidence changed");
                var context = new Context([
                    new Entry(1, Actor.Agent, [call]),
                    new Entry(2, Actor.Tool, [result])]);
                var returned = new Event(EventKind.ToolReturned, seq: 2, callID: result.CallID,
                    parts: result.Parts, job: result.Job, directoryListing: result.DirectoryListing,
                    fileEvidence: result.FileEvidence, searchEvidence: result.SearchEvidence);
                foreach (var snapshot in new[] { new HistorySnapshot([returned], context, 2), new HistorySnapshot([], context, 2) })
                {
                    var projected = Gui.Project(snapshot, "fixture").Frames.Single(f => f["type"]!.GetValue<string>() == "tool_finished");
                    if (!System.Text.Json.Nodes.JsonNode.DeepEquals(evidence, projected[field]))
                        throw new Exception("GUI projection lost " + field + " from real " + tool + " result");
                }
                Console.WriteLine("PASS real " + tool + " evidence through ordinary and snapshot-only GUI projection");
            }
            await Check("read_file", new { path, ai_callback_delay = 3 }, "file_evidence");
            await Check("list_directory", new { path = work, ai_callback_delay = 3 }, "directory_listing");
            await Check("search_files", new { path, pattern = "fmt", ai_callback_delay = 3 }, "search_evidence");
        }
        finally
        {
            root.Destroy();
            Directory.Delete(work, true);
        }
    }
}
