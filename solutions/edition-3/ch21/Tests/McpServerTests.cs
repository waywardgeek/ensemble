using System.Collections.Immutable;
using System.Diagnostics;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

static partial class McpTests
{
    public static async Task ServerSpecifications()
    {
        const string header = "---\nname: servers\ntype: loadable\n";
        const string body = "---\nManual keeps $TOOLS\n---\nremaining body";
        string block = "mcp_servers:\n  - name: local\n    transport: stdio\n    command: python3\n"
            + "    args: [\"scripts/peer.py\", \"two words\", \"$HOME;*.txt\", \"\"]\n"
            + "  - name: browser\n    transport: websocket\n";
        var skill = Skills.Parse(header + block + "tools: read_file\n" + body, "fixture");
        Check(skill.McpServers.Length == 2 && skill.Tools.SequenceEqual(new[] { "read_file" }),
            "Nested server parsing consumed a following top-level field");
        var local = skill.McpServers[0];
        Check(local.Name == "local" && local.Transport == "stdio" && local.Command == "python3",
            "Server launch fields changed");
        Check(local.Arguments.SequenceEqual(new[] { "scripts/peer.py", "two words", "$HOME;*.txt", "" }),
            "Arguments were expanded, split, or lost");
        Check(skill.McpServers[1].Command is null && skill.McpServers[1].Arguments.IsEmpty,
            "Websocket specification acquired process fields");
        Check(skill.Body == "Manual keeps $TOOLS\n---\nremaining body", "Server parser changed manual body");
        var trailingSpace = Skills.Parse(header + block.Replace("mcp_servers:\n", "mcp_servers:  \n") + body, "fixture");
        Check(trailingSpace.McpServers.Length == 2, "Harmless trailing field whitespace refused");
        var plain = Skills.Parse(header + body, "fixture");
        Check(plain.McpServers.IsEmpty, "Ordinary skill acquired server configuration");

        string[] invalid =
        [
            "mcp_servers:\n",
            "mcp_servers:\n  name: local\n",
            "mcp_servers:\n  - name: local\n    transport: http\n",
            "mcp_servers:\n  - name: local\n    transport: stdio\n",
            "mcp_servers:\n  - name: browser\n    transport: websocket\n    command: python3\n",
            "mcp_servers:\n  - name: browser\n    transport: websocket\n    args: []\n",
            "mcp_servers:\n  - name: local\n    transport: stdio\n    command: python3\n    args: [1]\n",
            "mcp_servers:\n  - name: local\n    transport: stdio\n    command: python3\n    args: [unquoted]\n",
            "mcp_servers:\n  - name: local\n    transport: stdio\n    command: python3\n    args: [\"\\u0000\"]\n",
            "mcp_servers:\n  - name: local\n    transport: websocket\n    transport: websocket\n",
            "mcp_servers:\n  - name: local\n    transport: websocket\n  - name: local\n    transport: websocket\n",
            "mcp_servers:\n  - name: local\n    transport: websocket\n    unknown: value\n",
            "mcp_servers:\n  - name: local\n    transport: websocket\n   command: bad\n",
            "mcp_servers:\n\t- name: local\n    transport: websocket\n"
        ];
        foreach (string value in invalid)
        {
            bool refused = false;
            try
            {
                Skills.Parse(header + value + body, "fixture");
            }
            catch (InvalidDataException)
            {
                refused = true;
            }
            Check(refused, "Malformed server configuration accepted: " + value);
        }
        string directory = Path.Combine(Path.GetTempPath(), "mcp-skill-refusal-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(Path.Combine(directory, "primary"));
        Directory.CreateDirectory(Path.Combine(directory, "remote"));
        File.WriteAllText(Path.Combine(directory, "primary", "SKILL.md"),
            "---\nname: primary\ntype: primary\nloadable-skills: remote\n---\nPrimary");
        File.WriteAllText(Path.Combine(directory, "remote", "SKILL.md"),
            "---\nname: remote\ntype: loadable\nmcp_servers:\n  - name: remote-peer\n"
            + "    transport: stdio\n    command: /no-such-mcp-server\n"
            + "    args: [\"-c\", \"exit 97\"]\n---\nRemote instructions");
        var root = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"),
                skillSpec: new SkillSpec(directory, "primary"));
            var before = agent.SkillManager()!.Active();
            var call = new ToolCallPart("refused-server-load", agent.Configuration().Target(), "load_skill", J("{\"name\":\"remote\"}"));
            var result = await agent.DispatchAsync(call);
            Check(result.IsError && agent.SkillManager()!.Active().SequenceEqual(before),
                "Failed server launch granted a skill");
            Check(agent.Mcp()!.Clients().IsEmpty && agent.SkillManager()!.Clients().IsEmpty,
                "Failed launch retained a connection");
        }
        finally
        {
            root.Destroy();
            Directory.Delete(directory, true);
        }
    }

    public static async Task ProcessBoundary()
    {
        string directory = Path.Combine(Path.GetTempPath(), "mcp-process-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(directory);
        string script = Path.Combine(directory, "peer.py");
        // A working external server: it checks handshake order, floods stderr
        // beyond pipe capacity, then speaks ordinary correlated JSON-RPC.
        File.WriteAllText(script, """
            import json, os, sys
            state = 0
            sys.stderr.write("diagnostic-only\n" * 20000)
            sys.stderr.flush()
            for line in sys.stdin:
                request = json.loads(line)
                method = request["method"]
                if method == "initialize":
                    assert state == 0
                    state = 1
                    result = {"protocolVersion": "2024-11-05", "capabilities": {}}
                elif method == "notifications/initialized":
                    assert state == 1
                    state = 2
                    continue
                elif method == "tools/list":
                    assert state == 2
                    state = 3
                    result = {"tools": []}
                elif method == "inspect":
                    assert state == 3
                    result = {"pid": os.getpid(), "cwd": os.getcwd(), "args": sys.argv[1:], "env": dict(os.environ)}
                elif method == "hold":
                    open("held", "w").close()
                    continue
                elif method == "exit":
                    break
                else:
                    raise AssertionError(method)
                print(json.dumps({"jsonrpc": "2.0", "id": request["id"], "result": result}), flush=True)
            """);
        var root = new Application(TextWriter.Null);
        string? previous = Environment.GetEnvironmentVariable("ANTHROPIC_API_KEY");
        string? previousCustom = Environment.GetEnvironmentVariable("MCP_TEST_PRIVATE");
        try
        {
            Environment.SetEnvironmentVariable("ANTHROPIC_API_KEY", "dummy-never-forward");
            Environment.SetEnvironmentVariable("MCP_TEST_PRIVATE", "dummy-private");
            bool startupRefused = false;
            try
            {
                await using var invalid = new McpProcessTransport(
                    new McpServerSpec("missing", "stdio", "/no-such-mcp-server", []), directory);
            }
            catch (System.ComponentModel.Win32Exception)
            {
                startupRefused = true;
            }
            Check(startupRefused, "Missing server executable was not refused");
            var specification = new McpServerSpec("peer", "stdio", "/usr/bin/python3",
                ["-u", "peer.py", "two words", "$HOME;*.txt", ""]);
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"));
            var manager = new Mcp(agent);
            var transport = new McpProcessTransport(specification, directory);
            var client = new McpClient(manager, "peer", transport);
            await client.ConnectAsync().WaitAsync(TimeSpan.FromSeconds(10));
            var inspected = await client.RequestAsync("inspect", J("{}")).WaitAsync(TimeSpan.FromSeconds(8));
            Check(inspected.GetProperty("cwd").GetString() == Path.GetFullPath(directory), "Wrong process workspace");
            string[] arguments = inspected.GetProperty("args").EnumerateArray().Select(value => value.GetString()!).ToArray();
            Check(arguments.SequenceEqual(new[] { "two words", "$HOME;*.txt", "" }), "Process arguments were shell-expanded");
            var environment = inspected.GetProperty("env");
            Check(!environment.TryGetProperty("ANTHROPIC_API_KEY", out _)
                && !environment.TryGetProperty("MCP_TEST_PRIVATE", out _), "Process inherited private environment");
            Check(environment.TryGetProperty("PATH", out _), "Basic executable path was lost");
            using var child = Process.GetProcessById(inspected.GetProperty("pid").GetInt32());
            Task held = client.RequestAsync("hold", J("{}"));
            using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(8));
            while (!File.Exists(Path.Combine(directory, "held")))
            {
                await Task.Delay(10, deadline.Token);
            }
            await agent.ShutdownAsync().WaitAsync(TimeSpan.FromSeconds(10));
            await Fails(held);
            await Task.WhenAll(transport.DisposeAsync().AsTask(), transport.DisposeAsync().AsTask());
            Check(child.HasExited && manager.Clients().IsEmpty, "Shutdown left a process or client attached");

            // EOF is a connection failure, not a fabricated reply. Destruction
            // must also join a child that exited naturally before cleanup.
            var second = new McpClient(new Mcp(new Agent(root,
                new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"))), "exit-peer",
                new McpProcessTransport(specification, directory));
            await second.ConnectAsync().WaitAsync(TimeSpan.FromSeconds(10));
            await Fails(second.RequestAsync("exit", J("{}")));
            await second.DestroyAsync().WaitAsync(TimeSpan.FromSeconds(10));
        }
        finally
        {
            Environment.SetEnvironmentVariable("ANTHROPIC_API_KEY", previous);
            Environment.SetEnvironmentVariable("MCP_TEST_PRIVATE", previousCustom);
            root.Destroy();
            Directory.Delete(directory, true);
        }
    }
}
