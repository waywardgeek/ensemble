using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class MeterTests
{
    // The actual confined Python server calls ordinary child tools. The local
    // provider supplies final text; public submission below isolates dispatch and
    // persistence from model tool-use ability. No paid inference occurs.
    public static async Task FictionNativeAssignments(string directory)
    {
        if (!OperatingSystem.IsMacOS())
        {
            throw new PlatformNotSupportedException("Workflow fixture requires macOS confinement");
        }
        string home = Path.Combine(directory, "fiction-native-assignments");
        Directory.CreateDirectory(home);
        string run = Path.Combine(home, "run");
        Directory.CreateDirectory(run);
        // Auxiliary recall is outside this dispatch fixture. Seed its policy
        // before spawn so only the two main child requests reach the fake.
        string childHome = Path.Combine(home, "children", "writer");
        Directory.CreateDirectory(childHome);
        var settings = new System.Text.Json.Nodes.JsonObject
        {
            ["version"] = 1,
            ["memory"] = JsonCodec.WriteMemorySettings(new MemorySettings()),
            ["recall"] = JsonCodec.WriteRecallSettings(new RecallSettings(disabled: true))
        };
        File.WriteAllText(Path.Combine(childHome, "settings.json"), settings.ToJsonString());
        File.Copy("agent/skills/shared/ensemble_mcp.py", Path.Combine(home, "ensemble_mcp.py"));
        File.Copy("agent/skills/fiction-workflow/contracts.json", Path.Combine(home, "contracts.json"));
        foreach (string name in new[] { "assignments.py", "child_calls.py", "run_state.py" })
        {
            File.Copy(Path.Combine("agent/skills/fiction-workflow", name), Path.Combine(home, name));
        }
        File.WriteAllText(Path.Combine(home, "workflow.py"), """
            import json
            from pathlib import Path
            from ensemble_mcp import Server, check_contracts
            from assignments import dispatch
            from run_state import RunStore
            server = Server("assignment-fixture")
            @server.tool("assignment_probe", "Exercise two owned-child assignments", {
                "type": "object", "properties": {}, "additionalProperties": False})
            async def probe(connection, arguments):
                check_contracts(json.loads(Path("contracts.json").read_text()),
                                await connection.list_tools())
                with RunStore("run") as store:
                    for turn in (1, 2):
                        schema = {"type": "object", "properties": {
                            "assignment_id": {"type": "string"}},
                            "required": ["assignment_id"], "additionalProperties": False}
                        payload = await dispatch(connection, store, child_id="writer",
                            phase="write", round_number=0, chapter=1, turn=turn,
                            task="Submit this fixture assignment; do not call other tools.",
                            new_child=turn == 1, submit_schema=schema if turn == 1 else None)
                        Path("run/result.txt").write_text(str(turn))
                        store.accept(payload, ["result.txt"])
                    return {"content": [{"type": "text", "text": json.dumps(store.snapshot())}]}
            server.run()
            """);
        McpTests.Manual(home, "primary", "type: primary\n"
            + "tools: spawn_sub_agent ask_agent get_submission wait_for_job agents_status shutdown_agent\n"
            + "loadable-skills: workflow\n", "Local child dispatch fixture");
        McpTests.Manual(home, "workflow", "type: loadable\ntools: assignment_probe\n"
            + "mcp_servers:\n  - name: assignments\n    transport: stdio\n"
            + "    command: /usr/bin/python3\n    args: [\"-u\", \"workflow.py\"]\n", "Fixture only");
        using var listener = Listen();
        var root = new Application(TextWriter.Null);
        try
        {
            var parent = new Agent(root, new Configuration(Vendor.OpenAI, "fixture",
                listener.Prefixes.Single(), "local-fixture", disableStreaming: true,
                sandboxRoot: home), dataDirectory: home, skillSpec: new SkillSpec(home, "primary"));
            var loaded = await parent.DispatchAsync(new ToolCallPart("load-workflow",
                parent.Configuration().Target(), "load_skill",
                JsonSerializer.SerializeToElement(new { name = "workflow" })))
                .WaitAsync(TimeSpan.FromSeconds(15));
            Check(!loaded.IsError, "Native assignment fixture failed to load: " + JsonCodec.Text(loaded.Parts));
            var client = parent.SkillManager()!.Clients().Single();
            var invocation = client.RequestAsync("tools/call", JsonSerializer.SerializeToElement(new
            {
                name = "assignment_probe", arguments = new { }
            }));
            for (int turn = 1; turn <= 2; turn++)
            {
                var request = await listener.GetContextAsync().WaitAsync(TimeSpan.FromSeconds(15));
                using var reader = new StreamReader(request.Request.InputStream);
                string wire = await reader.ReadToEndAsync();
                using var ledger = JsonDocument.Parse(File.ReadAllText(Path.Combine(run, "state.json")));
                var pending = ledger.RootElement.GetProperty("pending");
                string assignment = pending.GetProperty("id").GetString()!;
                Check(pending.GetProperty("turn").GetInt32() == turn,
                    "Wrong persisted native turn: expected " + turn + " actual "
                    + pending.GetProperty("turn").GetInt32());
                Check(wire.Contains(assignment, StringComparison.Ordinal), "Assignment identity missing from child request");
                var child = parent.FindChild("writer") ?? throw new Exception("Native child was not created");
                if (turn == 1)
                {
                    // A pending child question must wake supervision while the
                    // Python handler remains suspended in the original spawn.
                    using var questionDeadline = new CancellationTokenSource(TimeSpan.FromSeconds(10));
                    var announced = parent.WaitForChildChangeAsync(["writer"], timeoutSeconds: 1200,
                        stuckSeconds: 0, cancellation: questionDeadline.Token);
                    var answer = child.SendMessageToParentAsync("Which voice should I use?", questionDeadline.Token);
                    var change = await announced.WaitAsync(questionDeadline.Token);
                    Check(change.Reason == "question", "Long supervision wait did not yield for question");
                    var questions = change.Questions;
                    Check(questions.Single().AgentID == "writer", "Workflow question lost its owner");
                    Check(!invocation.IsCompleted, "Question prematurely completed workflow");
                    parent.RespondToQuestion("Use the supplied voice guide.", "writer", questions.Single().MessageID);
                    Check((await answer).Response == "Use the supplied voice guide.", "Workflow answer was lost");
                }
                child.PublishSubmission(JsonSerializer.SerializeToElement(new { assignment_id = assignment }));
                await Reply(request, JsonSerializer.Serialize(new
                {
                    choices = new[] { new { message = new { content = "Fixture submitted" }, finish_reason = "stop" } }
                }));
            }
            JsonElement result = await invocation.WaitAsync(TimeSpan.FromSeconds(15));
            Check(!result.TryGetProperty("isError", out var error) || !error.GetBoolean(),
                "Native Python assignment returned an error: " + result);
            using var final = JsonDocument.Parse(File.ReadAllText(Path.Combine(run, "state.json")));
            Check(final.RootElement.GetProperty("pending").ValueKind == JsonValueKind.Null,
                "Native assignment remained pending");
            Check(final.RootElement.GetProperty("accepted").GetArrayLength() == 2,
                "Native assignments did not produce two accepted receipts");
            Check(parent.Children().Length == 1, "Continued assignment created another child");
        }
        finally
        {
            // Join children and owned MCP processes even if the fixture fails.
            root.Destroy();
        }
    }
}
