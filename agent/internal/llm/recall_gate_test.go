package llm

import (
	"path/filepath"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// countingRecaller records whether retrieval was asked for at all. The count
// is the point of the test: the requirement is not merely that recall fails to
// reach the wire, but that we never pay to build it.
type countingRecaller struct{ calls int }

func (r *countingRecaller) Recall(query string, convo []string) common.PartList {
	r.calls++
	return common.PartList{common.TextPart{Text: "RECALLED-NEEDLE-7"}}
}

func newRecallEngine(t *testing.T, model string, r common.Recaller) *Engine {
	t.Helper()
	v, _ := common.VendorFor(model)
	eng := NewEngine(common.Config{
		Model:     model,
		Vendor:    v,
		Surface:   common.DefaultSurface(v),
		BaseURL:   "https://vendor.test",
		APIKey:    "key",
		Endpoints: map[common.Vendor]common.Endpoint{v: {BaseURL: "https://vendor.test", APIKey: "key"}},
	}, filepath.Join(t.TempDir(), "journal.jsonl"), nil, nil, nil)
	eng.Recall = r
	return eng
}

// A model that cannot take ephemera gets no auto recall, and the retrieval is
// skipped rather than discarded.
//
// Recall is not ephemera in this design: it lands as a permanent dialogue
// entry with its own kind, so the render time drop that removes ephemera
// sails straight past it. It has to be refused at its source. Asserting the
// call count rather than the absence of a recall entry is what distinguishes
// "we did the work and threw it away" from "we never did it", and only the
// second one saves a BM25 pass and a judge call on every turn.
func TestNoAutoRecallForModelThatCannotTakeEphemera(t *testing.T) {
	r := &countingRecaller{}
	eng := newRecallEngine(t, "claude-opus-5-5", r)

	eng.attachRecall("some query")

	if r.calls != 0 {
		t.Errorf("Recall was called %d time(s); want 0 for a model that cannot take ephemera", r.calls)
	}
}

// The companion case. Without this, deleting recall entirely would pass the
// test above, and a gate that is always closed is indistinguishable from a
// feature that was removed.
func TestAutoRecallStillRunsForOtherModels(t *testing.T) {
	r := &countingRecaller{}
	eng := newRecallEngine(t, "claude-opus-5", r)

	eng.attachRecall("some query")

	if r.calls != 1 {
		t.Errorf("Recall was called %d time(s); want 1 for a model that accepts recall", r.calls)
	}
}
