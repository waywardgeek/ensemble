using System.Collections.Immutable;
using System.Text;
using System.Text.Json;

namespace Ensemble;

public sealed partial class History
{
    private partial ImmutableArray<Event> ReadLog(string path)
    {
        // Invalid UTF-8 must not silently become replacement characters in a
        // supposedly exact canonical archive.
        using var input = Parent().Sandbox().OpenFile(path, FileMode.Open, FileAccess.Read);
        using var reader = new StreamReader(input, new UTF8Encoding(false, true));
        var loaded = new List<Event>();
        long previous = 0;
        int lineNumber = 0;
        while (reader.ReadLine() is string line)
        {
            lineNumber++;
            // An explicit unknown version is refused. Historical headerless logs
            // are current format, so their first event must not be discarded.
            if (lineNumber == 1)
            {
                using var first = JsonDocument.Parse(line);
                if (first.RootElement.TryGetProperty("log_version", out var version))
                {
                    if (!version.TryGetInt32(out var n) || n != 1)
                    {
                        throw new FormatException("Unsupported log version");
                    }

                    continue;
                }
            }
            if (string.IsNullOrWhiteSpace(line))
            {
                continue;
            }

            try
            {
                var value = JsonCodec.ReadEvent(line);
                if (value.Seq <= previous)
                {
                    throw new FormatException("Sequence must increase strictly");
                }

                previous = value.Seq;
                loaded.Add(value);
            }
            catch (Exception ex) when (ex is JsonException or FormatException or ArgumentException or InvalidOperationException or KeyNotFoundException)
            {
                throw new FormatException($"Invalid log at line {lineNumber}: {ex.Message}", ex);
            }
        }
        return loaded.ToImmutableArray();
    }

    private partial string ArchiveText(string path, HistorySnapshot snapshot)
    {
        ImmutableArray<Event> archived;
        try
        {
            archived = ReadLog(path);
        }
        catch (FileNotFoundException)
        {
            archived = [];
        }
        catch (DirectoryNotFoundException)
        {
            archived = [];
        }

        // An existing archive must itself be a full prefix. Do not conceal its
        // corruption merely because another source happens to fill a missing line.
        long expected = 0;
        foreach (var value in archived)
        {
            if (value.Seq != ++expected)
            {
                throw new InvalidDataException("Archive is not a complete event prefix");
            }
        }

        // These are temporary algorithm values, not another mutable History.
        // Canonical serialization compares all event fields, including private
        // provider replay material that working-context reduction may discard.
        var lines = new SortedDictionary<long, string>();
        void Include(IEnumerable<Event> source)
        {
            foreach (var value in source)
            {
                if (value.Seq > snapshot.Sequence)
                {
                    throw new InvalidDataException("Archive source extends beyond current History");
                }
                string line = JsonCodec.WriteEvent(value);
                if (lines.TryGetValue(value.Seq, out string? previous) && previous != line)
                {
                    throw new InvalidDataException($"Archive sources disagree at sequence {value.Seq}");
                }
                lines[value.Seq] = line;
            }
        }

        Include(archived);
        if (journalPath is not null)
        {
            Include(ReadJournal(journalPath));
        }
        Include(snapshot.Events);

        var text = new StringBuilder();
        text.AppendLine("{\"log_version\":1}");
        expected = 0;
        foreach (var (sequence, line) in lines)
        {
            if (sequence != ++expected)
            {
                throw new InvalidDataException("Complete event archive unavailable: historical events are missing");
            }
            text.AppendLine(line);
        }
        if (expected != snapshot.Sequence)
        {
            throw new InvalidDataException("Complete event archive unavailable: historical events are missing");
        }
        return text.ToString();
    }
}
