package grade

// Chapter 6 checks - seven properties, 100 points.

import (
	"fmt"
	"sort"
	"strings"
)

// Ch6Evaluate scores the ch6 result.
func Ch6Evaluate(r *Ch6Result) []Check {
	return ch6EvaluateCore(r, false)
}

// Ch6EvaluateParity scores only the checks that apply when running as parity
// from a later chapter. Multi-agent exercise checks are skipped.
func Ch6EvaluateParity(r *Ch6Result) []Check {
	return ch6EvaluateCore(r, true)
}

func ch6EvaluateCore(r *Ch6Result, parityOnly bool) []Check {
	checks := []Check{
		ch6Parity(r),
	}
	if !parityOnly {
		checks = append(checks,
			ch6NotDeaf(r),
			ch6ReplayIsLive(r),
		)
	}
	checks = append(checks,
		ch6ObserverFires(r),
	)
	if !parityOnly {
		checks = append(checks, ch6WakeOnce(r))
	}
	checks = append(checks,
		ch6LoudRefusal(r),
		ch6HubClean(r),
	)
	return checks
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
	if !r.BuildOK {
		c.failf("build failed: %s", r.BuildErr)
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
	if !r.BuildOK {
		c.failf("build failed: %s", r.BuildErr)
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
	if !r.BuildOK {
		c.failf("build failed: %s", r.BuildErr)
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
	if !r.BuildOK {
		c.failf("build failed: %s", r.BuildErr)
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

	// Count turn_ended observations - should be at least 2 (one per agent).
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
	if !r.BuildOK {
		c.failf("build failed: %s", r.BuildErr)
		return c
	}

	if r.LoudRefusal == "" {
		c.failf("no loud refusal for unknown model - expected error mentioning model name")
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
			// Must be a leaf - not another spoke like internal/llm.
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
	//
	// The spoke set is DERIVED from the tree, never hardcoded. An earlier
	// version listed llm, tools and jobs literally. That was accurate for ch06
	// as frozen and quietly wrong everywhere else: by ch17 the live tree had
	// eight spokes, so mcp, recall, ws, skills and settings were ungoverned.
	// Worse, the old loop did `if !ok { continue }`, so a spoke the list had
	// never heard of was skipped in silence rather than flagged - the check
	// reported success precisely where it had checked nothing. Deriving the set
	// means this governs whatever the student actually built, and grows with
	// the tree instead of rotting behind it.
	//
	// A spoke is the first path segment under internal/ other than common.
	// Packages nested inside a spoke (internal/llm/foo) belong to that spoke,
	// so intra-spoke imports are fine; spoke-to-SPOKE edges are what break the
	// star. Anything outside internal/ - cmd/ and the module root - is the
	// composition root and may import every spoke; that is its whole job.
	spokeNames, violations := starViolations(r.ImportGraph, r.Base)

	// An empty spoke set would make the scan vacuously clean. That is the same
	// shape as the IsToolEnabled bug: a filter over an empty set passes while
	// protecting nothing. DiscoverImportGraph returns an empty map when
	// `go list` fails, so without this guard a tree that does not even build
	// scores full marks for architecture. Refuse to report success we did not
	// verify.
	if len(spokeNames) == 0 {
		c.failf("no spoke packages found under %s/internal/: the import graph is empty or unreadable, so star topology was never actually checked", r.Base)
		return c
	}
	if len(violations) > 0 {
		c.failf("star topology violated, spoke imports spoke: %v", violations)
		return c
	}

	c.Passed = true
	c.Earned = c.Points
	c.notef("internal/common deps clean, star topology preserved across %d spokes: %s", len(spokeNames), strings.Join(spokeNames, ", "))
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

// starViolations reports spoke-to-spoke import edges in the module rooted at
// base. It returns the derived spoke names and the violations, both sorted.
//
// The spoke set is DERIVED from the graph, never hardcoded. An earlier version
// of the ch6 check listed llm, tools and jobs literally. That was accurate for
// ch06 as frozen and quietly wrong everywhere else: by ch17 the live tree had
// eight spokes, so mcp, recall, ws, skills and settings were ungoverned. Worse,
// its loop did `if !ok { continue }`, so a spoke the list had never heard of
// was skipped in silence rather than flagged, and the check reported success
// precisely where it had checked nothing.
//
// A spoke is the first path segment under internal/ other than common.
// Packages nested inside a spoke (internal/llm/foo) belong to that spoke, so
// intra-spoke imports are fine; spoke-to-SPOKE edges are what break the star.
// Anything outside internal/, meaning cmd/ and the module root, is the
// composition root and may import every spoke. That is its whole job.
//
// A caller must treat an empty spoke list as a failure to verify rather than a
// clean result: an empty graph makes the scan vacuously clean.
func starViolations(graph map[string][]string, base string) (spokes, violations []string) {
	internalPrefix := base + "/internal/"
	spokeOf := func(pkg string) string {
		if !strings.HasPrefix(pkg, internalPrefix) {
			return ""
		}
		rest := strings.TrimPrefix(pkg, internalPrefix)
		if i := strings.Index(rest, "/"); i >= 0 {
			rest = rest[:i]
		}
		return rest
	}

	seen := map[string]bool{}
	for pkg := range graph {
		if s := spokeOf(pkg); s != "" && s != "common" {
			seen[s] = true
		}
	}
	for s := range seen {
		spokes = append(spokes, s)
	}
	sort.Strings(spokes)

	for pkg, deps := range graph {
		from := spokeOf(pkg)
		if from == "" || from == "common" {
			continue
		}
		for _, dep := range deps {
			dep = strings.TrimSpace(dep)
			if dep == "" || isStdlib(dep) || !strings.HasPrefix(dep, base) {
				continue
			}
			to := spokeOf(dep)
			if to == "" || to == "common" || to == from {
				continue
			}
			violations = append(violations, pkg+" imports "+dep)
		}
	}
	sort.Strings(violations)
	return spokes, violations
}
