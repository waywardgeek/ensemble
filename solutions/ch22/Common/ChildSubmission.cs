using System.Text.Json;

namespace Ensemble;

public sealed partial class Agent
{
    // The contract is immutable and value-like. Publication and generation fields
    // are protected by ingress, including during disk replacement. The Agent's
    // explicit home owns the artifact path independently of History save paths.
    private readonly SubmissionContract submissionContract = new();
    // Includes schema, payload and provenance; underlying schema/payload limits
    // remain stricter. Readers and writers share this envelope bound.
    private const int SubmissionArtifactBytes = 2 * 1024 * 1024;
    /// <summary>Read a direct child's retained artifact freshly through this
    /// Agent's Sandbox. With no live child, use children/id beneath this Agent's
    /// home. Never constructs an Agent, activates skills or publishes completion.
    /// Missing, corrupt or foreign artifacts refuse. A custom live child home is
    /// respected while that child exists; no directory registry is invented.</summary>
    public partial StoredSubmission RetrieveSubmission(string childID);
    private JsonElement? submittedPayload;
    private long admittedWorkGeneration;
    private long runningWorkGeneration;
    // Consecutive protocol corrections reset only on a durable submission.
    // Failure is current-work state, not a fabricated submission or idle status.
    private int submissionReminders;
    private string? childFailure;
    /// <summary>The current delegated task's explicit failure, if any. New accepted
    /// work or durable submission clears it; shutdown alone does not. This is not
    /// a completion payload and does not imply quiescence.</summary>
    public partial string? ChildFailure();
    private partial bool CorrectChildCompletion();
    private partial void FailChildWork(Exception error);
    // A broadcast pulse, not a second completion record. Every observer rereads
    // submittedPayload under ingress; no observer can consume another's wake.
    private TaskCompletionSource submissionChanged = new(TaskCreationOptions.RunContinuationsAsynchronously);

    public partial JsonElement SubmissionSchema();
    // Tool admission uses the same immutable contract as durable publication.
    internal partial void ValidateSubmission(JsonElement payload);
    /// <summary>Return only an explicitly validated, durably stored submission.
    /// New admitted work invalidates completion without erasing saved files.</summary>
    public partial JsonElement? Submission();
    /// <summary>Wait for an explicit current submission. Cancellation detaches
    /// only this observer. Explicit task failure or shutdown without submission
    /// fails the wait rather than leaving it pending. This is
    /// not a quiescence barrier or question-aware fan-in; use ShutdownAsync to
    /// join owned work before destruction.</summary>
    public partial Task<JsonElement> WaitForSubmissionAsync(CancellationToken cancellation = default);
    private partial void SignalSubmissionChange();
    /// <summary>Validate the whole payload and atomically persist it before exposing
    /// completion. Standalone Agents have no Agent parent to submit to. An older
    /// turn cannot complete newer queued work. Shutdown closes publication.</summary>
    public partial JsonElement PublishSubmission(JsonElement payload);
    public partial JsonElement PublishSubmission(string text);
    private partial JsonElement StoreSubmission(JsonElement validated);
    private partial long AdmitChildWork();
    private partial void StartChildWork(long generation);
}

// Value-like historical data, not a child owner or a live completion record.
// ParentAddress is absent for v1 artifacts and unnamed transient parents.
public sealed record StoredSubmission(string AgentID, string? ParentAddress,
    long Generation, JsonElement Schema, JsonElement Payload);
