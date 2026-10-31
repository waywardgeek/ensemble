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
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
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

// modelStreamGate holds one real observation before the GUI consumes it. It
// controls scheduling without replacing either the actor or the GUI observer.
type modelStreamGate struct {
	entered, release chan struct{}
	once             sync.Once
}

// ObserveAgent gates the first streamed delta; all actual delivery still runs.
func (g *modelStreamGate) ObserveAgent(_ ensemble.Agent, observation ensemble.Observation) {
	if _, ok := observation.(ensemble.PartDelta); ok {
		g.once.Do(func() { close(g.entered); <-g.release })
	}
}

// TestModelSelectionDuringStream places a GUI setter behind a real observation.
// Busy rejection must not hold the mutex that the queued observation needs.
func TestModelSelectionDuringStream(t *testing.T) {
	finish := make(chan struct{})
	var finishOnce sync.Once
	stop := func() { finishOnce.Do(func() { close(finish) }) }
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		delta(w, map[string]string{"content": "streaming"})
		w.(http.Flusher).Flush()
		select {
		case <-finish:
			end(w)
		case <-r.Context().Done():
		}
	}))
	defer provider.Close()
	root := ensemble.New(io.Discard)
	gate := &modelStreamGate{entered: make(chan struct{}), release: make(chan struct{})}
	root.AddObserver(gate)
	a, err := root.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-4o", APIKey: "fake", BaseURL: provider.URL, DataDir: t.TempDir()})
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
	t.Cleanup(func() { release(); stop(); a.Shutdown(); hub.Close(); server.Close() })
	first, second := dial(t, server.URL), dial(t, server.URL)
	until(t, first, kind("current_settings"))
	until(t, second, kind("current_settings"))
	patch := func(model string) {
		t.Helper()
		if err := first.WriteJSON(map[string]any{"type": "update_settings", "settings": map[string]string{"model": model}}); err != nil {
			t.Fatal(err)
		}
	}
	selected := func(model string) {
		t.Helper()
		for _, peer := range []*websocket.Conn{first, second} {
			events := until(t, peer, kind("settings_changed"))
			if events[len(events)-1]["settings"].(map[string]any)["model"] != model || a.Config().Model != model {
				t.Fatal("selection broadcast before effective model")
			}
		}
	}
	patch("gpt-4o-mini")
	selected("gpt-4o-mini")
	t.Log("idle model selection acknowledged on both tabs")
	send(t, first, "prompt", "stream while changing model")
	select {
	case <-gate.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("stream never entered observer")
	}
	patch("gpt-5.6-sol")
	// Session model changes do not write preferences, so a disk barrier would
	// observe nothing. Wait for the actual socket setter's actor-reply wait.
	deadline := time.Now().Add(5 * time.Second)
	for {
		stack := make([]byte, 64*1024)
		stack = stack[:runtime.Stack(stack, true)]
		waiting := false
		for _, goroutine := range strings.Split(string(stack), "\n\n") {
			if strings.Contains(goroutine, ".SetModel(") && strings.Contains(goroutine, ".updateSettings(") && strings.Contains(goroutine, "[select]") {
				waiting = true
			}
		}
		if waiting {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("GUI model setter never waited for actor")
		}
		runtime.Gosched()
	}
	t.Log("GUI model setter is waiting behind gated observation")
	release()
	events := until(t, first, kind("settings_error"))
	if !strings.Contains(events[len(events)-1]["error"].(string), "between turns") || a.Config().Model != "gpt-4o-mini" {
		t.Fatal("busy selection changed model or lost its error")
	}
	until(t, second, kind("part_delta"))
	stop()
	until(t, first, kind("turn_ended"))
	patch("gpt-5.6-sol")
	selected("gpt-5.6-sol")
	if err := a.Shutdown(); err != nil {
		t.Fatal(err)
	}
	t.Log("busy rejection, progress, later selection and shutdown completed")
}
