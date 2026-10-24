package ensemble_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"ensemble"
)

func TestChapter5ExplicitGoKillReturnsBeforeHandler(t *testing.T) {
	app, a, dir := c4Agent(t)
	ready, release, returned := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	// Registered after the app cleanup: release even if the short assertion fails.
	t.Cleanup(unblock)
	c4Register(t, a, "held_kill", func(_ context.Context, call ensemble.ToolCall, _ json.RawMessage) error {
		call.SetExitCode(7)
		if _, err := io.WriteString(call, "before kill"); err != nil {
			return err
		}
		close(ready)
		<-release
		call.SetExitCode(99)
		_, err := io.WriteString(call, "FORBIDDEN LATE OUTPUT")
		close(returned)
		return err
	})
	c4Dispatch(t, a, "tool_limits", `{"ai_callback_delay":0}`, false)
	c4Dispatch(t, a, "held_kill", `{}`, false)
	c4Signal(t, ready)
	job, err := a.Jobs().Get(1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := job.Wait(context.Background(), ensemble.ToolLimits{MaxOutputBytes: 1024}); err != nil {
		t.Fatal(err)
	}
	observed, armed := c4Observe(t, job, "")
	c4Signal(t, armed)
	killed := make(chan error, 1)
	go func() {
		killed <- a.Tools().Invoke(context.Background(), ensemble.Part{Type: "tool_call", CallID: "explicit-go-kill", Name: "kill_job", Args: json.RawMessage(`{"handle":1}`)}, "")
	}()
	select {
	case err := <-killed:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("kill_job waited for the uncooperative handler")
	}
	select {
	case <-returned:
		t.Fatal("handler was released before kill assertions")
	default:
	}
	wake := c4Result(t, observed)
	if wake.Job.Status != ensemble.JobKilledStatus || wake.Job.ExitCode != nil || !strings.Contains(wake.Text, "Go cannot stop") {
		t.Fatalf("incorrect immediate waiter wake: %+v", wake)
	}
	c4Killed(t, a, job, "kill_job")
	recorded := false
	for _, e := range a.Log().Events() {
		if e.Type == ensemble.ToolReturned && e.Tool != nil && e.Tool.CallID == "explicit-go-kill" {
			text := ""
			for _, p := range e.Tool.Parts {
				text += p.Text
			}
			recorded = !e.Tool.IsError && e.Tool.Job == nil && strings.Contains(text, fmt.Sprintf("job %d killed;", job.Data().Handle)) && strings.Contains(text, "Go cannot stop")
		}
	}
	if !recorded {
		t.Fatal("actual Agent/tool report omitted the truthful Go-kill limitation")
	}
	repeated := make(chan error, 1)
	go func() { repeated <- job.Kill("kill_job") }()
	select {
	case err := <-repeated:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("repeat kill waited for the handler")
	}
	before := job.Data()
	disk, err := os.ReadFile(filepath.Join(dir, "cr", "io", "1"))
	if err != nil || string(disk) != "before kill" {
		t.Fatalf("pre-release output: %q %v", disk, err)
	}
	unblock()
	c4Signal(t, returned)
	if err := app.Shutdown(); err != nil {
		t.Fatal(err)
	}
	c4Killed(t, a, job, "kill_job")
	if after := job.Data(); !reflect.DeepEqual(before, after) {
		t.Fatalf("late handler overwrote killed metadata: before=%+v after=%+v", before, after)
	}
	disk, err = os.ReadFile(filepath.Join(dir, "cr", "io", "1"))
	if err != nil || string(disk) != "before kill" {
		t.Fatalf("late handler changed retained output: %q %v", disk, err)
	}
}
