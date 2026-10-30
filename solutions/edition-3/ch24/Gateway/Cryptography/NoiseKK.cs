using System.Security.Cryptography;

namespace Ensemble.Gateway;

public sealed partial class NoiseKK
{
    public partial NoiseKK(NoiseRole role, Keypair key, ReadOnlySpan<byte> peerPublic,
        ReadOnlySpan<byte> prologue)
    {
        ArgumentNullException.ThrowIfNull(key);
        if (role is not (NoiseRole.Initiator or NoiseRole.Responder))
        {
            throw new ArgumentOutOfRangeException(nameof(role));
        }
        if (peerPublic.Length != 32)
        {
            throw new ArgumentException("Peer public key must contain 32 X25519 bytes.", nameof(peerPublic));
        }
        Noise.Protocol protocol = Noise.Protocol.Parse(Suite);
        byte[] privateCopy = key.Private.ToArray();
        try
        {
            handshake = protocol.Create(role == NoiseRole.Initiator, prologue, privateCopy, peerPublic.ToArray());
        }
        finally
        {
            CryptographicOperations.ZeroMemory(privateCopy);
        }
    }

    internal partial NoiseKK(Client parent, NoiseRole role, Keypair key,
        ReadOnlySpan<byte> peerPublic, ReadOnlySpan<byte> prologue)
        : this(role, key, peerPublic, prologue)
    {
        try
        {
            parent.AddNoise(this);
        }
        catch
        {
            Dispose();
            throw;
        }
    }

    internal partial NoiseKK(Session parent, NoiseRole role, Keypair key,
        ReadOnlySpan<byte> peerPublic, ReadOnlySpan<byte> prologue)
        : this(role, key, peerPublic, prologue)
    {
        try { parent.AddNoise(this); }
        catch
        {
            Dispose();
            throw;
        }
    }

    private partial void EnsureLive()
    {
        ObjectDisposedException.ThrowIf(disposed, this);
        if (failed)
        {
            throw new CryptographicException("Noise session has failed; create a new connection.");
        }
    }

    private partial void Fail()
    {
        failed = true;
        handshake?.Dispose();
        handshake = null;
        transport?.Dispose();
        transport = null;
    }

    public partial byte[] WriteHandshakeMessage(ReadOnlySpan<byte> payload) => HandshakeMessage(true, payload);
    public partial byte[] ReadHandshakeMessage(ReadOnlySpan<byte> message) => HandshakeMessage(false, message);

    private partial byte[] HandshakeMessage(bool writing, ReadOnlySpan<byte> data)
    {
        lock (gate)
        {
            EnsureLive();
            if (handshake is null)
            {
                throw new InvalidOperationException("Noise handshake is already complete.");
            }
            byte[] output = new byte[Noise.Protocol.MaxMessageLength];
            try
            {
                // Upstream owns message order, hashing, DH and cipher splitting.
                var (size, _, completedTransport) = writing
                    ? handshake.WriteMessage(data, output)
                    : handshake.ReadMessage(data, output);
                if (completedTransport is not null)
                {
                    transport = completedTransport;
                    handshake.Dispose();
                    handshake = null;
                }
                return output.AsSpan(0, size).ToArray();
            }
            catch
            {
                Fail();
                throw;
            }
        }
    }

    public partial bool IsHandshakeComplete()
    {
        lock (gate)
        {
            return !disposed && !failed && transport is not null;
        }
    }

    public partial byte[] Encrypt(ReadOnlySpan<byte> data, ReadOnlySpan<byte> associated)
    {
        lock (gate)
        {
            EnsureLive();
            if (transport is null)
            {
                throw new InvalidOperationException("Noise handshake is incomplete.");
            }
            if (data.Length > Noise.Protocol.MaxMessageLength - 16)
            {
                throw new ArgumentException("Noise message is too large.", nameof(data));
            }
            byte[] output = new byte[data.Length + 16];
            try
            {
                int size = transport.WriteMessage(data, output, associated);
                return output.AsSpan(0, size).ToArray();
            }
            catch (CryptographicException)
            {
                Fail();
                throw;
            }
        }
    }

    public partial byte[] Decrypt(ReadOnlySpan<byte> data, ReadOnlySpan<byte> associated)
    {
        lock (gate)
        {
            EnsureLive();
            if (transport is null)
            {
                throw new InvalidOperationException("Noise handshake is incomplete.");
            }
            try
            {
                // Reject framing errors before allocating an attacker-sized buffer.
                if (data.Length < 16 || data.Length > Noise.Protocol.MaxMessageLength)
                {
                    throw new CryptographicException("Invalid Noise transport message length.");
                }
                byte[] output = new byte[data.Length - 16];
                int size = transport.ReadMessage(data, output, associated);
                return output.AsSpan(0, size).ToArray();
            }
            catch (CryptographicException)
            {
                Fail();
                throw;
            }
        }
    }

    public partial void Dispose()
    {
        lock (gate)
        {
            if (disposed)
            {
                return;
            }
            disposed = true;
            handshake?.Dispose();
            handshake = null;
            transport?.Dispose();
            transport = null;
        }
        parent?.RemoveNoise(this);
        sessionParent?.RemoveNoise(this);
    }
}
