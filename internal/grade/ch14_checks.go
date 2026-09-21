package grade

// Chapter 14 checks: the speech channel.
//
// Weighting note, deliberate and documented in book/review-ch14.md: the two
// heaviest checks are the two defects that survived a passing grader suite, an
// automated observer, and a human listening to the output every working day.

func Ch14Checks(r Ch14Result) []Check {
	return []Check{
		ch14BuffersFragments(r),
		ch14FiltersMarkup(r),
		ch14Boundaries(r),
		ch14SpeaksUnstreamed(r),
		ch14ExpandsIdentifiers(r),
		ch14GateBothCauses(r),
		ch14Ch13Parity(r),
	}
}

func ch14BuffersFragments(r Ch14Result) Check {
	c := Check{
		ID:     "tts-buffers-fragments",
		Title:  "fragments are buffered to phrase boundaries; no utterance breaks a word",
		Points: 20,
		Earned: 20,
		Passed: true,
	}
	if !r.BuffersOK {
		c.failf("tts-buffers-fragments: %s", r.BuffersErr)
	}
	return c
}

func ch14FiltersMarkup(r Ch14Result) Check {
	c := Check{
		ID:     "tts-filters-markup",
		Title:  "markup is filtered for the ear; a fenced block is named, not read",
		Points: 20,
		Earned: 20,
		Passed: true,
	}
	if !r.FiltersOK {
		c.failf("tts-filters-markup: %s", r.FiltersErr)
	}
	return c
}

func ch14Boundaries(r Ch14Result) Check {
	c := Check{
		ID:     "tts-boundaries",
		Title:  "a lone newline is a space, a blank line is a boundary, flush() speaks the rest",
		Points: 15,
		Earned: 15,
		Passed: true,
	}
	if !r.BoundariesOK {
		c.failf("tts-boundaries: %s", r.BoundariesErr)
	}
	return c
}

func ch14SpeaksUnstreamed(r Ch14Result) Check {
	c := Check{
		ID:     "tts-speaks-unstreamed",
		Title:  "a part arriving whole is spoken, and a part that streamed is not spoken twice",
		Points: 15,
		Earned: 15,
		Passed: true,
	}
	if !r.UnstreamedOK {
		c.failf("tts-speaks-unstreamed: %s", r.UnstreamedErr)
	}
	return c
}

func ch14ExpandsIdentifiers(r Ch14Result) Check {
	c := Check{
		ID:     "tts-expands-identifiers",
		Title:  "identifiers split into words; ordinary words and ALL CAPS are left alone",
		Points: 10,
		Earned: 10,
		Passed: true,
	}
	if !r.IdentifiersOK {
		c.failf("tts-expands-identifiers: %s", r.IdentifiersErr)
	}
	return c
}

func ch14GateBothCauses(r Ch14Result) Check {
	c := Check{
		ID:     "tts-gate-both-causes",
		Title:  "one edge-triggered predicate; a space counts as input; unpause needs both causes clear",
		Points: 10,
		Earned: 10,
		Passed: true,
	}
	if !r.GateOK {
		c.failf("tts-gate-both-causes: %s", r.GateErr)
	}
	return c
}

func ch14Ch13Parity(r Ch14Result) Check {
	c := Check{
		ID:     "ch13-parity",
		Title:  "chapter 13 still passes",
		Points: 10,
		Earned: 10,
		Passed: true,
	}
	if !r.Ch13Parity {
		c.failf("ch13-parity: %s", r.Ch13ParityErr)
	}
	return c
}
