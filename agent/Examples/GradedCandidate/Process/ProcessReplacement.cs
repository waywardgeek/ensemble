using System.ComponentModel;
using System.Diagnostics;
using System.Runtime.InteropServices;

namespace Ensemble.Grading;

internal static partial class ProcessReplacement
{
    internal static partial void Execute(ProcessStartInfo start)
    {
        var allocations = new List<nint>();
        try
        {
            string[] arguments = new[] { start.FileName }.Concat(start.ArgumentList).ToArray();
            string[] environment = start.Environment.Select(pair => pair.Key + "=" + pair.Value).ToArray();
            nint argv = Strings(arguments, allocations);
            nint envp = Strings(environment, allocations);
            Directory.SetCurrentDirectory(start.WorkingDirectory);
            // Success never returns: sandbox-exec then replaces itself with the
            // candidate. No intermediate supervisor can hide its exit status.
            _ = Exec(start.FileName, argv, envp);
            throw new Win32Exception(Marshal.GetLastPInvokeError());
        }
        finally
        {
            foreach (nint allocation in allocations)
            {
                Marshal.FreeCoTaskMem(allocation);
            }
        }
    }

    private static partial nint Strings(string[] values, List<nint> allocations)
    {
        nint array = Marshal.AllocCoTaskMem((values.Length + 1) * IntPtr.Size);
        allocations.Add(array);
        for (int index = 0; index < values.Length; index++)
        {
            nint text = Marshal.StringToCoTaskMemUTF8(values[index]);
            allocations.Add(text);
            Marshal.WriteIntPtr(array, index * IntPtr.Size, text);
        }
        Marshal.WriteIntPtr(array, values.Length * IntPtr.Size, IntPtr.Zero);
        return array;
    }
}
