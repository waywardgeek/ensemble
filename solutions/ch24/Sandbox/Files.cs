using Microsoft.Win32.SafeHandles;
using System.Runtime.InteropServices;

namespace Ensemble;

// A path check alone cannot protect a later open: another process can replace
// a checked component with a symlink. Resolve selects a physical in-root name;
// OpenDirectory walks that name again with no-follow directory opens. A changed
// component is refused, not retried through an ordinary host filesystem API.
public sealed partial class Sandbox
{
    public partial Sandbox(Agent parent)
    {
        ArgumentNullException.ThrowIfNull(parent);
        string requested = parent.Configuration().SandboxRoot;
        root = "";
        try
        {
            if (requested.Length != 0)
            {
                if (!OperatingSystem.IsMacOS() && !OperatingSystem.IsLinux())
                {
                    throw new PlatformNotSupportedException("Filesystem confinement requires macOS or Linux");
                }
                root = CanonicalPath(Path.GetFullPath(requested));
                // Bootstrap from / with the same no-follow component walk. This
                // avoids trusting mutable ancestors of the requested root too.
                rootHandle = CheckedHandle(NativeOpen("/", OpenFlags(true)), requested);
                using var ancestors = rootHandle;
                rootHandle = OpenDirectory(root.TrimStart('/'), false);
            }
            parent.AddSandbox(this);
        }
        catch
        {
            rootHandle?.Dispose();
            rootHandle = null;
            throw;
        }
    }

    public partial Agent Parent() => parent ?? throw new ObjectDisposedException(nameof(Sandbox));
    public partial string Root() => root;

    public partial string Resolve(string path)
    {
        ArgumentException.ThrowIfNullOrEmpty(path);
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (root.Length == 0)
            {
                return Path.GetFullPath(path);
            }
            string absolute = Path.GetFullPath(path, root);
            string canonical;
            try
            {
                canonical = CanonicalPath(absolute);
            }
            catch (IOException)
            {
                // Resolution failures may contain an absolute intermediate
                // name. Only the caller's original path belongs in feedback.
                throw new IOException($"Sandbox path could not be resolved: {path}");
            }
            string relative = Path.GetRelativePath(root, canonical);
            if (relative == ".." || relative.StartsWith("../", StringComparison.Ordinal)
                || Path.IsPathRooted(relative))
            {
                throw new IOException($"Sandbox path is outside the root: {path}");
            }
            return canonical;
        }
    }

    private partial string RelativePath(string path)
    {
        return Path.GetRelativePath(root, Resolve(path));
    }

    private static partial string CanonicalPath(string path)
    {
        // realpath resolves every existing ancestor, including absolute inward
        // symlinks. New files/directories retain the unresolved final suffix.
        // We climb only on ENOENT: permission errors and link loops fail closed.
        var suffix = new Stack<string>();
        string current = path;
        while (true)
        {
            IntPtr memory = NativeRealPath(current, IntPtr.Zero);
            if (memory != IntPtr.Zero)
            {
                string resolved;
                try
                {
                    resolved = Marshal.PtrToStringUTF8(memory)
                        ?? throw new IOException("Cannot decode filesystem path");
                }
                finally
                {
                    NativeFree(memory);
                }
                foreach (string name in suffix)
                {
                    resolved = Path.Combine(resolved, name);
                }
                return resolved;
            }
            if (Marshal.GetLastPInvokeError() != 2)
            {
                throw PathError(path);
            }
            string? ancestor = Path.GetDirectoryName(current);
            if (ancestor is null || ancestor == current)
            {
                throw PathError(path);
            }
            suffix.Push(Path.GetFileName(current));
            current = ancestor;
        }
    }

    private static partial int NativeOpenAt(SafeFileHandle directory, string path, int flags, uint mode)
    {
        // A fixed four-argument P/Invoke silently passes the wrong creation
        // permissions on Apple Silicon. The native mode is a variadic argument,
        // unlike mkdirat's fixed mode argument. Keep this ABI adaptation here.
        // https://developer.apple.com/documentation/apple-silicon/addressing-architectural-differences-in-your-macos-code
        if (OperatingSystem.IsMacOS() && RuntimeInformation.ProcessArchitecture == Architecture.Arm64)
        {
            return NativeOpenAtDarwinArm64(directory, path, flags, 0, 0, 0, 0, 0, mode);
        }
        return NativeOpenAtUnix(directory, path, flags, mode);
    }

    private static partial int OpenFlags(bool directory)
    {
        // O_NOFOLLOW | O_CLOEXEC, plus O_DIRECTORY for each parent. Numeric
        // values differ between Darwin and Linux; no ABI struct layout is used.
        if (OperatingSystem.IsMacOS())
        {
            return 0x100 | 0x1000000 | (directory ? 0x100000 : 0);
        }
        return 0x20000 | 0x80000 | (directory ? 0x10000 : 0);
    }

    private static partial IOException PathError(string requestedPath)
    {
        int error = Marshal.GetLastPInvokeError();
        // Never include a resolved target or native exception text. A hostile
        // symlink must not turn a refusal into a disclosure of its target path.
        string message = $"Sandbox filesystem operation refused: {requestedPath} (errno {error})";
        return error == 2 ? new FileNotFoundException(message) : new IOException(message);
    }

    private static partial SafeFileHandle CheckedHandle(int descriptor, string requestedPath)
    {
        if (descriptor < 0)
        {
            throw PathError(requestedPath);
        }
        return new SafeFileHandle((IntPtr)descriptor, ownsHandle: true);
    }

    private partial SafeFileHandle OpenDirectory(string relative, bool create)
    {
        return OpenDirectoryAt(rootHandle!, relative, create);
    }

    private static partial SafeFileHandle OpenDirectoryAt(SafeFileHandle root, string relative, bool create)
    {
        var current = CheckedHandle(NativeOpenAt(root, ".", OpenFlags(true), 0), relative);
        try
        {
            foreach (string component in relative.Split('/', StringSplitOptions.RemoveEmptyEntries))
            {
                if (component == ".")
                {
                    continue;
                }
                if (component == "..")
                {
                    throw new IOException("Sandbox directory traversal refused");
                }
                if (create && NativeMkdirAt(current, component, 0x1C0) != 0
                    && Marshal.GetLastPInvokeError() != 17) // EEXIST still needs a no-follow open.
                {
                    throw PathError(relative);
                }
                var next = CheckedHandle(NativeOpenAt(current, component, OpenFlags(true), 0), relative);
                current.Dispose();
                current = next;
            }
            return current;
        }
        catch
        {
            current.Dispose();
            throw;
        }
    }

    public partial FileStream OpenFile(string path, FileMode mode, FileAccess access)
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (root.Length == 0)
            {
                return new FileStream(path, mode, access, FileShare.ReadWrite | FileShare.Delete);
            }
            if (!Enum.IsDefined(mode) || !Enum.IsDefined(access)
                || (access == FileAccess.Read && mode != FileMode.Open)
                || (mode == FileMode.Append && access != FileAccess.Write))
            {
                throw new ArgumentException("Unsupported file mode/access combination");
            }
            string relative = RelativePath(path);
            using var directory = OpenDirectory(Path.GetDirectoryName(relative) ?? ".", false);
            int flags = OpenFlags(false) | (access == FileAccess.Read ? 0 : access == FileAccess.Write ? 1 : 2);
            bool mac = OperatingSystem.IsMacOS();
            // A named pipe can live inside the root (the .NET runtime itself
            // creates diagnostic FIFOs). Never block acquiring such a handle
            // while holding the Sandbox gate. Regular-file I/O ignores this
            // flag; non-seekable handles are refused below before publication.
            flags |= mac ? 0x4 : 0x800; // O_NONBLOCK
            if (mode is FileMode.Create or FileMode.CreateNew or FileMode.OpenOrCreate or FileMode.Append)
            {
                flags |= mac ? 0x200 : 0x40; // O_CREAT
            }
            if (mode is FileMode.Create or FileMode.Truncate)
            {
                flags |= mac ? 0x400 : 0x200; // O_TRUNC
            }
            if (mode == FileMode.CreateNew)
            {
                flags |= mac ? 0x800 : 0x80; // O_EXCL
            }
            if (mode == FileMode.Append)
            {
                flags |= mac ? 0x8 : 0x400; // O_APPEND
            }
            var handle = CheckedHandle(NativeOpenAt(directory, Path.GetFileName(relative), flags, 0x180), path);
            try
            {
                var stream = new FileStream(handle, access);
                if (!stream.CanSeek)
                {
                    stream.Dispose();
                    throw new IOException($"Sandbox file is not seekable: {path}");
                }
                if (mode == FileMode.Append)
                {
                    // FileStream's handle constructor starts its managed cursor
                    // at zero. Darwin's positional writes do not use O_APPEND,
                    // so initialize that cursor just as the path constructor does.
                    stream.Seek(0, SeekOrigin.End);
                }
                return stream;
            }
            catch
            {
                handle.Dispose();
                throw;
            }
        }
    }

    public partial FileStream OpenRegularRead(string path)
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (root.Length != 0 || (!OperatingSystem.IsMacOS() && !OperatingSystem.IsLinux()))
            {
                // Rooted OpenFile already uses nonblocking no-follow descriptors.
                return OpenFile(path, FileMode.Open, FileAccess.Read);
            }
            // Recursive readers must not wait for a FIFO writer. This explicit
            // operation leaves ordinary unconfined OpenFile semantics unchanged.
            int flags = OperatingSystem.IsMacOS() ? 0x1000004 : 0x80800; // CLOEXEC | NONBLOCK
            var handle = CheckedHandle(NativeOpen(path, flags), path);
            try
            {
                var input = new FileStream(handle, FileAccess.Read);
                if (!input.CanSeek)
                {
                    input.Dispose();
                    throw new IOException($"Sandbox file is not seekable: {path}");
                }
                return input;
            }
            catch
            {
                handle.Dispose();
                throw;
            }
        }
    }

    public partial byte[] ReadFile(string path)
    {
        using var input = OpenFile(path, FileMode.Open, FileAccess.Read);
        using var output = new MemoryStream();
        input.CopyTo(output);
        return output.ToArray();
    }

    public partial void WriteFile(string path, ReadOnlySpan<byte> content)
    {
        using var output = OpenFile(path, FileMode.Create, FileAccess.Write);
        output.Write(content);
    }

    public partial void CreateDirectory(string path)
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (root.Length == 0)
            {
                Directory.CreateDirectory(path);
                return;
            }
            using var directory = OpenDirectory(RelativePath(path), true);
        }
    }

    public partial void MoveFile(string source, string destination)
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (root.Length == 0)
            {
                File.Move(source, destination, overwrite: true);
                return;
            }
            string from = RelativePath(source);
            string to = RelativePath(destination);
            using var fromDirectory = OpenDirectory(Path.GetDirectoryName(from) ?? ".", false);
            using var toDirectory = OpenDirectory(Path.GetDirectoryName(to) ?? ".", false);
            // renameat operates on directory entries; it does not follow a
            // symlink substituted at either leaf after resolution.
            if (NativeRenameAt(fromDirectory, Path.GetFileName(from), toDirectory, Path.GetFileName(to)) != 0)
            {
                throw PathError(source);
            }
        }
    }

    public partial void DeleteFile(string path)
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (root.Length == 0)
            {
                File.Delete(path);
                return;
            }
            string relative = RelativePath(path);
            using var directory = OpenDirectory(Path.GetDirectoryName(relative) ?? ".", false);
            if (NativeUnlinkAt(directory, Path.GetFileName(relative), 0) != 0
                && Marshal.GetLastPInvokeError() != 2)
            {
                throw PathError(path);
            }
        }
    }

    public partial void Destroy()
    {
        lock (gate)
        {
            if (destroyed)
            {
                return;
            }
            destroyed = true;
            try
            {
                rootHandle?.Dispose();
                rootHandle = null;
            }
            finally
            {
                parent?.RemoveSandbox(this);
            }
        }
    }
}

public sealed partial class Agent
{
    public partial Sandbox Sandbox() => sandbox ?? throw new ObjectDisposedException(nameof(Agent));

    internal partial void AddSandbox(Sandbox child)
    {
        ObjectDisposedException.ThrowIf(destroyed, this);
        if (sandbox is not null || child.parent is not null || child.destroyed)
        {
            throw new InvalidOperationException("Sandbox already has an owner or Agent already has a Sandbox");
        }
        sandbox = child;
        child.parent = this;
    }

    internal partial void RemoveSandbox(Sandbox child)
    {
        if (!ReferenceEquals(sandbox, child) || !ReferenceEquals(child.parent, this))
        {
            throw new InvalidOperationException("Sandbox does not belong to this Agent");
        }
        sandbox = null;
        child.parent = null;
    }
}
