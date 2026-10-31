using Ensemble;
using Application = Ensemble.Ensemble;

static partial class McpTests
{
    public static async Task ChildSkills(string directory)
    {
        await using var peer = new HttpPeer();
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        string home = Path.Combine(directory, "child-skill-clients");
        string manuals = Path.Combine(home, "manuals");
        Manual(manuals, "primary", "type: primary\nloadable-skills: remote\n", "Primary");
        Manual(manuals, "remote", "type: loadable\ntools: remote_echo\nmcp_servers:\n  - name: peer\n    transport: http\n    url: "
            + peer.Url + "\n    auth-env: CHILD_TEST_TOKEN\n", "Remote $TOOLS");
        var log = new StringWriter();
        var app = new Application(log);
        try
        {
            var parent = new Agent(app, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"),
                dataDirectory: Path.Combine(home, "state"),
                skillSpec: new SkillSpec(manuals, "primary", StartupSkills: ["remote"],
                    McpEnvironment: new Dictionary<string, string> { ["CHILD_TEST_TOKEN"] = peer.Token }));
            parent.StartPersistence(new PersistenceOptions(Path.Combine(home, "state", "save.json")));
            await parent.SkillManager()!.InitializeAsync(deadline.Token);
            Directory.Delete(manuals, recursive: true);
            var child = await parent.CreateChildAsync("remote-worker", cancellation: deadline.Token);
            Check(child.SkillManager()!.Ready() && child.ToolEnabled("remote_echo")
                && child.SkillManager()!.Clients().Length == 1 && parent.SkillManager()!.Clients().Length == 1,
                "Child startup did not initialize independent MCP ownership");
            Check(child.SkillManager()!.Clients()[0] != parent.SkillManager()!.Clients()[0],
                "Child borrowed the parent's live MCP client");
            var result = await child.DispatchAsync(new ToolCallPart("child-call", child.Configuration().Target(),
                "remote_echo", J("{}")), deadline.Token);
            Check(!result.IsError && JsonCodec.Text(result.Parts).Contains("campaign=kept"), "Child remote tool did not execute");
            lock (peer.Requests)
            {
                Check(peer.Requests.All(request => request.Auth == "Bearer " + peer.Token),
                    "Explicit inherited MCP credential was not used");
            }
            string history = string.Join("\n", child.History().Events().Select(JsonCodec.WriteEvent));
            Check(!history.Contains(peer.Token) && !log.ToString().Contains(peer.Token), "Inherited credential leaked");
            child.Destroy();
            Check(parent.SkillManager()!.Clients().Length == 1 && parent.ToolEnabled("remote_echo"),
                "Child destruction detached parent client or grant");
            var parentResult = await parent.DispatchAsync(new ToolCallPart("parent-still-live", parent.Configuration().Target(),
                "remote_echo", J("{}")), deadline.Token);
            Check(!parentResult.IsError, "Parent client stopped with child");

            // Cancel only after the peer observes real startup, then require the
            // creating operation to join teardown before releasing this local ID.
            peer.HoldInitialization = true;
            using var cancelling = CancellationTokenSource.CreateLinkedTokenSource(deadline.Token);
            var pending = parent.CreateChildAsync("cancel-startup", cancellation: cancelling.Token);
            await peer.Held.Task.WaitAsync(deadline.Token);
            cancelling.Cancel();
            bool cancelled = false;
            try
            {
                await pending.WaitAsync(deadline.Token);
            }
            catch (OperationCanceledException) when (cancelling.IsCancellationRequested)
            {
                cancelled = true;
            }
            Check(cancelled && parent.Children().IsEmpty && app.Agents().Length == 1
                && parent.SkillManager()!.Clients().Length == 1,
                "Cancelled initialization returned before child cleanup or affected parent");
            peer.HoldInitialization = false;
            var retry = await parent.CreateChildAsync("cancel-startup", cancellation: deadline.Token);
            Check(retry.SkillManager()!.Ready(), "Cancelled creation prevented a clean same-ID retry");
        }
        finally
        {
            app.Destroy();
        }
    }
}
