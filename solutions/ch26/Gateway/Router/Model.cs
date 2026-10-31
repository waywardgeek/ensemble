using System.Collections.Immutable;

namespace Ensemble.Gateway;

internal sealed partial class RouterGroup
{
    internal partial RouterGroup(Router parent, string name, GroupDefinition definition, bool runtime)
    {
        Router.ValidateName(name);
        Router.ValidateGroup(definition);
        if (name.StartsWith("runtime:", StringComparison.Ordinal) != runtime)
        {
            throw new ArgumentException("Runtime groups require the runtime: namespace.");
        }
        this.name = name;
        this.definition = definition;
        this.runtime = runtime;
        parent.AddGroup(this);
    }

    internal partial Router Parent() => parent ?? throw new ObjectDisposedException(nameof(RouterGroup));
    internal partial GroupDefinition Definition()
    {
        Router owner = Parent();
        lock (owner.gate) { owner.RequireGroup(this); return definition; }
    }

    internal partial void Update(GroupDefinition value)
    {
        Router.ValidateGroup(value);
        Router owner = Parent();
        lock (owner.gate)
        {
            owner.RequireGroup(this);
            owner.RequireLive();
            if (!runtime) throw new InvalidOperationException("Config groups are immutable.");
            definition = value;
        }
    }

    internal partial void AddMember(RouterIdentity identity, GroupLink kind)
    {
        Router owner = Parent();
        lock (owner.gate)
        {
            owner.RequireGroup(this);
            owner.RequireLive();
            owner.RequireIdentity(identity);
            if (!Enum.IsDefined(kind)) throw new ArgumentOutOfRangeException(nameof(kind));
            if (!members.TryGetValue(identity, out HashSet<GroupLink>? links))
            {
                links = [];
                members.Add(identity, links);
            }
            // Both parent operations are symmetric and idempotent. The second
            // call sees its already-added link and terminates the recursion. No
            // observer can see the intermediate half-link under the root gate.
            if (links.Add(kind)) identity.AddGroup(this, kind);
        }
    }

    internal partial void RemoveMember(RouterIdentity identity, GroupLink kind)
    {
        Router owner = Parent();
        lock (owner.gate)
        {
            owner.RequireGroup(this);
            owner.RequireIdentity(identity);
            if (!Enum.IsDefined(kind)) throw new ArgumentOutOfRangeException(nameof(kind));
            if (!members.TryGetValue(identity, out HashSet<GroupLink>? links) || !links.Remove(kind)) return;
            if (links.Count == 0) members.Remove(identity);
            identity.RemoveGroup(this, kind);
        }
    }

    internal partial bool HasMember(RouterIdentity identity, GroupLink kind)
    {
        Router owner = Parent();
        lock (owner.gate) { owner.RequireGroup(this); return members.TryGetValue(identity, out var links) && links.Contains(kind); }
    }

    internal partial bool IsEmpty()
    {
        Router owner = Parent();
        lock (owner.gate) { owner.RequireGroup(this); return members.Count == 0; }
    }

    internal partial void Destroy()
    {
        Router? owner = parent;
        if (owner is null) return;
        lock (owner.gate)
        {
            if (parent is null) return;
            foreach (var (identity, links) in members.ToArray())
            {
                foreach (GroupLink kind in links.ToArray()) RemoveMember(identity, kind);
            }
            owner.RemoveGroup(this);
        }
    }
}

internal sealed partial class RouterIdentity
{
    internal partial RouterIdentity(Router parent, string id, ImmutableArray<byte> publicKey,
        string displayName, bool configured)
    {
        Router.ValidateName(id);
        if (publicKey.IsDefault || publicKey.Length != 32) throw new ArgumentException("Invalid public key.");
        ArgumentNullException.ThrowIfNull(displayName);
        this.id = id;
        this.publicKey = publicKey;
        this.displayName = displayName;
        this.configured = configured;
        parent.AddIdentity(this);
    }

    internal partial Router Parent() => parent ?? throw new ObjectDisposedException(nameof(RouterIdentity));
    private partial List<RouterGroup> Links(GroupLink kind) => kind switch
    {
        GroupLink.Base => bases,
        GroupLink.Grant => grants,
        GroupLink.Contact => contacts,
        _ => throw new ArgumentOutOfRangeException(nameof(kind))
    };

    internal partial RouterGroup[] Groups(GroupLink kind)
    {
        Router owner = Parent();
        lock (owner.gate) { return Links(kind).ToArray(); }
    }

    internal partial RouterGroup[] Memberships()
    {
        Router owner = Parent();
        lock (owner.gate) { return bases.Concat(grants).Distinct().ToArray(); }
    }

    internal partial bool Capable(string capability)
    {
        Router owner = Parent();
        lock (owner.gate)
        {
            return Memberships().Any(group => group.Definition().Capabilities.Contains(capability));
        }
    }

    internal partial void AddGroup(RouterGroup group, GroupLink kind)
    {
        Router owner = Parent();
        lock (owner.gate)
        {
            owner.RequireLive();
            owner.RequireIdentity(this);
            if (group.Parent() != owner) throw new InvalidOperationException("Group belongs to another router.");
            List<RouterGroup> links = Links(kind);
            if (links.Contains(group)) return;
            links.Add(group);
            group.AddMember(this, kind);
        }
    }

    internal partial void RemoveGroup(RouterGroup group, GroupLink kind)
    {
        Router owner = Parent();
        lock (owner.gate)
        {
            owner.RequireIdentity(this);
            if (group.Parent() != owner) throw new InvalidOperationException("Group belongs to another router.");
            if (Links(kind).Remove(group)) group.RemoveMember(this, kind);
        }
    }

    internal partial void ReplaceGroups(GroupLink kind, IReadOnlyList<RouterGroup> replacement)
    {
        Router owner = Parent();
        lock (owner.gate)
        {
            owner.RequireLive();
            owner.RequireIdentity(this);
            _ = Links(kind);
            // Validate the whole replacement before changing either endpoint.
            HashSet<RouterGroup> unique = [];
            foreach (RouterGroup group in replacement)
            {
                if (group.Parent() != owner || !unique.Add(group))
                    throw new ArgumentException("Groups must be unique and belong to this router.");
            }
            foreach (RouterGroup old in Links(kind).ToArray()) RemoveGroup(old, kind);
            foreach (RouterGroup group in replacement) AddGroup(group, kind);
        }
    }

    internal partial void AddQueue(DeliveryQueue child)
    {
        Router owner = Parent();
        lock (owner.gate)
        {
            owner.RequireLive();
            owner.RequireIdentity(this);
            if (queue is not null || child.parent is not null)
                throw new InvalidOperationException("Queue already has an owner or identity already has a queue.");
            queue = child;
            child.parent = this;
        }
    }

    internal partial void RemoveQueue(DeliveryQueue child)
    {
        Router owner = Parent();
        lock (owner.gate)
        {
            if (queue != child || child.parent != this) throw new InvalidOperationException("Queue owner mismatch.");
            queue = null;
            child.parent = null;
        }
    }

    internal partial DeliveryQueue Queue() => queue ?? throw new ObjectDisposedException(nameof(DeliveryQueue));

    internal partial void Destroy()
    {
        Router? owner = parent;
        if (owner is null) return;
        lock (owner.gate)
        {
            if (parent is null) return;
            // Router's asynchronous destruction cancels and joins these children
            // first. Refuse a synchronous partial teardown while they are live.
            if (sessions.Count != 0) throw new InvalidOperationException("Join identity sessions before destruction.");
            queue?.Destroy();
            foreach (GroupLink kind in Enum.GetValues<GroupLink>())
            {
                foreach (RouterGroup group in Links(kind).ToArray()) RemoveGroup(group, kind);
            }
            owner.RemoveIdentity(this);
        }
    }
}

internal sealed partial class DeliveryQueue
{
    internal partial DeliveryQueue(RouterIdentity parent, QueueState? saved)
    {
        if (saved is not null)
        {
            if (saved.Ack > saved.Next) throw new InvalidDataException("Receipt exceeds allocated sequence.");
            ulong previous = saved.Ack;
            foreach (Envelope entry in saved.Messages ?? [])
            {
                Envelope? message = entry is null ? null : Wire.NormalizeEnvelope(entry);
                if (message is null || previous == ulong.MaxValue || message.Seq != previous + 1 ||
                    message.Seq > saved.Next || message.ToAgent != parent.id ||
                    message.ID != $"{parent.id}:{message.Seq}" || message.Timestamp == default ||
                    message.Source is not { Type: "agent" } source || string.IsNullOrEmpty(source.UserID) ||
                    string.IsNullOrEmpty(source.ChannelID) || source.Trust is not ("external" or "trusted" or "owner") ||
                    message.Content is null || message.ConversationID is null || message.ResponseMode is null)
                {
                    throw new InvalidDataException("Invalid persisted delivery queue.");
                }
                // Old accepted policy may no longer exist: envelopes are stamped
                // historical values, not live membership links to reconstruct.
                _ = Wire.Encode(new Frame("deliver", Envelope: message));
                messages.Add(message);
                previous = message.Seq;
            }
            if (previous != saved.Next) throw new InvalidDataException("Persisted queue has missing deliveries.");
            next = saved.Next;
            ack = saved.Ack;
        }
        parent.AddQueue(this);
    }

    internal partial RouterIdentity Parent() => parent ?? throw new ObjectDisposedException(nameof(DeliveryQueue));
    private partial void RequireOwned(RouterIdentity identity, Router router)
    {
        // A parent can be detached while this caller waits for the root gate.
        // Recheck the actual relationship after locking, before reading/mutating.
        router.RequireIdentity(identity);
        if (parent != identity || identity.queue != this)
            throw new InvalidOperationException("Queue no longer belongs to this identity.");
    }

    internal partial QueueState Snapshot()
    {
        RouterIdentity identity = Parent();
        Router owner = identity.Parent();
        lock (owner.gate)
        {
            RequireOwned(identity, owner);
            return new(next, ack, messages.ToImmutableArray());
        }
    }

    internal partial Envelope? After(ulong sequence)
    {
        RouterIdentity identity = Parent();
        Router owner = identity.Parent();
        lock (owner.gate)
        {
            RequireOwned(identity, owner);
            return messages.FirstOrDefault(message => message.Seq > sequence);
        }
    }

    internal partial Envelope Enqueue(Source source, OutboundMessage message)
    {
        RouterIdentity identity = Parent();
        Router owner = identity.Parent();
        lock (owner.gate)
        {
            owner.RequireLive();
            RequireOwned(identity, owner);
            ulong sequence = checked(next + 1);
            Envelope envelope = new($"{identity.id}:{sequence}", sequence, identity.id,
                Protocol.ProvenanceHeader(source) + "\n\n" + message.Content,
                message.ConversationID, message.ResponseMode, DateTimeOffset.UtcNow, source);
            _ = Wire.Encode(new Frame("deliver", Envelope: envelope));
            messages.Add(envelope);
            next = sequence;
            return envelope;
        }
    }

    internal partial void Acknowledge(ulong sequence)
    {
        RouterIdentity identity = Parent();
        Router owner = identity.Parent();
        lock (owner.gate)
        {
            owner.RequireLive();
            RequireOwned(identity, owner);
            if (sequence > next) throw new InvalidOperationException("Receipt exceeds allocated sequence.");
            if (sequence <= ack) return;
            messages.RemoveAll(message => message.Seq <= sequence);
            ack = sequence;
        }
    }

    internal partial void Destroy()
    {
        RouterIdentity? owner = parent;
        if (owner is null) return;
        Router? root = owner.parent;
        if (root is null) return;
        lock (root.gate)
        {
            if (parent is null) return;
            owner.RemoveQueue(this);
            messages.Clear();
        }
    }
}
