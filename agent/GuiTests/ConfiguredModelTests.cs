using System.Net;
using System.Net.Sockets;
using System.Net.WebSockets;
using System.Text.Json;
using Ensemble;
using EnsembleGui;

static class ConfiguredModelTests
{
    static void Check(bool value, string message)
    {
        if (!value) throw new InvalidOperationException(message);
    }

    static async Task<JsonElement> Frame(ClientWebSocket socket, string type, CancellationToken token,
        string? endpoint = null)
    {
        byte[] buffer = new byte[8192];
        while (true)
        {
            using var bytes = new MemoryStream();
            WebSocketReceiveResult received;
            do
            {
                received = await socket.ReceiveAsync(buffer, token);
                Check(received.MessageType == WebSocketMessageType.Text, "Unexpected socket close");
                bytes.Write(buffer, 0, received.Count);
            } while (!received.EndOfMessage);
            using var document = JsonDocument.Parse(bytes.ToArray());
            var frame = document.RootElement;
            if (frame.GetProperty("type").GetString() != type) continue;
            if (endpoint is not null && (!frame.TryGetProperty("selected_offering", out var offering)
                || offering.ValueKind != JsonValueKind.Object
                || offering.GetProperty("endpoint").GetString() != endpoint)) continue;
            return frame.Clone();
        }
    }

    public static async Task RunAsync(string? fixtureDirectory = null)
    {
        string directory = Directory.CreateTempSubdirectory("ensemble-model-gui-").FullName;
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(fixtureDirectory is null ? 30 : 90));
        var token = deadline.Token;
        var root = new Ensemble.Ensemble(TextWriter.Null, dataDirectory: directory);
        using var stop = new CancellationTokenSource();
        Task? serving = null;
        try
        {
            var auth = new ModelAuthentication(root, "none");
            var vendor = new ModelVendor(root, "local", "Local inference");
            var a = new Model(new ModelEndpoint(vendor, "a", "http://127.0.0.1:1",
                Surface.ChatCompletions, auth), "small", "same-wire", "First", new ModelFeatures(Media.None));
            var b = new Model(new ModelEndpoint(vendor, "b", "http://127.0.0.1:2",
                Surface.ChatCompletions, auth), "small", "same-wire", "Second", new ModelFeatures(Media.None));
            var agent = new Agent(root, new Configuration(Vendor.Anthropic, "offline", "http://127.0.0.1:3", ""), name: "viewer");
            using var reservation = new TcpListener(IPAddress.Loopback, 0);
            reservation.Start();
            int port = ((IPEndPoint)reservation.LocalEndpoint).Port;
            reservation.Stop();
            string origin = $"http://127.0.0.1:{port}";
            serving = Gui.RunAsync(agent, new GuiOptions(port,
                SettingsPath: Path.Combine(directory, "appearance.json"),
                TracePath: Path.Combine(directory, "gui.log")), stop.Token);
            using var http = new HttpClient();
            while (true)
            {
                if (serving.IsCompleted) await serving;
                try { if ((await http.GetAsync(origin, token)).IsSuccessStatusCode) break; }
                catch (HttpRequestException) { }
                await Task.Delay(20, token);
            }
            if (fixtureDirectory is not null)
            {
                Directory.CreateDirectory(fixtureDirectory);
                File.WriteAllText(Path.Combine(fixtureDirectory, "ready.json"),
                    JsonSerializer.Serialize(new { origin }));
                bool removed = false;
                while (!File.Exists(Path.Combine(fixtureDirectory, "stop")))
                {
                    token.ThrowIfCancellationRequested();
                    if (!removed && File.Exists(Path.Combine(fixtureDirectory, "remove-a")))
                    {
                        ModelEndpoint oldEndpoint = a.Parent();
                        a.Destroy();
                        _ = new Model(oldEndpoint, "small", "same-wire", "Replacement first",
                            new ModelFeatures(Media.None));
                        removed = true;
                        File.WriteAllText(Path.Combine(fixtureDirectory, "removed-a"), "done");
                    }
                    await Task.Delay(20, token);
                }
                Check(ReferenceEquals(agent.SelectedModel(), b), "Browser did not recover to the second offering");
                Check(agent.History().Events().All(value => value.Kind != EventKind.RequestSent),
                    "Browser model selection initiated inference");
                Console.WriteLine("PASS configured model browser host: explicit recovery, no inference");
                return;
            }
            using var firstSocket = new ClientWebSocket();
            using var reconnected = new ClientWebSocket();
            ClientWebSocket socket = firstSocket;
            socket.Options.SetRequestHeader("Origin", origin);
            await socket.ConnectAsync(new Uri($"ws://127.0.0.1:{port}/ws"), token);
            await Frame(socket, "current_settings", token);
            async Task Select(Model model)
            {
                string command = new System.Text.Json.Nodes.JsonObject
                {
                    ["type"] = "set_model", ["agent_id"] = "viewer",
                    ["offering"] = JsonCodec.WriteModelIdentity(model.Identity())
                }.ToJsonString();
                await socket.SendAsync(System.Text.Encoding.UTF8.GetBytes(command), WebSocketMessageType.Text, true, token);
                await Frame(socket, "usage", token, model.Parent().ID);
                Check(ReferenceEquals(agent.SelectedModel(), model), "Socket selection did not publish the actual Model");
            }
            await Select(a);
            var frame = Gui.MeterFrame(agent);
            Check(frame["configured_models"]!.AsArray().Count == 2
                && frame["selection_available"]!.GetValue<bool>(), "Qualified choices missing from status");
            await socket.SendAsync(System.Text.Encoding.UTF8.GetBytes(
                "{\"type\":\"set_model\",\"agent_id\":\"viewer\",\"offering\":null}"),
                WebSocketMessageType.Text, true, token);
            await Frame(socket, "error", token);
            Check(ReferenceEquals(agent.SelectedModel(), a), "Malformed identity changed selection");
            a.Destroy();
            frame = Gui.MeterFrame(agent);
            Check(!frame["selection_available"]!.GetValue<bool>()
                && frame["selected_offering"]!["endpoint"]!.GetValue<string>() == "a",
                "Lost selection is not displayable for explicit recovery");
            firstSocket.Abort();
            reconnected.Options.SetRequestHeader("Origin", origin);
            await reconnected.ConnectAsync(new Uri($"ws://127.0.0.1:{port}/ws"), token);
            socket = reconnected;
            var unavailable = await Frame(socket, "usage", token, "a");
            Check(!unavailable.GetProperty("selection_available").GetBoolean(),
                "Reconnect misrepresented unavailable selection");
            await Select(b);
            Check(agent.History().Events().All(value => value.Kind != EventKind.RequestSent),
                "Status and selection unexpectedly initiated inference");
            socket.Abort();
            Console.WriteLine("PASS configured model GUI: identity, refusal, loss and explicit recovery");
        }
        finally
        {
            stop.Cancel();
            try { if (serving is not null) await serving.WaitAsync(TimeSpan.FromSeconds(10)); }
            finally { root.Destroy(); }
        }
    }
}
