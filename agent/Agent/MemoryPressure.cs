using System.Globalization;

namespace Ensemble;

public sealed partial class Agent
{
    private partial WorkingContextLimit EffectiveWorkingContextLimit()
    {
        // Callers hold ingress. Exact routing identity, not a billing alias.
        var configured = workingContextLimits.FirstOrDefault(limit =>
            limit.Vendor == configuration.Vendor && limit.Model == configuration.Model);
        if (configured is not null) return configured;
        global::Ensemble.Engine.LookupModel(configuration.Vendor, configuration.Model, out var features);
        long? capacity = features.ContextWindow;
        return new WorkingContextLimit(configuration.Vendor, configuration.Model,
            Math.Min(250_000, capacity ?? 250_000), capacity);
    }

    public partial WorkingContextStatus WorkingContextStatus()
    {
        lock (ingress)
        {
            var limit = EffectiveWorkingContextLimit();
            return new(configuration.Vendor, configuration.Model, mainContextInputTokens, limit.WorkingTokens,
                limit.CapacityTokens, mainContextOccupancy);
        }
    }

    internal partial bool RequiresMemoryHandoff()
    {
        lock (ingress) { return memoryHandoffRequired; }
    }

    private partial void RecalculateMemoryPressure()
    {
        lock (ingress)
        {
            var limit = EffectiveWorkingContextLimit();
            mainContextOccupancy = mainContextInputTokens is long input
                ? 100m * input / limit.WorkingTokens : null;
            memoryHandoffRequired = !memoryCompressor && mainContextOccupancy >= 100m;
        }
    }

    private partial void ObserveMemoryPressure(Event response)
    {
        lock (ingress)
        {
            // Usage is already normalized into disjoint provider-aware buckets.
            // Auxiliary sends and session totals never enter this measurement.
            mainContextInputTokens = memoryCompressor ? null : response.Usage?.AllInput();
            RecalculateMemoryPressure();
        }
    }

    private partial string MemoryPressureInstruction()
    {
        lock (ingress)
        {
            if (memoryCompressor || mainContextOccupancy is not >= 80m
                || mainContextOccupancy < 90m && memoryPressureReminderSent)
                return "";
            string percent = mainContextOccupancy.Value.ToString("0.0", CultureInfo.InvariantCulture);
            string cadence = mainContextOccupancy >= 90m
                ? "This reminder repeats on every model request until a checkpoint or lower reported usage."
                : "This is the once-per-checkpoint reminder.";
            return $"Ensemble context reminder: the last reported main request used {percent}% of your configured working input-token limit. Please call micro_handoff at the next good handoff point. {cadence} This is reported usage, not a prediction of the next request.";
        }
    }

    private partial void AcknowledgeMemoryPressure()
    {
        lock (ingress)
        {
            memoryHandoffRequired = false;
            memoryPressureReminderSent = false;
            // The measured request predates the cut. Unknown is honest; zero
            // would falsely claim that retained visible dialogue has no cost.
            mainContextInputTokens = null;
            mainContextOccupancy = null;
        }
    }

    private partial void ResetMemoryPressure() => AcknowledgeMemoryPressure();
}
