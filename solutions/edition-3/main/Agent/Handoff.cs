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
