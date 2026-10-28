namespace Ensemble;

public sealed partial class Agent
{
    private partial void StartMemoryCompression(bool includeConversation)
    {
        // Handoffs can also be applied while a turn is being stopped. Closing
        // admission must never create another child during owner teardown.
        lock (ingress)
        {
            if (admissionClosed || stopping || memoryCompressor || dataDirectory is null)
            {
                return;
            }
        }
        if (memoryCapture is not null || !memoryReady)
        {
            return;
        }

        var policy = MemorySettings();
        var context = History().View();
        MemoryCapture? selected = null;
        try
        {
            // Conversation has priority only at a checkpoint. Successful file
            // commits can cascade, but cannot invent a new dialogue checkpoint.
            if (includeConversation && !policy.Conversation.Disabled)
            {
                selected = Memory().CaptureConversation(context, policy.Conversation.Budget, DateTimeOffset.UtcNow);
                if (selected is not null && policy.Session.Disabled)
                {
                    Parent().Log("Memory compression deferred: session band is disabled");
                    selected = null;
                }
            }
            foreach (var source in new[] { MemoryBand.Session, MemoryBand.Archive8, MemoryBand.Archive64 })
            {
                if (selected is not null)
                {
                    break;
                }
                var band = policy.For(source);
                var destination = source switch
                {
                    MemoryBand.Session => MemoryBand.Archive8,
                    MemoryBand.Archive8 => MemoryBand.Archive64,
                    _ => MemoryBand.Memory
                };
                if (band.Disabled)
                {
                    continue;
                }
                selected = Memory().CaptureGraduation(context, source, band.HighWatermark, band.LowWatermark, DateTimeOffset.UtcNow);
                if (selected is not null && policy.For(destination).Disabled)
                {
                    Parent().Log($"Memory compression deferred: {destination} band is disabled");
                    selected = null;
                }
            }
        }
        catch (Exception error)
        {
            // Disk edits and malformed corpus names are recoverable selection
            // failures. Do not loop on them, or hide them as model success.
            Parent().Log($"Memory selection refused: {error.GetType().Name}");
            return;
        }
        if (selected is null)
        {
            return;
        }

        // Durable launch precedes the child and every external effect. A failed
        // append faults the actor and never starts unrecorded compression.
        Record(new Event(EventKind.CompressionLaunched, compression: selected.Launch));
        memoryCapture = selected;
        memoryLaunchSettings = policy;
        memoryCancellation = new CancellationTokenSource();
        memoryTask = MemoryWorker(selected, policy.For(selected.Launch.Destination), memoryCancellation.Token);
    }

    private partial async Task<MemoryCompletion> MemoryWorker(MemoryCapture capture,
        MemoryBandSettings destination, CancellationToken cancellation)
    {
        MemoryCompletion completed;
        try
        {
            var result = await RunMemoryCompressorAsync(capture, destination, cancellation);
            completed = new MemoryCompletion(capture.Launch.TaskID, result, null);
        }
        catch (Exception error)
        {
            // Construction can fail before the runner has a child to tear down.
            // The mailbox carries a bounded diagnostic, never a filesystem path.
            completed = new MemoryCompletion(capture.Launch.TaskID, null, error.GetType().Name);
        }
        mailbox.Writer.TryWrite(new Mail(MailKind.MemoryDone, MemoryCompletion: completed));
        return completed;
    }

    private partial void AcceptMemoryCompletion(MemoryCompletion completion)
    {
        if (memoryCapture?.Launch.TaskID != completion.TaskID)
        {
            return;
        }
        memoryCompletion = completion;
        if (!memoryUsageImported)
        {
            // Called only after the runner joins its child. Import before any
            // parent journal write, including when that journal has faulted.
            if (completion.Result is { } result)
            {
                engine!.ImportUsage(result.Meter);
            }
            memoryUsageImported = true;
        }
    }

    private partial void CompleteMemoryCompression()
    {
        if (memoryCapture is null || memoryCompletion is null)
        {
            return;
        }
        var capture = memoryCapture;
        var launch = capture.Launch;
        var result = memoryCompletion.Result;
        if (result?.Cancelled == true)
        {
            EndMemoryCompression(CompressionState.Canceled, "Compressor canceled");
            return;
        }
        if (result?.Summary is null || result.Error is not null || memoryCompletion.Error is not null)
        {
            EndMemoryCompression(CompressionState.Failed,
                memoryCompletion.Error ?? result?.Error ?? "Compressor returned no summary");
            return;
        }

        MemoryEntry? replacement = null;
        bool current = false;
        Exception? preparationError = null;
        try
        {
            current = Memory().StillCurrent(capture, History().View());
            if (current)
            {
                Memory().Stage(capture, result.Summary);
                replacement = new MemoryEntry(launch.Destination, launch.ID, launch.Timestamp, result.Summary);
                // Archive canonical events, not the lossy readable worker input,
                // before replay removes any selected conversation source.
                if (launch.Source is null)
                {
                    History().ArchiveMemory();
                }
                // Archival can take time. Recheck exact selected bytes after
                // that I/O and immediately before publishing the held removal.
                current = Memory().StillCurrent(capture, History().View());
            }
        }
        catch (Exception error)
        {
            preparationError = error;
        }
        // Recording failures here belong to the actor's fatal path, never to the
        // recoverable preparation catch above.
        if (preparationError is not null)
        {
            EndMemoryCompression(CompressionState.Failed,
                $"Preparation failed: {preparationError.GetType().Name}");
            return;
        }
        if (!current)
        {
            EndMemoryCompression(CompressionState.Invalidated, "Selected source or destination changed");
            return;
        }

        // These two facts are replay-atomic: the first holds a removal; only the
        // matching population applies it. Any recording failure remains fatal.
        if (launch.Source is null)
        {
            Record(new Event(EventKind.Redacted, redactFrom: launch.From, redactTo: launch.To,
                level: Redaction.Summary, reason: "memory compression", taskID: launch.TaskID));
        }
        else
        {
            Record(new Event(EventKind.MemoryRetired,
                retirement: new MemoryRetirement(launch.Source.Value, launch.SourceIDs),
                taskID: launch.TaskID));
        }
        Record(new Event(EventKind.MemoryPopulated, memory: replacement!, taskID: launch.TaskID));
        // After commit, never turn a disk conflict into a failed/abandoned cut.
        // Keep the committed receipt and fault for deterministic startup recovery.
        var outcome = Memory().Recover(History().View())
            ?? throw new InvalidOperationException("Committed compression has no recovery outcome");
        var receipt = new CompressionOutcome(outcome.TaskID, outcome.State, outcome.Error,
            result.Meter.Rows);
        Record(new Event(EventKind.CompressionFinished, outcome: receipt));
        Memory().Recover(History().View());
        ResolveMemoryHeadroom(launch.TaskID, committed: true);
        ReleaseMemoryCompression();
        StartMemoryCompression(includeConversation: false);
    }

    private partial async Task InvalidateMemoryPolicy(MemorySettings desired)
    {
        if (memoryCapture is null)
        {
            return;
        }
        var launch = memoryCapture.Launch;
        var previous = memoryLaunchSettings!;
        if (!previous.For(launch.Source).Same(desired.For(launch.Source))
            || !previous.For(launch.Destination).Same(desired.For(launch.Destination)))
        {
            await StopMemoryCompression(CompressionState.Invalidated, "Compression policy changed");
        }
    }

    private partial async Task StopMemoryCompression(CompressionState state, string reason, bool record)
    {
        if (memoryTask is null)
        {
            return;
        }
        memoryCancellation!.Cancel();
        AcceptMemoryCompletion(await memoryTask);
        if (record)
        {
            EndMemoryCompression(state, reason);
        }
        else
        {
            // Leave replay pending for recovery if the actor failed, especially
            // after commit. The worker is nevertheless fully joined and detached.
            ReleaseMemoryCompression();
        }
    }

    private partial void EndMemoryCompression(CompressionState state, string reason)
    {
        Record(new Event(EventKind.CompressionFinished,
            outcome: new CompressionOutcome(memoryCapture!.Launch.TaskID, state, reason,
                memoryCompletion?.Result?.Meter.Rows)));
        // Discard a staged but uncommitted candidate only after its terminal fact
        // is durable. It must never become an active scan input.
        Memory().Recover(History().View());
        ResolveMemoryHeadroom(memoryCapture!.Launch.TaskID, committed: false);
        ReleaseMemoryCompression();
    }

    private partial void ReleaseMemoryCompression()
    {
        memoryCancellation?.Dispose();
        memoryCancellation = null;
        memoryTask = null;
        memoryCompletion = null;
        memoryCapture = null;
        memoryLaunchSettings = null;
        memoryUsageImported = false;
    }
}
