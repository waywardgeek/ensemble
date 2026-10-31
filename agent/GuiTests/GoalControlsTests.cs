using System.Net;
using System.Net.Sockets;
using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using Ensemble;
using EnsembleGui;

// External public-API consumer. Real actors, GUI and sockets; no inference.
static class GoalControlsTests
{
    static void Check(bool condition, string message)
    {
        if (!condition) throw new InvalidOperationException(message);
    }

    static Task Send(ClientWebSocket socket, object value, CancellationToken token) =>
        socket.SendAsync(JsonSerializer.SerializeToUtf8Bytes(value), WebSocketMessageType.Text, true, token);

    static async Task<JsonElement> Frame(ClientWebSocket socket, string type, CancellationToken token,
        bool rejectUpdates = false)
    {
        byte[] buffer = new byte[8192];
        while (true)
        {
            using var bytes = new MemoryStream();
            WebSocketReceiveResult result;
            do
            {
                result = await socket.ReceiveAsync(buffer, token);
                Check(result.MessageType == WebSocketMessageType.Text, "Unexpected socket close");
                bytes.Write(buffer, 0, result.Count);
            } while (!result.EndOfMessage);
            using var document = JsonDocument.Parse(bytes.ToArray());
            string? actual = document.RootElement.GetProperty("type").GetString();
            Check(!rejectUpdates || actual != "goal_updated", "Unrelated or refused goal update acknowledged");
            if (actual == type) return document.RootElement.Clone();
        }
    }

    public static async Task RunAsync()
    {
        string directory = Directory.CreateTempSubdirectory("ensemble-goal-gui-").FullName;
        string skills = Path.Combine(directory, "skills");
        Directory.CreateDirectory(Path.Combine(skills, "base"));
        Directory.CreateDirectory(Path.Combine(skills, "goals"));
        File.WriteAllText(Path.Combine(skills, "base", "SKILL.md"),
            "---\nname: base\ntype: primary\ndescription: Goal GUI fixture\ntools: read_file\nloadable-skills: goals\n---\n$TOOLS\n$SKILLS");
        File.Copy(Path.Combine(Directory.GetCurrentDirectory(), "agent/Skills/goals/SKILL.md"),
            Path.Combine(skills, "goals", "SKILL.md"));
        string file = Path.Combine(directory, "GOAL.md");
        const string secret = "GOAL_BODY_NOT_STATUS";
        File.WriteAllText(file, secret);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(40));
        var token = deadline.Token;
        var app = new Ensemble.Ensemble(TextWriter.Null, dataDirectory: directory);
        var config = new Configuration(Vendor.OpenAI, "offline", baseUrl: "http://127.0.0.1:1");
        var initial = new Agent(app, config, name: "initial", skillSpec: new SkillSpec(skills, "base"));
        var peer = new Agent(app, config, name: "peer");
        using var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        int port = ((IPEndPoint)listener.LocalEndpoint).Port;
        listener.Stop();
        string origin = $"http://127.0.0.1:{port}";
        using var stop = new CancellationTokenSource();
        Task serving = Gui.RunAsync(initial, new GuiOptions(port,
            SettingsPath: Path.Combine(directory, "appearance.json"),
            TracePath: Path.Combine(directory, "gui.log")), stop.Token);
        try
        {
            using var http = new HttpClient();
            while (true)
            {
                if (serving.IsCompleted) await serving;
                try
                {
                    if ((await http.GetAsync(origin, token)).IsSuccessStatusCode) break;
                }
                catch (HttpRequestException) { }
                await Task.Delay(20, token);
            }
            Check((await http.GetStringAsync(origin, token)).Contains("goal-form"), "Goal controls absent");
            using var first = new ClientWebSocket();
            using var second = new ClientWebSocket();
            first.Options.SetRequestHeader("Origin", origin);
            second.Options.SetRequestHeader("Origin", origin);
            var endpoint = new Uri($"ws://127.0.0.1:{port}/ws");
            await first.ConnectAsync(endpoint, token);
            await second.ConnectAsync(endpoint, token);
            await Frame(first, "current_settings", token);
            await Frame(second, "current_settings", token);
            await Send(second, new { type = "select_agent", agent_id = "peer" }, token);
            await Frame(second, "current_settings", token);
            await Send(first, new { type = "get_goal", agent_id = "initial" }, token);
            var unloaded = (await Frame(first, "goal_status", token)).GetProperty("goal");
            Check(unloaded.GetProperty("phase").GetString() == "unloaded"
                && !unloaded.GetProperty("available").GetBoolean(), "Unloaded query granted control");
            await Send(first, new { type = "update_goal", agent_id = "initial", goal = new { file } }, token);
            await Frame(first, "error", token, rejectUpdates: true);
            var load = new ToolCallPart("load-goals", new Provenance(Vendor.OpenAI, "offline", Surface.ChatCompletions),
                "load_skill", JsonSerializer.SerializeToElement(new { name = "goals" }));
            Check(!(await initial.DispatchAsync(load)).IsError, "Skill load failed");
            await Send(first, new { type = "update_goal", agent_id = "initial", goal = new { file, enabled = true, idle_seconds = 600 } }, token);
            var applied = await Frame(first, "goal_updated", token);
            var status = await initial.GoalStatusAsync(token);
            Check(status.Enabled && status.Phase == "scheduled" && status.DueUtc.HasValue,
                "Acknowledgment preceded actor configuration");
            Check(!applied.GetRawText().Contains(secret), "Goal text leaked to status");
            var due = status.DueUtc;
            await Send(first, new { type = "get_goal", agent_id = "initial" }, token);
            await Frame(first, "goal_status", token);
            Check((await initial.GoalStatusAsync(token)).DueUtc == due, "Status polling reset deadline");
            await Send(second, new { type = "get_goal", agent_id = "peer" }, token);
            var other = await Frame(second, "goal_status", token, rejectUpdates: true);
            Check(other.GetProperty("goal").GetProperty("file").GetString() == "", "Goal leaked to peer");
            foreach (string patch in new[] { "{\"file\":null}", "{\"enabled\":true,\"enabled\":false}",
                "{\"idle_seconds\":0}", "{\"idle_seconds\":\"1\"}", "{\"extra\":true}", "{\"file\":\"/missing/goal-file\"}" })
            {
                string raw = "{\"type\":\"update_goal\",\"agent_id\":\"initial\",\"goal\":" + patch + "}";
                await first.SendAsync(Encoding.UTF8.GetBytes(raw), WebSocketMessageType.Text, true, token);
                await Frame(first, "error", token, rejectUpdates: true);
                Check(await initial.GoalStatusAsync(token) == status, "Refusal mutated goal or deadline");
            }
            await Send(first, new { type = "update_goal", agent_id = "peer", goal = new { file = "" } }, token);
            await Frame(first, "error", token, rejectUpdates: true);
            Check(await initial.GoalStatusAsync(token) == status, "Mismatched target changed selected goal");
            await initial.PostAsync(new AgentInput(InputKind.Pause));
            Check((await initial.GoalStatusAsync(token)).Phase == "paused", "Pause not reflected");
            // Reconnect observes current actor state; it cannot restart a deadline.
            using var reconnected = new ClientWebSocket();
            reconnected.Options.SetRequestHeader("Origin", origin);
            await reconnected.ConnectAsync(endpoint, token);
            await Frame(reconnected, "current_settings", token);
            await Send(reconnected, new { type = "get_goal", agent_id = "initial" }, token);
            Check((await Frame(reconnected, "goal_status", token)).GetProperty("goal").GetProperty("phase").GetString() == "paused",
                "Reconnect lost goal status");
            File.Delete(file);
            _ = initial.Engine().Render(initial.History().View());
            Check((await initial.GoalStatusAsync(token)).Phase == "file_error", "File error hidden as off");
            await Send(first, new { type = "update_goal", agent_id = "initial", goal = new { enabled = false } }, token);
            await Frame(first, "goal_updated", token);
            Check((await initial.GoalStatusAsync(token)).File == file, "Off cleared goal");
            await Send(first, new { type = "update_goal", agent_id = "initial", goal = new { file = "" } }, token);
            await Frame(first, "goal_updated", token);
            Check((await initial.GoalStatusAsync(token)).Phase == "unconfigured", "Clear failed with missing file");
            Check(!Directory.GetFiles(directory, "settings.json", SearchOption.AllDirectories).Any(), "Session goal persisted as policy");
            Console.WriteLine("PASS Goals GUI controls: availability, selection, refusal, deadline, reconnect, error/off/clear");
            first.Abort();
            second.Abort();
            reconnected.Abort();
        }
        finally
        {
            stop.Cancel();
            try { await serving.WaitAsync(TimeSpan.FromSeconds(10)); }
            finally { app.Destroy(); }
            Console.WriteLine("Goals GUI workspace: " + directory);
        }
    }
}
