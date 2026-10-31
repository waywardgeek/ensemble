using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Agent
{
    public partial Agent(Ensemble parent, Configuration configuration)
    {
        ArgumentNullException.ThrowIfNull(parent);
        ArgumentNullException.ThrowIfNull(configuration);
        this.configuration = configuration;
        messages = new List<Message>();
        parent.AddAgent(this);

        try
        {
            // Engine registers itself at both ends through AddEngine.
            _ = new Engine(this);
        }
        catch
        {
            // A failed constructor must not leave a partially built child
            // reachable from its owner.
            parent.RemoveAgent(this);
            throw;
        }
    }

    public partial Ensemble Parent()
    {
        return parent ?? throw new InvalidOperationException("Agent is detached.");
    }

    public partial Configuration Configuration()
    {
        return configuration;
    }

    public partial ImmutableArray<Message> Messages()
    {
        return messages.ToImmutableArray();
    }

    public partial Engine Engine()
    {
        return engine ?? throw new InvalidOperationException("Agent has no Engine.");
    }

    internal partial void AddEngine(Engine child)
    {
        ObjectDisposedException.ThrowIf(destroyed, this);
        if (engine is not null || child.parent is not null || child.destroyed)
        {
            throw new InvalidOperationException("Engine is destroyed or already attached.");
        }

        engine = child;
        child.parent = this;
    }

    internal partial void RemoveEngine(Engine child)
    {
        if (engine != child || child.parent != this)
        {
            throw new InvalidOperationException("Engine does not belong to this Agent.");
        }

        engine = null;
        child.parent = null;
    }

    public partial async Task<string> AskAsync(string question, CancellationToken cancellation)
    {
        ObjectDisposedException.ThrowIf(destroyed, this);
        if (string.IsNullOrWhiteSpace(question))
        {
            throw new ArgumentException("User message is empty.", nameof(question));
        }

        var user = new Message("user", question);
        ImmutableArray<Message> request = Messages().Add(user);
        string answer = await Engine().ExchangeAsync(request, cancellation);

        // Publication is deliberately after the entire exchange succeeds.
        // Failed requests never leave an unmatched user in the conversation.
        messages.Add(user);
        messages.Add(new Message("assistant", answer));
        return answer;
    }

    public partial void Destroy()
    {
        if (destroyed)
        {
            return;
        }

        destroyed = true;
        try
        {
            engine?.Destroy();
        }
        finally
        {
            messages.Clear();
            parent?.RemoveAgent(this);
        }
    }
}
