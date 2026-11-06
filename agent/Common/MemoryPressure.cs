namespace Ensemble;

public sealed partial class Agent
{
    // Transient main-actor measurement, never restored billing and never an
    // auxiliary/direct-library send. A decimal ratio avoids integer overflow.
    private decimal? mainContextOccupancy;
    private long? mainContextInputTokens;
    private bool memoryPressureReminderSent;
    // Ingress protects the public measurement snapshot; the actor stages the cut.
    // Stage automatic maintenance; do not manufacture a model tool grant.
    private bool memoryHandoffRequired;
    public partial WorkingContextStatus WorkingContextStatus();
    private partial WorkingContextLimit EffectiveWorkingContextLimit();
    private partial void RecalculateMemoryPressure();
    internal partial bool RequiresMemoryHandoff();
    private partial void ObserveMemoryPressure(Event response);
    private partial string MemoryPressureInstruction();
    private partial void AcknowledgeMemoryPressure();
    private partial void ResetMemoryPressure();
}
