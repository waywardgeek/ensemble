using System.Diagnostics;
using System.Text.Json;

namespace Ensemble;

public sealed partial class McpProcessTransport
{
    public partial McpProcessTransport(McpServerSpec specification, string workspace)
    {
        ArgumentNullException.ThrowIfNull(specification);
        if (specification.Transport != "stdio" || string.IsNullOrWhiteSpace(specification.Command))
        {
            throw new ArgumentException("A process transport requires a stdio command", nameof(specification));
        }
        string directory = Path.GetFullPath(workspace);
        var start = new ProcessStartInfo(specification.Command)
        {
            WorkingDirectory = directory,
            UseShellExecute = false,
            RedirectStandardInput = true,
            RedirectStandardOutput = true,
            RedirectStandardError = true,
            CreateNoWindow = true
        };
        foreach (string argument in specification.Arguments.IsDefault ? [] : specification.Arguments)
        {
            start.ArgumentList.Add(argument);
        }
        // An allowlist is safer than guessing every provider's credential name.
        // No model key, endpoint, custom variable, or inherited shell startup
        // hook belongs in a skill server's process environment.
        start.Environment.Clear();
        foreach (string name in new[] { "PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "LC_ALL" })
        {
            string? value = Environment.GetEnvironmentVariable(name);
            if (value is not null)
            {
                start.Environment[name] = value;
            }
        }
        process = new Process { StartInfo = start };
        bool started = false;
        try
        {
            if (!process.Start())
            {
                throw new IOException("MCP server process did not start");
            }
            started = true;
            wire = new McpStreamTransport(process.StandardOutput.BaseStream, process.StandardInput.BaseStream);
            stderrDrain = DrainAsync(process.StandardError.BaseStream);
        }
        catch
        {
            try
            {
                if (started)
                {
                    if (!process.HasExited)
                    {
                        process.Kill(entireProcessTree: true);
                    }
                    process.WaitForExit();
                }
            }
            finally
            {
                process.Dispose();
            }
            throw;
        }
    }

    public partial Task SendAsync(JsonElement message, CancellationToken cancellation)
    {
        return wire.SendAsync(message, cancellation);
    }

    public partial Task<JsonElement?> ReceiveAsync(CancellationToken cancellation)
    {
        return wire.ReceiveAsync(cancellation);
    }

    private static partial async Task DrainAsync(Stream stream)
    {
        // Never merge stderr with protocol stdout or accumulate unbounded text.
        // This transport does not publish a child's potentially sensitive logs.
        byte[] buffer = new byte[4096];
        while (await stream.ReadAsync(buffer) != 0)
        {
        }
    }

    public partial ValueTask DisposeAsync()
    {
        lock (gate)
        {
            destruction ??= DestroyCoreAsync();
            return new ValueTask(destruction);
        }
    }

    private partial async Task DestroyCoreAsync()
    {
        try
        {
            // Killing first releases a peer blocked on stdin or a full stdout
            // pipe. There is no implicit timeout or abandoned background join.
            if (!process.HasExited)
            {
                try
                {
                    process.Kill(entireProcessTree: true);
                }
                catch (InvalidOperationException) when (process.HasExited)
                {
                    // Natural exit raced the kill; the join is still required.
                }
            }
        }
        finally
        {
            try
            {
                await wire.DisposeAsync();
            }
            finally
            {
                try
                {
                    await process.WaitForExitAsync();
                    await stderrDrain;
                }
                finally
                {
                    process.Dispose();
                }
            }
        }
    }
}
