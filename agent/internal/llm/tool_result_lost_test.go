package llm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// A tool call with no result is a legal thing for the LOG to hold and an
// illegal thing to SEND. The log records what happened, and sometimes what
// happened is that a result went missing; the wire has no way to express it.
// Anthropic and OpenAI reject such a request outright, and Gemini has been
// seen to answer anyway, which is worse, because the conversation is then
// quietly wrong rather than loudly broken.
//
// These tests pin the repair: the gap is closed before a request is built,
// and closing it is RECORDED rather than done in silence.

func lostCallEngine(t *testing.T, entries ...common.Entry) *Engine {
	t.Helper()
	e := &Engine{Log: NewLog(), Ctx: common.NewContext(), Host: testHost{t}}
	e.Ctx.Dialogue = append(e.Ctx.Dialogue, entries...)
	return e
}

func callEntry(seq common.Seq, callID, name string) common.Entry {
	return common.Entry{Seq: seq, Actor: common.ActorAgent, Kind: common.KindDialogue, Parts: common.PartList{
		common.ToolCallPart{CallID: callID, Name: name},
	}}
}

func resultEntry(seq common.Seq, callID string) common.Entry {
	return common.Entry{Seq: seq, Actor: common.ActorTool, Kind: common.KindDialogue, Parts: common.PartList{
		common.ToolResultPart{CallID: callID},
	}}
}

func lostEvents(l *common.Log) []common.Event {
	var out []common.Event
	for _, ev := range l.Events {
		if ev.Type == common.ToolResultLost {
			out = append(out, ev)
		}
	}
	return out
}

// The repair itself: a call that never returned is closed, so the context
// becomes sendable.
func TestACallThatNeverReturnedIsClosedBeforeARequest(t *testing.T) {
	e := lostCallEngine(t, callEntry(1, "call_a", "read_file"))
	e.Ctx.Turn = common.ToolsPending

	if got := outstandingCalls(e.Ctx); got != 1 {
		t.Fatalf("setup: outstanding calls = %d, want 1", got)
	}
	if err := e.closeLostCalls(); err != nil {
		t.Fatalf("closeLostCalls: %v", err)
	}
	if got := outstandingCalls(e.Ctx); got != 0 {
		t.Fatalf("outstanding calls = %d, want 0: the context would still be refused by every vendor", got)
	}
}

// The anomaly is EVIDENCE, not a silent fix. A conversation that heals itself
// without saying so leaves nothing to debug the next time this happens.
func TestClosingALostCallIsRecordedInTheLog(t *testing.T) {
	e := lostCallEngine(t, callEntry(1, "call_a", "read_file"))
	if err := e.closeLostCalls(); err != nil {
		t.Fatalf("closeLostCalls: %v", err)
	}

	lost := lostEvents(e.Log)
	if len(lost) != 1 {
		t.Fatalf("tool_result_lost events = %d, want 1", len(lost))
	}
	if lost[0].Tool == nil {
		t.Fatal("event carries no tool payload, so nothing says WHICH call was closed")
	}
	if lost[0].Tool.CallID != "call_a" {
		t.Fatalf("event names call %q, want %q", lost[0].Tool.CallID, "call_a")
	}
}

// The standing result must be usable by the model: flagged as an error, and
// honest that the outcome is unknown rather than claiming a failure we did
// not observe.
func TestTheStandingResultSaysTheOutcomeIsUnknown(t *testing.T) {
	e := lostCallEngine(t, callEntry(1, "call_a", "read_file"))
	if err := e.closeLostCalls(); err != nil {
		t.Fatalf("closeLostCalls: %v", err)
	}

	last := e.Ctx.Dialogue[len(e.Ctx.Dialogue)-1]
	if len(last.Parts) != 1 {
		t.Fatalf("closing entry has %d parts, want 1", len(last.Parts))
	}
	res, ok := last.Parts[0].(common.ToolResultPart)
	if !ok {
		t.Fatalf("closing entry holds %T, want a ToolResultPart", last.Parts[0])
	}
	if !res.IsError {
		t.Error("standing result is not flagged as an error")
	}
	if res.CallID != "call_a" {
		t.Errorf("standing result answers %q, want %q", res.CallID, "call_a")
	}

	var text string
	for _, p := range res.Parts {
		if tp, ok := p.(common.TextPart); ok {
			text += tp.Text
		}
	}
	if !strings.Contains(text, "read_file") {
		t.Errorf("standing result does not name the tool, so the model cannot tell what was lost: %q", text)
	}
	if !strings.Contains(text, "NOT") && !strings.Contains(text, "not known") {
		t.Errorf("standing result should say the outcome is unknown, not assert a failure: %q", text)
	}
}

// The ordinary tool loop must be untouched. Every result is recorded before
// the next request is built, so there is nothing to close and no event to
// write.
func TestAnAnsweredCallIsLeftAlone(t *testing.T) {
	e := lostCallEngine(t,
		callEntry(1, "call_a", "read_file"),
		resultEntry(2, "call_a"),
	)
	before := len(e.Ctx.Dialogue)

	if err := e.closeLostCalls(); err != nil {
		t.Fatalf("closeLostCalls: %v", err)
	}
	if got := len(lostEvents(e.Log)); got != 0 {
		t.Fatalf("tool_result_lost events = %d, want 0: a normal tool loop must not be touched", got)
	}
	if got := len(e.Ctx.Dialogue); got != before {
		t.Fatalf("dialogue grew from %d to %d entries on a healthy conversation", before, got)
	}
}

// Only the unanswered call of a batch is closed. A parallel batch where one
// result arrived and one did not is the realistic shape of this failure.
func TestOnlyTheUnansweredCallOfABatchIsClosed(t *testing.T) {
	e := lostCallEngine(t,
		callEntry(1, "call_a", "read_file"),
		callEntry(2, "call_b", "run_command"),
		resultEntry(3, "call_a"),
	)
	if err := e.closeLostCalls(); err != nil {
		t.Fatalf("closeLostCalls: %v", err)
	}

	lost := lostEvents(e.Log)
	if len(lost) != 1 {
		t.Fatalf("tool_result_lost events = %d, want 1", len(lost))
	}
	if lost[0].Tool.CallID != "call_b" {
		t.Fatalf("closed %q, want the unanswered call %q", lost[0].Tool.CallID, "call_b")
	}
}

// Order is part of the contract. These calls become events carrying sequence
// numbers, so ranging over a map would number the same gap differently on
// every run and two replays of one log would disagree. Go randomises map
// iteration, so a loop is the honest way to ask.
func TestLostCallsComeBackInDialogueOrder(t *testing.T) {
	for i := 0; i < 50; i++ {
		c := common.NewContext()
		c.Dialogue = append(c.Dialogue,
			callEntry(1, "call_a", "t1"),
			callEntry(2, "call_b", "t2"),
			callEntry(3, "call_c", "t3"),
		)
		got := lostCalls(c)
		if len(got) != 3 {
			t.Fatalf("lostCalls returned %d calls, want 3", len(got))
		}
		for j, want := range []string{"call_a", "call_b", "call_c"} {
			if got[j].CallID != want {
				t.Fatalf("run %d: position %d is %q, want %q", i, j, got[j].CallID, want)
			}
		}
	}
}

// Closing the last outstanding call releases the turn, exactly as a real
// result would. Without this the state machine stays in ToolsPending waiting
// for something that is never coming.
func TestClosingTheLastCallReleasesTheTurn(t *testing.T) {
	c := common.NewContext()
	c.Dialogue = append(c.Dialogue, callEntry(1, "call_a", "read_file"))
	c.Turn = common.ToolsPending

	err := Apply(c, common.Event{
		Seq:  2,
		Type: common.ToolResultLost,
		Tool: &common.ToolData{CallID: "call_a", Name: "read_file"},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if c.Turn != common.InputPending {
		t.Fatalf("turn is %v, want InputPending: the turn is still waiting on a dead call", c.Turn)
	}
}

// The reducer refuses an event that names no call rather than inventing one.
func TestToolResultLostWithNoPayloadIsRefused(t *testing.T) {
	c := common.NewContext()
	err := Apply(c, common.Event{Seq: 1, Type: common.ToolResultLost})
	if err == nil {
		t.Fatal("Apply accepted a tool_result_lost with no tool payload")
	}
}

// unansweredOnTheWire walks a rendered Anthropic request and returns the
// tool_use ids that have no matching tool_result.
//
// It reads the BYTES rather than the context, because the context is the
// thing under test. Asserting that the projection looks right would restate
// the production code; this asks the only question the vendor asks.
func unansweredOnTheWire(t *testing.T, body string) []string {
	t.Helper()
	var req struct {
		Messages []struct {
			Content []struct {
				Type      string `json:"type"`
				ID        string `json:"id"`
				ToolUseID string `json:"tool_use_id"`
			} `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("request body is not JSON: %v", err)
	}
	var calls []string
	answered := map[string]bool{}
	for _, m := range req.Messages {
		for _, b := range m.Content {
			switch b.Type {
			case "tool_use":
				calls = append(calls, b.ID)
			case "tool_result":
				answered[b.ToolUseID] = true
			}
		}
	}
	var unanswered []string
	for _, id := range calls {
		if !answered[id] {
			unanswered = append(unanswered, id)
		}
	}
	return unanswered
}

// The WIRING, which is the part that actually broke.
//
// Every test above calls closeLostCalls directly, so deleting its one call
// site in Turn would leave all of them green while the bug walked straight
// back in. This drives a real request through the engine and asserts on the
// bytes that reach the vendor.
func TestARequestBuiltOverALostCallIsLegalOnTheWire(t *testing.T) {
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading request body: %v", err)
		}
		bodies = append(bodies, string(body))
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"msg_1","type":"message","role":"assistant","model":"claude-sonnet-5",
			"content":[{"type":"text","text":"ok"}],
			"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	t.Cleanup(srv.Close)

	const model = "claude-sonnet-5"
	e := &Engine{
		Log:  NewLog(),
		Ctx:  common.NewContext(),
		Host: testHost{t},
		HTTP: srv.Client(),
		Cfg: common.Config{
			Model:   model,
			Vendor:  common.VendorAnthropic,
			Surface: common.SurfaceForModel(model, common.VendorAnthropic),
			BaseURL: srv.URL,
			APIKey:  "test-key",
		},
	}
	// A call that never came back, exactly as a crashed or dropped dispatch
	// leaves it in save.json.
	e.Ctx.Dialogue = append(e.Ctx.Dialogue, callEntry(1, "call_a", "read_file"))
	e.Ctx.Turn = common.ToolsPending

	if _, err := e.Turn(common.StreamCallbacks{}); err != nil {
		t.Fatalf("Turn: %v", err)
	}
	if len(bodies) != 1 {
		t.Fatalf("vendor saw %d requests, want 1", len(bodies))
	}
	if bad := unansweredOnTheWire(t, bodies[0]); len(bad) != 0 {
		t.Fatalf("request carries tool calls with no result %v: every vendor refuses this, and one answers anyway", bad)
	}
}
