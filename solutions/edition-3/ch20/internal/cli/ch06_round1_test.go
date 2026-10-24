package cli

import (
	"bytes"
	"context"
	"ensemble"
	"ensemble/internal/common"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCh6R1RefusalHistoryAndChat(t *testing.T) {
	for _, content := range []string{"", "visible"} {
		t.Run("content="+content, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				if content != "" {
					fmt.Fprint(w, "data: {\"model\":\"actual\",\"choices\":[{\"delta\":{\"content\":\"visible\"}}]}\n\n")
				}
				fmt.Fprint(w, `data: {"model":"actual","choices":[{"delta":{"refusal":"cannot "}}]}`+"\n\n")
				w.(http.Flusher).Flush()
				fmt.Fprint(w, `data: {"choices":[{"delta":{"refusal":"comply"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`+"\n\ndata: [DONE]\n\n")
			}))
			defer server.Close()
			app := ensemble.New(nil)
			defer app.Shutdown()
			log := filepath.Join(t.TempDir(), "agent.jsonl")
			a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "offline", APIKey: "fake", BaseURL: server.URL, LogPath: log, StreamCapability: ensemble.StreamSSE})
			if err != nil {
				t.Fatal(err)
			}
			q := a.Observe()
			defer q.Close()
			var out, diagnostics bytes.Buffer
			if err := runChat(context.Background(), a, strings.NewReader("prompt\n"), &out, &diagnostics); err != nil {
				t.Fatal(err)
			}
			if strings.Count(out.String(), "cannot comply") != 1 || (content != "" && strings.Count(out.String(), content) != 1) {
				t.Fatalf("missing/duplicate final: %q", out.String())
			}
			var finalText []string
			var ids []uint64
			for _, o := range q.Drain() {
				if f, ok := o.(ensemble.PartFinal); ok && f.Part.Type == "text" {
					finalText = append(finalText, f.Part.Text)
					ids = append(ids, f.Part.ID)
				}
			}
			expected := 1
			if content != "" {
				expected = 2
			}
			if len(finalText) != expected || finalText[len(finalText)-1] != "cannot comply" {
				t.Fatalf("final observations: %q", finalText)
			}
			for i, id := range ids {
				if id != uint64(i) {
					t.Fatalf("final IDs: %v", ids)
				}
			}
			saved, err := os.ReadFile(log)
			if err != nil || !bytes.Contains(saved, []byte("cannot comply")) {
				t.Fatalf("refusal not saved: %v %s", err, saved)
			}
			var texts []string
			for _, e := range a.Log().Events() {
				if e.Type == common.ResponseEnded {
					for _, p := range e.Response.Parts {
						if p.Type == "text" {
							texts = append(texts, p.Text)
						}
					}
				}
			}
			if len(texts) != expected || texts[len(texts)-1] != "cannot comply" {
				t.Fatalf("refusal not in History: %q", texts)
			}
		})
	}
}
