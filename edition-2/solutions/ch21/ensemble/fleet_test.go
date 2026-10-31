package ensemble_test

import (
	"context"
	"encoding/json"
	"fmt"
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

	"ensemble/ensemble"
)

type humanInbox chan ensemble.FleetMessage

// Deliver models the external human's inbox without replacing fleet routing.
func (h humanInbox) Deliver(m ensemble.FleetMessage) error { h <- m; return nil }

func fleetCall(w http.ResponseWriter, name string, args any) {
	json.NewEncoder(w).Encode(map[string]any{"content": []any{map[string]any{"type": "tool_use", "id": "call", "name": name, "input": args}}, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
}

// Extract only provider-visible tool results: markers in earlier tool arguments
// do not prove that execution returned those bytes.
func fleetResultText(body []byte) string {
	var request struct {
		// Messages preserves the provider's dialogue order.
		Messages []struct {
			// Content contains text and tool-result blocks.
			Content []struct {
				// Type distinguishes a result from a requested call.
				Type string
				// Content is the returned tool payload, never its arguments.
				Content json.RawMessage
			}
		}
	}
	json.Unmarshal(body, &request)
	var result strings.Builder
	for _, m := range request.Messages {
		for _, p := range m.Content {
			if p.Type == "tool_result" {
				result.Write(p.Content)
			}
		}
	}
	return result.String()
}
func fleetAgent(t *testing.T, server *httptest.Server) (ensemble.Agent, humanInbox, *[]ensemble.Agent) {
	t.Helper()
	app := ensemble.New(io.Discard)
	inbox := make(humanInbox, 10)
	app.SetUserSink("Bill", inbox)
	var children []ensemble.Agent
	var mu sync.Mutex
	app.SetAgentFactory(func(parent ensemble.Agent, spec ensemble.AgentSpec) (ensemble.Agent, error) {
		a, err := app.NewChildAgent(parent, parent.Config(), spec)
		if err == nil {
			mu.Lock()
			children = append(children, a)
			mu.Unlock()
		}
		return a, err
	})
	a, err := app.NewAgent(ensemble.Config{Model: "fake", BaseURL: server.URL, DataDir: t.TempDir(), BuiltinTools: true, MaxToolRounds: 10})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := a.Shutdown(); err != nil {
			t.Error(err)
		}
	})
	return a, inbox, &children
}

// A real three-generation runtime exercises the recursive factory, early join,
// addressed reply, explicit fan-in and separate saved conversations together.
func TestFleetRecursiveQuestionsIsolation(t *testing.T) {
	var mu sync.Mutex
	turns := map[string]int{}
	var replySeen bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		text := string(body)
		var prompt struct {
			Messages []struct{ Content []struct{ Text string } }
		}
		json.Unmarshal(body, &prompt)
		first := ""
		if len(prompt.Messages) > 0 && len(prompt.Messages[0].Content) > 0 {
			first = prompt.Messages[0].Content[0].Text
		}
		role := "root"
		if first == "CHILD_TASK" {
			role = "child"
		}
		if first == "LEAF_TASK" {
			role = "leaf"
		}
		mu.Lock()
		n := turns[role]
		turns[role]++
		mu.Unlock()
		switch role {
		case "root":
			switch n {
			case 0:
				fleetCall(w, "spawn_sub_agent", map[string]string{"agent_id": "child", "task": "CHILD_TASK"})
			case 1:
				fleetCall(w, "join_agents", map[string]any{"agent_ids": []string{"child"}, "ai_callback_delay": 30})
			default:
				if !strings.Contains(fleetResultText(body), "CHILD_RESULT") {
					t.Error("root join lost child's submission")
				}
				fleetCall(w, "submit", map[string]string{"result": "ROOT_RESULT"})
			}
		case "child":
			switch n {
			case 0:
				fleetCall(w, "spawn_sub_agent", map[string]string{"agent_id": "leaf", "task": "LEAF_TASK"})
			case 1:
				fleetCall(w, "join_agents", map[string]any{"agent_ids": []string{"leaf"}, "ai_callback_delay": 30})
			case 2:
				if !strings.Contains(text, "leaf question") {
					t.Error("join did not yield question")
				}
				fleetCall(w, "respond_to_parent_message", map[string]string{"agent_id": "leaf", "response": "use cobalt"})
			case 3:
				fleetCall(w, "join_agents", map[string]any{"agent_ids": []string{"leaf"}, "ai_callback_delay": 30})
			default:
				if !strings.Contains(fleetResultText(body), "LEAF_RESULT") {
					t.Error("child join lost leaf submission")
				}
				fleetCall(w, "submit", map[string]string{"result": "CHILD_RESULT"})
			}
		case "leaf":
			if n == 0 {
				fleetCall(w, "send_message_to_parent", map[string]any{"content": "leaf question", "ai_callback_delay": 30})
			} else {
				// Check the actual returned payload, not the original request arguments.
				returned := fleetResultText(body)
				if strings.Contains(returned, "use cobalt") && strings.Contains(returned, "child") {
					mu.Lock()
					replySeen = true
					mu.Unlock()
				}

				fleetCall(w, "submit", map[string]string{"result": "LEAF_RESULT"})
			}
		}
	}))
	t.Cleanup(server.Close)
	a, _, children := fleetAgent(t, server)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	result, err := a.Ask(ctx, "ROOT_TASK")
	if err != nil || result != "ROOT_RESULT" {
		t.Fatalf("recursive fleet: %q %v", result, err)
	}
	if err := a.Shutdown(); err != nil {
		t.Fatal(err)
	}
	if len(*children) != 2 || !replySeen {
		t.Fatalf("factory depth=%d addressed reply=%v", len(*children), replySeen)
	}
	child, leaf := (*children)[0], (*children)[1]
	// The join handler itself must yield the question. Generic message wakeup
	// alone could hide a still-deadlocked background join.
	joinedQuestion := false
	for _, event := range child.History().EventsAfter(0) {
		if event.Type == "tool_called" && event.Tool.Name == "join_agents" && event.Tool.Job != nil {
			data, _ := os.ReadFile(event.Tool.Job.Output.Locator)
			joinedQuestion = joinedQuestion || strings.Contains(string(data), "leaf question")
		}
	}
	if !joinedQuestion {
		t.Error("join handler never returned the child question")
	}
	if child.ParentAgent() != a || leaf.ParentAgent() != child {
		t.Fatal("routing IDs replaced actual parent links")
	}
	if len(a.Fleet().List()) != 0 {
		t.Fatal("cascade retained live children")
	}
	for _, c := range *children {
		path := filepath.Join(c.Config().DataDir, "save.json")
		data, err := os.ReadFile(path)
		if err != nil || !strings.Contains(string(data), c.Spec().ID) {
			t.Fatalf("child save %s: %v", path, err)
		}
		if !strings.HasPrefix(c.Config().DataDir, c.ParentAgent().Config().DataDir+string(os.PathSeparator)) {
			t.Fatal("child directory escaped owner tree")
		}
	}
}

// A human-facing root finishes without submit or reactive correction. Provider
// failures likewise remain one failed request rather than becoming retries.
func TestFleetRootFinishesWithoutSubmit(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint(fail), func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if fail {
					http.Error(w, "provider refused", http.StatusBadRequest)
					return
				}
				replyText(w, "plain idle answer")
			}))
			t.Cleanup(server.Close)
			a, inbox, _ := fleetAgent(t, server)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			result, err := a.Ask(ctx, "idle test")
			if (err != nil) != fail || calls.Load() != 1 || (!fail && result != "plain idle answer") {
				t.Fatalf("root completion: calls=%d result=%q err=%v", calls.Load(), result, err)
			}
			select {
			case m := <-inbox:
				t.Fatalf("root escalated ordinary completion: %+v", m)
			default:
			}
			for _, e := range a.History().EventsAfter(0) {
				if e.Type == "fleet_notice" || (e.Message != nil && strings.Contains(fmt.Sprint(e.Message.Parts), "Framework wake ")) {
					t.Fatal("root received a submission correction", e)
				}
			}
		})
	}
}

// An ordinary incoming prompt wakes the current slow shell wait before its
// command finishes. The real continuing job later returns its actual result.
func TestFleetMessageWakesWaitAndRetainsResult(t *testing.T) {
	started := make(chan struct{})
	steered := make(chan struct{})
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		switch calls.Add(1) {
		case 1:
			fleetCall(w, "tool_limits", map[string]int{"ai_callback_delay": 60})
		case 2:
			fleetCall(w, "run_command", map[string]string{"command": "sleep 1; printf REAL_JOB_RESULT"})
			close(started)
		case 3:
			if !strings.Contains(string(body), "STEER_NOW") {
				t.Error("wake request lacks message")
			}
			close(steered)
			// The command is the first ordinary job; tool_limits is supervision only.
			fleetCall(w, "wait_for_job", map[string]any{"handle": 1, "ai_callback_delay": 3})
		default:
			if !strings.Contains(fleetResultText(body), "REAL_JOB_RESULT") {
				t.Error("background result lost")
			}
			fleetCall(w, "submit", map[string]string{"result": "finished"})
		}
	}))
	t.Cleanup(server.Close)
	a, _, _ := fleetAgent(t, server)
	reply := a.Submit(context.Background(), "start slow work")
	<-started
	// Wait for the actual running job rather than guessing when HTTP was parsed.
	deadline := time.Now().Add(time.Second)
	for {
		if job, err := a.Jobs().Find(1); err == nil && job.Data().Status == "running" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("shell job never started")
		}
		time.Sleep(time.Millisecond)
	}
	if err := a.Post(ensemble.UserMessage{Text: "STEER_NOW"}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-steered:
		if job, err := a.Jobs().Find(1); err != nil || job.Data().Status != "running" {
			t.Fatal("wake did not precede command completion")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("message waited for command or timer")
	}
	result := receive(t, reply)
	if result.Err != nil || result.Text != "finished" {
		t.Fatal(result)
	}
	job, _ := a.Jobs().Find(1)
	data, err := os.ReadFile(job.Data().Output.Locator)
	if err != nil || string(data) != "REAL_JOB_RESULT\nexit_code: 0\n" {
		t.Fatalf("real result %q %v", data, err)
	}
}

// Shutdown preserves a child's save and recursive owner lifetime, while a later
// same-ID spawn restores the actual previous dialogue. Submission is bottom-up.
func TestFleetLifecycleAndRestore(t *testing.T) {
	entered := make(chan struct{}, 2)
	restored := make(chan bool, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "RESTORE_TASK") {
			restored <- strings.Contains(string(body), "ORIGINAL_TASK")
			fleetCall(w, "submit", map[string]string{"result": "restored result"})
			return
		}
		entered <- struct{}{}
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)
	a, _, children := fleetAgent(t, server)
	if _, err := a.Fleet().Spawn("worker", "ORIGINAL_TASK", ""); err != nil {
		t.Fatal(err)
	}
	<-entered
	if got := a.Fleet().List(); len(got) != 1 || got[0].ID != "worker" || got[0].State != "running" {
		t.Fatal("live child invisible", got)
	}
	if err := a.Fleet().Submit("too early"); err == nil || !strings.Contains(err.Error(), "worker") {
		t.Fatal("submit accepted active child", err)
	}
	for _, id := range []string{"Bill", "../escape", "."} {
		if _, err := a.Fleet().Spawn(id, "task", ""); err == nil {
			t.Fatal("reserved child address accepted", id)
		}
	}
	if err := a.Fleet().ShutdownChild("Bill"); err == nil {
		t.Fatal("human treated as fleet member")
	}
	if err := a.Fleet().ShutdownChild("worker"); err != nil {
		t.Fatal(err)
	}
	if len(a.Fleet().List()) != 0 {
		t.Fatal("shutdown child retained")
	}
	original := (*children)[0]
	if err := original.Post(ensemble.UserMessage{Text: "must refuse"}); err != ensemble.ErrStopped {
		t.Fatal("shutdown left actor alive", err)
	}
	if _, err := a.Fleet().Spawn("worker", "RESTORE_TASK", ""); err != nil {
		t.Fatal(err)
	}
	select {
	case ok := <-restored:
		if !ok {
			t.Fatal("same-ID child lost previous dialogue")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("restored child never ran")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	result, err := a.Fleet().Join(ctx, []string{"worker"})
	if err != nil || !strings.Contains(result, "restored result") {
		t.Fatal(result, err)
	}
	if err := a.Fleet().Submit("parent complete"); err != nil {
		t.Fatal(err)
	}
	worker := (*children)[len(*children)-1]
	// A successful submission resets the literal correction counter for new work.
	for i := 0; i < 2; i++ {
		worker.Fleet().Begin()
		for wake := 1; wake <= 3; wake++ {
			correction, escalation := worker.Fleet().Idle()
			if correction == "" || escalation != "" {
				t.Fatalf("wake %d after submit: %q/%q", wake, correction, escalation)
			}
		}
		if correction, escalation := worker.Fleet().Idle(); correction != "" || escalation == "" {
			t.Fatal("unbounded correction")
		}
		if err := worker.Fleet().Submit("reset"); err != nil {
			t.Fatal(err)
		}
	}
}

// Park the parent's real HTTP request so generic tool-wait wakeup cannot mask
// fan-in's own early-return contract. Only the child and manager run this question.
func TestFleetJoinQuestionYield(t *testing.T) {
	parentEntered := make(chan struct{})
	releaseParent := make(chan struct{})
	var once sync.Once
	t.Cleanup(func() { once.Do(func() { close(releaseParent) }) })
	var leafRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "PARK_PARENT") {
			close(parentEntered)
			select {
			case <-releaseParent:
			case <-r.Context().Done():
				return
			}
			fleetCall(w, "submit", map[string]string{"result": "parent finished"})
		} else if leafRequests.Add(1) == 1 {
			fleetCall(w, "send_message_to_parent", map[string]any{"content": "DIRECT_QUESTION", "ai_callback_delay": 30})
		} else {
			fleetCall(w, "submit", map[string]string{"result": "child finished"})
		}
	}))
	t.Cleanup(server.Close)
	a, _, _ := fleetAgent(t, server)
	parentReply := a.Submit(context.Background(), "PARK_PARENT")
	<-parentEntered
	if _, err := a.Fleet().Spawn("child", "ASK_A_QUESTION", ""); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	result, err := a.Fleet().Join(ctx, []string{"child"})
	if err != nil || !strings.Contains(result, "DIRECT_QUESTION") {
		t.Fatalf("join did not yield pending question before child submission: %q %v", result, err)
	}
	if err := a.Fleet().Reply("child", "", "answer"); err != nil {
		t.Fatal(err)
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
	defer cancel2()
	result, err = a.Fleet().Join(ctx2, []string{"child"})
	if err != nil || !strings.Contains(result, "child finished") {
		t.Fatal(result, err)
	}
	once.Do(func() { close(releaseParent) })
	if result := receive(t, parentReply); result.Err != nil {
		t.Fatal(result.Err)
	}
}

// Reviewer found sender-local IDs collide when an accepted child ID equals its
// parent's. Two real questions must coexist and route replies to their own jobs.
func TestFleetSameNameQuestionCorrelation(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	t.Cleanup(func() { once.Do(func() { close(release) }) })
	var childCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "PARK_ROOT") {
			close(entered)
			select {
			case <-release:
			case <-r.Context().Done():
				return
			}
			fleetCall(w, "submit", map[string]string{"result": "root done"})
		} else if childCalls.Add(1) == 1 {
			fleetCall(w, "send_message_to_parent", map[string]any{"content": "CHILD_QUESTION", "ai_callback_delay": 30})
		} else {
			if !strings.Contains(fleetResultText(body), "CHILD_ANSWER") {
				t.Error("child received another question's reply")
			}
			fleetCall(w, "submit", map[string]string{"result": "child done"})
		}
	}))
	t.Cleanup(server.Close)
	a, inbox, _ := fleetAgent(t, server)
	rootTurn := a.Submit(context.Background(), "PARK_ROOT")
	<-entered
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	humanReply := make(chan ensemble.FleetMessage, 1)
	go func() { reply, _ := a.Fleet().AskParent(ctx, "ROOT_QUESTION"); humanReply <- reply }()
	var humanMessage ensemble.FleetMessage
	select {
	case humanMessage = <-inbox:
	case <-ctx.Done():
		t.Fatal("root question missing")
	}
	if _, err := a.Fleet().Spawn("main", "CHILD_TASK", ""); err != nil {
		t.Fatal(err)
	}
	// Join also sees the human question; wait for the child's actual request to
	// enter the shared recipient map before examining correlation.
	var messages []ensemble.FleetMessage
	for {
		result, err := a.Fleet().Join(ctx, []string{"main"})
		if err != nil {
			t.Fatalf("pending question correlation incomplete: %d questions after %d child requests: %v", len(messages), childCalls.Load(), err)
		}
		json.Unmarshal([]byte(result), &messages)
		if len(messages) == 2 {
			break
		}
		if ctx.Err() != nil {
			t.Fatalf("pending question correlation incomplete: %d questions after %d child requests", len(messages), childCalls.Load())
		}
		time.Sleep(time.Millisecond)
	}
	if messages[0].ID == messages[1].ID {
		t.Fatal("question IDs collided")
	}
	if err := a.Fleet().Reply("main", "", "CHILD_ANSWER"); err != nil {
		t.Fatal("single child question was ambiguous with human question", err)
	}
	if err := a.Fleet().Reply("", humanMessage.ID, "HUMAN_ANSWER"); err != nil {
		t.Fatal("root question was lost", err)
	}
	select {
	case reply := <-humanReply:
		if reply.Content != "HUMAN_ANSWER" || reply.From != "Bill" {
			t.Fatal("root received wrong reply", reply)
		}
	case <-ctx.Done():
		t.Fatal("root question remained blocked")
	}
	if _, err := a.Fleet().Join(ctx, []string{"main"}); err != nil {
		t.Fatal(err)
	}
	once.Do(func() { close(release) })
	if result := receive(t, rootTurn); result.Err != nil {
		t.Fatal(result.Err)
	}
}

// A child that exhausted correction is stopped without a submission, not still
// running. Preserve that failure for status/join after escalation reaches its parent.
func TestFleetEscalatedChildStatus(t *testing.T) {
	var requests atomic.Int32
	parentNotified := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests.Add(1) == 5 {
			close(parentNotified)
		}
		replyText(w, "unsubmitted prose")
	}))
	t.Cleanup(server.Close)
	a, _, _ := fleetAgent(t, server)
	if _, err := a.Fleet().Spawn("worker", "TASK", ""); err != nil {
		t.Fatal(err)
	}
	select {
	case <-parentNotified:
	case <-time.After(2 * time.Second):
		t.Fatal("escalation never reached parent")
	}
	statuses := a.Fleet().List()
	if len(statuses) != 1 || statuses[0].State != "failed" || statuses[0].Result != "" || !strings.Contains(statuses[0].Error, "three framework wakes") {
		t.Fatal("exhausted child still claims running", statuses)
	}
	// A follow-up ask waits behind the parent's escalation turn, so the count
	// checks completed turns rather than racing its response handling.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := a.Ask(ctx, "follow-up"); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 6 {
		t.Fatal("expected four child requests and two ordinary parent turns", requests.Load())
	}
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, err := a.Fleet().Join(ctx, []string{"worker"}); err == nil || !strings.Contains(err.Error(), "three framework wakes") {
		t.Fatal("join concealed child failure", err)
	}
}

// Role instructions follow the actual parent relationship, even when a primary
// skill replaces the prompt or a grandchild inherits its parent's configuration.
func TestFleetSubmissionInstructionsOnlyForChildren(t *testing.T) {
	for _, primary := range []bool{false, true} {
		t.Run(fmt.Sprint(primary), func(t *testing.T) {
			app := ensemble.New(io.Discard)
			spec := ensemble.AgentSpec{ID: "main", DataDir: t.TempDir()}
			if primary {
				var err error
				spec.SkillsDir, err = filepath.Abs("../skills")
				if err != nil {
					t.Fatal(err)
				}
				spec.PrimarySkill = "ensemble"
			}
			root, err := app.NewAgent(ensemble.Config{Model: "fake", BuiltinTools: true, SystemPrompt: "Original instructions."}, spec)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { root.Shutdown() })
			marker := "You are a delegated sub-agent."
			if strings.Contains(root.Engine().SystemPrompt(), marker) || strings.Contains(root.Engine().SystemPrompt(), "Ordinary text does not submit") {
				t.Fatal("root acquired child completion instructions")
			}
			parent := root
			for _, id := range []string{"child", "grandchild"} {
				spec.ID, spec.DataDir = id, t.TempDir()
				child, err := app.NewChildAgent(parent, parent.Config(), spec)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { child.Shutdown() })
				prompt := child.Engine().SystemPrompt()
				if strings.Count(prompt, marker) != 1 || !strings.Contains(prompt, "then call submit with your completed result") {
					t.Fatalf("%s missing or duplicated submission instructions: %q", id, prompt)
				}
				parent = child
			}
		})
	}
}

// Actual spawned children use the parent's confinement, not just copied IDs.
// Allowed reads remain usable beside outside refusal and safe-mode tool denial.
func TestFleetSandboxInheritance(t *testing.T) {
	workspace, outside := t.TempDir(), t.TempDir()
	inside := filepath.Join(workspace, "allowed.txt")
	secret := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(inside, []byte("CHILD_ALLOWED"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(secret, []byte("CHILD_OUTSIDE"), 0600); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		switch calls.Add(1) {
		case 1:
			fleetCall(w, "read_file", map[string]string{"path": inside})
		case 2:
			if !strings.Contains(fleetResultText(body), "CHILD_ALLOWED") {
				t.Error("allowed child read missing")
			}
			fleetCall(w, "read_file", map[string]string{"path": secret})
		default:
			if strings.Contains(fleetResultText(body), "CHILD_OUTSIDE") {
				t.Error("child escaped inherited root")
			}
			fleetCall(w, "submit", map[string]string{"result": "confined"})
		}
	}))
	defer server.Close()
	app := ensemble.New(io.Discard)
	var child ensemble.Agent
	app.SetAgentFactory(func(parent ensemble.Agent, spec ensemble.AgentSpec) (ensemble.Agent, error) {
		a, err := app.NewChildAgent(parent, parent.Config(), spec)
		child = a
		return a, err
	})
	parent, err := app.NewAgent(ensemble.Config{Model: "fake", BaseURL: server.URL, BuiltinTools: true, DataDir: filepath.Join(workspace, "state")}, ensemble.AgentSpec{ID: "main", SandboxRoot: workspace, SafeMode: true})
	if err != nil {
		t.Fatal(err)
	}
	defer parent.Shutdown()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err = parent.Fleet().Spawn("worker", "READ_TASK", ""); err != nil {
		t.Fatal(err)
	}
	if _, err = parent.Fleet().Join(ctx, nil); err != nil {
		t.Fatal(err)
	}
	if err = child.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	if child.ParentAgent() != parent || child.Ensemble() != app || !child.Spec().SafeMode || child.Spec().EnableWebSearch {
		t.Fatal("child ownership or policy lost")
	}
	for _, d := range child.Declarations() {
		if d.Name == "run_command" {
			t.Fatal("safe child declared shell execution")
		}
	}
	if _, err = sandboxCall(t, child, "run_command", map[string]any{"command": "printf forbidden"}); err == nil {
		t.Fatal("safe child executed shell")
	}
	refused := false
	outsideCall := ""
	for _, e := range child.History().EventsAfter(0) {
		if e.Type == "tool_called" && e.Tool.Name == "read_file" {
			var args struct {
				// Path distinguishes the outside request from the allowed control.
				Path string
			}
			if err := json.Unmarshal(e.Tool.Args, &args); err != nil {
				t.Fatal(err)
			}
			if args.Path == secret {
				outsideCall = e.Tool.CallID
			}
		}
		// Call IDs may repeat across responses. Match this return only after
		// the outside read was dispatched, then consume that pending identity.
		if e.Type == "tool_returned" && outsideCall != "" && e.Tool.CallID == outsideCall {
			refused = e.Tool.IsError
			outsideCall = ""
		}
	}
	if calls.Load() != 3 || !refused {
		t.Fatal("outside read was not actually refused", calls.Load(), refused)
	}
	wider := ensemble.AgentSpec{ID: "wider", DataDir: filepath.Join(outside, "state"), SandboxRoot: outside, EnableWebSearch: true}
	bad, err := app.NewChildAgent(parent, parent.Config(), wider)
	if bad != nil {
		bad.Shutdown()
	}
	if err == nil || !strings.Contains(err.Error(), "permissions narrowed") {
		t.Fatal("widening bypassed child boundary", err)
	}
}
