using System.Collections.Immutable;
using System.Text;
using System.Runtime.InteropServices;
using System.ComponentModel;

namespace Ensemble.Gateway;

public sealed partial class Keypair
{
    public partial Keypair(ReadOnlySpan<byte> privateKey)
    {
        using Noise.KeyPair derived = Noise.KeyPair.Generate(privateKey);
        Private = ImmutableArray.CreateRange(derived.PrivateKey);
        Public = ImmutableArray.CreateRange(derived.PublicKey);
    }
}

public static partial class Keys
{
    public static partial Keypair GenerateKeypair()
    {
        // Generation and derivation both remain in the selected crypto library.
        // The import constructor ensures the same identity invariant for fresh and
        // restored values; the temporary upstream private buffer is then erased.
        using Noise.KeyPair generated = Noise.KeyPair.Generate();
        return new Keypair(generated.PrivateKey);
    }

    public static partial string EncodePublicKey(ReadOnlySpan<byte> publicKey)
    {
        if (publicKey.Length != 32)
        {
            throw new ArgumentException("Public key must contain 32 X25519 bytes.", nameof(publicKey));
        }
        return "x25519:" + Convert.ToBase64String(publicKey);
    }

    public static partial ImmutableArray<byte> ParsePublicKey(string text)
    {
        ArgumentNullException.ThrowIfNull(text);
        text = text.Trim();
        int separator = text.IndexOf(':');
        if (separator >= 0)
        {
            text = text[(separator + 1)..];
        }
        byte[] decoded = Convert.FromBase64String(text);
        if (decoded.Length != 32)
        {
            throw new FormatException("Public key must encode 32 X25519 bytes.");
        }
        return ImmutableArray.CreateRange(decoded);
    }

    private static partial Keypair ReadKeypair(string path)
    {
        // The host selects this file explicitly. Bound decoding without accepting
        // a truncated prefix; ordinary I/O/permission failures are not absence.
        using FileStream input = File.OpenRead(path);
        byte[] data = new byte[1025];
        int size = input.ReadAtLeast(data, data.Length, throwOnEndOfStream: false);
        if (size > 1024)
        {
            throw new InvalidDataException("Private key file is too large.");
        }
        string encoded = new UTF8Encoding(false, true).GetString(data, 0, size).Trim();
        byte[] secret = Convert.FromBase64String(encoded);
        try
        {
            return new Keypair(secret);
        }
        finally
        {
            System.Security.Cryptography.CryptographicOperations.ZeroMemory(secret);
        }
    }

    public static partial Keypair LoadOrCreateKeypair(string path)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(path);
        // State the Windows exclusion explicitly for the Unix file-mode analyzer.
        if (OperatingSystem.IsWindows() || (!OperatingSystem.IsMacOS() && !OperatingSystem.IsLinux()))
        {
            throw new PlatformNotSupportedException("Private-key file ownership/publication is implemented for macOS and Linux.");
        }
        path = Path.GetFullPath(path);
        try
        {
            return ReadKeypair(path);
        }
        catch (FileNotFoundException) { }
        catch (DirectoryNotFoundException) { }

        string directory = Path.GetDirectoryName(path)!;
        // Existing directories are not chmod'd: they may belong to the caller.
        Directory.CreateDirectory(directory, UnixFileMode.UserRead | UnixFileMode.UserWrite | UnixFileMode.UserExecute);
        Keypair candidate = GenerateKeypair();
        string temporary = Path.Combine(directory, ".gateway-key-" + Guid.NewGuid().ToString("N"));
        Exception? originalFailure = null;
        bool ownsTemporary = false;
        try
        {
            FileStreamOptions options = new()
            {
                Mode = FileMode.CreateNew,
                Access = FileAccess.Write,
                Share = FileShare.None,
                UnixCreateMode = UnixFileMode.UserRead | UnixFileMode.UserWrite
            };
            using (FileStream output = new(temporary, options))
            {
                ownsTemporary = true;
                byte[] encoded = Encoding.ASCII.GetBytes(Convert.ToBase64String(candidate.Private.AsSpan()) + "\n");
                output.Write(encoded);
                output.Flush(flushToDisk: true);
            }
            // Publication is exclusive and atomic; readers never see the temporary
            // inode until its complete contents have been flushed. A concurrent
            // winner is loaded, not overwritten. This is not a directory-fsync or
            // power-loss guarantee.
            return PublishKeyFile(temporary, path) ? candidate : ReadKeypair(path);
        }
        catch (Exception error)
        {
            originalFailure = error;
            throw;
        }
        finally
        {
            try
            {
                if (ownsTemporary)
                {
                    File.Delete(temporary);
                }
            }
            catch (Exception) when (originalFailure is not null)
            {
                // Never hide a failed identity operation behind cleanup failure.
                // Any remaining temporary file was created owner-only.
            }
        }
    }

    private static partial bool PublishKeyFile(string temporary, string destination)
    {
        // File.Move(false) on Unix may perform a check followed by a replacing
        // rename. link is the exclusive publication operation we actually need.
        // The temporary file is in the same directory, hence on the same volume;
        // unsupported filesystems refuse instead of weakening atomicity.
        if (NativeLink(temporary, destination) == 0)
        {
            return true;
        }
        int code = Marshal.GetLastPInvokeError();
        if (code == 17) // EEXIST on the two supported Unix platforms.
        {
            return false;
        }
        throw new IOException("Could not publish the private-key file.", new Win32Exception(code));
    }
}
