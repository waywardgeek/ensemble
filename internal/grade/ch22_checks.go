package grade

import "fmt"

// ch22Table is chapter 22's published contract: seven checks, one hundred
// points. The point values are part of the chapter and must not drift; a
// student's score has to mean the same thing in six months.
var ch22Table = []struct {
	ID     string
	Points int
	Why    string
}{
	{"agent-builds", 5, "The agent builds"},
	{"back-pointer-chain", 20, "Capabilities are reached through a back-pointer chain, not stapled on"},
	{"single-composition-root", 15, "One composition root in the library, and the binary uses it"},
	{"agent-status-tool", 20, "agent_status reports model, usage, cache and cost"},
	{"reaches-through-the-chain", 15, "The tool package reaches the engine only through the hub"},
	{"per-model-cost", 15, "Session cost is the sum of per-model costs, not the total at the current price"},
	{"ch21-parity", 10, "Chapter 21's behaviour still works"},
}

// Ch22Result is what the runner observes. Errs maps a check ID to the reason
// it failed; a check ID present with an empty string passed.
type Ch22Result struct {
	Errs  map[string]string
	Fatal string
}

// ran marks a check as exercised. A check that never ran reports that fact
// rather than silently scoring zero, so a harness failure is distinguishable
// from a student failure. Chapter 21 lost time twice to harness bugs that
// looked exactly like student failures.
func (r *Ch22Result) ran(id string) {
	if r.Errs == nil {
		r.Errs = map[string]string{}
	}
	if _, ok := r.Errs[id]; !ok {
		r.Errs[id] = ""
	}
}

func (r *Ch22Result) fail(id, format string, args ...any) {
	r.ran(id)
	if existing := r.Errs[id]; existing != "" {
		return // keep the first failure: it is usually the cause of the rest
	}
	r.Errs[id] = fmt.Sprintf(format, args...)
}

// Ch22Checks converts observations into the graded result.
func Ch22Checks(r Ch22Result) []Check {
	out := make([]Check, 0, len(ch22Table))
	for _, w := range ch22Table {
		c := Check{ID: w.ID, Title: w.Why, Points: w.Points, Earned: w.Points, Passed: true}
		switch msg, ran := r.Errs[w.ID]; {
		case r.Fatal != "":
			c.failf("%s", r.Fatal)
		case !ran:
			c.failf("not exercised: an earlier failure stopped the run before this check")
		case msg != "":
			c.failf("%s", msg)
		}
		out = append(out, c)
	}
	return out
}
