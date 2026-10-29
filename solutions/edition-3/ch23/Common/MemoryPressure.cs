namespace Ensemble;

public sealed partial class Agent
{
    // Transient main-actor measurement, never restored billing and never an
    // auxiliary/direct-library send. A decimal ratio avoids integer overflow.
    private decimal? mainContextOccupancy;
    private bool memoryPressureAcknowledged;
    // Only this dispatch latch is read outside the actor; ingress protects it.
    // Restrict existing grants, never create a new grant for micro_handoff.
    private bool memoryHandoffRequired;
    // A foreground wait observes one existing compression attempt. Interrupt
    // detaches the wait, not the independently owned compressor. A successful
    // commit satisfies this measured request even if another band cascades.
    private string? memoryHeadroomTaskID;
    private string? memoryHeadroomError;
    private bool memoryHeadroomSatisfied;

    internal partial bool RequiresMemoryHandoff();
    internal partial bool AllowsMemoryTool(string name);
    private partial void ObserveMemoryPressure(Event response);
    private partial string MemoryPressureInstruction();
    private partial void AcknowledgeMemoryPressure();
    private partial void ResetMemoryPressure();
    private partial bool WaitForMemoryHeadroom();
    private partial void ResolveMemoryHeadroom(string taskID, bool committed);
}
