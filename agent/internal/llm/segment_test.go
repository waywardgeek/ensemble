package llm

import (
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// build a context of alternating dialogue and checkpoints.
func convo(t *testing.T, spec ...any) *common.Context {
	t.Helper()
	c := common.NewContext()
	seq := common.Seq(0)
	for _, item := range spec {
		seq++
		switch v := item.(type) {
		case string: // a checkpoint note
			if err := c.Apply(common.Event{Seq: seq, Type: common.MicroHandoff, Handoff: &common.MicroHandoffData{Text: v}}); err != nil {
				t.Fatalf("handoff: %v", err)
			}
		case int: // a dialogue entry of v bytes
			text := make([]byte, v)
			for i := range text {
				text[i] = 'x'
			}
			if err := c.Apply(common.Event{Seq: seq, Type: common.MessageReceived, Message: &common.MessageData{
				Actor: common.ActorHuman, Parts: common.PartList{common.TextPart{Text: string(text)}},
			}}); err != nil {
				t.Fatalf("message: %v", err)
			}
		}
	}
	return c
}

func TestSegmentsSplitAtCheckpoints(t *testing.T) {
	c := convo(t, 100, 100, "first checkpoint", 50, "second checkpoint", 70)
	segs := Segments(c)
	if len(segs) != 3 {
		t.Fatalf("got %d segments, want 3: %+v", len(segs), segs)
	}
	if segs[0].Bytes != 200 || segs[1].Bytes != 50 || segs[2].Bytes != 70 {
		t.Errorf("segment sizes %d/%d/%d, want 200/50/70", segs[0].Bytes, segs[1].Bytes, segs[2].Bytes)
	}
}

// The newest segment has no closing checkpoint: it is what the agent is
// working on. Compressing it would take away the detail in active use.
func TestSelectNeverTakesTheLiveSegment(t *testing.T) {
	c := convo(t, 100, "checkpoint", 5000)
	sel, ok := SelectForCompression(c, 50)
	if !ok {
		t.Fatal("expected a selection")
	}
	if sel.Bytes != 100 {
		t.Errorf("selected %d bytes, want the 100-byte closed segment, not the live one", sel.Bytes)
	}
	if sel.Split {
		t.Error("selection was marked split when a whole segment was available")
	}
}

// Selection takes whole segments, oldest first, and stops as soon as it has
// enough. It must not take a fourth segment to overshoot a target three
// already met.
func TestSelectTakesWholeSegmentsOldestFirst(t *testing.T) {
	c := convo(t, 100, "a", 100, "b", 100, "c", 100, "d", 40)
	sel, ok := SelectForCompression(c, 150)
	if !ok {
		t.Fatal("expected a selection")
	}
	if sel.Bytes != 200 {
		t.Errorf("selected %d bytes, want 200 (two whole segments)", sel.Bytes)
	}
	if sel.From != 1 {
		t.Errorf("selection starts at seq %d, want 1 (oldest first)", sel.From)
	}
	if sel.Split {
		t.Error("selection was marked split when whole segments sufficed")
	}
}

// With no checkpoint anywhere, there is no whole segment to take. Cutting one
// is the degraded path and must be reported, not hidden.
// TestSelectRefusesWhenNothingIsClosed pins the rule that work in progress
// is never compacted.
//
// An earlier version of this test asserted the opposite: that a conversation
// with no checkpoint would be cut rather than allowed to grow. That was the
// wrong trade. The live segment holds the detail the agent is reasoning with
// right now, and summarising it takes away the thing being used. Nothing has
// to be cut here, because the forcing rule removes every tool but
// micro_handoff before the window runs out, which turns this case into the
// ordinary one.
func TestSelectRefusesWhenNothingIsClosed(t *testing.T) {
	c := convo(t, 100, 100, 100)
	if sel, ok := SelectForCompression(c, 150); ok {
		t.Fatalf("compacted work in progress: took %d bytes from %d to %d", sel.Bytes, sel.From, sel.To)
	}
}

// TestSelectCutsASegmentTooBigToCompressWhole covers the one path that
// breaks the whole-segment rule, and checks that it admits to it.
func TestSelectCutsASegmentTooBigToCompressWhole(t *testing.T) {
	c := convo(t, 100, 100, 100, "cp")

	// One closed segment of 300 bytes, asked for a target so small that
	// taking the segment whole would mean handing the compressor far more
	// than it asked for.
	sel, ok := SelectForCompression(c, 10)
	if !ok {
		t.Fatal("expected a selection")
	}
	if !sel.Split {
		t.Error("cutting inside a segment was not reported as a split")
	}
	if sel.Bytes < 10 {
		t.Errorf("split selection took %d bytes, want at least 10", sel.Bytes)
	}
}

// An empty conversation has nothing to compress, which is the ordinary case
// on most checkpoints.
func TestSelectDeclinesWhenThereIsNothingToTake(t *testing.T) {
	c := common.NewContext()
	if _, ok := SelectForCompression(c, 100); ok {
		t.Error("selected a span from an empty conversation")
	}
}

// Memory must never be selected for compression: it is not conversation, and
// it is removed only by its own verb.
func TestSelectIgnoresMemoryBands(t *testing.T) {
	c := common.NewContext()
	if err := c.Apply(common.Event{Seq: 1, Type: common.BandPopulated, BandAdd: &common.BandPopulatedData{
		Band: common.BandSession, File: common.MemoryFileID{Date: "2026-09-26", Num: 1},
		Text: "a large and ancient memory that must not be recompressed",
	}}); err != nil {
		t.Fatalf("populate: %v", err)
	}
	if segs := Segments(c); len(segs) != 0 {
		t.Errorf("memory produced %d conversation segments, want 0: %+v", len(segs), segs)
	}
}
