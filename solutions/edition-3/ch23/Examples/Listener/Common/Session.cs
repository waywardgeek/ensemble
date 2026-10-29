using System.Diagnostics;
using System.Net.WebSockets;
using System.Text.Json;
using System.Text.Json.Serialization;
using Ensemble;
using Application = Ensemble.Ensemble;

namespace SpeechListener;

// Value-like semantic history. Heard means speech-channel admission, not hearing
// through a microphone. No model response, raw tool output or DOM text belongs here.
public sealed record Entry(
    [property: JsonPropertyName("seq")] int Seq,
    [property: JsonPropertyName("kind")] string Kind,
    [property: JsonPropertyName("source")] string Source,
    [property: JsonPropertyName("text")] string Text);

/// <summary>Application composition and owned I/O lifetime. The core root owns
/// the listener Agent, which owns exactly five custom Tools. The application owns
/// the target process, browser process/profile, and DevTools connection. These are
/// boundary resources, not additional Agent services or model-visible capabilities.</summary>
internal sealed partial class Session
{
    private readonly Application application;
    private readonly string directory;
    private readonly HttpClient http = new();
    private readonly ClientWebSocket socket = new();
    private readonly SemaphoreSlim protocol = new(1);
    private readonly SemaphoreSlim operations = new(1);
    private readonly List<Entry> history = [];
    private Process? target;
    private Process? browser;
    private Task<string>? targetOutput;
    private Task<string>? targetError;
    private Task<string>? browserOutput;
    private Task<string>? browserError;
    private int rpcID;
    private long speechSequence;
    private int heardCursor;
    private Task? destruction;
    private readonly object destructionGate = new();

    public partial Session(Configuration configuration, string directory);
    public partial Agent Listener();
    // Launch only caller-selected executables in a fresh caller-selected directory.
    // Target credentials are supplied explicitly; the browser gets no credentials.
    public partial Task StartAsync(string dotnet, string gui, string chrome,
        IReadOnlyDictionary<string, string> targetEnvironment, CancellationToken cancellation);
    public partial Task DestroyAsync();
    private partial Task DestroyCoreAsync();
    private partial Process Launch(string executable, IEnumerable<string> arguments,
        IReadOnlyDictionary<string, string> environment);
    private static partial Task StopAsync(Process? process);
    private partial Task<JsonElement> CommandAsync(string method, object parameters, CancellationToken cancellation);
    private partial Task<JsonElement> EvaluateAsync(string expression, CancellationToken cancellation);
    private partial Task<string> ExecuteAsync(Job job, string name, JsonElement arguments, CancellationToken cancellation);
    private partial Task<bool> CollectAsync(CancellationToken cancellation);
    private partial void Install();
    private static partial double Seconds(JsonElement arguments, string name);
}
