package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ensemble"
)

func TestCh16R2ResponseFraming(t *testing.T) {
	for _, tc := range []struct {
		name, media, prefix, body string
		status                    int
		stream, ok                bool
	}{
		{name: "absent SSE header", stream: true, ok: true},
		{name: "plain SSE header", media: "text/plain", prefix: ": keepalive\n\n", stream: true, ok: true},
		{name: "misleading JSON header", media: "application/json", prefix: "event: response.output_text.delta\n", stream: true, ok: true},
		{name: "complete JSON", media: "application/json", body: `{"status":"completed","model":"gpt-ch19-course","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"First."}]}],"usage":{"input_tokens":4,"output_tokens":1}}`, ok: true},
		{name: "headerless JSON", body: `  {"status":"completed","model":"gpt-ch19-course","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"First."}]}],"usage":{"input_tokens":4,"output_tokens":1}}`, ok: true},
		{name: "malformed JSON", media: "text/plain", body: `{"status":`},
		{name: "malformed SSE", media: "application/octet-stream", body: "data: not-json\n\n"},
		{name: "refused JSON", ok: true, media: "application/json", body: `{"status":"completed","model":"gpt-ch19-course","output":[{"type":"message","role":"assistant","content":[{"type":"refusal","refusal":"Not allowed."}]}],"usage":{"input_tokens":4,"output_tokens":1}}`},
		{name: "error envelope", body: `{"error":{"message":"Rejected."}}`},
		{name: "HTTP error SSE", status: 403, body: "data: {\"type\":\"response.output_text.delta\",\"delta\":\"must not appear\"}\n\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			first := make(chan struct{}, 1)
			beforeRemainder := make(chan bool, 1)
			peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/responses" || r.Header.Get("Authorization") != "Bearer dummy-r2" {
					t.Error("wrong HTTP route/credential")
				}
				if tc.media == "" {
					w.Header()["Content-Type"] = nil
				} else {
					w.Header().Set("Content-Type", tc.media)
				}
				if tc.status != 0 {
					w.WriteHeader(tc.status)
				}
				if !tc.stream {
					fmt.Fprint(w, tc.body)
					return
				}
				fmt.Fprint(w, tc.prefix+"data: {\"type\":\"response.output_text.delta\",\"output_index\":0,\"content_index\":0,\"delta\":\"First.\"}\n\n")
				w.(http.Flusher).Flush()
				// No remainder is available until a callback is observed. A whole-body
				// read deterministically reaches the bounded failure path instead.
				select {
				case <-first:
					beforeRemainder <- true
				case <-time.After(time.Second):
					beforeRemainder <- false
				case <-r.Context().Done():
					beforeRemainder <- false
					return
				}
				fmt.Fprint(w, "data: {\"type\":\"response.output_item.done\",\"output_index\":0,\"item\":{\"type\":\"message\",\"role\":\"assistant\",\"content\":[{\"type\":\"output_text\",\"text\":\"First.\"}]}}\n\n")
				fmt.Fprint(w, "data: {\"type\":\"response.completed\",\"response\":{\"status\":\"completed\",\"model\":\"gpt-ch19-course\",\"output\":[],\"usage\":{\"input_tokens\":4,\"output_tokens\":1}}}\n\n")
			}))
			defer peer.Close()
			app := ensemble.New(nil)
			defer app.Shutdown()
			a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-ch19-course", Surface: ensemble.Responses, AuthMode: ensemble.AuthAPI, APIKey: "dummy-r2", BaseURL: peer.URL, BaseDir: t.TempDir(), DataDir: t.TempDir(), Tools: []string{}})
			if err != nil {
				t.Fatal(err)
			}
			wire, err := a.Render()
			if err != nil {
				t.Fatal(err)
			}
			var deltas string
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			events, err := a.Engine().ExchangeRequest(ctx, wire, ensemble.ParseCallbacks{OnDelta: func(d ensemble.PartDelta) {
				if d.Kind == ensemble.DeltaText {
					deltas += d.Chunk
					select {
					case first <- struct{}{}:
					default:
					}
				}
			}})
			if tc.stream && !<-beforeRemainder {
				t.Error("no first callback before stream remainder")
			}
			if !tc.ok {
				if err == nil || deltas != "" {
					t.Fatalf("malformed/refused response accepted: err=%v deltas=%q", err, deltas)
				}
				return
			}
			wantText := "First."
			if tc.name == "refused JSON" {
				wantText = "Not allowed."
			}
			if err != nil || deltas != wantText {
				t.Fatalf("parse: err=%v deltas=%q", err, deltas)
			}

			ended := false
			for _, e := range events {
				if e.Type == ensemble.ResponseEnded {
					ended = true
					final := e.Response
					if final == nil || len(final.Parts) != 1 || final.Parts[0].Text != wantText || final.From.Model != "gpt-ch19-course" || final.Usage.Input != 4 || final.Usage.Output != 1 {
						t.Fatalf("wrong completed facts: %+v", final)
					}
				}
			}
			if !ended {
				t.Fatal("missing completed response")
			}

		})
	}
}

const ch16R2Constitution = "R2 constitution must occur exactly once."

func checkCh16R2CacheEnvelope(t *testing.T, body []byte, supported bool) string {
	t.Helper()
	var envelope struct {
		Model        string         `json:"model"`
		Instructions string         `json:"instructions"`
		Options      map[string]any `json:"prompt_cache_options"`
		Input        []struct {
			Role    string `json:"role"`
			Content []struct {
				Type, Text string
				Marker     map[string]any `json:"prompt_cache_breakpoint"`
			} `json:"content"`
		} `json:"input"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(body), ch16R2Constitution) != 1 {
		t.Errorf("constitution duplicated/lost: %s", body)
	}
	total := bytes.Count(body, []byte(`"prompt_cache_breakpoint"`))
	legal, developer := 0, 0
	for _, item := range envelope.Input {
		for _, block := range item.Content {
			if item.Role == "developer" && strings.Contains(block.Text, ch16R2Constitution) {
				developer++
				if block.Marker["mode"] != "explicit" {
					t.Error("constitution lacks explicit boundary")
				}
			}
			if block.Marker != nil && block.Marker["mode"] == "explicit" && (item.Role == "developer" || item.Role == "user" || item.Role == "assistant") && (block.Type == "input_text" || block.Type == "output_text") && block.Text != "" {
				legal++
			}
		}
	}
	if supported {
		if envelope.Options["mode"] != "explicit" || developer != 1 || envelope.Instructions != "" || total == 0 || total > 4 || total != legal {
			t.Errorf("unsupported explicit cache envelope: %s", body)
		}
	} else if envelope.Options != nil || total != 0 {
		t.Errorf("route acquired unsupported caching: %s", body)
	}
	return envelope.Model
}
