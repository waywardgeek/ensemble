using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;
using Application = Ensemble.Ensemble;

// Pending unload keeps teaching text, not permission. Exercise real dispatch and
// frozen provider projections rather than just checking the manual's marker.
internal static class SkillAuthorityTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition)
        {
            throw new Exception(message);
        }
    }

    private static ToolCallPart Call(Agent agent, string name, object arguments)
    {
        return new ToolCallPart(Guid.NewGuid().ToString("N"), agent.Configuration().Target(),
            name, JsonSerializer.SerializeToElement(arguments));
    }

    private static string[] Declarations(JsonObject request, Vendor vendor)
    {
        var tools = request["tools"]?.AsArray() ?? new JsonArray();
        if (vendor == Vendor.Gemini && tools.Count > 0)
        {
            tools = tools[0]!["functionDeclarations"]!.AsArray();
        }
        return tools.Select(tool => (vendor == Vendor.OpenAI ? tool!["function"]! : tool!)
            ["name"]!.GetValue<string>()).ToArray();
    }

    public static async Task PendingAuthority()
    {
        string root = SkillTests.Fixture();
        var app = new Application(TextWriter.Null);
        try
        {
            foreach (var vendor in Enum.GetValues<Vendor>())
            {
                var agent = new Agent(app, new Configuration(vendor, "model-a", "http://127.0.0.1:1",
                    deferredTools: vendor == Vendor.Anthropic), skillSpec: new SkillSpec(root, "root"));
                var manager = agent.SkillManager()!;
                var load = Call(agent, "load_skill", new { name = "editing" });
                var loaded = await agent.DispatchAsync(load);
                Check(!loaded.IsError, "Load refused");
                // Host controls are not model dialogue. Record a genuine-shaped
                // paired exchange to exercise historical discovery projection.
                agent.History().Append(new Event(EventKind.ResponseEnded, parts: [load]));
                agent.History().Append(new Event(EventKind.ToolReturned, callID: load.CallID,
                    parts: loaded.Parts, toolReferences: loaded.ToolReferences));
                var before = agent.History().View();
                string frozenWire = agent.Engine().Render(before).ToJsonString();
                string target = Path.Combine(root, vendor + "-effect.txt");
                Check(!(await agent.DispatchAsync(Call(agent, "write_file", new { path = target, content = "before" }))).IsError,
                    "Enabled writer refused");
                Check(!(await agent.DispatchAsync(Call(agent, "unload_skill", new { name = "editing" }))).IsError,
                    "Unload refused");
                Check(!agent.ToolEnabled("write_file") && !manager.Enabled("write_file")
                    && manager.Source("write_file") is null, "Pending writer retained authority");
                Check(agent.ToolEnabled("read_file") && manager.Source("read_file")?.Skill == "root",
                    "Shared grant was revoked");
                Check(agent.ToolEnabled("list_directory"), "Independent dependency was revoked");
                Check(!manager.LoadableSkills().Contains("extra"), "Pending manual still offered a child");
                var denied = await agent.DispatchAsync(Call(agent, "write_file", new { path = target, content = "after" }));
                Check(denied.IsError && File.ReadAllText(target) == "before", "Revoked writer had an effect");

                var after = agent.History().View();
                var request = agent.Engine().Render(after);
                var names = Declarations(request, vendor);
                Check(!names.Contains("write_file") && names.Contains("read_file"), "Wrong provider authority: " + vendor);
                string wire = request.ToJsonString();
                Check(wire.Contains("EDIT-") && wire.Split("EDIT-").Length == 2, "Manual lost or duplicated");
                Check(!wire.Contains("\"tool_name\":\"write_file\""), "Stale discovery reference survived");
                Check(agent.History().Events().Any(fact => fact.ToolReferences.Contains("write_file")),
                    "Filtering rewrote historical evidence");
                Check(agent.Engine().Render(before).ToJsonString() == frozenWire,
                    "Frozen request changed with live unload");

                // Snapshot-only persistence has no pending event to consult.
                string save = Path.Combine(root, vendor + "-save.json");
                agent.History().Save(new PersistenceOptions(save, RetainLog: 0));
                var replay = new Agent(app, new Configuration(vendor, "model-a", "http://127.0.0.1:1"));
                replay.InstallTools();
                Check(replay.History().LoadSave(save), "Snapshot not restored");
                Check(!replay.ToolEnabled("write_file") && replay.ToolEnabled("read_file"), "Replay restored revoked authority");
                Check(!Declarations(replay.Engine().Render(replay.History().View()), vendor).Contains("write_file"),
                    "Replay declared revoked writer");
            }
        }
        finally
        {
            app.Destroy();
            Directory.Delete(root, true);
        }
    }
}
