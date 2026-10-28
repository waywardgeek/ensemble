using System.Net;
using System.Net.Sockets;
using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using Ensemble;
using EnsembleGui;

// Outside-assembly regression consumer: it uses only the public core/GUI API.
// Creates its own empty working directory. All HTTP is literal loopback; the
// Agent has no provider credential and the test never submits an inference.

static void Check(bool condition, string message)
{
    if (!condition) throw new InvalidOperationException(message);
}
static int Port()
{
    using var listener = new TcpListener(IPAddress.Loopback, 0);
    listener.Start();
    return ((IPEndPoint)listener.LocalEndpoint).Port;
}
static async Task RpcReply(ClientWebSocket socket, JsonElement request, object result, CancellationToken cancellation)
{
    byte[] bytes = JsonSerializer.SerializeToUtf8Bytes(new
    {
        type = "mcp",
        data = new { jsonrpc = "2.0", id = request.GetProperty("id"), result }
    });
    await socket.SendAsync(bytes, WebSocketMessageType.Text, true, cancellation);
}

static async Task Ready(HttpClient client, string url, Task lifetime, CancellationToken cancellation)
{
    while (true)
    {
        if (lifetime.IsCompleted) await lifetime;
        try
        {
            using var response = await client.GetAsync(url, cancellation);
            if (response.IsSuccessStatusCode) return;
        }
        catch (HttpRequestException) { }
        await Task.Delay(20, cancellation);
    }
}
static async Task<JsonElement> Frame(ClientWebSocket socket, string type, CancellationToken cancellation)
{
    while (true)
    {
        using var bytes = new MemoryStream();
        var buffer = new byte[8192];
        WebSocketReceiveResult received;
        do
        {
            received = await socket.ReceiveAsync(buffer, cancellation);
            Check(received.MessageType == WebSocketMessageType.Text, "Socket closed before expected frame " + type);
            bytes.Write(buffer, 0, received.Count);
        } while (!received.EndOfMessage);
        using var json = JsonDocument.Parse(bytes.ToArray());
        if (json.RootElement.TryGetProperty("type", out var value) && value.GetString() == type)
            return json.RootElement.Clone();
    }
}
// Relative fixture paths must never overwrite a caller's shipped skills, even
// when a developer launches this consumer from the solution directory.
string originalDirectory = Environment.CurrentDirectory;
string workspace = Directory.CreateTempSubdirectory("ensemble-gui-consumer-").FullName;
Directory.SetCurrentDirectory(workspace);
Console.WriteLine("Consumer workspace: " + workspace);
using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(30));
var cancellation = deadline.Token;
using var http = new HttpClient();
var root = new Ensemble.Ensemble(TextWriter.Null);
try
{
    var agent = new Agent(root, new Configuration(Vendor.Anthropic, "offline", apiKey: "", baseUrl: "http://127.0.0.1:1"), outputDirectory: "jobs");
    for (int attempt = 0; attempt < 2; attempt++)
    {
        int port = Port();
        string origin = $"http://127.0.0.1:{port}";
        using var stop = CancellationTokenSource.CreateLinkedTokenSource(cancellation);
        Task hosting = Gui.RunAsync(agent, new GuiOptions(port, SettingsPath: "settings.json", TracePath: "gui.log"), stop.Token);
        try
        {
            await Ready(http, origin + "/", hosting, cancellation);
            string page = await http.GetStringAsync(origin + "/", cancellation);
            Check(page.Contains("/app.js", StringComparison.Ordinal), "Embedded index missing");
            foreach (string asset in new[] { "app.js", "auth.js", "mcp.js", "speech.js", "style.css" })
                Check((await http.GetStringAsync(origin + "/" + asset, cancellation)).Length > 100, "Asset missing: " + asset);
            using var foreign = new HttpRequestMessage(HttpMethod.Get, origin + "/");
            foreign.Headers.Host = "foreign.example";
            using var refused = await http.SendAsync(foreign, cancellation);
            Check(refused.StatusCode == HttpStatusCode.BadRequest, "Foreign Host accepted");
            using var socket = new ClientWebSocket();
            socket.Options.SetRequestHeader("Origin", origin);
            await socket.ConnectAsync(new Uri($"ws://127.0.0.1:{port}/ws"), cancellation);
            await Frame(socket, "current_settings", cancellation);
            Check(agent.Mcp()!.Browser() is not null, "Socket did not register browser relay");
            Check(agent.Mcp()!.Clients().Length == 0, "Discovery started without a granting skill");
            await socket.SendAsync(Encoding.UTF8.GetBytes("{\"type\":\"pause\"}"), WebSocketMessageType.Text, true, cancellation);
            while ((await Frame(socket, "agent_state", cancellation)).GetProperty("state").GetString() != "paused") { }
            socket.Abort();
        }
        finally
        {
            stop.Cancel();
            await hosting.WaitAsync(TimeSpan.FromSeconds(10));
        }
        Check(agent.Mcp()!.Browser() is null, "Browser registration survived GUI stop");
        Check(agent.Mcp()!.Clients().Length == 0, "Browser client survived GUI stop");
        await agent.PostAsync(new AgentInput(InputKind.Resume));
        Check(agent.Configuration().Model == "offline", "Host destroyed borrowed Agent");
    }
    // A caller can mount the public asset operation beside unrelated endpoints.
    var builder = WebApplication.CreateBuilder(Array.Empty<string>());
    builder.Logging.ClearProviders();
    int customPort = Port();
    builder.WebHost.UseUrls($"http://127.0.0.1:{customPort}");
    await using var custom = builder.Build();
    custom.MapGet("/health", () => "consumer-owned");
    Gui.MapAssets(custom.MapGroup("/embedded"));
    Directory.CreateDirectory("override");
    await File.WriteAllTextAsync("override/index.html", "OVERRIDE_PAGE", cancellation);
    Gui.MapAssets(custom.MapGroup("/disk"), Path.GetFullPath("override"));
    string skillDirectory = Path.Combine(Environment.CurrentDirectory, "skills");
    Directory.CreateDirectory(Path.Combine(skillDirectory, "browser"));
    await File.WriteAllTextAsync(Path.Combine(skillDirectory, "browser", "SKILL.md"), """
        ---
        name: browser
        description: External GUI consumer browser peer
        type: primary
        tools: page_echo
        mcp_servers:
          - name: page
            transport: websocket
        ---
        Use page_echo to read the connected test page.
        """, cancellation);
    var mountedAgent = new Agent(root, new Configuration(Vendor.OpenAI, "offline", baseUrl: "http://127.0.0.1:1"),
        name: "mounted", skillSpec: new SkillSpec(skillDirectory, "browser"));
    mountedAgent.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
        parts: [new TextPart("Canonical replay from the borrowed Agent")]));
    using var mountedStop = CancellationTokenSource.CreateLinkedTokenSource(cancellation);
    Task mounted = Gui.ServeAsync(custom, mountedAgent, new GuiOptions(customPort,
        SettingsPath: "mounted-settings.json", TracePath: "mounted.log", BasePath: "/assistant"), mountedStop.Token);
    if (mounted.IsCompleted)
    {
        await mounted;
    }
    await custom.StartAsync(cancellation);
    try
    {
        string url = $"http://127.0.0.1:{customPort}";
        Check(await http.GetStringAsync(url + "/health", cancellation) == "consumer-owned", "Custom handler missing");
        Check((await http.GetStringAsync(url + "/embedded/", cancellation)).Contains("/app.js"), "Mounted assets missing");
        Check(await http.GetStringAsync(url + "/disk/", cancellation) == "OVERRIDE_PAGE", "Override ignored");
        using var missing = await http.GetAsync(url + "/disk/app.js", cancellation);
        Check(missing.StatusCode == HttpStatusCode.NotFound, "Override silently fell back");

        // The complete mount uses its prefix, not the application root. Follow
        // the directory redirect just as a browser would before loading assets.
        using var mountedPage = await http.GetAsync(url + "/assistant", cancellation);
        Check(mountedPage.RequestMessage!.RequestUri!.AbsolutePath == "/assistant/", "Mount omitted directory redirect");
        Check((await mountedPage.Content.ReadAsStringAsync(cancellation)).Contains("./app.js"), "Page URLs escaped mount");
        Check((await http.GetStringAsync(url + "/assistant/app.js", cancellation)).Contains("""new URL("./ws", document.baseURI)"""), "Socket URL escaped mount");
        using var account = new HttpRequestMessage(HttpMethod.Get, url + "/assistant/auth/status");
        account.Headers.Add("X-Ensemble-Auth", "1");
        using var accountResponse = await http.SendAsync(account, cancellation);
        Check(accountResponse.IsSuccessStatusCode, "Mounted account status unavailable");
        using var accountJson = JsonDocument.Parse(await accountResponse.Content.ReadAsStringAsync(cancellation));
        Check(!accountJson.RootElement.GetProperty("enabled").GetBoolean(), "Unexpected credential store");
        using var unguarded = await http.GetAsync(url + "/assistant/auth/status", cancellation);
        Check(unguarded.StatusCode == HttpStatusCode.Forbidden, "Mounted account route lost guard");
        using var outside = await http.GetAsync(url + "/auth/status", cancellation);
        Check(outside.StatusCode == HttpStatusCode.NotFound, "Mounted account route leaked to root");
        using var foreignHealth = new HttpRequestMessage(HttpMethod.Get, url + "/health");
        foreignHealth.Headers.Host = "caller.example";
        using var healthResponse = await http.SendAsync(foreignHealth, cancellation);
        Check(healthResponse.IsSuccessStatusCode, "GUI guard changed caller route authority");
        using var foreignSocket = new ClientWebSocket();
        foreignSocket.Options.SetRequestHeader("Origin", "http://foreign.example");
        bool refusedOrigin = false;
        try
        {
            await foreignSocket.ConnectAsync(new Uri($"ws://127.0.0.1:{customPort}/assistant/ws"), cancellation);
        }
        catch (WebSocketException)
        {
            refusedOrigin = true;
        }
        Check(refusedOrigin, "Mounted socket accepted foreign origin");
        using var mountedSocket = new ClientWebSocket();
        mountedSocket.Options.SetRequestHeader("Origin", url);
        await mountedSocket.ConnectAsync(new Uri($"ws://127.0.0.1:{customPort}/assistant/ws"), cancellation);
        var replay = await Frame(mountedSocket, "replay", cancellation);
        var replayText = replay.GetProperty("data").GetProperty("events").EnumerateArray()
            .Single(frame => frame.GetProperty("type").GetString() == "part_final");
        Check(replayText.GetProperty("part").GetProperty("text").GetString() == "Canonical replay from the borrowed Agent",
            "Mounted handler did not replay canonical History");
        await Frame(mountedSocket, "current_settings", cancellation);
        Check(mountedAgent.Mcp()!.Browser() is not null, "Mounted socket did not register browser");
        var initialize = (await Frame(mountedSocket, "mcp", cancellation)).GetProperty("data");
        Check(initialize.GetProperty("method").GetString() == "initialize", "Browser initialization missing");
        await RpcReply(mountedSocket, initialize, new { protocolVersion = "2024-11-05", capabilities = new { } }, cancellation);
        var initialized = (await Frame(mountedSocket, "mcp", cancellation)).GetProperty("data");
        Check(initialized.GetProperty("method").GetString() == "notifications/initialized", "Browser initialization order");
        var listing = (await Frame(mountedSocket, "mcp", cancellation)).GetProperty("data");
        Check(listing.GetProperty("method").GetString() == "tools/list", "Browser discovery missing");
        await RpcReply(mountedSocket, listing, new { tools = new[]
        {
            new { name = "page_echo", description = "Read test page", inputSchema = new { type = "object" } }
        } }, cancellation);
        await mountedAgent.SkillManager()!.InitializeAsync(cancellation);
        Check(mountedAgent.ToolEnabled("page_echo"), "Browser skill did not grant tool");
        var call = new ToolCallPart("mounted-echo", mountedAgent.Configuration().Target(), "page_echo",
            JsonSerializer.SerializeToElement(new { }));
        Task<ToolResultPart> dispatch = mountedAgent.DispatchAsync(call, cancellation);
        var request = (await Frame(mountedSocket, "mcp", cancellation)).GetProperty("data");
        Check(request.GetProperty("method").GetString() == "tools/call"
            && request.GetProperty("params").GetProperty("name").GetString() == "page_echo", "Wrong browser call");
        await RpcReply(mountedSocket, request, new { content = new[] { new { type = "text", text = "Mounted browser result" } } }, cancellation);
        var result = await dispatch;
        Check(!result.IsError && result.Parts.OfType<TextPart>().Single().Text == "Mounted browser result",
            "Browser reply did not reach ordinary Agent dispatch");
        await mountedSocket.SendAsync(Encoding.UTF8.GetBytes("""{"type":"pause"}"""), WebSocketMessageType.Text, true, cancellation);
        while ((await Frame(mountedSocket, "agent_state", cancellation)).GetProperty("state").GetString() != "paused") { }

        // Cancel with an open socket. Completion must mean its request and
        // browser relay have detached, not merely that shutdown was requested.
        Task held = mountedAgent.Mcp()!.Browser()!.Client()!.RequestAsync("hold", JsonSerializer.SerializeToElement(new { }));
        var heldRequest = (await Frame(mountedSocket, "mcp", cancellation)).GetProperty("data");
        Check(heldRequest.GetProperty("method").GetString() == "hold", "Pending RPC never reached browser");
        var credentials = new Credentials(root, Path.GetFullPath("consumer-credentials"));
        using var login = new HttpRequestMessage(HttpMethod.Post, url + "/assistant/auth/login");
        login.Headers.Add("X-Ensemble-Auth", "1");
        login.Headers.Add("Origin", url);
        login.Content = new StringContent("""{"label":"local-only"}""", Encoding.UTF8, "application/json");
        login.Content.Headers.ContentType = new System.Net.Http.Headers.MediaTypeHeaderValue("application/json");
        using var loginResponse = await http.SendAsync(login, cancellation);
        Check(loginResponse.IsSuccessStatusCode, "Mounted login used an invalid callback URI");
        using var loginJson = JsonDocument.Parse(await loginResponse.Content.ReadAsStringAsync(cancellation));
        var authorization = new Uri(loginJson.RootElement.GetProperty("url").GetString()!);
        var query = Microsoft.AspNetCore.WebUtilities.QueryHelpers.ParseQuery(authorization.Query);
        Check(query["redirect_uri"] == url + "/auth/callback", "Mounted login changed fixed callback contract");
        using var wrongState = await http.GetAsync(url + "/auth/callback?state=wrong&error=access_denied", cancellation);
        Check(wrongState.StatusCode == HttpStatusCode.BadRequest && await credentials.LoginPendingAsync(cancellation),
            "Root callback missing or unsolicited callback consumed the attempt");
        mountedStop.Cancel();
        await mounted.WaitAsync(TimeSpan.FromSeconds(10));
        bool heldFailed = false;
        try
        {
            await held.WaitAsync(TimeSpan.FromSeconds(2));
        }
        catch (Exception error) when (error is IOException or ObjectDisposedException or OperationCanceledException)
        {
            heldFailed = true;
        }
        Check(heldFailed && mountedAgent.Mcp()!.Clients().IsEmpty
            && mountedAgent.SkillManager()!.Clients().IsEmpty, "Mounted teardown retained pending RPC or client owners");
        Check(mountedAgent.Mcp()!.Browser() is null, "Mounted browser survived serving task");
        Check(await credentials.LoginPendingAsync(cancellation), "Borrowed hosting cancelled caller's credential attempt");
        await credentials.CancelLoginAsync(cancellation);
        Check(await http.GetStringAsync(url + "/health", cancellation) == "consumer-owned", "GUI stopped caller application");
        using var closed = await http.GetAsync(url + "/assistant/", cancellation);
        Check(closed.StatusCode == HttpStatusCode.ServiceUnavailable, "Stopped mount admitted a request");
        await mountedAgent.PostAsync(new AgentInput(InputKind.Resume));
        Check(mountedAgent.Configuration().Model == "offline", "Mounted GUI destroyed borrowed Agent");
    }
    finally
    {
        mountedStop.Cancel();
        try
        {
            await mounted.WaitAsync(TimeSpan.FromSeconds(10));
        }
        finally
        {
            await custom.StopAsync(CancellationToken.None);
        }
    }
}
finally
{
    try
    {
        root.Destroy();
    }
    finally
    {
        Directory.SetCurrentDirectory(originalDirectory);
    }
}
Console.WriteLine("PASS external GUI consumer: embedded assets, disk override, full caller-owned mount, WebSocket controls, two hosting lifetimes, Agent survival");
