package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ensemble/ensemble"
	"ensemble/internal/common"
)

// A cut must affect the very next request, not just a later render or replay.
// Observe the real post-reduction boundary before the HTTP worker starts.
func TestContextSnapshotRefreshAfterLadder(t *testing.T) {
	var expected []byte
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil || !bytes.Equal(body, expected) {
			t.Error("request rendered a stale snapshot after the ladder cut")
		}
		replyText(writer, "done")
	}))
	defer server.Close()
	agent := actorAgent(t, server, func(config *ensemble.Config) { config.ContextTarget = 20000 })
	payload, _ := json.Marshal(strings.Repeat("OLD RESULT ", 1024))
	log := `{"seq":1,"type":"message_received","message":{"actor":"human","parts":[{"type":"text","text":"original"}]}}` + "\n" +
		`{"seq":2,"type":"request_sent","request":{"to":{"vendor":"anthropic","model":"fake","surface":"messages"}}}` + "\n" +
		`{"seq":3,"type":"response_ended","response":{"from":{"vendor":"anthropic","model":"fake","surface":"messages"},"parts":[{"type":"tool_call","call_id":"old","name":"echo","args":{},"from":{"vendor":"anthropic","model":"fake","surface":"messages"}}],"usage":{}}}` + "\n" +
		`{"seq":4,"type":"tool_returned","tool":{"call_id":"old","parts":[{"type":"text","text":` + string(payload) + `}]}}` + "\n" +
		`{"seq":5,"type":"request_sent","request":{"to":{"vendor":"anthropic","model":"fake","surface":"messages"}}}` + "\n" +
		`{"seq":6,"type":"response_ended","response":{"from":{"vendor":"anthropic","model":"fake","surface":"messages"},"parts":[{"type":"text","text":"previous answer"}],"usage":{}}}` + "\n"
	if err := agent.History().Load(strings.NewReader(log)); err != nil {
		t.Fatal(err)
	}
	cuts := 0
	answer, err := agent.AskWatching(context.Background(), "continue", common.StreamCallbacks{OnEvent: func(event common.Event) {
		if event.Redact != nil && event.Redact.Reason == "tool-byte ladder" {
			cuts++
		}
		if event.Type == "request_sent" {
			var renderErr error
			expected, renderErr = agent.Engine().Render(agent.History().Context())
			if renderErr != nil {
				t.Error(renderErr)
			}
		}
	}})
	if err != nil || answer != "done" || cuts == 0 || len(expected) == 0 {
		t.Fatalf("ladder turn: answer=%q err=%v cuts=%d", answer, err, cuts)
	}
}
