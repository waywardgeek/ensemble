package ensemble

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// A retained refusal is replay material, not answer text. Test the actual next
// HTTP request so accepting the first response cannot hide a broken renderer.
func TestChatRefusalReplaysIntoContinuationAndNextTurn(t *testing.T) {
	for _, plain := range []bool{false, true} {
		for _, tool := range []bool{false, true} {
			t.Run(fmt.Sprintf("plain=%t/tool=%t", plain, tool), func(t *testing.T) {
				count := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					count++
					var request struct {
						Messages []map[string]json.RawMessage `json:"messages"`
					}
					if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
						t.Error(err)
					}
					if count > 1 {
						found := 0
						for _, message := range request.Messages {
							if raw, ok := message["refusal"]; ok {
								found++
								if string(raw) != `"retained refusal"` || string(message["role"]) != `"assistant"` {
									t.Errorf("wrong refusal placement: %s", raw)
								}
								if tool && message["tool_calls"] == nil {
									t.Error("refusal detached from its call group")
								}
							}
						}
						if found != 1 {
							t.Errorf("request %d retained %d refusals", count, found)
						}
					}
					message := map[string]any{"role": "assistant", "content": "continued"}
					reason := "stop"
					if count == 1 {
						message["refusal"] = "retained refusal"
						message["content"] = "visible answer"
						if tool {
							reason = "tool_calls"
							message["content"] = nil
							message["tool_calls"] = []any{map[string]any{"id": "read-1", "type": "function", "function": map[string]any{"name": "read_file", "arguments": `{"path":"notes.txt"}`}}}
						}
					}
					usage := map[string]int{"prompt_tokens": 10, "completion_tokens": 2}
					if plain {
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(map[string]any{"model": "fixture", "choices": []any{map[string]any{"index": 0, "message": message, "finish_reason": reason}}, "usage": usage})
					} else {
						if calls, ok := message["tool_calls"].([]any); ok {
							calls[0].(map[string]any)["index"] = 0
						}
						w.Header().Set("Content-Type", "text/event-stream")
						body, _ := json.Marshal(map[string]any{"model": "fixture", "choices": []any{map[string]any{"index": 0, "delta": message, "finish_reason": reason}}})
						fmt.Fprintf(w, "data: %s\n\n", body)
						body, _ = json.Marshal(map[string]any{"choices": []any{}, "usage": usage})
						fmt.Fprintf(w, "data: %s\n\ndata: [DONE]\n\n", body)
					}
				}))
				defer server.Close()
				dir := t.TempDir()
				if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("read result"), 0600); err != nil {
					t.Fatal(err)
				}
				app := New(nil)
				defer app.Close()
				a, err := app.NewAgent(Config{DisableStreaming: plain, Vendor: "openai", Model: "fixture", APIKey: "fixture", BaseURL: server.URL, Workspace: dir, LogPath: filepath.Join(dir, "session.log"), Builtins: []string{"read_file"}})
				if err != nil {
					t.Fatal(err)
				}
				first, err := a.Ask(context.Background(), "first")
				if err != nil {
					t.Fatal(err)
				}
				expected := "visible answer"
				if tool {
					expected = "continued"
				}
				if first != expected {
					t.Fatal("refusal leaked into answer", first)
				}
				second, err := a.Ask(context.Background(), "next turn")
				if err != nil || second != "continued" {
					t.Fatal(second, err)
				}
				want := 2
				if tool {
					want = 3
				}
				if count != want {
					t.Fatal("HTTP count", count, want)
				}
			})
		}
	}
}
