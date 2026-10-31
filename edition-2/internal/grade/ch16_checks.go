package grade

// Chapter 16 checks.
//
// Every one of these is read off the wire: what the agent sent to the
// model, and what it stopped sending. None of them looks inside the
// student's tree for a function, a file name, or a directory layout. A
// student who builds the ladder differently from the reference and gets
// the same behaviour scores the same.

type ch16Check struct {
	ID     string
	Points int
	Why    string
}

var ch16Table = []ch16Check{
	{"micro-handoff-compresses", 15,
		"A checkpoint that crosses the budget turns the oldest finished work into a memory, " +
			"and the raw work it replaces leaves the context."},
	{"planted-fact-survives", 10,
		"The compressor is shown the work it is folding, and what it writes comes back as context."},
	{"graduation-fires-oldest-first", 12,
		"A band over its budget folds its OLDEST contents into the band above, leaving the newest alone."},
	{"disable-enable-idempotent", 12,
		"Switching bands off and back on - in any order - restores exactly the memory that was there before."},
	{"disabled-neighbor-refused", 8,
		"A band whose upper neighbour is switched off refuses to graduate rather than dropping memories on the floor."},
	{"forced-handoff", 12,
		"A nearly full context leaves micro_handoff as the only tool on offer, because asking nicely does not work."},
	{"replay-needs-no-llm", 10,
		"Rebuilding a context replays recorded decisions instead of re-running the model that made them."},
	{"abandon-on-restart", 8,
		"A fold that fails or is interrupted loses no work and is never applied twice."},
	{"memory-is-data", 5,
		"Memory arrives in the conversation, written in the first person, not bolted onto the system prompt."},
	{"fresh-start-populates", 5,
		"An agent starting with memory files on disk comes up with them already in context."},
	{"ch15-parity", 3,
		"Chapter 15's ladder and journal still work; memory is added on top of them, not instead of them."},
}

// Ch16Checks turns a run into the graded checks.
func Ch16Checks(r Ch16Result) []Check {
	out := make([]Check, 0, len(ch16Table))
	for _, w := range ch16Table {
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
