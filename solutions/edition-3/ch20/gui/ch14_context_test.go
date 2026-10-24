package gui

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestChapter14ContextTargetSettings(t *testing.T) {
	base := DefaultSettings()
	if base.ContextTarget != 0 {
		t.Fatal("default must retain zero sentinel")
	}
	for _, raw := range []string{`{"context_target":null}`, `{"context_target":-1}`, `{"context_target":1.5}`, `{"context_target":"20000"}`} {
		got, err := patchSettings(base, json.RawMessage(raw))
		if err == nil || got != base {
			t.Fatalf("invalid patch changed settings: %s", raw)
		}
	}
	next, err := patchSettings(base, json.RawMessage(`{"context_target":1}`))
	if err != nil || next.ContextTarget != 20000 {
		t.Fatalf("clamp: %+v %v", next, err)
	}
	store := settingsStore{path: filepath.Join(t.TempDir(), "preferences.data")}
	if err = store.save(next); err != nil {
		t.Fatal(err)
	}
	got, err := store.load()
	if err != nil || got.ContextTarget != 20000 {
		t.Fatalf("restore: %+v %v", got, err)
	}
	next, err = patchSettings(next, json.RawMessage(`{"context_target":0}`))
	if err != nil || next.ContextTarget != 0 {
		t.Fatal("zero sentinel rejected")
	}
}

// Feed real settings-owner frames through the production JS socket/input handlers.
func TestChapter14TargetDraftRealSettingsWS(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal("Node is required for this focused production-JS test")
	}
	app, a := idleAgent(t, false)
	server, err := NewServerWithSettings(app, "web/gui", filepath.Join(t.TempDir(), "preferences.store"))
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	server.DefaultAgent = a.ID()
	host := httptest.NewServer(server.Handler(server.Assets))
	defer host.Close()
	client, other := r4Dial(t, host.URL), r4Dial(t, host.URL)
	read := func(ws *websocket.Conn, kind string) Wire {
		t.Helper()
		_ = ws.SetReadDeadline(time.Now().Add(3 * time.Second))
		for {
			_, raw, err := ws.ReadMessage()
			if err != nil {
				t.Fatalf("waiting for %s: %v", kind, err)
			}
			var frame Wire
			if err := json.Unmarshal(raw, &frame); err != nil {
				t.Fatal(err)
			}
			if frame.Type == kind {
				return frame
			}
		}
	}
	snapshot := read(client, "current_settings")
	read(other, "current_settings")
	send := func(ws *websocket.Conn, settings any) {
		t.Helper()
		if err := ws.WriteJSON(map[string]any{"type": "update_settings", "settings": settings}); err != nil {
			t.Fatal(err)
		}
	}
	send(other, map[string]any{"theme": "light"})
	unrelated := read(client, "settings_changed")
	send(client, map[string]any{"context_target": 1})
	accepted := read(client, "settings_changed")
	if accepted.Settings == nil || accepted.Settings.ContextTarget != 20000 || a.Config().ContextTarget != 20000 {
		t.Fatal("accepted target did not reach settings and live Agent")
	}
	// Capture a real unrelated broadcast for the queued-frame ordering: the JS
	// consumer submits its negative draft before receiving this shared value.
	send(other, map[string]any{"theme": "dark"})
	unrelatedAtFloor := read(client, "settings_changed")
	if unrelatedAtFloor.Settings == nil || unrelatedAtFloor.Settings.ContextTarget != 20000 {
		t.Fatal("unrelated broadcast changed the accepted target")
	}
	send(client, map[string]any{"context_target": -1})
	refused := read(client, "settings_error")
	if a.Config().ContextTarget != 20000 {
		t.Fatal("refused target changed live policy")
	}
	frames, err := json.Marshal(map[string]Wire{"snapshot": snapshot, "unrelated": unrelated, "accepted": accepted, "unrelatedAtFloor": unrelatedAtFloor, "refused": refused})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "settings.frames")
	if err := os.WriteFile(path, frames, 0600); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(node, "ch14_context_test.cjs", path).CombinedOutput()
	t.Logf("%s", out)
	if err != nil {
		t.Fatal(err)
	}
}
