package gui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ensemble"
	"github.com/gorilla/websocket"
)

func TestMCPBrowserReloadSameSocket(t *testing.T) {
	var providerCalls atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providerCalls.Add(1)
		http.Error(w, "no provider calls expected", 500)
	}))
	defer provider.Close()
	t.Setenv("LLM_BASE_URL", provider.URL)
	t.Setenv("LLM_MODEL", "claude-fixture")
	t.Setenv("LLM_API_KEY", "fixture-not-a-key")
	dir := t.TempDir()
	for name, text := range map[string]string{
		"base":    "---\nname: base\ntype: primary\ndescription: base\nloadable-skills: browser\n---\nbase",
		"browser": "---\nname: browser\ntype: loadable\ndescription: browser\ntools: gui_snapshot\nmcp_servers:\n  - name: browser-debug\n    transport: websocket\n---\nRetain this browser manual while unloaded.",
	} {
		path := filepath.Join(dir, name, "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	en := ensemble.New(log.New(io.Discard, "", 0))
	defer en.Shutdown()
	a, err := en.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "offline", APIKey: "fixture", BaseURL: provider.URL, BaseDir: dir, DataDir: t.TempDir(), Skills: &ensemble.SkillSpec{Directory: dir, Primary: "base"}})
	if err != nil {
		t.Fatal(err)
	}
	server, err := NewServerWithSettings(en, "web/gui", filepath.Join(t.TempDir(), "settings.json"))
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
	var writeMu sync.Mutex
	write := func(value any) error { writeMu.Lock(); defer writeMu.Unlock(); return conn.WriteJSON(value) }
	frames := make(chan Wire, 128)
	failures := make(chan error, 1)
	stopped := make(chan struct{})
	var handshakes, calls atomic.Int32
	go func() {
		defer close(stopped)
		for {
			var wire Wire
			if err := conn.ReadJSON(&wire); err != nil {
				select {
				case failures <- err:
				default:
				}
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
				handshakes.Add(1)
				result = map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{"tools": map[string]any{}}}
			case "tools/list":
				result = map[string]any{"tools": []map[string]any{
					{"name": "gui_snapshot", "description": "fixture DOM observation", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false}},
					{"name": "tts_queue", "description": "fixture queue", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false}},
					{"name": "gui_click", "description": "fixture click", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"selector": map[string]any{"type": "string"}}, "required": []string{"selector"}, "additionalProperties": false}},
					{"name": "gui_input", "description": "fixture input", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"selector": map[string]any{"type": "string"}, "text": map[string]any{"type": "string"}}, "required": []string{"selector", "text"}, "additionalProperties": false}},
					{"name": "ungranted_browser_auto", "description": "not granted", "ephemeral": "round", "inputSchema": map[string]any{"type": "object"}},
				}}
			case "tools/call":
				calls.Add(1)
				result = map[string]any{"content": []map[string]any{{"type": "text", "text": fmt.Sprintf("same socket generation %d", handshakes.Load())}}}
			default:
				failures <- fmt.Errorf("unexpected method %s", request.Method)
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
					t.Fatalf("GUI error: %+v", frame)
				}
				if predicate(frame) {
					return
				}
			case err := <-failures:
				t.Fatal(err)
			case <-timer.C:
				t.Fatal("GUI protocol condition not reached")
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
	for generation := int32(1); generation <= 2; generation++ {
		send("load_skill")
		await(func(frame Wire) bool {
			if frame.Type != "skill_state" {
				return false
			}
			for _, skill := range frame.Skills {
				if skill.Name == "browser" {
					return skill.Enabled && !skill.PendingUnload
				}
			}
			return false
		})
		if handshakes.Load() != generation {
			t.Fatalf("handshakes %d, want %d", handshakes.Load(), generation)
		}
		before := calls.Load()
		round, turn := a.MCP().Gather(context.Background(), true)
		if len(round)+len(turn) != 0 || calls.Load() != before {
			t.Fatalf("ungranted automatic observation: %v %v", round, turn)
		}
		result, failed := a.Tools().Dispatch(context.Background(), "gui_snapshot", json.RawMessage(`{}`))
		data, _ := json.Marshal(result)
		if failed || !strings.Contains(string(data), fmt.Sprintf("same socket generation %d", generation)) {
			t.Fatalf("real tunneled job result: %s", data)
		}
		if generation == 2 {
			break
		}
		send("unload_skill")
		closed, pending := false, false
		await(func(frame Wire) bool {
			if frame.Type == "mcp_closed" {
				closed = true
			}
			if frame.Type == "skill_state" {
				for _, skill := range frame.Skills {
					if skill.Name == "browser" {
						pending = skill.Enabled && skill.PendingUnload
					}
				}
			}
			return closed && pending
		})
		refused, failed := a.Tools().Dispatch(context.Background(), "gui_snapshot", json.RawMessage(`{}`))
		if len(refused) != 1 || !failed {
			t.Fatalf("unloaded tool still executable: %+v", refused)
		}
	}
	if calls.Load() != 2 {
		t.Fatalf("unexpected browser calls: %d", calls.Load())
	}
	if providerCalls.Load() != 0 {
		t.Fatalf("unexpected provider calls: %d", providerCalls.Load())
	}
	conn.Close()
	<-stopped
}
