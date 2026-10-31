package gradeedition2

import (
	"testing"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// TestChapter3AcceptedSolution runs the real CLI, filesystem and tool loop.
// Inherited parity also exercises the capture fake's terminating reply.
func TestChapter3AcceptedSolution(t *testing.T) {
	bin, cleanup, err := grade.Build("../../solutions/ch03")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	report, err := Chapter3(bin)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Passed || report.Score != 100 || report.MaxScore != 100 || len(report.Checks) != 10 {
		t.Fatalf("corrected Chapter 3: %+v", report)
	}
}
