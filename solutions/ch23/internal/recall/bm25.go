// Package recall is passive memory retrieval: a BM25 keyword search over the
// agent's memory files, narrowed by a small model, attached to the
// conversation without anyone asking for it.
//
// It is a spoke. It imports the hub and the standard library and nothing
// else — in particular it does not import internal/llm, even though the
// judge is a model call, because the judge reaches it as a common.SnippetJudge
// handed in at construction. The inversion is the point: recall knows there is
// something that can rank snippets, and knows nothing about HTTP, vendors, or
// streaming.
package recall

import (
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// BM25 tuning. Robertson and Walker, 1994, and the standard values have not
// needed changing in thirty years.
const (
	// k1 controls term-frequency saturation: the tenth occurrence of a word
	// says much less than the second.
	k1 = 1.2
	// b controls length normalisation: how hard a long document is penalised
	// for having more surface area to match against.
	b = 0.75
	// minTokenLen drops one-character tokens, which are never discriminating
	// and are mostly list markers and stray letters from code.
	minTokenLen = 2
)

// stopWords is a const string rather than a package-level map on purpose.
//
// Idiomatic Go would write `var stopWords = map[string]bool{...}` here, and
// that would be a mutable global: any code in the process could add a word to
// it, and every index in the process would silently start filtering
// differently. The map is built per-Index in newIndex instead, so an index
// owns its own vocabulary and two indexes cannot interfere.
const stopWords = `a about above after again against all am an and any are aren't as at
be because been before being below between both but by
can cannot could couldn't
did didn't do does doesn't doing don't down during
each few for from further
had hadn't has hasn't have haven't having he her here hers herself him himself his how
i if in into is isn't it its itself
just
let's
me more most must my myself
no nor not now
of off on once only or other ought our ours ourselves out over own
same shan't she should shouldn't so some such
than that the their theirs them themselves then there these they this those through to too
under until up
very
was wasn't we were weren't what when where which while who whom why with won't would wouldn't
you your yours yourself yourselves
also get got how i'd i'll i'm i've like make made may might need needs one see something
thing things use used using want way well what's will
please thanks thank ok okay yes yeah sure`

// Chunk is one searchable piece of a file, with the statistics BM25 needs.
//
// The attribution fields — which file, which section — live on the hub, in
// common.Chunk, because they are what crosses a package boundary when a
// snippet is quoted back to the agent. Tokens, Length and TF are scoring
// state and never leave this package, so they are added here by embedding
// rather than by widening the hub type. The hub carries what two packages
// must agree on; an engine's working set is not that.
type Chunk struct {
	common.Chunk
	Tokens []string       // pre-computed tokenisation
	Length int            // token count, including stop words
	TF     map[string]int // term frequencies
}

// Index is a BM25 index over a set of chunks.
//
// One index per source. Merging every source into a single index is what the
// first version of this did, and design documents buried personal memories:
// BM25 rewards term density, and a 20KB design doc has far more surface area
// to match against than a 500-byte daily log that happens to contain the
// decision. Keeping the indexes separate is what makes a per-source quota
// expressible at all.
type Index struct {
	Chunks    []Chunk
	AvgLength float64        // average chunk length in tokens
	DocFreq   map[string]int // how many chunks contain each term
	Total     int            // total number of chunks

	// stop is this index's own stop-word set, built in newIndex from the
	// const above. A field, not a global: see the comment on stopWords.
	stop map[string]bool

	// totalLength is the running sum of chunk lengths, kept so AvgLength is
	// a division rather than a walk of every chunk on every Add.
	totalLength int
}

// newIndex returns an empty index with its stop-word set compiled.
func newIndex() *Index {
	ix := &Index{DocFreq: make(map[string]int)}
	ix.stop = make(map[string]bool)
	for _, w := range strings.Fields(stopWords) {
		ix.stop[w] = true
	}
	return ix
}

// NewIndex returns an empty index. Exported for tests and for callers that
// want to build an index from chunks they produced themselves.
func NewIndex() *Index { return newIndex() }

// Tokenize splits text into scoring terms: break on every non-alphanumeric
// rune, lowercase, drop anything shorter than two characters.
//
// Stop words are NOT removed here. They are removed from queries, in
// queryTerms, and deliberately kept in documents: a chunk's length is part of
// the BM25 normalisation, and a chunk that is half stop words really is
// longer — and therefore weaker evidence per match — than one that is not.
// Stripping them from documents would quietly inflate the score of prose
// relative to code.
func (ix *Index) Tokenize(text string) []string {
	fields := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if len(f) >= minTokenLen {
			out = append(out, f)
		}
	}
	return out
}

// queryTerms tokenizes a query and drops stop words.
//
// Duplicates are kept. A user who says "cache" twice means it more than a
// user who says it once, and BM25 sums over query terms, so the repetition
// weights the term without any special case.
func (ix *Index) queryTerms(query string) []string {
	toks := ix.Tokenize(query)
	out := make([]string, 0, len(toks))
	for _, t := range toks {
		if !ix.stop[t] {
			out = append(out, t)
		}
	}
	return out
}

// Add indexes one chunk.
//
// DocFreq counts CHUNKS CONTAINING a term, not occurrences of it — the
// distinction is the whole meaning of "document frequency", and counting
// occurrences instead is the single easiest way to get a ranking function
// that is subtly wrong and still returns plausible-looking results. That is
// exactly the failure mode worth fearing here: nothing crashes, nothing logs,
// the top hit is merely the wrong one, for weeks. Hence the seen set.
func (ix *Index) Add(c common.Chunk) {
	toks := ix.Tokenize(c.Content)
	tf := make(map[string]int, len(toks))
	for _, t := range toks {
		tf[t]++
	}
	for t := range tf {
		ix.DocFreq[t]++
	}
	ix.Chunks = append(ix.Chunks, Chunk{
		Chunk:  c,
		Tokens: toks,
		Length: len(toks),
		TF:     tf,
	})
	ix.Total = len(ix.Chunks)

	ix.totalLength += len(toks)
	if ix.Total > 0 {
		ix.AvgLength = float64(ix.totalLength) / float64(ix.Total)
	}
}

// idf is inverse document frequency, and it cannot go negative.
//
// The floor matters, and HOW it is achieved matters more. The textbook
// Robertson form, ln((N-df+0.5)/(df+0.5)), goes negative as soon as a term
// appears in more than half the corpus — a chunk would be PENALISED for
// containing a query word — so it is usually clamped at zero.
//
// Clamping is wrong here, and the reason only shows up on a small corpus.
// Clamped, the formula returns exactly zero for every term in half or more
// of the chunks, and on a corpus of three files that is nearly every term.
// A brand-new agent with a handful of memories would score everything at
// zero and recall nothing, forever, which is precisely when it most needs
// help. The failure is silent: no error, no empty index, just a ranking
// function that has quietly switched itself off.
//
// So this uses the smoothed form Lucene and Elasticsearch use,
// ln(1 + (N-df+0.5)/(df+0.5)), which is positive for every df and decays
// toward zero for ubiquitous terms instead of falling off a cliff at the
// halfway mark. Common words still stop discriminating; they just stop
// gradually. The explicit clamp below is kept as a guard, not as the
// mechanism — it is now unreachable, and that is the point.
func (ix *Index) idf(term string) float64 {
	df := ix.DocFreq[term]
	if df == 0 {
		return 0
	}
	v := math.Log(1 + (float64(ix.Total)-float64(df)+0.5)/(float64(df)+0.5))
	if v < 0 {
		return 0
	}
	return v
}

// Score scores one chunk against already-tokenized query terms.
func (ix *Index) Score(terms []string, i int) float64 {
	if i < 0 || i >= len(ix.Chunks) || ix.AvgLength == 0 {
		return 0
	}
	c := ix.Chunks[i]
	var score float64
	for _, t := range terms {
		f := c.TF[t]
		if f == 0 {
			continue
		}
		idf := ix.idf(t)
		if idf == 0 {
			continue
		}
		norm := float64(f) + k1*(1-b+b*float64(c.Length)/ix.AvgLength)
		score += idf * float64(f) * (k1 + 1) / norm
	}
	return score
}

// Hit is one scored chunk, carrying which source index produced it.
type Hit struct {
	Chunk  common.Chunk
	Score  float64
	Source string
}

// Search returns every chunk scoring above min, best first.
//
// It does not truncate. Quota allocation happens a level up, in the Recaller,
// which needs to know how many results each source COULD have supplied in
// order to reflow unused slots — a truncated list would hide that.
func (ix *Index) Search(query string, min float64) []Hit {
	terms := ix.queryTerms(query)
	if len(terms) == 0 {
		return nil
	}
	var hits []Hit
	for i := range ix.Chunks {
		s := ix.Score(terms, i)
		if s <= min {
			continue
		}
		hits = append(hits, Hit{Chunk: ix.Chunks[i].Chunk, Score: s})
	}
	sort.SliceStable(hits, func(a, c int) bool { return hits[a].Score > hits[c].Score })
	return hits
}
