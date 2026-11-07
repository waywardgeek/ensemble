namespace Ensemble;

public sealed partial class History
{
    private static partial List<Entry> HandoffDialogue(List<Entry> dialogue, Event value)
    {
        var retained = new List<Entry>();
        foreach (var entry in dialogue)
        {
            // The newest checkpoint replaces all older notes in active context.
            // Canonical events retain every note for archival replay. Media and
            // tool traffic leave active context too; accompanying prose, ordinary
            // system text and skills remain unchanged.
            var indices = Enumerable.Range(0, entry.Parts.Length)
                .Where(index => entry.Parts[index] is not ToolCallPart and not ToolResultPart
                    and not OpaquePart and not BlobPart and not HandoffPart)
                .ToArray();
            if (indices.Length == 0)
            {
                continue;
            }
            var ids = entry.PartIDs.IsEmpty ? [] : indices.Select(index => entry.PartIDs[index]);
            // Private signatures can also ride on visible text. Keep its prose
            // and display identity, not the provider replay envelope. Original
            // immutable parts remain in canonical events for historical replay.
            var parts = indices.Select(index => entry.Parts[index] is TextPart text
                ? new TextPart(text.Text) : entry.Parts[index]);
            retained.Add(new Entry(entry.Seq, entry.Actor, parts, entry.ResponseID, ids, entry.Time, entry.ObservationDuration));
        }
        retained.Add(new Entry(value.Seq, Actor.System, value.Parts, time: value.Time));
        return retained;
    }
}
