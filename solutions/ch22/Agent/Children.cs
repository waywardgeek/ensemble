using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Agent
{
    public partial Agent(Ensemble parent, Agent owner, string childID,
        Configuration configuration, string? outputDirectory, SkillSpec? skillSpec,
        System.Text.Json.JsonElement? submitSchema, string? dataDirectory)
    {
        // Schema admission precedes ownership, directory creation and other effects.
        submissionContract = new SubmissionContract(submitSchema);
        ArgumentNullException.ThrowIfNull(parent);
        ArgumentNullException.ThrowIfNull(owner);
        ArgumentNullException.ThrowIfNull(configuration);
        ValidateChildID(childID);
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
            string? home = dataDirectory;
            if (home is null && owner.dataDirectory is not null)
            {
                home = Path.Combine(owner.dataDirectory, "children", childID);
            }
            ConstructOwners(parent, owner, outputDirectory, skillSpec, home);
        }
    }

    public partial Agent? AgentParent() => agentParent;
    internal static partial void ValidateChildID(string childID)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(childID);
        // The same local syntax governs construction and durable child lookup.
        // An address of another branch can never become a filesystem path here.
        if (!childID.All(c => char.IsAsciiLetterOrDigit(c) || c is '-' or '_'))
        {
            throw new ArgumentException("Child ID must contain only ASCII letters, digits, '-' or '_'.", nameof(childID));
        }
    }

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
