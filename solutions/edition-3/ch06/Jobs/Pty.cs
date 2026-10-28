using System.Diagnostics;
using System.Text.Json;

namespace Ensemble;

public sealed partial class Job
{
    internal partial async Task<string> RunCommandAsync(JsonElement args, CancellationToken cancellation)
    {
        if (!OperatingSystem.IsMacOS())
        {
            throw new PlatformNotSupportedException("This chapter endpoint uses the macOS system PTY launcher");
        }
        var command = args.GetProperty("command").GetString()!;
        var directory = Environment.CurrentDirectory;
        if (args.TryGetProperty("cwd", out var cwd) && !string.IsNullOrEmpty(cwd.GetString()))
        {
            directory = Path.GetFullPath(cwd.GetString()!);
        }
        var groupFile = OutputPath + ".pgid";
        var start = new ProcessStartInfo("/usr/bin/script")
        {
            WorkingDirectory = directory,
            RedirectStandardInput = true,
            RedirectStandardOutput = true,
            RedirectStandardError = true,
            UseShellExecute = false
        };
        // script creates the session, controlling terminal and foreground group.
        // Its pipes are only the master's transport; the user's stdin/stdout are
        // attached to the slave PTY. Keep input open until the process exits.
        foreach (var argument in new[] { "-q", "-F", "/dev/null", "/bin/sh", "-c",
            "stty rows 50 cols 200; printf '%s\\n' \"$$\" > \"$1\"; exec /bin/sh -c \"$2\"",
            "ensemble-pty", groupFile, command })
        {
            start.ArgumentList.Add(argument);
        }
        foreach (var key in start.Environment.Keys.ToArray())
        {
            if (key.EndsWith("_API_KEY", StringComparison.OrdinalIgnoreCase)
                || key is "ANTHROPIC_AUTH_TOKEN" or "GOOGLE_API_KEY" or "GEMINI_API_KEY" or "OPENAI_API_KEY" or "ANTHROPIC_API_KEY")
            {
                start.Environment.Remove(key);
            }
        }
        start.Environment["TERM"] = "xterm-256color";
        var child = new Process { StartInfo = start };
        cancellation.ThrowIfCancellationRequested();
        if (!child.Start())
        {
            child.Dispose();
            throw new IOException("Could not start PTY launcher");
        }
        lock (gate)
        {
            process = child;
        }
        using var registration = cancellation.Register(StopProcess);
        var stdout = Pump(child.StandardOutput.BaseStream);
        var stderr = Pump(child.StandardError.BaseStream);
        try
        {
            while (!child.HasExited && !File.Exists(groupFile))
            {
                await Task.Delay(5, cancellation);
            }
            if (File.Exists(groupFile))
            {
                // Startup publishes the identity before exec. A partially written
                // line is retried; it is never guessed from unrelated processes.
                int group;
                while (!int.TryParse(await File.ReadAllTextAsync(groupFile, cancellation), out group))
                {
                    await Task.Delay(5, cancellation);
                }
                lock (gate)
                {
                    processGroup = group;
                }
            }
            await child.WaitForExitAsync(cancellation);
        }
        catch (OperationCanceledException) when (cancellation.IsCancellationRequested)
        {
            StopProcess();
            await child.WaitForExitAsync();
        }
        finally
        {
            await Task.WhenAll(stdout, stderr);
            var code = child.ExitCode;
            AppendOutput($"\nexit_code: {code}\n");
            // Wait for any cancellation callback before releasing its process.
            await registration.DisposeAsync();
            lock (gate)
            {
                exitCode = code;
                processGroup = 0;
                process = null;
            }
            child.Dispose();
            File.Delete(groupFile);
        }
        cancellation.ThrowIfCancellationRequested();
        return "";
    }

    private async Task Pump(Stream stream)
    {
        const byte carriageReturn = 13;
        const byte lineFeed = 10;
        var buffer = new byte[4096];
        var normalized = new byte[4097];
        bool pendingReturn = false;
        while (true)
        {
            int count = await stream.ReadAsync(buffer);
            int written = 0;
            for (int index = 0; index < count; index++)
            {
                byte value = buffer[index];
                if (pendingReturn)
                {
                    // A CRLF may straddle reads. Normalize at the PTY boundary
                    // before disk publication; generic tool bytes stay exact.
                    if (value != lineFeed)
                    {
                        normalized[written++] = carriageReturn;
                    }
                    pendingReturn = false;
                }
                if (value == carriageReturn)
                {
                    pendingReturn = true;
                }
                else
                {
                    normalized[written++] = value;
                }
            }
            if (count == 0)
            {
                if (pendingReturn)
                {
                    normalized[written++] = carriageReturn;
                }
                AppendBytes(normalized.AsSpan(0, written));
                return;
            }
            AppendBytes(normalized.AsSpan(0, written));
        }
    }

    private void StopProcess()
    {
        Process? child;
        int group;
        lock (gate)
        {
            child = process;
            group = processGroup;
        }
        if (group > 0)
        {
            // SIGKILL makes kill_job explicit and prompt, including descendants
            // that ignore TERM. Never signal group zero (our own group).
            _ = Signal(-group, 9);
        }
        if (child is not null)
        {
            try
            {
                // This also closes script if cancellation preceded group startup.
                if (!child.HasExited)
                {
                    child.Kill(entireProcessTree: true);
                }
            }
            catch (InvalidOperationException) { }
        }
    }
}
