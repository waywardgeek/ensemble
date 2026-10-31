using System.Collections.Immutable;

namespace Ensemble;

/// <summary>Independent scoring corpora. Memory includes current and archived
/// personal memory; its statistics must not be diluted by a large docs tree.</summary>
internal enum RecallSource { Memory, Docs, Handoffs, Learnings, Skills }

/// <summary>Immutable value-like policy. Zero selects the chapter default;
/// disabled bypasses recall, while disableJudge keeps mechanical retrieval.</summary>
public sealed partial class RecallSettings
{
    public readonly bool Disabled;
    public readonly bool DisableJudge;
    public readonly string JudgeModel;
    public readonly double MinScore;
    public readonly int MaxSnippets;
    public readonly int MinQueryLength;
    public readonly int Candidates;
    public readonly int ContextMessages;
    public readonly int MaxRecallBytes;
    public partial RecallSettings(bool disabled = false, bool disableJudge = false,
        string judgeModel = "", double minScore = 0, int maxSnippets = 0,
        int minQueryLength = 0, int candidates = 0, int contextMessages = 0,
        int maxRecallBytes = 0);
}

/// <summary>Value-like attributed source bytes. File is display metadata, not
/// authority to open a path. Content includes its heading so titles are searchable.</summary>
internal sealed partial class RecallSnippet
{
    internal readonly RecallSource Source;
    internal readonly string File;
    internal readonly string Header;
    internal readonly string Content;
    internal partial RecallSnippet(RecallSource source, string file, string header, string content);
    // Hash full content, not a truncated attachment: edits permit future recall.
    internal partial string Hash();
}

// Derived immutable statistics, not a live filesystem or object relationship.
internal sealed record RecallDocument(RecallSnippet Snippet,
    ImmutableDictionary<string, int> Terms, int Length);
internal sealed record RecallHit(RecallSnippet Snippet, double Score);

/// <summary>Value-like index snapshot for exactly one source. Recaller replaces
/// it after a fresh corpus scan; searching never mutates its statistics.</summary>
internal sealed partial class RecallIndex
{
    private readonly ImmutableArray<RecallDocument> documents;
    private readonly ImmutableDictionary<string, int> frequencies;
    private readonly double averageLength;
    internal partial RecallIndex(IEnumerable<RecallSnippet> snippets);
    internal partial ImmutableArray<RecallHit> Search(IEnumerable<string> query, double floor);
}

/// <summary>Recorded value, independent of disk/judge availability on replay.
/// Hashes identify full snippets actually represented, even when byte-capped.</summary>
public sealed partial class RecallAttachment
{
    public readonly string Text;
    public readonly ImmutableArray<string> Hashes;
    public partial RecallAttachment(string text, IEnumerable<string> hashes);
}

/// <summary>Pure retrieval operations. They neither read files nor publish
/// History, and successful judge choices never pass through mechanical quotas.</summary>
internal static partial class RecallText
{
    private const int ChunkBytes = 4096;
    private const string StopWords = " a an and are as at be been but by can could did do does for from had has have how if in into is it its may me my not of on or our should so that the their them then there these they this to was we were what when where which who will with would you your ";
    internal static partial ImmutableArray<string> Tokens(string text, bool query = false);
    internal static partial ImmutableArray<RecallSnippet> Chunk(RecallSource source, string file, string text);
    private static partial ImmutableArray<RecallSnippet> Sections(RecallSource source,
        string file, string text, string marker, string parent);
    internal static partial ImmutableArray<RecallHit> Rank(IEnumerable<RecallHit> hits);
    internal static partial ImmutableArray<RecallHit> Mechanical(IEnumerable<RecallHit> hits, int limit);
    // Invalid array syntax throws; a valid empty array returns no choices.
    internal static partial ImmutableArray<RecallHit> JudgeChoices(string reply,
        IReadOnlyList<RecallHit> candidates, int limit);
    internal static partial RecallAttachment? Format(IEnumerable<RecallHit> hits, int byteLimit);
    // Filter before candidate limiting or final quotas. Instructions are the
    // caller's actual projected instructions, not a second skill-policy model.
    internal static partial ImmutableArray<RecallHit> Available(IEnumerable<RecallHit> hits,
        Context context, string instructions);
    internal static partial string Unquote(string text);
    private static partial string Utf8Prefix(string text, int byteLimit);
}
