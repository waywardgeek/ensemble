using System.Collections.Immutable;
using System.Text.Json;
using System.Text.Json.Nodes;
namespace Ensemble;

/// <summary>Value-like file policy, not a service or owner. Negative retention
/// keeps all diagnostics, zero keeps none, and positive values keep the newest N.
/// StartPersistence adds a sibling .journal.jsonl for per-event crash recovery.
/// ArchivePath preserves a complete canonical event log for historical replay.
/// Missing earlier events cannot be synthesized from a reduced snapshot.
/// A null policy on a library Agent opts out of automatic lifecycle persistence;
/// explicit History.Save still writes a snapshot without activating a journal.</summary>
public sealed record PersistenceOptions(string Path, int RetainLog = -1, string? ArchivePath = null);

/// <summary>Immutable decoded save candidate. Context is already folded through
/// Sequence; Events may include older diagnostic facts as well as a newer tail.</summary>
public sealed record SaveDocument(Context? Context, long Sequence, JsonElement Configuration,
    ImmutableArray<Event> Events);

public sealed partial class Agent
{
    private PersistenceOptions? persistence;
    private bool destroying;
    // Configure startup once, before actor admission. Saving is armed only after
    // successful load and current-runtime reconfiguration. Shutdown saves only
    // after startup initialization succeeds. Failed startup preserves the previous
    // snapshot; admitted startup facts can still be present in the recovery journal.
    public partial bool StartPersistence(PersistenceOptions options);
}

public sealed partial class History
{
    // Save, append, journal setup and loading share writer serialization. Readers
    // remain free to inspect the previous committed state during disk I/O.
    // All path resolution and file operations reach Sandbox through this History's
    // Agent parent. Relative paths are rooted there; no second root is retained.
    public partial void Save(PersistenceOptions options);
    internal partial PersistenceOptions NormalizePersistence(PersistenceOptions options);
    // Validate primary and backup independently. Publish only a complete replay.
    // Replay the sibling journal exactly once beyond the chosen snapshot. Returns
    // false only when no snapshot or complete journal record exists. No writer may
    // be active, and candidate validation never rewrites the files.
    public partial bool LoadSave(string path);
    // Replace saved runtime capability state with the current startup snapshot.
    internal partial void ConfigureRuntime(ImmutableArray<SkillPart> skills);
    // Activate crash recovery after successful load, before runtime reconfiguration.
    // This is owned by the existing lifecycle, not a second persistence service.
    internal partial void StartJournal(string path);
    private partial ImmutableArray<Event> ReadJournal(string path);
    private partial (Context Context, long Sequence) ReplayJournal(
        Context context, long sequence, ImmutableArray<Event> journal);
    private partial void TrimJournal(long backupSequence);
    private static partial string JournalText(IEnumerable<Event> facts);
    private partial SaveDocument ReadSaveFile(string path);
    private partial (Context Context, long Sequence) ReplaySave(SaveDocument document);
    private partial void ReplaceFile(string path, string text);
    // Stage a complete archive before trimming. Sources must agree on overlaps;
    // a missing prefix is an error, not permission to invent historical events.
    private partial string ArchiveText(string path, HistorySnapshot snapshot);
}

public sealed partial class Engine
{
    // Effective neutral configuration, deliberately excluding keys and endpoints.
    public partial JsonElement SavedConfiguration(Context context);
}

public static partial class JsonCodec
{
    public static partial string WriteSave(SaveDocument document);
    public static partial SaveDocument ReadSave(string json);
    private static partial JsonObject WriteContext(Context context);
    private static partial Context ReadContext(JsonElement value, long sequence);
}
