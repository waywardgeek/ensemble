package grade

// ch15Weights: provisional per the TL;DR table; the P9 audit in
// book/review-ch15.md is what fixes them.
var ch15Weights = []struct {
	name  string
	desc  string
	score int
}{
	{"skill-survives-the-ladder", "A loaded skill appears exactly once, outside tool parts, before and after the ladder fires (rules 3, 4, 7)", 15},
	{"micro-handoff-shape", "After micro_handoff: no tool parts, no orphans, the handoff text exactly once (rule 5)", 15},
	{"keep-or-stub", "Unkept results are stubbed one request later on the stubbing model; kept ones and the no-stub model are not (rule 6)", 10},
	{"ladder-is-recorded", "Restarting under a larger target reproduces the recorded cuts byte for byte (rule 7)", 15},
	{"frozen-prefix", "System prompt and startup tools byte-identical across a skill load and an MCP connect; new tools reach the dialog (rules 1, 2)", 10},
	{"replay-equals-snapshot", "Loading with the context nulled equals loading the snapshot, on a save with handoff, redaction and skill events (rules 8, 9)", 15},
	{"crash-recovery", "SIGKILL after an anchored save loses no turn (rule 9)", 15},
	{"total-reducer", "Two parseable but unappliable events are skipped; the agent starts with the rest intact (rule 8)", 5},
}

// Ch15Checks turns a run into the scored check list.
func Ch15Checks(r Ch15Result) []Check {
	var out []Check
	for _, w := range ch15Weights {
		c := Check{ID: w.name, Title: w.desc, Points: w.score, Earned: w.score, Passed: true}
		switch msg, ran := r.Errs[w.name]; {
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
