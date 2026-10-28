using System.Collections.Immutable;
using System.Globalization;
using System.Text;

namespace Ensemble;

public sealed partial class Memory
{
    internal partial MemoryCapture? CaptureConversation(Context context, long budget, DateTimeOffset now)
    {
        ArgumentNullException.ThrowIfNull(context);
        lock (gate)
        {
            RequireRecovered(context);
            var selection = History.SelectConversation(context, budget);
            if (selection is null)
            {
                return null;
            }
            return Capture(context, null, NextSessionID(now), now,
                JsonCodec.WriteMemoryConversation(selection.Entries), []);
        }
    }

    internal partial MemoryCapture? CaptureGraduation(Context context, MemoryBand band,
        long highWatermark, long lowWatermark, DateTimeOffset now)
    {
        ArgumentNullException.ThrowIfNull(context);
        if (band is not (MemoryBand.Session or MemoryBand.Archive8 or MemoryBand.Archive64)
            || lowWatermark <= 0 || highWatermark < lowWatermark)
        {
            throw new ArgumentException("Graduation needs a compressible band and ordered positive watermarks");
        }
        lock (gate)
        {
            RequireRecovered(context);
            var files = context.Memories.Where(entry => entry.Band == band)
                .OrderBy(entry => Coverage(band, entry.ID).Start).ToArray();
            long bytes = files.Sum(entry => (long)Encoding.UTF8.GetByteCount(entry.Text));
            if (bytes <= highWatermark || files.Length < 2)
            {
                return null;
            }
            for (int index = 1; index < files.Length; index++)
            {
                if (Coverage(band, files[index - 1].ID).End.CompareTo(Coverage(band, files[index].ID).Start) >= 0)
                {
                    throw new FormatException("Overlapping selected corpus coverage");
                }
            }
            var selected = ImmutableArray.CreateBuilder<MemoryEntry>();
            foreach (var entry in files)
            {
                if (bytes <= lowWatermark && selected.Count >= 2)
                {
                    break;
                }
                selected.Add(entry);
                bytes -= Encoding.UTF8.GetByteCount(entry.Text);
            }
            var first = Coverage(band, selected[0].ID).Start;
            var last = Coverage(band, selected[^1].ID).End;
            string id = band == MemoryBand.Archive64 ? "MEMORY"
                : first.Day.ToString("yyyy-MM-dd", CultureInfo.InvariantCulture) + "-" + first.Number.ToString(CultureInfo.InvariantCulture)
                    + "--" + last.Day.ToString("yyyy-MM-dd", CultureInfo.InvariantCulture) + "-" + last.Number.ToString(CultureInfo.InvariantCulture);
            return Capture(context, band, id, now, "", selected.ToImmutable());
        }
    }

    private partial MemoryCapture Capture(Context context, MemoryBand? source, string id,
        DateTimeOffset now, string conversation, ImmutableArray<MemoryEntry> sources)
    {
        MemoryBand destination = source is null ? MemoryBand.Session : source.Value - 1;
        string? previous = ReadOptional(CorpusPath(destination, id));
        if (destination != MemoryBand.Memory && (previous is not null
            || context.Memories.Any(entry => entry.Band == destination && entry.ID == id)))
        {
            throw new IOException("Compression destination already exists");
        }
        if (destination == MemoryBand.Memory)
        {
            // Terminal curation must include the exact current document. Do not
            // silently merge human edits absent from the actor's captured view.
            var entries = context.Memories.Where(entry => entry.Band == destination).ToArray();
            if (entries.Length > 1 || (entries.Length == 0 ? null : entries[0].Text) != previous
                || entries.Any(entry => entry.ID != id))
            {
                throw new IOException("Curated memory changed before selection");
            }
        }
        foreach (var entry in sources)
        {
            if (ReadOptional(CorpusPath(entry.Band, entry.ID)) != entry.Text)
            {
                throw new IOException("Selected memory file changed before capture");
            }
        }
        long from = 0;
        long to = 0;
        if (source is null)
        {
            var entries = JsonCodec.ReadMemoryConversation(conversation).Dialogue;
            from = entries[0].Seq;
            to = entries[^1].Seq;
        }
        string hash = JsonCodec.MemoryFingerprint(source, destination, id, conversation, sources, previous);
        string taskID = Guid.NewGuid().ToString("N");
        var launch = new CompressionLaunch(taskID, source, destination, id, now, hash, from, to,
            sources.Select(entry => entry.ID), "memory/compactor-" + taskID + ".jsonl");
        return new MemoryCapture(launch, conversation, sources, previous);
    }

    internal partial bool StillCurrent(MemoryCapture capture, Context context)
    {
        ArgumentNullException.ThrowIfNull(capture);
        ArgumentNullException.ThrowIfNull(context);
        lock (gate)
        {
            var launch = capture.Launch;
            if (context.Compression is not { Committed: null } pending
                || !JsonCodec.SameCompressionLaunch(pending.Launch, launch))
            {
                return false;
            }
            if (launch.Source is null)
            {
                var entries = context.Dialogue.Where(entry => entry.Seq >= launch.From && entry.Seq <= launch.To
                    && !entry.Parts.Any(part => part is SkillPart));
                if (JsonCodec.WriteMemoryConversation(entries) != capture.Conversation)
                {
                    return false;
                }
            }
            else
            {
                var prefix = context.Memories.Where(entry => entry.Band == launch.Source)
                    .OrderBy(entry => Coverage(entry.Band, entry.ID).Start)
                    .Take(capture.Sources.Length).ToArray();
                if (prefix.Length != capture.Sources.Length)
                {
                    return false;
                }
                for (int index = 0; index < prefix.Length; index++)
                {
                    var before = capture.Sources[index];
                    var current = prefix[index];
                    if (current.ID != before.ID || current.Timestamp != before.Timestamp || current.Text != before.Text)
                    {
                        return false;
                    }
                }
            }
            var destination = context.Memories.SingleOrDefault(entry =>
                entry.Band == launch.Destination && entry.ID == launch.ID);
            if (destination?.Text != capture.DestinationText)
            {
                return false;
            }
            return CurrentFiles(capture, allowRetired: false)
                && ReadOptional(CorpusPath(launch.Destination, launch.ID)) == capture.DestinationText;
        }
    }

    private partial bool CurrentFiles(MemoryCapture capture, bool allowRetired)
    {
        foreach (var source in capture.Sources)
        {
            string path = CorpusPath(source.Band, source.ID);
            string? live = ReadOptional(path);
            string? retired = ReadOptional(path + ".retired");
            if (retired is not null && retired != source.Text)
            {
                return false;
            }
            if (live is not null)
            {
                if (live != source.Text)
                {
                    return false;
                }
            }
            else if (!allowRetired || retired != source.Text)
            {
                return false;
            }
        }
        return true;
    }
}
