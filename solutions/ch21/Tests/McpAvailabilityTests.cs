using Ensemble;
using Application = Ensemble.Ensemble;

static partial class McpTests
{
    public static async Task BrowserAvailability()
    {
        var root = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"));
            var managers = await Task.WhenAll(Enumerable.Range(0, 64).Select(_ => Task.Run(agent.EnsureMcp)));
            var manager = agent.Mcp()!;
            Check(managers.All(candidate => ReferenceEquals(candidate, manager)) && manager.Parent() == agent,
                "Concurrent lazy construction published competing MCP owners");
            var other = new Agent(root, agent.Configuration()).EnsureMcp();
            Task otherWait = other.WaitForBrowserAsync();
            using var cancel = new CancellationTokenSource();
            Task abandoned = manager.WaitForBrowserAsync(cancel.Token);
            Task waiting = manager.WaitForBrowserAsync();
            Check(!waiting.IsCompleted && !abandoned.IsCompleted, "Absent browser satisfied availability");
            cancel.Cancel();
            await Fails(abandoned);
            Check(!waiting.IsCompleted, "Observer cancellation changed other waits");
            var pair = await Pair();
            await using var peer = pair.Peer;
            var browser = new McpBrowser(manager, pair.Client);
            await waiting.WaitAsync(TimeSpan.FromSeconds(8));
            await manager.WaitForBrowserAsync().WaitAsync(TimeSpan.FromSeconds(8));
            Check(!otherWait.IsCompleted && browser.Client() is null && manager.Clients().IsEmpty,
                "Browser arrival escaped its owner or granted a client");
            await browser.DestroyAsync();
            Task afterRemoval = manager.WaitForBrowserAsync();
            Check(!afterRemoval.IsCompleted, "Removed browser left a sticky readiness signal");
            await manager.DestroyAsync();
            await Fails(afterRemoval);
            Check(agent.Mcp() is null, "Manager destruction retained reverse ownership");
            await other.DestroyAsync();
            await Fails(otherWait);
            await agent.ShutdownAsync();
            try
            {
                agent.EnsureMcp();
                throw new Exception("Closed Agent acquired a new manager");
            }
            catch (ObjectDisposedException) { }
        }
        finally
        {
            root.Destroy();
        }
    }

    public static async Task BrowserPeerStartup(string temp)
    {
        string directory = SkillFixture();
        Manual(directory, "plain", "type: primary\n", "Ordinary peer source");
        Manual(directory, "browser", "type: primary\ntools: page_echo\n" + BrowserServer, "Independent browser");
        var root = new Application(TextWriter.Null, dataDirectory: Path.Combine(temp, "browser-peers"));
        Task? serving = null;
        try
        {
            var source = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"),
                name: "source", skillSpec: new SkillSpec(directory, "plain"));
            await source.SkillManager()!.InitializeAsync();
            Task<Agent> creating = root.CreatePeerAsync(source, "browser-peer", skill: "browser", waitForBrowser: true);
            var pending = root.FindAgent("browser-peer")!;
            Check(pending is not null && !creating.IsCompleted && !pending.ToolEnabled("page_echo"),
                "Peer startup was invisible, failed before connection, or granted tools early");
            var pair = await Pair();
            await using var peer = pair.Peer;
            var browser = new McpBrowser(pending!.EnsureMcp(), pair.Client);
            serving = BrowserPeer(peer);
            var ready = await creating.WaitAsync(TimeSpan.FromSeconds(8));
            Check(ReferenceEquals(ready, pending) && ready.ToolEnabled("page_echo")
                && browser.Client()?.SkillOwner() == ready.SkillManager() && source.Mcp()?.Browser() is null,
                "Peer did not initialize through its own browser and skill owner");
            Check(File.Exists(Path.Combine(ready.DataDirectory(), "save.json")), "Startup did not reach durable commit");
            string registry = File.ReadAllText(Path.Combine(root.DataDirectory(), "ensemble.json"));
            await Fails(root.CreatePeerAsync(source, "headless", skill: "browser"));
            Check(root.FindAgent("headless") is null, "Headless default waited or retained a failed peer");
            using var cancel = new CancellationTokenSource();
            Task cancelled = root.CreatePeerAsync(source, "cancelled", skill: "browser", cancellation: cancel.Token, waitForBrowser: true);
            Check(root.FindAgent("cancelled") is not null && !cancelled.IsCompleted, "No pending cancellation control");
            cancel.Cancel();
            await Fails(cancelled);
            Check(root.FindAgent("cancelled") is null && File.ReadAllText(Path.Combine(root.DataDirectory(), "ensemble.json")) == registry,
                "Cancelled startup retained ownership or changed registry");
            Task closing = root.CreatePeerAsync(source, "closing", skill: "browser", waitForBrowser: true);
            Check(root.FindAgent("closing") is not null && !closing.IsCompleted, "No pending shutdown control");
            await Task.Run(root.Destroy).WaitAsync(TimeSpan.FromSeconds(8));
            await Fails(closing);
            await serving.WaitAsync(TimeSpan.FromSeconds(8));
            serving = null;
            Check(root.Agents().IsEmpty, "Ensemble shutdown retained a waiting peer");
        }
        finally
        {
            root.Destroy();
            if (serving is not null) await serving.WaitAsync(TimeSpan.FromSeconds(8));
            Directory.Delete(directory, recursive: true);
        }
    }
}
