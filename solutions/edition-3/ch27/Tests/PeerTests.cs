using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class PeerTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition)
        {
            throw new Exception(message);
        }
    }

    private static async Task Refuses(Func<Task> operation)
    {
        try
        {
            await operation();
        }
        catch (Exception error) when (error is ArgumentException or InvalidOperationException or IOException or InvalidDataException)
        {
            return;
        }
        throw new Exception("Expected peer operation refusal.");
    }

    private static string Text(Agent agent) => string.Join("\n", agent.History().Events()
        .SelectMany(fact => fact.Parts).OfType<TextPart>().Select(part => part.Text));

    public static async Task Persistence(string directory)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        string work = Path.Combine(directory, "peer-persistence");
        string container = Path.Combine(work, "ensemble");
        string manuals = Path.Combine(work, "manuals");
        McpTests.Manual(manuals, "primary", "type: primary\ntools: read_file\n", "Primary $ROLE");
        McpTests.Manual(manuals, "writer", "type: primary\ntools: list_directory\n", "Writer $ROLE");
        var config = new Configuration(Vendor.Anthropic, "launch-default", "http://127.0.0.1:1",
            safeMode: true, sandboxRoot: work);
        var specification = new SkillSpec(manuals, "primary", new Dictionary<string, string> { ["ROLE"] = "original" });
        string registry = Path.Combine(container, "ensemble.json");
        var app = new Application(TextWriter.Null, dataDirectory: container);
        string savedRegistry;
        try
        {
            var root = new Agent(app, config, name: "root", skillSpec: specification);
            root.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("root-only dialogue")]));
            var writer = await app.CreatePeerAsync(root, "writer", "launch-writer", "writer", deadline.Token);
            Check(writer.Parent() == app && writer.AgentParent() is null && root.Children().IsEmpty,
                "A peer acquired the source Agent as an owning parent.");
            Check(writer.DataDirectory() == Path.Combine(container, "agents", "writer")
                && File.Exists(Path.Combine(writer.DataDirectory(), "save.json")), "Creation returned before saving History.");
            Check(writer.SkillManager() != root.SkillManager() && writer.ToolEnabled("list_directory")
                && !writer.ToolEnabled("read_file") && !Text(writer).Contains("root-only dialogue"),
                "Peer borrowed activation or copied the source conversation.");
            writer.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("writer-owned dialogue")]));
            var child = await writer.CreateChildAsync("old-child", cancellation: deadline.Token);
            child.PublishSubmission("old child completion");
            await writer.SetModelAsync("later-running-selection");
            root.SkillManager()!.RegisterVariable("ROLE", "changed");
            // Parsing is a construction boundary, not a disk dependency on every
            // creation. A second peer must use the source's independent snapshot.
            Directory.Delete(manuals, recursive: true);
            var second = await app.CreatePeerAsync(root, "second", cancellation: deadline.Token);
            Check(second.SkillManager()!.Active().Any(skill => skill.Body.Contains("changed"))
                && writer.SkillManager()!.Active().Any(skill => skill.Body.Contains("original")),
                "Peer skill values were shared or source templates reread.");
            savedRegistry = File.ReadAllText(registry);
            using var json = JsonDocument.Parse(savedRegistry);
            Check(json.RootElement.GetArrayLength() == 2
                && json.RootElement[0].GetProperty("model").GetString() == "launch-writer"
                && !savedRegistry.Contains("later-running-selection") && !savedRegistry.Contains("http:"),
                "Saving another peer persisted mutable model selection or runtime endpoint.");
            await Refuses(() => app.CreatePeerAsync(root, "writer"));
            await Refuses(() => app.CreatePeerAsync(child, "not-a-peer-source"));
            Check(File.ReadAllText(registry) == savedRegistry, "Refusal changed the registry.");
        }
        finally
        {
            app.Destroy();
        }
        Check(File.ReadAllText(registry) == savedRegistry, "Ordinary close rewrote launch choices.");
        McpTests.Manual(manuals, "primary", "type: primary\ntools: read_file\n", "Current primary");
        McpTests.Manual(manuals, "writer", "type: primary\ntools: list_directory\n", "Current writer");
        app = new Application(TextWriter.Null, dataDirectory: container);
        try
        {
            var root = new Agent(app, config.WithModel("new-host-default"), name: "root", skillSpec: specification);
            var restored = await app.RestorePeersAsync(root, deadline.Token);
            var writer = restored.Single(peer => peer.Name == "writer");
            Check(restored.Length == 2 && app.Agents().Length == 3 && writer.Children().IsEmpty
                && writer.Configuration().Model == "launch-writer" && writer.Submission() is null,
                "Restore lost launch choices or reconstructed delegated work.");
            Check(Text(writer).Contains("writer-owned dialogue") && !Text(writer).Contains("root-only dialogue")
                && writer.SkillManager()!.Active().Any(skill => skill.Body.Contains("Current writer")),
                "Restore lost own conversation or retained old runtime instructions.");
            await Refuses(() => app.RestorePeersAsync(root));
            Check(app.Agents().Length == 3 && File.ReadAllText(registry) == savedRegistry,
                "Repeated restore changed live membership or registry.");
        }
        finally
        {
            app.Destroy();
        }
    }

    public static async Task Deletion(string directory)
    {
        string home = Path.Combine(directory, "peer-deletion");
        var configuration = new Configuration(Vendor.Anthropic, "model", "http://127.0.0.1:1");
        var app = new Application(TextWriter.Null, dataDirectory: home);
        string registry = Path.Combine(home, "ensemble.json");
        string peerHome;
        try
        {
            var root = new Agent(app, configuration, name: "root");
            var peer = await app.CreatePeerAsync(root, "writer");
            peerHome = peer.DataDirectory();
            var child = await peer.CreateChildAsync("unfinished");
            string original = File.ReadAllText(registry);
            await Refuses(() => app.DeletePeerAsync(root, "writer"));
            await Refuses(() => app.DeletePeerAsync(root, "root"));
            Check(app.FindAgent("writer") == peer && File.ReadAllText(registry) == original,
                "Refused deletion changed membership or launch choices.");
            var laterChild = await peer.CreateChildAsync("still-admitting");
            laterChild.Destroy();
            using var cancelled = new CancellationTokenSource();
            cancelled.Cancel();
            bool cancelledBeforeAdmission = false;
            try
            {
                await app.DeletePeerAsync(root, "writer", cancelled.Token);
            }
            catch (OperationCanceledException)
            {
                cancelledBeforeAdmission = true;
            }
            Check(cancelledBeforeAdmission && app.FindAgent("writer") == peer,
                "Cancelled deletion was admitted.");
            child.PublishSubmission("completed");
            await app.DeletePeerAsync(root, "writer");
            Check(app.FindAgent("writer") is null && app.Agents().Length == 1 && peer.Children().IsEmpty
                && File.ReadAllText(registry) == "[]" && File.Exists(Path.Combine(peerHome, "save.json"))
                && File.Exists(Path.Combine(peerHome, "children", "unfinished", "submit.md")),
                "Deletion failed to detach the subtree or erased its saved files.");
        }
        finally
        {
            app.Destroy();
        }
        app = new Application(TextWriter.Null, dataDirectory: home);
        try
        {
            var root = new Agent(app, configuration, name: "root");
            Check((await app.RestorePeersAsync(root)).IsEmpty && File.Exists(Path.Combine(peerHome, "save.json")),
                "Restart resurrected a deleted peer or removed retained conversation data.");
        }
        finally
        {
            app.Destroy();
        }
    }

    public static async Task Refusals(string directory)
    {
        string work = Path.Combine(directory, "peer-refusals");
        Directory.CreateDirectory(work);
        var app = new Application(TextWriter.Null, dataDirectory: work);
        try
        {
            var root = new Agent(app, new Configuration(Vendor.Anthropic, "model", "http://127.0.0.1:1"), name: "root");
            string path = Path.Combine(work, "ensemble.json");
            string valid = """{"name":"first","model":"model","skill":"","created":"2026-01-01T00:00:00Z"}""";
            foreach (string text in new[] { "{", "{}", "[" + valid + "," + valid + "]",
                "[" + valid.Replace("\"first\"", "\"../escape\"") + "]",
                "[" + valid.Replace("\"model\":", "\"name\":\"duplicate\",\"model\":") + "]" })
            {
                File.WriteAllText(path, text);
                await Refuses(() => app.RestorePeersAsync(root));
                await Refuses(() => app.CreatePeerAsync(root, "new"));
                Check(app.Agents().Length == 1 && File.ReadAllText(path) == text,
                    "Malformed registry changed owners or was silently replaced.");
            }
            // Full syntax validation precedes construction, but a later startup
            // can still fail. Roll back earlier peers without rewriting the file.
            var rows = new JsonArray(JsonNode.Parse(valid), JsonNode.Parse(valid));
            rows[1]!["name"] = "broken";
            rows[1]!["skill"] = "missing-primary";
            string failedStartup = rows.ToJsonString();
            File.WriteAllText(path, failedStartup);
            await Refuses(() => app.RestorePeersAsync(root));
            Check(app.Agents().Length == 1 && File.ReadAllText(path) == failedStartup,
                "Failed restoration retained a partial fleet or damaged launch choices.");
            File.Delete(path);
            await Refuses(() => app.CreatePeerAsync(root, "bad-skill", skill: "missing-primary"));
            Check(app.Agents().Length == 1 && !File.Exists(path), "Failed creation published a registry.");
            var peer = await app.CreatePeerAsync(root, "first");
            Check(peer.SkillManager() is null && app.Agents().Length == 2, "Rollback prevented clean creation.");
        }
        finally
        {
            app.Destroy();
        }
    }
}
