package gradeedition2

import (
	"github.com/waywardgeek/ensemble/internal/grade"
	"os"
	"testing"
)

// TestChapter20RootChat checks two completed real GUI turns with fleet enabled,
// so a root missing-submit loop cannot be concealed by early process EOF.
func TestChapter20RootChat(t *testing.T) {
	bin, cleanup, err := grade.Build("../../solutions/edition-3/ch20/gui/cmd")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	c := grade.Ch25Checks(grade.Ch25Result{})[6]
	c.Passed = true
	c.Details = nil
	fleetRootChatWith(bin, &c, true)
	if !c.Passed {
		t.Fatal(c.Details)
	}
}

// TestChapter20Structured runs the scored schema, validation, JSON fan-in and
// durable-row predicates through a real submitted CLI and its actual children.
func TestChapter20Structured(t *testing.T) {
	bin, cleanup, err := grade.Build("../../solutions/edition-3/ch20/cmd/ch02")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	checks := grade.Ch25Checks(grade.Ch25Result{})
	for i := range checks {
		checks[i].Passed = true
		checks[i].Details = nil
	}
	fleetStructuredWith(fleetDriver{bin: bin, current: true}, checks)
	for _, i := range []int{0, 1, 3} {
		if !checks[i].Passed {
			t.Errorf("%s: %v", checks[i].ID, checks[i].Details)
		}
	}
}

// These opt-in checks launch the real submitted binary; only its remote provider
// is replaced. They also permit focused restored-defect checks without GUI work.
func TestFleet22(t *testing.T) {
	bin := os.Getenv("EDITION2_FLEET_BIN")
	if bin == "" {
		t.Skip("set EDITION2_FLEET_BIN to a built submitted CLI")
	}
	checks := grade.Ch25Checks(grade.Ch25Result{})
	for i := range checks {
		checks[i].Passed = true
		checks[i].Earned = checks[i].Points
		checks[i].Details = nil
	}
	fleetChecks(bin, checks)
	for _, c := range checks {
		if !c.Passed {
			t.Errorf("%s: %v", c.ID, c.Details)
		}
	}
}

// The outside consumer and deployed GUI remain real programs. This entry point
// isolates their affected checks from the terminal fleet during sensitivity work.
func TestReuse22(t *testing.T) {
	core := os.Getenv("EDITION2_REUSE_CORE")
	if core == "" {
		t.Skip("set EDITION2_REUSE_CORE to a submitted core module")
	}
	checks := grade.Ch24Checks(grade.Ch24Result{})
	for i := range checks {
		checks[i].Passed = true
		checks[i].Earned = checks[i].Points
		checks[i].Details = nil
	}
	reuseChecks(core, core+"/gui", false, checks)
	for _, c := range checks {
		if !c.Passed {
			t.Errorf("%s: %v", c.ID, c.Details)
		}
	}
}
