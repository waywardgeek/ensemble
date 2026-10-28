namespace Ensemble;

public sealed partial class Agent
{
    public partial Agent(Ensemble parent, Configuration configuration, string? outputDirectory, string? name, SkillSpec? skillSpec, string? dataDirectory)
    {
        ArgumentNullException.ThrowIfNull(parent);
        ArgumentNullException.ThrowIfNull(configuration);
        if (name is not null)
        {
            ArgumentException.ThrowIfNullOrWhiteSpace(name);
        }
        Name = name;
        this.configuration = configuration;
        nextTurnToolLimit = configuration.ToolRoundLimit;
        string? namedHome = parent.AgentDirectory(name);
        if (namedHome is not null && dataDirectory is not null)
        {
            throw new ArgumentException(
                "Ensemble assigns the home of a named Agent; omit the explicit data directory.", nameof(dataDirectory));
        }
        ConstructOwners(parent, null, outputDirectory, skillSpec, namedHome ?? dataDirectory);
    }

    private partial void ConstructOwners(Ensemble root, Agent? owner,
        string? outputDirectory, SkillSpec? skillSpec, string? dataDirectory)
    {
        root.AddAgent(this);
        try
        {
            owner?.AddChild(this);
            _ = new Sandbox(this);
            if (dataDirectory is not null)
            {
                ArgumentException.ThrowIfNullOrWhiteSpace(dataDirectory);
                // Resolve and create through this Agent's actual filesystem owner.
                // This home is independent of the optional History save policy.
                this.dataDirectory = Sandbox().Resolve(dataDirectory);
                Sandbox().CreateDirectory(this.dataDirectory);
                outputDirectory ??= Path.Combine(this.dataDirectory, "jobs");
            }
            _ = new History(this);
            _ = new Engine(this);
            _ = new Jobs(this, outputDirectory);
            if (skillSpec is not null)
            {
                InstallTools();
                _ = new Skills(this, skillSpec);
            }
        }
        catch
        {
            // Partial construction has no admitted work. Do not start an actor
            // merely to clean up: its shutdown path requires all core owners.
            lock (ingress)
            {
                if (actorTask is null)
                {
                    admissionClosed = true;
                    mailbox.Writer.TryComplete();
                }
            }
            Destroy();
            throw;
        }
    }

    public partial int MaxToolRounds()
    {
        lock (ingress)
        {
            return nextTurnToolLimit;
        }
    }

    public partial void SetMaxToolRounds(int value)
    {
        ArgumentOutOfRangeException.ThrowIfNegative(value);
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            nextTurnToolLimit = value;
        }
    }

    public partial Ensemble Parent() => parent ?? throw new ObjectDisposedException(nameof(Agent));
    public partial Configuration Configuration()
    {
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            return configuration;
        }
    }

    internal partial string DiagnosticKey()
    {
        lock (ingress)
        {
            return configuration.ApiKey;
        }
    }

    public partial History History() => history ?? throw new ObjectDisposedException(nameof(History));
    public partial Engine Engine() => engine ?? throw new ObjectDisposedException(nameof(Engine));
    internal partial void AddHistory(History child)
    {
        ObjectDisposedException.ThrowIf(destroyed, this);
        if (history is not null || child.parent is not null || child.destroyed)
        {
            throw new InvalidOperationException("History already owned or wrong parent");
        }

        history = child;
        child.parent = this;
    }

    internal partial void RemoveHistory(History child)
    {
        if (!ReferenceEquals(history, child) || !ReferenceEquals(child.parent, this))
        {
            throw new InvalidOperationException("History is not owned here");
        }

        history = null;
        child.parent = null;
    }

    internal partial void AddEngine(Engine child)
    {
        ObjectDisposedException.ThrowIf(destroyed, this);
        if (engine is not null || child.parent is not null || child.destroyed)
        {
            throw new InvalidOperationException("Engine already owned or wrong parent");
        }

        engine = child;
        child.parent = this;
    }

    internal partial void RemoveEngine(Engine child)
    {
        if (!ReferenceEquals(engine, child) || !ReferenceEquals(child.parent, this))
        {
            throw new InvalidOperationException("Engine is not owned here");
        }

        engine = null;
        child.parent = null;
    }

    public partial async Task<string> SendAsync(string text, CancellationToken cancellation)
    {
        // Cancellation detaches this caller. Only an explicit interrupt changes
        // the Agent's turn; a timeout cannot silently cancel independent jobs.
        return await SubmitAsync([new TextPart(text)]).WaitAsync(cancellation);
    }

    public partial void SetEphemeral(string text)
    {
        PostAsync(new AgentInput(InputKind.Ephemeral, text)).GetAwaiter().GetResult();
    }

    public partial void Destroy()
    {
        if (destroyed || destroying)
        {
            return;
        }

        // Keep parent paths valid through joined shutdown and descendant cleanup.
        // Attempt every detach even after a failure. In particular, a child's
        // journal error must neither hide the parent's error nor strand siblings.
        destroying = true;
        List<Exception> failures = [];
        void Attempt(Action cleanup)
        {
            try
            {
                cleanup();
            }
            catch (Exception error)
            {
                failures.Add(error);
            }
        }
        try
        {
            Attempt(() => ShutdownAsync().GetAwaiter().GetResult());
            Attempt(DestroyChildren);
            destroyed = true;
            Attempt(() => jobs?.Destroy());
            Attempt(() => mcp?.DestroyAsync().GetAwaiter().GetResult());
            Attempt(() => skills?.Destroy());
            foreach (Tool tool in Tools())
            {
                Attempt(tool.Destroy);
            }
            Attempt(() => engine?.Destroy());
            Attempt(() => history?.Destroy());
            // Jobs and History finish with rooted files before the shared root
            // handle is closed. Only then remove both strong-parent memberships.
            Attempt(() => sandbox?.Destroy());
            observations.Writer.TryComplete();
            Attempt(() => agentParent?.RemoveChild(this));
            Attempt(() => parent?.RemoveAgent(this));
        }
        finally
        {
            destroying = false;
        }
        if (failures.Count == 1)
        {
            System.Runtime.ExceptionServices.ExceptionDispatchInfo.Capture(failures[0]).Throw();
        }
        if (failures.Count > 1)
        {
            throw new AggregateException("Agent cleanup failed.", failures);
        }
    }
}
