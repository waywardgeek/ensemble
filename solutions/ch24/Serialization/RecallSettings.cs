using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

public static partial class JsonCodec
{
    public static partial JsonObject WriteRecallSettings(RecallSettings settings)
    {
        ArgumentNullException.ThrowIfNull(settings);
        return new JsonObject
        {
            ["disabled"] = settings.Disabled,
            ["disable_judge"] = settings.DisableJudge,
            ["judge_model"] = settings.JudgeModel,
            ["min_score"] = settings.MinScore,
            ["max_snippets"] = settings.MaxSnippets,
            ["min_query_length"] = settings.MinQueryLength,
            ["candidates"] = settings.Candidates,
            ["context_messages"] = settings.ContextMessages,
            ["max_recall_bytes"] = settings.MaxRecallBytes
        };
    }

    public static partial RecallSettings ReadRecallSettings(JsonElement value)
    {
        // The same exact-field guard used by memory rejects duplicates, unknowns
        // and missing fields. Constructor validation resolves zero defaults.
        MemoryFields(value, "disabled", "disable_judge", "judge_model", "min_score",
            "max_snippets", "min_query_length", "candidates", "context_messages", "max_recall_bytes");
        return new RecallSettings(value.GetProperty("disabled").GetBoolean(),
            value.GetProperty("disable_judge").GetBoolean(), value.GetProperty("judge_model").GetString()!,
            value.GetProperty("min_score").GetDouble(), value.GetProperty("max_snippets").GetInt32(),
            value.GetProperty("min_query_length").GetInt32(), value.GetProperty("candidates").GetInt32(),
            value.GetProperty("context_messages").GetInt32(), value.GetProperty("max_recall_bytes").GetInt32());
    }
}
