using Microsoft.Win32.SafeHandles;
using System.Diagnostics;
using System.Runtime.InteropServices;

namespace Ensemble;

/// <summary>Agent-owned filesystem boundary. An empty launch root retains full
/// access. A confined owner retains a directory handle, not just a path prefix.
/// Streams returned by OpenFile belong to callers, which must dispose them.
/// Process confinement is a separate operation, not implied by Resolve.</summary>
public sealed partial class Sandbox
{
    internal Agent? parent;
    private readonly object gate = new();
    private readonly string root;
    private SafeFileHandle? rootHandle;
    internal bool destroyed;

    public partial Sandbox(Agent parent);
    public partial Agent Parent();
    public partial string Root();
    // Value derivation only, not child creation or registration. Null inherits
    // the current configuration. Refusal reports widening while Effective still
    // enforces the parent restrictions; ignoring Refusal cannot undo narrowing.
    public partial (Configuration Effective, string? Refusal) DeriveChild(Configuration? requested = null);
    // Canonicalization is useful for policy/reporting, but is NOT an open grant.
    // Every file operation repeats resolution and opens relative to rootHandle.
    public partial string Resolve(string path);
    // Confined opens never wait on a FIFO and refuse non-seekable handles.
    // The returned stream belongs to the caller, not to Sandbox.
    public partial FileStream OpenFile(string path, FileMode mode, FileAccess access);
    public partial byte[] ReadFile(string path);
    // Sorted immediate names, not reusable open grants. Consumers must route
    // every subsequent descent/read through this same owner.
    public partial string[] ReadDirectory(string path);
    public partial bool IsDirectory(string path);
    public partial void WriteFile(string path, ReadOnlySpan<byte> content);
    public partial void CreateDirectory(string path);
    public partial void MoveFile(string source, string destination);
    public partial void DeleteFile(string path);
    // Prepare a shell launch without starting it. Job retains process ownership.
    // Confined launches receive an allowlisted environment and inherited OS
    // filesystem/network policy; unsupported platforms refuse before execution.
    internal partial ProcessStartInfo PrepareCommand(string? directory);
    private static partial string ProfileLiteral(string value);
    private static partial string ToolchainTemporaryDirectory();

    public partial void Destroy();

    private static partial string CanonicalPath(string path);
    private partial string RelativePath(string path);
    private partial SafeFileHandle OpenDirectory(string relative, bool create);
    private static partial SafeFileHandle CheckedHandle(int descriptor, string requestedPath);
    private static partial int OpenFlags(bool directory);
    private static partial IOException PathError(string requestedPath);

    // These are platform ABI declarations, not module algorithms. SafeHandle
    // parameters keep descriptors alive across native calls. Bootstrap open
    // never creates a file, so it needs no variadic mode argument.
    [DllImport("libc", EntryPoint = "open", SetLastError = true)]
    private static extern int NativeOpen(string path, int flags);
    private static partial int NativeOpenAt(SafeFileHandle directory, string path, int flags, uint mode);
    [DllImport("libc", EntryPoint = "openat", SetLastError = true)]
    private static extern int NativeOpenAtUnix(SafeFileHandle directory, string path, int flags, uint mode);
    // Apple's ARM64 ABI puts varargs on the stack even when registers remain.
    // Five padding words fill x3-x7 so the mode occupies the first stack slot.
    [DllImport("libc", EntryPoint = "openat", SetLastError = true)]
    private static extern int NativeOpenAtDarwinArm64(SafeFileHandle directory, string path, int flags,
        nint unused3, nint unused4, nint unused5, nint unused6, nint unused7, uint mode);
    [DllImport("libc", EntryPoint = "mkdirat", SetLastError = true)]
    private static extern int NativeMkdirAt(SafeFileHandle directory, string path, uint mode);
    [DllImport("libc", EntryPoint = "renameat", SetLastError = true)]
    private static extern int NativeRenameAt(SafeFileHandle from, string source, SafeFileHandle to, string destination);
    [DllImport("libc", EntryPoint = "unlinkat", SetLastError = true)]
    private static extern int NativeUnlinkAt(SafeFileHandle directory, string path, int flags);
    // fdopendir takes ownership only on success; closedir then closes its fd.
    [DllImport("libc", EntryPoint = "fdopendir", SetLastError = true)]
    private static extern IntPtr NativeFdOpenDirectory(SafeFileHandle descriptor);
    [DllImport("libc", EntryPoint = "closedir", SetLastError = true)]
    private static extern int NativeCloseDirectory(IntPtr directory);
    [DllImport("libc", EntryPoint = "readdir", SetLastError = true)]
    private static extern IntPtr NativeReadDirectoryDarwin(IntPtr directory);
    [DllImport("libc", EntryPoint = "readdir$INODE64", SetLastError = true)]
    private static extern IntPtr NativeReadDirectoryDarwinX64(IntPtr directory);
    [DllImport("libc", EntryPoint = "readdir64", SetLastError = true)]
    private static extern IntPtr NativeReadDirectoryLinux(IntPtr directory);
    [DllImport("libc", EntryPoint = "realpath", SetLastError = true)]
    private static extern IntPtr NativeRealPath(string path, IntPtr buffer);
    [DllImport("libc", EntryPoint = "confstr")]
    private static extern nuint NativeConfstr(int name, [Out] byte[]? buffer, nuint length);
    [DllImport("libc", EntryPoint = "free")]
    private static extern void NativeFree(IntPtr memory);
}
