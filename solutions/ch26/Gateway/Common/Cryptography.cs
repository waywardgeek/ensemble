using System.Collections.Immutable;
using System.Runtime.InteropServices;

namespace Ensemble.Gateway;

/// <summary>
/// Value-like immutable X25519 identity, not an owned model node. Public bytes are
/// derived from the private half; callers cannot replace either half independently.
/// Private bytes are sensitive application data, never routing/configuration data.
/// Immutable copies follow ordinary GC lifetime; no memory-erasure promise is made.
/// </summary>
public sealed partial class Keypair
{
    public readonly ImmutableArray<byte> Private;
    public readonly ImmutableArray<byte> Public;

    /// <summary>Copy exactly 32 private bytes and derive the corresponding public key.</summary>
    public partial Keypair(ReadOnlySpan<byte> privateKey);
}

/// <summary>Explicit identity-file operations, independent of any Ensemble actor or store.</summary>
public static partial class Keys
{
    /// <summary>Obtain fresh private randomness from the existing Noise/libsodium implementation.</summary>
    public static partial Keypair GenerateKeypair();
    /// <summary>
    /// Load a host-selected private-key file, or atomically publish a new owner-only
    /// file without replacing a concurrent winner. Existing malformed files refuse.
    /// Implemented for macOS/Linux; other platforms refuse rather than weaken file
    /// privacy. This is not a sandbox, secret-store search, or permission repair.
    /// </summary>
    public static partial Keypair LoadOrCreateKeypair(string path);
    /// <summary>Encode exactly 32 public bytes with the unambiguous x25519 label.</summary>
    public static partial string EncodePublicKey(ReadOnlySpan<byte> publicKey);
    /// <summary>Decode 32 bytes; historical labels are accepted as in the Go client.</summary>
    public static partial ImmutableArray<byte> ParsePublicKey(string text);

    private static partial Keypair ReadKeypair(string path);
    // Return false only when another creator already published the destination.
    private static partial bool PublishKeyFile(string temporary, string destination);
    [DllImport("libc", EntryPoint = "link", SetLastError = true)]
    private static extern int NativeLink(
        [MarshalAs(UnmanagedType.LPUTF8Str)] string existing,
        [MarshalAs(UnmanagedType.LPUTF8Str)] string destination);
}

/// <summary>The two roles of the fixed, mutually authenticated KK exchange.</summary>
public enum NoiseRole { Initiator, Responder }

/// <summary>
/// Cryptographic session: a standalone root, Client child, or router Session child.
/// It owns upstream handshake/transport resources, not its caller's immutable keys
/// or files. Dispose ends that lifetime and detaches its owner. Calls are serialized,
/// and authentication failure is terminal.
/// No rekey operation is exposed. Socket owners must also close failed transports.
/// </summary>
public sealed partial class NoiseKK : IDisposable
{
    public const string Suite = "Noise_KK_25519_ChaChaPoly_SHA256";
    // Optional strong owner; standalone public sessions remain roots.
    internal Client? parent;
    internal Session? sessionParent; // Mutually exclusive with the Client owner.
    private readonly object gate = new();
    private Noise.HandshakeState? handshake;
    private Noise.Transport? transport;
    private bool disposed;
    private bool failed;

    /// <summary>Pin the peer's 32-byte public key and the application prologue before any traffic.</summary>
    public partial NoiseKK(NoiseRole role, Keypair key, ReadOnlySpan<byte> peerPublic,
        ReadOnlySpan<byte> prologue = default);
    /// <summary>Produce the next KK message, including an optional authenticated payload.</summary>
    public partial byte[] WriteHandshakeMessage(ReadOnlySpan<byte> payload);
    /// <summary>Authenticate the next KK message and return its payload; refusal is terminal.</summary>
    public partial byte[] ReadHandshakeMessage(ReadOnlySpan<byte> message);
    /// <summary>Construct with the strong owner and register both relationship ends.</summary>
    internal partial NoiseKK(Client parent, NoiseRole role, Keypair key,
        ReadOnlySpan<byte> peerPublic, ReadOnlySpan<byte> prologue);
    internal partial NoiseKK(Session parent, NoiseRole role, Keypair key,
        ReadOnlySpan<byte> peerPublic, ReadOnlySpan<byte> prologue);
    /// <summary>True only for a completed, live session; false after failure or disposal.</summary>
    public partial bool IsHandshakeComplete();
    /// <summary>Encrypt with authenticated, unencrypted associated data. Keep send order.</summary>
    public partial byte[] Encrypt(ReadOnlySpan<byte> data, ReadOnlySpan<byte> associated = default);
    /// <summary>Authenticate before returning plaintext. A refusal permanently fails this session.</summary>
    public partial byte[] Decrypt(ReadOnlySpan<byte> data, ReadOnlySpan<byte> associated = default);
    /// <summary>Release both possible upstream resources, idempotently; retain caller-owned keys/files.</summary>
    public partial void Dispose();

    private partial byte[] HandshakeMessage(bool writing, ReadOnlySpan<byte> data);
    private partial void EnsureLive();
    private partial void Fail();
}
