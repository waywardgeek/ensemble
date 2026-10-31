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
	"testing"
)

// Selection travels from one real WebSocket to the owning actor and is visible
// to another tab. The next request proves that this is more than a GUI echo.
func TestModelSelectionWireAndRequest(t *testing.T) {
	requested := make(chan string, 1)
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			// Model records the actual provider selection.
			Model string
		}
		json.NewDecoder(r.Body).Decode(&request)
		requested <- request.Model
		io.WriteString(w, `{"content":[{"type":"text","text":"selected"}],"usage":{"input_tokens":2,"output_tokens":1}}`)
	}))
	defer provider.Close()
	root := ensemble.New(io.Discard)
	a, err := root.NewAgent(ensemble.Config{Model: "claude-opus-5", BaseURL: provider.URL, DisableStreaming: true, DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	hub, err := gui.New(root, a, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Close()
	server := httptest.NewServer(hub.Handler("web/gui"))
	defer server.Close()
	first, second := dial(t, server.URL), dial(t, server.URL)
	until(t, first, kind("current_settings"))
	until(t, second, kind("current_settings"))
	patch := func(value any) {
		t.Helper()
		if err := first.WriteJSON(map[string]any{"type": "update_settings", "settings": value}); err != nil {
			t.Fatal(err)
		}
	}
	patch(map[string]any{"model": "claude-sonnet-5"})
	for _, peer := range []*websocket.Conn{first, second} {
		events := until(t, peer, kind("settings_changed"))
		if events[len(events)-1]["settings"].(map[string]any)["model"] != "claude-sonnet-5" {
			t.Fatal("selection not broadcast")
		}
	}
	send(t, second, "prompt", "answer")
	until(t, second, kind("turn_ended"))
	if got := <-requested; got != "claude-sonnet-5" {
		t.Fatal("selected model never reached provider", got)
	}
	// Runtime selection must not change the next launch's configured provider or
	// model. Saving an unrelated preference must not accidentally persist it.
	patch(map[string]any{"theme": "light"})
	until(t, first, kind("settings_changed"))
	data, err := os.ReadFile(filepath.Join(a.Config().DataDir, "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	var saved map[string]any
	json.Unmarshal(data, &saved)
	if _, exists := saved["model"]; exists {
		t.Fatal("session model persisted")
	}
	patch(map[string]any{"model": "claude-opus-5", "theme": "dark"})
	until(t, first, kind("settings_error"))
	if a.Config().Model != "claude-sonnet-5" {
		t.Fatal("rejected mixed update changed model")
	}
}
