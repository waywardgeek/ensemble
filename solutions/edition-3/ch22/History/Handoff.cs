namespace Ensemble;

public sealed partial class History
{
    private static partial List<Entry> HandoffDialogue(List<Entry> dialogue, Event value)
    {
        var retained = new List<Entry>();
        foreach (var entry in dialogue)
        {
            // Notes close completed segments. Retain older notes until a memory
            // replacement commits; tool clearing alone must not erase their
            // reasons. Ordinary system text and skills also remain unchanged.
            var indices = Enumerable.Range(0, entry.Parts.Length)
                .Where(index => entry.Parts[index] is not ToolCallPart and not ToolResultPart
                    and not OpaquePart)
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
            retained.Add(new Entry(entry.Seq, entry.Actor, parts, entry.ResponseID, ids));
        }
        retained.Add(new Entry(value.Seq, Actor.System, value.Parts));
        return retained;
    }
}
