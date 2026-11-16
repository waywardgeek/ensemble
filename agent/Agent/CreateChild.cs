namespace Ensemble;

public sealed partial class Agent
{
    public partial Task<Agent> CreateChildAsync(string childID, string? model,
        System.Text.Json.JsonElement? submitSchema, CancellationToken cancellation, long? workingContextTokens, ChildLaunchOptions? launch)
        => CreateChildCoreAsync(childID, model, submitSchema, null, cancellation,
            workingContextTokens: workingContextTokens, launch: launch);

    private partial async Task<Agent> CreateChildCoreAsync(string childID, string? model,
        System.Text.Json.JsonElement? submitSchema, ChildRestartState? restart, CancellationToken cancellation, bool waitForBrowser, long? workingContextTokens, ChildLaunchOptions? launch)
    {
        cancellation.ThrowIfCancellationRequested();
        ValidateChildID(childID);
        if (launch?.Offering is not null && model is not null)
        {
            throw new ArgumentException("Choose an exact offering or a legacy model, not both.");
        }
        string childDirectory = Path.Combine(DataDirectory(), "children", childID);
        ChildRestartState? savedLaunch = restart ?? ReadChildRestart(childID);
        if (restart is null && savedLaunch?.Version >= 3 && (launch is not null || model is not null))
        {
            throw new ArgumentException("Use a fresh child ID to select new launch options.");
        }
        // Snapshot values before admission: no Skills -> Agent lock-order reversal.
        // The child constructor still revalidates policy and both live parents.
        var inherited = Configuration();
        ModelIdentity? savedOffering = savedLaunch is null ? null
            : AgentSettings.Read(Sandbox(), Path.Combine(childDirectory, "settings.json"))?.SelectedOffering;
        if (savedOffering is not null)
        {
            inherited = new Configuration(inherited, Parent().ResolveModel(savedOffering));
        }
        else if (savedLaunch?.Offering is { } identity)
        {
            inherited = new Configuration(inherited, Parent().ResolveModel(identity));
            if (inherited.Model != savedLaunch.Model)
            {
                throw new InvalidDataException("Saved model name differs from its configured offering.");
            }
        }
        else if (launch?.Offering is { } selectedOffering)
        {
            inherited = new Configuration(inherited, Parent().ResolveModel(selectedOffering));
        }
        else if (model is not null)
        {
            inherited = inherited.WithModel(model);
        }
        // Saved policy is explicit. An unrestricted saved launch cannot silently
        // become an inherited launch under a newly confined creator.
        // A new assignment may reuse history, but must not widen the authority
        // under which that history and workspace were created. Completion and
        // pause state still come only from an explicit restoration.
        string? requestedRoot = savedLaunch?.SandboxRoot;
        if (requestedRoot == "" && Sandbox().Root().Length != 0)
        {
            throw new InvalidDataException("An unrestricted child cannot restore under a confined parent.");
        }
        if (launch?.Confined == true)
        {
            requestedRoot = childDirectory;
        }
        var imports = savedLaunch?.ProcessReadPaths ?? launch?.ProcessReadPaths;
        if (requestedRoot is not null || imports is not null)
        {
            var requested = new Configuration(inherited, inherited.SafeMode,
                inherited.EnableWebSearch, requestedRoot ?? inherited.SandboxRoot, imports);
            var derived = Sandbox().DeriveChild(requested);
            if (derived.Refusal is not null)
            {
                throw new ArgumentException(derived.Refusal);
            }
            inherited = derived.Effective;
        }
        // Validate against the actual child route before creating files or membership.
        // This is a working-input policy, not a change of model or funding.
        WorkingContextLimit? contextLimit = workingContextTokens is long tokens
            ? new WorkingContextLimit(inherited.Vendor, inherited.Model, tokens,
                inherited.ServingContextTokens ?? inherited.Features.ContextWindow, inherited.Offering)
            : null;
        var bootstrap = SkillManager()?.CaptureBootstrap();
        string? primary = savedLaunch?.PrimarySkill ?? launch?.PrimarySkill;
        if (primary is not null)
        {
            if (bootstrap is null || !bootstrap.Definitions.TryGetValue(primary, out var definition)
                || definition.Type != SkillType.Primary)
            {
                throw new ArgumentException("Child primary skill is missing or has the wrong type.");
            }
            // A selected primary is a complete starting contract. Do not attach
            // unrelated startup selections inherited from the creator.
            bootstrap = bootstrap with { Primary = primary, Startup = [] };
        }
        Agent? child = null;
        try
        {
            lock (ingress)
            {
                ObjectDisposedException.ThrowIf(destroyed || destroying || admissionClosed, this);
                cancellation.ThrowIfCancellationRequested();
                // Fresh confinement needs its root before Sandbox opens it.
                // Validate schema and duplicate ownership before any filesystem
                // effect. Restoration never recreates a missing saved root.
                if (launch?.Confined == true)
                {
                    _ = new SubmissionContract(submitSchema);
                    if (FindChild(childID) is not null)
                    {
                        throw new InvalidOperationException("Duplicate child Agent ID.");
                    }
                    Sandbox().CreateDirectory(childDirectory);
                }
                // Constructor revalidates schema, policy and both owners.
                child = new Agent(Parent(), this, childID, inherited,
                    outputDirectory: Path.Combine(childDirectory, "jobs"), submitSchema: submitSchema);
                child.childPrimarySkill = primary;
                child.childRestoring = restart is not null;
                if (restart is not null)
                {
                    child.paused = restart.Paused;
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
            if (contextLimit is not null)
            {
                // Persist before returning a ready child: the caller cannot admit
                // its first task under the old default, and restart keeps the limit.
                await child.SetWorkingContextLimitAsync(contextLimit, persist: true, cancellation: cancellation);
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
