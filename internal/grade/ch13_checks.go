package grade

// ch13Checks returns the check list for the gui-debug skill chapter.
func Ch13Checks(r Ch13Result) []Check {
	return []Check{
		ch13SkillLoads(r),
		ch13EphemeralInjected(r),
		ch13SnapshotUpdates(r),
		ch13TTSVisibility(r),
		ch13GUIInteraction(r),
		ch13SkillUnload(r),
		ch13Ch12Parity(r),
	}
}

func ch13SkillLoads(r Ch13Result) Check {
	c := Check{
		ID:     "skill-loads",
		Title:  "gui-debug skill discovered and loaded, MCP handshake completes",
		Points: 15,
		Earned: 15,
		Passed: true,
	}
	if !r.SkillLoadsOK {
		c.failf("skill-loads: %s", r.SkillLoadsErr)
	}
	return c
}

func ch13EphemeralInjected(r Ch13Result) Check {
	c := Check{
		ID:     "ephemeral-injected",
		Title:  "gui_snapshot result appears in context sent to LLM (in Ephemera, not dialogue)",
		Points: 20,
		Earned: 20,
		Passed: true,
	}
	if !r.EphemeralInjectedOK {
		c.failf("ephemeral-injected: %s", r.EphemeralInjectedErr)
	}
	return c
}

func ch13SnapshotUpdates(r Ch13Result) Check {
	c := Check{
		ID:     "snapshot-updates",
		Title:  "After gui_click, next round's gui_snapshot reflects the change",
		Points: 15,
		Earned: 15,
		Passed: true,
	}
	if !r.SnapshotUpdatesOK {
		c.failf("snapshot-updates: %s", r.SnapshotUpdatesErr)
	}
	return c
}

func ch13TTSVisibility(r Ch13Result) Check {
	c := Check{
		ID:     "tts-visibility",
		Title:  "tts_queue returns utterance data in ephemeral context",
		Points: 10,
		Earned: 10,
		Passed: true,
	}
	if !r.TTSVisibilityOK {
		c.failf("tts-visibility: %s", r.TTSVisibilityErr)
	}
	return c
}

func ch13GUIInteraction(r Ch13Result) Check {
	c := Check{
		ID:     "gui-interaction",
		Title:  "Agent calls gui_click/gui_input correctly, gets confirmation",
		Points: 15,
		Earned: 15,
		Passed: true,
	}
	if !r.GUIInteractionOK {
		c.failf("gui-interaction: %s", r.GUIInteractionErr)
	}
	return c
}

func ch13SkillUnload(r Ch13Result) Check {
	c := Check{
		ID:     "skill-unload",
		Title:  "Unloading gui-debug removes MCP tools and closes transport",
		Points: 15,
		Earned: 15,
		Passed: true,
	}
	if !r.SkillUnloadOK {
		c.failf("skill-unload: %s", r.SkillUnloadErr)
	}
	return c
}

func ch13Ch12Parity(r Ch13Result) Check {
	c := Check{
		ID:     "ch12-parity",
		Title:  "All ch12 behavior preserved",
		Points: 10,
		Earned: 10,
		Passed: true,
	}
	if !r.Ch12Parity {
		c.failf("ch12-parity: %s", r.Ch12ParityErr)
	}
	return c
}
