using System.Collections.Immutable;
using System.Text.Json;
using System.Text.Json.Nodes;
namespace Ensemble;

public static partial class JsonCodec
{
    private static partial JsonObject WriteContext(Context context)
    {
        var dialogue = new JsonArray();
        foreach (var entry in context.Dialogue)
        {
            dialogue.Add(new JsonObject
            {
                ["seq"] = entry.Seq,
                ["time"] = entry.Time,
                ["observation_ticks"] = entry.ObservationDuration?.Ticks,
                ["actor"] = entry.Actor.ToString().ToLowerInvariant(),
                ["parts"] = new JsonArray(entry.Parts.Select(WritePart).ToArray()),
                ["response_id"] = entry.ResponseID,
                ["part_ids"] = new JsonArray(entry.PartIDs.Select(id => (JsonNode?)System.Text.Json.Nodes.JsonValue.Create(id)).ToArray())
            });
        }
        return new JsonObject
        {
            ["dialogue"] = dialogue,
            ["ephemeral"] = new JsonArray(context.Ephemeral.Select(WritePart).ToArray()),
            ["usage"] = WriteUsage(context.Usage),
            ["phase"] = Label(context.Phase),
            ["memories"] = new JsonArray(context.Memories.Select(item => (JsonNode?)WriteMemory(item)).ToArray()),
            ["compression"] = context.Compression is null ? null : WritePendingCompression(context.Compression),
            ["compression_outcome"] = context.CompressionOutcome is null ? null : WriteCompressionOutcome(context.CompressionOutcome)
        };
    }

    private static partial Context ReadContext(JsonElement value, long sequence)
    {
        var entries = ImmutableArray.CreateBuilder<Entry>();
        long previous = 0;
        foreach (var raw in value.GetProperty("dialogue").EnumerateArray())
        {
            long seq = raw.GetProperty("seq").GetInt64();
            if (seq <= previous || seq > sequence)
            {
                throw new FormatException("Context entry sequence is outside its snapshot boundary");
            }
            previous = seq;
            var actor = EnumValue<Actor>(String(raw, "actor"));
            var parts = raw.GetProperty("parts").EnumerateArray().Select(ReadPart).ToImmutableArray();
            long responseID = raw.TryGetProperty("response_id", out var response) ? response.GetInt64() : 0;
            var ids = raw.TryGetProperty("part_ids", out var identities)
                ? identities.EnumerateArray().Select(id => id.GetInt32()).ToImmutableArray() : [];
            if (responseID < 0 || responseID > seq || ids.Any(id => id < 0)
                || ids.Distinct().Count() != ids.Length || (!ids.IsEmpty && ids.Length != parts.Length))
            {
                throw new FormatException("Invalid saved display identities");
            }
            DateTimeOffset? time = raw.TryGetProperty("time", out var recorded) && recorded.ValueKind != JsonValueKind.Null
                ? recorded.GetDateTimeOffset() : null;
            TimeSpan? duration = raw.TryGetProperty("observation_ticks", out var ticks) && ticks.ValueKind != JsonValueKind.Null
                ? TimeSpan.FromTicks(ticks.GetInt64()) : null;
            entries.Add(new Entry(seq, actor, parts, responseID, ids, time, duration));
        }
        var ephemeral = value.GetProperty("ephemeral").EnumerateArray().Select(ReadPart).ToImmutableArray();
        var phase = EnumValue<Phase>(String(value, "phase"));
        var usageValue = value.GetProperty("usage");
        foreach (string name in new[] { "input", "cache_write", "cache_read", "output" })
        {
            if (usageValue.TryGetProperty(name, out var count) && count.ValueKind != JsonValueKind.Null
                && (!count.TryGetInt64(out long number) || number < 0))
            {
                throw new FormatException("Invalid saved usage");
            }
        }
        var usage = ReadUsage(usageValue);
        // Older snapshots predate memory bands and remain readable.
        var memories = value.TryGetProperty("memories", out var corpus)
            ? corpus.EnumerateArray().Select(ReadMemory).ToImmutableArray() : [];
        var compression = value.TryGetProperty("compression", out var pending) && pending.ValueKind != JsonValueKind.Null
            ? ReadPendingCompression(pending) : null;
        var outcome = value.TryGetProperty("compression_outcome", out var terminal) && terminal.ValueKind != JsonValueKind.Null
            ? ReadCompressionOutcome(terminal) : null;
        if (compression is not null && (compression.Launch.To > sequence || compression.Removal?.Seq > sequence))
        {
            throw new FormatException("Compression selection exceeds snapshot boundary");
        }
        return new Context(entries, ephemeral, phase, usage, memories, compression, outcome);
    }

    public static partial string WriteSave(SaveDocument document)
    {
        // New snapshots contain state, never another copy of the event journal.
        // Legacy tails must be reduced before writing, not silently discarded.
        if (document.Context is null)
        {
            throw new ArgumentException("Snapshot writing requires folded Context");
        }
        if (document.Sequence < 0 || document.LatestHandoff < 0
            || document.LatestHandoff > document.Sequence || document.PreviousHandoff < 0
            || (document.PreviousHandoff != 0 && document.PreviousHandoff >= document.LatestHandoff))
        {
            throw new ArgumentException("Invalid snapshot handoff boundaries");
        }
        return new JsonObject
        {
            ["latest_handoff"] = document.LatestHandoff,
            ["previous_handoff"] = document.PreviousHandoff,
            ["context"] = WriteContext(document.Context),
            ["as_of"] = document.Sequence,
            ["config"] = JsonNode.Parse(document.Configuration.GetRawText())
        }.ToJsonString(new JsonSerializerOptions { WriteIndented = true }) + "\n";
    }

    public static partial SaveDocument ReadSave(string json)
    {
        try
        {
            using var parsed = JsonDocument.Parse(json);
            var root = parsed.RootElement;
            if (root.ValueKind != JsonValueKind.Object)
            {
                throw new FormatException("Save must be an object");
            }
            long sequence = root.GetProperty("as_of").GetInt64();
            if (sequence < 0)
            {
                throw new FormatException("Negative snapshot sequence");
            }
            long latest = root.TryGetProperty("latest_handoff", out var latestValue) ? latestValue.GetInt64() : 0;
            long preceding = root.TryGetProperty("previous_handoff", out var precedingValue) ? precedingValue.GetInt64() : 0;
            if (latest < 0 || latest > sequence || preceding < 0
                || (preceding != 0 && preceding >= latest))
            {
                throw new FormatException("Invalid snapshot handoff boundaries");
            }
            var rawContext = root.GetProperty("context");
            Context? context = rawContext.ValueKind == JsonValueKind.Null ? null : ReadContext(rawContext, sequence);
            var config = root.GetProperty("config");
            _ = String(config, "model");
            _ = String(config, "vendor");
            _ = String(config, "system_prompt");
            var toolNames = new HashSet<string>(StringComparer.Ordinal);
            foreach (var tool in config.GetProperty("tools").EnumerateArray())
            {
                string name = tool.GetProperty("name").GetString() ?? "";
                if (name.Length == 0 || !toolNames.Add(name)
                    || tool.GetProperty("description").ValueKind != JsonValueKind.String
                    || tool.GetProperty("schema").ValueKind != JsonValueKind.Object)
                {
                    throw new FormatException("Invalid saved tool declaration");
                }
            }
            // Old snapshots may carry diagnostics before the boundary and a
            // recovery tail after it. Validate their order, but never retain a
            // second event array. Only the tail contributes to restored state.
            long boundary = context is null ? 0 : sequence;
            long folded = boundary;
            long previous = 0;
            Context restored = context ?? new Context();
            if (root.TryGetProperty("log", out var legacyLog))
            {
                foreach (var raw in legacyLog.EnumerateArray())
                {
                    var fact = ReadEvent(raw.GetRawText());
                    if (fact.Seq <= previous)
                    {
                        throw new FormatException("Saved log sequences must increase strictly");
                    }
                    previous = fact.Seq;
                    // A legacy log can establish missing markers, including
                    // handoffs already folded into Context. Never duplicate one
                    // already recorded by modern snapshot metadata.
                    if (fact.Kind == EventKind.HandoffRecorded && fact.Seq > latest)
                    {
                        preceding = latest;
                        latest = fact.Seq;
                    }
                    if (fact.Seq <= boundary)
                    {
                        continue;
                    }
                    if (fact.Seq != checked(folded + 1))
                    {
                        throw new FormatException("Legacy save tail is incomplete");
                    }
                    restored = History.Reduce(restored, fact);
                    folded = fact.Seq;
                }
            }
            if (context is null && (folded == 0 || folded < sequence))
            {
                throw new FormatException("Save has no complete recovery state");
            }
            return new SaveDocument(restored, folded, config.Clone(), latest, preceding);
        }
        catch (Exception ex) when (ex is JsonException or InvalidOperationException or KeyNotFoundException
            or ArgumentException or OverflowException)
        {
            throw new FormatException("Malformed save document", ex);
        }
    }
}
