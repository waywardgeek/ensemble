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

}
