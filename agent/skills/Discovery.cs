namespace Ensemble;

public sealed partial class Skills
{
    // Called at actor admission boundaries, never by historical request rendering.
    // The configured directory is host-selected executable configuration, not a
    // workspace search path. Discovery itself grants no tools or MCP connections.
    private partial void RefreshCatalog()
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (Parent().Configuration().SandboxRoot.Length != 0) return;

            var candidate = new Dictionary<string, SkillDefinition>(StringComparer.Ordinal);
            foreach (string path in Directory.GetFiles(sourceDirectory, "SKILL.md", SearchOption.AllDirectories)
                         .Order(StringComparer.Ordinal))
            {
                var definition = Parse(File.ReadAllText(path), path);
                if (!candidate.TryAdd(definition.Name, definition))
                {
                    throw new InvalidDataException("Duplicate skill: " + definition.Name);
                }
            }
            // Even pending manuals remain retained until compaction. Preserve
            // their template relationships as well as their recorded text.
            foreach (var skill in Active())
            {
                if (definitions.TryGetValue(skill.Name, out var activeDefinition))
                {
                    candidate[skill.Name] = activeDefinition;
                }
            }
            // Publish only after every file has parsed and collisions have been
            // checked. Readers never observe a dictionary being cleared/refilled.
            definitions = candidate;
        }
    }

    internal partial string DiscoveryInstruction()
    {
        RefreshCatalog();
        lock (gate)
        {
            var names = LoadableSkills();
            return "Current skills available through load_skill (this current list supersedes skill lists in retained manuals; discovery grants nothing):\n"
                + (names.Length == 0 ? "None." : string.Join("\n", names.Select(name =>
                    $"- {name}: {definitions[name].Description}")));
        }
    }
}
