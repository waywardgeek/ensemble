using System.Text;
using System.Text.Json;

namespace Ensemble;

public sealed partial class Agent
{
    public partial string DataDirectory()
    {
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (persistence is null)
            {
                throw new InvalidOperationException("Configure Agent persistence before using child storage.");
            }
            return Path.GetDirectoryName(persistence.Path)!;
        }
    }

    public partial JsonElement SubmissionSchema() => submissionContract.Schema();

    internal partial void ValidateSubmission(JsonElement payload)
    {
        _ = submissionContract.Validate(payload);
    }

    public partial JsonElement? Submission()
    {
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            return submittedPayload?.Clone();
        }
    }

    public partial async Task<JsonElement> WaitForSubmissionAsync(CancellationToken cancellation)
    {
        while (true)
        {
            cancellation.ThrowIfCancellationRequested();
            Task changed;
            lock (ingress)
            {
                ObjectDisposedException.ThrowIf(destroyed, this);
                if (agentParent is null)
                {
                    throw new InvalidOperationException("Only a delegated Agent has a parent submission.");
                }
                if (submittedPayload is JsonElement payload)
                {
                    return payload.Clone();
                }
                if (childFailure is not null)
                {
                    throw new InvalidOperationException(childFailure);
                }
                ObjectDisposedException.ThrowIf(admissionClosed || destroying, this);
                // Capture the next pulse while examining the fact. Publication
                // cannot fall between the empty-state check and subscription.
                changed = submissionChanged.Task;
            }
            await changed.WaitAsync(cancellation).ConfigureAwait(false);
        }
    }

    private partial void SignalSubmissionChange()
    {
        lock (ingress)
        {
            TaskCompletionSource previous = submissionChanged;
            submissionChanged = new(TaskCreationOptions.RunContinuationsAsynchronously);
            previous.TrySetResult();
        }
    }

    public partial JsonElement PublishSubmission(string text)
    {
        // Text convenience cannot bypass a custom schema requiring other fields.
        return StoreSubmission(submissionContract.Validate(text));
    }

    public partial JsonElement PublishSubmission(JsonElement payload)
    {
        return StoreSubmission(submissionContract.Validate(payload));
    }

    private partial JsonElement StoreSubmission(JsonElement validated)
    {
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed || destroying || admissionClosed, this);
            if (agentParent is null)
            {
                throw new InvalidOperationException("Only a delegated Agent can submit; standalone Agents finish in conversation.");
            }
            if (runningWorkGeneration != admittedWorkGeneration)
            {
                throw new InvalidOperationException("Newer work is waiting; the older turn cannot submit for it.");
            }
            // Submitted descendants may still be draining their submit tool.
            // They need not be destroyed, but idle/stopped/failed without a
            // submission is unresolved, never implicit successful completion.
            string[] unresolved = children.Where(child => child.Submission() is null)
                .Select(child => child.childID!).ToArray();
            if (unresolved.Length != 0)
            {
                throw new InvalidOperationException("Join or shut down unresolved children before submitting: "
                    + string.Join(", ", unresolved));
            }
            string path = Path.Combine(DataDirectory(), "submit.md");
            // One JSON row retains the actual schema and typed payload. This is
            // a recovery artifact, not a Markdown rendering or a chat echo.
            string row = JsonSerializer.Serialize(new
            {
                version = 1,
                agent_id = childID,
                generation = admittedWorkGeneration,
                submit_schema = submissionContract.Schema(),
                payload = validated
            }) + "\n";
            Sandbox().ReplaceFile(path, Encoding.UTF8.GetBytes(row));
            // Publication is last. Validation, open, flush or rename refusal
            // leaves the earlier visible completion and saved file intact.
            submittedPayload = validated;
            submissionReminders = 0;
            childFailure = null;
            SignalSubmissionChange();
            return validated.Clone();
        }
    }

    private partial long AdmitChildWork()
    {
        lock (ingress)
        {
            if (agentParent is null)
            {
                return 0;
            }
            admittedWorkGeneration = checked(admittedWorkGeneration + 1);
            submittedPayload = null;
            childFailure = null;
            SignalSubmissionChange();
            return admittedWorkGeneration;
        }
    }

    private partial void StartChildWork(long generation)
    {
        lock (ingress)
        {
            // Actor-created continuations have generation zero. They continue
            // current work rather than inventing another external admission.
            runningWorkGeneration = generation == 0 ? admittedWorkGeneration : generation;
        }
    }
}
