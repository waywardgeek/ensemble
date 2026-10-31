package gradeedition2

import (
	"path/filepath"
	"testing"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// TestChapter11MCP exercises real commands, Agent/jobs, subprocess tools and GUI
// transport with only external protocol/provider peers replaced by small fakes.
func TestChapter11MCP(t *testing.T) {
	core, err := filepath.Abs("../../solutions/edition-3/ch11")
	if err != nil {
		t.Fatal(err)
	}
	bin, cleanup, err := grade.Build(filepath.Join(core, "cmd", "ch01"))
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	for _, c := range mcpChecks(bin, filepath.Join(core, "gui"), "claude-sonnet-4-6") {
		if !c.Passed {
			t.Errorf("%s: %v", c.ID, c.Details)
		}
	}
}
