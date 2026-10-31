using System.Net;
using System.Net.Sockets;
using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using Ensemble;
using EnsembleGui;

// Public consumer: real GUI, actors, filesystem and WebSockets. No inference.
static class MemorySettingsTests
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
            Check(!rejectUpdates || actual != "settings_updated", "Unrelated tab received settings acknowledgment");
            if (actual == type) return document.RootElement.Clone();
        }
    }

    public static async Task RunAsync()
    {
        string directory = Directory.CreateTempSubdirectory("ensemble-memory-gui-").FullName;
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(30));
        var token = deadline.Token;
        var app = new Ensemble.Ensemble(TextWriter.Null, dataDirectory: directory);
        var config = new Configuration(Vendor.OpenAI, "offline", baseUrl: "http://127.0.0.1:1");
        var initial = new Agent(app, config, name: "initial");
        var peer = new Agent(app, config, name: "peer");
        string soul = Path.Combine(initial.DataDirectory(), "SOUL.md");
        File.WriteAllText(soul, "ORIGINAL_GUI_SOUL");
        string settingsPath = Path.Combine(initial.DataDirectory(), "settings.json");
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
            string html = await http.GetStringAsync(origin, token);
            Check(html.Contains("recall-settings-form", StringComparison.Ordinal)
                && html.Contains("recall-max_recall_bytes", StringComparison.Ordinal), "Recall form missing");
            Check(html.Contains("memory-settings-form", StringComparison.Ordinal), "Memory controls not served");
            using var first = new ClientWebSocket();
            using var second = new ClientWebSocket();
            first.Options.SetRequestHeader("Origin", origin);
            second.Options.SetRequestHeader("Origin", origin);
            await first.ConnectAsync(new Uri($"ws://127.0.0.1:{port}/ws"), token);
            await second.ConnectAsync(new Uri($"ws://127.0.0.1:{port}/ws"), token);
            await Frame(first, "current_settings", token);
            await Frame(second, "current_settings", token);
            await Send(second, new { type = "select_agent", agent_id = "peer" }, token);
            Check((await Frame(second, "current_settings", token)).GetProperty("agent_id").GetString() == "peer",
                "Peer selection not acknowledged");
            await Send(first, new { type = "update_settings", agent_id = "initial",
                settings = new { memory = new { soul = new { disabled = true },
                    conversation = new { budget = 4567 } } } }, token);
            var ack = await Frame(first, "settings_updated", token);
            Check(ack.GetProperty("settings").GetProperty("memory").GetProperty("conversation")
                .GetProperty("budget").GetInt64() == 4567, "Budget acknowledgment incorrect");
            Check(initial.MemorySettings().Soul.Disabled && !initial.History().View().Memories.Any(),
                "Acknowledged before policy retirement");
            using (var saved = JsonDocument.Parse(File.ReadAllText(settingsPath)))
                Check(JsonCodec.ReadMemorySettings(saved.RootElement.GetProperty("memory")).Soul.Disabled,
                    "Acknowledged before durable save");
            Check(!File.Exists(Path.Combine(peer.DataDirectory(), "settings.json")), "Peer policy saved accidentally");
            await Send(second, new { type = "get_settings", agent_id = "peer" }, token);
            var other = await Frame(second, "current_settings", token, rejectUpdates: true);
            Check(!other.GetProperty("settings").GetProperty("memory").GetProperty("soul")
                .GetProperty("disabled").GetBoolean(), "Peer policy changed");
            File.WriteAllText(soul, "FRESH_GUI_SOUL");
            await Send(first, new { type = "update_settings", agent_id = "initial",
                settings = new { memory = new { soul = new { disabled = false } } } }, token);
            await Frame(first, "settings_updated", token);
            Check(initial.History().View().Memories.Single().Text == "FRESH_GUI_SOUL",
                "Acknowledged before fresh scan");
            // Recall uses this same durable Agent document, not appearance state.
            await Send(first, new { type = "update_settings", agent_id = "initial",
                settings = new { recall = new { disabled = true, disable_judge = true,
                    judge_model = "example-model", max_snippets = 7, min_score = 1.25 } } }, token);
            var recallAck = (await Frame(first, "settings_updated", token)).GetProperty("settings");
            Check(recallAck.GetProperty("recall").GetProperty("disabled").GetBoolean()
                && initial.RecallSettings().Disabled && initial.RecallSettings().MaxSnippets == 7,
                "Recall acknowledgment preceded application");
            using (var saved = JsonDocument.Parse(File.ReadAllText(settingsPath)))
            {
                Check(JsonCodec.ReadRecallSettings(saved.RootElement.GetProperty("recall")).JudgeModel == "example-model"
                    && JsonCodec.ReadMemorySettings(saved.RootElement.GetProperty("memory")).Conversation.Budget == 4567,
                    "Recall acknowledgment preceded save or lost memory settings");
            }
            await Send(second, new { type = "get_settings", agent_id = "peer" }, token);
            var peerRecall = await Frame(second, "current_settings", token, rejectUpdates: true);
            Check(!peerRecall.GetProperty("settings").GetProperty("recall").GetProperty("disabled").GetBoolean()
                && !peer.RecallSettings().Disabled, "Recall settings leaked to peer");
            await Send(first, new { type = "update_settings", agent_id = "initial",
                settings = new { memory = new { conversation = new { budget = 4568 } } } }, token);
            await Frame(first, "settings_updated", token);
            using (var saved = JsonDocument.Parse(File.ReadAllText(settingsPath)))
                Check(JsonCodec.ReadRecallSettings(saved.RootElement.GetProperty("recall")).Disabled,
                    "Memory patch erased recall settings");
            string before = File.ReadAllText(settingsPath);
            foreach (string patch in new[] {
                "{\"recall\":{\"max_snippets\":-1}}",
                "{\"recall\":{\"candidates\":1.5}}",
                "{\"recall\":{\"judge_model\":null}}",
                "{\"recall\":{\"unknown\":true}}",
                "{\"recall\":{\"disabled\":true,\"disabled\":false}}",
                "{\"recall\":{\"disabled\":false},\"memory\":{}}" })
            {
                // Raw JSON retains duplicate fields, unlike anonymous objects.
                string raw = "{\"type\":\"update_settings\",\"agent_id\":\"initial\",\"settings\":" + patch + "}";
                await first.SendAsync(Encoding.UTF8.GetBytes(raw), WebSocketMessageType.Text, true, token);
                await Frame(first, "error", token, rejectUpdates: true);
                Check(File.ReadAllText(settingsPath) == before && initial.RecallSettings().Disabled,
                    "Malformed recall patch changed file or live policy");
            }
            await Send(first, new { type = "update_settings", agent_id = "initial",
                settings = new { memory = new { soul = new { unknown = true } } } }, token);
            await Frame(first, "error", token, rejectUpdates: true);
            Check(File.ReadAllText(settingsPath) == before && !initial.MemorySettings().Soul.Disabled,
                "Invalid patch changed policy");
            File.Delete(settingsPath);
            Directory.CreateDirectory(settingsPath);
            await Send(first, new { type = "update_settings", agent_id = "initial",
                settings = new { recall = new { disabled = false } } }, token);
            await Frame(first, "error", token, rejectUpdates: true);
            Check(initial.RecallSettings().Disabled, "Failed recall save published success");
            await Send(first, new { type = "update_settings", agent_id = "initial",
                settings = new { memory = new { soul = new { disabled = true } } } }, token);
            await Frame(first, "error", token, rejectUpdates: true);
            Check(!initial.MemorySettings().Soul.Disabled, "Failed save changed live policy");
            Directory.Delete(settingsPath);
            Check(!initial.History().Events().Any(e => e.Kind == EventKind.RequestSent)
                && !peer.History().Events().Any(e => e.Kind == EventKind.RequestSent), "Settings caused inference");
            Console.WriteLine("PASS GUI memory/recall settings: durable acknowledgment, sibling preservation, fresh scan, selected-Agent isolation and refusal");
        }
        finally
        {
            stop.Cancel();
            try
            {
                await serving.WaitAsync(TimeSpan.FromSeconds(10));
                Check(initial.Configuration().Model == "offline", "GUI destroyed borrowed Agent");
            }
            finally { app.Destroy(); }
            Console.WriteLine("Memory GUI workspace: " + directory);
        }
    }
}
