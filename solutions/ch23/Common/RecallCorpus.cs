using System.Collections.Immutable;

namespace Ensemble;

/// <summary>Agent-owned local search indexes. Construction does no disk work.
/// History and Memory still own replay and recovery; this object only reads
/// eligible corpus bytes and caches derived, immutable search statistics.</summary>
public sealed partial class Recaller
{
    internal Agent? parent;
    internal bool destroyed;
    private readonly object gate = new();
    // A fingerprint covers filenames and exact file bytes, not lengths or dates.
    // Replace the entire set only after all sources have scanned successfully.
    private ImmutableDictionary<RecallSource, (string Fingerprint, RecallIndex Index)> indexes
        = ImmutableDictionary<RecallSource, (string, RecallIndex)>.Empty;

    public partial Recaller(Agent parent);
    public partial Agent Parent();
    public partial void Destroy();
    // Internal until the actor's startup/input barrier owns automatic retrieval.
    // Skills owns the separately selected, read-only skill-document corpus.
    internal partial ImmutableArray<RecallHit> Search(string query, double minimum,
        CancellationToken cancellation = default);
}

public sealed partial class Agent
{
    private Recaller? recaller;
    // Optional library Agents have no implicit filesystem corpus.
    internal partial bool HasDataDirectory();
    public partial Recaller Recaller();
    internal partial void AddRecaller(Recaller child);
    internal partial void RemoveRecaller(Recaller child);
}

public sealed partial class Memory
{
    // Serialize memory reads with stage/promotion/retirement. This never performs
    // recovery or publishes History on behalf of the actor.
    internal partial ImmutableArray<(string File, string Text, string Hash)> RecallFiles(
        CancellationToken cancellation);
}

internal static partial class RecallText
{
    // Temporary values, not another authoritative corpus. Sandbox is borrowed
    // for this operation through the caller's actual owning Agent.
    internal static partial ImmutableArray<(string File, string Text, string Hash)> ReadCorpus(
        Sandbox sandbox, string home, RecallSource source, CancellationToken cancellation);
    private static partial string? ResolveCorpusPath(Sandbox sandbox, string home, string path);
    private static partial void ReadCorpusDirectory(Sandbox sandbox, string home, string directory,
        RecallSource source, HashSet<string> visited,
        ImmutableArray<(string File, string Text, string Hash)>.Builder files,
        CancellationToken cancellation);
    private static partial void ReadCorpusFile(Sandbox sandbox, string home, string path,
        ImmutableArray<(string File, string Text, string Hash)>.Builder files,
        CancellationToken cancellation);
    // Shared bounded decoding; caller owns the stream and its read authority.
    internal static partial (string File, string Text, string Hash) ReadCorpusStream(Stream stream, string relative, CancellationToken cancellation);
    internal static partial string CorpusFingerprint(
        ImmutableArray<(string File, string Text, string Hash)> files);
}
