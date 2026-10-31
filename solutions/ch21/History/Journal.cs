using System.Collections.Immutable;
using System.Text;

namespace Ensemble;

public sealed partial class History
{
    internal partial void StartJournal(string path)
    {
        string destination = Parent().Sandbox().Resolve(path) + ".journal.jsonl";
        lock (saveGate)
        {
            ImmutableArray<Event> retained;
            lock (gate)
            {
                ObjectDisposedException.ThrowIf(destroyed, this);
                if (journalPath is not null || appendFaulted)
                {
                    throw new InvalidOperationException("History journal is already active or faulted");
                }
                retained = events.ToImmutableArray();
            }
            if (writerPath == destination)
            {
                throw new ArgumentException("Recovery journal must differ from the active diagnostic log");
            }
            var existing = ReadJournal(destination);
            long last = existing.IsEmpty ? 0 : existing[^1].Seq;
            // Fresh construction may already have recorded a local primary skill.
            // Include those facts too; after recovery the journal already contains
            // its tail, and saved diagnostic events at/before last are not repeated.
            var initial = existing.AddRange(retained.Where(value => value.Seq > last));
            ReplaceFile(destination, JournalText(initial));
            journalPath = destination;
        }
    }

    private partial ImmutableArray<Event> ReadJournal(string path)
    {
        byte[] bytes;
        try
        {
            bytes = Parent().Sandbox().ReadFile(path);
        }
        catch (FileNotFoundException)
        {
            return [];
        }
        catch (DirectoryNotFoundException)
        {
            return [];
        }
        // A killed writer can leave an incomplete UTF-8 character or JSON value.
        // Only newline-terminated records are committed. Ignore the final suffix,
        // never a malformed complete record or an invalid earlier sequence.
        int end = Array.LastIndexOf(bytes, (byte)'\n');
        if (end != bytes.Length - 1)
        {
            Parent().Parent().Log("persistence: ignored an incomplete final journal record");
        }
        if (end < 0)
        {
            return [];
        }
        string text;
        try
        {
            text = new UTF8Encoding(false, true).GetString(bytes, 0, end + 1);
        }
        catch (DecoderFallbackException error)
        {
            throw new InvalidDataException("Journal contains invalid UTF-8", error);
        }
        var facts = ImmutableArray.CreateBuilder<Event>();
        using var reader = new StringReader(text);
        string? line;
        long last = 0;
        while ((line = reader.ReadLine()) is not null)
        {
            Event fact;
            try
            {
                fact = JsonCodec.ReadEvent(line);
            }
            catch (Exception error) when (error is System.Text.Json.JsonException or FormatException
                or ArgumentException or InvalidOperationException or KeyNotFoundException)
            {
                throw new InvalidDataException("Journal contains an invalid complete record", error);
            }
            if (fact.Seq <= 0 || (last != 0 && fact.Seq != last + 1))
            {
                throw new InvalidDataException("Journal sequence is not contiguous");
            }
            facts.Add(fact);
            last = fact.Seq;
        }
        return facts.ToImmutable();
    }

    private partial (Context Context, long Sequence) ReplayJournal(
        Context context, long sequence, ImmutableArray<Event> journal)
    {
        foreach (var fact in journal)
        {
            if (fact.Seq <= sequence)
            {
                continue;
            }
            if (fact.Seq != sequence + 1)
            {
                throw new InvalidDataException("Journal does not continue the selected snapshot");
            }
            context = Apply(context, fact);
            sequence = fact.Seq;
        }
        return (context, sequence);
    }

    private static partial string JournalText(IEnumerable<Event> facts)
    {
        var text = new StringBuilder();
        foreach (var fact in facts)
        {
            text.Append(JsonCodec.WriteEvent(fact));
            text.Append('\n');
        }
        return text.ToString();
    }

    private partial void TrimJournal(long backupSequence)
    {
        if (journalPath is null)
        {
            return;
        }
        // Keep enough to recover from the older, validated BACKUP. Cutting at the
        // new primary would strand that backup if the primary were later damaged.
        var suffix = ReadJournal(journalPath).Where(fact => fact.Seq > backupSequence);
        ReplaceFile(journalPath, JournalText(suffix));
    }
}
