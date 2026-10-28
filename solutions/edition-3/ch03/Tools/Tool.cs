using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class Tool
{
    public partial Tool(Agent parent, ToolKind kind)
    {
        this.kind = kind;
        (name, description, schema) = Define(kind);
        parent.AddTool(this);
    }

    public partial Agent Parent() => parent ?? throw new InvalidOperationException("Tool is detached");
    public partial string Name() => name;

    public partial void Destroy()
    {
        if (destroyed)
        {
            return;
        }
        destroyed = true;
        parent?.RemoveTool(this);
    }

    // The declarations describe exactly the local validator's contract. No vendor
    // owns these definitions, and no renderer may mutate their stored JSON values.
    private static (string, string, JsonElement) Define(ToolKind kind)
    {
        var properties = new JsonObject();
        var required = new JsonArray();
        void Field(string name, string type, string explanation, bool mandatory = false)
        {
            properties[name] = new JsonObject
            {
                ["type"] = type,
                ["description"] = explanation
            };
            if (mandatory)
            {
                required.Add(name);
            }
        }
        string name;
        string description;
        switch (kind)
        {
            case ToolKind.ReadFile:
                name = "read_file";
                description = "Read a UTF-8 text file, optionally selecting a one-based inclusive line range.";
                Field("path", "string", "File path.", true);
                Field("start_line", "integer", "First line, default or zero means 1.");
                Field("end_line", "integer", "Last line, default or zero means end of file.");
                Field("max_bytes", "integer", "UTF-8 byte cap after line selection, default or zero means 32768; truncation is reported.");
                break;
            case ToolKind.ListDirectory:
                name = "list_directory";
                description = "List immediate directory entries sorted by name; directories end in /.";
                Field("path", "string", "Directory path.", true);
                break;
            case ToolKind.SearchFiles:
                name = "search_files";
                description = "Search text files recursively using a .NET regular expression; return grep-style lines and merged context windows.";
                Field("path", "string", "File or directory path.", true);
                Field("pattern", "string", "Regular expression.", true);
                Field("context_lines", "integer", "Nonnegative context radius, default 0.");
                break;
            case ToolKind.RunCommand:
                name = "run_command";
                description = "Run a fresh /bin/sh -c command with separate stdout, stderr and exit_code. No shell state survives.";
                Field("command", "string", "Shell command.", true);
                Field("cwd", "string", "Working directory for this call only.");
                break;
            case ToolKind.WriteFile:
                name = "write_file";
                description = "Create or write a UTF-8 file, creating parent directories. Existing files require overwrite:true or append:true.";
                Field("path", "string", "File path.", true);
                Field("content", "string", "Exact text to write.", true);
                Field("overwrite", "boolean", "Allow replacing an existing file, default false.");
                Field("append", "boolean", "Append instead of replacing, default false.");
                break;
            case ToolKind.EditFile:
                name = "edit_file";
                description = "Replace one exact unique nonempty old_text anchor. Zero or multiple sites refuse without changing bytes; read and narrow the anchor.";
                Field("path", "string", "File path.", true);
                Field("old_text", "string", "Exact nonempty anchor.", true);
                Field("new_text", "string", "Replacement text, possibly empty.", true);
                break;
            default:
                throw new ArgumentOutOfRangeException(nameof(kind));
        }
        var schema = new JsonObject
        {
            ["type"] = "object",
            ["properties"] = properties,
            ["required"] = required,
            ["additionalProperties"] = false
        };
        return (name, description, JsonSerializer.SerializeToElement(schema));
    }

    public partial JsonObject Declaration(Vendor vendor)
    {
        var parameters = JsonCodec.Canonical(schema);
        if (vendor == Vendor.Gemini)
        {
            parameters = GeminiSchema(schema);
        }
        var definition = new JsonObject
        {
            ["name"] = name,
            ["description"] = description,
            [vendor == Vendor.Anthropic ? "input_schema" : "parameters"] = parameters
        };
        if (vendor == Vendor.OpenAI)
        {
            return new JsonObject
            {
                ["type"] = "function",
                ["function"] = definition
            };
        }
        return definition;
    }

    // Apply the allow-list to schema nodes, not to property names. A user's
    // property may legitimately be named "additionalProperties" or "format".
    private static JsonObject GeminiSchema(JsonElement schema)
    {
        var projected = new JsonObject();
        foreach (var property in schema.EnumerateObject())
        {
            switch (property.Name)
            {
                case "type":
                case "description":
                case "enum":
                case "required":
                    projected[property.Name] = JsonCodec.Canonical(property.Value);
                    break;
                case "items":
                    projected["items"] = GeminiSchema(property.Value);
                    break;
                case "properties":
                    var properties = new JsonObject();
                    foreach (var child in property.Value.EnumerateObject())
                    {
                        properties[child.Name] = GeminiSchema(child.Value);
                    }
                    projected["properties"] = properties;
                    break;
            }
        }
        return projected;
    }

    private void Validate(JsonElement args)
    {
        if (args.ValueKind != JsonValueKind.Object)
        {
            throw new ArgumentException("Arguments must be a JSON object.");
        }
        var properties = schema.GetProperty("properties");
        var names = new HashSet<string>(StringComparer.Ordinal);
        foreach (var arg in args.EnumerateObject())
        {
            if (!names.Add(arg.Name) || !properties.TryGetProperty(arg.Name, out var field))
            {
                throw new ArgumentException("Unknown or duplicate argument: " + arg.Name);
            }
            var valid = field.GetProperty("type").GetString() switch
            {
                "string" => arg.Value.ValueKind == JsonValueKind.String,
                "boolean" => arg.Value.ValueKind is JsonValueKind.True or JsonValueKind.False,
                "integer" => arg.Value.ValueKind == JsonValueKind.Number && arg.Value.TryGetInt32(out _),
                _ => false
            };
            if (!valid)
            {
                throw new ArgumentException("Wrong argument type: " + arg.Name);
            }
        }
        foreach (var required in schema.GetProperty("required").EnumerateArray())
        {
            if (!args.TryGetProperty(required.GetString()!, out _))
            {
                throw new ArgumentException("Missing argument: " + required.GetString());
            }
        }
    }

    private static string String(JsonElement args, string name, string fallback = "")
    {
        return args.TryGetProperty(name, out var value) ? value.GetString()! : fallback;
    }

    private static int Integer(JsonElement args, string name, int fallback)
    {
        return args.TryGetProperty(name, out var value) ? value.GetInt32() : fallback;
    }

    private static bool Boolean(JsonElement args, string name)
    {
        return args.TryGetProperty(name, out var value) && value.GetBoolean();
    }

    public partial async Task<ToolResultPart> ExecuteAsync(ToolCallPart call, CancellationToken cancellation)
    {
        _ = Parent();
        try
        {
            cancellation.ThrowIfCancellationRequested();
            Validate(call.Args);
            var text = kind switch
            {
                ToolKind.ReadFile => ReadFile(call.Args),
                ToolKind.ListDirectory => ListDirectory(call.Args),
                ToolKind.SearchFiles => SearchFiles(call.Args),
                ToolKind.RunCommand => await RunCommand(call.Args, cancellation),
                ToolKind.WriteFile => WriteFile(call.Args),
                ToolKind.EditFile => EditFile(call.Args),
                _ => throw new InvalidOperationException("Unknown tool kind")
            };
            return new ToolResultPart(call.CallID, [new TextPart(text)]);
        }
        catch (Exception error) when (error is ArgumentException or IOException or UnauthorizedAccessException or System.ComponentModel.Win32Exception)
        {
            // These failures are useful model feedback, unlike a process crash.
            // Provider credentials are never included in this diagnostic path.
            return new ToolResultPart(call.CallID, [new TextPart(error.Message)], isError: true);
        }
    }
}
