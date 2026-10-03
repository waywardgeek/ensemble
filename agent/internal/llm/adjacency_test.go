package llm

import (
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// A result that arrives LATE is a different fault from one that never arrives,
// and the difference is easy to miss because both leave a call looking
// unanswered at the moment the request is built.
//
// The distinction matters because set membership hides it. Asking "is this call
// answered anywhere in the dialogue?" says yes for a late result, so a repair
// keyed on that question silently does nothing while the vendor keeps refusing
// the request. The rule every vendor actually enforces is ADJACENCY: the result
// must sit in the message immediately after the call.
//
// These tests pin the checker that tells the two apart.

func adjCall(seq common.Seq, callID, name string) common.Entry {
	return common.Entry{
		Seq:   seq,
		Actor: common.ActorAgent,
		Kind:  common.KindDialogue,
		Parts: []common.Part{common.ToolCallPart{CallID: callID, Name: name}},
	}
}

func adjResult(seq common.Seq, callID string) common.Entry {
	return common.Entry{
		Seq:   seq,
		Actor: common.ActorTool,
		Kind:  common.KindDialogue,
		Parts: []common.Part{common.ToolResultPart{CallID: callID}},
	}
}

func adjHuman(seq common.Seq, text string) common.Entry {
	return common.Entry{
		Seq:   seq,
		Actor: common.ActorHuman,
		Kind:  common.KindDialogue,
		Parts: []common.Part{common.TextPart{Text: text}},
	}
}

// TestHintDuringToolCallIsReportedAsMisplaced reproduces the shape that broke a
// real session: a hint typed while a tool was still running landed between the
// call and its result.
//
// This is not an exotic failure. Typing to the agent mid-tool-call is a normal
// thing to do -- it is the whole point of being able to steer a run in flight
// -- so the dialogue grows an entry in the one place no vendor allows one.
func TestHintDuringToolCallIsReportedAsMisplaced(t *testing.T) {
	ctx := common.NewContext()
	ctx.Dialogue = []common.Entry{
		adjCall(1, "call_ok", "read_file"),
		adjResult(2, "call_ok"),
		adjCall(3, "call_late", "run_command"),
		adjHuman(4, "How is it coming?"),
		adjHuman(5, "There was an error in the last call"),
		{Seq: 6, Actor: common.ActorSystem, Kind: common.KindRecall,
			Parts: []common.Part{common.TextPart{Text: "[Auto-recalled memories]"}}},
		adjHuman(7, "hi"),
		adjResult(8, "call_late"),
	}

	bad := misplacedCalls(ctx)
	if len(bad) != 1 {
		t.Fatalf("want exactly 1 misplaced call, got %d: %v", len(bad), bad)
	}
	got := bad[0]
	if got.CallID != "call_late" {
		t.Errorf("CallID = %q, want call_late", got.CallID)
	}
	// The healthy call must not be dragged into the report. A checker that
	// blames every call once one is wrong is noise, not a diagnosis.
	if got.CallAt != 2 || got.ResultAt != 7 {
		t.Errorf("CallAt/ResultAt = %d/%d, want 2/7", got.CallAt, got.ResultAt)
	}
	// Naming the intruder is the point: it turns "the context is malformed"
	// into "your hint landed while a tool was running".
	if got.BlockedAt != 3 {
		t.Errorf("BlockedAt = %d, want 3 (the first hint)", got.BlockedAt)
	}
	if !strings.Contains(got.BlockedBy, "human") {
		t.Errorf("BlockedBy = %q, want it to name the human entry", got.BlockedBy)
	}
	if !strings.Contains(got.String(), "call_late") {
		t.Errorf("report does not name the call: %s", got.String())
	}
}

// TestAdjacentResultIsNotReported keeps the checker quiet on healthy dialogue.
// A warning that fires on ordinary sessions gets filtered out by the reader
// long before the one that matters arrives.
func TestAdjacentResultIsNotReported(t *testing.T) {
	ctx := common.NewContext()
	ctx.Dialogue = []common.Entry{
		adjHuman(1, "go"),
		adjCall(2, "call_a", "read_file"),
		adjResult(3, "call_a"),
		adjHuman(4, "thanks"),
	}
	if bad := misplacedCalls(ctx); len(bad) != 0 {
		t.Fatalf("healthy dialogue reported as misplaced: %v", bad)
	}
}

// TestBatchAnsweredBySeveralToolEntriesIsAdjacent covers the ordinary parallel
// case: one assistant turn asking for several tools, answered by one tool entry
// each. Those entries are adjacent to the call as a group, and flagging them
// would fire on every parallel tool use in the system.
func TestBatchAnsweredBySeveralToolEntriesIsAdjacent(t *testing.T) {
	ctx := common.NewContext()
	ctx.Dialogue = []common.Entry{
		{Seq: 1, Actor: common.ActorAgent, Kind: common.KindDialogue, Parts: []common.Part{
			common.ToolCallPart{CallID: "a", Name: "read_file"},
			common.ToolCallPart{CallID: "b", Name: "list_directory"},
		}},
		adjResult(2, "a"),
		adjResult(3, "b"),
	}
	if bad := misplacedCalls(ctx); len(bad) != 0 {
		t.Fatalf("parallel batch reported as misplaced: %v", bad)
	}
}

// TestLostCallIsNotReportedAsMisplaced keeps one fault to one diagnosis. A call
// with no result anywhere is lost, and lostCalls already reports and repairs it;
// saying so twice in different words would leave a reader reconciling two
// messages about one problem.
func TestLostCallIsNotReportedAsMisplaced(t *testing.T) {
	ctx := common.NewContext()
	ctx.Dialogue = []common.Entry{
		adjCall(1, "call_gone", "run_command"),
		adjHuman(2, "still there?"),
	}
	if bad := misplacedCalls(ctx); len(bad) != 0 {
		t.Fatalf("lost call reported as misplaced: %v", bad)
	}
	if lost := lostCalls(ctx); len(lost) != 1 {
		t.Fatalf("want the lost call reported by lostCalls, got %d", len(lost))
	}
}

// TestMisplacedCallIsReportedNotRepaired pins the decision. The context is a
// projection of the event log, so reordering entries here would make a replay
// disagree with the session it replays. The checker diagnoses; the cure belongs
// upstream, where the entry is placed.
func TestMisplacedCallIsReportedNotRepaired(t *testing.T) {
	ctx := common.NewContext()
	ctx.Dialogue = []common.Entry{
		adjCall(1, "call_late", "run_command"),
		adjHuman(2, "hi"),
		adjResult(3, "call_late"),
	}
	before := len(ctx.Dialogue)
	if bad := misplacedCalls(ctx); len(bad) != 1 {
		t.Fatalf("want 1 misplaced call, got %d", len(bad))
	}
	if len(ctx.Dialogue) != before {
		t.Fatalf("checker changed the dialogue: %d entries, want %d", len(ctx.Dialogue), before)
	}
	if _, ok := ctx.Dialogue[1].Parts[0].(common.TextPart); !ok {
		t.Fatalf("checker moved an entry; entry 1 is no longer the human message")
	}
}
