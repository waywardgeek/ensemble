package grade

// Chapter 17 checks.
//
// Nine of these ten are read off the wire: what the agent sent to the model,
// and what it declined to send. None of them looks inside the student's tree
// for a function name, a file name, or a directory layout, so an implementation
// that is shaped differently from the reference and behaves the same scores the
// same.
//
// The tenth, recall-is-a-spoke, is openly structural, and it has to be. The
// star topology has no runtime symptom — code that violates it works fine until
// a later chapter needs the dependency to point the other way and finds a
// cycle. Chapter 6 grades the topology for the spokes that existed when it was
// written; a spoke added eleven chapters later has to grade itself.

type ch17Check struct {
	ID     string
	Points int
	Why    string
}

var ch17Table = []ch17Check{
	{"bm25-indexes", 10,
		"Archived text the user never mentioned reaches the model, so the archive is actually indexed and searched."},
	{"bm25-scores", 10,
		"Results are ranked by term rarity, term frequency and document length, not merely by containing the word."},
	{"bm25-stop-words", 5,
		"Words like \"the\" and \"we\" are stripped, so a short document full of them cannot outrank a real match."},
	{"chunking-splits", 10,
		"A long document is indexed section by section, so retrieving one part does not cost the context of the whole file."},
	{"per-source-quota", 10,
		"Each source keeps a share of the result, so the source with the most text on a topic cannot take every slot."},
	{"judge-filters", 15,
		"A model reads the candidates and its verdict decides what is attached; changing its answer changes the output."},
	{"judge-fallback", 10,
		"Prose, HTTP 500, a hang, and an empty archive all degrade to top-N BM25 or to silence. Never a crash, never a dead turn."},
	{"recall-is-a-spoke", 10,
		"internal/recall imports the hub and stdlib only, and no existing spoke imports it back."},
	{"recall-is-own-kind", 10,
		"Recalled material lands as its own permanent entry carrying its bytes, and is never merged into the user's message."},
	{"injection-capped", 10,
		"The attached block is bounded, so an archive that matches well everywhere cannot swallow the context window."},
}

// Ch17Checks turns a run into the graded checks.
func Ch17Checks(r Ch17Result) []Check {
	out := make([]Check, 0, len(ch17Table))
	for _, w := range ch17Table {
		c := Check{ID: w.ID, Title: w.Why, Points: w.Points, Earned: w.Points, Passed: true}
		switch msg, ran := r.Errs[w.ID]; {
		case r.Fatal != "":
			c.failf("%s", r.Fatal)
		case !ran:
			c.failf("check did not run")
		case msg != "":
			c.failf("%s", msg)
		}
		out = append(out, c)
	}
	return out
}
