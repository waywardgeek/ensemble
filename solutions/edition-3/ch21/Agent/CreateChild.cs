namespace Ensemble;

public sealed partial class Agent
{
    public partial async Task<Agent> CreateChildAsync(string childID, string? model,
        System.Text.Json.JsonElement? submitSchema, CancellationToken cancellation)
    {
        cancellation.ThrowIfCancellationRequested();
        // Snapshot values before admission: no Skills -> Agent lock-order reversal.
        // The child constructor still revalidates policy and both live parents.
        var inherited = Configuration();
        if (model is not null)
        {
            inherited = inherited.WithModel(model);
        }
        var bootstrap = SkillManager()?.CaptureBootstrap();
        Agent? child = null;
        try
        {
            lock (ingress)
            {
                ObjectDisposedException.ThrowIf(destroyed || destroying || admissionClosed, this);
                cancellation.ThrowIfCancellationRequested();
                string directory = DataDirectory();
                // Constructor validates schema and ID before creating any output.
                // Its registration refuses duplicate sibling IDs before file work.
                string childDirectory = Path.Combine(directory, "children", childID);
                child = new Agent(Parent(), this, childID, inherited,
                    outputDirectory: Path.Combine(childDirectory, "jobs"), submitSchema: submitSchema);
                child.InstallTools();
                if (bootstrap is not null)
                {
                    _ = new Skills(child, bootstrap);
                }
                string save = Path.Combine(childDirectory, "save.json");
                child.StartPersistence(new PersistenceOptions(save));
            }
            // No parent gate spans asynchronous startup. Shutdown can close child
            // admission and join its real actor/MCP work while this caller waits.
            if (child.SkillManager() is { } skills)
            {
                await skills.InitializeAsync(cancellation);
            }
            cancellation.ThrowIfCancellationRequested();
            lock (child.ingress)
            {
                ObjectDisposedException.ThrowIf(child.destroyed || child.destroying || child.admissionClosed, child);
            }
            return child;
        }
        catch (Exception creationError)
        {
            if (child is not null)
            {
                try
                {
                    child.Destroy();
                }
                catch (Exception cleanupError)
                {
                    throw new AggregateException("Child creation and cleanup failed.", creationError, cleanupError);
                }
            }
            throw;
        }
    }
}
