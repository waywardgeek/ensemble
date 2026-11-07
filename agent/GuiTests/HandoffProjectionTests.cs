using System.Net;
using System.Net.Sockets;
using System.Net.WebSockets;
using System.Text.Json;
using Ensemble;
using EnsembleGui;

internal static class HandoffProjectionTests
{
    public static async Task RunAsync()
    {
        string directory = Directory.CreateTempSubdirectory("handoff-gui-").FullName;
        var root = new Ensemble.Ensemble(TextWriter.Null, dataDirectory: directory);
        try
        {
            var agent = new Agent(root, new Configuration(Vendor.OpenAI, "offline", "http://127.0.0.1:1"));
            var history = agent.History();
            var stamp = DateTimeOffset.Parse("2026-01-01T00:00:00Z");
            history.Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("compressed conversation")]));
            history.Append(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("old checkpoint")]));
            history.Append(new Event(EventKind.HandoffRecorded, parts: [new HandoffPart("new checkpoint")]));
            var launch = new CompressionLaunch("fixture", null, MemoryBand.Session,
                "summary", stamp, new string('a', 64), 1, 2, tracePath: "memory/fixture.jsonl");
            history.Append(new Event(EventKind.CompressionLaunched, compression: launch));
            history.Append(new Event(EventKind.Redacted, redactFrom: 1, redactTo: 2,
                level: Redaction.Summary, reason: "memory compression", taskID: launch.TaskID));
            history.Append(new Event(EventKind.MemoryPopulated, taskID: launch.TaskID,
                memory: new MemoryEntry(MemoryBand.Session, "summary", stamp, "saved summary")));
            history.Append(new Event(EventKind.CompressionFinished,
                outcome: new CompressionOutcome(launch.TaskID, CompressionState.Committed)));
            history.Append(new Event(EventKind.TurnEnded, reason: "completed"));
            var snapshot = history.Snapshot();
            if (snapshot.Context.Dialogue.SelectMany(entry => entry.Parts).OfType<HandoffPart>()
                .Single().Text != "new checkpoint")
                throw new Exception("Core did not retain only the newest checkpoint");
            var projection = Gui.Project(snapshot, "fixture");
            string frames = string.Join("\n", projection.Frames.Select(frame => frame.ToJsonString()));
            if (!frames.Contains("new checkpoint") || frames.Contains("old checkpoint")
                || frames.Contains("compressed conversation") || projection.State != "idle")
                throw new Exception("GUI replay changed the authoritative transcript or final state");
            if (snapshot.Context.Memories.Single().Text != "saved summary"
                || history.Events().Count(fact => fact.Kind == EventKind.HandoffRecorded) != 2)
                throw new Exception("Display projection changed memory or canonical checkpoints");
            await ConnectAsync(agent, directory);
            Console.WriteLine("PASS GUI projects compressed history and delivers websocket replay after newest-only handoff");
        }
        finally { root.Destroy(); Directory.Delete(directory, true); }
    }

    private static async Task ConnectAsync(Agent agent, string directory)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(20));
        using var stop = new CancellationTokenSource();
        using var reservation = new TcpListener(IPAddress.Loopback, 0);
        reservation.Start();
        int port = ((IPEndPoint)reservation.LocalEndpoint).Port;
        reservation.Stop();
        string origin = $"http://127.0.0.1:{port}";
        var serving = Gui.RunAsync(agent, new GuiOptions(port,
            SettingsPath: Path.Combine(directory, "appearance.json"),
            TracePath: Path.Combine(directory, "gui.log")), stop.Token);
        using var socket = new ClientWebSocket();
        try
        {
            using var http = new HttpClient();
            while (true)
            {
                if (serving.IsCompleted) await serving;
                try
                {
                    using var response = await http.GetAsync(origin, deadline.Token);
                    if (response.IsSuccessStatusCode) break;
                }
                catch (HttpRequestException) { }
                await Task.Delay(20, deadline.Token);
            }
            socket.Options.SetRequestHeader("Origin", origin);
            await socket.ConnectAsync(new Uri($"ws://127.0.0.1:{port}/ws"), deadline.Token);
            var observed = new HashSet<string>();
            byte[] buffer = new byte[8192];
            while (!observed.IsSupersetOf(["connection", "replay", "current_settings"]))
            {
                using var body = new MemoryStream();
                WebSocketReceiveResult received;
                do
                {
                    received = await socket.ReceiveAsync(buffer, deadline.Token);
                    if (received.MessageType != WebSocketMessageType.Text)
                        throw new Exception("Socket closed before replay/settings delivery");
                    body.Write(buffer, 0, received.Count);
                    if (body.Length > 1024 * 1024) throw new Exception("Fixture frame exceeded limit");
                } while (!received.EndOfMessage);
                using var document = JsonDocument.Parse(body.ToArray());
                string type = document.RootElement.GetProperty("type").GetString()!;
                observed.Add(type);
                if (type == "replay")
                {
                    string replay = document.RootElement.GetProperty("data").GetProperty("events").GetRawText();
                    if (!replay.Contains("new checkpoint") || replay.Contains("old checkpoint"))
                        throw new Exception("Socket replay disagrees with retained checkpoint");
                }
            }
        }
        finally
        {
            socket.Abort();
            stop.Cancel();
            await serving.WaitAsync(TimeSpan.FromSeconds(10));
        }
    }

}
