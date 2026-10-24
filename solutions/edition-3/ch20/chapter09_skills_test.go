package ensemble_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ensemble"
)

func c9Skill(t *testing.T, dir, name, kind, meta, body string) {
	t.Helper()
	path := filepath.Join(dir, name, "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("---\nname: "+name+"\ntype: "+kind+"\ndescription: "+name+" summary\n"+meta+"---\n"+body), 0600); err != nil {
		t.Fatal(err)
	}
}
func c9Agent(t *testing.T, dir string, vars map[string]func() string) ensemble.Agent {
	t.Helper()
	app := ensemble.New(nil)
	t.Cleanup(func() { _ = app.Shutdown() })
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.Anthropic, Model: "claude-sonnet-4-6", BaseDir: t.TempDir(), DataDir: t.TempDir(), Skills: &ensemble.SkillSpec{Directory: dir, Primary: "base", Variables: vars}})
	if err != nil {
		t.Fatal(err)
	}
	return a
}
func c9Fixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	c9Skill(t, dir, "base", "primary", "tools: read_file\nloadable-skills: writer broken cycle dep missing\n", "BASE $TOOLS $SKILLS $CUSTOM_VAR $UNKNOWN")
	c9Skill(t, dir, "dep", "dependency", "tools:\n - write_file\n", "DEPENDENCY")
	c9Skill(t, dir, "writer", "loadable", "depends: dep\ntools: read_file\n", "WRITER $CUSTOM_VAR")
	c9Skill(t, dir, "broken", "loadable", "depends: dep absent\n", "BROKEN")
	c9Skill(t, dir, "cycle", "loadable", "depends: cycle\n", "CYCLE")
	c9Skill(t, dir, "hidden", "loadable", "tools: list_directory\n", "HIDDEN")
	return dir
}
func TestChapter09ActivationAndIsolation(t *testing.T) {
	dir := c9Fixture(t)
	value := "initial"
	a := c9Agent(t, dir, map[string]func() string{"CUSTOM_VAR": func() string { return value }})
	b := c9Agent(t, dir, nil)
	before := a.Skills().Snapshot()
	if !strings.Contains(before.Instructions, "initial $UNKNOWN") || strings.Contains(before.Instructions, "DEPENDENCY") {
		t.Fatalf("startup instructions: %s", before.Instructions)
	}
	for _, s := range a.Skills().LoadableSkills() {
		if s.Name == "dep" || s.Name == "hidden" {
			t.Fatalf("not directly loadable: %s", s.Name)
		}
	}
	for _, name := range []string{"hidden", "dep", "base", "broken", "cycle"} {
		if err := a.Skills().LoadSkill(name); err == nil {
			t.Fatalf("load %q should fail", name)
		}
	}
	if got := a.Skills().Snapshot(); len(got.Definitions) != len(before.Definitions) || got.Instructions != before.Instructions {
		t.Fatal("failed graph leaked capability")
	}
	path := filepath.Join(a.Config().BaseDir, "file.txt")
	_, failed := c3Call(t, a, "write_file", map[string]any{"path": path, "content": "before"})
	if !failed {
		t.Fatal("disabled writer dispatched")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("guessed writer touched file: %v", err)
	}
	value = "later"
	a.RegisterVar("REENTRANT", func() string {
		_ = a.Skills().Snapshot()
		a.RegisterVar("OTHER", func() string { return "ok" })
		return "safe"
	})
	notifications := 0
	cancel := a.Skills().SubscribeSkills(func() { notifications++; _ = a.Skills().Snapshot() })
	defer cancel()
	if err := a.Skills().LoadSkill("writer"); err != nil {
		t.Fatal(err)
	}
	after := a.Skills().Snapshot()
	if !strings.Contains(after.Instructions, "WRITER later") || !strings.Contains(after.Instructions, "initial $UNKNOWN") {
		t.Fatalf("render retention: %s", after.Instructions)
	}
	for _, tool := range after.Tools {
		if tool.Name == "read_file" && (tool.Origin != "initial" || tool.EventSeq != 0 || tool.Skill != "base") {
			t.Fatalf("shared tool rebased: %+v", tool)
		}
		if tool.Name == "write_file" && (!tool.Enabled || tool.Skill != "dep" || tool.Origin != "manual") {
			t.Fatalf("dependency tool: %+v", tool)
		}
	}
	if _, failed = c3Call(t, a, "write_file", map[string]any{"path": path, "content": "after"}); failed {
		t.Fatal("enabled real writer failed")
	}
	if text := c3Contents(t, path); text != "after" {
		t.Fatal(text)
	}
	if err := a.Skills().UnloadSkill("writer"); err != nil {
		t.Fatal(err)
	}
	pending := a.Skills().Snapshot()
	if pending.Instructions != after.Instructions || len(pending.Definitions) != len(after.Definitions) {
		t.Fatal("pending unload removed capability")
	}
	if notifications != 2 {
		t.Fatalf("notifications: %d", notifications)
	}
	if err := a.Skills().LoadSkill("writer"); err != nil {
		t.Fatal(err)
	}
	if a.Skills().Snapshot().Instructions != after.Instructions {
		t.Fatal("duplicate load rerendered bodies")
	}
	if strings.Contains(b.Skills().Snapshot().Instructions, "WRITER") {
		t.Fatal("cross-Agent activation")
	}
}
func TestChapter09EnvironmentAndStartupRenderer(t *testing.T) {
	dir := c9Fixture(t)
	inputs := map[string]string{"EN_SKILLS_DIR": dir, "EN_PRIMARY_SKILL": "base", "EN_CUSTOM_VAR": "explicit"}
	cfg, err := ensemble.ConfigFromEnv(func(key string) string { return inputs[key] })
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Skills == nil || cfg.Skills.Directory != dir || cfg.Skills.Variables["CUSTOM_VAR"]() != "explicit" {
		t.Fatal("environment skill inputs")
	}
	app := ensemble.New(nil)
	defer app.Shutdown()
	cfg.Vendor = ensemble.OpenAI
	cfg.Model = "fixture"
	cfg.APIKey = ""
	cfg.BaseURL = ""
	cfg.BaseDir = t.TempDir()
	cfg.DataDir = t.TempDir()
	cfg.Skills.Variables["ROOT_QUERY"] = func() string { _ = app.Agents(); return "safe" }
	if _, err := app.NewAgent(cfg); err != nil {
		t.Fatal(err)
	}
}
