package gui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"ensemble"
)

func TestSettingsPatchValidation(t *testing.T) {
	base := DefaultSettings()
	for _, raw := range []string{`null`, `[]`, `{"theme":false}`, `{"tts_enabled":null}`, `{"tts_speed":"fast"}`, `{"max_tool_rounds":2.5}`, `{"max_tool_rounds":1e100}`, `{"unknown":true}`, `{"theme":"light","tts_enabled":5}`} {
		got, err := patchSettings(base, json.RawMessage(raw))
		if err == nil || got != base {
			t.Fatalf("invalid patch %s: %+v, %v", raw, got, err)
		}
	}
	got, err := patchSettings(base, json.RawMessage(`{"theme":"unknown","tts_enabled":false,"tts_speed":8,"font_size":1,"max_tool_rounds":-4}`))
	if err != nil || got.Theme != "dark" || got.TTSEnabled || got.TTSSpeed != 5 || got.FontSize != 12 || got.MaxToolRounds != 0 {
		t.Fatalf("clamping: %+v, %v", got, err)
	}
	got, err = patchSettings(base, json.RawMessage(`{"max_tool_rounds":12345}`))
	if err != nil || got.MaxToolRounds != 12345 || got.Theme != base.Theme || got.TTSSpeed != base.TTSSpeed || got.TTSEnabled != base.TTSEnabled {
		t.Fatalf("partial update: %+v, %v", got, err)
	}
}

func TestSettingsPersistenceAndFailedSave(t *testing.T) {
	root := ensemble.New(nil)
	defer root.Shutdown()
	agent, err := root.NewAgent(ensemble.Config{Model: "offline", BaseDir: t.TempDir(), DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	baseline := agent.Log().Events()
	store := settingsStore{path: filepath.Join(t.TempDir(), "preferences", "settings.json")}
	h := NewHub(root)
	h.store = store
	defer h.Close()
	first, second := h.connect(), h.connect()
	defer h.disconnect(first)
	defer h.disconnect(second)
	drain(first)
	drain(second)
	if err = h.updateSettings(json.RawMessage(`{"theme":"light","max_tool_rounds":321,"tts_speed":1.2345,"font_size":18}`)); err != nil {
		t.Fatal(err)
	}
	for _, c := range []*client{first, second} {
		frames := drain(c)
		if len(frames) != 1 || frames[0].Type != "settings_changed" || frames[0].Settings.MaxToolRounds != 321 {
			t.Fatalf("missing accepted broadcast: %+v", frames)
		}
	}
	saved, err := store.load()
	if err != nil || saved != h.currentSettings() || saved.TTSSpeed != 1.2345 || saved.FontSize != 18 || agent.Config().MaxToolRounds != 321 {
		t.Fatalf("save/policy: %+v, %v", saved, err)
	}
	before := h.currentSettings()
	// An ordinary file cannot be used as a parent directory, even by a privileged user.
	obstacle := filepath.Join(t.TempDir(), "file")
	if err = os.WriteFile(obstacle, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	h.store.path = filepath.Join(obstacle, "settings.json")
	if err = h.updateSettings(json.RawMessage(`{"theme":"dark","max_tool_rounds":1}`)); err == nil {
		t.Fatal("failed save accepted")
	}
	if h.currentSettings() != before || agent.Config().MaxToolRounds != 321 || len(drain(first)) != 0 || len(drain(second)) != 0 {
		t.Fatal("failed save changed accepted state")
	}
	loaded, err := store.load()
	if err != nil || loaded != saved {
		t.Fatal("previous persisted file changed")
	}
	c := h.connect()
	defer h.disconnect(c)
	frames := drain(c)
	if len(frames) < 2 || frames[1].Type != "current_settings" || *frames[1].Settings != before {
		t.Fatalf("connect settings: %+v", frames)
	}
	if after := agent.Log().Events(); !reflect.DeepEqual(after, baseline) {
		t.Fatalf("settings changed canonical history: before=%+v after=%+v", baseline, after)
	}
}

func TestSettingsCorruptFileIsNotSilentlyReset(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte(`{"theme":`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := (settingsStore{path: path}).load(); err == nil {
		t.Fatal("malformed persisted state accepted")
	}
}
