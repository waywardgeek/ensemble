package ensemble_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"ensemble"
)

func TestChapter5ShutdownJoinsKilledFactDelivery(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "conversation.jsonl")
	app := ensemble.New(nil)
	a, err := app.NewAgent(ensemble.Config{Model: "offline", BaseDir: dir, DataDir: "cr", LogPath: logPath})
	if err != nil {
		t.Fatal(err)
	}
	releaseReport := make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(releaseReport) }) }
	killDone := make(chan struct{})
	killStarted := false
	var killErr error
	t.Cleanup(func() {
		release()
		if err := app.Shutdown(); err != nil {
			t.Error(err)
		}
		if killStarted {
			select {
			case <-killDone:
			case <-time.After(3 * time.Second):
				t.Error("external Kill did not finish during cleanup")
			}
		}
	})
	processReady, reaped, reportHeld := make(chan struct{}), make(chan struct{}), make(chan struct{})
	cmd := exec.Command("sleep", "60")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	c4Register(t, a, "managed", func(_ context.Context, call ensemble.ToolCall, _ json.RawMessage) error {
		if err := call.StartProcess(func() (io.Writer, func() error, error) {
			if err := cmd.Start(); err != nil {
				return nil, nil, err
			}
			stop := func() error {
				err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
				<-reaped
				// An ordinary application-provided stop callback is a controlled
				// boundary after real reap but before Kill can deliver its fact.
				close(reportHeld)
				<-releaseReport
				return err
			}
			return nil, stop, nil
		}); err != nil {
			return err
		}
		close(processReady)
		err := cmd.Wait()
		close(reaped)
		return err
	})
	shutdownEntered, watchReady := make(chan struct{}), make(chan struct{})
	c4Register(t, a, "shutdown_watch", func(ctx context.Context, call ensemble.ToolCall, _ json.RawMessage) error {
		close(watchReady)
		<-ctx.Done()
		close(shutdownEntered)
		if _, err := io.WriteString(call, "ordinary final output"); err != nil {
			return err
		}
		return errors.New("ordinary shutdown outcome")
	})
	start := func(name, id string) ensemble.Job {
		t.Helper()
		c4Dispatch(t, a, "tool_limits", `{"ai_callback_delay":0}`, false)
		d := a.Tools().Prepare(context.Background(), ensemble.Part{Type: "tool_call", CallID: id, Name: name, Args: json.RawMessage(`{}`)}, "", 0).Await(context.Background())
		if d.IsError || d.Job == nil {
			t.Fatalf("starting %s: %+v", name, d)
		}
		j, err := a.Jobs().Get(d.Job.Handle)
		if err != nil {
			t.Fatal(err)
		}
		return j
	}
	job := start("managed", "terminal-race")
	c4Signal(t, processReady)
	watcher := start("shutdown_watch", "shutdown-watch")
	c4Signal(t, watchReady)
	killStarted = true
	go func() { killErr = job.Kill("kill_job"); close(killDone) }()
	c4Signal(t, reportHeld)
	if err := syscall.Kill(cmd.Process.Pid, 0); !errors.Is(err, syscall.ESRCH) {
		t.Fatalf("process was not actually reaped: %v", err)
	}
	stopped := make(chan error, 1)
	go func() { stopped <- app.Shutdown() }()
	c4Signal(t, shutdownEntered)
	countKilled := func(events []ensemble.Event) int {
		count := 0
		for _, e := range events {
			if e.Type == ensemble.JobKilled && e.Job != nil && e.Job.Handle == job.Data().Handle {
				count++
			}
		}
		return count
	}
	saved := func() []ensemble.Event {
		t.Helper()
		raw, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatal(err)
		}
		var events []ensemble.Event
		for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
			if line != "" {
				var e ensemble.Event
				if err := json.Unmarshal([]byte(line), &e); err != nil {
					t.Fatal(err)
				}
				events = append(events, e)
			}
		}
		return events
	}
	// The real process has settled and shutdown is admitted, but its terminal
	// delivery is deliberately held. No final save/completion may overtake it.
	select {
	case err := <-stopped:
		t.Fatalf("shutdown overtook held terminal delivery: err=%v live killed facts=%d saved killed facts=%d", err, countKilled(a.Log().Events()), countKilled(saved()))
	case <-time.After(200 * time.Millisecond):
	}
	if countKilled(a.Log().Events()) != 0 {
		t.Fatal("killed fact arrived before the controlled reporting boundary")
	}
	release()
	c4Signal(t, killDone)
	if killErr != nil {
		t.Fatal(killErr)
	}
	select {
	case err := <-stopped:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("shutdown did not join released terminal delivery")
	}
	for _, events := range [][]ensemble.Event{a.Log().Events(), saved()} {
		if countKilled(events) != 1 {
			t.Fatalf("terminal killed fact count=%d, want exactly one", countKilled(events))
		}
		for _, e := range events {
			if e.Type == ensemble.JobKilled && e.Job != nil && e.Job.Handle == job.Data().Handle {
				if e.Job.Status != ensemble.JobKilledStatus || e.Job.Reason != "kill_job" || e.Job.ExitCode != nil || e.Job.CallID != "terminal-race" || e.Job.Name != "managed" || string(e.Job.Args) != "{}" {
					t.Fatalf("wrong retained kill fact: %+v", e)
				}
			}
		}
	}
	if watcher.Data().Status != ensemble.JobDone {
		t.Fatalf("ordinary shutdown lost real result: %+v", watcher.Data())
	}
	path := watcher.Data().Output.Locator
	if !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != "ordinary final output\nordinary shutdown outcome" {
		t.Fatalf("ordinary final output/error was discarded: %q %v", raw, err)
	}
}
