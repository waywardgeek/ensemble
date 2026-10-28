namespace Ensemble;

public sealed partial class Agent
{
    public partial Agent(Ensemble parent, Configuration configuration)
    {
        this.configuration = configuration;
        parent.AddAgent(this);
        try
        {
            _ = new History(this);
            _ = new Engine(this);
        }
        catch
        {
            // Construction may have registered History before Engine failed.
            Destroy();
            throw;
        }
    }

    public partial Ensemble Parent() => parent ?? throw new ObjectDisposedException(nameof(Agent));
    public partial Configuration Configuration()
    {
        ObjectDisposedException.ThrowIf(destroyed, this);
        return configuration;
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
        History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human, parts: [new TextPart(text)]));
        return await Engine().ExchangeAsync(cancellation);
    }

    public partial void SetEphemeral(string text)
    {
        var update = new Event(EventKind.MessageReceived,
            actor: Actor.System, parts: [new TextPart(text)]);
        History().Append(update);
    }

    public partial void Destroy()
    {
        if (destroyed)
        {
            return;
        }

        // Set the guard before cascading; repeated destruction cannot re-enter.
        destroyed = true;
        try
        {
            engine?.Destroy();
        }
        finally
        {
            try
            {
                history?.Destroy();
            }
            finally
            {
                // The parent removes both links. No second back-pointer write here.
                parent?.RemoveAgent(this);
            }
        }
    }
}
