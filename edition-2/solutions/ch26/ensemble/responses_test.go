package ensemble_test

import (
	"context"
	"encoding/json"
	"ensemble/ensemble"
	"ensemble/internal/common"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// A gated external server proves summary delivery precedes completion while the
// real actor executes the adjacent tool and renders its actual result next turn.
func TestResponsesSummaryToolLoop(t *testing.T) {
	gate := make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(gate) }) }
	defer release()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/responses" {
			t.Errorf("wrong transport: %s", r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["stream"] != true || body["store"] != false {
			t.Error("Responses delivery flags missing")
		}
		input := body["input"].([]any)
		first := input[0].(map[string]any)
		if first["role"] != "developer" || first["content"].([]any)[0].(map[string]any)["prompt_cache_breakpoint"] == nil {
			t.Error("developer cache block missing")
		}
		// Misleading metadata must not turn the framing fallback into a buffered
		// whole-response read; the summary itself releases the provider below.
		w.Header().Set("Content-Type", "application/json")
		if calls == 0 {
			calls++
			frame(w, wireFrame{"response.reasoning_summary_text.delta", decoded(`{"output_index":0,"delta":"consider "}`)})
			frame(w, wireFrame{"response.reasoning_summary_text.delta", decoded(`{"output_index":0,"delta":"the file"}`)})
			select {
			case <-gate:
			case <-r.Context().Done():
				return
			}
			frame(w, wireFrame{"response.output_item.added", decoded(`{"output_index":1,"item":{"type":"function_call","name":"echo"}}`)})
			frame(w, wireFrame{"response.function_call_arguments.delta", decoded(`{"output_index":1,"delta":"{}"}`)})
			frame(w, wireFrame{"response.completed", decoded(`{"response":{"status":"completed","model":"gpt-ch19-course","output":[{"type":"reasoning","id":"r1","summary":[{"type":"summary_text","text":"consider the file"}],"encrypted_content":"private"},{"type":"function_call","call_id":"c1","name":"echo","arguments":"{}"}],"usage":{"input_tokens":20,"output_tokens":4,"input_tokens_details":{"cached_tokens":5,"cache_write_tokens":3}}}}`)})
		} else {
			found, reasoning := false, false
			for _, raw := range input {
				item := raw.(map[string]any)
				if item["type"] == "function_call_output" && item["call_id"] == "c1" && item["output"] == "actual result" {
					found = true
				}
				if item["type"] == "reasoning" && item["encrypted_content"] == "private" {
					reasoning = true
				}
			}
			if !found || !reasoning {
				t.Errorf("result/reasoning replay absent: %v / %v", found, reasoning)
			}
			frame(w, wireFrame{"response.output_text.delta", decoded(`{"output_index":0,"content_index":0,"delta":"Done."}`)})
			frame(w, wireFrame{"response.completed", decoded(`{"response":{"status":"completed","model":"gpt-ch19-course","output":[{"type":"message","content":[{"type":"output_text","text":"Done."}]}],"usage":{"input_tokens":12,"output_tokens":2}}}`)})
		}
	}))
	defer server.Close()
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-ch19-course", BaseURL: server.URL, DataDir: t.TempDir(), Recall: ensemble.RecallSettings{Disabled: true}})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	if err := a.RegisterTool(ensemble.ToolDeclaration{Name: "echo", Schema: json.RawMessage(`{"type":"object"}`)}, func(ensemble.ToolContext, json.RawMessage) (string, error) {
		return "actual result", nil
	}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var chunks []string
	_, err = a.AskWatching(ctx, "Do it", ensemble.StreamCallbacks{OnDelta: func(_ uint64, k ensemble.DeltaKind, s string) {
		if k == ensemble.DeltaReasoningSummary {
			chunks = append(chunks, s)
			if len(chunks) == 2 {
				release()
			}
		}
	}})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(chunks, "") != "consider the file" {
		t.Fatalf("summaries: %q", chunks)
	}
	usage := a.Engine().Usage()
	if usage.Input != 24 || usage.CacheRead != 5 || usage.CacheWrite != 3 || usage.Output != 6 {
		t.Fatalf("disjoint usage: %+v", usage)
	}
}

// An HTTP 200 and a useful delta do not establish a completed response. Failed,
// incomplete and truncated streams must fail without recording response_ended.
func TestResponsesRequiresCompletion(t *testing.T) {
	for _, ending := range []string{"", "response.failed", "response.incomplete"} {
		t.Run(ending, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				frame(w, wireFrame{"response.output_text.delta", decoded(`{"output_index":0,"delta":"partial"}`)})
				frame(w, wireFrame{"response.output_item.done", decoded(`{"output_index":0,"item":{"type":"message","content":[{"type":"output_text","text":"partial"}]}}`)})
				if ending != "" {
					frame(w, wireFrame{ending, decoded(`{"response":{"status":"completed","output":[{"type":"message","content":[{"type":"output_text","text":"not completed"}]}],"usage":{"input_tokens":1,"output_tokens":1}}}`)})
				}
			}))
			defer server.Close()
			a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-ch19-course", BaseURL: server.URL, DataDir: t.TempDir(), Recall: ensemble.RecallSettings{Disabled: true}})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Shutdown()
			if _, err := a.Ask(context.Background(), "hello"); err == nil {
				t.Fatal("incomplete stream accepted")
			}
			for _, ev := range a.History().EventsAfter(0) {
				if ev.Response != nil {
					t.Fatal("failed stream recorded success")
				}
			}
		})
	}
}

// Earlier models retain their tested API surface; migration is not vendor-wide.
func TestResponsesModelBoundary(t *testing.T) {
	for _, model := range []string{"gpt-4o-mini", "gpt-5.6-sol", "gpt-6-astra", "gpt-6.1-sol", "gpt-ch19-course"} {
		t.Run(model, func(t *testing.T) {
			a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: model, ReasoningEffort: "high", DataDir: t.TempDir()})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Shutdown()
			body, err := a.Engine().Render(&ensemble.Context{})
			if err != nil {
				t.Fatal(err)
			}
			var req map[string]any
			json.Unmarshal(body, &req)
			want := model == "gpt-6-astra" || model == "gpt-6.1-sol" || model == "gpt-ch19-course"
			if (req["input"] != nil) != want {
				t.Fatalf("model transport: %s", body)
			}
			if want {
				reasoning, _ := req["reasoning"].(map[string]any)
				if req["model"] != model || reasoning["effort"] != "high" {
					t.Fatalf("configured model/reasoning lost: %s", body)
				}
			}
		})
	}
}

// A private reasoning item cannot cross model or transport boundaries, even
// though the ordinary visible messages remain usable in the same conversation.
func TestResponsesOpaqueProvenance(t *testing.T) {
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-ch19-course", DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	for _, from := range []common.Provenance{
		{Vendor: ensemble.OpenAI, Model: "other-model", Surface: common.Responses},
		{Vendor: ensemble.OpenAI, Model: "gpt-ch19-course", Surface: common.ChatCompletions},
	} {
		body, err := a.Engine().Render(&ensemble.Context{Dialogue: []ensemble.Entry{{Actor: common.AgentActor, Parts: []ensemble.Part{{Type: "opaque", From: from, Data: json.RawMessage(`{"type":"reasoning","encrypted_content":"foreign"}`)}}}}})
		if err != nil || strings.Contains(string(body), "foreign") {
			t.Fatalf("foreign reasoning replayed: %v", err)
		}
	}
}

// Responses input is prompt content, including its role-free replay items. The
// diagnostic must see appended bytes instead of reporting a tools-only 100% hit.
func TestResponsesLensIncludesInput(t *testing.T) {
	var log strings.Builder
	a, err := ensemble.New(&log).NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-ch19-course", DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	c := &ensemble.Context{Dialogue: []ensemble.Entry{{Actor: ensemble.Human, Parts: []ensemble.Part{{Type: "text", Text: "first"}}}}}
	first, err := a.Engine().Render(c)
	if err != nil {
		t.Fatal(err)
	}
	lens := a.Engine().CacheLens()
	lens.ObserveRequest("responses", first)
	c.Dialogue = append(c.Dialogue, ensemble.Entry{Actor: ensemble.Human, Parts: []ensemble.Part{{Type: "text", Text: "second response context"}}})
	second, err := a.Engine().Render(c)
	if err != nil {
		t.Fatal(err)
	}
	prediction := lens.ObserveRequest("responses", second)
	if prediction <= 0 || prediction >= 1 || !strings.Contains(log.String(), "input:appended") {
		t.Fatalf("Responses content omitted from prediction %.3f: %s", prediction, log.String())
	}
}

// This is the captured plan stream, with unrelated response metadata removed.
// Its completed read_file item lives in item.done; terminal output is empty.
// Controlled headers reproduce both parser defects without claiming knowledge
// of the live response's unrecorded header. Tools/files/actor remain real.
func TestResponsesCapturedPlanStream(t *testing.T) {
	fixture, err := os.ReadFile("testdata/ch19-plan-tool.sse")
	if err != nil {
		t.Fatal(err)
	}
	// The historical wire capture is unchanged. Preserve its parser/tool-loop
	// coverage with the course model, and prove a flagged real model refuses it.
	for _, model := range []string{"gpt-ch19-course", "gpt-6.1-sol"} {
		t.Run(model, func(t *testing.T) {
			for _, header := range []string{"application/json", "text/event-stream"} {
				t.Run(header, func(t *testing.T) {
					workspace := t.TempDir()
					t.Chdir(workspace)
					if err := os.WriteFile(filepath.Join(workspace, "puzzle.txt"), []byte("actual captured-plan tool result"), 0600); err != nil {
						t.Fatal(err)
					}
					requests := 0
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						requests++
						if requests == 1 {
							w.Header().Set("Content-Type", header)
							w.Write(fixture)
							return
						}
						body, _ := io.ReadAll(r.Body)
						var request struct {
							// Input contains the actual next-turn tool result, not earlier arguments.
							Input []struct {
								// Type selects a function result item.
								Type string
								// Output is the real handler's completed readout.
								Output string
								// CallID preserves provider correlation across the tool boundary.
								CallID string `json:"call_id"`
							} `json:"input"`
						}
						json.Unmarshal(body, &request)
						expected := "actual captured-plan tool result"
						if model == "gpt-6.1-sol" {
							expected = "Not executed:"
						}
						found := false
						for _, item := range request.Input {
							if item.Type == "function_call_output" && item.CallID == "call_u83Vpd7Ol6P1ixYjORjuBVXU" && strings.Contains(item.Output, expected) {
								found = true
							}
						}
						if !found {
							t.Errorf("completed streamed tool/result lost: %+v", request.Input)
						}
						w.Header().Set("Content-Type", "text/event-stream")
						frame(w, wireFrame{"response.completed", decoded(`{"response":{"status":"completed","output":[{"type":"message","content":[{"type":"output_text","text":"done"}]}],"usage":{"input_tokens":1,"output_tokens":1}}}`)})
					}))
					defer server.Close()
					a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: model, BaseURL: server.URL, Workspace: workspace, DataDir: t.TempDir(), BuiltinTools: true, Recall: ensemble.RecallSettings{Disabled: true}})
					if err != nil {
						t.Fatal(err)
					}
					defer a.Shutdown()
					answer, err := a.Ask(context.Background(), "read puzzle")
					if err != nil || answer != "done" || requests != 2 {
						t.Fatalf("captured plan stream failed: answer=%q requests=%d err=%v", answer, requests, err)
					}
					dispatched := 0
					for _, event := range events(t, a) {
						if event.Type == "tool_called" {
							dispatched++
						}
					}
					want := 1
					if model == "gpt-6.1-sol" {
						want = 0
					}
					if dispatched != want {
						t.Fatalf("dispatched=%d want=%d", dispatched, want)
					}
				})
			}
		})
	}
}
