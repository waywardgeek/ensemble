package grade

import "fmt"

// ch24Table is chapter 24's published contract: seven checks, one hundred
// points. The chapter's thesis is that the framework must not know the GUI
// exists, and check 1 is that thesis made mechanical. Every absence check
// here is paired with a positive control inside the harness, so a detector
// that goes blind fails the run instead of passing it.
var ch24Table = []struct {
	ID     string
	Points int
	Why    string
}{
	{"headless-linkage", 20, "A root-only consumer links no GUI package and no websocket library"},
	{"public-reuse", 15, "An application outside the module serves the GUI with fake hooks, and a planted event replays to a browser"},
	{"wire-compat", 15, "The browser wire protocol is unchanged: a ch22-style scripted session still works"},
	{"assets-embedded", 15, "The binary serves the GUI from an unrelated working directory"},
	{"components-without-shell", 10, "A consumer mounts the Artifact scroll from the exported FS without adopting index.html"},
	{"mcp-without-gorilla", 15, "MCP is in the headless closure without gorilla; virtual-user still dials through mcpws"},
	{"root-clean", 10, "The framework root exports no websocket-typed symbols and imports neither gui nor gorilla"},
}

// Ch24Result is what the runner observes. Errs maps a check ID to the reason
// it failed; a check ID present with an empty string passed.
type Ch24Result struct {
	Errs  map[string]string
	Fatal string
}

// ran marks a check as exercised. A check that never ran reports that fact
// rather than silently scoring zero, so a harness failure is distinguishable
// from a student failure.
func (r *Ch24Result) ran(id string) {
	if r.Errs == nil {
		r.Errs = map[string]string{}
	}
	if _, ok := r.Errs[id]; !ok {
		r.Errs[id] = ""
	}
}

func (r *Ch24Result) fail(id, format string, args ...any) {
	r.ran(id)
	if existing := r.Errs[id]; existing != "" {
		return // keep the first failure: it is usually the cause of the rest
	}
	r.Errs[id] = fmt.Sprintf(format, args...)
}

// failAll marks every check failed with the same reason. Used when the run
// cannot proceed at all (the submission does not build, the probe module
// cannot resolve), so no check silently passes by never being examined.
func (r *Ch24Result) failAll(format string, args ...any) {
	for _, w := range ch24Table {
		r.fail(w.ID, format, args...)
	}
}

// Ch24Checks converts observations into the graded result.
func Ch24Checks(r Ch24Result) []Check {
	out := make([]Check, 0, len(ch24Table))
	for _, w := range ch24Table {
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
