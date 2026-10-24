package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ensemble"
)

// Exercise the public no-argument adapter, real Agent, and HTTP parser. Both
// model responses retain native text part ID zero. Public display IDs are
// lifetime identities; the actual in_flight boundaries must still precede them.
func TestCh6R2ProtocolResponseBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name, model, disabled string
		streaming             bool
	}{
		{"stream", "gpt-5.4", "false", true},
		{"disabled_plain", "gpt-5.4", "true", false},
		{"unknown_plain", "unknown-local-model", "false", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			if err := os.WriteFile("input.txt", []byte("fixture content"), 0600); err != nil {
				t.Fatal(err)
			}
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n := requests.Add(1)
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
					http.Error(w, "read failure", 500)
					return
				}
				var request struct {
					Stream bool `json:"stream"`
				}
				if err := json.Unmarshal(body, &request); err != nil {
					t.Error(err)
					http.Error(w, "bad JSON", 400)
					return
				}
				if request.Stream != tc.streaming {
					t.Errorf("stream=%v, want %v", request.Stream, tc.streaming)
				}
				if n > 2 {
					t.Errorf("unexpected request %d", n)
					http.Error(w, "unexpected request", 500)
					return
				}
				if n == 2 && (!bytes.Contains(body, []byte(`"role":"tool"`)) || !bytes.Contains(body, []byte(`"tool_call_id":"read1"`))) {
					t.Errorf("second request lacks paired tool result: %s", body)
				}
				if !tc.streaming {
					w.Header().Set("Content-Type", "application/json")
					if n == 1 {
						fmt.Fprint(w, `{"model":"actual","choices":[{"message":{"content":"Looking up.","tool_calls":[{"id":"read1","type":"function","function":{"name":"read_file","arguments":"{\"path\":\"input.txt\"}"}}]},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":3,"completion_tokens":2}}`)
					} else {
						fmt.Fprint(w, `{"model":"actual","choices":[{"message":{"content":"done answer"},"finish_reason":"stop"}],"usage":{"prompt_tokens":4,"completion_tokens":2}}`)
					}
					return
				}
				w.Header().Set("Content-Type", "text/event-stream")
				if n == 1 {
					fmt.Fprint(w, "data: "+`{"model":"actual","choices":[{"delta":{"content":"Looking "}}]}`+"\n\n")
					w.(http.Flusher).Flush()
					fmt.Fprint(w, "data: "+`{"choices":[{"delta":{"content":"up."}}]}`+"\n\n")
					fmt.Fprint(w, "data: "+`{"choices":[{"delta":{"tool_calls":[{"index":0,"id":"read1","type":"function","function":{"name":"read_file","arguments":"{\"path\":\"input.txt\"}"}}]},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":3,"completion_tokens":2}}`+"\n\n")
				} else {
					fmt.Fprint(w, "data: "+`{"model":"actual","choices":[{"delta":{"content":"done "}}]}`+"\n\n")
					w.(http.Flusher).Flush()
					fmt.Fprint(w, "data: "+`{"choices":[{"delta":{"content":"answer"},"finish_reason":"stop"}],"usage":{"prompt_tokens":4,"completion_tokens":2}}`+"\n\n")
				}
				fmt.Fprint(w, "data: [DONE]\n\n")
			}))
			defer server.Close()
			t.Setenv("LLM_VENDOR", "openai")
			t.Setenv("LLM_MODEL", tc.model)
			t.Setenv("OPENAI_API_KEY", "dummy")
			t.Setenv("OPENAI_BASE_URL", server.URL)
			t.Setenv("EN_DISABLE_STREAMING", tc.disabled)
			t.Setenv("LLM_TOOLS", "read_file")
			var out, diagnostics bytes.Buffer
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := Run(ctx, nil, strings.NewReader("{\"kind\":\"prompt\",\"text\":\"Read input.txt then answer\"}\n"), &out, &diagnostics); err != nil {
				t.Fatalf("Run: %v; diagnostics=%s", err, &diagnostics)
			}
			if requests.Load() != 2 {
				t.Fatalf("requests=%d", requests.Load())
			}
			decoder := json.NewDecoder(&out)
			boundaries, answers, modelFinals, toolResults := uint64(0), 0, 0, 0
			type textKey struct{ response, display uint64 }
			textIDs := map[uint64]uint64{}
			textDeltas, textFinals := map[textKey]string{}, map[textKey]string{}
			seenFinals := map[textKey]bool{}
			for {
				var record struct {
					Observation string             `json:"observation"`
					From        string             `json:"from"`
					To          string             `json:"to"`
					ResponseID  uint64             `json:"response_id"`
					PartID      *uint64            `json:"part_id"`
					Kind        ensemble.DeltaKind `json:"kind"`
					Chunk       string             `json:"chunk"`
					Part        ensemble.Part      `json:"part"`
					Assistant   *string            `json:"assistant"`
				}
				if err := decoder.Decode(&record); err == io.EOF {
					break
				} else if err != nil {
					t.Fatal(err)
				}
				if record.Observation == "state_changed" && record.To == "in_flight" {
					if record.From == "" {
						t.Fatal("missing real transition source")
					}
					boundaries++
				}
				if record.Observation == "part_delta" || record.Observation == "part_final" {
					if record.Part.Type == "tool_result" {
						toolResults++
						continue
					}
					if record.ResponseID == 0 || record.ResponseID != boundaries {
						t.Fatalf("%s for response %d before matching in_flight boundary (seen %d)", record.Observation, record.ResponseID, boundaries)
					}
					if record.PartID == nil {
						t.Fatal("missing public display ID")
					}
					key := textKey{record.ResponseID, *record.PartID}
					if record.Observation == "part_delta" && record.Kind == ensemble.DeltaText {
						if id, exists := textIDs[record.ResponseID]; exists && id != *record.PartID {
							t.Fatal("text display ID changed within response")
						}
						textIDs[record.ResponseID] = *record.PartID
						textDeltas[key] += record.Chunk
					}
					if record.Observation == "part_final" {
						modelFinals++
						if record.Part.DisplayID != *record.PartID {
							t.Fatal("final display ID changed")
						}
						if seenFinals[key] {
							t.Fatal("duplicate final")
						}
						seenFinals[key] = true
						if record.Part.Type == "text" {
							id, exists := textIDs[record.ResponseID]
							if !exists || id != *record.PartID || record.Part.ID != 0 {
								t.Fatal("final text lost display identity or native ID zero")
							}
							textFinals[key] = record.Part.Text
						}
					}
				}
				if record.Assistant != nil {
					answers++
					if *record.Assistant != "done answer" || textFinals[textKey{2, textIDs[2]}] != "done answer" {
						t.Fatalf("answer before matching final: %q, finals=%v", *record.Assistant, textFinals)
					}
				}
			}
			if boundaries != 2 || answers != 1 || modelFinals != 3 || toolResults != 1 {
				t.Fatalf("boundaries=%d answers=%d model finals=%d tool results=%d", boundaries, answers, modelFinals, toolResults)
			}
			if len(textIDs) != 2 || textIDs[1] == textIDs[2] {
				t.Fatalf("response display IDs not distinct: %v", textIDs)
			}
			for response, text := range map[uint64]string{1: "Looking up.", 2: "done answer"} {
				key := textKey{response, textIDs[response]}
				if textDeltas[key] != text || textFinals[key] != text {
					t.Fatalf("response %d display ID %d: deltas=%q finals=%q", response, key.display, textDeltas[key], textFinals[key])
				}
			}
		})
	}
}
