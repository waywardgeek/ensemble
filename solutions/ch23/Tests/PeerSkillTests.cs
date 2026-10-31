using Ensemble;
using Application = Ensemble.Ensemble;

static partial class McpTests
{
    public static async Task PeerStartup(string directory)
    {
        await using var remote = new HttpPeer();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        string home = Path.Combine(directory, "peer-startup");
        string manuals = Path.Combine(home, "manuals");
        Manual(manuals, "primary", "type: primary\nloadable-skills: remote\n", "Primary");
        Manual(manuals, "remote", "type: loadable\ntools: remote_echo\nmcp_servers:\n  - name: peer\n    transport: http\n    url: "
            + remote.Url + "\n    auth-env: PEER_TEST_TOKEN\n", "Remote $TOOLS");
        var app = new Application(TextWriter.Null, dataDirectory: Path.Combine(home, "ensemble"));
        try
        {
            var root = new Agent(app, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"),
                name: "root", skillSpec: new SkillSpec(manuals, "primary", StartupSkills: ["remote"],
                    McpEnvironment: new Dictionary<string, string> { ["PEER_TEST_TOKEN"] = remote.Token }));
            await root.SkillManager()!.InitializeAsync(deadline.Token);
            remote.HoldInitialization = true;
            using var cancelling = CancellationTokenSource.CreateLinkedTokenSource(deadline.Token);
            var pending = app.CreatePeerAsync(root, "cancelled", cancellation: cancelling.Token);
            await remote.Held.Task.WaitAsync(deadline.Token);
            cancelling.Cancel();
            bool cancelled = false;
            try
            {
                await pending.WaitAsync(deadline.Token);
            }
            catch (OperationCanceledException) when (cancelling.IsCancellationRequested && !deadline.IsCancellationRequested)
            {
                cancelled = true;
            }
            Check(cancelled && app.Agents().Length == 1 && root.SkillManager()!.Clients().Length == 1
                && !File.Exists(Path.Combine(app.DataDirectory(), "ensemble.json")),
                "Cancelled peer startup escaped joined rollback or published a registry.");
            remote.HoldInitialization = false;
            var peer = await app.CreatePeerAsync(root, "cancelled", cancellation: deadline.Token);
            Check(peer.SkillManager()!.Ready() && peer.SkillManager()!.Clients()[0] != root.SkillManager()!.Clients()[0],
                "Retry failed or borrowed the source MCP connection.");
        }
        finally
        {
            app.Destroy();
        }

        // Use a fresh fake so its observed initialize gate cannot be a stale
        // signal from the cancellation test above.
        await using var held = new HttpPeer();
        string closing = Path.Combine(directory, "peer-closing");
        string closingManuals = Path.Combine(closing, "manuals");
        Manual(closingManuals, "primary", "type: primary\nmcp_servers:\n  - name: peer\n    transport: http\n    url: "
            + held.Url + "\n    auth-env: PEER_TEST_TOKEN\n", "Primary $TOOLS");
        var closingApp = new Application(TextWriter.Null, dataDirectory: Path.Combine(closing, "ensemble"));
        try
        {
            var root = new Agent(closingApp, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"),
                name: "root", skillSpec: new SkillSpec(closingManuals, "primary",
                    McpEnvironment: new Dictionary<string, string> { ["PEER_TEST_TOKEN"] = held.Token }));
            await root.SkillManager()!.InitializeAsync(deadline.Token);
            held.HoldInitialization = true;
            var starting = closingApp.CreatePeerAsync(root, "closing", cancellation: deadline.Token);
            await held.Held.Task.WaitAsync(deadline.Token);
            var queued = closingApp.CreatePeerAsync(root, "queued", cancellation: deadline.Token);
            await Task.Run(closingApp.Destroy).WaitAsync(deadline.Token);
            foreach (var operation in new[] { starting, queued })
            {
                bool cancelled = false;
                try
                {
                    await operation.WaitAsync(deadline.Token);
                }
                catch (OperationCanceledException) when (!deadline.IsCancellationRequested)
                {
                    cancelled = true;
                }
                Check(cancelled, "Application destruction did not cancel pending peer creation.");
            }
            Check(closingApp.Agents().IsEmpty && !File.Exists(Path.Combine(closingApp.DataDirectory(), "ensemble.json")),
                "Application destruction returned before startup cleanup or persisted an unfinished peer.");
        }
        finally
        {
            closingApp.Destroy();
        }
    }
}
