using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;

namespace EnsembleGui;

public static partial class Gui
{
    private static partial JsonObject AgentSettings(Preferences preferences, Agent agent)
    {
        var result = Settings(preferences, agent.MaxToolRounds(), agent.ContextByteBudget());
        result["memory"] = JsonCodec.WriteMemorySettings(agent.MemorySettings());
        result["recall"] = JsonCodec.WriteRecallSettings(agent.RecallSettings());
        return result;
    }

    private static partial MemorySettings MergeMemorySettings(JsonElement patch, MemorySettings current)
    {
        if (patch.ValueKind != JsonValueKind.Object)
        {
            throw new InvalidDataException("Expected memory settings object");
        }
        var merged = JsonCodec.WriteMemorySettings(current);
        var bands = new HashSet<string>(StringComparer.Ordinal);
        foreach (var band in patch.EnumerateObject())
        {
            if (!bands.Add(band.Name) || merged[band.Name] is not JsonObject policy
                || band.Value.ValueKind != JsonValueKind.Object)
            {
                throw new InvalidDataException("Invalid or duplicate memory band");
            }
            var fields = new HashSet<string>(StringComparer.Ordinal);
            foreach (var field in band.Value.EnumerateObject())
            {
                if (!fields.Add(field.Name) || !policy.ContainsKey(field.Name))
                {
                    throw new InvalidDataException("Invalid or duplicate memory setting");
                }
                policy[field.Name] = JsonNode.Parse(field.Value.GetRawText());
            }
        }
        // Shared codec validates exact integer, boolean and watermark contracts.
        // There is no forgiving fallback which could silently enable a band.
        using var document = JsonDocument.Parse(merged.ToJsonString());
        return JsonCodec.ReadMemorySettings(document.RootElement);
    }
}
