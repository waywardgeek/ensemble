using System.Collections.Immutable;
using System.Globalization;
using System.Text;

namespace Ensemble;

public sealed partial class Memory
{
    public partial Memory(Agent parent)
    {
        ArgumentNullException.ThrowIfNull(parent);
        parent.AddMemory(this);
    }

    public partial Agent Parent() => parent ?? throw new ObjectDisposedException(nameof(Memory));

    public partial void Destroy()
    {
        Agent? owner;
        lock (gate)
        {
            if (destroyed)
            {
                return;
            }
            destroyed = true;
            owner = parent;
        }
        // Do not acquire Agent ingress while retaining the disk-operation gate.
        // Logical destruction preserves the durable corpus and recovery receipt.
        owner?.RemoveMemory(this);
    }

    private partial string HomePath(string relative)
    {
        ObjectDisposedException.ThrowIf(destroyed, this);
        return Path.Combine(Parent().DataDirectory(), relative);
    }

    private static partial string BandDirectory(MemoryBand band) => band switch
    {
        MemoryBand.Session => "memory",
        MemoryBand.Archive8 => "bucket-0",
        MemoryBand.Archive64 => "bucket-1",
        MemoryBand.Soul or MemoryBand.Memory => "",
        _ => throw new ArgumentException("Unknown corpus band", nameof(band))
    };

    private static partial (DateOnly Day, int Number) ParsePoint(string value)
    {
        if (value.Length < 12 || value[10] != '-'
            || !DateOnly.TryParseExact(value[..10], "yyyy-MM-dd", CultureInfo.InvariantCulture,
                DateTimeStyles.None, out var day)
            || !int.TryParse(value.AsSpan(11), NumberStyles.None, CultureInfo.InvariantCulture, out int number)
            || number <= 0 || number.ToString(CultureInfo.InvariantCulture) != value[11..])
        {
            throw new FormatException("Invalid dated corpus identity");
        }
        return (day, number);
    }

    private static partial ((DateOnly Day, int Number) Start, (DateOnly Day, int Number) End)
        Coverage(MemoryBand band, string id)
    {
        if (band == MemoryBand.Session)
        {
            var point = ParsePoint(id);
            return (point, point);
        }
        if (band is not (MemoryBand.Archive8 or MemoryBand.Archive64))
        {
            throw new ArgumentException("Stable files do not have coverage ranges", nameof(band));
        }
        string[] range = id.Split("--", StringSplitOptions.None);
        if (range.Length != 2)
        {
            throw new FormatException("Corpus range needs two dated endpoints");
        }
        var start = ParsePoint(range[0]);
        var end = ParsePoint(range[1]);
        if (start.CompareTo(end) > 0)
        {
            throw new FormatException("Reversed corpus range");
        }
        return (start, end);
    }

    private static partial DateTimeOffset FileTimestamp(MemoryBand band, string id)
    {
        if (band is MemoryBand.Soul or MemoryBand.Memory)
        {
            return DateTimeOffset.UnixEpoch;
        }
        var point = Coverage(band, id).Start;
        // This is an identity-derived ordering timestamp, not filesystem mtime.
        // One tick per positive int keeps numeric order within the named day;
        // even int.MaxValue fits well inside a day. Context can then use its
        // existing timestamp/ordinal ordering without placing 10 before 2.
        return new DateTimeOffset(point.Day.ToDateTime(TimeOnly.MinValue, DateTimeKind.Utc)).AddTicks(point.Number);
    }

    private partial string CorpusPath(MemoryBand band, string id)
    {
        if (band is MemoryBand.Soul or MemoryBand.Memory)
        {
            string expected = band == MemoryBand.Soul ? "SOUL" : "MEMORY";
            if (id != expected)
            {
                throw new FormatException("Invalid stable corpus identity");
            }
            return HomePath(expected + ".md");
        }
        _ = Coverage(band, id); // Validate before using any supplied identity as a path.
        return HomePath(Path.Combine(BandDirectory(band), id + ".md"));
    }

    private partial string? ReadOptional(string path)
    {
        try
        {
            using var input = Parent().Sandbox().OpenFile(path, FileMode.Open, FileAccess.Read);
            if (input.Length > MaximumFileBytes)
            {
                throw new IOException("Memory file exceeds the disk safety limit");
            }
            using var output = new MemoryStream();
            byte[] buffer = new byte[16 * 1024];
            int count;
            while ((count = input.Read(buffer, 0, buffer.Length)) != 0)
            {
                if (output.Length + count > MaximumFileBytes)
                {
                    throw new IOException("Memory file grew beyond the disk safety limit");
                }
                output.Write(buffer, 0, count);
            }
            return new UTF8Encoding(false, true).GetString(output.ToArray());
        }
        catch (FileNotFoundException)
        {
            return null;
        }
        catch (DirectoryNotFoundException)
        {
            return null;
        }
    }

    private partial void RequireRecovered(Context context)
    {
        if (context.Compression is not null || ReadOptional(HomePath(".memory-stage.json")) is not null)
        {
            throw new InvalidOperationException("Recover memory transactions before scanning or selecting files");
        }
    }

    internal partial ImmutableArray<MemoryEntry> Scan(Context recovered, MemoryBand band,
        MemoryCapture? live)
    {
        ArgumentNullException.ThrowIfNull(recovered);
        lock (gate)
        {
            if (live is null)
            {
                RequireRecovered(recovered);
            }
            else if (recovered.Compression is not { Committed: null } pending
                || !JsonCodec.SameCompressionLaunch(live.Launch, pending.Launch)
                || band == live.Launch.Source || band == live.Launch.Destination
                || ReadOptional(HomePath(".memory-stage.json")) is not null)
            {
                throw new InvalidOperationException("Only an unrelated band may be scanned during a live unstaged attempt");
            }
            var entries = new List<MemoryEntry>();
            if (band is MemoryBand.Soul or MemoryBand.Memory)
            {
                string id = band == MemoryBand.Soul ? "SOUL" : "MEMORY";
                string? text = ReadOptional(CorpusPath(band, id));
                if (text is not null)
                {
                    entries.Add(new MemoryEntry(band, id, FileTimestamp(band, id), text));
                }
                return entries.ToImmutableArray();
            }

            string directory = HomePath(BandDirectory(band));
            string[] names;
            try
            {
                names = Parent().Sandbox().ReadDirectory(directory);
            }
            catch (FileNotFoundException) { return []; }
            catch (DirectoryNotFoundException) { return []; }
            foreach (string name in names)
            {
                if (!name.EndsWith(".md", StringComparison.Ordinal))
                {
                    continue; // .retired, temporary files and traces are not active input.
                }
                string id = name[..^3];
                string? text = ReadOptional(CorpusPath(band, id));
                if (text is null)
                {
                    throw new IOException("Corpus changed during scan");
                }
                entries.Add(new MemoryEntry(band, id, FileTimestamp(band, id), text));
            }
            entries.Sort((left, right) => Coverage(band, left.ID).Start.CompareTo(Coverage(band, right.ID).Start));
            for (int index = 1; index < entries.Count; index++)
            {
                if (Coverage(band, entries[index - 1].ID).End.CompareTo(Coverage(band, entries[index].ID).Start) >= 0)
                {
                    throw new FormatException("Overlapping corpus coverage");
                }
            }
            return entries.ToImmutableArray();
        }
    }

    private partial string NextSessionID(DateTimeOffset now)
    {
        string day = now.UtcDateTime.ToString("yyyy-MM-dd", CultureInfo.InvariantCulture);
        int next = 1;
        string[] names;
        try
        {
            names = Parent().Sandbox().ReadDirectory(HomePath("memory"));
        }
        catch (FileNotFoundException) { names = []; }
        catch (DirectoryNotFoundException) { names = []; }
        foreach (string name in names)
        {
            string stem;
            if (name.EndsWith(".md.retired", StringComparison.Ordinal))
            {
                stem = name[..^11];
            }
            else if (name.EndsWith(".md", StringComparison.Ordinal))
            {
                stem = name[..^3];
            }
            else
            {
                continue;
            }
            var point = ParsePoint(stem);
            if (stem.StartsWith(day + "-", StringComparison.Ordinal))
            {
                next = Math.Max(next, checked(point.Number + 1));
            }
        }
        return day + "-" + next.ToString(CultureInfo.InvariantCulture);
    }
}

public sealed partial class Agent
{
    public partial Memory Memory()
    {
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            return memory ?? throw new InvalidOperationException("Agent has no memory owner");
        }
    }

    internal partial void AddMemory(Memory child)
    {
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed || admissionClosed, this);
            if (child.destroyed || child.parent is not null || memory is not null)
            {
                throw new InvalidOperationException("Memory ownership conflict");
            }
            memory = child;
            child.parent = this;
        }
    }

    internal partial void RemoveMemory(Memory child)
    {
        lock (ingress)
        {
            if (!ReferenceEquals(memory, child) || !ReferenceEquals(child.parent, this))
            {
                throw new InvalidOperationException("Memory does not belong to this Agent");
            }
            memory = null;
            child.parent = null;
        }
    }
}
