package grade

import (
	"path/filepath"
	"runtime"
	"testing"
)

// TestLiveAgentIsAStar enforces the chapter 6 star topology on the LIVE agent/
// tree, which no grader checks.
//
// Why this exists as a test rather than a grader check. Graders grade student
// trees, and each chapter's grader runs against its canonical frozen target
// (the list lives in scripts/gradesweep.sh); ch6's is solutions/ch06/agent. So
// the architectural invariant the book's central argument rests on was enforced
// only on a snapshot frozen at chapter 6, while agent/ -- the code a reader
// actually reads, and the code we keep changing -- was governed by nothing.
//
// That gap is not theoretical. The 2026-09-27 refactor that pulled skills and
// settings out of internal/common could have introduced a spoke-to-spoke import
// and every grader would have stayed green. It was caught by the compiler and
// by hand, not by the suite.
//
// This shares starViolations with the ch6 check on purpose. Two copies of an
// invariant drift, and the copy nobody reads is the one that rots.
func TestLiveAgentIsAStar(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate this test file")
	}
	agentDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "agent")

	base := DiscoverBase(agentDir)
	if base == "" {
		t.Fatalf("cannot determine module path for %s", agentDir)
	}

	graph := DiscoverImportGraph(agentDir)
	spokes, violations := starViolations(graph, base)

	// DiscoverImportGraph returns an empty map when `go list` fails, and an
	// empty graph makes the scan vacuously clean. Without this guard a tree
	// that does not compile would pass silently, which is the failure mode this
	// test exists to prevent.
	if len(spokes) == 0 {
		t.Fatalf("no spoke packages found under %s/internal/ (graph has %d entries): "+
			"the import graph is empty or unreadable, so star topology was never "+
			"actually checked", base, len(graph))
	}

	for _, v := range violations {
		t.Errorf("star topology violated: %s", v)
	}
	if len(violations) > 0 {
		t.Fatalf("%d spoke-to-spoke import(s) in agent/. A spoke may import only "+
			"internal/common. If two spokes need to share behavior, invert it: put "+
			"the interface in common and let the composition root in cmd/ wire the "+
			"implementation, the way skills and settings were extracted.", len(violations))
	}

	t.Logf("agent/ is a clean star: %d spokes around internal/common: %v", len(spokes), spokes)
}
