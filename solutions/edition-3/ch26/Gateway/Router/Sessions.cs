using System.Net.WebSockets;

namespace Ensemble.Gateway;

public sealed partial class Router
{
    public partial bool ContainsIdentity(string id)
    {
        lock (gate) { return !stopped && identities.ContainsKey(id); }
    }

    internal partial Keypair StaticKey() => key;

    public partial async Task AcceptAsync(string id, WebSocket socket, CancellationToken cancellation)
    {
        ArgumentNullException.ThrowIfNull(socket);
        Session? session = null;
        try
        {
            lock (gate)
            {
                RequireLive();
                session = new Session(Identity(id), socket, cancellation);
            }
            await session.RunAsync();
        }
        finally
        {
            // Constructor failure never transfers the socket into the graph.
            if (session is null) socket.Dispose();
        }
    }

    public partial ValueTask DisposeAsync()
    {
        Session[] children;
        TaskCompletionSource completion;
        lock (gate)
        {
            if (closing is not null) return new ValueTask(closing.Task);
            stopped = true;
            closing = completion = new(TaskCreationOptions.RunContinuationsAsynchronously);
            children = identities.Values.SelectMany(identity => identity.Sessions()).ToArray();
        }
        // Cancellation/Abort and joins happen outside the policy gate. Even a
        // peer holding a partial handshake must not prevent orderly shutdown.
        _ = FinishCloseAsync(children, completion);
        return new ValueTask(completion.Task);
    }

    private partial async Task FinishCloseAsync(Session[] sessions, TaskCompletionSource completion)
    {
        List<Exception> failures = [];
        foreach (Session session in sessions)
        {
            try { session.Close(); }
            catch (Exception error) { failures.Add(error); }
        }
        foreach (Session session in sessions)
        {
            try { await session.Done(); }
            catch (Exception error) { failures.Add(error); }
        }
        lock (gate)
        {
            try { SaveLocked(); }
            catch (Exception error) { failures.Add(error); }
            try { DestroyModel(); }
            catch (Exception error) { failures.Add(error); }
            disposed = true;
        }
        if (failures.Count == 0) completion.TrySetResult();
        else if (failures.Count == 1) completion.TrySetException(failures[0]);
        else completion.TrySetException(new AggregateException(failures));
    }
}

internal sealed partial class RouterIdentity
{
    internal partial void AddSession(Session child)
    {
        Router owner = Parent();
        lock (owner.gate)
        {
            owner.RequireLive();
            owner.RequireIdentity(this);
            if (child.parent is not null || sessions.Contains(child))
                throw new InvalidOperationException("Session is already owned.");
            sessions.Add(child);
            child.parent = this;
        }
    }

    internal partial void RemoveSession(Session child)
    {
        Router owner = Parent();
        lock (owner.gate)
        {
            owner.RequireIdentity(this);
            if (child.parent != this || !sessions.Contains(child) || child.HasNoise())
                throw new InvalidOperationException("Detach session children before removing ownership.");
            if (current == child) current = null;
            sessions.Remove(child);
            child.parent = null;
        }
    }

    internal partial Session[] Sessions()
    {
        Router owner = Parent();
        lock (owner.gate) { owner.RequireIdentity(this); return sessions.ToArray(); }
    }

    internal partial bool IsCurrent(Session child)
    {
        Router owner = Parent();
        lock (owner.gate) { return child.parent == this && current == child && sessions.Contains(child); }
    }

    internal partial (Session? Previous, WelcomePayload Welcome, ulong Receipt) Activate(Session child, ulong receipt)
    {
        Router owner = Parent();
        lock (owner.gate)
        {
            owner.RequireLive();
            owner.RequireIdentity(this);
            if (child.parent != this || !sessions.Contains(child))
                throw new InvalidOperationException("Session is not owned by this identity.");
            QueueState state = Queue().Snapshot();
            if (receipt > state.Next) throw new InvalidDataException("Hello receipt exceeds allocated sequence.");
            Queue().Acknowledge(receipt);
            ulong acknowledged = Math.Max(state.Ack, receipt);
            int depth = state.Messages!.Value.Count(message => message.Seq > acknowledged);
            Session? previous = current;
            current = child;
            return (previous, new WelcomePayload(id, depth, "1"), acknowledged);
        }
    }

    internal partial void Notify()
    {
        Router owner = Parent();
        lock (owner.gate) { current?.Notify(); }
    }
}
