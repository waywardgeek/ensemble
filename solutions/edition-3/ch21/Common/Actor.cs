using System.Collections.Immutable;
using System.Threading.Channels;

namespace Ensemble;

// Value-like public input. Live messages become hints when a turn is active;
// SubmitAsync, in contrast, always reserves a separate request and reply.
public enum InputKind
{
    Message, Hint, Pause, Resume, InterruptNow, InterruptAfterTool, Ephemeral,
    // Admission is acknowledged immediately; the recorded tool-only cut waits
    // until current batch outcomes are paired. It does not cancel owned Jobs.
    ClearTools,
    // Interactive text: starts an idle turn with a reliable answer, or becomes
    // a hint to the active turn and returns an empty acknowledgment immediately.
    Chat
}
public sealed record AgentInput(InputKind Kind, string Text = "");

// Observations are disposable progress, not acknowledgments. Part carries the
// full neutral value; Seq identifies its durable event, not a second transcript.
public enum ObservationKind { Input, StateChanged, PartDelta, PartFinal, TurnEnded, EventRecorded, ActivityChanged }
public sealed record Observation(ObservationKind Kind, long Seq = 0, Part? Part = null,
    string Text = "", Phase? From = null, Phase? To = null, bool? Accepted = null, Chunk? Delta = null,
    int? PartID = null, long ResponseID = 0, bool? Succeeded = null, bool HasReply = false, Event? Fact = null, WorkActivity? Activity = null);
public sealed record AgentObservation(Agent Agent, Observation Observation);

public sealed partial class Ensemble
{
    private readonly Channel<AgentObservation> observations = Channel.CreateBounded<AgentObservation>(
        new BoundedChannelOptions(256) { FullMode = BoundedChannelFullMode.DropOldest });
    // Agent identities are actual references supplied by their owning parent.
    internal partial void Observe(Agent child, Observation value);
    public partial Task<AgentObservation> WaitAnyAsync(Func<AgentObservation, bool> predicate,
        CancellationToken cancellation = default);
}

public sealed partial class Agent
{
    // The ingress lock protects lifecycle admission and public configuration;
    // child completion fields use that same gate (see ChildSubmission). Turn
    // scheduling fields below are touched exclusively by the actor task.
    private readonly object ingress = new();
    private readonly Channel<Mail> mailbox = Channel.CreateUnbounded<Mail>(
        new UnboundedChannelOptions { SingleReader = true });
    private readonly Channel<Observation> observations = Channel.CreateBounded<Observation>(
        new BoundedChannelOptions(128) { FullMode = BoundedChannelFullMode.DropOldest });
    private Task? actorTask;
    private bool admissionClosed;
    private readonly Queue<Mail> requests = new();
    private readonly List<TaskCompletionSource<string>> flushReplies = [];
    // Caller-owned job controls may observe/kill independent jobs while a turn
    // is busy or paused. Their replies identify those outstanding observations.
    private readonly Dictionary<TaskCompletionSource<ToolResultPart>, ToolCallPart> controlReplies = [];
    // Initialization has no tool identity. Track its ordinary reliable replies
    // separately while joining it at the same shutdown/flush boundary.
    private readonly HashSet<TaskCompletionSource<string>> startupReplies = [];
    private readonly CancellationTokenSource controlCancellation = new();
    private Mail? active;
    private bool paused;
    private bool stopping;
    private bool workerPending;
    private bool transportPending;
    // Automatic values are turn/request-local; History's ephemera slot is the
    // authoritative projection once these are recorded just before a request.
    private bool automaticPending;
    private bool turnSnapshotsCollected;
    private bool roundSnapshotsCollected;
    private ImmutableArray<Part> turnSnapshots = [];
    private ImmutableArray<Part> roundSnapshots = [];
    private bool toolBeforeText;
    private long responseID;
    private bool interruptPending;
    private bool discardResponse;
    private bool skipTail;
    private bool hintPending;
    private CancellationTokenSource? observationCancellation;
    private ImmutableArray<ToolCallPart> calls = [];
    private int nextCall;
    // Staged only after a successful control result is recorded. The actor
    // publishes it after all companion calls have recorded terminal results.
    private HandoffPart? pendingHandoff;
    // A host clear waits for the current batch's paired outcomes. It does not
    // cancel independent Jobs, discard dialogue, or revoke skill instructions.
    private bool pendingToolClear;
    private partial void ApplyToolClear();
    private partial void StageHandoff(ToolCallPart call, ToolResultPart result);
    private partial void ApplyHandoff();
    private int batches;
    // Configuration seeds the policy; this field alone owns runtime updates.
    // The actor captures it once per turn, so a settings change cannot move an
    // in-flight turn's boundary. Negative constructor seeds retain Chapter 5's
    // library-only unlimited behavior; the runtime setter accepts >= 0 only.
    private int nextTurnToolLimit;
    // Policy is ingress-protected; the watermark is actor-owned. After restart
    // zero conservatively requires a successful presentation before any cut.
    public const int DefaultContextByteBudget = 400_000;
    private int contextByteBudget = DefaultContextByteBudget;
    private long presentedThrough;
    // One target owns all derived thresholds: result band T/8, stub/call band
    // T/16, and opt-in immediate result threshold T/100. Zero disables policy.
    public partial int ContextByteBudget();
    public partial void SetContextByteBudget(int bytes);
    public partial (int ResultsBytes, int CallsBytes) ToolCompactionBudgets();
    // Immutable proposed cuts for the batch in the just-consumed request.
    // They are not History until all companion calls have recorded outcomes.
    private ImmutableArray<Event> pendingResultStubs = [];
    internal partial ImmutableArray<Event> SelectImmediateCuts(Context context, long requestSequence, int threshold);
    private partial void StageKeepResults(ToolCallPart call, ToolResultPart result);
    private partial void ApplyResultStubs();
    private partial void CompactTools();
    internal static partial Event? SelectToolCut(Context context, long presentedThrough, int budget, Redaction level);
    private int capturedToolLimit;
    public partial int MaxToolRounds();
    public partial void SetMaxToolRounds(int value);
    private string responseText = "";

    // Transient immutable transfer envelope. Completion sources belong to the
    // caller's request, not the persistent object graph. Exactly one terminal
    // result is published; cancelling a caller's wait never cancels its work.
    private enum MailKind { Submit, Input, DirectTool, ReverseTool, RegisterMcpTools, CommitSkill, InitializeSkills, Delta, HttpDone, ToolDone, AutomaticDone, ControlDone, Flush, Shutdown, Wake, SelectModel, WorkStatus, IncomingQuestion }
    private sealed record Mail(MailKind Kind, ImmutableArray<Part> Parts = default,
        AgentInput? Input = null, ToolCallPart? Call = null, string? Model = null,
        TaskCompletionSource<string>? Reply = null,
        TaskCompletionSource<ToolResultPart>? ToolReply = null,
        ImmutableArray<Event> Events = default, ToolResultPart? Result = null,
        Exception? Error = null, ParsedResponse? Parsed = null, Chunk? Chunk = null, McpClient? Client = null, ImmutableArray<AutomaticResult> AutomaticResults = default, SkillActivation? Activation = null, TaskCompletionSource<ImmutableArray<string>>? SkillReply = null,
        long WorkGeneration = 0, TaskCompletionSource<AgentWorkStatus>? StatusReply = null, QuestionInfo? Question = null);

    // Human session control, not a model tool. Reject busy/queued work; retain
    // conversation, Engine, meter, lens, provider, surface and funding route.
    public partial Task SetModelAsync(string model);
    private partial void SelectModel(Mail message);

    // Reliable request reply. The actor remains alive after a turn completes.
    public partial Task<string> SubmitAsync(IEnumerable<Part> parts);
    // Controls acknowledge acceptance independently of the active turn's reply.
    // Chat alone returns the answer when it starts a turn; busy Chat returns ""
    // after recording its hint. The actor decides, not an off-thread idle check.
    public partial Task<string> PostAsync(AgentInput input);
    // A sequential channel facade over SubmitAsync, never a second agent loop.
    // Cancellation stops this facade's observation, not the Agent or its jobs.
    public partial IAsyncEnumerable<string> RunAsync(ChannelReader<string> input,
        CancellationToken cancellation = default);
    // Shared bounded progress stream: competing waiters may consume each other's
    // observations. Use SubmitAsync for lossless request completion instead.
    public partial Task<Observation> WaitAsync(Func<Observation, bool> predicate,
        CancellationToken cancellation = default);
    // Wait for quiescence of turns and foreground controls, not background Jobs.
    // Does not close admission or resume paused work; later inputs may extend it.
    public partial Task FlushAsync(CancellationToken cancellation = default);
    // Forceful lifetime teardown: close admission, cancel active work and owned
    // jobs, then join the actor and descendants. This is not graceful draining.
    // The shutdown_agent tool first awaits FlushAsync; kill_agent skips that wait.
    // Repeated calls join the same lifetime; no actor can be resurrected.
    public partial Task ShutdownAsync();

    // Every public submission publishes one admission outcome, including refusal.
    private partial void ObserveInput(Mail message, bool accepted);
    private partial void ValidateInput(Mail message);
    private partial void Enqueue(Mail message);
    private partial Task ActorLoop();
    private partial void Receive(Mail message);
    private partial void Advance();
    private partial bool BeginAutomaticCollection();
    private partial Task AutomaticWorker(ImmutableArray<(McpToolDefinition Definition, Job Job)> jobs,
        CancellationToken cancellation);
    private partial void ReceiveAutomatic(Mail message);
    private partial void Finish(Exception? error = null);
    private partial void Record(Event fact, ImmutableArray<int> partIDs = default);
    internal partial void Publish(Observation value);
    private partial Task TransportWorker(string body, CancellationToken cancellation);
    private partial Task ToolWorker(Task<ToolResultPart> work);
    private partial Task ControlWorker(Mail request, Task<ToolResultPart> work);
    private partial Task StartupWorker(Mail request, Task work);
    private partial void SkipRemaining(string reason);
}
