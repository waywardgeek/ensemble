using System.Collections.Immutable;
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

    private static void RecallContent()
    {
        const string text = "[Auto-recalled memories]\n> Exact supplied text <img src=x>\n> Second line.";
        var part = new RecallPart(new RecallAttachment(text, [new string('a', 64)]));
        var fact = new Event(EventKind.RecallAdded, seq: 1, parts: [part]);
        var context = History.Reduce(new Context(), fact);
        Check(JsonCodec.Text(context.Dialogue.Single().Parts) == text,
            "Recall content changed when recorded in Context.");
        foreach (var snapshot in new[] {
            new HistorySnapshot([fact], context, 1),
            new HistorySnapshot([], context, 1) })
        {
            var view = Gui.Project(snapshot, "parent", 1);
            var frame = view.Frames.Single(item => item["type"]?.GetValue<string>() == "part_final");
            Check(frame["part"]?["type"]?.GetValue<string>() == "recall"
                && frame["part"]?["text"]?.GetValue<string>() == text
                && frame["part"]?["role"]?.GetValue<string>() == "human"
                && view.FinalIDs.SetEquals(["parent:e1:p0"]),
                "Recall projection lost content, role or stable identity.");
        }
    }

    public static async Task RunAsync()
    {
        SeparatePaneWindows();
        RecallContent();
        // Errors belong to the chat window, not the short-lived status tail.
        var failed = new HistorySnapshot([
            new Event(EventKind.ErrorOccurred, seq: 1, error: "Safe provider failure"),
            new Event(EventKind.ToolObserved, seq: 2, callID: "background", parts: [new TextPart("output")]),
            new Event(EventKind.ToolObserved, seq: 3, callID: "background", parts: [new TextPart("output")])
        ], new Context(), 3);
        var failureView = Gui.Project(failed, "parent", 1);
        Check(failureView.FinalIDs.Contains("parent:e1:error"),
            "Tool traffic displaced a retained chat error.");
        var errorFrame = failureView.Frames.Single(frame => frame["kind"]?.GetValue<string>() == "error_occurred");
        Check(errorFrame["part_id"]?.GetValue<string>() == "parent:e1:error"
            && errorFrame["message"]?.GetValue<string>() == "Safe provider failure",
            "Error replay lost its stable identity or sanitized message.");
        var later = failed with { Events = [.. failed.Events,
            new Event(EventKind.ErrorOccurred, seq: 4, error: "Later failure")], Sequence = 4 };
        var laterView = Gui.Project(later, "parent", 1);
        Check(!laterView.FinalIDs.Contains("parent:e1:error") && laterView.FinalIDs.Contains("parent:e4:error"),
            "Errors bypassed the bounded chat window.");
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

        // Reverse MCP calls have no assistant card. The recorded name must
        // survive a display cutoff so the browser can recognize a later wait.
        var waitArguments = JsonSerializer.SerializeToElement(new { handle = 32 });
        var external = new HistorySnapshot([
            new Event(EventKind.ToolCalled, seq: 1, callID: "mcp:wait", name: "wait_for_job", args: waitArguments),
            new Event(EventKind.ToolObserved, seq: 2, callID: "mcp:wait", parts: [new TextPart("running")],
                job: new JobInfo(32, "running", "/jobs/32", 91, OriginCallID: "spawn", OriginTool: "spawn_sub_agent"))
        ], new Context(), 2);
        var background = Gui.Project(external, "parent", 1).Frames.Single(frame => frame["type"]!.GetValue<string>() == "tool_finished");
        Check(background["name"]?.GetValue<string>() == "wait_for_job" && background["tool_part_id"] is null,
            "Background wait lost its tool name or invented a model call card.");

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
            for (int i = 0; i < 150; i++)
                child.History().Append(new Event(EventKind.ToolObserved, callID: "background-" + i,
                    parts: [new TextPart("tool output " + i)]));
            string newestTool = $"parent/child:e{child.History().Snapshot().Sequence}:result";
            do { window = await Read(socket, "display_window", stop.Token); }
            while (!window.GetProperty("part_ids").EnumerateArray().Any(id => id.GetString() == newestTool));
            Check(window.GetProperty("part_ids").GetArrayLength() == 200
                && window.GetProperty("part_ids").EnumerateArray().Any(id => id.GetString() == newest),
                "Heavy live tool traffic evicted chat or ignored independent defaults.");
            for (int i = 0; i < 125; i++)
                child.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                    parts: [new TextPart("chat after tools " + i)]));
            newest = $"parent/child:e{child.History().Snapshot().Sequence}:p0";
            do { window = await Read(socket, "display_window", stop.Token); }
            while (!window.GetProperty("part_ids").EnumerateArray().Any(id => id.GetString() == newest));
            Check(window.GetProperty("part_ids").GetArrayLength() == 200
                && window.GetProperty("part_ids").EnumerateArray().Any(id => id.GetString() == newestTool),
                "Heavy live chat traffic evicted tool history.");
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

    static void SeparatePaneWindows()
    {
        var target = new Configuration(Vendor.OpenAI, "offline", "http://127.0.0.1:1").Target();
        var call = new ToolCallPart("call", target, "read_file", JsonSerializer.SerializeToElement(new { path = "fixture" }));
        var context = new Context([
            new Entry(1, Actor.Human, [new TextPart("first")]),
            new Entry(2, Actor.Agent, [new TextPart("mixed response"), call]),
            new Entry(4, Actor.Tool, [new ToolResultPart("call", [new TextPart("result")])]),
            new Entry(5, Actor.Human, [new TextPart("last")])
        ]);
        var snapshot = new HistorySnapshot([], context, 5);
        var projection = Gui.Project(snapshot, "agent", 2, 1);
        Check(projection.FinalIDs.SetEquals(["agent:e2:p0", "agent:e4:result", "agent:e5:p0"]),
            "Tool window removed mixed-response prose or kept an expired call.");
        projection = Gui.Project(snapshot, "agent", 1, 2);
        Check(projection.FinalIDs.SetEquals(["agent:e2:p1", "agent:e4:result", "agent:e5:p0"]),
            "Chat window removed mixed-response tool content or kept expired prose.");
        var options = new GuiOptions { HistoryEventLimit = 7 };
        Check(options.ChatHistoryEventLimit == 7 && options.ToolHistoryEventLimit == 7,
            "Legacy single limit did not configure both independent panes.");
        var notices = Enumerable.Range(6, 1000).Select(sequence =>
            new Event(EventKind.JobKilled, seq: sequence, reason: "fixture",
                job: new JobInfo(sequence, "killed", "/jobs/fixture", 0)));
        projection = Gui.Project(new HistorySnapshot(notices.ToImmutableArray(), context, 1005), "agent", 2, 1);
        Check(projection.FinalIDs.Count == 3 && projection.Frames.Count(frame =>
            frame["type"]!.GetValue<string>() == "lifecycle") == 2,
            "Preserving quiet chat made intervening lifecycle replay unbounded.");
        Console.WriteLine("PASS independent pane limits, mixed response parts and legacy option compatibility");
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
