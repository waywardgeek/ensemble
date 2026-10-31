package grade

import "fmt"

// ch25Table is chapter 25's published contract: seven checks, one hundred
// points. The chapter's thesis is that a fleet must report what the framework
// watched happen rather than what an agent claims, and check 3 is that thesis
// made mechanical: an agent that stops without submitting has not finished,
// however confidently it stopped.
//
// Everything here is graded through the seven pinned tool names, which are
// the only names the contract fixes. Type names, file layout and how the
// manager is wired are free, so the harness drives the real binary and never
// reaches into the student's packages.
var ch25Table = []struct {
	ID     string
	Points int
	Why    string
}{
	{"spawn-reports-back", 15, "A spawned child runs and reports back: the parent calls spawn_sub_agent and the child's submitted payload reaches the parent"},
	{"child-isolated", 15, "Children are isolated and factory built: each child owns its data directory and save file, and an AgentFactory constructs it"},
	{"fleet-visibility", 10, "Fleet visibility and teardown: agents_status lists a live child, and shutdown_agent removes it"},
	{"join-fans-in", 15, "join_agents blocks until every listed child finishes, then returns each child's result"},
	{"addressed-messaging", 15, "Agents address their parent and the user: send_message_to_parent blocks the child, respond_to_parent_message resumes it with a reply that carries its sender's address"},
	{"wait-interrupt", 15, "A message arriving during a slow tool wakes the agent before that tool finishes, and the backgrounded tool's result still arrives"},
	{"idle-corrected", 15, "A child that goes idle without submitting is woken rather than treated as finished, and after three wakes it escalates to its parent"},
}

// Ch25Result is what the runner observes. Errs maps a check ID to the reason
// it failed; a check ID present with an empty string passed.
type Ch25Result struct {
	Errs  map[string]string
	Fatal string
}

// ran marks a check as exercised. A check that never ran reports that fact
// rather than silently scoring zero, so a harness failure is distinguishable
// from a student failure.
func (r *Ch25Result) ran(id string) {
	if r.Errs == nil {
		r.Errs = map[string]string{}
	}
	if _, ok := r.Errs[id]; !ok {
		r.Errs[id] = ""
	}
}

func (r *Ch25Result) fail(id, format string, args ...any) {
	r.ran(id)
	if existing := r.Errs[id]; existing != "" {
		return // keep the first failure: it is usually the cause of the rest
	}
	r.Errs[id] = fmt.Sprintf(format, args...)
}

// failAll marks every check failed with the same reason. Used when the run
// cannot proceed at all (the submission does not build, the binary will not
// start), so no check silently passes by never being examined.
func (r *Ch25Result) failAll(format string, args ...any) {
	for _, w := range ch25Table {
		r.fail(w.ID, format, args...)
	}
}

// Ch25Checks converts observations into the graded result.
func Ch25Checks(r Ch25Result) []Check {
	out := make([]Check, 0, len(ch25Table))
	for _, w := range ch25Table {
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
