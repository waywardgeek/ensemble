using System.Net;
using System.Net.Sockets;
using System.Net.WebSockets;
using Ensemble;
using EnsembleGui;

// An authenticated proxy terminates TLS, preserves Host, and connects locally.
// This fixture simulates only that transport; no Cloudflare or provider login.
static class PublicOriginTests
{
    static void Check(bool value, string message)
    {
        if (!value) throw new InvalidOperationException(message);
    }

    public static async Task RunAsync()
    {
        string work = Path.Combine(Path.GetTempPath(), "gui-origin-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(work);
        var root = new Ensemble.Ensemble(TextWriter.Null, dataDirectory: Path.Combine(work, "fleet"));
        var agent = new Agent(root, new Configuration(Vendor.OpenAI, "offline", "http://127.0.0.1:1"), name: "source");
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(45));
        var token = deadline.Token;
        using var probe = new TcpListener(IPAddress.Loopback, 0);
        probe.Start();
        int port = ((IPEndPoint)probe.LocalEndpoint).Port;
        probe.Stop();
        string local = $"http://127.0.0.1:{port}";
        const string publicHost = "ensemble.example.test";
        const string publicOrigin = "https://ensemble.example.test";
        var builder = WebApplication.CreateBuilder();
        builder.Logging.ClearProviders();
        builder.WebHost.UseUrls(local);
        await using var host = builder.Build();
        using var stop = new CancellationTokenSource();
        using var http = new HttpClient(new HttpClientHandler { UseProxy = false });

        // Refuse malformed configuration before registering a mount or starting.
        foreach (string invalid in new[] { "http://ensemble.example.test", "https://ensemble.example.test/path",
            "https://user:secret@ensemble.example.test", "https://ensemble.example.test?q=1", "https://localhost",
            "https://ensemble.example.test#fragment", "https://*.example.test" })
        {
            bool refused = false;
            try { await Gui.ServeAsync(host, agent, new GuiOptions(port, PublicOrigin: invalid), token); }
            catch (ArgumentException) { refused = true; }
            Check(refused, "Invalid public-origin configuration was accepted");
        }
        var serving = Gui.ServeAsync(host, agent, new GuiOptions(port,
            SettingsPath: Path.Combine(work, "settings.json"), TracePath: Path.Combine(work, "gui.log"),
            BasePath: "/assistant", PublicOrigin: publicOrigin), stop.Token);
        async Task Expect(string path, HttpStatusCode status, string? origin = null,
            string? guard = null, bool post = false, string? hostName = publicHost, bool forwarded = false)
        {
            using var request = new HttpRequestMessage(post ? HttpMethod.Post : HttpMethod.Get, local + "/assistant" + path);
            if (post) request.Content = new StringContent("{}", System.Text.Encoding.UTF8, "application/json");
            request.Headers.Host = hostName ?? $"127.0.0.1:{port}";
            if (origin is not null) request.Headers.Add("Origin", origin);
            if (guard is not null) request.Headers.Add(guard, "1");
            if (forwarded)
            {
                request.Headers.Add("X-Forwarded-Host", publicHost);
                request.Headers.Add("X-Forwarded-Proto", "https");
            }
            using var response = await http.SendAsync(request, token);
            Check(response.StatusCode == status, $"{path}: expected {status}, got {response.StatusCode}");
        }
        async Task Socket(string? origin, bool allowed, string hostName = publicHost)
        {
            using var socket = new ClientWebSocket();
            socket.Options.Proxy = null;
            socket.Options.SetRequestHeader("Host", hostName);
            if (origin is not null) socket.Options.SetRequestHeader("Origin", origin);
            bool connected = false;
            try { await socket.ConnectAsync(new Uri($"ws://127.0.0.1:{port}/assistant/ws"), token); connected = true; }
            catch (WebSocketException) when (!allowed) { }
            finally { socket.Abort(); }
            Check(connected == allowed, "Unexpected WebSocket origin decision");
        }
        try
        {
            await host.StartAsync(token);
            while (true)
            {
                if (serving.IsCompleted) { await serving; throw new InvalidOperationException("Mount ended before readiness"); }
                try { using var response = await http.GetAsync(local + "/assistant/", token); if (response.IsSuccessStatusCode) break; }
                catch (HttpRequestException) { }
                await Task.Delay(30, token);
            }
            await Expect("/", HttpStatusCode.OK);
            await Expect("/", HttpStatusCode.OK, hostName: null);
            await Expect("/", HttpStatusCode.BadRequest, hostName: "untrusted.example.test", forwarded: true);
            await Socket(publicOrigin, true);
            await Socket(null, false);
            await Socket("http://ensemble.example.test", false);
            await Socket("https://untrusted.example.test", false);
            await Socket(local, true, $"127.0.0.1:{port}");
            await Expect("/auth/status", HttpStatusCode.OK, guard: "X-Ensemble-Auth");
            await Expect("/auth/status", HttpStatusCode.Forbidden);
            await Expect("/auth/status", HttpStatusCode.Forbidden, origin: "https://untrusted.example.test", guard: "X-Ensemble-Auth");
            await Expect("/auth/callback?code=dummy&state=dummy", HttpStatusCode.Forbidden);
            // Refusal inside the operation (400), rather than its CSRF guard
            // (403), demonstrates that an allowed request reaches real routing.
            await Expect("/peers/create", HttpStatusCode.BadRequest, publicOrigin, "X-Ensemble-Peers", post: true);
            await Expect("/peers/create", HttpStatusCode.Forbidden, guard: "X-Ensemble-Peers", post: true);
            await Expect("/peers/create", HttpStatusCode.Forbidden, "https://untrusted.example.test", "X-Ensemble-Peers", post: true);
            await Expect("/jobs/output", HttpStatusCode.BadRequest, guard: "X-Ensemble-Output");
            await Expect("/jobs/output", HttpStatusCode.Forbidden);
            await Expect("/jobs/output", HttpStatusCode.Forbidden, "https://untrusted.example.test", "X-Ensemble-Output");
            Console.WriteLine("PASS public origin: HTTPS proxy, local access, WebSocket, guarded routes, wrong hosts/origins, spoofed forwarding, local-only callback");
        }
        finally
        {
            stop.Cancel();
            try { await serving.WaitAsync(TimeSpan.FromSeconds(10)); }
            finally { await host.StopAsync(CancellationToken.None); root.Destroy(); }
        }
    }
}
