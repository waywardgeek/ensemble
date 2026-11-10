using System.Collections.Immutable;
using System.Text.Json;
using System.Text.Json.Nodes;
namespace Ensemble;

/// <summary>Value-like snapshot and journal policy. Audit keeps the same journal
/// untruncated on disk; it never enlarges working History in memory. Default
/// retention keeps the preceding handoff window and the validated backup suffix.
/// A null Agent policy opts out of lifecycle persistence; explicit Save alone
/// does not activate a journal.</summary>
public sealed record PersistenceOptions(string Path, bool Audit = false);

/// <summary>Value-like self-contained recovery snapshot. No event collection is
/// retained here. Legacy input is validated and folded at the decoding boundary.</summary>
public sealed record SaveDocument(
    Context Context,
    long Sequence,
    JsonElement Configuration,
    long LatestHandoff = 0,
    long PreviousHandoff = 0);

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
    private partial IEnumerable<Event> ReadJournal(string path);
    private partial HistorySnapshot ReplayJournal(
        Context context, long sequence, IEnumerable<Event> journal,
        long latestHandoff = 0, long previousHandoff = 0);
    private partial void TrimJournal(long backupSequence);
    private partial void WriteJournal(string path, IEnumerable<Event> facts);
    private partial SaveDocument ReadSaveFile(string path);
    private partial void ReplaceFile(string path, string text);
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
