package gui

import (
	"context"
	"ensemble"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

type ch19PreloadAuth struct {
	ensemble.Auth
	onMode func() error
}

func (a ch19PreloadAuth) Mode(context.Context) (ensemble.AuthMode, error) {
	if err := a.onMode(); err != nil {
		return "", err
	}
	return ensemble.AuthAPI, nil
}

func TestCh19R7PreloadSnapshot(t *testing.T) {
	assets, err := filepath.Abs("web/gui")
	if err != nil {
		t.Fatal(err)
	}
	caller := t.TempDir()
	base := filepath.Join(caller, "work")
	if err = os.MkdirAll(base, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(base, "settings.json")
	if err = os.WriteFile(path, []byte(`{"safe_mode":true,"enable_web_search":false,"theme":"dark"}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(caller)
	app := ensemble.New(nil)
	defer app.Shutdown()
	var calls atomic.Int32
	// Normal NewAgent consults Auth.Mode after preload. Mutate the owned settings
	// file then: Hub must use the parsed snapshot, not reread the replacement.
	if err = app.SetAuth(ch19PreloadAuth{Auth: app.Auth(), onMode: func() error {
		calls.Add(1)
		return os.WriteFile(path, []byte(`{"safe_mode":false,"enable_web_search":true,"theme":"light"}`), 0600)
	}}); err != nil {
		t.Fatal(err)
	}
	cfg := ensemble.Config{Model: "fixture", Vendor: ensemble.OpenAI, Surface: ensemble.ChatCompletions, BaseURL: "http://127.0.0.1:1", BaseDir: base, SandboxRoot: caller, DataDir: "data", SavePath: "caller.save", EnableWebSearch: true}
	server, a, err := NewServerFromConfig(app, cfg, assets, "settings.json")
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	if calls.Load() != 1 {
		t.Fatalf("original auth selection was normalized away: Mode calls=%d", calls.Load())
	}
	if !a.Config().SafeMode || a.Config().EnableWebSearch || a.Config().Surface != cfg.Surface || a.Config().Model != cfg.Model {
		t.Fatal("construction did not retain original config plus preloaded policy")
	}
	if !server.Hub.settings.SafeMode || server.Hub.settings.EnableWebSearch || server.Hub.settings.Theme != "dark" {
		t.Fatalf("Hub reread settings: %+v", server.Hub.settings)
	}
	if server.Hub.store.agent != a || server.Hub.store.path != path {
		t.Fatal("final store is not bound to final owner/resolved path")
	}
	if a.ID() != "agent-1" || a.Config().DataDir != filepath.Join(base, "data") || a.Config().SavePath != filepath.Join(caller, "caller.save") {
		t.Fatalf("construction paths changed: id=%s data=%s save=%s", a.ID(), a.Config().DataDir, a.Config().SavePath)
	}
	if err = a.Save(); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(caller, "caller.save")); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(base, "caller.save")); !os.IsNotExist(err) {
		t.Fatalf("explicit save moved under BaseDir: %v", err)
	}
}

func TestCh19R7MalformedSettingsBeforeAgent(t *testing.T) {
	for _, raw := range []string{"", "{]", `{"safe_mode":null}`, `{"enable_web_search":"yes"}`} {
		t.Run(raw, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(raw), 0600); err != nil {
				t.Fatal(err)
			}
			app := ensemble.New(nil)
			defer app.Shutdown()
			server, _, err := NewServerFromConfig(app, ensemble.Config{Model: "offline", BaseDir: dir, SandboxRoot: dir, DataDir: filepath.Join(dir, "data")}, "web/gui", "settings.json")
			if server != nil {
				defer server.Close()
			}
			if err == nil || len(app.Agents()) != 0 {
				t.Fatalf("malformed settings reached construction: %q %v", raw, err)
			}
		})
	}
}
