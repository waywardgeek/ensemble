using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Ensemble
{
    public partial Ensemble(TextWriter diagnostics)
    {
        ArgumentNullException.ThrowIfNull(diagnostics);
        this.diagnostics = diagnostics;
        agents = new List<Agent>();
    }

    public partial ImmutableArray<Agent> Agents()
    {
        return agents.ToImmutableArray();
    }

    // Validate before changing either end. No caller has a half-relationship
    // to finish, and a refused operation leaves the previous ownership intact.
    internal partial void AddAgent(Agent child)
    {
        ObjectDisposedException.ThrowIf(destroyed, this);
        if (child.destroyed || child.parent is not null || agents.Contains(child))
        {
            throw new InvalidOperationException("Agent is destroyed or already attached.");
        }

        agents.Add(child);
        child.parent = this;
    }

    internal partial void RemoveAgent(Agent child)
    {
        if (child.parent != this || !agents.Contains(child))
        {
            throw new InvalidOperationException("Agent does not belong to this Ensemble.");
        }

        agents.Remove(child);
        child.parent = null;
    }

    public partial void Log(string safeMessage)
    {
        diagnostics.WriteLine(safeMessage);
    }

    public partial void Destroy()
    {
        if (destroyed)
        {
            return;
        }

        destroyed = true;
        // Each child removes itself, so iterate a snapshot, not the live list.
        foreach (Agent child in agents.ToArray())
        {
            child.Destroy();
        }
    }
}
