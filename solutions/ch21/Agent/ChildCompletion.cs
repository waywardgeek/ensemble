namespace Ensemble;

public sealed partial class Agent
{
    public partial string? ChildFailure()
    {
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            return childFailure;
        }
    }

    // Only the actor calls this at a successful model-turn boundary. Returning
    // true keeps its reliable reply pending and lets Advance render the durable
    // correction as another request. There is no timer or second task loop.
    private partial bool CorrectChildCompletion()
    {
        lock (ingress)
        {
            if (agentParent is null || submittedPayload is not null || admissionClosed
                || runningWorkGeneration != admittedWorkGeneration)
            {
                return false;
            }
            if (submissionReminders == 3)
            {
                childFailure = "Delegated task ended without submit after three framework corrections.";
                Record(new Event(EventKind.ErrorOccurred, error: childFailure));
                SignalSubmissionChange();
                return false;
            }
            string[] unresolved = children.Where(child => child.Submission() is null)
                .Select(child => child.childID!).ToArray();
            string instruction = unresolved.Length == 0
                ? "Your delegated task is not complete. Call submit with your result. Ordinary final text is not submission."
                : "Your delegated task is not complete. Join or shut down unresolved children before calling submit: "
                    + string.Join(", ", unresolved);
            submissionReminders++;
            Record(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart($"Framework correction {submissionReminders}/3: {instruction}")]));
            return true;
        }
    }

    // Failure belongs only to the current generation. An older queued turn
    // cannot poison newer work. An accepted submission remains authoritative
    // even if later draining fails; the reliable turn reply still reports that.
    private partial void FailChildWork(Exception error)
    {
        lock (ingress)
        {
            if (agentParent is not null && submittedPayload is null
                && runningWorkGeneration == admittedWorkGeneration)
            {
                childFailure = error.Message;
                SignalSubmissionChange();
            }
        }
    }
}
