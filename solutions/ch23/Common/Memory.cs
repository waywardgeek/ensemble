using System.Collections.Immutable;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

/// <summary>Stable-to-recent provider order. Conversation is not a memory band.</summary>
public enum MemoryBand { Soul = 1, Memory, Archive64, Archive8, Session }

/// <summary>Immutable value-like memory data, not a grant or an owning relationship.
/// ID is stable within its band; it is never a path or permission to open a file.
/// Timestamp orders entries oldest-first; exact text survives replay and save.</summary>
public sealed partial class MemoryEntry
{
    public readonly MemoryBand Band;
    public readonly string ID;
    public readonly DateTimeOffset Timestamp;
    public readonly string Text;
    public partial MemoryEntry(MemoryBand band, string id, DateTimeOffset timestamp, string text);
}

/// <summary>Value-like explicit retirement. Empty IDs means the entire band,
/// useful for disabling a corpus without deleting its source files.</summary>
public sealed partial class MemoryRetirement
{
    public readonly MemoryBand Band;
    public readonly ImmutableArray<string> IDs;
    public partial MemoryRetirement(MemoryBand band, IEnumerable<string>? ids = null);
}

public static partial class JsonCodec
{
    private static partial JsonObject WriteMemory(MemoryEntry value);
    private static partial MemoryEntry ReadMemory(JsonElement value);
    private static partial JsonObject WriteMemoryRetirement(MemoryRetirement value);
    private static partial MemoryRetirement ReadMemoryRetirement(JsonElement value);
}

/// <summary>Value-like, immutable selected conversation source. It owns no Agent
/// or file relationships. Exact entries (including their sequence/part identities)
/// are retained for later stale-result validation; readable text is derived only
/// when preparing compressor input. Selection alone never removes source.</summary>
internal sealed partial class ConversationSelection
{
    internal readonly ImmutableArray<Entry> Entries;
    internal partial ConversationSelection(IEnumerable<Entry> entries);
    internal partial string Text();
}

public sealed partial class History
{
    // Select the oldest whole completed prefix when readable dialogue exceeds
    // budget. Handoff notes join the source but do not consume dialogue budget.
    // Skills and the unfinished tail are never selected. No I/O or mutation.
    internal static partial ConversationSelection? SelectConversation(Context context, long budget);
}
