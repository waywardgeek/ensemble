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
            ["phase"] = Label(context.Phase)
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
            entries.Add(new Entry(seq, actor, parts, responseID, ids));
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
        return new Context(entries, ephemeral, phase, usage);
    }

    public static partial string WriteSave(SaveDocument document)
    {
        return new JsonObject
        {
            ["context"] = document.Context is null ? null : WriteContext(document.Context),
            ["as_of"] = document.Sequence,
            ["config"] = JsonNode.Parse(document.Configuration.GetRawText()),
            ["log"] = new JsonArray(document.Events.Select(value => JsonNode.Parse(WriteEvent(value))).ToArray())
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
            var events = ImmutableArray.CreateBuilder<Event>();
            long previous = 0;
            foreach (var raw in root.GetProperty("log").EnumerateArray())
            {
                var fact = ReadEvent(raw.GetRawText());
                if (fact.Seq <= previous)
                {
                    throw new FormatException("Saved log sequences must increase strictly");
                }
                previous = fact.Seq;
                events.Add(fact);
            }
            return new SaveDocument(context, sequence, config.Clone(), events.ToImmutable());
        }
        catch (Exception ex) when (ex is JsonException or InvalidOperationException or KeyNotFoundException
            or ArgumentException or OverflowException)
        {
            throw new FormatException("Malformed save document", ex);
        }
    }
}
