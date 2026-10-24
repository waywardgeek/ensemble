package ensemble_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ensemble"
)

func TestChapter14SoleGrantAndRestore(t *testing.T) {
	dir := t.TempDir()
	c9Skill(t, dir, "main", "primary", "loadable-skills: extra\n", "PRIMARY")
	c9Skill(t, dir, "extra", "loadable", "tools: write_file\n", "EXTRA-BODY")
	path := filepath.Join(t.TempDir(), "session.json")
	cfg := ensemble.Config{Model: "gpt-5.4", Vendor: ensemble.OpenAI, Surface: ensemble.ChatCompletions, BaseDir: t.TempDir(), DataDir: t.TempDir(), SavePath: path, Skills: &ensemble.SkillSpec{Directory: dir, Primary: "main"}}
	e := ensemble.New(nil)
	a, err := e.NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Skills().LoadSkill("extra"); err != nil {
		t.Fatal(err)
	}
	if !c14Has(a, "write_file") {
		t.Fatal("sole grant not enabled")
	}
	if err = a.Skills().UnloadSkill("extra"); err != nil {
		t.Fatal(err)
	}
	if c14Has(a, "write_file") {
		t.Fatal("pending unload retained compiled grant")
	}
	blocked := filepath.Join(cfg.BaseDir, "must-not-exist")
	raw, _ := json.Marshal(map[string]string{"path": blocked, "content": "not admitted"})
	if _, failed := a.Tools().Dispatch(context.Background(), "write_file", raw); !failed {
		t.Fatal("withdrawn grant dispatched")
	}
	if _, err = os.Stat(blocked); !os.IsNotExist(err) {
		t.Fatal("withdrawn tool had filesystem effect")
	}
	wire, err := a.Render()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(wire), "EXTRA-BODY") {
		t.Fatal("pending instructions disappeared before compaction")
	}
	if err = a.Save(); err != nil {
		t.Fatal(err)
	}
	if err = e.Shutdown(); err != nil {
		t.Fatal(err)
	}
	restoredRoot := ensemble.New(nil)
	defer restoredRoot.Shutdown()
	restored, err := restoredRoot.NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if c14Has(restored, "write_file") {
		t.Fatal("recovery restored a withdrawn compiled grant")
	}
	wire, err = restored.Render()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(wire), "EXTRA-BODY") {
		t.Fatal("recovery lost pending instruction survivor")
	}
	if _, failed := c3Call(t, restored, "keep_tool_results", map[string]any{}); failed {
		t.Fatal("context tool not public-dispatchable")
	}
	other, err := restoredRoot.NewAgent(ensemble.Config{Model: "offline", BaseDir: t.TempDir(), DataDir: t.TempDir(), Tools: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	if len(other.Log().Context().ActiveSkills) != 0 {
		t.Fatal("sibling inherited survivor state")
	}
}

func c14Has(a ensemble.Agent, name string) bool {
	for _, tool := range a.Tools().Definitions() {
		if tool.Name == name {
			return true
		}
	}
	return false
}
