using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

public static partial class JsonCodec
{
    public static partial JsonObject WriteMemorySettings(MemorySettings settings)
    {
        ArgumentNullException.ThrowIfNull(settings);
        return new JsonObject
        {
            ["conversation"] = WriteMemoryBandSettings(settings.Conversation),
            ["soul"] = WriteMemoryBandSettings(settings.Soul),
            ["memory"] = WriteMemoryBandSettings(settings.Memory),
            ["archive64"] = WriteMemoryBandSettings(settings.Archive64),
            ["archive8"] = WriteMemoryBandSettings(settings.Archive8),
            ["session"] = WriteMemoryBandSettings(settings.Session)
        };
    }

    private static partial JsonObject WriteMemoryBandSettings(MemoryBandSettings settings)
    {
        return new JsonObject
        {
            ["budget"] = settings.Budget,
            ["disabled"] = settings.Disabled
        };
    }

    public static partial MemorySettings ReadMemorySettings(JsonElement value)
    {
        MemoryFields(value, "conversation", "soul", "memory", "archive64", "archive8", "session");
        var defaults = new MemorySettings();
        return new MemorySettings(
            ReadMemoryBandSettings(value.GetProperty("conversation"), defaults.Conversation.Budget),
            ReadMemoryBandSettings(value.GetProperty("soul"), defaults.Soul.Budget),
            ReadMemoryBandSettings(value.GetProperty("memory"), defaults.Memory.Budget),
            ReadMemoryBandSettings(value.GetProperty("archive64"), defaults.Archive64.Budget),
            ReadMemoryBandSettings(value.GetProperty("archive8"), defaults.Archive8.Budget),
            ReadMemoryBandSettings(value.GetProperty("session"), defaults.Session.Budget));
    }

    private static partial MemoryBandSettings ReadMemoryBandSettings(JsonElement value, long defaultBudget)
    {
        if (value.ValueKind != JsonValueKind.Object)
        {
            throw new FormatException("Invalid memory band settings");
        }
        bool legacy = value.TryGetProperty("high_watermark", out _)
            || value.TryGetProperty("low_watermark", out _);
        if (legacy)
        {
            // Read old settings without rewriting disk or restoring removed knobs.
            // Still validate the complete old shape; malformed input must not turn
            // into a successful migration merely because the fields are obsolete.
            MemoryFields(value, "budget", "high_watermark", "low_watermark", "disabled");
            long high = value.GetProperty("high_watermark").GetInt64();
            long low = value.GetProperty("low_watermark").GetInt64();
            ArgumentOutOfRangeException.ThrowIfNegative(high);
            ArgumentOutOfRangeException.ThrowIfNegative(low);
            long budget = new MemoryBandSettings(value.GetProperty("budget").GetInt64())
                .Effective(defaultBudget).Budget;
            long effectiveHigh = high == 0 ? checked(2 * budget) : high;
            long effectiveLow = low == 0 ? budget : low;
            if (effectiveLow > effectiveHigh)
            {
                throw new FormatException("Invalid legacy memory thresholds");
            }
        }
        else
        {
            MemoryFields(value, "budget", "disabled");
        }
        return new MemoryBandSettings(value.GetProperty("budget").GetInt64(),
            value.GetProperty("disabled").GetBoolean());
    }
}
