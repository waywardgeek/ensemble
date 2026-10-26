package gui_test

import (
	"encoding/json"
	gui "ensemble-gui-stub"
	"ensemble/ensemble"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
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
		if s["theme"] != "system" || s["tts_enabled"] != true || s["tts_speed"] != float64(1) || s["font_size"] != float64(16) || s["max_tool_rounds"] != float64(0) {
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
	patch(map[string]any{"theme": "light", "tts_speed": 4.2, "font_size": 20, "max_tool_rounds": 3, "context_target": 100, "log_retention": 7, "memory": map[string]any{"session": map[string]any{"budget": 80, "disabled": true}}})
	for _, peer := range []*websocket.Conn{first, second} {
		s := read(peer, "settings_changed")["settings"].(map[string]any)
		if s["theme"] != "light" || s["tts_speed"] != 4.2 || s["font_size"] != float64(20) || s["max_tool_rounds"] != float64(3) || s["context_target"] != float64(20000) || s["log_retention"] != float64(7) {
			t.Fatal("roundtrip/broadcast", s)
		}
		memory := s["memory"].(map[string]any)["session"].(map[string]any)
		if memory["budget"] != float64(80) || memory["disabled"] != true {
			t.Fatal("memory patch not effective", memory)
		}
	}
	freshMemory := read(dial(t, server.URL), "current_settings")["settings"].(map[string]any)["memory"].(map[string]any)["session"].(map[string]any)
	if freshMemory["budget"] != float64(80) || freshMemory["disabled"] != true {
		t.Fatal("memory patch not effective", freshMemory)
	}
	patch(map[string]any{"tts_enabled": false, "max_tool_rounds": 0, "memory": map[string]any{"session": map[string]any{"disabled": false}}})
	for _, peer := range []*websocket.Conn{first, second} {
		s := read(peer, "settings_changed")["settings"].(map[string]any)
		if s["tts_enabled"] != false || s["max_tool_rounds"] != float64(0) || s["theme"] != "light" || s["tts_speed"] != 4.2 {
			t.Fatal("patch lost false/zero or absent fields", s)
		}
	}
	_, restored := start()
	peer := dial(t, restored.URL)
	s := read(peer, "current_settings")["settings"].(map[string]any)
	if s["theme"] != "light" || s["tts_enabled"] != false || s["tts_speed"] != 4.2 || s["context_target"] != float64(20000) || s["log_retention"] != float64(7) {
		t.Fatal("restart lost preferences", s)
	}
	memory := s["memory"].(map[string]any)["session"].(map[string]any)
	if memory["budget"] != float64(80) || memory["disabled"] != false {
		t.Fatal("memory restart lost patch", memory)
	}
	patch(map[string]any{"theme": "invalid", "tts_speed": -5, "font_size": 99, "max_tool_rounds": -5})
	s = read(first, "settings_changed")["settings"].(map[string]any)
	if s["theme"] != "system" || s["tts_speed"] != 0.5 || s["font_size"] != float64(24) || s["max_tool_rounds"] != float64(0) {
		t.Fatal("entry not clamped", s)
	}
	patch(map[string]any{"context_target": 0, "log_retention": 0})
	s = read(first, "settings_changed")["settings"].(map[string]any)
	if s["context_target"] != float64(0) || s["log_retention"] != float64(0) {
		t.Fatal("context zero defaults lost", s)
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
	if s["theme"] != "system" || s["tts_speed"] != 0.5 {
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
	if s["theme"] != "system" || s["tts_enabled"] != false || s["tts_speed"] != float64(5) || s["font_size"] != float64(12) || s["max_tool_rounds"] != float64(0) {
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

// streamGate is a recording display subscriber placed before the GUI observer.
// Its first delta gate lets the test coordinate real delivery with a disk write.
type streamGate struct {
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

// ObserveAgent records when streamed delivery begins, then waits for the test's
// one release. It neither replaces nor changes the actual Agent or GUI observer.
func (g *streamGate) ObserveAgent(_ ensemble.Agent, o ensemble.Observation) {
	if _, ok := o.(ensemble.PartDelta); ok {
		g.once.Do(func() { close(g.entered); <-g.release })
	}
}

// TestMemorySettingsDuringStream keeps the actual provider stream open until
// settings finish. Both tabs must see each effective patch while progress runs.
// This covers the reviewer-found GUI mutex/actor acknowledgment cycle.
func TestMemorySettingsDuringStream(t *testing.T) {
	finish := make(chan struct{})
	var once sync.Once
	stop := func() { once.Do(func() { close(finish) }) }
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if r.Method != "POST" || r.URL.Path != "/v1/chat/completions" || json.NewDecoder(r.Body).Decode(&body) != nil || body["model"] != "gpt-4o-mini" || body["stream"] != true || len(body["messages"].([]any)) == 0 {
			http.Error(w, "invalid streaming chat request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-r.Context().Done():
				return
			case <-finish:
				end(w)
				return
			case <-ticker.C:
				delta(w, map[string]string{"content": "word "})
			}
		}
	}))
	defer provider.Close()
	defer stop()
	root := ensemble.New(io.Discard)
	gate := &streamGate{entered: make(chan struct{}), release: make(chan struct{})}
	root.AddObserver(gate)
	a, err := root.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-4o-mini", APIKey: "fake", BaseURL: provider.URL, DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	hub, err := gui.New(root, a, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(hub.Handler("web/gui"))
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(gate.release) }) }
	t.Cleanup(func() { release(); a.Shutdown(); hub.Close(); server.Close() })
	first, second := dial(t, server.URL), dial(t, server.URL)
	until(t, first, kind("current_settings"))
	until(t, second, kind("current_settings"))
	patch := func(n int) {
		t.Helper()
		writer := first
		if n%2 != 0 {
			writer = second
		}
		if err := writer.WriteJSON(map[string]any{"type": "update_settings", "settings": map[string]any{"memory": map[string]any{"session": map[string]any{"budget": 100 + n}}}}); err != nil {
			t.Fatal(err)
		}
		for _, peer := range []*websocket.Conn{first, second} {
			events := until(t, peer, kind("settings_changed"))
			memory := events[len(events)-1]["settings"].(map[string]any)["memory"].(map[string]any)["session"].(map[string]any)
			if memory["budget"] != float64(100+n) || a.Config().Memory.Session.Budget != 100+n {
				t.Fatal("acknowledgment preceded effective settings")
			}
		}
	}
	patch(0)
	t.Log("idle settings control acknowledged")
	send(t, first, "prompt", "stream while settings change")
	select {
	case <-gate.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("stream did not enter observer")
	}
	// The real settings file is our barrier. In the defective source, completing
	// this write proves the socket holds the mutex needed by the next GUI observer.
	if err := first.WriteJSON(map[string]any{"type": "update_settings", "settings": map[string]any{"memory": map[string]any{"session": map[string]any{"budget": 101}}}}); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		data, _ := os.ReadFile(filepath.Join(a.Config().DataDir, "settings.json"))
		var saved gui.Settings
		if json.Unmarshal(data, &saved) == nil && saved.Memory.Session.Budget == 101 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("active settings did not reach persistence barrier")
		}
		time.Sleep(time.Millisecond)
	}
	release()
	for _, peer := range []*websocket.Conn{first, second} {
		until(t, peer, kind("settings_changed"))
	}
	until(t, first, kind("part_delta"))
	for i := 2; i <= 32; i++ {
		patch(i)
	}
	t.Log("all active-stream patches acknowledged on both tabs")
	// Subscribe after the patches while the response is still streaming.
	fresh := dial(t, server.URL)
	events := until(t, fresh, kind("current_settings"))
	memory := events[len(events)-1]["settings"].(map[string]any)["memory"].(map[string]any)["session"].(map[string]any)
	if memory["budget"] != float64(132) {
		t.Fatal("subscription missed effective settings")
	}
	stop()
	until(t, first, kind("turn_ended"))
	if err := a.Shutdown(); err != nil {
		t.Fatal(err)
	}
	t.Log("stream completed and Agent shutdown returned")
}
