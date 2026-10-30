using System.Collections.Immutable;
using System.Security.Cryptography;
using System.Text;

namespace Ensemble;

public sealed partial class RecallSettings
{
    public partial RecallSettings(bool disabled, bool disableJudge, string judgeModel,
        double minScore, int maxSnippets, int minQueryLength, int candidates,
        int contextMessages, int maxRecallBytes)
    {
        ArgumentNullException.ThrowIfNull(judgeModel);
        if (!double.IsFinite(minScore) || minScore < 0 || maxSnippets < 0
            || minQueryLength < 0 || candidates < 0 || contextMessages < 0 || maxRecallBytes < 0)
        {
            throw new ArgumentOutOfRangeException(nameof(minScore), "Recall limits must be finite and nonnegative");
        }
        Disabled = disabled;
        DisableJudge = disableJudge;
        JudgeModel = judgeModel;
        MinScore = minScore == 0 ? 3 : minScore;
        MaxSnippets = maxSnippets == 0 ? 3 : maxSnippets;
        MinQueryLength = minQueryLength == 0 ? 80 : minQueryLength;
        Candidates = candidates == 0 ? 20 : candidates;
        ContextMessages = contextMessages == 0 ? 3 : contextMessages;
        MaxRecallBytes = maxRecallBytes == 0 ? 6144 : maxRecallBytes;
    }
}

internal sealed partial class RecallSnippet
{
    internal partial RecallSnippet(RecallSource source, string file, string header, string content)
    {
        if (!Enum.IsDefined(source)) throw new ArgumentOutOfRangeException(nameof(source));
        ArgumentException.ThrowIfNullOrEmpty(file);
        ArgumentNullException.ThrowIfNull(header);
        ArgumentException.ThrowIfNullOrEmpty(content);
        Source = source;
        File = file;
        Header = header;
        Content = content;
    }

    internal partial string Hash()
    {
        return Convert.ToHexStringLower(SHA256.HashData(Encoding.UTF8.GetBytes(Content)));
    }
}

internal sealed partial class RecallIndex
{
    internal partial RecallIndex(IEnumerable<RecallSnippet> snippets)
    {
        ArgumentNullException.ThrowIfNull(snippets);
        var entries = ImmutableArray.CreateBuilder<RecallDocument>();
        var counts = new Dictionary<string, int>(StringComparer.Ordinal);
        RecallSource? source = null;
        double totalLength = 0;
        foreach (var snippet in snippets)
        {
            ArgumentNullException.ThrowIfNull(snippet);
            if (source is not null && snippet.Source != source)
            {
                throw new ArgumentException("Each recall index must contain exactly one source");
            }
            source = snippet.Source;
            var words = RecallText.Tokens(snippet.Content);
            var terms = new Dictionary<string, int>(StringComparer.Ordinal);
            foreach (var word in words)
            {
                terms[word] = terms.GetValueOrDefault(word) + 1;
            }
            // DF counts documents containing a term, not occurrences of that term.
            foreach (var word in terms.Keys)
            {
                counts[word] = counts.GetValueOrDefault(word) + 1;
            }
            entries.Add(new RecallDocument(snippet, terms.ToImmutableDictionary(StringComparer.Ordinal), words.Length));
            totalLength += words.Length;
        }
        documents = entries.ToImmutable();
        frequencies = counts.ToImmutableDictionary(StringComparer.Ordinal);
        averageLength = documents.Length == 0 ? 0 : totalLength / documents.Length;
    }

    internal partial ImmutableArray<RecallHit> Search(IEnumerable<string> query, double floor)
    {
        ArgumentNullException.ThrowIfNull(query);
        if (!double.IsFinite(floor) || floor < 0) throw new ArgumentOutOfRangeException(nameof(floor));
        if (averageLength == 0) return [];
        var terms = query.Distinct(StringComparer.Ordinal).ToArray();
        var hits = ImmutableArray.CreateBuilder<RecallHit>();
        foreach (var document in documents)
        {
            double score = 0;
            foreach (var term in terms)
            {
                var frequency = document.Terms.GetValueOrDefault(term);
                if (frequency == 0) continue;
                var documentFrequency = frequencies[term];
                // Smoothed IDF stays positive in a one-document corpus. k1=1.2
                // saturates repetition; b=.75 discounts long keyword-heavy prose.
                var idf = Math.Log(1 + (documents.Length - documentFrequency + 0.5)
                    / (documentFrequency + 0.5));
                var denominator = frequency + 1.2 * (0.25 + 0.75 * document.Length / averageLength);
                score += idf * frequency * 2.2 / denominator;
            }
            if (score > 0 && score >= floor)
            {
                hits.Add(new RecallHit(document.Snippet, score));
            }
        }
        return RecallText.Rank(hits);
    }
}

internal static partial class RecallText
{
    internal static partial ImmutableArray<string> Tokens(string text, bool query)
    {
        ArgumentNullException.ThrowIfNull(text);
        var output = ImmutableArray.CreateBuilder<string>();
        var seen = new HashSet<string>(StringComparer.Ordinal);
        var word = new StringBuilder();
        var length = 0;
        void Flush()
        {
            if (length >= 2)
            {
                var token = word.ToString();
                if (!query || (!StopWords.Contains(" " + token + " ", StringComparison.Ordinal) && seen.Add(token)))
                {
                    output.Add(token);
                }
            }
            word.Clear();
            length = 0;
        }
        // Count Unicode scalars, not UTF-16 code units or UTF-8 bytes. A single
        // supplementary letter is still a one-character word and is discarded.
        foreach (var rune in text.EnumerateRunes())
        {
            if (Rune.IsLetterOrDigit(rune))
            {
                word.Append(Rune.ToLowerInvariant(rune).ToString());
                length++;
            }
            else
            {
                Flush();
            }
        }
        Flush();
        return output.ToImmutable();
    }

    internal static partial ImmutableArray<RecallHit> Rank(IEnumerable<RecallHit> hits)
    {
        // LINQ's stable ordering preserves deterministic corpus/file/chunk order
        // on ties, rather than introducing a hash-table-dependent tie breaker.
        return hits.OrderByDescending(hit => hit.Score).ToImmutableArray();
    }
}
