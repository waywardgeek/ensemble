using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class SubmissionRetrievalTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition) throw new Exception(message);
    }

    private static Agent Parent(Application app, string directory, string name)
    {
        Directory.CreateDirectory(directory);
        var config = new Configuration(Vendor.Anthropic, "retrieval-fixture", "http://127.0.0.1:1",
            sandboxRoot: directory);
        var parent = new Agent(app, config, name: name, dataDirectory: directory);
        parent.InstallTools();
        return parent;
    }

    public static async Task FreshHistory(string directory)
    {
        var app = new Application(TextWriter.Null);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        try
        {
            Agent parent = Parent(app, Path.Combine(directory, "retrieval-fresh"), "review");
            Agent child = await parent.CreateChildAsync("writer", cancellation: deadline.Token);
            child.PublishSubmission("first");
            string path = Path.Combine(child.DataDirectory(), "submit.md");
            StoredSubmission first = parent.RetrieveSubmission("writer");
            Check(first.ParentAddress == "review" && first.Generation == 0
                && first.Payload.GetProperty("result").GetString() == "first", "Wrong saved provenance or value.");
            var changed = JsonNode.Parse(File.ReadAllText(path))!.AsObject();
            changed["payload"]!["result"] = "fresh from disk";
            File.WriteAllText(path, changed.ToJsonString());
            Check(parent.RetrieveSubmission("writer").Payload.GetProperty("result").GetString() == "fresh from disk"
                && child.Submission()!.Value.GetProperty("result").GetString() == "first"
                && first.Payload.GetProperty("result").GetString() == "first",
                "Retrieval cached an old row or mutated live/previous values.");
            await child.PostAsync(new AgentInput(InputKind.Pause));
            Task<string> admitted = child.SubmitAsync([new TextPart("new work must not use saved completion")]);
            _ = await child.WorkStatusAsync(deadline.Token);
            Check(child.Submission() is null, "New work retained live completion.");
            _ = parent.RetrieveSubmission("writer");
            Check(child.Submission() is null, "Reading historical data completed current work.");
            await child.ShutdownAsync();
            bool stopped = false;
            try { await admitted.WaitAsync(deadline.Token); }
            catch (ObjectDisposedException) { stopped = true; }
            Check(stopped, "Shutdown did not reject the queued task.");
            child.Destroy();
            int owners = app.Agents().Length;
            Check(parent.RetrieveSubmission("writer").Payload.GetProperty("result").GetString() == "fresh from disk"
                && parent.Children().Length == 0 && app.Agents().Length == owners,
                "Saved result was lost or retrieval resurrected an Agent.");
            Agent replacement = await parent.CreateChildAsync("writer", cancellation: deadline.Token);
            Check(replacement.Submission() is null, "Reconstruction revived completion.");
            _ = parent.RetrieveSubmission("writer");
            Check(replacement.Submission() is null, "Retrieval revived reconstructed completion.");
        }
        finally
        {
            app.Destroy();
        }
    }

    public static async Task Validation(string directory)
    {
        var app = new Application(TextWriter.Null);
        try
        {
            Agent parent = Parent(app, Path.Combine(directory, "retrieval-validation"), "owner");
            var schema = JsonSerializer.Deserialize<JsonElement>("""
                {"type":"object","properties":{"count":{"type":"integer"}},"required":["count"],"additionalProperties":false}
                """);
            Agent child = await parent.CreateChildAsync("typed", submitSchema: schema);
            child.PublishSubmission(JsonSerializer.Deserialize<JsonElement>("{\"count\":9007199254740993}"));
            string path = Path.Combine(child.DataDirectory(), "submit.md");
            string valid = File.ReadAllText(path);
            Check(parent.RetrieveSubmission("typed").Payload.GetProperty("count").GetRawText() == "9007199254740993",
                "Typed result lost exact integer precision.");
            var legacy = JsonNode.Parse(valid)!.AsObject();
            legacy["version"] = 1;
            legacy.Remove("parent_address");
            File.WriteAllText(path, legacy.ToJsonString());
            Check(parent.RetrieveSubmission("typed").ParentAddress is null,
                "Version 1 invented parent provenance.");
            foreach (Action<JsonObject> corrupt in new Action<JsonObject>[] {
                row => row["version"] = 99,
                row => row["version"] = "2",
                row => row["agent_id"] = "elsewhere",
                row => row["parent_address"] = "other/owner",
                row => row["parent_address"] = null,
                row => row["generation"] = -1,
                row => row["payload"]!["count"] = "not integer",
                row => row.Remove("submit_schema"),
                row => row["unknown"] = true })
            {
                var row = JsonNode.Parse(valid)!.AsObject();
                corrupt(row);
                File.WriteAllText(path, row.ToJsonString());
                Refused(() => parent.RetrieveSubmission("typed"));
            }
            foreach (string corrupt in new[] { "{}", "[]", valid.TrimEnd().TrimEnd('}') + ",\"version\":2}",
                valid.Replace("9007199254740993", "1,\"count\":2"), new string(' ', 2 * 1024 * 1024 + 1) })
            {
                File.WriteAllText(path, corrupt);
                Refused(() => parent.RetrieveSubmission("typed"));
            }
            File.WriteAllText(path, valid);
            foreach (string id in new[] { "../typed", "owner/typed", "typed/", "..", "/typed" })
            {
                Refused(() => parent.RetrieveSubmission(id));
            }
            File.Delete(path);
            string outside = Path.Combine(directory, "outside-retrieval.json");
            File.WriteAllText(outside, valid);
            File.CreateSymbolicLink(path, outside);
            Refused(() => parent.RetrieveSubmission("typed"));
            File.Delete(path);
            Refused(() => parent.RetrieveSubmission("typed"));
            Check(parent.Children().Single() == child && child.Submission() is not null,
                "Refusal changed ownership or live submission.");
        }
        finally
        {
            app.Destroy();
        }
    }

    public static async Task Tool(string directory,
        [System.Runtime.CompilerServices.CallerFilePath] string sourceFile = "")
    {
        var app = new Application(TextWriter.Null);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        try
        {
            Agent parent = Parent(app, Path.Combine(directory, "retrieval-tool"), "owner");
            string skillDirectory = Path.GetFullPath(Path.Combine(Path.GetDirectoryName(sourceFile)!, "..", "Skills"));
            _ = new Skills(parent, new SkillSpec(skillDirectory, "ensemble"));
            await parent.SkillManager()!.InitializeAsync(deadline.Token);
            Check(parent.ToolEnabled("get_submission")
                && parent.Engine().Render(parent.History().View()).ToJsonString().Contains("get_submission", StringComparison.Ordinal),
                "Shipped ensemble manual does not expose retrieval.");
            Agent child = await parent.CreateChildAsync("writer", cancellation: deadline.Token);
            child.PublishSubmission("retained result");
            child.Destroy();
            async Task<ToolResultPart> Call(string id)
            {
                ToolResultPart result = await parent.DispatchAsync(new ToolCallPart(Guid.NewGuid().ToString("N"),
                    parent.Configuration().Target(), "get_submission", JsonSerializer.SerializeToElement(new { agent_id = id })), deadline.Token);
                if (result.Job?.Status == "running")
                {
                    result = await parent.Jobs().Find(result.Job.Handle).CompletionAsync(deadline.Token);
                }
                return result;
            }
            ToolResultPart report = await Call("writer");
            string text = string.Concat(report.Parts.OfType<TextPart>().Select(part => part.Text));
            Check(!report.IsError && text.Contains("\"historical\":true", StringComparison.Ordinal)
                && text.Contains("retained result", StringComparison.Ordinal) && parent.Children().Length == 0,
                "Real dispatch failed or revived the child.");
            Check((await Call("elsewhere/writer")).IsError, "Tool accepted another branch's address.");
            File.WriteAllText(Path.Combine(parent.DataDirectory(), "children", "writer", "submit.md"),
                "{\"version\":99,\"private-marker\":true}");
            ToolResultPart corrupt = await Call("writer");
            string corruptText = string.Concat(corrupt.Parts.OfType<TextPart>().Select(part => part.Text));
            Check(corrupt.IsError && !corruptText.Contains("private-marker", StringComparison.Ordinal),
                "Corrupt artifact succeeded or echoed its content.");
            ToolResultPart missing = await Call("missing");
            string error = string.Concat(missing.Parts.OfType<TextPart>().Select(part => part.Text));
            Check(missing.IsError && !error.Contains(parent.DataDirectory(), StringComparison.Ordinal),
                "Missing artifact was successful or leaked host path.");
        }
        finally
        {
            app.Destroy();
        }
    }

    private static void Refused(Action action)
    {
        try { action(); }
        catch (Exception error) when (error is ArgumentException or IOException or InvalidDataException or UnauthorizedAccessException) { return; }
        throw new Exception("Invalid retrieval was accepted.");
    }
}
