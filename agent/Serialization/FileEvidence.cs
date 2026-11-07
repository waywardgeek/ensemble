using System.Text;

namespace Ensemble;

public sealed partial class FileExcerpt
{
    public partial FileExcerpt(string text, long totalBytes)
    {
        ArgumentNullException.ThrowIfNull(text);
        int bytes = Encoding.UTF8.GetByteCount(text);
        if (bytes > MaximumBytes || totalBytes < bytes)
            throw new ArgumentException("Invalid file excerpt bounds");
        Text = text;
        TotalBytes = totalBytes;
    }

    public static partial FileExcerpt Capture(string text)
    {
        ArgumentNullException.ThrowIfNull(text);
        long total = Encoding.UTF8.GetByteCount(text);
        int length = 0, bytes = 0;
        foreach (var rune in text.EnumerateRunes())
        {
            if (bytes + rune.Utf8SequenceLength > MaximumBytes) break;
            bytes += rune.Utf8SequenceLength;
            length += rune.Utf16SequenceLength;
        }
        return new FileExcerpt(text[..length], total);
    }
}
public sealed partial class FileEvidence
{
    public partial FileEvidence(FileOperation operation, string path, long startLine,
        long startColumn, FileExcerpt content, FileExcerpt? removed)
    {
        ArgumentNullException.ThrowIfNull(content);
        if (!Enum.IsDefined(operation) || string.IsNullOrEmpty(path) || path.Contains('\0')
            || Encoding.UTF8.GetByteCount(path) > 32768 || startLine < 1 || startColumn < 1
            || (operation == FileOperation.Edit) != (removed is not null))
            throw new ArgumentException("Invalid file operation evidence");
        Operation = operation;
        Path = path;
        StartLine = startLine;
        StartColumn = startColumn;
        Content = content;
        Removed = removed;
    }
}
