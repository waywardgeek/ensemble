using System.Net;
using System.Net.Sockets;
using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using Ensemble;
using EnsembleGui;

static class HistoryWindowTests
{
    static void Check(bool condition, string message)
    {
        if (!condition) throw new InvalidOperationException(message);
    }

    public static async Task RunAsync()
    {
        // Large snapshot-only histories must be bounded before JSON projection.
        var entries = Enumerable.Range(1, 10000).Select(i =>
            new Entry(i, Actor.Human, [new TextPart("message " + i)])).ToArray();
        var snapshot = new HistorySnapshot([], new Context(entries), 10000);
        var projected = Gui.Project(snapshot, "parent");
        Check(projected.FinalIDs.Count == 100 && projected.Frames.All(frame => frame["seq"]!.GetValue<long>() >= 9901),
            "Default replay did not retain the newest 100 display events.");
        Check(Gui.Project(snapshot, "parent", 7).FinalIDs.Count == 7 && snapshot.Context.Dialogue.Length == 10000,
            "Display limit changed saved history or ignored the explicit limit.");
        var paused = snapshot with { Events = [new Event(EventKind.AgentPaused, seq: 1)] };
        Check(Gui.Project(paused, "parent").State == "paused", "Window forgot earlier pause state.");

        string directory = Directory.CreateTempSubdirectory("gui-history-window-").FullName;
        var root = new Ensemble.Ensemble(TextWriter.Null);
        using var stop = new CancellationTokenSource(TimeSpan.FromSeconds(30));
        Task? serving = null;
        try
        {
            var config = new Configuration(Vendor.OpenAI, "offline", "http://127.0.0.1:1");
            var parent = new Agent(root, config, name: "parent");
            var child = new Agent(root, parent, "child", config);
            foreach (var agent in new[] { parent, child })
                for (int i = 0; i < 250; i++)
                    agent.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                        parts: [new TextPart(agent.Address() + " " + i)]));
            using var reservation = new TcpListener(IPAddress.Loopback, 0);
            reservation.Start();
            int port = ((IPEndPoint)reservation.LocalEndpoint).Port;
            reservation.Stop();
            serving = Gui.RunAsync(parent, new GuiOptions(port, SettingsPath: Path.Combine(directory, "settings.json"),
                TracePath: Path.Combine(directory, "gui.log")), stop.Token);
            using var http = new HttpClient();
            while (true)
            {
                if (serving.IsCompleted) await serving;
                try { if ((await http.GetAsync($"http://127.0.0.1:{port}", stop.Token)).IsSuccessStatusCode) break; }
                catch (HttpRequestException) { }
                await Task.Delay(20, stop.Token);
            }
            using var socket = new ClientWebSocket();
            await socket.ConnectAsync(new Uri($"ws://127.0.0.1:{port}/ws"), stop.Token);
            var replay = await Read(socket, "replay", stop.Token);
            Check(replay.GetProperty("data").GetProperty("events").EnumerateArray()
                .Count(frame => frame.TryGetProperty("part_id", out _)) == 100, "Socket replay is unbounded.");
            await socket.SendAsync(Encoding.UTF8.GetBytes("{\"type\":\"select_agent\",\"agent_id\":\"parent/child\"}"),
                WebSocketMessageType.Text, true, stop.Token);
            replay = await Read(socket, "replay", stop.Token);
            Check(replay.GetProperty("data").GetProperty("events").EnumerateArray()
                .All(frame => frame.GetProperty("agent_id").GetString() == "parent/child"), "Replay mixed agents.");
            for (int i = 0; i < 125; i++)
                foreach (var agent in new[] { parent, child })
                    agent.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                        parts: [new TextPart("live " + i)]));
            string newest = $"parent/child:e{child.History().Snapshot().Sequence}:p0";
            JsonElement window;
            do { window = await Read(socket, "display_window", stop.Token); }
            while (!window.GetProperty("part_ids").EnumerateArray().Any(id => id.GetString() == newest));
            Check(window.GetProperty("agent_id").GetString() == "parent/child"
                && window.GetProperty("part_ids").GetArrayLength() == 100,
                "Live window is unbounded or attributed to another agent.");
            Check(parent.History().View().Dialogue.Length == 375 && child.History().View().Dialogue.Length == 375,
                "Display eviction altered either real history.");
            socket.Abort();
        }
        finally
        {
            stop.Cancel();
            if (serving is not null) await serving;
            root.Destroy();
            Directory.Delete(directory, true);
        }
        Console.WriteLine("PASS bounded snapshot/live display, preserved history, parent/child routing and earlier pause state");
    }

    static async Task<JsonElement> Read(ClientWebSocket socket, string type, CancellationToken token)
    {
        byte[] buffer = new byte[8192];
        while (true)
        {
            using var bytes = new MemoryStream();
            WebSocketReceiveResult part;
            do
            {
                part = await socket.ReceiveAsync(buffer, token);
                Check(part.MessageType == WebSocketMessageType.Text, "Unexpected socket close.");
                bytes.Write(buffer, 0, part.Count);
            } while (!part.EndOfMessage);
            using var document = JsonDocument.Parse(bytes.ToArray());
            if (document.RootElement.GetProperty("type").GetString() == type) return document.RootElement.Clone();
        }
    }
}
