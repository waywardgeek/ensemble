namespace Ensemble;

public sealed partial class History
{
    internal static partial void ValidateCompressionRemoval(CompressionLaunch launch, Event removal)
    {
        if (removal.TaskID != launch.TaskID)
        {
            throw new ArgumentException("Removal has no matching compression task");
        }
        if (launch.Source is null)
        {
            if (removal.Kind != EventKind.Redacted || removal.Level != Redaction.Summary
                || removal.RedactFrom != launch.From || removal.RedactTo != launch.To
                || !removal.Parts.IsEmpty || removal.Reason != "memory compression")
            {
                throw new ArgumentException("Cut differs from the launched conversation selection");
            }
        }
        else if (removal.Kind != EventKind.MemoryRetired || removal.Retirement is not MemoryRetirement retirement
            || retirement.Band != launch.Source || !retirement.IDs.SequenceEqual(launch.SourceIDs))
        {
            throw new ArgumentException("Retirement differs from the launched file selection");
        }
    }

    internal static partial void ValidateCompressionReplacement(CompressionLaunch launch, MemoryEntry replacement)
    {
        if (replacement.Band != launch.Destination || replacement.ID != launch.ID
            || replacement.Timestamp != launch.Timestamp || string.IsNullOrWhiteSpace(replacement.Text))
        {
            throw new ArgumentException("Replacement differs from the launched destination");
        }
    }

    private static partial void ValidateCompressionEvent(Context before, Event value)
    {
        var pending = before.Compression;
        if (value.Compression is CompressionLaunch launch)
        {
            if (pending is not null)
            {
                throw new ArgumentException("Previous compression has no terminal outcome");
            }
            if (before.CompressionOutcome?.TaskID == launch.TaskID)
            {
                throw new ArgumentException("A terminal compression cannot be relaunched");
            }
        }
        if (value.TaskID.Length != 0)
        {
            if (pending is null || pending.Launch.TaskID != value.TaskID || pending.Committed is not null)
            {
                throw new ArgumentException("Compression fact has no matching uncommitted launch");
            }
            if (value.Kind == EventKind.MemoryPopulated)
            {
                if (pending.Removal is null || value.Memory is null)
                {
                    throw new ArgumentException("Replacement requires its held source removal");
                }
                ValidateCompressionReplacement(pending.Launch, value.Memory);
            }
            else
            {
                if (pending.Removal is not null)
                {
                    throw new ArgumentException("Compression already has a held source removal");
                }
                ValidateCompressionRemoval(pending.Launch, value);
            }
        }
        else if (pending is not null && value.Kind == EventKind.Redacted && value.Reason == "memory compression")
        {
            // An older uncorrelated fact cannot borrow this launch's eventual
            // population. Normal redaction and scans retain their old meanings.
            throw new ArgumentException("Managed source removal requires a task identity");
        }

        if (value.Outcome is CompressionOutcome outcome)
        {
            if (pending is null)
            {
                var last = before.CompressionOutcome;
                if (last is null || last.TaskID != outcome.TaskID || last.State != outcome.State || last.Error != outcome.Error)
                {
                    throw new ArgumentException("Outcome has no matching compression");
                }
                return; // Exact repeated finalization is harmless and bounded.
            }
            if (pending.Launch.TaskID != outcome.TaskID
                || (pending.Committed is not null) != (outcome.State == CompressionState.Committed))
            {
                throw new ArgumentException("Outcome disagrees with compression commit state");
            }
        }
    }
}
