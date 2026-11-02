namespace Ensemble;

public sealed partial class Agent
{
    public partial Task<Agent> CreateChildAsync(string childID, string? model,
        System.Text.Json.JsonElement? submitSchema, CancellationToken cancellation)
        => CreateChildCoreAsync(childID, model, submitSchema, null, cancellation);

    private partial async Task<Agent> CreateChildCoreAsync(string childID, string? model,
        System.Text.Json.JsonElement? submitSchema, ChildRestartState? restart, CancellationToken cancellation, bool waitForBrowser)
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
                child.childRestoring = restart is not null;
                if (restart is not null)
                {
                    child.paused = child.explicitPause = restart.Paused;
                    child.admittedWorkGeneration = restart.Generation;
                    child.runningWorkGeneration = restart.Generation;
                    if (restart.Submitted)
                    {
                        StoredSubmission saved = RetrieveSubmission(childID);
                        if (saved.Generation != restart.Generation)
                        {
                            throw new InvalidDataException("Child submission generation differs from restart registration.");
                        }
                        string hash = Convert.ToHexString(System.Security.Cryptography.SHA256.HashData(
                            System.Text.Json.JsonSerializer.SerializeToUtf8Bytes(saved.Payload)));
                        if (hash != restart.SubmissionHash)
                        {
                            throw new InvalidDataException("Child submission differs from its acknowledged restart record.");
                        }
                        child.submittedPayload = child.submissionContract.Validate(saved.Payload);
                    }
                }
                child.InstallTools();
                if (bootstrap is not null)
                {
                    _ = new Skills(child, bootstrap);
                }
                string save = Path.Combine(childDirectory, "save.json");
                child.StartPersistence(new PersistenceOptions(save));
                if (Parent().agentDiagnostics) child.StartDiagnostics();
            }
            // No parent gate spans asynchronous startup. Shutdown can close child
            // admission and join its real actor/MCP work while this caller waits.
            if (child.SkillManager() is { } skills)
            {
                if (waitForBrowser && skills.RequiresBrowserStartup())
                {
                    await child.EnsureMcp().WaitForBrowserAsync(cancellation).ConfigureAwait(false);
                }
                await skills.InitializeAsync(cancellation);
            }
            cancellation.ThrowIfCancellationRequested();
            lock (child.ingress)
            {
                ObjectDisposedException.ThrowIf(child.destroyed || child.destroying || child.admissionClosed, child);
            }
            lock (child.ingress)
            {
                child.childRestartEnabled = true;
                child.childRestoring = false;
                if (restart is null) { child.SaveChildRestart(); }
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
