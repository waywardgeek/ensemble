package ensemble_test

import (
	"strings"
	"testing"

	"ensemble"
)

func TestStartupSkillAdmissionAndInstructions(t *testing.T) {
	dir := c9Fixture(t)
	config := ensemble.Config{Model: "fixture", BaseDir: t.TempDir(), DataDir: t.TempDir(), Skills: &ensemble.SkillSpec{Directory: dir, Primary: "base"}, StartupSkills: []string{"writer"}}
	app := ensemble.New(nil)
	t.Cleanup(func() { _ = app.Shutdown() })
	a, err := app.NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}
	config.StartupSkills[0] = "missing"
	copy := a.Config()
	copy.StartupSkills[0] = "other"
	if a.Config().StartupSkills[0] != "writer" {
		t.Fatal("startup configuration aliases caller memory")
	}
	prompt, err := a.Engine().EffectiveSystemPrompt(a.Log().Context())
	if err != nil || !strings.Contains(prompt, "WRITER") || !strings.Contains(prompt, "BASE") {
		t.Fatalf("startup SYSTEM: %q, %v", prompt, err)
	}
	for _, skill := range a.Skills().ListSkills() {
		if skill.Name == "writer" && (!skill.Enabled || skill.Origin != "initial" || skill.EventSeq != 0) {
			t.Fatalf("startup provenance: %+v", skill)
		}
	}
	hasWriter := func() bool {
		for _, tool := range a.Skills().Snapshot().Tools {
			if tool.Name == "write_file" && tool.Enabled {
				return true
			}
		}
		return false
	}
	if !hasWriter() {
		t.Fatal("startup skill tool not enabled")
	}
	if err := a.Skills().UnloadSkill("writer"); err != nil {
		t.Fatal(err)
	}
	if !hasWriter() || !strings.Contains(a.Skills().Snapshot().Instructions, "WRITER") {
		t.Fatal("pending unload removed compiled capability or instructions")
	}
}

func TestStartupSkillMustBeOffered(t *testing.T) {
	dir := c9Fixture(t)
	app := ensemble.New(nil)
	t.Cleanup(func() { _ = app.Shutdown() })
	_, err := app.NewAgent(ensemble.Config{Model: "fixture", BaseDir: t.TempDir(), DataDir: t.TempDir(), Skills: &ensemble.SkillSpec{Directory: dir, Primary: "base"}, StartupSkills: []string{"hidden"}})
	if err == nil || !strings.Contains(err.Error(), "not offered") {
		t.Fatalf("startup bypassed admission: %v", err)
	}
}
