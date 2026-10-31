using System.Collections.Immutable;

namespace Ensemble;

// Value-like construction inputs. Library code never reads process environment.
public sealed record SkillSpec(string Directory, string Primary,
    IReadOnlyDictionary<string, string>? Variables = null);
public enum SkillType { Primary, Loadable, Dependency }

// Templates grant nothing. Activation creates a separate historical snapshot;
// later variable changes cannot rewrite instructions already given to a model.
public sealed record SkillDefinition(string Name, string Description, SkillType Type,
    ImmutableArray<string> Tools, ImmutableArray<string> Depends,
    ImmutableArray<string> Loadable, string Body);
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
    private readonly object gate = new();
    private readonly Dictionary<string, SkillDefinition> definitions;
    private readonly Dictionary<string, string> variables;
    // Validate, attach, and publish the complete initial closure; roll back attachment on failure.
    public partial Skills(Agent parent, SkillSpec specification);
    public partial Agent Parent();
    // Detach the manager, not the Agent or recorded activation values. Idempotent.
    public partial void Destroy();
    // Parse only the documented front-matter subset; retain the remaining body verbatim.
    public static partial SkillDefinition Parse(string text, string path);
    // Affect future rendering only; never rewrite an already recorded manual.
    public partial void RegisterVariable(string name, string value);
    // Read immutable activation facts from History, including pending manuals.
    public partial ImmutableArray<SkillPart> Active();
    // Progressive discovery: offered by active manuals and not already active.
    public partial ImmutableArray<string> LoadableSkills();
    // Reachable declarations alone never grant dispatch.
    // Reachable declarations for optional deferral; this does not enable dispatch.
    public partial ImmutableArray<string> CatalogTools();
    // Admission permission, independent of whether a deferred schema is advertised.
    public partial bool Enabled(string tool);
    // Earliest recorded grant, including grants whose manuals are pending removal.
    public partial ToolSource? Source(string tool);
    // A pending marker is diagnostic, not revocation in this chapter.
    public partial bool Pending(string skill);
    // Called synchronously by actor admission, never by an execution worker.
    // Called by actor admission after argument validation, never by a Job worker.
    internal partial ToolResultPart Control(ToolCallPart call);
    private partial ImmutableArray<SkillPart> Activate(string name, bool initial);
    private partial List<SkillDefinition> Resolve(string name, bool initial);
    private partial ImmutableArray<string> Available(IEnumerable<string> activeNames);
    private partial string Render(string body, IEnumerable<string> tools, IEnumerable<string> available);
}
