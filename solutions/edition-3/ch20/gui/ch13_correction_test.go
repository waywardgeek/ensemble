package gui

import (
	"context"
	"ensemble"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCh13FailedToolSpeech(t *testing.T) {
	app, a := idleAgent(t, true)
	h := NewHub(app)
	defer h.Close()
	h.mu.Lock()
	defer h.mu.Unlock()
	v := h.view(a)
	h.narrate(v, 1, Card{Kind: "tool_result", Role: "tool", Text: "ordinary output", TurnID: 1}, true)
	if h.speech.active[a.ID()] != "" {
		t.Fatal("successful result narrated")
	}
	h.narrate(v, 2, Card{Kind: "tool_result", Role: "tool", Text: "permission denied", IsError: true, TurnID: 1}, true)
	task := h.speech.tasks[h.speech.active[a.ID()]]
	if task == nil || len(task.segments) != 1 || task.segments[0].Source != "error" || !strings.Contains(task.text, "permission denied") {
		t.Fatalf("failed tool not admitted as error: %+v", task)
	}
}
func TestCh13CancelledQueue(t *testing.T) {
	app, a := idleAgent(t, true)
	h := NewHub(app)
	defer h.Close()
	c := h.connect()
	defer h.disconnect(c)
	_ = replayData(t, c)
	h.mu.Lock()
	h.speech.enqueue(a.ID(), spokenPart{Text: "active"}, "automatic", 1)
	id := h.speech.active[a.ID()]
	h.mu.Unlock()
	if err := h.speech.claim(c, a.ID(), id); err != nil {
		t.Fatal(err)
	}
	h.mu.Lock()
	h.speech.enqueue(a.ID(), spokenPart{Text: "must not restart"}, "automatic", 1)
	h.mu.Unlock()
	if err := h.speech.complete(c, a.ID(), id, "cancelled"); err != nil {
		t.Fatal(err)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.speech.active[a.ID()] != "" || h.speech.pending[a.ID()] != nil {
		t.Fatal("cancelled owner queue promoted pending text")
	}

}
func TestCh13UnchangedPartialBoundary(t *testing.T) {
	next := make(chan struct{})
	release := make(chan struct{})
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"reasoning_content\":\"unfinished thought\"}}]}\n\n")
		w.(http.Flusher).Flush()
		select {
		case <-next:
		case <-r.Context().Done():
			return
		}
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Answer. \",\"role\":\"assistant\"}}]}\n\n")
		w.(http.Flusher).Flush()
		select {
		case <-release:
		case <-r.Context().Done():
			return
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	app := ensemble.New(nil)
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "local", APIKey: "not-real", BaseURL: peer.URL, StreamCapability: ensemble.StreamSSE, BaseDir: t.TempDir(), Speaking: true})
	if err != nil {
		t.Fatal(err)
	}
	defer peer.Close()
	defer app.Shutdown()
	h := NewHub(app)
	defer h.Close()
	q := a.Observe()
	defer q.Close()
	h.mu.Lock()
	v := h.view(a)
	result := a.Submit(context.Background(), "prompt")
	waitDelta(t, q, 1)
	h.sync(v, a.View())
	close(next)
	waitDelta(t, q, 1)
	h.sync(v, a.View())
	var spoken string
	if task := h.speech.tasks[h.speech.active[a.ID()]]; task != nil {
		spoken += task.text
	}
	if task := h.speech.pending[a.ID()]; task != nil {
		spoken += task.text
	}
	h.mu.Unlock()
	close(release)
	select {
	case <-result:
	case <-time.After(3 * time.Second):
		t.Fatal("turn did not end")
	}
	thought, answer := strings.Index(spoken, "unfinished thought"), strings.Index(spoken, "Answer.")
	if thought < 0 || answer < thought {
		t.Fatalf("later part overtook unchanged tail: %q", spoken)
	}
}

func TestCh13ToolAndProviderErrorsFromHistory(t *testing.T) {
	requests := 0
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		if requests == 1 {
			fmt.Fprint(w, `{"choices":[{"message":{"role":"assistant","tool_calls":[{"id":"bad","type":"function","function":{"name":"missing_ch13_tool","arguments":"{}"}}]},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`)
			return
		}
		w.WriteHeader(500)
		fmt.Fprint(w, `{"error":{"message":"provider failed too"}}`)
	}))
	defer peer.Close()
	app := ensemble.New(nil)
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "local", APIKey: "not-real", BaseURL: peer.URL, StreamCapability: ensemble.StreamPlain, DisableStreaming: true, BaseDir: t.TempDir(), Speaking: true, MaxToolRounds: 3})
	if err != nil {
		t.Fatal(err)
	}
	defer app.Shutdown()
	h := NewHub(app)
	defer h.Close()
	h.mu.Lock()
	v := h.view(a)
	h.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result := <-a.Submit(ctx, "try the missing tool")
	if result.Err == nil {
		t.Fatal("provider error missing")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sync(v, a.View())
	var errorsText []string
	for _, task := range h.speech.tasks {
		for _, part := range task.segments {
			if part.Source == "error" {
				errorsText = append(errorsText, part.Text)
			}
		}
	}
	if pending := h.speech.pending[a.ID()]; pending != nil {
		for _, part := range pending.segments {
			if part.Source == "error" {
				errorsText = append(errorsText, part.Text)
			}
		}
	}
	expected := 1 // failed-tool result, then the actual public provider-error events
	for _, event := range a.Log().Events() {
		if event.Type == ensemble.ErrorOccurred {
			expected++
		}
	}
	if len(errorsText) != expected || !strings.Contains(errorsText[0], "missing ch13 tool") {
		t.Fatalf("failed-tool/provider admission: %q", errorsText)
	}
	for _, text := range errorsText[1:] {
		if !strings.Contains(text, "HTTP status 500") {
			t.Fatalf("provider order/content: %q", errorsText)
		}
	}

}
