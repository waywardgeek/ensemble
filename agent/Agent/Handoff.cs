namespace Ensemble;

public sealed partial class Agent
{
    private partial void StageHandoff(ToolCallPart call, ToolResultPart result)
    {
        // A remote/custom tool merely named micro_handoff is not this control.
        if (!result.IsError && Tools().Any(tool => tool.Name() == call.Name && tool.IsHandoff()))
        {
            pendingHandoff = new HandoffPart(call.Args.GetProperty("text").GetString()!);
        }
    }

    private partial void ApplyHandoff()
    {
        if (pendingHandoff is null && RequiresMemoryHandoff())
        {
            // Advance/Finish call here only after companion results are paired.
            // This is an Agent action, not a model-authored tool invocation.
            pendingHandoff = new HandoffPart("Automatic working-context checkpoint: the configured input-token limit was reached without a model-authored handoff. No model-authored continuation note was supplied. The checkpoint is completed; continue from retained visible conversation and canonical History without repeating this handoff.");
        }
        if (pendingHandoff is null)
        {
            return;
        }
        // Record owns durable-before-publication. An append failure must not
        // clear the staged note or pretend that reduction succeeded.
        Record(new Event(EventKind.HandoffRecorded, parts: [pendingHandoff]));
        pendingHandoff = null;
        AcknowledgeMemoryPressure();
        StartMemoryCompression(includeConversation: true);
    }
}
