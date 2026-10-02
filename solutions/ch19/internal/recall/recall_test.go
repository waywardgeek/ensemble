package recall

// Scoring is tested against numbers computed by hand, on paper, from the
// published BM25 formula — not against whatever this implementation happened
// to produce on the day it was written.
//
// That distinction is the whole point of this file. A ranking function that
// is slightly wrong still returns plausible results: the top hit is merely
// the wrong one, nothing crashes, nothing logs, and a golden test that
// records the current output will happily lock the bug in and call it
// regression coverage. The reference implementation this chapter is drawn
// from carried an off-by-one for weeks exactly that way.

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// tinyIndex is the fixture every hand computation below refers to.
//
//	A: "the cat sat on the mat"  -> 6 tokens, TF{the:2, cat:1, sat:1, on:1, mat:1}
//	B: "the dog sat"             -> 3 tokens
//	C: "birds fly"               -> 2 tokens
//
// N = 3, total length 11, AvgLength = 11/3.
func tinyIndex() *Index {
	ix := newIndex()
	ix.Add(common.Chunk{Filename: "a.md", Header: "A", Content: "the cat sat on the mat"})
	ix.Add(common.Chunk{Filename: "b.md", Header: "B", Content: "the dog sat"})
	ix.Add(common.Chunk{Filename: "c.md", Header: "C", Content: "birds fly"})
	return ix
}

func TestIndexStatistics(t *testing.T) {
	ix := tinyIndex()
	if ix.Total != 3 {
		t.Fatalf("Total = %d, want 3", ix.Total)
	}
	if want := 11.0 / 3.0; math.Abs(ix.AvgLength-want) > 1e-9 {
		t.Errorf("AvgLength = %v, want %v", ix.AvgLength, want)
	}
	// Stop words are kept in documents. "the" and "on" are both stop words;
	// if they were stripped at index time this would be 3, and every length
	// normalisation in the corpus would be quietly wrong.
	if got := ix.Chunks[0].Length; got != 6 {
		t.Errorf("chunk A length = %d, want 6 (stop words retained in documents)", got)
	}
	// DocFreq counts chunks containing a term, not occurrences of it. "the"
	// occurs three times across two chunks.
	if got := ix.DocFreq["the"]; got != 2 {
		t.Errorf("DocFreq[the] = %d, want 2 (chunks containing, not occurrences)", got)
	}
	if got := ix.DocFreq["sat"]; got != 2 {
		t.Errorf("DocFreq[sat] = %d, want 2", got)
	}
}

// TestScoreAgainstHandComputation is the check that matters.
//
//	idf(cat) = ln(1 + (3 - 1 + 0.5) / (1 + 0.5)) = ln(1 + 5/3) = ln(8/3)
//	         = 0.9808292530
//	norm     = 1 + 1.2 * (1 - 0.75 + 0.75 * 6 / (11/3))
//	         = 1 + 1.2 * (0.25 + 1.2272727273)
//	         = 2.7727272727
//	score    = 0.9808292530 * 1 * 2.2 / 2.7727272727 = 0.7782317
func TestScoreAgainstHandComputation(t *testing.T) {
	ix := tinyIndex()
	got := ix.Score([]string{"cat"}, 0)
	const want = 0.7782317
	if math.Abs(got-want) > 1e-6 {
		t.Errorf("Score(cat, A) = %.7f, want %.7f", got, want)
	}
	// "cat" appears nowhere else.
	if s := ix.Score([]string{"cat"}, 1); s != 0 {
		t.Errorf("Score(cat, B) = %v, want 0", s)
	}
}

// TestIDFIsPositiveAndDecaying pins the property the smoothed form buys:
// a term in every chunk still scores above zero, so a tiny corpus can still
// rank, while a rare term is worth much more than a common one.
func TestIDFIsPositiveAndDecaying(t *testing.T) {
	ix := tinyIndex()
	rare := ix.idf("cat") // df 1 of 3
	ubiquitous := ix.idf("the")
	middling := ix.idf("sat") // df 2 of 3
	if rare <= middling {
		t.Errorf("idf(cat)=%v should exceed idf(sat)=%v", rare, middling)
	}
	for name, v := range map[string]float64{"cat": rare, "the": ubiquitous, "sat": middling} {
		if v < 0 {
			t.Errorf("idf(%s) = %v, must never be negative", name, v)
		}
	}
	// The case clamping gets wrong: a term in EVERY chunk must still be
	// scoreable, or a three-file corpus ranks nothing at all.
	ix2 := newIndex()
	for _, c := range []string{"alpha beta", "alpha gamma", "alpha delta"} {
		ix2.Add(common.Chunk{Filename: "f.md", Content: c})
	}
	if v := ix2.idf("alpha"); v <= 0 {
		t.Errorf("idf of a term in every chunk = %v, want > 0", v)
	}
	if s := ix2.Score([]string{"alpha"}, 0); s <= 0 {
		t.Errorf("score for a term in every chunk = %v, want > 0", s)
	}
}

// TestStopWordsFilteredFromQueriesOnly is the asymmetry in one test.
func TestStopWordsFilteredFromQueriesOnly(t *testing.T) {
	ix := tinyIndex()
	if got := ix.queryTerms("the cat"); len(got) != 1 || got[0] != "cat" {
		t.Errorf("queryTerms(the cat) = %v, want [cat]", got)
	}
	// A query made only of stop words retrieves nothing rather than
	// everything.
	if hits := ix.Search("the on of and", 0); len(hits) != 0 {
		t.Errorf("all-stop-word query returned %d hits, want 0", len(hits))
	}
}

func TestTokenizeDropsShortTokens(t *testing.T) {
	ix := newIndex()
	got := ix.Tokenize("Go 1.22, a b cd -- E_f")
	joined := strings.Join(got, ",")
	if strings.Contains(joined, ",a,") || strings.HasPrefix(joined, "a,") {
		t.Errorf("Tokenize kept a one-character token: %v", got)
	}
	for _, tok := range got {
		if len(tok) < minTokenLen {
			t.Errorf("Tokenize returned %q, shorter than %d", tok, minTokenLen)
		}
		if tok != strings.ToLower(tok) {
			t.Errorf("Tokenize returned non-lowercase %q", tok)
		}
	}
}

func TestSearchRanksRelevantHigher(t *testing.T) {
	ix := newIndex()
	ix.Add(common.Chunk{Filename: "x.md", Header: "Cache", Content: "prompt cache prefix invalidation cache cache"})
	ix.Add(common.Chunk{Filename: "y.md", Header: "Dinner", Content: "tomatoes basil olive oil pasta"})
	hits := ix.Search("prompt cache invalidation", 0)
	if len(hits) == 0 {
		t.Fatal("no hits")
	}
	if hits[0].Chunk.Filename != "x.md" {
		t.Errorf("top hit = %s, want x.md", hits[0].Chunk.Filename)
	}
	for _, h := range hits {
		if h.Chunk.Filename == "y.md" && h.Score > 0 {
			t.Errorf("irrelevant chunk scored %v", h.Score)
		}
	}
}

// ----------------------------------------------------------------
// Chunking
// ----------------------------------------------------------------

func TestChunkMarkdownSplitsOnHeaders(t *testing.T) {
	doc := "preamble text here\n\n## First\nalpha\n\n## Second\nbeta\n"
	got := ChunkMarkdown("f.md", doc)
	if len(got) != 3 {
		t.Fatalf("got %d chunks, want 3: %+v", len(got), got)
	}
	if got[0].Header != "" || !strings.Contains(got[0].Content, "preamble") {
		t.Errorf("preamble chunk = %+v", got[0])
	}
	if got[1].Header != "First" || got[1].Content != "alpha" {
		t.Errorf("chunk 1 = %+v", got[1])
	}
	if got[2].Header != "Second" || got[2].Content != "beta" {
		t.Errorf("chunk 2 = %+v", got[2])
	}
}

func TestChunkMarkdownBreadcrumbsLargeSections(t *testing.T) {
	big := strings.Repeat("padding words to make this section large. ", 200)
	doc := "## Parent\n### Sub A\n" + big + "\n### Sub B\nshort\n"
	got := ChunkMarkdown("f.md", doc)
	found := false
	for _, c := range got {
		if c.Header == "Parent > Sub A" {
			found = true
		}
		if len(c.Content) > maxChunkBytes {
			t.Errorf("chunk %q is %d bytes, over the %d limit", c.Header, len(c.Content), maxChunkBytes)
		}
	}
	if !found {
		t.Errorf("no breadcrumb header 'Parent > Sub A' in %d chunks", len(got))
	}
}

func TestChunkMarkdownSplitsParagraphsWhenNoSubHeaders(t *testing.T) {
	para := strings.Repeat("sentence about caches. ", 120) // ~2.8KB
	doc := "## Only\n" + para + "\n\n" + para + "\n\n" + para
	got := ChunkMarkdown("f.md", doc)
	if len(got) < 2 {
		t.Fatalf("got %d chunks, want the oversized section split", len(got))
	}
	for _, c := range got {
		if len(c.Content) > maxChunkBytes {
			t.Errorf("chunk is %d bytes, over %d", len(c.Content), maxChunkBytes)
		}
	}
}

// ----------------------------------------------------------------
// Quotas
// ----------------------------------------------------------------

func TestAllocateGivesMemoryHalf(t *testing.T) {
	r := &Recaller{cfg: DefaultConfig()}
	names := []string{SourceMemory, SourceDocs, SourceHandoffs, SourceLearnings, SourceSkills}
	avail := map[string]int{
		SourceMemory: 50, SourceDocs: 50, SourceHandoffs: 50,
		SourceLearnings: 50, SourceSkills: 50,
	}
	q := r.allocate(20, names, avail)
	if q[SourceMemory] < 10 {
		t.Errorf("memory quota = %d, want at least 10 of 20", q[SourceMemory])
	}
	total := 0
	for _, n := range names {
		total += q[n]
	}
	if total != 20 {
		t.Errorf("quotas total %d, want 20", total)
	}
}

func TestAllocateReflowsUnusedSlots(t *testing.T) {
	r := &Recaller{cfg: DefaultConfig()}
	names := []string{SourceMemory, SourceDocs}
	// Memory has only 3 results; docs has plenty. Memory's unused 7 slots
	// must flow to docs rather than evaporating.
	avail := map[string]int{SourceMemory: 3, SourceDocs: 100}
	q := r.allocate(20, names, avail)
	if q[SourceMemory] != 3 {
		t.Errorf("memory quota = %d, want 3 (all it has)", q[SourceMemory])
	}
	if q[SourceDocs] != 17 {
		t.Errorf("docs quota = %d, want 17 (10 + memory's unused 7)", q[SourceDocs])
	}
}

// ----------------------------------------------------------------
// Judge reply parsing
// ----------------------------------------------------------------

func TestParseIndices(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []int
		ok   bool
	}{
		{"plain", "[0, 3, 7]", []int{0, 3, 7}, true},
		{"quoted", `["0", "3"]`, []int{0, 3}, true},
		{"mixed", `[0, "3", 7]`, []int{0, 3, 7}, true},
		{"wrapped in prose", "Sure! Here you go: [1, 2]. Hope that helps.", []int{1, 2}, true},
		{"empty is a judgment", "[]", []int{}, true},
		{"out of range filtered", "[0, 99, 2]", []int{0, 2}, true},
		{"duplicates filtered", "[1, 1, 2]", []int{1, 2}, true},
		{"no brackets is a failure", "I could not decide.", nil, false},
		{"garbage inside brackets", "[not json at all", nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := parseIndices(c.in, 10)
			if ok != c.ok {
				t.Fatalf("ok = %v, want %v (got %v)", ok, c.ok, got)
			}
			if len(got) != len(c.want) {
				t.Fatalf("got %v, want %v", got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Fatalf("got %v, want %v", got, c.want)
				}
			}
		})
	}
}

// ----------------------------------------------------------------
// Pipeline
// ----------------------------------------------------------------

// fakeJudge is a SnippetJudge that answers from a script.
type fakeJudge struct {
	reply  string
	err    error
	delay  time.Duration
	prompt string
	calls  int
}

func (f *fakeJudge) Pick(prompt string) (string, error) {
	f.calls++
	f.prompt = prompt
	if f.delay > 0 {
		time.Sleep(f.delay)
	}
	return f.reply, f.err
}

// pipelineRecaller builds a Recaller over a corpus big enough to be honest.
//
// The filler is not padding for its own sake. MinScore is 3.0, a production
// value tuned against a corpus of a few hundred chunks, and BM25 scores are
// not normalised: on a three-chunk fixture every IDF is small and nothing
// clears the floor, so a test built on three chunks would "prove" that
// recall never fires. Twenty chunks of unrelated prose put the inverse
// document frequencies in the range the threshold was chosen for.
func pipelineRecaller(t *testing.T, judge common.SnippetJudge) *Recaller {
	t.Helper()
	cfg := DefaultConfig()
	cfg.JudgeTimeout = 200 * time.Millisecond
	r := &Recaller{judge: judge, cfg: cfg}
	ix := newIndex()
	for _, c := range []common.Chunk{
		{Filename: "m1.md", Header: "Cache", Content: "prompt cache prefix invalidation is a prefix property"},
		{Filename: "m2.md", Header: "Cache again", Content: "the prefix cache invalidation cost dominates"},
		{Filename: "m3.md", Header: "Pantry", Content: "cache of tomatoes in the pantry prefix invalidation"},
	} {
		ix.Add(c)
	}
	for _, filler := range fillerChunks(20) {
		ix.Add(filler)
	}
	r.sources = []*source{{name: SourceMemory, index: ix}}
	return r
}

// fillerChunks is unrelated prose, to give the corpus a realistic size.
func fillerChunks(n int) []common.Chunk {
	words := []string{
		"harbour", "lantern", "meadow", "quarry", "thistle", "orchard",
		"granite", "willow", "compass", "ferry", "bramble", "kiln",
	}
	out := make([]common.Chunk, 0, n)
	for i := 0; i < n; i++ {
		var b strings.Builder
		for j := 0; j < 8; j++ {
			b.WriteString(words[(i+j)%len(words)])
			b.WriteByte(' ')
		}
		out = append(out, common.Chunk{
			Filename: "filler.md",
			Header:   "Filler",
			Content:  strings.TrimSpace(b.String()),
		})
	}
	return out
}

const longQuery = "I want to understand prompt cache prefix invalidation and what it costs us when we delete bytes"

func TestRecallSkipsShortQueries(t *testing.T) {
	r := pipelineRecaller(t, nil)
	if parts := r.Recall("fix it", nil); parts != nil {
		t.Errorf("short query produced recall: %v", parts)
	}
}

func TestRecallWithoutJudgeReturnsTopN(t *testing.T) {
	r := pipelineRecaller(t, nil)
	parts := r.Recall(longQuery, nil)
	if len(parts) == 0 {
		t.Fatal("no recall")
	}
	text := parts[0].(common.TextPart).Text
	if !strings.Contains(text, recallHeader) {
		t.Errorf("missing header: %q", text)
	}
}

func TestJudgeFiltersCandidates(t *testing.T) {
	j := &fakeJudge{reply: "[0]"}
	r := pipelineRecaller(t, j)
	parts := r.Recall(longQuery, nil)
	if len(parts) == 0 {
		t.Fatal("no recall")
	}
	text := parts[0].(common.TextPart).Text
	if strings.Count(text, "From ") != 1 {
		t.Errorf("judge picked 1 but %d snippets attached:\n%s", strings.Count(text, "From "), text)
	}
	if j.calls != 1 {
		t.Errorf("judge called %d times, want 1", j.calls)
	}
	if !strings.Contains(j.prompt, "CANDIDATE SNIPPETS:") {
		t.Errorf("judge prompt missing candidates section")
	}
}

func TestJudgeEmptyAnswerIsObeyed(t *testing.T) {
	r := pipelineRecaller(t, &fakeJudge{reply: "[]"})
	if parts := r.Recall(longQuery, nil); len(parts) != 0 {
		t.Errorf("empty judgment ignored: %v", parts)
	}
}

func TestJudgeFallbacks(t *testing.T) {
	cases := map[string]common.SnippetJudge{
		"garbage": &fakeJudge{reply: "I am not going to answer that."},
		"error":   &fakeJudge{err: errTest{}},
		"timeout": &fakeJudge{reply: "[0]", delay: 2 * time.Second},
		"nil":     nil,
	}
	for name, j := range cases {
		t.Run(name, func(t *testing.T) {
			r := pipelineRecaller(t, j)
			parts := r.Recall(longQuery, nil)
			if len(parts) == 0 {
				t.Fatal("fallback produced no recall; must degrade to top-N BM25")
			}
			text := parts[0].(common.TextPart).Text
			if n := strings.Count(text, "From "); n == 0 || n > DefaultConfig().MaxSnippets {
				t.Errorf("fallback attached %d snippets, want 1..%d", n, DefaultConfig().MaxSnippets)
			}
		})
	}
}

type errTest struct{}

func (errTest) Error() string { return "judge exploded" }

func TestRecallDedupesAgainstConversation(t *testing.T) {
	r := pipelineRecaller(t, nil)
	first := r.Recall(longQuery, nil)
	if len(first) == 0 {
		t.Fatal("no recall")
	}
	text := first[0].(common.TextPart).Text
	// Feed the previous recall back as conversation. Everything it
	// contained must now be skipped.
	second := r.Recall(longQuery, []string{text})
	if len(second) > 0 {
		got := second[0].(common.TextPart).Text
		for _, line := range strings.Split(text, "\n") {
			if strings.HasPrefix(line, "From ") && strings.Contains(got, line) {
				t.Errorf("re-attached a snippet already in the conversation: %q", line)
			}
		}
	}
}

func TestRecallHonoursByteCap(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxRecallBytes = 300
	// The floor is not what is under test here, and 3.0 is tuned for a
	// corpus far larger than three chunks. Lower it so the cap is what
	// decides the outcome.
	cfg.MinScore = 0.1
	r := &Recaller{cfg: cfg}
	ix := newIndex()
	body := strings.Repeat("prompt cache prefix invalidation bytes. ", 60)
	for _, n := range []string{"a.md", "b.md", "c.md"} {
		ix.Add(common.Chunk{Filename: n, Header: "Big", Content: body})
	}
	r.sources = []*source{{name: SourceMemory, index: ix}}
	parts := r.Recall(longQuery, nil)
	if len(parts) == 0 {
		t.Fatal("no recall")
	}
	text := parts[0].(common.TextPart).Text
	if len(text) > cfg.MaxRecallBytes {
		t.Errorf("recall is %d bytes, over the %d cap", len(text), cfg.MaxRecallBytes)
	}
}
