using System.Runtime.InteropServices;

namespace Ensemble;

public sealed partial class Sandbox
{
    public partial string[] ReadDirectory(string path)
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (root.Length == 0)
            {
                return Directory.EnumerateFileSystemEntries(path)
                    .Select(entry => Path.GetFileName(entry)).Order(StringComparer.Ordinal).ToArray();
            }

            // A fresh open has its own enumeration offset. Duplicating the
            // retained root descriptor would share an offset across callers.
            using var handle = OpenDirectory(RelativePath(path), false);
            IntPtr directory = NativeFdOpenDirectory(handle);
            if (directory == IntPtr.Zero)
            {
                throw PathError(path);
            }
            handle.SetHandleAsInvalid(); // ownership transferred to closedir
            try
            {
                var names = new List<string>();
                while (true)
                {
                    // Darwin ARM64 has only the 64-bit inode ABI under the
                    // plain symbol. Darwin x64 needs its explicit suffix.
                    // Linux readdir64 fixes the inode/offset layout as well.
                    IntPtr entry;
                    if (OperatingSystem.IsMacOS())
                    {
                        if (RuntimeInformation.ProcessArchitecture == Architecture.X64)
                        {
                            entry = NativeReadDirectoryDarwinX64(directory);
                        }
                        else
                        {
                            entry = NativeReadDirectoryDarwin(directory);
                        }
                    }
                    else
                    {
                        entry = NativeReadDirectoryLinux(directory);
                    }
                    if (entry == IntPtr.Zero)
                    {
                        // SetLastError clears errno before the native call on
                        // modern .NET: null + zero is EOF, not a failed read.
                        int error = Marshal.GetLastPInvokeError();
                        if (error != 0)
                        {
                            throw PathError(path);
                        }
                        break;
                    }

                    // sys/dirent.h: Darwin dirent64 has d_namlen before d_type;
                    // Linux dirent64 does not. Copy before the next readdir,
                    // which can reuse the native buffer. Metadata is not used
                    // as authority for later opens.
                    int nameOffset = OperatingSystem.IsMacOS() ? 21 : 19;
                    string name = Marshal.PtrToStringUTF8(entry + nameOffset)
                        ?? throw new IOException($"Invalid directory entry: {path}");
                    if (name is not "." and not "..")
                    {
                        names.Add(name);
                    }
                }
                return names.Order(StringComparer.Ordinal).ToArray();
            }
            finally
            {
                NativeCloseDirectory(directory);
            }
        }
    }

    public partial bool IsDirectory(string path)
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (root.Length == 0)
            {
                return Directory.Exists(path);
            }

            string relative = RelativePath(path);
            using var parent = OpenDirectory(Path.GetDirectoryName(relative) ?? "", false);
            int descriptor = NativeOpenAt(parent, Path.GetFileName(relative),
                OpenFlags(true), 0);
            if (descriptor < 0)
            {
                int error = Marshal.GetLastPInvokeError();
                if (error is 2 or 20) // ENOENT, ENOTDIR
                {
                    return false;
                }
                throw PathError(path);
            }
            using var directory = CheckedHandle(descriptor, path);
            return true;
        }
    }
}
