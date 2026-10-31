package gradeedition2

import (
	"runtime"
	"testing"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// TestChapter19Sandbox exercises both real CLI policies and an outside public
// consumer. Seatbelt is the actual boundary, so other platforms skip this test
// rather than replacing confinement with a test implementation.
func TestChapter19Sandbox(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS Seatbelt exercise")
	}
	core := "../../solutions/edition-3/ch19"
	bin, clean, err := grade.Build(core + "/cmd/ch01")
	if err != nil {
		t.Fatal(err)
	}
	defer clean()
	report, err := Chapter19CLI(bin, core)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range report.Checks {
		if !c.Passed {
			t.Errorf("%s: %v", c.ID, c.Details)
		}
	}
}
