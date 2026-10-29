using System.Diagnostics;
using System.Net;
using System.Net.Sockets;
using System.Text.Json;
using Ensemble.Gateway;

internal static partial class Program
{
    private static async Task GoRouter(string executable)
    {
        string directory = Path.Combine(Path.GetTempPath(), "gateway-go-client-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(directory);
        Keypair router = Keys.LoadOrCreateKeypair(Path.Combine(directory, "router", "identity.key"));
        Keypair aliceKey = Keys.GenerateKeypair(), bobKey = Keys.GenerateKeypair();
        using TcpListener reservation = new(IPAddress.Loopback, 0);
        reservation.Start();
        int port = ((IPEndPoint)reservation.LocalEndpoint).Port;
        reservation.Stop();
        string configuration = Path.Combine(directory, "gateway.json");
        await File.WriteAllTextAsync(configuration, JsonSerializer.Serialize(new
        {
            gateway = new { listen = "127.0.0.1:" + port, data_dir = Path.Combine(directory, "router") },
            groups = new Dictionary<string, object>
            {
                ["agents"] = new { description = "Fixture peers", capabilities = new[] { "send-message", "declare-contactable" }, trust = "trusted" }
            },
            agents = new Dictionary<string, object>
            {
                ["alice"] = new { public_key = Keys.EncodePublicKey(aliceKey.Public.AsSpan()), groups = new[] { "agents" } },
                ["bob"] = new { public_key = Keys.EncodePublicKey(bobKey.Public.AsSpan()), groups = new[] { "agents" } }
            }
        }));
        ProcessStartInfo start = new(executable)
        {
            UseShellExecute = false, RedirectStandardOutput = true, RedirectStandardError = true,
            WorkingDirectory = directory
        };
        start.ArgumentList.Add("-config");
        start.ArgumentList.Add(configuration);
        using Process process = Process.Start(start) ?? throw new IOException("Could not start Go router fixture.");
        Task<string> stderr = process.StandardError.ReadToEndAsync();
        Task<string>? remainingOutput = null;
        bool cleanExit = false;
        try
        {
            string? printedKey = await process.StandardOutput.ReadLineAsync().WaitAsync(TimeSpan.FromSeconds(5));
            Check(printedKey == Keys.EncodePublicKey(router.Public.AsSpan()), "Go router did not load the shared private-only identity format");
            remainingOutput = process.StandardOutput.ReadToEndAsync();
            // A printed key is not listener readiness. Probe only our chosen
            // loopback endpoint while requiring the new process to remain alive.
            using HttpClient http = new() { Timeout = TimeSpan.FromMilliseconds(200) };
            using CancellationTokenSource ready = new(TimeSpan.FromSeconds(5));
            while (true)
            {
                Check(!process.HasExited, "Go router exited before binding its listener");
                try
                {
                    using HttpResponseMessage health = await http.GetAsync($"http://127.0.0.1:{port}/v1/ws?agent_id=alice", ready.Token);
                    if (health.StatusCode == HttpStatusCode.BadRequest) break;
                }
                catch (HttpRequestException) { }
                catch (OperationCanceledException) when (!ready.IsCancellationRequested) { }
                await Task.Delay(20, ready.Token);
            }
            string url = $"ws://127.0.0.1:{port}/v1/ws";
            RecordingHandler aliceHandler = new();
            await using (Client alice = await Client.ConnectAsync(new(url, "alice", aliceKey, router.Public), aliceHandler))
            {
                JsonElement initialized = await alice.RequestAsync("initialize", new { });
                Check(initialized.GetProperty("serverInfo").GetProperty("name").GetString() == "ensemble-gateway", "initialize wire differs");
                JsonElement tools = await alice.RequestAsync("tools/list", new { });
                Check(tools.GetProperty("tools").EnumerateArray().Any(t => t.GetProperty("name").GetString() == "send-message"), "missing send-message tool");
                await alice.DeclareContactableAsync(["agents"]);
                await using (Client setup = await Client.ConnectAsync(new(url, "bob", bobKey, router.Public), new NopHandler()))
                {
                    await setup.DeclareContactableAsync(["agents"]);
                }
                await alice.SendAsync(new("bob", "offline one", "thread"));
                await alice.SendAsync(new("bob", "offline two", "thread"));
                RecordingHandler bobHandler = new();
                ulong receipt;
                await using (Client bob = await Client.ConnectAsync(new(url, "bob", bobKey, router.Public), bobHandler))
                {
                    Check(bobHandler.Welcome?.QueueDepth == 2, "offline queue depth differs");
                    Envelope first = await bobHandler.Deliveries.Reader.ReadAsync();
                    Envelope second = await bobHandler.Deliveries.Reader.ReadAsync();
                    await File.WriteAllTextAsync(Path.Combine(directory, "deliveries.json"), JsonSerializer.Serialize(new[] { first, second }));
                    const string aliceHeader = "[GATEWAY source=agents user=alice trust=trusted]";
                    Check(first.Seq == 1 && second.Seq == 2 && first.Content == aliceHeader + "\n\noffline one"
                        && second.Content == aliceHeader + "\n\noffline two", "offline queue order/content differs");
                    Check(first.Source.UserID == "alice" && first.Source.Trust == "trusted"
                        && Protocol.ProvenanceHeader(first.Source) == aliceHeader, "authenticated provenance differs");
                    // The reply is read only after the receiver's preceding ack
                    // write and receipt publication, giving an actor-free barrier.
                    _ = await bob.RequestAsync("tools/list", new { });
                    receipt = bob.LastAckedSeq();
                    Check(receipt == 2, "last accepted receipt differs");
                }
                await alice.SendAsync(new("bob", "offline three", "thread"));
                RecordingHandler resumedHandler = new();
                await using (Client resumed = await Client.ConnectAsync(new(url, "bob", bobKey, router.Public, receipt), resumedHandler))
                {
                    Envelope third = await resumedHandler.Deliveries.Reader.ReadAsync();
                    Check(third.Seq == 3 && third.Content == "[GATEWAY source=agents user=alice trust=trusted]\n\noffline three", "resume replayed an accepted item or skipped queued delivery");
                    await resumed.DeclareContactableAsync(["agents"]);
                    await resumed.RespondAsync(third, "answer");
                    Envelope answer = await aliceHandler.Deliveries.Reader.ReadAsync();
                    await File.WriteAllTextAsync(Path.Combine(directory, "reply.json"), JsonSerializer.Serialize(answer));
                    Check(answer.Source.UserID == "bob" && answer.Content == "[GATEWAY source=agents user=bob trust=trusted]\n\nanswer" && answer.ConversationID == "thread"
                        && answer.ResponseMode == "", "response correlation/provenance differs");
                    _ = await resumed.RequestAsync("tools/list", new { });
                    _ = await alice.RequestAsync("tools/list", new { });
                    Check(resumed.LastAckedSeq() == 3 && resumedHandler.Events.Count(e => e == "deliver") == 1, "resume accepted duplicate delivery");
                }
            }
            // SIGTERM joins the reference router's final state save. We signal
            // only the process just created here, never discover or reuse a PID.
            ProcessStartInfo stop = new("/bin/kill") { UseShellExecute = false };
            stop.ArgumentList.Add("-TERM");
            stop.ArgumentList.Add(process.Id.ToString(System.Globalization.CultureInfo.InvariantCulture));
            using Process signal = Process.Start(stop)!;
            await signal.WaitForExitAsync();
            Check(signal.ExitCode == 0, "Could not signal owned Go router");
            await process.WaitForExitAsync().WaitAsync(TimeSpan.FromSeconds(5));
            Check(process.ExitCode == 0, "Go router failed joined shutdown");
            cleanExit = true;
            using JsonDocument state = JsonDocument.Parse(await File.ReadAllTextAsync(Path.Combine(directory, "router", "queues", "bob.json")));
            Check(state.RootElement.GetProperty("ack").GetUInt64() == 3, "Go router did not persist C# acknowledgments");
            await File.WriteAllTextAsync(Path.Combine(directory, "result.json"), "{\"passed\":true,\"exit_code\":0,\"bob_receipt\":3}\n");
        }
        finally
        {
            if (!cleanExit && !process.HasExited)
            {
                process.Kill(entireProcessTree: true);
                await process.WaitForExitAsync();
            }
            await File.WriteAllTextAsync(Path.Combine(directory, "router.stderr"), await stderr);
            if (remainingOutput is not null)
            {
                await File.WriteAllTextAsync(Path.Combine(directory, "router.stdout"), await remainingOutput);
            }
            Console.WriteLine("Go router fixture: " + directory);
        }
    }
}
