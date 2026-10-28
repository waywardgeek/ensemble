using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Agent
{
    public partial Agent(Ensemble parent, Agent owner, string childID,
        Configuration configuration, string? outputDirectory, SkillSpec? skillSpec,
        System.Text.Json.JsonElement? submitSchema)
    {
        // Schema admission precedes ownership, directory creation and other effects.
        submissionContract = new SubmissionContract(submitSchema);
        ArgumentNullException.ThrowIfNull(parent);
        ArgumentNullException.ThrowIfNull(owner);
        ArgumentNullException.ThrowIfNull(configuration);
        ArgumentException.ThrowIfNullOrWhiteSpace(childID);
        // IDs are addresses, never paths. Restrict their syntax before any file
        // or relationship is created so later per-child storage can use them.
        if (!childID.All(c => char.IsAsciiLetterOrDigit(c) || c is '-' or '_'))
        {
            throw new ArgumentException("Child ID must contain only ASCII letters, digits, '-' or '_'.", nameof(childID));
        }
        this.childID = childID;
        var derived = owner.Sandbox().DeriveChild(configuration);
        if (derived.Refusal is not null)
        {
            throw new ArgumentException(derived.Refusal, nameof(configuration));
        }
        lock (owner.ingress)
        {
            ObjectDisposedException.ThrowIf(owner.destroyed || owner.destroying || owner.admissionClosed, owner);
            if (!ReferenceEquals(owner.Parent(), parent))
            {
                throw new InvalidOperationException("Both Agent parents must belong to the same Ensemble.");
            }
            this.configuration = derived.Effective;
            nextTurnToolLimit = this.configuration.ToolRoundLimit;
            // Registration and construction hold the owning Agent's admission
            // gate: shutdown cannot miss a child whose construction has begun.
            ConstructOwners(parent, owner, outputDirectory, skillSpec);
        }
    }

    public partial Agent? AgentParent() => agentParent;
    public partial string? ChildID() => childID;

    public partial ImmutableArray<Agent> Children()
    {
        lock (ingress)
        {
            return children.ToImmutableArray();
        }
    }

    public partial Agent? FindChild(string id)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(id);
        lock (ingress)
        {
            return children.FirstOrDefault(child => child.childID == id);
        }
    }

    internal partial void AddChild(Agent child)
    {
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed || destroying || admissionClosed, this);
            if (child.destroyed || child.agentParent is not null || child.childID is null
                || !ReferenceEquals(child.parent, parent) || children.Contains(child)
                || children.Any(existing => existing.childID == child.childID))
            {
                throw new InvalidOperationException("Invalid or duplicate child Agent registration.");
            }
            children.Add(child);
            child.agentParent = this;
        }
    }

    internal partial void RemoveChild(Agent child)
    {
        lock (ingress)
        {
            if (!ReferenceEquals(child.agentParent, this) || !children.Contains(child))
            {
                throw new InvalidOperationException("Child Agent is not registered here.");
            }
            children.Remove(child);
            child.agentParent = null;
        }
    }

    private partial void DestroyChildren()
    {
        List<Exception>? failures = null;
        foreach (Agent child in Children())
        {
            try
            {
                child.Destroy();
            }
            catch (Exception error)
            {
                failures ??= [];
                failures.Add(error);
            }
        }
        if (failures is not null)
        {
            throw new AggregateException("Child Agent cleanup failed.", failures);
        }
    }
}
