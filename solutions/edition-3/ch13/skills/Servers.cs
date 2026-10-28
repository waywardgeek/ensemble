using System.Collections.Immutable;
using System.Text.Json;
using System.Text.RegularExpressions;

namespace Ensemble;

public sealed partial class Skills
{
    // This is the documented nested list, not an arbitrary YAML object graph.
    // Consistent indentation keeps a misspelled nested field from becoming an
    // unrelated top-level skill field. JSON quoting is used only for argv.
    private static partial ImmutableArray<McpServerSpec> ParseServers(
        string[] lines, ref int index, int end, string path)
    {
        InvalidDataException Invalid(string message) => new($"{path}: {message}");
        var result = ImmutableArray.CreateBuilder<McpServerSpec>();
        var names = new HashSet<string>(StringComparer.Ordinal);
        Dictionary<string, string>? fields = null;
        int itemIndent = -1;
        int fieldIndent = -1;

        void Finish()
        {
            if (fields is null)
            {
                return;
            }
            string name = fields.GetValueOrDefault("name", "");
            if (!Regex.IsMatch(name, "^[A-Za-z0-9_-]+$") || !names.Add(name))
            {
                throw Invalid("Invalid or duplicate MCP server name");
            }
            string transport = fields.GetValueOrDefault("transport", "");
            string? command = fields.GetValueOrDefault("command");
            if (transport is not ("stdio" or "websocket"))
            {
                throw Invalid("MCP transport must be stdio or websocket");
            }
            if (transport == "stdio" && string.IsNullOrWhiteSpace(command))
            {
                throw Invalid("A stdio MCP server requires a command");
            }
            if (transport == "websocket" && (command is not null || fields.ContainsKey("args")))
            {
                throw Invalid("A websocket MCP server cannot declare command or args");
            }
            ImmutableArray<string> arguments = [];
            if (fields.TryGetValue("args", out string? encoded))
            {
                try
                {
                    using var document = JsonDocument.Parse(encoded);
                    if (document.RootElement.ValueKind != JsonValueKind.Array
                        || document.RootElement.EnumerateArray().Any(value => value.ValueKind != JsonValueKind.String))
                    {
                        throw Invalid("MCP args must be a JSON string array");
                    }
                    arguments = document.RootElement.EnumerateArray()
                        .Select(value => value.GetString()!).ToImmutableArray();
                }
                catch (JsonException)
                {
                    throw Invalid("MCP args must be a JSON string array");
                }
            }
            if (command?.Contains('\0') == true || arguments.Any(value => value.Contains('\0')))
            {
                throw Invalid("MCP command and arguments cannot contain NUL");
            }
            result.Add(new McpServerSpec(name, transport, command, arguments));
        }

        int cursor = index + 1;
        for (; cursor < end; cursor++)
        {
            string raw = lines[cursor];
            string line = raw.Trim();
            if (line.Length == 0 || line.StartsWith('#'))
            {
                continue;
            }
            int indent = raw.Length - raw.TrimStart(' ').Length;
            if (indent == 0 && !raw.StartsWith('\t'))
            {
                break;
            }
            if (raw.Contains('\t'))
            {
                throw Invalid("MCP server indentation requires spaces");
            }
            bool newItem = line.StartsWith("- ", StringComparison.Ordinal);
            if (newItem)
            {
                if (itemIndent < 0)
                {
                    itemIndent = indent;
                }
                if (indent != itemIndent)
                {
                    throw Invalid("Inconsistent MCP server list indentation");
                }
                Finish();
                fields = new Dictionary<string, string>(StringComparer.Ordinal);
                fieldIndent = -1;
                line = line[2..].Trim();
            }
            else
            {
                if (fields is null || indent <= itemIndent)
                {
                    throw Invalid("Expected an MCP server list item");
                }
                if (fieldIndent < 0)
                {
                    fieldIndent = indent;
                }
                if (indent != fieldIndent)
                {
                    throw Invalid("Inconsistent MCP server field indentation");
                }
            }
            int colon = line.IndexOf(':');
            if (colon <= 0)
            {
                throw Invalid("Expected MCP server field: value");
            }
            string key = line[..colon].Trim();
            string value = line[(colon + 1)..].Trim();
            if (key is not ("name" or "transport" or "command" or "args"))
            {
                throw Invalid("Unknown MCP server field " + key);
            }
            if (!fields!.TryAdd(key, value))
            {
                throw Invalid("Duplicate MCP server field " + key);
            }
        }
        Finish();
        if (result.Count == 0)
        {
            throw Invalid("mcp_servers must contain at least one server");
        }
        // The outer parser increments once after returning. Leave it on the
        // final consumed line, not on the next top-level field.
        index = cursor - 1;
        return result.ToImmutable();
    }
}
