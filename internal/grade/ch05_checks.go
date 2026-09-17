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
		c := Check{ID: "agent-builds", Title: "Agent builds from ./cmd/", Points: 10}
		if r.AgentBuildOK {
			c.Passed = true
			c.Earned = c.Points
		} else {
			c.failf("agent binary failed to build: %s", r.AgentBuildErr)
		}
		checks = append(checks, c)
	}

	// ---- exercise-builds (10 pts) -----------------------------------------
	{
		c := Check{ID: "exercise-builds", Title: "Exercise binary imports and builds", Points: 10}
		if r.ExBuildOK {
			c.Passed = true
			c.Earned = c.Points
		} else {
			c.failf("exercise binary failed to build: %s", r.ExBuildErr)
		}
		checks = append(checks, c)
	}

	// ---- star-topology (25 pts) -------------------------------------------
	{
		c := Check{ID: "star-topology", Title: "Star topology: impl packages import only common", Points: 25}
		if r.Base == "" {
			c.failf("could not determine module path from agent/go.mod")
		} else {
			commonPkg := ""
			// Find the common package (whatever it's called).
			for pkg := range r.ImportGraph {
				if strings.HasSuffix(pkg, "/common") && strings.Contains(pkg, "/internal/") {
					commonPkg = pkg
					break
				}
			}
			if commonPkg == "" {
				c.failf("no internal/common package found in import graph")
			} else {
				// Find all internal implementation packages.
				var violations []string
				for pkg, imports := range r.ImportGraph {
					if !strings.Contains(pkg, "/internal/") || pkg == commonPkg {
						continue
					}
					if strings.HasSuffix(pkg, "/cmd") || strings.Contains(pkg, "/cmd/") {
						continue
					}
					for _, imp := range imports {
						if imp == commonPkg || imp == pkg {
							continue
						}
						if strings.Contains(imp, r.Base+"/internal/") && imp != commonPkg {
							violations = append(violations, fmt.Sprintf("%s imports %s", pkg, imp))
						}
					}
				}
				if len(violations) > 0 {
					c.failf("packages import each other:\n  %s", strings.Join(violations, "\n  "))
				} else {
					c.Passed = true
					c.Earned = c.Points
					c.notef("all implementation packages import only %s", commonPkg)
				}
			}
		}
		checks = append(checks, c)
	}

	// ---- custom-tool-called (25 pts) --------------------------------------
	{
		c := Check{ID: "custom-tool-called", Title: "Custom tool is called by exercise binary", Points: 25}
		if !r.ExBuildOK {
			c.failf("exercise binary did not build")
		} else if !r.ExToolCalled {
			c.failf("the custom tool was never called")
		} else {
			c.Passed = true
			c.Earned = c.Points
			c.notef("tool returned: %s", r.ExToolOutput)
		}
		checks = append(checks, c)
	}

	// ---- logger-accessible (10 pts) ----------------------------------------
	{
		c := Check{ID: "logger-accessible", Title: "Logger reachable through parent interface chain", Points: 10}
		if !r.AgentBuildOK {
			c.failf("agent binary did not build")
		} else if r.HasLogf {
			c.Passed = true
			c.Earned = c.Points
			c.notef("Host interface with Logf found in common, embedded in Call")
		} else {
			c.failf("no Host interface with Logf found in common, or not embedded in Call")
		}
		checks = append(checks, c)
	}

	// ---- no-mutable-globals (10 pts) --------------------------------------
	{
		c := Check{ID: "no-mutable-globals", Title: "No mutable package-level variables", Points: 10}
		if !r.AgentBuildOK {
			c.failf("agent binary did not build")
		} else if len(r.MutableGlobals) == 0 {
			c.Passed = true
			c.Earned = c.Points
		} else {
			c.failf("mutable globals found: %s", strings.Join(r.MutableGlobals, ", "))
		}
		checks = append(checks, c)
	}

	// ---- ch4-parity (30 pts) ----------------------------------------------
	{
		c := Check{ID: "ch4-parity", Title: "Agent passes ch4 behavioral checks", Points: 30}
		if !r.AgentBuildOK {
			c.failf("agent binary did not build")
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
