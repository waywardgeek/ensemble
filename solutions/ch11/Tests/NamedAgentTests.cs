using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    public static async Task NamedAgents()
    {
        var app = new Application(TextWriter.Null);
        var other = new Application(TextWriter.Null);
        try
        {
            foreach (var vendor in new[] { Vendor.Anthropic, Vendor.OpenAI, Vendor.Gemini })
            {
                string name = vendor.ToString();
                string instructions = "ROLE_" + name + "_ONLY";
                var configuration = new Configuration(vendor, "text-only", "http://127.0.0.1:1",
                    primaryInstructions: instructions);
                var agent = new Agent(app, configuration, name: name);
                Check(ReferenceEquals(app.FindAgent(name), agent), "Named lookup lost actual Agent identity.");
                Check(ReferenceEquals(agent.Parent(), app), "Named Agent lost parent.");
                agent.InstallTools();
                agent.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                    parts: [new TextPart("hello")]));
                string body = agent.Engine().Render(agent.History().View()).ToJsonString();
                Check(body.Split(instructions).Length == 2, "Tools erased or duplicated role instructions.");
                Check(body.Contains("Before each tool batch"), "Tools lost their use guidance.");
                int count = app.Agents().Length;
                try
                {
                    _ = new Agent(app, configuration, name: name);
                    throw new Exception("Duplicate name accepted.");
                }
                catch (InvalidOperationException) { }
                Check(app.Agents().Length == count && ReferenceEquals(app.FindAgent(name), agent),
                    "Duplicate-name refusal changed ownership.");

                // Removal completes both ends. The detached object retains its
                // immutable name but cannot be found through its former parent.
                app.RemoveAgent(agent);
                Check(agent.parent is null && app.FindAgent(name) is null, "Incomplete named removal.");
                other.AddAgent(agent);
                Check(ReferenceEquals(agent.Parent(), other) && ReferenceEquals(other.FindAgent(name), agent),
                    "Reattachment did not establish both ends.");
                var replacement = new Agent(app, configuration, name: name);
                agent.Destroy();
                Check(other.FindAgent(name) is null && ReferenceEquals(app.FindAgent(name), replacement),
                    "Destroy crossed root ownership or retained stale lookup.");
            }

            var listener = new TcpListener(IPAddress.Loopback, 0);
            listener.Start();
            using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(10));
            try
            {
                int port = ((IPEndPoint)listener.LocalEndpoint).Port;
                var author = new Agent(app, new Configuration(Vendor.Anthropic, "local",
                    "http://127.0.0.1:" + port, "dummy", primaryInstructions: "AUTHOR_ROLE"), name: "author");
                var editor = new Agent(app, new Configuration(Vendor.Anthropic, "local",
                    "http://127.0.0.1:" + port, "dummy", primaryInstructions: "EDITOR_ROLE"), name: "editor");
                var server = Task.Run(async () =>
                {
                    using var first = await listener.AcceptTcpClientAsync(deadline.Token);
                    using var firstStream = first.GetStream();
                    string firstBody = await ReadRequest(firstStream, deadline.Token);
                    using var second = await listener.AcceptTcpClientAsync(deadline.Token);
                    using var secondStream = second.GetStream();
                    string secondBody = await ReadRequest(secondStream, deadline.Token);
                    using var firstRequest = JsonDocument.Parse(firstBody);
                    using var secondRequest = JsonDocument.Parse(secondBody);
                    string firstRole = firstRequest.RootElement.GetProperty("system").GetString()!;
                    string secondRole = secondRequest.RootElement.GetProperty("system").GetString()!;
                    Check(new HashSet<string> { firstRole, secondRole }.SetEquals(["AUTHOR_ROLE", "EDITOR_ROLE"]),
                        "Concurrent Agents sent the wrong roles.");
                    await Reply(firstStream, firstRole, deadline.Token);
                    await Reply(secondStream, secondRole, deadline.Token);
                });
                Task<string> authorReply = author.SendAsync("author input", deadline.Token);
                Task<string> editorReply = editor.SendAsync("editor input", deadline.Token);
                await Task.WhenAll(authorReply, editorReply, server);
                Check(await authorReply == "AUTHOR_ROLE" && await editorReply == "EDITOR_ROLE", "Replies crossed Agents.");
                Check(!author.History().Events().SelectMany(fact => fact.Parts).OfType<TextPart>()
                    .Any(part => part.Text == "editor input"), "Cross-Agent history.");
                Check(app.FindAgent("missing") is null, "Missing name returned an unrelated Agent.");
            }
            finally
            {
                listener.Stop();
                deadline.Cancel();
            }
        }
        finally
        {
            app.Destroy();
            other.Destroy();
        }
    }
}
