package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"ensemble/ensemble"
)

// The HTTP peer waits for real actor observations before sending more fragments.
// Delays between parts and after content must not inflate their own spans.
func TestArtifactProductionTiming(t *testing.T) {
	for _, disabled := range []bool{false, true} {
		t.Run(fmt.Sprintf("disabled=%t", disabled), func(t *testing.T) {
			observed := make(chan struct{}, 1)
			var requests atomic.Int32
			peer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				var body map[string]any
				if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
					t.Error(err)
					return
				}
				if (body["stream"] == true) == disabled {
					t.Error("unexpected streaming selection")
				}
				final := requests.Add(1) > 1
				plain, frames := streamFixture(ensemble.Anthropic, final)
				if disabled {
					json.NewEncoder(writer).Encode(plain)
					return
				}
				writer.Header().Set("Content-Type", "text/event-stream")
				if final {
					frames = []wireFrame{
						{"message_start", decoded(`{"type":"message_start","message":{"usage":{"input_tokens":8,"output_tokens":1}}}`)},
						{"content_block_start", decoded(`{"type":"content_block_start","index":0,"content_block":{"type":"thinking"}}`)},
						{"content_block_delta", decoded(`{"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"one burst"}}`)},
						{"content_block_stop", decoded(`{"type":"content_block_stop","index":0}`)},
						{"content_block_start", decoded(`{"type":"content_block_start","index":1,"content_block":{"type":"text"}}`)},
						{"content_block_delta", decoded(`{"type":"content_block_delta","index":1,"delta":{"type":"text_delta","text":"Done."}}`)},
						{"content_block_stop", decoded(`{"type":"content_block_stop","index":1}`)},
						{"message_stop", decoded(`{"type":"message_stop"}`)},
					}
				}
				for _, next := range frames {
					data := next.data.(map[string]any)
					delta, _ := data["delta"].(map[string]any)
					chunk, _ := delta["text"].(string)
					if thought, ok := delta["thinking"].(string); ok {
						chunk = thought
					}
					if chunk == "this" {
						time.Sleep(80 * time.Millisecond)
					} else if chunk == "ing." {
						time.Sleep(120 * time.Millisecond)
					}
					if block, ok := data["content_block"].(map[string]any); ok && block["type"] == "text" {
						time.Sleep(250 * time.Millisecond)
					}
					if next.event == "message_stop" {
						time.Sleep(150 * time.Millisecond)
					}
					frame(writer, next)
					if chunk != "" {
						select {
						case <-observed:
						case <-request.Context().Done():
							return
						}
					}
				}
			}))
			defer peer.Close()
			root := ensemble.New(io.Discard)
			agent, err := root.NewAgent(ensemble.Config{Model: "claude-sonnet-4-6", BaseURL: peer.URL, DataDir: t.TempDir(), DisableStreaming: disabled, Recall: ensemble.RecallSettings{Disabled: true}})
			if err != nil {
				t.Fatal(err)
			}
			defer agent.Shutdown()
			agent.RegisterTool(ensemble.ToolDeclaration{Name: "echo", Schema: json.RawMessage(`{"type":"object"}`)}, func(ensemble.ToolContext, json.RawMessage) (string, error) { return "complete", nil })
			arrivals := map[string]int64{}
			var parts []ensemble.Part
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			answer, err := agent.AskWatching(ctx, "use echo", ensemble.StreamCallbacks{
				OnDelta: func(_ uint64, kind ensemble.DeltaKind, chunk string) {
					if kind == ensemble.DeltaThinking || kind == ensemble.DeltaText {
						arrivals[chunk] = time.Now().UnixMilli()
						if !disabled {
							select {
							case observed <- struct{}{}:
							case <-ctx.Done():
							}
						}
					}
				},
				OnPartFinal: func(_ uint64, part ensemble.Part) { parts = append(parts, part) },
			})
			if err != nil || answer != "Done." {
				t.Fatalf("answer=%q error=%v", answer, err)
			}
			wantParts := 5
			if disabled {
				wantParts = 4
			}
			if len(parts) != wantParts {
				t.Fatalf("got %d parts, want %d", len(parts), wantParts)
			}
			if disabled {
				for _, part := range parts {
					if part.Started != 0 || part.Elapsed != 0 {
						t.Fatal("unstreamed part invented a duration", part)
					}
				}
			} else {
				for index, chunks := range [][2]string{{"consider ", "this"}, {"Check", "ing."}} {
					part := parts[index]
					want := arrivals[chunks[1]] - arrivals[chunks[0]]
					if absMilliseconds(part.Started-arrivals[chunks[0]]) > 50 || part.Elapsed < 50 || absMilliseconds(part.Elapsed-want) > 50 {
						t.Errorf("part %d span=%d start=%d, observed fragment span=%d", index, part.Elapsed, part.Started, want)
					}
				}
				for _, part := range parts[3:] {
					if part.Started != 0 || part.Elapsed != 0 {
						t.Error("single-burst response reused or invented a span", part)
					}
				}
			}
			var saved bytes.Buffer
			if err := agent.History().Dump(&saved); err != nil {
				t.Fatal(err)
			}
			restored, err := root.NewAgent(agent.Config())
			if err != nil {
				t.Fatal(err)
			}
			defer restored.Shutdown()
			if err := restored.History().Load(bytes.NewReader(saved.Bytes())); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(agent.History().EventsAfter(0), restored.History().EventsAfter(0)) {
				t.Fatal("history replay changed timing facts")
			}
			wire, err := agent.Engine().Render(agent.History().Context())
			if err != nil || bytes.Contains(wire, []byte(`"elapsed"`)) || bytes.Contains(wire, []byte(`"started"`)) {
				t.Fatalf("display timing leaked onto provider wire: %v", err)
			}
		})
	}
}

func absMilliseconds(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}
