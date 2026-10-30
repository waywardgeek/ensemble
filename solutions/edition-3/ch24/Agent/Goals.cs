using System.Text;

namespace Ensemble;

public sealed partial class Agent
{
    public partial Task<GoalStatus> GoalControlAsync(GoalUpdate? update, CancellationToken cancellation)
    {
        cancellation.ThrowIfCancellationRequested();
        var reply = new TaskCompletionSource<GoalStatus>(TaskCreationOptions.RunContinuationsAsynchronously);
        Enqueue(new Mail(MailKind.GoalControl, GoalUpdate: update, GoalReply: reply));
        return reply.Task.WaitAsync(cancellation);
    }

    public partial Task<GoalStatus> GoalStatusAsync(CancellationToken cancellation)
    {
        cancellation.ThrowIfCancellationRequested();
        var reply = new TaskCompletionSource<GoalStatus>(TaskCreationOptions.RunContinuationsAsynchronously);
        Enqueue(new Mail(MailKind.GoalStatus, GoalReply: reply));
        return reply.Task.WaitAsync(cancellation);
    }

    internal partial bool GoalAvailable(Context? context)
    {
        // The legacy library mode permits its ordinary catalog without manuals.
        // Goals is deliberately not part of that fallback: an attached Skills
        // owner and a retained, active grant are both required.
        return skills is not null && AllowsMemoryTool("goal_control")
            && (context ?? History().View()).EnabledTools(Configuration()).Contains("goal_control");
    }

    private partial bool GoalIdle()
    {
        // Independent Jobs may be running: the reminder must inspect them, not
        // cancel them. Foreground controls, admitted turns and direct sends are busy.
        lock (ingress)
        {
            return !admissionClosed && !destroyed && !stopping && !paused
                && active is null && requests.Count == 0 && !workerPending
                && !transportPending && !automaticPending && startupReplies.Count == 0
                && controlReplies.Count == 0 && engineSends == 0;
        }
    }

    private partial GoalStatus CaptureGoalStatus()
    {
        bool available = GoalAvailable();
        lock (ingress)
        {
            string phase = !available ? "unloaded"
                : goalConfiguration.File.Length == 0 ? "unconfigured"
                : goalConfiguration.LastError.Length != 0 ? "file_error"
                : paused ? "paused"
                : !GoalIdle() ? "busy"
                : !goalConfiguration.Enabled ? "off"
                : goalDue.HasValue ? "scheduled" : "busy";
            // A submission can disable configuration before the actor drains its
            // refresh mail. Never expose that invalidated timer as a future wake.
            DateTimeOffset? due = phase == "scheduled" ? goalDue : null;
            return new GoalStatus(goalConfiguration.File, goalConfiguration.Enabled,
                goalConfiguration.IdleSeconds, phase, due, goalConfiguration.LastError, available);
        }
    }

    private partial void ConfigureGoal(Mail message)
    {
        try
        {
            ObjectDisposedException.ThrowIf(stopping, this);
            if (!GoalAvailable())
            {
                throw new ArgumentException("Load the goals skill before using goal_control.");
            }
            GoalUpdate? update = message.GoalUpdate;
            if (update is not null && (update.File is not null || update.Enabled.HasValue || update.IdleSeconds.HasValue))
            {
                GoalConfiguration previous;
                lock (ingress)
                {
                    previous = goalConfiguration;
                }
                double seconds = update.IdleSeconds ?? previous.IdleSeconds;
                // System.Threading.Timer uses a finite millisecond interval.
                // Reject values that round down to zero or overflow that contract.
                if (!double.IsFinite(seconds) || seconds < .001 || seconds > (uint.MaxValue - 1) / 1000d)
                {
                    throw new ArgumentException("idle_seconds must be finite, at least .001, and fit the timer interval.");
                }
                string file = update.File?.Trim() ?? previous.File;
                bool enabled = update.Enabled ?? previous.Enabled;
                if (update.File is not null && file.Length == 0)
                {
                    enabled = false;
                }
                // Off works even after a file disappears. Enabling or selecting
                // a nonempty file verifies it before changing current settings.
                if (file.Length != 0 && (update.File is not null || update.Enabled == true))
                {
                    file = Sandbox().Resolve(file);
                    _ = ReadGoalText(file);
                }
                if (enabled && file.Length == 0)
                {
                    throw new ArgumentException("Configure a readable goal file before enabling reminders.");
                }
                StopGoalTimer();
                lock (ingress)
                {
                    goalConfiguration = new GoalConfiguration(file, enabled, seconds);
                }
            }
            RefreshGoalTimer();
            message.GoalReply!.TrySetResult(CaptureGoalStatus());
        }
        catch (Exception error) when (error is ArgumentException or IOException
            or UnauthorizedAccessException or InvalidOperationException)
        {
            message.GoalReply!.TrySetException(error);
        }
    }

    private partial string ReadGoalText(string file)
    {
        // The same retained-descriptor boundary used by corpus reads refuses
        // FIFOs and outward paths. Read at most one MiB plus one sentinel byte;
        // strict UTF-8 and nonblank text prevent silent loss of the objective.
        const int maximum = 1024 * 1024;
        using var stream = Sandbox().OpenRegularRead(file);
        using var bytes = new MemoryStream();
        var buffer = new byte[8192];
        while (bytes.Length <= maximum)
        {
            int count = stream.Read(buffer, 0, (int)Math.Min(buffer.Length, maximum + 1L - bytes.Length));
            if (count == 0)
            {
                break;
            }
            bytes.Write(buffer, 0, count);
        }
        if (bytes.Length > maximum)
        {
            throw new InvalidDataException("Goal file exceeds the one MiB read limit.");
        }
        string text = new UTF8Encoding(false, true).GetString(bytes.ToArray()).TrimStart('\uFEFF');
        if (string.IsNullOrWhiteSpace(text))
        {
            throw new InvalidDataException("Goal file is empty.");
        }
        return text;
    }

    internal partial string GoalInstruction(bool automatic)
    {
        if (!GoalAvailable())
        {
            return "";
        }
        GoalConfiguration captured;
        lock (ingress)
        {
            captured = goalConfiguration;
        }
        if (captured.File.Length == 0)
        {
            return "";
        }
        try
        {
            string text = ReadGoalText(captured.File);
            // Configuration can change during a public library render. Never
            // publish stale bytes after a clear/reconfiguration wins that race.
            lock (ingress)
            {
                if (!ReferenceEquals(captured, goalConfiguration))
                {
                    return "";
                }
            }
            return $"Current goal ({captured.File}):\n{text}";
        }
        catch (Exception error) when (error is IOException or UnauthorizedAccessException or ArgumentException)
        {
            lock (ingress)
            {
                if (!ReferenceEquals(captured, goalConfiguration))
                {
                    return "";
                }
            }
            // Filesystem exception messages may disclose resolved host paths.
            // Keep a stable diagnostic; never copy arbitrary OS text to History.
            const string diagnostic = "Configured goal file is unavailable, empty, too large or not valid UTF-8. Reminders are off; repair it and explicitly re-enable.";
            var failure = new GoalReadFailure(captured, diagnostic);
            mailbox.Writer.TryWrite(new Mail(MailKind.GoalReadFailed, GoalFailure: failure));
            if (automatic)
            {
                throw new InvalidDataException(diagnostic);
            }
            return diagnostic;
        }
    }

    private partial void FailGoalRead(GoalReadFailure failure)
    {
        lock (ingress)
        {
            if (!ReferenceEquals(goalConfiguration, failure.Configuration))
            {
                return;
            }
            goalConfiguration = goalConfiguration with { Enabled = false, LastError = failure.Error };
        }
        StopGoalTimer();
    }

    private partial void StopGoalTimer()
    {
        goalGeneration++;
        goalTimer?.Dispose();
        goalTimer = null;
        goalDue = null;
    }

    private partial void GoalSubmitted()
    {
        // StoreSubmission already holds ingress and has durably saved its result.
        // Publish disabled configuration before waiters can observe submission.
        // A queued wake checks this fact; only the actor disposes/rearms the timer.
        lock (ingress)
        {
            if (!goalConfiguration.Enabled)
            {
                return;
            }
            goalConfiguration = goalConfiguration with { Enabled = false };
            mailbox.Writer.TryWrite(new Mail(MailKind.GoalRefresh));
        }
    }

    private partial void DisableGoal()
    {
        StopGoalTimer();
        lock (ingress)
        {
            goalConfiguration = goalConfiguration with { Enabled = false };
        }
    }

    private partial void ClearGoal()
    {
        StopGoalTimer();
        lock (ingress)
        {
            goalConfiguration = new GoalConfiguration();
        }
    }

    internal partial void GoalSkillsChanged()
    {
        // Skill transitions already publish revocation before returning. Retained
        // manuals grant nothing. A queued clear cannot be undone by a quick reload.
        if (!GoalAvailable())
        {
            mailbox.Writer.TryWrite(new Mail(MailKind.GoalClear));
        }
    }

    private partial void RefreshGoalTimer()
    {
        GoalConfiguration configuration;
        lock (ingress)
        {
            configuration = goalConfiguration;
        }
        if (!configuration.Enabled || configuration.File.Length == 0 || !GoalAvailable() || !GoalIdle())
        {
            if (goalTimer is not null)
            {
                StopGoalTimer();
            }
            return;
        }
        if (goalTimer is not null)
        {
            return; // Status reads never reset the clock.
        }
        long generation = ++goalGeneration;
        TimeSpan delay = TimeSpan.FromSeconds(configuration.IdleSeconds);
        goalDue = DateTimeOffset.UtcNow + delay;
        goalTimer = new Timer(_ => mailbox.Writer.TryWrite(new Mail(MailKind.GoalWake,
            GoalGeneration: generation)), null, delay, Timeout.InfiniteTimeSpan);
    }

    private partial void WakeGoal(Mail message)
    {
        if (message.GoalGeneration != goalGeneration)
        {
            return;
        }
        StopGoalTimer();
        GoalConfiguration configuration;
        lock (ingress)
        {
            configuration = goalConfiguration;
        }
        if (!configuration.Enabled || !GoalAvailable() || !GoalIdle())
        {
            return;
        }
        // No reliable human reply is invented. This is a normal actor turn with
        // an explicitly labelled runtime reminder, not a completion operation.
        try
        {
            _ = GoalInstruction(automatic: true);
        }
        catch (InvalidDataException)
        {
            // The renderer posted the correlated failure; no turn or recall has
            // been admitted. The mailbox will disable this configuration.
            return;
        }
        Receive(new Mail(MailKind.Submit, Parts: [new TextPart(GoalReminder)],
            AutomaticGoal: true));
    }
}
