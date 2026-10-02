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

// onlyTools is a skill filter that enables exactly the names it lists. It
// embeds the interface so any method the test does not expect panics.
type onlyTools struct {
	common.Skills
	names map[string]bool
}

func (o onlyTools) IsToolEnabled(name string) bool { return o.names[name] }

// The registry keys tools by a normalized name ("view_gui" -> "viewgui"), but
// a skill lists the real one. Asking the filter about the key, as Declarations
// once did, silently dropped every non-builtin tool whose name has an
// underscore: view_gui shipped registered, listed in the primary skill, and
// absent from every request.
func TestSkillFilterUsesTheDeclaredName(t *testing.T) {
	r := NewBareRegistry()
	r.SetSkillRegistry(onlyTools{names: map[string]bool{"view_gui": true}})
	r.RegisterInitial(common.Tool{
		Name:   "view_gui",
		Schema: json.RawMessage(`{"type":"object","properties":{}}`),
		Run:    func(*common.Call, json.RawMessage) (string, error) { return "", nil },
	})
	for name, decls := range map[string][]common.ToolDecl{
		"Declarations":        r.Declarations(),
		"InitialDeclarations": r.InitialDeclarations(),
	} {
		if len(decls) != 1 || decls[0].Name != "view_gui" {
			t.Errorf("%s = %v, want exactly view_gui", name, decls)
		}
	}
}
