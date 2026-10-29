using System.Collections.Immutable;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Runtime.CompilerServices;

[assembly: InternalsVisibleTo("RouterTests")]

namespace Ensemble.Gateway;

// Configuration and disk snapshots are immutable values, not a second mutable
// policy owner. The live graph below replaces them after validation on load.
public sealed record GroupDefinition(
    [property: JsonPropertyName("capabilities"), JsonRequired] ImmutableArray<string> Capabilities,
    [property: JsonPropertyName("trust"), JsonRequired] string Trust,
    [property: JsonPropertyName("description")] string Description = "");
public sealed record IdentityDefinition(
    [property: JsonPropertyName("public_key"), JsonRequired] string PublicKey,
    [property: JsonPropertyName("groups"), JsonRequired] ImmutableArray<string> Groups,
    [property: JsonPropertyName("display_name")] string DisplayName = "");
public sealed record RouterEndpoint(
    [property: JsonPropertyName("listen"), JsonRequired] string Listen,
    [property: JsonPropertyName("data_dir"), JsonRequired] string DataDirectory);
public sealed record RouterConfig(
    [property: JsonPropertyName("gateway"), JsonRequired] RouterEndpoint Gateway,
    [property: JsonPropertyName("groups"), JsonRequired] ImmutableDictionary<string, GroupDefinition> Groups,
    [property: JsonPropertyName("agents"), JsonRequired] ImmutableDictionary<string, IdentityDefinition> Agents);
internal sealed record RouterState(
    [property: JsonPropertyName("groups"), JsonRequired] ImmutableDictionary<string, GroupDefinition> Groups,
    [property: JsonPropertyName("children"), JsonRequired] ImmutableDictionary<string, IdentityDefinition> Children,
    [property: JsonPropertyName("memberships"), JsonRequired] ImmutableDictionary<string, ImmutableArray<string>> Memberships,
    [property: JsonPropertyName("contactable"), JsonRequired] ImmutableDictionary<string, ImmutableArray<string>> Contactable);
internal sealed record QueueState(
    [property: JsonPropertyName("next"), JsonRequired] ulong Next,
    [property: JsonPropertyName("ack"), JsonRequired] ulong Ack,
    [property: JsonPropertyName("messages"), JsonRequired] ImmutableArray<Envelope>? Messages);

/// <summary>
/// Standalone routing root. It owns policy identities/groups, and each identity
/// owns its delivery queue. It owns no Ensemble Agent and performs no inference.
/// One gate covers authorization and admission; network I/O must stay outside it.
/// Snapshots are periodic/orderly-restart protection, not a multi-file transaction.
/// </summary>
public sealed partial class Router : IAsyncDisposable
{
    internal readonly object gate = new();
    private readonly string directory;
    private readonly Keypair key;
    private readonly Dictionary<string, RouterGroup> groups = new(StringComparer.Ordinal);
    private readonly Dictionary<string, RouterIdentity> identities = new(StringComparer.Ordinal);
    private bool stopped;
    private bool disposed;

    /// <summary>Validate bootstrap and restore state before admitting any work.
    /// A broken existing file refuses startup; it is never replaced with defaults.</summary>
    public partial Router(RouterConfig config, Keypair key);
    public static partial RouterConfig ReadConfig(string path);
    /// <summary>Individually replace private queue/policy files; retain receipts
    /// even when all messages have been accepted. The host schedules snapshots.</summary>
    public partial void Save();
    /// <summary>End admission, cancel and join sessions, save, then detach all
    /// owned model objects even if saving fails. Concurrent callers share the
    /// close result. Durable files and caller-owned immutable keys are retained.</summary>
    public partial ValueTask DisposeAsync();

    internal partial RouterIdentity Identity(string id);
    internal partial RouterGroup Group(string name);
    internal partial void AddIdentity(RouterIdentity identity);
    internal partial void RemoveIdentity(RouterIdentity identity);
    internal partial void AddGroup(RouterGroup group);
    internal partial void RemoveGroup(RouterGroup group);
    internal partial void RequireLive();
    internal partial void RequireIdentity(RouterIdentity identity);
    internal partial void RequireGroup(RouterGroup group);
    internal static partial void ValidateName(string name);
    internal static partial void ValidateGroup(GroupDefinition definition);
    internal partial RpcMessage? Request(RouterIdentity caller, RpcMessage request);
    internal partial Envelope Route(RouterIdentity caller, OutboundMessage message);
    private partial object Call(RouterIdentity caller, string name, JsonElement args);
    private partial object Admin(string name, JsonElement args);
    private partial RouterGroup[] ResolveGroups(ImmutableArray<string> names, bool runtimeOnly = false);
    private partial void Load(RouterConfig config);
    private partial void SaveLocked();
    private partial void DestroyModel();
    private static partial T? ReadOptional<T>(string path) where T : class;
    private static partial void ReplaceFile<T>(string path, T value);
    private static partial JsonElement Schemas();
    private static partial string Text(JsonElement args, string name, bool optional = false);
    private static partial ImmutableArray<string> Names(JsonElement args, string name);
    private static partial void Arguments(JsonElement args, params string[] allowed);
}

// Base memberships belong to bootstrap or an attenuated child registration.
// Runtime grants can be replaced independently. Contact links authorize incoming
// senders; they do not confer the group's capabilities on the recipient.
internal enum GroupLink { Base, Grant, Contact }

/// <summary>Router-owned group with weak reverse links to identities. Destroying
/// a group detaches those links; it does not destroy the identities or queues.</summary>
internal sealed partial class RouterGroup
{
    internal Router? parent;
    internal readonly string name;
    internal readonly bool runtime;
    private GroupDefinition definition;
    private readonly Dictionary<RouterIdentity, HashSet<GroupLink>> members = new();
    internal partial RouterGroup(Router parent, string name, GroupDefinition definition, bool runtime);
    internal partial Router Parent();
    internal partial GroupDefinition Definition();
    internal partial void Update(GroupDefinition value);
    internal partial void AddMember(RouterIdentity identity, GroupLink kind);
    internal partial void RemoveMember(RouterIdentity identity, GroupLink kind);
    internal partial bool HasMember(RouterIdentity identity, GroupLink kind);
    internal partial bool IsEmpty();
    internal partial void Destroy();
}

/// <summary>A public authenticated identity, owned by Router, not by its original
/// registrar. Registration attenuates authority once; it does not imply lifetime
/// ownership or later inheritance. Group links are weak in both directions.</summary>
internal sealed partial class RouterIdentity
{
    internal Router? parent;
    internal readonly string id;
    internal readonly bool configured;
    internal readonly ImmutableArray<byte> publicKey;
    internal readonly string displayName;
    private readonly List<RouterGroup> bases = [];
    private readonly List<RouterGroup> grants = [];
    private readonly List<RouterGroup> contacts = [];
    internal DeliveryQueue? queue;
    internal partial RouterIdentity(Router parent, string id, ImmutableArray<byte> publicKey,
        string displayName, bool configured);
    internal partial Router Parent();
    internal partial RouterGroup[] Groups(GroupLink kind);
    internal partial RouterGroup[] Memberships();
    internal partial bool Capable(string capability);
    internal partial void AddGroup(RouterGroup group, GroupLink kind);
    internal partial void RemoveGroup(RouterGroup group, GroupLink kind);
    internal partial void ReplaceGroups(GroupLink kind, IReadOnlyList<RouterGroup> replacement);
    internal partial void AddQueue(DeliveryQueue child);
    internal partial void RemoveQueue(DeliveryQueue child);
    internal partial DeliveryQueue Queue();
    internal partial void Destroy();
    private partial List<RouterGroup> Links(GroupLink kind);
}

/// <summary>Identity-owned ordered at-least-once delivery state. Envelopes are
/// immutable values. Receipts mean handler acceptance, never model completion.</summary>
internal sealed partial class DeliveryQueue
{
    internal RouterIdentity? parent;
    private ulong next;
    private ulong ack;
    private readonly List<Envelope> messages = [];
    internal partial DeliveryQueue(RouterIdentity parent, QueueState? saved = null);
    internal partial RouterIdentity Parent();
    private partial void RequireOwned(RouterIdentity identity, Router router);
    internal partial QueueState Snapshot();
    internal partial Envelope? After(ulong sequence);
    internal partial Envelope Enqueue(Source source, OutboundMessage message);
    internal partial void Acknowledge(ulong sequence);
    internal partial void Destroy();
}
