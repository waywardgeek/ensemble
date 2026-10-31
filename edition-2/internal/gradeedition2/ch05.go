package gradeedition2

import (
	"strings"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter5 grades original Chapter 6's actor/media exercise under the corrected
// Edition 2 scope. Ownership, private request replies and actual role-tool handoffs
// still require source review alongside these observable boundary checks.
func Chapter5(bin, core, consumer, workflow string, example ConsumerCase, models map[string]string) (grade.Report, error) {
	previous, err := Chapter4(bin, core, consumer, example)
	if err != nil {
		return grade.Report{}, err
	}
	parity := grade.Check{ID: "ch4-parity", Title: "All corrected Chapter 4 checks pass", Points: 10, Earned: 10, Passed: true}
	for _, check := range previous.Checks {
		if !check.Passed {
			fail(&parity, "%s: %s", check.ID, strings.Join(check.Details, "; "))
		}
	}
	workflowChecks := exerciseWorkflow(core, workflow)
	checks := []grade.Check{parity, workflowChecks[0], replayRequests(bin), workflowChecks[1], workflowChecks[2], mediaCapabilities(bin, models), cleanHub(core)}
	return grade.NewTitledReport("Edition 2, Chapter 5 — Two Seams and a Loop", checks, ""), nil
}

// Chapter5Executable reuses the unchanged language-neutral boundary checks.
// Its 25-point report does not credit Go imports, topology, workflow or parity.
// A non-Go implementation needs separate public-consumer and ownership evidence.
func Chapter5Executable(bin string, models map[string]string) grade.Report {
	return grade.NewTitledReport("Edition 2, Chapter 5 — executable replay/media only (excludes Go consumer, workflow, topology and Chapter 4 parity)",
		[]grade.Check{replayRequests(bin), mediaCapabilities(bin, models)}, "")
}

func cleanHub(core string) grade.Check {
	check := grade.Check{ID: "hub-clean", Title: "Common depends only on standard library and first-party leaves", Points: 10, Earned: 10, Passed: true}
	base := grade.DiscoverBase(core)
	packages, err := listPackages(core, "-deps", "./internal/common")
	if err != nil {
		fail(&check, "%v", err)
		return check
	}
	found := false
	for _, pkg := range packages {
		if pkg.ImportPath == base+"/internal/common" {
			found = true
			continue
		}
		// Module names such as "ensemble" contain no dot too. Go's Standard
		// metadata establishes this boundary; spelling heuristics do not.
		if pkg.Standard {
			continue
		}
		if strings.HasPrefix(pkg.ImportPath, base+"/") && !strings.Contains(pkg.ImportPath, "/internal/") {
			// Preserve the original allowance for first-party vocabulary leaves.
			// Their transitive dependencies are still inspected by this walk.
			continue
		}
		fail(&check, "common has nonstandard implementation/dependency %s", pkg.ImportPath)
	}
	if base == "" || !found {
		fail(&check, "common package was not found")
	}
	check.Details = append(check.Details, "full spoke-to-spoke topology remains checked by corrected Chapter 4 parity; behavior placement remains source review")
	return check
}
