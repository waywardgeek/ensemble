package grade

import "fmt"

// ch26Table is chapter 26's published contract: seven checks, one hundred
// points. The chapter's thesis is that managing a swarm requires activity
// visibility, persistent identity, ownership enforcement, reactive monitoring,
// a working composition root, GUI observability, and regression safety for
// the existing fleet lifecycle.
var ch26Table = []struct {
	ID     string
	Points int
	Why    string
}{
	{"activity-observed", 15, "Activity tracking emits observable transitions: a scripted turn produces at least one activity report through get_sub_agent_status"},
	{"ensemble-persists", 15, "Named agents persist across restart: create an agent, restart the Ensemble from the same directory, verify the agent exists"},
	{"ownership-enforced", 15, "Ownership enforcement rejects cross-tree access: a tool call naming a non-child returns an error"},
	{"wait-change-wakes", 15, "wait_for_agent_change wakes on state change: spawn a child, call wait_for_agent_change, verify it wakes when the child transitions"},
	{"composition-root", 15, "Composition root produces a working binary: build, start, send a prompt, receive a response"},
	{"gui-tree-message", 15, "GUI tree message on WebSocket connect: connect via WebSocket, receive an agents_tree message with at least one agent entry"},
	{"submit-idle-regression", 10, "Bottom-up submit and idle correction still work through the Ensemble layer"},
}

// Ch26Result is what the runner observes. Errs maps a check ID to the reason
// it failed; a check ID present with an empty string passed.
type Ch26Result struct {
	Errs  map[string]string
	Fatal string
}

func (r *Ch26Result) ran(id string) {
	if r.Errs == nil {
		r.Errs = map[string]string{}
	}
	if _, ok := r.Errs[id]; !ok {
		r.Errs[id] = ""
	}
}

func (r *Ch26Result) fail(id, format string, args ...any) {
	r.ran(id)
	if existing := r.Errs[id]; existing != "" {
		return
	}
	r.Errs[id] = fmt.Sprintf(format, args...)
}

func (r *Ch26Result) failAll(format string, args ...any) {
	for _, w := range ch26Table {
		r.fail(w.ID, format, args...)
	}
}

func (r *Ch26Result) passed(id string) bool {
	msg, ran := r.Errs[id]
	return ran && msg == ""
}

// Ch26Checks converts observations into the graded result.
func Ch26Checks(r Ch26Result) []Check {
	out := make([]Check, 0, len(ch26Table))
	for _, w := range ch26Table {
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
