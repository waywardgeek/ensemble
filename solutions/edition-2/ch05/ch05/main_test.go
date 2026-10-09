package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ensemble/ensemble"
)

// The external peer supplies a tool call and echoes the result it actually
// received. All framework ownership, dispatch, jobs and wire code remain real.
func TestExternalToolLoop(t *testing.T) {
	var logs bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Tools []struct {
				Name, Description string
				Schema            json.RawMessage `json:"input_schema"`
			}
			Messages []struct {
				Content []struct {
					Type    string
					ID      string `json:"tool_use_id"`
					Content []struct{ Text string }
				}
			}
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		if len(request.Tools) != 1 || request.Tools[0].Name != "shout" || request.Tools[0].Description != "Return text converted to uppercase" || !bytes.Contains(request.Tools[0].Schema, []byte(`"text"`)) {
			t.Errorf("custom declaration missing or changed: %+v", request.Tools)
		}
		for _, message := range request.Messages {
			for _, part := range message.Content {
				if part.Type == "tool_result" {
					if part.ID != "external-1" || len(part.Content) != 1 || !strings.HasPrefix(part.Content[0].Text, "HELLO FRAMEWORK") {
						t.Errorf("custom result did not reach provider: %+v", part)
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"content": []map[string]string{{"type": "text", "text": part.Content[0].Text}}, "usage": map[string]int{"input_tokens": 2, "output_tokens": 1}})
					return
				}
			}
		}
		_, _ = io.WriteString(w, `{"content":[{"type":"tool_use","id":"external-1","name":"shout","input":{"text":"hello framework"}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	defer server.Close()
	app := ensemble.New(&logs)
	a, err := app.NewAgent(ensemble.Config{Model: "fake", BaseURL: server.URL, DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	if err := registerShout(a); err != nil {
		t.Fatal(err)
	}
	answer, err := a.Ask(context.Background(), "use the tool")
	if err != nil || !strings.HasPrefix(answer, "HELLO FRAMEWORK") {
		t.Fatalf("answer = %q, %v", answer, err)
	}
	if !strings.HasSuffix(logs.String(), " shout called\n") || len(logs.String()) < 20 {
		t.Fatalf("tool did not reach root logger: %q", logs.String())
	}
	if _, err := time.Parse("2006/01/02 15:04:05", logs.String()[:19]); err != nil {
		t.Fatalf("logger omitted timestamp: %q", logs.String())
	}
}
