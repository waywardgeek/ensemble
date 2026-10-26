package agent

import (
	"encoding/json"
	"reflect"

	"ensemble/internal/common"
)

func (r *runtime) contextTarget() int {
	target := r.owner.Config().ContextTarget
	if target == 0 {
		return 400000
	}
	return max(20000, target)
}

// Policy records decisions; only History mutates the projection. Comparing the
// visible registry at this boundary also covers application-owned MCP connects.
func (r *runtime) prepareContext() error {
	defer r.phase("context_preparation")()
	if err := r.syncMemory(); err != nil {
		return err
	}
	if err := r.forceCheckpoint(); err != nil {
		return err
	}
	c := r.contextSnapshot()
	if c.StartupTools != nil {
		visible := r.owner.Declarations()
		delta := common.ToolsChangedData{}
		old := make(map[string]common.ToolDeclaration)
		for _, d := range c.Tools {
			old[d.Name] = d
		}
		for _, d := range visible {
			if previous, exists := old[d.Name]; !exists || !reflect.DeepEqual(previous, d) {
				delta.Added = append(delta.Added, d)
			}
			delete(old, d.Name)
		}
		for _, d := range c.Tools {
			if _, gone := old[d.Name]; gone {
				delta.Removed = append(delta.Removed, d.Name)
			}
		}
		if len(delta.Added)+len(delta.Removed) > 0 {
			if err := r.append(common.Event{Type: "tools_changed", Tools: &delta}); err != nil {
				return err
			}
		}
	}
	if err := r.ladder("redact_result", r.contextTarget()/8); err != nil {
		return err
	}
	return r.ladder("redact_tool", r.contextTarget()/16)
}

// The model decides about the results it just saw, not the new calls beside
// keep_tool_results. Each request takes this small candidate snapshot once.
func (r *runtime) stubPrevious(t *turn) error {
	if !r.owner.engine.Features().StubsToolResults {
		return nil
	}
	for _, call := range t.calls {
		if call.Name == "keep_tool_results" {
			return nil
		}
	}
	for _, entry := range t.presented {
		for _, p := range entry.Parts {
			if p.Type != "tool_result" {
				continue
			}
			bytes, _ := json.Marshal(p.Parts)
			if len(bytes) <= r.contextTarget()/100 {
				continue
			}
			if err := r.append(common.Event{Type: "redacted", Redact: &common.RedactData{From: entry.Seq, To: entry.Seq, Level: "redact_result", Reason: "round trip: not kept"}}); err != nil {
				return err
			}
		}
	}
	return nil
}

// A batch is the smallest removable group: assistant calls share an entry and
// each must take its result with it. Sizes count serialized tool parts, never
// dialogue, instruction survivors, guessed tokens or external output files.
type toolBatch struct {
	last  uint64
	bytes int
	whole bool
	calls map[string]bool
}

func batches(c *common.Context) []toolBatch {
	var out []toolBatch
	var batch *toolBatch
	for _, entry := range c.Dialogue {
		for _, p := range entry.Parts {
			if p.Type == "tool_call" {
				if batch == nil || len(batch.calls) == 0 {
					out = append(out, toolBatch{calls: map[string]bool{}})
					batch = &out[len(out)-1]
				}
				batch.calls[p.CallID] = true
			}
			if batch == nil || (p.Type != "tool_call" && p.Type != "tool_result") {
				continue
			}
			if p.Type == "tool_result" {
				if !batch.calls[p.CallID] {
					continue
				}
				delete(batch.calls, p.CallID)
				for _, nested := range p.Parts {
					if nested.Type != "redacted" {
						batch.whole = true
					}
				}
			}
			data, _ := json.Marshal(p)
			batch.bytes += len(data)
			batch.last = max(batch.last, entry.Seq)
		}
	}
	return out
}

func (r *runtime) ladder(level string, budget int) error {
	c := r.contextSnapshot()
	var eligible []toolBatch
	total := 0
	for _, batch := range batches(c) {
		if len(batch.calls) != 0 || batch.last > c.SentThrough {
			continue
		}
		if batch.whole != (level == "redact_result") {
			continue
		}
		eligible = append(eligible, batch)
		total += batch.bytes
	}
	if total <= 2*budget {
		return nil
	}
	// Cut an old prefix in one step, not one event per entry. The recorded Seq
	// makes later replay independent of whatever target the user chooses then.
	var to uint64
	for _, batch := range eligible {
		if total <= budget {
			break
		}
		total -= batch.bytes
		to = batch.last
	}
	return r.append(common.Event{Type: "redacted", Redact: &common.RedactData{From: 0, To: to, Level: level, Reason: "tool-byte ladder"}})
}
