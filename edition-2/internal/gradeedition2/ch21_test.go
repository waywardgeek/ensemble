package gradeedition2

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// TestStructuredManagement23 exercises corrected submission, owned durable
// retrieval and exact GUI replay through the real optional application.
func TestStructuredManagement23(t *testing.T) {
	bin, cleanup, err := grade.Build("../../solutions/ch21/gui")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	checks := grade.Ch26Checks(grade.Ch26Result{})
	for i := range checks {
		checks[i].Passed = true
		checks[i].Details = nil
	}
	managementGUI(bin, checks)
	for _, i := range []int{0, 1, 2, 3, 5} {
		if !checks[i].Passed {
			t.Errorf("%s: %v", checks[i].ID, checks[i].Details)
		}
	}
}

// An opt-in real executable keeps mutation runs focused on changed management
// observations, without rerunning unchanged recursive CLI and idle-correction checks.
func TestManagement23(t *testing.T) {
	bin := os.Getenv("EDITION2_MANAGEMENT_BIN")
	if bin == "" {
		t.Skip("set EDITION2_MANAGEMENT_BIN to an actual optional GUI executable")
	}
	checks := grade.Ch26Checks(grade.Ch26Result{})
	for i := range checks {
		checks[i].Passed = true
		checks[i].Earned = checks[i].Points
		checks[i].Details = nil
	}
	managementGUI(bin, checks)
	data, _ := json.Marshal(checks)
	fmt.Println(string(data))
	for _, i := range []int{0, 1, 2, 3, 5} {
		if !checks[i].Passed {
			t.Errorf("%s: %v", checks[i].ID, checks[i].Details)
		}
	}
}

// A recalled submit argument is not evidence that the submission/result was restored.
func TestManagementReplayText23(t *testing.T) {
	for _, event := range []map[string]any{
		{"type": "part_final", "part": map[string]any{"type": "text", "text": "RESTORED"}},
		{"type": "tool_finished", "result": "RESTORED"},
		{"type": "turn_ended", "text": "RESTORED"},
	} {
		if !managementContains([]map[string]any{event}, "RESTORED") {
			t.Fatalf("valid rendered event rejected: %v", event)
		}
	}
	argument := map[string]any{"type": "part_final", "part": map[string]any{"type": "tool_call", "args": map[string]any{"result": "RESTORED"}}}
	if managementContains([]map[string]any{argument}, "RESTORED") {
		t.Fatal("tool-call argument falsely proves restored result")
	}
}

// TestManagementRootBuildFailure23 prevents an unexecuted root-chat predicate
// from earning the combined regression row when only the core can be built.
func TestManagementRootBuildFailure23(t *testing.T) {
	bin, cleanup, err := grade.Build("../../solutions/ch21")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	report, err := Chapter21(bin, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range report.Checks {
		if c.ID == "composition-root" && !c.Passed {
			t.Fatalf("valid core composition control failed: %v", c.Details)
		}
		if c.ID == "submit-idle-regression" {
			if c.Passed || c.Earned != 0 || len(c.Details) != 1 || !strings.Contains(c.Details[0], "go build") {
				t.Fatalf("unexecuted root-chat predicate earned credit: %+v", c)
			}
			return
		}
	}
	t.Fatal("combined submit/idle/root-chat row absent")
}
