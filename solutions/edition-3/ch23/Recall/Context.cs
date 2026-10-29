using System.Collections.Immutable;
using System.Text;

namespace Ensemble;

internal static partial class RecallText
{
    internal static partial ImmutableArray<RecallHit> Available(IEnumerable<RecallHit> hits,
        Context context, string instructions)
    {
        ArgumentNullException.ThrowIfNull(hits);
        ArgumentNullException.ThrowIfNull(context);
        ArgumentNullException.ThrowIfNull(instructions);

        // Text and identities answer different questions. Complete readable
        // context suppresses an already-visible snippet; a retained hash also
        // suppresses the unseen suffix of a byte-capped attachment. Neither is a
        // persistent cache. Removing the context entry permits later retrieval.
        var text = new StringBuilder(instructions);
        foreach (var memory in context.Memories)
        {
            text.Append('\n').Append(memory.Text);
        }
        var hashes = new HashSet<string>(StringComparer.Ordinal);
        foreach (var entry in context.Dialogue)
        {
            // JsonCodec.Text includes nested tool results, notes and recall but
            // not opaque provider replay or tool arguments. Manuals are supplied
            // in the actual instruction projection, not counted a second time.
            text.Append('\n').Append(JsonCodec.Text(entry.Parts));
            foreach (var recall in entry.Parts.OfType<RecallPart>())
            {
                hashes.UnionWith(recall.Attachment.Hashes);
            }
        }
        text.Append('\n').Append(JsonCodec.Text(context.Ephemeral));
        string present = Unquote(text.ToString());
        var available = ImmutableArray.CreateBuilder<RecallHit>();
        foreach (var hit in hits)
        {
            if (!hashes.Contains(hit.Snippet.Hash())
                && !present.Contains(Unquote(hit.Snippet.Content), StringComparison.Ordinal))
            {
                available.Add(hit);
            }
        }
        // Keep input order and the full candidate pool. The judge limit and
        // mechanical source quotas belong to their later, separate operations.
        return available.ToImmutable();
    }
}
