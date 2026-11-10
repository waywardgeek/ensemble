using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

static partial class McpTests
{
    // Run from the repository root. Copy the public helper into the sandbox,
    // exactly as an external skill would distribute its supporting library.
    public static async Task PythonWorkflow()
    {
        if (!OperatingSystem.IsMacOS())
        {
            throw new PlatformNotSupportedException("Python workflow test requires the macOS sandbox");
        }
        string tree = Path.Combine(Path.GetTempPath(), "mcp-python-" + Guid.NewGuid().ToString("N"));
        string inside = Path.Combine(tree, "inside");
        Directory.CreateDirectory(inside);
        string outside = Path.Combine(tree, "outside.txt");
        File.WriteAllText(outside, "unchanged");
        File.Copy("agent/skills/shared/ensemble_mcp.py", Path.Combine(inside, "ensemble_mcp.py"));
        File.WriteAllText(Path.Combine(inside, "workflow.py"), """
            import json
            from ensemble_mcp import Server, check_contracts
            server = Server("workflow-test")
            with open("contracts.json") as source:
                contracts = json.load(source)
            @server.tool("workflow_probe", "Exercise real reverse dispatch", {
                "type": "object", "properties": {
                    "tool": {"type": "string"},
                    "arguments": {"type": "object", "additionalProperties": True}
                }, "required": ["tool", "arguments"], "additionalProperties": False})
            async def probe(connection, arguments):
                check_contracts(contracts, await connection.list_tools())
                return await connection.call_tool(arguments["tool"], arguments["arguments"])
            server.run()
            """);
        Manual(inside, "primary", "type: primary\ntools: write_file\nloadable-skills: workflow\n", "Primary");
        Manual(inside, "workflow", "type: loadable\ntools: workflow_probe\n"
            + "mcp_servers:\n  - name: workflow\n    transport: stdio\n"
            + "    command: /usr/bin/python3\n    args: [\"-u\", \"workflow.py\"]\n", "Workflow fixture");
        var root = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1",
                sandboxRoot: inside), Path.Combine(inside, "jobs"),
                skillSpec: new SkillSpec(inside, "primary"));
            // This integration fixture supplies today's actual declaration.
            // Independent Python tests prove stale contracts refuse; this is
            // not a golden contract that would detect future schema drift.
            var write = agent.Tools().Single(tool => tool.Name() == "write_file");
            var schema = write.Declaration(Vendor.Anthropic)["input_schema"];
            File.WriteAllText(Path.Combine(inside, "contracts.json"),
                JsonSerializer.Serialize(new Dictionary<string, object?> { ["write_file"] = schema }));
            var loaded = await agent.DispatchAsync(SkillCall(agent, "load_skill", "workflow"))
                .WaitAsync(TimeSpan.FromSeconds(15));
            Check(!loaded.IsError, "Python workflow load failed: " + JsonCodec.Text(loaded.Parts));
            var client = agent.SkillManager()!.Clients().Single();
            async Task<JsonElement> Invoke(string tool, object arguments)
            {
                return await client.RequestAsync("tools/call", JsonSerializer.SerializeToElement(new
                {
                    name = "workflow_probe",
                    arguments = new { tool, arguments }
                })).WaitAsync(TimeSpan.FromSeconds(15));
            }
            string target = Path.Combine(inside, "written.txt");
            // Admit the outer workflow through ordinary tool dispatch too.
            // Its suspended Python handler must not block reverse Agent work.
            var allowed = await agent.DispatchAsync(new ToolCallPart("python-workflow", agent.Configuration().Target(),
                "workflow_probe", JsonSerializer.SerializeToElement(new
                {
                    tool = "write_file",
                    arguments = new { path = target, content = "through Python" }
                }))).WaitAsync(TimeSpan.FromSeconds(15));
            Check(!allowed.IsError, "Allowed workflow refused: " + JsonCodec.Text(allowed.Parts));
            Check(File.ReadAllText(target) == "through Python", "Reverse write did not execute");
            var refused = await Invoke("write_file", new { path = outside, content = "changed", overwrite = true });
            Check(refused.GetProperty("isError").GetBoolean(), "Outside reverse write was allowed");
            Check(File.ReadAllText(outside) == "unchanged", "Outside sentinel changed");
            var ungranted = await Invoke("read_file", new { path = target });
            Check(ungranted.GetProperty("isError").GetBoolean(), "Ungranted reverse tool was allowed");
            var unloaded = await agent.DispatchAsync(SkillCall(agent, "unload_skill", "workflow"));
            Check(!unloaded.IsError && agent.SkillManager()!.Clients().IsEmpty,
                "Python workflow unload did not join/detach its connection");
        }
        finally
        {
            root.Destroy();
            Directory.Delete(tree, true);
        }
    }
}
