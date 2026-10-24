package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The real driver's fixed CDP operations over a local WebSocket. This is not
// browser/native/model evidence; it controls only the public page replies.
func driverPage(t *testing.T, outcome string) *driver {
	t.Helper()
	state := map[string]any{"ready": true, "agent_id": "target", "selected_agent_id": "target", "turn": "idle", "turn_id": 3, "outcome": "success", "prompt": "previous", "revision": 1, "ended": 0, "busy": false, "input_revision": 0, "draft": ""}
	upgrade := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrade.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
		for {
			var request struct {
				ID     int    `json:"id"`
				Method string `json:"method"`
				Params struct {
					Expression string `json:"expression"`
					Text       string `json:"text"`
				} `json:"params"`
			}
			if c.ReadJSON(&request) != nil {
				return
			}
			var value any = true
			switch request.Method {
			case "Input.insertText":
				state["draft"] = request.Params.Text
				state["input_revision"] = 1
			case "Runtime.evaluate":
				if strings.Contains(request.Params.Expression, "requestSubmit()") {
					state["turn_id"] = 4
					state["outcome"] = outcome
					if outcome == "capture-fails" {
						state["outcome"] = "success"
					}
					if outcome == "unrelated" {
						state["agent_id"] = "other"
					}
					state["prompt"] = state["draft"]
					state["draft"] = ""
					state["ended"] = 1
					state["revision"] = 2
				}
				if strings.Contains(request.Params.Expression, "speechChannel.status(") {
					value = state
				}
				if strings.Contains(request.Params.Expression, "speechChannel.read(") {
					if outcome == "capture-fails" {
						return
					}
					entry := map[string]any{"seq": 1, "event": "utterance", "text": "transport fixture"}
					if outcome == "error" {
						entry["text"] = "Error: target rejected request"
						entry["source"] = "error"
					}
					value = map[string]any{"cursor": 1, "events": []any{entry}}
				}
			}
			result := map[string]any{}
			if request.Method == "Runtime.evaluate" {
				result["result"] = map[string]any{"value": value}
			}
			if c.WriteJSON(map[string]any{"id": request.ID, "result": result}) != nil {
				return
			}
		}
	}))
	t.Cleanup(server.Close)
	c, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return &driver{page: &page{conn: c}, speech: json.NewEncoder(&strings.Builder{}), actions: json.NewEncoder(&strings.Builder{})}
}
func TestTargetFailureIsNotCompletion(t *testing.T) {
	for _, outcome := range []string{"error", "interrupted"} {
		t.Run(outcome, func(t *testing.T) {
			d := driverPage(t, outcome)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if _, err := d.act(ctx, operation{Op: "type", Text: "question"}); err != nil {
				t.Fatal(err)
			}
			if _, err := d.act(ctx, operation{Op: "submit"}); err != nil {
				t.Fatal(err)
			}
			result, err := d.act(ctx, operation{Op: "wait", TimeoutMS: 100})
			if err == nil {
				t.Fatalf("target %s reported success: %#v", outcome, result)
			}
		})
	}
}

func TestDriverSettlement(t *testing.T) {
	for _, outcome := range []string{"success", "error", "interrupted", "unrelated", "pending"} {
		t.Run(outcome, func(t *testing.T) {
			d := driverPage(t, outcome)
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			if _, err := d.act(ctx, operation{Op: "type", Text: "question"}); err != nil {
				t.Fatal(err)
			}
			if _, err := d.act(ctx, operation{Op: "submit"}); err != nil {
				t.Fatal(err)
			}
			err := d.settle(ctx)
			if (err == nil) != (outcome == "success") {
				t.Fatalf("settle %s: %v", outcome, err)
			}
		})
	}
}
func TestJSONLImmediateEOFStopsSession(t *testing.T) {
	for _, outcome := range []string{"success", "error", "capture-fails"} {
		t.Run(outcome, func(t *testing.T) {
			d := driverPage(t, outcome)
			input, err := os.CreateTemp(t.TempDir(), "empty")
			if err != nil {
				t.Fatal(err)
			}
			defer input.Close()
			previous := os.Stdin
			os.Stdin = input
			defer func() { os.Stdin = previous }()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err = runTestSession(ctx, d.page, "question", filepath.Join(t.TempDir(), "speech.jsonl"), "transcript", "", true, false, t.TempDir())
			if !errors.Is(err, errPeerEOF) {
				t.Fatalf("interactive EOF did not end session %s: %v", outcome, err)
			}
		})
	}
}

type failedSpeechWriter struct{}

func (failedSpeechWriter) Write([]byte) (int, error) { return 0, errors.New("transcript write failed") }
func TestTranscriptWriteErrorAndUnsubmittedDraft(t *testing.T) {
	d := driverPage(t, "success")
	d.speech = json.NewEncoder(failedSpeechWriter{})
	if _, err := d.hear(context.Background()); err == nil {
		t.Fatal("capture write error ignored")
	}
	if d.cursor != 0 {
		t.Fatal("failed write advanced capture cursor")
	}
	d.ownsDraft = true
	if err := d.settle(context.Background()); err == nil {
		t.Fatal("unsubmitted owned input reported complete")
	}
}
