package gui_test

import (
	"encoding/json"
	gui "ensemble-gui-stub"
	"ensemble/ensemble"
	"github.com/gorilla/websocket"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestSettingsWireAndDisk follows a patch through both clients and a fresh
// server. Malformed input and failed writes must leave the accepted state intact.
func TestSettingsWireAndDisk(t *testing.T) {
	dir := t.TempDir()
	start := func() (*gui.Server, *httptest.Server) {
		root := ensemble.New(io.Discard)
		a, err := root.NewAgent(ensemble.Config{Model: "offline", DataDir: dir})
		if err != nil {
			t.Fatal(err)
		}
		hub, err := gui.New(root, a, io.Discard)
		if err != nil {
			t.Fatal(err)
		}
		server := httptest.NewServer(hub.Handler("web/gui"))
		t.Cleanup(func() { hub.Close(); server.Close(); a.Shutdown() })
		return hub, server
	}
	_, server := start()
	first, second := dial(t, server.URL), dial(t, server.URL)
	for _, peer := range []*websocket.Conn{first, second} {
		messages := until(t, peer, kind("current_settings"))
		s := messages[len(messages)-1]["settings"].(map[string]any)
		if s["theme"] != "dark" || s["tts_enabled"] != true || s["tts_speed"] != float64(1) || s["font_size"] != float64(16) || s["max_tool_rounds"] != float64(0) {
			t.Fatal("defaults", s)
		}
	}
	patch := func(value any) {
		t.Helper()
		if err := first.WriteJSON(map[string]any{"type": "update_settings", "settings": value}); err != nil {
			t.Fatal(err)
		}
	}
	read := func(peer *websocket.Conn, expected string) map[string]any {
		t.Helper()
		events := until(t, peer, kind(expected))
		return events[len(events)-1]
	}
	patch(map[string]any{"theme": "light", "tts_speed": 1.7, "font_size": 20, "max_tool_rounds": 3})
	for _, peer := range []*websocket.Conn{first, second} {
		s := read(peer, "settings_changed")["settings"].(map[string]any)
		if s["theme"] != "light" || s["tts_speed"] != 1.7 || s["font_size"] != float64(20) || s["max_tool_rounds"] != float64(3) {
			t.Fatal("roundtrip/broadcast", s)
		}
	}
	patch(map[string]any{"tts_enabled": false, "max_tool_rounds": 0})
	for _, peer := range []*websocket.Conn{first, second} {
		s := read(peer, "settings_changed")["settings"].(map[string]any)
		if s["tts_enabled"] != false || s["max_tool_rounds"] != float64(0) || s["theme"] != "light" || s["tts_speed"] != 1.7 {
			t.Fatal("patch lost false/zero or absent fields", s)
		}
	}
	_, restored := start()
	peer := dial(t, restored.URL)
	s := read(peer, "current_settings")["settings"].(map[string]any)
	if s["theme"] != "light" || s["tts_enabled"] != false || s["tts_speed"] != 1.7 {
		t.Fatal("restart lost preferences", s)
	}
	patch(map[string]any{"theme": "invalid", "tts_speed": -5, "font_size": 99, "max_tool_rounds": -5})
	s = read(first, "settings_changed")["settings"].(map[string]any)
	if s["theme"] != "dark" || s["tts_speed"] != 0.5 || s["font_size"] != float64(24) || s["max_tool_rounds"] != float64(0) {
		t.Fatal("entry not clamped", s)
	}
	// A type error after a valid field must not partially change the object.
	patch(map[string]any{"theme": "light", "tts_speed": "bad"})
	read(first, "settings_error")
	// A directory at the replacement path forces a real write failure even when
	// tests run with permissions that would bypass an ordinary read-only mode.
	path := filepath.Join(dir, "settings.json")
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	patch(map[string]any{"theme": "light"})
	read(first, "settings_error")
	fresh := dial(t, server.URL)
	s = read(fresh, "current_settings")["settings"].(map[string]any)
	if s["theme"] != "dark" || s["tts_speed"] != 0.5 {
		t.Fatal("failed patch changed accepted state", s)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"theme":"wrong","tts_enabled":false,"tts_speed":9,"font_size":0,"max_tool_rounds":-1}`), 0600); err != nil {
		t.Fatal(err)
	}
	_, loaded := start()
	s = read(dial(t, loaded.URL), "current_settings")["settings"].(map[string]any)
	if s["theme"] != "dark" || s["tts_enabled"] != false || s["tts_speed"] != float64(2) || s["font_size"] != float64(12) || s["max_tool_rounds"] != float64(0) {
		t.Fatal("disk not clamped", s)
	}
	if err := os.WriteFile(path, json.RawMessage(`{"theme":`), 0600); err != nil {
		t.Fatal(err)
	}
	root := ensemble.New(io.Discard)
	a, _ := root.NewAgent(ensemble.Config{Model: "offline", DataDir: dir})
	defer a.Shutdown()
	if hub, err := gui.New(root, a, io.Discard); err == nil {
		hub.Close()
		t.Fatal("corrupt settings silently accepted")
	}
}
