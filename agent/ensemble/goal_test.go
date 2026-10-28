package ensemble_test

import (
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"ensemble/ensemble"
	"ensemble/internal/common"
)

type goalFixture struct {
	a        ensemble.Agent
	control  ensemble.GoalController
	file     string
	requests chan map[string]any
}

func newGoalFixture(t *testing.T, hold <-chan struct{}) goalFixture {
	return goalFixtureWith(t, hold, nil, nil)
}

func goalFixtureWith(t *testing.T, hold <-chan struct{}, configure func(*ensemble.Config, *ensemble.AgentSpec), respond func(http.ResponseWriter, map[string]any)) goalFixture {
	t.Helper()
	requests := make(chan map[string]any, 32)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
			http.Error(w, "bad request", 400)
			return
		}
		select {
		case requests <- body:
		default:
			t.Error("unexpected request flood")
		}
		if hold != nil {
			select {
			case <-hold:
			case <-r.Context().Done():
				return
			}
		}
		if respond != nil {
			respond(w, body)
			return
		}
		_, _ = io.WriteString(w, `{"content":[{"type":"text","text":"done"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	t.Cleanup(server.Close)
	dir := skillFiles(t, map[string]string{
		"base":  "---\nname: base\ntype: primary\ntools: read_file write_file\nloadable-skills: goals\n---\nPRIMARY_SENTINEL",
		"goals": "---\nname: goals\ntype: loadable\ntools: goal_control\n---\nGOAL_SKILL_SENTINEL",
	})
	root := ensemble.New(io.Discard)
	workspace := t.TempDir()
	cfg := ensemble.Config{Model: "fake", BuiltinTools: true, BaseURL: server.URL, DataDir: t.TempDir(), Workspace: workspace, Recall: ensemble.RecallSettings{Disabled: true}}
	spec := ensemble.AgentSpec{SkillsDir: dir, PrimarySkill: "base"}
	if configure != nil {
		configure(&cfg, &spec)
	}
	a, err := root.NewAgent(cfg, spec)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := a.Shutdown(); err != nil {
			t.Error(err)
		}
	})
	control, ok := a.(ensemble.GoalController)
	if !ok {
		t.Fatal("missing goal owner")
	}
	file := filepath.Join(workspace, "GOAL.md")
	if err := os.WriteFile(file, []byte("GOAL_ONE"), 0600); err != nil {
		t.Fatal(err)
	}
	return goalFixture{a, control, file, requests}
}
func (f goalFixture) load(t *testing.T) {
	t.Helper()
	if _, err := f.a.Skills().LoadDynamic("goals", 0); err != nil {
		t.Fatal(err)
	}
}
func (f goalFixture) configure(t *testing.T, enabled bool, delay float64) ensemble.GoalStatus {
	t.Helper()
	s, err := f.control.ConfigureGoal(ensemble.GoalConfig{File: f.file, Enabled: enabled, IdleSeconds: delay})
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func goalRequest(t *testing.T, requests <-chan map[string]any) map[string]any {
	t.Helper()
	select {
	case r := <-requests:
		return r
	case <-time.After(3 * time.Second):
		t.Fatal("no request")
		return nil
	}
}
func noGoalRequest(t *testing.T, requests <-chan map[string]any, delay time.Duration) {
	t.Helper()
	select {
	case <-requests:
		t.Fatal("unexpected automatic request")
	case <-time.After(delay):
	}
}
func goalStatus(t *testing.T, f goalFixture, phase string) ensemble.GoalStatus {
	t.Helper()
	end := time.Now().Add(3 * time.Second)
	for time.Now().Before(end) {
		s := f.control.GoalStatus()
		if s.Phase == phase {
			return s
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("want %s, got %+v", phase, f.control.GoalStatus())
	return ensemble.GoalStatus{}
}
func goalSystem(body map[string]any) string { b, _ := json.Marshal(body["system"]); return string(b) }

func TestGoalsOptInFreshSystemAndUnload(t *testing.T) {
	f := newGoalFixture(t, nil)
	if got := f.control.GoalStatus().IdleSeconds; got != 480 {
		t.Fatalf("default idle delay = %v", got)
	}
	if slices.Contains(toolNames(f.a), "goal_control") {
		t.Fatal("goal control disclosed before load")
	}
	if _, err := f.control.ConfigureGoal(ensemble.GoalConfig{File: f.file, Enabled: true, IdleSeconds: .03}); err == nil {
		t.Fatal("enabled without skill")
	}
	noGoalRequest(t, f.requests, 40*time.Millisecond)
	f.load(t)
	if !slices.Contains(toolNames(f.a), "goal_control") {
		t.Fatal("loaded control missing")
	}
	if s := f.control.GoalStatus(); s.Enabled || s.File != "" {
		t.Fatalf("load activated goal: %+v", s)
	}
	f.configure(t, false, .05)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, text := range []string{"GOAL_ONE", "GOAL_TWO"} {
		if err := os.WriteFile(f.file, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := f.a.Ask(ctx, "ordinary request"); err != nil {
			t.Fatal(err)
		}
		body := goalRequest(t, f.requests)
		system := goalSystem(body)
		if strings.Count(system, text) != 1 || !strings.Contains(system, "PRIMARY_SENTINEL") {
			t.Fatalf("wrong SYSTEM: %s", system)
		}
		messages, _ := json.Marshal(body["messages"])
		if strings.Contains(string(messages), text) {
			t.Fatal("goal copied into dialogue")
		}
		if text == "GOAL_TWO" && strings.Contains(system, "GOAL_ONE") {
			t.Fatal("stale goal")
		}
	}
	noGoalRequest(t, f.requests, 80*time.Millisecond)
	if err := f.a.Skills().MarkUnload("goals"); err != nil {
		t.Fatal(err)
	}
	if slices.Contains(toolNames(f.a), "goal_control") {
		t.Fatal("pending unload retained goal control")
	}
	if _, err := f.a.Ask(ctx, "after unload"); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(goalSystem(goalRequest(t, f.requests)), "GOAL_TWO") {
		t.Fatal("unload retained goal in SYSTEM")
	}
	goalStatus(t, f, "none")
	if _, err := os.Stat(f.file); err != nil {
		t.Fatal("unload changed goal file", err)
	}
}

func TestGoalsIdlePauseDisableAndFileFailure(t *testing.T) {
	f := newGoalFixture(t, nil)
	f.load(t)
	f.a.Pause()
	f.configure(t, true, .06)
	goalStatus(t, f, "paused")
	noGoalRequest(t, f.requests, 100*time.Millisecond)
	f.a.Unpause()
	body := goalRequest(t, f.requests)
	if !strings.Contains(goalSystem(body), "GOAL_ONE") {
		t.Fatal("wake did not read goal")
	}
	messages, _ := json.Marshal(body["messages"])
	if !strings.Contains(string(messages), "Automatic goal reminder") {
		t.Fatal("missing labelled wake")
	}
	f.configure(t, false, .06)
	noGoalRequest(t, f.requests, 100*time.Millisecond)
	if err := os.Remove(f.file); err != nil {
		t.Fatal(err)
	}
	// An off switch cannot depend on reading a now-missing configured file.
	f.configure(t, false, .06)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := f.a.Ask(ctx, "help repair the goal"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(goalSystem(goalRequest(t, f.requests)), "Current goal unavailable") {
		t.Fatal("missing repair notice")
	}
	s := goalStatus(t, f, "error")
	if s.Enabled || s.Error == "" {
		t.Fatalf("untruthful error status: %+v", s)
	}
}

func TestGoalsDoNotWakeBusyTurn(t *testing.T) {
	release := make(chan struct{})
	f := newGoalFixture(t, release)
	f.load(t)
	f.configure(t, true, .08)
	done := make(chan error, 1)
	go func() { _, err := f.a.Ask(context.Background(), "held request"); done <- err }()
	goalRequest(t, f.requests)
	noGoalRequest(t, f.requests, 160*time.Millisecond)
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	// A new full idle interval, not the expired interval from before the turn.
	noGoalRequest(t, f.requests, 30*time.Millisecond)
	goalRequest(t, f.requests)
	f.configure(t, false, .08)
	noGoalRequest(t, f.requests, 120*time.Millisecond)
}

// HTTP failure returns the real actor to idle; a held request is covered separately.
func TestGoalsFailedRequestRearms(t *testing.T) {
	f := goalFixtureWith(t, nil, nil, func(w http.ResponseWriter, _ map[string]any) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	})
	f.load(t)
	if _, err := f.control.ConfigureGoal(ensemble.GoalConfig{File: f.file, Enabled: true, IdleSeconds: .1}); err != nil {
		t.Fatal(err)
	}
	goalRequest(t, f.requests)
	status := goalStatus(t, f, "waiting")
	if !status.Enabled || status.NextReminder.IsZero() {
		t.Fatalf("not rearmed: %+v", status)
	}
	goalRequest(t, f.requests)
	if _, err := f.control.ConfigureGoal(ensemble.GoalConfig{Enabled: false}); err != nil {
		t.Fatal(err)
	}
	noGoalRequest(t, f.requests, 150*time.Millisecond)
}

func TestGoalsInterruptShutdownAndIsolation(t *testing.T) {
	f := newGoalFixture(t, nil)
	f.load(t)
	other := newGoalFixture(t, nil)
	other.load(t)
	cfg := ensemble.GoalConfig{File: f.file, Enabled: true, IdleSeconds: .1}
	if _, err := f.control.ConfigureGoal(cfg); err != nil {
		t.Fatal(err)
	}
	if err := f.a.Post(ensemble.Interrupt{}); err != nil {
		t.Fatal(err)
	}
	goalStatus(t, f, "off")
	noGoalRequest(t, f.requests, 150*time.Millisecond)
	if s := other.control.GoalStatus(); s.File != "" || s.Enabled {
		t.Fatalf("inherited goal: %+v", s)
	}
	if _, err := f.control.ConfigureGoal(cfg); err != nil {
		t.Fatal(err)
	}
	f.a.Shutdown()
	noGoalRequest(t, f.requests, 150*time.Millisecond)
	if _, err := f.control.ConfigureGoal(cfg); err == nil {
		t.Fatal("configured stopped actor")
	}
}

func TestGoalsSystemWireLanes(t *testing.T) {
	for _, tc := range []struct {
		name         string
		vendor       ensemble.Vendor
		model, field string
	}{
		{"anthropic", ensemble.Anthropic, "claude-sonnet-4-6", "system"},
		{"chat", ensemble.OpenAI, "gpt-4o-mini", "messages"},
		{"responses", ensemble.OpenAI, "gpt-ch19-course", "input"},
		{"codex", ensemble.OpenAI, "gpt-ch19-course", "instructions"},
		{"gemini", ensemble.Gemini, "gemini-2.5-pro", "systemInstruction"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := goalFixtureWith(t, nil, func(c *ensemble.Config, _ *ensemble.AgentSpec) {
				c.Vendor = tc.vendor
				c.Model = tc.model
				if tc.name == "codex" {
					c.Credentials = goalCodexCredentials{}
				}
			}, nil)
			f.load(t)
			if _, err := f.control.ConfigureGoal(ensemble.GoalConfig{File: f.file}); err != nil {
				t.Fatal(err)
			}
			for _, text := range []string{"first goal", "edited goal"} {
				if err := os.WriteFile(f.file, []byte(text), 0600); err != nil {
					t.Fatal(err)
				}
				wire, err := f.a.Engine().Render(f.a.History().Context())
				if err != nil {
					t.Fatal(err)
				}
				var body map[string]json.RawMessage
				if err := json.Unmarshal(wire, &body); err != nil {
					t.Fatal(err)
				}
				lane := body[tc.field]
				if tc.field == "messages" || tc.field == "input" {
					var items []map[string]json.RawMessage
					if err := json.Unmarshal(lane, &items); err != nil {
						t.Fatal(err)
					}
					if len(items) == 0 {
						t.Fatalf("missing instruction lane: %s", wire)
					}
					want := `"system"`
					if tc.field == "input" {
						want = `"developer"`
					}
					if string(items[0]["role"]) != want {
						t.Fatalf("wrong role: %s", items[0]["role"])
					}
					lane = items[0]["content"]
				}
				if strings.Count(string(lane), text) != 1 || strings.Count(string(wire), text) != 1 {
					t.Fatalf("goal not once in SYSTEM: %s", wire)
				}
				history, _ := json.Marshal(f.a.History().Context().Dialogue)
				if strings.Contains(string(history), text) {
					t.Fatal("goal copied into history")
				}
			}
		})
	}
}

func TestGoalsPreflightFailureRearms(t *testing.T) {
	save := filepath.Join(t.TempDir(), "save.json")
	f := goalFixtureWith(t, nil, func(c *ensemble.Config, _ *ensemble.AgentSpec) { c.SavePath = save }, nil)
	f.load(t)
	if err := f.a.History().Checkpoint(save); err != nil {
		t.Fatal(err)
	}
	journal := save + ".journal"
	if err := os.Remove(journal); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if err := os.Mkdir(journal, 0700); err != nil {
		t.Fatal(err)
	}
	initial := f.configure(t, true, .08)
	noGoalRequest(t, f.requests, 140*time.Millisecond)
	s := f.control.GoalStatus()
	if s.Phase != "waiting" || !s.NextReminder.After(initial.NextReminder) {
		t.Errorf("preflight failure stranded reminder: %+v", s)
	}
	// Even after an automatic preflight refusal, an ordinary render is not automatic.
	if err := os.Remove(f.file); err != nil {
		t.Fatal(err)
	}
	prompt, err := f.control.GoalSystemPrompt()
	if err != nil || !strings.Contains(prompt, "Current goal unavailable") {
		t.Errorf("automatic flag leaked beyond refused turn: %q %v", prompt, err)
	}
}

func TestGoalsDispatchedControlAndSameTurnEdit(t *testing.T) {
	var file string
	var calls int
	f := goalFixtureWith(t, nil, nil, func(w http.ResponseWriter, body map[string]any) {
		calls++
		switch calls {
		case 1:
			input, _ := json.Marshal(map[string]any{"file": file, "enabled": false})
			io.WriteString(w, `{"content":[{"type":"tool_use","id":"configure","name":"goal_control","input":`+string(input)+`}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		case 2:
			if !strings.Contains(goalSystem(body), "GOAL_ONE") {
				t.Error("dispatched configuration absent from next request")
			}
			input, _ := json.Marshal(map[string]any{"path": file, "content": "SAME_TURN_EDIT", "overwrite": true})
			io.WriteString(w, `{"content":[{"type":"tool_use","id":"edit","name":"write_file","input":`+string(input)+`}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		case 3:
			if !strings.Contains(goalSystem(body), "SAME_TURN_EDIT") || strings.Contains(goalSystem(body), "GOAL_ONE") {
				t.Error("same-turn edit not reflected in SYSTEM")
			}
			io.WriteString(w, `{"content":[{"type":"tool_use","id":"status","name":"goal_control","input":{}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		default:
			io.WriteString(w, `{"content":[{"type":"text","text":"done"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		}
	})
	file = f.file
	f.load(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := f.a.Ask(ctx, "configure, edit the goal, and report status"); err != nil {
		t.Fatal(err)
	}
	if calls != 4 {
		t.Fatalf("got %d requests", calls)
	}
	// Inspect the actual tool result on the fourth request, not just the local API.
	var last map[string]any
	for i := 0; i < calls; i++ {
		last = goalRequest(t, f.requests)
	}
	messages, _ := json.Marshal(last["messages"])
	if !strings.Contains(string(messages), `\"idle_seconds\":480`) || !strings.Contains(string(messages), `\"enabled\":false`) {
		t.Fatalf("missing paired control status: %s", messages)
	}
	if s := f.control.GoalStatus(); s.Enabled || s.File != file {
		t.Fatalf("wrong control state: %+v", s)
	}
}

type goalObserver func(ensemble.Observation)

func (f goalObserver) Observe(o ensemble.Observation) { f(o) }

func TestGoalsQueuedCallbackYieldsToInterrupt(t *testing.T) {
	f := newGoalFixture(t, nil)
	f.load(t)
	armed, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	f.a.Observe(goalObserver(func(o ensemble.Observation) {
		if g, ok := o.(ensemble.GoalChanged); ok && g.Status.Phase == "waiting" {
			once.Do(func() { close(armed); <-release })
		}
	}))
	done := make(chan error, 1)
	go func() {
		_, err := f.control.ConfigureGoal(ensemble.GoalConfig{File: f.file, Enabled: true, IdleSeconds: .02})
		done <- err
	}()
	<-armed
	// The real timer callback can enqueue while publication holds the actor.
	time.Sleep(60 * time.Millisecond)
	if err := f.a.Post(ensemble.Interrupt{}); err != nil {
		t.Fatal(err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	goalStatus(t, f, "off")
	noGoalRequest(t, f.requests, 70*time.Millisecond)
	// A new generation must receive its own complete delay.
	f.configure(t, true, .1)
	noGoalRequest(t, f.requests, 45*time.Millisecond)
	goalRequest(t, f.requests)
	f.configure(t, false, .1)
}

func TestGoalsInvalidUpdatesAndRepair(t *testing.T) {
	f := newGoalFixture(t, nil)
	f.load(t)
	original := f.configure(t, false, 480)
	for _, delay := range []float64{-1, math.NaN(), math.Inf(1), 1e20, 1e-20} {
		if _, err := f.control.ConfigureGoal(ensemble.GoalConfig{File: f.file, IdleSeconds: delay}); err == nil {
			t.Fatalf("accepted delay %v", delay)
		}
		if got := f.control.GoalStatus(); got != original {
			t.Fatalf("invalid update changed state: %+v", got)
		}
	}
	if err := os.WriteFile(f.file, nil, 0600); err != nil {
		t.Fatal(err)
	}
	prompt, err := f.control.GoalSystemPrompt()
	if err != nil || !strings.Contains(prompt, "Current goal unavailable") {
		t.Fatalf("missing repair notice: %q %v", prompt, err)
	}
	goalStatus(t, f, "error")
	if _, err := f.control.ConfigureGoal(ensemble.GoalConfig{File: f.file, Enabled: false}); err != nil {
		t.Fatalf("off switch required readable goal: %v", err)
	}
	if err := os.WriteFile(f.file, []byte("REPAIRED"), 0600); err != nil {
		t.Fatal(err)
	}
	prompt, err = f.control.GoalSystemPrompt()
	if err != nil || !strings.Contains(prompt, "REPAIRED") {
		t.Fatal(prompt, err)
	}
	if f.control.GoalStatus().Enabled {
		t.Fatal("repair re-enabled reminders")
	}
	if _, err := f.control.ConfigureGoal(ensemble.GoalConfig{}); err != nil {
		t.Fatal(err)
	}
	if s := f.control.GoalStatus(); s.File != "" || s.Enabled {
		t.Fatal(s)
	}
}

func TestGoalsCompactionIsolationAndHandoff(t *testing.T) {
	f := newGoalFixture(t, nil)
	f.load(t)
	f.configure(t, false, 480)
	if err := f.a.History().Append(ensemble.Event{Type: "micro_handoff", Handoff: &common.MicroHandoffData{Text: "current note"}}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := f.a.Ask(ctx, "after handoff"); err != nil {
		t.Fatal(err)
	}
	if system := goalSystem(goalRequest(t, f.requests)); strings.Count(system, "GOAL_ONE") != 1 {
		t.Fatal("handoff lost goal")
	}
	child, done, err := f.a.Fleet().StartCompaction("goal-isolation", "summarize only this synthetic source")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { child.Shutdown() })
	body := goalRequest(t, f.requests)
	if strings.Contains(goalSystem(body), "GOAL_ONE") {
		t.Fatal("compaction inherited goal prompt")
	}
	controller := child.(ensemble.GoalController)
	if s := controller.GoalStatus(); s.File != "" || s.Enabled {
		t.Fatalf("compaction inherited state: %+v", s)
	}
	if _, err := controller.ConfigureGoal(ensemble.GoalConfig{File: f.file, Enabled: true}); err == nil {
		t.Fatal("compaction acquired reminders")
	}
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("compaction did not finish")
	}
}

func TestGoalsDispatchDenialBeforeLoad(t *testing.T) {
	var calls int
	f := goalFixtureWith(t, nil, nil, func(w http.ResponseWriter, body map[string]any) {
		calls++
		if calls == 1 {
			tools, _ := json.Marshal(body["tools"])
			if strings.Contains(string(tools), "goal_control") {
				t.Error("unloaded control disclosed")
			}
			io.WriteString(w, `{"content":[{"type":"tool_use","id":"refused","name":"goal_control","input":{}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		} else {
			io.WriteString(w, `{"content":[{"type":"text","text":"done"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := f.a.Ask(ctx, "try an ungranted control"); err != nil {
		t.Fatal(err)
	}
	goalRequest(t, f.requests)
	body := goalRequest(t, f.requests)
	var refused bool
	for _, item := range body["messages"].([]any) {
		content, ok := item.(map[string]any)["content"].([]any)
		if !ok {
			continue
		}
		for _, part := range content {
			p := part.(map[string]any)
			if p["tool_use_id"] == "refused" && p["is_error"] == true {
				refused = true
			}
		}
	}
	if !refused {
		t.Fatal("ungranted call lacked paired refusal")
	}
}

func TestGoalsOptionalBuiltinsAndRootedRead(t *testing.T) {
	f := goalFixtureWith(t, nil, func(c *ensemble.Config, s *ensemble.AgentSpec) {
		c.BuiltinTools = false
		c.DataDir = filepath.Join(c.Workspace, "agent-data")
		s.SandboxRoot = c.Workspace
	}, nil)
	f.load(t)
	if !strings.Contains(strings.Join(toolNames(f.a), ","), "goal_control") {
		t.Fatal("loaded control depends on file builtins")
	}
	original := f.configure(t, false, 480)
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(outside, []byte("OUTSIDE_GOAL_SECRET"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(filepath.Dir(f.file), "outside-link")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{outside, link, "../outside.md"} {
		if _, err := f.control.ConfigureGoal(ensemble.GoalConfig{File: path}); err == nil {
			t.Fatalf("accepted outside goal %q", path)
		}
		if got := f.control.GoalStatus(); got != original {
			t.Fatal("refusal changed goal state")
		}
	}
	prompt, err := f.control.GoalSystemPrompt()
	if err != nil || !strings.Contains(prompt, "GOAL_ONE") || strings.Contains(prompt, "OUTSIDE_GOAL_SECRET") {
		t.Fatal(prompt, err)
	}
}

func TestGoalsQueuedOrdinaryInputWins(t *testing.T) {
	f := newGoalFixture(t, nil)
	f.load(t)
	armed, release := make(chan struct{}), make(chan struct{})
	t.Cleanup(func() {
		select {
		case <-release:
		default:
			close(release)
		}
	})
	var once sync.Once
	f.a.Observe(goalObserver(func(o ensemble.Observation) {
		if g, ok := o.(ensemble.GoalChanged); ok && g.Status.Phase == "waiting" {
			once.Do(func() { close(armed); <-release })
		}
	}))
	configured := make(chan error, 1)
	go func() {
		_, err := f.control.ConfigureGoal(ensemble.GoalConfig{File: f.file, Enabled: true, IdleSeconds: .1})
		configured <- err
	}()
	<-armed
	time.Sleep(140 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result := f.a.Submit(ctx, "ORDINARY_INPUT_WINS")
	close(release)
	if err := <-configured; err != nil {
		t.Fatal(err)
	}
	body := goalRequest(t, f.requests)
	wire, _ := json.Marshal(body["messages"])
	if !strings.Contains(string(wire), "ORDINARY_INPUT_WINS") || strings.Contains(string(wire), "Automatic goal reminder") {
		t.Fatalf("wrong admission order: %s", wire)
	}
	if r := <-result; r.Err != nil {
		t.Fatal(r.Err)
	}
	noGoalRequest(t, f.requests, 45*time.Millisecond)
	goalRequest(t, f.requests)
	f.configure(t, false, .1)
}

func TestGoalsWakeLeavesBackgroundJobAlone(t *testing.T) {
	f := newGoalFixture(t, nil)
	f.load(t)
	job, err := f.a.Jobs().Start()
	if err != nil {
		t.Fatal(err)
	}
	defer job.Finish("owned job completed", nil)
	before := job.Data()
	f.configure(t, true, .03)
	goalRequest(t, f.requests)
	f.configure(t, false, .03)
	if job.Context().Err() != nil {
		t.Fatal("reminder canceled background job")
	}
	if after := job.Data(); !reflect.DeepEqual(after, before) {
		t.Fatalf("job changed: %+v => %+v", before, after)
	}
}

type goalCodexCredentials struct{}

func (goalCodexCredentials) Kind() ensemble.CredentialKind                  { return ensemble.CredentialCodex }
func (goalCodexCredentials) GetBearerToken(context.Context) (string, error) { return "dummy", nil }
