package gradeedition2

import (
	"path/filepath"
	"testing"
)

// TestChapter8Settings runs real optional-GUI processes and their WebSocket
// settings path, including process restart in the same disposable directory.
func TestChapter8Settings(t *testing.T) {
	dir, err := filepath.Abs("../../solutions/ch08/gui")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range settingsChecks(dir, "claude-sonnet-5") {
		if !c.Passed {
			t.Errorf("%s: %v", c.ID, c.Details)
		}
	}
}
