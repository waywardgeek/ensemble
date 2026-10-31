package ensemble

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func ch10AdmissionCatalog(t *testing.T) SessionOptions {
	t.Helper()
	w := t.TempDir()
	o := SessionOptions{Config: Config{Vendor: "openai", Model: "fixture", APIKey: "admission-noncredential", BaseURL: "http://127.0.0.1:1", Workspace: w, DataDir: filepath.Join(w, "session")}}
	names, deps := []string{}, []string{}
	for i := 0; i < 128; i++ {
		names = append(names, fmt.Sprintf("offer%03d", i))
	}
	for i := 0; i < 126; i++ {
		deps = append(deps, fmt.Sprintf("dep%03d", i))
	}
	catalog := map[string][]byte{
		"base": []byte("---\nname: base\ndescription: Base\ntype: primary\nloadable-skills: bulk offer000 offer001\n---\nBase."),
		"bulk": []byte("---\nname: bulk\ndescription: Bulk\ntype: loadable\ndepends: " + strings.Join(deps, " ") + "\nloadable-skills: " + strings.Join(names, " ") + "\n---\n" + strings.Repeat("${PAD}", 16)),
	}
	for _, name := range names {
		catalog[name] = []byte("---\nname: " + name + "\ndescription: \"" + strings.Repeat("<", 256) + "\"\ntype: loadable\n---\n")
	}
	for _, name := range deps {
		catalog[name] = []byte("---\nname: " + name + "\ndescription: Dependency\ntype: dependency\nloadable-skills: " + strings.Join(names, " ") + "\n---\n" + strings.Repeat("${PAD}", 16))
	}
	o.Config.Builtins = []string{"load_skill", "unload_skill"}
	o.Config.Skills = &SkillConfig{Primary: "base", Catalog: catalog, Variables: map[string]string{"PAD": strings.Repeat("<", 4096)}}
	return o
}

func TestCh10AdmissionSkillsVersusStorage(t *testing.T) {
	root := New(io.Discard)
	t.Cleanup(func() { _ = root.Close() })
	o := ch10AdmissionCatalog(t)
	a, err := root.OpenSession(o)
	if err != nil {
		t.Fatal(err)
	}
	before, err := a.InspectSkills()
	if err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(o.Config.DataDir, "events.log")
	original, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	_, err = a.LoadSkill("bulk")
	var refusal *SkillError
	if !errors.As(err, &refusal) || refusal.Code != "skill_too_large" || !strings.Contains(refusal.Detail, "record") {
		t.Fatalf("expected controlled whole-record preflight refusal: %v", err)
	}
	after, err := a.InspectSkills()
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(log)
	if err != nil || !bytes.Equal(got, original) || !reflect.DeepEqual(before, after) {
		t.Fatal("skill preflight changed authority or durable bytes")
	}
	if _, err = a.LoadSkill("offer000"); err != nil {
		t.Fatalf("controlled skill preflight poisoned admission: %v", err)
	}
	state, err := a.SkillState()
	if err != nil || state.Revision != 1 || len(state.Active) != 2 || state.Active[1].Activation != 2 {
		t.Fatal("controlled preflight consumed revision/activation")
	}
	if _, err = a.Checkpoint(); err != nil {
		t.Fatal(err)
	}
	checkpoint := filepath.Join(o.Config.DataDir, "checkpoint.json")
	priorCheckpoint, err := os.ReadFile(checkpoint)
	if err != nil {
		t.Fatal(err)
	}
	priorState := a.Snapshot()
	priorLog, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	// The same healthy session now attempts a genuinely oversized ordinary
	// encoded record. Earlier decoded Skills bounds have not been enlarged.
	err = a.Append(Event{Type: "message_received", Message: &Entry{Actor: "system", Purpose: "instruction", Parts: []Part{Text(strings.Repeat("x", 64<<20))}}})
	ch10RemainingCode(t, err, "session_limit")
	if !reflect.DeepEqual(priorState, a.Snapshot()) {
		t.Fatal("storage overflow changed accepted state")
	}
	got, err = os.ReadFile(log)
	if err != nil || !bytes.Equal(got, priorLog) {
		t.Fatal("storage overflow appended bytes")
	}
	if _, err = a.LoadSkill("offer001"); err == nil {
		t.Fatal("terminal storage overflow retained skill admission")
	}
	ch10RemainingCode(t, a.Close(), "session_limit")
	got, err = os.ReadFile(checkpoint)
	if err != nil || !bytes.Equal(got, priorCheckpoint) {
		t.Fatal("terminal storage overflow replaced checkpoint")
	}
	if _, err = root.InspectSession(o.Config.DataDir); err != nil {
		t.Fatalf("unchanged complete prefix cannot be inspected: %v", err)
	}
}
