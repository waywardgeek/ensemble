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
        ImmutableArray<McpServerSpec> servers = [];
        var renderers = ImmutableArray.CreateBuilder<SkillRenderer>();
        for (int index = 1; index < end; index++)
        {
            string line = lines[index].Trim();
            if (line.Length == 0 || line.StartsWith('#'))
            {
                continue;
            }
            if (line == "renderers:")
            {
                if (char.IsWhiteSpace(lines[index][0]) || !fields.TryAdd("renderers", ""))
                    throw Invalid("Duplicate or indented renderers field");
                var names = new HashSet<string>(StringComparer.Ordinal);
                int indent = -1;
                while (index + 1 < end && (string.IsNullOrWhiteSpace(lines[index + 1])
                    || char.IsWhiteSpace(lines[index + 1][0])))
                {
                    string raw = lines[++index];
                    string item = raw.Trim();
                    if (item.Length == 0 || item.StartsWith('#')) continue;
                    int currentIndent = raw.Length - raw.TrimStart().Length;
                    if (raw.Contains('\t') || (indent >= 0 && currentIndent != indent))
                        throw Invalid("Renderer entries require consistent space indentation");
                    indent = currentIndent;
                    int separator = item.IndexOf(':');
                    if (separator < 1) throw Invalid("Expected renderer name: entry");
                    string renderer = item[..separator].Trim();
                    string entry = item[(separator + 1)..].Trim();
                    if (!Regex.IsMatch(renderer, "^[A-Za-z0-9_-]+$") || !names.Add(renderer))
                        throw Invalid("Invalid or duplicate renderer name");
                    if (entry.Length > 256 || !entry.EndsWith(".js", StringComparison.Ordinal)
                        || !Regex.IsMatch(entry, "^[A-Za-z0-9_./-]+$")
                        || entry.Split('/').Any(segment => segment is "" or "." or ".."))
                        throw Invalid("Renderer entry must be a relative JavaScript file");
                    if (renderers.Count >= 64) throw Invalid("Too many renderer entries");
                    renderers.Add(new SkillRenderer(renderer, entry));
                }
                if (renderers.Count == 0) throw Invalid("Empty renderers mapping");
                list = null;
                continue;
            }
            if (line == "mcp_servers:")
            {
                if (char.IsWhiteSpace(lines[index][0]) || !fields.TryAdd("mcp_servers", ""))
                {
                    throw Invalid("Duplicate or indented mcp_servers field");
                }
                servers = ParseServers(lines, ref index, end, path);
                list = null;
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
            string.Join('\n', lines.Skip(end + 1)), servers, renderers.ToImmutable());
    }
}
