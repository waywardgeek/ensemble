using System.Text;
using System.Text.RegularExpressions;

namespace Ensemble;

public sealed partial class Job
{
    internal partial Job(Jobs parent, int handle, ToolCallPart call)
    {
        Handle = handle;
        this.call = call;
        OutputPath = Path.Combine(parent.OutputDirectory, handle.ToString(System.Globalization.CultureInfo.InvariantCulture));
        output = new FileStream(OutputPath, FileMode.Create, FileAccess.Write, FileShare.Read);
        try
        {
            parent.AddJob(this);
        }
        catch
        {
            output.Dispose();
            cancellation.Dispose();
            throw;
        }
    }

    public partial Jobs Parent() => parent ?? throw new ObjectDisposedException(nameof(Job));

    public partial JobInfo Snapshot()
    {
        lock (gate)
        {
            string cwd = "";
            if (call.Name == "run_command" && call.Args.ValueKind == System.Text.Json.JsonValueKind.Object
                && call.Args.TryGetProperty("cwd", out var value) && value.ValueKind == System.Text.Json.JsonValueKind.String)
            {
                cwd = value.GetString()!;
                if (cwd.Length > 0)
                {
                    try
                    {
                        cwd = Path.GetFullPath(cwd);
                    }
                    catch (Exception ex) when (ex is ArgumentException or IOException)
                    {
                        // Keep malformed input visible in the admitted call.
                        // Execution validation returns its paired error.
                    }
                }
            }
            return new JobInfo(Handle, StatusName(status), OutputPath, bytes,
                ExitCode: exitCode, Cwd: cwd);
        }
    }

    private static string StatusName(JobStatus value) => value.ToString().ToLowerInvariant();

    // All notification replacements occur under gate. Observers capture the
    // current task with their state snapshot, so a wake cannot be lost between
    // checking the file length and beginning to wait.
    private void Notify()
    {
        var previous = changed;
        changed = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        previous.TrySetResult();
    }

    internal partial void Start(Tool? tool, ToolCallPart sanitized, Exception? admissionError)
    {
        lock (gate)
        {
            if (worker is not null || destroyed)
            {
                throw new InvalidOperationException("Job already started or destroyed");
            }
            commandJob = tool?.Name() == "run_command";
            worker = Task.Run(async () =>
            {
                try
                {
                    if (admissionError is not null)
                    {
                        throw admissionError;
                    }
                    cancellation.Token.ThrowIfCancellationRequested();
                    if (tool is null)
                    {
                        throw new ArgumentException("Unknown tool: " + call.Name);
                    }
                    var result = await tool.ExecuteJobAsync(this, sanitized, cancellation.Token);
                    AppendOutput(JsonCodec.Text(result.Parts));
                    lock (gate)
                    {
                        if (status == JobStatus.Running)
                        {
                            isError = result.IsError;
                        }
                    }
                }
                catch (OperationCanceledException) when (cancellation.IsCancellationRequested)
                {
                    // KillAsync records the reason. A cancellation exception
                    // does not supply a new lifecycle outcome.
                }
                catch (Exception ex) when (ex is ArgumentException or IOException
                    or InvalidDataException or UnauthorizedAccessException
                    or System.Text.Json.JsonException or System.ComponentModel.Win32Exception)
                {
                    AppendOutput("Tool failed: " + ex.Message);
                    lock (gate)
                    {
                        if (status == JobStatus.Running)
                        {
                            isError = true;
                        }
                    }
                }
                finally
                {
                    lock (gate)
                    {
                        if (status == JobStatus.Running)
                        {
                            status = JobStatus.Done;
                            output.Flush();
                            Notify();
                        }
                        // Readers reopen the durable file; a settled job does
                        // not need to retain an OS writer handle.
                        output.Dispose();
                    }
                    lock (gate)
                    {
                        if (!destroyed)
                        {
                            Parent().Parent().Parent().Log($"job {Handle} {StatusName(status)} ({bytes} bytes)");
                        }
                    }
                }
            });
        }
    }

    public partial void AppendOutput(string text)
    {
        if (text.Length != 0)
        {
            AppendBytes(Encoding.UTF8.GetBytes(text));
        }
    }

    internal partial void AppendBytes(ReadOnlySpan<byte> data)
    {
        lock (gate)
        {
            // A late managed callback may finish after cancellation. It cannot
            // resurrect the job or reopen the retained output artifact.
            if (destroyed || (status != JobStatus.Running && !(commandJob && status == JobStatus.Killed)))
            {
                return;
            }
            output.Write(data);
            output.Flush();
            bytes += data.Length;
            Notify();
        }
    }

    public partial async Task KillAsync(string reason)
    {
        Task? running;
        bool drain;
        bool stop = false;
        Exception? recordFailure = null;
        lock (gate)
        {
            running = worker;
            // Completed work is joined too: a faulted task must never disappear
            // merely because its status publication preceded its exception.
            drain = commandJob || status == JobStatus.Done || running?.IsCompleted == true;
            if (status == JobStatus.Running)
            {
                stop = true;
                if (worker is not null)
                {
                    var notice = commandJob ? "Job killed: " + reason + "\n"
                        : "Job killed: " + reason + "; .NET cannot forcibly stop an uncooperative managed handler. Cancellation requested; late output is ignored.\n";
                    AppendOutput(notice);
                    var info = new JobInfo(Handle, "killed", OutputPath, bytes, ExitCode: exitCode);
                    try
                    {
                        Parent().Parent().History().Append(new Event(EventKind.JobKilled, reason: reason, job: info));
                    }
                    catch (Exception error)
                    {
                        recordFailure = error;
                    }
                }
                status = JobStatus.Killed;
                isError = false;
                if (!commandJob)
                {
                    // Late managed output is ignored. A killed process keeps
                    // its writer only until the PTY reader has drained.
                    output.Dispose();
                }
                Notify();
            }
        }
        try
        {
            if (stop)
            {
                cancellation.Cancel();
            }
            if (drain && running is not null)
            {
                await running;
            }
        }
        finally
        {
            if (recordFailure is not null)
            {
                throw new IOException("Job stopped, but its kill event could not be recorded", recordFailure);
            }
        }
    }

    public partial async Task<ToolResultPart> ObserveAsync(string callID, double delay, string pattern,
        int maxBytes, CancellationToken cancellation, bool wakeOnOutput)
    {
        if (!double.IsFinite(delay) || delay < 0 || maxBytes <= 0)
        {
            throw new ArgumentException("Invalid observation options");
        }
        var regex = pattern == "" ? null : new Regex(pattern, RegexOptions.None, TimeSpan.FromSeconds(1));
        await observation.WaitAsync(cancellation);
        try
        {
            return await ObserveLocked(callID, delay, regex, maxBytes, cancellation, wakeOnOutput);
        }
        finally
        {
            observation.Release();
        }
    }

    private async Task<ToolResultPart> ObserveLocked(string callID, double delay, Regex? pattern,
        int maxBytes, CancellationToken token, bool wakeOnOutput)
    {
        var expires = Task.Delay(TimeSpan.FromSeconds(delay), token);
        while (true)
        {
            Task signal;
            Task? finished;
            JobInfo info;
            byte[] fresh;
            bool error;
            lock (gate)
            {
                signal = changed.Task;
                info = Snapshot();
                finished = status == JobStatus.Done ? worker : null;
                error = isError;
                using var reader = new FileStream(OutputPath, FileMode.Open, FileAccess.Read, FileShare.ReadWrite);
                reader.Position = cursor;
                using var buffer = new MemoryStream();
                reader.CopyTo(buffer);
                fresh = buffer.ToArray();
            }
            if (finished is not null)
            {
                // Await outside gate: the worker's final diagnostic may still
                // be completing. Unexpected exceptions belong to the host.
                await finished.WaitAsync(token);
            }
            var text = Encoding.UTF8.GetString(fresh);
            bool ready = info.Status != "running" || expires.IsCompleted
                || (pattern is not null ? pattern.IsMatch(text) : wakeOnOutput && fresh.Length > 0);
            if (ready)
            {
                token.ThrowIfCancellationRequested();
                bool truncated = fresh.Length > maxBytes;
                if (truncated)
                {
                    // Keep the most useful ends; the full bytes stay on disk.
                    int head = maxBytes / 2;
                    int tail = maxBytes - head;
                    while (head > 0 && head < fresh.Length && (fresh[head] & 0xc0) == 0x80)
                    {
                        head--;
                    }
                    int tailStart = fresh.Length - tail;
                    while (tailStart < fresh.Length && (fresh[tailStart] & 0xc0) == 0x80)
                    {
                        tailStart++;
                    }
                    text = Encoding.UTF8.GetString(fresh.AsSpan(0, head))
                        + "\n[output truncated; full output on disk]\n"
                        + Encoding.UTF8.GetString(fresh.AsSpan(tailStart));
                }
                bool firstReport;
                lock (gate)
                {
                    firstReport = cursor == 0;
                    cursor += fresh.Length;
                }
                bool bareResult = firstReport && info.Status == "done" && !truncated && info.Cwd.Length == 0;
                var exit = info.Status != "killed" && info.ExitCode.HasValue ? $", exit_code: {info.ExitCode}" : "";
                var cwdNotice = info.Cwd.Length == 0 ? "" : $"cwd: {info.Cwd}\n";
                var header = cwdNotice + $"job {Handle}, status: {info.Status}, bytes: {info.Bytes}{exit}, truncated: {truncated.ToString().ToLowerInvariant()}, output_path: {OutputPath}\n";
                return new ToolResultPart(callID, [new TextPart(bareResult ? text : header + text)], error, info with
                {
                    Truncated = truncated
                });
            }
            await Task.WhenAny(signal, expires).WaitAsync(token);
        }
    }

    public partial async Task<ToolResultPart> SendInputAsync(string callID, string text, double delay,
        string pattern, int maxBytes, CancellationToken cancellation)
    {
        var regex = pattern == "" ? null : new Regex(pattern, RegexOptions.None, TimeSpan.FromSeconds(1));
        await observation.WaitAsync(cancellation);
        try
        {
            lock (gate)
            {
                if (status != JobStatus.Running || process is null || process.HasExited)
                {
                    throw new ArgumentException("Job has no running PTY stdin");
                }
                // Start after the displayed prompt and submit one terminal line.
                // Exact raw-byte input is deliberately not part of this chapter.
                cursor = bytes;
                process.StandardInput.Write(text.EndsWith("\n", StringComparison.Ordinal) ? text : text + "\n");
                process.StandardInput.Flush();
            }
            return await ObserveLocked(callID, delay, regex, maxBytes, cancellation, wakeOnOutput: true);
        }
        finally
        {
            observation.Release();
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
        }
        try
        {
            KillAsync("shutdown").GetAwaiter().GetResult();
        }
        finally
        {
            output.Dispose();
            process?.Dispose();
            // Do not dispose the token source under a late managed callback.
            // Its managed state is reclaimed with the Job; OS resources close now.
            parent?.RemoveJob(this);
        }
        // Retain the output artifact: recorded reports still point to it.
    }
}
