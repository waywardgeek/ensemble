package ch09_test

// External consumer: only the new chapter's public values and actor operations.
import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"example.com/ensemble"
)

func catalog() map[string][]byte {
	return map[string][]byte{
		"base": []byte("---\nname: base\ndescription: Base\ntype: primary\ntools: read_file\nloadable-skills: edit\n---\nProject=$PROJECT.\n"),
		"dep":  []byte("---\nname: dep\ndescription: Dependency\ntype: dependency\ntools: read_file\n---\nDependency manual.\n"),
		"edit": []byte("---\nname: edit\ndescription: Edit\ntype: loadable\ntools: write_file\ndepends: dep\n---\nEdit project $PROJECT.\n"),
	}
}
func configuration(t *testing.T, name string) ensemble.Config {
	t.Helper()
	dir := t.TempDir()
	return ensemble.Config{Vendor: "openai", Model: "fixture-public", APIKey: "local-only", BaseURL: "http://127.0.0.1:1", Workspace: dir, LogPath: filepath.Join(dir, name+".jsonl"), Builtins: []string{"load_skill", "read_file", "unload_skill", "write_file"}, Skills: &ensemble.SkillConfig{Catalog: catalog(), Primary: "base", Variables: map[string]string{"PROJECT": name + "-$TOOLS"}}}
}
func agent(t *testing.T, app *ensemble.Ensemble, c ensemble.Config) *ensemble.Agent {
	t.Helper()
	a, err := app.NewAgent(c)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = a.Close() })
	return a
}
func inspection(t *testing.T, a *ensemble.Agent) ensemble.SkillInspection {
	t.Helper()
	s, err := a.InspectSkills()
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func requireCode(t *testing.T, err error, code string) {
	t.Helper()
	var skill *ensemble.SkillError
	if !errors.As(err, &skill) || skill.Code != code {
		t.Fatalf("wanted %s, got %v", code, err)
	}
}
func load(t *testing.T, a *ensemble.Agent) {
	t.Helper()
	r, err := a.LoadSkill("edit")
	if err != nil || r.Status != "loaded" || r.Revision != 1 || !r.Changed {
		t.Fatalf("load: %+v %v", r, err)
	}
}
func TestCh09PublicOwnedCopiesAndContributors(t *testing.T) {
	app := ensemble.New(io.Discard)
	defer app.Close()
	c := configuration(t, "alpha")
	a := agent(t, app, c)
	c.Skills.Catalog["base"][0] = 'X'
	c.Skills.Catalog["edit"][0] = 'X'
	c.Skills.Variables["PROJECT"] = "wrong"
	c.Builtins[0] = "wrong"
	load(t, a)
	original := inspection(t, a)
	if len(original.Material) != 3 || original.Material[0].Record.Body != "Project=alpha-$TOOLS.\n" || original.Material[2].Record.Body != "Edit project alpha-$TOOLS.\n" {
		t.Fatalf("inputs aliased or recursively expanded: %+v", original)
	}
	var read []uint64
	for _, v := range original.Contributors {
		if v.Tool == "read_file" {
			read = v.Activations
		}
	}
	if !reflect.DeepEqual(read, []uint64{1, 2}) {
		t.Fatalf("contributors: %+v", original.Contributors)
	}
	for _, name := range []string{"load_skill", "unload_skill"} {
		found := false
		for _, v := range original.Contributors {
			if v.Tool == name {
				found = v.Mandatory
			}
		}
		if !found {
			t.Fatalf("mandatory grant missing: %s", name)
		}
	}
	changed := inspection(t, a)
	changed.State.Active[0].Name = "wrong"
	changed.State.Roots[0] = "wrong"
	changed.State.Tools[0] = "wrong"
	changed.Material[0].Record.Body = "wrong"
	changed.Material[0].Record.Offers[0].Description = "wrong"
	changed.Material[2].Record.Dependencies[0] = 999
	changed.Material[2].Record.Tools[0] = "wrong"
	changed.Contributors[0].Tool = "wrong"
	for i := range changed.Contributors {
		if len(changed.Contributors[i].Activations) > 0 {
			changed.Contributors[i].Activations[0] = 999
		}
	}
	if got := inspection(t, a); !reflect.DeepEqual(got, original) {
		t.Fatal("returned inspection aliases owner storage")
	}
	state, err := a.SkillState()
	if err != nil {
		t.Fatal(err)
	}
	state.Active[0].Name = "changed again"
	if !reflect.DeepEqual(inspection(t, a), original) {
		t.Fatal("state getter aliases owner storage")
	}
	if _, err = a.UnloadSkill("edit"); err != nil {
		t.Fatal(err)
	}
	retired := inspection(t, a)
	if len(retired.State.Retired) != 2 {
		t.Fatal("dependency and root not retired")
	}
	retired.State.Retired[0].Name = "wrong"
	retired.Material[1].Retired = false
	if fresh := inspection(t, a); fresh.State.Retired[0].Name == "wrong" || !fresh.Material[1].Retired {
		t.Fatal("retired inspection aliases owner")
	}
}
func TestCh09PublicCreationConfigAtomic(t *testing.T) {
	app := ensemble.New(io.Discard)
	defer app.Close()
	a := agent(t, app, configuration(t, "config"))
	original := a.Config()
	if err := a.SetConfig(original); err != nil {
		t.Fatalf("unchanged roundtrip: %v", err)
	}
	empty := a.Config()
	empty.System = ""
	if err := a.SetConfig(empty); err != nil {
		t.Fatalf("empty retains primary: %v", err)
	}
	cases := map[string]func(*ensemble.Config){"primary": func(c *ensemble.Config) { c.Skills.Primary = "edit" }, "variables": func(c *ensemble.Config) { c.Skills.Variables["PROJECT"] = "wrong" }, "source": func(c *ensemble.Config) { c.Skills.Catalog["base"][0] = 'X' }, "mode": func(c *ensemble.Config) { c.Skills = nil }, "ceiling": func(c *ensemble.Config) { c.Builtins = append(c.Builtins, "list_directory") }, "system": func(c *ensemble.Config) { c.System = "competing" }}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			before := inspection(t, a)
			c := a.Config()
			change(&c)
			if err := a.SetConfig(c); err == nil {
				t.Fatal("creation-only mutation accepted")
			}
			if !reflect.DeepEqual(a.Config(), original) || !reflect.DeepEqual(inspection(t, a), before) {
				t.Fatal("refusal changed owner facts")
			}
		})
	}
	c := configuration(t, "competing")
	c.System = "explicit competing instruction"
	if bad, err := app.NewAgent(c); err == nil {
		bad.Close()
		t.Fatal("competing constructor System accepted")
	}
}
func TestCh09PublicTwoAgentsAndDisabled(t *testing.T) {
	app := ensemble.New(io.Discard)
	defer app.Close()
	a := agent(t, app, configuration(t, "alpha"))
	c := configuration(t, "beta")
	c.Builtins = []string{"load_skill", "read_file", "unload_skill"}
	b := agent(t, app, c)
	load(t, a)
	before := inspection(t, b)
	_, err := b.LoadSkill("edit")
	requireCode(t, err, "skill_tool_unavailable")
	if !reflect.DeepEqual(inspection(t, b), before) || before.Material[0].Record.Body != "Project=beta-$TOOLS.\n" {
		t.Fatal("agent ceiling/scalar authority leaked")
	}
	c = configuration(t, "disabled")
	c.Skills = nil
	c.Builtins = []string{"read_file"}
	plain := agent(t, app, c)
	state, err := plain.SkillState()
	if err != nil || state != nil {
		t.Fatalf("no-skills state: %v %v", state, err)
	}
	for _, operation := range []func(string) (ensemble.SkillResult, error){plain.LoadSkill, plain.UnloadSkill} {
		_, err = operation("edit")
		requireCode(t, err, "skills_disabled")
	}
}
func TestCh09PublicFrozenDirectory(t *testing.T) {
	app := ensemble.New(io.Discard)
	defer app.Close()
	c := configuration(t, "disk")
	root := filepath.Join(c.Workspace, "catalog")
	for name, body := range c.Skills.Catalog {
		dir := filepath.Join(root, name)
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	c.Skills.Catalog = nil
	c.Skills.Directory = "catalog"
	a := agent(t, app, c)
	if err := os.RemoveAll(root); err != nil {
		t.Fatal(err)
	}
	if err := a.SetConfig(a.Config()); err != nil {
		t.Fatalf("roundtrip reread catalog: %v", err)
	}
	load(t, a)
	if got := inspection(t, a).Material[2].Record.Body; got != "Edit project disk-$TOOLS.\n" {
		t.Fatalf("not frozen: %q", got)
	}
}
func TestCh09PublicPausedPublicationAndNoop(t *testing.T) {
	app := ensemble.New(io.Discard)
	defer app.Close()
	a := agent(t, app, configuration(t, "paused"))
	pause, err := a.RegisterPause()
	if err != nil {
		t.Fatal(err)
	}
	defer pause.Close()
	if _, err = pause.Update(true, false); err != nil {
		t.Fatal(err)
	}
	_, watch, err := a.Watch()
	if err != nil {
		t.Fatal(err)
	}
	defer watch.Close()
	done := make(chan error, 1)
	go func() { _, err := a.LoadSkill("edit"); done <- err }()
	select {
	case err = <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("typed control blocked behind pause")
	}
	snap, second, err := a.Watch()
	if err != nil {
		t.Fatal(err)
	}
	second.Close()
	if snap.State.Skills == nil || snap.State.Skills.Revision != 1 || !snap.State.Paused {
		t.Fatalf("success before state publication: %+v", snap.State)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	record, err := watch.Next(ctx)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(record.Observation)
	if !strings.Contains(string(raw), "skills_changed") {
		t.Fatalf("missing change observation: %s", raw)
	}
	baseline, _ := a.Dump()
	result, err := a.LoadSkill("edit")
	if err != nil || result.Changed || result.Revision != 1 {
		t.Fatalf("noop: %+v %v", result, err)
	}
	after, _ := a.Dump()
	if string(after) != string(baseline) {
		t.Fatal("public noop appended history")
	}
}
func TestCh09PublicConcurrentSingleCommit(t *testing.T) {
	app := ensemble.New(io.Discard)
	defer app.Close()
	a := agent(t, app, configuration(t, "concurrent"))
	start := make(chan struct{})
	results := make(chan ensemble.SkillResult, 16)
	failures := make(chan error, 16)
	var group sync.WaitGroup
	for range 16 {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			r, err := a.LoadSkill("edit")
			if err != nil {
				failures <- err
			} else {
				results <- r
			}
		}()
	}
	close(start)
	group.Wait()
	close(results)
	close(failures)
	for err := range failures {
		t.Fatal(err)
	}
	changes := 0
	for r := range results {
		if r.Revision != 1 {
			t.Fatal(r)
		}
		if r.Changed {
			changes++
		}
	}
	if changes != 1 {
		t.Fatalf("%d commits for concurrent same root", changes)
	}
	s := inspection(t, a)
	if len(s.Material) != 3 || s.Material[2].Record.Activation != 3 {
		t.Fatalf("candidate identity leak: %+v", s.Material)
	}
}
func TestCh09PublicCustomScalarBounds(t *testing.T) {
	cases := []struct {
		name      string
		variables map[string]string
		valid     bool
	}{{"empty", map[string]string{"PROJECT": ""}, true}, {"multiline", map[string]string{"PROJECT": "a\nb"}, true}, {"single-exact", map[string]string{"PROJECT": strings.Repeat("x", 4096)}, true}, {"single-over", map[string]string{"PROJECT": strings.Repeat("x", 4097)}, false}, {"nul", map[string]string{"PROJECT": "a\x00b"}, false}, {"reserved", map[string]string{"PROJECT": "x", "TOOLS": "wrong"}, false}, {"bad-name", map[string]string{"PROJECT": "x", "lower": "x"}, false}}
	for _, count := range []int{32, 33} {
		vars := map[string]string{"PROJECT": "x"}
		for i := 1; i < count; i++ {
			vars[fmt.Sprintf("V%d", i)] = ""
		}
		cases = append(cases, struct {
			name      string
			variables map[string]string
			valid     bool
		}{fmt.Sprintf("count-%d", count), vars, count == 32})
	}
	for _, over := range []bool{false, true} {
		vars := map[string]string{"PROJECT": strings.Repeat("x", 4096)}
		for i := 1; i < 16; i++ {
			vars[fmt.Sprintf("V%d", i)] = strings.Repeat("x", 4096)
		}
		if over {
			vars["EXTRA"] = "x"
		}
		cases = append(cases, struct {
			name      string
			variables map[string]string
			valid     bool
		}{fmt.Sprintf("aggregate-over-%v", over), vars, !over})
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			app := ensemble.New(io.Discard)
			defer app.Close()
			c := configuration(t, test.name)
			c.Skills.Variables = test.variables
			a, err := app.NewAgent(c)
			if (err == nil) != test.valid {
				t.Fatalf("valid=%v err=%v", test.valid, err)
			}
			if a != nil {
				a.Close()
			}
		})
	}
}

func TestCh09PublicAppendFrozenAuthority(t *testing.T) {
	app := ensemble.New(io.Discard)
	defer app.Close()
	c := configuration(t, "append")
	target := agent(t, app, c)
	c.LogPath = filepath.Join(c.Workspace, "donor.jsonl")
	donor := agent(t, app, c)
	load(t, donor)
	data, err := donor.Dump()
	if err != nil {
		t.Fatal(err)
	}
	var initialization, transition map[string]any
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var row map[string]any
		if err = json.Unmarshal([]byte(line), &row); err != nil {
			t.Fatal(err)
		}
		switch row["type"] {
		case "skills_initialized":
			initialization = row
		case "skills_changed":
			transition = row
		}
	}
	if initialization == nil || transition == nil {
		t.Fatal("missing durable skill facts")
	}
	original, _ := json.Marshal(transition)
	records := transition["skills"].(map[string]any)["activated"].([]any)
	changed := records[len(records)-1].(map[string]any)
	changed["body"] = "Shape-valid forged instruction."
	changed["sha256"] = fmt.Sprintf("%x", sha256.Sum256([]byte(changed["body"].(string))))
	raw, _ := json.Marshal(transition)
	var event ensemble.Event
	if err = json.Unmarshal(raw, &event); err != nil {
		t.Fatal(err)
	}
	before, _ := target.Dump()
	state := inspection(t, target)
	if err = target.Append(event); err == nil {
		t.Fatal("live append accepted body outside frozen catalog candidate")
	}
	after, _ := target.Dump()
	if string(before) != string(after) || !reflect.DeepEqual(state, inspection(t, target)) {
		t.Fatal("rejected public append mutated history or state")
	}
	if err = json.Unmarshal(original, &event); err != nil {
		t.Fatal(err)
	}
	if err = target.Append(event); err != nil {
		t.Fatalf("matching public transition refused: %v", err)
	}
	if inspection(t, target).State.Revision != 1 {
		t.Fatal("matching append did not commit")
	}
	c = configuration(t, "no-skills-append")
	c.Skills = nil
	c.Builtins = []string{"read_file"}
	plain := agent(t, app, c)
	raw, _ = json.Marshal(initialization)
	if err = json.Unmarshal(raw, &event); err != nil {
		t.Fatal(err)
	}
	before, _ = plain.Dump()
	if err = plain.Append(event); err == nil {
		t.Fatal("public append initialized exposed no-skills Agent")
	}
	after, _ = plain.Dump()
	if string(before) != string(after) {
		t.Fatal("initializer refusal changed log")
	}
}
