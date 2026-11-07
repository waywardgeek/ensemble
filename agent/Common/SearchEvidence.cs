using System.Collections.Immutable;

namespace Ensemble;

public readonly record struct SearchSpan(int Start, int Length);
public sealed partial class SearchLine
{
    public readonly string Path;
    public readonly long Number;
    public readonly string Text;
    public readonly ImmutableArray<SearchSpan> Matches;
    public partial SearchLine(string path, long number, string text, IEnumerable<SearchSpan> matches);
}
public sealed partial class SearchEvidence
{
    public const int MaximumLines = 256;
    public const int MaximumBytes = 32768;
    public const int MaximumSpansPerLine = 256;
    public readonly ImmutableArray<SearchLine> Lines;
    public readonly long TotalLines;
    public bool Complete => Lines.Length == TotalLines;
    public partial SearchEvidence(IEnumerable<SearchLine> lines, long totalLines);
}
