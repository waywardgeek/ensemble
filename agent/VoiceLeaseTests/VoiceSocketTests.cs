using System.Net;
using System.Net.Sockets;
using System.Net.WebSockets;
using System.Text;
using System.Text.Json.Nodes;
using Ensemble;
using EnsembleGui;

internal static class VoiceSocketTests
{
    internal static async Task RunAsync(PcmVoiceAudio audio)
    {
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(35));
        var token = deadline.Token;
        using var stop = new CancellationTokenSource();
        var root = new Ensemble.Ensemble(TextWriter.Null);
        var handler = new Handler();
        _ = new VoiceTranscriber(root, "synthetic-wire-key", handler, TimeSpan.FromSeconds(15));
        var target = new Agent(root, new Configuration(Vendor.Anthropic, "offline", apiKey: "", baseUrl: "http://127.0.0.1:1"), outputDirectory: "socket-first-jobs", name: "socket-first");
        _ = new Agent(root, new Configuration(Vendor.Anthropic, "offline", apiKey: "", baseUrl: "http://127.0.0.1:1"), outputDirectory: "socket-other-jobs", name: "socket-other");
        var listener = new TcpListener(IPAddress.Loopback, 0);
        listener.Start();
        int port = ((IPEndPoint)listener.LocalEndpoint).Port;
        listener.Stop();
        string origin = $"http://127.0.0.1:{port}";
        string trace = Path.GetFullPath("voice-socket-trace.log");
        Task? hosting = null;
        TaskCompletionSource? release = null;
        using var socket = new ClientWebSocket();
        try
        {
            await target.PostAsync(new AgentInput(InputKind.Pause)).WaitAsync(token);
            hosting = Gui.RunAsync(target, new GuiOptions(port, SettingsPath: "voice-socket-settings.json", TracePath: trace), stop.Token);
            using var http = new HttpClient();
            while (true)
            {
                token.ThrowIfCancellationRequested();
                if (hosting.IsCompleted) await hosting;
                try { using var reply = await http.GetAsync(origin + "/", token); if (reply.IsSuccessStatusCode) break; }
                catch (HttpRequestException) { }
                await Task.Delay(20, token);
            }
            socket.Options.SetRequestHeader("Origin", origin);
            await socket.ConnectAsync(new Uri($"ws://127.0.0.1:{port}/ws"), token);
            var seen = new List<JsonObject>();
            var replay = await Frame(socket, seen, "replay", token);
            string address = replay["agent_id"]!.GetValue<string>();
            JsonObject Command(string type, long seq) => new() { ["type"] = type, ["agent_id"] = address, ["sequence"] = seq };
            JsonObject Audio(long seq)
            {
                var command = Command("voice_audio", seq);
                command["audio"] = new JsonObject { ["mime"] = PcmVoiceAudio.MimeType, ["data"] = Convert.ToBase64String(audio.WaveBytes.AsSpan()) };
                return command;
            }
            async Task Expect(JsonObject command, string type)
            {
                await Send(socket, command, token);
                var frame = await Frame(socket, seen, type, token);
                string sequence = type == "error" ? "voice_sequence" : "sequence";
                if (command["sequence"] is not null)
                    Check(frame[sequence]?.GetValue<long>() == command["sequence"]!.GetValue<long>(), "Misdirected voice acknowledgment");
            }
            await Expect(Command("voice_begin", 1), "voice_begun");
            await Expect(Command("voice_commit", 1), "error");
            await Expect(Audio(1), "voice_ready");
            Check(!seen.Last(f => f["type"]?.GetValue<string>() == "voice_ready").ToJsonString().Contains("synthetic lease transcript"), "Ready frame leaked transcript");
            await Expect(Command("voice_commit", 2), "error");
            var wrong = Command("voice_commit", 1); wrong["agent_id"] = "socket-other";
            await Expect(wrong, "error");
            await Expect(Command("voice_commit", 1), "voice_accepted");
            await Expect(Command("voice_commit", 1), "error");
            await Expect(Audio(1), "error");
            await Expect(Command("voice_begin", 1), "error");
            Check(handler.Calls == 1, "Duplicate audio caused a second provider request");
            Console.WriteLine("PASS WebSocket readiness, addressed admission, duplicate and monotonic refusal");

            using (var second = new ClientWebSocket())
            {
                second.Options.SetRequestHeader("Origin", origin);
                await second.ConnectAsync(new Uri($"ws://127.0.0.1:{port}/ws"), token);
                await Frame(second, new(), "replay", token);
                await Send(second, Command("voice_commit", 1), token);
                await Frame(second, new(), "error", token);
                second.Abort();
            }
            await Expect(Command("voice_begin", 2), "voice_begun");
            var malformed = Audio(2); malformed["audio"]!["data"] = "AAAA";
            await Expect(malformed, "error");
            var wrongMime = Audio(2); wrongMime["audio"]!["mime"] = "audio/webm";
            await Expect(wrongMime, "error");
            var oversized = Audio(2);
            oversized["audio"]!["data"] = new string('A', ((PcmVoiceAudio.MaximumWaveBytes + 2) / 3) * 4 + 4);
            await Expect(oversized, "error");
            var smuggled = Audio(2); smuggled["type"] = "prompt"; smuggled["text"] = "must not admit audio";
            await Send(socket, smuggled, token);
            await Frame(socket, seen, "error", token);
            Check(handler.Calls == 1, "Rejected audio reached provider");
            await Expect(Command("voice_cancel", 2), "voice_cancelled");
            await Expect(Audio(2), "error");
            Console.WriteLine("PASS cross-socket, malformed audio and cancelled upload refusals");

            var entered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            var heldRelease = release;
            handler.Respond = async _ => { entered.TrySetResult(); await heldRelease.Task; return Response(); };
            await Expect(Command("voice_begin", 3), "voice_begun");
            await Send(socket, Audio(3), token);
            await entered.Task.WaitAsync(token);
            await Expect(Command("voice_cancel", 3), "voice_cancelled");
            await Expect(Command("voice_commit", 3), "error");
            await Expect(Command("voice_begin", 4), "error");
            release.TrySetResult();
            // Probe readiness with fresh sequence numbers, not retransmitted audio.
            long next = 5;
            while (true)
            {
                await Send(socket, Command("voice_begin", next), token);
                var result = await Frame(socket, seen, new[] { "voice_begun", "error" }, token);
                if (result["type"]!.GetValue<string>() == "voice_begun") break;
                next++;
                await Task.Delay(10, token);
            }
            Check(!seen.Any(f => f["type"]?.GetValue<string>() == "voice_ready" && f["sequence"]?.GetValue<long>() == 3), "Cancelled request published late ready");
            handler.Respond = _ => Task.FromResult(Response());
            await Expect(Audio(next), "voice_ready");
            await Send(socket, new JsonObject { ["type"] = "select_agent", ["agent_id"] = "socket-other" }, token);
            await Frame(socket, seen, "replay", token);
            await Send(socket, new JsonObject { ["type"] = "select_agent", ["agent_id"] = address }, token);
            await Frame(socket, seen, "replay", token);
            await Expect(Command("voice_commit", next), "error");
            Console.WriteLine("PASS responsive cancellation, late result suppression and away/back selection invalidation");

            entered = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            var cancelled = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            release = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            heldRelease = release;
            handler.Respond = async cancellation =>
            {
                using var registration = cancellation.Register(() => cancelled.TrySetResult());
                entered.TrySetResult();
                await heldRelease.Task;
                return Response();
            };
            await Expect(Command("voice_begin", next + 1), "voice_begun");
            await Send(socket, Audio(next + 1), token);
            await entered.Task.WaitAsync(token);
            socket.Abort();
            await cancelled.Task.WaitAsync(token);
            stop.Cancel();
            Check(!hosting.IsCompleted, "Host failed to join held transcription");
            release.TrySetResult();
            await hosting.WaitAsync(token);
            string log = await File.ReadAllTextAsync(trace, token);
            Check(!log.Contains(Convert.ToBase64String(audio.WaveBytes.AsSpan()), StringComparison.Ordinal), "Trace leaked WAV bytes");
            Console.WriteLine("PASS disconnect cancellation, joined host teardown and media-free trace");
        }
        finally
        {
            release?.TrySetResult();
            socket.Abort();
            stop.Cancel();
            try { if (hosting is not null) await hosting.WaitAsync(TimeSpan.FromSeconds(10)); }
            finally { root.Destroy(); }
        }
        Check(handler.Disposed, "Socket fixture leaked HTTP transport");
    }

    private static HttpResponseMessage Response() => new(HttpStatusCode.OK)
    { Content = new StringContent("{\"text\":\"synthetic lease transcript\"}") };
    private static void Check(bool value, string message) { if (!value) throw new Exception(message); }
    private static Task Send(ClientWebSocket socket, JsonObject command, CancellationToken token) =>
        socket.SendAsync(Encoding.UTF8.GetBytes(command.ToJsonString()), WebSocketMessageType.Text, true, token);
    private static Task<JsonObject> Frame(ClientWebSocket socket, List<JsonObject> seen, string type, CancellationToken token) =>
        Frame(socket, seen, new[] { type }, token);
    private static async Task<JsonObject> Frame(ClientWebSocket socket, List<JsonObject> seen, string[] types, CancellationToken token)
    {
        while (true)
        {
            using var bytes = new MemoryStream();
            var buffer = new byte[8192];
            WebSocketReceiveResult part;
            do
            {
                part = await socket.ReceiveAsync(buffer, token);
                if (part.MessageType == WebSocketMessageType.Close) throw new IOException("Socket closed before expected frame");
                bytes.Write(buffer, 0, part.Count);
                if (bytes.Length > 2 * 1024 * 1024) throw new IOException("Fixture frame overflow");
            } while (!part.EndOfMessage);
            var frame = JsonNode.Parse(bytes.ToArray())!.AsObject();
            seen.Add(frame);
            if (types.Contains(frame["type"]?.GetValue<string>())) return frame;
            if (frame["type"]?.GetValue<string>() == "error")
                throw new Exception("Unexpected server error: " + frame.ToJsonString());
        }
    }
}
