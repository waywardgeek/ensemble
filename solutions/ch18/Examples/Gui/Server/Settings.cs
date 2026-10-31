using Ensemble;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace EnsembleGui;

public static partial class Gui
{
    private static partial (Preferences Preferences, int Budget, int ContextTarget) MergeSettings(
        JsonElement patch, Preferences current, int budget, int contextTarget)
    {
        if (patch.ValueKind != JsonValueKind.Object)
        {
            throw new ArgumentException("Settings must be an object");
        }
        var next = current;
        foreach (var field in patch.EnumerateObject())
        {
            var value = field.Value;
            switch (field.Name)
            {
                case "theme" when value.ValueKind == JsonValueKind.String:
                    string? theme = value.GetString();
                    if (theme is "system" or "light" or "dark")
                    {
                        next = next with { Theme = theme };
                    }
                    break;
                case "tts_enabled" when value.ValueKind is JsonValueKind.True or JsonValueKind.False:
                    next = next with { TtsEnabled = value.GetBoolean() };
                    break;
                case "pause_while_speaking" when value.ValueKind is JsonValueKind.True or JsonValueKind.False:
                    next = next with { PauseWhileSpeaking = value.GetBoolean() };
                    break;
                case "tts_speed" when value.ValueKind == JsonValueKind.Number:
                    if (value.TryGetDouble(out double speed) && double.IsFinite(speed))
                    {
                        next = next with { TtsSpeed = Math.Clamp(speed, 0.25, 4) };
                    }
                    break;
                case "font_size" when value.ValueKind == JsonValueKind.Number:
                    if (value.TryGetInt32(out int font))
                    {
                        next = next with { FontSize = Math.Clamp(font, 12, 32) };
                    }
                    break;
                case "context_target_bytes" when value.ValueKind == JsonValueKind.Number:
                    if (value.TryGetInt32(out int bytes))
                    {
                        contextTarget = ContextTarget(bytes);
                    }
                    break;
                case "max_tool_rounds" when value.ValueKind == JsonValueKind.Number:
                    if (value.TryGetInt32(out int limit) && limit >= 0)
                    {
                        budget = limit; // Zero stays zero on wire and disk.
                    }
                    break;
            }
        }
        return (next, budget, contextTarget);
    }

    private static partial JsonObject Settings(Preferences preferences, int budget, int contextTarget) => new()
    {
        ["theme"] = preferences.Theme,
        ["tts_enabled"] = preferences.TtsEnabled,
        ["tts_speed"] = preferences.TtsSpeed,
        ["font_size"] = preferences.FontSize,
        ["pause_while_speaking"] = preferences.PauseWhileSpeaking,
        ["max_tool_rounds"] = budget,
        ["context_target_bytes"] = contextTarget
    };

    private static partial (Preferences Preferences, int Budget, int ContextTarget) LoadSettings(string path)
    {
        var defaults = new Preferences();
        try
        {
            using var json = JsonDocument.Parse(File.ReadAllText(path));
            return MergeSettings(json.RootElement, defaults, 0, Agent.DefaultContextByteBudget);
        }
        catch (Exception error) when (error is IOException or UnauthorizedAccessException or JsonException or ArgumentException)
        {
            // Missing, unreadable or malformed preferences do not prevent startup.
            // Do not rewrite the file here: it may contain recoverable user data.
            return (defaults, 0, Agent.DefaultContextByteBudget);
        }
    }

    private static partial int ContextTarget(int bytes)
    {
        if (bytes <= 0)
        {
            return Agent.DefaultContextByteBudget;
        }
        // Use a wider intermediate so rounding near Int32.MaxValue cannot
        // overflow. The largest supported target is a whole thousand too.
        long rounded = ((long)bytes + 500) / 1000 * 1000;
        if (rounded == 0)
        {
            return Agent.DefaultContextByteBudget;
        }
        return (int)Math.Min(rounded, int.MaxValue / 1000L * 1000);
    }

    private static partial void SaveSettings(string path, JsonObject settings)
    {
        string fullPath = Path.GetFullPath(path);
        string directory = Path.GetDirectoryName(fullPath)!;
        Directory.CreateDirectory(directory);
        string temporary = Path.Combine(directory, ".settings-" + Guid.NewGuid().ToString("N") + ".tmp");
        try
        {
            File.WriteAllText(temporary, settings.ToJsonString(new JsonSerializerOptions { WriteIndented = true }) + "\n");
            // Same-directory replacement avoids exposing a half-written JSON file.
            // This is not a power-loss durability or backup guarantee.
            File.Move(temporary, fullPath, overwrite: true);
        }
        finally
        {
            if (File.Exists(temporary))
            {
                File.Delete(temporary);
            }
        }
    }
}
