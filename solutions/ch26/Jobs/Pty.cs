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
        string? directory = args.TryGetProperty("cwd", out var cwd) ? cwd.GetString() : null;
        var start = Parent().Parent().Sandbox().PrepareCommand(directory);
        start.RedirectStandardInput = true;
        start.RedirectStandardOutput = true;
        start.RedirectStandardError = true;
        // script creates the session, controlling terminal and foreground group.
        // Its trusted bootstrap publishes the group before user code or public
        // stdin can run. No writable filesystem entry is a signal target.
        start.ArgumentList.Add("-c");
        start.ArgumentList.Add("printf 'ENSEMBLE_PGID=%s\\n' \"$$\"; stty rows 50 cols 200; exec /bin/sh -c \"$1\"");
        start.ArgumentList.Add("ensemble-pty");
        start.ArgumentList.Add(command);
        start.ArgumentList.Insert(0, start.FileName);
        start.ArgumentList.Insert(0, "/dev/null");
        start.ArgumentList.Insert(0, "-F");
        start.ArgumentList.Insert(0, "-q");
        start.FileName = "/usr/bin/script";
        using var child = new Process { StartInfo = start };
        cancellation.ThrowIfCancellationRequested();
        if (!child.Start())
        {
            throw new IOException("Could not start PTY launcher");
        }
        lock (gate)
        {
            process = child;
        }
        using var registration = cancellation.Register(StopProcess);
        Task stdout = Task.CompletedTask;
        Task stderr = PumpAsync(child.StandardError.BaseStream);
        bool pumpingOutput = false;
        try
        {
            int group = await ReadProcessGroupAsync(child.StandardOutput.BaseStream, cancellation);
            lock (gate)
            {
                processGroup = group;
            }
            stdout = PumpAsync(child.StandardOutput.BaseStream);
            pumpingOutput = true;
            if (cancellation.IsCancellationRequested)
            {
                StopProcess();
            }
            await child.WaitForExitAsync(cancellation);
        }
        catch (OperationCanceledException) when (cancellation.IsCancellationRequested)
        {
            StopProcess();
            await child.WaitForExitAsync();
        }
        catch
        {
            StopProcess();
            await child.WaitForExitAsync();
            throw;
        }
        finally
        {
            try
            {
                if (!pumpingOutput)
                {
                    stdout = PumpAsync(child.StandardOutput.BaseStream);
                }
                await Task.WhenAll(stdout, stderr);
                AppendOutput($"\nexit_code: {child.ExitCode}\n");
            }
            finally
            {
                // Even a failed output write must release the live relationship.
                // Join any cancellation callback before disposing its process.
                await registration.DisposeAsync();
                lock (gate)
                {
                    exitCode = child.ExitCode;
                    processGroup = 0;
                    process = null;
                }
            }
        }
        cancellation.ThrowIfCancellationRequested();
        return "";
    }

    private partial async Task<int> ReadProcessGroupAsync(Stream source, CancellationToken token)
    {
        // Consume exactly the first line: subsequent command output must never
        // be reinterpreted as process control. Failed startup remains bounded.
        var line = new System.Text.StringBuilder();
        var next = new byte[1];
        while (line.Length < 96 && await source.ReadAsync(next, token) != 0)
        {
            if (next[0] == '\n')
            {
                const string prefix = "ENSEMBLE_PGID=";
                string text = line.ToString().TrimEnd('\r');
                if (text.StartsWith(prefix, StringComparison.Ordinal))
                {
                    string number = text[prefix.Length..];
                    if (number.Length > 0 && number.All(char.IsAsciiDigit)
                        && int.TryParse(number, out int group) && group > 1)
                    {
                        return group;
                    }
                }
                break;
            }
            line.Append((char)next[0]);
        }
        AppendOutput(line.ToString() + "\n");
        throw new IOException("PTY did not establish its process group");
    }

    private partial async Task PumpAsync(Stream stream)
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

    private partial void StopProcess()
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
