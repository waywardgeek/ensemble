using System.Text;

namespace Ensemble;

public sealed partial class Memory
{
    internal partial void Stage(MemoryCapture capture, string summary)
    {
        ArgumentNullException.ThrowIfNull(capture);
        var stage = new MemoryStage(capture, summary);
        byte[] bytes = new UTF8Encoding(false, true).GetBytes(JsonCodec.WriteMemoryStage(stage));
        if (bytes.Length > MaximumFileBytes)
        {
            throw new IOException("Memory candidate receipt exceeds the disk safety limit");
        }
        lock (gate)
        {
            string path = HomePath(".memory-stage.json");
            if (ReadOptional(path) is not null)
            {
                throw new InvalidOperationException("A memory stage already exists; recover it before staging another");
            }
            // Flush and atomic rename happen below the Agent-owned Sandbox.
            // The candidate is outside every active corpus directory.
            Parent().Sandbox().ReplaceFile(path, bytes);
        }
    }

    private partial MemoryStage ReadStage()
    {
        string text = ReadOptional(HomePath(".memory-stage.json"))
            ?? throw new IOException("Committed memory transaction has no staged receipt");
        return JsonCodec.ReadMemoryStage(text);
    }

    private partial void DropStage()
    {
        Parent().Sandbox().DeleteFile(HomePath(".memory-stage.json"));
    }

    internal partial CompressionOutcome? Recover(Context context)
    {
        ArgumentNullException.ThrowIfNull(context);
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            var pending = context.Compression;
            if (pending is null)
            {
                // No replay authority can accept these bytes. This also releases
                // a completed receipt only AFTER its terminal fact was journaled.
                DropStage();
                return null;
            }
            if (pending.Committed is null)
            {
                // A launched worker cannot survive process restart. No removal
                // was applied by replay, so abandoning even a malformed candidate
                // is safe and leaves every source active.
                DropStage();
                return new CompressionOutcome(pending.Launch.TaskID, CompressionState.Abandoned);
            }
            var stage = ReadStage();
            FinalizeStage(pending, stage);
            return new CompressionOutcome(pending.Launch.TaskID, CompressionState.Committed);
        }
    }

    private partial void FinalizeStage(PendingCompression pending, MemoryStage stage)
    {
        var capture = stage.Capture;
        var launch = capture.Launch;
        if (!JsonCodec.SameCompressionLaunch(launch, pending.Launch)
            || pending.Committed is null || pending.Removal is null
            || pending.Committed.Band != launch.Destination || pending.Committed.ID != launch.ID
            || pending.Committed.Timestamp != launch.Timestamp || pending.Committed.Text != stage.Summary)
        {
            throw new IOException("Staged memory candidate does not match the committed replay receipt");
        }
        string destination = CorpusPath(launch.Destination, launch.ID);
        string? current = ReadOptional(destination);
        if (current != capture.DestinationText && current != stage.Summary)
        {
            throw new IOException("Committed memory destination changed before recovery");
        }
        if (!CurrentFiles(capture, allowRetired: true))
        {
            throw new IOException("Committed memory source changed before recovery");
        }

        // Preflight every source before the first mutation. A prior crash may
        // already have promoted the same bytes or retired any prefix of files.
        if (current != stage.Summary)
        {
            Parent().Sandbox().ReplaceFile(destination, new UTF8Encoding(false, true).GetBytes(stage.Summary));
        }
        foreach (var source in capture.Sources)
        {
            string path = CorpusPath(source.Band, source.ID);
            string? live = ReadOptional(path);
            string? retired = ReadOptional(path + ".retired");
            if (live is null && retired == source.Text)
            {
                continue;
            }
            if (live != source.Text || (retired is not null && retired != source.Text))
            {
                throw new IOException("Memory source changed during finalization");
            }
            if (retired is null)
            {
                // Same-directory rename is atomic, avoiding a partial archive
                // copy after a crash. Checks are optimistic, not an OS lock on
                // external editors racing this final operation.
                Parent().Sandbox().MoveFile(path, path + ".retired");
            }
            else
            {
                // Identical archived bytes already exist; retaining another live
                // name would resurrect the source on the next scan.
                Parent().Sandbox().DeleteFile(path);
            }
        }
        // The actor must durably record the returned outcome before releasing
        // this receipt. A crash here can safely repeat the entire operation.
    }
}
