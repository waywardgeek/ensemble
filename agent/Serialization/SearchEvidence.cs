using System.Collections.Immutable;

using System.Text;

namespace Ensemble;

public sealed partial class SearchLine
{
    public partial SearchLine(string path, long number, string text, IEnumerable<SearchSpan> matches)
    {
        ArgumentNullException.ThrowIfNull(text);
        ArgumentNullException.ThrowIfNull(matches);
        if (string.IsNullOrEmpty(path) || path.Contains('\0') || number < 1
            || Encoding.UTF8.GetByteCount(path) + (long)Encoding.UTF8.GetByteCount(text) > SearchEvidence.MaximumBytes)
            throw new ArgumentException("Invalid search line");
        var spans = matches.Take(SearchEvidence.MaximumSpansPerLine + 1).ToImmutableArray();
        if (spans.Length > SearchEvidence.MaximumSpansPerLine)
            throw new ArgumentException("Too many search spans");
        int end = 0;
        foreach (var span in spans)
        {
            if (span.Start < end || span.Length < 0 || (long)span.Start + span.Length > text.Length)
                throw new ArgumentException("Invalid search span");
            end = span.Start + span.Length;
        }
        Path = path; Number = number; Text = text; Matches = spans;
    }
}
public sealed partial class SearchEvidence
{
    public partial SearchEvidence(IEnumerable<SearchLine> lines, long totalLines)
    {
        ArgumentNullException.ThrowIfNull(lines);
        var captured = lines.Take(MaximumLines + 1).ToImmutableArray();
        long bytes = 0;
        foreach (var line in captured)
        {
            ArgumentNullException.ThrowIfNull(line);
            bytes += Encoding.UTF8.GetByteCount(line.Path) + (long)Encoding.UTF8.GetByteCount(line.Text);
        }
        if (captured.Length > MaximumLines || totalLines < captured.Length || bytes > MaximumBytes)
            throw new ArgumentException("Invalid search evidence bounds");
        Lines = captured; TotalLines = totalLines;
    }
}
