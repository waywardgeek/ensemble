using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static class SafetyTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition)
        {
            throw new Exception(message);
        }
    }

    public static async Task SafeMode()
    {
        string directory = Path.Combine(Path.GetTempPath(), "safe-policy-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(directory);
        var root = new Application(TextWriter.Null);
        try
        {
            var config = new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1",
                safeMode: true, deferredTools: true);
            var agent = new Agent(root, config, Path.Combine(directory, "jobs"));
            agent.InstallTools();
            Check(!config.EnableWebSearch, "Library omission must not grant web-search");
            Check(!agent.ToolEnabled("run_command") && !agent.ToolEnabled("send_input"), "Shell policy missing");
            string file = Path.Combine(directory, "allowed.txt");
            var write = new ToolCallPart("write", config.Target(), "write_file",
                JsonSerializer.SerializeToElement(new { path = file, content = "allowed" }));
            Check(!(await agent.DispatchAsync(write)).IsError && File.ReadAllText(file) == "allowed",
                "Safe mode must still allow real file work");

            string marker = Path.Combine(directory, "forbidden.txt");
            var command = new ToolCallPart("shell", config.Target(), "run_command",
                JsonSerializer.SerializeToElement(new { command = "touch '" + marker + "'" }));
            var refused = await agent.DispatchAsync(command);
            Check(refused.IsError && JsonCodec.Text(refused.Parts).Contains("not enabled") && !File.Exists(marker),
                "Shell dispatch must refuse before effects");
            // Refuse every shell/job-control entry point before looking up a
            // handle. Hiding run_command alone leaves saved grants too broad.
            string[] forbidden = ["run_command", "send_input", "wait_for_job", "kill_job"];
            foreach (string name in forbidden)
            {
                Check(!agent.ToolEnabled(name), $"Safe mode still enables {name}");
                var control = await agent.DispatchAsync(new ToolCallPart(name, config.Target(), name,
                    JsonSerializer.SerializeToElement(new { handle = 999, input = "x" })));
                Check(control.IsError && JsonCodec.Text(control.Parts).Contains("not enabled"),
                    $"{name} policy must precede argument/handle lookup");
            }

            // Saved grants and discovery references are evidence, not permission
            // to widen a later launch. Render a genuine paired exchange.
            var context = new Context([
                new Entry(1, Actor.System, [new SkillPart("primary", "Retained manual", ["run_command", "send_input", "wait_for_job", "kill_job", "read_file"], primary: true)]),
                new Entry(2, Actor.Agent, [new ToolCallPart("discover", config.Target(), "read_file", JsonSerializer.SerializeToElement(new { path = file }))]),
                new Entry(3, Actor.Tool, [new ToolResultPart("discover", [new TextPart("catalog")], toolReferences: ["run_command", "send_input", "wait_for_job", "kill_job", "read_file"])])
            ]);
            string wire = agent.Engine().Render(context).ToJsonString();
            Check(forbidden.All(name => !wire.Contains(name))
                && wire.Contains("read_file") && wire.Contains("Retained manual"),
                "Declarations/references must narrow without erasing manuals");
            Check(context.EnabledTools().Contains("run_command") && !context.EnabledTools(config).Contains("run_command"),
                "Policy projection must not mutate recorded grants");
            await agent.SetModelAsync("another-local-model");
            Check(agent.Configuration().SafeMode && !agent.Configuration().EnableWebSearch
                && !agent.ToolEnabled("run_command"), "Model selection widened launch policy");
        }
        finally
        {
            root.Destroy();
        }
    }
}

static partial class McpTests
{
    public static async Task WebPolicy()
    {
        await using var peer = new HttpPeer();
        string directory = Path.Combine(Path.GetTempPath(), "web-policy-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(directory);
        Manual(directory, "primary", "type: primary\nloadable-skills: web-search bridge\n", "Available: $SKILLS\nTools: $TOOLS");
        Manual(directory, "web-search", "type: loadable\ntools: remote_echo\nmcp_servers:\n  - name: web-peer\n    transport: http\n    url: " + peer.Url + "\n", "Remote manual");
        Manual(directory, "bridge", "type: loadable\ndepends: web-search\n", "Bridge manual");
        var root = new Application(TextWriter.Null);
        try
        {
            var disabled = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1", deferredTools: true),
                skillSpec: new SkillSpec(directory, "primary"));
            Check(!disabled.SkillManager()!.LoadableSkills().Contains("web-search"), "Disabled web skill advertised");
            Check(!disabled.Engine().Render(disabled.History().View()).ToJsonString().Contains("remote_echo"), "Disabled deferred tool advertised");
            foreach (string name in new[] { "web-search", "bridge" })
            {
                var refused = await disabled.DispatchAsync(SkillCall(disabled, "load_skill", name));
                Check(refused.IsError && !disabled.ToolEnabled("remote_echo"), "Direct/dependency load widened web authority");
            }
            Check(peer.Count == 0, "Disabled activation contacted MCP peer");
            try
            {
                _ = new Agent(root, disabled.Configuration(), skillSpec: new SkillSpec(directory, "primary", StartupSkills: ["bridge"]));
                throw new Exception("Startup dependency bypassed web policy");
            }
            catch (InvalidDataException) { }
            Check(peer.Count == 0, "Failed startup contacted MCP peer");

            var saved = new Context([
                new Entry(1, Actor.System, [new SkillPart("web-search", "Saved web manual", ["remote_echo"])])
            ]);
            Check(saved.EnabledTools().Contains("remote_echo") && !saved.EnabledTools(disabled.Configuration()).Contains("remote_echo"),
                "Saved web grants widened current policy");
            Check(saved.RevokedTools(disabled.Configuration()).Contains("remote_echo"), "Saved discovery reference remains authorized");
            string replay = disabled.Engine().Render(saved).ToJsonString();
            Check(replay.Contains("Saved web manual") && !replay.Contains("remote_echo"), "Replay erased manual or advertised disabled tool");

            // Positive control: the same files and real HTTP transport work when
            // explicitly authorized. Denying everything would fail this check.
            var enabled = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1", enableWebSearch: true),
                skillSpec: new SkillSpec(directory, "primary"));
            var loaded = await enabled.DispatchAsync(SkillCall(enabled, "load_skill", "web-search")).WaitAsync(TimeSpan.FromSeconds(10));
            Check(!loaded.IsError && enabled.ToolEnabled("remote_echo") && peer.Count > 0, "Explicit web activation failed");
            var echo = await enabled.DispatchAsync(new ToolCallPart("web", enabled.Configuration().Target(), "remote_echo", J("{}")));
            Check(!echo.IsError, "Authorized HTTP tool failed");
            await enabled.SetModelAsync("another-local-model");
            Check(enabled.Configuration().EnableWebSearch && enabled.ToolEnabled("remote_echo"), "Model change lost explicit web authority");
            Check(!disabled.ToolEnabled("remote_echo"), "Sibling activation widened disabled Agent");
        }
        finally
        {
            root.Destroy();
        }
    }
}
