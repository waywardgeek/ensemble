package llm

import (
	"fmt"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// MisplacedCall is a tool call whose result EXISTS but does not sit where every
// vendor requires it: in the message immediately after the call.
//
// This is a different fault from a lost result, which lostCalls reports, and
// keeping the two apart is the whole reason this checker exists. A lost result
// never arrived, so the repair is to synthesize one. A misplaced result arrived
// late, after other entries had already landed between the call and its answer,
// so synthesizing anything would be wrong: the answer is right there, in the
// wrong place.
//
// Set membership cannot tell them apart. Asking "is this call answered
// anywhere?" returns yes for a misplaced result, so the fault goes unreported
// and any repair keyed on it silently does nothing. The rule is ADJACENCY, and
// it is quoted from the refusal itself: "Each tool_use block must have a
// corresponding tool_result block in the next message."
type MisplacedCall struct {
	CallID string
	Name   string

	// CallAt and ResultAt are dialogue indices, not Seqs. A reader chasing this
	// report wants to look at neighbouring entries, and neighbours are an index
	// apart, not a Seq apart.
	CallAt   int
	ResultAt int

	// BlockedAt is the first entry that landed between the call and its result,
	// and BlockedBy describes it. Naming the intruder is the point of the
	// report: it is the difference between "the context is malformed" and "a
	// human message landed while a tool was still running".
	BlockedAt int
	BlockedBy string
}

func (m MisplacedCall) String() string {
	return fmt.Sprintf(
		"call %s to %s is answered at entry %d, but a %s landed first at entry %d; the result must sit immediately after the call at entry %d",
		m.CallID, m.Name, m.ResultAt, m.BlockedBy, m.BlockedAt, m.CallAt)
}

// misplacedCalls reports every call whose result exists but is not adjacent to
// it, in dialogue order.
//
// It deliberately says nothing about a call with no result at all. That is
// lostCalls' business, and reporting it here too would make one fault produce
// two different diagnoses that a reader would have to reconcile.
func misplacedCalls(c *common.Context) []MisplacedCall {
	if c == nil {
		return nil
	}
	var out []MisplacedCall
	for i, entry := range c.Dialogue {
		want := callsIn(entry)
		if len(want) == 0 {
			continue
		}
		// Walk forward from the call. A batch's results may span more than one
		// tool entry, so several in a row are fine; anything that is not a
		// result breaks the adjacency the vendor requires.
		answered := make(map[string]bool, len(want))
		blockedAt, blockedBy := -1, ""
		for j := i + 1; j < len(c.Dialogue) && len(answered) < len(want); j++ {
			next := c.Dialogue[j]
			ids := resultIDsIn(next)
			if len(ids) > 0 {
				for _, id := range ids {
					answered[id] = true
				}
				continue
			}
			blockedAt, blockedBy = j, describeEntry(next)
			break
		}
		if blockedAt < 0 {
			// Either every call was answered adjacently, or the dialogue ended
			// first. A dialogue that ends mid-batch has lost its results, not
			// misplaced them.
			continue
		}
		for _, call := range want {
			if answered[call.CallID] {
				continue
			}
			at := findResultIndex(c, call.CallID)
			if at < 0 {
				continue // lost, not misplaced
			}
			out = append(out, MisplacedCall{
				CallID:    call.CallID,
				Name:      call.Name,
				CallAt:    i,
				ResultAt:  at,
				BlockedAt: blockedAt,
				BlockedBy: blockedBy,
			})
		}
	}
	return out
}

func callsIn(e common.Entry) []common.ToolCallPart {
	var calls []common.ToolCallPart
	for _, p := range e.Parts {
		if call, ok := p.(common.ToolCallPart); ok {
			calls = append(calls, call)
		}
	}
	return calls
}

func resultIDsIn(e common.Entry) []string {
	var ids []string
	for _, p := range e.Parts {
		if res, ok := p.(common.ToolResultPart); ok {
			ids = append(ids, res.CallID)
		}
	}
	return ids
}

func findResultIndex(c *common.Context, callID string) int {
	for i, e := range c.Dialogue {
		for _, p := range e.Parts {
			if res, ok := p.(common.ToolResultPart); ok && res.CallID == callID {
				return i
			}
		}
	}
	return -1
}

// describeEntry names an entry the way a person reading a diagnostic would:
// who spoke, and what kind of thing they said.
func describeEntry(e common.Entry) string {
	return fmt.Sprintf("%v %v", e.Actor, e.Kind)
}
