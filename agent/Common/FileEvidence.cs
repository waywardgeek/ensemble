namespace Ensemble;

// Bounded operation evidence for presentation; never reread a mutable path.
public enum FileOperation { Read = 1, Write, Append, Edit }
public sealed partial class FileExcerpt
{
    public const int MaximumBytes = 32768;
    public readonly string Text;
    public readonly long TotalBytes;
    public bool Complete => System.Text.Encoding.UTF8.GetByteCount(Text) == TotalBytes;
    public partial FileExcerpt(string text, long totalBytes);
    public static partial FileExcerpt Capture(string text);
}
public sealed partial class FileEvidence
{
    public readonly FileOperation Operation;
    public readonly string Path;
    public readonly long StartLine;
    public readonly long StartColumn;
    public readonly FileExcerpt Content;
    public readonly FileExcerpt? Removed;
    public partial FileEvidence(FileOperation operation, string path, long startLine,
        long startColumn, FileExcerpt content, FileExcerpt? removed = null);
}
