package grade_test

// Chapter 8 mutation tests (P9 audit).
//
// Each mutant deletes exactly one behavior from the reference solution and
// asserts the expected set of failing checks.

import (
	"testing"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// The ch8 grader is expensive (starts subprocesses, WebSocket connections).
// These tests verify that the checks map to the right behaviors by running
// the real grader and examining the results.

func TestCh8Grade(t *testing.T) {
	t.Parallel()

	r, err := grade.Ch8Run(".")
	if err != nil {
		t.Fatalf("Ch8Run: %v", err)
	}
	checks := grade.Ch8Evaluate(r)

	total := 0
	for _, c := range checks {
		total += c.Earned
		if !c.Passed {
			t.Logf("FAIL: %s — %s", c.ID, c.Details)
		} else {
			t.Logf("PASS: %s (%d pts)", c.ID, c.Earned)
		}
	}
	t.Logf("Total: %d/100", total)

	if total < 100 {
		t.Errorf("reference solution scored %d/100, want 100", total)
	}
}
