package ensemble_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ensemble/ensemble"
	"ensemble/internal/common"
)

// TestCompactionProviderCompletion supplies valid submit JSON even on a length
// stop. Only a real successful terminal response may execute that submission;
// both successful and refused responses retain their measured billable usage.
func TestCompactionProviderCompletion(t *testing.T) {
	for _, provider := range []struct {
		name, model string
		vendor      ensemble.Vendor
	}{
		{"anthropic", "claude-sonnet-5", ensemble.Anthropic},
		{"chat", "gpt-4o", ensemble.OpenAI},
		{"responses", "gpt-6-astra", ensemble.OpenAI},
		{"gemini", "gemini-3.8-flash", ensemble.Gemini},
	} {
		for _, streamed := range []bool{false, true} {
			for _, limited := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/stream-%t/limit-%t", provider.name, streamed, limited), func(t *testing.T) {
					memory := strings.Repeat("I kept the decision and its reason. ", 1500)
					peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						var request map[string]any
						if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
							t.Error(err)
							http.Error(w, "JSON", 400)
							return
						}
						checkCompactionCap(t, provider.name, request, true)
						args, _ := json.Marshal(map[string]string{"memory": memory})
						compactionProviderReply(w, provider.name, streamed, limited, string(args))
					}))
					t.Cleanup(peer.Close)
					app := ensemble.New(io.Discard)
					parent, err := app.NewAgent(ensemble.Config{Model: provider.model, Vendor: provider.vendor, BaseURL: peer.URL, APIKey: "fake", DataDir: t.TempDir(), DisableStreaming: !streamed})
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() { parent.Shutdown() })
					raw, err := parent.Engine().Render(parent.History().Context())
					if err != nil {
						t.Fatal(err)
					}
					var ordinary map[string]any
					if err := json.Unmarshal(raw, &ordinary); err != nil {
						t.Fatal(err)
					}
					checkCompactionCap(t, provider.name, ordinary, false)
					child, err := app.CreateChild(parent, common.AgentSpec{ID: "compact", Compaction: true, DataDir: filepath.Join(parent.Config().DataDir, "children", "compact")})
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() { child.Shutdown() })
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					_, err = child.Ask(ctx, "SOURCE DATA: preserve this completed work and why.")
					status := child.Fleet().Status()
					if limited {
						if err == nil || status.State == "submitted" {
							t.Fatalf("length-limited reply submitted: %s %v", status.State, err)
						}
						for _, event := range child.History().EventsAfter(0) {
							if event.Type == "tool_called" {
								t.Fatal("length-limited submit was dispatched")
							}
						}
					} else {
						var payload struct {
							Memory string `json:"memory"`
						}
						if err != nil || json.Unmarshal(status.Payload, &payload) != nil || payload.Memory != memory {
							t.Fatalf("complete large memory rejected/truncated: state=%s err=%v bytes=%d", status.State, err, len(payload.Memory))
						}
					}
					want := common.Usage{Input: 80, CacheRead: 20, Output: 9000}
					if got := child.Engine().Usage(); got != want {
						t.Fatalf("measured worker usage=%+v want=%+v", got, want)
					}
					if parent.Engine().Usage() != (common.Usage{}) || parent.History().Context().LastUsage != (common.Usage{}) {
						t.Fatal("child request changed main usage/occupancy before managed import")
					}
				})
			}
		}
	}
}

func checkCompactionCap(t *testing.T, provider string, wire map[string]any, worker bool) {
	t.Helper()
	switch provider {
	case "anthropic":
		want := float64(4096)
		if worker {
			want = 16384
		}
		if wire["max_tokens"] != want {
			t.Errorf("Anthropic max_tokens=%v want=%v", wire["max_tokens"], want)
		}
	case "chat":
		value, exists := wire["max_completion_tokens"]
		if worker && exists || !worker && value != float64(1024) {
			t.Errorf("Chat output allowance=%v exists=%t worker=%t", value, exists, worker)
		}
	case "gemini":
		generation, _ := wire["generationConfig"].(map[string]any)
		value, exists := generation["maxOutputTokens"]
		if worker && exists || !worker && value != float64(4096) {
			t.Errorf("Gemini output allowance=%v exists=%t worker=%t", value, exists, worker)
		}
	case "responses":
		if _, exists := wire["max_output_tokens"]; exists {
			t.Error("Responses acquired an unverified output cap")
		}
	}
}

// A small HTTP peer emits each adapter's actual terminal and usage vocabulary.
// Streaming paths exercise the real SSE reconstruction before semantic parsing.
func compactionProviderReply(w http.ResponseWriter, provider string, streamed, limited bool, arguments string) {
	frame := func(kind string, value any) {
		data, _ := json.Marshal(value)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", kind, data)
	}
	explanation := "I will retain the decision and the reason for it."
	args := []byte(arguments)
	if streamed {
		w.Header().Set("Content-Type", "text/event-stream")
	} else {
		w.Header().Set("Content-Type", "application/json")
	}
	switch provider {
	case "anthropic":
		stop := "tool_use"
		if limited {
			stop = "max_tokens"
		}
		usage := map[string]int{"input_tokens": 80, "cache_read_input_tokens": 20, "output_tokens": 9000}
		text := map[string]any{"type": "text", "text": explanation}
		call := map[string]any{"type": "tool_use", "id": "call", "name": "submit", "input": json.RawMessage(args)}
		if !streamed {
			json.NewEncoder(w).Encode(map[string]any{"content": []any{text, call}, "stop_reason": stop, "usage": usage})
			return
		}
		frame("message_start", map[string]any{"type": "message_start", "message": map[string]any{"usage": map[string]int{"input_tokens": 80, "cache_read_input_tokens": 20, "output_tokens": 0}}})
		for i, block := range []any{text, call} {
			frame("content_block_start", map[string]any{"type": "content_block_start", "index": i, "content_block": block})
			frame("content_block_stop", map[string]any{"type": "content_block_stop", "index": i})
		}
		frame("message_delta", map[string]any{"type": "message_delta", "delta": map[string]string{"stop_reason": stop}, "usage": map[string]int{"output_tokens": 9000}})
		frame("message_stop", map[string]string{"type": "message_stop"})
	case "chat":
		stop := "tool_calls"
		if limited {
			stop = "length"
		}
		usage := map[string]any{"prompt_tokens": 100, "prompt_tokens_details": map[string]int{"cached_tokens": 20}, "completion_tokens": 9000}
		call := map[string]any{"index": 0, "id": "call", "type": "function", "function": map[string]string{"name": "submit", "arguments": string(args)}}
		if !streamed {
			json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]any{"role": "assistant", "content": explanation, "tool_calls": []any{call}}, "finish_reason": stop}}, "usage": usage})
			return
		}
		frame("message", map[string]any{"choices": []any{map[string]any{"delta": map[string]string{"content": explanation}}}})
		frame("message", map[string]any{"choices": []any{map[string]any{"delta": map[string]any{"tool_calls": []any{call}}}}})
		frame("message", map[string]any{"choices": []any{map[string]any{"delta": map[string]any{}, "finish_reason": stop}}})
		frame("message", map[string]any{"choices": []any{}, "usage": usage})
		fmt.Fprint(w, "data: [DONE]\n\n")
	case "gemini":
		stop := "STOP"
		if limited {
			stop = "MAX_TOKENS"
		}
		usage := map[string]int{"promptTokenCount": 100, "cachedContentTokenCount": 20, "candidatesTokenCount": 8900, "thoughtsTokenCount": 100}
		content := map[string]any{"role": "model", "parts": []any{map[string]string{"text": explanation}, map[string]any{"functionCall": map[string]any{"name": "submit", "args": json.RawMessage(args)}}}}
		response := map[string]any{"candidates": []any{map[string]any{"content": content, "finishReason": stop}}, "usageMetadata": usage}
		if streamed {
			frame("message", response)
		} else {
			json.NewEncoder(w).Encode(response)
		}
	case "responses":
		status := "completed"
		if limited {
			status = "incomplete"
		}
		message := map[string]any{"type": "message", "content": []any{map[string]string{"type": "output_text", "text": explanation}}}
		call := map[string]any{"type": "function_call", "call_id": "call", "name": "submit", "arguments": string(args)}
		response := map[string]any{"status": status, "output": []any{message, call}, "usage": map[string]any{"input_tokens": 100, "input_tokens_details": map[string]int{"cached_tokens": 20}, "output_tokens": 9000}}
		if limited {
			response["incomplete_details"] = map[string]string{"reason": "max_output_tokens"}
		}
		if !streamed {
			json.NewEncoder(w).Encode(response)
			return
		}
		frame("response.output_text.delta", map[string]any{"type": "response.output_text.delta", "output_index": 0, "content_index": 0, "delta": explanation})
		frame("response.output_item.added", map[string]any{"type": "response.output_item.added", "output_index": 1, "item": call})
		frame("response."+status, map[string]any{"type": "response." + status, "response": response})
	}
}

// TestCompactionMalformedArguments retains measured usage when a completed
// provider envelope contains invalid JSON arguments, without dispatching tools.
func TestCompactionMalformedArguments(t *testing.T) {
	for _, provider := range []struct{ name, model string }{{"chat", "gpt-4o"}, {"responses", "gpt-6-astra"}} {
		for _, streamed := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/stream-%t", provider.name, streamed), func(t *testing.T) {
				peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					compactionProviderReply(w, provider.name, streamed, false, `{"memory":`)
				}))
				t.Cleanup(peer.Close)
				app := ensemble.New(io.Discard)
				parent, err := app.NewAgent(ensemble.Config{Model: provider.model, Vendor: ensemble.OpenAI, BaseURL: peer.URL, APIKey: "fake", DataDir: t.TempDir(), DisableStreaming: !streamed})
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { parent.Shutdown() })
				child, err := app.CreateChild(parent, common.AgentSpec{ID: "compact", Compaction: true, DataDir: filepath.Join(parent.Config().DataDir, "children", "compact")})
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { child.Shutdown() })
				if _, err := child.Ask(submissionContext(t), "preserve this source"); err == nil || child.Fleet().Status().State == "submitted" {
					t.Fatalf("malformed arguments accepted: %v", err)
				}
				if got := child.Engine().Usage(); got != (common.Usage{Input: 80, CacheRead: 20, Output: 9000}) {
					t.Fatalf("failed response usage=%+v", got)
				}
				for _, event := range child.History().EventsAfter(0) {
					if event.Type == "tool_called" || event.Type == "response_ended" {
						t.Fatalf("invalid response published %s", event.Type)
					}
				}
			})
		}
	}
}
