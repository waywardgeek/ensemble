using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Skills
{
    public partial bool Ready()
    {
        return Active().Any(skill => skill.Primary);
    }

    public partial bool RequiresBrowserStartup()
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            return !Ready() && Resolve(primaryName, initial: true)
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
            var additions = Resolve(primaryName, initial: true).ToImmutableArray();
            var values = initialVariables
                ?? throw new InvalidOperationException("Primary startup has no construction snapshot.");
            var manager = Parent().Mcp() ?? new Mcp(Parent());
            string workspace = Directory.GetCurrentDirectory();
            // The actor keeps processing reverse calls and shutdown while the
            // worker prepares. Commit returns to this same actor before grants.
            initialization = Task.Run(() => PrepareAsync(primaryName, additions, values,
                manager, workspace, cancellation, initial: true));
            return initialization;
        }
    }
}
