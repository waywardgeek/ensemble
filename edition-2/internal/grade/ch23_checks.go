package grade

import "fmt"

// ch23Table is chapter 23's published contract: seven checks, one hundred
// points. The chapter prints this table, so the IDs and the point values are
// part of the text and must not drift.
//
// Every check here is behavioural. That is deliberate and it is the whole
// lesson of the chapter: a sandbox is not a property of the source, it is a
// property of what the running program can reach. A structural check that
// grepped for a call to sandbox.Resolve would pass a student who called it and
// ignored the answer, which is precisely the security theatre section 23.8
// warns about.
var ch23Table = []struct {
	ID     string
	Points int
	Why    string
}{
	{"path-confinement", 15, "File tools refuse to escape the sandbox root"},
	{"kernel-confinement", 20, "Commands and their descendants cannot reach outside the root"},
	{"no-network", 15, "A sandboxed command cannot reach the network"},
	{"no-credentials", 15, "No credential reaches a tool result"},
	{"child-cannot-widen", 10, "A child's permissions are clamped, and an over-request is reported"},
	{"safe-mode-absence", 15, "Safe mode removes the exec tools and keeps the file tools"},
	{"no-host-path-leak", 10, "A refusal names the path the model asked for, not the host path"},
}

// Ch23Result is what the runner observes. Errs maps a check ID to the reason it
// failed; a check ID present with an empty string passed.
type Ch23Result struct {
	Errs  map[string]string
	Fatal string
}

// ran marks a check as exercised. A check that never ran says so rather than
// silently scoring zero, so a harness bug is distinguishable from a student
// failure. Chapters 21 and 22 each lost time to harness faults that looked
// exactly like student faults.
func (r *Ch23Result) ran(id string) {
	if r.Errs == nil {
		r.Errs = map[string]string{}
	}
	if _, ok := r.Errs[id]; !ok {
		r.Errs[id] = ""
	}
}

func (r *Ch23Result) fail(id, format string, args ...any) {
	r.ran(id)
	if existing := r.Errs[id]; existing != "" {
		return // keep the first failure: it is usually the cause of the rest
	}
	r.Errs[id] = fmt.Sprintf(format, args...)
}

// Ch23Checks converts observations into the graded result.
func Ch23Checks(r Ch23Result) []Check {
	out := make([]Check, 0, len(ch23Table))
	for _, w := range ch23Table {
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
