using System.Collections.Immutable;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

/// <summary>Agent-owned corpus and transaction I/O. History remains the authority
/// for acceptance; this child never appends events or starts model work. All paths
/// follow Parent → Sandbox and Parent → DataDirectory, not copied services.</summary>
public sealed partial class Memory
{
    internal Agent? parent;
    internal bool destroyed;
    private readonly object gate = new();
    // Safety limit for individual corpus files and the complete hidden receipt,
    // not a compression budget. Oversized input is refused, never truncated.
    internal const int MaximumFileBytes = 64 * 1024 * 1024;

    public partial Memory(Agent parent);
    public partial Agent Parent();
    public partial void Destroy();
    internal partial ImmutableArray<MemoryEntry> Scan(Context recovered, MemoryBand band);
    internal partial MemoryCapture? CaptureConversation(Context context, long budget, DateTimeOffset now);
    internal partial MemoryCapture? CaptureGraduation(Context context, MemoryBand band,
        long highWatermark, long lowWatermark, DateTimeOffset now);
    // Checks selected replay/disk bytes only. The actor must separately recheck
    // the launch policy and task lifetime before committing a worker result.
    internal partial bool StillCurrent(MemoryCapture capture, Context context);
    internal partial void Stage(MemoryCapture capture, string summary);
    // Call at startup, for a committed result, or after recording a terminal
    // outcome, never to inspect an uncommitted task whose worker is still live.
    // Recovery abandons that state because no worker survives process restart.
    // It leaves a committed stage until the returned outcome has been journaled;
    // calling again with that terminal Context releases the receipt.
    internal partial CompressionOutcome? Recover(Context context);

    private partial string HomePath(string relative);
    private partial string CorpusPath(MemoryBand band, string id);
    private static partial string BandDirectory(MemoryBand band);
    private static partial (DateOnly Day, int Number) ParsePoint(string value);
    private static partial ((DateOnly Day, int Number) Start, (DateOnly Day, int Number) End) Coverage(MemoryBand band, string id);
    private static partial DateTimeOffset FileTimestamp(MemoryBand band, string id);
    private partial string? ReadOptional(string path);
    private partial void RequireRecovered(Context context);
    private partial string NextSessionID(DateTimeOffset now);
    private partial MemoryCapture Capture(Context context, MemoryBand? source, string id,
        DateTimeOffset now, string conversation, ImmutableArray<MemoryEntry> sources);
    private partial bool CurrentFiles(MemoryCapture capture, bool allowRetired);
    private partial void FinalizeStage(PendingCompression pending, MemoryStage stage);
    private partial MemoryStage ReadStage();
    private partial void DropStage();
}

/// <summary>Value-like exact captured input. Conversation uses the existing
/// Context codec to retain part identities, not just readable text. File entries
/// retain exact UTF-8 text; destination null means absence, not an empty file.
/// The worker may read this immutable value but has no disk authority from it.</summary>
internal sealed partial class MemoryCapture
{
    internal readonly CompressionLaunch Launch;
    internal readonly string Conversation;
    internal readonly ImmutableArray<MemoryEntry> Sources;
    internal readonly string? DestinationText;
    internal partial MemoryCapture(CompressionLaunch launch, string conversation,
        IEnumerable<MemoryEntry> sources, string? destinationText);
    internal partial string Input();
}

/// <summary>Value-like hidden candidate. It cannot become active corpus data
/// without a matching committed replay receipt.</summary>
internal sealed partial class MemoryStage
{
    internal readonly MemoryCapture Capture;
    internal readonly string Summary;
    internal partial MemoryStage(MemoryCapture capture, string summary);
}

public sealed partial class Agent
{
    private Memory? memory;
    public partial Memory Memory();
    internal partial void AddMemory(Memory child);
    internal partial void RemoveMemory(Memory child);
}

public static partial class JsonCodec
{
    internal static partial string WriteMemoryConversation(IEnumerable<Entry> entries);
    internal static partial Context ReadMemoryConversation(string text);
    internal static partial string MemoryFingerprint(MemoryBand? source, MemoryBand destination,
        string id, string conversation, IEnumerable<MemoryEntry> sources, string? destinationText);
    internal static partial bool SameCompressionLaunch(CompressionLaunch left, CompressionLaunch right);
    internal static partial string WriteMemoryStage(MemoryStage stage);
    internal static partial MemoryStage ReadMemoryStage(string text);
    private static partial JsonObject WriteMemoryInputs(MemoryBand? source, MemoryBand destination,
        string id, string conversation, IEnumerable<MemoryEntry> sources, string? destinationText);
    private static partial void MemoryFields(JsonElement value, params string[] names);
    private static partial void MemoryUnique(JsonElement value);
}
