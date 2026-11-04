using System.Collections.Immutable;
using System.Diagnostics;
using System.Text.Json;
using System.Threading.Channels;
using Ensemble.Gateway;
using Microsoft.AspNetCore.Hosting.Server;
using Microsoft.AspNetCore.Hosting.Server.Features;

// A separate assembly: no internal access, reflection, or candidate test code.
internal static class Program
{
    private static readonly TimeSpan Limit = TimeSpan.FromSeconds(30);
    private static void Check(bool value, string reason)
    {
        if (!value) throw new InvalidDataException(reason);
    }
    private static JsonElement State(string root)
    {
        using JsonDocument document = JsonDocument.Parse(File.ReadAllText(Path.Combine(root, "gateway-state.json")));
        return document.RootElement.Clone();
    }
    private static JsonElement Queue(string root) => State(root).GetProperty("queues").GetProperty("bob");
    private static Keypair Key(string root, string name) => Keys.LoadOrCreateKeypair(Path.Combine(root, name + ".key"));

    public static async Task<int> Main(string[] args)
    {
        try
        {
            if (args.Length == 2 && args[0] == "--serve") return await Serve(args[1]);
            if (args.Length != 1) throw new ArgumentException("Expected isolated state directory.");
            await Grade(args[0]);
            Console.WriteLine(JsonSerializer.Serialize(new { scope = "gateway-complete-state-v1", passed = true }));
            return 0;
        }
        catch (Exception error)
        {
            Console.Error.WriteLine(error);
            return 1;
        }
    }

    private static async Task<int> Serve(string root)
    {
        await using Router router = new(Router.ReadConfig(Path.Combine(root, "config.json")), Key(root, "gateway"));
        WebApplicationBuilder builder = WebApplication.CreateSlimBuilder();
        builder.Configuration.Sources.Clear();
        builder.Logging.ClearProviders();
        builder.WebHost.ConfigureKestrel(options => options.Listen(System.Net.IPAddress.Loopback, 0));
        await using WebApplication app = builder.Build();
        app.UseWebSockets();
        app.MapGet("/v1/ws", async context =>
        {
            string id = context.Request.Query["agent_id"].ToString();
            if (!router.ContainsIdentity(id) || !context.WebSockets.IsWebSocketRequest)
            {
                context.Response.StatusCode = 403;
                return;
            }
            await router.AcceptAsync(id, await context.WebSockets.AcceptWebSocketAsync(), context.RequestAborted);
        });
        await app.StartAsync();
        string address = app.Services.GetRequiredService<IServer>().Features.Get<IServerAddressesFeature>()!.Addresses.Single();
        Console.WriteLine(address.Replace("http://", "ws://", StringComparison.Ordinal) + "/v1/ws");
        Console.Out.Flush();
        await Task.Delay(Timeout.InfiniteTimeSpan);
        return 0;
    }

    private sealed class Host : IAsyncDisposable
    {
        private readonly Process process;
        private readonly Task<string> errors;
        internal string Url = "";
        private Host(Process process) { this.process = process; errors = process.StandardError.ReadToEndAsync(); }
        internal static async Task<Host> Start(string root)
        {
            ProcessStartInfo start = new(Environment.ProcessPath!) { UseShellExecute = false, RedirectStandardOutput = true, RedirectStandardError = true };
            start.ArgumentList.Add(typeof(Program).Assembly.Location);
            start.ArgumentList.Add("--serve");
            start.ArgumentList.Add(root);
            Host host = new(Process.Start(start) ?? throw new IOException("Host launch failed"));
            try
            {
                host.Url = await host.process.StandardOutput.ReadLineAsync().WaitAsync(Limit) ?? "";
                Check(host.Url.StartsWith("ws://127.0.0.1:", StringComparison.Ordinal), "Host did not become ready");
                return host;
            }
            catch { await host.DisposeAsync(); throw; }
        }
        public async ValueTask DisposeAsync()
        {
            if (!process.HasExited) process.Kill(entireProcessTree: true);
            await process.WaitForExitAsync();
            string diagnostic = await errors;
            process.Dispose();
            if (diagnostic.Length != 0) Console.Error.WriteLine(diagnostic);
        }
    }
    private sealed class Deliveries : NopHandler
    {
        internal readonly Channel<Envelope> Received = Channel.CreateUnbounded<Envelope>();
        public override void OnDeliver(Envelope envelope) => Received.Writer.TryWrite(envelope);
    }
    private static ClientConfig Connection(Host host, string root, string id) => new(host.Url, id, Key(root, id), Key(root, "gateway").Public);

    private static async Task Grade(string root)
    {
        Directory.CreateDirectory(root);
        Keypair aliceKey = Key(root, "alice"), bobKey = Key(root, "bob"), gatewayKey = Key(root, "gateway");
        RouterConfig config = new(new("127.0.0.1:0", root),
            new Dictionary<string, GroupDefinition> { ["agents"] = new(["send-message", "declare-contactable"], "trusted") }.ToImmutableDictionary(),
            new Dictionary<string, IdentityDefinition>
            {
                ["alice"] = new(Keys.EncodePublicKey(aliceKey.Public.AsSpan()), ["agents"]),
                ["bob"] = new(Keys.EncodePublicKey(bobKey.Public.AsSpan()), ["agents"])
            }.ToImmutableDictionary());
        File.WriteAllText(Path.Combine(root, "config.json"), JsonSerializer.Serialize(config));
        string payload = Guid.NewGuid().ToString("N");
        using CancellationTokenSource deadline = new(TimeSpan.FromSeconds(90));
        await using (Host host = await Host.Start(root))
        {
            await using (Client bob = await Client.ConnectAsync(Connection(host, root, "bob"), new NopHandler(), deadline.Token))
                await bob.DeclareContactableAsync(["agents"], deadline.Token);
            await using Client alice = await Client.ConnectAsync(Connection(host, root, "alice"), new NopHandler(), deadline.Token);
            await alice.SendAsync(new("bob", payload), deadline.Token);
            JsonElement queue = Queue(root);
            Check(queue.GetProperty("next").GetUInt64() == 1 && queue.GetProperty("ack").GetUInt64() == 0 &&
                queue.GetProperty("messages")[0].GetProperty("content").GetString()!.EndsWith(payload, StringComparison.Ordinal),
                "Admission RPC completed without the exact durable delivery");
        } // Host is killed, never gracefully stopped or saved.
        await using (Host host = await Host.Start(root))
        {
            Deliveries handler = new();
            await using Client bob = await Client.ConnectAsync(Connection(host, root, "bob"), handler, deadline.Token);
            Envelope delivered = await handler.Received.Reader.ReadAsync(deadline.Token);
            Check(delivered.Seq == 1 && delivered.Content.EndsWith(payload, StringComparison.Ordinal) && delivered.Source.UserID == "alice",
                "Crash replay lost content, sequence or authenticated provenance");
            while (bob.LastAckedSeq() != 1) await Task.Delay(5, deadline.Token);
            _ = await bob.RequestAsync("tools/list", new { }, deadline.Token);
            Check(Queue(root).GetProperty("ack").GetUInt64() == 1 && Queue(root).GetProperty("messages").GetArrayLength() == 0,
                "Receipt barrier completed without durable removal");
        }
        await using (Router reopened = new(config, gatewayKey))
            Check(Queue(root).GetProperty("ack").GetUInt64() == 1, "Abrupt restart lost receipt");
        string primary = Path.Combine(root, "gateway-state.json");
        File.WriteAllText(primary, "{broken");
        bool refused = false;
        try { await using Router invalid = new(config, gatewayKey); }
        catch (JsonException) { refused = true; }
        catch (InvalidDataException) { refused = true; }
        Check(refused && File.ReadAllText(primary) == "{broken", "Corruption silently became empty or repaired state");
        await using (Router recovered = Router.RecoverBackup(config, gatewayKey))
            Check(Queue(root).GetProperty("next").GetUInt64() == 1 && Queue(root).GetProperty("ack").GetUInt64() == 0,
                "Explicit backup recovery did not restore prior complete generation");
        Check(Directory.GetFiles(root, "gateway-state.recovery-*.json").Any(path => File.ReadAllText(path) == "{broken"),
            "Recovery did not preserve damaged primary evidence");
        string beforeFailure = File.ReadAllText(primary);
        await using (Host host = await Host.Start(root))
        {
            await using Client alice = await Client.ConnectAsync(Connection(host, root, "alice"), new NopHandler(), deadline.Token);
            string backup = Path.Combine(root, "gateway-state.backup.json");
            File.Delete(backup);
            Directory.CreateDirectory(backup);
            bool writeRefused = false;
            try { await alice.SendAsync(new("bob", "must-not-be-accepted"), deadline.Token); }
            catch (Exception error) when (error is InvalidOperationException or IOException or System.Net.WebSockets.WebSocketException)
            { writeRefused = true; }
            Check(writeRefused && File.ReadAllText(primary) == beforeFailure,
                "Failed persistence was acknowledged or changed the primary");
            Directory.Delete(backup);
            bool stillClosed = false;
            try { await alice.SendAsync(new("bob", "must-stay-closed"), deadline.Token); }
            catch (Exception error) when (error is InvalidOperationException or IOException or System.Net.WebSockets.WebSocketException)
            { stillClosed = true; }
            Check(stillClosed && File.ReadAllText(primary) == beforeFailure,
                "Removing obstruction reopened an ambiguous writer");
        }
        Check(File.ReadAllText(primary) == beforeFailure, "Teardown rewrote failed-save evidence");

        string legacy = Path.Combine(root, "legacy-import");
        Directory.CreateDirectory(Path.Combine(legacy, "queues"));
        RouterConfig legacyConfig = config with { Gateway = config.Gateway with { DataDirectory = legacy } };
        string legacyQueue = Path.Combine(legacy, "queues", "bob.json");
        const string original = "{\"next\":7,\"ack\":7,\"messages\":null}";
        File.WriteAllText(legacyQueue, original);
        bool migrationRequired = false;
        try { await using Router implicitImport = new(legacyConfig, gatewayKey); }
        catch (InvalidDataException) { migrationRequired = true; }
        Check(migrationRequired, "Legacy state imported without explicit migration");
        await using (Router migrated = Router.MigrateLegacy(legacyConfig, gatewayKey))
            Check(Queue(legacy).GetProperty("next").GetUInt64() == 7 && Queue(legacy).GetProperty("ack").GetUInt64() == 7,
                "Migration lost drained sequence or receipt");
        Check(File.ReadAllText(legacyQueue) == original, "Migration changed legacy original");
        await using Router legacyRestart = new(legacyConfig, gatewayKey);
        Check(Queue(legacy).GetProperty("next").GetUInt64() == 7, "Migrated state cannot restart");
    }
}
