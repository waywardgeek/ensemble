using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    public static async Task SkillExchange()
    {
        foreach (bool deferred in new[] { false, true })
        {
            string root = SkillTests.Fixture();
            string target = Path.Combine(root, "written.txt");
            var listener = new TcpListener(IPAddress.Loopback, 0);
            listener.Start();
            using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
            var app = new Application(TextWriter.Null);
            try
            {
                var server = Task.Run(async () =>
                {
                    string primary = "";
                    for (int round = 0; round < 3; round++)
                    {
                        using var client = await listener.AcceptTcpClientAsync(deadline.Token);
                        using var stream = client.GetStream();
                        using var request = JsonDocument.Parse(await ReadRequest(stream, deadline.Token));
                        var body = request.RootElement;
                        string instructions = body.GetProperty("system").GetString()!;
                        var tools = body.GetProperty("tools").EnumerateArray().ToArray();
                        bool Has(string name) => tools.Any(tool => tool.GetProperty("name").GetString() == name);
                        Check(Has("read_file") && Has("load_skill"), "Eager declaration missing");
                        if (round == 0)
                        {
                            primary = instructions;
                            Check(Has("write_file") == deferred, "Initial disabled schema contract wrong");
                            if (deferred)
                            {
                                Check(tools.Single(tool => tool.GetProperty("name").GetString() == "write_file")
                                    .GetProperty("defer_loading").GetBoolean(), "Later tool was not deferred");
                                Check(!tools.Single(tool => tool.GetProperty("name").GetString() == "load_skill")
                                    .TryGetProperty("defer_loading", out _), "Loader was not eager");
                            }
                            await ReplyJson(stream, new
                            {
                                content = new object[]
                                {
                                    new { type = "text", text = "I will load editing before writing." },
                                    new { type = "tool_use", id = "load-1", name = "load_skill", input = new { name = "editing" } }
                                },
                                usage = new { input_tokens = 1, output_tokens = 2 }
                            }, deadline.Token);
                        }
                        else
                        {
                            Check(instructions.StartsWith(primary, StringComparison.Ordinal), "Primary instructions changed");
                            Check(instructions.Contains("EDIT-once") && Has("write_file"), "Activation not visible next round");
                            var messages = body.GetProperty("messages").EnumerateArray().ToArray();
                            Check(!body.GetProperty("messages").GetRawText().Contains("EDIT-once"), "Instructions leaked into ordinary dialogue");
                            var result = messages.SelectMany(message => message.GetProperty("content").EnumerateArray())
                                .Single(part => part.GetProperty("type").GetString() == "tool_result"
                                    && part.GetProperty("tool_use_id").GetString() == "load-1");
                            var content = result.GetProperty("content").EnumerateArray().ToArray();
                            Check(content.Any(part => part.GetProperty("type").GetString() == "tool_reference") == deferred,
                                "Discovery result used the wrong provider projection");
                            if (deferred)
                            {
                                var names = content.Select(part => part.GetProperty("tool_name").GetString()).Order().ToArray();
                                Check(names.SequenceEqual(new[] { "list_directory", "write_file" }), "References did not match newly granted tools");
                            }
                            if (round == 1)
                            {
                                await ReplyJson(stream, new
                                {
                                    content = new object[]
                                    {
                                        new { type = "tool_use", id = "write-1", name = "write_file", input = new { path = target, content = "actual effect" } }
                                    },
                                    usage = new { input_tokens = 1, output_tokens = 2 }
                                }, deadline.Token);
                            }
                            else
                            {
                                Check(File.ReadAllText(target) == "actual effect", "Tool effect missing before final request");
                                await Reply(stream, "skill task complete", deadline.Token);
                            }
                        }
                    }
                });
                var agent = new Agent(app, new Configuration(Vendor.Anthropic, "model-a",
                    $"http://127.0.0.1:{((IPEndPoint)listener.LocalEndpoint).Port}", deferredTools: deferred),
                    skillSpec: new SkillSpec(root, "root", new Dictionary<string, string> { ["WHO"] = "once" }));
                string answer = await agent.SendAsync("Make the file").WaitAsync(deadline.Token);
                Check(answer == "skill task complete", "Reliable final reply changed");
                await server.WaitAsync(deadline.Token);
                var discovered = agent.History().Events().Where(value => value.Kind == EventKind.ToolReturned)
                    .Single(result => result.CallID == "load-1");
                Check(discovered.ToolReferences.Length == 2, "Canonical discovery metadata missing");
                string line = JsonCodec.WriteEvent(agent.History().Events().First(value => value.Kind == EventKind.ToolReturned));
                Check(JsonCodec.WriteEvent(JsonCodec.ReadEvent(line)) == line, "Discovery metadata did not round-trip");
            }
            finally
            {
                deadline.Cancel();
                listener.Stop();
                app.Destroy();
                Directory.Delete(root, true);
            }
        }
    }
}
