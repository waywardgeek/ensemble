using System.Net;
using System.Net.Sockets;
using System.Net.WebSockets;
using System.Text.Json;
using Ensemble;
using EnsembleGui;

static class DiagnosticRoutingTests
{
    public static async Task RunAsync()
    {
        string work = Directory.CreateTempSubdirectory("ensemble-gui-diagnostics-").FullName;
        var app = new Ensemble.Ensemble(TextWriter.Null, dataDirectory: work, agentDiagnostics: true);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(40));
        using var stop = new CancellationTokenSource();
        using var socket = new ClientWebSocket();
        Task? serving = null;
        try
        {
            app.StartDiagnostics();
            var config = new Configuration(Vendor.OpenAI, "gpt-4.1", "http://127.0.0.1:1", sandboxRoot: work);
            var first = new Agent(app, config, name: "first");
            var second = new Agent(app, config, name: "second");
            first.StartDiagnostics();
            second.StartDiagnostics();
            using var probe = new TcpListener(IPAddress.Loopback, 0);
            probe.Start();
            int port = ((IPEndPoint)probe.LocalEndpoint).Port;
            probe.Stop();
            string url = $"http://127.0.0.1:{port}";
            serving = Gui.RunAsync(first, new GuiOptions(port, SettingsPath: Path.Combine(work, "settings.json")), stop.Token);
            using var http = new HttpClient();
            while (true)
            {
                if (serving.IsCompleted) await serving;
                try
                {
                    using var response = await http.GetAsync(url, deadline.Token);
                    if (response.IsSuccessStatusCode) break;
                }
                catch (HttpRequestException) { }
                await Task.Delay(20, deadline.Token);
            }
            socket.Options.SetRequestHeader("Origin", url);
            await socket.ConnectAsync(new Uri($"ws://127.0.0.1:{port}/ws"), deadline.Token);
            await Replay();
            await Send(new { type = "subscribe", agent_id = first.Address(), marker = "first-only-marker" });
            await Send(new { type = "select_agent", agent_id = second.Address() });
            await Replay();
            await Send(new { type = "subscribe", agent_id = second.Address(), marker = "second-only-marker" });
            await Send(new { type = "mcp", agent_id = first.Address(),
                data = new { jsonrpc = "2.0", method = "notifications/fixture", marker = "relay-first-marker" } });
            // Client addressing must not redirect diagnostics to another Agent.
            await Send(new { type = "subscribe", agent_id = first.Address(), marker = "refused-second-marker" });
            await Send(new { type = "select_agent", agent_id = first.Address() });
            await Replay();
            string firstLog = File.ReadAllText(Path.Combine(first.DataDirectory(), "gui.log"));
            string secondLog = File.ReadAllText(Path.Combine(second.DataDirectory(), "gui.log"));
            Check(firstLog.Contains("first-only-marker") && !firstLog.Contains("second-only-marker")
                && !firstLog.Contains("refused-second-marker"), "First Agent received another conversation's control");
            Check(secondLog.Contains("second-only-marker") && secondLog.Contains("refused-second-marker")
                && !secondLog.Contains("first-only-marker"), "Selected Agent control routing failed");
            Check(firstLog.Contains("relay-first-marker") && !secondLog.Contains("relay-first-marker"),
                "MCP relay diagnostics followed display selection");
            Check(firstLog.Contains("<queued") && secondLog.Contains("<queued"), "Owned outgoing frames were not traced");
            string hostLog = File.ReadAllText(Path.Combine(work, "ensemble.log"));
            Check(!hostLog.Contains("first-only-marker") && !hostLog.Contains("second-only-marker"), "Host received Agent controls");
            Console.WriteLine("PASS Agent-owned GUI selected/refused/queued diagnostic routing");

            Task Send(object value) => socket.SendAsync(JsonSerializer.SerializeToUtf8Bytes(value),
                WebSocketMessageType.Text, true, deadline.Token);
            async Task Replay()
            {
                byte[] buffer = new byte[8192];
                while (true)
                {
                    using var bytes = new MemoryStream();
                    WebSocketReceiveResult received;
                    do
                    {
                        received = await socket.ReceiveAsync(buffer, deadline.Token);
                        Check(received.MessageType == WebSocketMessageType.Text, "Socket closed before replay");
                        bytes.Write(buffer, 0, received.Count);
                    } while (!received.EndOfMessage);
                    using var json = JsonDocument.Parse(bytes.ToArray());
                    if (json.RootElement.GetProperty("type").GetString() == "replay") return;
                }
            }
        }
        finally
        {
            socket.Abort();
            stop.Cancel();
            try { if (serving is not null) await serving; }
            catch (OperationCanceledException) when (stop.IsCancellationRequested) { }
            finally
            {
                try { app.Destroy(); }
                finally { Directory.Delete(work, true); }
            }
        }
    }

    static void Check(bool condition, string message)
    {
        if (!condition) throw new InvalidOperationException(message);
    }
}
