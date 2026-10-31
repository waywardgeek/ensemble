using System.Collections.Immutable;
using System.Diagnostics;
using System.Text.Json;
using System.Text.RegularExpressions;

namespace Ensemble;

public enum JobStatus
{
    Running, Done, Killed
}

// Value-like recorded report. It cannot restart a process during replay.
public sealed record JobInfo(int Handle, string Status, string OutputPath, long Bytes,
    bool Truncated = false, int? ExitCode = null, string Cwd = "");

// Value-like observation policy. Delay wakes the caller; it never stops work.
public sealed record JobLimits(double Delay = 3, string Pattern = "", int MaxBytes = 16384);

// Jobs strongly owns every admitted Job until explicit destruction. The lock
// protects membership and pending observation settings; it is never held while work drains.
public sealed partial class Jobs
{
    internal Agent? parent;
    internal bool destroyed;
    private readonly object gate = new();
    private readonly List<Job> children = [];
    private JobLimits? nextLimits;
    public readonly string OutputDirectory;

    // Create output through the owning Agent's Sandbox, then register both ends.
    // Confined defaults live under its root; explicit outside paths refuse.
    public partial Jobs(Agent parent, string? outputDirectory = null);
    public partial Agent Parent();
    public partial ImmutableArray<Job> Children();
    // Membership changes validate first and complete both relationship ends.
    internal partial void AddJob(Job child);
    internal partial void RemoveJob(Job child);
    // Only jobs belonging to this Agent can be found, even though Ensemble
    // allocates globally unique handles across its Agents.
    public partial Job Find(int handle);
    // Consume pending options, record the call, then permit effects. Controls
    // consume options too, but neither reserve handles nor produce job records.
    // Actor-only admission. Automatic work uses normal Job lifetime/output, but
    // is never admitted by a guessed tool name or an ordinary model call.
    internal partial Job StartAutomatic(AutomaticTool automatic);
    internal partial Task<ToolResultPart> DispatchAsync(ToolCallPart call, CancellationToken cancellation = default, string? refusal = null);
    // Validate before replacing the one pending observation override.
    public partial void SetLimits(JobLimits limits);
    // Cap current observers at the default delay measured from their start.
    // Execution is untouched. Later observers also check Agent hint state.
    internal partial void WakeObservers();
    // Stop owned work, close OS resources, detach both relationship ends.
    // Output files intentionally remain available to recorded references.
    public partial void Destroy();
}

// Each Job owns its process, cancellation and output stream. Bytes on disk are
// authoritative; notifications contain no output. A single cursor records what
// observers have already received, independent of canonical conversation events.
public sealed partial class Job
{
    internal Jobs? parent;
    internal bool destroyed;
    private readonly object gate = new();
    private readonly SemaphoreSlim observation = new(1, 1);
    private readonly CancellationTokenSource cancellation = new();
    private readonly FileStream output;
    private readonly ToolCallPart call;
    private TaskCompletionSource changed = new(TaskCreationOptions.RunContinuationsAsynchronously);
    // Replaceable pulse, captured once by the active serialized observer.
    // Unlike output attention, a message wake is not retained for later callers.
    private TaskCompletionSource observationWake = new(TaskCreationOptions.RunContinuationsAsynchronously);
    private Task? worker;
    private JobStatus status = JobStatus.Running;
    private bool isError;
    private bool commandJob;
    // Remote completion is ordered structured content, not a process byte stream.
    // The first terminal observation advances the cursor past this value. Retain
    // it until destruction so automatic collection can read it independently;
    // later ordinary observations still return only unseen byte output.
    private ToolResultPart? structuredResult;
    private long bytes;
    private long cursor;
    // End of the latest urgent progress range. The shared byte cursor consumes
    // attention with its output; this is not a second question/completion store.
    private long attentionEnd;
    private int? exitCode;
    private Process? process;
    private int processGroup;
    public readonly int Handle;
    public readonly string OutputPath;

    // Native libc entry point has no managed body. Only Job uses it to signal
    // the process group created by the system PTY launcher.
    [System.Runtime.InteropServices.DllImport("libSystem.B.dylib", EntryPoint = "kill")]
    private static extern int Signal(int pid, int signal);

    internal partial Job(Jobs parent, int handle, ToolCallPart call);
    public partial Jobs Parent();
    public partial JobInfo Snapshot();
    internal partial void Start(Tool? tool, ToolCallPart sanitized, Exception? admissionError,
        McpClient? automaticClient = null);
    // Wait for the real worker; cancellation detaches this wait, not the Job.
    internal partial Task WaitForCompletionAsync(CancellationToken cancellation);
    // Wait for completed output without consuming the public observation cursor.
    internal partial Task<ToolResultPart> CompletionAsync(CancellationToken cancellation);
    // Custom handlers may stream progress. Publication flushes disk before waking.
    public partial void AppendOutput(string text);
    // Actionable progress must yield even with a nonmatching observer pattern.
    // Flush the text before marking its unread byte range for immediate delivery.
    internal partial void AppendAttention(string text);
    internal partial void WakeObservation();
    // Bootstrap publishes its process group before executing the command. The
    // host consumes it once; it never becomes model-visible output or a file.
    private partial Task<int> ReadProcessGroupAsync(Stream source, CancellationToken token);
    private partial Task PumpAsync(Stream stream);
    private partial void StopProcess();
    internal partial void AppendBytes(ReadOnlySpan<byte> data);
    // Serialize observers and advance the byte cursor only when returning a
    // report. Caller cancellation leaves the job alive and its cursor intact.
    public partial Task<ToolResultPart> ObserveAsync(string callID, double delay = 3,
        string pattern = "", int maxBytes = 16384, CancellationToken cancellation = default, bool wakeOnOutput = true);
    // Submit one newline-terminated input, then observe only fresh output.
    public partial Task<ToolResultPart> SendInputAsync(string callID, string text,
        double delay = 3, string pattern = "", int maxBytes = 16384, CancellationToken cancellation = default);
    // Record explicit termination, stop the process group, and drain it.
    // Managed callbacks are cancelled cooperatively; late output is ignored.
    public partial Task KillAsync(string reason = "kill_job");
    internal partial Task<string> RunCommandAsync(JsonElement args, CancellationToken cancellation);
    // Stop owned work, close OS resources, detach both relationship ends.
    // Output files intentionally remain available to recorded references.
    public partial void Destroy();
}
