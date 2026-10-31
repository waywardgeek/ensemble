namespace Ensemble;

public sealed partial class Agent
{
    public partial Task SetModelAsync(string model)
    {
        var reply = new TaskCompletionSource<string>(TaskCreationOptions.RunContinuationsAsynchronously);
        Enqueue(new Mail(MailKind.SelectModel, Model: model, Reply: reply));
        return reply.Task;
    }

    private partial void SelectModel(Mail message)
    {
        try
        {
            ObjectDisposedException.ThrowIf(stopping, this);
            // Queued prompts count as admitted work too. A selection must not
            // jump ahead of a paused turn or a skill transaction still starting.
            if (active is not null || requests.Count != 0 || workerPending
                || transportPending || automaticPending || startupReplies.Count != 0
                || controlReplies.Count != 0)
            {
                throw new InvalidOperationException("Model selection requires an idle Agent between turns.");
            }
            lock (ingress)
            {
                ObjectDisposedException.ThrowIf(destroyed, this);
                if (engineSends != 0)
                {
                    throw new InvalidOperationException("Model selection cannot change an active Engine send.");
                }
                // Validation constructs a complete immutable candidate first.
                // The meter/lens stay on the existing Engine; neither History
                // nor saved preferences becomes a second selection authority.
                Configuration candidate = configuration.WithModel(message.Model!);
                configuration = candidate;
                ResetMemoryPressure();
            }
            message.Reply!.TrySetResult("");
        }
        catch (Exception error) when (error is ArgumentException or InvalidOperationException)
        {
            message.Reply!.TrySetException(error);
        }
    }

    internal partial Configuration BeginEngineSend()
    {
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            engineSends++;
            if (goalConfiguration.Enabled) { mailbox.Writer.TryWrite(new Mail(MailKind.GoalRefresh)); }
            return configuration;
        }
    }

    internal partial void EndEngineSend()
    {
        lock (ingress)
        {
            // No cancellation or failure may strand a selection lock. This is
            // released by Engine's finally, including credential-acquisition and
            // response parsing failures. It never cancels the request itself.
            engineSends--;
            if (goalConfiguration.Enabled) { mailbox.Writer.TryWrite(new Mail(MailKind.GoalRefresh)); }
        }
    }
}
