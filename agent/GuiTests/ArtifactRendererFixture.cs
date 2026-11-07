using System.Net;
using System.Net.Sockets;
using Ensemble;
using EnsembleGui;
using System.Text.Json;
using System.Threading.Channels;

// No provider calls: browser tests inject recorded facts into the real handler.
internal static class ArtifactRendererFixture
{
    // Only the remote transport is fake; handshake and Tool registration are real.
    private sealed class RendererPeer(string[] names) : IMcpTransport
    {
        private readonly Channel<JsonElement> replies = Channel.CreateUnbounded<JsonElement>();
        public Task SendAsync(JsonElement message, CancellationToken cancellation = default)
        {
            if (!message.TryGetProperty("id", out var id)) return Task.CompletedTask;
            string method = message.GetProperty("method").GetString()!;
            object result = method switch
            {
                "initialize" => new { protocolVersion = "2024-11-05", capabilities = new { tools = new { } } },
                "tools/list" => new { tools = names.Select(name => new
                    { name, description = "Synthetic renderer fixture", inputSchema = new { type = "object" } }) },
                _ => throw new InvalidOperationException("Unexpected renderer fixture RPC")
            };
            if (!replies.Writer.TryWrite(JsonSerializer.SerializeToElement(new { jsonrpc = "2.0", id, result })))
                throw new IOException("Fixture closed");
            return Task.CompletedTask;
        }
        public async Task<JsonElement?> ReceiveAsync(CancellationToken cancellation = default)
        {
            try { return await replies.Reader.ReadAsync(cancellation); }
            catch (ChannelClosedException) { return null; }
        }
        public ValueTask DisposeAsync() { replies.Writer.TryComplete(); return ValueTask.CompletedTask; }
    }

    internal static async Task RunAsync(string directory)
    {
        Directory.CreateDirectory(directory);
        using var reservation = new TcpListener(IPAddress.Loopback, 0);
        reservation.Start();
        int port = ((IPEndPoint)reservation.LocalEndpoint).Port;
        reservation.Stop();
        using var stop = new CancellationTokenSource(TimeSpan.FromMinutes(3));
        var root = new Ensemble.Ensemble(TextWriter.Null);
        Task? hosting = null;
        try
        {
            var agent = new Agent(root,
                new Configuration(Vendor.Anthropic, "offline", apiKey: "", baseUrl: "http://127.0.0.1:1"),
                outputDirectory: Path.Combine(directory, "jobs"));
            var remote = new McpClient(agent.EnsureMcp(), "renderer-fixture", new RendererPeer(["remote_weather", "command"]));
            await remote.InstallToolsAsync(stop.Token);
            if (!agent.Tools().Any(tool => tool.Name() == "remote_weather" && tool.Remote() == remote))
                throw new InvalidOperationException("MCP renderer fixture tool not installed");
            var peerAgent = new Agent(root,
                new Configuration(Vendor.Anthropic, "offline", apiKey: "", baseUrl: "http://127.0.0.1:1"),
                name: "renderer-peer", outputDirectory: Path.Combine(directory, "peer-jobs"));
            var peerRemote = new McpClient(peerAgent.EnsureMcp(), "peer-renderer-fixture",
                new RendererPeer(["remote_weather", "peer_only"]));
            await peerRemote.InstallToolsAsync(stop.Token);
            hosting = Gui.RunAsync(agent, new GuiOptions(port,
                SettingsPath: Path.Combine(directory, "settings.json"),
                TracePath: Path.Combine(directory, "gui.log"),
                RendererDirectory: Path.Combine(directory, "renderers")), stop.Token);
            string url = $"http://127.0.0.1:{port}/";
            using var client = new HttpClient();
            while (true)
            {
                if (hosting.IsCompleted) { await hosting; throw new InvalidOperationException("Host exited before readiness"); }
                try
                {
                    using var response = await client.GetAsync(url, stop.Token);
                    if (response.IsSuccessStatusCode) break;
                }
                catch (HttpRequestException) { }
                await Task.Delay(25, stop.Token);
            }
            await File.WriteAllTextAsync(Path.Combine(directory, "url"), url, stop.Token);
            while (!File.Exists(Path.Combine(directory, "stop")))
            {
                if (hosting.IsCompleted) { await hosting; throw new InvalidOperationException("Host exited before stop"); }
                await Task.Delay(25, stop.Token);
            }
        }
        finally
        {
            stop.Cancel();
            try { if (hosting is not null) await hosting; }
            catch (OperationCanceledException) when (stop.IsCancellationRequested) { }
            finally { root.Destroy(); }
        }
        Console.WriteLine("STOPPED artifact renderer fixture and owned Agent");
    }
}
