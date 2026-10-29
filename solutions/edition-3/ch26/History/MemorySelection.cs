using System.Collections.Immutable;
using System.Text;

namespace Ensemble;

internal sealed partial class ConversationSelection
{
    internal partial ConversationSelection(IEnumerable<Entry> entries)
    {
        Entries = entries.ToImmutableArray();
        if (Entries.IsEmpty || !Entries[^1].Parts.Any(part => part is HandoffPart))
        {
            throw new ArgumentException("Conversation selection must end at a checkpoint", nameof(entries));
        }
    }

    internal partial string Text()
    {
        var text = new StringBuilder();
        foreach (var entry in Entries)
        {
            // Only readable material is compressor input. The exact source
            // entries are retained separately for validation and archival.
            text.Append(entry.Actor).Append(": ")
                .Append(JsonCodec.Text(entry.Parts)).Append('\n');
        }
        return text.ToString();
    }
}

public sealed partial class History
{
    internal static partial ConversationSelection? SelectConversation(Context context, long budget)
    {
        ArgumentNullException.ThrowIfNull(context);
        if (budget <= 0)
        {
            throw new ArgumentOutOfRangeException(nameof(budget), "Selection requires a resolved positive byte budget");
        }

        // A skill entry is an instruction/grant, not conversation to summarize.
        // Notes describe segment boundaries but do not inflate dialogue usage.
        var dialogue = context.Dialogue
            .Where(entry => !entry.Parts.Any(part => part is SkillPart))
            .ToArray();
        static long DialogueBytes(Entry entry) => Encoding.UTF8.GetByteCount(
            JsonCodec.Text(entry.Parts.Where(part => part is not HandoffPart)));
        long total = dialogue.Sum(DialogueBytes);
        if (total <= budget)
        {
            return null;
        }

        var selected = new List<Entry>();
        int completeCount = 0;
        long removed = 0;
        long completeBytes = 0;
        foreach (var entry in dialogue)
        {
            // An unpaired/unpresented batch cannot be a completed prefix. Actor
            // handoff normally strips these parts first; pure callers and replay
            // still get a conservative refusal rather than a partial batch cut.
            if (entry.Parts.Any(part => part is ToolCallPart or ToolResultPart or OpaquePart
                || part is TextPart { From: not null } or TextPart { Opaque: not null }))
            {
                break;
            }
            selected.Add(entry);
            removed += DialogueBytes(entry);
            if (entry.Parts.Any(part => part is HandoffPart))
            {
                completeCount = selected.Count;
                completeBytes = removed;
                if (total - removed <= budget)
                {
                    break;
                }
            }
        }
        if (completeCount == 0 || completeBytes == 0)
        {
            return null;
        }

        // If the open tail alone exceeds budget, all completed segments may be
        // selected, but never part of that tail. No mutation occurs here.
        return new ConversationSelection(selected.Take(completeCount));
    }
}
