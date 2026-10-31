using System.Collections.Immutable;
using System.Text.RegularExpressions;

namespace Ensemble;

public sealed partial class SkillPart
{
    public partial SkillPart(string name, string body, IEnumerable<string> tools, bool primary, bool pending)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(name);
        Name = name;
        Body = body;
        Tools = tools.ToImmutableArray();
        Primary = primary;
        Pending = pending;
    }
}

public sealed partial class Skills
{
    // Intentionally the chapter's small YAML subset, not a general YAML parser.
    // Only frontmatter boundaries are consumed; later --- remains body text.
    public static partial SkillDefinition Parse(string text, string path)
    {
        var lines = text.Replace("\r\n", "\n", StringComparison.Ordinal).Split('\n');
        InvalidDataException Invalid(string message) => new($"{path}: {message}");
        if (lines.Length < 3 || lines[0] != "---")
        {
            throw Invalid("Expected opening frontmatter boundary");
        }
        int end = Array.FindIndex(lines, 1, line => line == "---");
        if (end < 0)
        {
            throw Invalid("Missing closing frontmatter boundary");
        }
        var fields = new Dictionary<string, string>(StringComparer.Ordinal);
        var lists = new Dictionary<string, List<string>>(StringComparer.Ordinal);
        string? list = null;
        for (int index = 1; index < end; index++)
        {
            string line = lines[index].Trim();
            if (line.Length == 0 || line.StartsWith('#'))
            {
                continue;
            }
            if (line.StartsWith("- ", StringComparison.Ordinal))
            {
                if (list is null)
                {
                    throw Invalid("List item without a list field");
                }
                string item = line[2..].Trim();
                if (!Regex.IsMatch(item, "^[A-Za-z0-9_-]+$"))
                {
                    throw Invalid("Invalid list item");
                }
                lists[list].Add(item);
                continue;
            }
            int colon = line.IndexOf(':');
            if (colon <= 0)
            {
                throw Invalid("Expected field: value");
            }
            string key = line[..colon].Trim();
            string value = line[(colon + 1)..].Trim();
            if (!fields.TryAdd(key, value))
            {
                throw Invalid("Duplicate field " + key);
            }
            list = null;
            if (key is "tools" or "depends" or "loadable-skills")
            {
                var items = value.Split((char[]?)null, StringSplitOptions.RemoveEmptyEntries).ToList();
                if (items.Any(item => !Regex.IsMatch(item, "^[A-Za-z0-9_-]+$")))
                {
                    throw Invalid("Expected space-separated names or a YAML list");
                }
                lists[key] = items;
                if (value.Length == 0)
                {
                    list = key;
                }
            }
            else if (key is not ("name" or "description" or "type"))
            {
                throw Invalid("Unknown field " + key);
            }
        }
        string name = fields.GetValueOrDefault("name", "");
        if (!Regex.IsMatch(name, "^[A-Za-z0-9_-]+$"))
        {
            throw Invalid("Invalid skill name");
        }
        SkillType type = fields.GetValueOrDefault("type") switch
        {
            "primary" => SkillType.Primary,
            "loadable" => SkillType.Loadable,
            "dependency" => SkillType.Dependency,
            _ => throw Invalid("Expected primary, loadable or dependency type")
        };
        ImmutableArray<string> Names(string key) => lists.TryGetValue(key, out var values)
            ? values.Distinct(StringComparer.Ordinal).ToImmutableArray() : [];
        return new SkillDefinition(name, fields.GetValueOrDefault("description", ""), type,
            Names("tools"), Names("depends"), Names("loadable-skills"),
            string.Join('\n', lines.Skip(end + 1)));
    }
}
