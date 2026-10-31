using System.Collections.Immutable;
using System.Text.Json;

namespace Ensemble;

public sealed partial class Agent
{
    internal partial async Task RegisterMcpToolsAsync(McpClient client, CancellationToken cancellation)
    {
        cancellation.ThrowIfCancellationRequested();
        var reply = new TaskCompletionSource<string>(TaskCreationOptions.RunContinuationsAsynchronously);
        Enqueue(new Mail(MailKind.RegisterMcpTools, Reply: reply, Client: client));
        // Like ordinary host input, cancelling this waiter does not undo an
        // already admitted mutation. The reliable reply still has one outcome.
        await reply.Task.WaitAsync(cancellation);
    }
}

public sealed partial class McpClient
{
    public partial async Task InstallToolsAsync(CancellationToken cancellation)
    {
        await ConnectAsync(cancellation);
        await Parent().Parent().RegisterMcpToolsAsync(this, cancellation);
    }

    public partial ImmutableArray<Tool> Tools()
    {
        lock (gate)
        {
            return tools.ToImmutableArray();
        }
    }

    public partial void AddTool(Tool tool)
    {
        lock (gate)
        {
            if (destroyed || ended || tool.destroyed || !tool.HasStructuredResult() || tool.remote is not null
                || tool.parent != Parent().Parent()
                || (SkillName is not null && (skillOwner is null || !skillOwner.Grants(SkillName, tool.Name())))
                || !catalog.Any(item => item.Name == tool.Name() && item.Ephemeral == "")
                || tools.Any(item => item.Name() == tool.Name()))
            {
                throw new InvalidOperationException("Invalid MCP tool relationship.");
            }
            tools.Add(tool);
            tool.remote = this;
        }
    }

    public partial void RemoveTool(Tool tool)
    {
        lock (gate)
        {
            if (tool.remote != this || !tools.Contains(tool))
            {
                throw new InvalidOperationException("Tool does not belong to this MCP client.");
            }
            tools.Remove(tool);
            tool.remote = null;
        }
    }

    internal partial void InstallTools()
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed || ended, this);
            if (SkillName is not null && skillOwner is null)
            {
                throw new InvalidOperationException("Skill connection is detached from its granting owner.");
            }
        }
        var agent = Parent().Parent();
        var definitions = Catalog().Where(item => item.Ephemeral == ""
            && (skillOwner is null || skillOwner.Grants(SkillName!, item.Name))).ToArray();
        // Validate the complete namespace before attaching any new Tool. Calling
        // this operation again is harmless, but a local/foreign name collision
        // must not leave a partially installed catalog.
        var existing = agent.Tools().ToDictionary(tool => tool.Name(), StringComparer.Ordinal);
        foreach (var definition in definitions)
        {
            if (existing.TryGetValue(definition.Name, out var tool) && tool.Remote() != this)
            {
                throw new ArgumentException("MCP tool name conflicts with an existing tool.");
            }
        }
        var added = new List<Tool>();
        try
        {
            foreach (var definition in definitions)
            {
                if (!existing.ContainsKey(definition.Name))
                {
                    added.Add(new Tool(agent, this, definition));
                }
            }
        }
        catch
        {
            foreach (var tool in added)
            {
                tool.Destroy();
            }
            throw;
        }
    }

    internal partial async Task<ToolResultPart> CallAsync(ToolCallPart call, CancellationToken cancellation)
    {
        var parameters = JsonSerializer.SerializeToElement(new
        {
            name = call.Name, arguments = call.Args
        });
        JsonElement result;
        try
        {
            result = await RequestAsync("tools/call", parameters, cancellation);
        }
        catch (ObjectDisposedException error)
        {
            // A Job can be admitted immediately before unload/shutdown closes
            // this peer. That is an external availability failure, not a broken
            // Job invariant; preserve it in the ordinary error/output channel.
            throw new IOException("MCP connection closed before tool execution.", error);
        }
        return McpContent.Decode(call.CallID, result);
    }
}
