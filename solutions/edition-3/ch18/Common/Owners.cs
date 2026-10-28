using System.Collections.Immutable;
using System.Text.Json.Nodes;

namespace Ensemble;

/// <summary>Application root. Strongly owns Agents and supplies diagnostics.
/// The caller owns the diagnostic writer (for example Console.Error).</summary>
public sealed partial class Ensemble
{
    private readonly TextWriter diagnostics;
    private readonly TextWriter? apiTrace;
    private readonly object traceGate = new();
    private readonly object handleGate = new();
    private int nextJobHandle = 1;
    private readonly List<Agent> agents;
    private bool destroyed;

    // Begin with no Agents; the supplied writer remains owned by the caller.
    public partial Ensemble(TextWriter diagnostics, TextWriter? apiTrace = null);
    // Optional sanitized wire-payload trace. The caller owns the writer; Engine
    // reaches this facility through Agent, rather than copying a logging service.
    internal partial void Trace(Agent origin, string direction, string payload);

    // A membership snapshot cannot be used to edit the relationship list.
    public partial ImmutableArray<Agent> Agents();

    // Names are optional for standalone Agents. Named lookup uses this actual
    // ownership collection; there is no second global Agent registry.
    public partial Agent? FindAgent(string name);

    // Validate first, then update both ends. Remove detaches without destroying.
    internal partial void AddAgent(Agent child);
    internal partial void RemoveAgent(Agent child);

    // Children reach diagnostics through their parents, never a copied writer.
    public partial void Log(string message);
    // Handles identify work across all Agents owned by this Ensemble.
    internal partial int AllocateJobHandle();

    // Destroy every owned Agent, even if one cleanup fails. Repeated calls are safe.
    public partial void Destroy();
}

/// <summary>Strong child of Ensemble and strong parent of History, Engine, Jobs and Tools.
/// Configuration lives here; each relationship is maintained at both ends.</summary>
public sealed partial class Agent
{
    // Only Ensemble's relationship methods write this back-reference.
    internal Ensemble? parent;
    // One authoritative immutable selection, replaced only by the actor.
    private Configuration configuration;
    // Also protects direct public Engine sends from a between-turn selection.
    // The ingress lock guards this count and replacement of configuration.
    private int engineSends;
    internal partial Configuration BeginEngineSend();
    internal partial void EndEngineSend();
    public readonly string? Name;
    private History? history;
    private Jobs? jobs;
    private Skills? skills;
    private Engine? engine;
    internal bool destroyed;

    // Register with Ensemble and construct History, Engine and Jobs. Failed construction
    // tears down any children already made and removes the Agent from Ensemble.
    public partial Agent(Ensemble parent, Configuration configuration, string? outputDirectory = null, string? name = null, SkillSpec? skillSpec = null);
    public partial Jobs Jobs();
    internal partial void AddJobs(Jobs child);
    internal partial void RemoveJobs(Jobs child);

    // These access paths fail when their relationship is no longer present.
    public partial Ensemble Parent();
    public partial Configuration Configuration();
    public partial History History();
    public partial Engine Engine();

    // Each operation completes both ends before returning, or changes neither
    // on refusal. Removing a child does not dispose it or mark it destroyed.
    internal partial void AddHistory(History child);
    internal partial void RemoveHistory(History child);
    internal partial void AddEngine(Engine child);
    internal partial void RemoveEngine(Engine child);

    // Record the human input, then attempt one exchange. Unlike Chapter 1's
    // completed-pair list, History retains admitted input even if the send fails.
    public partial Task<string> SendAsync(string text, CancellationToken cancellation = default);

    // Replace the one-request status slot; empty text clears it. This is not
    // an update to the stable primary instructions and not a queue of notices.
    public partial void SetEphemeral(string text);

    // Release both children and detach from Ensemble, including on cleanup error.
    public partial void Destroy();
}

/// <summary>Agent-owned canonical log and its derived context. The optional writer
/// belongs to History; Append writes before publishing new in-memory state.</summary>
public sealed partial class History
{
    // Only Agent's relationship methods write this back-reference.
    internal Agent? parent;
    // Snapshot readers (for example reconnecting GUI clients) never race the
    // actor's append. The same gate covers events, context, and log lifetime.
    private readonly object gate = new();
    private readonly List<Event> events;
    private Context context;
    // Sequence folded into context; retained log length is not sequence identity.
    private long foldedSequence;
    // All writers share this gate. Disk work releases gate (the committed-view
    // lock), so readers can still see the previous complete snapshot.
    private readonly object saveGate = new();
    private string? journalPath;
    private bool appendFaulted;
    private StreamWriter? writer;
    private string? writerPath;
    internal bool destroyed;

    // Construct an empty log and context, then register with the owning Agent.
    public partial History(Agent parent);
    public partial Agent Parent();

    // Immutable values/snapshots: neither grants callers mutation of the log.
    public partial Context View();
    public partial ImmutableArray<Event> Events();
    // Capture facts and their derived view together; neither can be mutated.
    public partial HistorySnapshot Snapshot();

    // Start a versioned diagnostic log, first writing retained startup facts.
    // This truncates the chosen file; save.json, not this log, is recovery state.
    public partial void StartLog(string path);

    // Validate and reduce the whole input before replacing current state.
    // Loading is offline and is refused while an output log is active.
    public partial void Load(string path);
    // Shared structural decoder. Archive saving additionally requires a complete
    // prefix; ordinary diagnostic log loading keeps its existing gap tolerance.
    private static partial ImmutableArray<Event> ReadLog(string path);

    // Write canonical events in order, not the derived or vendor-projected view.
    public partial void Dump(TextWriter destination);

    // Assign sequence/time and reduce; write and flush before publication.
    // A failed write leaves the in-memory state unchanged, but does not promise
    // rollback of a partially written file. Stronger recovery comes later.
    public partial Event Append(Event value);

    // Close the owned writer and detach from Agent even if closing throws.
    public partial void Destroy();

    // Pure, total replay step for known facts: inapplicable payloads return the
    // original Context unchanged. Structural decoding/sequence checks stay strict.
    public static partial Context Reduce(Context before, Event value);
    private static partial string? ReductionProblem(Context before, Event value);
    // Owner path adds safe diagnostics without giving the pure reducer services.
    private partial Context Apply(Context before, Event value);
    private static partial List<Entry> HandoffDialogue(List<Entry> dialogue, Event value);
}

/// <summary>Agent-owned vendor boundary and HTTP lifetime. Render and Parse do not
/// mutate History. Only the Agent applies the immutable facts returned by transport.</summary>
public sealed partial class Engine
{
    // Only Agent's relationship methods write this back-reference.
    internal Agent? parent;
    private readonly HttpClient http;
    internal bool destroyed;

    // Own one HTTP client, then register with Agent. The generation has no
    // blanket deadline; explicit caller cancellation still reaches network I/O.
    public partial Engine(Agent parent);
    public partial Agent Parent();

    // Build a disposable request. Refuse unsupported media or ambiguous calls
    // locally; do not consume ephemera or alter recorded events.
    public partial JsonObject Render(Context context, string requestInstruction = "");

    // Normalize a completed HTTP reply into unnumbered events. Invalid replies
    // become safe error facts, never raw remote bodies or invented completions.
    public partial ImmutableArray<Event> Parse(int status, string body);

    // Send a frozen request; plan HTTP 401 permits one same-source renewal and
    // byte-identical retry before any response facts are admitted. Return unnumbered facts;
    // never read or mutate History. The Agent records RequestSent before calling
    // this operation, then applies the returned facts in mailbox order.
    public partial Task<ParsedResponse> SendAsync(string requestBody, CancellationToken cancellation = default,
        Action<Chunk>? onChunk = null);

    private partial Task<ParsedResponse> SendCoreAsync(Configuration config, string requestBody,
        CancellationToken cancellation, Action<Chunk>? onChunk);

    // Keep credential failures token-free on initial acquisition and 401 renewal.
    private partial Task<string> PlanBearerAsync(string? rejectedBearer, CancellationToken cancellation);

    // Dispose the owned client and detach from Agent; do not destroy the parent.
    public partial void Destroy();
}
