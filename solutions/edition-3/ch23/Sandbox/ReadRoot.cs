using Microsoft.Win32.SafeHandles;

namespace Ensemble;

public sealed partial class Sandbox
{
    internal static partial SafeFileHandle OpenReadRoot(string path)
    {
        if (!OperatingSystem.IsMacOS() && !OperatingSystem.IsLinux())
        {
            throw new PlatformNotSupportedException("Rooted skill-document reads require macOS or Linux");
        }
        // Resolve the host's selected root once. Descents thereafter use this
        // retained descriptor, not the possibly renamed/replaced pathname.
        string canonical = CanonicalPath(Path.GetFullPath(path));
        using var ancestors = CheckedHandle(NativeOpen("/", OpenFlags(true)), path);
        return OpenDirectoryAt(ancestors, canonical.TrimStart('/'), false);
    }

    internal static partial SafeFileHandle OpenReadDirectory(SafeFileHandle root, string relative)
    {
        if (Path.IsPathRooted(relative))
        {
            throw new IOException("Read-root paths must be relative");
        }
        return OpenDirectoryAt(root, relative, false);
    }

    internal static partial string[] ReadRootNames(SafeFileHandle root)
    {
        // Opening '.' creates an independent enumeration offset; dup would not.
        using var scan = OpenReadDirectory(root, ".");
        return ReadOpenedDirectory(scan, ".");
    }

    internal static partial FileStream OpenReadFile(SafeFileHandle directory, string name)
    {
        if (name.Length == 0 || name is "." or ".." || name.Contains('/'))
        {
            throw new IOException("Read-root files must be single directory entries");
        }
        int flags = OpenFlags(false) | (OperatingSystem.IsMacOS() ? 0x4 : 0x800); // O_NONBLOCK
        var handle = CheckedHandle(NativeOpenAt(directory, name, flags, 0), name);
        try
        {
            var stream = new FileStream(handle, FileAccess.Read);
            if (!stream.CanSeek)
            {
                stream.Dispose();
                throw new IOException("Skill document is not a seekable file");
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
