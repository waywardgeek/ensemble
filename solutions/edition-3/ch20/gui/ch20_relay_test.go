package gui_test

import (
	"encoding/json"
	"io"
	"net"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ensemble"
	"ensemble/gui"
	"github.com/gorilla/websocket"
)

func TestCh20RelaySessionCorrelation(t *testing.T) {
	root := ensemble.New(nil)
	defer root.Shutdown()
	base := t.TempDir()
	agent, err := root.NewAgent(ensemble.Config{ID: "root", Model: "offline", BaseDir: base, DataDir: filepath.Join(base, "root"), SavePath: filepath.Join(base, "root", "save.json")})
	if err != nil {
		t.Fatal(err)
	}
	app, err := gui.New(root, agent, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	httpServer := httptest.NewServer(app.Handler(""))
	defer httpServer.Close()
	browser, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(httpServer.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer browser.Close()
	_ = browser.SetReadDeadline(time.Now().Add(5 * time.Second))
	// The replay proves the connection has entered the Hub before relay admission.
	if _, _, err := browser.ReadMessage(); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- app.ServeMCP(listener) }()
	streams := make([]ensemble.MCPTransport, 2)
	for i := range streams {
		conn, err := net.DialTimeout("tcp", listener.Addr().String(), time.Second)
		if err != nil {
			t.Fatal(err)
		}
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
		stream := ensemble.NewMCPStream(conn)
		streams[i] = stream
		defer stream.Close()
		request, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 7, "method": "fixture", "params": map[string]int{"session": i}})
		if err := stream.Send(request); err != nil {
			t.Fatal(err)
		}
	}
	type frame struct {
		Type    string          `json:"type"`
		AgentID string          `json:"agent_id"`
		Payload json.RawMessage `json:"payload"`
	}
	var routed []frame
	for len(routed) < 2 {
		_, raw, err := browser.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}
		var f frame
		if err := json.Unmarshal(raw, &f); err != nil {
			t.Fatal(err)
		}
		if f.Type == "jsonrpc" {
			routed = append(routed, f)
		}
	}
	if routed[0].AgentID == routed[1].AgentID {
		t.Fatal("relay routing key reused")
	}
	for i := len(routed) - 1; i >= 0; i-- {
		f := routed[i]
		var request struct {
			Params struct {
				Session int `json:"session"`
			} `json:"params"`
		}
		if err := json.Unmarshal(f.Payload, &request); err != nil {
			t.Fatal(err)
		}
		reply := map[string]any{"type": "jsonrpc", "agent_id": f.AgentID, "payload": map[string]any{"jsonrpc": "2.0", "id": 7, "result": request.Params.Session}}
		if err := browser.WriteJSON(reply); err != nil {
			t.Fatal(err)
		}
	}
	for i, stream := range streams {
		raw, err := stream.Recv()
		if err != nil {
			t.Fatal(err)
		}
		var reply struct {
			ID     int `json:"id"`
			Result int `json:"result"`
		}
		if err := json.Unmarshal(raw, &reply); err != nil {
			t.Fatal(err)
		}
		if reply.ID != 7 || reply.Result != i {
			t.Fatalf("session %d got %s", i, raw)
		}
	}
	app.Close()
	select {
	case err := <-served:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("relay did not stop")
	}
	if err := agent.Post(ensemble.Hint{Text: "view closed, Agent remains owned by application"}); err != nil {
		t.Fatal(err)
	}
}
