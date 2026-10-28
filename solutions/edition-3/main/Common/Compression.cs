using System.Collections.Immutable;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

/// <summary>Committed is final disk promotion/archival, not just population.</summary>
public enum CompressionState { Committed = 1, Failed, Canceled, Invalidated, Abandoned }

/// <summary>Value-like parent-authored selection identity. Null Source means
/// conversation; file sources graduate one band. IDs grant no filesystem access.
/// SelectionHash binds the exact selected source and destination bytes in the
/// staged disk record to this launch. Replay never reads those files.</summary>
public sealed partial class CompressionLaunch
{
    public readonly string TaskID;
    public readonly MemoryBand? Source;
    public readonly MemoryBand Destination;
    public readonly string ID;
    public readonly DateTimeOffset Timestamp;
    public readonly string SelectionHash;
    public readonly long From;
    public readonly long To;
    public readonly ImmutableArray<string> SourceIDs;
    public readonly string TracePath;
    public partial CompressionLaunch(string taskID, MemoryBand? source, MemoryBand destination,
        string id, DateTimeOffset timestamp, string selectionHash, long from = 0, long to = 0,
        IEnumerable<string>? sourceIDs = null, string tracePath = "");
}

/// <summary>Immutable replay receipt, not a live worker. Removal is held without
/// changing active source. Committed retains the replacement for crash recovery
/// until a terminal outcome confirms disk finalization.</summary>
public sealed partial class PendingCompression
{
    public readonly CompressionLaunch Launch;
    public readonly Event? Removal;
    public readonly MemoryEntry? Committed;
    public partial PendingCompression(CompressionLaunch launch, Event? removal = null, MemoryEntry? committed = null);
}

/// <summary>Value-like bounded last outcome. It cannot revive live work.</summary>
public sealed partial class CompressionOutcome
{
    public readonly string TaskID;
    public readonly CompressionState State;
    public readonly string Error;
    public partial CompressionOutcome(string taskID, CompressionState state, string error = "");
}

public sealed partial class History
{
    internal static partial void ValidateCompressionRemoval(CompressionLaunch launch, Event removal);
    internal static partial void ValidateCompressionReplacement(CompressionLaunch launch, MemoryEntry replacement);
    private static partial void ValidateCompressionEvent(Context before, Event value);
}

public static partial class JsonCodec
{
    private static partial JsonObject WriteCompressionLaunch(CompressionLaunch value);
    private static partial CompressionLaunch ReadCompressionLaunch(JsonElement value);
    private static partial JsonObject WriteCompressionOutcome(CompressionOutcome value);
    private static partial CompressionOutcome ReadCompressionOutcome(JsonElement value);
    private static partial JsonObject WritePendingCompression(PendingCompression value);
    private static partial PendingCompression ReadPendingCompression(JsonElement value);
}
