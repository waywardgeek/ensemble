package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"ensemble"
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
)

func TestOpenAIBusyHintToolAdjacencyHTTP(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprintf("stream_%v", stream), func(t *testing.T) {
			const hint = "Keep the actual job observation and use the new direction."
			const callID = "held-call"
			entered, release := make(chan struct{}, 1), make(chan struct{})
			var releaseOnce sync.Once
			unblock := func() { releaseOnce.Do(func() { close(release) }) }
			var requests, effects atomic.Int32
			second := make(chan []map[string]json.RawMessage, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body struct {
					Messages []map[string]json.RawMessage `json:"messages"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
					w.WriteHeader(400)
					return
				}
				n := requests.Add(1)
				if n == 2 {
					second <- body.Messages
					valid := len(body.Messages) == 5
					if valid {
						var role, id, content string
						_ = json.Unmarshal(body.Messages[3]["role"], &role)
						_ = json.Unmarshal(body.Messages[3]["tool_call_id"], &id)
						_ = json.Unmarshal(body.Messages[4]["content"], &content)
						valid = role == "tool" && id == callID && content == hint
					}
					if !valid {
						http.Error(w, "tool results must immediately follow assistant tool_calls", 400)
						return
					}
				} else if n != 1 {
					http.Error(w, "unexpected extra request", 500)
					return
				}
				if stream {
					w.Header().Set("Content-Type", "text/event-stream")
					if n == 1 {
						io.WriteString(w, "data: {\"model\":\"served\",\"choices\":[{\"delta\":{\"content\":\"Starting the job.\"}}]}\n\n")
						io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"id\":\"held-call\",\"type\":\"function\",\"function\":{\"name\":\"held\",\"arguments\":\"{}\"}}]},\"finish_reason\":\"tool_calls\"}],\"usage\":{\"prompt_tokens\":3,\"completion_tokens\":1}}\n\ndata: [DONE]\n\n")
					} else {
						io.WriteString(w, "data: {\"model\":\"served\",\"choices\":[{\"delta\":{\"content\":\"continued with hint\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":3,\"completion_tokens\":1}}\n\ndata: [DONE]\n\n")
					}
				} else {
					w.Header().Set("Content-Type", "application/json")
					if n == 1 {
						io.WriteString(w, `{"model":"served","choices":[{"message":{"content":"Starting the job.","tool_calls":[{"id":"held-call","type":"function","function":{"name":"held","arguments":"{}"}}]},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":3,"completion_tokens":1}}`)
					} else {
						io.WriteString(w, `{"model":"served","choices":[{"message":{"content":"continued with hint"},"finish_reason":"stop"}],"usage":{"prompt_tokens":3,"completion_tokens":1}}`)
					}
				}
			}))
			defer server.Close()
			dir := t.TempDir()
			logPath := filepath.Join(dir, "history.jsonl")
			app := ensemble.New(nil)
			defer app.Shutdown()
			defer unblock()
			a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-5.4", APIKey: "dummy", BaseURL: server.URL, BaseDir: dir, DataDir: filepath.Join(dir, "jobs"), LogPath: logPath, DisableStreaming: !stream, Tools: []string{}})
			if err != nil {
				t.Fatal(err)
			}
			err = a.Tools().Register(ensemble.ToolDefinition{Name: "held", Description: "Hold a real job for a hint.", Parameters: json.RawMessage(`{"type":"object"}`)}, func(ctx context.Context, _ ensemble.ToolCall, _ json.RawMessage) error {
				effects.Add(1)
				select {
				case entered <- struct{}{}:
				default:
				}
				select {
				case <-release:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			reply := a.Submit(ctx, "start a job then accept a busy hint")
			select {
			case <-entered:
			case <-ctx.Done():
				t.Fatal("handler did not start")
			}
			if err = a.Post(ensemble.Hint{Text: hint}); err != nil {
				t.Fatal(err)
			}
			var r ensemble.Result
			select {
			case r = <-reply:
			case <-ctx.Done():
				t.Fatal("turn did not finish")
			}
			if r.Err != nil || r.Text != "continued with hint" {
				t.Fatalf("actual continuation rejected: %+v", r)
			}
			if effects.Load() != 1 || requests.Load() != 2 {
				t.Fatalf("effects=%d requests=%d", effects.Load(), requests.Load())
			}
			var messages []map[string]json.RawMessage
			select {
			case messages = <-second:
			default:
				t.Fatal("missing second HTTP request")
			}
			var wireResult string
			if err = json.Unmarshal(messages[3]["content"], &wireResult); err != nil {
				t.Fatal(err)
			}
			contextBefore := a.Log().Context()
			before, _ := json.Marshal(contextBefore)
			callAt, hintAt, resultAt := -1, -1, -1
			for i, e := range contextBefore.Dialogue {
				for _, p := range e.Parts {
					if p.Type == "tool_call" && p.CallID == callID {
						callAt = i
					}
					if p.Type == "text" && p.Text == hint {
						hintAt = i
					}
					if p.Type == "tool_result" && p.CallID == callID {
						resultAt = i
						var texts []string
						for _, part := range p.Parts {
							texts = append(texts, part.Text)
						}
						if wireResult != strings.Join(texts, "\n") || !strings.Contains(wireResult, "running") {
							t.Errorf("real running result altered: %q vs %+v", wireResult, p)
						}
					}
				}
			}
			if !(callAt >= 0 && callAt < hintAt && hintAt < resultAt) {
				t.Fatalf("History projection no longer causal: %s", before)
			}
			if _, err = a.Render(); err != nil {
				t.Fatal(err)
			}
			after, _ := json.Marshal(a.Log().Context())
			if !bytes.Equal(before, after) {
				t.Fatal("render mutated History projection")
			}
			data, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatal(err)
			}
			calledAt, receivedAt, returnedAt := -1, -1, -1
			for i, line := range bytes.Split(bytes.TrimSpace(data), []byte("\n")) {
				var e ensemble.Event
				if err = json.Unmarshal(line, &e); err != nil {
					t.Fatal(err)
				}
				switch e.Type {
				case ensemble.ToolCalled:
					calledAt = i
				case ensemble.HintReceived:
					receivedAt = i
				case ensemble.ToolReturned:
					returnedAt = i
					if e.Tool.CallID != callID || e.Tool.Job == nil || e.Tool.Job.Status != ensemble.JobRunning {
						t.Errorf("saved result upgraded or lost: %+v", e.Tool)
					}
				}
			}
			if !(calledAt >= 0 && calledAt < receivedAt && receivedAt < returnedAt) {
				t.Fatalf("saved causal order changed: %s", data)
			}
			want := ensemble.Usage{Input: 6, Output: 2}
			if a.Engine().Usage() != want || a.Log().Context().Usage != want {
				t.Errorf("usage changed: engine=%+v history=%+v", a.Engine().Usage(), a.Log().Context().Usage)
			}
		})
	}
}
