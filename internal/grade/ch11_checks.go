package grade

// ch11Checks returns the check list for the persistence chapter.
func Ch11Checks(r Ch11Result) []Check {
	return []Check{
		ch11Deterministic(r),
		ch11SaveConfig(r),
		ch11Roundtrip(r),
		ch11Resume(r),
		ch11LogNotNeeded(r),
		ch11PartialReplay(r),
		ch11Ch10Parity(r),
	}
}

func ch11Deterministic(r Ch11Result) Check {
	c := Check{
		ID:     "deterministic-context",
		Title:  "Context is deterministically built from the event log",
		Points: 20,
		Earned: 20,
		Passed: true,
	}
	if !r.DeterministicOK {
		c.failf("deterministic rebuild: %s", r.DeterministicErr)
	}
	return c
}

func ch11SaveConfig(r Ch11Result) Check {
	c := Check{
		ID:     "save-config",
		Title:  "Save file captures model, vendor, system prompt, and tools",
		Points: 15,
		Earned: 15,
		Passed: true,
	}
	if !r.SaveConfigOK {
		c.failf("save config: %s", r.SaveConfigErr)
	}
	return c
}

func ch11Roundtrip(r Ch11Result) Check {
	c := Check{
		ID:     "save-roundtrip",
		Title:  "Save file survives JSON roundtrip",
		Points: 10,
		Earned: 10,
		Passed: true,
	}
	if !r.RoundtripOK {
		c.failf("roundtrip: %s", r.RoundtripErr)
	}
	return c
}

func ch11Resume(r Ch11Result) Check {
	c := Check{
		ID:     "resume-continues",
		Title:  "Agent resumes conversation from loaded save file",
		Points: 20,
		Earned: 20,
		Passed: true,
	}
	if !r.ResumeOK {
		c.failf("resume: %s", r.ResumeErr)
	}
	return c
}

func ch11LogNotNeeded(r Ch11Result) Check {
	c := Check{
		ID:     "log-not-needed",
		Title:  "Context alone (no event log) is sufficient to continue",
		Points: 15,
		Earned: 15,
		Passed: true,
	}
	if !r.LogNotNeededOK {
		c.failf("log-not-needed: %s", r.LogNotNeededErr)
	}
	return c
}

func ch11PartialReplay(r Ch11Result) Check {
	c := Check{
		ID:     "partial-replay",
		Title:  "Event log maintains monotonic sequence ordering",
		Points: 5,
		Earned: 5,
		Passed: true,
	}
	if !r.PartialReplayOK {
		c.failf("partial replay: %s", r.PartialReplayErr)
	}
	return c
}

func ch11Ch10Parity(r Ch11Result) Check {
	c := Check{
		ID:     "ch10-parity",
		Title:  "Chapter 10 checks still pass",
		Points: 15,
		Earned: 15,
		Passed: true,
	}
	if !r.Ch10Parity {
		c.failf("ch10 parity: %s", r.Ch10ParityErr)
	}
	return c
}
