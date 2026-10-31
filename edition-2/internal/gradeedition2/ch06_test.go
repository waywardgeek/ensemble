package gradeedition2

import (
	"github.com/waywardgeek/ensemble/internal/grade"
	"testing"
)

// TestChapter6Streaming exercises actual core and separate public shape clients
// against the same provider replies, preserving real actor/tool/parse behavior.
func TestChapter6Streaming(t *testing.T) {
	const core = "../../solutions/ch06"
	bin, cleanup, err := grade.Build(core)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	for _, check := range streamingChecks(bin, core, core+"/ch07", "claude-sonnet-5") {
		if !check.Passed {
			t.Errorf("%s: %v", check.ID, check.Details)
		}
	}
}
