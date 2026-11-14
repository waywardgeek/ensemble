using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;
using System.Text.RegularExpressions;

namespace Ensemble;

internal static partial class ToolSchemaCompatibility
{
    // Never rewrite below conditional/reference evaluation: enforcing one branch
    // unconditionally would reject valid arguments from a different branch.
    private static partial bool Conditional(IEnumerable<string> keywords) => keywords.Any(
        name => name is "$ref" or "$dynamicRef" or "anyOf" or "oneOf" or "allOf"
            or "not" or "if" or "then" or "else" or "dependentSchemas" or "prefixItems");

    internal static partial void Project(JsonObject schema) => ProjectNode(schema, 0);

    private static partial void ProjectNode(JsonObject schema, int depth)
    {
        if (depth > 64) throw new ArgumentException("Tool schema compatibility depth exceeded.");
        if (Conditional(schema.Select(field => field.Key))) return;
        if (schema["type"] is JsonValue type && type.TryGetValue<string>(out var name) && name == "string"
            && schema["pattern"] is JsonValue pattern && pattern.TryGetValue<string>(out var expression)
            && schema["maxLength"] is JsonValue maximum && maximum.TryGetValue<int>(out var limit) && limit >= 0)
        {
            // Compile before sending a broader declaration. Host validation uses
            // the same bounded .NET regex convention as submission contracts.
            _ = new Regex(expression, RegexOptions.CultureInvariant, TimeSpan.FromMilliseconds(100));
            string description = schema["description"]?.GetValue<string>() ?? "";
            schema["description"] = description + $" Maximum length: {limit} Unicode characters; enforced before execution.";
            schema.Remove("maxLength");
        }
        if (schema["properties"] is JsonObject properties)
            foreach (var field in properties)
                if (field.Value is JsonObject child) ProjectNode(child, depth + 1);
        if (schema["items"] is JsonObject item) ProjectNode(item, depth + 1);
    }

    internal static partial void Validate(JsonElement schema, JsonElement value) => ValidateNode(schema, value, 0);

    private static partial void ValidateNode(JsonElement schema, JsonElement value, int depth)
    {
        if (depth > 64) throw new ArgumentException("Tool schema compatibility depth exceeded.");
        if (schema.ValueKind != JsonValueKind.Object || Conditional(schema.EnumerateObject().Select(field => field.Name))) return;
        if (schema.TryGetProperty("type", out var type) && type.ValueKind == JsonValueKind.String && type.GetString() == "string"
            && schema.TryGetProperty("pattern", out var pattern) && pattern.ValueKind == JsonValueKind.String
            && schema.TryGetProperty("maxLength", out var maximum) && maximum.ValueKind == JsonValueKind.Number && maximum.TryGetInt32(out int limit) && limit >= 0)
        {
            if (value.ValueKind != JsonValueKind.String)
                throw new ArgumentException("Tool argument must be a string at a pattern/length constrained location.");
            string text = value.GetString()!;
            int length = text.EnumerateRunes().Count();
            if (length > limit)
                throw new ArgumentException("Tool argument exceeds its declared maximum string length.");
            if (schema.TryGetProperty("minLength", out var minimum) && minimum.TryGetInt32(out int min) && length < min)
                throw new ArgumentException("Tool argument is shorter than its declared minimum string length.");
            try
            {
                if (!Regex.IsMatch(text, pattern.GetString()!, RegexOptions.CultureInvariant, TimeSpan.FromMilliseconds(100)))
                    throw new ArgumentException("Tool argument does not match its declared string pattern.");
            }
            catch (RegexMatchTimeoutException)
            {
                throw new ArgumentException("Tool argument string-pattern evaluation limit exceeded.");
            }
        }
        if (value.ValueKind == JsonValueKind.Object && schema.TryGetProperty("properties", out var properties)
            && properties.ValueKind == JsonValueKind.Object)
        {
            var names = new HashSet<string>(StringComparer.Ordinal);
            foreach (var field in value.EnumerateObject())
            {
                if (!names.Add(field.Name))
                    throw new ArgumentException("Duplicate nested tool argument.");
                if (properties.TryGetProperty(field.Name, out var rule)) ValidateNode(rule, field.Value, depth + 1);
            }
        }
        if (value.ValueKind == JsonValueKind.Array && schema.TryGetProperty("items", out var item)
            && item.ValueKind == JsonValueKind.Object)
        {
            foreach (var child in value.EnumerateArray()) ValidateNode(item, child, depth + 1);
        }
    }
}
