package gui_test

import (
	"encoding/json"
	gui "ensemble-gui-stub"
	"ensemble/ensemble"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func applicationFixture(t *testing.T) string {
	t.Helper()
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "api-key"), []byte("fixture-key"), 0600); err != nil {
		t.Fatal(err)
	}
	config := map[string]any{
		"version":  1,
		"provider": map[string]any{"vendor": "openai", "model": "gpt-6-astra", "billing_mode": "api", "credential_file": "api-key", "reasoning_effort": "high", "reasoning_summary": "auto"},
		"gui":      map[string]any{"theme": "light", "tts_speed": 2.6},
	}
	path := filepath.Join(directory, "settings.json")
	data, _ := json.Marshal(config)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestApplicationSettingsWireAndDisk(t *testing.T) {
	path := applicationFixture(t)
	store, err := gui.LoadApplicationConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	cfg, spec, err := store.Snapshot().RuntimeConfig()
	if err != nil {
		t.Fatal(err)
	}
	root := ensemble.New(io.Discard)
	a, err := root.NewAgent(cfg, spec)
	if err != nil {
		t.Fatal(err)
	}
	// A stale, legacy preferences file must not override file-mode preferences.
	os.MkdirAll(cfg.DataDir, 0700)
	os.WriteFile(filepath.Join(cfg.DataDir, "settings.json"), []byte(`{"theme":"dark","tts_speed":1}`), 0600)
	hub, err := gui.NewConfigured(root, a, io.Discard, store)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(hub.Handler("web/gui"))
	t.Cleanup(func() { hub.Close(); server.Close(); a.Shutdown() })
	peer := dial(t, server.URL)
	messages := until(t, peer, kind("current_settings"))
	first := messages[len(messages)-1]["settings"].(map[string]any)
	if first["theme"] != "light" || first["tts_speed"] != 2.6 {
		t.Fatalf("wrong preference source: %v", first)
	}
	app := first["application"].(map[string]any)
	if app["restart_required"] != false || app["path"] != path {
		t.Fatalf("initial application: %v", app)
	}
	patch := func(value any, event string) map[string]any {
		t.Helper()
		if err := peer.WriteJSON(map[string]any{"type": "update_settings", "settings": value, "request_id": "fixture-save"}); err != nil {
			t.Fatal(err)
		}
		events := until(t, peer, kind(event))
		reply := events[len(events)-1]
		if reply["request_id"] != "fixture-save" {
			t.Fatalf("missing save acknowledgement: %v", reply)
		}
		return reply
	}
	updated := patch(map[string]any{"provider": map[string]any{"reasoning_summary": "detailed"}}, "settings_changed")["settings"].(map[string]any)
	app = updated["application"].(map[string]any)
	if app["restart_required"] != true || app["active_provider"].(map[string]any)["reasoning_summary"] != "auto" || app["provider"].(map[string]any)["reasoning_summary"] != "detailed" {
		t.Fatalf("restart contract: %v", app)
	}
	if a.Config().ReasoningSummary != "auto" {
		t.Fatal("provider patch changed running request configuration")
	}
	patch(map[string]any{"theme": "dark", "tts_speed": 5}, "settings_changed")
	reopened, err := gui.LoadApplicationConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	saved := reopened.Snapshot()
	if saved.Provider.ReasoningSummary != "detailed" || saved.GUI.Theme != "dark" || saved.GUI.TTSSpeed != 5 {
		t.Fatalf("lost settings: %+v", saved)
	}
	restartCfg, _, err := saved.RuntimeConfig()
	if err != nil {
		t.Fatal(err)
	}
	if restartCfg.ReasoningSummary != "detailed" {
		t.Fatal("restart did not pick up provider change")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file map[string]json.RawMessage
	json.Unmarshal(data, &file)
	var prefs map[string]json.RawMessage
	json.Unmarshal(file["gui"], &prefs)
	if _, ok := prefs["application"]; ok {
		t.Fatal("transient application metadata saved into preferences")
	}
	for _, bad := range []any{
		map[string]any{"model": "gpt-5.6"},
		map[string]any{"provider": map[string]any{"api_key": "do-not-save"}},
		map[string]any{"provider": map[string]any{"reasoning_summary": "bad"}},
		map[string]any{"provider": map[string]any{"model": "gpt-5.6"}, "theme": "light"},
	} {
		patch(bad, "settings_error")
	}
	backup := path + ".backup"
	if err := os.Rename(path, backup); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	patch(map[string]any{"provider": map[string]any{"model": "gpt-5.6"}}, "settings_error")
	patch(map[string]any{"theme": "light"}, "settings_error")
	if store.Snapshot().Provider.Model != "gpt-6-astra" || store.Snapshot().GUI.Theme != "dark" {
		t.Fatal("failed write changed accepted settings")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(backup, path); err != nil {
		t.Fatal(err)
	}
}
