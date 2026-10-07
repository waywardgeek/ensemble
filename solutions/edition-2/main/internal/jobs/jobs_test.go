package jobs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
	"unicode/utf8"

	"example.com/ensemble/internal/common"
)

type testRoot struct {
	mu   sync.Mutex
	next uint64
}

func (*testRoot) Logf(string, ...any)          {}
func (*testRoot) Publish(string, common.Event) {}
func (r *testRoot) AllocateHandle() uint64     { r.mu.Lock(); defer r.mu.Unlock(); r.next++; return r.next }

type testAgent struct {
	root         *testRoot
	workspace    string
	mu           sync.Mutex
	events       []common.Event
	fault        error
	failTerminal bool
	registry     *testRegistry
}

func (a *testAgent) Ensemble() common.Ensemble { return a.root }
func (a *testAgent) Config() common.Config     { return common.Config{Workspace: a.workspace} }
func (a *testAgent) Workspace() string         { return a.workspace }
func (a *testAgent) Registry() common.Registry { return a.registry }
func (a *testAgent) Fault(err error)           { a.mu.Lock(); defer a.mu.Unlock(); a.fault = err }
func (a *testAgent) RecordJob(e common.Event) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.failTerminal && strings.HasPrefix(e.Type, "job_") {
		return errors.New("injected terminal persistence failure")
	}
	a.events = append(a.events, e)
	return nil
}

type testRegistry struct {
	parent  *testAgent
	handler func(common.Part, common.Job) *common.ExecutionResult
}

func (r *testRegistry) ResolveLimits(common.Part) (common.Limits, string, error) {
	return common.Limits{}, "", nil
}
func (r *testRegistry) Supervise(common.Part, common.Limits, string) error { return nil }
func (r *testRegistry) Agent() common.Agent                                { return r.parent }
func (*testRegistry) Declarations() []common.ToolDefinition                { return nil }
func (r *testRegistry) Execute(p common.Part) common.ToolEvent {
	result := r.handler(p, nil)
	text := result.Text + result.Note
	return common.ToolEvent{Parts: []common.Part{{Type: "text", Text: &text}}, IsError: result.IsError}
}
func (r *testRegistry) ExecuteJob(p common.Part, j common.Job) *common.ExecutionResult {
	return r.handler(p, j)
}
func (*testRegistry) Kind(name string) (bool, bool) {
	return true, name == "wait_for_job" || name == "send_input" || name == "kill_job" || name == "tool_limits"
}
func harness(t *testing.T) (*Service, *testAgent) {
	t.Helper()
	a := &testAgent{root: &testRoot{}, workspace: t.TempDir()}
	a.registry = &testRegistry{parent: a}
	s := New(a)
	t.Cleanup(func() { s.Close() })
	return s, a
}
func local(text string, failed bool) *common.ExecutionResult {
	return &common.ExecutionResult{IsError: failed, Text: text}
}
func launch(t *testing.T, s *Service, a *testAgent, command string) common.Job {
	t.Helper()
	a.registry.handler = func(p common.Part, j common.Job) *common.ExecutionResult {
		if err := j.StartProcess(command, ""); err != nil {
			return local(err.Error(), true)
		}
		return nil
	}
	j, err := s.Create()
	if err != nil {
		t.Fatal(err)
	}
	s.Start(j, common.Part{Name: "run_command"})
	return j
}
func report(t *testing.T, s *Service, j common.Job, delay time.Duration, budget int) common.ToolEvent {
	t.Helper()
	if err := s.Report(j, common.JobReport{CallID: "original", Limits: common.Limits{Delay: delay, MaxBytes: budget}, Original: true, MatchStart: -1}); err != nil {
		t.Fatal(err)
	}
	a := s.parent.(*testAgent)
	a.mu.Lock()
	defer a.mu.Unlock()
	return *a.events[len(a.events)-1].Tool
}

// Test helper expresses typed lifecycle operations; wire validation is tested
// through the Registry and public Agent integration paths.
func supervise(t *testing.T, s *Service, name, args string, limits common.Limits) common.ToolEvent {
	t.Helper()
	var data struct {
		Handle        uint64
		Input         string
		AppendNewline *bool `json:"append_newline"`
	}
	if err := json.Unmarshal([]byte(args), &data); err != nil {
		t.Fatal(err)
	}
	j, err := s.Lookup(data.Handle)
	if err != nil {
		t.Fatal(err)
	}
	request := common.JobReport{CallID: name, Limits: limits, MatchStart: -1}
	switch name {
	case "send_input":
		if data.AppendNewline == nil || *data.AppendNewline {
			data.Input += "\n"
		}
		request.MatchStart, err = s.Send(j, data.Input)
	case "kill_job":
		err = s.Kill(j, "kill_job")
	}
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Report(j, request); err != nil {
		t.Fatal(err)
	}
	a := s.parent.(*testAgent)
	a.mu.Lock()
	defer a.mu.Unlock()
	return *a.events[len(a.events)-1].Tool
}
func output(e common.ToolEvent) string { return *e.Parts[0].Text }
func TestProcessDrainsStatusAndRetainedOutput(t *testing.T) {
	s, a := harness(t)
	j := launch(t, s, a, "i=0; while [ $i -lt 20000 ]; do printf abcdefgh; printf ijklmnop >&2; i=$((i+1)); done; printf drained > marker; exit 7")
	result := report(t, s, j, 5*time.Second, 8)
	if result.IsError || result.Job.Status != "done" || result.Job.ExitCode == nil || *result.Job.ExitCode != 7 {
		t.Fatal(result)
	}
	if len(output(result)) > 350 || !strings.Contains(output(result), "truncated:") {
		t.Fatal(output(result))
	}
	data, err := os.ReadFile(filepath.Join(a.workspace, "cr/io/1"))
	if err != nil || len(data) != 320014 || !strings.HasSuffix(string(data), "\nexit_code: 7\n") {
		t.Fatal(len(data), err)
	}
	marker, _ := os.ReadFile(filepath.Join(a.workspace, "marker"))
	if string(marker) != "drained" {
		t.Fatal("did not drain")
	}
	again := supervise(t, s, "wait_for_job", `{"handle":1}`, common.Limits{Delay: time.Second, MaxBytes: 8})
	if strings.Contains(output(again), "abcdefgh") || again.IsError {
		t.Fatal(output(again))
	}
}
func TestCursorCapAndUTF8(t *testing.T) {
	for _, text := range []string{"abcdefghij", "ééééé"} {
		t.Run(text, func(t *testing.T) {
			s, a := harness(t)
			a.registry.handler = func(common.Part, common.Job) *common.ExecutionResult { return local(text, false) }
			j, err := s.Create()
			if err != nil {
				t.Fatal(err)
			}
			s.Start(j, common.Part{})
			first := report(t, s, j, time.Second, 4)
			got := output(first)
			if !utf8.ValidString(got) || strings.Contains(got, "\uFFFD") || !strings.Contains(got, "6 bytes omitted") {
				t.Fatal(got)
			}
			if text == "abcdefghij" && (!strings.Contains(got, "\nab\n") || !strings.HasSuffix(got, "ij")) {
				t.Fatal(got)
			}
			next := supervise(t, s, "wait_for_job", `{"handle":1}`, common.Limits{MaxBytes: 4})
			if strings.Contains(output(next), "omitted") || strings.HasSuffix(output(next), "ij") {
				t.Fatal(output(next))
			}
		})
	}
	s, a := harness(t)
	a.registry.handler = func(common.Part, common.Job) *common.ExecutionResult { return local("éX", false) }
	j, _ := s.Create()
	s.Start(j, common.Part{})
	if got := output(report(t, s, j, time.Second, 3)); got != "éX" {
		t.Fatal(got)
	}
}
func TestPatternInputAndKillGroup(t *testing.T) {
	s, a := harness(t)
	j := launch(t, s, a, "printf 'ready>'; while IFS= read -r line; do sleep 0.15; printf 'reply:%s ready>' \"$line\"; done")
	limits := common.Limits{Delay: 2 * time.Second, Pattern: regexp.MustCompile("ready>"), MaxBytes: 16384}
	var err error
	if err = s.Report(j, common.JobReport{Limits: limits, Original: true, MatchStart: -1}); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	reply := supervise(t, s, "send_input", `{"handle":1,"input":"hello"}`, limits)
	if time.Since(start) < 140*time.Millisecond || !strings.Contains(output(reply), "reply:hello ready>") {
		t.Fatal("old prompt matched", output(reply))
	}
	exact := supervise(t, s, "send_input", `{"handle":1,"input":"exact\n","append_newline":false}`, limits)
	if !strings.Contains(output(exact), "reply:exact") {
		t.Fatal(output(exact))
	}
	killed := supervise(t, s, "kill_job", `{"handle":1}`, limits)
	if !strings.Contains(output(killed), "status: killed") || j.Snapshot().Status != "killed" {
		t.Fatal(output(killed))
	}
	again := supervise(t, s, "wait_for_job", `{"handle":1}`, limits)
	if again.IsError || strings.Contains(output(again), "exit_code:") {
		t.Fatal(output(again))
	}
	supervise(t, s, "kill_job", `{"handle":1}`, limits)
	terminal := 0
	for _, e := range a.events {
		if e.Type == "job_killed" {
			terminal++
		}
	}
	if terminal != 1 {
		t.Fatal(terminal)
	}
}
func TestShutdownKillsChildAndFailureStillCleans(t *testing.T) {
	for _, fault := range []bool{false, true} {
		t.Run(fmt.Sprint(fault), func(t *testing.T) {
			s, a := harness(t)
			j := launch(t, s, a, "sleep 60 & child=$!; echo $child > child.pid; echo $$ > shell.pid; wait")
			report(t, s, j, 100*time.Millisecond, 1024)
			childBytes, err := os.ReadFile(filepath.Join(a.workspace, "child.pid"))
			if err != nil {
				t.Fatal(err)
			}
			var child int
			fmt.Sscan(string(childBytes), &child)
			a.failTerminal = fault
			err = s.Close()
			if (err != nil) != fault {
				t.Fatal(err)
			}
			if j.Snapshot().Status != "killed" || j.Snapshot().Reason != "shutdown" {
				t.Fatal(j.Snapshot())
			}
			deadline := time.Now().Add(time.Second)
			for syscall.Kill(child, 0) == nil && time.Now().Before(deadline) {
				time.Sleep(10 * time.Millisecond)
			}
			if syscall.Kill(child, 0) == nil {
				t.Fatal("child survived process-group kill", child)
			}
		})
	}
}
func TestLocalKillDiscardsLateResultAndErrorWaitSucceeds(t *testing.T) {
	s, a := harness(t)
	gate := make(chan struct{})
	started := make(chan struct{})
	a.registry.handler = func(common.Part, common.Job) *common.ExecutionResult {
		close(started)
		<-gate
		return local("late must not appear", false)
	}
	j, _ := s.Create()
	s.Start(j, common.Part{})
	<-started
	if err := s.kill(j.(*job), "kill_job"); err != nil {
		t.Fatal(err)
	}
	close(gate)
	time.Sleep(20 * time.Millisecond)
	data, _ := os.ReadFile(filepath.Join(a.workspace, "cr/io/1"))
	if len(data) != 0 {
		t.Fatal(string(data))
	}
	if !strings.Contains(output(report(t, s, j, 0, 100)), "may continue") {
		t.Fatal("dishonest local kill")
	}
	a.registry.handler = func(common.Part, common.Job) *common.ExecutionResult { return local("missing file", true) }
	failed, _ := s.Create()
	s.Start(failed, common.Part{})
	first := report(t, s, failed, time.Second, 100)
	if !first.IsError {
		t.Fatal(first)
	}
	later := supervise(t, s, "wait_for_job", `{"handle":2}`, common.Limits{MaxBytes: 100})
	if later.IsError || !strings.Contains(output(later), "execution failed") {
		t.Fatal(later)
	}
}
func TestPendingLimitsConsumedAndExplicitPrecedence(t *testing.T) {
	s, _ := harness(t)
	delay := 200 * time.Millisecond
	budget := 4
	s.SetLimits(common.LimitOverrides{Delay: &delay, MaxBytes: &budget})
	limits, consumed := s.Resolve(common.LimitOverrides{})
	if !consumed || limits.Delay != delay || limits.MaxBytes != budget {
		t.Fatal(limits, consumed)
	}
	defaults, consumed := s.Resolve(common.LimitOverrides{})
	if consumed || defaults.Delay != 3*time.Second || defaults.MaxBytes != 16384 {
		t.Fatal(defaults, consumed)
	}
	s.SetLimits(common.LimitOverrides{Delay: &delay})
	explicit := 10 * time.Second
	limits, consumed = s.Resolve(common.LimitOverrides{Delay: &explicit})
	if !consumed || limits.Delay != explicit {
		t.Fatal(limits, consumed)
	}
}

func TestSourceNoticeSurvivesInitialRunningReport(t *testing.T) {
	s, a := harness(t)
	gate := make(chan struct{})
	a.registry.handler = func(common.Part, common.Job) *common.ExecutionResult {
		<-gate
		return &common.ExecutionResult{Text: "ABCD", Note: "\n[truncated: max_bytes limit reached]\n"}
	}
	j, err := s.Create()
	if err != nil {
		t.Fatal(err)
	}
	s.Start(j, common.Part{})
	first := report(t, s, j, 0, 100)
	if first.Job.Status != "running" || strings.Contains(output(first), "max_bytes") {
		t.Fatal(first)
	}
	close(gate)
	later := supervise(t, s, "wait_for_job", `{"handle":1}`, common.Limits{Delay: time.Second, MaxBytes: 2})
	if got := output(later); !strings.Contains(got, "2 bytes omitted; total bytes: 4") || !strings.Contains(got, "max_bytes limit reached") || !strings.Contains(got, "\nA\n") || !strings.Contains(got, "\nD\n") {
		t.Fatal(got)
	}
	data, err := os.ReadFile(filepath.Join(a.workspace, "cr/io/1"))
	if err != nil || string(data) != "ABCD" || j.Snapshot().Bytes != 4 {
		t.Fatal(string(data), j.Snapshot(), err)
	}
}
