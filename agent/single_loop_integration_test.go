package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/waywardgeek/ensemble/agent/internal/common"
	"github.com/waywardgeek/ensemble/agent/internal/llm"
	"github.com/waywardgeek/ensemble/agent/internal/settings"
)

func loopTestAgent(t *testing.T, reply func(map[string]any) []map[string]any) *Agent {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"id": "test", "type": "message", "role": "assistant", "model": "claude-sonnet-5", "content": reply(req), "stop_reason": "end_turn", "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
	}))
	t.Cleanup(srv.Close)
	a := NewBareAgent(Config{Model: "claude-sonnet-5", Vendor: common.VendorAnthropic, BaseURL: srv.URL, APIKey: "test", DisableStreaming: true}, filepath.Join(t.TempDir(), "events.json"))
	t.Cleanup(func() { a.Shutdown() })
	return a
}
func textReply(text string) []map[string]any { return []map[string]any{{"type": "text", "text": text}} }

func TestBlockingAskSurvivesObservationOverflow(t *testing.T) {
	a := loopTestAgent(t, func(map[string]any) []map[string]any {
		var parts []map[string]any
		for i := 0; i < 300; i++ {
			parts = append(parts, map[string]any{"type": "text", "text": "x"})
		}
		return parts
	})
	actor := a.NewActor()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go actor.Run(ctx)
	done := make(chan error, 1)
	go func() {
		text, err := actor.Ask("hello")
		if err == nil && text != strings.Repeat("x", 300) {
			err = fmt.Errorf("wrong reply length: %d", len(text))
		}
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("blocking completion lost behind full observation buffer")
	}
}

func TestAgentHasOneActor(t *testing.T) {
	a := loopTestAgent(t, func(map[string]any) []map[string]any { return textReply("done") })
	if a.NewActor() != a.NewActor() {
		t.Fatal("Agent creates multiple execution owners")
	}
	if text, err := a.Ask("hello"); err != nil || text != "done" {
		t.Fatalf("Ask: %q, %v", text, err)
	}
}

func TestConcurrentBlockingRequestsGetOwnReply(t *testing.T) {
	a := loopTestAgent(t, func(req map[string]any) []map[string]any {
		messages := req["messages"].([]any)
		last := messages[len(messages)-1].(map[string]any)
		b, _ := json.Marshal(last["content"])
		if strings.Contains(string(b), "question one") {
			return textReply("answer one")
		}
		return textReply("answer two")
	})
	var wg sync.WaitGroup
	for _, name := range []string{"one", "two"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			text, err := a.Ask("question " + name)
			if err != nil || text != "answer "+name {
				t.Errorf("%s: %q, %v", name, text, err)
			}
		}()
	}
	wg.Wait()
}

func TestToolRoundLimitAndResume(t *testing.T) {
	for _, tc := range []struct {
		name           string
		limit, batches int
		stop           bool
	}{
		{"configured_above_16", 200, 18, false}, {"zero_default", 0, 18, false},
		{"exact_boundary", 2, 2, false}, {"reject_next_batch", 2, 3, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := 0
			a := loopTestAgent(t, func(map[string]any) []map[string]any {
				request++
				if request > tc.batches {
					return textReply("finished")
				}
				return []map[string]any{{"type": "text", "text": "I will run the next tool."}, {"type": "tool_use", "id": fmt.Sprintf("call%d", request), "name": "count", "input": map[string]any{}}}
			})
			a.eng.Cfg.MaxToolRounds = tc.limit
			runs := 0
			a.reg.RegisterInitial(common.Tool{Name: "count", Description: "Count one execution.", Schema: json.RawMessage(`{"type":"object","properties":{}}`), Run: func(*common.Call, json.RawMessage) (string, error) { runs++; return "counted", nil }})
			a.eng.Cfg.Tools = a.reg.Declarations()
			text, err := a.Ask("start")
			if tc.stop {
				var limitErr *ToolRoundLimitError
				if !errors.As(err, &limitErr) || limitErr.Limit != tc.limit {
					t.Fatalf("expected limit error, got %q %v", text, err)
				}
				if runs != tc.limit {
					t.Fatalf("executed %d batches, limit %d", runs, tc.limit)
				}
				if calls := a.eng.PendingCalls(); len(calls) != 0 {
					t.Fatalf("dangling calls: %+v", calls)
				}
				if text, err := a.Ask("continue"); err != nil || text != "finished" {
					t.Fatalf("resume: %q %v", text, err)
				}
			} else if err != nil || text != "finished" || runs != tc.batches {
				t.Fatalf("reply %q err %v runs %d", text, err, runs)
			}
			if _, err := llm.LoadLogFile(a.eng.Path); err != nil {
				t.Fatalf("saved log: %v", err)
			}
		})
	}
}

func TestBlockingAskReturnsSaveFailure(t *testing.T) {
	a := loopTestAgent(t, func(map[string]any) []map[string]any { return textReply("done") })
	a.eng.Path = filepath.Join(t.TempDir(), "missing", "events.json")
	if _, err := a.Ask("hello"); err == nil {
		t.Fatal("save failure reported as success")
	}
}

func TestBlockingAskCancellationAndStoppedActor(t *testing.T) {
	for _, streaming := range []bool{false, true} {
		t.Run(fmt.Sprintf("streaming_%t", streaming), func(t *testing.T) {
			started, disconnected, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Consume the POST, as a real provider does. net/http cannot start
				// its background disconnect read while an unread body owns the
				// connection; waiting only on Context.Done would deadlock the fake.
				if _, err := io.Copy(io.Discard, r.Body); err != nil {
					t.Error(err)
					return
				}
				if streaming {
					w.Header().Set("Content-Type", "text/event-stream")
					fmt.Fprint(w, "data: {\"type\":\"message_start\",\"message\":{\"id\":\"cancel\",\"type\":\"message\",\"role\":\"assistant\",\"model\":\"claude-sonnet-5\",\"content\":[],\"usage\":{\"input_tokens\":1,\"output_tokens\":0}}}\n\n")
					w.(http.Flusher).Flush()
				}
				close(started)
				select {
				case <-r.Context().Done():
					close(disconnected)
				case <-release: // failure cleanup only, never evidence of cancellation
				}
			}))
			defer srv.Close()
			defer close(release)
			a := NewBareAgent(Config{Vendor: common.VendorAnthropic, Model: "claude-sonnet-5", BaseURL: srv.URL, APIKey: "test"}, filepath.Join(t.TempDir(), "events.json"))
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() { _, err := a.NewActor().AskContext(ctx, "hello"); done <- err }()
			select {
			case <-started:
			case <-time.After(2 * time.Second):
				t.Fatal("provider never received request")
			}
			cancel()
			select {
			case err := <-done:
				if !errors.Is(err, context.Canceled) {
					t.Fatal(err)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("Ask ignored cancellation")
			}
			select {
			case <-disconnected:
			case <-time.After(2 * time.Second):
				t.Fatal("request did not propagate cancellation to provider")
			}
			stopped := make(chan error, 1)
			go func() { stopped <- a.Shutdown() }()
			select {
			case err := <-stopped:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("actor shutdown did not join execution")
			}
			if err := a.Shutdown(); err != nil {
				t.Fatal(err)
			}
			if _, err := a.Ask("too late"); !errors.Is(err, ErrActorStopped) {
				t.Fatalf("Ask after shutdown: %v", err)
			}
		})
	}
}

func TestShutdownJoinsToolBeforeSaving(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	finish := func() { once.Do(func() { close(release) }) }
	defer finish()
	a := loopTestAgent(t, func(map[string]any) []map[string]any {
		return []map[string]any{{"type": "text", "text": "I will wait for the tool."}, {"type": "tool_use", "id": "blocked", "name": "block", "input": map[string]any{}}}
	})
	a.reg.RegisterInitial(common.Tool{Name: "block", Description: "Wait.", Schema: json.RawMessage(`{"type":"object","properties":{}}`), Run: func(*common.Call, json.RawMessage) (string, error) {
		close(started)
		<-release
		return "actual output", nil
	}})
	a.eng.Cfg.Tools = a.reg.Declarations()
	ask := make(chan error, 1)
	go func() { _, err := a.Ask("start"); ask <- err }()
	<-started
	stopped := make(chan error, 1)
	go func() { stopped <- a.Shutdown() }()
	select {
	case err := <-stopped:
		t.Fatalf("shutdown returned before tool exited: %v", err)
	case <-time.After(30 * time.Millisecond):
	}
	finish()
	select {
	case err := <-stopped:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown did not join")
	}
	select {
	case err := <-ask:
		if err == nil {
			t.Fatal("interrupted Ask succeeded")
		}
	case <-time.After(time.Second):
		t.Fatal("Ask stranded")
	}
	log, err := llm.LoadLogFile(a.eng.Path)
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, e := range log.Events {
		if e.Type == common.ToolReturned && e.Tool.CallID == "blocked" {
			found++
			if e.Tool.Parts[0] != (common.TextPart{Text: "actual output"}) {
				t.Fatalf("lost real result: %+v", e.Tool)
			}
		}
	}
	if found != 1 {
		t.Fatalf("saved %d results, want one", found)
	}
}

// The callback deadline is a return of control, not cancellation. A Go tool
// that remains blocked must not keep Ask from receiving a running-job report.
func TestBlockingAskReceivesJobCallbackBeforeToolReturns(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	a := loopTestAgent(t, func(req map[string]any) []map[string]any {
		messages := req["messages"].([]any)
		last, _ := json.Marshal(messages[len(messages)-1])
		if strings.Contains(string(last), "tool_result") {
			if !strings.Contains(string(last), "still running") {
				t.Errorf("not a running job report: %s", last)
			}
			return textReply("job is still running")
		}
		return []map[string]any{{"type": "text", "text": "I will start a tool with a short callback."}, {"type": "tool_use", "id": "callback", "name": "slow", "input": map[string]any{"ai_callback_delay": 0.02}}}
	})
	defer close(release) // before the agent cleanup joins its Go worker
	a.reg.RegisterInitial(common.Tool{Name: "slow", Description: "Wait.", Schema: json.RawMessage(`{"type":"object","properties":{}}`), Run: func(*common.Call, json.RawMessage) (string, error) {
		close(started)
		<-release
		return "eventual output", nil
	}})
	a.eng.Cfg.Tools = a.reg.Declarations()
	done := make(chan error, 1)
	go func() {
		text, err := a.Ask("start")
		if err == nil && text != "job is still running" {
			err = fmt.Errorf("reply = %q", text)
		}
		done <- err
	}()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("tool did not start")
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("callback delay waited for the tool function to return")
	}
}

func TestRuntimeRoundLimitChangesAtHumanTurnBoundary(t *testing.T) {
	store := settings.NewSettingsStore(filepath.Join(t.TempDir(), "settings.json"))
	store.ApplyRaw(json.RawMessage(`{"max_tool_rounds":1}`))
	var requests atomic.Int32
	a := loopTestAgent(t, func(map[string]any) []map[string]any {
		n := requests.Add(1)
		if n == 1 {
			// The GUI writes the store while a request is in flight. This must not
			// enlarge the active turn's budget or require restarting the actor.
			store.ApplyRaw(json.RawMessage(`{"max_tool_rounds":3}`))
		}
		if n <= 5 {
			return []map[string]any{
				{"type": "text", "text": "I will run the next step."},
				{"type": "tool_use", "id": fmt.Sprintf("runtime%d", n), "name": "step", "input": map[string]any{}},
			}
		}
		return textReply("done")
	})
	a.eng.ToolRoundLimit = func() int { return store.Get().MaxToolRounds }
	var dispatched atomic.Int32
	a.RegisterTool("step", "step", json.RawMessage(`{"type":"object"}`), func(json.RawMessage) (string, error) {
		dispatched.Add(1)
		return "ok", nil
	})
	_, err := a.Ask("first turn")
	var stopped *ToolRoundLimitError
	if !errors.As(err, &stopped) || stopped.Limit != 1 || requests.Load() != 2 || dispatched.Load() != 1 {
		t.Fatalf("active budget changed: err=%v requests=%d dispatched=%d", err, requests.Load(), dispatched.Load())
	}
	text, err := a.Ask("next turn")
	if err != nil || text != "done" || requests.Load() != 6 || dispatched.Load() != 4 {
		t.Fatalf("next turn missed updated budget: text=%q err=%v requests=%d dispatched=%d", text, err, requests.Load(), dispatched.Load())
	}
	if pending := a.eng.PendingCalls(); len(pending) != 0 {
		t.Fatalf("unpaired calls: %+v", pending)
	}
}
