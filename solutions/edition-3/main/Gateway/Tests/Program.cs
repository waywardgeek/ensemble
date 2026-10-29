using System.Diagnostics;
using System.Runtime.InteropServices;
using System.Security.Cryptography;
using System.Text;
using System.Text.Json;
using Noise;

// Real library tests and a pipe-connected real Go implementation. No provider,
// network port or production actor is involved. Identity files use owned temporary homes.
internal static partial class Program
{
    private const string Suite = "Noise_KK_25519_ChaChaPoly_SHA256";
    private static readonly byte[] Prologue = Encoding.UTF8.GetBytes("Gateway interoperability — edition 3");

    private static int Main(string[] args)
    {
        Console.WriteLine($"Platform: {RuntimeInformation.OSDescription}; {RuntimeInformation.ProcessArchitecture}");
        int failures = 0;
        Run("legacy and associated-data transport", TransportRoundTrips, ref failures);
        Run("authentication rejection preserves receive nonce", AuthenticationFailures, ref failures);
        Run("KK identities and prologue are authenticated", HandshakeFailures, ref failures);
        Run("public key values and import", PublicKeys, ref failures);
        Run("private identity persistence and concurrent creation", KeyFiles, ref failures);
        Run("public session lifetime and terminal authentication failure", PublicSessions, ref failures);
        if (args.Length == 2 && args[0] == "--interop")
        {
            Run("C# initiator with existing Go responder", () => Interoperate(args[1], true), ref failures);
            Run("Go initiator with C# responder", () => Interoperate(args[1], false), ref failures);
        }
        else if (args.Length != 0)
        {
            Console.Error.WriteLine("Usage: NoiseTests [--interop GO_PEER_EXECUTABLE]");
            return 2;
        }
        else
        {
            Console.WriteLine("Go interoperability not requested in this run.");
        }
        Console.WriteLine($"Failures: {failures}");
        return failures == 0 ? 0 : 1;
    }

    private static void Run(string name, Action test, ref int failures)
    {
        try
        {
            test();
            Console.WriteLine($"PASS {name}");
        }
        catch (Exception error)
        {
            failures++;
            Console.WriteLine($"FAIL {name}: {error}");
        }
    }

    private static void Require(bool condition, string reason)
    {
        if (!condition)
        {
            throw new InvalidOperationException(reason);
        }
    }

    private static void Reject<T>(Action operation, string reason) where T : Exception
    {
        try
        {
            operation();
        }
        catch (T)
        {
            return;
        }
        throw new InvalidOperationException($"Accepted {reason}");
    }

    private static (Transport Initiator, Transport Responder) Connect()
    {
        using KeyPair alice = KeyPair.Generate();
        using KeyPair bob = KeyPair.Generate();
        Protocol protocol = Protocol.Parse(Suite);
        using HandshakeState a = protocol.Create(true, Prologue, alice.PrivateKey, bob.PublicKey);
        using HandshakeState b = protocol.Create(false, Prologue, bob.PrivateKey, alice.PublicKey);
        byte[] wire = new byte[Protocol.MaxMessageLength];
        byte[] plain = new byte[Protocol.MaxMessageLength];
        byte[] hello = Encoding.UTF8.GetBytes("first handshake payload");
        var first = a.WriteMessage(hello, wire);
        var received = b.ReadMessage(wire.AsSpan(0, first.BytesWritten), plain);
        Require(first.Transport is null && received.Transport is null, "KK finished early");
        Require(plain.AsSpan(0, received.BytesRead).SequenceEqual(hello), "first payload differs");
        byte[] welcome = Encoding.UTF8.GetBytes("second handshake payload");
        var second = b.WriteMessage(welcome, wire);
        var finished = a.ReadMessage(wire.AsSpan(0, second.BytesWritten), plain);
        Require(plain.AsSpan(0, finished.BytesRead).SequenceEqual(welcome), "second payload differs");
        Require(second.HandshakeHash.AsSpan().SequenceEqual(finished.HandshakeHash), "transcript hash differs");
        return (finished.Transport, second.Transport);
    }

    private static void TransportRoundTrips()
    {
        var pair = Connect();
        using Transport a = pair.Initiator;
        using Transport b = pair.Responder;
        byte[] wire = new byte[Protocol.MaxMessageLength];
        byte[] plain = new byte[Protocol.MaxMessageLength];
        byte[][] messages = [[], Encoding.UTF8.GetBytes("Mac ↔ Pi"), RandomNumberGenerator.GetBytes(65519)];
        byte[][] bindings = [[], Encoding.UTF8.GetBytes("binding"), [0, 255, 128, 0, 7]];
        foreach (byte[] message in messages)
        {
            // Original methods and new methods agree on empty associated data.
            int size = a.WriteMessage(message, wire);
            int count = b.ReadMessage(wire.AsSpan(0, size), plain, []);
            Require(plain.AsSpan(0, count).SequenceEqual(message), "legacy write/new read differs");
            size = b.WriteMessage(message, wire, []);
            count = a.ReadMessage(wire.AsSpan(0, size), plain);
            Require(plain.AsSpan(0, count).SequenceEqual(message), "new write/legacy read differs");
            foreach (byte[] binding in bindings)
            {
                RoundTrip(a, b, message, binding, wire, plain);
                RoundTrip(b, a, message, binding, wire, plain);
            }
        }
        // Argument refusal occurs before nonce mutation; another valid message still works.
        Reject<ArgumentException>(() => a.WriteMessage(new byte[65520], wire, bindings[1]), "oversized message");
        Reject<ArgumentException>(() => a.WriteMessage(new byte[1], new byte[16], bindings[1]), "short output buffer");
        RoundTrip(a, b, messages[1], bindings[1], wire, plain);
        a.Dispose();
        Reject<ObjectDisposedException>(() => a.WriteMessage([], wire, bindings[1]), "disposed transport");
    }

    private static void RoundTrip(Transport sender, Transport receiver, byte[] message,
        byte[] binding, byte[] wire, byte[] plain)
    {
        int size = sender.WriteMessage(message, wire, binding);
        Require(size == message.Length + 16, "associated data changed ciphertext size");
        int count = receiver.ReadMessage(wire.AsSpan(0, size), plain, binding);
        Require(plain.AsSpan(0, count).SequenceEqual(message), "plaintext differs");
    }

    private static void AuthenticationFailures()
    {
        var pair = Connect();
        using Transport a = pair.Initiator;
        using Transport b = pair.Responder;
        byte[] message = Encoding.UTF8.GetBytes("authenticated transport");
        byte[] binding = Encoding.UTF8.GetBytes("binding");
        byte[] wire = new byte[message.Length + 16];
        byte[] plain = new byte[message.Length];
        int size = a.WriteMessage(message, wire, binding);
        Reject<CryptographicException>(() => b.ReadMessage(wire, plain, "changed"u8), "changed associated data");
        Reject<CryptographicException>(() => b.ReadMessage(wire, plain), "omitted associated data");
        byte[] damaged = (byte[])wire.Clone();
        damaged[^1] ^= 1;
        Reject<CryptographicException>(() => b.ReadMessage(damaged, plain, binding), "altered tag");
        Require(b.ReadMessage(wire, plain, binding) == message.Length, "failure consumed nonce");
        Require(plain.AsSpan().SequenceEqual(message), "corrected input differs");
        Reject<CryptographicException>(() => b.ReadMessage(wire, plain, binding), "replayed frame");
        RoundTrip(a, b, message, binding, wire, plain);
        // The other direction has an independent counter and authenticates AD too.
        size = b.WriteMessage(message, wire, binding);
        Reject<CryptographicException>(() => a.ReadMessage(wire.AsSpan(0, size), plain, "changed"u8), "reverse AD mismatch");
        Require(a.ReadMessage(wire.AsSpan(0, size), plain, binding) == message.Length, "reverse receive nonce changed");
    }

    private static void HandshakeFailures()
    {
        using KeyPair alice = KeyPair.Generate();
        using KeyPair bob = KeyPair.Generate();
        using KeyPair stranger = KeyPair.Generate();
        Protocol protocol = Protocol.Parse(Suite);
        for (int scenario = 0; scenario < 3; scenario++)
        {
            using HandshakeState a = protocol.Create(true, Prologue, alice.PrivateKey,
                scenario == 0 ? stranger.PublicKey : bob.PublicKey);
            using HandshakeState b = protocol.Create(false, scenario == 2 ? "wrong application"u8 : Prologue,
                bob.PrivateKey, scenario == 1 ? stranger.PublicKey : alice.PublicKey);
            byte[] wire = new byte[Protocol.MaxMessageLength];
            byte[] plain = new byte[Protocol.MaxMessageLength];
            var first = a.WriteMessage([], wire);
            Reject<CryptographicException>(() => b.ReadMessage(wire.AsSpan(0, first.BytesWritten), plain),
                $"wrong handshake identity/prologue scenario {scenario}");
        }
    }

    private static void Interoperate(string executable, bool initiator)
    {
        using var process = new Process();
        process.StartInfo = new ProcessStartInfo(executable)
        {
            RedirectStandardInput = true,
            RedirectStandardOutput = true,
            RedirectStandardError = true,
            UseShellExecute = false
        };
        Require(process.Start(), "Go peer failed to start");
        Task<string> errors = process.StandardError.ReadToEndAsync();
        try
        {
            var generated = global::Ensemble.Gateway.Keys.GenerateKeypair();
            // Restore from private bytes, exercising the newly exposed existing
            // derivation operation against an independent implementation.
            var local = new global::Ensemble.Gateway.Keypair(generated.Private.AsSpan());
            byte[] peer = Ask(process, "start", role: initiator ? "responder" : "initiator",
                peer: local.Public.ToArray(), prologue: Prologue);
            using var session = new global::Ensemble.Gateway.NoiseKK(
                initiator ? global::Ensemble.Gateway.NoiseRole.Initiator : global::Ensemble.Gateway.NoiseRole.Responder,
                local, peer, Prologue);
            byte[] hello = Encoding.UTF8.GetBytes("independent implementation handshake");
            if (initiator)
            {
                byte[] first = session.WriteHandshakeMessage(hello);
                Require(Ask(process, "read", first).AsSpan().SequenceEqual(hello), "Go handshake read differs");
                byte[] second = Ask(process, "write", hello, expectComplete: true);
                Require(session.ReadHandshakeMessage(second).AsSpan().SequenceEqual(hello), "C# handshake read differs");
            }
            else
            {
                byte[] first = Ask(process, "write", hello);
                Require(session.ReadHandshakeMessage(first).AsSpan().SequenceEqual(hello), "C# initial read differs");
                byte[] second = session.WriteHandshakeMessage(hello);
                Require(Ask(process, "read", second, expectComplete: true).AsSpan().SequenceEqual(hello), "Go final read differs");
            }
            Require(session.IsHandshakeComplete(), "C# wrapper did not publish completed handshake");
            foreach (byte[] binding in new byte[][] { [], "binding"u8.ToArray(), [0, 255, 128, 0] })
            {
                foreach (byte[] message in new byte[][] { [], "C# ↔ Go"u8.ToArray(), RandomNumberGenerator.GetBytes(4096) })
                {
                    byte[] sent = session.Encrypt(message, binding);
                    Ask(process, "decrypt", sent, "not the binding"u8.ToArray(), expectError: true);
                    byte[] received = Ask(process, "decrypt", sent, binding);
                    Require(received.AsSpan().SequenceEqual(message), "Go decrypt differs");
                    byte[] reverse = Ask(process, "encrypt", message, binding);
                    Require(session.Decrypt(reverse, binding).AsSpan().SequenceEqual(message), "C# decrypt differs");
                }
            }
            byte[] last = Ask(process, "encrypt", "last"u8.ToArray(), "binding"u8.ToArray());
            Reject<CryptographicException>(() => session.Decrypt(last, "wrong"u8), "C# public wrapper wrong AD");
            Reject<CryptographicException>(() => session.Decrypt(last, "binding"u8), "failed wrapper must not resume");
            process.StandardInput.Close();
            Require(process.WaitForExit(10000), "Go peer did not exit after EOF");
            Require(process.ExitCode == 0, $"Go peer exit {process.ExitCode}: {errors.GetAwaiter().GetResult()}");
        }
        finally
        {
            if (!process.HasExited)
            {
                process.Kill(entireProcessTree: true);
                process.WaitForExit();
            }
        }
    }

    private static byte[] Ask(Process process, string operation, byte[]? data = null,
        byte[]? associated = null, string? role = null, byte[]? peer = null,
        byte[]? prologue = null, bool expectError = false, bool expectComplete = false)
    {
        string request = JsonSerializer.Serialize(new { operation, data, associated, role, peer, prologue });
        process.StandardInput.WriteLine(request);
        process.StandardInput.Flush();
        string? line = process.StandardOutput.ReadLineAsync().WaitAsync(TimeSpan.FromSeconds(10)).GetAwaiter().GetResult();
        Require(line is not null, "Go peer closed before replying");
        using JsonDocument document = JsonDocument.Parse(line!);
        JsonElement reply = document.RootElement;
        string error = reply.GetProperty("error").GetString()!;
        if (expectError)
        {
            Require(error.Length != 0, "Go accepted incorrect associated data");
            return [];
        }
        Require(error.Length == 0, $"Go refused request: {error}");
        if (expectComplete)
        {
            Require(reply.GetProperty("complete").GetBoolean(), "Go handshake is not complete");
        }
        JsonElement result = reply.GetProperty("data");
        return result.ValueKind == JsonValueKind.Null ? [] : result.GetBytesFromBase64();
    }
}
