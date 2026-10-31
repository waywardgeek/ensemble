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
	"sync/atomic"
	"testing"
	"time"

	"ensemble/ensemble"
)

// These tests keep the real HTTP parser, actor, history and tool supervisor.
// Only the external provider is a fake; an actual file proves execution.
func explanationText(text string) any {
	return map[string]any{"type": "message", "role": "assistant", "content": []any{map[string]any{"type": "output_text", "text": text}}}
}
func explanationCall(id string) any {
	return map[string]any{"type": "function_call", "call_id": id, "name": "mark", "arguments": "{}"}
}
func explanationResponse(w http.ResponseWriter, model string, output ...any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"status": "completed", "model": model, "output": output, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
}
func explanationAgent(t *testing.T, server *httptest.Server, model string, rounds int) (ensemble.Agent, ensemble.Config, string, *atomic.Int32) {
	t.Helper()
	dir := t.TempDir()
	cfg := ensemble.Config{Vendor: ensemble.OpenAI, Model: model, ResolvedModel: model, BillingMode: ensemble.CredentialAPIKey, APIKey: "unused-http-fixture-key", BaseURL: server.URL, DataDir: dir, SavePath: filepath.Join(dir, "save.json"), MaxToolRounds: rounds}
	a, err := ensemble.New(io.Discard).NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { a.Shutdown() })
	if err := a.Load(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "executed.txt")
	calls := new(atomic.Int32)
	if err := a.RegisterTool(ensemble.ToolDeclaration{Name: "mark", Schema: json.RawMessage(`{"type":"object"}`)}, func(ensemble.ToolContext, json.RawMessage) (string, error) {
		calls.Add(1)
		return "marker written", os.WriteFile(path, []byte("executed"), 0600)
	}); err != nil {
		t.Fatal(err)
	}
	return a, cfg, path, calls
}
func explanationAsk(t *testing.T, a ensemble.Agent) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	text, err := a.Ask(ctx, "I already explained the task: inspect state so we can diagnose it. Use mark.")
	if err != nil {
		t.Fatal(err)
	}
	return text
}

func TestVisibleReasoningRejectsWholeBatchAndAllowsRetry(t *testing.T) {
	for _, kind := range []string{"absent", "whitespace", "summary-only", "after-call", "between-calls"} {
		t.Run(kind, func(t *testing.T) {
			var requests atomic.Int32
			var marker string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				switch requests.Add(1) {
				case 1:
					output := []any{explanationCall("blocked-a"), explanationCall("blocked-b")}
					switch kind {
					case "whitespace":
						output = append([]any{explanationText(" \n\t ")}, output...)
					case "summary-only":
						output = append([]any{map[string]any{"type": "reasoning", "id": "r1", "encrypted_content": "opaque-fixture", "summary": []any{map[string]any{"type": "summary_text", "text": "Inspecting state to diagnose the problem"}}}}, output...)
					case "after-call":
						output = append(output, explanationText("I will inspect state because it explains the problem."))
					case "between-calls":
						output = []any{output[0], explanationText("I will inspect state because it explains the problem."), output[1]}
					}
					explanationResponse(w, "gpt-6-astra", output...)
				case 2:
					if _, err := os.Stat(marker); !os.IsNotExist(err) {
						t.Errorf("blocked batch caused a side effect: %v", err)
					}
					count := 0
					for _, raw := range body["input"].([]any) {
						item := raw.(map[string]any)
						if item["type"] == "function_call_output" {
							count++
							if !strings.Contains(item["output"].(string), "Not executed:") {
								t.Errorf("missing corrective error: %v", item)
							}
						}
					}
					if count != 2 {
						t.Errorf("got %d paired errors, want 2", count)
					}
					if kind == "summary-only" {
						encoded, _ := json.Marshal(body)
						if !strings.Contains(string(encoded), "opaque-fixture") {
							t.Error("rejection discarded opaque provider replay data")
						}
					}
					explanationResponse(w, "gpt-6-astra", explanationText("I will write the marker to verify that the repaired call really executes."), explanationCall("accepted"))
				case 3:
					explanationResponse(w, "gpt-6-astra", explanationText("Done."))
				default:
					t.Error("unexpected retry")
					explanationResponse(w, "gpt-6-astra", explanationText("Stop."))
				}
			}))
			defer server.Close()
			a, cfg, path, calls := explanationAgent(t, server, "gpt-6-astra", 4)
			marker = path
			if got := explanationAsk(t, a); got != "Done." {
				t.Fatalf("reply %q", got)
			}
			if calls.Load() != 1 {
				t.Fatalf("tool executions=%d, want exactly one", calls.Load())
			}
			if _, err := os.Stat(path); err != nil {
				t.Fatal(err)
			}
			rejected, dispatched := 0, 0
			for _, e := range events(t, a) {
				if e.Type == "tool_called" {
					dispatched++
					if e.Tool.CallID != "accepted" {
						t.Errorf("blocked call dispatched: %s", e.Tool.CallID)
					}
				}
				if e.Type == "tool_returned" && strings.HasPrefix(e.Tool.CallID, "blocked-") {
					rejected++
					if !e.Tool.IsError || e.Tool.Job != nil {
						t.Errorf("blocked call wasn't a job-free error: %+v", e.Tool)
					}
				}
			}
			if rejected != 2 || dispatched != 1 {
				t.Fatalf("rejected=%d dispatched=%d", rejected, dispatched)
			}
			// The next accepted tool must receive the first job: rejection allocated none.
			if _, err := a.Jobs().Find(2); err == nil {
				t.Error("rejection consumed job handles")
			}
			if err := a.Shutdown(); err != nil {
				t.Fatal(err)
			}
			if err := a.Save(); err != nil {
				t.Fatal(err)
			}
			restored, err := ensemble.New(io.Discard).NewAgent(cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer restored.Shutdown()
			if err := restored.Load(); err != nil {
				t.Fatal(err)
			}
			var log strings.Builder
			if err := restored.History().Dump(&log); err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"blocked-a", "Not executed:", "verify that the repaired call really executes"} {
				if !strings.Contains(log.String(), want) {
					t.Errorf("saved explanation/error missing %q", want)
				}
			}
		})
	}
}

func TestVisibleReasoningPresenceNotLengthAndDefaultDisabled(t *testing.T) {
	for _, tc := range []struct{ name, model, text string }{
		{"one-character", "gpt-6-astra", "x"},
		{"synthetic-default-disabled", "gpt-ch19-course", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if requests.Add(1) == 1 {
					output := []any{explanationCall("one")}
					if tc.text != "" {
						output = append([]any{explanationText(tc.text)}, output...)
					}
					explanationResponse(w, tc.model, output...)
				} else {
					explanationResponse(w, tc.model, explanationText("Done."))
				}
			}))
			defer server.Close()
			a, _, _, calls := explanationAgent(t, server, tc.model, 2)
			explanationAsk(t, a)
			if calls.Load() != 1 {
				t.Fatalf("tool executions=%d", calls.Load())
			}
		})
	}
}

func TestVisibleReasoningRequiredAgainForNextResponse(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch requests.Add(1) {
		case 1:
			explanationResponse(w, "gpt-6-astra", explanationText("I will mark the first step so we can verify progress."), explanationCall("first"))
		case 2:
			explanationResponse(w, "gpt-6-astra", explanationCall("unexplained-next"))
		default:
			explanationResponse(w, "gpt-6-astra", explanationText("Done."))
		}
	}))
	defer server.Close()
	a, _, _, calls := explanationAgent(t, server, "gpt-6-astra", 3)
	explanationAsk(t, a)
	if calls.Load() != 1 {
		t.Fatalf("previous response authorized unexplained next call: %d executions", calls.Load())
	}
}

func TestVisibleReasoningRejectionsUseRoundLimit(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := requests.Add(1)
		if n > 3 {
			explanationResponse(w, "gpt-6-astra", explanationText("Unexpected extra retry."))
			return
		}
		explanationResponse(w, "gpt-6-astra", explanationCall("blocked"))
	}))
	defer server.Close()
	a, _, path, calls := explanationAgent(t, server, "gpt-6-astra", 2)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := a.Ask(ctx, "Act"); !errors.Is(err, ensemble.ErrRoundLimit) {
		t.Fatalf("want round limit, got %v", err)
	}
	if requests.Load() != 3 || calls.Load() != 0 {
		t.Fatalf("requests=%d executions=%d", requests.Load(), calls.Load())
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("blocked tool created marker: %v", err)
	}
	if _, err := a.Jobs().Find(1); err == nil {
		t.Error("blocked calls allocated jobs")
	}
}

func TestVisibleReasoningChatCompletionsAndStreamOrder(t *testing.T) {
	for _, kind := range []string{"json-no-text", "json-explained", "stream-tool-first", "stream-whitespace-first", "stream-explained"} {
		t.Run(kind, func(t *testing.T) {
			allowed := strings.HasSuffix(kind, "-explained")
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n := requests.Add(1)
				call := map[string]any{"id": "one", "type": "function", "function": map[string]string{"name": "mark", "arguments": "{}"}}
				if n == 1 && strings.HasPrefix(kind, "stream-") {
					w.Header().Set("Content-Type", "text/event-stream")
					frame := func(delta map[string]any, finish any) {
						data, _ := json.Marshal(map[string]any{"model": "gpt-4o", "choices": []any{map[string]any{"index": 0, "delta": delta, "finish_reason": finish}}})
						io.WriteString(w, "data: "+string(data)+"\n\n")
						w.(http.Flusher).Flush()
					}
					if kind == "stream-explained" {
						frame(map[string]any{"content": "I will write a marker so we can verify execution."}, nil)
					}
					if kind == "stream-whitespace-first" {
						frame(map[string]any{"content": " \n\t "}, nil)
					}
					call["index"] = 0
					frame(map[string]any{"tool_calls": []any{call}}, nil)
					if !allowed {
						frame(map[string]any{"content": "I will write a marker so we can verify execution, but this explanation arrived too late."}, nil)
					}
					frame(map[string]any{}, "tool_calls")
					io.WriteString(w, "data: {\"model\":\"gpt-4o\",\"choices\":[],\"usage\":{\"prompt_tokens\":1,\"completion_tokens\":1}}\n\ndata: [DONE]\n\n")
					return
				}
				message := map[string]any{"role": "assistant", "content": "Done."}
				if n == 1 {
					message["content"] = nil
					if allowed {
						message["content"] = "I will write a marker so we can verify execution."
					}
					message["tool_calls"] = []any{call}
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{"model": "gpt-4o", "choices": []any{map[string]any{"message": message}}, "usage": map[string]int{"prompt_tokens": 1, "completion_tokens": 1}})
			}))
			defer server.Close()
			a, _, _, calls := explanationAgent(t, server, "gpt-4o", 2)
			explanationAsk(t, a)
			expected := int32(0)
			if allowed {
				expected = 1
			}
			if calls.Load() != expected {
				t.Fatalf("tool executions=%d want %d", calls.Load(), expected)
			}
			if !allowed {
				paired := false
				for _, e := range events(t, a) {
					if e.Type == "tool_returned" && e.Tool.CallID == "one" {
						paired = e.Tool.IsError
						if strings.HasPrefix(kind, "stream-") && !strings.Contains(e.Tool.Parts[0].Text, "arrived in the stream") {
							t.Error("missing durable reason for stream-order rejection")
						}
					}
				}
				if !paired {
					t.Error("missing paired rejection")
				}
			}
		})
	}
}
