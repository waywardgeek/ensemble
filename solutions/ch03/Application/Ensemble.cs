using System.Collections.Immutable;
namespace Ensemble;

public sealed partial class Ensemble
{
    public partial Ensemble(TextWriter diagnostics)
    {
        this.diagnostics = diagnostics;
        agents = [];
    }

    public partial ImmutableArray<Agent> Agents() => agents.ToImmutableArray();
    internal partial void AddAgent(Agent child)
    {
        ObjectDisposedException.ThrowIf(destroyed, this);
        if (child.destroyed || child.parent is not null || agents.Contains(child))
        {
            throw new InvalidOperationException("Invalid Agent registration");
        }

        agents.Add(child);
        child.parent = this;
    }

    internal partial void RemoveAgent(Agent child)
    {
        if (!ReferenceEquals(child.parent, this) || !agents.Contains(child))
        {
            throw new InvalidOperationException("Agent is not registered here");
        }

        agents.Remove(child);
        child.parent = null;
    }

    public partial void Log(string message) => diagnostics.WriteLine(message);
    public partial void Destroy()
    {
        if (destroyed)
        {
            return;
        }

        destroyed = true;
        // Children remove themselves from this collection. Enumerate a snapshot.
        List<Exception>? failures = null;
        foreach (var child in agents.ToArray())
        {
            try
            {
                child.Destroy();
            }
            catch (Exception ex)
            {
                failures ??= [];
                failures.Add(ex);
            }
        }
        // Complete every ownership cascade even if one log cannot close cleanly.
        if (failures is not null)
        {
            throw new AggregateException("Agent cleanup failed", failures);
        }
    }
}
