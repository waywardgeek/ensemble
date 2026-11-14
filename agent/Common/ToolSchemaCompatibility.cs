using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

// A bounded compatibility subset, not a general JSON Schema validator. Only
// unconditional properties/items paths and explicit pattern+maxLength strings
// are adapted. Tool keeps the sole authoritative schema; projection is a copy.
internal static partial class ToolSchemaCompatibility
{
    internal static partial void Project(JsonObject schema);
    internal static partial void Validate(JsonElement schema, JsonElement value);
    private static partial bool Conditional(IEnumerable<string> keywords);
    private static partial void ProjectNode(JsonObject schema, int depth);
    private static partial void ValidateNode(JsonElement schema, JsonElement value, int depth);
}
