package gui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"ensemble"
	"github.com/gorilla/websocket"
)

func TestChapter12GenericBrowserTunnel(t *testing.T) {
	dir := t.TempDir()
	for name, text := range map[string]string{
		"base":    "---\nname: base\ntype: primary\ndescription: base\nloadable-skills: browser\n---\nbase",
		"browser": "---\nname: browser\ntype: loadable\ndescription: generic browser peer\ntools: peer_echo peer_round gui_click\nmcp_servers:\n  - name: browser-debug\n    transport: websocket\n---\nGENERIC:peer instructions",
	} {
		path := filepath.Join(dir, name, "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	root := ensemble.New(log.New(io.Discard, "", 0))
	defer root.Shutdown()
	a, err := root.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "fixture", APIKey: "fixture", BaseDir: dir, DataDir: t.TempDir(), Skills: &ensemble.SkillSpec{Directory: dir, Primary: "base"}})
	if err != nil {
		t.Fatal(err)
	}
	// Unknown required names must not be invented for an unattached browser.
	if err := a.Skills().LoadSkill("browser"); err == nil || !strings.Contains(err.Error(), "missing tool") {
		t.Fatalf("unknown cold catalog admitted: %v", err)
	}
	server, err := NewServerWithSettings(root, "web/gui", filepath.Join(t.TempDir(), "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	httpServer := httptest.NewServer(server.Handler(server.Assets))
	defer httpServer.Close()
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(httpServer.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	var writeMu, callsMu sync.Mutex
	write := func(value any) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		return conn.WriteJSON(value)
	}
	calls := map[string]int{}
	frames := make(chan Wire, 128)
	failures := make(chan error, 1)
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		for {
			var wire Wire
			if err := conn.ReadJSON(&wire); err != nil {
				failures <- err
				return
			}
			if wire.Type != "jsonrpc" {
				frames <- wire
				continue
			}
			var request struct {
				ID     json.RawMessage
				Method string
				Params struct{ Name string }
			}
			if err := json.Unmarshal(wire.Payload, &request); err != nil {
				failures <- err
				return
			}
			if len(request.ID) == 0 {
				continue
			}
			var result any
			switch request.Method {
			case "initialize":
				result = map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{"tools": map[string]any{}}}
			case "tools/list":
				result = map[string]any{"tools": []map[string]any{
					{"name": "peer_echo", "description": "ordinary peer action", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"message": map[string]any{"type": "string"}}}},
					{"name": "peer_round", "description": "hidden round context", "ephemeral": "round", "inputSchema": map[string]any{"type": "object"}},
					{"name": "gui_click", "description": "generic hidden turn context", "ephemeral": "turn", "inputSchema": map[string]any{"type": "object"}},
					{"name": "ungranted_context", "description": "not granted", "ephemeral": "round", "inputSchema": map[string]any{"type": "object"}},
				}}
			case "tools/call":
				callsMu.Lock()
				calls[request.Params.Name]++
				callsMu.Unlock()
				result = map[string]any{"content": []map[string]string{{"type": "text", "text": "PEER:" + request.Params.Name}}}
			default:
				failures <- fmt.Errorf("unexpected peer method %s", request.Method)
				return
			}
			if err := write(map[string]any{"type": "jsonrpc", "agent_id": a.ID(), "payload": map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": result}}); err != nil {
				failures <- err
				return
			}
		}
	}()
	await := func(predicate func(Wire) bool) {
		t.Helper()
		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()
		for {
			select {
			case frame := <-frames:
				if frame.Type == "error" {
					t.Fatalf("GUI tunnel error: %+v", frame)
				}
				if predicate(frame) {
					return
				}
			case err := <-failures:
				t.Fatal(err)
			case <-timer.C:
				t.Fatal("generic browser tunnel deadline")
			}
		}
	}
	send := func(kind string) {
		t.Helper()
		if err := write(map[string]any{"type": kind, "agent_id": a.ID(), "name": "browser"}); err != nil {
			t.Fatal(err)
		}
	}
	await(func(frame Wire) bool { return frame.Type == "replay" })
	send("mcp_open")
	// The command owner waits for this tunnel's real handshake before invoking
	// skill discovery; the peer is attached, not a cold four-tool expectation.
	send("load_skill")
	await(func(frame Wire) bool {
		for _, skill := range frame.Skills {
			if skill.Name == "browser" && skill.Enabled && !skill.PendingUnload {
				return true
			}
		}
		return false
	})
	ordinary := false
	for _, definition := range a.Tools().Definitions() {
		if definition.Name == "peer_echo" {
			ordinary = strings.Contains(string(definition.Parameters), "message")
		}
		if definition.Name == "peer_round" || definition.Name == "gui_click" || definition.Name == "ungranted_context" {
			t.Fatalf("hidden/ungranted tool exposed to model: %s", definition.Name)
		}
	}
	if !ordinary {
		t.Fatal("real peer's ordinary schema not discovered")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, failed := a.Tools().Dispatch(ctx, "peer_echo", json.RawMessage(`{"message":"requested"}`))
	if failed || !strings.Contains(fmt.Sprint(result), "PEER:peer_echo") {
		t.Fatalf("ordinary granted execution: %v failed=%v", result, failed)
	}
	for _, name := range []string{"peer_round", "gui_click", "ungranted_context", "gui_snapshot"} {
		if _, failed := a.Tools().Dispatch(ctx, name, json.RawMessage(`{}`)); !failed {
			t.Fatalf("guessed hidden/ungranted tool executable: %s", name)
		}
	}
	round, turn := a.MCP().Gather(ctx, false)
	if !strings.Contains(fmt.Sprint(round), "PEER:peer_round") || len(turn) != 0 {
		t.Fatalf("round-only context: %v / %v", round, turn)
	}
	round, turn = a.MCP().Gather(ctx, true)
	if !strings.Contains(fmt.Sprint(round), "PEER:peer_round") || !strings.Contains(fmt.Sprint(turn), "PEER:gui_click") {
		t.Fatalf("granted automatic context: %v / %v", round, turn)
	}
	callsMu.Lock()
	counts := fmt.Sprint(calls)
	correct := calls["peer_echo"] == 1 && calls["peer_round"] == 2 && calls["gui_click"] == 1 && len(calls) == 3
	callsMu.Unlock()
	if !correct {
		t.Fatalf("unexpected automatic/guessed calls: %s", counts)
	}
	send("unload_skill")
	await(func(frame Wire) bool {
		for _, skill := range frame.Skills {
			if skill.Name == "browser" && skill.PendingUnload {
				return true
			}
		}
		return false
	})
	if !strings.Contains(a.Skills().Snapshot().Instructions, "GENERIC:peer instructions") {
		t.Fatal("pending unload lost instructions")
	}
	if _, failed := a.Tools().Dispatch(ctx, "peer_echo", json.RawMessage(`{}`)); !failed {
		t.Fatal("pending unload left external action enabled")
	}
	round, turn = a.MCP().Gather(ctx, true)
	if len(round)+len(turn) != 0 {
		t.Fatalf("pending unload gathered external context: %v / %v", round, turn)
	}
	conn.Close()
	<-stopped
}
