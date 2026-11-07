using System.Net;
using System.Net.Sockets;
using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using Ensemble;
using EnsembleGui;

// Real mounted HTTP/WebSocket consumer. Only the browser's MCP protocol is
// replaced; ownership, skill startup, persistence and joined teardown are real.
static class PeerControlsTests
{
    static void Check(bool value, string message)
    {
        if (!value) throw new InvalidOperationException(message);
    }

    static async Task<JsonElement> Frame(ClientWebSocket socket, string type, CancellationToken token)
    {
        byte[] buffer = new byte[8192];
        while (true)
        {
            using var bytes = new MemoryStream();
            WebSocketReceiveResult received;
            do
            {
                received = await socket.ReceiveAsync(buffer, token);
                Check(received.MessageType == WebSocketMessageType.Text, "Unexpected socket close");
                bytes.Write(buffer, 0, received.Count);
            } while (!received.EndOfMessage);
            using var json = JsonDocument.Parse(bytes.ToArray());
            if (json.RootElement.GetProperty("type").GetString() == type) return json.RootElement.Clone();
        }
    }

    static Task Send(ClientWebSocket socket, object value, CancellationToken token) =>
        socket.SendAsync(JsonSerializer.SerializeToUtf8Bytes(value), WebSocketMessageType.Text, true, token);

    public static async Task RunAsync()
    {
        string work = Path.Combine(Path.GetTempPath(), "gui-peers-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(work);
        Console.WriteLine("Peer controls workspace: " + work);
        string skills = Path.Combine(work, "skills");
        foreach (string skill in new[] { "plain", "browser", "extra" }) Directory.CreateDirectory(Path.Combine(skills, skill));
        File.WriteAllText(Path.Combine(skills, "plain", "SKILL.md"), "---\nname: plain\ndescription: Plain\ntype: primary\n---\nPlain\n");
        File.WriteAllText(Path.Combine(skills, "browser", "SKILL.md"), "---\nname: browser\ndescription: Browser\ntype: primary\ntools: page_echo\nmcp_servers:\n  - name: page\n    transport: websocket\n---\nBrowser\n");
        File.WriteAllText(Path.Combine(skills, "extra", "SKILL.md"), "---\nname: extra\ndescription: Loadable only\ntype: loadable\n---\nExtra\n");
        var root = new Ensemble.Ensemble(TextWriter.Null, dataDirectory: Path.Combine(work, "fleet"));
        string globalPath = Path.Combine(work, "global.json");
        var globalSettings = new GlobalSettings(root, globalPath);
        var localAuth = new ModelAuthentication(root, "local");
        var localVendor = new ModelVendor(root, "ollama", "Ollama");
        var localEndpoint = new ModelEndpoint(localVendor, "test", "http://127.0.0.1:2",
            Surface.ChatCompletions, localAuth);
        var localModel = new Model(localEndpoint, "mimo", "fixture-mimo", "MiMo fixture",
            new ModelFeatures(Media.None), servingContextTokens: 4096);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(50));
        var token = deadline.Token;
        var source = new Agent(root, new Configuration(Vendor.OpenAI, "offline", "http://127.0.0.1:1"),
            name: "source", skillSpec: new SkillSpec(skills, "plain"));
        await source.SkillManager()!.InitializeAsync(token);
        using var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        int port = ((IPEndPoint)listener.LocalEndpoint).Port;
        listener.Stop();
        string url = $"http://127.0.0.1:{port}";
        var builder = WebApplication.CreateBuilder();
        builder.Logging.ClearProviders();
        builder.WebHost.UseUrls(url);
        await using var host = builder.Build();
        host.MapGet("/health", () => "caller alive");
        using var stop = new CancellationTokenSource();
        var serving = Gui.ServeAsync(host, source, new GuiOptions(port,
            SettingsPath: Path.Combine(work, "settings.json"), TracePath: Path.Combine(work, "gui.log"), BasePath: "/assistant"), stop.Token);
        using var http = new HttpClient();
        async Task<HttpResponseMessage> Post(string operation, string body, string? header = "1", string? origin = null,
            string? hostHeader = null, string contentType = "application/json", CancellationToken cancellation = default)
        {
            using var request = new HttpRequestMessage(HttpMethod.Post, url + "/assistant/peers/" + operation);
            request.Content = new StringContent(body, Encoding.UTF8, contentType);
            if (header is not null) request.Headers.Add("X-Ensemble-Peers", header);
            if (origin is not null) request.Headers.Add("Origin", origin);
            if (hostHeader is not null) request.Headers.Host = hostHeader;
            return await http.SendAsync(request, cancellation == default ? token : cancellation);
        }
        async Task Expect(string operation, string body, HttpStatusCode status, string? header = "1", string? origin = null,
            string? hostHeader = null, string contentType = "application/json")
        {
            using var response = await Post(operation, body, header, origin, hostHeader, contentType);
            Check(response.StatusCode == status, $"{operation} returned {response.StatusCode}, expected {status}");
            if (hostHeader is null) Check(response.Headers.CacheControl?.NoStore == true, "Management response was cacheable");
        }
        string registry = Path.Combine(root.DataDirectory(), "ensemble.json");
        try
        {
            await host.StartAsync(token);
            using (var denied = await http.GetAsync(url + "/assistant/peers/choices", token))
                Check(denied.StatusCode == HttpStatusCode.Forbidden, "Unguarded choices request was accepted");
            using (var request = new HttpRequestMessage(HttpMethod.Get, url + "/assistant/peers/choices"))
            {
                request.Headers.Add("X-Ensemble-Peers", "1");
                request.Headers.Add("Origin", url);
                using var response = await http.SendAsync(request, token);
                Check(response.StatusCode == HttpStatusCode.OK && response.Headers.CacheControl?.NoStore == true,
                    "Choices were unavailable or cacheable");
                using var choices = JsonDocument.Parse(await response.Content.ReadAsStringAsync(token));
                var offered = choices.RootElement.GetProperty("configured_models").EnumerateArray().Single();
                Check(offered.GetProperty("vendor").GetString() == "Ollama"
                    && offered.GetProperty("enabled").GetBoolean()
                    && JsonCodec.ReadModelIdentity(offered.GetProperty("offering")) == localModel.Identity(),
                    "Choices lost the configured vendor or offering identity");
                Check(choices.RootElement.GetProperty("primary_skills").EnumerateArray()
                    .Select(value => value.GetProperty("name").GetString()).SequenceEqual(new[] { "browser", "plain" }),
                    "Primary choices must exclude loadable templates");
            }
            await Expect("create", "{\"name\":\"bad\"}", HttpStatusCode.Forbidden, header: null);
            await Expect("create", "{\"name\":\"bad\"}", HttpStatusCode.Forbidden, origin: "http://foreign.example");
            await Expect("create", "{\"name\":\"bad\"}", HttpStatusCode.BadRequest, hostHeader: "foreign.example:" + port);
            await Expect("create", "{}", HttpStatusCode.UnsupportedMediaType, contentType: "text/plain");
            foreach (string malformed in new[] { "[]", "{}", "{\"name\":\"a\",\"name\":\"b\"}",
                "{\"name\":\"a\",\"path\":\"outside\"}", "{\"name\":7}", "{\"name\":\"../escape\"}",
                "{\"name\":\"a\",\"skill\":null}", "{\"name\":\"" + new string('x', 4200) + "\"}" })
            {
                await Expect("create", malformed, HttpStatusCode.BadRequest);
            }
            Check(root.Agents().Length == 1 && !File.Exists(registry), "Refused requests changed ownership/registry");
            using (var outside = await http.PostAsync(url + "/peers/create", new StringContent("{}"), token))
                Check(outside.StatusCode == HttpStatusCode.NotFound, "Management escaped mounted prefix");
            using (var get = await http.GetAsync(url + "/assistant/peers/create", token))
                Check(get.StatusCode == HttpStatusCode.MethodNotAllowed, "Mutation accepted GET");

            string offeringJson = JsonCodec.WriteModelIdentity(localModel.Identity()).ToJsonString();
            await Expect("create", "{\"name\":\"local-peer\",\"offering\":" + offeringJson
                + ",\"skill\":\"plain\"}", HttpStatusCode.OK, origin: url);
            var local = root.FindAgent("local-peer")!;
            Check(ReferenceEquals(local.SelectedModel(), localModel)
                && local.Configuration().Billing == BillingMode.None && source.Configuration().Model == "offline",
                "Qualified creation did not select the complete local route independently");
            await Expect("delete", "{\"name\":\"local-peer\"}", HttpStatusCode.OK);
            Check(globalSettings.FundingDefaults().IsEmpty, "Creating an Agent changed global funding.");
            async Task<HttpResponseMessage> Funding(string body, string? origin = null, bool header = true)
            {
                using var request = new HttpRequestMessage(HttpMethod.Post, url + "/assistant/models/funding");
                request.Content = new StringContent(body, Encoding.UTF8, "application/json");
                if (header) request.Headers.Add("X-Ensemble-Peers", "1");
                if (origin is not null) request.Headers.Add("Origin", origin);
                return await http.SendAsync(request, token);
            }
            string fundingBody = "{\"offering\":" + offeringJson + "}";
            using (var denied = await Funding(fundingBody, header: false))
                Check(denied.StatusCode == HttpStatusCode.Forbidden, "Funding change lacked guard.");
            using (var denied = await Funding(fundingBody, origin: "https://foreign.example"))
                Check(denied.StatusCode == HttpStatusCode.Forbidden, "Cross-origin funding change accepted.");
            using (var saved = await Funding(fundingBody, origin: url))
                Check(saved.StatusCode == HttpStatusCode.OK, "Funding settings save failed.");
            Check(globalSettings.FundingDefaults()["ollama"] == "local" && source.SelectedModel() is null,
                "Funding save did not persist independently of the active model.");
            byte[] savedFunding = File.ReadAllBytes(globalPath);
            using (var held = new FileStream(globalPath + ".lock", FileMode.OpenOrCreate,
                FileAccess.ReadWrite, FileShare.None))
            using (var refused = await Funding(fundingBody))
            {
                Check(refused.StatusCode == HttpStatusCode.Conflict, "Locked funding write was accepted.");
                Check(File.ReadAllBytes(globalPath).SequenceEqual(savedFunding), "Refused funding save changed defaults.");
            }
            // Existing legacy peers can be repaired independently while this GUI stays live.
            var stoppedPeer = await root.CreatePeerAsync(source, "recover-me", cancellation: token);
            stoppedPeer.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("Recovered conversation")]));
            await stoppedPeer.ShutdownAsync();
            stoppedPeer.Destroy();
            using (var request = new HttpRequestMessage(HttpMethod.Get, url + "/assistant/peers/saved"))
            {
                request.Headers.Add("X-Ensemble-Peers", "1");
                using var response = await http.SendAsync(request, token);
                string listed = await response.Content.ReadAsStringAsync(token);
                Check(response.IsSuccessStatusCode && listed.Contains("recover-me"), "Stopped peer missing from recovery UI.");
            }
            await Expect("restore", "{\"name\":\"recover-me\",\"offering\":" + offeringJson + "}", HttpStatusCode.OK);
            Check(ReferenceEquals(root.FindAgent("recover-me")!.SelectedModel(), localModel), "HTTP recovery lost explicit route.");
            Check(JsonCodec.Text(root.FindAgent("recover-me")!.History().View().Dialogue.SelectMany(entry => entry.Parts))
                .Contains("Recovered conversation"), "HTTP recovery lost conversation.");
            await Expect("restore", "{\"name\":\"recover-me\"}", HttpStatusCode.Conflict);
            await Expect("delete", "{\"name\":\"recover-me\"}", HttpStatusCode.OK);
            const string preferenceJson = "{\"vendor\":\"ollama\",\"model\":\"fixture-mimo\"}";
            await Expect("create", "{\"name\":\"preferred-peer\",\"preference\":" + preferenceJson
                + ",\"skill\":\"plain\"}", HttpStatusCode.OK);
            Check(ReferenceEquals(root.FindAgent("preferred-peer")!.SelectedModel(), localModel)
                && source.SelectedModel() is null, "Remembered funding did not create an independent configured peer");
            await Expect("delete", "{\"name\":\"preferred-peer\"}", HttpStatusCode.OK);
            File.WriteAllText(globalPath, "{\"model_funding_defaults\":{\"ollama\":\"missing\"}}");
            await Expect("create", "{\"name\":\"missing-funding\",\"preference\":" + preferenceJson + "}",
                HttpStatusCode.Conflict);
            Check(root.FindAgent("missing-funding") is null, "Unavailable funding created a peer through fallback");
            await Expect("create", "{\"name\":\"mixed-choice\",\"preference\":" + preferenceJson
                + ",\"offering\":" + offeringJson + "}", HttpStatusCode.BadRequest);
            globalSettings.RememberFunding(localModel);
            await Expect("create", "{\"name\":\"conflict\",\"model\":\"offline\",\"offering\":"
                + offeringJson + "}", HttpStatusCode.BadRequest);
            await Expect("create", "{\"name\":\"bad-offering\",\"offering\":null}", HttpStatusCode.BadRequest);
            await Expect("create", "{\"name\":\"plain-peer\",\"model\":\"chosen-model\"}", HttpStatusCode.OK, origin: url);
            var plain = root.FindAgent("plain-peer")!;
            Check(plain.Configuration().Model == "chosen-model" && File.Exists(Path.Combine(plain.DataDirectory(), "save.json")), "Creation acknowledged before durable startup");
            string committed = File.ReadAllText(registry);
            await Expect("create", "{\"name\":\"plain-peer\"}", HttpStatusCode.Conflict);
            await Expect("delete", "{\"name\":\"source\"}", HttpStatusCode.Conflict);
            var child = await plain.CreateChildAsync("worker", cancellation: token);
            await Expect("delete", "{\"name\":\"plain-peer\"}", HttpStatusCode.Conflict);
            Check(File.ReadAllText(registry) == committed && root.FindAgent("plain-peer") == plain,
                "Refused deletion changed durable/live membership");
            child.Destroy();
            string retainedHome = plain.DataDirectory();
            await Expect("delete", "{\"name\":\"plain-peer\"}", HttpStatusCode.OK);
            Check(root.FindAgent("plain-peer") is null && File.Exists(Path.Combine(retainedHome, "save.json")), "Deletion lost files or retained membership");

            using var socket = new ClientWebSocket();
            socket.Options.SetRequestHeader("Origin", url);
            await socket.ConnectAsync(new Uri($"ws://127.0.0.1:{port}/assistant/ws"), token);
            await Frame(socket, "replay", token);
            using var creatingResponse = await StartBrowserPeer("web");
            Check(creatingResponse.StatusCode == HttpStatusCode.OK && root.FindAgent("web")!.ToolEnabled("page_echo"),
                "Browser peer did not finish real skill initialization");

            // A deleted display returns through the same acknowledged replay
            // admission as an explicit switch, without retargeting this relay.
            await Expect("create", "{\"name\":\"victim\"}", HttpStatusCode.OK);
            await Send(socket, new { type = "select_agent", agent_id = "victim" }, token);
            Check((await Frame(socket, "replay", token)).GetProperty("agent_id").GetString() == "victim", "Victim selection failed");
            await Expect("delete", "{\"name\":\"victim\"}", HttpStatusCode.OK);
            Check((await Frame(socket, "replay", token)).GetProperty("agent_id").GetString() == "source", "Deleted display did not recover through replay");

            // Cancel the mount with another HTTP creation waiting for a browser.
            // ServeAsync must join rollback without destroying the borrowed owners.
            Task<HttpResponseMessage> pending = Post("create", "{\"name\":\"pending\",\"skill\":\"browser\"}");
            await TreeContains("pending");
            Check(!pending.IsCompleted && !File.ReadAllText(registry).Contains("pending", StringComparison.Ordinal), "Pending startup falsely committed");
            stop.Cancel();
            await serving.WaitAsync(TimeSpan.FromSeconds(10));
            using var stopped = await pending;
            Check(stopped.StatusCode == HttpStatusCode.ServiceUnavailable && root.FindAgent("pending") is null,
                "Mount teardown did not cancel/join creation");
            Check(await http.GetStringAsync(url + "/health", token) == "caller alive"
                && source.Configuration().Model == "offline" && root.FindAgent("web") is not null,
                "Mount acquired caller/peer lifetime");

            async Task TreeContains(string address)
            {
                while (true)
                {
                    var tree = await Frame(socket, "agent_tree", token);
                    if (tree.GetProperty("agents").EnumerateArray().Any(row => row.GetProperty("key").GetString() == address)) return;
                }
            }
            async Task<HttpResponseMessage> StartBrowserPeer(string name)
            {
                Task<HttpResponseMessage> creating = Post("create", JsonSerializer.Serialize(new { name, skill = "browser" }));
                await TreeContains(name);
                Check(!creating.IsCompleted, "Creation did not await its own browser");
                await Send(socket, new { type = "select_agent", agent_id = name }, token);
                Check((await Frame(socket, "replay", token)).GetProperty("agent_id").GetString() == name,
                    "Pending HTTP creation blocked display selection");
                Check(!creating.IsCompleted && root.FindAgent(name)!.Mcp()!.Browser() is null, "Display selection borrowed source browser");
                using var relay = new ClientWebSocket();
                relay.Options.SetRequestHeader("Origin", url);
                await relay.ConnectAsync(new Uri($"ws://127.0.0.1:{port}/assistant/ws?agent_id={name}"), token);
                var initialize = (await Frame(relay, "mcp", token)).GetProperty("data");
                Check(initialize.GetProperty("method").GetString() == "initialize", "Wrong initialization request");
                await Send(relay, new { type = "mcp", data = new { jsonrpc = "2.0", id = initialize.GetProperty("id"), result = new { protocolVersion = "2024-11-05", capabilities = new { } } } }, token);
                await Frame(relay, "mcp", token); // notifications/initialized
                var listing = (await Frame(relay, "mcp", token)).GetProperty("data");
                Check(listing.GetProperty("method").GetString() == "tools/list", "Wrong discovery request");
                await Send(relay, new { type = "mcp", data = new { jsonrpc = "2.0", id = listing.GetProperty("id"), result = new { tools = new[] { new { name = "page_echo", description = "Echo", inputSchema = new { type = "object" } } } } } }, token);
                var response = await creating;
                relay.Abort();
                return response;
            }
        }
        finally
        {
            stop.Cancel();
            try { await serving.WaitAsync(TimeSpan.FromSeconds(10)); }
            finally
            {
                await host.StopAsync(CancellationToken.None);
                root.Destroy();
            }
        }
        Console.WriteLine("PASS mounted peer controls: guards, persistence, refusal, independent browser startup, joined mount cancellation");
    }
}
