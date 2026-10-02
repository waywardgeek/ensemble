package llm

import (
	"encoding/json"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// The property Chapter 16 is graded on: the files on disk uniquely describe
// what memory looks like when it is loaded. Switch every band off, switch it
// back on in a deliberately wrong order, and the context must be the one you
// started with — byte for byte.
//
// This is the test that makes the canonical ordering load-bearing instead of
// decorative. An implementation that appends band entries as they arrive
// passes every other test in this file and fails this one.
func TestBandRestoreOutOfOrderIsIdentical(t *testing.T) {
	populate := func(c *common.Context, seq common.Seq, b common.Band, date string, num int, text string) {
		t.Helper()
		err := Apply(c, common.Event{Seq: seq, Type: common.BandPopulated, BandAdd: &common.BandPopulatedData{
			Band: b, File: common.MemoryFileID{Date: date, Num: num}, Text: text, Source: "startup",
		}})
		if err != nil {
			t.Fatalf("populate %s %s-%d: %v", b, date, num, err)
		}
	}

	// A context with all five bands and a scrap of conversation.
	original := common.NewContext()
	populate(original, 1, common.BandSoul, "", 0, "I would rather be honest than comfortable.")
	populate(original, 2, common.BandMemory, "", 0, "Bill owns the repo and pushes.")
	populate(original, 3, common.Band64x, "2026-09-01", 1, "Chapters one through nine shipped.")
	populate(original, 4, common.Band8x, "2026-09-20", 1, "The grader measured the ladder at 35%.")
	populate(original, 5, common.Band8x, "2026-09-22", 2, "Bands replaced the on/off switch.")
	populate(original, 6, common.BandSession, "2026-09-26", 1, "I found six dead settings fields.")
	populate(original, 7, common.BandSession, "2026-09-26", 2, "BlobPart never resolves to bytes.")
	if err := Apply(original, common.Event{Seq: 8, Type: common.MessageReceived, Message: &common.MessageData{
		Actor: common.ActorHuman, Parts: common.PartList{common.TextPart{Text: "carry on"}},
	}}); err != nil {
		t.Fatalf("message: %v", err)
	}

	before := mustJSON(t, original)

	// Switch every band off. Depopulating with no Thru means "all of it",
	// and the event names nothing it wiped.
	for _, b := range []common.Band{common.BandSoul, common.BandMemory, common.Band64x, common.Band8x, common.BandSession} {
		if err := Apply(original, common.Event{Seq: 100, Type: common.BandDepopulated, BandDrop: &common.BandDepopulatedData{
			Band: b, Reason: "disabled",
		}}); err != nil {
			t.Fatalf("depopulate %s: %v", b, err)
		}
	}
	for _, e := range original.Dialogue {
		if common.BandForKind(e.Kind) != 0 {
			t.Fatalf("band entry %s survived being switched off", e.Kind)
		}
	}

	// Switch them back on in an order chosen to be wrong in every way:
	// newest file first, and the bands themselves back to front.
	populate(original, 200, common.BandSession, "2026-09-26", 2, "BlobPart never resolves to bytes.")
	populate(original, 201, common.Band8x, "2026-09-22", 2, "Bands replaced the on/off switch.")
	populate(original, 202, common.BandSession, "2026-09-26", 1, "I found six dead settings fields.")
	populate(original, 203, common.BandMemory, "", 0, "Bill owns the repo and pushes.")
	populate(original, 204, common.Band8x, "2026-09-20", 1, "The grader measured the ladder at 35%.")
	populate(original, 205, common.Band64x, "2026-09-01", 1, "Chapters one through nine shipped.")
	populate(original, 206, common.BandSoul, "", 0, "I would rather be honest than comfortable.")

	if after := mustJSON(t, original); after != before {
		t.Errorf("restoring bands out of order produced a different context.\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// Landing the same file twice must say the same thing as landing it once,
// or "the files describe the memory" is false the moment a restore overlaps
// a band that was never emptied.
func TestBandPopulateIsIdempotent(t *testing.T) {
	c := common.NewContext()
	for i := 0; i < 3; i++ {
		if err := Apply(c, common.Event{Seq: common.Seq(i + 1), Type: common.BandPopulated, BandAdd: &common.BandPopulatedData{
			Band: common.BandSession, File: common.MemoryFileID{Date: "2026-09-26", Num: 1}, Text: "one memory",
		}}); err != nil {
			t.Fatalf("populate %d: %v", i, err)
		}
	}
	if got := len(c.Dialogue); got != 1 {
		t.Errorf("populating one file three times produced %d entries, want 1", got)
	}
}

// A graduation retires its sources oldest-first, through Thru inclusive, and
// leaves everything newer alone.
func TestBandDepopulateThruIsInclusiveAndOldestFirst(t *testing.T) {
	c := common.NewContext()
	for i := 1; i <= 4; i++ {
		if err := Apply(c, common.Event{Seq: common.Seq(i), Type: common.BandPopulated, BandAdd: &common.BandPopulatedData{
			Band: common.BandSession, File: common.MemoryFileID{Date: "2026-09-26", Num: i}, Text: "memory",
		}}); err != nil {
			t.Fatalf("populate %d: %v", i, err)
		}
	}
	thru := common.MemoryFileID{Date: "2026-09-26", Num: 2}
	if err := Apply(c, common.Event{Seq: 10, Type: common.BandDepopulated, BandDrop: &common.BandDepopulatedData{
		Band: common.BandSession, Thru: &thru, Reason: "graduation",
	}}); err != nil {
		t.Fatalf("graduation: %v", err)
	}
	var left []int
	for _, e := range c.Dialogue {
		if e.Kind == common.KindSession {
			left = append(left, e.File.Num)
		}
	}
	if len(left) != 2 || left[0] != 3 || left[1] != 4 {
		t.Errorf("after graduation through %s, session band holds %v, want [3 4]", thru, left)
	}
}

// A graduation that consumed nothing is a stale event wearing the costume of
// one that worked. It must say so.
func TestBandDepopulateGraduationMatchingNothingIsLoud(t *testing.T) {
	c := common.NewContext()
	thru := common.MemoryFileID{Date: "2026-09-26", Num: 9}
	err := Apply(c, common.Event{Seq: 1, Type: common.BandDepopulated, BandDrop: &common.BandDepopulatedData{
		Band: common.BandSession, Thru: &thru, Reason: "graduation",
	}})
	if err == nil {
		t.Error("a graduation that named no entry was applied silently")
	}
}

// Switching off a band that is already empty is ordinary, and must not be
// confused with the case above.
func TestBandDepopulateDisableOfEmptyBandIsQuiet(t *testing.T) {
	c := common.NewContext()
	if err := Apply(c, common.Event{Seq: 1, Type: common.BandDepopulated, BandDrop: &common.BandDepopulatedData{
		Band: common.BandSession, Reason: "disabled",
	}}); err != nil {
		t.Errorf("switching off an empty band complained: %v", err)
	}
}

// A populate with no bytes describes memory nothing can render. The reducer
// has no filesystem and no renderer dereferences anything, so this is our own
// builder being broken, and it must be impossible to ignore.
func TestBandPopulateWithoutBytesPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("band_populated with no text was accepted quietly")
		}
	}()
	c := common.NewContext()
	_ = Apply(c, common.Event{Seq: 1, Type: common.BandPopulated, BandAdd: &common.BandPopulatedData{
		Band: common.BandSession, File: common.MemoryFileID{Date: "2026-09-26", Num: 1},
	}})
}

// Tool clearing must not reach memory. A checkpoint strips the conversation's
// tool traffic; the bands are removed only by their own verb.
func TestMicroHandoffDoesNotClearBands(t *testing.T) {
	c := common.NewContext()
	if err := Apply(c, common.Event{Seq: 1, Type: common.BandPopulated, BandAdd: &common.BandPopulatedData{
		Band: common.BandSession, File: common.MemoryFileID{Date: "2026-09-26", Num: 1}, Text: "a memory",
	}}); err != nil {
		t.Fatalf("populate: %v", err)
	}
	if err := Apply(c, common.Event{Seq: 2, Type: common.MicroHandoff, Handoff: &common.MicroHandoffData{
		Text: "checkpoint",
	}}); err != nil {
		t.Fatalf("handoff: %v", err)
	}
	found := false
	for _, e := range c.Dialogue {
		if e.Kind == common.KindSession {
			found = true
		}
	}
	if !found {
		t.Error("micro_handoff cleared a memory band")
	}
}

func mustJSON(t *testing.T, c *common.Context) string {
	t.Helper()
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal context: %v", err)
	}
	return string(b)
}
