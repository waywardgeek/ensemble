namespace Ensemble;

public sealed partial class History
{
    private partial Context Apply(Context before, Event value)
    {
        var after = Reduce(before, value);
        if (ReferenceEquals(before, after))
        {
            Parent().Parent().Log($"history: skipped event {value.Seq} ({value.Kind}): {ReductionProblem(before, value)}");
        }
        return after;
    }

    private static partial string? ReductionProblem(Context before, Event value)
    {
        // Validation is before mutation. In particular, a malformed runtime
        // configuration must not first remove the current instruction/grant set.
        switch (value.Kind)
        {
            case EventKind.CompressionLaunched:
                if (value.Compression is null)
                {
                    return "missing compression launch";
                }
                break;
            case EventKind.CompressionFinished:
                if (value.Outcome is null)
                {
                    return "missing compression outcome";
                }
                break;
            case EventKind.MemoryPopulated:
                if (value.Memory is null)
                {
                    return "missing memory population";
                }
                break;
            case EventKind.MemoryRetired:
                if (value.Retirement is null)
                {
                    return "missing memory retirement";
                }
                if (!before.Memories.Any(item => item.Band == value.Retirement.Band
                    && (value.Retirement.IDs.IsEmpty || value.Retirement.IDs.Contains(item.ID))))
                {
                    return "memory retirement has no retained entries";
                }
                break;
            case EventKind.MessageReceived:
            case EventKind.HintReceived:
                if (value.Actor is null)
                {
                    return "missing message actor";
                }
                break;
            case EventKind.HandoffRecorded:
                if (value.Parts.Length != 1 || value.Parts[0] is not HandoffPart)
                {
                    return "expected one continuation note";
                }
                break;
            case EventKind.SkillLoaded:
            case EventKind.RuntimeConfigured:
                var skills = value.Parts.OfType<SkillPart>().ToArray();
                if ((value.Kind == EventKind.SkillLoaded && skills.Length == 0)
                    || skills.Length != value.Parts.Length
                    || skills.Select(skill => skill.Name).Distinct(StringComparer.Ordinal).Count() != skills.Length
                    || skills.Count(skill => skill.Primary) > 1)
                {
                    return "expected distinct skill snapshots";
                }
                break;
            case EventKind.SkillUnloadPending:
                if (string.IsNullOrWhiteSpace(value.Name)
                    || !before.Dialogue.SelectMany(entry => entry.Parts).OfType<SkillPart>()
                        .Any(skill => skill.Name == value.Name))
                {
                    return "pending unload has no retained manual";
                }
                break;
            case EventKind.Redacted:
                if (value.RedactFrom <= 0 || value.RedactTo < value.RedactFrom || value.Level is null)
                {
                    return "invalid redaction range or missing level";
                }
                if (!before.Dialogue.Any(entry => entry.Seq >= value.RedactFrom && entry.Seq <= value.RedactTo))
                {
                    return "redaction range has no retained entries";
                }
                if (value.Level == Redaction.Summary
                    && value.Parts.Any(part => part is ToolCallPart or ToolResultPart or SkillPart))
                {
                    return "summary cannot introduce tool traffic or skill grants";
                }
                break;
        }
        return null;
    }
}
