package gui

import (
	"encoding/json"
	"ensemble"
	"fmt"
	"testing"
)

func TestLegacyDisplayIDsAreStableWithoutChangingHistory(t *testing.T) {
	events := []ensemble.Event{}
	for i, text := range []string{"FIRST", "SECOND", "zero"} {
		id := uint64(7)
		if i == 2 {
			id = 0
		}
		var e ensemble.Event
		raw := fmt.Sprintf(`{"seq":%d,"type":"response_ended","response":{"parts":[{"type":"text","display_id":%d,"text":%q}]}}`, i+1, id, text)
		if err := json.Unmarshal([]byte(raw), &e); err != nil {
			t.Fatal(err)
		}
		events = append(events, e)
	}
	ids := newDisplayIDs(events)
	used := map[uint64]bool{}
	for _, e := range events {
		a := ids.project("agent", e)
		b := ids.project("agent", e)
		if len(a) != 1 || a[0].PartID == 0 || used[a[0].PartID] || a[0].PartID != b[0].PartID {
			t.Fatalf("unstable/duplicate frames %+v", a)
		}
		used[a[0].PartID] = true
	}
	if events[0].Response.Parts[0].DisplayID != 7 || events[1].Response.Parts[0].DisplayID != 7 || events[2].Response.Parts[0].DisplayID != 0 {
		t.Fatal("canonical IDs mutated")
	}
	live := ids.partial(8)
	if used[live] {
		t.Fatal("live collision")
	}
}
