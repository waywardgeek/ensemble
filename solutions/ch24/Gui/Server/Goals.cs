using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;

namespace EnsembleGui;

public static partial class Gui
{
    private static partial GoalUpdate ReadGoalUpdate(JsonElement patch)
    {
        // Decode the complete patch before sending anything to the actor. Null,
        // unknown and duplicate fields are not omissions and must not mutate it.
        if (patch.ValueKind != JsonValueKind.Object || patch.GetRawText().Length > 8192)
        {
            throw new InvalidDataException("Invalid goal update");
        }
        string? file = null;
        bool? enabled = null;
        double? seconds = null;
        var names = new HashSet<string>(StringComparer.Ordinal);
        foreach (var field in patch.EnumerateObject())
        {
            if (!names.Add(field.Name))
            {
                throw new InvalidDataException("Duplicate goal field");
            }
            switch (field.Name)
            {
                case "file" when field.Value.ValueKind == JsonValueKind.String:
                    file = field.Value.GetString();
                    break;
                case "enabled" when field.Value.ValueKind is JsonValueKind.True or JsonValueKind.False:
                    enabled = field.Value.GetBoolean();
                    break;
                case "idle_seconds" when field.Value.ValueKind == JsonValueKind.Number
                    && field.Value.TryGetDouble(out double value) && double.IsFinite(value):
                    seconds = value;
                    break;
                default:
                    throw new InvalidDataException("Invalid goal field");
            }
        }
        return new GoalUpdate(file, enabled, seconds);
    }

    private static partial JsonObject GoalValue(GoalStatus status)
    {
        // A snapshot only: the Agent owns availability, configuration and deadline.
        // No text from the goal file is exposed through the status channel.
        return new JsonObject
        {
            ["available"] = status.Available,
            ["file"] = status.File,
            ["enabled"] = status.Enabled,
            ["idle_seconds"] = status.IdleSeconds,
            ["phase"] = status.Phase,
            ["due_utc"] = status.DueUtc?.ToString("O"),
            ["last_error"] = status.LastError
        };
    }
}
