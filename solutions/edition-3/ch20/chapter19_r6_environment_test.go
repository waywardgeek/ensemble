package ensemble_test

import (
	"ensemble"
	"ensemble/internal/common"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCh19R6HiddenSkillCredentialEnvironment(t *testing.T) {
	const authName = "R6_WEB_AUTH"
	t.Setenv(authName, "harmless-web-fixture-value")
	t.Setenv("R6_BUILD_SETTING", "ordinary-build-value")
	t.Setenv("GOFLAGS", "-mod=readonly -p=1")
	dir := t.TempDir()
	skills := filepath.Join(dir, "skills")
	for name, doc := range map[string]string{
		"primary": "---\nname: primary\ntype: primary\ndescription: local fixture\nloadable-skills: web-search\n---\nLocal policy fixture.",
		"web":     "---\nname: web-search\ntype: loadable\ndescription: local web metadata\ntools: r6_lookup\nmcp_servers:\n  - name: r6-web\n    transport: url\n    url: http://127.0.0.1:1/mcp\n    auth-env: " + authName + "\n---\nMetadata only; never activated.",
	} {
		path := filepath.Join(skills, name, "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(doc), 0600); err != nil {
			t.Fatal(err)
		}
	}
	app := ensemble.New(nil)
	defer app.Shutdown()
	a, err := app.NewAgent(ensemble.Config{Model: "offline", AuthMode: ensemble.AuthAPI, BaseURL: "http://127.0.0.1:1", BaseDir: dir, DataDir: filepath.Join(dir, "data"), EnableWebSearch: false, Skills: &ensemble.SkillSpec{Directory: skills, Primary: "primary"}})
	if err != nil {
		t.Fatal(err)
	}
	if a.Config().SandboxRoot != "" || len(a.Config().MCPServers) != 0 {
		t.Fatal("fixture must be unconfined with skill-only credential metadata")
	}
	for _, state := range []struct {
		name string
		web  bool
	}{{"disabled-initial", false}, {"enabled-unloaded", true}, {"disabled-again", false}} {
		t.Run(state.name, func(t *testing.T) {
			a.SetSecurityPolicy(false, state.web)
			visible := false
			for _, skill := range a.Skills().ListSkills() {
				if skill.Name == "web-search" {
					visible = true
					if skill.Enabled {
						t.Error("web skill was activated")
					}
				}
			}
			if visible != state.web {
				t.Errorf("web visibility=%t want %t", visible, state.web)
			}
			snapshot := a.Skills().Snapshot()
			for _, tool := range snapshot.Definitions {
				if tool.Name == "r6_lookup" {
					t.Error("unloaded web tool exposed")
				}
			}
			for _, tool := range snapshot.Tools {
				if tool.Name == "r6_lookup" && tool.Enabled {
					t.Error("unloaded web tool enabled")
				}
			}
			for _, skill := range snapshot.Skills {
				if skill.Name == "web-search" && !state.web {
					t.Error("disabled web declaration exposed in snapshot")
				}
			}
			env := map[string]string{}
			for _, item := range common.ProcessEnvironment(a) {
				key, value, _ := strings.Cut(item, "=")
				env[key] = value
			}
			if _, found := env[authName]; found {
				t.Errorf("known skill credential variable inherited while web=%t", state.web)
			}
			for key, want := range map[string]string{"R6_BUILD_SETTING": "ordinary-build-value", "GOFLAGS": "-mod=readonly -p=1"} {
				if env[key] != want {
					t.Errorf("ordinary %s=%q want %q", key, env[key], want)
				}
			}
		})
	}
}
