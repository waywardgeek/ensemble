package gui

import (
	"encoding/json"
	"ensemble"
	"os"
	"path/filepath"
	"testing"
)

func TestCh19SettingsPolicy(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		t.Run(map[bool]string{false: "missing", true: "legacy"}[legacy], func(t *testing.T) {
			dir := t.TempDir()
			app := ensemble.New(nil)
			defer app.Shutdown()
			a, err := app.NewAgent(ensemble.Config{Model: "offline", BaseDir: dir, SandboxRoot: dir, SafeMode: true, EnableWebSearch: false})
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, "settings.json")
			if legacy {
				if err = os.WriteFile(path, []byte(`{"theme":"light"}`), 0600); err != nil {
					t.Fatal(err)
				}
			}
			server, err := NewServerWithSettings(app, "web/gui", path)
			if err != nil {
				t.Fatal(err)
			}
			defer server.Hub.Close()
			server.Hub.mu.Lock()
			server.Hub.view(a)
			server.Hub.mu.Unlock()
			if !a.Config().SafeMode || a.Config().EnableWebSearch {
				t.Fatalf("launch policy overwritten: %+v", a.Config())
			}
			if err = server.Hub.updateSettings(json.RawMessage(`{"safe_mode":false,"enable_web_search":true}`)); err != nil {
				t.Fatal(err)
			}
			if a.Config().SafeMode || !a.Config().EnableWebSearch {
				t.Fatal("patch not applied")
			}
			loaded, err := server.Hub.store.load()
			if err != nil || loaded.SafeMode || !loaded.EnableWebSearch {
				t.Fatalf("roundtrip: %+v %v", loaded, err)
			}
		})
	}
	base := DefaultSettings()
	for _, raw := range []string{`{"safe_mode":null}`, `{"enable_web_search":"yes"}`} {
		if _, err := patchSettings(base, []byte(raw)); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}
func TestCh19SettingsLoadSymlink(t *testing.T) {
	dir := t.TempDir()
	app := ensemble.New(nil)
	defer app.Shutdown()
	a, err := app.NewAgent(ensemble.Config{Model: "offline", BaseDir: dir, SandboxRoot: dir})
	if err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside.json")
	if err = os.WriteFile(outside, []byte(`{"theme":"light"}`), 0600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "settings.json")
	if err = os.Symlink(outside, path); err != nil {
		t.Fatal(err)
	}
	if _, err = (settingsStore{path: path, agent: a}).load(); err == nil {
		t.Fatal("outside settings followed")
	}
}
