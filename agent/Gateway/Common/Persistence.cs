using System.Collections.Immutable;
using System.Text.Json.Serialization;

namespace Ensemble.Gateway;

// Complete durable values. Bootstrap is included so restarting with a different
// configuration cannot silently reinterpret existing authority or receipts.
internal sealed record GatewayState(
    [property: JsonPropertyName("version"), JsonRequired] int Version,
    [property: JsonPropertyName("revision"), JsonRequired] long Revision,
    [property: JsonPropertyName("gateway_key"), JsonRequired] string GatewayKey,
    [property: JsonPropertyName("bootstrap_groups"), JsonRequired] ImmutableDictionary<string, GroupDefinition> BootstrapGroups,
    [property: JsonPropertyName("bootstrap_agents"), JsonRequired] ImmutableDictionary<string, IdentityDefinition> BootstrapAgents,
    [property: JsonPropertyName("policy"), JsonRequired] RouterState Policy,
    [property: JsonPropertyName("queues"), JsonRequired] ImmutableDictionary<string, QueueState> Queues);

internal enum StartupMode { Open, MigrateLegacy, RecoverBackup }

public sealed partial class Router
{
    private partial Router(RouterConfig config, Keypair key, StartupMode mode);
    /// <summary>Explicitly import legacy policy/queue files without changing them.
    /// Refuse if complete state already exists. The returned Router owns the directory.</summary>
    public static partial Router MigrateLegacy(RouterConfig config, Keypair key);
    /// <summary>Explicitly roll back to a validated backup. Preserve the current
    /// primary as evidence. Recovery may replay previously accepted deliveries.</summary>
    public static partial Router RecoverBackup(RouterConfig config, Keypair key);
    private partial GatewayState Snapshot();
    private partial void Restore(GatewayState state);
    private static partial void ValidateState(GatewayState state);
    private partial void ValidateBootstrap(GatewayState state, RouterConfig config);
    private partial GatewayState? ReadState(string path);
    private partial void Commit(GatewayState candidate, Action publish);
    internal partial void CommitQueue(RouterIdentity identity, QueueState candidate, Action publish);
    private partial void AcquireStateLease();
}

internal sealed partial class DeliveryQueue
{
    internal static partial QueueState ValidateState(string identity, QueueState saved);
}
