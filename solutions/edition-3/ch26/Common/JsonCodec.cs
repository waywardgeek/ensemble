using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

/// <summary>Stateless JSON vocabulary conversion. No vendor request algorithms live here.
/// Exact JsonElement numbers avoid rounding tool arguments through floating point.</summary>
public static partial class JsonCodec
{
    // The public JSON-lines vocabulary is independent of C# enum/member spelling.
    // Reject unknown discriminators rather than silently losing recorded facts.
    public static partial Event ReadEvent(string line);
    public static partial string WriteEvent(Event value);
    public static partial JsonObject WritePart(Part part);
    public static partial Part ReadPart(JsonElement value);
    public static partial JsonObject WriteUsage(Usage value);
    public static partial JsonObject WriteProvenance(Provenance value);
    public static partial Provenance ReadProvenance(JsonElement value);
    // Sort object keys recursively for stable output; preserve arrays and exact
    // numeric tokens. Returned nodes do not borrow a JsonDocument lifetime.
    public static partial JsonNode? Canonical(JsonElement value);
    // Readable text and safe redaction placeholders only; opaque values and blob
    // locators never become prose. Tool-result content is visited recursively.
    public static partial string Text(IEnumerable<Part> parts);
}
