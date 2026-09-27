package recall

// The pipeline: search, judge, attach.
//
// Three stages, and each one degrades into the next rather than failing. If
// the query is too short, nothing happens. If BM25 finds nothing, nothing
// happens. If the judge breaks, times out, or was never wired, the BM25
// results go through unfiltered. If recall is switched off entirely, the
// agent works and simply does not remember passively. Every failure mode
// produces a working agent with less context, never a broken one — which is
// the only defensible shape for a feature nobody asked for by name.

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// recallHeader labels the block so the model can tell retrieved text from
// anything a human typed. It is the first thing a reader of the log sees too.
const recallHeader = "[Auto-recalled memories]"

// Config is the tuning. These are not guesses; each one moved over months of
// daily use, and §17.6 tells the story of every number.
type Config struct {
	// MinScore is the BM25 floor when there is no judge. High, because with
	// nothing downstream to filter false positives the floor IS the filter.
	MinScore float64
	// JudgeMinScore is the floor when a judge is present. Low, deliberately:
	// cast a wide net and let the judge do the discriminating, since it can
	// see context that a score cannot.
	JudgeMinScore float64
	// MaxSnippets caps how many snippets are attached.
	MaxSnippets int
	// MinQueryLength skips recall for short messages. "fix it" and "run
	// tests" contain no searchable concept, and matching on them produces
	// confident noise.
	MinQueryLength int
	// Candidates is how many BM25 hits are offered to the judge.
	Candidates int
	// ContextMsgs is how many recent messages the judge is shown.
	ContextMsgs int
	// MaxRecallBytes is the hard ceiling on attached text.
	MaxRecallBytes int
	// MemoryQuota is memory's guaranteed share of candidate slots.
	MemoryQuota float64
	// JudgeTimeout bounds the judge call.
	//
	// Enforced HERE rather than inside whatever implements SnippetJudge, so
	// that a slow judge cannot wedge a turn no matter who supplies it. A
	// deadline belongs to the caller that has something else to be doing.
	JudgeTimeout time.Duration
}

// DefaultConfig returns the production values.
//
// A function, not a package-level struct: a var would be a mutable global
// that any caller could retune for every agent in the process.
func DefaultConfig() Config {
	return Config{
		MinScore:       3.0,
		JudgeMinScore:  0.5,
		MaxSnippets:    3,
		MinQueryLength: 80,
		Candidates:     20,
		ContextMsgs:    3,
		MaxRecallBytes: 6144,
		MemoryQuota:    0.5,
		JudgeTimeout:   15 * time.Second,
	}
}

// Recaller implements common.Recaller.
type Recaller struct {
	sources []*source
	judge   common.SnippetJudge
	cfg     Config
}

// New builds a Recaller over the given sources.
//
// judge may be nil, which means no judge: recall still works, filtered by
// BM25 score alone. Nothing here dials anything or starts a goroutine —
// indexing is the only work, and it is a few hundred files of markdown.
func New(specs []SourceSpec, judge common.SnippetJudge, cfg Config) *Recaller {
	r := &Recaller{judge: judge, cfg: cfg}
	for _, spec := range specs {
		s := loadSource(spec)
		if s.chunkCount() == 0 {
			continue
		}
		r.sources = append(r.sources, s)
	}
	return r
}

// Indexed reports the total number of chunks across all sources. Zero means
// there is nothing to recall and the caller may skip recall entirely.
func (r *Recaller) Indexed() int {
	n := 0
	for _, s := range r.sources {
		n += s.chunkCount()
	}
	return n
}

// Recall returns the snippets that belong with a user message.
//
// convo is the conversation so far as plain text, oldest first. The tail of
// it goes to the judge as context; all of it is used to avoid attaching
// something the agent is already looking at.
func (r *Recaller) Recall(query string, convo []string) common.PartList {
	if len(strings.TrimSpace(query)) < r.cfg.MinQueryLength {
		return nil
	}
	cands := r.candidates(query)
	if len(cands) == 0 {
		return nil
	}
	picked := r.pick(query, convo, cands)
	text := r.format(picked, convo)
	if text == "" {
		return nil
	}
	return common.PartList{common.TextPart{Text: text}}
}

// ----------------------------------------------------------------
// Stage 1: BM25 with per-source quotas
// ----------------------------------------------------------------

// candidates runs one BM25 search per source and allocates slots by quota.
func (r *Recaller) candidates(query string) []Hit {
	min := r.cfg.MinScore
	if r.judge != nil {
		min = r.cfg.JudgeMinScore
	}

	perSource := make(map[string][]Hit, len(r.sources))
	avail := make(map[string]int, len(r.sources))
	names := make([]string, 0, len(r.sources))
	for _, s := range r.sources {
		hits := s.index.Search(query, min)
		for i := range hits {
			hits[i].Source = s.name
		}
		perSource[s.name] = hits
		avail[s.name] = len(hits)
		names = append(names, s.name)
	}

	quota := r.allocate(r.cfg.Candidates, names, avail)

	var out []Hit
	for _, n := range names {
		hits := perSource[n]
		if q := quota[n]; q < len(hits) {
			hits = hits[:q]
		}
		out = append(out, hits...)
	}
	// Present best-first across sources. The quota decided WHICH hits get a
	// seat; score decides the order they are read in, which matters because
	// the no-judge and fallback paths both take the front of this list.
	sortHits(out)
	if len(out) > r.cfg.Candidates {
		out = out[:r.cfg.Candidates]
	}
	return out
}

// allocate hands out candidate slots.
//
// Memory takes half, rounded up, and the rest is split evenly. Then unused
// quota flows to sources that have more results than their share — otherwise
// an agent with no handoffs and no docs would send the judge six candidates
// when twenty were available.
//
// This is a policy decision, not a retrieval improvement. Merged BM25 already
// ranks correctly by its own lights; the problem is that its lights are term
// density, and a design document has more of that than a daily log which
// happens to contain the actual decision. No amount of tuning fixes a
// mismatch between what the formula rewards and what the reader values, so
// the fix is structural: guarantee the seats.
func (r *Recaller) allocate(total int, names []string, avail map[string]int) map[string]int {
	quota := make(map[string]int, len(names))
	if len(names) == 0 || total <= 0 {
		return quota
	}

	var others []string
	hasMemory := false
	for _, n := range names {
		if n == SourceMemory {
			hasMemory = true
			continue
		}
		others = append(others, n)
	}

	rest := total
	if hasMemory {
		quota[SourceMemory] = int(math.Ceil(float64(total) * r.cfg.MemoryQuota))
		rest = total - quota[SourceMemory]
	}
	if len(others) > 0 && rest > 0 {
		base, rem := rest/len(others), rest%len(others)
		for i, n := range others {
			quota[n] = base
			if i < rem {
				quota[n]++
			}
		}
	}

	// Reflow. Cap each source at what it actually has, pool the surplus, and
	// hand it round-robin to sources that still have results left over.
	for pass := 0; pass <= len(names); pass++ {
		spare := 0
		var hungry []string
		for _, n := range names {
			switch {
			case quota[n] > avail[n]:
				spare += quota[n] - avail[n]
				quota[n] = avail[n]
			case avail[n] > quota[n]:
				hungry = append(hungry, n)
			}
		}
		if spare == 0 || len(hungry) == 0 {
			break
		}
		moved := false
		for spare > 0 {
			progressed := false
			for _, n := range hungry {
				if spare == 0 {
					break
				}
				if avail[n] > quota[n] {
					quota[n]++
					spare--
					progressed = true
					moved = true
				}
			}
			if !progressed {
				break
			}
		}
		if !moved {
			break
		}
	}
	return quota
}

// sortHits orders by score, best first, with a stable tie-break on
// attribution so a run is reproducible.
func sortHits(h []Hit) {
	for i := 1; i < len(h); i++ {
		for j := i; j > 0; j-- {
			a, bb := h[j-1], h[j]
			if a.Score > bb.Score {
				break
			}
			if a.Score == bb.Score && attribution(a) <= attribution(bb) {
				break
			}
			h[j-1], h[j] = h[j], h[j-1]
		}
	}
}

// ----------------------------------------------------------------
// Stage 2: the judge
// ----------------------------------------------------------------

// pick narrows candidates to the snippets worth attaching.
//
// Three ways this ends up as plain top-N BM25, and all three are normal
// operation rather than error handling: no judge was wired, the judge call
// failed or timed out, or the judge answered with something unparseable. A
// noisier recall beats no recall, and either beats a turn that does not run.
func (r *Recaller) pick(query string, convo []string, cands []Hit) []Hit {
	if r.judge == nil {
		return r.topNWithQuota(cands, r.cfg.MaxSnippets)
	}
	reply, err := r.askJudge(r.judgePrompt(query, convo, cands))
	if err != nil {
		return r.topNWithQuota(cands, r.cfg.MaxSnippets)
	}
	idx, ok := parseIndices(reply, len(cands))
	if !ok {
		return r.topNWithQuota(cands, r.cfg.MaxSnippets)
	}
	// An empty array is a judgment, not a failure: the judge looked and
	// found nothing worth surfacing. Obeying it is the point of having one.
	out := make([]Hit, 0, len(idx))
	for _, i := range idx {
		if len(out) >= r.cfg.MaxSnippets {
			break
		}
		out = append(out, cands[i])
	}
	return out
}

// askJudge calls the judge under a deadline.
//
// The channel is buffered so that a judge which answers after the deadline
// writes into the buffer and exits, rather than blocking forever on a
// receiver that has already given up. A timeout that leaks a goroutine per
// occurrence is a slow crash.
func (r *Recaller) askJudge(prompt string) (string, error) {
	type result struct {
		text string
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		text, err := r.judge.Pick(prompt)
		ch <- result{text, err}
	}()

	timeout := r.cfg.JudgeTimeout
	if timeout <= 0 {
		timeout = DefaultConfig().JudgeTimeout
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case res := <-ch:
		return res.text, res.err
	case <-timer.C:
		return "", fmt.Errorf("recall judge timed out after %s", timeout)
	}
}

// topN takes the first n hits, which are already the best-scoring ones.
// topNWithQuota selects up to n hits, best first, but refuses to let a single
// source take every slot.
//
// This runs on the MECHANICAL paths only — no judge configured, or a judge
// that errored, timed out, or answered nonsense. On those paths the ranking is
// pure lexical overlap, and pure lexical overlap has a pathology: whichever
// source happens to hold the most text about a topic sweeps the board. An
// agent with nine months of daily logs and one design document will surface
// daily logs forever, and the document it should have quoted never gets a
// seat, because it was outnumbered rather than outranked.
//
// It deliberately does NOT run after a judge has spoken. If a judge actually
// read the candidates and concluded that the three relevant ones all came from
// memory, then they did, and forcing in a document to satisfy a ratio would be
// overruling the only component that assessed relevance rather than counted
// words. Diversity is a tie-breaker for a mechanism that cannot tell, not a
// correction to one that can.
//
// Note the second pass. A quota may leave seats empty when a source has fewer
// hits than its share — and an empty seat helps nobody, so anything unfilled
// is handed back to the best remaining hit regardless of source. The quota
// caps the greedy case; it never shrinks the result.
func (r *Recaller) topNWithQuota(hits []Hit, n int) []Hit {
	if n <= 0 || len(hits) == 0 {
		return nil
	}
	avail := map[string]int{}
	var order []string
	for _, h := range hits {
		if _, seen := avail[h.Source]; !seen {
			order = append(order, h.Source)
		}
		avail[h.Source]++
	}
	if len(order) < 2 {
		// A single source is not crowding anybody out.
		return topN(hits, n)
	}
	quota := r.allocate(n, order, avail)

	out := make([]Hit, 0, n)
	used := map[string]int{}
	taken := make([]bool, len(hits))
	for i, h := range hits {
		if len(out) >= n {
			break
		}
		if used[h.Source] >= quota[h.Source] {
			continue
		}
		used[h.Source]++
		taken[i] = true
		out = append(out, h)
	}
	for i, h := range hits {
		if len(out) >= n {
			break
		}
		if !taken[i] {
			out = append(out, h)
		}
	}
	return out
}

func topN(h []Hit, n int) []Hit {
	if n < 0 {
		n = 0
	}
	if len(h) > n {
		return h[:n]
	}
	return h
}

// ----------------------------------------------------------------
// Stage 3: formatting and the byte cap
// ----------------------------------------------------------------

// attribution is the source line above a quoted snippet.
func attribution(h Hit) string {
	if strings.TrimSpace(h.Chunk.Header) == "" {
		return fmt.Sprintf("From %s:", h.Chunk.Filename)
	}
	return fmt.Sprintf("From %s — %s:", h.Chunk.Filename, h.Chunk.Header)
}

// format renders the chosen snippets as blockquoted text under a header.
//
// Two things happen here that are easy to mistake for polish and are not.
//
// Dedupe: because these blocks persist, the same memory recalled on turns 3,
// 7 and 12 would appear three times. Retrieval should surface what the agent
// does not already have in front of it, so a snippet already present in the
// conversation is skipped. This is what makes keeping a block strictly better
// than clearing it — the cost of retention is bounded by the corpus, not by
// the number of turns.
//
// The cap: MaxRecallBytes is a hard ceiling on the whole block, enforced
// after assembly as well as during it. A ceiling that only holds when the
// arithmetic upstream was right is not a ceiling.
func (r *Recaller) format(hits []Hit, convo []string) string {
	if len(hits) == 0 {
		return ""
	}
	seen := strings.Join(convo, "\n")

	var b strings.Builder
	b.WriteString(recallHeader)
	b.WriteString("\n")
	wrote := 0
	for _, h := range hits {
		if alreadyPresent(seen, h) {
			continue
		}
		piece := "\n" + attribution(h) + "\n" + blockquote(h.Chunk.Content) + "\n"
		if b.Len()+len(piece) > r.cfg.MaxRecallBytes {
			room := r.cfg.MaxRecallBytes - b.Len()
			if room <= len(recallHeader) {
				break
			}
			b.WriteString(truncateAt(piece, room))
			wrote++
			break
		}
		b.WriteString(piece)
		wrote++
	}
	if wrote == 0 {
		return ""
	}
	return capBytes(b.String(), r.cfg.MaxRecallBytes)
}

// alreadyPresent reports whether this snippet is already in the conversation,
// either because it was recalled before (same attribution line) or because
// its text is literally there.
func alreadyPresent(convo string, h Hit) bool {
	if convo == "" {
		return false
	}
	if strings.Contains(convo, attribution(h)) {
		return true
	}
	body := strings.TrimSpace(h.Chunk.Content)
	if body == "" {
		return true
	}
	probe := body
	if len(probe) > 120 {
		probe = truncateAt(probe, 120)
	}
	return strings.Contains(convo, probe)
}

// blockquote prefixes every line with "> ".
func blockquote(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i, ln := range lines {
		lines[i] = "> " + ln
	}
	return strings.Join(lines, "\n")
}

// truncateAt cuts to at most n bytes without splitting a rune.
func truncateAt(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if len(s) <= n {
		return s
	}
	cut := n
	for cut > 0 && !utf8Start(s[cut]) {
		cut--
	}
	return s[:cut]
}

// capBytes is the last line of defence on the ceiling: cut at a line
// boundary if one is near, otherwise cut hard.
func capBytes(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	s = truncateAt(s, max)
	if i := strings.LastIndex(s, "\n"); i > max/2 {
		return s[:i]
	}
	return s
}
