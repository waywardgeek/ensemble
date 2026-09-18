package grade

// Chapter 5 checks: pure predicates over Ch5Result.

import (
	"fmt"
	"strings"
)

func Ch5Evaluate(r *Ch5Result) []Check {
	var checks []Check

	// ---- agent-builds (10 pts) --------------------------------------------
	{
		c := Check{ID: "agent-builds", Title: "Binary builds", Points: 10}
		if r.AgentBuildOK {
			c.Passed = true
			c.Earned = c.Points
		} else {
			c.failf("binary failed to build: %s", r.AgentBuildErr)
		}
		checks = append(checks, c)
	}

	// ---- star-topology (25 pts) -------------------------------------------
	{
		c := Check{ID: "star-topology", Title: "Star topology: impl packages import only the hub", Points: 25}
		if r.Base == "" {
			c.failf("could not determine module path from go.mod")
		} else {
			// Find the hub package: the internal package imported by the most
			// other internal packages. This discovers the student's layout
			// rather than assuming "internal/common".
			hubPkg := discoverHub(r.Base, r.ImportGraph)
			if hubPkg == "" {
				c.failf("no hub package found (an internal package imported by ≥2 others)")
			} else {
				var violations []string
				for pkg, imports := range r.ImportGraph {
					if !strings.Contains(pkg, "/internal/") || pkg == hubPkg {
						continue
					}
					// Skip cmd packages.
					if strings.HasSuffix(pkg, "/cmd") || strings.Contains(pkg, "/cmd/") {
						continue
					}
					for _, imp := range imports {
						if imp == hubPkg || imp == pkg {
							continue
						}
						if strings.Contains(imp, r.Base+"/internal/") && imp != hubPkg {
							violations = append(violations, fmt.Sprintf("%s imports %s", pkg, imp))
						}
					}
				}
				if len(violations) > 0 {
					c.failf("packages import each other:\n  %s", strings.Join(violations, "\n  "))
				} else {
					c.Passed = true
					c.Earned = c.Points
					c.notef("hub=%s, all impl packages import only the hub", hubPkg)
				}
			}
		}
		checks = append(checks, c)
	}

	// ---- tool-called (25 pts) ---------------------------------------------
	{
		c := Check{ID: "tool-called", Title: "Binary executes a tool call", Points: 25}
		if !r.AgentBuildOK {
			c.failf("binary did not build")
		} else if !r.ToolCalled {
			c.failf("the tool call was never executed")
		} else {
			c.Passed = true
			c.Earned = c.Points
			c.notef("tool returned: %s", r.ToolOutput)
		}
		checks = append(checks, c)
	}

	// ---- logger-accessible (10 pts) ----------------------------------------
	{
		c := Check{ID: "logger-accessible", Title: "Logger reachable through parent interface chain", Points: 10}
		if !r.AgentBuildOK {
			c.failf("binary did not build")
		} else if r.HasLogf {
			c.Passed = true
			c.Earned = c.Points
			c.notef("Host interface with Logf found, embedded in Call")
		} else {
			c.failf("no Host interface with Logf found, or not embedded in Call")
		}
		checks = append(checks, c)
	}

	// ---- no-mutable-globals (10 pts) --------------------------------------
	{
		c := Check{ID: "no-mutable-globals", Title: "No mutable package-level variables", Points: 10}
		if !r.AgentBuildOK {
			c.failf("binary did not build")
		} else if len(r.MutableGlobals) == 0 {
			c.Passed = true
			c.Earned = c.Points
		} else {
			c.failf("mutable globals found: %s", strings.Join(r.MutableGlobals, ", "))
		}
		checks = append(checks, c)
	}

	// ---- ch4-parity (40 pts) ----------------------------------------------
	{
		c := Check{ID: "ch4-parity", Title: "Passes ch4 behavioral checks", Points: 40}
		if !r.AgentBuildOK {
			c.failf("binary did not build")
		} else if r.Ch4Result == nil {
			c.failf("ch4 parity test did not run")
		} else {
			ch4checks := Ch4Evaluate(r.Ch4Result)
			earned := 0
			total := 0
			var fails []string
			for _, ch := range ch4checks {
				total += ch.Points
				earned += ch.Earned
				if !ch.Passed {
					fails = append(fails, ch.ID)
				}
			}
			pct := 0
			if total > 0 {
				pct = earned * 100 / total
			}
			if pct >= 80 {
				c.Passed = true
				c.Earned = c.Points
				c.notef("ch4 checks: %d/%d (%d%%)", earned, total, pct)
			} else {
				c.failf("ch4 checks: %d/%d (%d%%), failing: %s",
					earned, total, pct, strings.Join(fails, ", "))
			}
		}
		checks = append(checks, c)
	}

	return checks
}

// discoverHub finds the internal package imported by the most other internal
// packages. This is the student's hub/common package, whatever they named it.
func discoverHub(base string, graph map[string][]string) string {
	importedBy := make(map[string]int)
	internalPrefix := base + "/internal/"

	for pkg, imports := range graph {
		if !strings.HasPrefix(pkg, internalPrefix) {
			continue
		}
		for _, imp := range imports {
			if strings.HasPrefix(imp, internalPrefix) {
				importedBy[imp]++
			}
		}
	}

	best := ""
	bestCount := 0
	for pkg, count := range importedBy {
		if count > bestCount {
			best = pkg
			bestCount = count
		}
	}
	if bestCount >= 2 {
		return best
	}
	return ""
}
