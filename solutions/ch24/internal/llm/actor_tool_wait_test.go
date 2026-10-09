package llm

import (
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// A completion left over from an interrupted turn must not release the next
// tool's wait. The observer queues the current completion only AFTER the stale
// one is delivered, so returning on the stale message leaves it unprocessed.
func TestToolWaitIgnoresStaleCompletion(t *testing.T) {
	e := lostCallEngine(t)
	a := NewActor(e, e.parent)
	var finished []string
	a.Attach(ObserverFunc(func(obs common.Observation) {
		if f, ok := obs.(common.ToolFinished); ok {
			finished = append(finished, f.CallID)
			if f.CallID == "old" {
				a.mb.Post(common.Hint{Text: "Check the new tool, not the old one."})
				a.mb.Post(common.ToolCompleted{CallID: "current", Result: "done"})
			}
		}
	}))
	a.mb.Post(common.ToolCompleted{CallID: "old", Result: "late result"})

	if !a.waitForTool("current") {
		t.Fatal("tool wait was interrupted")
	}
	if len(finished) != 2 || finished[0] != "old" || finished[1] != "current" {
		t.Fatalf("delivered completions = %v; wait must consume current, not just old", finished)
	}
	if len(e.Log.Events) != 1 {
		t.Fatalf("recorded %d events, want the intervening hint", len(e.Log.Events))
	}
	ev := e.Log.Events[0]
	if ev.Type != common.MessageReceived || ev.Message == nil || ev.Message.Actor != common.ActorSystem {
		t.Fatalf("hint was not recorded as a system message: %+v", ev)
	}
	if len(ev.Message.Parts) != 1 || ev.Message.Parts[0] != (common.TextPart{Text: "Check the new tool, not the old one."}) {
		t.Fatalf("recorded hint parts = %#v", ev.Message.Parts)
	}
}

// Receiving the right completion still ends the wait without requiring any
// other mailbox message.
func TestToolWaitAcceptsCurrentCompletion(t *testing.T) {
	e := lostCallEngine(t)
	a := NewActor(e, e.parent)
	a.mb.Post(common.ToolCompleted{CallID: "current", Result: "done"})
	if !a.waitForTool("current") {
		t.Fatal("current completion did not finish the wait")
	}
}

// An interrupted tool can finish while the actor is idle. Consuming its result
// must not requeue it forever or leave it to satisfy the next tool's wait.
func TestActorConsumesIdleToolCompletion(t *testing.T) {
	e := lostCallEngine(t)
	a := NewActor(e, e.parent)
	var finished []common.ToolFinished
	a.Attach(ObserverFunc(func(obs common.Observation) {
		if f, ok := obs.(common.ToolFinished); ok {
			finished = append(finished, f)
		}
	}))
	a.handle(common.ToolCompleted{CallID: "old", Result: "late failure", IsError: true})
	if pending := a.mb.Drain(); len(pending) != 0 {
		t.Fatalf("idle completion was requeued: %+v", pending)
	}
	if len(finished) != 1 || finished[0].CallID != "old" || finished[0].Result != "late failure" || !finished[0].IsError {
		t.Fatalf("idle completion was not reported intact: %+v", finished)
	}
}
