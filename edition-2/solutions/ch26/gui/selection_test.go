package gui_test

import (
	"context"
	"encoding/json"
	gui "ensemble-gui-stub"
	"ensemble/ensemble"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type selectionInbox chan ensemble.FleetMessage

// Deliver retains human-directed messages while the actual fleet does its work.
func (inbox selectionInbox) Deliver(message ensemble.FleetMessage) error {
	inbox <- message
	return nil
}

// A submitted tool batch ends the turn without another provider reply. Its
// current state must survive selection and restart, while real work stays busy.
// This extends the independent reviewer's real-runtime/socket reproduction.
func TestSelectionSubmittedTurnState(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "NEXT_TASK") {
			close(entered)
			select {
			case <-release:
			case <-r.Context().Done():
				return
			}
		}
		json.NewEncoder(w).Encode(map[string]any{"content": []any{map[string]any{"type": "tool_use", "id": "submit-call", "name": "submit", "input": map[string]string{"result": "DONE"}}}, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
	}))
	defer provider.Close()
	dir := t.TempDir()
	cfg := ensemble.Config{Model: "fake", BaseURL: provider.URL, BuiltinTools: true}
	start := func() (*ensemble.Ensemble, ensemble.Agent) {
		e := ensemble.New(io.Discard)
		a, err := e.Start(dir, cfg, ensemble.AgentSpec{}, make(selectionInbox, 1))
		if err != nil {
			t.Fatal(err)
		}
		return e, a
	}
	app, root := start()
	defer func() { app.Close() }()
	peer, err := app.CreateAgent("console", "fake", "")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if result, err := peer.Ask(ctx, "TASK"); err != nil || result != "DONE" {
		t.Fatal(result, err)
	}
	check := func(stage, want string) {
		t.Helper()
		if got := peer.History().Context().Turn; got != want {
			t.Errorf("%s retained turn = %s, want %s", stage, got, want)
		}
		hub, err := gui.New(app, root, io.Discard)
		if err != nil {
			t.Fatal(err)
		}
		server := httptest.NewServer(hub.Handler(""))
		defer server.Close()
		defer hub.Close()
		conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/ws", nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		if err = conn.WriteJSON(map[string]string{"type": "select_agent", "key": "console"}); err != nil {
			t.Fatal(err)
		}
		frames := until(t, conn, kind("state_changed"))
		if got := frames[len(frames)-1]["to"]; got != want {
			t.Errorf("%s selected state = %v, want %s", stage, got, want)
		}
	}
	check("completed", "idle")
	if err = app.Close(); err != nil {
		t.Fatal(err)
	}
	app, root = start()
	peer, err = app.FindAgent("console", root)
	if err != nil {
		t.Fatal(err)
	}
	check("restarted", "idle")
	done := make(chan error, 1)
	go func() { _, err := peer.Ask(ctx, "NEXT_TASK"); done <- err }()
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	check("request in flight", "in_flight")
	close(release)
	if err = <-done; err != nil {
		t.Fatal(err)
	}
	check("next completion", "idle")
}

// Reviewer found repeated selection could fill the reliable replay queue while
// holding the observation mutex. A slow receiver must disconnect, not deadlock
// every Agent and Server.Close. Keep the real socket writer and its backpressure.
func TestSelectionSlowReceiverClose(t *testing.T) {
	root := ensemble.New(io.Discard)
	a, err := root.NewAgent(ensemble.Config{Model: "fake", DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	if err = a.Record(ensemble.Event{Type: "message_received", Message: &ensemble.MessageData{Actor: ensemble.Human, Parts: []ensemble.Part{{Type: "text", Text: strings.Repeat("x", 64*1024)}}}}); err != nil {
		t.Fatal(err)
	}
	hub, err := gui.New(root, a, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(hub.Handler(""))
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	for i := 0; i < 320; i++ {
		if err = conn.WriteJSON(map[string]string{"type": "select_agent", "key": "main"}); err != nil {
			break
		}
	}
	// Allow the admitted requests to fill the socket and application queues. The
	// peer intentionally reads no replay; no fixture replaces the actual pressure.
	time.Sleep(3 * time.Second)
	conn.Close()
	done := make(chan struct{})
	go func() { hub.Close(); close(done) }()
	select {
	case <-done:
		server.Close()
	case <-time.After(time.Second):
		t.Fatal("GUI close blocked after slow selection client disconnected")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err = a.Flush(ctx); err != nil {
		t.Fatal("slow receiver stranded actor", err)
	}
}
