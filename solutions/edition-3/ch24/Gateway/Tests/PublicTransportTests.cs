using Noise;
using System.Security.Cryptography;
using Gateway = global::Ensemble.Gateway;

internal static partial class Program
{
    private static void PublicKeys()
    {
        Gateway.Keypair original = Gateway.Keys.GenerateKeypair();
        byte[] supplied = original.Private.ToArray();
        Gateway.Keypair restored = new(supplied);
        Require(restored.Public.AsSpan().SequenceEqual(original.Public.AsSpan()), "private-key import changed identity");
        supplied[0] ^= 255;
        Require(restored.Private.AsSpan().SequenceEqual(original.Private.AsSpan()), "identity retained mutable caller bytes");
        string encoded = Gateway.Keys.EncodePublicKey(original.Public.AsSpan());
        Require(encoded.StartsWith("x25519:"), "public key label differs");
        foreach (string spelling in new[] { encoded, encoded[7..], "historical-label:" + encoded[7..], "  " + encoded + "\n" })
        {
            Require(Gateway.Keys.ParsePublicKey(spelling).AsSpan().SequenceEqual(original.Public.AsSpan()), "public key parse differs");
        }
        Reject<ArgumentException>(() => new Gateway.Keypair(new byte[31]), "short private key");
        Reject<ArgumentException>(() => new Gateway.Keypair(new byte[33]), "long private key");
        Reject<ArgumentException>(() => Gateway.Keys.EncodePublicKey(new byte[31]), "short public encoding");
        Reject<FormatException>(() => Gateway.Keys.ParsePublicKey("x25519:not-base64"), "malformed public encoding");
        Reject<FormatException>(() => Gateway.Keys.ParsePublicKey(Convert.ToBase64String(new byte[31])), "short public decoding");
    }

    private static void KeyFiles()
    {
        string root = Directory.CreateTempSubdirectory("gateway-key-tests-").FullName;
        bool complete = false;
        try
        {
            string directory = Path.Combine(root, "identity");
            string path = Path.Combine(directory, "key");
            using var start = new ManualResetEventSlim();
            Task<Gateway.Keypair>[] creators = Enumerable.Range(0, 12).Select(_ => Task.Run(() =>
            {
                start.Wait();
                return Gateway.Keys.LoadOrCreateKeypair(path);
            })).ToArray();
            start.Set();
            Gateway.Keypair[] identities = Task.WhenAll(creators).WaitAsync(TimeSpan.FromSeconds(15)).GetAwaiter().GetResult();
            Gateway.Keypair first = identities[0];
            foreach (Gateway.Keypair identity in identities)
            {
                Require(identity.Private.AsSpan().SequenceEqual(first.Private.AsSpan()), "concurrent creators returned different identities");
            }
            string disk = File.ReadAllText(path);
            Require(disk == Convert.ToBase64String(first.Private.AsSpan()) + "\n", "private-only file format changed");
            Require(Directory.GetFiles(directory).Length == 1, "successful creation left a temporary key file");
            Require(Gateway.Keys.LoadOrCreateKeypair(path).Public.AsSpan().SequenceEqual(first.Public.AsSpan()), "reload changed public identity");
            if (!OperatingSystem.IsWindows())
            {
                Require(File.GetUnixFileMode(path) == (UnixFileMode.UserRead | UnixFileMode.UserWrite), "new key file is not owner-only");
                Require(File.GetUnixFileMode(directory) == (UnixFileMode.UserRead | UnixFileMode.UserWrite | UnixFileMode.UserExecute), "new private directory mode differs");
            }
            File.WriteAllText(path, "broken-but-existing");
            Reject<FormatException>(() => Gateway.Keys.LoadOrCreateKeypair(path), "malformed file must not rotate identity");
            Require(File.ReadAllText(path) == "broken-but-existing", "malformed private file was overwritten");
            File.WriteAllText(path, Convert.ToBase64String(new byte[31]));
            Reject<ArgumentException>(() => Gateway.Keys.LoadOrCreateKeypair(path), "wrong private key length");
            string oversized = new('A', 1025);
            File.WriteAllText(path, oversized);
            Reject<InvalidDataException>(() => Gateway.Keys.LoadOrCreateKeypair(path), "oversized key file");
            Require(File.ReadAllText(path) == oversized, "oversized key file was rewritten");
            complete = true;
        }
        finally
        {
            if (complete)
            {
                Directory.Delete(root, recursive: true);
            }
            else
            {
                Console.WriteLine("Failed key-file fixture retained at " + root);
            }
        }
    }

    private static (Gateway.NoiseKK Alice, Gateway.NoiseKK Bob) PublicPair()
    {
        Gateway.Keypair alice = Gateway.Keys.GenerateKeypair();
        Gateway.Keypair bob = Gateway.Keys.GenerateKeypair();
        Gateway.NoiseKK a = new(Gateway.NoiseRole.Initiator, alice, bob.Public.AsSpan(), Prologue);
        Gateway.NoiseKK b = new(Gateway.NoiseRole.Responder, bob, alice.Public.AsSpan(), Prologue);
        try
        {
            Require(!a.IsHandshakeComplete() && !b.IsHandshakeComplete(), "new session already complete");
            Reject<InvalidOperationException>(() => a.Encrypt("early"u8), "transport before handshake");
            byte[] first = a.WriteHandshakeMessage("hello"u8);
            Require(b.ReadHandshakeMessage(first).AsSpan().SequenceEqual("hello"u8), "first handshake payload differs");
            byte[] second = b.WriteHandshakeMessage("reply"u8);
            Require(a.ReadHandshakeMessage(second).AsSpan().SequenceEqual("reply"u8), "second handshake payload differs");
            Require(a.IsHandshakeComplete() && b.IsHandshakeComplete(), "public handshake did not complete");
            return (a, b);
        }
        catch
        {
            a.Dispose();
            b.Dispose();
            throw;
        }
    }

    private static void PublicSessions()
    {
        var (a, b) = PublicPair();
        using (a)
        using (b)
        {
            foreach (byte[] text in new byte[][] { [], "public transport π"u8.ToArray(), new byte[Protocol.MaxMessageLength - 16] })
            {
                Require(b.Decrypt(a.Encrypt(text, "binding"u8), "binding"u8).AsSpan().SequenceEqual(text), "public forward transport differs");
                Require(a.Decrypt(b.Encrypt(text)).AsSpan().SequenceEqual(text), "public reverse transport differs");
            }
            Reject<InvalidOperationException>(() => a.WriteHandshakeMessage([]), "handshake after completion");
            byte[] valid = a.Encrypt("private payload"u8, "binding"u8);
            Reject<CryptographicException>(() => b.Decrypt(valid, "wrong"u8), "wrong associated data");
            Require(!b.IsHandshakeComplete(), "failed session remains usable");
            Reject<CryptographicException>(() => b.Decrypt(valid, "binding"u8), "failed session accepted retry");
            Reject<CryptographicException>(() => b.Encrypt("after failure"u8), "failed session encrypted");
        }
        a.Dispose();
        Require(!a.IsHandshakeComplete(), "disposed session remains complete");
        Reject<ObjectDisposedException>(() => a.Encrypt([]), "disposed public session");

        Gateway.Keypair alice = Gateway.Keys.GenerateKeypair();
        Gateway.Keypair bob = Gateway.Keys.GenerateKeypair();
        using Gateway.NoiseKK initiator = new(Gateway.NoiseRole.Initiator, alice, bob.Public.AsSpan(), "first"u8);
        using Gateway.NoiseKK responder = new(Gateway.NoiseRole.Responder, bob, alice.Public.AsSpan(), "different"u8);
        byte[] mismatched = initiator.WriteHandshakeMessage([]);
        Reject<CryptographicException>(() => responder.ReadHandshakeMessage(mismatched), "public prologue mismatch");
        Reject<CryptographicException>(() => responder.WriteHandshakeMessage([]), "failed handshake resumed");
    }
}
