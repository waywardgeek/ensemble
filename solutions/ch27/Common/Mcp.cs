using System.Collections.Immutable;
using System.Diagnostics;
using System.Net.WebSockets;
using System.Text.Json;

namespace Ensemble;

// A replaceable I/O boundary, not a copied Agent service. One message is one
// JSON-RPC envelope. The client owns its transport and closes it on destruction.
// Closing must release a blocked receive; sends may overlap at the caller.
public interface IMcpTransport : IAsyncDisposable
{
    Task SendAsync(JsonElement message, CancellationToken cancellation = default);
    Task<JsonElement?> ReceiveAsync(CancellationToken cancellation = default);
}

// Newline-delimited UTF-8 JSON over a caller-supplied duplex boundary. The
// transport takes ownership of both streams, which may be the same instance.
public sealed partial class McpStreamTransport : IMcpTransport
{
    private readonly Stream input;
    private readonly Stream output;
    private readonly StreamReader reader;
    private readonly SemaphoreSlim sending = new(1, 1);
    private readonly CancellationTokenSource closing = new();
    private int closed;
    public partial McpStreamTransport(Stream input, Stream output);
    public partial Task SendAsync(JsonElement message, CancellationToken cancellation = default);
    public partial Task<JsonElement?> ReceiveAsync(CancellationToken cancellation = default);
    public partial ValueTask DisposeAsync();
}

// Value-like launch configuration from a skill, not a live connection or grant.
// Arguments are literal argv elements; no shell performs expansion.
public sealed record McpServerSpec(string Name, string Transport, string? Command,
    ImmutableArray<string> Arguments, string? Url = null, string? AuthEnvironment = null);

// Owns one subprocess, its protocol streams, and its stderr drain. The stream
// transport handles framing; this boundary additionally joins process lifetime.
public sealed partial class McpProcessTransport : IMcpTransport
{
    private readonly Process process;
    private readonly McpStreamTransport wire;
    private readonly Task stderrDrain;
    private readonly object gate = new();
    private Task? destruction;
    // Workspace is explicit and applies only to this child process.
    public partial McpProcessTransport(McpServerSpec specification, string workspace);
    public partial Task SendAsync(JsonElement message, CancellationToken cancellation = default);
    public partial Task<JsonElement?> ReceiveAsync(CancellationToken cancellation = default);
    public partial ValueTask DisposeAsync();
    private partial Task DestroyCoreAsync();
    private static partial Task DrainAsync(Stream stream);
}

// Owns an already-connected WebSocket. HTTP upgrade and browser identity belong
// to the hosting application, not to JSON-RPC correlation or the core Agent.
public sealed partial class McpWebSocketTransport : IMcpTransport
{
    private readonly WebSocket socket;
    private readonly SemaphoreSlim writes = new(1, 1);
    private int disposed;
    public partial McpWebSocketTransport(WebSocket socket);
    public partial Task SendAsync(JsonElement message, CancellationToken cancellation = default);
    public partial Task<JsonElement?> ReceiveAsync(CancellationToken cancellation = default);
    public partial ValueTask DisposeAsync();
}

// Protocol/content conversion is stateless. Unsupported blocks are refused as a
// whole result; no OfType/filter may silently discard an image or error payload.
public static partial class McpContent
{
    public static partial ToolResultPart Decode(string callID, JsonElement result);
    public static partial JsonElement Encode(ToolResultPart result);
}

// Value-like discovery data. A catalog record grants no execution permission.
public sealed record McpToolDefinition(string Name, string Description,
    JsonElement InputSchema, string Ephemeral);

public sealed partial class Agent
{
    private Mcp? mcp;
    public partial Mcp? Mcp();
    // Lazy construction is serialized with Agent admission, so hosting and
    // skill startup cannot publish two managers for the same owning Agent.
    public partial Mcp EnsureMcp();
    // Prefix and discovered annotation classification; never an execution grant.
    internal partial bool IsAutomaticTool(string name);
    // Reverse RPC enters the actor's independent observation lane. It still
    // passes ordinary tool admission; it cannot queue behind its own MCP caller.
    internal partial Task<ToolResultPart> DispatchReverseAsync(ToolCallPart call,
        CancellationToken cancellation);
    internal partial Task RegisterMcpToolsAsync(McpClient client, CancellationToken cancellation);
    public partial void AddMcp(Mcp child);
    public partial void RemoveMcp(Mcp child);
}

// Agent strongly owns this manager; the manager strongly owns its clients.
// Clients never copy Agent, History, Jobs, or logging services into fields.
public sealed partial class Mcp
{
    internal Agent? parent;
    internal bool destroyed;
    private readonly object gate = new();
    private readonly List<McpClient> clients = [];
    private McpBrowser? browser;
    private TaskCompletionSource browserChanged = new(TaskCreationOptions.RunContinuationsAsynchronously);
    private Task? destruction;
    public partial Mcp(Agent parent);
    public partial Agent Parent();
    public partial ImmutableArray<McpClient> Clients();
    // At most one browser target. Registration is not discovery or a tool grant.
    public partial McpBrowser? Browser();
    // Wait for this owner's actual browser relationship. Cancellation abandons
    // only this observer; availability is not a lease or a tool grant.
    public partial Task WaitForBrowserAsync(CancellationToken cancellation = default);
    private partial void SignalBrowserChanged();
    public partial void AddBrowser(McpBrowser child);
    public partial void RemoveBrowser(McpBrowser child);
    internal partial McpClient ClaimBrowser(Skills owner, string skill, string server);
    public partial void AddClient(McpClient child);
    public partial void RemoveClient(McpClient child);
    public partial Task DestroyAsync();
    private partial Task DestroyCoreAsync();
    // Published prefixes identify cadence. Selection additionally checks the
    // live granting skill; merely discovering a tool grants no authority.
    internal static partial string AutomaticFrequency(string name);
    internal partial ImmutableArray<AutomaticTool> AutomaticTools(bool includeTurn);
}

// Value-like, short-lived selections/outcomes, not additional owners or registries.
internal sealed record AutomaticTool(McpClient Client, McpToolDefinition Definition);
internal sealed record AutomaticResult(McpToolDefinition Definition, ToolResultPart Result);

// One connection owns one reader and a reply table. Reverse requests execute
// outside the reader so an outbound RPC can still receive its correlated reply.
public sealed partial class McpClient
{
    internal Mcp? parent;
    internal bool destroyed;
    public readonly string Name;
    internal Skills? skillOwner;
    public readonly string? SkillName;
    public partial Skills? SkillOwner();
    internal readonly IMcpTransport transport;
    internal McpBrowser? browserOwner;
    public partial McpBrowser? BrowserOwner();
    private readonly object gate = new();
    private readonly CancellationTokenSource lifetime = new();
    private readonly Dictionary<long, TaskCompletionSource<JsonElement>> pending = [];
    private readonly HashSet<Task> reverseCalls = [];
    private Task? readerTask;
    private Task? connection;
    private Task? destruction;
    private long nextID;
    private bool ended;
    private bool serving;
    private ImmutableArray<McpToolDefinition> catalog = [];
    private readonly List<Tool> tools = [];
    // Construction attaches only; ConnectAsync performs handshake and discovery.
    public partial McpClient(Mcp parent, string name, IMcpTransport transport,
        Skills? skillOwner = null, string? skillName = null, McpBrowser? browserOwner = null);
    public partial Mcp Parent();
    public partial ImmutableArray<McpToolDefinition> Catalog();
    public partial Task ConnectAsync(CancellationToken cancellation = default);
    // Incoming stdin-style service, not an outgoing discovery connection.
    // Completes after normal EOF, or faults on malformed protocol. Cancellation
    // ends this connection; callers still explicitly destroy/join its owner.
    public partial Task ServeAsync(CancellationToken cancellation = default);
    private partial Task ServeCoreAsync(CancellationToken cancellation);
    // Installation enters the Agent mailbox. Discovery alone grants nothing;
    // automatic definitions are never ordinary Tools, even without skills.
    public partial Task InstallToolsAsync(CancellationToken cancellation = default);
    internal partial void InstallTools();
    public partial ImmutableArray<Tool> Tools();
    public partial void AddTool(Tool tool);
    public partial void RemoveTool(Tool tool);
    internal partial Task<ToolResultPart> CallAsync(ToolCallPart call, CancellationToken cancellation);
    // Raw RPC is public for explicit host extensions. It never enables a tool.
    public partial Task<JsonElement> RequestAsync(string method, JsonElement parameters,
        CancellationToken cancellation = default);
    public partial Task NotifyAsync(string method, JsonElement parameters,
        CancellationToken cancellation = default);
    public partial Task DestroyAsync();
    private partial Task ConnectCoreAsync(CancellationToken cancellation);
    private partial Task ReadAsync();
    private partial Task ReverseAsync(JsonElement request);
    private partial void End(Exception error);
    private partial Task DestroyCoreAsync();
}

// Concrete owned HTTP boundary. Each POST has its own lifetime; the receive
// queue joins replies back to the existing client's numeric request identities.
// No GET stream, event resumption, redirects, cookies or automatic retry.
public sealed partial class McpHttpTransport : IMcpTransport
{
    private readonly Uri endpoint;
    private readonly string? bearer;
    private readonly HttpClient http;
    private readonly CancellationTokenSource closing = new();
    private readonly object gate = new();
    private readonly HashSet<Task> posts = [];
    private readonly System.Threading.Channels.Channel<JsonElement> incoming =
        System.Threading.Channels.Channel.CreateBounded<JsonElement>(64);
    private string? session;
    private Task? destruction;
    private bool closed;
    public partial McpHttpTransport(string url, string? bearer = null);
    public partial Task SendAsync(JsonElement message, CancellationToken cancellation);
    public partial Task<JsonElement?> ReceiveAsync(CancellationToken cancellation);
    private partial Task PostAsync(JsonElement message, bool request, CancellationToken cancellation);
    private partial Task ReadReplyAsync(HttpResponseMessage response, JsonElement id, CancellationToken cancellation);
    private partial Task PublishAsync(JsonElement message, CancellationToken cancellation);
    private partial Task<string> ReadErrorAsync(HttpResponseMessage response, CancellationToken cancellation);
    public partial ValueTask DisposeAsync();
    private partial Task DestroyAsync();
}
