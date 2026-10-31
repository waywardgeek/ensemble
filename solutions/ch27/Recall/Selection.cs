using System.Collections.Immutable;
using System.Globalization;
using System.Text;
using System.Text.Json;

namespace Ensemble;

public sealed partial class RecallAttachment
{
    public partial RecallAttachment(string text, IEnumerable<string> hashes)
    {
        ArgumentException.ThrowIfNullOrEmpty(text);
        ArgumentNullException.ThrowIfNull(hashes);
        Text = text;
        Hashes = hashes.ToImmutableArray();
        // Identities travel with the attachment, not a session-wide cache.
        if (Hashes.IsEmpty || Hashes.Any(hash => hash is null || hash.Length != 64
            || hash.Any(character => !char.IsAsciiHexDigitLower(character)))
            || Hashes.Distinct(StringComparer.Ordinal).Count() != Hashes.Length)
        {
            throw new ArgumentException("Recall hashes must be unique SHA-256 lowercase hex values");
        }
    }
}

internal static partial class RecallText
{
    private static partial ImmutableArray<RecallSnippet> Sections(RecallSource source,
        string file, string text, string marker, string parent)
    {
        var output = ImmutableArray.CreateBuilder<RecallSnippet>();
        var header = parent;
        var body = new StringBuilder();
        void Flush()
        {
            var content = body.ToString().Trim();
            if (content.Length > 0) output.Add(new RecallSnippet(source, file, header, content));
            body.Clear();
        }
        foreach (var line in text.Split('\n'))
        {
            if (line.StartsWith(marker, StringComparison.Ordinal))
            {
                Flush();
                var title = line[marker.Length..].Trim();
                header = parent.Length == 0 ? title : parent + " > " + title;
            }
            // Keep heading text in the chunk; the breadcrumb is attribution,
            // not a replacement for searchable source content.
            body.Append(line).Append('\n');
        }
        Flush();
        return output.ToImmutable();
    }

    internal static partial ImmutableArray<RecallSnippet> Chunk(RecallSource source, string file, string text)
    {
        ArgumentNullException.ThrowIfNull(text);
        var output = ImmutableArray.CreateBuilder<RecallSnippet>();
        foreach (var section in Sections(source, file, text, "## ", ""))
        {
            var parts = Encoding.UTF8.GetByteCount(section.Content) <= ChunkBytes
                ? ImmutableArray.Create(section)
                : Sections(source, file, section.Content, "### ", section.Header);
            foreach (var part in parts)
            {
                if (Encoding.UTF8.GetByteCount(part.Content) <= ChunkBytes)
                {
                    output.Add(part);
                    continue;
                }
                // Preserve coherent paragraphs. A single oversized paragraph is
                // indivisible here; Format enforces the later hard wire-byte cap.
                var body = new StringBuilder();
                var bytes = 0;
                foreach (var paragraph in part.Content.Split("\n\n", StringSplitOptions.None))
                {
                    var paragraphBytes = Encoding.UTF8.GetByteCount(paragraph);
                    if (body.Length > 0 && bytes + 2L + paragraphBytes > ChunkBytes)
                    {
                        output.Add(new RecallSnippet(source, file, part.Header, body.ToString()));
                        body.Clear();
                        bytes = 0;
                    }
                    if (body.Length > 0)
                    {
                        body.Append("\n\n");
                        bytes += 2;
                    }
                    body.Append(paragraph);
                    bytes += paragraphBytes;
                }
                if (body.Length > 0)
                {
                    output.Add(new RecallSnippet(source, file, part.Header, body.ToString()));
                }
            }
        }
        return output.ToImmutable();
    }

    internal static partial ImmutableArray<RecallHit> Mechanical(IEnumerable<RecallHit> hits, int limit)
    {
        if (limit <= 0) throw new ArgumentOutOfRangeException(nameof(limit));
        var ranked = Rank(hits);
        var quotas = new Dictionary<RecallSource, int>
        {
            [RecallSource.Memory] = limit / 2 + limit % 2
        };
        var secondary = new[] { RecallSource.Docs, RecallSource.Handoffs, RecallSource.Learnings, RecallSource.Skills }
            .Where(source => ranked.Any(hit => hit.Snippet.Source == source)).ToArray();
        var remaining = limit - quotas[RecallSource.Memory];
        for (var index = 0; index < secondary.Length; index++)
        {
            quotas[secondary[index]] = remaining / secondary.Length;
            if (index < remaining % secondary.Length) quotas[secondary[index]]++;
        }
        var output = ImmutableArray.CreateBuilder<RecallHit>();
        var selected = new HashSet<int>();
        // Reserve the final source shares before spending spare slots. The
        // broad judge candidate pool is deliberately not used as this input.
        for (var index = 0; index < ranked.Length; index++)
        {
            var hit = ranked[index];
            if (quotas.GetValueOrDefault(hit.Snippet.Source) > 0)
            {
                output.Add(hit);
                selected.Add(index);
                quotas[hit.Snippet.Source]--;
            }
        }
        for (var index = 0; index < ranked.Length && output.Count < limit; index++)
        {
            if (!selected.Contains(index)) output.Add(ranked[index]);
        }
        return output.ToImmutable();
    }

    internal static partial ImmutableArray<RecallHit> JudgeChoices(string reply,
        IReadOnlyList<RecallHit> candidates, int limit)
    {
        ArgumentNullException.ThrowIfNull(reply);
        ArgumentNullException.ThrowIfNull(candidates);
        if (limit <= 0) throw new ArgumentOutOfRangeException(nameof(limit));
        var start = reply.IndexOf('[');
        var end = reply.LastIndexOf(']');
        if (start < 0 || end < start) throw new FormatException("Recall judge reply has no array");
        using var document = JsonDocument.Parse(reply[start..(end + 1)]);
        var output = ImmutableArray.CreateBuilder<RecallHit>();
        var seen = new HashSet<int>();
        foreach (var element in document.RootElement.EnumerateArray())
        {
            int index;
            // A syntactically valid mixed array can be partially useful. Null,
            // fractional, overflowing and non-numeric values never become zero.
            var valid = element.ValueKind == JsonValueKind.Number
                ? element.TryGetInt32(out index)
                : int.TryParse(element.ValueKind == JsonValueKind.String ? element.GetString() : null,
                    NumberStyles.Integer, CultureInfo.InvariantCulture, out index);
            if (valid && index >= 0 && index < candidates.Count && seen.Add(index))
            {
                output.Add(candidates[index]);
                if (output.Count == limit) break;
            }
        }
        return output.ToImmutable();
    }

    private static partial string Utf8Prefix(string text, int byteLimit)
    {
        var output = new StringBuilder();
        var remaining = byteLimit;
        foreach (var rune in text.EnumerateRunes())
        {
            if (rune.Utf8SequenceLength > remaining) break;
            output.Append(rune.ToString());
            remaining -= rune.Utf8SequenceLength;
        }
        return output.ToString();
    }

    internal static partial RecallAttachment? Format(IEnumerable<RecallHit> hits, int byteLimit)
    {
        if (byteLimit <= 0) throw new ArgumentOutOfRangeException(nameof(byteLimit));
        const string title = "[Auto-recalled memories]\n";
        var output = new StringBuilder(title);
        var bytes = Encoding.UTF8.GetByteCount(title);
        var hashes = new List<string>();
        var seen = new HashSet<string>(StringComparer.Ordinal);
        foreach (var hit in hits)
        {
            var snippet = hit.Snippet;
            var hash = snippet.Hash();
            if (seen.Contains(hash)) continue;
            var prefix = $"\nFrom {snippet.File} — {snippet.Header}:\n> ";
            var prefixBytes = Encoding.UTF8.GetByteCount(prefix);
            var room = byteLimit - (long)bytes - prefixBytes;
            if (room <= 0) break;
            var body = snippet.Content.Replace("\n", "\n> ", StringComparison.Ordinal);
            var capped = Utf8Prefix(body, (int)room);
            if (capped.Length == 0) break;
            output.Append(prefix).Append(capped);
            bytes += prefixBytes + Encoding.UTF8.GetByteCount(capped);
            hashes.Add(hash);
            seen.Add(hash);
        }
        return hashes.Count == 0 ? null : new RecallAttachment(output.ToString(), hashes);
    }

    internal static partial string Unquote(string text)
    {
        return string.Join('\n', text.Split('\n').Select(line =>
            line.StartsWith("> ", StringComparison.Ordinal) ? line[2..] : line)).Trim();
    }
}
