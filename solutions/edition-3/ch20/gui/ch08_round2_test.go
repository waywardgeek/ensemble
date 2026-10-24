package gui

import (
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"ensemble"
	"github.com/gorilla/websocket"
)

// The real GUI websocket path supplies the reason and explicit-Continue
// authority consumed by the page tests. No model request or native voice runs.
func TestChapter8R2SpeechAndContinueProtocol(t *testing.T) {
	app, a := idleAgent(t, false)
	s, err := NewServerWithSettings(app, "web/gui", filepath.Join(t.TempDir(), "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	s.DefaultAgent = a.ID()
	host := httptest.NewServer(s.Handler(s.Assets))
	t.Cleanup(host.Close)
	first, second := r4Dial(t, host.URL), r4Dial(t, host.URL)
	clients := []*websocket.Conn{first, second}
	readUntil := func(ws *websocket.Conn, accept func(Wire) bool) Wire {
		t.Helper()
		for {
			var w Wire
			if err := ws.ReadJSON(&w); err != nil {
				t.Fatal(err)
			}
			if w.Type == "error" || w.Type == "settings_error" {
				t.Fatalf("unexpected GUI error: %+v", w)
			}
			if accept(w) {
				return w
			}
		}
	}
	write := func(ws *websocket.Conn, fields map[string]any) {
		t.Helper()
		fields["agent_id"] = a.ID()
		if err := ws.WriteJSON(fields); err != nil {
			t.Fatal(err)
		}
	}
	decode := func(w Wire, dst any) {
		t.Helper()
		raw, err := json.Marshal(w.Data)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(raw, dst); err != nil {
			t.Fatal(err)
		}
	}
	ids := make([]string, 2)
	for i, ws := range clients {
		var replay replayEnvelope
		decode(readUntil(ws, func(w Wire) bool { return w.Type == "replay" }), &replay)
		ids[i] = replay.Client
		readUntil(ws, func(w Wire) bool { return w.Type == "current_settings" })
	}
	write(first, map[string]any{"type": "update_settings", "settings": map[string]any{"tts_enabled": false}})
	for _, ws := range clients {
		readUntil(ws, func(w Wire) bool { return w.Type == "settings_changed" && w.Settings != nil && !w.Settings.TTSEnabled })
	}

	write(first, map[string]any{"type": "pause"})
	var before controlState
	for _, ws := range clients {
		w := readUntil(ws, func(w Wire) bool { return w.Type == "state_changed" && w.State != nil && w.State.Paused })
		before = *w.State
	}
	write(first, map[string]any{"type": "unpause", "automatic": true})
	for _, ws := range clients {
		readUntil(ws, func(w Wire) bool {
			if w.Type == "continued" {
				t.Fatal("automatic release must not broadcast explicit Continue authority")
			}
			return w.Type == "state_changed" && w.State != nil && !w.State.Paused && w.State.ContinueEpoch > before.ContinueEpoch
		})
	}

	const completeText = "Read this complete displayed text."
	if err := a.Post(ensemble.Ephemeral{Text: completeText}); err != nil {
		t.Fatal(err)
	}
	var partID uint64
	for _, ws := range clients {
		w := readUntil(ws, func(w Wire) bool { return w.Type == "part_final" && w.Part != nil && w.Part.Text == completeText })
		partID = w.PartID
	}
	if partID == 0 {
		t.Fatal("displayed card needs a durable part ID")
	}
	write(first, map[string]any{"type": "speak", "part_id": partID})
	var requested struct{ ID, Reason string }
	for _, ws := range clients {
		decode(readUntil(ws, func(w Wire) bool { return w.Type == "speech_requested" }), &requested)
		if requested.Reason != "on demand" || requested.ID == "" {
			t.Fatalf("production explicit Speak payload: %+v", requested)
		}
	}
	write(first, map[string]any{"type": "claim_speech", "speech_id": requested.ID})
	for _, ws := range clients {
		var claim struct{ ID, Owner, Reason, Text string }
		decode(readUntil(ws, func(w Wire) bool { return w.Type == "speech_claimed" }), &claim)
		if claim.ID != requested.ID || claim.Owner != ids[0] || claim.Reason != "on demand" || claim.Text != completeText {
			t.Fatalf("wrong speech ownership: %+v", claim)
		}
	}
	// Keep a pending explicit request behind the owned one. Continue must cancel
	// both without letting a late completion activate queued stale speech.
	write(first, map[string]any{"type": "speak", "part_id": partID})
	write(first, map[string]any{"type": "pause"})
	for _, ws := range clients {
		w := readUntil(ws, func(w Wire) bool { return w.Type == "state_changed" && w.State != nil && w.State.Paused })
		before = *w.State
	}
	write(second, map[string]any{"type": "unpause"})
	for _, ws := range clients {
		cancelled := false
		w := readUntil(ws, func(w Wire) bool {
			if w.Type == "state_changed" && w.State != nil && w.State.ContinueEpoch > before.ContinueEpoch {
				t.Fatal("state acknowledgement preceded explicit Continue authority")
			}
			if w.Type == "speech_cancelled" {
				cancelled = true
			}
			if w.Type == "continued" && !cancelled {
				t.Fatal("explicit Continue did not cancel owned speech immediately")
			}
			return w.Type == "continued"
		})
		var authority struct {
			Client string
			Epoch  uint64
		}
		decode(w, &authority)
		if w.AgentID != a.ID() || authority.Client != ids[1] || authority.Epoch != before.Epoch {
			t.Fatalf("wrong Continue authority: %+v", authority)
		}
		readUntil(ws, func(w Wire) bool {
			return w.Type == "state_changed" && w.State != nil && !w.State.Paused && w.State.ContinueEpoch > before.ContinueEpoch
		})
	}
	write(first, map[string]any{"type": "speech_done", "speech_id": requested.ID, "status": "done"})
	write(first, map[string]any{"type": "pause"})
	readUntil(first, func(w Wire) bool {
		if w.Type == "speech_requested" || w.Type == "speech_claimed" {
			t.Fatal("late completion resurrected cancelled speech")
		}
		return w.Type == "state_changed" && w.State != nil && w.State.Paused
	})
}
