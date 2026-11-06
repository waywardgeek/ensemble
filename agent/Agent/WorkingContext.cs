using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Agent
{
    public partial ImmutableArray<WorkingContextLimit> WorkingContextLimits()
    {
        lock (ingress) { return workingContextLimits; }
    }

    private partial void RestoreWorkingContextLimits(ImmutableArray<WorkingContextLimit> limits)
    {
        lock (ingress) { workingContextLimits = limits.IsDefault ? [] : limits; }
    }

    public partial async Task SetWorkingContextLimitAsync(WorkingContextLimit limit,
        bool persist, CancellationToken cancellation)
    {
        ArgumentNullException.ThrowIfNull(limit);
        cancellation.ThrowIfCancellationRequested();
        var reply = new TaskCompletionSource<string>(TaskCreationOptions.RunContinuationsAsynchronously);
        Enqueue(new Mail(MailKind.WorkingContextSettings, Reply: reply,
            WorkingContextLimit: limit, PersistWorkingContext: persist));
        // Cancellation detaches this caller, never retracts an admitted save.
        await reply.Task.WaitAsync(cancellation);
    }

    private partial void ApplyWorkingContextLimit(Mail message)
    {
        try
        {
            var limit = message.WorkingContextLimit!;
            var current = WorkingContextLimits();
            var next = current.Where(item => item.Vendor != limit.Vendor || item.Model != limit.Model)
                .Append(limit).ToImmutableArray();
            if (next.Length > 64) throw new InvalidDataException("At most 64 model context limits are supported");
            if (message.PersistWorkingContext)
            {
                var saved = Memory().ReadSettings();
                Memory().WriteSettings(new AgentSettings(saved?.Memory ?? MemorySettings(),
                    saved?.Recall ?? RecallSettings(), next));
            }
            lock (ingress)
            {
                RestoreWorkingContextLimits(next);
                RecalculateMemoryPressure();
            }
            message.Reply!.TrySetResult("");
        }
        catch (Exception error)
        {
            message.Reply!.TrySetException(error);
        }
    }
}
