using System.Collections.Immutable;
using System.Text;
namespace Ensemble;

public sealed partial class History
{
    private partial (Context Context, long Sequence) ReplaySave(SaveDocument document)
    {
        Context view = document.Context ?? new Context();
        long boundary = document.Context is null ? 0 : document.Sequence;
        long sequence = boundary;
        foreach (var value in document.Events)
        {
            if (value.Seq <= boundary)
            {
                continue; // Retained diagnostics are not applied twice.
            }
            if (value.Seq != checked(sequence + 1))
            {
                throw new FormatException("Save tail is incomplete");
            }
            view = Apply(view, value);
            sequence = value.Seq;
        }
        return (view, Math.Max(document.Sequence, sequence));
    }

    private partial SaveDocument ReadSaveFile(string path)
    {
        using var input = Parent().Sandbox().OpenFile(path, FileMode.Open, FileAccess.Read);
        using var reader = new StreamReader(input);
        var document = JsonCodec.ReadSave(reader.ReadToEnd());
        _ = ReplaySave(document);
        return document;
    }

    public partial bool LoadSave(string path)
    {
        path = Parent().Sandbox().Resolve(path);
        lock (saveGate)
        {
            lock (gate)
            {
                ObjectDisposedException.ThrowIf(destroyed, this);
                if (writer is not null || journalPath is not null)
                {
                    throw new InvalidOperationException("Load before opening History writers");
                }
            }
            var journal = ReadJournal(path + ".journal.jsonl");
            SaveDocument? loaded = null;
            Context restored = new();
            long sequence = 0;
            long snapshotSequence = 0;
            bool found = false;
            foreach (string candidate in new[] { path, path + ".bak" })
            {
                try
                {
                    var document = ReadSaveFile(candidate);
                    var snapshot = ReplaySave(document);
                    var replayed = ReplayJournal(snapshot.Context, snapshot.Sequence, journal);
                    loaded = document;
                    restored = replayed.Context;
                    sequence = replayed.Sequence;
                    snapshotSequence = snapshot.Sequence;
                    if (candidate != path)
                    {
                        Parent().Parent().Log($"persistence: recovered backup and journal at sequence {sequence}");
                    }
                    break;
                }
                catch (Exception error) when (error is FileNotFoundException or DirectoryNotFoundException)
                {
                    continue;
                }
                catch (Exception error) when (error is IOException or InvalidDataException or UnauthorizedAccessException or FormatException)
                {
                    found = true;
                    Parent().Parent().Log("persistence: invalid save candidate; trying backup");
                }
            }
            if (loaded is null)
            {
                if (found)
                {
                    throw new InvalidDataException("No valid save candidate; files left untouched");
                }
                if (journal.IsEmpty)
                {
                    return false;
                }
                // A process can die before its first snapshot. Only a complete
                // journal prefix beginning at one can recover that first session.
                (restored, sequence) = ReplayJournal(new Context(), 0, journal);
            }
            var retained = (loaded?.Events ?? [])
                .AddRange(journal.Where(fact => fact.Seq > snapshotSequence));
            lock (gate)
            {
                events.Clear();
                events.AddRange(retained);
                context = restored;
                foldedSequence = sequence;
            }
            return true;
        }
    }

    internal partial void ConfigureRuntime(ImmutableArray<SkillPart> skills)
    {
        Append(new Event(EventKind.RuntimeConfigured, parts: skills));
    }

    private partial void ReplaceFile(string path, string text)
    {
        Parent().Sandbox().ReplaceFile(path, new UTF8Encoding(false).GetBytes(text));
    }

    internal partial PersistenceOptions NormalizePersistence(PersistenceOptions options)
    {
        ArgumentNullException.ThrowIfNull(options);
        ArgumentException.ThrowIfNullOrWhiteSpace(options.Path);
        var files = Parent().Sandbox();
        string path = files.Resolve(options.Path);
        string? archive = options.ArchivePath is null ? null : files.Resolve(options.ArchivePath);
        if (archive == path || archive == path + ".bak" || archive == path + ".journal.jsonl")
        {
            throw new ArgumentException("Archive must not replace recovery files");
        }
        return new PersistenceOptions(path, options.RetainLog, archive);
    }

    public partial void Save(PersistenceOptions options)
    {
        options = NormalizePersistence(options);
        lock (saveGate)
        {
            if (appendFaulted)
            {
                throw new IOException("Cannot save after an uncertain History append");
            }
            if (journalPath is not null && journalPath != options.Path + ".journal.jsonl")
            {
                throw new ArgumentException("Save path must match the active recovery journal");
            }
            // Replacing an open diagnostic log would leave its writer attached
            // to the old file, splitting later facts from the visible file.
            if (writerPath is not null && (writerPath == options.Path ||
                writerPath == options.Path + ".bak" || writerPath == options.ArchivePath))
            {
                throw new ArgumentException("Persistence files must differ from the active diagnostic log");
            }
            var snapshot = Snapshot();
            var configuration = Parent().Engine().SavedConfiguration(snapshot.Context);
            var retained = options.RetainLog < 0 ? snapshot.Events
                : snapshot.Events.TakeLast(options.RetainLog).ToImmutableArray();
            var candidate = new SaveDocument(snapshot.Context, snapshot.Sequence, configuration, retained);
            string text = JsonCodec.WriteSave(candidate);
            _ = ReplaySave(JsonCodec.ReadSave(text)); // Validate before touching disk.

            if (options.ArchivePath is not null)
            {
                ReplaceFile(options.ArchivePath, ArchiveText(options.ArchivePath, snapshot));
            }
            string? previous = null;
            long backupSequence = 0;
            try
            {
                using var input = Parent().Sandbox().OpenFile(options.Path, FileMode.Open, FileAccess.Read);
                using var reader = new StreamReader(input);
                previous = reader.ReadToEnd();
                backupSequence = ReplaySave(JsonCodec.ReadSave(previous)).Sequence;
            }
            catch (Exception error) when (error is FileNotFoundException or DirectoryNotFoundException or FormatException)
            {
                // Missing or invalid primaries must not replace a good backup.
                previous = null;
            }
            if (previous is not null)
            {
                ReplaceFile(options.Path + ".bak", previous);
            }
            else
            {
                try
                {
                    backupSequence = ReplaySave(ReadSaveFile(options.Path + ".bak")).Sequence;
                }
                catch (Exception error) when (error is IOException or InvalidDataException or UnauthorizedAccessException or FormatException)
                {
                    // No validated backup boundary: retain the whole journal.
                    backupSequence = 0;
                }
            }
            ReplaceFile(options.Path, text);
            TrimJournal(backupSequence);
            lock (gate)
            {
                // Every disk operation succeeded before diagnostics are trimmed.
                // Append cannot race this writer, and Context is never trimmed here.
                events.RemoveAll(value => value.Seq <= snapshot.Sequence
                    && (retained.IsEmpty || value.Seq < retained[0].Seq));
            }
        }
    }
}
