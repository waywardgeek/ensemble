package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"ensemble/ensemble"
	"ensemble/internal/common"
)

func replyText(w http.ResponseWriter, text string) {
	json.NewEncoder(w).Encode(map[string]any{"content": []map[string]string{{"type": "text", "text": text}}, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
}
func lastHuman(r *http.Request) string {
	var in struct {
		// Messages retains provider message order, including tool-result boundaries.
		Messages []struct {
			// Content preserves the provider's ordered blocks or message text.
			Content []struct {
				// Type selects the wire block variant before its other fields are interpreted. Text
				// holds visible provider or user content, including an explicitly empty string.
				Type, Text string
			}
		}
	}
	json.NewDecoder(r.Body).Decode(&in)
	value := ""
	for _, m := range in.Messages {
		for _, p := range m.Content {
			if p.Type == "text" {
				value = p.Text
			}
		}
	}
	return value
}
func actorAgent(t *testing.T, server *httptest.Server, change func(*ensemble.Config)) ensemble.Agent {
	t.Helper()
	cfg := ensemble.Config{Model: "fake", BaseURL: server.URL, DataDir: t.TempDir()}
	if change != nil {
		change(&cfg)
	}
	a, err := ensemble.New(io.Discard).NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := a.Shutdown(); err != nil {
			t.Error(err)
		}
	})
	return a
}
func receive(t *testing.T, ch <-chan ensemble.Result) ensemble.Result {
	t.Helper()
	select {
	case result := <-ch:
		return result
	case <-time.After(3 * time.Second):
		t.Fatal("request did not complete")
		return ensemble.Result{}
	}
}

// The provider and tool are external peers. Channels establish actual overlap:
// the new turn must finish while the earlier Go handler is still blocked.
func TestInterruptKeepsAgentAndJobAlive(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	defer unblock()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		text := lastHuman(r)
		if text == "start" {
			io.WriteString(w, `{"content":[{"type":"tool_use","id":"slow-call","name":"slow","input":{"ai_callback_delay":0.1}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		} else {
			replyText(w, "new turn answered")
		}
	}))
	defer func() { unblock(); server.Close() }()
	a := actorAgent(t, server, nil)
	// Cleanup releases the handler BEFORE the Agent cleanup joins it on failures.
	t.Cleanup(unblock)
	a.RegisterTool(ensemble.ToolDeclaration{Name: "slow", Schema: json.RawMessage(`{"type":"object"}`)}, func(_ ensemble.ToolContext, _ json.RawMessage) (string, error) {
		close(entered)
		<-release
		return "real late result", nil
	})
	first := a.Submit(context.Background(), "start")
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("handler never started")
	}
	a.Post(ensemble.Hint{Text: "heard while busy"})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := a.Wait(ctx, func(o ensemble.Observation) bool {
		d, ok := o.(ensemble.PartDelta)
		return ok && d.Chunk == "hint:heard while busy"
	}); err != nil {
		t.Fatal("hint was deaf:", err)
	}
	a.Post(ensemble.Interrupt{})
	if result := receive(t, first); !errors.Is(result.Err, ensemble.ErrInterrupted) {
		t.Fatalf("interrupt reply: %+v", result)
	}
	if result := receive(t, a.Submit(context.Background(), "new turn")); result.Err != nil || result.Text != "new turn answered" {
		t.Fatalf("restart: %+v", result)
	}
	job, err := a.Jobs().Find(1)
	if err != nil {
		t.Fatal(err)
	}
	if job.Data().Status != "running" {
		t.Fatal("interrupt killed continuing job")
	}
	stopped := make(chan error, 1)
	go func() { stopped <- a.Shutdown() }()
	select {
	case <-stopped:
		t.Fatal("shutdown failed to join blocked Go handler")
	case <-time.After(30 * time.Millisecond):
	}
	unblock()
	select {
	case err := <-stopped:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown did not finish after handler")
	}
	data := job.Data()
	raw, err := os.ReadFile(data.Output.Locator)
	if err != nil || data.Status != "done" || string(raw) != "real late result" {
		t.Fatalf("late result lost: %s %q %v", data.Status, raw, err)
	}
	if result := receive(t, a.Submit(context.Background(), "after shutdown")); !errors.Is(result.Err, ensemble.ErrStopped) {
		t.Fatal("stopped agent accepted request")
	}
	if err := a.Shutdown(); err != nil {
		t.Fatal("repeated shutdown:", err)
	}
}

// A callback must reach the model before a blocked Go handler returns.
func TestCallbackReportsWhileHandlerBlocked(t *testing.T) {
	release := make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	var second string
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		raw, _ := io.ReadAll(r.Body)
		if calls == 1 {
			io.WriteString(w, `{"content":[{"type":"tool_use","id":"slow","name":"slow","input":{"ai_callback_delay":0.01}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		} else {
			second = string(raw)
			replyText(w, "callback received")
		}
	}))
	defer func() { unblock(); server.Close() }()
	a := actorAgent(t, server, nil)
	t.Cleanup(unblock)
	a.RegisterTool(ensemble.ToolDeclaration{Name: "slow", Schema: json.RawMessage(`{"type":"object"}`)}, func(_ ensemble.ToolContext, _ json.RawMessage) (string, error) { <-release; return "done later", nil })
	result := receive(t, a.Submit(context.Background(), "start"))
	if result.Err != nil || result.Text != "callback received" || !strings.Contains(second, "running") || !strings.Contains(second, "wait_for_job") {
		t.Fatalf("callback did not reach model: %+v %s", result, second)
	}
	job, _ := a.Jobs().Find(1)
	if job.Data().Status != "running" {
		t.Fatal("callback canceled execution")
	}
	unblock()
}

// Each queued caller must receive its own reply even when nobody consumes bounded
// progress.
func TestRequestsCompletePrivatelyWhenProgressOverflows(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { replyText(w, lastHuman(r)) }))
	defer server.Close()
	a := actorAgent(t, server, nil)
	replies := make([]<-chan ensemble.Result, 90)
	for i := range replies {
		replies[i] = a.Submit(context.Background(), strings.Repeat("x", i+1))
	}
	// No observation consumer: more transitions/finals than the progress buffer.
	for i, ch := range replies {
		result := receive(t, ch)
		if result.Err != nil || result.Text != strings.Repeat("x", i+1) {
			t.Fatalf("request %d received another completion: %+v", i, result)
		}
	}
}

// A mid-request hint must appear exactly once in the next request and reproduce from
// log replay.
func TestHintDuringHTTPReplaysAtNextRequest(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	var requests [][]byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		requests = append(requests, raw)
		if len(requests) == 1 {
			close(entered)
			<-release
		}
		replyText(w, "answer")
	}))
	defer func() { unblock(); server.Close() }()
	a := actorAgent(t, server, nil)
	t.Cleanup(unblock)
	pending := a.Submit(context.Background(), "first")
	<-entered
	a.Post(ensemble.UserMessage{Text: "next-request-hint"})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := a.Wait(ctx, func(o ensemble.Observation) bool {
		d, ok := o.(ensemble.PartDelta)
		return ok && d.Chunk == "hint:next-request-hint"
	}); err != nil {
		t.Fatal(err)
	}
	unblock()
	if result := receive(t, pending); result.Err != nil {
		t.Fatal(result.Err)
	}
	if len(requests) != 2 || bytes.Contains(requests[0], []byte("next-request-hint")) || !bytes.Contains(requests[1], []byte("next-request-hint")) {
		t.Fatal("hint delivery crossed request boundary")
	}
	// Replay each prefix immediately before request_sent through the real engine.
	var log bytes.Buffer
	a.History().Dump(&log)
	lines := bytes.Split(bytes.TrimSpace(log.Bytes()), []byte("\n"))
	index := 0
	for i, line := range lines {
		if bytes.Contains(line, []byte(`"type":"request_sent"`)) {
			replay := actorAgent(t, server, nil)
			if err := replay.History().Load(bytes.NewReader(bytes.Join(lines[:i], []byte("\n")))); err != nil {
				t.Fatal(err)
			}
			body, err := replay.Engine().Render(replay.History().Context())
			if err != nil || !bytes.Equal(body, requests[index]) {
				t.Fatalf("request %d replay differs: %v", index, err)
			}
			index++
		}
	}
	if _, err := a.Ask(context.Background(), "third"); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(requests[2], []byte("next-request-hint")) {
		t.Fatal("hint became permanent conversation")
	}
}

// The limit must reject every excess call, save its result and leave the next turn
// usable.
func TestRoundLimitPairsEveryRejectedCallAndAllowsNextTurn(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls > 2 {
			replyText(w, "recovered")
			return
		}
		ids := []string{"a"}
		if calls == 2 {
			ids = []string{"b", "c"}
		}
		var content []map[string]any
		for _, id := range ids {
			content = append(content, map[string]any{"type": "tool_use", "id": id, "name": "count", "input": map[string]any{}})
		}
		json.NewEncoder(w).Encode(map[string]any{"content": content, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
	}))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "events.jsonl")
	a := actorAgent(t, server, func(c *ensemble.Config) { c.MaxToolRounds = 1; c.LogPath = path })
	executions := 0
	a.RegisterTool(ensemble.ToolDeclaration{Name: "count", Schema: json.RawMessage(`{"type":"object"}`)}, func(_ ensemble.ToolContext, _ json.RawMessage) (string, error) { executions++; return "executed", nil })
	_, err := a.Ask(context.Background(), "limit")
	if !errors.Is(err, ensemble.ErrRoundLimit) || calls != 2 || executions != 1 {
		t.Fatalf("limit: %d requests %d executions %v", calls, executions, err)
	}
	raw, err := os.ReadFile(path)
	if err != nil || bytes.Count(raw, []byte("not executed: tool round limit reached")) != 2 {
		t.Fatalf("rejected batch not saved before completion: %s %v", raw, err)
	}
	if text, err := a.Ask(context.Background(), "again"); err != nil || text != "recovered" {
		t.Fatalf("next turn: %q %v", text, err)
	}
}

// A completed model reply must not acknowledge success when its required save
// failed.
func TestSaveFailureIsRequestFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { replyText(w, "unsaved") }))
	defer server.Close()
	cfg := ensemble.Config{Model: "fake", BaseURL: server.URL, LogPath: filepath.Join(t.TempDir(), "missing", "events.jsonl"), DataDir: t.TempDir()}
	a, err := ensemble.New(io.Discard).NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown() // The same expected save failure is returned at shutdown.
	if _, err := a.Ask(context.Background(), "save"); err == nil {
		t.Fatal("unsaved response acknowledged as successful")
	}
}

// Interrupting an active turn must leave a prior managed process interactive and
// recoverable.
func TestInterruptDoesNotKillManagedProcess(t *testing.T) {
	entered := make(chan struct{})
	first := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		text := lastHuman(r)
		if first {
			first = false
			io.WriteString(w, `{"content":[{"type":"tool_use","id":"process","name":"run_command","input":{"command":"printf 'READY\\n'; read line; printf 'REAL=%s\\n' \"$line\"","ai_callback_delay":0.02}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		} else if text == "busy" {
			close(entered)
			<-r.Context().Done()
			return
		} else {
			replyText(w, "ready for new work")
		}
	}))
	defer server.Close()
	a := actorAgent(t, server, func(c *ensemble.Config) { c.BuiltinTools = true })
	defer a.Shutdown()
	if result := receive(t, a.Submit(context.Background(), "start process")); result.Err != nil {
		t.Fatal(result.Err)
	}
	job, err := a.Jobs().Find(1)
	if err != nil {
		t.Fatal(err)
	}
	// Start another HTTP turn and interrupt it while the prior process continues.
	// The callback has already returned; ownership is job-wide, not turn-wide.
	busy := a.Submit(context.Background(), "busy")
	<-entered
	a.Post(ensemble.Interrupt{})
	if result := receive(t, busy); !errors.Is(result.Err, ensemble.ErrInterrupted) {
		t.Fatal("busy turn did not interrupt")
	}
	if result := receive(t, a.Submit(context.Background(), "new work")); result.Err != nil {
		t.Fatal(result.Err)
	}
	if job.Data().Status != "running" {
		t.Fatal("interrupt stopped another managed job")
	}
	if err := job.Send("kept"); err != nil {
		t.Fatal(err)
	}
	text, err := job.Wait(common.Limits{Delay: time.Second, MaxBytes: 1024})
	if err != nil || !strings.Contains(text, "REAL=kept") || job.Data().Status != "done" {
		t.Fatalf("continuing process result: %s %v", text, err)
	}
}

// Merged progress must retain the actual Agent, stored part identity, tool content
// and state changes.
func TestObserverRoutesIdentityAndFinalParts(t *testing.T) {
	first := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if first {
			first = false
			io.WriteString(w, `{"content":[{"type":"tool_use","id":"echo","name":"echo","input":{}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
			return
		}
		replyText(w, "final")
	}))
	defer server.Close()
	app := ensemble.New(io.Discard)
	a, _ := app.NewAgent(ensemble.Config{Model: "first", BaseURL: server.URL, DataDir: t.TempDir()})
	defer a.Shutdown()
	b, _ := app.NewAgent(ensemble.Config{Model: "second", BaseURL: server.URL, DataDir: t.TempDir()})
	defer b.Shutdown()
	a.RegisterTool(ensemble.ToolDeclaration{Name: "echo", Schema: json.RawMessage(`{"type":"object"}`)}, func(_ ensemble.ToolContext, _ json.RawMessage) (string, error) { return "tool-visible", nil })
	if _, err := a.Ask(context.Background(), "call"); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Ask(context.Background(), "text"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	seen := map[ensemble.Agent]map[string]bool{a: {}, b: {}}
	transitions := map[string]bool{}
	finals := 0
	for finals < 4 {
		routed, err := app.WaitAny(ctx, func(ensemble.AgentObservation) bool { return true })
		if err != nil {
			t.Fatal("missing routed observations:", err)
		}
		if _, exists := seen[routed.Agent]; !exists {
			t.Fatal("observation lost actual Agent identity")
		}
		switch event := routed.Observation.(type) {
		case ensemble.PartFinal:
			if event.Seq == 0 || event.PartID == 0 {
				t.Fatal("final part has no stored identity")
			}
			seen[routed.Agent][event.Part.Type] = true
			finals++
			if event.Part.Type == "tool_result" && event.Part.Parts[0].Text != "tool-visible" {
				t.Fatal("observer lost actual tool output")
			}
		case ensemble.StateChanged:
			transitions[event.To] = true
		}
	}
	if !seen[a]["tool_call"] || !seen[a]["tool_result"] || !seen[a]["text"] || !seen[b]["text"] {
		t.Fatal("observer lost parts or mixed Agents", seen)
	}
	for _, state := range []string{"input_pending", "in_flight", "tools_pending", "idle"} {
		if !transitions[state] {
			t.Fatal("missing state transition", state)
		}
	}
}

// Queued blocking callers must be released on shutdown, and caller cancellation
// must not strand the actor before that shutdown can complete.
func TestShutdownResolvesQueuedRequests(t *testing.T) {
	entered := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		close(entered)
		<-r.Context().Done()
	}))
	defer server.Close()
	a := actorAgent(t, server, nil)
	defer a.Shutdown()
	first := a.Submit(context.Background(), "active")
	<-entered
	var queued []<-chan ensemble.Result
	for i := 0; i < 5; i++ {
		queued = append(queued, a.Submit(context.Background(), "queued"))
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	cancelReply := make(chan error, 1)
	go func() { _, err := a.Ask(canceled, "canceled caller"); cancelReply <- err }()
	select {
	case err := <-cancelReply:
		if !errors.Is(err, context.Canceled) {
			t.Fatal("caller cancellation did not return")
		}
	case <-time.After(time.Second):
		t.Fatal("caller cancellation did not return")
	}

	if err := a.Shutdown(); err != nil {
		t.Fatal(err)
	}
	for _, ch := range append(queued, first) {
		if result := receive(t, ch); !errors.Is(result.Err, ensemble.ErrStopped) {
			t.Fatal("shutdown lost queued completion", result.Err)
		}
	}
}
