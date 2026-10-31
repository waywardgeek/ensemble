package grade

import "fmt"

// ch27Table is the chapter 27 grading contract, in the order the TL;DR
// states it. The point values are published, so they are pinned here rather
// than derived from anything that could drift.
var ch27Table = []struct {
	ID     string
	Why    string
	Points int
}{
	{"C1", "Noise-KK handshake: two peers mutually authenticate, wrong keys are rejected", 15},
	{"C2", "Prompt envelope carries provenance, trust, conversation and response mode", 10},
	{"C3", "Gateway server accepts Noise-KK connections from multiple agents", 15},
	{"C4", "Group-based routing delivers only from accepted groups and rejects the rest", 15},
	{"C5", "Delivery semantics: ack, queue while offline, resume from last acked", 15},
	{"C6", "Client library exposes Client and Handler and is usable outside the module", 10},
	{"C7", "Parent registers a child and cannot grant groups it does not hold", 20},
}

// Ch27Result accumulates observations made while driving the student's
// gateway. Every check starts un-run; a check that is never exercised fails,
// so an aborted run can never read as a pass.
type Ch27Result struct {
	Errs map[string]string
	// Fatal records a failure that prevented the run entirely, such as the
	// gateway binary not building.
	Fatal string
}

// NewCh27Result returns an empty result.
func NewCh27Result() Ch27Result {
	return Ch27Result{Errs: map[string]string{}}
}

// ran marks a check as exercised without recording a failure.
func (r *Ch27Result) ran(id string) {
	if _, ok := r.Errs[id]; !ok {
		r.Errs[id] = ""
	}
}

// pass marks a check exercised and passing.
func (r *Ch27Result) pass(id string) { r.ran(id) }

// fail records the first failure for a check. The first is kept because it
// is usually the cause, and later ones are usually its consequences.
func (r *Ch27Result) fail(id, format string, args ...any) {
	r.ran(id)
	if existing := r.Errs[id]; existing != "" {
		return
	}
	r.Errs[id] = fmt.Sprintf(format, args...)
}

// require records a failure when cond is false, and returns cond so callers
// can stop early.
func (r *Ch27Result) require(id string, cond bool, format string, args ...any) bool {
	if !cond {
		r.fail(id, format, args...)
		return false
	}
	r.ran(id)
	return true
}

func (r *Ch27Result) failAll(format string, args ...any) {
	for _, w := range ch27Table {
		r.fail(w.ID, format, args...)
	}
}

// Ch27Checks converts observations into the graded result.
func Ch27Checks(r Ch27Result) []Check {
	out := make([]Check, 0, len(ch27Table))
	for _, w := range ch27Table {
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
