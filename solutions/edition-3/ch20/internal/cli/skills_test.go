package cli

import (
	"os"
	"path/filepath"
	"testing"

	"ensemble"
)

func TestClientSkillSelection(t *testing.T) {
	t.Chdir(t.TempDir())
	var config ensemble.Config
	if err := ConfigureSkills(&config, "", false); err != nil || config.Skills != nil {
		t.Fatalf("plain launch without skills: %+v, %v", config.Skills, err)
	}
	if err := ConfigureSkills(&config, "", true); err == nil {
		t.Fatal("debug launch without a primary succeeded")
	}
	if err := os.MkdirAll(filepath.Join("skills", "ensemble"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("skills", "ensemble", "SKILL.md"), []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ConfigureSkills(&config, "", false); err != nil || config.Skills == nil || config.Skills.Directory != "skills" || config.Skills.Primary != "ensemble" || len(config.StartupSkills) != 0 {
		t.Fatalf("default launch: %+v, %v", config, err)
	}
	if err := ConfigureSkills(&config, "chosen", true); err != nil || config.Skills.Directory != "chosen" || len(config.StartupSkills) != 1 || config.StartupSkills[0] != "gui-debug" {
		t.Fatalf("explicit debug launch: %+v, %v", config, err)
	}
	if err := ConfigureSkills(&config, "", true); err != nil || len(config.StartupSkills) != 1 {
		t.Fatalf("duplicate startup selection: %+v, %v", config.StartupSkills, err)
	}
}

func TestClientExplicitPrimaryWins(t *testing.T) {
	t.Chdir(t.TempDir())
	selection := &ensemble.SkillSpec{Directory: "environment", Primary: "custom"}
	config := ensemble.Config{Skills: selection}
	if err := ConfigureSkills(&config, "flag", false); err != nil {
		t.Fatal(err)
	}
	if config.Skills.Directory != "flag" || config.Skills.Primary != "custom" || selection.Directory != "environment" {
		t.Fatalf("selection precedence or ownership: %+v, %+v", config.Skills, selection)
	}
}
