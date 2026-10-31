using System.Collections.Immutable;
using System.Text.RegularExpressions;

namespace Ensemble;

public sealed partial class Skills
{
    public partial Skills(Agent parent, SkillSpec specification)
    {
        ArgumentNullException.ThrowIfNull(parent);
        ArgumentNullException.ThrowIfNull(specification);
        primaryName = specification.Primary;
        startupSkills = specification.StartupSkills?.ToImmutableArray() ?? [];
        definitions = new Dictionary<string, SkillDefinition>(StringComparer.Ordinal);
        variables = new Dictionary<string, string>(StringComparer.Ordinal);
        if (specification.Variables is not null)
        {
            foreach (var (name, value) in specification.Variables)
            {
                RegisterVariable(name, value);
            }
        }
        // Sorting makes diagnostics and duplicate detection reproducible. A host
        // chooses one root; no implicit home-directory or environment fallback.
        foreach (string path in Directory.GetFiles(specification.Directory, "SKILL.md", SearchOption.AllDirectories)
                     .Order(StringComparer.Ordinal))
        {
            var definition = Parse(File.ReadAllText(path), path);
            if (!definitions.TryAdd(definition.Name, definition))
            {
                throw new InvalidDataException("Duplicate skill: " + definition.Name);
            }
        }
        if (!definitions.TryGetValue(specification.Primary, out var primary)
            || primary.Type != SkillType.Primary)
        {
            throw new InvalidDataException("Primary skill is missing or has the wrong type");
        }
        // The host supplies environment values explicitly. Only named MCP
        // credentials survive construction, separate from rendered variables.
        foreach (var server in definitions.Values.SelectMany(value => value.McpServers))
        {
            if (server.AuthEnvironment is { } key
                && specification.McpEnvironment?.TryGetValue(key, out var token) == true
                && !string.IsNullOrEmpty(token))
            {
                mcpCredentials[key] = token;
            }
        }
        Attach(parent);
    }

    internal partial Skills(Agent parent, SkillBootstrap bootstrap)
    {
        ArgumentNullException.ThrowIfNull(parent);
        ArgumentNullException.ThrowIfNull(bootstrap);
        primaryName = bootstrap.Primary;
        startupSkills = bootstrap.Startup;
        definitions = new Dictionary<string, SkillDefinition>(bootstrap.Definitions, StringComparer.Ordinal);
        variables = new Dictionary<string, string>(bootstrap.Variables, StringComparer.Ordinal);
        foreach (var (name, value) in bootstrap.Credentials)
        {
            mcpCredentials.Add(name, value);
        }
        Attach(parent);
    }

    internal partial SkillBootstrap CaptureBootstrap()
    {
        // Capture before the owning Agent's admission gate is acquired. Skill
        // operations can reach that Agent while holding this gate, not vice versa.
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            return new SkillBootstrap(primaryName, startupSkills,
                definitions.ToImmutableDictionary(StringComparer.Ordinal),
                variables.ToImmutableDictionary(StringComparer.Ordinal),
                mcpCredentials.ToImmutableDictionary(StringComparer.Ordinal));
        }
    }

    private partial void Attach(Agent parent)
    {
        parent.AddSkills(this);
        var controls = new List<Tool>();
        try
        {
            // These are concrete ordinary catalog entries, but dispatch treats
            // them as actor controls. No background job owns skill state.
            controls.Add(new Tool(parent, ToolKind.LoadSkill));
            controls.Add(new Tool(parent, ToolKind.UnloadSkill));
            var additions = ResolveStartup();
            if (additions.Any(skill => !skill.McpServers.IsDefaultOrEmpty))
            {
                initialVariables = variables.ToImmutableDictionary(StringComparer.Ordinal);
            }
            else
            {
                Activate(primaryName, initial: true);
            }
        }
        catch
        {
            foreach (var control in controls)
            {
                control.Destroy();
            }
            Destroy();
            throw;
        }
    }

    public partial Agent Parent() => parent ?? throw new ObjectDisposedException(nameof(Skills));

    public partial void Destroy()
    {
        if (destroyed)
        {
            return;
        }
        destroyed = true;
        try
        {
            Task.WhenAll(Clients().Select(client => client.DestroyAsync())).GetAwaiter().GetResult();
        }
        finally
        {
            lock (gate)
            {
                mcpCredentials.Clear();
            }
            parent?.RemoveSkills(this);
        }
    }

    public partial void RegisterVariable(string name, string value)
    {
        if (!Regex.IsMatch(name, "^[A-Z][A-Z0-9_]*$") || name is "TOOLS" or "SKILLS")
        {
            throw new ArgumentException("Variable name is invalid or reserved", nameof(name));
        }
        ArgumentNullException.ThrowIfNull(value);
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            variables[name] = value;
        }
    }

    public partial ImmutableArray<SkillPart> Active()
    {
        return Parent().History().View().Dialogue.SelectMany(entry => entry.Parts)
            .OfType<SkillPart>().ToImmutableArray();
    }

    private partial ImmutableArray<string> Available(IEnumerable<string> activeNames)
    {
        var active = activeNames.ToHashSet(StringComparer.Ordinal);
        return active.SelectMany(name => definitions[name].Loadable)
            .Where(name => !active.Contains(name) && Parent().Configuration().AllowsSkill(name)).Distinct(StringComparer.Ordinal)
            .Order(StringComparer.Ordinal).ToImmutableArray();
    }

    public partial ImmutableArray<string> LoadableSkills()
    {
        var retained = Active();
        return Available(retained.Where(skill => !skill.Pending).Select(skill => skill.Name))
            .Where(name => !retained.Any(skill => skill.Name == name)).ToImmutableArray();
    }

    public partial bool Enabled(string tool)
    {
        return Parent().ToolEnabled(tool);
    }

    private partial ImmutableArray<string> UniversalControls()
    {
        ImmutableArray<string> names = ["load_skill", "unload_skill", "micro_handoff"];
        return Parent().Configuration().StubsToolResults ? names.Add("keep_tool_results") : names;
    }

    public partial ImmutableArray<string> CatalogTools(Context? context)
    {
        context ??= Parent().History().View();
        var retained = context.Dialogue.SelectMany(entry => entry.Parts).OfType<SkillPart>().ToArray();
        var pending = retained.Where(skill => skill.Pending).Select(skill => skill.Name).ToHashSet(StringComparer.Ordinal);
        var visited = new HashSet<string>(StringComparer.Ordinal);
        // Initial grants include the universal controls even when no skills are offered.
        var policy = Parent().Configuration();
        var names = context.EnabledTools(policy).ToHashSet(StringComparer.Ordinal);
        void Visit(string name)
        {
            if (!policy.AllowsSkill(name) || pending.Contains(name) || !visited.Add(name) || !definitions.TryGetValue(name, out var definition))
            {
                return;
            }
            foreach (string tool in definition.Tools)
            {
                if (Parent().Tools().Any(candidate => candidate.Name() == tool))
                {
                    names.Add(tool);
                }
            }
            foreach (string related in definition.Depends.Concat(definition.Loadable))
            {
                Visit(related);
            }
        }
        foreach (var skill in retained.Where(skill => !skill.Pending))
        {
            Visit(skill.Name);
        }
        names.ExceptWith(context.RevokedTools(policy));
        return names.Where(name => policy.AllowsTool(name) && !Parent().IsAutomaticTool(name))
            .Order(StringComparer.Ordinal).ToImmutableArray();
    }

    public partial ToolSource? Source(string tool)
    {
        foreach (var entry in Parent().History().View().Dialogue)
        {
            foreach (var skill in entry.Parts.OfType<SkillPart>())
            {
                if (!skill.Pending && skill.Tools.Contains(tool, StringComparer.Ordinal))
                {
                    return new ToolSource(skill.Name, entry.Seq);
                }
            }
        }
        return null;
    }

    public partial bool Pending(string skill)
    {
        return Active().Any(value => value.Name == skill && value.Pending);
    }

    private partial List<SkillDefinition> Resolve(string name, bool initial, IEnumerable<string>? preceding)
    {
        var active = Active().Select(skill => skill.Name).ToHashSet(StringComparer.Ordinal);
        if (preceding is not null)
        {
            active.UnionWith(preceding);
        }
        var visiting = new HashSet<string>(StringComparer.Ordinal);
        var visited = new HashSet<string>(StringComparer.Ordinal);
        var ordered = new List<SkillDefinition>();
        void Visit(string current, bool root)
        {
            // Refuse before preparing dependencies or external clients, even if
            // an earlier session recorded a successful activation of this skill.
            if (!Parent().Configuration().AllowsSkill(current))
            {
                throw new InvalidDataException("Skill is disabled by launch policy: " + current);
            }
            if (active.Contains(current) || visited.Contains(current))
            {
                return;
            }
            if (!definitions.TryGetValue(current, out var definition))
            {
                throw new InvalidDataException("Missing skill: " + current);
            }
            if (definition.Type == SkillType.Primary && !(initial && root))
            {
                throw new InvalidDataException("A primary skill cannot be dynamically loaded or used as a dependency");
            }
            if (!visiting.Add(current))
            {
                throw new InvalidDataException("Dependency cycle at " + current);
            }
            foreach (string offered in definition.Loadable)
            {
                if (!definitions.TryGetValue(offered, out var target) || target.Type != SkillType.Loadable)
                {
                    throw new InvalidDataException("Invalid loadable skill: " + offered);
                }
            }
            foreach (string tool in definition.Tools)
            {
                if (definition.McpServers.IsDefaultOrEmpty
                    && !Parent().Tools().Any(candidate => candidate.Name() == tool))
                {
                    throw new InvalidDataException("Unknown tool: " + tool);
                }
            }
            foreach (string dependency in definition.Depends)
            {
                Visit(dependency, root: false);
            }
            visiting.Remove(current);
            visited.Add(current);
            ordered.Add(definition);
        }
        Visit(name, root: true);
        return ordered;
    }

    private partial string Render(string body, IEnumerable<string> tools, IEnumerable<string> available,
        IReadOnlyDictionary<string, string>? values)
    {
        // A single regex pass avoids recursively interpreting substituted text.
        // Unknown names remain visible rather than silently becoming empty.
        return Regex.Replace(body, @"\$([A-Z][A-Z0-9_]*)", match =>
        {
            string key = match.Groups[1].Value;
            if (key == "TOOLS")
            {
                return string.Join("\n", tools.Where(name => Parent().Configuration().AllowsTool(name) && !Parent().IsAutomaticTool(name))
                    .Select(name => "- " + name));
            }
            if (key == "SKILLS")
            {
                return string.Join("\n", available.Select(name => $"- {name}: {definitions[name].Description}"));
            }
            return (values ?? variables).TryGetValue(key, out string? value) ? value : match.Value;
        });
    }

    private partial ImmutableArray<SkillPart> Activate(string name, bool initial)
    {
        lock (gate)
        {
            var additions = initial ? ResolveStartup() : Resolve(name, initial: false);
            if (additions.Count == 0)
            {
                return [];
            }
            var active = Active();
            // All manuals in this batch see the same prospective capability
            // set. Dependencies do not receive a partially enabled snapshot.
            var names = active.Select(skill => skill.Name)
                .Concat(additions.Select(skill => skill.Name)).ToArray();
            var available = Available(names);
            var tools = active.SelectMany(skill => skill.Tools)
                .Concat(additions.SelectMany(skill => skill.Tools)).ToHashSet(StringComparer.Ordinal);
            tools.UnionWith(UniversalControls());
            var enabled = tools.Order(StringComparer.Ordinal).ToArray();

            var rendered = ImmutableArray.CreateBuilder<SkillPart>();
            foreach (var skill in additions)
            {
                bool primary = skill.Type == SkillType.Primary;
                IEnumerable<string> grants = skill.Tools;
                if (primary)
                {
                    // The primary owns the universal control grants. They do
                    // not need to be repeated in every dynamic manual.
                    grants = grants.Concat(UniversalControls());
                }
                string body = Render(skill.Body, enabled, available);
                rendered.Add(new SkillPart(skill.Name, body, grants, primary));
            }
            var parts = rendered.ToImmutable();
            // One durable-before-publication operation commits the whole closure.
            Parent().History().Append(new Event(EventKind.SkillLoaded, parts: parts));
            return parts;
        }
    }

    internal partial Task<ToolResultPart> ControlAsync(ToolCallPart call, CancellationToken cancellation)
    {
        string name = call.Args.GetProperty("name").GetString()!;
        string message;
        IEnumerable<string> discovered = [];
        var active = Active();
        if (call.Name == "load_skill")
        {
            if (active.Any(skill => skill.Name == name))
            {
                message = "Already loaded: " + name;
            }
            else
            {
                if (!LoadableSkills().Contains(name, StringComparer.Ordinal))
                {
                    throw new ArgumentException("Skill is not currently offered: " + name);
                }
                var closure = Resolve(name, initial: false).ToImmutableArray();
                if (closure.Any(skill => !skill.McpServers.IsDefaultOrEmpty))
                {
                    var manager = Parent().Mcp() ?? new Mcp(Parent());
                    ImmutableDictionary<string, string> values;
                    lock (gate)
                    {
                        values = variables.ToImmutableDictionary(StringComparer.Ordinal);
                    }
                    string workspace = Directory.GetCurrentDirectory();
                    // Preparation cannot publish History. Commit returns through the mailbox.
                    return Task.Run(async () =>
                    {
                        var references = await PrepareAsync(name, closure, values, manager, workspace, cancellation);
                        string loaded = "Loaded: " + string.Join(", ", closure.Select(skill => skill.Name));
                        return new ToolResultPart(call.CallID, [new TextPart(loaded)], toolReferences: references);
                    });
                }
                var added = Activate(name, initial: false);
                var previous = active.Where(skill => !skill.Pending).SelectMany(skill => skill.Tools).ToHashSet(StringComparer.Ordinal);
                discovered = added.SelectMany(skill => skill.Tools).Where(tool => !previous.Contains(tool));
                message = "Loaded: " + string.Join(", ", added.Select(skill => skill.Name));
            }
        }
        else
        {
            var skill = active.FirstOrDefault(skill => skill.Name == name)
                ?? throw new ArgumentException("Skill is not loaded: " + name);
            if (skill.Primary)
            {
                throw new ArgumentException("The primary skill cannot be unloaded");
            }
            if (!Pending(name))
            {
                Parent().History().Append(new Event(EventKind.SkillUnloadPending, name: name));
            }
            message = "Pending unload at later compaction: " + name + ". Instructions remain active; tool grants are revoked and external servers have stopped.";
        }
        var result = new ToolResultPart(call.CallID, [new TextPart(message)], toolReferences: discovered);
        return call.Name == "unload_skill"
            ? StopClientsAsync(name, result)
            : Task.FromResult(result);
    }
}
