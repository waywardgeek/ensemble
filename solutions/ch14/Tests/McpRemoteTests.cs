using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

static partial class McpTests
{
    static async Task Discover(McpClient client, IMcpTransport peer, string definitions)
    {
        Task connect = client.ConnectAsync();
        var initialize = await Read(peer);
        await Reply(peer, initialize, new { protocolVersion = "2024-11-05" });
        var ready = await Read(peer);
        Check(ready.GetProperty("method").GetString() == "notifications/initialized", "initialized notification");
        var list = await Read(peer);
        await Reply(peer, list, J("{\"tools\":" + definitions + "}"));
        await connect.WaitAsync(TimeSpan.FromSeconds(8));
    }

    static readonly string MixedContent = """
        {"isError":true,"content":[{"type":"text","text":"first"},
        {"type":"image","mimeType":"image/png","data":"AQID"},
        {"type":"audio","mimeType":"audio/wav","data":"BAUG"},
        {"type":"text","text":"last"}]}
        """;

    static void CheckMixed(ToolResultPart result)
    {
        Check(result.IsError && result.Parts.Length == 4, "error flag and every ordered block survive");
        Check(result.Parts[0] is TextPart { Text: "first" }
            && result.Parts[1] is BlobPart { Mime: "image/png" }
            && result.Parts[2] is BlobPart { Mime: "audio/wav" }
            && result.Parts[3] is TextPart { Text: "last" }, "text/image/audio/text order");
        Check(JsonElement.DeepEquals(McpContent.Encode(result), J(MixedContent)),
            "MCP payload round-trips without reordering or flattening");
    }

    public static async Task RemoteJobs()
    {
        var streams = await Pair();
        await using var peer = streams.Peer;
        var root = new Application(TextWriter.Null);
        try
        {
            var configuration = new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1");
            var agent = new Agent(root, configuration);
            agent.InstallTools();
            var manager = new Mcp(agent);
            var client = new McpClient(manager, "media", streams.Client);
            await Discover(client, peer, """
                [{"name":"mixed","inputSchema":{"type":"object"}},
                 {"name":"snapshot","inputSchema":{"type":"object"},"ephemeral":"round"}]
                """);
            await client.InstallToolsAsync().WaitAsync(TimeSpan.FromSeconds(8));
            await client.InstallToolsAsync().WaitAsync(TimeSpan.FromSeconds(8));
            var remote = client.Tools().Single();
            Check(remote.Parent() == agent && remote.Remote() == client, "two actual strong parents");
            client.RemoveTool(remote);
            Check(remote.Remote() is null && client.Tools().IsEmpty && agent.Tools().Contains(remote),
                "removing one ownership edge completes both ends and preserves the other owner");
            client.AddTool(remote);
            agent.RemoveTool(remote);
            var foreign = new Agent(root, configuration);
            bool foreignRefused = false;
            try
            {
                foreign.AddTool(remote);
            }
            catch (InvalidOperationException)
            {
                foreignRefused = true;
            }
            Check(foreignRefused && foreign.Tools().IsEmpty && client.Tools().Contains(remote),
                "foreign Agent cannot borrow a connection's Tool");
            agent.AddTool(remote);
            Check(remote.Parent() == agent && remote.Remote() == client, "explicit reattachment restores both relationships");
            Check(!agent.Tools().Any(tool => tool.Name() == "snapshot"), "automatic tool never enters ordinary registry");

            var call = new ToolCallPart("remote-1", configuration.Target(), "mixed", J("{}"));
            var work = agent.DispatchAsync(call);
            var request = await Read(peer);
            Check(request.GetProperty("method").GetString() == "tools/call"
                && request.GetProperty("params").GetProperty("name").GetString() == "mixed", "ordinary MCP invocation");
            await Reply(peer, request, J(MixedContent));
            var result = await work.WaitAsync(TimeSpan.FromSeconds(8));
            CheckMixed(result);
            Check(result.Job is { Status: "done" }, "remote call has a real completed Job");
            var fact = agent.History().Events().Last(e => e.Kind == EventKind.ToolObserved);
            Check(fact.IsError && fact.Parts.Length == 4 && fact.Parts[1] is BlobPart, "canonical host result retains media");
            string output = File.ReadAllText(result.Job!.OutputPath);
            Check(output.Contains("AQID", StringComparison.Ordinal), "output file retains full structured payload");
            var again = await agent.Jobs().Find(result.Job.Handle).ObserveAsync("again", delay: 0);
            Check(again.Parts.All(part => part is TextPart) && !JsonCodec.Text(again.Parts).Contains("AQID"),
                "completed content is consumed once");

            // A server can call one of our enabled tools while its reader
            // remains available for the nested outbound RPC on this same link.
            await peer.SendAsync(J("""
                {"jsonrpc":"2.0","id":"reverse-media","method":"tools/call",
                 "params":{"name":"mixed","arguments":{}}}
                """));
            var nested = await Read(peer);
            Check(nested.GetProperty("method").GetString() == "tools/call", "reverse dispatch reaches ordinary remote Job");
            await Reply(peer, nested, J(MixedContent));
            var reverseMedia = await Read(peer);
            Check(reverseMedia.GetProperty("id").GetString() == "reverse-media"
                && JsonElement.DeepEquals(reverseMedia.GetProperty("result"), J(MixedContent)),
                "reverse conversion retains all media/error content without blocking the reader");

            // A foreground timeout is not remote cancellation. The running Job
            // later supplies its complete typed result to wait_for_job.
            var held = await agent.DispatchAsync(new ToolCallPart("remote-2", configuration.Target(), "mixed",
                J("{\"ai_callback_delay\":0}"))).WaitAsync(TimeSpan.FromSeconds(8));
            var heldRequest = await Read(peer);
            Check(held.Job is { Status: "running" }, "observation can return before remote completion");
            await Reply(peer, heldRequest, J(MixedContent));
            var waited = await agent.DispatchAsync(new ToolCallPart("wait", configuration.Target(), "wait_for_job",
                JsonSerializer.SerializeToElement(new { handle = held.Job!.Handle, ai_callback_delay = 5 })))
                .WaitAsync(TimeSpan.FromSeconds(8));
            CheckMixed(waited);
            Check(waited.CallID == "wait", "observation uses its caller's correlation identity");

            var killing = await agent.DispatchAsync(new ToolCallPart("remote-3", configuration.Target(), "mixed",
                J("{\"ai_callback_delay\":0}"))).WaitAsync(TimeSpan.FromSeconds(8));
            var killedRequest = await Read(peer);
            var kill = agent.DispatchAsync(new ToolCallPart("kill", configuration.Target(), "kill_job",
                JsonSerializer.SerializeToElement(new { handle = killing.Job!.Handle })));
            var cancel = await Read(peer);
            Check(cancel.GetProperty("method").GetString() == "$/cancelRequest"
                && cancel.GetProperty("params").GetProperty("id").GetInt64() == killedRequest.GetProperty("id").GetInt64(),
                "kill sends the exact remote request identity");
            await kill.WaitAsync(TimeSpan.FromSeconds(8));
            Check(agent.Jobs().Find(killing.Job!.Handle).Snapshot().Status == "killed", "kill remains terminal");
            await Reply(peer, killedRequest, J(MixedContent));
            var next = client.RequestAsync("ping", J("{}"));
            await Reply(peer, await Read(peer), new { ok = true });
            Check((await next.WaitAsync(TimeSpan.FromSeconds(8))).GetProperty("ok").GetBoolean(), "late cancelled result cannot poison another RPC");

            await client.DestroyAsync().WaitAsync(TimeSpan.FromSeconds(8));
            Check(remote.Remote() is null && client.Tools().IsEmpty
                && !agent.Tools().Contains(remote), "connection teardown detaches both Tool parents");
        }
        finally
        {
            root.Destroy();
        }
    }

    public static async Task RemoteRegistration()
    {
        var streams = await Pair();
        await using var peer = streams.Peer;
        var root = new Application(TextWriter.Null);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "fake", "http://127.0.0.1"));
            agent.InstallTools();
            var client = new McpClient(new Mcp(agent), "collision", streams.Client);
            await Discover(client, peer, """
                [{"name":"unique","inputSchema":{"type":"object"}},
                 {"name":"read_file","inputSchema":{"type":"object"}}]
                """);
            bool refused = false;
            try
            {
                await client.InstallToolsAsync().WaitAsync(TimeSpan.FromSeconds(8));
            }
            catch (ArgumentException)
            {
                refused = true;
            }
            Check(refused && client.Tools().IsEmpty && !agent.Tools().Any(t => t.Name() == "unique"),
                "collision refuses the whole registration before mutation");
            // Provider projection uses the same neutral blobs. Anthropic
            // supports images inside tool results; the other current adapters
            // must refuse that placement rather than silently keep text only.
            foreach (var model in new[]
            {
                (Vendor.Anthropic, "claude-3-haiku-20240307"),
                (Vendor.OpenAI, "gpt-4o"),
                (Vendor.Gemini, "gemini-2.5-flash")
            })
            {
                var renderer = new Agent(root, new Configuration(model.Item1, model.Item2, "http://127.0.0.1"));
                var payload = McpContent.Decode("image", J("""
                    {"isError":true,"content":[{"type":"text","text":"before"},
                    {"type":"image","mimeType":"image/png","data":"AQID"},
                    {"type":"text","text":"after"}]}
                    """));
                var invocation = new ToolCallPart("image", renderer.Configuration().Target(), "remote", J("{}"));
                var context = new Context([
                    new Entry(1, Actor.Agent, [invocation]),
                    new Entry(2, Actor.Tool, [payload])]);
                bool unsupported = false;
                try
                {
                    var request = renderer.Engine().Render(context);
                    var block = request["messages"]![1]!["content"]![0]!;
                    var content = block["content"]!.AsArray();
                    Check(block["is_error"]!.GetValue<bool>() && content.Count == 3
                        && content[0]!["text"]!.GetValue<string>() == "before"
                        && content[1]!["source"]!["data"]!.GetValue<string>() == "AQID"
                        && content[2]!["text"]!.GetValue<string>() == "after", "provider retains image/error/order");
                }
                catch (NotSupportedException)
                {
                    unsupported = true;
                }
                Check(unsupported == (model.Item1 != Vendor.Anthropic), "unsupported placement is an explicit refusal");
                Check(((BlobPart)payload.Parts[1]).Ref.Data().SequenceEqual(new byte[] { 1, 2, 3 }), "inline bytes decode exactly");
            }
            foreach (string invalid in new[]
            {
                "{\"isError\":\"false\",\"content\":[]}",
                "{\"content\":[{\"type\":\"image\",\"mimeType\":\"image/png\",\"data\":\"not base64\"}]}",
                "{\"content\":[{\"type\":\"text\",\"text\":\"keep\"},{\"type\":\"resource\"}]}"
            })
            {
                bool invalidResult = false;
                try
                {
                    McpContent.Decode("invalid", J(invalid));
                }
                catch (InvalidDataException)
                {
                    invalidResult = true;
                }
                Check(invalidResult, "invalid/unsupported content refuses the whole result, not just one block");
            }
        }
        finally
        {
            root.Destroy();
        }
    }
}
