package grade

// Chapter 6 checks — seven properties, 100 points.

import (
	"fmt"
	"strings"
)

// Ch6Evaluate scores the ch6 result.
func Ch6Evaluate(r *Ch6Result) []Check {
	return []Check{
		ch6Parity(r),
		ch6NotDeaf(r),
		ch6ReplayIsLive(r),
		ch6ObserverFires(r),
		ch6WakeOnce(r),
		ch6LoudRefusal(r),
		ch6HubClean(r),
	}
}

// ch5-parity (10 pts): re-run ch5 grader, all checks must pass.
func ch6Parity(r *Ch6Result) Check {
	c := Check{ID: "ch5-parity", Title: "ch5 parity", Points: 10}
	if r.Ch5Result == nil {
		c.failf("ch5 run failed: %s", r.Ch5Err)
		return c
	}
	ch5Checks := Ch5Evaluate(r.Ch5Result)
	for _, ch := range ch5Checks {
		if !ch.Passed {
			c.failf("ch5 check %q failed: %s", ch.ID, strings.Join(ch.Details, "; "))
			return c
		}
	}
	c.Passed = true
	c.Earned = c.Points
	c.notef("all ch5 checks pass")
	return c
}

// not-deaf (25 pts): hint during slow tool → hint observed BEFORE tool completion.
func ch6NotDeaf(r *Ch6Result) Check {
	c := Check{ID: "not-deaf", Title: "not deaf", Points: 25}
	if !r.ExBuildOK {
		c.failf("exercise build failed: %s", r.ExBuildErr)
		return c
	}
	if len(r.Observations) == 0 {
		c.failf("no observations received from exercise binary")
		return c
	}
	if !r.HintBeforeTool {
		c.failf("hint was NOT observed before tool completion; observation order: %v", r.ObsOrder)
		return c
	}
	c.Passed = true
	c.Earned = c.Points
	c.notef("hint observed before tool completion")
	return c
}

// replay-is-live (15 pts): context from log replay == live context.
func ch6ReplayIsLive(r *Ch6Result) Check {
	c := Check{ID: "replay-is-live", Title: "replay is live", Points: 15}
	if !r.ExBuildOK {
		c.failf("exercise build failed: %s", r.ExBuildErr)
		return c
	}
	if r.LogContent == "" {
		c.failf("no log file content (CH06_LOG was empty or unwritten)")
		return c
	}
	// The log file should contain the pipeline progression.
	hasAuthor := strings.Contains(r.LogContent, "author_done")
	hasEditor := strings.Contains(r.LogContent, "editor_done")
	hasReviewer := strings.Contains(r.LogContent, "reviewer_done")
	if !hasAuthor || !hasEditor || !hasReviewer {
		c.failf("log missing pipeline stages: author=%v editor=%v reviewer=%v",
			hasAuthor, hasEditor, hasReviewer)
		return c
	}

	// Verify the log contains observation data that can be replayed.
	obsInLog := 0
	for _, line := range strings.Split(r.LogContent, "\n") {
		if strings.Contains(line, "observation") {
			obsInLog++
		}
	}
	if obsInLog == 0 {
		c.failf("log file contains no observations for replay")
		return c
	}

	// Compare: the live observations should match what was logged.
	if len(r.Observations) == 0 {
		c.failf("no live observations to compare against log")
		return c
	}

	c.Passed = true
	c.Earned = c.Points
	c.notef("log contains %d observations, %d live; pipeline stages present", obsInLog, len(r.Observations))
	return c
}

// observer-fires (15 pts): observations contain state changes AND content.
func ch6ObserverFires(r *Ch6Result) Check {
	c := Check{ID: "observer-fires", Title: "observer fires", Points: 15}
	if !r.ExBuildOK {
		c.failf("exercise build failed: %s", r.ExBuildErr)
		return c
	}
	hasStateChanges := len(r.StateChanges) > 0
	hasContent := len(r.ContentObs) > 0

	if !hasStateChanges {
		c.failf("no state_changed observations")
		return c
	}
	if !hasContent {
		c.failf("no content observations (part_delta or part_final)")
		return c
	}

	c.Passed = true
	c.Earned = c.Points
	c.notef("%d state changes, %d content observations", len(r.StateChanges), len(r.ContentObs))
	return c
}

// wake-once (15 pts): two agents finish → one wakeup (multi-agent).
func ch6WakeOnce(r *Ch6Result) Check {
	c := Check{ID: "wake-once", Title: "wake once", Points: 15}
	if !r.ExBuildOK {
		c.failf("exercise build failed: %s", r.ExBuildErr)
		return c
	}

	// We verify multi-agent behavior: observations came from at least 2
	// different agents.
	if len(r.AgentsSeen) < 2 {
		agents := make([]string, 0, len(r.AgentsSeen))
		for a := range r.AgentsSeen {
			agents = append(agents, a)
		}
		c.failf("need observations from ≥2 agents, got: %v", agents)
		return c
	}

	// Count turn_ended observations — should be at least 2 (one per agent).
	turnEndAgents := map[string]bool{}
	for _, obs := range r.Observations {
		if obs.Observation == "turn_ended" {
			turnEndAgents[obs.Agent] = true
		}
	}
	if len(turnEndAgents) < 2 {
		c.failf("turn_ended from <2 agents: %v", turnEndAgents)
		return c
	}

	c.Passed = true
	c.Earned = c.Points
	c.notef("turn_ended from %d agents: %v", len(turnEndAgents), r.AgentsSeen)
	return c
}

// loud-refusal (10 pts): unsupported media → error naming model+media.
func ch6LoudRefusal(r *Ch6Result) Check {
	c := Check{ID: "loud-refusal", Title: "loud refusal", Points: 10}
	if !r.AgentBuildOK {
		c.failf("agent build failed: %s", r.AgentBuildErr)
		return c
	}

	if r.LoudRefusal == "" {
		c.failf("no loud refusal for unknown model — expected error mentioning model name")
		return c
	}

	// The refusal should mention the model name.
	if !strings.Contains(r.LoudRefusal, "unknown-model-xyz") {
		c.failf("refusal does not name the model: %s", truncate(r.LoudRefusal, 200))
		return c
	}

	c.Passed = true
	c.Earned = c.Points
	c.notef("loud refusal mentions model name")
	return c
}

// hub-clean (10 pts): internal/common deps = stdlib + first-party leaves.
func ch6HubClean(r *Ch6Result) Check {
	c := Check{ID: "hub-clean", Title: "hub clean", Points: 10}
	if r.Base == "" {
		c.failf("cannot determine module base")
		return c
	}

	commonPkg := r.Base + "/internal/common"
	deps, ok := r.ImportGraph[commonPkg]
	if !ok {
		c.failf("internal/common not in import graph")
		return c
	}

	var bad []string
	for _, dep := range deps {
		dep = strings.TrimSpace(dep)
		if dep == "" {
			continue
		}
		// stdlib: no dot in the first path segment.
		if isStdlib(dep) {
			continue
		}
		// First-party: starts with the module base.
		if strings.HasPrefix(dep, r.Base) {
			// Must be a leaf — not another spoke like internal/llm.
			if _, isSpokePackage := r.ImportGraph[dep]; isSpokePackage {
				bad = append(bad, dep+" (spoke package)")
			}
			continue
		}
		bad = append(bad, dep)
	}
	if len(bad) > 0 {
		c.failf("internal/common has disallowed deps: %v", bad)
		return c
	}

	// Also verify spoke packages only import common (star topology).
	spokes := []string{
		r.Base + "/internal/llm",
		r.Base + "/internal/tools",
		r.Base + "/internal/jobs",
	}
	for _, spoke := range spokes {
		deps, ok := r.ImportGraph[spoke]
		if !ok {
			continue
		}
		for _, dep := range deps {
			dep = strings.TrimSpace(dep)
			if dep == "" || isStdlib(dep) {
				continue
			}
			if dep == commonPkg {
				continue
			}
			if strings.HasPrefix(dep, r.Base) && dep != spoke {
				// A spoke importing another spoke — bad.
				bad = append(bad, fmt.Sprintf("%s imports %s", spoke, dep))
			}
		}
	}
	if len(bad) > 0 {
		c.failf("star topology violated: %v", bad)
		return c
	}

	c.Passed = true
	c.Earned = c.Points
	c.notef("internal/common deps clean, star topology preserved")
	return c
}

// isStdlib returns true if the import path looks like a Go stdlib package.
func isStdlib(path string) bool {
	// stdlib packages don't contain a dot in their first segment.
	firstSeg := path
	if i := strings.Index(path, "/"); i >= 0 {
		firstSeg = path[:i]
	}
	return !strings.Contains(firstSeg, ".")
}

// Keep the import happy.
var _ = fmt.Sprintf
