using System.Collections.Immutable;
using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

static partial class McpTests
{
    // This is an OS policy test, not a profile-string assertion. The shared
    // module and single file remain outside the writable root throughout.
    public static async Task ReadOnlyProcessPaths(string directory)
    {
        string tree = Path.Combine(directory, "process-read-paths");
        string inside = Path.Combine(tree, "inside");
        string shared = Path.Combine(tree, "shared");
        Directory.CreateDirectory(inside);
        Directory.CreateDirectory(shared);
        string module = Path.Combine(shared, "lesson.py");
        string single = Path.Combine(tree, "approved.txt");
        string sibling = Path.Combine(tree, "private.txt");
        File.WriteAllText(module, "VALUE = 42\n");
        File.WriteAllText(single, "approved");
        File.WriteAllText(sibling, "private");
        Directory.CreateSymbolicLink(Path.Combine(inside, "shared"), shared);
        File.CreateSymbolicLink(Path.Combine(shared, "escape.txt"), sibling);
        File.WriteAllText(Path.Combine(inside, "probe.py"), """
            import errno, json, os, socket, subprocess, sys
            sys.dont_write_bytecode = True

            def denied(operation):
                try:
                    operation()
                    return False
                except OSError as error:
                    return error.errno in (errno.EPERM, errno.EACCES)

            def probe(args):
                sys.path.insert(0, 'shared')
                import lesson
                checks = {'import': lesson.VALUE == 42}
                with open(args['single']) as source:
                    checks['single_read'] = source.read() == 'approved'
                def write(path):
                    with open(path, 'w') as output:
                        output.write('changed')
                for label, path in [('module', args['module']), ('single', args['single'])]:
                    checks[label + '_write_denied'] = denied(lambda: write(path))
                    checks[label + '_delete_denied'] = denied(lambda: os.unlink(path))
                    checks[label + '_rename_denied'] = denied(lambda: os.rename(path, 'stolen'))
                    checks[label + '_chmod_denied'] = denied(lambda: os.chmod(path, 0o777))
                checks['create_denied'] = denied(lambda: write('shared/new.py'))
                checks['sibling_denied'] = denied(lambda: open(args['sibling']).close())
                checks['symlink_escape_denied'] = denied(lambda: open('shared/escape.txt').close())
                with open('workspace.txt', 'w') as output:
                    output.write('allowed')
                checks['workspace_write'] = open('workspace.txt').read() == 'allowed'
                def connect():
                    with socket.socket() as client:
                        client.settimeout(2)
                        client.connect(('127.0.0.1', args['port']))
                checks['network_denied'] = denied(connect)
                return checks

            if len(sys.argv) > 1:
                print(json.dumps(probe(json.loads(sys.argv[1]))), flush=True)
                sys.exit(0)
            for line in sys.stdin:
                request = json.loads(line)
                method = request['method']
                if method == 'notifications/initialized':
                    continue
                if method == 'initialize':
                    result = {'protocolVersion': '2024-11-05', 'capabilities': {}}
                elif method == 'tools/list':
                    result = {'tools': []}
                elif method == 'probe':
                    args = request['params']
                    child = subprocess.run([sys.executable, 'probe.py', json.dumps(args)],
                        capture_output=True, text=True, check=True)
                    result = {'self': probe(args), 'descendant': json.loads(child.stdout)}
                else:
                    raise AssertionError(method)
                print(json.dumps({'jsonrpc': '2.0', 'id': request['id'], 'result': result}), flush=True)
            """);
        Manual(inside, "primary", "type: primary\nloadable-skills: reader\n", "Read-only process test");
        Manual(inside, "reader", "type: loadable\n"
            + "mcp_servers:\n  - name: reader\n    transport: stdio\n"
            + "    command: /usr/bin/python3\n    args: [\"-u\", \"probe.py\"]\n", "External reader");
        var root = new Application(TextWriter.Null);
        var listener = new TcpListener(IPAddress.Loopback, 0);
        try
        {
            var policy = new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1",
                sandboxRoot: inside, processReadPaths: ImmutableArray.Create(shared, single));
            var agent = new Agent(root, policy, dataDirectory: inside,
                skillSpec: new SkillSpec(inside, "primary"));
            Check(policy.WithModel("another").ProcessReadPaths!.Value.SequenceEqual(new[] { shared, single }),
                "Model change discarded process policy");
            var inherited = agent.Sandbox().DeriveChild();
            Check(inherited.Refusal is null && inherited.Effective.ProcessReadPaths!.Value.Length == 2,
                "Child did not inherit process reads");
            var dropped = agent.Sandbox().DeriveChild(new Configuration(Vendor.Anthropic, "fake",
                "http://127.0.0.1:1", sandboxRoot: inside, processReadPaths: ImmutableArray<string>.Empty));
            Check(dropped.Refusal is null && dropped.Effective.ProcessReadPaths!.Value.IsEmpty,
                "Explicit empty process reads did not narrow policy");
            var widened = agent.Sandbox().DeriveChild(new Configuration(Vendor.Anthropic, "fake",
                "http://127.0.0.1:1", sandboxRoot: inside, processReadPaths: ImmutableArray.Create(tree)));
            Check(widened.Refusal is not null && widened.Effective.ProcessReadPaths!.Value.Length == 2,
                "Child widened inherited process reads");
            var unrelated = agent.Sandbox().DeriveChild(new Configuration(Vendor.Anthropic, "fake",
                "http://127.0.0.1:1", sandboxRoot: inside, processReadPaths: ImmutableArray.Create(sibling)));
            Check(unrelated.Refusal is not null && unrelated.Effective.ProcessReadPaths!.Value.Length == 2,
                "Child gained an unrelated non-overlapping file grant");
            var narrower = agent.Sandbox().DeriveChild(new Configuration(Vendor.Anthropic, "fake",
                "http://127.0.0.1:1", sandboxRoot: inside, processReadPaths: ImmutableArray.Create(module)));
            Check(narrower.Refusal is null && narrower.Effective.ProcessReadPaths!.Value.Length == 1,
                "Child could not narrow a directory grant to one file");
            bool relativeRefused = false;
            try
            {
                _ = new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1",
                    sandboxRoot: inside, processReadPaths: ImmutableArray.Create("../shared"));
            }
            catch (ArgumentException)
            {
                relativeRefused = true;
            }
            Check(relativeRefused, "Relative process grant was accepted");
            bool nativeReadRefused = false;
            try
            {
                agent.Sandbox().ReadFile(single);
            }
            catch (IOException)
            {
                nativeReadRefused = true;
            }
            Check(nativeReadRefused, "Process policy silently widened native file tools");
            bool nativeWriteRefused = false;
            try
            {
                agent.Sandbox().WriteFile(single, "changed"u8);
            }
            catch (IOException)
            {
                nativeWriteRefused = true;
            }
            Check(nativeWriteRefused, "Native tools wrote an external process input");
            listener.Start();
            int port = ((IPEndPoint)listener.LocalEndpoint).Port;
            var loaded = await agent.DispatchAsync(SkillCall(agent, "load_skill", "reader"))
                .WaitAsync(TimeSpan.FromSeconds(15));
            Check(!loaded.IsError, "Read-only process skill load failed: " + JsonCodec.Text(loaded.Parts));
            var arguments = JsonSerializer.SerializeToElement(new { module, single, sibling, port });
            var response = await agent.SkillManager()!.Clients().Single().RequestAsync("probe", arguments)
                .WaitAsync(TimeSpan.FromSeconds(15));
            foreach (string role in new[] { "self", "descendant" })
            {
                foreach (var check in response.GetProperty(role).EnumerateObject())
                {
                    Check(check.Value.GetBoolean(), role + " failed " + check.Name);
                }
            }
            // A real native child inherits the same process policy. No task or
            // provider is needed to exercise its graph-authorized skill launch.
            var child = await agent.CreateChildAsync("reader-child", null, null, CancellationToken.None);
            var childLoad = await child.DispatchAsync(SkillCall(child, "load_skill", "reader"));
            Check(!childLoad.IsError, "Inherited process grant did not reach native child");
            var childResponse = await child.SkillManager()!.Clients().Single().RequestAsync("probe", arguments)
                .WaitAsync(TimeSpan.FromSeconds(15));
            foreach (string role in new[] { "self", "descendant" })
            {
                foreach (var check in childResponse.GetProperty(role).EnumerateObject())
                {
                    Check(check.Value.GetBoolean(), "native child " + role + " failed " + check.Name);
                }
            }
            child.Destroy();
            Check(File.ReadAllText(module) == "VALUE = 42\n" && File.ReadAllText(single) == "approved",
                "Shared input bytes changed");
            var unloaded = await agent.DispatchAsync(SkillCall(agent, "unload_skill", "reader"));
            Check(!unloaded.IsError && agent.SkillManager()!.Clients().IsEmpty,
                "Read-only process was not joined on unload");
        }
        finally
        {
            root.Destroy();
            listener.Stop();
        }
    }
}
