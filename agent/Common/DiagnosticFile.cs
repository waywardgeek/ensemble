using System.Text;

namespace Ensemble;

// An owned diagnostic file resource, not another logger identity or service bag.
// Callers transfer a seekable file on successful construction and close this
// writer only after their work joins. Borrowed diagnostic writers are not wrapped.
public sealed partial class DiagnosticFile : TextWriter
{
    public const long MaximumBytes = 64L * 1024 * 1024;
    internal const string ResetMarker = "{\"diagnostic\":\"reset\",\"reason\":\"size_limit\"}\n";
    private static readonly Encoding Utf8 = new UTF8Encoding(false);
    private readonly Stream stream;
    private readonly long maximumBytes;
    private readonly object gate = new();
    private bool closed;
    private bool failed;

    public partial DiagnosticFile(FileStream stream);
    internal partial DiagnosticFile(Stream stream, long maximumBytes);
    public override partial Encoding Encoding { get; }
    public override partial void WriteLine(string? value);
    public override partial void WriteLine();
    public override partial void Write(char value);
    public override partial void Write(string? value);
    public override partial void Flush();
    protected override partial void Dispose(bool disposing);
}
