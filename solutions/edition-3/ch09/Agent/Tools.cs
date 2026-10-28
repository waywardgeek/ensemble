using System.Collections.Immutable;
namespace Ensemble;

public sealed partial class Agent
{
    public partial ImmutableArray<Tool> Tools() => tools.ToImmutableArray();

    public partial void AddTool(Tool tool)
    {
        if (destroyed || tool.destroyed || tool.parent is not null
            || tools.Any(existing => existing.Name() == tool.Name()))
        {
            throw new InvalidOperationException("Tool is destroyed, already attached, or has a duplicate name");
        }
        tools.Add(tool);
        tool.parent = this;
    }

    public partial void RemoveTool(Tool tool)
    {
        if (tool.parent != this || !tools.Contains(tool))
        {
            throw new InvalidOperationException("Tool belongs to another Agent");
        }
        tools.Remove(tool);
        tool.parent = null;
    }

    public partial void InstallTools()
    {
        if (destroyed || tools.Count != 0)
        {
            throw new InvalidOperationException("Install tools only on an empty live Agent");
        }
        foreach (var kind in Enum.GetValues<ToolKind>().Where(kind => kind is not (ToolKind.Custom or ToolKind.LoadSkill or ToolKind.UnloadSkill)))
        {
            _ = new Tool(this, kind);
        }
    }

    public partial async Task<ToolResultPart> DispatchAsync(ToolCallPart call, CancellationToken cancellation)
    {
        var reply = new TaskCompletionSource<ToolResultPart>(TaskCreationOptions.RunContinuationsAsynchronously);
        Enqueue(new Mail(MailKind.DirectTool, Call: call, ToolReply: reply));
        // Like SendAsync, cancellation detaches this waiter only. Explicit
        // interrupt is the operation that wakes the actor's foreground wait.
        return await reply.Task.WaitAsync(cancellation);
    }
}
