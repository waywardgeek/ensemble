package ensemble_test

import (
	"context"
	"encoding/json"
	"ensemble/ensemble"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type activityLog struct {
	mu     sync.Mutex
	phases []string
}

// Observe records actual actor phases independently of status tool text.
func (l *activityLog) Observe(o ensemble.Observation) {
	if a, ok := o.(ensemble.ActivityChanged); ok {
		l.mu.Lock()
		l.phases = append(l.phases, a.Activity.Phase)
		l.mu.Unlock()
	}
}

// Only the provider is replaced: real actors, tools, jobs and durable submission
// establish the phase, output cursor and disk fallback contracts together.
func TestManagementActivityOutputAndSubmission(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(fleetResultText(body), "tool text") {
			fleetCall(w, "run_command", map[string]string{"command": "printf 'tool text'"})
		} else {
			fleetCall(w, "submit", map[string]string{"result": "DONE_RESULT"})
		}
	}))
	t.Cleanup(server.Close)
	a, _, _ := fleetAgent(t, server)
	phases := &activityLog{}
	root := a.Ensemble().(*ensemble.Ensemble)
	root.SetAgentFactory(func(parent ensemble.Agent, spec ensemble.AgentSpec) (ensemble.Agent, error) {
		c, err := root.NewChildAgent(parent, parent.Config(), spec)
		if err == nil {
			c.Observe(phases)
		}
		return c, err
	})
	if _, err := a.Fleet().Spawn("worker", "TASK", ""); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := a.Fleet().Join(ctx, []string{"worker"})
	if err != nil || !strings.Contains(result, "DONE_RESULT") {
		t.Fatal(result, err)
	}
	child, _ := a.Fleet().Find("worker")
	if err = child.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	phases.mu.Lock()
	got := strings.Join(phases.phases, ",")
	phases.mu.Unlock()
	if !strings.Contains(got, "thinking,tool_call") || !strings.HasSuffix(got, "idle") {
		t.Fatal("missing ordered phases", got)
	}
	output, err := a.Fleet().Output(ctx, "worker", 0, false)
	if err != nil || !strings.Contains(output.Content, "tool text") || output.Status.Usage.Input == 0 {
		t.Fatal(output, err)
	}
	next, err := a.Fleet().Output(ctx, "worker", output.Cursor, false)
	if err != nil || next.Content != "" {
		t.Fatal("cursor repeats output", next, err)
	}
	child.Fleet().Begin()
	submission, err := a.Fleet().Submission("worker")
	if err != nil || submission != "DONE_RESULT" {
		t.Fatal("durable fallback", submission, err)
	}
}

// Guessing another tree's ID must reveal no more than a genuinely absent name.
func TestManagementOwnership(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fleetCall(w, "submit", map[string]string{"result": "owned"})
	}))
	t.Cleanup(server.Close)
	a, _, _ := fleetAgent(t, server)
	if _, err := a.Fleet().Spawn("owned", "OWN", ""); err != nil {
		t.Fatal(err)
	}
	b, _, _ := fleetAgent(t, server)
	if _, err := b.Fleet().Spawn("foreign", "TASK", ""); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	for _, name := range []string{"shutdown_agent", "kill_agent", "interrupt_agent", "send_message", "ask_agent", "wait_for_agent", "get_submission", "get_sub_agent_status", "respond_to_parent_message", "join_agents", "wait_for_agent_change"} {
		var messages []string
		for _, id := range []string{"foreign", "absent"} {
			args, _ := json.Marshal(map[string]any{"agent_id": id, "agent_ids": []string{id}, "content": "hello", "response": "reply", "timeout_seconds": 0.1})
			result, err := a.Invoke(ctx, name, args)
			message := result
			if err != nil {
				message = err.Error()
			}
			if !strings.Contains(message, "unknown agent") {
				t.Fatalf("%s wrong rejection: %s", name, message)
			}
			messages = append(messages, strings.ReplaceAll(message, id, "TARGET"))
		}
		if messages[0] != messages[1] {
			t.Fatalf("%s distinguishes foreign and absent: %q", name, messages)
		}
	}
}

// A parked real HTTP request lets us distinguish waiting for completion from
// canceling it. The stuck detector is a timeout property with that working baseline.
func TestManagementWaitAndShutdownLifetimes(t *testing.T) {
	entered := make(chan struct{}, 2)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		entered <- struct{}{}
		select {
		case <-release:
			fleetCall(w, "submit", map[string]string{"result": "finished"})
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(server.Close)
	a, _, _ := fleetAgent(t, server)
	if _, err := a.Fleet().Spawn("worker", "TASK", ""); err != nil {
		t.Fatal(err)
	}
	<-entered
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	stuck, err := a.Fleet().WaitChange(ctx, []string{"worker"}, "", 100*time.Millisecond)
	if err != nil || !strings.Contains(stuck, `"reason":"stuck"`) {
		t.Fatal("stuck wait", stuck, err)
	}
	graceful := make(chan error, 1)
	go func() { graceful <- a.Fleet().GracefulShutdown(ctx, "worker") }()
	select {
	case err := <-graceful:
		t.Fatal("graceful canceled active turn", err)
	case <-time.After(30 * time.Millisecond):
	}
	close(release)
	if err := <-graceful; err != nil {
		t.Fatal(err)
	}
	if _, err := a.Fleet().Find("worker"); err == nil {
		t.Fatal("graceful retained child")
	}
}

// Reviewer found manager-wide message counting attributed an unselected child's
// question to the watched child. Keep the two real HTTP requests parked so the
// only change is the other child's question; it must not satisfy this wait.
func TestManagementWaitIgnoresOtherChildMessage(t *testing.T) {
	entered := make(chan struct{}, 3)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		entered <- struct{}{}
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)
	a, _, _ := fleetAgent(t, server)
	for _, id := range []string{"watched", "other"} {
		if _, err := a.Fleet().Spawn(id, "PARK", ""); err != nil {
			t.Fatal(err)
		}
		<-entered
	}
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		result, err := a.Fleet().WaitChange(ctx, []string{"watched"}, "", 0)
		if err == nil {
			done <- fmt.Errorf("unselected child woke watch: %s", result)
		} else {
			done <- err
		}
	}()
	time.Sleep(20 * time.Millisecond)
	other, _ := a.Fleet().Find("other")
	go other.Fleet().AskParent(ctx, "OTHER_QUESTION")
	if err := <-done; err != context.DeadlineExceeded {
		t.Fatal(err)
	}
}

type dispatchedTool chan struct{}

// Observe exposes the real serial-dispatch boundary before the shell completes.
func (c dispatchedTool) Observe(o ensemble.Observation) {
	if d, ok := o.(ensemble.ToolDispatched); ok && d.Name == "run_command" {
		c <- struct{}{}
	}
}

// Both interrupt modes retain the child for another task. after_tool keeps the
// actual tool result; now cancels the in-flight HTTP request instead of waiting.
func TestManagementInterruptModesAndAsk(t *testing.T) {
	for _, mode := range []string{"now", "after_tool"} {
		t.Run(mode, func(t *testing.T) {
			entered := make(chan struct{}, 1)
			var calls atomic.Int32
			var thirdSeen atomic.Bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				if strings.Contains(string(body), "THIRD") {
					thirdSeen.Store(true)
				}
				if calls.Add(1) == 1 {
					if mode == "now" {
						entered <- struct{}{}
						<-r.Context().Done()
						return
					}
					fleetCall(w, "run_command", map[string]any{"command": "sleep 0.15; printf TOOL_FINISHED", "ai_callback_delay": 2})
					return
				}
				fleetCall(w, "submit", map[string]string{"result": "SECOND_RESULT"})
			}))
			t.Cleanup(server.Close)
			a, _, _ := fleetAgent(t, server)
			root := a.Ensemble().(*ensemble.Ensemble)
			root.SetAgentFactory(func(parent ensemble.Agent, spec ensemble.AgentSpec) (ensemble.Agent, error) {
				child, err := root.NewChildAgent(parent, parent.Config(), spec)
				if err == nil && mode == "after_tool" {
					child.Observe(dispatchedTool(entered))
				}
				return child, err
			})
			if _, err := a.Fleet().Spawn("worker", "FIRST", ""); err != nil {
				t.Fatal(err)
			}
			<-entered
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			args, _ := json.Marshal(map[string]string{"agent_id": "worker", "mode": mode})
			if _, err := a.Invoke(ctx, "interrupt_agent", args); err != nil {
				t.Fatal(err)
			}
			child, _ := a.Fleet().Find("worker")
			if err := child.Flush(ctx); err != nil {
				t.Fatal(err)
			}
			interrupted, toolResult := false, false
			for _, event := range child.History().EventsAfter(0) {
				interrupted = interrupted || event.Type == "turn_interrupted"
				if event.Type == "tool_returned" {
					for _, part := range event.Tool.Parts {
						toolResult = toolResult || strings.Contains(part.Text, "TOOL_FINISHED")
					}
				}
			}
			if !interrupted || mode == "after_tool" && !toolResult {
				t.Fatalf("interrupt=%v completed tool=%v", interrupted, toolResult)
			}
			answer, err := a.Invoke(ctx, "ask_agent", json.RawMessage(`{"agent_id":"worker","content":"SECOND"}`))
			if err != nil || !strings.Contains(answer, "SECOND_RESULT") {
				t.Fatal("interrupt killed child", answer, err)
			}
			if _, err = a.Invoke(ctx, "send_message", json.RawMessage(`{"agent_id":"worker","content":"THIRD"}`)); err != nil {
				t.Fatal(err)
			}
			if err = child.Flush(ctx); err != nil || !thirdSeen.Load() {
				t.Fatal("idle send_message was not delivered", err)
			}

		})
	}
}

// Monitor three asynchronous children through change waits, then collect their
// actual submissions through the inherited join path. Each provider is released
// separately, making the observed child identity meaningful.
func TestManagementThreeChildMonitor(t *testing.T) {
	gates := map[string]chan struct{}{"one": make(chan struct{}), "two": make(chan struct{}), "three": make(chan struct{})}
	entered := make(chan string, 3)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		for id, gate := range gates {
			if strings.Contains(string(body), "TASK_"+id) {
				entered <- id
				select {
				case <-gate:
					fleetCall(w, "submit", map[string]string{"result": "RESULT_" + id})
				case <-r.Context().Done():
				}
				return
			}
		}
	}))
	t.Cleanup(server.Close)
	a, _, _ := fleetAgent(t, server)
	for _, id := range []string{"one", "two", "three"} {
		if _, err := a.Fleet().Spawn(id, "TASK_"+id, ""); err != nil {
			t.Fatal(err)
		}
		<-entered
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	for _, id := range []string{"one", "two", "three"} {
		done := make(chan string, 1)
		go func() {
			result, err := a.Fleet().WaitChange(ctx, []string{id}, "", 0)
			if err != nil {
				result = err.Error()
			}
			done <- result
		}()
		close(gates[id])
		result := <-done
		if !strings.Contains(result, `"agent_id":"`+id+`"`) {
			t.Fatal("change did not identify released child", result)
		}
	}
	joined, err := a.Fleet().Join(ctx, []string{"one", "two", "three"})
	if err != nil {
		t.Fatal(err)
	}
	for id := range gates {
		if !strings.Contains(joined, "RESULT_"+id) {
			t.Fatal("join lost result", joined)
		}
	}
	pattern, err := a.Fleet().WaitChange(ctx, []string{"one"}, "RESULT_one", 0)
	if err != nil || !strings.Contains(pattern, `"reason":"pattern"`) {
		t.Fatal("pattern match missing", pattern, err)
	}
}
