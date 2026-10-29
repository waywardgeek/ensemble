using System.Text;
using System.Text.Json;
using Ensemble;

// Pure algorithms operate on actual immutable source snapshots. Filesystem,
// actor integration and live judging are deliberately outside this slice.
internal static class RecallTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition) throw new Exception(message);
    }

    private static void Refuses(Action action)
    {
        try { action(); }
        catch (Exception error) when (error is ArgumentException or FormatException or JsonException) { return; }
        throw new Exception("Expected invalid recall input to refuse");
    }

    private static RecallSnippet Snippet(string content, RecallSource source = RecallSource.Memory,
        string file = "memory/decision.md", string header = "Decision") => new(source, file, header, content);

    private static RecallHit Hit(RecallSource source, double score, string content) => new(Snippet(content, source), score);

    public static void Values()
    {
        var settings = new RecallSettings();
        Check(!settings.Disabled && !settings.DisableJudge && settings.JudgeModel == ""
            && settings.MinScore == 3 && settings.MaxSnippets == 3 && settings.MinQueryLength == 80
            && settings.Candidates == 20 && settings.ContextMessages == 3 && settings.MaxRecallBytes == 6144,
            "zero selects each documented default");
        var explicitSettings = new RecallSettings(true, true, "chosen", 0.5, 5, 7, 8, 9, 100);
        Check(explicitSettings.Disabled && explicitSettings.DisableJudge && explicitSettings.JudgeModel == "chosen"
            && explicitSettings.MinScore == .5 && explicitSettings.MaxSnippets == 5 && explicitSettings.MaxRecallBytes == 100,
            "explicit values survive");
        Refuses(() => new RecallSettings(minScore: double.NaN));
        Refuses(() => new RecallSettings(minScore: double.PositiveInfinity));
        Refuses(() => new RecallSettings(minScore: -1));
        Refuses(() => new RecallSettings(maxSnippets: -1));
        Refuses(() => new RecallSettings(minQueryLength: -1));
        Refuses(() => new RecallSettings(candidates: -1));
        Refuses(() => new RecallSettings(contextMessages: -1));
        Refuses(() => new RecallSettings(maxRecallBytes: -1));
        var original = Snippet("Original exact bytes");
        var list = new List<RecallSnippet> { original };
        var index = new RecallIndex(list);
        list.Clear();
        Check(index.Search(["original"], 0).Single().Snippet == original, "index freezes caller collection");
        var hashes = new List<string> { original.Hash() };
        var attachment = new RecallAttachment("recorded", hashes);
        hashes.Clear();
        Check(attachment.Hashes.Single() == original.Hash(), "attachment freezes hashes");
        Check(Snippet(original.Content, RecallSource.Docs).Hash() == original.Hash(), "identity follows content, not source path");
        Check(Snippet(original.Content + " ").Hash() != original.Hash(), "even a byte edit changes identity");
        Refuses(() => new RecallAttachment("recorded", []));
        Refuses(() => new RecallAttachment("recorded", [original.Hash(), original.Hash()]));
        Refuses(() => new RecallAttachment("recorded", ["not-a-hash"]));
        Refuses(() => new RecallIndex([original, Snippet("other", RecallSource.Docs)]));
    }

    public static void Chunks()
    {
        Check(RecallText.Tokens("I A x # Redis's r2 cache-line CAFÉ 𐐀 𐐀𐐀").SequenceEqual(
            new[] { "redis", "r2", "cache", "line", "café", "𐐨𐐨" }), "alphanumeric Unicode runs and rune length");
        Check(RecallText.Tokens("THE the Redis redis and why", query: true).SequenceEqual(
            new[] { "redis", "why" }), "only query drops stop words and repeated terms");
        Check(RecallText.Tokens("the the Redis").SequenceEqual(new[] { "the", "the", "redis" }),
            "documents retain stop words and repetition for statistics");
        var shortText = "Preface fact.\n\n## Architecture\nSQLite chosen.\n### Details\nReason remains together.\n## Next\nLater fact.";
        var shortChunks = RecallText.Chunk(RecallSource.Docs, "docs/spec.md", shortText);
        Check(shortChunks.Length == 3 && shortChunks[0].Header == "" && shortChunks[1].Header == "Architecture"
            && shortChunks[1].Content.Contains("### Details") && shortChunks[2].Header == "Next", "small sections stay coherent with preamble");
        Check(shortChunks.All(chunk => chunk.File == "docs/spec.md" && chunk.Source == RecallSource.Docs), "source attribution survives splitting");
        var largeText = "## Architecture\n### Storage\n" + new string('a', 2600)
            + "\n\n" + new string('b', 2600) + "\n### Deployment\nLocal only.";
        var largeChunks = RecallText.Chunk(RecallSource.Memory, "memory/big.md", largeText);
        Check(largeChunks.Count(chunk => chunk.Header == "Architecture > Storage") == 2,
            "oversized section splits by child heading then paragraphs");
        Check(largeChunks.Last().Header == "Architecture > Deployment" && largeChunks.Last().Content.Contains("Local only"), "breadcrumb follows parent");
        Check(largeChunks.All(chunk => Encoding.UTF8.GetByteCount(chunk.Content) <= 4096), "divisible paragraphs honor chunk size");
        var indivisible = RecallText.Chunk(RecallSource.Memory, "memory/long.md", "## Long\n" + new string('x', 5000));
        Check(indivisible.Single().Content.Length > 4096, "indivisible paragraph awaits attachment cap rather than losing text");
        Check(RecallText.Chunk(RecallSource.Docs, "empty.md", " \n\n ").IsEmpty, "empty sections discarded");
    }

    public static void Scoring()
    {
        var singleton = new RecallIndex([Snippet("quartz")]);
        var score = singleton.Search(["quartz"], 0).Single().Score;
        Check(Math.Abs(score - Math.Log(1 + .5 / 1.5)) < 1e-12, "exact smoothed singleton BM25");
        Check(singleton.Search(["quartz", "quartz"], 0).Single().Score == score, "query repetition cannot multiply score");
        Check(singleton.Search(["missing"], 0).IsEmpty && singleton.Search(["quartz"], score + .01).IsEmpty, "missing terms and score floor");
        Check(new RecallIndex([]).Search(["quartz"], 0).IsEmpty
            && new RecallIndex([Snippet("!!!")]).Search(["quartz"], 0).IsEmpty, "empty and zero-word corpora");
        // CodeRhapsody's older clamped Robertson IDF gives zero when df >= N/2.
        // Ensemble ch23 deliberately uses positive log(1+ratio) smoothing. This
        // concrete unequal-length corpus checks DF, TF and average length jointly.
        var arithmetic = new RecallIndex([Snippet("quartz quartz amber"), Snippet("quartz slate"),
            Snippet("slate slate slate slate slate")]).Search(["quartz"], 0);
        Check(arithmetic.Length == 2
            && Math.Abs(arithmetic[0].Score - Math.Log(1.6) * 4.4 / 3.11) < 1e-12
            && Math.Abs(arithmetic[1].Score - Math.Log(1.6) * 2.2 / 1.84) < 1e-12,
            "N=3 DF=2 avgDL=10/3 uses smoothed IDF and exact TF/length normalization");
        var rarity = new RecallIndex([Snippet("common rare"), Snippet("common common"), Snippet("common common")]);
        Check(rarity.Search(["rare"], 0)[0].Score > rarity.Search(["common"], 0)[0].Score, "rare term outweighs ubiquitous repetition");
        var shortSnippet = Snippet("quartz decision");
        var longSnippet = Snippet("quartz " + string.Join(' ', Enumerable.Repeat("filler", 1000)));
        var normalized = new RecallIndex([longSnippet, shortSnippet]).Search(["quartz"], 0);
        Check(normalized[0].Snippet == shortSnippet && normalized[0].Score > normalized[1].Score, "length normalization ranks concise decision first");
        var saturated = new RecallIndex([Snippet(string.Join(' ', Enumerable.Repeat("quartz", 1000)))]).Search(["quartz"], 0).Single();
        Check(saturated.Score > score && saturated.Score < score * 2.2, "TF saturates below k1+1, not linearly");
        var first = Snippet("equal quartz", file: "first.md");
        var second = Snippet("equal quartz", file: "second.md");
        Check(new RecallIndex([first, second]).Search(["quartz"], 0).Select(hit => hit.Snippet).SequenceEqual(new[] { first, second }), "ties preserve original deterministic order");
        var docs = new RecallIndex(Enumerable.Range(0, 1000).Select(number => Snippet("quartz docs " + number, RecallSource.Docs)));
        Check(docs.Search(["quartz"], 0).Length == 1000 && singleton.Search(["quartz"], 0).Single().Score == score,
            "large independent docs corpus cannot change memory statistics");
        Refuses(() => singleton.Search(["quartz"], double.NaN));
    }

    public static void Quotas()
    {
        var hits = Enumerable.Range(0, 20).Select(number => Hit(RecallSource.Docs, 100 - number, "doc " + number))
            .Concat(new[] { Hit(RecallSource.Memory, 2, "memory one"), Hit(RecallSource.Memory, 1, "memory two") }).ToArray();
        var selected = RecallText.Mechanical(hits, 3);
        Check(selected.Length == 3 && selected.Count(hit => hit.Snippet.Source == RecallSource.Memory) == 2,
            "memory share is enforced on final selection despite high-scoring docs flood");
        var sources = new[] { RecallSource.Docs, RecallSource.Handoffs, RecallSource.Learnings, RecallSource.Skills };
        var mixed = sources.SelectMany(source => Enumerable.Range(0, 4).Select(number => Hit(source, 20 - number, source + " " + number)))
            .Concat(Enumerable.Range(0, 5).Select(number => Hit(RecallSource.Memory, 1, "mem " + number))).ToArray();
        var distributed = RecallText.Mechanical(mixed, 9);
        Check(distributed.Count(hit => hit.Snippet.Source == RecallSource.Memory) == 5
            && sources.All(source => distributed.Count(hit => hit.Snippet.Source == source) == 1), "half memory, remaining shares cover secondary corpora");
        Check(RecallText.Mechanical(hits.Where(hit => hit.Snippet.Source == RecallSource.Docs), 3).Length == 3, "unused memory slots return to score pool");
        var sparse = RecallText.Mechanical(hits.Take(1).Concat(hits.Skip(20)), 10);
        Check(sparse.Length == 3 && sparse.Distinct().Count() == 3, "sparse corpora do not duplicate hits");
    }

    public static void Judge()
    {
        var hits = Enumerable.Range(0, 5).Select(number => Hit(RecallSource.Docs, 10 - number, "candidate " + number)).ToArray();
        var chosen = RecallText.JudgeChoices("Here is my choice:\n```json\n[\"3\", 1, 3, -1, 99, null, 2.5, {}, \"bad\", \"2\"]\n```", hits, 3);
        Check(chosen.SequenceEqual(new[] { hits[3], hits[1], hits[2] }), "quoted/mixed choices retain valid ordered unique indices");
        Check(RecallText.JudgeChoices("[]", hits, 3).IsEmpty, "valid empty is not fallback");
        Check(RecallText.JudgeChoices("[null, false, -1, 99, 2147483648]", hits, 3).IsEmpty, "invalid elements never select zero");
        Check(RecallText.JudgeChoices("[4,3,2,1]", hits, 2).SequenceEqual(new[] { hits[4], hits[3] }), "judge order and max selections respected without quotas");
        Check(RecallText.JudgeChoices("[\" 2 \"]", hits, 3).Single() == hits[2], "quoted integer whitespace is tolerated");
        Refuses(() => RecallText.JudgeChoices("not an array", hits, 3));
        Refuses(() => RecallText.JudgeChoices("[1,]", hits, 3));
        Refuses(() => RecallText.JudgeChoices("[1] [2]", hits, 3));
    }

    public static void Formatting()
    {
        var snippet = Snippet("first line\nsecond line\n" + string.Concat(Enumerable.Repeat("é界🙂", 2000)));
        var hit = new RecallHit(snippet, 1);
        var attachment = RecallText.Format([hit], 6144)!;
        var strict = new UTF8Encoding(false, true);
        Check(attachment.Text.StartsWith("[Auto-recalled memories]\n\nFrom memory/decision.md — Decision:\n> first line\n> second line\n> ", StringComparison.Ordinal), "title attribution and blockquote every line");
        Check(strict.GetByteCount(attachment.Text) <= 6144 && strict.GetByteCount(attachment.Text) > 6139, "entire wire text counted in hard byte cap");
        Check(attachment.Hashes.Single() == snippet.Hash(), "partial attachment carries full snippet identity");
        Check(RecallText.Format([hit], 2) is null, "no title-only attachment");
        var prefixBytes = Encoding.UTF8.GetByteCount("[Auto-recalled memories]\n\nFrom memory/decision.md — Decision:\n> ");
        Check(RecallText.Format([new RecallHit(Snippet("🙂"), 1)], prefixBytes + 3) is null,
            "no identity when first source rune does not fit");
        for (var limit = prefixBytes + 1; limit < prefixBytes + 30; limit++)
        {
            var capped = RecallText.Format([new RecallHit(Snippet("é界🙂é界🙂é界🙂"), 1)], limit);
            if (capped is not null) Check(strict.GetByteCount(capped.Text) <= limit, "every multibyte boundary remains valid");
        }
        var small = new RecallHit(Snippet("One\nTwo"), 1);
        var duplicate = RecallText.Format([small, small], 1000)!;
        Check(duplicate.Hashes.Length == 1 && duplicate.Text.LastIndexOf("From", StringComparison.Ordinal)
            == duplicate.Text.IndexOf("From", StringComparison.Ordinal), "same content is represented once");
        Check(RecallText.Unquote("> One\n> Two") == "One\nTwo", "quote normalization supports exact-content deduplication");
    }
}
