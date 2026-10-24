package gui

import (
	"ensemble"
	"os"
	"path/filepath"
	"testing"
)

func TestSkillStateTracksPendingAndLoadable(t *testing.T) {
	dir := t.TempDir()
	for name, contents := range map[string]string{"base": "---\nname: base\ntype: primary\ndescription: base\nloadable-skills: next\n---\nBASE", "next": "---\nname: next\ntype: loadable\ndescription: next\n---\nNEXT"} {
		path := filepath.Join(dir, name, "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	app := ensemble.New(nil)
	defer app.Shutdown()
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "fixture", BaseDir: t.TempDir(), DataDir: t.TempDir(), Skills: &ensemble.SkillSpec{Directory: dir, Primary: "base"}})
	if err != nil {
		t.Fatal(err)
	}
	f := skillFrame(a)
	if f.Type != "skill_state" || f.AgentID != a.ID() || len(f.LoadableSkills) != 1 {
		t.Fatalf("initial frame: %+v", f)
	}
	if err := a.Skills().LoadSkill("next"); err != nil {
		t.Fatal(err)
	}
	if err := a.Skills().UnloadSkill("next"); err != nil {
		t.Fatal(err)
	}
	f = skillFrame(a)
	if len(f.LoadableSkills) != 0 {
		t.Fatal("pending still counts as loaded")
	}
	for _, s := range f.Skills {
		if s.Name == "next" && (!s.Enabled || !s.PendingUnload) {
			t.Fatalf("pending: %+v", s)
		}
	}
}
