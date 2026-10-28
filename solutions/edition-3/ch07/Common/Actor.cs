using System.Collections.Immutable;
using System.Threading.Channels;

namespace Ensemble;

// Value-like public input. Live messages become hints when a turn is active;
// SubmitAsync, in contrast, always reserves a separate request and reply.
public enum InputKind { Message, Hint, Pause, Resume, InterruptNow, InterruptAfterTool, Ephemeral }
public sealed record AgentInput(InputKind Kind, string Text = "");

// Observations are disposable progress, not acknowledgments. Part carries the
// full neutral value; Seq identifies its durable event, not a second transcript.
public enum ObservationKind { Input, StateChanged, PartDelta, PartFinal, TurnEnded, EventRecorded }
public sealed record Observation(ObservationKind Kind, long Seq = 0, Part? Part = null,
    string Text = "", Phase? From = null, Phase? To = null, bool? Accepted = null, Chunk? Delta = null,
    int? PartID = null, long ResponseID = 0, bool? Succeeded = null, bool HasReply = false, Event? Fact = null);
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
    // The ingress lock protects lazy startup and closing admission only. All
    // turn scheduling fields below are touched exclusively by the actor task.
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
    private readonly HashSet<TaskCompletionSource<ToolResultPart>> controlReplies = [];
    private readonly CancellationTokenSource controlCancellation = new();
    private Mail? active;
    private bool paused;
    private bool stopping;
    private bool workerPending;
    private bool transportPending;
    private bool toolBeforeText;
    private long responseID;
    private bool interruptPending;
    private bool discardResponse;
    private bool skipTail;
    private bool hintPending;
    private CancellationTokenSource? observationCancellation;
    private ImmutableArray<ToolCallPart> calls = [];
    private int nextCall;
    private int batches;
    private string responseText = "";

    // Transient immutable transfer envelope. Completion sources belong to the
    // caller's request, not the persistent object graph. Exactly one terminal
    // result is published; cancelling a caller's wait never cancels its work.
    private enum MailKind { Submit, Input, DirectTool, Delta, HttpDone, ToolDone, ControlDone, Flush, Shutdown, Wake }
    private sealed record Mail(MailKind Kind, ImmutableArray<Part> Parts = default,
        AgentInput? Input = null, ToolCallPart? Call = null,
        TaskCompletionSource<string>? Reply = null,
        TaskCompletionSource<ToolResultPart>? ToolReply = null,
        ImmutableArray<Event> Events = default, ToolResultPart? Result = null,
        Exception? Error = null, ParsedResponse? Parsed = null, Chunk? Chunk = null);

    // Reliable request reply. The actor remains alive after a turn completes.
    public partial Task<string> SubmitAsync(IEnumerable<Part> parts);
    // Controls acknowledge acceptance independently of the active turn's reply.
    public partial Task PostAsync(AgentInput input);
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
    // Close admission, settle active work, terminate owned jobs, then join the
    // actor. Repeated calls join the same lifetime; no actor can be resurrected.
    public partial Task ShutdownAsync();

    // Every public submission publishes one admission outcome, including refusal.
    private partial void ObserveInput(Mail message, bool accepted);
    private partial void ValidateInput(Mail message);
    private partial void Enqueue(Mail message);
    private partial Task ActorLoop();
    private partial void Receive(Mail message);
    private partial void Advance();
    private partial void Finish(Exception? error = null);
    private partial void Record(Event fact, ImmutableArray<int> partIDs = default);
    internal partial void Publish(Observation value);
    private partial Task TransportWorker(string body, CancellationToken cancellation);
    private partial Task ToolWorker(Task<ToolResultPart> work);
    private partial Task ControlWorker(Mail request, Task<ToolResultPart> work);
    private partial void SkipRemaining(string reason);
}
