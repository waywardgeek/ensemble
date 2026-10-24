package ensemble_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"ensemble"
)

func TestChapter5ShutdownInitiatesAllOwnedStops(t *testing.T) {
	app := ensemble.New(nil)
	release := make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer func() {
		unblock()
		if err := app.Shutdown(); err != nil {
			t.Error(err)
		}
	}()
	dir := t.TempDir()
	makeAgent := func(name string) ensemble.Agent {
		t.Helper()
		a, err := app.NewAgent(ensemble.Config{Model: "offline", APIKey: "dummy", BaseURL: "http://127.0.0.1:1", BaseDir: dir, DataDir: filepath.Join(dir, name), LogPath: filepath.Join(dir, name+".jsonl")})
		if err != nil {
			t.Fatal(err)
		}
		return a
	}
	first, second := makeAgent("one"), makeAgent("two")
	started := make(chan struct{})
	if err := first.Tools().Register(ensemble.ToolDefinition{Name: "held", Description: "Held Go execution.", Parameters: json.RawMessage(`{"type":"object"}`)}, func(_ context.Context, c ensemble.ToolCall, _ json.RawMessage) error {
		close(started)
		<-release
		_, _ = io.WriteString(c, "real final output")
		return errors.New("real handler outcome")
	}); err != nil {
		t.Fatal(err)
	}
	limit := first.Tools().Prepare(context.Background(), ensemble.Part{Type: "tool_call", CallID: "limit", Name: "tool_limits", Args: json.RawMessage(`{"ai_callback_delay":0}`)}, "", 0).Await(context.Background())
	if limit.IsError {
		t.Fatalf("setting observation limit: %+v", limit)
	}
	held := first.Tools().Prepare(context.Background(), ensemble.Part{Type: "tool_call", CallID: "held", Name: "held", Args: json.RawMessage(`{}`)}, "", 0).Await(context.Background())
	if held.IsError {
		t.Fatalf("held dispatch failed: %+v", held)
	}
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("handler not started")
	}
	var processes []ensemble.Job
	var pids []int
	for i, a := range []ensemble.Agent{first, second} {
		result := a.Tools().Prepare(context.Background(), ensemble.Part{Type: "tool_call", CallID: fmt.Sprintf("process-%d", i), Name: "run_command", Args: json.RawMessage(`{"command":"echo $$; exec sleep 60","ai_callback_delay":0}`)}, "", 0).Await(context.Background())
		if result.IsError || result.Job == nil {
			t.Fatalf("process start: %+v", result)
		}
		j, err := a.Jobs().Get(result.Job.Handle)
		if err != nil {
			t.Fatal("missing process job")
		}
		processes = append(processes, j)
		var pid int
		ch5Eventually(t, "real process PID", func() bool {
			d := j.Data()
			b, err := os.ReadFile(filepath.Join(dir, d.Output.Locator))
			if filepath.IsAbs(d.Output.Locator) {
				b, err = os.ReadFile(d.Output.Locator)
			}
			if err != nil {
				return false
			}
			fields := strings.Fields(string(b))
			if len(fields) == 0 {
				return false
			}
			pid, err = strconv.Atoi(fields[0])
			return err == nil
		})
		pids = append(pids, pid)
	}
	done := make(chan error, 1)
	go func() { done <- app.Shutdown() }()
	ch5Eventually(t, "all managed processes reaped while Go work remains held", func() bool {
		for i, j := range processes {
			if j.Data().Status == ensemble.JobRunning || !errors.Is(syscall.Kill(pids[i], 0), syscall.ESRCH) {
				return false
			}
		}
		return true
	})
	select {
	case err := <-done:
		t.Fatalf("shutdown claimed held work completed: %v", err)
	default:
	}
	heldJob, _ := first.Jobs().Get(held.Job.Handle)
	if heldJob.Data().Status != ensemble.JobRunning {
		t.Fatalf("held Go result was fabricated: %+v", heldJob.Data())
	}
	terminal := []ensemble.JobData{processes[0].Data(), processes[1].Data()}
	unblock()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("shutdown did not join released handler")
	}

	d := heldJob.Data()
	if d.Status != ensemble.JobDone {
		t.Fatalf("lost real Go outcome: %+v", d)
	}
	output := d.Output.Locator
	if !filepath.IsAbs(output) {
		output = filepath.Join(dir, output)
	}
	raw, err := os.ReadFile(output)
	if err != nil || string(raw) != "real final output\nreal handler outcome" {
		t.Fatalf("lost retained handler outcome: %q %v", raw, err)
	}
	for i, j := range processes {
		if after := j.Data(); !reflect.DeepEqual(after, terminal[i]) {
			t.Fatalf("terminal metadata overwritten: before=%+v after=%+v", terminal[i], after)
		}
	}
	saved, err := os.ReadFile(filepath.Join(dir, "one.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	recorded := false
	for _, line := range strings.Split(string(saved), "\n") {
		var e ensemble.Event
		if json.Unmarshal([]byte(line), &e) == nil && e.Type == ensemble.JobReported && e.Job != nil && e.Job.CallID == "held" && e.Job.Status == ensemble.JobDone && e.Job.Bytes == int64(len(raw)) {
			recorded = true
		}
	}
	if !recorded {
		t.Fatalf("shutdown did not save the real final job fact: %s", saved)
	}

}

func TestChapter5HintedBatchDoesNotSpendRound(t *testing.T) {
	for _, extra := range []bool{false, true} {
		t.Run(fmt.Sprintf("extra_batch_%v", extra), func(t *testing.T) {
			entered, release := make(chan struct{}), make(chan struct{})
			var once sync.Once
			unblock := func() { once.Do(func() { close(release) }) }
			defer unblock()
			var requests, effects atomic.Int32
			bodies := make(chan string, 4)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, _ := io.ReadAll(r.Body)
				n := requests.Add(1)
				bodies <- string(b)
				switch n {
				case 1:
					close(entered)
					<-release
					ch5Reply(w, ch5Call("skipped-effect", "effect", map[string]any{}), ch5Call("skipped-setter", "tool_limits", map[string]any{"max_output_bytes": 1}))
				case 2:
					ch5Reply(w, ch5Call("actual-first", "effect", map[string]any{}))
				default:
					if extra {
						ch5Reply(w, ch5Call("over-limit", "effect", map[string]any{}))
					} else {
						ch5Reply(w, map[string]string{"type": "text", "text": "first dispatched batch consumed"})
					}
				}
			}))
			defer func() { unblock(); server.Close() }()
			app := ensemble.New(nil)
			defer func() { unblock(); app.Shutdown() }()
			dir := t.TempDir()
			a, err := app.NewAgent(ensemble.Config{Model: "offline", APIKey: "dummy", BaseURL: server.URL, BaseDir: dir, DataDir: filepath.Join(dir, "jobs"), MaxToolRounds: 1})
			if err != nil {
				t.Fatal(err)
			}
			if err = a.Tools().Register(ensemble.ToolDefinition{Name: "effect", Description: "Count a real execution.", Parameters: json.RawMessage(`{"type":"object"}`)}, func(_ context.Context, c ensemble.ToolCall, _ json.RawMessage) error {
				effects.Add(1)
				_, err := io.WriteString(c, "real-output-not-truncated")
				return err
			}); err != nil {
				t.Fatal(err)
			}
			reply := a.Submit(context.Background(), "one dispatched batch allowed")
			select {
			case <-entered:
			case <-time.After(3 * time.Second):
				t.Fatal("no initial HTTP")
			}
			if err = a.Post(ensemble.Hint{Text: "revise before dispatch"}); err != nil {
				t.Fatal(err)
			}
			ch5Eventually(t, "hint admitted during initial HTTP", func() bool {
				for _, e := range a.Log().Events() {
					if e.Type == ensemble.HintReceived {
						return true
					}
				}
				return false
			})
			unblock()
			select {
			case r := <-reply:
				if extra {
					if !errors.Is(r.Err, ensemble.ErrToolRoundLimit) {
						t.Fatalf("missing paired round refusal: %+v", r)
					}
				} else if r.Err != nil || r.Text != "first dispatched batch consumed" {
					t.Fatalf("zero-dispatch batch spent allowance: %+v", r)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("turn did not finish")
			}
			if effects.Load() != 1 || len(app.JobTable()) != 1 || requests.Load() != 3 {
				t.Fatalf("effects=%d jobs=%d requests=%d", effects.Load(), len(app.JobTable()), requests.Load())
			}
			returned := map[string]ensemble.ToolData{}
			for _, e := range a.Log().Events() {
				if e.Type == ensemble.ToolReturned {
					returned[e.Tool.CallID] = *e.Tool
				}
			}
			for _, id := range []string{"skipped-effect", "skipped-setter"} {
				d, ok := returned[id]
				if !ok || !d.IsError || d.Job != nil || !strings.Contains(ch5ToolText(d), "interrupted, not executed") {
					t.Fatalf("bad skipped pair %s: %+v", id, d)
				}
			}
			if d := returned["actual-first"]; d.IsError || ch5ToolText(d) != "real-output-not-truncated" {
				t.Fatalf("stale skipped setter affected actual call: %+v", d)
			}
			if extra {
				d, ok := returned["over-limit"]
				if !ok || !d.IsError || d.Job != nil {
					t.Fatalf("unpaired limit refusal: %+v", d)
				}
			}
		})
	}
}

func TestChapter5LiveAdmissionAndQueuedCancellation(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		n := requests.Add(1)
		if strings.Contains(string(b), "cancelled-separate-turn") {
			t.Errorf("cancelled queue entry reached HTTP: %s", b)
		}
		if n == 1 {
			close(entered)
			<-release
		}
		if n == 2 && !strings.Contains(string(b), "active-note") {
			t.Errorf("joined hint missing: %s", b)
		}
		ch5Reply(w, map[string]string{"type": "text", "text": fmt.Sprintf("answer-%d", n)})
	}))
	defer func() { unblock(); server.Close() }()
	app := ensemble.New(nil)
	defer func() { unblock(); app.Shutdown() }()
	dir := t.TempDir()
	a, err := app.NewAgent(ensemble.Config{Model: "offline", APIKey: "dummy", BaseURL: server.URL, BaseDir: dir, DataDir: filepath.Join(dir, "jobs")})
	if err != nil {
		t.Fatal(err)
	}
	if a.Config().MaxToolRounds != 200 {
		t.Fatalf("zero default=%d", a.Config().MaxToolRounds)
	}
	if err = a.Post(ensemble.Hint{Text: "idle-hint-starts-request"}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("idle hint was stranded")
	}
	queuedCtx, cancel := context.WithCancel(context.Background())
	queued := a.Submit(queuedCtx, "cancelled-separate-turn")
	cancel()
	select {
	case r := <-queued:
		if !errors.Is(r.Err, context.Canceled) {
			t.Fatalf("queued cancel: %+v", r)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("queued cancellation was not reliable")
	}
	joined := a.SubmitLive(context.Background(), "active-note")
	ch5Eventually(t, "joined live admission", func() bool {
		for _, e := range a.Log().Events() {
			if e.Type == ensemble.HintReceived && e.Message.Parts[0].Text == "active-note" {
				return true
			}
		}
		return false
	})
	unblock()
	select {
	case r := <-joined:
		if r.Err != nil || r.Text != "answer-2" || r.TurnID != 1 {
			t.Fatalf("joined completion: %+v", r)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("lost joined completion")
	}
	select {
	case r := <-a.SubmitLive(context.Background(), "idle-second"):
		if r.Err != nil || r.Text != "answer-3" || r.TurnID != 2 {
			t.Fatalf("idle live completion: %+v", r)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("lost idle completion")
	}
}

func ch5ToolText(d ensemble.ToolData) string {
	var b strings.Builder
	for _, p := range d.Parts {
		if p.Type == "text" {
			b.WriteString(p.Text)
		}
	}
	return b.String()
}
