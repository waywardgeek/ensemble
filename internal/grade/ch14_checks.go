package grade

// Chapter 14 checks: the speech channel.
//
// These checks read a log, not a source tree. The student declares a way to
// run their system with the speech engine replaced by a recorder; the grader
// scripts the model's side of the conversation and reads what came out. No
// check here names a method, a module, or a language.
//
// Weighting, deliberate and recorded in book/review-ch14.md:
//
//   - tts-discrimination is the heaviest because it is the one check that
//     cannot be satisfied by accident. The same string arrives twice, once
//     as a tool result and once as the model's own words. Speak everything
//     and half of it fails; speak nothing and the other half fails.
//   - tts-gate-both-causes and tts-speaks-errors are worth ten each because
//     they cover the same failure: a wire connected at one end. Both were
//     found by reading the log rather than the pipeline, which is the whole
//     argument for grading the channel.
//   - The two checks that test pipeline niceties rather than shipped defects
//     (boundaries, identifiers) are worth five each.

func Ch14Checks(r Ch14Result) []Check {
	return []Check{
		ch14HarnessRuns(r),
		ch14Discrimination(r),
		ch14FiltersMarkup(r),
		ch14BuffersFragments(r),
		ch14SpeaksUnstreamed(r),
		ch14GateBothCauses(r),
		ch14SpeaksErrors(r),
		ch14Boundaries(r),
		ch14ExpandsIdentifiers(r),
		ch14Ch13Parity(r),
	}
}

// ch14HarnessRuns is worth nothing and reported first on purpose. Without it
// a harness that fails to launch reports eight mysterious zeros; with it, the
// first line of the report says so.
func ch14HarnessRuns(r Ch14Result) Check {
	c := Check{
		ID:     "tts-harness-runs",
		Title:  "the speech harness launches the system and produces a log",
		Points: 0,
		Earned: 0,
		Passed: true,
	}
	if !r.HarnessOK {
		c.failf("tts-harness-runs: %s", r.HarnessErr)
	}
	return c
}

func ch14Discrimination(r Ch14Result) Check {
	c := Check{
		ID:     "tts-discrimination",
		Title:  "tool results stay silent; the model's own words are spoken",
		Points: 20,
		Earned: 20,
		Passed: true,
	}
	if !r.DiscriminationOK {
		c.failf("tts-discrimination: %s", r.DiscriminationErr)
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

func ch14BuffersFragments(r Ch14Result) Check {
	c := Check{
		ID:     "tts-buffers-fragments",
		Title:  "fragments are assembled before they are spoken",
		Points: 10,
		Earned: 10,
		Passed: true,
	}
	if !r.BuffersOK {
		c.failf("tts-buffers-fragments: %s", r.BuffersErr)
	}
	return c
}

func ch14SpeaksUnstreamed(r Ch14Result) Check {
	c := Check{
		ID:     "tts-speaks-unstreamed",
		Title:  "a reply that arrives whole is still spoken",
		Points: 15,
		Earned: 15,
		Passed: true,
	}
	if !r.UnstreamedOK {
		c.failf("tts-speaks-unstreamed: %s", r.UnstreamedErr)
	}
	return c
}

func ch14GateBothCauses(r Ch14Result) Check {
	c := Check{
		ID:     "tts-gate-both-causes",
		Title:  "typing pauses speech, the log says why, and speech resumes",
		Points: 10,
		Earned: 10,
		Passed: true,
	}
	if !r.GateOK {
		c.failf("tts-gate-both-causes: %s", r.GateErr)
	}
	return c
}

func ch14SpeaksErrors(r Ch14Result) Check {
	c := Check{
		ID:     "tts-speaks-errors",
		Title:  "a failure is spoken rather than leaving the listener waiting",
		Points: 10,
		Earned: 10,
		Passed: true,
	}
	if !r.ErrorsOK {
		c.failf("tts-speaks-errors: %s", r.ErrorsErr)
	}
	return c
}

func ch14Boundaries(r Ch14Result) Check {
	c := Check{
		ID:     "tts-boundaries",
		Title:  "sentences are separate utterances a listener can interrupt",
		Points: 5,
		Earned: 5,
		Passed: true,
	}
	if !r.BoundariesOK {
		c.failf("tts-boundaries: %s", r.BoundariesErr)
	}
	return c
}

func ch14ExpandsIdentifiers(r Ch14Result) Check {
	c := Check{
		ID:     "tts-expands-identifiers",
		Title:  "run-together identifiers are spoken as words, all-caps left alone",
		Points: 5,
		Earned: 5,
		Passed: true,
	}
	if !r.IdentifiersOK {
		c.failf("tts-expands-identifiers: %s", r.IdentifiersErr)
	}
	return c
}

func ch14Ch13Parity(r Ch14Result) Check {
	c := Check{
		ID:     "ch13-parity",
		Title:  "chapter 13 still passes",
		Points: 5,
		Earned: 5,
		Passed: true,
	}
	if !r.Ch13Parity {
		c.failf("ch13-parity: %s", r.Ch13ParityErr)
	}
	return c
}
