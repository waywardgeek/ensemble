package agent

import (
	"ensemble/internal/common"
	"testing"
)

func TestLosslessObservationQueue(t *testing.T) {
	a := &agent{queues: make(map[*observationQueue]struct{})}
	q := a.Observe()
	defer q.Close()
	// Exercise more than the compatibility subscription's bounded buffer.
	for i := 0; i < 4096; i++ {
		for queue := range a.queues {
			queue.push(common.PartDelta{PartID: uint64(i), Kind: common.DeltaText, Chunk: "x"})
		}
	}
	got := q.Drain()
	if len(got) != 4096 {
		t.Fatalf("lost observations: %d", len(got))
	}
	for i, o := range got {
		if o.(common.PartDelta).PartID != uint64(i) {
			t.Fatal("reordered observations")
		}
	}
	if len(q.Drain()) != 0 {
		t.Fatal("duplicate delivery")
	}
}
