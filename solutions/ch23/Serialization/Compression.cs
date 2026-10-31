using System.Collections.Immutable;
using System.Globalization;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class CompressionLaunch
{
    public partial CompressionLaunch(string taskID, MemoryBand? source, MemoryBand destination,
        string id, DateTimeOffset timestamp, string selectionHash, long from, long to,
        IEnumerable<string>? sourceIDs, string tracePath)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(taskID);
        ArgumentException.ThrowIfNullOrWhiteSpace(id);
        ArgumentException.ThrowIfNullOrWhiteSpace(tracePath);
        ArgumentNullException.ThrowIfNull(selectionHash);
        if (selectionHash.Length != 64 || selectionHash.Any(character => !char.IsAsciiHexDigit(character)))
        {
            throw new ArgumentException("Selection needs a SHA-256 fingerprint", nameof(selectionHash));
        }
        var ids = sourceIDs?.ToImmutableArray() ?? [];
        if (source is null)
        {
            if (destination != MemoryBand.Session || from <= 0 || to < from || !ids.IsEmpty)
            {
                throw new ArgumentException("Invalid conversation selection boundaries");
            }
        }
        else if (source < MemoryBand.Archive64 || source > MemoryBand.Session
            || destination != source - 1 || from != 0 || to != 0 || ids.Length < 2
            || ids.Any(string.IsNullOrWhiteSpace) || ids.Distinct(StringComparer.Ordinal).Count() != ids.Length)
        {
            throw new ArgumentException("Invalid graduation selection");
        }
        TaskID = taskID;
        Source = source;
        Destination = destination;
        ID = id;
        Timestamp = timestamp.ToUniversalTime();
        SelectionHash = selectionHash.ToLowerInvariant();
        From = from;
        To = to;
        SourceIDs = ids;
        TracePath = tracePath;
    }
}

public sealed partial class PendingCompression
{
    public partial PendingCompression(CompressionLaunch launch, Event? removal, MemoryEntry? committed)
    {
        ArgumentNullException.ThrowIfNull(launch);
        if (removal is not null)
        {
            History.ValidateCompressionRemoval(launch, removal);
        }
        if (committed is not null)
        {
            if (removal is null)
            {
                throw new ArgumentException("Committed replacement requires its held removal");
            }
            History.ValidateCompressionReplacement(launch, committed);
        }
        Launch = launch;
        Removal = removal;
        Committed = committed;
    }
}

public sealed partial class CompressionOutcome
{
    public partial CompressionOutcome(string taskID, CompressionState state, string error,
        IEnumerable<MeterEntry>? usage)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(taskID);
        ArgumentNullException.ThrowIfNull(error);
        if (!Enum.IsDefined(state))
        {
            throw new ArgumentException("Unknown compression outcome", nameof(state));
        }
        var rows = usage?.ToImmutableArray() ?? [];
        if (rows.Any(row => row is null || row.Identity is null || row.Counts is null
            || row.Identity.Model is null || !Enum.IsDefined(row.Identity.Vendor)
            || !Enum.IsDefined(row.Identity.Billing))
            || rows.Select(row => row.Identity).Distinct().Count() != rows.Length)
        {
            throw new ArgumentException("Usage receipt needs distinct valid billing identities", nameof(usage));
        }
        TaskID = taskID;
        State = state;
        Error = error;
        Usage = rows;
    }
}

public static partial class JsonCodec
{
    private static partial JsonObject WriteCompressionLaunch(CompressionLaunch value) => new()
    {
        ["task_id"] = value.TaskID,
        ["source"] = value.Source is MemoryBand source ? Label(source) : null,
        ["destination"] = Label(value.Destination),
        ["id"] = value.ID,
        ["timestamp"] = value.Timestamp.ToString("O", CultureInfo.InvariantCulture),
        ["selection_hash"] = value.SelectionHash,
        ["from"] = value.From,
        ["to"] = value.To,
        ["source_ids"] = new JsonArray(value.SourceIDs.Select(id => (JsonNode?)JsonValue.Create(id)).ToArray()),
        ["trace_path"] = value.TracePath
    };

    private static partial CompressionLaunch ReadCompressionLaunch(JsonElement value) => new(
        String(value, "task_id"), value.GetProperty("source").ValueKind == JsonValueKind.Null
            ? null : EnumValue<MemoryBand>(String(value, "source")),
        EnumValue<MemoryBand>(String(value, "destination")), String(value, "id"),
        value.GetProperty("timestamp").GetDateTimeOffset(), String(value, "selection_hash"),
        value.GetProperty("from").GetInt64(), value.GetProperty("to").GetInt64(),
        value.GetProperty("source_ids").EnumerateArray().Select(id => id.GetString()!), String(value, "trace_path"));

    private static partial JsonObject WriteCompressionOutcome(CompressionOutcome value) => new()
    {
        ["task_id"] = value.TaskID,
        ["state"] = Label(value.State),
        ["error"] = value.Error,
        ["usage"] = new JsonArray(value.Usage.Select(row => (JsonNode?)WriteCompressionUsage(row)).ToArray())
    };

    private static partial CompressionOutcome ReadCompressionOutcome(JsonElement value) => new(
        String(value, "task_id"), EnumValue<CompressionState>(String(value, "state")), String(value, "error"),
        value.TryGetProperty("usage", out var usage)
            ? usage.EnumerateArray().Select(ReadCompressionUsage) : null);

    private static partial JsonObject WriteCompressionUsage(MeterEntry value) => new()
    {
        ["vendor"] = Label(value.Identity.Vendor),
        ["model"] = value.Identity.Model,
        ["service_tier"] = value.Identity.ServiceTier,
        ["billing"] = Label(value.Identity.Billing),
        ["counts"] = WriteUsage(value.Counts)
    };

    private static partial MeterEntry ReadCompressionUsage(JsonElement value)
    {
        var tier = value.GetProperty("service_tier");
        var counts = value.GetProperty("counts");
        // Unlike old optional response fields, a supplied receipt has all four
        // buckets. Null means unknown; an omitted bucket is damaged data.
        foreach (string key in new[] { "input", "cache_write", "cache_read", "output" })
        {
            _ = counts.GetProperty(key);
        }
        return new MeterEntry(new UsageIdentity(
            EnumValue<Vendor>(String(value, "vendor")), value.GetProperty("model").GetString()!,
            tier.ValueKind == JsonValueKind.Null ? null : tier.GetString(),
            EnumValue<BillingMode>(String(value, "billing"))), ReadUsage(counts));
    }

    private static partial JsonObject WritePendingCompression(PendingCompression value) => new()
    {
        ["launch"] = WriteCompressionLaunch(value.Launch),
        ["removal"] = value.Removal is null ? null : JsonNode.Parse(WriteEvent(value.Removal)),
        ["committed"] = value.Committed is null ? null : WriteMemory(value.Committed)
    };

    private static partial PendingCompression ReadPendingCompression(JsonElement value) => new(
        ReadCompressionLaunch(value.GetProperty("launch")),
        value.GetProperty("removal").ValueKind == JsonValueKind.Null ? null : ReadEvent(value.GetProperty("removal").GetRawText()),
        value.GetProperty("committed").ValueKind == JsonValueKind.Null ? null : ReadMemory(value.GetProperty("committed")));
}
