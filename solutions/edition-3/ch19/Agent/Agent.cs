namespace Ensemble;

public sealed partial class Agent
{
    public partial Agent(Ensemble parent, Configuration configuration, string? outputDirectory, string? name, SkillSpec? skillSpec)
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
        parent.AddAgent(this);
        try
        {
            _ = new Sandbox(this);
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

        // Keep public parent paths valid until joined shutdown has saved.
        destroying = true;
        try
        {
            try
            {
                ShutdownAsync().GetAwaiter().GetResult();
            }
            finally
            {
                destroyed = true;
                try
                {
                    jobs?.Destroy();
                }
                finally
                {
                    try
                    {
                        mcp?.DestroyAsync().GetAwaiter().GetResult();
                    }
                    finally
                    {
                        skills?.Destroy();
                        foreach (var tool in Tools())
                        {
                            tool.Destroy();
                        }
                        engine?.Destroy();
                    }
                }
            }
        }
        finally
        {
            try
            {
                history?.Destroy();
            }
            finally
            {
                try
                {
                    // History and Jobs have finished using rooted files before
                    // their shared root handle is closed and detached.
                    sandbox?.Destroy();
                }
                finally
                {
                    // The parent removes both links.
                    observations.Writer.TryComplete();
                    parent?.RemoveAgent(this);
                }
            }
        }
    }
}
