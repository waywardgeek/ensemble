using System.Collections.Immutable;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

// Built-ins remain concrete operations. Custom is the chapter's explicit
// external-assembly extension boundary, not internal service wiring.
public enum ToolKind
{
    ReadFile,
    ListDirectory,
    SearchFiles,
    RunCommand,
    WriteFile,
    EditFile,
    WaitForJob,
    SendInput,
    KillJob,
    ToolLimits,
    LoadSkill,
    UnloadSkill,
    Custom,
    Remote
}

public sealed partial class Tool
{
    internal Agent? parent;
    // A remote Tool has two strong parents: the Agent and its connection.
    internal McpClient? remote;
    internal bool destroyed;
    private readonly object destructionGate = new();
    private readonly ToolKind kind;
    private readonly bool waitForCompletion;
    private readonly string name;
    private readonly string description;
    private readonly JsonElement schema;
    private readonly Func<Job, JsonElement, CancellationToken, Task<string>>? handler;

    // Construction registers both links. The definition is immutable and belongs
    // here once; renderers derive fresh JSON trees rather than retaining copies.
    public partial Tool(Agent parent, ToolKind kind);
    // A custom callback receives the real Job, so logging follows its parents.
    // Opted-in channel operations return one atomic terminal result instead of
    // a timed running report. They remain independently owned, cancellable Jobs.
    public partial Tool(Agent parent, string name, string description, JsonElement schema,
        Func<Job, JsonElement, CancellationToken, Task<string>> handler, bool waitForCompletion = false);
    internal partial Tool(Agent parent, McpClient client, McpToolDefinition definition);
    public partial McpClient? Remote();
    internal partial bool HasStructuredResult();
    internal partial bool WaitsForCompletion();
    public partial Agent Parent();
    public partial bool IsControl();
    // Admission validates the immutable declared argument schema before any
    // started fact or handler effect. Invocation does not repeat this check.
    internal partial void Validate(JsonElement args);
    internal partial Task<ToolResultPart> ExecuteJobAsync(Job job, ToolCallPart call, CancellationToken cancellation);
    public partial string Name();
    public partial JsonObject Declaration(Vendor vendor);

    // Expected argument and operating-system failures become IsError results.
    // Cancellation is not a fabricated tool failure; it propagates to the caller.
    public partial Task<ToolResultPart> ExecuteAsync(ToolCallPart call, CancellationToken cancellation = default);
    public partial void Destroy();
}

public sealed partial class Agent
{
    private readonly List<Tool> tools = [];
    private readonly object toolsGate = new();

    // Registry mutations complete both directions or leave both unchanged.
    public partial ImmutableArray<Tool> Tools();
    public partial void AddTool(Tool tool);
    public partial void RemoveTool(Tool tool);
    public partial void InstallTools();
    public partial Task<ToolResultPart> DispatchAsync(ToolCallPart call, CancellationToken cancellation = default);
}
