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
	"testing"
	"time"

	"ensemble/ensemble"
	"ensemble/internal/common"
)

// Posting from the dispatch notification orders the hint before the real setter's
// report, including when that handler returns immediately. It never blocks the owner.
type batchHintObserver struct {
	agent ensemble.Agent
	hint  string
}

func (o batchHintObserver) Observe(event ensemble.Observation) {
	if event, ok := event.(ensemble.ToolDispatched); ok && event.CallID == "observed" {
		_ = o.agent.Post(ensemble.Hint{Text: o.hint})
		_ = o.agent.Post(ensemble.Hint{Text: o.hint + " (second hint)"})
	}
}

// Real Agent, PTY, jobs, parsing and rendering; only the model is local HTTP.
// The original job cannot finish until the test sends input after the turn.
func TestHintInterruptsUndispatchedBatch(t *testing.T) {
	for _, observation := range []string{"initial", "wait_for_job", "setter-race"} {
		t.Run(observation, func(t *testing.T) {
			dir := t.TempDir()
			ready := filepath.Join(dir, "ready")
			marker := filepath.Join(dir, "must-not-exist")
			fresh := filepath.Join(dir, "fresh")
			const hint = "Genuine human hint: abandon the queued tail and write fresh instead."
			const interrupted = "Not executed: interrupted by a hint before dispatch. This call did not run."
			type wireBlock struct {
				Type    string      `json:"type"`
				ID      string      `json:"tool_use_id"`
				Content []wireBlock `json:"content"`
				Text    string      `json:"text"`
				IsError bool        `json:"is_error"`
			}
			resultText := func(block wireBlock) string {
				var text string
				for _, part := range block.Content {
					text += part.Text
				}
				return text
			}
			type wireRequest struct {
				Messages []struct {
					Role    string      `json:"role"`
					Content []wireBlock `json:"content"`
				} `json:"messages"`
			}
			requests := make(chan wireRequest, 4)
			requestCount := 0
			batchRequest := 1
			if observation != "initial" {
				batchRequest = 2
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++
				var request wireRequest
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
				}
				blocks := []map[string]any{{"type": "text", "text": "I will exercise the requested batch."}}
				call := func(id, name string, args map[string]any) {
					blocks = append(blocks, map[string]any{"type": "tool_use", "id": id, "name": name, "input": args})
				}
				shell := map[string]any{"command": fmt.Sprintf("printf READY > %s; printf READY; read line; printf 'RESUMED:%%s' \"$line\"", ready), "ai_callback_delay": 30}
				switch {
				case requestCount < batchRequest:
					shell["ai_callback_pattern"] = "READY"
					call("launch", "run_command", shell)
				case requestCount == batchRequest:
					call("completed", "tool_limits", map[string]any{"max_output_bytes": 1000})
					if observation == "initial" {
						call("observed", "run_command", shell)
					} else if observation == "setter-race" {
						call("observed", "tool_limits", map[string]any{"max_output_bytes": 1})
					} else {
						call("observed", "wait_for_job", map[string]any{"handle": 1, "ai_callback_delay": 30})
					}
					call("tail-wait", "wait_for_job", map[string]any{"handle": 1, "ai_callback_delay": 30})
					call("tail-file", "write_file", map[string]any{"path": marker, "content": "wrong"})
					call("tail-shell", "run_command", map[string]any{"command": "printf wrong >> " + marker})
					call("tail-setter", "tool_limits", map[string]any{"max_output_bytes": 2})
				case requestCount == batchRequest+1:
					requests <- request
					call("fresh", "write_file", map[string]any{"path": fresh, "content": "new decision"})
				default:
					requests <- request
					blocks = []map[string]any{{"type": "text", "text": "done"}}
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"content": blocks, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
			}))
			defer server.Close()
			a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{DataDir: dir, Model: "fake", BaseURL: server.URL, BuiltinTools: true})
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := a.Shutdown(); err != nil {
					t.Error(err)
				}
			}()
			if observation == "setter-race" {
				a.Observe(batchHintObserver{a, hint})
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			done := make(chan common.Result, 1)
			go func() { text, err := a.Ask(ctx, "start the batch"); done <- common.Result{Text: text, Err: err} }()
			for {
				for _, event := range a.History().EventsAfter(0) {
					if event.Type == "tool_called" && event.Tool.CallID == "observed" {
						goto observing
					}
				}
				select {
				case result := <-done:
					t.Fatalf("turn ended before observation: %+v; history: %+v", result, a.History().EventsAfter(0))
				case <-ctx.Done():
					t.Fatalf("observation never dispatched; history: %+v", a.History().EventsAfter(0))
				case <-time.After(time.Millisecond):
				}
			}
		observing:
			for {
				if _, err := os.Stat(ready); err == nil {
					break
				}
				select {
				case <-ctx.Done():
					t.Fatal("real shell never started")
				case <-time.After(time.Millisecond):
				}
			}
			if observation != "setter-race" {
				if err := a.Post(ensemble.Hint{Text: hint}); err != nil {
					t.Fatal(err)
				}
			}
			var next wireRequest
			select {
			case next = <-requests:
			case <-time.After(time.Second):
				var woke, tail bool
				for _, event := range a.History().EventsAfter(0) {
					if event.Type == "tool_returned" && event.Tool.CallID == "observed" {
						woke = true
						t.Logf("real observed report: %+v", event.Tool.Parts)
					}
					if event.Type == "tool_called" && event.Tool.CallID == "tail-wait" {
						tail = true
					}
				}
				t.Fatalf("hint failed to reach next model request: observed-woke=%t tail-wait-dispatched=%t", woke, tail)
			}
			job, err := a.Jobs().Find(1)
			if err != nil || job.Data().Status != "running" {
				t.Fatalf("original job not still running: %v", err)
			}
			results := map[string]wireBlock{}
			hintSeen := false
			for _, message := range next.Messages {
				for _, block := range message.Content {
					if block.Type == "tool_result" {
						if _, duplicate := results[block.ID]; duplicate {
							t.Errorf("duplicate result for %s", block.ID)
						}
						results[block.ID] = block
					}
					if block.Type == "text" && strings.Contains(block.Text, hint) && message.Role == "user" {
						hintSeen = true
					}
				}
			}
			if !hintSeen {
				t.Error("next request omitted genuine user hint")
			}
			if got := results["completed"]; got.IsError || !strings.Contains(resultText(got), "Limits set") {
				t.Errorf("completed result changed: %+v", got)
			}
			observedText := "job 1 running"
			if observation == "setter-race" {
				observedText = "Limits set"
			}
			if got := results["observed"]; got.IsError || !strings.Contains(resultText(got), observedText) {
				t.Errorf("dispatched result changed: %+v", got)
			}
			for _, id := range []string{"tail-wait", "tail-file", "tail-shell", "tail-setter"} {
				want := interrupted
				if observation == "setter-race" && id == "tail-wait" {
					want = `[tool_limits consumed by this wait_for_job call: ai_callback_delay 3s, ai_callback_pattern "", max_output_bytes 1]` + "\n" + want
				}
				if got := results[id]; !got.IsError || resultText(got) != want {
					t.Errorf("tail correlation/truth %s: %+v", id, got)
				}
			}
			select {
			case result := <-done:
				if result.Err != nil || result.Text != "done" {
					t.Fatalf("new calls stalled: %+v", result)
				}
			case <-ctx.Done():
				t.Fatal("new calls waited for original job")
			}
			if _, err := os.Stat(marker); !os.IsNotExist(err) {
				t.Errorf("undispatched tail had effects: %v", err)
			}
			if data, err := os.ReadFile(fresh); err != nil || string(data) != "new decision" {
				t.Errorf("fresh call did not run: %q %v", data, err)
			}
			if _, err := a.Jobs().Find(3); err == nil {
				t.Error("undispatched tail created a job")
			}
			returned := map[string]int{}
			for _, event := range a.History().EventsAfter(0) {
				if event.Type == "tool_returned" {
					returned[event.Tool.CallID]++
					if got, ok := results[event.Tool.CallID]; ok {
						var text string
						for _, part := range event.Tool.Parts {
							text += part.Text
						}
						if text != resultText(got) || event.Tool.IsError != got.IsError {
							t.Errorf("wire changed real result %s", event.Tool.CallID)
						}
					}
					if event.Tool.CallID == "fresh" && strings.Contains(event.Tool.Parts[0].Text, "tool_limits consumed") {
						t.Error("pending one-shot leaked into new call")
					}
				}
				if event.Type == "tool_called" && strings.HasPrefix(event.Tool.CallID, "tail-") {
					t.Errorf("tail dispatched/replayed: %s", event.Tool.CallID)
				}
			}
			for _, id := range []string{"completed", "observed", "tail-wait", "tail-file", "tail-shell", "tail-setter", "fresh"} {
				if returned[id] != 1 {
					t.Errorf("result %s occurred %d times", id, returned[id])
				}
			}
			if job.Data().Status != "running" {
				t.Fatal("hint canceled original job")
			}
			started := time.Now()
			text, err := a.Invoke(ctx, "wait_for_job", json.RawMessage(`{"handle":1,"ai_callback_delay":0.08}`))
			if err != nil || time.Since(started) < 60*time.Millisecond || !strings.Contains(text, "running") {
				t.Fatalf("hint leaked into later timer: %q %v elapsed %s", text, err, time.Since(started))
			}
			if err := job.Send("continue"); err != nil {
				t.Fatal(err)
			}
			text, err = a.Invoke(ctx, "wait_for_job", json.RawMessage(`{"handle":1,"ai_callback_delay":2}`))
			if err != nil || !strings.Contains(text, "RESUMED:continue") {
				t.Fatalf("job resumption: %q %v", text, err)
			}
		})
	}
}

func TestHintBatchWithoutHintExecutesNormally(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "normal")
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls != 1 {
			replyText(w, "normal done")
			return
		}
		blocks := []map[string]any{{"type": "text", "text": "I will write, read, then append to verify ordinary serial dispatch."}}
		for i, call := range []toolCall{
			{"tool_limits", map[string]any{"max_output_bytes": 1000}},
			{"run_command", map[string]any{"command": "printf first > " + path, "ai_callback_delay": 2}},
			{"read_file", map[string]any{"path": path}},
			{"run_command", map[string]any{"command": "printf second >> " + path, "ai_callback_delay": 2}},
		} {
			blocks = append(blocks, map[string]any{"type": "tool_use", "id": fmt.Sprint(i), "name": call.name, "input": call.args})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"content": blocks, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
	}))
	defer server.Close()
	a := actorAgent(t, server, func(cfg *ensemble.Config) { cfg.BuiltinTools = true; cfg.Workspace = dir })
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if text, err := a.Ask(ctx, "normal batch"); err != nil || text != "normal done" {
		t.Fatalf("normal batch: %q %v", text, err)
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != "firstsecond" {
		t.Fatalf("ordered effects: %q %v", data, err)
	}
	called, returned := map[string]int{}, map[string]int{}
	for _, event := range a.History().EventsAfter(0) {
		switch event.Type {
		case "tool_called":
			called[event.Tool.CallID]++
		case "tool_returned":
			returned[event.Tool.CallID]++
			if event.Tool.IsError {
				t.Errorf("ordinary call failed: %+v", event.Tool)
			}
			var text string
			for _, part := range event.Tool.Parts {
				text += part.Text
			}
			if event.Tool.CallID == "1" && !strings.HasPrefix(text, "[tool_limits consumed by this run_command call:") {
				t.Errorf("normal one-shot not consumed: %q", text)
			}
			if event.Tool.CallID == "2" && text != "first" {
				t.Errorf("serial read: %q", text)
			}
		}
	}
	for _, id := range []string{"0", "1", "2", "3"} {
		if called[id] != 1 || returned[id] != 1 {
			t.Errorf("call %s dispatched %d, returned %d", id, called[id], returned[id])
		}
	}
}
