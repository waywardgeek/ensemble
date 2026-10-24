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

	"ensemble"
)

func ch5Eventually(t *testing.T, what string, predicate func() bool) {
	t.Helper()
	deadline := time.NewTimer(4 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for !predicate() {
		select {
		case <-deadline.C:
			t.Fatal("timed out: " + what)
		case <-tick.C:
		}
	}
}
func ch5Body(t *testing.T, ch <-chan string) string {
	t.Helper()
	select {
	case b := <-ch:
		return b
	case <-time.After(4 * time.Second):
		t.Fatal("no resumed provider request")
		return ""
	}
}
func ch5Call(id, name string, args any) map[string]any {
	return map[string]any{"type": "tool_use", "id": id, "name": name, "input": args}
}
func ch5Reply(w http.ResponseWriter, parts ...any) {
	_ = json.NewEncoder(w).Encode(map[string]any{"content": parts, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
}

// These are real Agent/registry/job paths. Only the provider and the handler's
// completion gate are fake; the process case launches a real retained PTY job.
func TestChapter5WakeInitialAndLaterObservation(t *testing.T) {
	for _, kind := range []string{"handler", "command", "retired-timer"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			release := make(chan struct{})
			var releaseOnce sync.Once
			marker := filepath.Join(dir, "release")
			unblock := func() { releaseOnce.Do(func() { close(release); _ = os.WriteFile(marker, []byte("go"), 0600) }) }
			defer unblock()
			var starts, effects, fresh atomic.Int32
			bodies := make(chan string, 8)
			var requests atomic.Int32
			var heldHandle atomic.Int64
			longDelay := float64(3600)
			if kind == "retired-timer" {
				longDelay = .2
			}
			limits := func(id string) any {
				return ch5Call(id, "tool_limits", map[string]any{"ai_callback_delay": longDelay, "max_output_bytes": 4})
			}
			held := ch5Call("held", "held", map[string]any{})
			if kind == "command" {
				held = ch5Call("held", "run_command", map[string]any{"command": "printf abcd; while [ ! -f '" + marker + "' ]; do sleep 0.01; done; printf efghijkl"})
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, _ := io.ReadAll(r.Body)
				n := requests.Add(1)
				if n > 1 {
					bodies <- string(b)
				}
				switch n {
				case 1:
					ch5Reply(w, map[string]any{"type": "text", "text": "I will complete one call, start a held job, then do the remaining work."},
						ch5Call("complete", "complete", map[string]any{}), limits("long-limit"), held,
						ch5Call("skip-effect", "effect", map[string]any{}), ch5Call("skip-setter", "tool_limits", map[string]any{"max_output_bytes": 1}))
				case 2:
					ch5Reply(w, map[string]any{"type": "text", "text": "I will run fresh work and observe the same held job."},
						ch5Call("fresh", "fresh", map[string]any{}), ch5Call("later", "wait_for_job", map[string]any{"handle": heldHandle.Load(), "ai_callback_delay": 3600, "max_output_bytes": 4}))
				default:
					ch5Reply(w, map[string]any{"type": "text", "text": "Turn complete; the held job is still available."})
				}
			}))
			defer server.Close()
			app := ensemble.New(nil)
			defer func() {
				unblock()
				if err := app.Shutdown(); err != nil {
					t.Error(err)
				}
			}()
			a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.Anthropic, Model: "offline-test", APIKey: "dummy", BaseURL: server.URL, BaseDir: dir, DataDir: filepath.Join(dir, "jobs")})
			if err != nil {
				t.Fatal(err)
			}
			register := func(name string, h ensemble.ToolHandler) {
				t.Helper()
				if err := a.Tools().Register(ensemble.ToolDefinition{Name: name, Description: "Focused offline fixture.", Parameters: json.RawMessage(`{"type":"object"}`)}, h); err != nil {
					t.Fatal(err)
				}
			}
			register("held", func(ctx context.Context, c ensemble.ToolCall, _ json.RawMessage) error {
				starts.Add(1)
				_, _ = io.WriteString(c, "abcd")
				select {
				case <-release:
				case <-ctx.Done():
					return ctx.Err()
				}
				_, err := io.WriteString(c, "efghijkl")
				return err
			})
			register("complete", func(_ context.Context, c ensemble.ToolCall, _ json.RawMessage) error {
				_, err := io.WriteString(c, "completed-real")
				return err
			})
			register("effect", func(context.Context, ensemble.ToolCall, json.RawMessage) error { effects.Add(1); return nil })
			register("fresh", func(_ context.Context, c ensemble.ToolCall, _ json.RawMessage) error {
				fresh.Add(1)
				_, err := io.WriteString(c, "fresh-output-unclipped")
				return err
			})
			reply := a.Submit(context.Background(), "exercise the batch")
			var job ensemble.Job
			ch5Eventually(t, "held job output", func() bool {
				for _, j := range app.JobTable() {
					d := j.Data()
					if d.CallID == "held" && d.Bytes >= 4 {
						job = j
						heldHandle.Store(int64(d.Handle))
						return true
					}
				}
				return false
			})
			first := ensemble.Inbound(ensemble.Hint{Text: "initial-live-hint"})
			second := ensemble.Inbound(ensemble.Question{Text: "later-addressed-question", Sender: "human", Recipient: "author", RequestID: "question-7", TurnID: "source-turn-8"})
			if kind == "command" {
				first, second = second, first
			}
			if err := a.Post(first); err != nil {
				t.Fatal(err)
			}
			request := ch5Body(t, bodies)
			for _, fragment := range []string{"completed-real", "held", "skip-effect", "skip-setter", "interrupted, not executed", "abcd"} {
				if !strings.Contains(request, fragment) {
					t.Fatalf("missing %q in resumed request: %s", fragment, request)
				}
			}
			if effects.Load() != 0 {
				t.Fatal("undispatched side effect ran")
			}
			ch5Eventually(t, "later wait admitted", func() bool {
				for _, e := range a.Log().Events() {
					if e.Type == ensemble.ToolCalled && e.Tool.CallID == "later" {
						return true
					}
				}
				return false
			})
			if kind == "retired-timer" {
				select {
				case body := <-bodies:
					t.Fatalf("retired initial timer answered later wait: %s", body)
				case <-time.After(300 * time.Millisecond):
				}
			}
			if err := a.Post(second); err != nil {
				t.Fatal(err)
			}
			request = ch5Body(t, bodies)
			for _, fragment := range []string{"initial-live-hint", "later-addressed-question", "question-7", "source-turn-8", "human", "author", "fresh-output-unclipped"} {
				if !strings.Contains(request, fragment) {
					t.Fatalf("missing %q after second wake: %s", fragment, request)
				}
			}
			select {
			case result := <-reply:
				if result.Err != nil {
					t.Fatal(result.Err)
				}
			case <-time.After(4 * time.Second):
				t.Fatal("private completion did not arrive")
			}
			if job.Data().Status != ensemble.JobRunning || effects.Load() != 0 || fresh.Load() != 1 {
				t.Fatal("wake changed execution, or stale work replayed")
			}
			if kind != "command" && starts.Load() != 1 {
				t.Fatal("handler was executed more than once")
			}
			counts := map[string]int{}
			for _, e := range a.Log().Events() {
				if e.Type == ensemble.ToolReturned {
					counts[e.Tool.CallID]++
				}
			}
			for _, id := range []string{"complete", "long-limit", "held", "skip-effect", "skip-setter", "fresh", "later"} {
				if counts[id] != 1 {
					t.Fatalf("%s has %d correlated returns", id, counts[id])
				}
			}
			before := len(app.JobTable())
			unblock()
			ch5Eventually(t, "job settles", func() bool { return job.Data().Status != ensemble.JobRunning })

			path := job.Data().Output.Locator
			if !filepath.IsAbs(path) {
				path = filepath.Join(dir, path)
			}
			data, err := os.ReadFile(path)
			if err != nil || !strings.HasPrefix(string(data), "abcdefghijkl") {
				t.Fatalf("retained output %q: %v", data, err)
			}
			report, failed := a.Tools().Dispatch(context.Background(), "wait_for_job", json.RawMessage(fmt.Sprintf(`{"handle":%d,"max_output_bytes":4,"ai_callback_delay":0}`, job.Data().Handle)))
			if failed {
				t.Fatalf("finished read: %+v", report)
			}
			text := ""
			for _, p := range report {
				text += p.Text
			}
			if !strings.Contains(text, "\nef\n[output truncated") || !strings.Contains(text, "output truncated") || !strings.HasSuffix(text, string(data[len(data)-2:])) {
				t.Fatalf("bounded head/tail observation: %s", text)
			}
			report, failed = a.Tools().Dispatch(context.Background(), "wait_for_job", json.RawMessage(fmt.Sprintf(`{"handle":%d,"max_output_bytes":4,"ai_callback_delay":0}`, job.Data().Handle)))
			if failed {
				t.Fatalf("second finished read: %+v", report)
			}
			for _, p := range report {
				if p.Type == "text" && strings.Contains(p.Text, "ef") {
					t.Fatalf("omitted bytes were replayed: %+v", report)
				}
			}
			if len(app.JobTable()) != before {
				t.Fatal("job control allocated a new handle")
			}

			d := job.Data()
			if d.TurnID == 0 || d.CallID != "held" || d.Name == "" || len(d.Args) == 0 {
				t.Fatalf("missing originating metadata: %+v", d)
			}
		})
	}
}

func TestChapter5HTTPFreezeAndReliableQueuedReplies(t *testing.T) {
	entered := make(chan string, 4)
	release := make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	var n atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		entered <- string(b)
		i := n.Add(1)
		if i == 1 {
			select {
			case <-release:
			case <-r.Context().Done():
				return
			}
		}
		ch5Reply(w, map[string]any{"type": "text", "text": fmt.Sprintf("reply-%d", i)})
	}))
	defer server.Close()
	app := ensemble.New(nil)
	defer func() { unblock(); _ = app.Shutdown() }()
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.Anthropic, Model: "offline-test", APIKey: "dummy", BaseURL: server.URL, BaseDir: t.TempDir(), Tools: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	first := a.Submit(context.Background(), "first-private-question")
	frozen := ch5Body(t, entered)
	second := a.Submit(context.Background(), "second-private-question")
	if err := a.Post(ensemble.Hint{Text: "frozen-live-hint"}); err != nil {
		t.Fatal(err)
	}
	ch5Eventually(t, "hint admitted", func() bool {
		for _, e := range a.Log().Events() {
			if e.Type == ensemble.HintReceived {
				return true
			}
		}
		return false
	})
	if strings.Contains(frozen, "frozen-live-hint") || strings.Contains(frozen, "second-private-question") {
		t.Fatal("request mutated after freeze")
	}
	unblock()
	continuation := ch5Body(t, entered)
	if !strings.Contains(continuation, "frozen-live-hint") || !strings.Contains(continuation, "reply-1") || strings.Contains(continuation, "second-private-question") {
		t.Fatalf("hint lost or Submit merged: %s", continuation)
	}
	queued := ch5Body(t, entered)
	if !strings.Contains(queued, "second-private-question") {
		t.Fatal("queued Submit lost")
	}
	for i, ch := range []<-chan ensemble.Result{first, second} {
		select {
		case r := <-ch:
			if r.Err != nil || r.Text != fmt.Sprintf("reply-%d", i+2) {
				t.Fatalf("private reply %d: %+v", i, r)
			}
		case <-time.After(4 * time.Second):
			t.Fatal("lost private reply")
		}
	}
}
