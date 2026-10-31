package ensemble_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ensemble/ensemble"
	"ensemble/internal/common"
)

// TestCompactionBackgroundProgress holds only the memory HTTP exchange and proves
// the real owner completes its main turn before the worker is allowed to reply.
func TestCompactionBackgroundProgress(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	var main atomic.Int32
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "JSON", 400)
			return
		}
		tools, _ := request["tools"].([]any)
		managed := false
		if len(tools) == 1 {
			tool := tools[0].(map[string]any)
			schema, _ := tool["input_schema"].(map[string]any)
			properties, _ := schema["properties"].(map[string]any)
			managed = tool["name"] == "submit" && properties["memory"] != nil
		}
		if managed {
			once.Do(func() { close(entered) })
			select {
			case <-release:
			case <-r.Context().Done():
				return
			}
			compactionProviderReply(w, "anthropic", false, false, `{"memory":"I retained the decision and its reason."}`)
			return
		}
		var content []any
		if main.Add(1) == 1 {
			content = []any{map[string]any{"type": "tool_use", "id": "checkpoint", "name": "micro_handoff", "input": map[string]string{"text": "I finished the segment and retained its reason."}}}
		} else {
			content = []any{map[string]string{"type": "text", "text": "main continued"}}
		}
		json.NewEncoder(w).Encode(map[string]any{"content": content, "usage": map[string]int{"input_tokens": 100, "output_tokens": 5}})
	}))
	t.Cleanup(peer.Close)
	t.Cleanup(func() {
		select {
		case <-release:
		default:
			close(release)
		}
	})
	app := ensemble.New(io.Discard)
	a, err := app.NewAgent(ensemble.Config{Model: "claude-sonnet-5", BaseURL: peer.URL, DataDir: t.TempDir(), BuiltinTools: true, DisableStreaming: true, Memory: ensemble.MemorySettings{Conversation: ensemble.BandSettings{Budget: 40}}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { a.Shutdown() })
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	done := a.Submit(ctx, strings.Repeat("SOURCE-REASON: completed work. ", 8))
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("memory worker did not start")
	}
	select {
	case result := <-done:
		if result.Err != nil || result.Text != "main continued" {
			t.Fatalf("main result: %+v", result)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("held memory request blocked the main turn")
	}
	if len(a.Fleet().Children()) != 1 {
		t.Fatal("main finish discarded or concealed worker")
	}
	close(release)
	deadline := time.Now().Add(5 * time.Second)
	for {
		outcome := a.History().Context().CompactionOutcome
		if outcome != nil {
			if outcome.State != "committed" {
				t.Fatalf("worker outcome: %+v", outcome)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("background result was not committed while owner idle")
		}
		time.Sleep(time.Millisecond)
	}
	if len(a.Fleet().Children()) != 0 {
		t.Fatal("committed worker retained")
	}
}

// backgroundOwner uses a concurrent provider boundary: holding a memory request
// never serializes a main request behind the fake's bookkeeping lock.
func backgroundOwner(t *testing.T, holdMain bool, usage ensemble.Usage, memory string, truncated bool) (ensemble.Agent, <-chan struct{}, chan struct{}, <-chan struct{}, chan struct{}) {
	t.Helper()
	workerEntered, workerRelease := make(chan struct{}), make(chan struct{})
	mainEntered, mainRelease := make(chan struct{}), make(chan struct{})
	var main atomic.Int32
	var first sync.Once
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "JSON", 400)
			return
		}
		tools, _ := request["tools"].([]any)
		managed := false
		if len(tools) == 1 {
			tool := tools[0].(map[string]any)
			schema, _ := tool["input_schema"].(map[string]any)
			properties, _ := schema["properties"].(map[string]any)
			managed = tool["name"] == "submit" && properties["memory"] != nil
		}
		if managed {
			first.Do(func() { close(workerEntered) })
			select {
			case <-workerRelease:
			case <-r.Context().Done():
				return
			}
			payload, _ := json.Marshal(map[string]string{"memory": memory})
			compactionProviderReply(w, "anthropic", false, truncated, string(payload))
			return
		}
		n := main.Add(1)
		if n == 2 {
			close(mainEntered)
			if holdMain {
				select {
				case <-mainRelease:
				case <-r.Context().Done():
					return
				}
			}
		}
		content := []any{map[string]string{"type": "text", "text": "main continued"}}
		if n == 1 {
			content = []any{map[string]any{"type": "tool_use", "id": "checkpoint", "name": "micro_handoff", "input": map[string]string{"text": "I completed the source segment."}}}
		}
		json.NewEncoder(w).Encode(map[string]any{"content": content, "usage": map[string]int{"input_tokens": usage.Input, "cache_read_input_tokens": usage.CacheRead, "cache_creation_input_tokens": usage.CacheWrite, "output_tokens": 5}})
	}))
	t.Cleanup(peer.Close)
	t.Cleanup(func() {
		for _, ch := range []chan struct{}{workerRelease, mainRelease} {
			select {
			case <-ch:
			default:
				close(ch)
			}
		}
	})
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "claude-sonnet-5", BaseURL: peer.URL, DataDir: t.TempDir(), BuiltinTools: true, DisableStreaming: true, Memory: ensemble.MemorySettings{Conversation: ensemble.BandSettings{Budget: 40}}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { a.Shutdown() })
	return a, workerEntered, workerRelease, mainEntered, mainRelease
}

func backgroundOutcome(t *testing.T, a ensemble.Agent) *common.CompactionOutcomeData {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if outcome := a.History().Context().CompactionOutcome; outcome != nil {
			return outcome
		}
		if time.Now().After(deadline) {
			t.Fatal("no terminal compaction outcome")
		}
		time.Sleep(time.Millisecond)
	}
}

// TestCompactionOwnerLifetimes distinguishes turn interruption from worker kill,
// owner shutdown, policy invalidation and failed measured worker submissions.
func TestCompactionOwnerLifetimes(t *testing.T) {
	for _, mode := range []string{"interrupt", "kill", "shutdown", "disable-reenable", "empty", "length"} {
		t.Run(mode, func(t *testing.T) {
			memory := "I kept the source reason."
			if mode == "empty" {
				memory = "  "
			}
			a, entered, release, mainEntered, _ := backgroundOwner(t, mode == "interrupt", ensemble.Usage{Input: 100}, memory, mode == "length")
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			done := a.Submit(ctx, strings.Repeat("RETAIN-SOURCE-REASON ", 10))
			select {
			case <-entered:
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			children := a.Fleet().Children()
			if len(children) != 1 || !children[0].Spec().Compaction {
				t.Fatal("actor did not launch one managed worker")
			}
			child := children[0]
			want := "failed"
			if mode == "interrupt" {
				select {
				case <-mainEntered:
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				}
				if err := a.Post(ensemble.Interrupt{}); err != nil {
					t.Fatal(err)
				}
				result := <-done
				if !errors.Is(result.Err, ensemble.ErrInterrupted) {
					t.Fatalf("main interrupt: %+v", result)
				}
				if len(a.Fleet().Children()) != 1 {
					t.Fatal("ordinary interrupt killed worker")
				}
				if result, err := a.Ask(ctx, "NEW-UNRELATED-DIALOGUE"); err != nil || result != "main continued" {
					t.Fatal(result, err)
				}
				want = "committed"
			} else {
				select {
				case result := <-done:
					if result.Err != nil {
						t.Fatal(result.Err)
					}
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				}
			}
			switch mode {
			case "kill":
				if err := a.Fleet().ShutdownChild(child.Spec().ID); err != nil {
					t.Fatal(err)
				}
				want = "canceled"
			case "shutdown":
				if err := a.Shutdown(); err != nil {
					t.Fatal(err)
				}
				want = "canceled"
			case "disable-reenable":
				settings := a.Config().Memory
				settings.Conversation.Disabled = true
				if err := a.SetMemorySettings(settings); err != nil {
					t.Fatal(err)
				}
				settings.Conversation.Disabled = false
				if err := a.SetMemorySettings(settings); err != nil {
					t.Fatal(err)
				}
				want = "invalidated"
			}
			close(release)
			outcome := backgroundOutcome(t, a)
			if outcome.State != want {
				t.Fatalf("mode=%s outcome=%+v", mode, outcome)
			}
			if len(a.Fleet().Children()) != 0 || a.Ensemble().(*ensemble.Ensemble).ContainsAgent(child) {
				t.Fatal("terminal worker retained")
			}
			contextText := wireText(map[string]any{"context": a.History().Context()})
			if want != "committed" && !strings.Contains(contextText, "RETAIN-SOURCE-REASON") {
				t.Fatal("failed work cut source")
			}
			if mode == "interrupt" && !strings.Contains(contextText, "NEW-UNRELATED-DIALOGUE") {
				t.Fatal("commit removed newer dialogue")
			}
			if mode == "empty" || mode == "length" || mode == "interrupt" {
				wantUsage := ensemble.Usage{Input: 280, CacheRead: 20, Output: 9010}
				if a.Engine().Usage() != wantUsage {
					t.Fatalf("measured usage=%+v want %+v", a.Engine().Usage(), wantUsage)
				}
				if a.Engine().Meter().Usage != (ensemble.Usage{Input: 100, Output: 5}) || a.History().Context().LastUsage != (ensemble.Usage{Input: 100, Output: 5}) {
					t.Fatal("worker usage changed latest main occupancy")
				}
				if len(outcome.Usage) != 1 || outcome.Usage[0].Usage != (ensemble.Usage{Input: 80, CacheRead: 20, Output: 9000}) {
					t.Fatalf("outcome lost measured receipt: %+v", outcome.Usage)
				}
			}
		})
	}
}

// TestCompactionWaitsForSafeBoundary imports completed work's usage while a main
// response is in flight, but does not retire source until that response finishes.
func TestCompactionWaitsForSafeBoundary(t *testing.T) {
	a, entered, release, mainEntered, mainRelease := backgroundOwner(t, true, ensemble.Usage{Input: 100}, "I retained the reason.", false)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	done := a.Submit(ctx, strings.Repeat("SOURCE-BEFORE-CUT ", 10))
	for _, ch := range []<-chan struct{}{entered, mainEntered} {
		select {
		case <-ch:
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	close(release)
	deadline := time.Now().Add(5 * time.Second)
	for a.Engine().Usage().Output < 9000 {
		if time.Now().After(deadline) {
			t.Fatal("worker completion not accounted while main in flight")
		}
		time.Sleep(time.Millisecond)
	}
	c := a.History().Context()
	if c.CompactionOutcome != nil || c.Compression == nil || c.Compression.Committed != nil || !strings.Contains(wireText(map[string]any{"context": c}), "SOURCE-BEFORE-CUT") {
		t.Fatal("worker committed during sent main response")
	}
	close(mainRelease)
	if result := <-done; result.Err != nil {
		t.Fatal(result.Err)
	}
	if outcome := backgroundOutcome(t, a); outcome.State != "committed" {
		t.Fatalf("safe completion: %+v", outcome)
	}
}

// TestCompactionContextLimit holds the next main request only at the known
// window threshold, counts cache buckets once, and keeps the wait interruptible.
func TestCompactionContextLimit(t *testing.T) {
	for _, mode := range []string{"commit", "failure", "interrupt", "below"} {
		t.Run(mode, func(t *testing.T) {
			usage := ensemble.Usage{Input: 100, CacheRead: 3800}
			memory := "I kept the source reason."
			if mode == "failure" {
				memory = " "
			}
			if mode == "below" {
				usage = ensemble.Usage{Input: 100, CacheRead: 1800, CacheWrite: 1000}
			}
			a, entered, release, mainEntered, _ := backgroundOwner(t, false, usage, memory, false)
			// This existing fixture model has a known 4096-token context window.
			if err := a.SetModel("claude-ch16-course"); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			done := a.Submit(ctx, strings.Repeat("THRESHOLD-SOURCE ", 10))
			select {
			case <-entered:
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			if mode == "below" {
				select {
				case <-mainEntered:
				case <-ctx.Done():
					t.Fatal("cache was double-counted")
				}
				if result := <-done; result.Err != nil {
					t.Fatal(result.Err)
				}
			} else {
				for a.Activity().Phase != common.ActivityWaiting {
					select {
					case <-mainEntered:
						t.Fatal("sent main request before headroom")
					case <-ctx.Done():
						t.Fatal(ctx.Err())
					default:
						time.Sleep(time.Millisecond)
					}
				}
				select {
				case <-mainEntered:
					t.Fatal("headroom wait sent request")
				default:
				}
			}
			if mode == "interrupt" {
				if err := a.Post(ensemble.Interrupt{}); err != nil {
					t.Fatal(err)
				}
				if result := <-done; !errors.Is(result.Err, ensemble.ErrInterrupted) {
					t.Fatalf("wait interrupt: %+v", result)
				}
				if len(a.Fleet().Children()) != 1 {
					t.Fatal("headroom interrupt killed worker")
				}
			}
			close(release)
			if mode == "commit" || mode == "failure" {
				result := <-done
				if mode == "commit" && (result.Err != nil || result.Text != "main continued") {
					t.Fatalf("headroom commit: %+v", result)
				}
				if mode == "failure" {
					if result.Err == nil || !strings.Contains(result.Err.Error(), "compaction") {
						t.Fatalf("failed headroom wait: %+v", result)
					}
					select {
					case <-mainEntered:
						t.Fatal("failed compaction resent full context")
					default:
					}
				}
			}
			outcome := backgroundOutcome(t, a)
			want := "committed"
			if mode == "failure" {
				want = "failed"
			}
			if outcome.State != want {
				t.Fatalf("headroom outcome: %+v", outcome)
			}
		})
	}
}

// TestCompactionActorPersistence uses real file failures after a worker starts,
// proving only a saved valid candidate can cut source and that a committed disk
// failure holds new render/re-enable until repair. Reload never imports old usage.
func TestCompactionActorPersistence(t *testing.T) {
	for _, mode := range []string{"worker-save", "worker-log", "stage", "destination-edit", "finish"} {
		t.Run(mode, func(t *testing.T) {
			a, entered, release, _, _ := backgroundOwner(t, false, ensemble.Usage{Input: 100}, "EXACT REPLACEMENT REASON", false)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			done := a.Submit(ctx, strings.Repeat("PERSIST-SOURCE ", 12))
			select {
			case <-entered:
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			if result := <-done; result.Err != nil {
				t.Fatal(result.Err)
			}
			child := a.Fleet().Children()[0]
			var repair func()
			switch mode {
			case "worker-save", "worker-log":
				path := child.Config().SavePath
				if mode == "worker-log" {
					path = child.Config().LogPath
				}
				if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
					t.Fatal(err)
				}
				if err := os.Mkdir(path, 0700); err != nil {
					t.Fatal(err)
				}
				repair = func() { os.Remove(path) }
			case "stage":
				path := filepath.Join(a.Config().DataDir, ".compaction")
				if err := os.WriteFile(path, []byte("blocked staging directory"), 0600); err != nil {
					t.Fatal(err)
				}
				repair = func() { os.Remove(path) }
			case "destination-edit", "finish":
				dir := filepath.Join(a.Config().DataDir, "memory")
				if err := os.MkdirAll(dir, 0700); err != nil {
					t.Fatal(err)
				}
				if mode == "finish" {
					if err := os.Chmod(dir, 0500); err != nil {
						t.Fatal(err)
					}
					repair = func() { os.Chmod(dir, 0700) }
				} else {
					task, err := a.Memory().Conversation()
					if err != nil || task == nil {
						t.Fatalf("selection: %+v %v", task, err)
					}
					path := filepath.Join(a.Config().DataDir, task.Selection.Destination.Path)
					if err := os.WriteFile(path, []byte("HUMAN CORRECTION"), 0600); err != nil {
						t.Fatal(err)
					}
					repair = func() {
						data, err := os.ReadFile(path)
						if err != nil || string(data) != "HUMAN CORRECTION" {
							t.Errorf("human bytes lost: %q %v", data, err)
						}
					}
				}
			}
			defer repair()
			close(release)
			if mode == "finish" {
				for {
					c := a.History().Context()
					if c.Compression != nil && c.Compression.Committed != nil {
						break
					}
					if c.CompactionOutcome != nil {
						t.Fatalf("expected disk hold, got %+v", c.CompactionOutcome)
					}
					select {
					case <-ctx.Done():
						t.Fatal(ctx.Err())
					default:
						time.Sleep(time.Millisecond)
					}
				}
				if _, err := a.Ask(ctx, "must remain blocked"); err == nil {
					t.Fatal("unfinalized corpus sent main request")
				}
				if err := a.SetMemorySettings(a.Config().Memory); err == nil {
					t.Fatal("settings bypassed recovery hold")
				}
				sent := 0
				for _, event := range a.History().EventsAfter(0) {
					if event.Type == "request_sent" {
						sent++
					}
				}
				if sent != 2 {
					t.Fatalf("blocked turn sent request: %d", sent)
				}
				if c := a.History().Context(); c.Compression == nil || c.Compression.Committed == nil || c.CompactionOutcome != nil {
					t.Fatal("failed finalization abandoned committed receipt")
				}
				repair()
				if text, err := a.Ask(ctx, "continue after disk repair"); err != nil || text != "main continued" {
					t.Fatal(text, err)
				}
			}
			outcome := backgroundOutcome(t, a)
			want := "failed"
			if mode == "finish" {
				want = "committed"
			}
			if mode == "destination-edit" {
				want = "invalidated"
			}
			if outcome.State != want {
				t.Fatalf("disk outcome=%+v want %s", outcome, want)
			}
			if mode != "finish" && !strings.Contains(wireText(map[string]any{"context": a.History().Context()}), "PERSIST-SOURCE") {
				t.Fatal("failed persistence cut source")
			}
			if len(outcome.Usage) != 1 || outcome.Usage[0].Usage.Output != 9000 {
				t.Fatalf("failed work lost measured receipt: %+v", outcome.Usage)
			}
			if a.Engine().Usage().Output < 9000 {
				t.Fatal("failed work not accounted")
			}
			if mode == "stage" {
				repair()
				if err := a.SetMemorySettings(a.Config().Memory); err != nil {
					t.Fatal(err)
				}
			}
			if err := a.Shutdown(); err != nil {
				t.Fatal(err)
			}
			if err := a.Save(); err != nil {
				t.Fatal(err)
			}
			b, err := ensemble.New(io.Discard).NewAgent(a.Config())
			if err != nil {
				t.Fatal(err)
			}
			defer b.Shutdown()
			if err := b.Load(); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(b.History().Context().CompactionOutcome, outcome) {
				t.Fatal("reload lost terminal result/usage receipt")
			}
			if b.Engine().Usage() != (ensemble.Usage{}) {
				t.Fatal("reload billed historical work again")
			}
		})
	}
}
