package gradeedition2

import (
	"path/filepath"
	"testing"
)

// TestChapter13Speech runs the shipped executable harness, real GUI and renderer.
// Its provider and synthesis boundaries are fakes; native listening is separate.
func TestChapter13Speech(t *testing.T) {
	core, err := filepath.Abs("../../solutions/edition-3/ch13")
	if err != nil {
		t.Fatal(err)
	}
	for _, check := range speechChecks(core) {
		if !check.Passed {
			t.Errorf("%s: %v", check.ID, check.Details)
		}
	}
}
