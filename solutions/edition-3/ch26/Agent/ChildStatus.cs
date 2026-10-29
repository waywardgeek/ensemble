namespace Ensemble;

public sealed partial class Agent
{
    public partial async Task<AgentWorkStatus> WorkStatusAsync(CancellationToken cancellation)
    {
        cancellation.ThrowIfCancellationRequested();
        Task<AgentWorkStatus> response;
        lock (ingress)
        {
            if (admissionClosed || destroyed || destroying)
            {
                return CaptureWorkStatus(false, false);
            }
            var reply = new TaskCompletionSource<AgentWorkStatus>(TaskCreationOptions.RunContinuationsAsynchronously);
            // Enqueue under the same admission lock: shutdown cannot close the
            // mailbox between this check and admission of the typed query.
            Enqueue(new Mail(MailKind.WorkStatus, StatusReply: reply));
            response = reply.Task;
        }
        try
        {
            return await response.WaitAsync(cancellation);
        }
        catch (ObjectDisposedException)
        {
            // A query admitted just before shutdown can be rejected by the
            // final drain. Report that closed lifetime, never strand the caller.
            return CaptureWorkStatus(false, false);
        }
    }

    private partial void ReadWorkStatus(Mail message)
    {
        // These fields belong to the actor and must not be sampled by the tool's
        // worker thread. Direct Engine sends are counted under ingress below.
        bool busy = active is not null || requests.Count != 0 || workerPending
            || transportPending || automaticPending || startupReplies.Count != 0
            || controlReplies.Count != 0;
        message.StatusReply!.TrySetResult(CaptureWorkStatus(busy, paused));
    }

    private partial AgentWorkStatus CaptureWorkStatus(bool busy, bool isPaused)
    {
        lock (ingress)
        {
            AgentActivity activity;
            if (admissionClosed || destroyed || destroying)
            {
                activity = actorTask is null || actorTask.IsCompleted
                    ? AgentActivity.Stopped : AgentActivity.Stopping;
                isPaused = false; // Pause has no scheduling meaning after closure.
            }
            else
            {
                activity = busy || engineSends != 0 ? AgentActivity.Busy : AgentActivity.Idle;
            }
            WorkActivity work = activity is AgentActivity.Stopping or AgentActivity.Stopped
                ? lastEmittedActivity : ReadWorkActivity();
            return new AgentWorkStatus(activity, isPaused, submittedPayload.HasValue, childFailure, work);
        }
    }
}
