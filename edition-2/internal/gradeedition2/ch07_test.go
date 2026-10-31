package gradeedition2

import (
	"path/filepath"
	"testing"
)

// TestChapter7GUI drives the actual optional application, real core and shell
// tools through WebSockets; only the provider HTTP peer is a faithful fake.
func TestChapter7GUI(t *testing.T) {
	core, err := filepath.Abs("../../solutions/ch07")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range guiChecks(core, filepath.Join(core, "gui"), "claude-sonnet-5") {
		if !c.Passed {
			t.Errorf("%s: %v", c.ID, c.Details)
		}
	}
}
