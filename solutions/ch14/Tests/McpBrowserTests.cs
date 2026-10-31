using System.Net.WebSockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

static partial class McpTests
{
    const string BrowserServer = "mcp_servers:\n  - name: page\n    transport: websocket\n";

    // A working browser-side protocol boundary. The real skill manager,
    // client reader, actor commit, tools and Jobs remain under test.
    static async Task BrowserPeer(IMcpTransport peer, Task? release = null,
        TaskCompletionSource? discovering = null)
    {
        try
        {
            var initialize = await Read(peer);
            Check(initialize.GetProperty("method").GetString() == "initialize", "Browser handshake order");
            await Reply(peer, initialize, new { protocolVersion = "2024-11-05", capabilities = new { } });
            Check((await Read(peer)).GetProperty("method").GetString() == "notifications/initialized",
                "Browser initialized notification missing");
            var list = await Read(peer);
            Check(list.GetProperty("method").GetString() == "tools/list", "Browser discovery order");
            discovering?.TrySetResult();
            if (release is not null)
            {
                await release;
            }
            await Reply(peer, list, new { tools = new object[]
            {
                new { name = "page_echo", description = "Browser text", inputSchema = J("{\"type\":\"object\"}") },
                new { name = "unrequested", description = "No grant", inputSchema = J("{\"type\":\"object\"}") }
            } });
            while (await peer.ReceiveAsync() is { } request)
            {
                if (request.GetProperty("method").GetString() == "hold")
                {
                    continue;
                }
                await Reply(peer, request, new { content = new[] { new { type = "text", text = "browser result" } } });
            }
        }
        catch (Exception error) when (error is IOException or WebSocketException or ObjectDisposedException)
        {
            // Disconnect is a real boundary outcome, not an internally mocked
            // completion. Tests assert the client's reliable outcome separately.
        }
    }

    public static async Task BrowserSkillBinding()
    {
        string directory = SkillFixture();
        Manual(directory, "primary", "type: primary\nloadable-skills: web\n", "Primary");
        Manual(directory, "web", "type: loadable\ntools: page_echo\n" + BrowserServer, "BROWSER_MANUAL");
        var root = new Application(TextWriter.Null);
        var pair = await WebSocketPair();
        await using var peer = new McpWebSocketTransport(pair.Peer);
        Task? serving = null;
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"),
                skillSpec: new SkillSpec(directory, "primary"));
            var unavailable = await agent.DispatchAsync(SkillCall(agent, "load_skill", "web"));
            Check(unavailable.IsError && !agent.ToolEnabled("page_echo"), "Absent browser granted a skill");
            var manager = agent.Mcp()!;
            var browser = new McpBrowser(manager, pair.Client);
            Check(manager.Browser() == browser && browser.Parent() == manager && browser.Client() is null
                && manager.Clients().IsEmpty && !agent.ToolEnabled("page_echo"),
                "Registration granted tools or omitted its parent link");
            var refused = await agent.DispatchAsync(new ToolCallPart("early", agent.Configuration().Target(), "page_echo", J("{}")));
            Check(refused.IsError, "Unloaded browser tool became callable");

            // A second registration and a foreign removal must leave both
            // original links intact. Failed construction does not own its I/O.
            var spare = await WebSocketPair();
            await using var spareClient = spare.Client;
            await using var sparePeer = new McpWebSocketTransport(spare.Peer);
            try
            {
                _ = new McpBrowser(manager, spareClient);
                throw new Exception("Accepted a second browser");
            }
            catch (InvalidOperationException) { }
            await spareClient.SendAsync(J("{\"still\":\"open\"}"));
            Check((await Read(sparePeer)).GetProperty("still").GetString() == "open", "Refused registration closed caller I/O");
            var foreignAgent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"));
            var foreign = new Mcp(foreignAgent);
            try
            {
                foreign.RemoveBrowser(browser);
                throw new Exception("Accepted foreign removal");
            }
            catch (InvalidOperationException) { }
            Check(manager.Browser() == browser && browser.Parent() == manager && foreign.Browser() is null,
                "Refused removal mutated a browser parent");

            serving = BrowserPeer(peer);
            var loaded = await agent.DispatchAsync(SkillCall(agent, "load_skill", "web")).WaitAsync(TimeSpan.FromSeconds(10));
            Check(!loaded.IsError, "Connected browser skill did not load");
            var client = browser.Client()!;
            Check(client.BrowserOwner() == browser && client.Parent() == manager
                && client.SkillOwner() == agent.SkillManager() && agent.SkillManager()!.Clients().Contains(client),
                "Browser client omitted an owning relationship");
            Check(agent.ToolEnabled("page_echo") && !agent.ToolEnabled("unrequested")
                && agent.Tools().All(tool => tool.Name() != "unrequested"), "Discovery bypassed the allowlist");
            var result = await agent.DispatchAsync(new ToolCallPart("echo", agent.Configuration().Target(), "page_echo", J("{}")));
            Check(!result.IsError && result.Parts.OfType<TextPart>().Single().Text == "browser result",
                "Browser call did not return through the ordinary Job path");
            Task held = client.RequestAsync("hold", J("{}"));
            var unloaded = await agent.DispatchAsync(SkillCall(agent, "unload_skill", "web"));
            await Fails(held);
            Check(!unloaded.IsError && browser.Client() is null && client.BrowserOwner() is null
                && manager.Clients().IsEmpty && agent.SkillManager()!.Clients().IsEmpty
                && agent.Tools().All(tool => tool.Name() != "page_echo"), "Unload retained browser tools or ownership");
            await Task.WhenAll(browser.DestroyAsync(), browser.DestroyAsync()).WaitAsync(TimeSpan.FromSeconds(10));
            Check(manager.Browser() is null, "Disconnect retained registration");
            await serving.WaitAsync(TimeSpan.FromSeconds(10));
        }
        finally
        {
            root.Destroy();
            if (serving is not null) await serving.WaitAsync(TimeSpan.FromSeconds(10));
            Directory.Delete(directory, true);
        }
    }

    public static async Task BrowserPrimaryStartup()
    {
        string directory = SkillFixture();
        Manual(directory, "primary", "type: primary\ntools: page_echo\n" + BrowserServer, "BROWSER_PRIMARY");
        var root = new Application(TextWriter.Null);
        var pair = await WebSocketPair();
        await using var peer = new McpWebSocketTransport(pair.Peer);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var discovering = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        Task serving = BrowserPeer(peer, release.Task, discovering);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"),
                skillSpec: new SkillSpec(directory, "primary"));
            var manager = new Mcp(agent);
            var browser = new McpBrowser(manager, pair.Client);
            Check(agent.SkillManager()!.RequiresBrowserStartup(), "Browser primary omitted its host prerequisite");
            Task startup = agent.SkillManager()!.InitializeAsync();
            await discovering.Task.WaitAsync(TimeSpan.FromSeconds(10));
            Check(!agent.SkillManager()!.Ready() && !agent.ToolEnabled("page_echo")
                && !agent.History().View().Dialogue.SelectMany(entry => entry.Parts).OfType<SkillPart>().Any(),
                "Held browser startup published provisional grants or instructions");
            release.SetResult();
            await startup.WaitAsync(TimeSpan.FromSeconds(10));
            Check(agent.SkillManager()!.Ready() && agent.ToolEnabled("page_echo"), "Browser primary did not activate");
            await agent.ShutdownAsync().WaitAsync(TimeSpan.FromSeconds(10));
            Check(manager.Browser() is null && manager.Clients().IsEmpty && browser.Client() is null && agent.Mcp() is null,
                "Shutdown retained browser ownership");
            await serving.WaitAsync(TimeSpan.FromSeconds(10));
        }
        finally
        {
            release.TrySetResult();
            root.Destroy();
            await serving.WaitAsync(TimeSpan.FromSeconds(10));
            Directory.Delete(directory, true);
        }
    }

    public static async Task BrowserDiscoveryDisconnect()
    {
        string directory = SkillFixture();
        Manual(directory, "primary", "type: primary\nloadable-skills: web\n", "Primary");
        Manual(directory, "web", "type: loadable\ntools: page_echo\n" + BrowserServer, "Must not publish");
        var root = new Application(TextWriter.Null);
        var pair = await WebSocketPair();
        await using var peer = new McpWebSocketTransport(pair.Peer);
        var release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var discovering = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        Task serving = BrowserPeer(peer, release.Task, discovering);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1:1"),
                skillSpec: new SkillSpec(directory, "primary"));
            var manager = new Mcp(agent);
            var browser = new McpBrowser(manager, pair.Client);
            var loading = agent.DispatchAsync(SkillCall(agent, "load_skill", "web"));
            await discovering.Task.WaitAsync(TimeSpan.FromSeconds(10));
            await browser.DestroyAsync().WaitAsync(TimeSpan.FromSeconds(10));
            var result = await loading.WaitAsync(TimeSpan.FromSeconds(10));
            Check(result.IsError && manager.Browser() is null && manager.Clients().IsEmpty
                && agent.SkillManager()!.Clients().IsEmpty && !agent.ToolEnabled("page_echo")
                && agent.History().View().Dialogue.SelectMany(entry => entry.Parts).OfType<SkillPart>().All(part => part.Name != "web"),
                "Discovery disconnect leaked provisional activation");
            // An unused replacement registration is still owned and must close
            // when the Agent shuts down; no skill has to claim it first.
            var replacement = await WebSocketPair();
            await using var otherPeer = new McpWebSocketTransport(replacement.Peer);
            _ = new McpBrowser(manager, replacement.Client);
            await agent.ShutdownAsync().WaitAsync(TimeSpan.FromSeconds(10));
            Check(manager.Browser() is null, "Shutdown retained unused browser registration");
            await Fails(Read(otherPeer));
        }
        finally
        {
            release.TrySetResult();
            root.Destroy();
            await serving.WaitAsync(TimeSpan.FromSeconds(10));
            Directory.Delete(directory, true);
        }
    }
}
