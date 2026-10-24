package gui

import (
	"bytes"
	"context"
	"encoding/json"
	"ensemble"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func streamingAgent(t *testing.T, url string) (*ensemble.Ensemble, ensemble.Agent) {
	t.Helper()
	app := ensemble.New(nil)
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "offline", APIKey: "fake", BaseURL: url, StreamCapability: ensemble.StreamSSE, BaseDir: t.TempDir()})
	if err != nil {
		_ = app.Shutdown()
		t.Fatal(err)
	}
	return app, a
}

func TestCh7R2ReplayMinimumVisibleEvents(t *testing.T) {
	app, a := idleAgent(t, false)
	var log bytes.Buffer
	enc := json.NewEncoder(&log)
	_ = enc.Encode(map[string]int{"log_version": 1})
	seq := ensemble.Seq(0)
	old := time.Now().Add(-2 * time.Hour)
	add := func(e ensemble.Event) {
		seq++
		e.Seq = seq
		e.Time = old
		if err := enc.Encode(e); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 120; i++ {
		add(ensemble.Event{Type: ensemble.MessageReceived, Message: &ensemble.MessageData{Actor: "human", Parts: ensemble.PartList{{Type: "text", Text: fmt.Sprintf("visible-%03d", i), DisplayID: uint64(i + 1)}}}})
		add(ensemble.Event{Type: ensemble.TurnEnded, Message: &ensemble.MessageData{Actor: "agent", Parts: ensemble.PartList{}}})
		add(ensemble.Event{Type: ensemble.ToolReturned, Stale: true, Tool: &ensemble.ToolData{Name: "stale", Parts: ensemble.PartList{{Type: "text", Text: "not visible"}}}})
		add(ensemble.Event{Type: ensemble.ToolReturned, Stale: true, Tool: &ensemble.ToolData{Name: "also-stale", Parts: ensemble.PartList{{Type: "text", Text: "not visible"}}}})
		add(ensemble.Event{Type: ensemble.HintReceived, Message: &ensemble.MessageData{Actor: ensemble.System, Parts: ensemble.PartList{}}})
	}
	if err := a.Log().Load(&log); err != nil {
		t.Fatal(err)
	}
	_ = a.Post(ensemble.Ephemeral{Text: "current"})
	eventually(t, func() bool { return a.View().Through > seq })
	h := NewHub(app)
	defer h.Close()
	c := h.connect()
	defer h.disconnect(c)
	replay := replayData(t, c)
	var visible []string
	for _, w := range replay.Events {
		if w.Type == "part_final" {
			visible = append(visible, decodeCard(t, w).Text)
		}
	}
	if len(visible) != 100 {
		t.Fatalf("visible replay events=%d want100", len(visible))
	}
	if visible[0] != "visible-021" || visible[99] != "current" {
		t.Fatalf("wrong window endpoints %q .. %q", visible[0], visible[99])
	}
}

func TestCh7R2PublicWireFields(t *testing.T) {
	next, finish := make(chan struct{}), make(chan struct{})
	defer func() {
		select {
		case <-next:
		default:
			close(next)
		}
		select {
		case <-finish:
		default:
			close(finish)
		}
	}()
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		send := func(v string) { fmt.Fprintf(w, "data: %s\n\n", v); w.(http.Flusher).Flush() }
		send(`{"choices":[{"delta":{"content":"hel"}}]}`)
		select {
		case <-next:
		case <-r.Context().Done():
			return
		}
		send(`{"choices":[{"delta":{"content":"lo"}}]}`)
		select {
		case <-finish:
		case <-r.Context().Done():
			return
		}
		send(`{"choices":[{"delta":{},"finish_reason":"stop"}]}`)
		send(`{"choices":[],"usage":{"prompt_tokens":1,"completion_tokens":2}}`)
		send(`[DONE]`)
	}))
	defer peer.Close()
	app, a := streamingAgent(t, peer.URL)
	defer app.Shutdown()
	s, err := NewServer(app, "web/gui")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	host := httptest.NewServer(s.Handler(s.Assets))
	defer host.Close()
	ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(host.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	_ = ws.SetReadDeadline(time.Now().Add(4 * time.Second))
	waiting := "replay"
	read := func() map[string]any {
		var m map[string]any
		if err := ws.ReadJSON(&m); err != nil {
			t.Fatalf("waiting for %s: %v", waiting, err)
		}
		if m["type"] == "error" {
			t.Fatalf("waiting for %s: server error %v", waiting, m)
		}
		return m
	}
	replay := read()
	if replay["type"] != "replay" {
		t.Fatal("missing replay")
	}
	for _, item := range replay["data"].(map[string]any)["events"].([]any) {
		m := item.(map[string]any)
		if m["type"] == "state_changed" && (m["from"] != m["to"] || m["state"] == nil || m["data"] != nil) {
			t.Fatal("nonpublic replay state", m)
		}
	}
	until := func(kind string) map[string]any {
		waiting = kind
		for {
			m := read()
			if m["type"] == "state_changed" {
				if _, ok := m["from"].(string); !ok {
					t.Fatal("missing from", m)
				}
				if _, ok := m["to"].(string); !ok {
					t.Fatal("missing to", m)
				}
				if m["data"] != nil {
					t.Fatal("nested private state", m)
				}
			}
			if m["type"] == kind {
				return m
			}
		}
	}
	if err := ws.WriteJSON(map[string]any{"type": "prompt", "agent_id": a.ID(), "text": "wire stream"}); err != nil {
		t.Fatal(err)
	}
	first := until("part_delta")
	if first["kind"] != "text" || first["chunk"] != "hel" || first["data"] != nil {
		t.Fatal("nonpublic first delta", first)
	}
	close(next)
	second := until("part_delta")
	if second["kind"] != "text" || second["chunk"] != "lo" || second["part_id"] != first["part_id"] || second["data"] != nil {
		t.Fatal("delta must be incremental with stable ID", second)
	}
	close(finish)
	final := until("part_final")
	part, ok := final["part"].(map[string]any)
	if !ok || part["type"] != "text" || part["text"] != "hello" || final["part_id"] != first["part_id"] || final["data"] != nil {
		t.Fatal("nonpublic final", final)
	}
	ended := until("turn_ended")
	if ended["text"] != "hello" || ended["data"] != nil {
		t.Fatal("turn text missing or numeric", ended)
	}
	err = a.Tools().Register(ensemble.ToolDefinition{Name: "wire_effect", Description: "wire probe", Parameters: json.RawMessage(`{"type":"object"}`)}, func(ctx context.Context, call ensemble.ToolCall, _ json.RawMessage) error {
		_, err := io.WriteString(call, "tool-readable")
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		done <- a.Tools().Invoke(context.Background(), ensemble.Part{Type: "tool_call", CallID: "wire-call", Name: "wire_effect", Args: json.RawMessage(`{"x":1}`)}, "")
	}()
	dispatched := until("tool_dispatched")
	input, ok := dispatched["input"].(map[string]any)
	if !ok || input["x"] != float64(1) || dispatched["call_id"] != "wire-call" || dispatched["name"] != "wire_effect" || dispatched["data"] != nil {
		t.Fatal("nonpublic dispatch", dispatched)
	}
	finished := until("tool_finished")
	isError, ok := finished["is_error"].(bool)
	result, hasResult := finished["result"].(string)
	if !ok || isError || !hasResult || !strings.Contains(result, "tool-readable") || finished["data"] != nil {
		t.Fatal("nonpublic result", finished)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestCh7R2AbsoluteAssetsOutsideCWD(t *testing.T) {
	assets, err := filepath.Abs("web/gui")
	if err != nil {
		t.Fatal(err)
	}
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	app, _ := idleAgent(t, false)
	s, err := NewServer(app, assets)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	host := httptest.NewServer(s.Handler(s.Assets))
	defer host.Close()
	for _, path := range []string{"/", "/app.js", "/style.css"} {
		res, err := http.Get(host.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil || res.StatusCode != 200 || len(body) == 0 {
			t.Fatalf("asset %s: status=%d bytes=%d err=%v", path, res.StatusCode, len(body), err)
		}
	}
}

func TestCh7R2TerminalInterrupt(t *testing.T) {
	entered := make(chan struct{})
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"waiting\"}}]}\n\n")
		w.(http.Flusher).Flush()
		close(entered)
		<-r.Context().Done()
	}))
	defer peer.Close()
	app, a := streamingAgent(t, peer.URL)
	defer app.Shutdown()
	s, err := NewServer(app, "web/gui")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.DefaultAgent = a.ID()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	s.Terminal(ctx, reader, io.Discard)
	result := a.Submit(context.Background(), "wait for interrupt")
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("request did not enter")
	}
	if _, err := io.WriteString(writer, "/interrupt\n"); err != nil {
		t.Fatal(err)
	}
	writer.Close()
	select {
	case r := <-result:
		if r.Err == nil {
			t.Fatal("/interrupt did not cancel active request")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("terminal /interrupt not handled")
	}
	host := httptest.NewServer(s.Handler(s.Assets))
	defer host.Close()
	res, err := http.Get(host.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatal("EOF stopped GUI hosting")
	}
}
