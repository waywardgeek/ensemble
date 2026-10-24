package ensemble_test

import (
	"context"
	"encoding/json"
	"ensemble"
	"ensemble/internal/common"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const ch19R3Token = "fixture-Credential-603"

func ch19R3Frames(w http.ResponseWriter, toolFirst bool) {
	w.Header().Set("Content-Type", "text/event-stream")
	frame := func(delta map[string]any) {
		raw, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"index": 0, "delta": delta}}})
		fmt.Fprintf(w, "data: %s\n\n", raw)
		w.(http.Flusher).Flush()
	}
	text := func(s string) { frame(map[string]any{"content": s}) }
	tool := func(name, args string) {
		frame(map[string]any{"tool_calls": []any{map[string]any{"index": 0, "id": "local-call", "type": "function", "function": map[string]any{"name": name, "arguments": args}}}})
	}
	if toolFirst {
		tool("fi", "")
		text("I will inspect safely.")
		tool("xture_probe", `{}`)
	} else {
		text("fi")
		tool("local_probe", `{}`)
		text("rst, I will inspect safely.")
	}
	fmt.Fprint(w, "data: {\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"tool_calls\"}],\"usage\":{\"prompt_tokens\":1,\"completion_tokens\":1}}\n\ndata: [DONE]\n\n")
}
func TestCh19R3InterleavedDeltaOrder(t *testing.T) {
	for _, toolFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("tool-first=%t", toolFirst), func(t *testing.T) {
			peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ch19R3Frames(w, toolFirst) }))
			defer peer.Close()
			app := ensemble.New(nil)
			defer app.Shutdown()
			_ = app.SetAuth(ch19Auth{token: ch19R3Token})
			a, err := app.NewAgent(ensemble.Config{Model: "fixture", Vendor: ensemble.OpenAI, Surface: ensemble.ChatCompletions, AuthMode: ensemble.AuthAPI, BaseURL: peer.URL, BaseDir: t.TempDir(), StreamCapability: ensemble.StreamSSE})
			if err != nil {
				t.Fatal(err)
			}
			var deltas []common.PartDelta
			events, err := a.Engine().ExchangeRequest(context.Background(), []byte(`{"messages":[],"stream":true}`), common.ParseCallbacks{OnDelta: func(d common.PartDelta) {
				if d.Chunk != "" {
					deltas = append(deltas, d)
				}
			}})
			if err != nil {
				t.Fatal(err)
			}
			want := []common.PartDelta{{Kind: common.DeltaText, PartID: 0, Chunk: "fi"}, {Kind: common.DeltaToolCall, PartID: 1, Name: "local_probe", Chunk: "local_probe"}, {Kind: common.DeltaToolCall, PartID: 1, Name: "local_probe", Chunk: `{}`}, {Kind: common.DeltaText, PartID: 0, Chunk: "rst, I will inspect safely."}}
			toolName := "local_probe"
			if toolFirst {
				toolName = "fixture_probe"
				want = []common.PartDelta{{Kind: common.DeltaToolCall, PartID: 0, Name: "fi", Chunk: "fi"}, {Kind: common.DeltaText, PartID: 1, Chunk: "I will inspect safely."}, {Kind: common.DeltaToolCall, PartID: 0, Name: "fixture_probe", Chunk: "xture_probe"}, {Kind: common.DeltaToolCall, PartID: 0, Name: "fixture_probe", Chunk: `{}`}}
			}
			if !reflect.DeepEqual(deltas, want) {
				t.Errorf("arrival metadata/order changed:\n got %#v\nwant %#v", deltas, want)
			}
			found := false
			for _, event := range events {
				if event.Response != nil {
					for _, p := range event.Response.Parts {
						if p.Type == "tool_call" {
							found = true
							if p.Name != toolName || string(p.Args) != `{}` {
								t.Errorf("final call corrupted: %+v", p)
							}
						}
					}
				}
			}
			if !found {
				t.Error("missing final tool call")
			}
		})
	}
}
func TestCh19R3ExplanationArrivalPolicy(t *testing.T) {
	for _, toolFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("tool-first=%t", toolFirst), func(t *testing.T) {
			var requests, effects atomic.Int32
			peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if requests.Add(1) == 1 {
					ch19R3Frames(w, toolFirst)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{"choices":[{"message":{"role":"assistant","content":"Done."},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`)
			}))
			defer peer.Close()
			app := ensemble.New(nil)
			defer app.Shutdown()
			_ = app.SetAuth(ch19Auth{token: ch19R3Token})
			a, err := app.NewAgent(ensemble.Config{Model: "fixture", Vendor: ensemble.OpenAI, Surface: ensemble.ChatCompletions, AuthMode: ensemble.AuthAPI, BaseURL: peer.URL, BaseDir: t.TempDir(), DataDir: t.TempDir(), StreamCapability: ensemble.StreamSSE, RequireToolExplanation: true, Tools: []string{}})
			if err != nil {
				t.Fatal(err)
			}
			name := "local_probe"
			if toolFirst {
				name = "fixture_probe"
			}
			err = a.RegisterExternalTool(ensemble.ToolDefinition{Name: name, Description: "In-process fixture only", Parameters: json.RawMessage(`{"type":"object"}`)}, func(ctx context.Context, call ensemble.ToolCall, args json.RawMessage) error {
				if string(args) != `{}` {
					t.Errorf("args changed: %s", args)
				}
				effects.Add(1)
				_, err := io.WriteString(call, "fixture result")
				return err
			})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			reply, err := a.Ask(ctx, "Use the in-process fixture.")
			if err != nil || reply != "Done." {
				t.Fatalf("turn result %q: %v", reply, err)
			}
			want := int32(1)
			if toolFirst {
				want = 0
			}
			if effects.Load() != want {
				t.Errorf("arrival policy effects=%d want=%d", effects.Load(), want)
			}
			if requests.Load() != 2 {
				t.Errorf("unexpected request count %d", requests.Load())
			}
		})
	}
}

func TestCh19R3InterleavedSecretDelivery(t *testing.T) {
	for _, toolFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("tool-first=%t", toolFirst), func(t *testing.T) {
			peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				frame := func(delta map[string]any) {
					raw, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"index": 0, "delta": delta}}})
					fmt.Fprintf(w, "data: %s\n\n", raw)
					w.(http.Flusher).Flush()
				}
				text := func(s string) { frame(map[string]any{"content": s}) }
				tool := func(name, args string) {
					frame(map[string]any{"tool_calls": []any{map[string]any{"index": 0, "id": "secret-call", "type": "function", "function": map[string]any{"name": name, "arguments": args}}}})
				}
				if toolFirst {
					tool(ch19R3Token[:8], "")
					text("ordinary explanation")
					tool(ch19R3Token[8:], `{}`)
				} else {
					text(ch19R3Token[:8])
					tool("local_probe", `{}`)
					text(ch19R3Token[8:] + " tail")
				}
				fmt.Fprint(w, "data: {\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"tool_calls\"}],\"usage\":{\"prompt_tokens\":1,\"completion_tokens\":1}}\n\ndata: [DONE]\n\n")
			}))
			defer peer.Close()
			app := ensemble.New(nil)
			defer app.Shutdown()
			_ = app.SetAuth(ch19Auth{token: ch19R3Token})
			a, err := app.NewAgent(ensemble.Config{Model: "fixture", Vendor: ensemble.OpenAI, Surface: ensemble.ChatCompletions, AuthMode: ensemble.AuthAPI, BaseURL: peer.URL, BaseDir: t.TempDir(), StreamCapability: ensemble.StreamSSE})
			if err != nil {
				t.Fatal(err)
			}
			var deltas []common.PartDelta
			var chunks []string
			events, err := a.Engine().ExchangeRequest(context.Background(), []byte(`{"messages":[],"stream":true}`), common.ParseCallbacks{OnDelta: func(d common.PartDelta) {
				deltas = append(deltas, d)
				if d.Chunk != "" {
					chunks = append(chunks, d.Chunk)
				}
			}})
			if err != nil {
				t.Fatal(err)
			}
			want := []string{"[REDACTED]", "local_probe", `{}`, " tail"}
			if toolFirst {
				want = []string{"[REDACTED]", "ordinary explanation", `{}`}
			}
			if !reflect.DeepEqual(chunks, want) {
				t.Errorf("redaction moved across an intervening part: got %q want %q", chunks, want)
			}
			for _, value := range []any{deltas, events} {
				raw, _ := json.Marshal(value)
				if strings.Contains(string(raw), ch19R3Token) {
					t.Error("exact credential reached public stream/event")
				}
			}
		})
	}
}

func TestCh19R3TerminalPrefixFlush(t *testing.T) {
	for _, mode := range []string{"complete", "error", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				fmt.Fprint(w, "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"fi\"}}]}\n\n")
				w.(http.Flusher).Flush()
				fmt.Fprint(w, "data: {\"choices\":[{\"index\":0,\"delta\":{\"tool_calls\":[{\"index\":0,\"id\":\"end-call\",\"type\":\"function\",\"function\":{\"name\":\"local_probe\",\"arguments\":\"{}\"}}]}}]}\n\n")
				w.(http.Flusher).Flush()
				if mode == "cancel" {
					<-r.Context().Done()
					return
				}
				if mode == "error" {
					fmt.Fprint(w, "data: {\"error\":{\"message\":\"fixture terminal failure\"}}\n\n")
					return
				}
				fmt.Fprint(w, "data: {\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"tool_calls\"}],\"usage\":{\"prompt_tokens\":1,\"completion_tokens\":1}}\n\ndata: [DONE]\n\n")
			}))
			defer peer.Close()
			app := ensemble.New(nil)
			defer app.Shutdown()
			_ = app.SetAuth(ch19Auth{token: ch19R3Token})
			a, err := app.NewAgent(ensemble.Config{Model: "fixture", Vendor: ensemble.OpenAI, Surface: ensemble.ChatCompletions, AuthMode: ensemble.AuthAPI, BaseURL: peer.URL, BaseDir: t.TempDir(), StreamCapability: ensemble.StreamSSE})
			if err != nil {
				t.Fatal(err)
			}
			var chunks []string
			frames := 0
			_, err = a.Engine().ExchangeRequest(ctx, []byte(`{"messages":[],"stream":true}`), common.ParseCallbacks{OnDelta: func(d common.PartDelta) {
				if d.Chunk != "" {
					chunks = append(chunks, d.Chunk)
				}
			}, OnFrame: func(frame common.Frame) {
				frames++
				if len(chunks) != 0 {
					t.Errorf("non-resuming ambiguous prefix was bypassed before parse completion: %q", chunks)
				}
				if mode == "cancel" && frames == 2 {
					cancel()
				}
			}})
			if (err == nil) != (mode == "complete") {
				t.Fatalf("completion/error contract mode=%s err=%v", mode, err)
			}
			if frames < 2 {
				t.Fatal("fixture did not reach both interleaved parts")
			}
			if want := []string{"fi", "local_probe", `{}`}; !reflect.DeepEqual(chunks, want) {
				t.Errorf("terminal flush changed arrival order: got %q want %q", chunks, want)
			}
		})
	}
}
