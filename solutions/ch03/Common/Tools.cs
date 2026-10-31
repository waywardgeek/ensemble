using System.Collections.Immutable;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

// The fixed Chapter 3 capabilities. This is deliberately not an extensibility
// framework: a Tool is an Agent-owned object with one concrete implementation.
public enum ToolKind
{
    ReadFile,
    ListDirectory,
    SearchFiles,
    RunCommand,
    WriteFile,
    EditFile
}

public sealed partial class Tool
{
    internal Agent? parent;
    internal bool destroyed;
    private readonly ToolKind kind;
    private readonly string name;
    private readonly string description;
    private readonly JsonElement schema;

    // Construction registers both links. The definition is immutable and belongs
    // here once; renderers derive fresh JSON trees rather than retaining copies.
    public partial Tool(Agent parent, ToolKind kind);
    public partial Agent Parent();
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

    // Registry mutations complete both directions or leave both unchanged.
    public partial ImmutableArray<Tool> Tools();
    public partial void AddTool(Tool tool);
    public partial void RemoveTool(Tool tool);
    public partial void InstallTools();
    public partial Task<ToolResultPart> DispatchAsync(ToolCallPart call, CancellationToken cancellation = default);
}
