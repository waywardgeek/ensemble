using System.Collections.Immutable;
using System.Text.Json;
using System.Text.Json.Nodes;
namespace Ensemble;

/// <summary>Value-like file policy, not a service or owner. Negative retention
/// keeps all diagnostics, zero keeps none, and positive values keep the newest N.
/// A null policy on a library Agent opts out of automatic lifecycle persistence.</summary>
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
    // after primary initialization succeeds; failed startup preserves prior disk state.
    public partial bool StartPersistence(PersistenceOptions options);
}

public sealed partial class History
{
    // Save is serialized with other saves, not with readers across disk I/O.
    // A candidate is captured once; events appended later remain in live History.
    public partial void Save(PersistenceOptions options);
    internal static partial PersistenceOptions NormalizePersistence(PersistenceOptions options);
    // Validate primary and backup independently. Publish only a complete replay.
    // Returns false only when neither file exists. Writer must not be active.
    public partial bool LoadSave(string path);
    // Replace saved runtime capability state with the current startup snapshot.
    internal partial void ConfigureRuntime(ImmutableArray<SkillPart> skills);
    private static partial SaveDocument ReadSaveFile(string path);
    private static partial (Context Context, long Sequence) ReplaySave(SaveDocument document);
    private static partial void ReplaceFile(string path, string text);
    private static partial string ArchiveText(Context context);
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
