package gui

import (
	"bytes"
	"context"
	"encoding/json"
	"ensemble"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCh16ResponsesPublicWire(t *testing.T) {
	const secret = "private-continuation-must-not-reach-GUI"
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: {\"type\":\"response.reasoning_summary_text.delta\",\"output_index\":0,\"summary_index\":0,\"delta\":\"Public summary\"}\n\ndata: {\"type\":\"response.completed\",\"response\":{\"status\":\"completed\",\"model\":\"gpt-ch19-course\",\"output\":[{\"type\":\"reasoning\",\"id\":\"r1\",\"summary\":[{\"type\":\"summary_text\",\"text\":\"Public summary\"}],\"encrypted_content\":%q},{\"type\":\"message\",\"role\":\"assistant\",\"content\":[{\"type\":\"output_text\",\"text\":\"Visible answer\"}]}],\"usage\":{\"input_tokens\":8,\"output_tokens\":3,\"input_tokens_details\":{\"cached_tokens\":2}}}}\n\n", secret)
	}))
	defer peer.Close()
	app := ensemble.New(nil)
	defer app.Shutdown()
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-ch19-course", AuthMode: ensemble.AuthAPI, APIKey: "dummy-api", BaseURL: peer.URL + "/responses", BaseDir: t.TempDir(), DataDir: t.TempDir(), Tools: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	server, err := NewServerWithSettings(app, "web/gui", filepath.Join(t.TempDir(), "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	host := httptest.NewServer(server.Handler(server.Assets))
	defer host.Close()
	if _, err = a.Ask(context.Background(), "Answer locally"); err != nil {
		t.Fatal(err)
	}
	ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(host.URL, "http")+"/ws", http.Header{"Origin": {host.URL}})
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, raw, err := ws.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte(secret)) || bytes.Contains(raw, []byte("dummy-api")) || bytes.Contains(raw, []byte("encrypted_content")) {
		t.Fatalf("private data on GUI wire: %s", raw)
	}
	var frame struct {
		Type string
		Data struct {
			Agents []struct {
				Config struct {
					BillingMode string `json:"billing_mode"`
				}
			}
			Events []Wire
		}
	}
	if err = json.Unmarshal(raw, &frame); err != nil {
		t.Fatal(err)
	}
	if frame.Type != "replay" || len(frame.Data.Agents) != 1 || frame.Data.Agents[0].Config.BillingMode != "api" {
		t.Fatalf("missing billing state: %s", raw)
	}
	summary, answer, placeholder, usage := false, false, false, false
	for _, w := range frame.Data.Events {
		if w.Part != nil {
			summary = summary || w.Part.Kind == "reasoning_summary" && w.Part.Text == "Public summary"
			answer = answer || w.Part.Kind == "text" && w.Part.Text == "Visible answer"
			placeholder = placeholder || w.Part.Kind == "continuation" && strings.Contains(w.Part.Text, "private continuation")
		}
		if w.Type == "usage" && w.Meter != nil && w.Meter.Requests == 1 {
			usage = true
		}
	}
	if !summary || !answer || !placeholder || !usage {
		t.Fatalf("incomplete public projection: summary=%t answer=%t placeholder=%t usage=%t %s", summary, answer, placeholder, usage, raw)
	}
}
