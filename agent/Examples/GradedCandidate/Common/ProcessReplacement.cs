using System.Diagnostics;
using System.Runtime.InteropServices;

namespace Ensemble.Grading;

// Stateless host operation, not an owned model object. Replaces this launcher
// so the grader retains the real candidate PID, stdio and exit status.
internal static partial class ProcessReplacement
{
    internal static partial void Execute(ProcessStartInfo start);
    private static partial nint Strings(string[] values, List<nint> allocations);
    [DllImport("libc", EntryPoint = "execve", SetLastError = true)]
    private static extern int Exec(string path, nint arguments, nint environment);
}
