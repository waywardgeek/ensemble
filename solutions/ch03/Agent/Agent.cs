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
        // A round limit bounds model repetition, not HTTP or command duration.
        for (int round = 0; round < 16; round++)
        {
            // Engine performs and records one exchange. The loop reads that
            // canonical response rather than retaining a second conversation.
            var reply = await Engine().ExchangeAsync(cancellation);
            var response = History().Events().Last(e => e.Kind == EventKind.ResponseEnded);
            var calls = response.Parts.OfType<ToolCallPart>().ToArray();
            if (calls.Length == 0)
            {
                return reply;
            }
            if (!string.IsNullOrWhiteSpace(reply))
            {
                // Keep JSON stdout reserved for the final response. The CLI's
                // existing diagnostics channel exposes narration before effects.
                Parent().Log(reply);
            }
            foreach (var call in calls)
            {
                // Publish dispatch before its effect, then the real result before
                // another request. Failure is still a result, not a lost batch.
                History().Append(new Event(EventKind.ToolCalled, callID: call.CallID, name: call.Name, args: call.Args));
                // Some selected models require an enforced visible preamble.
                // Return a repairable result without any tool effect when absent.
                ToolResultPart result;
                if (configuration.RequireToolExplanation && string.IsNullOrWhiteSpace(reply))
                {
                    result = new ToolResultPart(call.CallID,
                        [new TextPart("Tool not executed: first explain to the user what you will do and why, then retry.")], isError: true);
                }
                else
                {
                    result = await DispatchAsync(call, cancellation);
                }
                History().Append(new Event(EventKind.ToolReturned, parts: result.Parts, callID: call.CallID, isError: result.IsError));
            }
        }
        const string error = "Tool round limit reached (16); no further request was sent.";
        History().Append(new Event(EventKind.ErrorOccurred, error: error));
        throw new InvalidOperationException(error);
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
            foreach (var tool in tools.ToArray())
            {
                tool.Destroy();
            }
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
