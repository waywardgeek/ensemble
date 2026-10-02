package llm

import (
	"path/filepath"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// countingRecaller records whether retrieval was asked for at all.
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

// Auto recall keeps running for a model that cannot take ephemera, and this
// test exists to stop someone pairing the two.
//
// They look like the same category and they are not. The criterion is whether
// the content survives into the next request. Ephemera appear once and vanish,
// which rewrites a prefix the model has already seen. A recall entry is
// appended to the dialogue, rendered inline in conversation order by every
// vendor, and removed only at a checkpoint, which already rewrites the prefix
// and is already paying for that cache miss. So recall only ever extends the
// prefix, and a model that tolerates no change to it is unaffected.
//
// This is a property of THIS design, not of auto recall as an idea. The same
// feature built as ephemera, delivered fresh each turn and gone by the next
// round trip, would have to be switched off for such a model. Gating recall
// here would have cost the feature for no caching benefit at all.
func TestAutoRecallRunsForModelThatCannotTakeEphemera(t *testing.T) {
	if f, ok := common.LookupModel("claude-opus-5-5"); !ok || !f.NoEphemera {
		t.Fatalf("precondition: claude-opus-5-5 must declare NoEphemera, got ok=%v", ok)
	}

	r := &countingRecaller{}
	eng := newRecallEngine(t, "claude-opus-5-5", r)

	eng.attachRecall("some query")

	if r.calls != 1 {
		t.Errorf("Recall was called %d time(s); want 1. Recall is appended and permanent, so it does not destabilise the prefix and must not be gated on NoEphemera", r.calls)
	}
}

// The ordinary case, so that a recall feature deleted outright cannot pass the
// test above by accident.
func TestAutoRecallRunsForOtherModels(t *testing.T) {
	r := &countingRecaller{}
	eng := newRecallEngine(t, "claude-opus-5", r)

	eng.attachRecall("some query")

	if r.calls != 1 {
		t.Errorf("Recall was called %d time(s); want 1", r.calls)
	}
}
