package ensemble

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"example.com/ensemble/internal/common"
	"example.com/ensemble/internal/skills"
)

// This explicit owner-state test seam fails one durable write, with the same
// production append/apply path. It is not a second runtime persistence owner.
type skillFailLog struct {
	common.EventLog
	kind string
}

func (l *skillFailLog) Append(e Event) error {
	if e.Type == l.kind {
		return fmt.Errorf("injected skill write failure")
	}
	return l.EventLog.Append(e)
}
func persistenceSkillConfig(t *testing.T) Config {
	t.Helper()
	return Config{DisableStreaming: true, Model: "fixture", APIKey: "fixture", Workspace: t.TempDir(), LogPath: filepath.Join(t.TempDir(), "events"), Builtins: []string{"load_skill", "unload_skill", "write_file"}, Skills: &SkillConfig{Primary: "base", Catalog: map[string][]byte{
		"base": []byte("---\nname: base\ndescription: Base\ntype: primary\nloadable-skills: edit\n---\nIdentity."),
		"edit": []byte("---\nname: edit\ndescription: Editor\ntype: loadable\ntools: write_file\n---\nUse write_file."),
	}}}
}
func TestSkillDurableCommitFailureBoundaries(t *testing.T) {
	for _, kind := range []string{"skills_changed", "tool_returned"} {
		t.Run(kind, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"content":[{"type":"tool_use","id":"c","name":"load_skill","input":{"name":"edit"}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
			}))
			defer server.Close()
			app := New(nil)
			defer app.Close()
			config := persistenceSkillConfig(t)
			config.BaseURL = server.URL
			a, err := app.NewAgent(config)
			if err != nil {
				t.Fatal(err)
			}
			before := (turnAgent{a}).SkillView()
			a.mu.Lock()
			a.log = &skillFailLog{EventLog: a.log, kind: kind}
			a.mu.Unlock()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if _, err = a.Prompt(ctx, "load edit"); err == nil {
				t.Fatal("failed persistence reported success")
			}
			after := (turnAgent{a}).SkillView()
			changed := 0
			returns := 0
			for _, e := range a.Events() {
				if e.Type == "skills_changed" {
					changed++
				}
				if e.Type == "tool_returned" {
					returns++
				}
			}
			if kind == "skills_changed" {
				if !reflect.DeepEqual(before, after) || changed != 0 {
					t.Fatal("failed transition changed authority")
				}
			} else if after.State.Revision != 1 || len(after.Material) != 2 || changed != 1 {
				t.Fatal("durable transition was rolled back after result failure")
			}
			if returns != 0 {
				t.Fatal("failed result entered history")
			}
		})
	}
}
func TestSkillInitialWriteFailureDoesNotApply(t *testing.T) {
	app := New(nil)
	defer app.Close()
	a, err := app.construct(persistenceSkillConfig(t))
	if err != nil {
		t.Fatal(err)
	}
	a.skills, err = skills.New(skillAgent{a})
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := a.skills.Prepare(common.SkillOperation{Action: "initialize", Name: "base"})
	if err != nil {
		t.Fatal(err)
	}
	a.log = &skillFailLog{kind: "skills_initialized"}
	fact := candidate.Transition()
	err = a.appendPrepared(Event{Type: "skills_initialized", Skills: &fact}, true, false, nil, candidate)
	if err == nil || a.skills.Inspect().State != nil || len(a.Events()) != 0 || a.context.SkillMode {
		t.Fatal("failed initialization published state")
	}
}
