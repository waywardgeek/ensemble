using System.Collections.Immutable;
using System.Text;
namespace Ensemble;

public sealed partial class History
{
    private static partial (Context Context, long Sequence) ReplaySave(SaveDocument document)
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
            view = Reduce(view, value);
            sequence = value.Seq;
        }
        return (view, Math.Max(document.Sequence, sequence));
    }

    private static partial SaveDocument ReadSaveFile(string path)
    {
        var document = JsonCodec.ReadSave(File.ReadAllText(path));
        _ = ReplaySave(document);
        return document;
    }

    public partial bool LoadSave(string path)
    {
        lock (saveGate)
        {
            SaveDocument? loaded = null;
            bool found = false;
            foreach (string candidate in new[] { path, path + ".bak" })
            {
                try
                {
                    loaded = ReadSaveFile(candidate);
                    if (candidate != path)
                    {
                        Parent().Parent().Log($"persistence: recovered backup at sequence {loaded.Sequence}");
                    }
                    break;
                }
                catch (Exception error) when (error is FileNotFoundException or DirectoryNotFoundException)
                {
                    continue;
                }
                catch (Exception ex) when (ex is IOException or UnauthorizedAccessException or FormatException)
                {
                    // Never include remote content, credential strings, or corrupt
                    // file bytes in the diagnostic channel.
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
                return false;
            }
            var restored = ReplaySave(loaded);
            lock (gate)
            {
                ObjectDisposedException.ThrowIf(destroyed, this);
                if (writer is not null)
                {
                    throw new InvalidOperationException("Cannot load while an output log is active");
                }
                events.Clear();
                events.AddRange(loaded.Events);
                context = restored.Context;
                foldedSequence = restored.Sequence;
            }
            return true;
        }
    }

    internal partial void ConfigureRuntime(ImmutableArray<SkillPart> skills)
    {
        Append(new Event(EventKind.RuntimeConfigured, parts: skills));
    }

    private static partial void ReplaceFile(string path, string text)
    {
        string fullPath = System.IO.Path.GetFullPath(path);
        string directory = System.IO.Path.GetDirectoryName(fullPath)!;
        Directory.CreateDirectory(directory);
        string temporary = System.IO.Path.Combine(directory, "." + System.IO.Path.GetFileName(path) + "." + Guid.NewGuid().ToString("N") + ".tmp");
        try
        {
            using (var stream = new FileStream(temporary, FileMode.CreateNew, FileAccess.Write, FileShare.None))
            {
                byte[] bytes = new UTF8Encoding(false).GetBytes(text);
                stream.Write(bytes);
                stream.Flush(flushToDisk: true);
            }
            // Same-directory replacement avoids exposing a partly written JSON file.
            File.Move(temporary, fullPath, overwrite: true);
        }
        finally
        {
            if (File.Exists(temporary))
            {
                File.Delete(temporary);
            }
        }
    }

    private static partial string ArchiveText(Context context)
    {
        // This derivative contains only dialogue snapshots. It is not accepted by
        // LoadSave and cannot become an alternative recovery authority.
        var text = new StringBuilder();
        foreach (var entry in context.Dialogue)
        {
            if (entry.Parts.Any(part => part is SkillPart))
            {
                continue;
            }
            var fact = new Event(EventKind.MessageReceived, seq: entry.Seq, actor: entry.Actor, parts: entry.Parts);
            text.AppendLine(JsonCodec.WriteEvent(fact));
        }
        return text.ToString();
    }

    internal static partial PersistenceOptions NormalizePersistence(PersistenceOptions options)
    {
        ArgumentNullException.ThrowIfNull(options);
        ArgumentException.ThrowIfNullOrWhiteSpace(options.Path);
        string path = System.IO.Path.GetFullPath(options.Path);
        string? archive = options.ArchivePath is null ? null : System.IO.Path.GetFullPath(options.ArchivePath);
        if (archive == path || archive == path + ".bak")
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
            var snapshot = Snapshot();
            var configuration = Parent().Engine().SavedConfiguration(snapshot.Context);
            var retained = options.RetainLog < 0 ? snapshot.Events
                : snapshot.Events.TakeLast(options.RetainLog).ToImmutableArray();
            var candidate = new SaveDocument(snapshot.Context, snapshot.Sequence, configuration, retained);
            string text = JsonCodec.WriteSave(candidate);
            _ = ReplaySave(JsonCodec.ReadSave(text)); // Validate before touching disk.

            if (options.ArchivePath is not null)
            {
                ReplaceFile(options.ArchivePath, ArchiveText(snapshot.Context));
            }
            if (File.Exists(options.Path))
            {
                string previous = File.ReadAllText(options.Path);
                try
                {
                    _ = ReplaySave(JsonCodec.ReadSave(previous));
                    ReplaceFile(options.Path + ".bak", previous);
                }
                catch (FormatException)
                {
                    // Recovery may have selected the backup. Never rotate damage
                    // over that known-good file when publishing the repaired primary.
                    Parent().Parent().Log("persistence: corrupt primary not promoted to backup");
                }
            }
            ReplaceFile(options.Path, text);
            lock (gate)
            {
                // Publication succeeded. Keep any events appended during disk I/O;
                // trim only the captured diagnostic prefix, never live Context.
                events.RemoveAll(value => value.Seq <= snapshot.Sequence
                    && (retained.IsEmpty || value.Seq < retained[0].Seq));
            }
        }
    }
}
