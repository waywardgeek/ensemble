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
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ensemble"
)

func TestChapter8CapturedBudgetPrivateFinalRequest(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	var count, effects atomic.Int32
	var mu sync.Mutex
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		mu.Lock()
		requests = append(requests, string(b))
		mu.Unlock()
		n := count.Add(1)
		switch n {
		case 1:
			close(entered)
			<-release
			ch5Reply(w, ch5Call("first", "effect", map[string]any{}))
		case 2:
			ch5Reply(w, map[string]string{"type": "text", "text": "first turn finished"})
		case 3:
			ch5Reply(w, ch5Call("next-one", "effect", map[string]any{}))
		case 4:
			ch5Reply(w, ch5Call("next-two", "effect", map[string]any{}))
		case 5:
			ch5Reply(w, ch5Call("refused-one", "effect", map[string]any{}), ch5Call("refused-two", "effect", map[string]any{}))
		default:
			ch5Reply(w, map[string]string{"type": "text", "text": "ordinary later turn"})
		}
	}))
	defer func() { unblock(); server.Close() }()
	app := ensemble.New(nil)
	defer func() { unblock(); app.Shutdown() }()
	dir := t.TempDir()
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.Anthropic, Model: "offline", APIKey: "dummy", BaseURL: server.URL, BaseDir: dir, DataDir: filepath.Join(dir, "jobs"), LogPath: filepath.Join(dir, "history.jsonl"), MaxToolRounds: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Tools().Register(ensemble.ToolDefinition{Name: "effect", Description: "Count real job execution", Parameters: json.RawMessage(`{"type":"object"}`)}, func(_ context.Context, c ensemble.ToolCall, _ json.RawMessage) error {
		effects.Add(1)
		_, err := io.WriteString(c, "real output")
		return err
	}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	first := a.Submit(ctx, "first turn")
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("first request was not admitted")
	}
	if err = a.SetMaxToolRounds(2); err != nil {
		t.Fatal(err)
	}
	unblock()
	select {
	case result := <-first:
		if result.Err != nil || result.Text != "first turn finished" {
			t.Fatalf("first turn: %+v", result)
		}
	case <-ctx.Done():
		t.Fatal("first turn did not finish")
	}
	if _, err = a.Ask(ctx, "second turn"); !errors.Is(err, ensemble.ErrToolRoundLimit) {
		t.Fatalf("second turn limit: %v", err)
	}
	if effects.Load() != 3 || len(app.JobTable()) != 3 {
		t.Fatalf("effects=%d jobs=%d, want exactly three actual executions", effects.Load(), len(app.JobTable()))
	}
	if err = a.SetMaxToolRounds(0); err != nil {
		t.Fatal(err)
	}
	if text, err := a.Ask(ctx, "later ordinary turn"); err != nil || text != "ordinary later turn" {
		t.Fatalf("later turn: %q, %v", text, err)
	}
	mu.Lock()
	captured := append([]string(nil), requests...)
	mu.Unlock()
	if len(captured) != 6 {
		t.Fatalf("requests=%d, want 6", len(captured))
	}
	// Inspect the provider request, not an internal field or mock Render callback.
	for i, body := range captured {
		lowered := strings.ToLower(body)
		stop := strings.Contains(lowered, "tool-round limit") && strings.Contains(lowered, "do not call")
		want := i == 1 || i == 4
		if stop != want {
			t.Errorf("request %d private stop=%v want %v", i+1, stop, want)
		}
	}
	returned := map[string]ensemble.ToolData{}
	returnCounts, callCounts := map[string]int{}, map[string]int{}
	for _, event := range a.Log().Events() {
		if event.Type == ensemble.ToolReturned {
			returned[event.Tool.CallID] = *event.Tool
			returnCounts[event.Tool.CallID]++
		}
		if event.Type == ensemble.ResponseEnded && event.Response != nil {
			for _, part := range event.Response.Parts {
				if part.Type == "tool_call" {
					callCounts[part.CallID]++
				}
			}
		}
		raw, _ := json.Marshal(event)
		if strings.Contains(strings.ToLower(string(raw)), "do not call") {
			t.Fatal("private stop instruction leaked into History")
		}
	}
	for _, id := range []string{"refused-one", "refused-two"} {
		result, ok := returned[id]
		if !ok || returnCounts[id] != 1 || callCounts[id] != 1 || !result.IsError || result.Dispatched || result.Job != nil {
			t.Fatalf("unpaired or executed refusal %s: %+v", id, result)
		}
	}
	durable, err := os.ReadFile(filepath.Join(dir, "history.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToLower(string(durable)), "do not call") {
		t.Fatal("private stop instruction leaked into persisted log")
	}
	snapshot, _ := json.Marshal(a.History())
	if strings.Contains(strings.ToLower(string(snapshot)), "do not call") {
		t.Fatal("private stop instruction leaked into snapshot")
	}
}
