package grade

// Chapter 18 checks: the invisible invoice.
//
// Every other chapter's grader can be satisfied by an agent that behaves
// correctly. This one cannot, because the defect it grades produces correct
// behaviour: the tool call works, the edit lands, the test passes, and the
// only symptom is the bill. So these checks read the bytes on the wire, the
// counts in the event log, and the numbers on the meter, and never ask the
// agent how it thinks it is doing.
var ch18Table = []struct {
	ID     string
	Points int
	Why    string
}{
	{"cache-lens-exists", 15,
		"a cache lens observes consecutive requests, so a broken prefix can be found"},
	{"prefix-is-stable", 20,
		"the tool and system sections are byte-identical across turns, so they can be cached at all"},
	{"system-has-breakpoint", 15,
		"the system prompt carries a cache_control marker on the wire"},
	{"cost-is-computed-not-stored", 15,
		"the event log holds token counts only; cost is derived at display time"},
	{"usage-is-session-scoped", 10,
		"the meter reports the session, not the lifetime restored from disk"},
	{"cache-read-rate", 10,
		"cache reads are zero on a cold turn and reported faithfully on a warm one"},
	{"all-vendors-report-usage", 15,
		"every vendor parser extracts all four usage categories from its own reply shape"},
}

func Ch18Checks(r Ch18Result) []Check {
	out := make([]Check, 0, len(ch18Table))
	for _, w := range ch18Table {
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
