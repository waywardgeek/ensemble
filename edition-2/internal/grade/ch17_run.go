package grade

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Chapter 17 scenarios.
//
// Everything here is observed from outside the agent: what the model was sent,
// and what landed in the event log. Nothing reaches into the student's types.
// The one exception is recall-is-a-spoke, which is explicitly a structural
// check because the property it defends — the star topology — is invisible at
// runtime and only becomes visible as a compile-time import cycle six chapters
// later.

// Marker tokens. Each is a nonsense word so that finding one in a request
// proves it came from the archive rather than from the prompt, the system
// prompt, or the model's own output.
const (
	ch17MarkStrong = "markalpha"       // the best lexical match
	ch17MarkWeak   = "markbeta"        // a genuine but weaker match
	ch17MarkTrap   = "marktrap"        // matches the query on STOP WORDS only
	ch17MarkDocs   = "markdocsource"   // lives outside memory, for quota testing
	ch17MarkSect   = "marksectionfour" // one section of a long document
	ch17MarkOther  = "marksectionone"  // a different section of that document

	// ch17MarkBulk labels a document that mentions the query's rarest term
	// MORE times than the best match does, while being an order of magnitude
	// longer. Raw term frequency ranks it first; BM25, with TF saturation and
	// length normalisation, ranks it below the short dense note. It exists so
	// that the scoring check can assert an order that a naive "count the
	// occurrences" scorer gets BACKWARDS rather than merely imprecise.
	ch17MarkBulk = "markbulkymatch"

	// These two live in sources that no other scenario touches, so the
	// indexing check can assert that every source root is walked without its
	// verdict depending on ranking, quota, judging or capping.
	ch17MarkHandoff  = "markhandoffsrc"
	ch17MarkLearning = "marklearningsrc"
)

// ch17Query is the prompt used by most scenarios.
//
// It is deliberately over 80 characters, because a sane implementation refuses
// to run retrieval on "ok" or "thanks" — a two-word query matches everything
// weakly and wastes a judge call to discover that. Chapter fixtures that used
// a short prompt would silently test the skip path instead of recall.
//
// It also quotes none of the archive verbatim, so nothing is dropped by the
// "already in the conversation" filter.
const ch17Query = "Remind me please, what did we settle on for the zorblax rollout, and why did we choose that approach?"

const ch17GardenFiller = "Tomato seedlings prefer warm soil and steady watering through the summer. " +
	"Prune the lower leaves and stake the vines before the fruit sets. " +
	"Compost improves drainage in heavy clay beds and feeds the roots slowly.\n\n"

// ch17Filler returns n sections of prose that share no content words with
// ch17Query, so it pads the corpus without competing for a slot.
//
// Padding is not cosmetic. BM25 weights a term by how RARE it is, and rarity
// is measured against the corpus. In a three-chunk archive nothing is rare,
// every score collapses toward zero, and a correct implementation looks broken.
func ch17Filler(n int) string {
	var sb strings.Builder
	sb.WriteString("# Garden notes\n\n")
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&sb, "## Garden note %d\n\n%s", i, ch17GardenFiller)
	}
	return sb.String()
}

func ch17Write(dir string, files map[string]string) error {
	for rel, body := range files {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// ch17Archive is the fixture used by the scoring, filtering and own-kind
// scenarios.
func ch17Archive() map[string]string {
	return map[string]string{
		"docs/filler.md": ch17Filler(24),

		// Strong: short, and dense in the query's content words.
		"docs/decision.md": "## Zorblax rollout decision\n\n" +
			ch17MarkStrong + " The zorblax rollout was settled in favour of a staged approach. " +
			"We chose zorblax staging because the zorblax cutover risk was judged too high. " +
			"The zorblax rollout therefore proceeds in phases.\n",

		// Weak: mentions zorblax once, buried in unrelated prose. A correct
		// scorer ranks this below the decision note; a scorer that only
		// checks for the presence of a term cannot tell them apart.
		"docs/background.md": "## Zorblax background reading\n\n" +
			ch17MarkWeak + " A single mention of zorblax appears here. " +
			strings.Repeat(ch17GardenFiller, 4),

		// The stop-word trap. Its ONLY overlap with the query is words like
		// "the", "we", "did" and "for". It is short, which under BM25 length
		// normalisation makes it score very well indeed — if those words are
		// indexed at all. Strip them and this file scores exactly zero.
		"docs/trap.md": "## Trap\n\n" + ch17MarkTrap +
			" the of and to is it that this with from for was are be on at as by an or not but " +
			"they we you what did why we for the and that the of the to the is it was for we did\n",

		// The term-frequency trap. This mentions "zorblax" twelve times —
		// more than docs/decision.md does — but buries them in ten times the
		// prose. Counting occurrences ranks it FIRST. BM25 ranks it below
		// decision.md, because the twelfth mention adds almost nothing once
		// TF saturates and because the score is normalised by length.
		//
		// The order of these two is therefore a genuine discriminator: it is
		// not merely wrong under a bad formula, it is REVERSED.
		"docs/bulk.md": "## Zorblax rollout appendix\n\n" + ch17MarkBulk + " " +
			strings.Repeat("The zorblax rollout is mentioned here again. "+ch17GardenFiller, 12),
	}
}

// ch17IndexArchive puts exactly three matching chunks in three DIFFERENT
// source roots, and nothing else that matches at all.
//
// Three candidates for three slots means every one of them is attached no
// matter how they are ranked, whether the quota runs, or what the judge says.
// That is the point: this fixture isolates "was the archive walked and
// indexed" from every downstream behaviour, so the indexing check can be
// killed by deleting a source root and by nothing else.
func ch17IndexArchive() map[string]string {
	return map[string]string{
		"docs/filler.md": ch17Filler(24),
		"docs/decision.md": "## Zorblax rollout decision\n\n" +
			ch17MarkStrong + " The zorblax rollout was settled in favour of a staged approach.\n",
		"handoffs/handoff-2026-01-01.md": "## Zorblax rollout handoff\n\n" +
			ch17MarkHandoff + " The zorblax rollout was handed over mid-flight.\n",
		"learnings/learnings.md": "## Zorblax rollout learning\n\n" +
			ch17MarkLearning + " The zorblax rollout taught us to stage the cutover.\n",
	}
}

// ch17LongDoc is one document with several distinct sections, for chunking.
func ch17LongDoc() map[string]string {
	return map[string]string{
		"docs/filler.md": ch17Filler(24),
		"docs/manual.md": "# Operations manual\n\n" +
			"## Section one: mulching\n\n" + ch17MarkOther + " " + strings.Repeat(ch17GardenFiller, 3) +
			"## Section two: pruning\n\n" + strings.Repeat(ch17GardenFiller, 3) +
			"## Section three: watering\n\n" + strings.Repeat(ch17GardenFiller, 3) +
			"## Section four: the zorblax rollout\n\n" + ch17MarkSect +
			" The zorblax rollout was settled here in favour of a staged approach. " +
			"We chose zorblax staging because the zorblax cutover risk was too high. " +
			"The zorblax rollout proceeds in phases.\n\n" +
			"## Section five: composting\n\n" + strings.Repeat(ch17GardenFiller, 3),
	}
}

// ch17QuotaArchive floods MEMORY with strong matches and puts a single much
// weaker match in docs.
//
// The crowding has to be real. An earlier version of this fixture used a short,
// dense docs note, which BM25 ranked inside the top three on merit — so the
// note appeared whether or not a quota existed, and deleting the quota
// outright still scored full marks. The check was decoration.
//
// So docs/note.md now mentions zorblax exactly once, diluted through several
// paragraphs of unrelated prose, while every memory log is short and says
// almost nothing else. Ranked on score alone the four surviving logs take all
// three slots and the note is nowhere. It can only appear if something
// reserves a share of the result for each source.
//
// SkipNewestMemories drops the two newest daily logs — they are already
// verbatim in the system prompt — so six files are written to leave four.
func ch17QuotaArchive() map[string]string {
	files := map[string]string{"docs/filler.md": ch17Filler(24)}
	for i := 1; i <= 6; i++ {
		files[fmt.Sprintf("memory/memory/2026-01-%02d-1.md", i)] =
			fmt.Sprintf("## Zorblax rollout log %d\n\nThe zorblax rollout was settled: we chose zorblax staging "+
				"because the zorblax cutover risk was too high. The zorblax rollout proceeds in phases.\n", i)
	}
	files["docs/note.md"] = "## Garden and operations miscellany\n\n" + ch17MarkDocs + " " +
		ch17GardenFiller + "A zorblax note kept outside memory. " +
		strings.Repeat(ch17GardenFiller, 3)
	return files
}

// ch17FloodArchive is a deliberately abusive archive: a great deal of text
// that all matches the query well.
func ch17FloodArchive() map[string]string {
	files := map[string]string{"docs/filler.md": ch17Filler(24)}
	for i := 1; i <= 40; i++ {
		files[fmt.Sprintf("docs/flood-%02d.md", i)] = fmt.Sprintf(
			"## Zorblax rollout volume %d\n\nThe zorblax rollout was settled in favour of a staged approach, "+
				"and we chose zorblax staging for the zorblax rollout. %s\n",
			i, strings.Repeat("The zorblax rollout approach we chose was settled deliberately. ", 120))
	}
	return files
}

// ---------------------------------------------------------------------------
// Judge routes
// ---------------------------------------------------------------------------

func ch17JudgeReply(text string) func([]byte) *fakevendor.Reply {
	return func(body []byte) *fakevendor.Reply {
		if ch17IsJudge(body) {
			return &fakevendor.Reply{Text: text}
		}
		return nil
	}
}

func ch17JudgeError() func([]byte) *fakevendor.Reply {
	return func(body []byte) *fakevendor.Reply {
		if ch17IsJudge(body) {
			return &fakevendor.Reply{Status: 500, ErrBody: `{"error":"judge exploded"}`}
		}
		return nil
	}
}

func ch17JudgeHang(d time.Duration) func([]byte) *fakevendor.Reply {
	return func(body []byte) *fakevendor.Reply {
		if ch17IsJudge(body) {
			time.Sleep(d)
			return &fakevendor.Reply{Text: "[0]"}
		}
		return nil
	}
}

// ---------------------------------------------------------------------------
// Extracting the recalled block from a request
// ---------------------------------------------------------------------------

const ch17RecallHeader = "[Auto-recalled memories]"

// ch17RecallMsg returns the message carrying the recalled block.
func ch17RecallMsg(body []byte) (ch17Msg, bool) {
	for _, m := range ch17Parse(body).Messages {
		if strings.Contains(m.Text(), ch17RecallHeader) {
			return m, true
		}
	}
	return ch17Msg{}, false
}

func ch17RecallText(body []byte) string {
	m, ok := ch17RecallMsg(body)
	if !ok {
		return ""
	}
	return m.Text()
}

// ---------------------------------------------------------------------------
// Result plumbing
// ---------------------------------------------------------------------------

type Ch17Result struct {
	Errs  map[string]string
	Fatal string
}

func (r *Ch17Result) ran(id string)               { r.Errs[id] = "" }
func (r *Ch17Result) fail(id, f string, a ...any) { r.Errs[id] = fmt.Sprintf(f, a...) }

// ensure marks a check failed only if it has not already failed, so the first
// and most specific diagnosis survives.
func (r *Ch17Result) ensure(id string, ok bool, f string, a ...any) {
	if _, seen := r.Errs[id]; !seen {
		r.Errs[id] = ""
	}
	if !ok && r.Errs[id] == "" {
		r.Errs[id] = fmt.Sprintf(f, a...)
	}
}

func Ch17Run(dir string) Ch17Result {
	res := Ch17Result{Errs: map[string]string{}}

	// Structural check first: it needs no binary and no model, and if the
	// package is not there at all the rest of the diagnosis is noise.
	ch17CheckSpoke(dir, &res)

	bin, cleanup, err := Build(dir)
	if err != nil {
		res.Fatal = fmt.Sprintf("build failed: %v", err)
		return res
	}
	defer cleanup()
	skills, err := ch17Skills()
	if err != nil {
		res.Fatal = fmt.Sprintf("skills: %v", err)
		return res
	}
	gui := filepath.Join(dir, "web")

	ch17Scoring(bin, skills, gui, &res)
	ch17Indexing(bin, skills, gui, &res)
	ch17Filtering(bin, skills, gui, &res)
	ch17Chunking(bin, skills, gui, &res)
	ch17Quota(bin, skills, gui, &res)
	ch17Fallback(bin, skills, gui, &res)
	ch17Flood(bin, skills, gui, &res)
	return res
}

func ch17Skills() (string, error) {
	dir, err := os.MkdirTemp("", "ch17-skills")
	if err != nil {
		return "", err
	}
	base := filepath.Join(dir, "base")
	if err := os.MkdirAll(base, 0o755); err != nil {
		return "", err
	}
	body := "---\nname: base\ndescription: base skill\n---\n\nYou are a helpful assistant.\n"
	if err := os.WriteFile(filepath.Join(base, "SKILL.md"), []byte(body), 0o644); err != nil {
		return "", err
	}
	return dir, nil
}

// ---------------------------------------------------------------------------
// Scenario: scoring
// ---------------------------------------------------------------------------

// ch17Scoring drives the mechanical path by giving the judge a reply it cannot
// possibly parse, so selection falls back to pure BM25 order.
//
// Using the fallback here is deliberate: it makes the assertion about RANKING
// rather than about the judge's taste. With a judge in the loop the order of
// the attached snippets is the judge's order, and a scoring bug would be
// invisible behind it.
func ch17Scoring(bin, skills, gui string, res *Ch17Result) {
	ids := []string{"bm25-scores", "bm25-stop-words", "recall-is-own-kind"}
	dir, cleanDir := freshRunDir("ch17-scoring")
	defer cleanDir()
	if err := ch17Write(dir, ch17Archive()); err != nil {
		for _, id := range ids {
			res.fail(id, "fixture: %v", err)
		}
		return
	}
	out := ch17Launch(bin, skills, gui, ch17Opts{
		dir:     dir,
		model:   "claude-sonnet-5-course",
		prompts: []string{ch17Query},
		turns:   []int{1},
		replies: []fakevendor.Reply{{Text: "Understood."}},
		route:   ch17JudgeReply("I am afraid I cannot determine relevance from this material."),
	})
	if out.fatal != "" {
		for _, id := range ids {
			res.fail(id, "%s", out.fatal)
		}
		return
	}
	turn, ok := ch17LastTurn(out.reqs)
	if !ok {
		for _, id := range ids {
			res.fail(id, "no turn request reached the vendor")
		}
		return
	}
	text := ch17RecallText(turn.Body)

	// bm25-indexes is NOT checked here. It has its own scenario, because the
	// question "was the archive indexed at all" must not be answerable only
	// through the ranked, quota'd, judged, capped output of this one.

	// bm25-scores: the dense short match must outrank both the incidental
	// mention and the long document that says "zorblax" more often than it
	// does; and generic filler must not appear at all.
	res.ran("bm25-scores")
	switch {
	case text == "":
		res.fail("bm25-scores", "no recalled block reached the model")
	case strings.Contains(text, "Garden note"):
		res.fail("bm25-scores", "recall surfaced filler that shares no content words with the query,\n"+
			"which means results are not being ranked by relevance.\nrecalled block was:\n%s", ch17Trunc(text))
	case !strings.Contains(text, ch17MarkStrong):
		res.fail("bm25-scores", "the strongest match (docs/decision.md) was not recalled at all")
	case strings.Contains(text, ch17MarkWeak) &&
		strings.Index(text, ch17MarkWeak) < strings.Index(text, ch17MarkStrong):
		res.fail("bm25-scores", "docs/background.md, which mentions zorblax once in four paragraphs of unrelated prose,\n"+
			"was ranked above docs/decision.md, which is about nothing else.\n"+
			"Term frequency and document length are not affecting the score.\nrecalled block was:\n%s", ch17Trunc(text))
	case strings.Contains(text, ch17MarkBulk) &&
		strings.Index(text, ch17MarkBulk) < strings.Index(text, ch17MarkStrong):
		res.fail("bm25-scores", "docs/bulk.md was ranked above docs/decision.md.\n"+
			"bulk.md contains the word \"zorblax\" MORE times than decision.md does, but it is ten\n"+
			"times longer and says little else about it. Ranking it first is what counting raw\n"+
			"occurrences produces: the score is not saturating repeated terms, and it is not being\n"+
			"normalised by document length, so a long document can outrank a precise one simply by\n"+
			"repeating the query.\nrecalled block was:\n%s", ch17Trunc(text))
	}

	// bm25-stop-words
	res.ran("bm25-stop-words")
	if text != "" && strings.Contains(text, ch17MarkTrap) {
		res.fail("bm25-stop-words", "docs/trap.md was recalled. Its only overlap with the query is words like\n"+
			"\"the\", \"we\", \"did\" and \"for\". It is short, so BM25 length normalisation scores it highly\n"+
			"once those words are indexed. Stop words are not being removed.\nrecalled block was:\n%s", ch17Trunc(text))
	}

	// recall-is-own-kind: permanent, separate, and carrying its bytes.
	res.ran("recall-is-own-kind")
	ch17CheckOwnKind(out, turn, res)
}

func ch17CheckOwnKind(out ch17Out, turn fakevendor.Recorded, res *Ch17Result) {
	const id = "recall-is-own-kind"
	m, ok := ch17RecallMsg(turn.Body)
	if !ok {
		res.fail(id, "no recalled block reached the model, so it cannot be checked for placement")
		return
	}
	// The property that matters: recalled material is never fused into the
	// human's own message. If it were, the model could not tell the archive's
	// guesses from the user's instructions, and any text in the archive would
	// be able to impersonate the user.
	if strings.Contains(m.Text(), ch17Query) {
		res.fail(id, "the recalled block and the user's own words arrived in the SAME message.\n"+
			"Recalled text must never be merged into the user's turn: a model cannot obey\n"+
			"instructions from the user if anything in the archive can forge them.")
		return
	}
	if m.Role == "user" {
		// Allowed on vendors with no system role, but not on Anthropic.
		res.fail(id, "the recalled block arrived with role %q on a vendor that supports a\n"+
			"mid-conversation system message; it should not be presented as the user speaking", m.Role)
		return
	}
	// Permanent, and replayable without a model call: the event carries the
	// rendered bytes rather than the query that produced them.
	var found bool
	for _, e := range out.log {
		if e.Recall == nil {
			continue
		}
		found = true
		if len(e.Recall.Parts) == 0 || string(e.Recall.Parts) == "null" {
			res.fail(id, "the recall event was logged with no parts. Replay would have to re-run\n"+
				"retrieval against today's archive to reconstruct it, so history would change\n"+
				"whenever a memory file changed.")
			return
		}
	}
	if !found {
		res.fail(id, "no recall event appears in the event log. Recalled material must land in the\n"+
			"dialogue as its own permanent entry, not as turn-scoped scaffolding that replay cannot see.")
	}
}

// ---------------------------------------------------------------------------
// Scenario: the archive is indexed, across every source
// ---------------------------------------------------------------------------

// ch17Indexing asks the narrowest question in the chapter: was the archive
// walked and indexed at all, and was EVERY source root walked?
//
// It deliberately owns its own fixture, in which exactly three chunks match
// the query and three slots are available. Every matching chunk is therefore
// attached regardless of how it was scored, whether a per-source quota ran,
// what the judge decided, or where the byte cap fell. Strip any one of those
// behaviours out and this check still passes; fail to index a source root and
// it is the only check that fails.
//
// That independence is the whole point. The previous version of this check
// read the same recalled block as the scoring check, so the only mutant that
// killed it was one that killed seven other checks at the same time — which
// tells you nothing about what the check actually defends.
func ch17Indexing(bin, skills, gui string, res *Ch17Result) {
	const id = "bm25-indexes"
	res.ran(id)
	dir, cleanDir := freshRunDir("ch17-indexing")
	defer cleanDir()
	if err := ch17Write(dir, ch17IndexArchive()); err != nil {
		res.fail(id, "fixture: %v", err)
		return
	}
	out := ch17Launch(bin, skills, gui, ch17Opts{
		dir:     dir,
		model:   "claude-sonnet-5-course",
		prompts: []string{ch17Query},
		turns:   []int{1},
		replies: []fakevendor.Reply{{Text: "Understood."}},
		route:   ch17JudgeReply("I am afraid I cannot determine relevance from this material."),
	})
	if out.fatal != "" {
		res.fail(id, "%s", out.fatal)
		return
	}
	turn, ok := ch17LastTurn(out.reqs)
	if !ok {
		res.fail(id, "no turn request reached the vendor")
		return
	}
	text := ch17RecallText(turn.Body)
	if text == "" {
		res.fail(id, "no recalled block reached the model: the archive was never indexed or never searched\n"+
			"(the request carried no message containing %q)", ch17RecallHeader)
		return
	}
	for _, want := range []struct{ mark, where string }{
		{ch17MarkStrong, "docs/decision.md"},
		{ch17MarkHandoff, "handoffs/handoff-2026-01-01.md"},
		{ch17MarkLearning, "learnings/learnings.md"},
	} {
		if !strings.Contains(text, want.mark) {
			res.fail(id, "%s matches the query as well as anything in the archive, and there were exactly\n"+
				"three matching chunks competing for three slots, so nothing crowded it out — yet it was\n"+
				"not recalled. That source root is not being indexed.\nrecalled block was:\n%s",
				want.where, ch17Trunc(text))
			return
		}
	}
}

// ---------------------------------------------------------------------------
// Scenario: the judge actually filters
// ---------------------------------------------------------------------------

// ch17Filtering runs the SAME fixture and query as ch17Scoring, changing only
// the judge's answer.
//
// That is what makes it a real test. Scoring showed the weak match being
// attached; here the judge keeps only the first candidate, so the weak match
// must disappear. An implementation that runs the judge and ignores its verdict
// produces identical output in both scenarios and fails exactly here.
func ch17Filtering(bin, skills, gui string, res *Ch17Result) {
	const id = "judge-filters"
	res.ran(id)
	dir, cleanDir := freshRunDir("ch17-filtering")
	defer cleanDir()
	if err := ch17Write(dir, ch17Archive()); err != nil {
		res.fail(id, "fixture: %v", err)
		return
	}
	out := ch17Launch(bin, skills, gui, ch17Opts{
		dir:     dir,
		model:   "claude-sonnet-5-course",
		prompts: []string{ch17Query},
		turns:   []int{1},
		replies: []fakevendor.Reply{{Text: "Understood."}},
		route:   ch17JudgeReply("[0]"),
	})
	if out.fatal != "" {
		res.fail(id, "%s", out.fatal)
		return
	}
	if len(ch17Judges(out.reqs)) == 0 {
		res.fail(id, "no judge call was made. Every request the vendor saw carried tools, so nothing\n"+
			"asked a model which snippets were worth keeping; BM25 output went straight to the turn.")
		return
	}
	turn, ok := ch17LastTurn(out.reqs)
	if !ok {
		res.fail(id, "no turn request reached the vendor")
		return
	}
	text := ch17RecallText(turn.Body)
	if text == "" {
		res.fail(id, "the judge selected candidate 1 but nothing was attached to the turn")
		return
	}
	if strings.Contains(text, ch17MarkWeak) {
		res.fail(id, "the judge was asked for candidate 1 only, and docs/background.md was attached anyway.\n"+
			"The judge's verdict is being discarded: the same snippets are attached whatever it says.\nrecalled block was:\n%s", ch17Trunc(text))
	}
}

// ---------------------------------------------------------------------------
// Scenario: chunking
// ---------------------------------------------------------------------------

// ch17Chunking checks that a long document is split, and that the relevant
// SECTION is what gets recalled rather than the whole file.
func ch17Chunking(bin, skills, gui string, res *Ch17Result) {
	const id = "chunking-splits"
	res.ran(id)
	dir, cleanDir := freshRunDir("ch17-chunking")
	defer cleanDir()
	if err := ch17Write(dir, ch17LongDoc()); err != nil {
		res.fail(id, "fixture: %v", err)
		return
	}
	out := ch17Launch(bin, skills, gui, ch17Opts{
		dir:     dir,
		model:   "claude-sonnet-5-course",
		prompts: []string{ch17Query},
		turns:   []int{1},
		replies: []fakevendor.Reply{{Text: "Understood."}},
		route:   ch17JudgeReply("I am afraid I cannot determine relevance from this material."),
	})
	if out.fatal != "" {
		res.fail(id, "%s", out.fatal)
		return
	}
	turn, ok := ch17LastTurn(out.reqs)
	if !ok {
		res.fail(id, "no turn request reached the vendor")
		return
	}
	text := ch17RecallText(turn.Body)
	switch {
	case text == "":
		res.fail(id, "nothing was recalled from a document containing an exactly relevant section")
	case !strings.Contains(text, ch17MarkSect):
		res.fail(id, "the relevant section of docs/manual.md was not recalled.\nrecalled block was:\n%s", ch17Trunc(text))
	case strings.Contains(text, ch17MarkOther):
		res.fail(id, "recalling the zorblax section of docs/manual.md also dragged in its mulching section.\n"+
			"The file is being indexed as one unit, so retrieving any part of a document costs the\n"+
			"context budget of the whole document.\nrecalled block was:\n%s", ch17Trunc(text))
	}
}

// ---------------------------------------------------------------------------
// Scenario: per-source quota
// ---------------------------------------------------------------------------

func ch17Quota(bin, skills, gui string, res *Ch17Result) {
	const id = "per-source-quota"
	res.ran(id)
	dir, cleanDir := freshRunDir("ch17-quota")
	defer cleanDir()
	if err := ch17Write(dir, ch17QuotaArchive()); err != nil {
		res.fail(id, "fixture: %v", err)
		return
	}
	out := ch17Launch(bin, skills, gui, ch17Opts{
		dir:     dir,
		model:   "claude-sonnet-5-course",
		prompts: []string{ch17Query},
		turns:   []int{1},
		replies: []fakevendor.Reply{{Text: "Understood."}},
		route:   ch17JudgeReply("I am afraid I cannot determine relevance from this material."),
	})
	if out.fatal != "" {
		res.fail(id, "%s", out.fatal)
		return
	}
	turn, ok := ch17LastTurn(out.reqs)
	if !ok {
		res.fail(id, "no turn request reached the vendor")
		return
	}
	text := ch17RecallText(turn.Body)
	switch {
	case text == "":
		res.fail(id, "nothing was recalled from an archive full of strong matches")
	case !strings.Contains(text, ch17MarkDocs):
		res.fail(id, "four daily logs took every available slot and docs/note.md got none.\n"+
			"Ranked on score alone the source holding the most text about a topic sweeps the board,\n"+
			"so a document is never quoted once memory has enough to say about it. Reserve a share\n"+
			"of the result for each source present.\nrecalled block was:\n%s", ch17Trunc(text))
	}
}

// ---------------------------------------------------------------------------
// Scenario: fallback
// ---------------------------------------------------------------------------

// ch17Fallback checks that every way the judge can let us down degrades to
// top-N BM25, and that none of them takes the turn down with it.
//
// The judge is the only part of recall that depends on a remote service
// answering sensibly, and it runs on EVERY turn. If its failure modes are not
// all handled, recall converts an intermittent vendor problem into an agent
// that cannot hold a conversation.
func ch17Fallback(bin, skills, gui string, res *Ch17Result) {
	const id = "judge-fallback"
	res.ran(id)

	cases := []struct {
		name  string
		route func([]byte) *fakevendor.Reply
		why   string
	}{
		{"garbage", ch17JudgeReply("Sorry, I can't help with that request."),
			"the judge answered in prose instead of with numbers"},
		{"error", ch17JudgeError(),
			"the judge call returned HTTP 500"},
		{"hang", ch17JudgeHang(20 * time.Second),
			"the judge stopped responding and had to be abandoned"},
	}
	for _, c := range cases {
		dir, cleanDir := freshRunDir("ch17-fallback-" + c.name)
		defer cleanDir()
		if err := ch17Write(dir, ch17Archive()); err != nil {
			res.fail(id, "fixture: %v", err)
			return
		}
		out := ch17Launch(bin, skills, gui, ch17Opts{
			dir:     dir,
			model:   "claude-sonnet-5-course",
			prompts: []string{ch17Query},
			turns:   []int{1},
			replies: []fakevendor.Reply{{Text: "Understood."}},
			route:   c.route,
		})
		if out.fatal != "" {
			res.fail(id, "when %s, the turn did not complete: %s", c.why, out.fatal)
			return
		}
		turn, ok := ch17LastTurn(out.reqs)
		if !ok {
			res.fail(id, "when %s, no turn request reached the vendor at all.\n"+
				"A judge that misbehaves must not be able to stop the agent from talking.", c.why)
			return
		}
		if text := ch17RecallText(turn.Body); !strings.Contains(text, ch17MarkStrong) {
			res.fail(id, "when %s, recall gave up entirely instead of falling back to the top BM25 results.\n"+
				"The lexical ranking is still perfectly good on its own; the judge only trims it.\nrecalled block was:\n%s",
				c.why, ch17Trunc(text))
			return
		}
	}

	// A brand-new agent with nothing archived. Recall must be inert rather
	// than attaching an empty block or failing the turn.
	dir, cleanDir := freshRunDir("ch17-fallback-empty")
	defer cleanDir()
	out := ch17Launch(bin, skills, gui, ch17Opts{
		dir:     dir,
		model:   "claude-sonnet-5-course",
		prompts: []string{ch17Query},
		turns:   []int{1},
		replies: []fakevendor.Reply{{Text: "Understood."}},
		route:   ch17JudgeReply("[0]"),
	})
	if out.fatal != "" {
		res.fail(id, "with an empty archive the turn did not complete: %s", out.fatal)
		return
	}
	turn, ok := ch17LastTurn(out.reqs)
	if !ok {
		res.fail(id, "with an empty archive no turn request reached the vendor.\n"+
			"An agent on its first run has nothing to recall, and that is the normal case, not an error.")
		return
	}
	if strings.Contains(ch17AllText(turn.Body), ch17RecallHeader) {
		res.fail(id, "with an empty archive an empty recall block was still attached to the turn.\n"+
			"Finding nothing is not the same as having something to say.")
	}
}

// ---------------------------------------------------------------------------
// Scenario: the byte cap
// ---------------------------------------------------------------------------

// ch17Flood points recall at an archive where a great deal of text matches the
// query well, and checks that the amount attached stays bounded.
//
// Without a cap the failure is quiet and expensive: every turn silently grows
// by however much the archive happened to match, the context fills with
// retrieved text, and compaction starts discarding the actual conversation to
// make room for guesses about it.
// The bound must sit BELOW what an uncapped implementation would attach, or
// the check cannot fail no matter how abusive the archive is. Chunks are at
// most ~4KB and three snippets are selected, so an implementation with no cap
// at all attaches roughly 12KB; anything at or above that is unfalsifiable.
// 8KB is comfortably above a sane cap and comfortably below the uncapped size.
const ch17MaxRecallBytes = 8192

func ch17Flood(bin, skills, gui string, res *Ch17Result) {
	const id = "injection-capped"
	res.ran(id)
	dir, cleanDir := freshRunDir("ch17-flood")
	defer cleanDir()
	if err := ch17Write(dir, ch17FloodArchive()); err != nil {
		res.fail(id, "fixture: %v", err)
		return
	}
	out := ch17Launch(bin, skills, gui, ch17Opts{
		dir:     dir,
		model:   "claude-sonnet-5-course",
		prompts: []string{ch17Query},
		turns:   []int{1},
		replies: []fakevendor.Reply{{Text: "Understood."}},
		route:   ch17JudgeReply("[0,1,2]"),
	})
	if out.fatal != "" {
		res.fail(id, "%s", out.fatal)
		return
	}
	turn, ok := ch17LastTurn(out.reqs)
	if !ok {
		res.fail(id, "no turn request reached the vendor")
		return
	}
	text := ch17RecallText(turn.Body)
	if text == "" {
		res.fail(id, "nothing was recalled from an archive of strong matches")
		return
	}
	if len(text) > ch17MaxRecallBytes {
		res.fail(id, "%d bytes of recalled text were attached to a single turn (limit %d).\n"+
			"The archive grows without bound, so anything proportional to how much of it matched\n"+
			"is unbounded too. Cap the block and drop whatever does not fit.", len(text), ch17MaxRecallBytes)
	}
}

// ---------------------------------------------------------------------------
// Structural: recall is a spoke
// ---------------------------------------------------------------------------

// ch17CheckSpoke is the one deliberately structural check in this chapter.
//
// Every other check here asks what the agent DID. This one asks what the code
// is allowed to import, because the property it protects has no runtime
// symptom: a retrieval package that reaches into the model package works
// perfectly, right up until some later chapter needs the model package to know
// about retrieval and discovers the cycle. Chapter 6 grades the star topology
// for the spokes that existed when it was written, and it cannot know about a
// spoke added eleven chapters later. So this chapter grades its own.
func ch17CheckSpoke(dir string, res *Ch17Result) {
	const id = "recall-is-a-spoke"
	res.ran(id)

	const hub = "internal/common"
	pkgDir := filepath.Join(dir, "internal", "recall")
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		res.fail(id, "no package at agent/internal/recall: %v\n"+
			"Retrieval belongs in its own spoke, not folded into the package that talks to models.", err)
		return
	}
	fset := token.NewFileSet()
	var files int
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		files++
		f, err := parser.ParseFile(fset, filepath.Join(pkgDir, name), nil, parser.ImportsOnly)
		if err != nil {
			res.fail(id, "parsing %s: %v", name, err)
			return
		}
		for _, im := range f.Imports {
			path, err := strconv.Unquote(im.Path.Value)
			if err != nil {
				continue
			}
			if !strings.Contains(path, "/ensemble/") {
				continue // stdlib or third party
			}
			if strings.HasSuffix(path, hub) {
				continue
			}
			res.fail(id, "agent/internal/recall/%s imports %q.\n"+
				"A spoke may import the hub and nothing else. Spokes that import each other turn the\n"+
				"star into a graph, and the cost is not paid here: it is paid by whichever later\n"+
				"chapter first needs the dependency to run the other way.", name, path)
			return
		}
	}
	if files == 0 {
		res.fail(id, "agent/internal/recall contains no Go files")
		return
	}

	// And the other direction: no existing spoke may depend on the new one.
	for _, spoke := range []string{"llm", "tools", "jobs"} {
		bad, where := ch17ImportsRecall(filepath.Join(dir, "internal", spoke))
		if bad {
			res.fail(id, "agent/internal/%s imports the recall spoke (%s).\n"+
				"The engine must reach retrieval through an interface declared in the hub, so that it\n"+
				"depends on the idea of recall rather than on one implementation of it.", spoke, where)
			return
		}
	}
}

func ch17ImportsRecall(pkgDir string) (bool, string) {
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return false, ""
	}
	fset := token.NewFileSet()
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(pkgDir, name), nil, parser.ImportsOnly)
		if err != nil {
			continue
		}
		for _, im := range f.Imports {
			path, _ := strconv.Unquote(im.Path.Value)
			if strings.HasSuffix(path, "internal/recall") {
				return true, name
			}
		}
	}
	return false, ""
}

func ch17Trunc(s string) string {
	const max = 1200
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n... (truncated)"
}
