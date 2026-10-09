package llm

import (
	"example.com/ensemble/internal/common"
	"testing"
)

func TestAdmissionIndexMatchesRetainedState(t *testing.T) {
	// These accepted reducer transitions exercise moving, consuming and dropping
	// state; recomputation is a separate oracle for the incremental admission cache.
	c := common.Context{}
	events := []common.Event{
		{Type: "session_initialized", Session: &common.SessionFact{}},
		{Type: "message_received", Message: &common.Entry{Actor: "system", Purpose: "instruction", Parts: []common.Part{Text("instruction")}}},
		{Type: "message_received", Message: &common.Entry{Actor: "system", Purpose: "ephemeral", Parts: []common.Part{Text("one request")}}},
		{Type: "hint_received", Hint: &common.HintEvent{Text: "hint"}},
		{Type: "message_received", Message: &common.Entry{Actor: "human", Purpose: "dialogue", Parts: []common.Part{Text("question")}}},
		{Type: "request_sent", Request: &common.RequestEvent{Ephemera: []uint64{3}, Hints: []uint64{4}}},
		{Type: "response_ended", Response: &common.Response{Parts: []common.Part{{Type: "tool_call", CallID: "c"}}, Usage: &common.Usage{}}},
		{Type: "tool_called", Tool: &common.ToolEvent{CallID: "c"}},
		{Type: "tool_returned", Tool: &common.ToolEvent{CallID: "c", Parts: []common.Part{Text("result"), {Type: "blob"}}}},
		{Type: "request_sent", Request: &common.RequestEvent{}},
		{Type: "response_ended", Response: &common.Response{Parts: []common.Part{Text("answer")}, Usage: &common.Usage{}}},
		{Type: "message_received", Message: &common.Entry{Actor: "human", Purpose: "dialogue", Parts: []common.Part{Text("aborted")}}},
		{Type: "error_occurred"},
	}
	for i, e := range events {
		e.Seq = uint64(i + 1)
		Apply(nil, &c, e)
		expected := c
		ReindexContext(nil, &expected)
		if c.Index != expected.Index {
			t.Fatalf("after %s: incremental %+v recomputed %+v", e.Type, c.Index, expected.Index)
		}
	}
}
