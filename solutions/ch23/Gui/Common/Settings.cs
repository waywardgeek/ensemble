using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;

namespace EnsembleGui;

// Value-like presentation preferences, not an owner or an Agent service. The
// tool budget deliberately does not live here: the Agent owns that live fact.
public sealed record Preferences(string Theme = "system", bool TtsEnabled = false,
    double TtsSpeed = 1, int FontSize = 18, bool PauseWhileSpeaking = true);

public static partial class Gui
{
    // Shared validation for disk and wire patches. Invalid/unknown fields leave
    // their previous values intact; a well-typed out-of-range number is clamped.
    private static partial (Preferences Preferences, int Budget, int ContextTarget) MergeSettings(
        JsonElement patch, Preferences current, int budget, int contextTarget);
    private static partial JsonObject Settings(Preferences preferences, int budget, int contextTarget);
    private static partial (Preferences Preferences, int Budget, int ContextTarget) LoadSettings(Agent agent, string path);
    // GUI zero means reset to the default, unlike library zero (disabled).
    // Disk and wire share this normalization; Agent owns the effective value.
    private static partial int ContextTarget(int bytes);
    private static partial void SaveSettings(Agent agent, string path, JsonObject settings);
    // Runtime memory/recall comes from the selected Agent, never Preferences or the
    // shared appearance file. Patches are strict and preserve untouched fields.
    private static partial JsonObject AgentSettings(Preferences preferences, Agent agent);
    private static partial MemorySettings MergeMemorySettings(JsonElement patch, MemorySettings current);
    private static partial RecallSettings MergeRecallSettings(JsonElement patch, RecallSettings current);
}
