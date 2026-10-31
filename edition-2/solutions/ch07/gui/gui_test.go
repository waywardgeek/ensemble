package gui_test

import (
	"bytes"
	"context"
	"encoding/json"
	gui "ensemble-gui-stub"
	"ensemble/ensemble"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// observationSink is a real display consumer; it records the public lifecycle
// while the real Agent, parser and tools perform the work.
type observationSink struct{ events chan ensemble.Observation }

// Observe retains a bounded test transcript without delaying execution.
func (s *observationSink) Observe(o ensemble.Observation) { s.events <- o }

type trace struct {
	mu sync.Mutex
	// Buffer retains transport trace bytes under mu.
	bytes.Buffer
}

// Write serializes concurrent transport log writes with transcript reads.
func (b *trace) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Buffer.Write(p)
}
func (b *trace) text() string { b.mu.Lock(); defer b.mu.Unlock(); return b.String() }
func dial(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	c, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(url, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	send(t, c, "subscribe", "")
	return c
}
func send(t *testing.T, c *websocket.Conn, kind, text string) {
	t.Helper()
	if err := c.WriteJSON(map[string]string{"type": kind, "text": text}); err != nil {
		t.Fatal(err)
	}
}
func until(t *testing.T, c *websocket.Conn, predicate func(map[string]any) bool) []map[string]any {
	t.Helper()
	c.SetReadDeadline(time.Now().Add(5 * time.Second))
	var out []map[string]any
	for {
		var m map[string]any
		if err := c.ReadJSON(&m); err != nil {
			t.Fatal("waiting for wire event:", err)
		}
		out = append(out, m)
		if predicate(m) {
			return out
		}
	}
}
func kind(value string) func(map[string]any) bool {
	return func(m map[string]any) bool { return m["type"] == value }
}
func frame(w http.ResponseWriter, value any) {
	data, _ := json.Marshal(value)
	fmt.Fprintf(w, "data: %s\n\n", data)
	w.(http.Flusher).Flush()
}
func delta(w http.ResponseWriter, value any) {
	frame(w, map[string]any{"model": "gpt-4o-mini", "choices": []any{map[string]any{"delta": value}}})
}
func end(w http.ResponseWriter) {
	frame(w, map[string]any{"choices": []any{map[string]any{"delta": map[string]any{}, "finish_reason": "stop"}}})
	frame(w, map[string]any{"choices": []any{}, "usage": map[string]int{"prompt_tokens": 2, "completion_tokens": 2}})
	io.WriteString(w, "data: [DONE]\n\n")
}
func setup(t *testing.T, provider *httptest.Server) (ensemble.Agent, *gui.Server, *httptest.Server, *trace) {
	t.Helper()
	root := ensemble.New(io.Discard)
	a, err := root.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-4o-mini", APIKey: "fake", BaseURL: provider.URL, BuiltinTools: true, Workspace: t.TempDir(), DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	log := &trace{}
	hub := gui.New(root, a, log)
	server := httptest.NewServer(hub.Handler("web/gui"))
	t.Cleanup(func() { a.Unpause(); a.Shutdown(); hub.Close(); server.Close() })
	return a, hub, server, log
}

// TestStreamReplayPause drives real HTTP parsing, actor, shell tools and websocket
// peers. The first HTTP stream and shell command have explicit release points,
// making reconnect and pause assertions causal rather than timing guesses.
func TestStreamReplayPause(t *testing.T) {
	release := make(chan struct{})
	var once sync.Once
	var requests atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		if requests.Add(1) == 1 {
			delta(w, map[string]string{"content": "Hello "})
			<-release
			delta(w, map[string]string{"content": "world."})
			calls := []any{}
			for i, command := range []string{"while [ ! -f release ]; do sleep .01; done; printf FIRST", "printf SECOND > second"} {
				args, _ := json.Marshal(map[string]any{"command": command})
				calls = append(calls, map[string]any{"index": i, "id": fmt.Sprint("call", i), "type": "function", "function": map[string]string{"name": "run_command", "arguments": string(args)}})
			}
			delta(w, map[string]any{"tool_calls": calls})
		} else {
			delta(w, map[string]string{"content": "Finished."})
		}
		end(w)
	}))
	defer func() { once.Do(func() { close(release) }); provider.Close() }()
	a, _, server, log := setup(t, provider)
	sink := &observationSink{events: make(chan ensemble.Observation, 256)}
	a.Observe(sink)
	first := dial(t, server.URL)
	until(t, first, kind("pause_changed"))
	send(t, first, "prompt", "exercise")
	partial := until(t, first, kind("part_delta"))
	firstID := partial[len(partial)-1]["part_id"]
	second := dial(t, server.URL)
	replay := until(t, second, kind("pause_changed"))
	found := false
	for _, m := range replay {
		if m["type"] == "part_delta" && m["chunk"] == "Hello " && m["part_id"] == firstID {
			found = true
		}
	}
	if !found {
		t.Fatal("reconnect lost in-flight partial")
	}
	send(t, second, "hint", "keep it short")
	until(t, first, func(m map[string]any) bool {
		return m["type"] == "user_message" && strings.Contains(fmt.Sprint(m["part"]), "keep it short")
	})
	once.Do(func() { close(release) })
	announced := false
	deadline := time.After(time.Second)
	for !announced {
		select {
		case o := <-sink.events:
			_, announced = o.(ensemble.ToolDispatched)
		case <-deadline:
			t.Fatal("observer did not announce dispatch before execution")
		}
	}
	until(t, first, kind("tool_dispatched"))
	send(t, first, "pause", "")
	until(t, first, func(m map[string]any) bool { return m["type"] == "pause_changed" && m["paused"] == true })
	during := dial(t, server.URL)
	for _, m := range until(t, during, kind("pause_changed")) {
		if m["type"] == "part_delta" {
			t.Fatal("finalized partial survived replay during tools")
		}
	}
	os.WriteFile(filepath.Join(a.Config().Workspace, "release"), nil, 0600)
	until(t, first, kind("tool_finished"))
	time.Sleep(100 * time.Millisecond) // Negative boundary: completion must not start tool two.
	if _, err := os.Stat(filepath.Join(a.Config().Workspace, "second")); err == nil {
		t.Fatal("pause did not hold next tool")
	}
	send(t, second, "unpause", "")
	ending := until(t, first, kind("turn_ended"))
	if _, err := os.Stat(filepath.Join(a.Config().Workspace, "second")); err != nil {
		t.Fatal("other tab did not release global pause")
	}
	if ending[len(ending)-1]["text"] != "Finished." {
		t.Fatal("wrong final reply")
	}
	// A second response reuses local parser ID 1, but may not replace the first card.
	send(t, first, "prompt", "another turn")
	next := until(t, first, kind("turn_ended"))
	for _, m := range next {
		if m["type"] == "part_delta" && m["part_id"] == firstID {
			t.Fatal("artifact identity collided across responses")
		}
	}
	third := dial(t, server.URL)
	history := until(t, third, kind("pause_changed"))
	seen := map[string]bool{}
	for _, m := range history {
		seen[fmt.Sprint(m["type"])] = true
		if m["type"] == "part_delta" {
			t.Fatal("finalized partial survived replay")
		}
	}
	for _, want := range []string{"part_final", "user_message", "tool_dispatched", "tool_finished", "state_changed"} {
		if !seen[want] {
			t.Fatal("replay missing", want)
		}
	}
	counts := map[string]int{"dispatch": 1}
	for len(sink.events) > 0 {
		switch (<-sink.events).(type) {
		case ensemble.ToolDispatched:
			counts["dispatch"]++
		case ensemble.ToolFinished:
			counts["finish"]++
		}
	}
	if counts["dispatch"] != 2 || counts["finish"] != 2 {
		t.Fatal("observer lost tool lifecycle", counts)
	}
	text := log.text()
	if !strings.Contains(text, " > {") || !strings.Contains(text, " < {") {
		t.Fatal("gui.log missing direction")
	}
	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
		fields := strings.SplitN(line, " ", 3)
		if len(fields) != 3 {
			t.Fatal("bad trace line")
		}
		if _, err := time.Parse(time.RFC3339Nano, fields[0]); err != nil || !json.Valid([]byte(fields[2])) {
			t.Fatal("bad timestamp or JSON")
		}
	}
}

// TestPausedInterruptAndReplayBatch checks interruption without unpausing and
// reliable replay larger than the live queue. A later turn uses the same Agent.
func TestPausedInterruptAndReplayBatch(t *testing.T) {
	var requests atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		if requests.Add(1) == 1 {
			delta(w, map[string]any{"tool_calls": []any{map[string]any{"index": 0, "id": "blocked", "function": map[string]string{"name": "run_command", "arguments": `{"command":"touch forbidden"}`}}}})
		} else {
			delta(w, map[string]string{"content": "usable"})
		}
		end(w)
	}))
	defer provider.Close()
	a, hub, server, _ := setup(t, provider)
	c := dial(t, server.URL)
	until(t, c, kind("pause_changed"))
	send(t, c, "pause", "")
	until(t, c, kind("pause_changed"))
	send(t, c, "prompt", "blocked")
	until(t, c, kind("part_final"))
	send(t, c, "interrupt", "")
	done := until(t, c, kind("turn_ended"))
	if !strings.Contains(fmt.Sprint(done[len(done)-1]["error"]), "interrupted") {
		t.Fatal("paused turn did not interrupt")
	}
	if _, err := os.Stat(filepath.Join(a.Config().Workspace, "forbidden")); err == nil {
		t.Fatal("paused tool executed")
	}
	if !a.Paused() {
		t.Fatal("interrupt implicitly unpaused Agent")
	}
	answer, err := a.Ask(context.Background(), "new request")
	if err != nil || answer != "usable" {
		t.Fatal("Agent not usable after interrupt", answer, err)
	}
	for i := 0; i < 270; i++ {
		if err := a.Record(ensemble.Event{Type: "message_received", Message: &ensemble.MessageData{Actor: ensemble.Human, Parts: []ensemble.Part{{Type: "text", Text: fmt.Sprint("history-", i)}}}}); err != nil {
			t.Fatal(err)
		}
	}
	d := dial(t, server.URL)
	replay := until(t, d, kind("pause_changed"))
	count := 0
	for _, m := range replay {
		if strings.Contains(fmt.Sprint(m["part"]), "history-") {
			count++
		}
	}
	if count != 270 {
		t.Fatal("replay silently dropped history", count)
	}
	// Exercise disconnect against active root delivery; data queues stay open.
	for i := 0; i < 20; i++ {
		peer := dial(t, server.URL)
		peer.Close()
		a.Record(ensemble.Event{Type: "message_received", Message: &ensemble.MessageData{Actor: ensemble.Human, Parts: []ensemble.Part{{Type: "text", Text: "disconnect"}}}})
	}
	hub.Close()
}

// TestReplayStatus checks that a new subscriber receives a complete, truthful
// current-state envelope even before the Agent has made any transition.
func TestReplayStatus(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer provider.Close()
	a, _, server, _ := setup(t, provider)
	peer := dial(t, server.URL)
	messages := until(t, peer, kind("state_changed"))
	status := messages[len(messages)-1]
	current := string(a.History().Context().Turn)
	if status["from"] != current || status["to"] != current || status["replay"] != true {
		t.Fatalf("replay status must describe current state %q at both endpoints: %v", current, status)
	}
}

// TestReplayWindow checks both sides of the fixed OR policy using real imported
// history timestamps. It does not sleep an hour or replace the history owner.
func TestReplayWindow(t *testing.T) {
	for _, recent := range []int{0, 110, 130} {
		t.Run(fmt.Sprint(recent), func(t *testing.T) {
			provider := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
			defer provider.Close()
			a, _, server, _ := setup(t, provider)
			var log bytes.Buffer
			encoder := json.NewEncoder(&log)
			for i := 0; i < 130; i++ {
				stamp := time.Now().Add(-2 * time.Hour)
				if i >= 130-recent {
					stamp = time.Now()
				}
				encoder.Encode(ensemble.Event{Seq: uint64(i + 1), Time: stamp, Type: "message_received", Message: &ensemble.MessageData{Actor: ensemble.Human, Parts: []ensemble.Part{{Type: "text", Text: fmt.Sprint(i)}}}})
			}
			if err := a.History().Load(&log); err != nil {
				t.Fatal(err)
			}
			peer := dial(t, server.URL)
			messages := until(t, peer, kind("pause_changed"))
			count := 0
			for _, m := range messages {
				if m["type"] == "user_message" {
					count++
				}
			}
			want := 100
			if recent > want {
				want = recent
			}
			if count != want {
				t.Fatalf("replay window got %d want %d", count, want)
			}
		})
	}
}
