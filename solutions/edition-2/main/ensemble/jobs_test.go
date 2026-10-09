package ensemble_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"testing/synctest"
	"time"

	"ensemble/ensemble"
	"ensemble/internal/common"
)

// Tests drive the existing HTTP fake and the real tools. The disk log is
// independent of report text, so a plausible wrapper cannot conceal data loss.
func events(t *testing.T, a common.Agent) []common.Event {
	t.Helper()
	var log bytes.Buffer
	if err := a.History().Dump(&log); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(log.String()), "\n")
	var result []common.Event
	for _, line := range lines[1:] {
		var event common.Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatal(err)
		}
		result = append(result, event)
	}
	return result
}
func TestJobFileCapCursorAndRedaction(t *testing.T) {
	path := filepath.Join(t.TempDir(), "large.txt")
	full := "FIRST\n" + strings.Repeat("middle\n", 10000) + "LAST\n"
	if err := os.WriteFile(path, []byte(full), 0600); err != nil {
		t.Fatal(err)
	}
	a, results := toolSession(t,
		toolCall{"tool_limits", map[string]any{"max_output_bytes": 128}},
		toolCall{"read_file", map[string]any{"path": path}},
		toolCall{"wait_for_job", map[string]any{"handle": 1}},
		toolCall{"read_file", map[string]any{"path": path, "start_line": 1, "end_line": 1}},
		toolCall{"read_file", map[string]any{"path": path}},
	)
	report := resultText(results[1])
	if !strings.HasPrefix(report, "[tool_limits consumed by this read_file call:") || !strings.Contains(report, "FIRST") || !strings.Contains(report, "LAST") || !strings.Contains(report, fmt.Sprintf("[%d bytes omitted; %d bytes total", len(full)-128, len(full))) || len(report) > 700 {
		t.Fatalf("cap/notice: %q", report)
	}
	if strings.Contains(resultText(results[2]), "FIRST") || !strings.Contains(resultText(results[2]), "no new output") {
		t.Fatal("finished wait repeated output")
	}
	if resultText(results[3]) != "FIRST\n" {
		t.Fatal("one-shot limits remained or small result changed")
	}
	if got := resultText(results[4]); !strings.Contains(got, fmt.Sprintf("[%d bytes omitted; %d bytes total", len(full)-16384, len(full))) {
		t.Fatal("default cap differs from 16384")
	}
	var returned common.Event
	for _, event := range events(t, a) {
		if event.Tool == nil {
			continue
		}
		if event.Tool.CallID == "b" {
			if event.Tool.Job == nil || event.Tool.Job.Handle != 1 {
				t.Fatal("ordinary tool has no job")
			}
			if event.Type == "tool_called" && (event.Tool.Job.Status != "running" || event.Tool.Job.Bytes != 0) {
				t.Fatal("handle was not recorded before execution")
			}
			if event.Type == "tool_returned" {
				returned = event
			}
		} else if (event.Tool.CallID == "a" || event.Tool.CallID == "c") && event.Tool.Job != nil {
			t.Fatal("supervision became a job")
		}
	}
	data, err := os.ReadFile(returned.Tool.Job.Output.Locator)
	if err != nil || string(data) != full || returned.Tool.Job.Bytes != int64(len(full)) {
		t.Fatal("job disk output lost bytes", err)
	}
	if err := a.History().Append(common.Event{Type: "redacted", Redact: &common.RedactData{From: returned.Seq, To: returned.Seq, Level: "redact_result"}}); err != nil {
		t.Fatal(err)
	}
	for _, entry := range a.History().Context().Dialogue {
		if entry.Seq == returned.Seq && entry.Parts[0].Parts[0].Ref != returned.Tool.Job.Output {
			t.Fatal("redaction lost recovery address")
		}
	}
}
func TestPTYInteractionUnseenPatternAndCwd(t *testing.T) {
	dir := t.TempDir()
	a, results := toolSession(t,
		toolCall{"run_command", map[string]any{"command": "printf 'READY> '; read line; sleep 0.1; printf 'VALUE=%s\\nDONE> ' \"$line\"; read quit", "cwd": dir, "ai_callback_pattern": "READY> ", "ai_callback_delay": 2}},
		toolCall{"wait_for_job", map[string]any{"handle": 1, "ai_callback_pattern": "READY> ", "ai_callback_delay": 0.05}},
		toolCall{"send_input", map[string]any{"handle": 1, "input": "hello", "ai_callback_pattern": "DONE> ", "ai_callback_delay": 2}},
		toolCall{"send_input", map[string]any{"handle": 1, "input": "quit", "ai_callback_delay": 2}},
		toolCall{"run_command", map[string]any{"command": "pwd; printf '%s' \"$TERM\"; stty size"}},
		toolCall{"run_command", map[string]any{"command": "printf should-not-run", "cwd": filepath.Join(dir, "absent")}},
	)
	if !strings.Contains(resultText(results[0]), "READY> ") || !strings.Contains(resultText(results[0]), "cwd "+dir) {
		t.Fatal("prompt/cwd not reported")
	}
	if strings.Contains(resultText(results[1]), "READY>") || !strings.Contains(resultText(results[1]), "no new output") {
		t.Fatal("old prompt repeated")
	}
	if !strings.Contains(resultText(results[2]), "hello\nVALUE=hello\nDONE> ") {
		t.Fatalf("input/echo/pattern/CRLF: %q", resultText(results[2]))
	}
	if !strings.Contains(resultText(results[3]), "exit_code: 0") {
		t.Fatal("PTY exit lost")
	}
	workspace, _ := os.Getwd()
	if !strings.HasPrefix(resultText(results[4]), workspace+"\n") || !strings.Contains(resultText(results[4]), "dumb50 200") {
		t.Fatalf("cwd isolation/terminal: %q", resultText(results[4]))
	}
	if !results[5].IsError || strings.Contains(resultText(results[5]), "should-not-run") {
		t.Fatal("missing cwd did not fail")
	}
	for _, event := range events(t, a) {
		if event.Type == "tool_returned" && event.Tool.CallID == "e" && event.Tool.Job.Cwd != "" {
			t.Fatal("cwd override stuck")
		}
	}
}
func TestKillWakesWaiterAndKillsGroup(t *testing.T) {
	t.Run("already blocked waiter", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "fake", DataDir: t.TempDir()})
			if err != nil {
				t.Fatal(err)
			}
			job, err := a.Jobs().Start()
			if err != nil {
				t.Fatal(err)
			}
			type waitResult struct {
				text string
				err  error
			}
			finished := make(chan waitResult, 1)
			go func() {
				text, err := job.Wait(common.Limits{Delay: 30 * time.Second, MaxBytes: 100})
				finished <- waitResult{text, err}
			}()
			// This is the real Wait, with its real notification channel. synctest
			// establishes durable blocking; sleeping or merely starting a goroutine
			// does not. Keep OS process I/O outside this scheduling-only subtest.
			synctest.Wait()
			select {
			case <-finished:
				t.Fatal("wait returned before kill")
			default:
			}
			if err := job.Kill("kill_job"); err != nil {
				t.Fatal(err)
			}
			synctest.Wait()
			var result waitResult
			select {
			case result = <-finished:
			default:
				t.Error("kill did not wake already-blocked waiter")
				// Let a broken implementation's waiter reach its virtual deadline.
				// Receive its result even on failure: no shared variable is read
				// before a later cleanup write, and the waiter can exit cleanly.
				time.Sleep(30 * time.Second)
				result = <-finished
			}
			if result.err != nil || !strings.Contains(result.text, "killed") {
				t.Fatalf("kill report: %s (%v)", result.text, result.err)
			}
		})
	})
	a, results := toolSession(t, toolCall{"run_command", map[string]any{"command": "sleep 2 & child=$!; printf 'child:%s\\n' \"$child\"; wait", "ai_callback_pattern": "child:[0-9]+", "ai_callback_delay": 2}})
	var pid int
	report := resultText(results[0])
	index := strings.Index(report, "child:")
	if index < 0 {
		t.Fatal("child not started")
	}
	fmt.Sscanf(report[index:], "child:%d", &pid)
	job, err := a.Jobs().Find(1)
	if err != nil {
		t.Fatal(err)
	}
	killStarted := time.Now()
	if err := job.Kill("kill_job"); err != nil {
		t.Fatal(err)
	}
	if time.Since(killStarted) > time.Second {
		t.Fatal("process-group kill left the child running")
	}
	// The subtest above establishes the pre-kill waiting case. Here the real
	// PTY test also checks that an already-killed process stays observable.
	result, err := job.Wait(common.Limits{MaxBytes: 100})
	if err != nil || !strings.Contains(result, "killed") || strings.Contains(result, "done") || strings.Contains(result, "exit_code") {
		t.Fatalf("kill report: %s (%v)", result, err)
	}
	if err := syscall.Kill(pid, 0); err == nil {
		t.Fatal("process-group child survived")
	}
	if data := job.Data(); data.Status != "killed" || data.ExitCode != nil {
		t.Fatal("natural completion overwrote killed state")
	}
	found := false
	for _, event := range events(t, a) {
		if event.Type == "job_killed" {
			found = true
			if event.Job.Reason != "kill_job" || event.Job.Status != "killed" {
				t.Fatal("kill fact incorrect")
			}
		}
	}
	if !found {
		t.Fatal("kill fact absent")
	}
}
func TestLimitsBurnAndExplicitOverride(t *testing.T) {
	_, results := toolSession(t,
		toolCall{"tool_limits", map[string]any{"max_output_bytes": 1}},
		toolCall{"tool_limits", map[string]any{"ai_callback_delay": 0, "max_output_bytes": 1}},
		toolCall{"run_command", map[string]any{"command": "printf complete", "ai_callback_delay": 2, "max_output_bytes": 128}},
		toolCall{"wait_for_job", map[string]any{"handle": 1}},
	)
	if !strings.Contains(resultText(results[1]), "consumed by this tool_limits") {
		t.Fatal("setter did not consume preceding setting")
	}
	if !strings.Contains(resultText(results[2]), "complete\nexit_code: 0") || strings.Contains(resultText(results[2]), "omitted") {
		t.Fatal("explicit arguments did not win")
	}
	if strings.Contains(resultText(results[3]), "consumed") {
		t.Fatal("one-shot setting was reused")
	}
}

func TestWakeLeavesJobRunningAndPatternConsumes(t *testing.T) {
	a, _ := toolSession(t, toolCall{"run_command", map[string]any{"command": "printf 'READY> '; read line; sleep 0.2; printf 'REPLY> '; read quit", "ai_callback_pattern": "READY> ", "ai_callback_delay": 2}})
	job, err := a.Jobs().Find(1)
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	report, err := job.Wait(common.Limits{Delay: 80 * time.Millisecond, Pattern: "READY> ", MaxBytes: 128})
	if err != nil || time.Since(started) < 60*time.Millisecond || !strings.Contains(report, "running") || job.Data().Status != "running" {
		t.Fatalf("old prompt woke again or wake killed work: %q %v elapsed %s", report, err, time.Since(started))
	}
	if err := job.Send("hello"); err != nil {
		t.Fatal(err)
	}
	started = time.Now()
	report, err = job.Wait(common.Limits{Delay: 2 * time.Second, Pattern: "REPLY> ", MaxBytes: 128})
	if err != nil || time.Since(started) > time.Second || !strings.Contains(report, "REPLY> ") {
		t.Fatalf("new prompt did not wake: %q %v", report, err)
	}
	if err := job.Send("quit"); err != nil {
		t.Fatal(err)
	}
	if _, err := job.Wait(common.Limits{Delay: 2 * time.Second, MaxBytes: 128}); err != nil {
		t.Fatal(err)
	}
}

func TestBlockedFileToolCanBeSupervised(t *testing.T) {
	fifo := filepath.Join(t.TempDir(), "pipe")
	if err := syscall.Mkfifo(fifo, 0600); err != nil {
		t.Fatal(err)
	}
	a, results := toolSession(t,
		toolCall{"tool_limits", map[string]any{"ai_callback_delay": 0.02}},
		toolCall{"read_file", map[string]any{"path": fifo}},
		toolCall{"list_directory", map[string]any{"path": filepath.Dir(fifo)}},
		toolCall{"kill_job", map[string]any{"handle": 1}},
	)
	if !strings.Contains(resultText(results[1]), "running") || !strings.Contains(resultText(results[2]), "pipe") || !strings.Contains(resultText(results[3]), "Go cannot stop") {
		t.Fatal("blocked file tool escaped supervision")
	}
	// Unblock the real FIFO read ourselves: marking killed cannot cancel Go I/O.
	// This prevents the regression fixture from leaking the contained goroutine.
	if err := os.WriteFile(fifo, []byte("released"), 0600); err != nil {
		t.Fatal(err)
	}
	job, err := a.Jobs().Find(1)
	if err != nil {
		t.Fatal(err)
	}
	// Releasing a FIFO proves only that its writer returned, not that dispatch
	// has called Finish. Exercise that same real completion operation directly
	// and synchronously before checking state; a late result must be harmless.
	job.Finish("released", nil)
	if job.Data().Status != "killed" {
		t.Fatal("late file result changed killed state")
	}
}

func TestShutdownRecordsKilledReason(t *testing.T) {
	a, _ := toolSession(t, toolCall{"run_command", map[string]any{"command": "sleep 2", "ai_callback_delay": 0.02}})
	if err := a.Shutdown(); err != nil {
		t.Fatal(err)
	}
	job, err := a.Jobs().Find(1)
	if err != nil {
		t.Fatal(err)
	}
	if job.Data().Status != "killed" {
		t.Fatal("shutdown left job running")
	}
	found := false
	for _, event := range events(t, a) {
		if event.Type == "job_killed" && event.Job.Reason == "shutdown" {
			found = true
		}
	}
	if !found {
		t.Fatal("shutdown reason absent")
	}
}
