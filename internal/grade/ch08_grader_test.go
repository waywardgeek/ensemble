package grade_test

// Chapter 8 mutation tests (P9 audit).
//
// Each mutant deletes exactly one behavior from the reference solution and
// asserts the expected set of failing checks.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// The ch8 grader is expensive (starts subprocesses, WebSocket connections).
// These tests verify that the checks map to the right behaviors by running
// the real grader and examining the results.

func TestCh8Grade(t *testing.T) {
	// Deliberately NOT t.Parallel(). This grader starts subprocesses and
	// WebSocket connections, and grader runs that overlap produce false scores:
	// ch16 once measured 11/100 under a concurrent sweep and 100/100 alone. A
	// number measured alongside other graders is not a measurement.

	// Ch8Run wants the agent module directory. This passed ".", which under
	// `go test` is the package directory internal/grade, not an agent tree, so
	// DiscoverBase found no go.mod and every check collapsed. Stale since
	// 34ef50b ("Unify grading: every chapter grades one directory").
	//
	// The canonical target is ./agent, per scripts/gradesweep.sh: ch1-ch4
	// grade solutions/chNN, ch6 alone grades a frozen snapshot, and ch5 plus
	// ch7-ch17 grade the live tree. solutions/ch08 also scores 100 here, but it
	// is not the canonical target and grading it would let the live tree rot
	// unnoticed.
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate this test file")
	}
	target := filepath.Join(filepath.Dir(thisFile), "..", "..", "agent")
	if _, err := os.Stat(filepath.Join(target, "go.mod")); err != nil {
		t.Fatalf("ch8 target %s has no go.mod: %v", target, err)
	}

	r, err := grade.Ch8Run(target)
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

// Replayed IDs occupy a different namespace from live IDs. A numeric-only
// decoder silently dropped every replay final before the grader could see it.
func TestCh8ReadsLiveAndReplayFinals(t *testing.T) {
	for _, id := range []string{`4`, `"r9.0"`} {
		var msg grade.Ch8WsMsg
		wire := `{"type":"part_final","part_id":` + id + `,"seq":9,"text":"finished"}`
		if err := json.Unmarshal([]byte(wire), &msg); err != nil {
			t.Fatal(err)
		}
		if msg.Type != "part_final" || msg.Text != "finished" || msg.Seq != 9 || string(msg.PartID) != id {
			t.Fatalf("final lost or misidentified: %+v", msg)
		}
	}
}
