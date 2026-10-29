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
            ["high_watermark"] = settings.HighWatermark,
            ["low_watermark"] = settings.LowWatermark,
            ["disabled"] = settings.Disabled
        };
    }

    public static partial MemorySettings ReadMemorySettings(JsonElement value)
    {
        MemoryFields(value, "conversation", "soul", "memory", "archive64", "archive8", "session");
        return new MemorySettings(
            ReadMemoryBandSettings(value.GetProperty("conversation")),
            ReadMemoryBandSettings(value.GetProperty("soul")),
            ReadMemoryBandSettings(value.GetProperty("memory")),
            ReadMemoryBandSettings(value.GetProperty("archive64")),
            ReadMemoryBandSettings(value.GetProperty("archive8")),
            ReadMemoryBandSettings(value.GetProperty("session")));
    }

    private static partial MemoryBandSettings ReadMemoryBandSettings(JsonElement value)
    {
        MemoryFields(value, "budget", "high_watermark", "low_watermark", "disabled");
        return new MemoryBandSettings(value.GetProperty("budget").GetInt64(),
            value.GetProperty("high_watermark").GetInt64(),
            value.GetProperty("low_watermark").GetInt64(), value.GetProperty("disabled").GetBoolean());
    }
}
