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
            // Stream the validated prefix, omitting only an uncommitted final
            // suffix. Fresh startup facts may follow that prefix. Do not buffer
            // a perpetual audit journal merely to reopen its append path.
            IEnumerable<Event> Initial()
            {
                long last = 0;
                foreach (var fact in ReadJournal(destination))
                {
                    last = fact.Seq;
                    yield return fact;
                }
                foreach (var fact in retained)
                {
                    if (fact.Seq > last)
                    {
                        yield return fact;
                    }
                }
            }
            WriteJournal(destination, Initial());
            journalPath = destination;
        }
    }

    private partial IEnumerable<Event> ReadJournal(string path)
    {
        FileStream input;
        try
        {
            input = Parent().Sandbox().OpenFile(path, FileMode.Open, FileAccess.Read);
        }
        catch (FileNotFoundException)
        {
            yield break;
        }
        catch (DirectoryNotFoundException)
        {
            yield break;
        }
        using (input)
        using (var buffered = new BufferedStream(input))
        using (var record = new MemoryStream())
        {
            var encoding = new UTF8Encoding(false, true);
            long last = 0;
            int next;
            while ((next = buffered.ReadByte()) != -1)
            {
                if (next != '\n')
                {
                    record.WriteByte((byte)next);
                    continue;
                }
                Event fact;
                try
                {
                    string line = encoding.GetString(record.GetBuffer(), 0, checked((int)record.Length));
                    fact = JsonCodec.ReadEvent(line);
                }
                catch (Exception error) when (error is DecoderFallbackException or System.Text.Json.JsonException
                    or FormatException or ArgumentException or InvalidOperationException or KeyNotFoundException)
                {
                    throw new InvalidDataException("Journal contains an invalid complete record", error);
                }
                if (fact.Seq <= 0 || (last != 0 && fact.Seq != last + 1))
                {
                    throw new InvalidDataException("Journal sequence is not contiguous");
                }
                record.SetLength(0);
                last = fact.Seq;
                yield return fact;
            }
            // Decode only committed records. A torn final UTF-8 character is
            // just as uncommitted as a torn JSON object. Earlier corruption is
            // never excused by this rule.
            if (record.Length != 0)
            {
                Parent().Parent().Log("persistence: ignored an incomplete final journal record");
            }
        }
    }

    private partial HistorySnapshot ReplayJournal(
        Context context, long sequence, IEnumerable<Event> journal,
        long latestHandoff, long previousHandoff)
    {
        var retained = new List<Event>();
        foreach (var fact in journal)
        {
            if (fact.Seq > sequence)
            {
                if (fact.Seq != sequence + 1)
                {
                    throw new InvalidDataException("Journal does not continue the selected snapshot");
                }
                context = Apply(context, fact);
                sequence = fact.Seq;
            }
            if (fact.Kind == EventKind.HandoffRecorded && fact.Seq > latestHandoff)
            {
                previousHandoff = latestHandoff;
                latestHandoff = fact.Seq;
                retained.RemoveAll(value => value.Seq < previousHandoff);
            }
            if (fact.Seq >= previousHandoff)
            {
                retained.Add(fact);
            }
        }
        return new HistorySnapshot(retained.ToImmutableArray(), context, sequence, latestHandoff, previousHandoff);
    }

    private partial void WriteJournal(string path, IEnumerable<Event> facts)
    {
        var files = Parent().Sandbox();
        string directory = Path.GetDirectoryName(path)!;
        files.CreateDirectory(directory);
        string temporary = Path.Combine(directory, "." + Path.GetFileName(path) + "." + Guid.NewGuid().ToString("N") + ".tmp");
        try
        {
            using (var output = files.OpenFile(temporary, FileMode.CreateNew, FileAccess.Write))
            {
                using (var writer = new StreamWriter(output, new UTF8Encoding(false), leaveOpen: true))
                {
                    foreach (var fact in facts)
                    {
                        writer.WriteLine(JsonCodec.WriteEvent(fact));
                    }
                    writer.Flush();
                }
                output.Flush(flushToDisk: true);
            }
            // Enumerating/validating the input and flushing the replacement both
            // finish before publication. A failure leaves the old journal intact.
            files.MoveFile(temporary, path);
        }
        catch
        {
            try
            {
                files.DeleteFile(temporary);
            }
            catch (Exception cleanup) when (cleanup is IOException or UnauthorizedAccessException)
            {
                // Preserve the original error; a leftover temporary is not a
                // recovery candidate.
            }
            throw;
        }
    }

    private partial void TrimJournal(long backupSequence)
    {
        if (journalPath is null || backupSequence == 0 || previousHandoff == 0)
        {
            return;
        }
        // Keep the preceding handoff itself and every fact needed to advance
        // the validated backup. Neither requirement can weaken the other.
        long cutoff = Math.Min(backupSequence, previousHandoff - 1);
        WriteJournal(journalPath, ReadJournal(journalPath).Where(fact => fact.Seq > cutoff));
    }
}
