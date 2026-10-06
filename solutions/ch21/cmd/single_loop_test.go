package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
	"github.com/waywardgeek/ensemble/agent/internal/llm"
)

// Exercise the actual settings file and subprocess entry points, not just
// Config.MaxToolRounds. A small saved limit distinguishes live settings from
// an implementation that merely raised its hard-coded default to 200.
func TestCLISavedRoundLimitAndResume(t *testing.T) {
	bin := buildAs(t, "single-loop")
	for _, prompt := range []string{`{"kind":"prompt","text":"work"}`, `{"user":"work"}`} {
		for _, limit := range []int{1, 200} {
			t.Run(fmt.Sprintf("%s/limit%d", prompt, limit), func(t *testing.T) {
				work := t.TempDir()
				var requests atomic.Int32
				var resumed atomic.Bool
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body, err := io.ReadAll(r.Body)
					if err != nil {
						t.Error(err)
						http.Error(w, "read request", 400)
						return
					}
					n := requests.Add(1)
					if n == 1 && !strings.Contains(string(body), "ephemeral-marker") {
						t.Error("legacy ephemeral attachment did not reach the first request")
					}
					parts := []map[string]any{{"type": "text", "text": "COMPLETE"}}
					if !resumed.Load() && n <= 18 {
						parts = []map[string]any{
							{"type": "text", "text": "I will inspect the directory."},
							{"type": "tool_use", "id": fmt.Sprintf("call%d", n), "name": "list_directory", "input": map[string]any{"path": work}},
						}
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]any{"id": "test", "type": "message", "role": "assistant", "model": "claude-sonnet-5", "content": parts, "stop_reason": "end_turn", "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
				}))
				defer srv.Close()
				if err := os.WriteFile(filepath.Join(work, "settings.json"), []byte(fmt.Sprintf(`{"max_tool_rounds":%d}`, limit)), 0600); err != nil {
					t.Fatal(err)
				}
				env := []string{"LLM_VENDOR=anthropic", "LLM_MODEL=claude-sonnet-5", "LLM_API_KEY=test", "LLM_BASE_URL=" + srv.URL}
				stdout, stderr, code := runBin(t, bin, work, "{\"ephemeral\":\"ephemeral-marker\"}\n"+prompt+"\n", env...)
				wantRequests, wantCode := int32(19), 0
				if limit == 1 {
					wantRequests, wantCode = 2, 1
				}
				if code != wantCode || requests.Load() != wantRequests {
					t.Fatalf("exit %d, requests %d; want %d, %d\nstdout: %s\nstderr: %s", code, requests.Load(), wantCode, wantRequests, stdout, stderr)
				}
				answers, failures, ended := 0, 0, 0
				for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
					var msg map[string]json.RawMessage
					if err := json.Unmarshal([]byte(line), &msg); err != nil {
						t.Fatal(err)
					}
					if observation, ok := msg["observation"]; ok {
						if string(observation) == `"turn_ended"` {
							ended++
							_, hasError := msg["error"]
							if hasError != (limit == 1) {
								t.Errorf("terminal observation disagrees with outcome: %s", line)
							}
						}
						continue
					}
					if text, ok := msg["assistant"]; ok {
						answers++
						if string(text) != `"COMPLETE"` {
							t.Errorf("unexpected final answer: %s", text)
						}
					}
					if text, ok := msg["error"]; ok {
						failures++
						if !strings.Contains(string(text), "stopped after 1 rounds") {
							t.Errorf("wrong error: %s", text)
						}
					}
				}
				if ended != 1 {
					t.Fatalf("got %d terminal observations, want one", ended)
				}
				if limit == 1 {
					if answers != 0 || failures != 1 {
						t.Fatalf("limit stop: %d answers, %d errors", answers, failures)
					}
				} else if answers != 1 || failures != 0 {
					t.Fatalf("normal completion: %d answers, %d errors", answers, failures)
				}
				data, err := os.ReadFile(filepath.Join(work, "save.json"))
				if err != nil {
					t.Fatal(err)
				}
				var saved llm.SaveFile
				if err := json.Unmarshal(data, &saved); err != nil {
					t.Fatal(err)
				}
				ctx := saved.Restore(func(err error) { t.Error(err) })
				restored := &llm.Engine{Ctx: ctx}
				if pending := restored.PendingCalls(); len(pending) != 0 {
					t.Fatalf("unpaired calls after save: %+v", pending)
				}
				results, refused := 0, 0
				for _, event := range saved.Log {
					if event.Type != common.ToolReturned {
						continue
					}
					results++
					if event.Tool.IsError {
						refused++
						if !strings.Contains(event.Tool.Parts[0].(common.TextPart).Text, "Not executed") {
							t.Errorf("wrong refusal: %+v", event.Tool)
						}
					}
				}
				wantResults, wantRefused := 18, 0
				if limit == 1 {
					wantResults, wantRefused = 2, 1
				}
				if results != wantResults || refused != wantRefused {
					t.Fatalf("results/refused %d/%d, want %d/%d", results, refused, wantResults, wantRefused)
				}
				resumed.Store(true)
				stdout, stderr, code = runBin(t, bin, work, prompt+"\n", env...)
				if code != 0 || !strings.Contains(stdout, `"assistant":"COMPLETE"`) || requests.Load() != wantRequests+1 {
					t.Fatalf("restart did not resume cleanly: exit %d\nstdout: %s\nstderr: %s", code, stdout, stderr)
				}
			})
		}
	}
}
