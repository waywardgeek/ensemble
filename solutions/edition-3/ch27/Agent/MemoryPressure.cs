using System.Globalization;

namespace Ensemble;

public sealed partial class Agent
{
    internal partial bool RequiresMemoryHandoff()
    {
        lock (ingress)
        {
            return memoryHandoffRequired;
        }
    }

    internal partial bool AllowsMemoryTool(string name)
    {
        return !RequiresMemoryHandoff() || name == "micro_handoff";
    }

    private partial void ObserveMemoryPressure(Event response)
    {
        mainContextOccupancy = null;
        memoryPressureAcknowledged = false;
        memoryHeadroomSatisfied = false;
        // The selected routing model owns this request's documented window.
        // Actual billing identity may be a dated provider model; it owns prices,
        // not an inferred capability alias. Unknown windows stay unknown.
        if (memoryCompressor || !global::Ensemble.Engine.LookupModel(configuration.Vendor,
                configuration.Model, out var features)
            || features.ContextWindow is not > 0
            || response.Usage?.AllInput() is not long input)
        {
            return;
        }
        mainContextOccupancy = 100m * input / features.ContextWindow.Value;
        if (mainContextOccupancy >= 95m)
        {
            lock (ingress)
            {
                memoryHandoffRequired = true;
            }
        }
    }

    private partial string MemoryPressureInstruction()
    {
        // A later short/usage-less reply does not discharge the checkpoint.
        // Explain the retained latch even when the latest measurement is low.
        if (RequiresMemoryHandoff() && mainContextOccupancy is not >= 90m)
        {
            return "An earlier main request requires a context checkpoint. Only micro_handoff is currently enabled. Call it before other tools.";
        }
        if (memoryPressureAcknowledged || mainContextOccupancy is not >= 90m)
        {
            return "";
        }
        string percent = mainContextOccupancy.Value.ToString("0.0", CultureInfo.InvariantCulture);
        string action = RequiresMemoryHandoff()
            ? "Only micro_handoff is currently enabled. Call it before other tools."
            : "Call micro_handoff soon to preserve a continuation and free context headroom.";
        return $"Context window warning: the last main request used {percent}% of its known window. {action}";
    }

    private partial void AcknowledgeMemoryPressure()
    {
        lock (ingress)
        {
            memoryHandoffRequired = false;
        }
        memoryPressureAcknowledged = true;
        // A handoff strips tools/private replay, but does not prove that the
        // pending compressor has replaced its retained source. Keep that wait.
    }

    private partial void ResetMemoryPressure()
    {
        AcknowledgeMemoryPressure();
        mainContextOccupancy = null;
        memoryHeadroomTaskID = null;
        memoryHeadroomError = null;
        memoryHeadroomSatisfied = false;
    }

    private partial bool WaitForMemoryHeadroom()
    {
        if (memoryCompressor || memoryCapture is null || memoryHeadroomSatisfied
            || mainContextOccupancy is not >= 95m)
        {
            return false;
        }
        memoryHeadroomTaskID = memoryCapture.Launch.TaskID;
        return true;
    }

    private partial void ResolveMemoryHeadroom(string taskID, bool committed)
    {
        if (memoryHeadroomTaskID != taskID)
        {
            return;
        }
        memoryHeadroomTaskID = null;
        if (committed)
        {
            memoryHeadroomSatisfied = true;
        }
        else
        {
            memoryHeadroomError = "Memory headroom unavailable: compression did not commit. Source was retained; no main request was sent.";
        }
    }
}
