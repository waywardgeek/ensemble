package gradeedition2

import (
	"testing"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// The CLI, local files, actor, jobs and separate application remain real. Only
// HTTP is replaced, and the replay oracle is the body captured at that boundary.
func TestChapter5Boundaries(t *testing.T) {
	const core = "../../solutions/ch05"
	bin, cleanup, err := grade.Build(core)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	models, err := ReadMediaModels("")
	if err != nil {
		t.Fatal(err)
	}
	checks := exerciseWorkflow(core, core+"/ch06")
	checks = append(checks, replayRequests(bin), mediaCapabilities(bin, models), cleanHub(core))
	for _, check := range checks {
		if !check.Passed {
			t.Errorf("%s: %v", check.ID, check.Details)
		}
	}
}
