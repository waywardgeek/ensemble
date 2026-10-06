package grade

import "fmt"

// Chapter 21 checks: giving the agent the web.
//
// The chapter's claim is that twelve chapters of MCP work mean a whole new
// capability costs one transport and one configuration file. So the checks are
// weighted toward the things a shortcut would skip: reaching a server over a
// transport that did not exist before, keeping the capability off until it is
// asked for, carrying fetched bytes all the way to the model without losing
// them, and making both kinds of failure visible.
//
// Every check reads either the wire the student's agent produced or the calls
// a fake MCP server received. None of them reads the student's source. The
// grader supplies the skill file, so it owns the endpoint and the tool names:
// a student who chose a different search backend is graded on their wiring,
// not on their vendor.
//
// One check is unfakeable by construction. The fake server plants a token
// generated fresh each run inside one page; a stub, a cached fixture or the
// model's own knowledge cannot produce it.

// ch21Table is the published contract: seven checks, one hundred points.
var ch21Table = []struct {
	ID     string
	Points int
	Why    string
}{
	{"tools-gated-by-skill", 15,
		"The web tools are absent until the web-search skill is loaded. Off is a " +
			"statement about what this feature initiates, and the honest way to show it " +
			"is that there is no tool to call."},
	{"url-transport-connects", 15,
		"A skill declaring transport: url reaches a hosted MCP server and its tools " +
			"appear in the list sent to the model. This is the seam Chapter 12 left " +
			"open, filled in."},
	{"search-dispatches", 15,
		"A search call reaches the server and its results come back to the model as a " +
			"tool result, rather than being dispatched and dropped."},
	{"fetch-returns-planted-token", 20,
		"Content fetched from a page arrives at the model intact. Verified with a token " +
			"minted this run, so a stub or a cached fixture cannot produce it."},
	{"fetched-content-is-a-tool-result", 15,
		"Fetched bytes stay inside a tool result and never reach the system prompt or a " +
			"user turn. Web pages contain instructions aimed at the model; keeping them " +
			"structurally separate from the operator's words is what makes them data."},
	{"tool-error-is-reported", 10,
		"A tool-level failure (isError on a 200) is told to the model. A failure the " +
			"model cannot see is one it will confidently narrate around."},
	{"transport-error-is-reported", 10,
		"An HTTP-level failure reaches the model with the reason intact. The response " +
			"body is where quota and auth refusals say what went wrong; a transport that " +
			"discards it turns a diagnosable refusal into 'request failed'."},
}

// Ch21Result is what the runner observes. Errs maps a check ID to the reason
// it failed; a check ID present with an empty string passed.
type Ch21Result struct {
	Errs  map[string]string
	Fatal string
}

// ran marks a check as exercised. A check that never ran reports that fact
// rather than silently scoring zero, so a harness failure is distinguishable
// from a student failure.
func (r *Ch21Result) ran(id string) {
	if r.Errs == nil {
		r.Errs = map[string]string{}
	}
	if _, ok := r.Errs[id]; !ok {
		r.Errs[id] = ""
	}
}

func (r *Ch21Result) fail(id, format string, args ...any) {
	r.ran(id)
	if existing := r.Errs[id]; existing != "" {
		return // keep the first failure: it is usually the cause of the rest
	}
	r.Errs[id] = fmt.Sprintf(format, args...)
}

// Ch21Checks converts observations into the graded result.
func Ch21Checks(r Ch21Result) []Check {
	out := make([]Check, 0, len(ch21Table))
	for _, w := range ch21Table {
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
