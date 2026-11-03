using System.Text;
using Ensemble;

internal static class DiagnosticFileTests
{
    private sealed class FailingStream : MemoryStream
    {
        public bool FailWrites;
        public bool FailReset;
        public override void SetLength(long value)
        {
            if (FailReset) throw new IOException("Synthetic reset failure");
            base.SetLength(value);
        }
        public int Attempts;
        public override void Write(byte[] buffer, int offset, int count)
        {
            Attempts++;
            if (FailWrites) throw new IOException("Synthetic diagnostic write failure");
            base.Write(buffer, offset, count);
        }
        public override void Write(ReadOnlySpan<byte> buffer)
        {
            Attempts++;
            if (FailWrites) throw new IOException("Synthetic diagnostic write failure");
            base.Write(buffer);
        }
    }

    public static async Task Bounds()
    {
        const int limit = 512;
        var stream = new MemoryStream();
        using (var writer = new DiagnosticFile(stream, limit))
        {
            writer.WriteLine(new string('é', 150));
            Check(stream.Length == 301, "Limit must count UTF-8 bytes and newline, not characters");
            writer.WriteLine(new string('界', 100));
            string reset = Encoding.UTF8.GetString(stream.ToArray());
            Check(reset.StartsWith(DiagnosticFile.ResetMarker, StringComparison.Ordinal), "Reset marker missing");
            Check(!reset.Contains('é') && !reset.Contains('\0'), "Reset retained old bytes or a sparse offset");
            Check(stream.Length <= limit, "Reset exceeded limit");
            writer.WriteLine(new string('x', 1000));
            string omitted = Encoding.UTF8.GetString(stream.ToArray());
            Check(omitted.Contains("record_omitted") && omitted.Contains("1001"), "Oversized record lacked truthful byte count");
            Check(!omitted.Contains(new string('x', 20)), "Oversized payload leaked into notice");
            await Task.WhenAll(Enumerable.Range(0, 8).Select(index => Task.Run(() =>
            {
                for (int record = 0; record < 30; record++) writer.WriteLine($"record:{index}:{record}");
            })));
            Check(stream.Length <= limit, "Concurrent writes exceeded limit");
            string[] lines = Encoding.UTF8.GetString(stream.ToArray()).Split('\n', StringSplitOptions.RemoveEmptyEntries);
            Check(lines.All(line => line.StartsWith("record:", StringComparison.Ordinal) || line == DiagnosticFile.ResetMarker.TrimEnd('\n')),
                "Concurrent reset split or mixed records");
        }
        Check(!stream.CanWrite, "Owned stream survived writer disposal");
        Check(DiagnosticFile.MaximumBytes == 64L * 1024 * 1024, "Unexpected production cap");

        var failing = new FailingStream();
        using var failedWriter = new DiagnosticFile(failing, limit);
        failing.FailWrites = true;
        try { failedWriter.WriteLine("first"); throw new Exception("Write failure was hidden"); }
        catch (IOException) { }
        int attempts = failing.Attempts;
        failing.FailWrites = false;
        try { failedWriter.WriteLine("must not resume"); throw new Exception("Poisoned writer resumed"); }
        catch (IOException) { }
        Check(failing.Attempts == attempts, "Writer attempted another effect after failed write");

        var resetFailure = new FailingStream();
        using var resetWriter = new DiagnosticFile(resetFailure, limit);
        resetWriter.WriteLine(new string('a', 300));
        byte[] preserved = resetFailure.ToArray();
        resetFailure.FailReset = true;
        try { resetWriter.WriteLine(new string('b', 300)); throw new Exception("Reset failure was hidden"); }
        catch (IOException) { }
        Check(resetFailure.ToArray().SequenceEqual(preserved), "Refused reset changed existing bytes");
        resetFailure.FailReset = false;
        try { resetWriter.WriteLine("must not resume"); throw new Exception("Failed reset writer resumed"); }
        catch (IOException) { }

    }

    public static Task RealFile()
    {
        string directory = Path.Combine(Path.GetTempPath(), "diagnostic-bound-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(directory);
        string path = Path.Combine(directory, "api.log");
        string sentinel = Path.Combine(directory, "save.json");
        File.WriteAllText(sentinel, "durable sentinel");
        try
        {
            using (var writer = new DiagnosticFile(new FileStream(path, FileMode.Create, FileAccess.Write, FileShare.Read), 512))
            {
                for (int record = 0; record < 40; record++) writer.WriteLine(new string('界', 100));
                Check(new FileInfo(path).Length <= 512, "Real file exceeded bound");
            }
            byte[] bytes = File.ReadAllBytes(path);
            Check(!bytes.Contains((byte)0), "Real reset left a sparse hole");
            Check(Encoding.UTF8.GetString(bytes).StartsWith(DiagnosticFile.ResetMarker, StringComparison.Ordinal), "Real reset marker absent");
            Check(Directory.GetFiles(directory).Length == 2, "Diagnostic reset created a prior file");
            Check(File.ReadAllText(sentinel) == "durable sentinel", "Diagnostic reset touched saved state");
        }
        finally { Directory.Delete(directory, true); }
        return Task.CompletedTask;
    }

    private static void Check(bool valid, string message)
    {
        if (!valid) throw new Exception(message);
    }
}
