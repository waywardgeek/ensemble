package llm

import (
	"fmt"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// replaceLostResult lets an observed outcome supersede an unknown one. Keep
// the standing result's position (adjacent to the call), but use the real
// result's sequence so later redactions refer to the bytes that actually won.
// An ordinary tool error is NOT a lost result and is never replaced here.
func replaceLostResult(c *common.Context, seq common.Seq, result common.ToolResultPart) bool {
	for i := range c.Dialogue {
		for j, p := range c.Dialogue[i].Parts {
			if old, ok := p.(common.ToolResultPart); ok && old.CallID == result.CallID && old.Lost {
				c.Dialogue[i].Parts[j] = result
				c.Dialogue[i].Seq = seq
				return true
			}
		}
	}
	return false
}

// repairLateResults migrates snapshots written before lost results carried
// provenance. The log, not the wording or error flag, identifies placeholders.
// Keep the log unchanged: both the premature loss and the return happened.
// With a retained log that no longer contains the loss event we cannot guess;
// new snapshots carry Lost themselves and need no log for this distinction.
func repairLateResults(c *common.Context, events []common.Event, asOf common.Seq, diag func(error)) {
	lost := make(map[common.Seq]string)
	for _, e := range events {
		if e.Seq <= asOf && e.Type == common.ToolResultLost && e.Tool != nil {
			lost[e.Seq] = e.Tool.CallID
		}
	}
	for i := range c.Dialogue {
		entry := &c.Dialogue[i]
		id, ok := lost[entry.Seq]
		if !ok {
			continue
		}
		for j, p := range entry.Parts {
			if result, ok := p.(common.ToolResultPart); ok && result.CallID == id {
				result.Lost = true
				entry.Parts[j] = result
			}
		}
	}
	// Remove only real-result parts successfully moved into a known placeholder.
	// Other parts in their entries, and every unrelated entry, are untouched.
	for i := 0; i < len(c.Dialogue); i++ {
		entry := c.Dialogue[i]
		var kept common.PartList
		changed := false
		for _, p := range entry.Parts {
			result, ok := p.(common.ToolResultPart)
			if ok && !result.Lost && replaceLostResult(c, entry.Seq, result) {
				changed = true
				if diag != nil {
					diag(fmt.Errorf("restore: call %s: replaced synthetic lost result with real result at seq %d", result.CallID, entry.Seq))
				}
			} else {
				kept = append(kept, p)
			}
		}
		if !changed {
			continue
		}
		if len(kept) != 0 {
			c.Dialogue[i].Parts = kept
		} else {
			c.Dialogue = append(c.Dialogue[:i], c.Dialogue[i+1:]...)
			i--
		}
	}
}
