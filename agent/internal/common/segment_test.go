package common

import "testing"

// build a context of alternating dialogue and checkpoints.
func convo(t *testing.T, spec ...any) *Context {
	t.Helper()
	c := NewContext()
	seq := Seq(0)
	for _, item := range spec {
		seq++
		switch v := item.(type) {
		case string: // a checkpoint note
			if err := c.Apply(Event{Seq: seq, Type: MicroHandoff, Handoff: &MicroHandoffData{Text: v}}); err != nil {
				t.Fatalf("handoff: %v", err)
			}
		case int: // a dialogue entry of v bytes
			text := make([]byte, v)
			for i := range text {
				text[i] = 'x'
			}
			if err := c.Apply(Event{Seq: seq, Type: MessageReceived, Message: &MessageData{
				Actor: ActorHuman, Parts: PartList{TextPart{Text: string(text)}},
			}}); err != nil {
				t.Fatalf("message: %v", err)
			}
		}
	}
	return c
}

func TestSegmentsSplitAtCheckpoints(t *testing.T) {
	c := convo(t, 100, 100, "first checkpoint", 50, "second checkpoint", 70)
	segs := c.Segments()
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
	sel, ok := c.SelectForCompression(50)
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
	sel, ok := c.SelectForCompression(150)
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
func TestSelectSplitsOnlyWhenNothingIsClosed(t *testing.T) {
	c := convo(t, 100, 100, 100)
	sel, ok := c.SelectForCompression(150)
	if !ok {
		t.Fatal("expected a selection")
	}
	if !sel.Split {
		t.Error("cutting a live segment was not reported as a split")
	}
	if sel.Bytes < 150 {
		t.Errorf("split selection took %d bytes, want at least 150", sel.Bytes)
	}
}

// An empty conversation has nothing to compress, which is the ordinary case
// on most checkpoints.
func TestSelectDeclinesWhenThereIsNothingToTake(t *testing.T) {
	c := NewContext()
	if _, ok := c.SelectForCompression(100); ok {
		t.Error("selected a span from an empty conversation")
	}
}

// Memory must never be selected for compression: it is not conversation, and
// it is removed only by its own verb.
func TestSelectIgnoresMemoryBands(t *testing.T) {
	c := NewContext()
	if err := c.Apply(Event{Seq: 1, Type: BandPopulated, BandAdd: &BandPopulatedData{
		Band: BandSession, File: MemoryFileID{Date: "2026-09-26", Num: 1},
		Text: "a large and ancient memory that must not be recompressed",
	}}); err != nil {
		t.Fatalf("populate: %v", err)
	}
	if segs := c.Segments(); len(segs) != 0 {
		t.Errorf("memory produced %d conversation segments, want 0: %+v", len(segs), segs)
	}
}
