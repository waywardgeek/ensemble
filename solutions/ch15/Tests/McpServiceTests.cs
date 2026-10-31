using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

static partial class McpTests
{
    public static async Task IncomingService()
    {
        string directory = Path.Combine(Path.GetTempPath(), "mcp-service-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(Path.Combine(directory, "skills", "primary"));
        File.WriteAllText(Path.Combine(directory, "skills", "primary", "SKILL.md"),
            "---\nname: primary\ntype: primary\ntools: read_file\n---\nRead only.\n");
        string marker = Path.Combine(directory, "marker.txt");
        string forbidden = Path.Combine(directory, "forbidden.txt");
        File.WriteAllText(marker, "SERVICE_REAL_FILE");
        var pair = await Pair();
        var root = new Application(TextWriter.Null);
        var agent = new Agent(root, new Configuration(Vendor.OpenAI, "offline", "http://127.0.0.1:1", "dummy"),
            Path.Combine(directory, "io"));
        agent.InstallTools();
        var manager = new Mcp(agent);
        var service = new McpClient(manager, "service", pair.Client);
        try
        {
            _ = new Skills(agent, new SkillSpec(Path.Combine(directory, "skills"), "primary"));
            Task serving = service.ServeAsync();
            Check(ReferenceEquals(serving, service.ServeAsync()), "Repeated service start did not join");
            await pair.Peer.SendAsync(J("{\"jsonrpc\":\"2.0\",\"id\":0,\"method\":\"tools/list\"}"));
            var refused = await Read(pair.Peer);
            Check(refused.GetProperty("id").GetInt64() == 0 && refused.TryGetProperty("error", out _),
                "Service admitted discovery before initialization or lost ID zero");
            await pair.Peer.SendAsync(J("{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"initialize\",\"params\":{\"protocolVersion\":\"unsupported\"}}"));
            Check((await Read(pair.Peer)).TryGetProperty("error", out _), "Unsupported protocol was accepted");
            await pair.Peer.SendAsync(J("{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"initialize\",\"params\":{\"protocolVersion\":\"2024-11-05\"}}"));
            var initialized = await Read(pair.Peer);
            Check(initialized.GetProperty("id").GetInt64() == 2
                && initialized.GetProperty("result").GetProperty("protocolVersion").GetString() == "2024-11-05",
                "Service initialize response was not correlated");
            await pair.Peer.SendAsync(J("{\"jsonrpc\":\"2.0\",\"id\":3,\"method\":\"tools/list\"}"));
            Check((await Read(pair.Peer)).TryGetProperty("error", out _), "Service skipped initialized notification");
            await pair.Peer.SendAsync(J("{\"jsonrpc\":\"2.0\",\"method\":\"notifications/initialized\"}"));
            await pair.Peer.SendAsync(J("{\"jsonrpc\":\"2.0\",\"id\":4,\"method\":\"tools/list\"}"));
            var listed = await Read(pair.Peer);
            var names = listed.GetProperty("result").GetProperty("tools").EnumerateArray()
                .Select(tool => tool.GetProperty("name").GetString()).ToArray();
            Check(names.Contains("read_file") && !names.Contains("write_file"), "Service discovery bypassed skill grants");
            await pair.Peer.SendAsync(JsonSerializer.SerializeToElement(new
            {
                jsonrpc = "2.0", id = 5, method = "tools/call",
                @params = new { name = "write_file", arguments = new { path = forbidden, content = "NO" } }
            }));
            var denied = await Read(pair.Peer);
            Check(denied.GetProperty("id").GetInt64() == 5
                && denied.GetProperty("result").GetProperty("isError").GetBoolean()
                && !File.Exists(forbidden), "Denied service tool had effects");
            await pair.Peer.SendAsync(JsonSerializer.SerializeToElement(new
            {
                jsonrpc = "2.0", id = 6, method = "tools/call",
                @params = new { name = "read_file", arguments = new { path = marker } }
            }));
            var read = await Read(pair.Peer);
            Check(read.GetProperty("id").GetInt64() == 6
                && !read.GetProperty("result").GetProperty("isError").GetBoolean()
                && read.GetProperty("result").GetRawText().Contains("SERVICE_REAL_FILE"),
                "Service did not return the real correlated file result");
            await pair.Peer.SendAsync(J("{\"jsonrpc\":\"2.0\",\"id\":7,\"method\":\"unknown\"}"));
            Check((await Read(pair.Peer)).GetProperty("error").GetProperty("code").GetInt32() == -32601,
                "Service unknown-method error was missing");
            try
            {
                await service.ConnectAsync();
                throw new Exception("Service started a second protocol reader");
            }
            catch (InvalidOperationException) { }
            try
            {
                await service.RequestAsync("tools/list", J("{}"));
                throw new Exception("Service initiated outgoing RPC");
            }
            catch (InvalidOperationException) { }
            Check(!agent.History().Events().Any(item => item.Kind == EventKind.RequestSent),
                "Incoming service invented a model request");
            await pair.Peer.DisposeAsync();
            await serving.WaitAsync(TimeSpan.FromSeconds(8));
            await service.DestroyAsync();
            Check(manager.Clients().Length == 0, "EOF service failed to detach");
        }
        finally
        {
            await pair.Peer.DisposeAsync();
            root.Destroy();
        }
    }

    public static async Task ServiceTermination()
    {
        foreach (string ending in new[] { "malformed", "cancel", "shutdown" })
        {
            var pair = await Pair();
            var root = new Application(TextWriter.Null);
            var agent = new Agent(root, new Configuration(Vendor.OpenAI, "offline", "http://127.0.0.1:1", "dummy"),
                Path.Combine(Path.GetTempPath(), "mcp-service-stop-" + Guid.NewGuid().ToString("N")));
            var manager = new Mcp(agent);
            var service = new McpClient(manager, "service", pair.Client);
            using var cancellation = new CancellationTokenSource();
            try
            {
                Task serving = service.ServeAsync(cancellation.Token);
                if (ending == "malformed")
                {
                    await pair.Peer.SendAsync(J("{\"jsonrpc\":\"2.0\",\"id\":1.5,\"method\":\"initialize\"}"));
                }
                else if (ending == "cancel")
                {
                    cancellation.Cancel();
                }
                else
                {
                    await agent.ShutdownAsync().WaitAsync(TimeSpan.FromSeconds(8));
                }
                await Fails(serving);
                await service.DestroyAsync();
                Check(manager.Clients().Length == 0, "Terminated service remained attached: " + ending);
            }
            finally
            {
                await pair.Peer.DisposeAsync();
                root.Destroy();
            }
        }
    }
}
