package gradeedition2

import (
	"path/filepath"
	"testing"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// TestChapter12GUIDebug exercises both actual startup clients and external peers.
func TestChapter12GUIDebug(t *testing.T) {
	core, err := filepath.Abs("../../solutions/edition-3/ch12")
	if err != nil {
		t.Fatal(err)
	}
	bin, cleanup, err := grade.Build(filepath.Join(core, "cmd", "ch01"))
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	for _, c := range guiDebugChecks(bin, filepath.Join(core, "gui"), "claude-sonnet-4-6") {
		if !c.Passed {
			t.Errorf("%s: %v", c.ID, c.Details)
		}
	}
}
