using System.Collections.Immutable;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace Ensemble.Gateway;

public sealed partial class Router
{
    private partial Router(RouterConfig config, Keypair key, StartupMode mode)
    {
        ArgumentNullException.ThrowIfNull(config);
        ArgumentNullException.ThrowIfNull(config.Gateway);
        ArgumentNullException.ThrowIfNull(key);
        if (string.IsNullOrWhiteSpace(config.Gateway.DataDirectory))
            throw new ArgumentException("Gateway requires an explicit data directory.");
        directory = Path.GetFullPath(config.Gateway.DataDirectory);
        this.key = key;
        try
        {
            AcquireStateLease();
            string primary = Path.Combine(directory, "gateway-state.json");
            string backup = Path.Combine(directory, "gateway-state.backup.json");
            if (mode == StartupMode.RecoverBackup)
            {
                GatewayState restored = ReadState(backup) ?? throw new InvalidDataException("No Gateway backup exists.");
                ValidateBootstrap(restored, config);
                // Copy, never rename away, the damaged primary. Failure leaves it
                // in place. No corrupt primary is promoted into the backup slot.
                if (File.Exists(primary))
                    File.Copy(primary, Path.Combine(directory, "gateway-state.recovery-" + Guid.NewGuid().ToString("N") + ".json"));
                ReplaceFile(primary, restored);
            }
            GatewayState? saved = ReadState(primary);
            if (saved is not null)
            {
                if (mode == StartupMode.MigrateLegacy)
                    throw new InvalidOperationException("Complete Gateway state already exists; migration refused.");
                ValidateBootstrap(saved, config);
                Restore(saved);
                revision = saved.Revision;
            }
            else
            {
                if (ReadState(backup) is not null)
                    throw new InvalidDataException("Primary Gateway state is missing; use explicit backup recovery.");
                bool legacy = Path.Exists(Path.Combine(directory, "state.json")) ||
                    Path.Exists(Path.Combine(directory, "queues")) ||
                    new FileInfo(Path.Combine(directory, "state.json")).LinkTarget is not null;
                if (legacy != (mode == StartupMode.MigrateLegacy))
                    throw new InvalidDataException(legacy
                        ? "Legacy Gateway state requires explicit migration."
                        : "No legacy Gateway state exists to migrate.");
                // This loader validates legacy relationships and queue continuity.
                // It reads originals only; the complete file has a distinct name.
                LoadLegacy(config);
                GatewayState initial = Snapshot() with { Revision = 1 };
                ValidateState(initial);
                ValidateBootstrap(initial, config);
                ReplaceFile(backup, initial);
                ReplaceFile(primary, initial);
                revision = 1;
            }
        }
        catch
        {
            try { DestroyModel(); }
            finally { stateLease?.Dispose(); stateLease = null; disposed = true; }
            throw;
        }
    }

    public static partial Router MigrateLegacy(RouterConfig config, Keypair key) => new(config, key, StartupMode.MigrateLegacy);
    public static partial Router RecoverBackup(RouterConfig config, Keypair key) => new(config, key, StartupMode.RecoverBackup);

    private partial void AcquireStateLease()
    {
        if (OperatingSystem.IsWindows() || (!OperatingSystem.IsMacOS() && !OperatingSystem.IsLinux()))
            throw new PlatformNotSupportedException("Gateway persistence supports macOS/Linux.");
        Directory.CreateDirectory(directory, UnixFileMode.UserRead | UnixFileMode.UserWrite | UnixFileMode.UserExecute);
        // Hold the handle until joined teardown: another process must not load a
        // stale generation and replace this writer's complete state.
        stateLease = new FileStream(Path.Combine(directory, "gateway-state.lock"), new FileStreamOptions
        {
            Mode = FileMode.OpenOrCreate, Access = FileAccess.ReadWrite, Share = FileShare.None,
            UnixCreateMode = UnixFileMode.UserRead | UnixFileMode.UserWrite
        });
    }

    private partial GatewayState? ReadState(string path)
    {
        // ReadOptional detects duplicates, null documents and unavailable paths.
        using JsonDocument? document = ReadOptional<JsonDocument>(path);
        if (document is null) return null;
        GatewayState value = document.RootElement.Deserialize<GatewayState>(new JsonSerializerOptions
        {
            UnmappedMemberHandling = JsonUnmappedMemberHandling.Disallow
        }) ?? throw new InvalidDataException("Null Gateway state.");
        ValidateState(value);
        return value;
    }

    private partial void ValidateBootstrap(GatewayState state, RouterConfig config)
    {
        bool sameGroups = JsonElement.DeepEquals(JsonSerializer.SerializeToElement(state.BootstrapGroups), JsonSerializer.SerializeToElement(config.Groups));
        bool sameAgents = JsonElement.DeepEquals(JsonSerializer.SerializeToElement(state.BootstrapAgents), JsonSerializer.SerializeToElement(config.Agents));
        if (!sameGroups || !sameAgents || state.GatewayKey != Keys.EncodePublicKey(key.Public.AsSpan()))
            throw new InvalidDataException("Gateway bootstrap/key differs from durable state; implicit authority migration refused.");
    }

    private partial GatewayState Snapshot()
    {
        IdentityDefinition Definition(RouterIdentity identity) => new(
            Keys.EncodePublicKey(identity.publicKey.AsSpan()),
            identity.Groups(GroupLink.Base).Select(group => group.name).ToImmutableArray(), identity.displayName);
        RouterState policy = new(
            groups.Values.Where(group => group.runtime).ToImmutableDictionary(group => group.name, group => group.Definition()),
            identities.Values.Where(identity => !identity.configured).ToImmutableDictionary(identity => identity.id, Definition),
            identities.Values.ToImmutableDictionary(identity => identity.id, identity => identity.Groups(GroupLink.Grant).Select(group => group.name).ToImmutableArray()),
            identities.Values.ToImmutableDictionary(identity => identity.id, identity => identity.Groups(GroupLink.Contact).Select(group => group.name).ToImmutableArray()));
        return new(1, revision, Keys.EncodePublicKey(key.Public.AsSpan()),
            groups.Values.Where(group => !group.runtime).ToImmutableDictionary(group => group.name, group => group.Definition()),
            identities.Values.Where(identity => identity.configured).ToImmutableDictionary(identity => identity.id, Definition),
            policy, identities.Values.ToImmutableDictionary(identity => identity.id, identity => identity.Queue().Snapshot()));
    }

    private static partial void ValidateState(GatewayState state)
    {
        if (state.Version != 1 || state.Revision < 1 || state.BootstrapGroups is null || state.BootstrapAgents is null ||
            state.Policy is null || state.Policy.Groups is null || state.Policy.Children is null ||
            state.Policy.Memberships is null || state.Policy.Contactable is null || state.Queues is null)
            throw new InvalidDataException("Invalid or unsupported complete Gateway state.");
        _ = Keys.ParsePublicKey(state.GatewayKey);
        var allGroups = new HashSet<string>(StringComparer.Ordinal);
        void Groups(ImmutableDictionary<string, GroupDefinition> values, bool runtime)
        {
            foreach (var (name, definition) in values)
            {
                ValidateName(name);
                ValidateGroup(definition);
                if (name.StartsWith("runtime:", StringComparison.Ordinal) != runtime || !allGroups.Add(name))
                    throw new InvalidDataException("Invalid or duplicate persisted group.");
            }
        }
        Groups(state.BootstrapGroups, false);
        Groups(state.Policy.Groups, true);
        void Links(ImmutableArray<string> links, bool runtimeOnly, bool bootstrapOnly = false)
        {
            if (links.IsDefault || links.Distinct(StringComparer.Ordinal).Count() != links.Length || links.Any(name =>
                !allGroups.Contains(name) || (runtimeOnly && !state.Policy.Groups.ContainsKey(name)) ||
                (bootstrapOnly && !state.BootstrapGroups.ContainsKey(name))))
                throw new InvalidDataException("Invalid persisted group links.");
        }
        HashSet<string> ids = new(StringComparer.Ordinal);
        void Identities(ImmutableDictionary<string, IdentityDefinition> values, bool configured)
        {
            foreach (var (id, definition) in values)
            {
                ValidateName(id);
                if (!ids.Add(id) || definition is null || definition.DisplayName is null)
                    throw new InvalidDataException("Invalid or duplicate persisted identity.");
                _ = Keys.ParsePublicKey(definition.PublicKey);
                Links(definition.Groups, false, configured);
            }
        }
        Identities(state.BootstrapAgents, true);
        Identities(state.Policy.Children, false);
        if (!ids.SetEquals(state.Queues.Keys) || !ids.SetEquals(state.Policy.Memberships.Keys) || !ids.SetEquals(state.Policy.Contactable.Keys))
            throw new InvalidDataException("Complete state must contain exactly one queue and link set per identity.");
        foreach (string id in ids)
        {
            Links(state.Policy.Memberships[id], true);
            Links(state.Policy.Contactable[id], false);
            QueueState queue = state.Queues[id];
            if (queue is null || queue.Messages is null) throw new InvalidDataException("Complete queue messages must be an array.");
            _ = DeliveryQueue.ValidateState(id, queue);
        }
    }

    private partial void Restore(GatewayState state)
    {
        foreach (var (name, definition) in state.BootstrapGroups) _ = new RouterGroup(this, name, definition, false);
        foreach (var (name, definition) in state.Policy.Groups) _ = new RouterGroup(this, name, definition, true);
        void Identity(string id, IdentityDefinition definition, bool configured)
        {
            RouterIdentity identity = new(this, id, Keys.ParsePublicKey(definition.PublicKey), definition.DisplayName, configured);
            identity.ReplaceGroups(GroupLink.Base, ResolveGroups(definition.Groups));
            identity.ReplaceGroups(GroupLink.Grant, ResolveGroups(state.Policy.Memberships[id], runtimeOnly: true));
            identity.ReplaceGroups(GroupLink.Contact, ResolveGroups(state.Policy.Contactable[id]));
            _ = new DeliveryQueue(identity, state.Queues[id]);
        }
        foreach (var (id, definition) in state.BootstrapAgents) Identity(id, definition, true);
        foreach (var (id, definition) in state.Policy.Children) Identity(id, definition, false);
    }

    private partial void Commit(GatewayState candidate, Action publish)
    {
        RequireLive();
        candidate = candidate with { Revision = checked(revision + 1) };
        ValidateState(candidate);
        try
        {
            string primary = Path.Combine(directory, "gateway-state.json");
            GatewayState previous = ReadState(primary) ?? throw new InvalidDataException("Durable Gateway state disappeared.");
            if (previous.Revision != revision) throw new InvalidDataException("Gateway state revision changed outside its owner.");
            ReplaceFile(Path.Combine(directory, "gateway-state.backup.json"), previous);
            ReplaceFile(primary, candidate);
            // Both persistence and publication are within the root gate. An
            // unexpected apply failure is terminal too; never overwrite the
            // committed candidate using the incomplete live graph at teardown.
            publish();
            revision = candidate.Revision;
        }
        catch
        {
            persistenceFailed = true;
            throw;
        }
    }

    internal partial void CommitQueue(RouterIdentity identity, QueueState candidate, Action publish)
    {
        RequireIdentity(identity);
        GatewayState state = Snapshot();
        Commit(state with { Queues = state.Queues.SetItem(identity.id, candidate) }, publish);
    }
}
