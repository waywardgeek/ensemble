using System.Collections.Immutable;
namespace Ensemble;

public sealed partial class Agent
{
    public partial ImmutableArray<Tool> Tools()
    {
        lock (toolsGate)
        {
            return tools.ToImmutableArray();
        }
    }

    public partial void AddTool(Tool tool)
    {
        lock (toolsGate)
        {
            if (destroyed || tool.destroyed || tool.parent is not null
                || tools.Any(existing => existing.Name() == tool.Name())
                || (tool.remote is not null && tool.remote.Parent().Parent() != this))
            {
                throw new InvalidOperationException("Invalid or duplicate tool");
            }
            tools.Add(tool);
            tool.parent = this;
        }
    }

    public partial void RemoveTool(Tool tool)
    {
        lock (toolsGate)
        {
            if (tool.parent != this || !tools.Contains(tool))
            {
                throw new InvalidOperationException("Tool does not belong to this Agent");
            }
            tools.Remove(tool);
            tool.parent = null;
        }
    }

    public partial void InstallTools()
    {
        if (destroyed || !Tools().IsEmpty)
        {
            throw new InvalidOperationException("Install tools only on an empty live Agent");
        }
        foreach (var kind in Enum.GetValues<ToolKind>().Where(kind => kind is not (ToolKind.Custom or ToolKind.Remote or ToolKind.LoadSkill or ToolKind.UnloadSkill)))
        {
            if (kind == ToolKind.KeepToolResults && !configuration.StubsToolResults)
            {
                continue;
            }
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
