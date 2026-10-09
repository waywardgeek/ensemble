package gui

import (
	"context"
	"encoding/json"
	"example.com/ensemble"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func sessionSocket(t *testing.T, owner ensemble.ClientOwner, id string) (*Server, *websocket.Conn) {
	t.Helper()
	httpServer := httptest.NewUnstartedServer(nil)
	origin := "http://" + httpServer.Listener.Addr().String()
	server, err := NewServer(owner, id, origin, nil, ServerOptions{PreferencesPath: filepath.Join(t.TempDir(), "preferences.json")})
	if err != nil {
		t.Fatal(err)
	}
	httpServer.Config.Handler = server
	httpServer.Start()
	t.Cleanup(httpServer.Close)
	t.Cleanup(func() { server.Close() })
	socket, _, err := websocket.DefaultDialer.Dial(strings.Replace(origin, "http:", "ws:", 1)+"/ws", http.Header{"Origin": []string{origin}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { socket.Close() })
	return server, socket
}
func frame(t *testing.T, s *websocket.Conn) map[string]json.RawMessage {
	t.Helper()
	s.SetReadDeadline(time.Now().Add(2 * time.Second))
	var out map[string]json.RawMessage
	if err := s.ReadJSON(&out); err != nil {
		t.Fatal(err)
	}
	return out
}
func value(f map[string]json.RawMessage, key string) string {
	var s string
	json.Unmarshal(f[key], &s)
	return s
}
func subscribeSession(t *testing.T, s *websocket.Conn) map[string]json.RawMessage {
	t.Helper()
	if err := s.WriteJSON(map[string]string{"type": "subscribe", "id": "sub"}); err != nil {
		t.Fatal(err)
	}
	var state map[string]json.RawMessage
	for {
		f := frame(t, s)
		if value(f, "type") == "snapshot_begin" {
			if err := json.Unmarshal(f["state"], &state); err != nil {
				t.Fatal(err)
			}
		}
		if value(f, "type") == "snapshot_end" {
			break
		}
	}
	return state
}
func TestSessionBrowserCheckpointOrder(t *testing.T) {
	root := ensemble.New(io.Discard)
	defer root.Close()
	a, err := root.OpenSession(ensemble.SessionOptions{Config: ensemble.Config{DataDir: filepath.Join(t.TempDir(), "session"), APIKey: "local-only", Model: "fixture"}})
	if err != nil {
		t.Fatal(err)
	}
	_, socket := sessionSocket(t, root, a.ID())
	state := subscribeSession(t, socket)
	var session ensemble.SessionState
	if err = json.Unmarshal(state["session"], &session); err != nil || session.ID == "" || session.CheckpointSeq != nil {
		t.Fatal("session state", err)
	}
	if string(state["job_access"]) != "[]" {
		t.Fatal("missing separate job access")
	}
	if err = socket.WriteJSON(map[string]string{"type": "checkpoint", "id": "save"}); err != nil {
		t.Fatal(err)
	}
	var applied uint64
	for {
		f := frame(t, socket)
		switch value(f, "type") {
		case "observation":
			var o ensemble.Observation
			json.Unmarshal(f["observation"], &o)
			if o.Kind == "session_changed" {
				json.Unmarshal(f["revision"], &applied)
				if o.Session == nil || o.Session.ID != session.ID || o.Session.CheckpointSeq == nil || *o.Session.CheckpointSeq != 1 {
					t.Fatal("invalid applied session")
				}
			}
		case "command_ack":
			var revision, asof uint64
			json.Unmarshal(f["watch_revision"], &revision)
			json.Unmarshal(f["as_of"], &asof)
			if value(f, "id") != "save" || value(f, "status") != "saved" || revision != applied || applied == 0 || asof != 1 {
				t.Fatal("ack preceded applied state")
			}
			return
		case "command_error":
			t.Fatal("checkpoint refused", string(f["code"]))
		}
	}
}

type canceledSaveOwner struct {
	ensemble.ClientOwner
	entered chan struct{}
	left    chan struct{}
}

func (o *canceledSaveOwner) CheckpointContext(ctx context.Context, id string) (ensemble.CheckpointAck, error) {
	close(o.entered)
	<-ctx.Done()
	close(o.left)
	return ensemble.CheckpointAck{}, ctx.Err()
}
func TestSessionBrowserDisconnectReleasesSaveWait(t *testing.T) {
	root := ensemble.New(io.Discard)
	defer root.Close()
	a, err := root.OpenSession(ensemble.SessionOptions{Config: ensemble.Config{DataDir: filepath.Join(t.TempDir(), "session"), APIKey: "local-only", Model: "fixture"}})
	if err != nil {
		t.Fatal(err)
	}
	owner := &canceledSaveOwner{ClientOwner: root, entered: make(chan struct{}), left: make(chan struct{})}
	server, socket := sessionSocket(t, owner, a.ID())
	subscribeSession(t, socket)
	socket.WriteJSON(map[string]string{"type": "checkpoint", "id": "save"})
	<-owner.entered
	socket.Close()
	select {
	case <-owner.left:
	case <-time.After(time.Second):
		t.Fatal("connection retained save waiter")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err = server.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err = a.Checkpoint(); err != nil {
		t.Fatal("connection shutdown stopped Agent", err)
	}
}
