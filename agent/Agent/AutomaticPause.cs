namespace Ensemble;

public sealed partial class Agent
{
    public partial Task SetAutomaticPauseAsync(object owner, bool held)
    {
        ArgumentNullException.ThrowIfNull(owner);
        var reply = new TaskCompletionSource<string>(TaskCreationOptions.RunContinuationsAsynchronously);
        Enqueue(new Mail(MailKind.AutomaticPause, Reply: reply, PauseOwner: owner, PauseHeld: held));
        return reply.Task;
    }

    private partial void RefreshPause(string reason)
    {
        bool effective = explicitPause || automaticPauseOwners.Count != 0;
        if (paused == effective) return;
        if (effective) StopGoalTimer();
        paused = effective;
        Record(new Event(effective ? EventKind.AgentPaused : EventKind.AgentResumed, reason: reason));
    }
}
