using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;

namespace EnsembleGui;

public static partial class Gui
{
    private static partial RecallSettings MergeRecallSettings(JsonElement patch, RecallSettings current)
    {
        if (patch.ValueKind != JsonValueKind.Object)
        {
            throw new InvalidDataException("Expected recall settings object");
        }
        var merged = JsonCodec.WriteRecallSettings(current);
        var fields = new HashSet<string>(StringComparer.Ordinal);
        foreach (var field in patch.EnumerateObject())
        {
            if (!fields.Add(field.Name) || !merged.ContainsKey(field.Name))
            {
                throw new InvalidDataException("Invalid or duplicate recall setting");
            }
            merged[field.Name] = JsonNode.Parse(field.Value.GetRawText());
        }
        using var document = JsonDocument.Parse(merged.ToJsonString());
        return JsonCodec.ReadRecallSettings(document.RootElement);
    }
}
