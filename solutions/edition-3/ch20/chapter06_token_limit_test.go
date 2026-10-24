package ensemble_test

import (
	"context"
	"encoding/json"
	"ensemble"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Exercise the shared Render through actual HTTP, not a proxy rewrite or a
// hand-built request. The documented OpenAI replacement keeps the native cap
// and both delivery modes; other vendors retain their existing wire fields.
func TestCompletionTokenLimitHTTP(t *testing.T) {
	for _, tc := range []struct {
		name, model, plain, stream string
		vendor                     ensemble.Vendor
	}{
		{"gpt-5.4", "gpt-5.4", "", "", ensemble.OpenAI},
		{"gpt-5.4-mini", "gpt-5.4-mini", "", "", ensemble.OpenAI},
		{"legacy-openai", "gpt-4.1", "", "", ensemble.OpenAI},
		{"anthropic", "claude-sonnet-4-6", `{"model":"served-anthropic","content":[{"type":"text","text":"hello"}],"stop_reason":"end_turn","usage":{"input_tokens":23,"output_tokens":7}}`, "data: {\"type\":\"message_start\",\"message\":{\"model\":\"served-anthropic\",\"usage\":{\"input_tokens\":23}}}\n\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"hel\"}}\n\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"lo\"}}\n\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":7}}\n\ndata: {\"type\":\"message_stop\"}\n\n", ensemble.Anthropic},
		{"gemini", "gemini-3.8-flash", `{"modelVersion":"served-gemini","candidates":[{"content":{"parts":[{"text":"hello"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":23,"candidatesTokenCount":7}}`, "data: {\"modelVersion\":\"served-gemini\",\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"hel\"}]}}]}\n\ndata: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"lo\"}]},\"finishReason\":\"STOP\"}],\"usageMetadata\":{\"promptTokenCount\":23,\"candidatesTokenCount\":7}}\n\n", ensemble.Gemini},
	} {
		if tc.vendor == ensemble.OpenAI {
			// Identical wire response makes aliases/default/legacy requests
			// comparable without inferring producer identity from their names.
			tc.plain = `{"model":"served-openai","choices":[{"message":{"content":"hello"},"finish_reason":"stop"}],"usage":{"prompt_tokens":23,"completion_tokens":7}}`
			tc.stream = "data: {\"model\":\"served-openai\",\"choices\":[{\"delta\":{\"content\":\"hel\"}}]}\n\ndata: {\"choices\":[{\"delta\":{\"content\":\"lo\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":23,\"completion_tokens\":7}}\n\ndata: [DONE]\n\n"
		}
		for _, streaming := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/stream=%t", tc.name, streaming), func(t *testing.T) {
				requests := make(chan map[string]any, 1)
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Errorf("decode actual request: %v", err)
						http.Error(w, "bad request", http.StatusBadRequest)
						return
					}
					requests <- body
					if streaming {
						w.Header().Set("Content-Type", "text/event-stream")
						fmt.Fprint(w, tc.stream)
						w.(http.Flusher).Flush()
					} else {
						w.Header().Set("Content-Type", "application/json")
						fmt.Fprint(w, tc.plain)
					}
				}))
				defer server.Close()
				app := ensemble.New(nil)
				defer app.Shutdown()
				a, err := app.NewAgent(ensemble.Config{Vendor: tc.vendor, Model: tc.model, APIKey: "dummy", BaseURL: server.URL, Tools: []string{}, StreamCapability: ensemble.StreamSSE, DisableStreaming: !streaming})
				if err != nil {
					t.Fatal(err)
				}
				q := a.Observe()
				defer q.Close()
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				result := <-a.Submit(ctx, "hello")
				wantUsage := ensemble.Usage{Input: 23, Output: 7}
				if result.Err != nil || result.Text != "hello" || a.Engine().Usage() != wantUsage {
					t.Fatalf("reply=%+v engine usage=%+v", result, a.Engine().Usage())
				}
				responseCount := 0
				var savedTexts []ensemble.Part
				for _, event := range a.Log().Events() {
					if event.Type == "response_ended" && event.Response != nil {
						responseCount++
						for _, part := range event.Response.Parts {
							if part.Type == "text" {
								savedTexts = append(savedTexts, part)
							}
						}
						if event.Response.Usage != wantUsage {
							t.Errorf("saved response usage=%+v", event.Response.Usage)
						}
					}
				}
				if responseCount != 1 {
					t.Errorf("saved responses=%d", responseCount)
				}
				body := <-requests
				switch tc.vendor {
				case ensemble.OpenAI:
					_, legacyFieldPresent := body["max_tokens"]
					if body["max_completion_tokens"] != float64(1024) || legacyFieldPresent {
						t.Errorf("completion cap=%v wrong legacy field=%v", body["max_completion_tokens"], body["max_tokens"])
					}
					if streaming {
						opts, ok := body["stream_options"].(map[string]any)
						if !ok || len(opts) != 1 || opts["include_usage"] != true {
							t.Errorf("stream usage options=%v", body["stream_options"])
						}
					} else if body["stream_options"] != nil {
						t.Errorf("plain request has stream options: %v", body["stream_options"])
					}
				case ensemble.Anthropic:
					_, replacementPresent := body["max_completion_tokens"]
					if body["max_tokens"] != float64(1024) || replacementPresent {
						t.Errorf("Anthropic cap changed: %v", body)
					}
				case ensemble.Gemini:
					config, ok := body["generationConfig"].(map[string]any)
					_, legacyPresent := body["max_tokens"]
					_, replacementPresent := body["max_completion_tokens"]
					if !ok || config["maxOutputTokens"] != float64(1024) || legacyPresent || replacementPresent {
						t.Errorf("Gemini cap changed: %v", body)
					}
				}
				if tc.vendor != ensemble.Gemini {
					if body["model"] != tc.model || streaming && body["stream"] != true || !streaming && body["stream"] != nil {
						t.Errorf("model/stream changed: %v", body)
					}
				}
				var text strings.Builder
				var displayID uint64
				sawText, finals := false, 0
				for _, o := range q.Drain() {
					switch o := o.(type) {
					case ensemble.PartDelta:
						if o.Kind != ensemble.DeltaText {
							continue
						}
						if !sawText {
							displayID, sawText = o.PartID, true
						}
						if o.PartID != displayID {
							t.Errorf("text display ID changed: %d -> %d", displayID, o.PartID)
						}
						text.WriteString(o.Chunk)
					case ensemble.PartFinal:
						if o.Part.Type != "text" {
							continue
						}
						finals++
						if !sawText || o.PartID != displayID || o.Part.DisplayID != displayID || o.Part.ID != 0 || o.Part.Text != "hello" {
							t.Errorf("final does not match text display ID %d/native ID zero: %+v", displayID, o)
						}
					}
				}
				if !sawText || finals != 1 || text.String() != "hello" {
					t.Errorf("text observed=%t finals=%d delta content=%q", sawText, finals, text.String())
				}
				if len(savedTexts) != 1 {
					t.Fatalf("saved text parts=%d", len(savedTexts))
				}
				if part := savedTexts[0]; part.DisplayID != displayID || part.ID != 0 || part.Text != "hello" {
					t.Errorf("saved text does not match display ID %d/native ID zero: %+v", displayID, part)
				}
			})
		}
	}
}
