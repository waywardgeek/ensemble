using System.Collections.Immutable;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class Ensemble
{
    private partial CancellationTokenSource BeginPeerOperation(Agent source, CancellationToken cancellation)
    {
        ArgumentNullException.ThrowIfNull(source);
        cancellation.ThrowIfCancellationRequested();
        // Check real membership, not a name lookup or a retained defaults pointer.
        lock (agents)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (!ReferenceEquals(source.parent, this) || !agents.Contains(source) || source.ChildID() is not null)
            {
                throw new ArgumentException("Peer defaults must come from a top-level Agent owned here.", nameof(source));
            }
            _ = DataDirectory();
            return CancellationTokenSource.CreateLinkedTokenSource(cancellation, peerLifetime.Token);
        }
    }

    public partial async Task<Agent> CreatePeerAsync(Agent source, string name, string? model,
        string? skill, CancellationToken cancellation, bool waitForBrowser, ModelIdentity? offering)
    {
        using var operation = BeginPeerOperation(source, cancellation);
        await peerOperations.WaitAsync(operation.Token).ConfigureAwait(false);
        Agent? peer = null;
        try
        {
            operation.Token.ThrowIfCancellationRequested();
            _ = AgentDirectory(name);
            var entries = ReadPeerLaunches(source);
            if (entries.Any(entry => entry.Name == name) || FindAgent(name) is not null)
            {
                throw new InvalidOperationException("An Agent with that name already exists.");
            }
            var configuration = source.Configuration();
            var bootstrap = source.SkillManager()?.CaptureBootstrap();
            if (offering is not null && !string.IsNullOrEmpty(model))
            {
                throw new ArgumentException("Choose a qualified offering or a legacy model name, not both.");
            }
            // Resolve in this root; a model name must never stand in for its
            // endpoint and funding. Projection validates the actual binding.
            var selection = offering is null
                ? configuration.WithModel(string.IsNullOrEmpty(model) ? configuration.Model : model)
                : new Configuration(configuration, ResolveModel(offering));
            string selectedSkill = string.IsNullOrEmpty(skill) ? bootstrap?.Primary ?? "" : skill;
            var entry = new PeerLaunch(name, selection.Model, selectedSkill, DateTimeOffset.UtcNow, selection.Offering);
            peer = await ConstructPeerAsync(entry, configuration, bootstrap, operation.Token, waitForBrowser, source.Sandbox()).ConfigureAwait(false);
            operation.Token.ThrowIfCancellationRequested();
            // This replacement is the commit point. Do not throw cancellation
            // afterward and tell the caller a durably committed creation failed.
            SavePeerLaunches(source, entries.Add(entry));
            return peer;
        }
        catch (Exception creationError)
        {
            if (peer is not null)
            {
                try
                {
                    peer.Destroy();
                }
                catch (Exception cleanupError)
                {
                    throw new AggregateException("Peer creation and cleanup failed.", creationError, cleanupError);
                }
            }
            throw;
        }
        finally
        {
            peerOperations.Release();
        }
    }

    public partial async Task<ImmutableArray<Agent>> RestorePeersAsync(Agent source, CancellationToken cancellation, bool waitForBrowser)
    {
        using var operation = BeginPeerOperation(source, cancellation);
        await peerOperations.WaitAsync(operation.Token).ConfigureAwait(false);
        var restored = ImmutableArray.CreateBuilder<Agent>();
        try
        {
            operation.Token.ThrowIfCancellationRequested();
            var entries = ReadPeerLaunches(source);
            // Refuse collisions before any peer has filesystem or startup effects.
            foreach (var entry in entries)
            {
                if (FindAgent(entry.Name) is not null)
                {
                    throw new InvalidOperationException("A registry peer is already running: " + entry.Name);
                }
            }
            var configuration = source.Configuration();
            // Resolve the entire launch set before any peer startup or file writes.
            foreach (var entry in entries)
            {
                _ = PeerConfiguration(entry, configuration, source.Sandbox());
            }
            var bootstrap = source.SkillManager()?.CaptureBootstrap();
            foreach (var entry in entries)
            {
                restored.Add(await ConstructPeerAsync(entry, configuration, bootstrap, operation.Token, waitForBrowser, source.Sandbox())
                    .ConfigureAwait(false));
            }
            operation.Token.ThrowIfCancellationRequested();
            return restored.ToImmutable();
        }
        catch (Exception restoreError)
        {
            var failures = new List<Exception> { restoreError };
            foreach (var peer in restored)
            {
                try
                {
                    peer.Destroy();
                }
                catch (Exception cleanupError)
                {
                    failures.Add(cleanupError);
                }
            }
            if (failures.Count > 1)
            {
                throw new AggregateException("Peer restoration and cleanup failed.", failures);
            }
            throw;
        }
        finally
        {
            peerOperations.Release();
        }
    }

    public partial async Task DeletePeerAsync(Agent source, string name, CancellationToken cancellation)
    {
        using var operation = BeginPeerOperation(source, cancellation);
        await peerOperations.WaitAsync(operation.Token).ConfigureAwait(false);
        try
        {
            operation.Token.ThrowIfCancellationRequested();
            _ = AgentDirectory(name);
            var entries = ReadPeerLaunches(source);
            var entry = entries.SingleOrDefault(candidate => candidate.Name == name)
                ?? throw new InvalidOperationException("No persistent peer has that name.");
            var peer = FindAgent(name)
                ?? throw new InvalidOperationException("The persistent peer is not running here.");
            operation.Token.ThrowIfCancellationRequested();
            await peer.ShutdownForPeerDeletionAsync().ConfigureAwait(false);
            // Admission is now closed and cleanup joined. A failed replacement
            // keeps both the old durable entry and the stopped live object; the
            // caller can correct the I/O failure and retry without a phantom success.
            SavePeerLaunches(source, entries.Remove(entry));
            peer.Destroy();
        }
        finally
        {
            peerOperations.Release();
        }
    }

    private partial Configuration PeerConfiguration(PeerLaunch entry, Configuration policy, Sandbox sandbox)
    {
        var saved = AgentSettings.Read(sandbox, Path.Combine(AgentDirectory(entry.Name)!, "settings.json"));
        if (saved?.SelectedOffering is { } selected)
        {
            return new Configuration(policy, ResolveModel(selected));
        }
        if (entry.Offering is { } identity)
        {
            Model model = ResolveModel(identity);
            if (!string.Equals(model.Name, entry.Model, StringComparison.Ordinal))
            {
                throw new InvalidDataException("Peer offering does not match its saved model name.");
            }
            return new Configuration(policy, model);
        }
        if (policy.Offering is not null)
        {
            throw new InvalidDataException("A legacy peer launch cannot identify a configured offering.");
        }
        return policy.WithModel(entry.Model);
    }

    private partial async Task<Agent> ConstructPeerAsync(PeerLaunch entry, Configuration configuration,
        SkillBootstrap? bootstrap, CancellationToken cancellation, bool waitForBrowser, Sandbox sandbox)
    {
        cancellation.ThrowIfCancellationRequested();
        // A primary override selects from immutable parsed definitions. It does
        // not share live activation, grants, variables, or MCP connections.
        if (entry.Skill.Length != 0)
        {
            if (bootstrap is null || !bootstrap.Definitions.TryGetValue(entry.Skill, out var primary)
                || primary.Type != SkillType.Primary)
            {
                throw new InvalidDataException("Peer primary skill is missing or has the wrong type.");
            }
            bootstrap = bootstrap with { Primary = entry.Skill };
        }
        else
        {
            bootstrap = null;
        }
        var peer = new Agent(this, PeerConfiguration(entry, configuration, sandbox), name: entry.Name);
        try
        {
            peer.InstallTools();
            if (bootstrap is not null)
            {
                _ = new Skills(peer, bootstrap);
            }
            var persistence = new PersistenceOptions(Path.Combine(peer.DataDirectory(), "save.json"));
            peer.StartPersistence(persistence);
            if (agentDiagnostics) peer.StartDiagnostics();
            if (peer.SkillManager() is { } skills)
            {
                if (waitForBrowser && skills.RequiresBrowserStartup())
                {
                    // Already owned and addressable, this peer can receive its
                    // own relay while creation waits. Never borrow the source's
                    // browser or publish a tool grant before skill startup.
                    await peer.EnsureMcp().WaitForBrowserAsync(cancellation).ConfigureAwait(false);
                }
                await skills.InitializeAsync(cancellation).ConfigureAwait(false);
            }
            cancellation.ThrowIfCancellationRequested();
            peer.History().Save(persistence);
            peer.History().StartLog(Path.Combine(peer.DataDirectory(), "events.jsonl"));
            await peer.RestoreChildrenAsync(cancellation, waitForBrowser).ConfigureAwait(false);
            return peer;
        }
        catch (Exception startupError)
        {
            try
            {
                // InitializeAsync cancellation detaches its waiter. Destroy joins
                // the actual actor and MCP startup before removing ownership.
                peer.Destroy();
            }
            catch (Exception cleanupError)
            {
                throw new AggregateException("Peer startup and cleanup failed.", startupError, cleanupError);
            }
            throw;
        }
    }

    private partial ImmutableArray<PeerLaunch> ReadPeerLaunches(Agent source)
    {
        byte[] bytes;
        try
        {
            bytes = source.Sandbox().ReadFile(Path.Combine(DataDirectory(), "ensemble.json"));
        }
        catch (FileNotFoundException)
        {
            return [];
        }
        // Only absence is an empty registry. Malformed, inaccessible and damaged
        // files must not be silently replaced by a new registry.
        try
        {
            using var document = JsonDocument.Parse(bytes);
            if (document.RootElement.ValueKind != JsonValueKind.Array)
            {
                throw new InvalidDataException("The peer registry must be a JSON array.");
            }
            var entries = ImmutableArray.CreateBuilder<PeerLaunch>();
            var names = new HashSet<string>(StringComparer.Ordinal);
            foreach (var row in document.RootElement.EnumerateArray())
            {
                if (row.ValueKind != JsonValueKind.Object)
                {
                    throw new InvalidDataException("A peer registry entry must be an object.");
                }
                var fields = new HashSet<string>(StringComparer.Ordinal);
                foreach (var property in row.EnumerateObject())
                {
                    bool known = property.Name is "name" or "model" or "skill" or "created" or "offering";
                    bool validKind = property.Name == "offering"
                        ? property.Value.ValueKind == JsonValueKind.Object
                        : property.Value.ValueKind == JsonValueKind.String;
                    if (!known || !validKind || !fields.Add(property.Name))
                    {
                        throw new InvalidDataException("Invalid or duplicate peer registry field.");
                    }
                }
                if (!fields.Contains("name") || !fields.Contains("model")
                    || !fields.Contains("skill") || !fields.Contains("created"))
                {
                    throw new InvalidDataException("A peer registry entry requires name, model, skill and created.");
                }
                string name = row.GetProperty("name").GetString()!;
                string model = row.GetProperty("model").GetString()!;
                string skill = row.GetProperty("skill").GetString()!;
                _ = AgentDirectory(name);
                if (string.IsNullOrWhiteSpace(model) || !names.Add(name)
                    || !row.GetProperty("created").TryGetDateTimeOffset(out var created))
                {
                    throw new InvalidDataException("Invalid or duplicate peer launch identity.");
                }
                ModelIdentity? offering = row.TryGetProperty("offering", out var savedOffering)
                    ? JsonCodec.ReadModelIdentity(savedOffering) : null;
                entries.Add(new PeerLaunch(name, model, skill, created, offering));
            }
            return entries.ToImmutable();
        }
        catch (Exception error) when (error is JsonException or ArgumentException or FormatException)
        {
            throw new InvalidDataException("Invalid peer registry.", error);
        }
    }

    private partial void SavePeerLaunches(Agent source, ImmutableArray<PeerLaunch> entries)
    {
        var rows = new JsonArray();
        foreach (var entry in entries)
        {
            var row = new JsonObject
            {
                ["name"] = entry.Name,
                ["model"] = entry.Model,
                ["skill"] = entry.Skill,
                ["created"] = entry.Created
            };
            if (entry.Offering is { } offering)
            {
                row["offering"] = JsonCodec.WriteModelIdentity(offering);
            }
            rows.Add(row);
        }
        // The supplied owner's Sandbox enforces the same host policy as its
        // History. No unconfined fallback or new filesystem service is needed.
        source.Sandbox().ReplaceFile(Path.Combine(DataDirectory(), "ensemble.json"),
            JsonSerializer.SerializeToUtf8Bytes(rows));
    }
}
