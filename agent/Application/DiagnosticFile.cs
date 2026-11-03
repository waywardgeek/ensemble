using System.Text;

namespace Ensemble;

public sealed partial class DiagnosticFile
{
    public partial DiagnosticFile(FileStream stream) : this(stream, MaximumBytes) { }

    internal partial DiagnosticFile(Stream stream, long maximumBytes)
    {
        ArgumentNullException.ThrowIfNull(stream);
        if (!stream.CanWrite || !stream.CanSeek)
            throw new ArgumentException("Diagnostic files require a writable, seekable stream.", nameof(stream));
        if (maximumBytes < 256 || maximumBytes > MaximumBytes)
            throw new ArgumentOutOfRangeException(nameof(maximumBytes));
        this.stream = stream;
        this.maximumBytes = maximumBytes;
    }

    public override partial Encoding Encoding { get => Utf8; }

    public override partial void WriteLine(string? value)
    {
        value ??= "";
        long originalBytes = Utf8.GetByteCount(value) + 1L;
        int markerBytes = Utf8.GetByteCount(ResetMarker);
        if (originalBytes > maximumBytes - markerBytes)
        {
            value = "{\"diagnostic\":\"record_omitted\",\"reason\":\"oversized\",\"utf8_bytes\":"
                + originalBytes.ToString(System.Globalization.CultureInfo.InvariantCulture) + "}";
        }
        byte[] record = Utf8.GetBytes(value + "\n");
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(closed, this);
            if (failed) throw new IOException("Diagnostic file is unavailable after a write failure.");
            try
            {
                if (stream.Length + record.Length > maximumBytes)
                {
                    // Keep the already authorized handle. Reopening a pathname
                    // would need another Sandbox admission and could target a
                    // replacement file. Reset length AND offset under this lock.
                    stream.Flush();
                    stream.SetLength(0);
                    stream.Position = 0;
                    stream.Write(Utf8.GetBytes(ResetMarker));
                }
                else
                {
                    stream.Position = stream.Length;
                }
                stream.Write(record);
                stream.Flush();
            }
            catch
            {
                // An I/O failure may have left a partial record. Never assume a
                // known append offset or silently resume after that failure.
                failed = true;
                throw;
            }
        }
    }

    public override partial void WriteLine() => WriteLine("");
    public override partial void Write(char value) => throw new NotSupportedException("Diagnostic files accept complete WriteLine records only.");
    public override partial void Write(string? value) => throw new NotSupportedException("Diagnostic files accept complete WriteLine records only.");

    public override partial void Flush()
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(closed, this);
            if (failed) throw new IOException("Diagnostic file is unavailable after a write failure.");
            try { stream.Flush(); }
            catch { failed = true; throw; }
        }
    }

    protected override partial void Dispose(bool disposing)
    {
        if (disposing)
        {
            lock (gate)
            {
                if (closed) return;
                closed = true;
                stream.Dispose();
            }
        }
        base.Dispose(disposing);
    }
}
