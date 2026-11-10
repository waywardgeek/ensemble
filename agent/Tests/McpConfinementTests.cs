using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

static partial class McpTests
{
    // A real stdio peer and its descendant attempt harmless operations against
    // disposable files and a listening local socket. Only permission refusal
    // counts: absent files and closed ports would manufacture a passing test.
    public static async Task ConfinedProcess()
    {
        string tree = Path.Combine(Path.GetTempPath(), "mcp-confined-" + Guid.NewGuid().ToString("N"));
        string inside = Path.Combine(tree, "inside");
        Directory.CreateDirectory(inside);
        string sentinel = Path.Combine(tree, "outside.txt");
        File.WriteAllText(sentinel, "outside-sentinel");
        File.WriteAllText(Path.Combine(inside, "probe.py"), """
            import errno, json, os, socket, subprocess, sys

            def probe(outside, port):
                result = {}
                with open("inside.txt", "w") as output:
                    output.write("allowed")
                with open("inside.txt") as source:
                    result["inside"] = source.read() == "allowed"
                try:
                    with open(outside) as source:
                        source.read()
                    result["outside_denied"] = False
                except OSError as error:
                    result["outside_denied"] = error.errno in (errno.EPERM, errno.EACCES)
                try:
                    with socket.socket() as connection:
                        connection.settimeout(2)
                        connection.connect(("127.0.0.1", int(port)))
                    result["network_denied"] = False
                except OSError as error:
                    result["network_denied"] = error.errno in (errno.EPERM, errno.EACCES)
                return result

            if len(sys.argv) > 1:
                print(json.dumps(probe(sys.argv[1], sys.argv[2])), flush=True)
                sys.exit(0)

            for line in sys.stdin:
                request = json.loads(line)
                method = request["method"]
                if method == "notifications/initialized":
                    continue
                if method == "initialize":
                    result = {"protocolVersion": "2024-11-05", "capabilities": {}}
                elif method == "tools/list":
                    result = {"tools": []}
                elif method == "probe":
                    args = request["params"]
                    result = {"self": probe(args["outside"], args["port"])}
                    child = subprocess.run(
                        [sys.executable, "probe.py", args["outside"], str(args["port"])],
                        capture_output=True, text=True, check=True)
                    result["child"] = json.loads(child.stdout)
                else:
                    raise AssertionError(method)
                print(json.dumps({"jsonrpc": "2.0", "id": request["id"], "result": result}), flush=True)
            """);
        Manual(inside, "primary", "type: primary\nloadable-skills: confined\n", "Primary");
        Manual(inside, "confined", "type: loadable\n"
            + "mcp_servers:\n  - name: confined\n    transport: stdio\n"
            + "    command: /usr/bin/python3\n    args: [\"-u\", \"probe.py\"]\n", "Confined peer");
        var root = new Application(TextWriter.Null);
        var listener = new TcpListener(IPAddress.Loopback, 0);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1",
                sandboxRoot: inside), Path.Combine(inside, "jobs"),
                skillSpec: new SkillSpec(inside, "primary"));
            var specification = new McpServerSpec("confined", "stdio", "/usr/bin/python3", ["-u", "probe.py"]);
            if (!OperatingSystem.IsMacOS())
            {
                bool refused = false;
                try
                {
                    await using var unexpected = new McpProcessTransport(specification, inside, agent.Sandbox());
                }
                catch (PlatformNotSupportedException)
                {
                    refused = true;
                }
                Check(refused, "Unsupported confined MCP launch did not refuse");
                return;
            }
            listener.Start();
            int port = ((IPEndPoint)listener.LocalEndpoint).Port;
            // Actual skill admission must select the confined workspace and
            // supply this Agent's Sandbox, not merely a manual test constructor.
            var loaded = await agent.DispatchAsync(SkillCall(agent, "load_skill", "confined"))
                .WaitAsync(TimeSpan.FromSeconds(15));
            Check(!loaded.IsError, "Confined skill failed to load: " + JsonCodec.Text(loaded.Parts));
            var client = agent.SkillManager()!.Clients().Single();
            var arguments = JsonSerializer.SerializeToElement(new { outside = sentinel, port });
            var result = await client.RequestAsync("probe", arguments).WaitAsync(TimeSpan.FromSeconds(15));
            foreach (string role in new[] { "self", "child" })
            {
                JsonElement observed = result.GetProperty(role);
                Check(observed.GetProperty("inside").GetBoolean(), role + " lost sandbox workspace access");
                Check(observed.GetProperty("outside_denied").GetBoolean(), role + " read outside sandbox");
                Check(observed.GetProperty("network_denied").GetBoolean(), role + " reached local listener");
            }
            Check(File.ReadAllText(sentinel) == "outside-sentinel", "Outside sentinel changed");
            var unloaded = await agent.DispatchAsync(SkillCall(agent, "unload_skill", "confined"));
            Check(!unloaded.IsError && agent.SkillManager()!.Clients().IsEmpty,
                "Confined skill unload did not detach its joined connection");
        }
        finally
        {
            // Even assertion/timeout failures close and join the owned peer.
            root.Destroy();
            listener.Stop();
            Directory.Delete(tree, true);
        }
    }
}
