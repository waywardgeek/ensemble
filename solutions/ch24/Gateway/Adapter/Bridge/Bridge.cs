using Ensemble;
using Ensemble.Gateway;

namespace Ensemble.GatewayAdapter;

public sealed partial class Bridge
{
    public partial Bridge() { }

    public partial Task Ready() => ready.Task;

    public partial async Task RunAsync(Agent agent, ClientConfig configuration,
        IEnumerable<string> groups, CancellationToken cancellation)
    {
        ArgumentNullException.ThrowIfNull(agent);
        ArgumentNullException.ThrowIfNull(configuration);
        ArgumentNullException.ThrowIfNull(groups);
        string[] acceptedGroups = groups.ToArray();
        lock (gate)
        {
            if (started)
                throw new InvalidOperationException("A Gateway bridge is single-use.");
            started = true;
            this.agent = agent;
            lifetime = CancellationTokenSource.CreateLinkedTokenSource(cancellation);
            accepting = true;
        }
        List<Exception> failures = [];
        try
        {
            client = await Client.ConnectAsync(configuration, this, lifetime.Token);
            // Offline deliveries may arrive before ConnectAsync returns. Their
            // turns may start immediately; replies await this publication only.
            connected.TrySetResult();
            using var setup = CancellationTokenSource.CreateLinkedTokenSource(lifetime.Token);
            setup.CancelAfter(TimeSpan.FromSeconds(10));
            await client.DeclareContactableAsync(acceptedGroups, setup.Token);
            ready.TrySetResult();
            await client.Done();
            if (!cancellation.IsCancellationRequested)
                throw new IOException("Gateway connection ended.");
        }
        catch (OperationCanceledException) when (cancellation.IsCancellationRequested) { }
        catch (Exception error)
        {
            failures.Add(error);
        }
        finally
        {
            lock (gate)
            {
                accepting = false;
            }
            lifetime.Cancel();
            if (failures.Count > 0)
                ready.TrySetException(failures[0]);
            else
                ready.TrySetCanceled();
            // RunAsync remains the authoritative outcome even if nobody waits
            // for readiness (for example, a caller cancelled during startup).
            if (ready.Task.IsFaulted)
                _ = ready.Task.Exception;
            connected.TrySetCanceled();
            // Join callbacks before capturing the reply list. A callback that
            // won the admission lock is thus included even during disconnect.
            if (client is not null)
            {
                try
                {
                    await client.DisposeAsync();
                }
                catch (Exception error)
                {
                    failures.Add(error);
                }
            }
            Task[] pending;
            lock (gate)
            {
                pending = replies.ToArray();
            }
            try
            {
                await Task.WhenAll(pending);
            }
            catch (Exception error)
            {
                failures.Add(error);
            }
            lock (gate)
            {
                replies.Clear();
                client = null;
                this.agent = null;
            }
            lifetime.Dispose();
            lifetime = null;
        }
        if (failures.Count == 1)
            System.Runtime.ExceptionServices.ExceptionDispatchInfo.Capture(failures[0]).Throw();
        if (failures.Count > 1)
            throw new AggregateException("Gateway bridge failed.", failures);
    }

    public partial void OnConnected(WelcomePayload welcome) { }

    public partial void OnDeliver(Envelope envelope)
    {
        lock (gate)
        {
            if (!accepting || agent is null || lifetime!.IsCancellationRequested)
                throw new ObjectDisposedException(nameof(Bridge));
            // The stamped envelope is ordinary input, never a grant or system
            // instruction. SubmitAsync synchronously admits to the actor mailbox;
            // its reliable task belongs to THIS turn, not the next observation.
            Task<string> result = agent.SubmitAsync([new TextPart(envelope.Content)]);
            // If disconnect cancels our wait, admitted work still belongs to the
            // Agent. Observe a later fault without retaining the bridge or Agent.
            _ = result.ContinueWith(static completed => { _ = completed.Exception; },
                CancellationToken.None, TaskContinuationOptions.OnlyOnFaulted |
                TaskContinuationOptions.ExecuteSynchronously, TaskScheduler.Default);
            replies.RemoveAll(static reply => reply.IsCompletedSuccessfully);
            replies.Add(ReplyAsync(agent, envelope, result, lifetime.Token));
            // Returning accepts delivery; it does not claim model success or a
            // durable Agent commit. A crash can lose or repeat admitted work.
        }
    }

    private partial async Task ReplyAsync(Agent actor, Envelope envelope, Task<string> result,
        CancellationToken cancellation)
    {
        try
        {
            string text = await result.WaitAsync(cancellation);
            if (envelope.ResponseMode != Protocol.ResponseAsync)
                return;
            await connected.Task.WaitAsync(cancellation);
            await client!.RespondAsync(envelope, text, cancellation);
        }
        catch (OperationCanceledException) when (cancellation.IsCancellationRequested) { }
        catch (Exception error)
        {
            // Do not send provider exceptions or secrets to a remote peer.
            // Existing parent diagnostics apply their normal redaction policy.
            actor.Parent().Log($"gateway turn/reply failed: {error.Message}");
        }
    }

    public partial void OnDisconnected(string reason)
    {
        lock (gate)
        {
            accepting = false;
            lifetime?.Cancel();
        }
    }

    public partial void OnError(Exception error)
    {
        lock (gate)
        {
            agent?.Parent().Log($"gateway connection error: {error.Message}");
        }
    }
}
