using System.Collections.Immutable;
using System.Security.Cryptography;
using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

internal sealed partial class MemoryCapture
{
    internal partial MemoryCapture(CompressionLaunch launch, string conversation,
        IEnumerable<MemoryEntry> sources, string? destinationText)
    {
        ArgumentNullException.ThrowIfNull(launch);
        ArgumentNullException.ThrowIfNull(conversation);
        var entries = sources.ToImmutableArray();
        if (entries.Any(entry => entry is null))
        {
            throw new FormatException("Null captured memory source");
        }
        if (launch.Source is null)
        {
            var selected = JsonCodec.ReadMemoryConversation(conversation).Dialogue;
            if (!entries.IsEmpty || selected.IsEmpty || selected[0].Seq != launch.From
                || selected[^1].Seq != launch.To || !selected[^1].Parts.Any(part => part is HandoffPart)
                || selected.Any(entry => entry.Parts.Any(part => part is SkillPart or ToolCallPart
                    or ToolResultPart or OpaquePart || part is TextPart { From: not null }
                    or TextPart { Opaque: not null })))
            {
                throw new FormatException("Invalid captured conversation prefix");
            }
        }
        else if (conversation.Length != 0 || entries.Any(entry => entry.Band != launch.Source)
            || !entries.Select(entry => entry.ID).SequenceEqual(launch.SourceIDs))
        {
            throw new FormatException("Captured files disagree with launch");
        }
        string fingerprint = JsonCodec.MemoryFingerprint(launch.Source, launch.Destination,
            launch.ID, conversation, entries, destinationText);
        if (fingerprint != launch.SelectionHash)
        {
            throw new FormatException("Captured input fingerprint mismatch");
        }
        Launch = launch;
        Conversation = conversation;
        Sources = entries;
        DestinationText = destinationText;
    }

    internal partial string Input()
    {
        if (Launch.Source is null)
        {
            return new ConversationSelection(JsonCodec.ReadMemoryConversation(Conversation).Dialogue).Text();
        }
        var text = new StringBuilder();
        foreach (var source in Sources)
        {
            text.Append("Memory ").Append(source.ID).Append(": ").Append(source.Text).Append('\n');
        }
        if (Launch.Destination == MemoryBand.Memory && DestinationText is not null)
        {
            text.Append("Existing curated memory: ").Append(DestinationText).Append('\n');
        }
        return text.ToString();
    }
}

internal sealed partial class MemoryStage
{
    internal partial MemoryStage(MemoryCapture capture, string summary)
    {
        ArgumentNullException.ThrowIfNull(capture);
        ArgumentException.ThrowIfNullOrWhiteSpace(summary);
        Capture = capture;
        Summary = summary; // Exact accepted bytes, not trimmed or rewritten.
    }
}

public static partial class JsonCodec
{
    internal static partial string WriteMemoryConversation(IEnumerable<Entry> entries)
    {
        return WriteContext(new Context(entries)).ToJsonString();
    }

    internal static partial Context ReadMemoryConversation(string text)
    {
        using var document = JsonDocument.Parse(text);
        MemoryUnique(document.RootElement);
        var context = ReadContext(document.RootElement, long.MaxValue);
        if (!context.Ephemeral.IsEmpty || !context.Memories.IsEmpty || context.Compression is not null
            || context.CompressionOutcome is not null || context.Phase != Phase.Idle)
        {
            throw new FormatException("Captured conversation contains unrelated projection state");
        }
        // Only the canonical, dialogue-only representation is accepted. Unknown
        // fields cannot hide information that disappears when we recompute a hash.
        if (WriteMemoryConversation(context.Dialogue) != text)
        {
            throw new FormatException("Captured conversation is not canonical");
        }
        return context;
    }

    private static partial JsonObject WriteMemoryInputs(MemoryBand? source, MemoryBand destination,
        string id, string conversation, IEnumerable<MemoryEntry> sources, string? destinationText) => new()
    {
        ["source"] = source is null ? null : Label(source.Value),
        ["destination"] = Label(destination),
        ["id"] = id,
        ["conversation"] = conversation,
        ["sources"] = new JsonArray(sources.Select(entry => (JsonNode?)WriteMemory(entry)).ToArray()),
        ["destination_text"] = destinationText
    };

    internal static partial string MemoryFingerprint(MemoryBand? source, MemoryBand destination,
        string id, string conversation, IEnumerable<MemoryEntry> sources, string? destinationText)
    {
        string encoded = WriteMemoryInputs(source, destination, id, conversation, sources, destinationText).ToJsonString();
        return Convert.ToHexStringLower(SHA256.HashData(Encoding.UTF8.GetBytes(encoded)));
    }

    internal static partial bool SameCompressionLaunch(CompressionLaunch left, CompressionLaunch right)
    {
        return WriteCompressionLaunch(left).ToJsonString() == WriteCompressionLaunch(right).ToJsonString();
    }

    internal static partial string WriteMemoryStage(MemoryStage stage) => new JsonObject
    {
        ["version"] = 1,
        ["launch"] = WriteCompressionLaunch(stage.Capture.Launch),
        ["conversation"] = stage.Capture.Conversation,
        ["sources"] = new JsonArray(stage.Capture.Sources.Select(entry => (JsonNode?)WriteMemory(entry)).ToArray()),
        ["destination_text"] = stage.Capture.DestinationText,
        ["summary"] = stage.Summary
    }.ToJsonString();

    internal static partial MemoryStage ReadMemoryStage(string text)
    {
        using var document = JsonDocument.Parse(text);
        var root = document.RootElement;
        MemoryUnique(root);
        MemoryFields(root, "version", "launch", "conversation", "sources", "destination_text", "summary");
        if (root.GetProperty("version").GetInt32() != 1)
        {
            throw new FormatException("Unsupported memory stage version");
        }
        var launch = root.GetProperty("launch");
        MemoryFields(launch, "task_id", "source", "destination", "id", "timestamp", "selection_hash",
            "from", "to", "source_ids", "trace_path");
        if (launch.GetProperty("trace_path").ValueKind != JsonValueKind.String)
        {
            throw new FormatException("Memory receipt trace must be a string");
        }
        var sources = root.GetProperty("sources").EnumerateArray().Select(value =>
        {
            MemoryFields(value, "band", "id", "timestamp", "text");
            if (value.GetProperty("text").ValueKind != JsonValueKind.String)
            {
                throw new FormatException("Captured source text must be a string");
            }
            return ReadMemory(value);
        }).ToArray();
        var destination = root.GetProperty("destination_text");
        string? destinationText = destination.ValueKind == JsonValueKind.Null ? null : destination.GetString();
        string conversation = root.GetProperty("conversation").GetString()
            ?? throw new FormatException("Captured conversation must be a string");
        string summary = root.GetProperty("summary").GetString()
            ?? throw new FormatException("Staged summary must be a string");
        var capture = new MemoryCapture(ReadCompressionLaunch(launch), conversation, sources, destinationText);
        return new MemoryStage(capture, summary);
    }

    private static partial void MemoryFields(JsonElement value, params string[] names)
    {
        if (value.ValueKind != JsonValueKind.Object
            || !value.EnumerateObject().Select(property => property.Name).Order(StringComparer.Ordinal)
                .SequenceEqual(names.Order(StringComparer.Ordinal)))
        {
            throw new FormatException("Invalid memory receipt fields");
        }
    }

    private static partial void MemoryUnique(JsonElement value)
    {
        if (value.ValueKind == JsonValueKind.Object)
        {
            var names = new HashSet<string>(StringComparer.Ordinal);
            foreach (var property in value.EnumerateObject())
            {
                if (!names.Add(property.Name))
                {
                    throw new FormatException("Duplicate memory receipt field");
                }
                MemoryUnique(property.Value);
            }
        }
        else if (value.ValueKind == JsonValueKind.Array)
        {
            foreach (var item in value.EnumerateArray())
            {
                MemoryUnique(item);
            }
        }
    }
}
