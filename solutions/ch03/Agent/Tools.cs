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
        foreach (var kind in Enum.GetValues<ToolKind>())
        {
            _ = new Tool(this, kind);
        }
    }

    public partial async Task<ToolResultPart> DispatchAsync(ToolCallPart call, CancellationToken cancellation)
    {
        var tool = tools.SingleOrDefault(candidate => candidate.Name() == call.Name);
        if (tool is null)
        {
            return new ToolResultPart(call.CallID,
                [new TextPart("Unknown tool: " + call.Name)], isError: true);
        }
        return await tool.ExecuteAsync(call, cancellation);
    }
}
