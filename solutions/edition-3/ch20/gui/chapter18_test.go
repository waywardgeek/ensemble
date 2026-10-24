package gui

import (
	"context"
	"encoding/json"
	"ensemble"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestChapter18WSModelProjection(t *testing.T) {
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"model":"claude-sonnet-4-6","content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":4,"output_tokens":2}}`)
	}))
	defer peer.Close()
	app := ensemble.New(nil)
	defer app.Shutdown()
	a, err := app.NewAgent(ensemble.Config{Model: "claude-sonnet-4-6", APIKey: "dummy", BaseURL: peer.URL, BaseDir: t.TempDir(), DataDir: t.TempDir(), Tools: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = a.Ask(context.Background(), "fixture"); err != nil {
		t.Fatal(err)
	}
	server, err := NewServerWithSettings(app, "web/gui", filepath.Join(t.TempDir(), "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	server.DefaultAgent = a.ID()
	host := httptest.NewServer(server.Handler(server.Assets))
	defer host.Close()
	ws := r4Dial(t, host.URL)
	var frames []json.RawMessage
	read := func() struct {
		Type  string
		Meter ensemble.EngineStatus
		State struct{ Paused bool }
	} {
		t.Helper()
		var raw json.RawMessage
		if err := ws.ReadJSON(&raw); err != nil {
			t.Fatal(err)
		}
		frames = append(frames, raw)
		var v struct {
			Type  string
			Meter ensemble.EngineStatus
			State struct{ Paused bool }
		}
		if err := json.Unmarshal(raw, &v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	for read().Type != "replay" {
	}
	send := func(kind, text string) {
		t.Helper()
		if err := ws.WriteJSON(map[string]string{"type": kind, "agent_id": a.ID(), "text": text}); err != nil {
			t.Fatal(err)
		}
	}
	send("set_model", "new-private-model")
	for {
		v := read()
		if v.Type == "usage" && v.Meter.SelectedModel == "new-private-model" {
			if v.Meter.LastModel != "claude-sonnet-4-6" {
				t.Fatal("producer identity changed")
			}
			break
		}
	}
	send("agent_status", "")
	send("pause", "") // Ordered state update after both human commands.
	for {
		v := read()
		if v.Type == "state_changed" && v.State.Paused {
			break
		}
	}
	path := filepath.Join(t.TempDir(), "model-frames.json")
	b, _ := json.Marshal(frames)
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command("node", "chapter18_test.cjs", path).CombinedOutput(); err != nil {
		t.Fatalf("production JS consumer: %v\n%s", err, output)
	}
}
