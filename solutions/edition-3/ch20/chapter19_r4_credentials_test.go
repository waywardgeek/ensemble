package ensemble_test

import (
	"context"
	"encoding/json"
	"ensemble"
	"ensemble/internal/common"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestCh19R4OverlappingRenewalPublication(t *testing.T) {
	for _, old := range []string{"fixture-overlap", "RED"} {
		for _, stream := range []bool{false, true} {
			t.Run(fmt.Sprintf("old=%s/stream=%t", old, stream), func(t *testing.T) {
				renewed := old + "ACTED-renewed"
				auth := &ch19RenewAuth{old: old, new: renewed, renew: true}
				var requests atomic.Int32
				argsValue, _ := json.Marshal(renewed)
				args := fmt.Sprintf(`{"echo":%s,"keep":[1,2],"label":"ordinary"}`, argsValue)
				const wantArgs = `{"echo":"[REDACTED]","keep":[1,2],"label":"ordinary"}`
				const wantText = "ordinary [REDACTED] then [REDACTED] done"
				peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if requests.Add(1) == 1 {
						if r.Header.Get("Authorization") != "Bearer "+old {
							t.Error("wrong initial token")
						}
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					if r.Header.Get("Authorization") != "Bearer "+renewed {
						t.Error("wrong renewed token")
					}
					if !stream {
						w.Header().Set("Content-Type", "application/json")
						text, _ := json.Marshal("ordinary " + renewed + " then " + old + " done")
						name, _ := json.Marshal(renewed)
						wireArgs, _ := json.Marshal(args)
						fmt.Fprintf(w, `{"choices":[{"message":{"role":"assistant","content":%s,"tool_calls":[{"id":"overlap-call","type":"function","function":{"name":%s,"arguments":%s}}]},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`, text, name, wireArgs)
						return
					}
					w.Header().Set("Content-Type", "text/event-stream")
					frame := func(delta map[string]any) {
						raw, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"index": 0, "delta": delta}}})
						fmt.Fprintf(w, "data: %s\n\n", raw)
						w.(http.Flusher).Flush()
					}
					text := func(s string) { frame(map[string]any{"content": s}) }
					tool := func(name, arguments string) {
						frame(map[string]any{"tool_calls": []any{map[string]any{"index": 0, "id": "overlap-call", "type": "function", "function": map[string]any{"name": name, "arguments": arguments}}}})
					}
					text("ordinary ")
					text(old)
					text(renewed[len(old):] + " then ")
					text(old + " done")
					tool(old, "")
					tool(renewed[len(old):], args)
					fmt.Fprint(w, "data: {\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"tool_calls\"}],\"usage\":{\"prompt_tokens\":1,\"completion_tokens\":1}}\n\ndata: [DONE]\n\n")
				}))
				defer peer.Close()
				app := ensemble.New(nil)
				defer app.Shutdown()
				if err := app.SetAuth(auth); err != nil {
					t.Fatal(err)
				}
				capability := ensemble.StreamPlain
				if stream {
					capability = ensemble.StreamSSE
				}
				a, err := app.NewAgent(ensemble.Config{Model: "fixture", Vendor: ensemble.OpenAI, Surface: ensemble.ChatCompletions, AuthMode: ensemble.AuthAPI, BaseURL: peer.URL, BaseDir: t.TempDir(), StreamCapability: capability})
				if err != nil {
					t.Fatal(err)
				}
				var callbacks []common.Event
				var text, tool strings.Builder
				lastName := ""
				events, err := a.Engine().ExchangeRequest(context.Background(), []byte(fmt.Sprintf(`{"messages":[],"stream":%t}`, stream)), common.ParseCallbacks{OnEvent: func(e common.Event) { callbacks = append(callbacks, e) }, OnDelta: func(d common.PartDelta) {
					switch d.Kind {
					case common.DeltaText:
						text.WriteString(d.Chunk)
					case common.DeltaToolCall:
						tool.WriteString(d.Chunk)
						lastName = d.Name
					}
				}})
				if err != nil {
					t.Fatal(err)
				}
				for name, got := range map[string]string{"delta": text.String(), "OnEvent": ch19EventText(callbacks), "returned": ch19EventText(events)} {
					if got != wantText {
						t.Errorf("%s exact-content mismatch: got %q want %q", name, got, wantText)
					}
				}
				if got := tool.String(); got != "[REDACTED]"+wantArgs {
					t.Errorf("tool delta mismatch: %q", got)
				}
				if lastName != "[REDACTED]" {
					t.Errorf("delta Name mismatch: %q", lastName)
				}
				for label, list := range map[string][]common.Event{"OnEvent": callbacks, "returned": events} {
					found := false
					for _, event := range list {
						if event.Response != nil {
							for _, part := range event.Response.Parts {
								if part.Type == "tool_call" {
									found = true
									if part.Name != "[REDACTED]" || part.CallID != "overlap-call" || string(part.Args) != wantArgs {
										t.Errorf("%s complete tool changed/inconsistent: %+v", label, part)
									}
								}
							}
						}
					}
					if !found {
						t.Errorf("%s missing tool", label)
					}
				}
				if requests.Load() != 2 || !auth.rejected.Load() {
					t.Fatal("fixture did not exercise renewal")
				}
			})
		}
	}
}
