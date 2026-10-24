package gui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ensemble"
	"github.com/gorilla/websocket"
)

func r4Dial(t *testing.T, host string) *websocket.Conn {
	t.Helper()
	ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(host, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ws.Close() })
	_ = ws.SetReadDeadline(time.Now().Add(4 * time.Second))
	return ws
}

func TestCh7R4PongPauseOrderingAndCleanup(t *testing.T) {
	entered, release, effect := make(chan struct{}), make(chan struct{}), make(chan struct{})
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()
	var requests atomic.Int32
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if requests.Add(1) == 1 {
			close(entered)
			select {
			case <-release:
			case <-r.Context().Done():
				return
			}
			fmt.Fprint(w, `{"choices":[{"message":{"role":"assistant","tool_calls":[{"id":"pause-call","type":"function","function":{"name":"effect","arguments":"{}"}}]},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":2,"completion_tokens":1}}`)
		} else {
			fmt.Fprint(w, `{"choices":[{"message":{"role":"assistant","content":"done"},"finish_reason":"stop"}],"usage":{"prompt_tokens":3,"completion_tokens":1}}`)
		}
	}))
	t.Cleanup(peer.Close)
	app := ensemble.New(nil)
	t.Cleanup(func() { _ = app.Shutdown() })
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "offline", APIKey: "fake", BaseURL: peer.URL, StreamCapability: ensemble.StreamPlain, BaseDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Tools().Register(ensemble.ToolDefinition{Name: "effect", Description: "local ordering probe", Parameters: json.RawMessage(`{"type":"object"}`)}, func(context.Context, ensemble.ToolCall, json.RawMessage) error { close(effect); return nil }); err != nil {
		t.Fatal(err)
	}
	s, err := NewServer(app, "web/gui")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	host := httptest.NewServer(s.Handler(s.Assets))
	t.Cleanup(host.Close)
	ws := r4Dial(t, host.URL)
	pongs, ended, readerDone := make(chan string, 4), make(chan struct{}, 1), make(chan struct{})
	ws.SetPongHandler(func(payload string) error { pongs <- payload; return nil })
	go func() {
		defer close(readerDone)
		for {
			var w Wire
			if err := ws.ReadJSON(&w); err != nil {
				return
			}
			if w.Type == "turn_ended" {
				ended <- struct{}{}
			}
		}
	}()
	if err := ws.WriteJSON(command{Type: "prompt", AgentID: a.ID(), Text: "run the effect"}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("model request did not start")
	}
	if err := ws.WriteJSON(command{Type: "pause", AgentID: a.ID()}); err != nil {
		t.Fatal(err)
	}
	for _, payload := range []string{"pause\x00\xff", "", strings.Repeat("p", 125)} {
		if err := ws.WriteControl(websocket.PingMessage, []byte(payload), time.Now().Add(time.Second)); err != nil {
			t.Fatal(err)
		}
		select {
		case got := <-pongs:
			if got != payload {
				t.Fatalf("Pong payload %q, want %q", got, payload)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("no matching Pong after pause command")
		}
	}
	close(release)
	eventually(t, func() bool {
		for _, e := range a.Log().EventsRange(0, a.View().Through) {
			if e.Type == ensemble.ResponseEnded {
				return true
			}
		}
		return false
	})
	select {
	case <-effect:
		t.Fatal("tool ran through the pause/Ping barrier")
	case <-time.After(100 * time.Millisecond):
	}
	if err := ws.WriteJSON(command{Type: "unpause", AgentID: a.ID()}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-effect:
	case <-time.After(2 * time.Second):
		t.Fatal("tool did not resume")
	}
	select {
	case <-ended:
	case <-time.After(2 * time.Second):
		t.Fatal("turn did not finish")
	}
	if requests.Load() != 2 {
		t.Fatalf("model requests=%d", requests.Load())
	}
	_ = ws.Close()
	select {
	case <-readerDone:
	case <-time.After(time.Second):
		t.Fatal("client reader did not stop")
	}
	eventually(t, func() bool { s.Hub.mu.Lock(); defer s.Hub.mu.Unlock(); return len(s.Hub.clients) == 0 })
}

type r4GatedTrace struct {
	entered, release chan struct{}
	block, open      sync.Once
}

func (g *r4GatedTrace) Write(p []byte) (int, error) {
	g.block.Do(func() { close(g.entered); <-g.release })
	return len(p), nil
}
func (g *r4GatedTrace) unblock() { g.open.Do(func() { close(g.release) }) }

func TestCh7R4PongOverloadStopsClient(t *testing.T) {
	app, _ := idleAgent(t, false)
	s, err := NewServer(app, "web/gui")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	gate := &r4GatedTrace{entered: make(chan struct{}), release: make(chan struct{})}
	s.Trace = gate
	host := httptest.NewServer(s.Handler(s.Assets))
	t.Cleanup(host.Close)
	t.Cleanup(gate.unblock)
	ws := r4Dial(t, host.URL)
	select {
	case <-gate.entered:
	case <-time.After(2 * time.Second):
		t.Fatal("socket writer did not reach trace gate")
	}
	s.Hub.mu.Lock()
	var c *client
	for c = range s.Hub.clients {
		break
	}
	s.Hub.mu.Unlock()
	if c == nil {
		t.Fatal("no connected client")
	}
	pongs, readerDone := make(chan struct{}, 128), make(chan struct{})
	ws.SetPongHandler(func(string) error { pongs <- struct{}{}; return nil })
	go func() {
		defer close(readerDone)
		for {
			if _, _, err := ws.ReadMessage(); err != nil {
				return
			}
		}
	}()
	// The existing writer is held before its initial replay write. Control
	// admission must remain bounded/nonblocking, not write from this reader.
	for i := 0; i < 128; i++ {
		if err := ws.WriteControl(websocket.PingMessage, []byte(fmt.Sprint(i)), time.Now().Add(time.Second)); err != nil {
			break
		}
	}
	select {
	case <-c.done:
	case <-time.After(2 * time.Second):
		t.Fatal("control overload did not stop client while writer was held")
	}
	select {
	case <-pongs:
		t.Fatal("Pong bypassed the held sole writer")
	default:
	}
	gate.unblock()
	select {
	case <-readerDone:
	case <-time.After(2 * time.Second):
		t.Fatal("overloaded socket did not close")
	}
	eventually(t, func() bool { s.Hub.mu.Lock(); defer s.Hub.mu.Unlock(); return len(s.Hub.clients) == 0 })
}

func TestCh7R4EmptyStartLiveAndReplay(t *testing.T) {
	for _, kind := range []string{"text", "thinking"} {
		t.Run(kind, func(t *testing.T) {
			release := make(chan struct{})
			defer func() {
				select {
				case <-release:
				default:
					close(release)
				}
			}()
			peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				send := func(v string) { fmt.Fprintf(w, "data: %s\n\n", v); w.(http.Flusher).Flush() }
				send(`{"type":"message_start","message":{"model":"claude-sonnet-4-6","usage":{"input_tokens":2}}}`)
				send(fmt.Sprintf(`{"type":"content_block_start","index":0,"content_block":{"type":%q,%q:""}}`, kind, kind))
				select {
				case <-release:
				case <-r.Context().Done():
					return
				}
				send(fmt.Sprintf(`{"type":"content_block_delta","index":0,"delta":{"type":%q,%q:"readable content"}}`, kind+"_delta", kind))
				if kind == "thinking" {
					send(`{"type":"content_block_delta","index":0,"delta":{"type":"signature_delta","signature":"private-signature"}}`)
				}
				send(`{"type":"content_block_stop","index":0}`)
				send(`{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":3}}`)
				send(`{"type":"message_stop"}`)
			}))
			t.Cleanup(peer.Close)
			app := ensemble.New(nil)
			t.Cleanup(func() { _ = app.Shutdown() })
			a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.Anthropic, Model: "claude-sonnet-4-6", APIKey: "fake", BaseURL: peer.URL, StreamCapability: ensemble.StreamSSE, BaseDir: t.TempDir()})
			if err != nil {
				t.Fatal(err)
			}
			s, err := NewServer(app, "web/gui")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(s.Close)
			host := httptest.NewServer(s.Handler(s.Assets))
			t.Cleanup(host.Close)
			live := r4Dial(t, host.URL)
			read := func(ws *websocket.Conn) map[string]any {
				var m map[string]any
				if err := ws.ReadJSON(&m); err != nil {
					t.Fatal(err)
				}
				return m
			}
			if m := read(live); m["type"] != "replay" {
				t.Fatal("missing initial replay")
			}
			q := a.Observe()
			defer q.Close()
			reply := a.Submit(context.Background(), "empty start")
			var first map[string]any
			for {
				first = read(live)
				if first["type"] == "part_delta" && first["kind"] == kind {
					break
				}
			}
			if chunk, ok := first["chunk"].(string); !ok || chunk != "" {
				t.Fatalf("empty live start omitted/changed chunk: %v", first)
			}
			id, ok := first["part_id"].(float64)
			if !ok {
				t.Fatal("missing display ID")
			}
			replaySocket := r4Dial(t, host.URL)
			batch := read(replaySocket)
			if batch["type"] != "replay" {
				t.Fatal("missing batch replay")
			}
			partials := batch["data"].(map[string]any)["partials"].([]any)
			if len(partials) != 1 {
				t.Fatalf("active partials=%v", partials)
			}
			partial := partials[0].(map[string]any)
			if chunk, ok := partial["chunk"].(string); !ok || chunk != "" || partial["part_id"] != id || partial["kind"] != kind {
				t.Fatalf("empty replay start lost envelope/identity: %v", partial)
			}
			close(release)
			for _, ws := range []*websocket.Conn{live, replaySocket} {
				var text strings.Builder
				for {
					m := read(ws)
					if m["type"] == "part_delta" && m["kind"] == kind {
						chunk, ok := m["chunk"].(string)
						if !ok || m["part_id"] != id {
							t.Fatal("later delta lost chunk/identity", m)
						}
						text.WriteString(chunk)
					}
					if m["type"] == "part_final" && m["part_id"] == id {
						part := m["part"].(map[string]any)
						if part["type"] != kind || part["text"] != "readable content" {
							t.Fatal("final changed content/type", m)
						}
						break
					}
				}
				// Wakeups may coalesce the last live delta into the durable final.
				if text.Len() != 0 && text.String() != "readable content" {
					t.Fatalf("partial content=%q", text.String())
				}
			}
			if r := <-reply; r.Err != nil {
				t.Fatal(r.Err)
			}
			finals := 0
			for _, o := range q.Drain() {
				if f, ok := o.(ensemble.PartFinal); ok {
					finals++
					if f.PartID != uint64(id) || f.Part.DisplayID != uint64(id) || f.Part.ID != 0 {
						t.Fatalf("native/display identity changed: %+v", f)
					}
				}
			}
			if finals != 1 {
				t.Fatalf("model finals=%d", finals)
			}
			for _, e := range a.Log().EventsRange(0, a.View().Through) {
				if e.Type == ensemble.ResponseEnded {
					for _, p := range e.Response.Parts {
						if p.ID != 0 || p.DisplayID != uint64(id) {
							t.Fatal("saved identity changed")
						}
						if kind == "thinking" && !strings.Contains(string(p.Data), "private-signature") {
							t.Fatal("thinking signature lost from history")
						}
					}
				}
			}
		})
	}
}
