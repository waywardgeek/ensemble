using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Agent
{
    public partial MemorySettings MemorySettings()
    {
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            return memorySettings;
        }
    }

    public partial async Task SetMemorySettingsAsync(MemorySettings settings, CancellationToken cancellation)
    {
        ArgumentNullException.ThrowIfNull(settings);
        cancellation.ThrowIfCancellationRequested();
        var reply = new TaskCompletionSource<string>(TaskCreationOptions.RunContinuationsAsynchronously);
        Enqueue(new Mail(MailKind.MemorySettings, Reply: reply, MemorySettings: settings));
        await reply.Task.WaitAsync(cancellation);
    }

    public partial async Task SaveMemorySettingsAsync(MemorySettings settings, CancellationToken cancellation)
    {
        ArgumentNullException.ThrowIfNull(settings);
        cancellation.ThrowIfCancellationRequested();
        var reply = new TaskCompletionSource<string>(TaskCreationOptions.RunContinuationsAsynchronously);
        Enqueue(new Mail(MailKind.MemorySettings, Reply: reply, MemorySettings: settings,
            PersistMemorySettings: true));
        await reply.Task.WaitAsync(cancellation);
    }

    private partial bool SaveMemorySettingsBeforeApply(Mail message)
    {
        // Recovery can publish journal facts. A recording failure here must
        // fault the actor, not masquerade as a recoverable save refusal.
        if (!memoryReady)
        {
            RecoverMemory();
        }
        try
        {
            // Preflight before saving or canceling a live attempt.
            if (History().View().Compression is not null && memoryCapture is null)
            {
                throw new InvalidOperationException("Memory recovery must complete before changing settings.");
            }
            ReadMemoryChanges(message.MemorySettings!, startup: !memoryReady);
            var saved = Memory().ReadSettings();
            Memory().WriteSettings(new AgentSettings(message.MemorySettings!,
                saved?.Recall ?? RecallSettings()));
            return true;
        }
        catch (Exception error) when (error is IOException or InvalidDataException or UnauthorizedAccessException
            or ArgumentException or InvalidOperationException or FormatException or System.Text.Json.JsonException)
        {
            message.Reply!.TrySetException(error);
            return false;
        }
    }

    private partial void RecoverMemory()
    {
        // This is startup-only. A process-local worker cannot have been launched
        // yet. Never use crash recovery to cancel still-running compression.
        if (dataDirectory is null)
        {
            if (History().View().Compression is not null)
            {
                throw new InvalidOperationException("Memory recovery requires an Agent data home");
            }
            return;
        }
        var outcome = Memory().Recover(History().View());
        if (outcome is not null)
        {
            Record(new Event(EventKind.CompressionFinished, outcome: outcome));
            // Release staging only after the terminal outcome is durable. A
            // failed append leaves the committed receipt available at restart.
            Memory().Recover(History().View());
        }
    }

    private partial void EnsureMemoryReady()
    {
        if (memoryReady)
        {
            return;
        }
        RecoverMemory();
        // Internal compressors retain their deliberately disabled policy.
        // Transient library Agents have no implicit filesystem configuration.
        var saved = dataDirectory is null || memoryCompressor ? null : Memory().ReadSettings();
        var desired = saved?.Memory ?? MemorySettings();
        var changes = ReadMemoryChanges(desired, startup: true);
        PublishMemorySettings(desired, changes);
        if (saved is not null)
        {
            SetRecallSettings(saved.Recall);
        }
    }

    private partial void ApplyMemorySettings(Mail message)
    {
        ImmutableArray<Event> changes;
        RecallSettings? startupRecall = null;
        var desired = message.MemorySettings!;
        if (!memoryReady)
        {
            // Recovery includes durable publication. A journal failure here
            // belongs to the actor's fatal path, not a recoverable scan refusal.
            RecoverMemory();
        }
        try
        {
            // Read every required file before recording anything. A malformed
            // or unreadable newly enabled band must not partly apply policy.
            if (!memoryReady && dataDirectory is not null && !memoryCompressor)
            {
                startupRecall = Memory().ReadSettings()?.Recall;
            }
            changes = ReadMemoryChanges(desired, startup: !memoryReady);
        }
        catch (Exception error)
        {
            message.Reply!.TrySetException(error);
            return;
        }
        // Recording failures are fatal to the actor, not a successful settings
        // update. The loop settles this reply and drains other admitted work.
        PublishMemorySettings(desired, changes);
        if (startupRecall is not null)
        {
            SetRecallSettings(startupRecall);
        }
        message.Reply!.TrySetResult("");
    }

    private partial ImmutableArray<Event> ReadMemoryChanges(MemorySettings desired, bool startup)
    {
        var context = History().View();
        if (context.Compression is { } pending
            && (pending.Committed is not null || memoryCapture is null
                || !JsonCodec.SameCompressionLaunch(pending.Launch, memoryCapture.Launch)))
        {
            // Only this actor's live, uncommitted attempt can coexist with a
            // settings update. Other replay receipts require recovery first.
            throw new InvalidOperationException("Recover pending memory work before changing settings");
        }
        var previous = MemorySettings();
        var changes = ImmutableArray.CreateBuilder<Event>();
        foreach (var band in Enum.GetValues<MemoryBand>())
        {
            var policy = desired.For(band);
            var old = context.Memories.Where(entry => entry.Band == band).ToArray();
            if (policy.Disabled)
            {
                if (old.Length != 0)
                {
                    changes.Add(new Event(EventKind.MemoryRetired,
                        retirement: new MemoryRetirement(band), reason: "disabled"));
                }
                continue;
            }
            if (dataDirectory is null || (!startup && !previous.For(band).Disabled))
            {
                // Pure library replay without a home remains pure. Ordinary
                // size changes also need no disk read; re-enable and startup do.
                continue;
            }
            var fresh = Memory().Scan(context, band, memoryCapture);
            var ids = fresh.Select(entry => entry.ID).ToHashSet(StringComparer.Ordinal);
            var missing = old.Where(entry => !ids.Contains(entry.ID)).Select(entry => entry.ID).ToArray();
            if (missing.Length != 0)
            {
                changes.Add(new Event(EventKind.MemoryRetired,
                    retirement: new MemoryRetirement(band, missing), reason: "disk scan"));
            }
            foreach (var entry in fresh)
            {
                var saved = old.FirstOrDefault(item => item.ID == entry.ID);
                if (saved is null || saved.Timestamp != entry.Timestamp || saved.Text != entry.Text)
                {
                    changes.Add(new Event(EventKind.MemoryPopulated, memory: entry, reason: "disk scan"));
                }
            }
        }
        return changes.ToImmutable();
    }

    private partial void PublishMemorySettings(MemorySettings desired, ImmutableArray<Event> changes)
    {
        foreach (var fact in changes)
        {
            Record(fact);
        }
        lock (ingress)
        {
            memorySettings = desired;
        }
        memoryReady = true;
    }
}
