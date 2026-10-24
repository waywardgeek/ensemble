package gui

import (
	"encoding/json"
	"ensemble"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func ch19StartGUI(root *ensemble.Ensemble, cfg ensemble.Config, path string) (*Server, ensemble.Agent, error) {
	return NewServerFromConfig(root, cfg, "web/gui", path)
}
func TestCh19R2SettingsBeforeStartup(t *testing.T) {
	var handshakes atomic.Int32
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
		}
		_ = json.NewDecoder(r.Body).Decode(&m)
		if len(m.ID) == 0 {
			w.WriteHeader(202)
			return
		}
		result := `{}`
		switch m.Method {
		case "initialize":
			handshakes.Add(1)
			result = `{"protocolVersion":"2025-03-26","capabilities":{"tools":{}}}`
		case "tools/list":
			result = `{"tools":[{"name":"lookup","description":"fixture","inputSchema":{"type":"object"}}]}`
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":%s}`, m.ID, result)
	}))
	defer peer.Close()
	for _, mode := range []string{"safe-initial", "startup-web", "recovery-web"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			skills := filepath.Join(dir, "skills")
			for name, doc := range map[string]string{"primary": "---\nname: primary\ntype: primary\ndescription: fixture\ntools: run_command\nloadable-skills: web-search\n---\nFixture", "web": "---\nname: web-search\ntype: loadable\ndescription: fixture\ntools: lookup\nmcp_servers:\n  - name: startup-web\n    transport: url\n    url: " + peer.URL + "\n---\nFixture"} {
				path := filepath.Join(skills, name, "SKILL.md")
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(doc), 0600); err != nil {
					t.Fatal(err)
				}
			}
			cfg := ensemble.Config{Model: "offline", BaseDir: dir, SandboxRoot: dir, DataDir: filepath.Join(dir, "data"), SavePath: filepath.Join(dir, "save.json"), EnableWebSearch: true, Skills: &ensemble.SkillSpec{Directory: skills, Primary: "primary"}}
			if mode == "startup-web" {
				cfg.StartupSkills = []string{"web-search"}
			}
			if mode == "recovery-web" {
				seed := ensemble.New(nil)
				a, err := seed.NewAgent(cfg)
				if err != nil {
					t.Fatal(err)
				}
				if err = a.Skills().LoadSkill("web-search"); err != nil {
					t.Fatal(err)
				}
				if err = seed.Shutdown(); err != nil {
					t.Fatal(err)
				}
			}
			path := filepath.Join(dir, "settings.json")
			if err := os.WriteFile(path, []byte(`{"safe_mode":true,"enable_web_search":false}`), 0600); err != nil {
				t.Fatal(err)
			}
			handshakes.Store(0)
			app := ensemble.New(nil)
			defer app.Shutdown()
			server, a, err := ch19StartGUI(app, cfg, path)
			if server != nil {
				defer server.Close()
			}
			if got := handshakes.Load(); got != 0 {
				t.Errorf("persisted web-off arrived after %d startup handshakes", got)
			}
			if mode == "safe-initial" {
				if err != nil {
					t.Fatal(err)
				}
				if !a.Config().SafeMode || a.Config().EnableWebSearch {
					t.Fatal("saved security policy was not initial Agent policy")
				}
				for _, event := range a.Log().Events() {
					if event.Tools != nil {
						for _, decl := range event.Tools.Added {
							if decl.Name == "run_command" {
								t.Fatal("unsafe tool was published before saved policy")
							}
						}
					}
				}
			} else if err == nil {
				t.Error("disabled requested/recovered web skill was not refused before startup")
			}
		})
	}
}

func TestCh19R2LaunchPolicyDefaults(t *testing.T) {
	for _, mode := range []string{"missing", "legacy", "outside"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "settings.json")
			if mode == "legacy" {
				if err := os.WriteFile(path, []byte(`{"theme":"light"}`), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "outside" {
				outside := filepath.Join(t.TempDir(), "settings.json")
				if err := os.WriteFile(outside, []byte(`{"safe_mode":false,"enable_web_search":true}`), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, path); err != nil {
					t.Fatal(err)
				}
			}
			app := ensemble.New(nil)
			defer app.Shutdown()
			server, a, err := ch19StartGUI(app, ensemble.Config{Model: "offline", BaseDir: dir, SandboxRoot: dir, DataDir: filepath.Join(dir, "data"), SafeMode: true, EnableWebSearch: false}, path)
			if server != nil {
				defer server.Close()
			}
			if mode == "outside" {
				if err == nil || len(app.Agents()) != 0 {
					t.Fatal("outside settings was not refused before Agent installation")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !a.Config().SafeMode || a.Config().EnableWebSearch {
				t.Fatal("missing/legacy settings overwrote deliberate launch policy")
			}
		})
	}
}
