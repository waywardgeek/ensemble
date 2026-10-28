using System.Collections.Immutable;
using System.Globalization;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class MemoryEntry
{
    public partial MemoryEntry(MemoryBand band, string id, DateTimeOffset timestamp, string text)
    {
        if (!Enum.IsDefined(band))
        {
            throw new ArgumentException("Unknown memory band", nameof(band));
        }
        ArgumentException.ThrowIfNullOrWhiteSpace(id);
        ArgumentNullException.ThrowIfNull(text);
        Band = band;
        ID = id;
        Timestamp = timestamp.ToUniversalTime();
        Text = text;
    }
}

public sealed partial class MemoryRetirement
{
    public partial MemoryRetirement(MemoryBand band, IEnumerable<string>? ids)
    {
        if (!Enum.IsDefined(band))
        {
            throw new ArgumentException("Unknown memory band", nameof(band));
        }
        var captured = ids?.ToImmutableArray() ?? [];
        if (captured.Any(string.IsNullOrWhiteSpace)
            || captured.Distinct(StringComparer.Ordinal).Count() != captured.Length)
        {
            throw new ArgumentException("Retirement needs distinct nonempty identities", nameof(ids));
        }
        Band = band;
        IDs = captured;
    }
}

public static partial class JsonCodec
{
    private static partial JsonObject WriteMemory(MemoryEntry value) => new()
    {
        ["band"] = Label(value.Band),
        ["id"] = value.ID,
        ["timestamp"] = value.Timestamp.ToString("O", CultureInfo.InvariantCulture),
        ["text"] = value.Text
    };

    private static partial MemoryEntry ReadMemory(JsonElement value) => new(
        EnumValue<MemoryBand>(String(value, "band")), String(value, "id"),
        value.GetProperty("timestamp").GetDateTimeOffset(), String(value, "text"));

    private static partial JsonObject WriteMemoryRetirement(MemoryRetirement value) => new()
    {
        ["band"] = Label(value.Band),
        ["ids"] = new JsonArray(value.IDs.Select(id => (JsonNode?)JsonValue.Create(id)).ToArray())
    };

    private static partial MemoryRetirement ReadMemoryRetirement(JsonElement value) => new(
        EnumValue<MemoryBand>(String(value, "band")),
        value.GetProperty("ids").EnumerateArray().Select(id => id.GetString()!));
}
