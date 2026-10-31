using System.Collections.Immutable;
using System.Text.Json.Nodes;

namespace Ensemble;

/// <summary>Application root. Strongly owns Agents and supplies diagnostics.
/// The caller owns the diagnostic writer (for example Console.Error).</summary>
public sealed partial class Ensemble
{
    private readonly TextWriter diagnostics;
    private readonly object handleGate = new();
    private int nextJobHandle = 1;
    private readonly List<Agent> agents;
    private bool destroyed;

    // Begin with no Agents; the supplied writer remains owned by the caller.
    public partial Ensemble(TextWriter diagnostics);

    // A membership snapshot cannot be used to edit the relationship list.
    public partial ImmutableArray<Agent> Agents();

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
    private readonly Configuration configuration;
    private History? history;
    private Jobs? jobs;
    private Engine? engine;
    internal bool destroyed;

    // Register with Ensemble and construct History, Engine and Jobs. Failed construction
    // tears down any children already made and removes the Agent from Ensemble.
    public partial Agent(Ensemble parent, Configuration configuration, string? outputDirectory = null);
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
    private readonly List<Event> events;
    private Context context;
    private StreamWriter? writer;
    internal bool destroyed;

    // Construct an empty log and context, then register with the owning Agent.
    public partial History(Agent parent);
    public partial Agent Parent();

    // Immutable values/snapshots: neither grants callers mutation of the log.
    public partial Context View();
    public partial ImmutableArray<Event> Events();

    // Create a versioned output log before any event is recorded. This truncates
    // the chosen file; it is not a resume, backup, or crash-recovery operation.
    public partial void StartLog(string path);

    // Validate and reduce the whole input before replacing current state.
    // Loading is offline and is refused while an output log is active.
    public partial void Load(string path);

    // Write canonical events in order, not the derived or vendor-projected view.
    public partial void Dump(TextWriter destination);

    // Assign sequence/time and reduce; write and flush before publication.
    // A failed write leaves the in-memory state unchanged, but does not promise
    // rollback of a partially written file. Stronger recovery comes later.
    public partial Event Append(Event value);

    // Close the owned writer and detach from Agent even if closing throws.
    public partial void Destroy();

    // Pure replay step: no I/O, clock, or mutation of the input values.
    public static partial Context Reduce(Context before, Event value);
}

/// <summary>Agent-owned vendor boundary and HTTP lifetime. Render and Parse do not
/// mutate History; only the live exchange appends their events via its parent.</summary>
public sealed partial class Engine
{
    private const string PrimaryInstructions = "You are a helpful assistant.";
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
    public partial JsonObject Render(Context context);

    // Normalize a completed HTTP reply into unnumbered events. Invalid replies
    // become safe error facts, never raw remote bodies or invented completions.
    public partial ImmutableArray<Event> Parse(int status, string body);

    // Render, record the attempt, send once, and append normalized events before
    // returning text. No tools, retries, or streaming are performed here.
    public partial Task<string> ExchangeAsync(CancellationToken cancellation = default);

    // Dispose the owned client and detach from Agent; do not destroy the parent.
    public partial void Destroy();
}
