namespace Ensemble;

public sealed partial class Agent
{
    private partial void ApplyToolClear()
    {
        if (!pendingToolClear)
        {
            return;
        }
        var snapshot = History().Snapshot();
        if (snapshot.Context.Dialogue.Any(entry =>
            entry.Parts.Any(part => part is ToolCallPart or ToolResultPart)))
        {
            // The actor reaches this point only after paired batch outcomes.
            // Capture the range now, not when the host requested the clear:
            // otherwise late results could survive without their calls. Replay
            // uses this fixed range and never consults a live policy or clock.
            Record(new Event(EventKind.Redacted, redactFrom: 1,
                redactTo: snapshot.Sequence, level: Redaction.Tool,
                reason: "manual tool clear"));
        }
        // A failed append propagates before this reset. Empty/repeated clears
        // are harmless and need no manufactured reduction event.
        pendingToolClear = false;
    }
}
