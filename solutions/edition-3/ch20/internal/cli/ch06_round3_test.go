package cli

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ensemble"
)

func TestCh6R3ChatToolNameChunks(t *testing.T) {
	for _, plain := range []bool{false, true} {
		t.Run(fmt.Sprintf("plain=%t", plain), func(t *testing.T) {
			t.Chdir(t.TempDir())
			if err := os.WriteFile("input.txt", []byte("file body"), 0600); err != nil {
				t.Fatal(err)
			}
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n := requests.Add(1)
				if n > 2 {
					t.Errorf("unexpected request %d", n)
					http.Error(w, "too many requests", 500)
					return
				}
				if plain || n == 2 {
					w.Header().Set("Content-Type", "application/json")
					if n == 1 {
						fmt.Fprint(w, `{"model":"actual","choices":[{"message":{"content":"Reading.","tool_calls":[{"id":"call-1","type":"function","function":{"name":"read_file","arguments":"{\"path\":\"input.txt\"}"}}]},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`)
					} else {
						fmt.Fprint(w, `{"model":"actual","choices":[{"message":{"content":"done"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`)
					}
					return
				}
				w.Header().Set("Content-Type", "text/event-stream")
				for _, frame := range []string{
					`{"model":"actual","choices":[{"delta":{"content":"Reading."}}]}`,
					`{"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call-1","type":"function","function":{"name":"read_"}}]}}]}`,
					`{"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"name":"file","arguments":"{\"path\":"}}]}}]}`,
					`{"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\"input.txt\"}"}}]}}]}`,
					`{"choices":[{"delta":{},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`,
					`[DONE]`,
				} {
					fmt.Fprint(w, "data: "+frame+"\n\n")
					w.(http.Flusher).Flush()
				}
			}))
			defer server.Close()
			app := ensemble.New(nil)
			defer app.Shutdown()
			a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Surface: ensemble.ChatCompletions, Model: "gpt-5.4", BaseURL: server.URL, APIKey: "dummy", Tools: []string{"read_file"}, DisableStreaming: plain})
			if err != nil {
				t.Fatal(err)
			}
			var out, diagnostics bytes.Buffer
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := runChat(ctx, a, strings.NewReader("read input.txt\n"), &out, &diagnostics); err != nil {
				t.Fatal(err)
			}
			if requests.Load() != 2 {
				t.Fatalf("requests=%d", requests.Load())
			}
			stdout, stderr := out.String(), diagnostics.String()
			if strings.Count(stderr, "\ntool> ") != 1 || strings.Count(stderr, "read_file") != 1 {
				t.Fatalf("tool chunks duplicated/missing: %q", stderr)
			}
			_, display, _ := strings.Cut(stderr, "\ntool> ")
			line, _, _ := strings.Cut(display, "\n")
			args := strings.TrimSpace(strings.TrimPrefix(line, "read_file"))
			if args != `{"path":"input.txt"}` {
				t.Fatalf("tool name/args display: %q", line)
			}
			if strings.Contains(stdout, "read_file") || strings.Contains(stdout, "input.txt") || strings.Count(stdout, "Reading.") != 1 || strings.Count(stdout, "done") != 1 {
				t.Fatalf("stdout routing or duplication: %q", stdout)
			}
		})
	}
}
