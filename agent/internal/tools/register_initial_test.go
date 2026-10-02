package tools

import (
	"encoding/json"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// A tool the host builds after the registry (view_gui needs the GUI hub) is
// still part of the agent as created. Stamped Dynamic, it would drop out of
// the cache-stable initial declarations and read as if a skill brought it.
func TestRegisterInitialIsInitialProvenance(t *testing.T) {
	r := NewBareRegistry()
	r.RegisterInitial(common.Tool{
		Name:   "view_gui",
		Schema: json.RawMessage(`{"type":"object","properties":{}}`),
		Run:    func(*common.Call, json.RawMessage) (string, error) { return "", nil },
	})
	has := func(decls []common.ToolDecl) bool {
		for _, d := range decls {
			if d.Name == "view_gui" {
				return true
			}
		}
		return false
	}
	if !has(r.InitialDeclarations()) {
		t.Fatal("view_gui missing from InitialDeclarations")
	}
	if has(r.DynamicDeclarations()) {
		t.Fatal("view_gui listed in DynamicDeclarations, as if a skill had loaded it")
	}
	if tool, err := r.Lookup("view_gui"); err != nil || tool.NoJob {
		t.Fatalf("view_gui must be registered and run as a job (err=%v, NoJob=%v)", err, tool.NoJob)
	}
}
