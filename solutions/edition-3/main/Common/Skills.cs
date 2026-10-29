using System.Collections.Immutable;

namespace Ensemble;

// Value-like construction inputs. Library code never reads process environment.
public sealed record SkillSpec(string Directory, string Primary,
    IReadOnlyDictionary<string, string>? Variables = null,
    IReadOnlyList<string>? StartupSkills = null,
    IReadOnlyDictionary<string, string>? McpEnvironment = null);
// Value-like, short-lived construction snapshot. It contains immutable templates
// and explicit values, never live grants, clients, owners or service references.
internal sealed record SkillBootstrap(string Directory, string Primary, ImmutableArray<string> Startup,
    ImmutableDictionary<string, SkillDefinition> Definitions,
    ImmutableDictionary<string, string> Variables,
    ImmutableDictionary<string, string> Credentials);

public enum SkillType { Primary, Loadable, Dependency }

// Templates grant nothing. Activation creates a separate historical snapshot;
// later variable changes cannot rewrite instructions already given to a model.
public sealed record SkillDefinition(string Name, string Description, SkillType Type,
    ImmutableArray<string> Tools, ImmutableArray<string> Depends,
    ImmutableArray<string> Loadable, string Body, ImmutableArray<McpServerSpec> McpServers = default);
public sealed partial class SkillPart : Part
{
    public string Name { get; }
    public string Body { get; }
    public ImmutableArray<string> Tools { get; }
    public bool Primary { get; }
    // Pending is reduced conversation state, not a query over retained diagnostics.
    public bool Pending { get; }
    public partial SkillPart(string name, string body, IEnumerable<string> tools, bool primary = false, bool pending = false);
}
public readonly record struct ToolSource(string Skill, long Sequence);

public sealed partial class Agent
{
    public partial Skills? SkillManager();
    public partial void AddSkills(Skills child);
    public partial void RemoveSkills(Skills child);
    public partial bool ToolEnabled(string name);
}

// Strong child of Agent. History, not this manager, owns active instructions,
// capability grants and pending-unload markers. Definitions are only templates.
public sealed partial class Skills
{
    internal Agent? parent;
    internal bool destroyed;
    // Host-selected documentation provenance. It is not Agent file-tool authority.
    // Templates remain immutable even when recall observes newer manuals on disk.
    private readonly string sourceDirectory;
    private Microsoft.Win32.SafeHandles.SafeFileHandle? recallRoot;
    private readonly string primaryName;
    // Frozen host selection, not a second active-skill registry. History owns activation.
    private readonly ImmutableArray<string> startupSkills;
    // A server-backed startup closure freezes creation-time variables before any await.
    private ImmutableDictionary<string, string>? initialVariables;
    private Task? initialization;
    private readonly object gate = new();
    private readonly Dictionary<string, SkillDefinition> definitions;
    private readonly Dictionary<string, string> variables;
    // Only referenced auth-env values; never manual variables or saved context.
    private readonly Dictionary<string, string> mcpCredentials = new(StringComparer.Ordinal);
    // Connections have two strong owners: this skill manager and the MCP manager.
    private readonly List<McpClient> clients = [];
    public partial ImmutableArray<McpClient> Clients();
    public partial void AddClient(McpClient child);
    public partial void RemoveClient(McpClient child);
    internal partial bool Grants(string skill, string tool);
    internal partial Task<ToolResultPart> ControlAsync(ToolCallPart call, CancellationToken cancellation);
    private partial Task<ImmutableArray<string>> PrepareAsync(string name,
        ImmutableArray<SkillDefinition> additions, ImmutableDictionary<string, string> values,
        Mcp manager, string workspace, CancellationToken cancellation, bool initial = false);
    internal partial ImmutableArray<string> Commit(SkillActivation activation);
    private partial Task<ToolResultPart> StopClientsAsync(string name, ToolResultPart result);

    // Validate and attach; local-only startup selections publish synchronously. Server-backed
    // selections require InitializeAsync after persistence load and before user work.
    // Failed construction detaches; failed async initialization requires destruction.
    public partial Skills(Agent parent, SkillSpec specification);
    internal partial Skills(Agent parent, SkillBootstrap bootstrap);
    internal partial SkillBootstrap CaptureBootstrap();
    // Lazy read-only root is owned by Skills and closed by Destroy. Descendants
    // inherit the selected path, not this descriptor or any active MCP client.
    internal partial ImmutableArray<(string File, string Text, string Hash)> RecallFiles(CancellationToken cancellation);
    private partial void ReadRecallDirectory(Microsoft.Win32.SafeHandles.SafeFileHandle directory,
        string relative, List<(string File, string Text, string Hash)> files, Configuration policy, CancellationToken cancellation);
    private partial void Attach(Agent parent);
    public partial Agent Parent();
    /// <summary>Join the initial skill batch before admitting user work. Cancellation
    /// detaches only this waiter; Agent shutdown cancels and joins preparation.</summary>
    public partial Task InitializeAsync(CancellationToken cancellation = default);
    public partial bool Ready();
    // An initial closure with a WebSocket server needs a browser registration
    // before initialization. Stdio-only startup needs no GUI connection.
    public partial bool RequiresBrowserStartup();
    internal partial Task BeginInitialization(CancellationToken cancellation);
    // Detach the manager, not the Agent or recorded activation values. Idempotent.
    public partial void Destroy();
    // Parse only the documented front-matter subset; retain the remaining body verbatim.
    public static partial SkillDefinition Parse(string text, string path);
    private static partial ImmutableArray<McpServerSpec> ParseServers(string[] lines, ref int index, int end, string path);
    // Affect future rendering only; never rewrite an already recorded manual.
    public partial void RegisterVariable(string name, string value);
    // Read immutable activation facts from History, including pending manuals.
    public partial ImmutableArray<SkillPart> Active();
    // Progressive discovery: offered by active manuals and not already active.
    public partial ImmutableArray<string> LoadableSkills();
    // Reachable declarations alone never grant dispatch.
    // Reachable declarations for optional deferral; this does not enable dispatch.
    public partial ImmutableArray<string> CatalogTools(Context? context = null);
    private partial ImmutableArray<string> UniversalControls();
    // Admission permission, independent of whether a deferred schema is advertised.
    public partial bool Enabled(string tool);
    // Earliest still-active grant; pending manuals no longer supply authority.
    public partial ToolSource? Source(string tool);
    // Manuals remain pending; all grants revoke and external connections stop at unload.
    public partial bool Pending(string skill);
    // Local-only activation is synchronous on actor admission. Server-backed
    // activation returns to the actor through Commit before recording grants.
    private partial ImmutableArray<SkillPart> Activate(string name, bool initial);
    private partial List<SkillDefinition> Resolve(string name, bool initial, IEnumerable<string>? preceding = null);
    // Validate the whole ordered startup selection before any instruction publication.
    private partial List<SkillDefinition> ResolveStartup();
    private partial ImmutableArray<string> Available(IEnumerable<string> activeNames);
    private partial string Render(string body, IEnumerable<string> tools, IEnumerable<string> available,
        IReadOnlyDictionary<string, string>? values = null);
}

// Short-lived mailbox transfer, not another owner. Clients remain attached to
// their real parents until commit succeeds or the preparing worker joins cleanup.
internal sealed record SkillActivation(Skills Owner, string Name,
    ImmutableArray<SkillDefinition> Additions, ImmutableDictionary<string, string> Variables,
    ImmutableArray<McpClient> Clients, CancellationToken Cancellation, bool Initial = false);

public sealed partial class Agent
{
    internal partial Task<ImmutableArray<string>> CommitSkillAsync(SkillActivation activation);
    internal partial Task InitializeSkillsAsync();
}
