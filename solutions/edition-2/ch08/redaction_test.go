package ensemble_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"example.com/ensemble"
)

func TestRedactionRefusalsAreActionableAndDoNotMutate(t *testing.T) {
	var diagnostics bytes.Buffer
	app := ensemble.New(&diagnostics)
	path := filepath.Join(t.TempDir(), "log")
	agent, err := app.NewAgent(ensemble.Config{DisableStreaming: true, Model: "fixture", APIKey: "credential-marker", LogPath: path})
	if err != nil {
		t.Fatal(err)
	}
	defer agent.Close()
	provenance := ensemble.Provenance{Vendor: "anthropic", Model: "fixture", Surface: "messages"}
	for _, event := range []ensemble.Event{
		{Type: "message_received", Message: &ensemble.Entry{Actor: "human", Purpose: "dialogue", Parts: []ensemble.Part{ensemble.Text("private-prompt")}}},
		{Type: "response_ended", Response: &ensemble.Response{From: provenance, Usage: &ensemble.Usage{Input: 3, Output: 2}, Parts: []ensemble.Part{{Type: "tool_call", CallID: "private-call", Name: "inspect", Args: json.RawMessage(`{}`), From: &provenance}}}},
		{Type: "tool_returned", Tool: &ensemble.ToolEvent{CallID: "private-call", Parts: []ensemble.Part{ensemble.Text("private-result")}}},
	} {
		if err := agent.Append(event); err != nil {
			t.Fatal(err)
		}
	}
	before, _ := os.ReadFile(path)
	snapshot, usage, events := agent.Snapshot(), agent.Usage(), agent.Events()
	for _, test := range []struct {
		redaction ensemble.Redaction
		want      string
	}{
		{ensemble.Redaction{From: 0, To: 3, Level: "redact_result", Reason: "private-reason"}, "positive ordered sequences"},
		{ensemble.Redaction{From: 3, To: 2, Level: "redact_result", Reason: "private-reason"}, "positive ordered sequences"},
		{ensemble.Redaction{From: 3, To: 4, Level: "redact_result", Reason: "private-reason"}, "already in this log"},
		{ensemble.Redaction{From: 3, To: 3, Level: "private-level", Reason: "private-reason"}, "level must be redact_result"},
		{ensemble.Redaction{From: 3, To: 3, Level: "redact_result"}, "reason must be nonempty"},
		{ensemble.Redaction{From: 3, To: 3, Level: "redact_result", Reason: " \t"}, "reason must be nonempty"},
		{ensemble.Redaction{From: 1, To: 2, Level: "redact_result", Reason: "private-reason"}, "no tool results"},
	} {
		diagnostics.Reset()
		err := agent.Redact(test.redaction)
		if err == nil || !strings.Contains(err.Error(), test.want) || !strings.Contains(diagnostics.String(), test.want) {
			t.Fatalf("want %q, got %v / %s", test.want, err, diagnostics.String())
		}
		if strings.Contains(err.Error()+diagnostics.String(), "private-") || strings.Contains(err.Error()+diagnostics.String(), "credential-marker") {
			t.Fatal("refusal exposed supplied content")
		}
		after, _ := os.ReadFile(path)
		if !bytes.Equal(before, after) || !reflect.DeepEqual(snapshot, agent.Snapshot()) || !reflect.DeepEqual(events, agent.Events()) || usage != agent.Usage() {
			t.Fatal("rejected redaction changed owned state")
		}
	}
	if err := agent.Redact(ensemble.Redaction{From: 3, To: 3, Level: "redact_result", Reason: "controlled positive"}); err != nil {
		t.Fatal("rejections prevented a later valid directive", err)
	}
	body, err := agent.Render(agent.Config())
	if err != nil || bytes.Contains(body, []byte("private-result")) || !bytes.Contains(body, []byte("[redacted]")) {
		t.Fatal("valid target failed", err)
	}
}
