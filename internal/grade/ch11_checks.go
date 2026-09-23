package grade

// ch11_checks.go — the Chapter 11 check table.
//
// The names, points and order below are the contract printed in the
// chapter's TL;DR. Total 100.

// Ch11Checks returns the check list for the persistence chapter.
func Ch11Checks(r Ch11Result) []Check {
	return []Check{
		ch11SaveShape(r),
		ch11DefaultLoad(r),
		ch11ReplayEqualsSnapshot(r),
		ch11TailAppliedOnce(r),
		ch11LogNotNeeded(r),
		ch11BadSaveRefused(r),
		ch11Ch10Parity(r),
	}
}

func ch11SaveShape(r Ch11Result) Check {
	c := Check{
		ID:     "save-shape",
		Title:  "Save file records config, anchors at the last log Seq, numbers strictly up",
		Points: 15,
		Earned: 15,
		Passed: true,
	}
	if !r.SaveShapeOK {
		c.failf("save-shape: %s", r.SaveShapeErr)
	}
	return c
}

func ch11DefaultLoad(r Ch11Result) Check {
	c := Check{
		ID:     "default-load",
		Title:  "A second start in the same directory, no flags, resumes the conversation",
		Points: 20,
		Earned: 20,
		Passed: true,
	}
	if !r.DefaultLoadOK {
		c.failf("default-load: %s", r.DefaultLoadErr)
	}
	return c
}

func ch11ReplayEqualsSnapshot(r Ch11Result) Check {
	c := Check{
		ID:     "replay-equals-snapshot",
		Title:  "Loading a save, and loading it with context null, render identical requests",
		Points: 20,
		Earned: 20,
		Passed: true,
	}
	if !r.ReplayEqualsSnapshotOK {
		c.failf("replay-equals-snapshot: %s", r.ReplayEqualsSnapshotErr)
	}
	return c
}

func ch11TailAppliedOnce(r Ch11Result) Check {
	c := Check{
		ID:     "tail-applied-once",
		Title:  "An older snapshot spliced onto a newer log replays the tail exactly once",
		Points: 15,
		Earned: 15,
		Passed: true,
	}
	if !r.TailAppliedOnceOK {
		c.failf("tail-applied-once: %s", r.TailAppliedOnceErr)
	}
	return c
}

func ch11LogNotNeeded(r Ch11Result) Check {
	c := Check{
		ID:     "log-not-needed",
		Title:  "A snapshot with an empty log is a complete save, and numbering continues",
		Points: 10,
		Earned: 10,
		Passed: true,
	}
	if !r.LogNotNeededOK {
		c.failf("log-not-needed: %s", r.LogNotNeededErr)
	}
	return c
}

func ch11BadSaveRefused(r Ch11Result) Check {
	c := Check{
		ID:     "bad-save-refused",
		Title:  "A save file that does not parse is fatal, and its bytes are left alone",
		Points: 5,
		Earned: 5,
		Passed: true,
	}
	if !r.BadSaveRefusedOK {
		c.failf("bad-save-refused: %s", r.BadSaveRefusedErr)
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
