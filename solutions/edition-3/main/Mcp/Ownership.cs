using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Agent
{
    public partial Mcp? Mcp()
    {
        lock (ingress)
        {
            return mcp;
        }
    }

    public partial Mcp EnsureMcp()
    {
        lock (ingress)
        {
            if (admissionClosed || destroying || destroyed)
            {
                throw new ObjectDisposedException(nameof(Agent));
            }
            // The constructor registers through AddMcp under this reentrant lock.
            return mcp ?? new Mcp(this);
        }
    }

    internal partial async Task<ToolResultPart> DispatchReverseAsync(ToolCallPart call,
        CancellationToken cancellation)
    {
        var reply = new TaskCompletionSource<ToolResultPart>(TaskCreationOptions.RunContinuationsAsynchronously);
        Enqueue(new Mail(MailKind.ReverseTool, Call: call, ToolReply: reply));
        return await reply.Task.WaitAsync(cancellation);
    }

    public partial void AddMcp(Mcp child)
    {
        lock (ingress)
        {
            if (admissionClosed || destroying || destroyed || child.destroyed || mcp is not null || child.parent is not null)
            {
                throw new InvalidOperationException("MCP manager is closed, destroyed or already owned.");
            }
            mcp = child;
            child.parent = this;
        }
    }

    public partial void RemoveMcp(Mcp child)
    {
        lock (ingress)
        {
            if (!ReferenceEquals(mcp, child) || !ReferenceEquals(child.parent, this))
            {
                throw new InvalidOperationException("MCP manager belongs to another Agent.");
            }
            mcp = null;
            child.parent = null;
        }
    }

}

public sealed partial class Mcp
{
    public partial Mcp(Agent parent)
    {
        ArgumentNullException.ThrowIfNull(parent);
        parent.AddMcp(this);
    }

    public partial Agent Parent() => parent ?? throw new ObjectDisposedException(nameof(Mcp));

    public partial ImmutableArray<McpClient> Clients()
    {
        lock (gate)
        {
            return clients.ToImmutableArray();
        }
    }

    public partial void AddClient(McpClient child)
    {
        lock (gate)
        {
            if (destroyed || child.destroyed || child.parent is not null
                // A configured server name belongs to its skill, not a global namespace.
                || clients.Any(existing => existing.Name == child.Name && existing.SkillName == child.SkillName))
            {
                throw new InvalidOperationException("MCP client is destroyed, owned, or has a duplicate name.");
            }
            clients.Add(child);
            child.parent = this;
        }
    }

    public partial void RemoveClient(McpClient child)
    {
        lock (gate)
        {
            if (!ReferenceEquals(child.parent, this) || !clients.Contains(child))
            {
                throw new InvalidOperationException("MCP client belongs to another manager.");
            }
            clients.Remove(child);
            child.parent = null;
        }
    }

    public partial Task DestroyAsync()
    {
        lock (gate)
        {
            destroyed = true;
            SignalBrowserChanged();
            destruction ??= Task.Run(DestroyCoreAsync);
            return destruction;
        }
    }

    private partial async Task DestroyCoreAsync()
    {
        List<Exception> errors = [];
        var registeredBrowser = Browser();
        if (registeredBrowser is not null)
        {
            try
            {
                await registeredBrowser.DestroyAsync();
            }
            catch (Exception error)
            {
                errors.Add(error);
            }
        }
        // Browser destruction already detached its multiply-owned client.
        // Snapshot only the remaining clients instead of destroying it twice.
        foreach (var client in Clients())
        {
            try
            {
                await client.DestroyAsync();
            }
            catch (Exception error)
            {
                // Cleanup must attempt every child, not conceal a broken peer
                // by abandoning the rest of the owned connection graph.
                errors.Add(error);
            }
        }
        parent?.RemoveMcp(this);
        if (errors.Count > 0)
        {
            throw new AggregateException(errors);
        }
    }
}
