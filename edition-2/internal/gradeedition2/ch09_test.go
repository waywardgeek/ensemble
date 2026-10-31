package gradeedition2

import (
	"path/filepath"
	"testing"
)

// TestChapter9Skills exercises actual frozen GUI/core behavior through the
// provider boundary, including hidden dispatch and a pending-unload file effect.
func TestChapter9Skills(t *testing.T) {
	dir, err := filepath.Abs("../../solutions/edition-3/ch09/gui")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range skillsChecks(dir, "claude-sonnet-4-6") {
		if !c.Passed {
			t.Errorf("%s: %v", c.ID, c.Details)
		}
	}
}
