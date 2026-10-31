using System.Collections.Immutable;
using System.Text.Json;
using System.Text.RegularExpressions;

namespace Ensemble.Gateway;

public sealed partial class Router
{
    public partial Router(RouterConfig config, Keypair key)
    {
        ArgumentNullException.ThrowIfNull(config);
        ArgumentNullException.ThrowIfNull(config.Gateway);
        ArgumentNullException.ThrowIfNull(key);
        if (string.IsNullOrWhiteSpace(config.Gateway.DataDirectory))
            throw new ArgumentException("Gateway requires an explicit data directory.");
        directory = Path.GetFullPath(config.Gateway.DataDirectory);
        this.key = key;
        try { Load(config); }
        catch
        {
            DestroyModel();
            disposed = true;
            throw;
        }
    }

    public static partial RouterConfig ReadConfig(string path) =>
        ReadOptional<RouterConfig>(path) ?? throw new FileNotFoundException("Gateway configuration is missing.", path);

    internal static partial void ValidateName(string name)
    {
        if (name is null || !Regex.IsMatch(name, @"\A[A-Za-z0-9][A-Za-z0-9_.:-]*\z", RegexOptions.CultureInvariant))
            throw new ArgumentException("Invalid Gateway identity/group name.");
    }

    internal static partial void ValidateGroup(GroupDefinition definition)
    {
        ArgumentNullException.ThrowIfNull(definition);
        if (definition.Trust is not ("external" or "trusted" or "owner") || definition.Description is null ||
            definition.Capabilities.IsDefault || definition.Capabilities.Any(capability =>
                capability is not ("send-message" or "declare-contactable" or "group-admin")))
        {
            throw new ArgumentException("Invalid group trust or capabilities.");
        }
    }

    internal partial void RequireLive()
    {
        if (stopped || disposed) throw new ObjectDisposedException(nameof(Router));
    }

    internal partial void RequireIdentity(RouterIdentity identity)
    {
        if (identity.parent != this || !identities.TryGetValue(identity.id, out var found) || found != identity)
            throw new InvalidOperationException("Identity does not belong to this router.");
    }

    internal partial void RequireGroup(RouterGroup group)
    {
        if (group.parent != this || !groups.TryGetValue(group.name, out var found) || found != group)
            throw new InvalidOperationException("Group does not belong to this router.");
    }

    internal partial RouterIdentity Identity(string id)
    {
        lock (gate)
        {
            RequireLive();
            return identities.TryGetValue(id, out var identity) ? identity : throw new ArgumentException("Unknown agent.");
        }
    }

    internal partial RouterGroup Group(string name)
    {
        lock (gate)
        {
            RequireLive();
            return groups.TryGetValue(name, out var group) ? group : throw new ArgumentException("Unknown group.");
        }
    }

    internal partial void AddIdentity(RouterIdentity identity)
    {
        lock (gate)
        {
            RequireLive();
            if (identity.parent is not null || identities.ContainsKey(identity.id))
                throw new InvalidOperationException("Identity is already registered or owned.");
            identities.Add(identity.id, identity);
            identity.parent = this;
        }
    }

    internal partial void RemoveIdentity(RouterIdentity identity)
    {
        lock (gate)
        {
            RequireIdentity(identity);
            if (identity.queue is not null || identity.Sessions().Length != 0 ||
                Enum.GetValues<GroupLink>().Any(kind => identity.Groups(kind).Length != 0))
                throw new InvalidOperationException("Detach the identity's children and group links first.");
            identities.Remove(identity.id);
            identity.parent = null;
        }
    }

    internal partial void AddGroup(RouterGroup group)
    {
        lock (gate)
        {
            RequireLive();
            if (group.parent is not null || groups.ContainsKey(group.name))
                throw new InvalidOperationException("Group is already registered or owned.");
            groups.Add(group.name, group);
            group.parent = this;
        }
    }

    internal partial void RemoveGroup(RouterGroup group)
    {
        lock (gate)
        {
            if (group.parent != this || !groups.TryGetValue(group.name, out var found) || found != group || !group.IsEmpty())
                throw new InvalidOperationException("Group owner mismatch or remaining members.");
            groups.Remove(group.name);
            group.parent = null;
        }
    }

    private partial RouterGroup[] ResolveGroups(ImmutableArray<string> names, bool runtimeOnly)
    {
        if (names.IsDefault) throw new ArgumentException("A group array is required.");
        HashSet<string> unique = new(StringComparer.Ordinal);
        List<RouterGroup> resolved = [];
        foreach (string name in names)
        {
            ValidateName(name);
            RouterGroup group = Group(name);
            if (runtimeOnly && !group.runtime) throw new ArgumentException("Only runtime groups can be assigned.");
            // Go memberships form a set; preserve first occurrence/order without
            // duplicate links or duplicated effective authority.
            if (unique.Add(name)) resolved.Add(group);
        }
        return resolved.ToArray();
    }

    private partial void Load(RouterConfig config)
    {
        ArgumentNullException.ThrowIfNull(config.Groups);
        ArgumentNullException.ThrowIfNull(config.Agents);
        foreach (var (name, definition) in config.Groups)
            _ = new RouterGroup(this, name, definition, runtime: false);
        RouterState? saved = ReadOptional<RouterState>(Path.Combine(directory, "state.json"));
        if (saved is not null)
        {
            if (saved.Groups is null || saved.Children is null || saved.Memberships is null || saved.Contactable is null)
                throw new InvalidDataException("Runtime policy maps must not be null.");
            foreach (var (name, definition) in saved.Groups)
                _ = new RouterGroup(this, name, definition, runtime: true);
        }
        void LoadIdentity(string id, IdentityDefinition definition, bool configured)
        {
            ArgumentNullException.ThrowIfNull(definition);
            RouterGroup[] membership = ResolveGroups(definition.Groups);
            if (configured && membership.Any(group => group.runtime))
                throw new ArgumentException("Bootstrap memberships must refer to bootstrap groups.");
            RouterIdentity identity = new(this, id, Keys.ParsePublicKey(definition.PublicKey),
                definition.DisplayName ?? "", configured);
            identity.ReplaceGroups(GroupLink.Base, membership);
            // Names exclude separators and leading dots; this is the Go filename,
            // including an unescaped colon in a legitimate child ID.
            QueueState? queue = ReadOptional<QueueState>(Path.Combine(directory, "queues", id + ".json"));
            _ = new DeliveryQueue(identity, queue);
        }
        foreach (var (id, definition) in config.Agents) LoadIdentity(id, definition, configured: true);
        if (saved is null) return;
        foreach (var (id, definition) in saved.Children) LoadIdentity(id, definition, configured: false);
        foreach (var (id, names) in saved.Memberships)
            Identity(id).ReplaceGroups(GroupLink.Grant, ResolveGroups(names, runtimeOnly: true));
        foreach (var (id, names) in saved.Contactable)
            Identity(id).ReplaceGroups(GroupLink.Contact, ResolveGroups(names));
    }

    public partial void Save()
    {
        lock (gate)
        {
            RequireLive();
            SaveLocked();
        }
    }

    private partial void SaveLocked()
    {
        // No fsync or cross-file atomicity is claimed. Queue-first matches the
        // reference's loss window; a snapshot error is reported, never ignored.
        foreach (RouterIdentity identity in identities.Values)
            ReplaceFile(Path.Combine(directory, "queues", identity.id + ".json"), identity.Queue().Snapshot());
        RouterState state = new(
            groups.Values.Where(group => group.runtime).ToImmutableDictionary(group => group.name, group => group.Definition()),
            identities.Values.Where(identity => !identity.configured).ToImmutableDictionary(identity => identity.id,
                identity => new IdentityDefinition(Keys.EncodePublicKey(identity.publicKey.AsSpan()),
                    identity.Groups(GroupLink.Base).Select(group => group.name).ToImmutableArray(), identity.displayName)),
            identities.Values.ToImmutableDictionary(identity => identity.id,
                identity => identity.Groups(GroupLink.Grant).Select(group => group.name).ToImmutableArray()),
            identities.Values.ToImmutableDictionary(identity => identity.id,
                identity => identity.Groups(GroupLink.Contact).Select(group => group.name).ToImmutableArray()));
        ReplaceFile(Path.Combine(directory, "state.json"), state);
    }

    private partial void DestroyModel()
    {
        foreach (RouterIdentity identity in identities.Values.ToArray()) identity.Destroy();
        foreach (RouterGroup group in groups.Values.ToArray()) group.Destroy();
    }

    private static partial T? ReadOptional<T>(string path) where T : class
    {
        path = Path.GetFullPath(path);
        byte[] bytes;
        try { bytes = File.ReadAllBytes(path); }
        catch (IOException error) when (error is FileNotFoundException or DirectoryNotFoundException)
        {
            // ENOTDIR can surface as DirectoryNotFoundException too. A queue
            // directory replaced by a regular file is damage, not an empty queue.
            // Also refuse an existing dangling link instead of treating its target
            // as permission to initialize new state. Other I/O errors propagate.
            for (string? candidate = path; candidate is not null; candidate = Path.GetDirectoryName(candidate))
            {
                FileAttributes attributes;
                try { attributes = File.GetAttributes(candidate); }
                catch (IOException missing) when (missing is FileNotFoundException or DirectoryNotFoundException) { continue; }
                if (candidate == path || (attributes & FileAttributes.Directory) == 0)
                    throw new IOException("Gateway snapshot path exists but is unavailable or is not a directory.", error);
                return null;
            }
            return null;
        }
        using JsonDocument document = JsonDocument.Parse(bytes);
        Wire.UniqueNames(document.RootElement);
        return document.RootElement.Deserialize<T>() ?? throw new JsonException("Null Gateway state.");
    }

    private static partial void ReplaceFile<T>(string path, T value)
    {
        if (OperatingSystem.IsWindows() || (!OperatingSystem.IsMacOS() && !OperatingSystem.IsLinux()))
            throw new PlatformNotSupportedException("Gateway private files currently support macOS/Linux.");
        byte[] bytes = JsonSerializer.SerializeToUtf8Bytes(value);
        string parent = Path.GetDirectoryName(path)!;
        Directory.CreateDirectory(parent, UnixFileMode.UserRead | UnixFileMode.UserWrite | UnixFileMode.UserExecute);
        string temporary = Path.Combine(parent, ".gateway-state-" + Guid.NewGuid().ToString("N"));
        bool owned = false;
        Exception? failure = null;
        try
        {
            using (FileStream stream = new(temporary, new FileStreamOptions
            {
                Mode = FileMode.CreateNew, Access = FileAccess.Write, Share = FileShare.None,
                UnixCreateMode = UnixFileMode.UserRead | UnixFileMode.UserWrite
            }))
            {
                owned = true;
                stream.Write(bytes);
                stream.WriteByte((byte)'\n');
                stream.Flush();
            }
            File.Move(temporary, path, overwrite: true);
            owned = false;
        }
        catch (Exception error) { failure = error; throw; }
        finally
        {
            try { if (owned) File.Delete(temporary); }
            catch when (failure is not null) { /* Preserve the save failure, not a cleanup symptom. */ }
        }
    }
}
