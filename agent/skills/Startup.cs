using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Skills
{
    // Resolve in host-supplied order: a later selection can be offered by an earlier
    // one. The prospective name set is algorithmic only; History remains the sole
    // active-state authority. Nothing is published until the whole batch validates.
    private partial List<SkillDefinition> ResolveStartup()
    {
        var additions = Resolve(primaryName, initial: true);
        var selected = Active().Select(skill => skill.Name)
            .Concat(additions.Select(skill => skill.Name)).ToHashSet(StringComparer.Ordinal);
        foreach (string name in startupSkills)
        {
            if (string.IsNullOrWhiteSpace(name) || !definitions.TryGetValue(name, out var definition)
                || definition.Type != SkillType.Loadable)
            {
                throw new InvalidDataException("Startup skill is missing or not loadable: " + name);
            }
            if (selected.Contains(name))
            {
                continue; // Repeated selections or already included dependencies activate once.
            }
            if (!Available(selected).Contains(name, StringComparer.Ordinal))
            {
                throw new InvalidDataException("Startup skill is not currently offered: " + name);
            }
            var next = Resolve(name, initial: false, preceding: selected);
            additions.AddRange(next);
            selected.UnionWith(next.Select(skill => skill.Name));
        }
        return additions;
    }

    public partial bool Ready()
    {
        var active = Active();
        return active.Any(skill => skill.Primary && skill.Name == primaryName)
            && startupSkills.All(name => active.Any(skill => skill.Name == name));
    }

    public partial bool RequiresBrowserStartup()
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            return !Ready() && ResolveStartup()
                .Any(skill => !skill.McpServers.IsDefaultOrEmpty
                    && skill.McpServers.Any(server => server.Transport == "websocket"));
        }
    }

    public partial async Task InitializeAsync(CancellationToken cancellation)
    {
        // Cancellation releases this observer, not the owned startup operation.
        // Shutdown cancels the shared operation and waits for every process.
        await Parent().InitializeSkillsAsync().WaitAsync(cancellation);
    }

    internal partial Task BeginInitialization(CancellationToken cancellation)
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (Ready())
            {
                return Task.CompletedTask;
            }
            if (initialization is not null)
            {
                return initialization;
            }
            var additions = ResolveStartup().ToImmutableArray();
            var values = initialVariables
                ?? throw new InvalidOperationException("Primary startup has no construction snapshot.");
            var manager = Parent().EnsureMcp();
            string workspace = Directory.GetCurrentDirectory();
            // The actor keeps processing reverse calls and shutdown while the
            // worker prepares. Commit returns to this same actor before grants.
            initialization = Task.Run(() => PrepareAsync(primaryName, additions, values,
                manager, workspace, cancellation, initial: true));
            return initialization;
        }
    }
}
