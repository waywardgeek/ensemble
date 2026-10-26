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
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ensemble/ensemble"
)

type wireFrame struct {
	event string
	data  any
}

func frame(w http.ResponseWriter, f wireFrame) {
	raw, _ := json.Marshal(f.data)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", f.event, raw)
	w.(http.Flusher).Flush()
}
func decoded(s string) any {
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		panic(err)
	}
	return v
}

// Fixtures follow the vendor stream references cited in ch07.md. Both delivery
// modes describe one thought/text/call response and the same final reply. Tool
// fragments deliberately precede OpenAI text, where eventual list positions fail.
func streamFixture(v ensemble.Vendor, final bool) (any, []wireFrame) {
	if v == ensemble.Anthropic {
		blocks := []any{decoded(`{"type":"thinking","thinking":"consider this","signature":"signed"}`), decoded(`{"type":"text","text":"Checking."}`), decoded(`{"type":"tool_use","id":"c1","name":"echo","input":{"value":"complete"}}`)}
		if final {
			blocks = []any{decoded(`{"type":"text","text":"Done."}`)}
		}
		response := map[string]any{"model": "claude-sonnet-4-6", "content": blocks, "usage": decoded(`{"input_tokens":8,"output_tokens":4}`)}
		frames := []wireFrame{{"message_start", decoded(`{"type":"message_start","message":{"model":"claude-sonnet-4-6","usage":{"input_tokens":8,"output_tokens":1}}}`)}}
		for i, b := range blocks {
			block := b.(map[string]any)
			typ := block["type"].(string)
			start := map[string]any{"type": typ}
			if typ == "tool_use" {
				start["id"], start["name"], start["input"] = "c1", "echo", map[string]any{}
			}
			frames = append(frames, wireFrame{"content_block_start", map[string]any{"type": "content_block_start", "index": i, "content_block": start}})
			field, dtype, chunks := "text", "text_delta", []string{"Check", "ing."}
			if final {
				chunks = []string{"Do", "ne."}
			}
			if typ == "thinking" {
				field, dtype, chunks = "thinking", "thinking_delta", []string{"consider ", "this"}
			}
			if typ == "tool_use" {
				field, dtype, chunks = "partial_json", "input_json_delta", []string{`{"value":`, `"complete"}`}
			}
			for _, chunk := range chunks {
				frames = append(frames, wireFrame{"content_block_delta", map[string]any{"type": "content_block_delta", "index": i, "delta": map[string]any{"type": dtype, field: chunk}}})
			}
			if typ == "thinking" {
				frames = append(frames, wireFrame{"content_block_delta", map[string]any{"type": "content_block_delta", "index": i, "delta": decoded(`{"type":"signature_delta","signature":"signed"}`)}})
			}
			frames = append(frames, wireFrame{"content_block_stop", map[string]any{"type": "content_block_stop", "index": i}})
		}
		return response, append(frames, wireFrame{"message_delta", decoded(`{"type":"message_delta","usage":{"output_tokens":4}}`)}, wireFrame{"message_stop", decoded(`{"type":"message_stop"}`)})
	}
	if v == ensemble.OpenAI {
		message := decoded(`{"content":"Checking.","tool_calls":[{"id":"c1","type":"function","function":{"name":"echo","arguments":"{\"value\":\"complete\"}"}}]}`)
		deltas := []any{decoded(`{"tool_calls":[{"index":0,"id":"c1","function":{"name":"echo","arguments":"{\"value\":"}}]}`), decoded(`{"content":"Check"}`), decoded(`{"tool_calls":[{"index":0,"function":{"arguments":"\"complete\"}"}}]}`), decoded(`{"content":"ing."}`)}
		if final {
			message = decoded(`{"content":"Done."}`)
			deltas = []any{decoded(`{"content":"Do"}`), decoded(`{"content":"ne."}`)}
		}
		usage := decoded(`{"prompt_tokens":8,"completion_tokens":4,"prompt_tokens_details":{"cached_tokens":2}}`)
		response := map[string]any{"model": "gpt-4o-mini", "choices": []any{map[string]any{"message": message}}, "usage": usage}
		var frames []wireFrame
		for _, d := range deltas {
			frames = append(frames, wireFrame{"", map[string]any{"model": "gpt-4o-mini", "choices": []any{map[string]any{"delta": d}}}})
		}
		return response, append(frames, wireFrame{"", decoded(`{"choices":[{"delta":{},"finish_reason":"stop"}]}`)}, wireFrame{"", map[string]any{"choices": []any{}, "usage": usage}})
	}
	parts := []any{decoded(`{"text":"consider this","thought":true,"thoughtSignature":"signed"}`), decoded(`{"text":"Checking.","thoughtSignature":"text-signed"}`), decoded(`{"functionCall":{"id":"c1","name":"echo","args":{"value":"complete"}},"thoughtSignature":"call-signed"}`)}
	pieces := []any{decoded(`{"text":"consider ","thought":true}`), decoded(`{"text":"this","thought":true,"thoughtSignature":"signed"}`), decoded(`{"text":"Check"}`), decoded(`{"text":"ing."}`), decoded(`{"text":"","thoughtSignature":"text-signed"}`), parts[2]}
	if final {
		parts = []any{decoded(`{"text":"Done."}`)}
		pieces = []any{decoded(`{"text":"Do"}`), decoded(`{"text":"ne."}`)}
	}
	usage := decoded(`{"promptTokenCount":8,"candidatesTokenCount":4,"thoughtsTokenCount":3}`)
	response := map[string]any{"modelVersion": "gemini-2.5-flash", "candidates": []any{map[string]any{"content": map[string]any{"parts": parts}}}, "usageMetadata": usage}
	var frames []wireFrame
	for i, p := range pieces {
		candidate := map[string]any{"content": map[string]any{"parts": []any{p}}}
		f := map[string]any{"modelVersion": "gemini-2.5-flash", "candidates": []any{candidate}}
		if i == len(pieces)-1 {
			candidate["finishReason"] = "STOP"
			f["usageMetadata"] = usage
		}
		frames = append(frames, wireFrame{"", f})
	}
	return response, frames
}

// TestStreamingDeliveryAndContent checks real vendor delivery through the actor:
// chunks arrive before completion, IDs retain their kinds and content, tools wait
// for complete arguments, and changing delivery preserves finalized facts/replay.
func TestStreamingDeliveryAndContent(t *testing.T) {
	for _, vendor := range []ensemble.Vendor{ensemble.Anthropic, ensemble.OpenAI, ensemble.Gemini} {
		t.Run(fmt.Sprint(vendor), func(t *testing.T) {
			var baseline []ensemble.Event
			counts := []int{}
			for _, disabled := range []bool{false, true} {
				t.Run(fmt.Sprint("disabled=", disabled), func(t *testing.T) {
					first := make(chan struct{})
					release := make(chan struct{})
					var once sync.Once
					defer once.Do(func() { close(release) })
					var requests atomic.Int32
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						var body map[string]any
						json.NewDecoder(r.Body).Decode(&body)
						streaming := body["stream"] == true || r.URL.Query().Get("alt") == "sse"
						if streaming == disabled {
							t.Errorf("request ignored effective streaming: %v", body)
						}
						if streaming && vendor == ensemble.OpenAI && !reflect.DeepEqual(body["stream_options"], map[string]any{"include_usage": true}) {
							t.Error("streamed usage was not requested")
						}
						if vendor == ensemble.Gemini && streaming && !strings.HasSuffix(r.URL.Path, ":streamGenerateContent") {
							t.Error("wrong Gemini streaming endpoint", r.URL.Path)
						}
						requestNumber := requests.Add(1)
						plain, frames := streamFixture(vendor, requestNumber > 1)
						if !streaming {
							json.NewEncoder(w).Encode(plain)
							return
						}
						w.Header().Set("Content-Type", "text/event-stream")
						for i, f := range frames {
							frame(w, f)
							// Stop the provider before finalization. The test cannot release it
							// until the real actor has delivered at least one actual fragment.
							gate := map[ensemble.Vendor]int{ensemble.Anthropic: 11, ensemble.OpenAI: 0, ensemble.Gemini: 2}[vendor]
							if requestNumber == 1 && i == gate {
								select {
								case <-release:
								case <-r.Context().Done():
									return
								}
							}
						}
						if vendor == ensemble.OpenAI {
							fmt.Fprint(w, "data: [DONE]\n\n")
						}
					}))
					defer server.Close()
					model := map[ensemble.Vendor]string{ensemble.Anthropic: "claude-sonnet-4-6", ensemble.OpenAI: "gpt-4o-mini", ensemble.Gemini: "gemini-2.5-flash"}[vendor]
					var trace bytes.Buffer
					app := ensemble.New(io.Discard)
					app.SetAPILog(&trace)
					a, err := app.NewAgent(ensemble.Config{Vendor: vendor, Model: model, BaseURL: server.URL, DataDir: t.TempDir()})
					if err != nil {
						t.Fatal(err)
					}
					// Configuration belongs to Agent, so set the opt-out before construction.
					if disabled {
						a.Shutdown()
						a, err = app.NewAgent(ensemble.Config{Vendor: vendor, Model: model, BaseURL: server.URL, DataDir: t.TempDir(), DisableStreaming: true})
						if err != nil {
							t.Fatal(err)
						}
					}
					defer a.Shutdown()
					var executed atomic.Bool
					a.RegisterTool(ensemble.ToolDeclaration{Name: "echo", Schema: json.RawMessage(`{"type":"object","properties":{"value":{"type":"string"}}}`)}, func(_ ensemble.ToolContext, args json.RawMessage) (string, error) {
						if string(args) != `{"value":"complete"}` {
							t.Errorf("executed incomplete arguments: %s", args)
						}
						executed.Store(true)
						return "complete", nil
					})
					var events []ensemble.Event
					chunks := map[uint64]string{}
					kinds := map[uint64]ensemble.DeltaKind{}
					deltas := 0
					signaled := false
					callbacks := ensemble.StreamCallbacks{
						OnDelta: func(id uint64, kind ensemble.DeltaKind, chunk string) {
							if old, ok := kinds[id]; ok && old != kind {
								t.Error("part changed kind", id)
							}
							kinds[id] = kind
							chunks[id] += chunk
							deltas++
							// A fragment of arguments has reached the observer while the
							// provider still withholds the closing JSON and response end.
							ready := disabled || (vendor == ensemble.Gemini && kind == ensemble.DeltaText) || (kind == ensemble.DeltaToolCall && chunk == `{"value":`)
							if !signaled && ready {
								signaled = true
								close(first)
							}
						},
						OnEvent: func(event ensemble.Event) {
							if event.Response != nil {
								events = append(events, event)
							}
						},
						OnPartFinal: func(id uint64, p ensemble.Part) {
							c := a.History().Context()
							if len(c.Dialogue) == 0 || len(events) == 0 {
								t.Error("final preceded stored response")
								return
							}
							last := events[len(events)-1].Response
							found := false
							for _, stored := range last.Parts {
								if reflect.DeepEqual(stored, p) {
									found = true
								}
							}
							if !found {
								t.Error("final content absent from event")
							}
							want := ""
							wantKind := ensemble.DeltaText
							switch p.Type {
							case "text":
								want = p.Text
							case "tool_call":
								wantKind = ensemble.DeltaToolCall
								want = p.Name + " " + string(p.Args)
							case "opaque":
								wantKind = ensemble.DeltaThinking
								want = "consider this"
							}
							if kinds[id] != wantKind {
								t.Errorf("part %d channel %d != expected %d", id, kinds[id], wantKind)
							}
							if chunks[id] != want {
								t.Errorf("part %d deltas %q != final %q", id, chunks[id], want)
							}
							delete(chunks, id)
							delete(kinds, id)
						},
					}
					result := make(chan error, 1)
					go func() {
						answer, err := a.AskWatching(context.Background(), "use echo", callbacks)
						if err == nil && answer != "Done." {
							err = fmt.Errorf("reply contaminated: %q", answer)
						}
						result <- err
					}()
					if !disabled {
						select {
						case <-first:
						case <-time.After(3 * time.Second):
							once.Do(func() { close(release) })
							t.Fatal("no incremental observation while provider remained open")
						}
						if executed.Load() {
							t.Error("tool executed before response completion")
						}
						once.Do(func() { close(release) })
					}
					if err := <-result; err != nil {
						t.Fatal(err)
					}
					// This OpenAI SSE fixture deliberately delivers the call before
					// text. Keep observing/reconstructing it, but the flagged model
					// must not execute it. JSON has no streamed ordering violation.
					wantExecution := vendor != ensemble.OpenAI || disabled
					if executed.Load() != wantExecution {
						t.Fatalf("tool execution=%t, want %t", executed.Load(), wantExecution)
					}
					if !wantExecution {
						paired := false
						var dump bytes.Buffer
						if err := a.History().Dump(&dump); err != nil {
							t.Fatal(err)
						}
						decoder := json.NewDecoder(&dump)
						for {
							var e ensemble.Event
							if err := decoder.Decode(&e); err == io.EOF {
								break
							} else if err != nil {
								t.Fatal(err)
							}
							if e.Type == "tool_returned" && e.Tool.IsError && strings.Contains(e.Tool.Parts[0].Text, "arrived in the stream") {
								paired = true
							}
						}
						if !paired {
							t.Fatal("late explanation did not produce a paired stream-order rejection")
						}
					}
					if !bytes.Contains(trace.Bytes(), []byte(`"event":`)) || !bytes.Contains(trace.Bytes(), []byte("Done")) && !bytes.Contains(trace.Bytes(), []byte("ne.")) {
						t.Fatal("API trace lost response frames")
					}
					if len(events) != 2 {
						t.Fatal("missing response facts", len(events))
					}
					for i := range events {
						events[i].Seq = 0
						// Display identity follows arrival order, not provider content semantics.
						for j := range events[i].Response.Parts {
							events[i].Response.Parts[j].ArtifactID = 0
							// Arrival measurements differ by delivery mode, not content.
							events[i].Response.Parts[j].Started = 0
							events[i].Response.Parts[j].Elapsed = 0
						}

					}
					if baseline == nil {
						baseline = events
					} else if !reflect.DeepEqual(baseline, events) {
						t.Fatalf("delivery changed finalized facts\nstream: %#v\nplain: %#v", baseline, events)
					}
					var replay bytes.Buffer
					if err := a.History().Dump(&replay); err != nil {
						t.Fatal(err)
					}
					if bytes.Contains(replay.Bytes(), []byte("part_delta")) {
						t.Fatal("transient deltas entered history")
					}
					b, _ := app.NewAgent(a.Config())
					for _, d := range a.Declarations() {
						b.RegisterTool(d, func(ensemble.ToolContext, json.RawMessage) (string, error) { return "", nil })
					}
					defer b.Shutdown()
					if err := b.History().Load(bytes.NewReader(replay.Bytes())); err != nil {
						t.Fatal(err)
					}
					live, _ := a.Engine().Render(a.History().Context())
					rebuilt, _ := b.Engine().Render(b.History().Context())
					if !bytes.Equal(live, rebuilt) {
						t.Fatal("streamed history did not replay to identical request")
					}
					if a.Engine().Usage() != a.History().Context().Usage {
						t.Fatal("stream usage diverged from replay accounting")
					}
					counts = append(counts, deltas)
				})
			}
			if len(counts) == 2 && counts[0] <= counts[1] {
				t.Fatal("delivery did not become incremental", counts)
			}
		})
	}
}

// TestUnknownModelAndStreamFailures protects conservative capability fallback and
// rejects incomplete or unmeasured streams before they can publish final parts.
func TestUnknownModelAndStreamFailures(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["stream"] == true {
			t.Error("unknown model requested streaming")
		}
		json.NewEncoder(w).Encode(decoded(`{"content":[{"type":"text","text":"works"}],"usage":{"input_tokens":2,"output_tokens":1}}`))
	}))
	defer server.Close()
	app := ensemble.New(io.Discard)
	a, _ := app.NewAgent(ensemble.Config{Model: "unknown-future-model", BaseURL: server.URL, DataDir: t.TempDir()})
	defer a.Shutdown()
	if answer, err := a.Ask(context.Background(), "hello"); err != nil || answer != "works" || requests != 1 {
		t.Fatal("unknown text model refused", answer, err)
	}
	if ensemble.StreamingFor(a.Config()) != 0 || ensemble.StreamingFor(ensemble.Config{Model: "claude-sonnet-4-6", DisableStreaming: true}) != 0 {
		t.Fatal("capability and caller policy did not intersect")
	}
	if _, err := json.Marshal(ensemble.PartDelta{}); err == nil {
		t.Fatal("unset delta kind was silently accepted")
	}
	for _, tc := range []struct {
		name   string
		vendor ensemble.Vendor
		body   string
	}{
		{"anthropic error", ensemble.Anthropic, "event: error\ndata: {\"type\":\"error\",\"error\":{\"message\":\"untrusted\"}}\n\n"},
		{"anthropic truncated", ensemble.Anthropic, "data: {\"type\":\"message_start\",\"message\":{\"usage\":{\"input_tokens\":1,\"output_tokens\":1}}}\n\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"partial\"}}\n\n"},
		{"openai missing usage", ensemble.OpenAI, "data: {\"choices\":[{\"delta\":{\"content\":\"answer\"},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n"},
		{"gemini truncated", ensemble.Gemini, "data: {\"usageMetadata\":{\"promptTokenCount\":1},\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"partial\"}]}}]}\n\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, _ := app.NewAgent(ensemble.Config{Vendor: tc.vendor, Model: "parse-fixture", DataDir: t.TempDir()})
			defer a.Shutdown()
			finals, logged := 0, 0
			resp := &http.Response{Header: http.Header{"Content-Type": []string{"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(tc.body))}
			err := a.Engine().Parse(resp, ensemble.StreamCallbacks{OnFrame: func(string, []byte) { logged++ }, OnPartFinal: func(uint64, ensemble.Part) { finals++ }})
			if err == nil || finals != 0 || logged == 0 {
				t.Fatal("failed stream masqueraded as final or lost trace", err, finals, logged)
			}
		})
	}
}

// TestAPIFrameRedaction verifies that the real HTTP/owner logging path retains an
// error frame while removing a credential deliberately reflected by the peer.
func TestAPIFrameRedaction(t *testing.T) {
	// The peer deliberately reflects its credential in an error frame. Keep raw
	// diagnostics useful without letting provider-controlled text disclose the key.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		frame(w, wireFrame{"error", map[string]any{"type": "error", "error": map[string]string{"message": r.Header.Get("x-api-key")}}})
	}))
	defer server.Close()
	var trace, diagnostics bytes.Buffer
	app := ensemble.New(&diagnostics)
	app.SetAPILog(&trace)
	a, _ := app.NewAgent(ensemble.Config{Model: "claude-sonnet-4-6", BaseURL: server.URL, APIKey: "private-test-token", DataDir: t.TempDir()})
	defer a.Shutdown()
	_, err := a.Ask(context.Background(), "hello")
	if err == nil {
		t.Fatal("provider error accepted")
	}
	if strings.Contains(trace.String()+diagnostics.String()+err.Error(), "private-test-token") {
		t.Fatal("credential escaped into diagnostics")
	}
	if !strings.Contains(trace.String(), "[REDACTED]") || !strings.Contains(trace.String(), `"event":"error"`) {
		t.Fatal("sanitized error frame missing", trace.String())
	}
}
