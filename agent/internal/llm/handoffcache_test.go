package llm

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// Tests for the breakpoint that sits on the compaction bound.
//
// The rolling pair covers the part of the conversation that grows. These cover
// the part that has stopped changing, which is the part a cache read is least
// likely to reach on its own: the read walks backward a bounded number of
// positions looking for a prior write, and a turn that appends a lot of blocks
// at once pushes the settled prefix out of that window.

// workEntries models a stretch of real work: a request, an agent turn that
// calls a tool, and the tool's result. The tool traffic is the point of the
// fixture. It is what a handoff strips, and the totality of that strip is what
// makes the bound stable.
func workEntries(seq common.Seq, label string) []common.Entry {
	return []common.Entry{
		{Seq: seq, Kind: common.KindDialogue, Actor: common.ActorHuman,
			Parts: []common.Part{common.TextPart{Text: "please handle " + label}}},
		{Seq: seq + 1, Kind: common.KindDialogue, Actor: common.ActorAgent,
			Parts: []common.Part{
				common.TextPart{Text: "working on " + label},
				common.ToolCallPart{CallID: "call-" + label, Name: "read_file",
					Args: json.RawMessage(`{"path":"` + label + `"}`)},
			}},
		{Seq: seq + 2, Kind: common.KindDialogue, Actor: common.ActorTool,
			Parts: []common.Part{common.ToolResultPart{CallID: "call-" + label,
				Parts: []common.Part{common.TextPart{Text: "contents of " + label}}}}},
	}
}

// applyHandoff compacts through the REAL reducer rather than hand-assembling
// what a compacted context looks like. A fixture built by hand would encode my
// belief about what a handoff leaves behind, and the code under test encodes
// the same belief; if that belief is wrong both agree and the test passes.
func applyHandoff(t *testing.T, c *common.Context, seq common.Seq, text string) {
	t.Helper()
	err := Apply(c, common.Event{
		Seq:     seq,
		Type:    common.MicroHandoff,
		Handoff: &common.MicroHandoffData{Text: text},
	})
	if err != nil {
		t.Fatalf("Apply(MicroHandoff): %v", err)
	}
}

// prefixThrough returns a comparable rendering of every block up to and
// including the message carrying needle. Cache markers are deliberately not
// part of it: markers move between turns by design, so comparing raw bytes
// would report every healthy turn as a rewrite.
func prefixThrough(t *testing.T, w anthWireMsgs, needle string) string {
	t.Helper()
	var b strings.Builder
	for _, m := range w.Messages {
		found := false
		for _, blk := range m.Content {
			b.WriteString(m.Role + "|" + blk.Type + "|" + blk.Text + "\n")
			if strings.Contains(blk.Text, needle) {
				found = true
			}
		}
		if found {
			return b.String()
		}
	}
	t.Fatalf("no message on the wire contained %q", needle)
	return ""
}

func ctxWith(entries ...[]common.Entry) *common.Context {
	c := &common.Context{}
	for _, e := range entries {
		c.Dialogue = append(c.Dialogue, e...)
	}
	return c
}

func TestHandoffCarriesItsOwnBreakpoint(t *testing.T) {
	c := ctxWith(workEntries(1, "alpha"))
	applyHandoff(t, c, 10, "HANDOFF-ONE summary of the work so far")
	// Two stretches of work since the compaction, which is the situation the
	// marker exists for. With only one, the rolling anchor lands on the
	// handoff message itself and the positions coincide; see the test below.
	c.Dialogue = append(c.Dialogue, workEntries(20, "beta")...)
	c.Dialogue = append(c.Dialogue, workEntries(30, "gamma")...)

	w := renderAnthWire(t, c)
	marks := w.marked()
	if len(marks) != 3 {
		t.Fatalf("want 3 breakpoints (handoff bound + rolling pair), got %d: %q", len(marks), marks)
	}
	var onHandoff bool
	for _, m := range marks {
		if strings.Contains(m, "HANDOFF-ONE") {
			onHandoff = true
		}
	}
	if !onHandoff {
		t.Errorf("no breakpoint landed on the handoff bound; marked blocks were %q", marks)
	}
}

// TestHandoffAtTheAnchorCollapses records that the three positions are allowed
// to coincide. Immediately after a compaction the rolling anchor resolves to
// the handoff message itself, both calls mark the same block, and the request
// carries one marker rather than two. That is why the placement code needs no
// "is this the same position" guard: setting the marker twice is idempotent,
// and a guard would be untested code protecting against nothing.
func TestHandoffAtTheAnchorCollapses(t *testing.T) {
	c := ctxWith(workEntries(1, "alpha"))
	applyHandoff(t, c, 10, "HANDOFF-ONE summary of the work so far")
	c.Dialogue = append(c.Dialogue, workEntries(20, "beta")...)

	marks := renderAnthWire(t, c).marked()
	if len(marks) != 2 {
		t.Fatalf("want 2 distinct breakpoints when the bound and the anchor "+
			"coincide, got %d: %q", len(marks), marks)
	}
	if !strings.Contains(strings.Join(marks, "\n"), "HANDOFF-ONE") {
		t.Errorf("the surviving marker should still sit on the handoff bound, got %q", marks)
	}
}

func TestNoHandoffLeavesTheRollingPairAlone(t *testing.T) {
	c := ctxWith(workEntries(1, "alpha"), workEntries(10, "beta"))

	marks := renderAnthWire(t, c).marked()
	if len(marks) != 2 {
		t.Fatalf("a conversation that has never been compacted should carry only "+
			"the rolling pair, got %d breakpoints: %q", len(marks), marks)
	}
}

// TestHandoffPrefixSurvivesALaterHandoff is the claim the whole breakpoint
// rests on: a cache entry written behind one handoff is still readable after
// the conversation has been compacted again.
//
// This is exactly the case that fails when a handoff keeps a few recent tool
// pairs behind the boundary as evidence that tools exist. Those pairs are
// counted back from a boundary that moves to the newest handoff, so the pairs
// surviving behind an OLDER handoff are dropped as soon as a NEWER one exists,
// the prefix ending at the older handoff changes, and its entry dies. It fires
// on every compaction after the first, which is most of a long session.
func TestHandoffPrefixSurvivesALaterHandoff(t *testing.T) {
	c := ctxWith(workEntries(1, "alpha"))
	applyHandoff(t, c, 10, "HANDOFF-ONE summary of the work so far")
	c.Dialogue = append(c.Dialogue, workEntries(20, "beta")...)

	before := prefixThrough(t, renderAnthWire(t, c), "HANDOFF-ONE")

	// The session keeps going and compacts again.
	c.Dialogue = append(c.Dialogue, workEntries(30, "gamma")...)
	applyHandoff(t, c, 40, "HANDOFF-TWO a later summary")
	c.Dialogue = append(c.Dialogue, workEntries(50, "delta")...)

	after := prefixThrough(t, renderAnthWire(t, c), "HANDOFF-ONE")

	if before != after {
		t.Errorf("the prefix ending at the first handoff was rewritten by a later "+
			"compaction, so its cache entry is dead.\nbefore:\n%s\nafter:\n%s", before, after)
	}

	// The bound tracks the NEWEST handoff. An older one is still a sound
	// bound, so nothing breaks if this slips, but the cached span shrinks to
	// the oldest compaction and the marker quietly stops earning its slot.
	marks := strings.Join(renderAnthWire(t, c).marked(), "\n")
	if !strings.Contains(marks, "HANDOFF-TWO") {
		t.Errorf("the bound should have advanced to the newest handoff, got %q", marks)
	}
}

// TestHandoffBoundIsTerminal asserts the PROPERTY that makes the bound sound,
// not the placement that follows from it. The breakpoint is safe only because
// nothing behind a handoff can still be removed, and that fact is owned by
// land() in context_ops.go, a different file with different reasons to change.
// Asserting placement alone would let a change there silently turn every
// post-compaction request into a full-price one.
func TestHandoffBoundIsTerminal(t *testing.T) {
	c := ctxWith(workEntries(1, "alpha"), workEntries(10, "beta"))
	applyHandoff(t, c, 20, "HANDOFF-ONE summary of the work so far")
	c.Dialogue = append(c.Dialogue, workEntries(30, "gamma")...)

	w := renderAnthWire(t, c)
	for _, m := range w.Messages {
		for _, blk := range m.Content {
			if strings.Contains(blk.Text, "HANDOFF-ONE") {
				return // reached the bound with nothing disqualifying before it
			}
			if blk.Type == "tool_use" || blk.Type == "tool_result" {
				t.Fatalf("a %s block survives behind the handoff bound, so the "+
					"region is not the terminal state of the redaction ladder and a "+
					"breakpoint there will be invalidated by the next compaction", blk.Type)
			}
		}
	}
	t.Fatal("the handoff never appeared on the wire")
}
