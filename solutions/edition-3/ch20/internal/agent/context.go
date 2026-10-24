package agent

import (
	"context"
	"encoding/json"
	"ensemble/internal/common"
	"fmt"
	"strings"
)

const defaultContextTarget = 400000

func effectiveTarget(n int) int {
	if n == 0 {
		return defaultContextTarget
	}
	return max(n, 20000)
}
func (a *agent) SetContextTarget(n int) error {
	if n < 0 {
		return fmt.Errorf("negative context_target")
	}
	if n != 0 {
		n = effectiveTarget(n)
	}
	a.configMu.Lock()
	a.config.ContextTarget = n
	a.configMu.Unlock()
	return nil
}

type contextRequest struct {
	ctx             context.Context
	operation, text string
	reply           chan contextResult
}
type contextResult struct {
	text string
	err  error
}

func (a *agent) CompactContext(ctx context.Context, operation, text string) (string, error) {
	reply := make(chan contextResult, 1)
	if err := a.box.post(contextRequest{ctx, operation, text, reply}, true); err != nil {
		return "", err
	}
	select {
	case r := <-reply:
		return r.text, r.err
	case <-ctx.Done():
		return "", ctx.Err()
	case <-a.done:
		return "", common.ErrStopped
	}
}
func (a *agent) contextRequest(r contextRequest) {
	result := contextResult{}
	if err := r.ctx.Err(); err != nil {
		result.err = err
	} else {
		switch r.operation {
		case "keep_tool_results":
			result.text = "The previous tool-result batch is kept for the following round trip."
		case "micro_handoff":
			if strings.TrimSpace(r.text) == "" {
				result.err = fmt.Errorf("handoff text must not be blank")
				break
			}
			// The ordinary receipt is recorded first by returned; only then is this fact appended.
			a.handoffText = r.text
			result.text = "Checkpoint accepted; the complete batch will be cleared after its last receipt."
		case "compact_context":
			result.err = a.compactBands()
			result.text = "Context tool-traffic bands checked; only complete, already presented batches can be cleared."
		default:
			result.err = fmt.Errorf("unknown context operation")
		}
	}
	r.reply <- result
}

type toolBatch struct {
	from, to       common.Seq
	whole, small   int
	complete, live bool
	results        []common.Entry
}

func toolBatches(c common.Context, live func(string) bool) []toolBatch {
	var batches []toolBatch
	var batch *toolBatch
	pending := map[string]bool{}
	calls := map[string]int{}
	for _, entry := range c.Dialogue {
		if entry.Kind != "" && entry.Kind != common.KindDialogue {
			continue
		}
		hasCall := false
		for _, p := range entry.Parts {
			if p.Type == "tool_call" {
				hasCall = true
			}
		}
		if hasCall {
			batches = append(batches, toolBatch{from: entry.Seq})
			batch = &batches[len(batches)-1]
			pending = map[string]bool{}
			calls = map[string]int{}
		}
		if batch == nil {
			continue
		}
		hasResult := false
		for _, p := range entry.Parts {
			if p.Type != "tool_call" && p.Type != "tool_result" {
				continue
			}
			data, _ := json.Marshal(p)
			if p.Type == "tool_call" {
				pending[p.CallID] = true
				calls[p.CallID] = len(data)
				batch.whole += len(data)
			} else {
				delete(pending, p.CallID)
				hasResult = true
				if p.Stub != "" || len(data) < 256 {
					batch.whole -= calls[p.CallID]
					batch.small += len(data) + calls[p.CallID]
				} else {
					batch.whole += len(data)
				}
			}
			if live != nil && live(p.CallID) {
				batch.live = true
			}
		}
		if hasResult {
			batch.results = append(batch.results, entry)
		}
		if hasCall || hasResult {
			batch.to = entry.Seq
		}
		batch.complete = len(pending) == 0
	}
	return batches
}

// A model has consumed a batch only after RequestSent records its numeric boundary.
// Watermarks are resolved here; replay never consults current policy settings.
func (a *agent) compactBands() error {
	target := effectiveTarget(a.Config().ContextTarget)
	for _, band := range []struct {
		level  common.Redaction
		budget int
	}{{common.RedactResult, target / 8}, {common.RedactTool, target / 16}} {
		c := a.history.Context()
		batches := toolBatches(c, a.liveToolCall)
		total := 0
		for _, b := range batches {
			if band.level == common.RedactResult {
				total += b.whole
			} else {
				total += b.small
			}
		}
		if total <= 2*band.budget {
			continue
		}
		var first, last common.Seq
		for _, b := range batches {
			if total <= band.budget {
				break
			}
			if !b.complete || b.to > c.SentThrough || b.live || (c.KeepAt == c.SentThrough && b.from == c.KeptBatch) {
				break
			}
			size := b.whole
			if band.level == common.RedactTool {
				size = b.small
			}
			if size == 0 {
				continue
			}
			if first == 0 {
				first = b.from
			}
			last = b.to
			total -= size
		}
		if last != 0 {
			if _, err := a.record(common.Event{Type: common.Redacted, Redact: &common.RedactData{From: first, To: last, Level: band.level}}); err != nil {
				return err
			}
		}
	}
	return nil
}

// RequestSent advances this one-shot window. The bands may still reclaim old
// kept results, but this round-trip rule does not reconsider them.
func (a *agent) stubPresentedResults() error {
	cfg := a.Config()
	model := cfg.ResolvedModel
	if model == "" {
		model = cfg.Model
	}
	features, _ := common.LookupModel(model)
	if !features.StubsToolResults {
		return nil
	}
	c := a.history.Context()
	threshold := effectiveTarget(cfg.ContextTarget) / 100
	for _, b := range toolBatches(c, a.liveToolCall) {
		if !b.complete || b.to > c.SentThrough || b.to <= c.PreviousSentThrough || b.live || (c.KeepAt == c.SentThrough && b.from == c.KeptBatch) {
			continue
		}
		for _, entry := range b.results {
			var indexes []int
			for i, p := range entry.Parts {
				if p.Type != "tool_result" || p.Name == "keep_tool_results" || p.Stub != "" {
					continue
				}
				data, _ := json.Marshal(p.Parts)
				if len(data) > threshold {
					indexes = append(indexes, i)
				}
			}
			if len(indexes) > 0 {
				if _, err := a.record(common.Event{Type: common.Redacted, Redact: &common.RedactData{From: entry.Seq, To: entry.Seq, Parts: indexes, Level: common.RedactResult}}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
func (a *agent) prepareContext() (common.Context, error) {
	if err := a.syncDeclarations(); err != nil {
		return common.Context{}, err
	}
	if err := a.stubPresentedResults(); err != nil {
		return common.Context{}, err
	}
	if err := a.compactBands(); err != nil {
		return common.Context{}, err
	}
	return a.history.Context(), nil
}
func (a *agent) syncDeclarations() error {
	if a.skills == nil {
		return nil
	}
	c := a.history.Context()
	desired := a.skills.Snapshot().Definitions
	current := map[string]common.ToolDeclaration{}
	for _, t := range c.Tools {
		current[t.Name] = t
	}
	delta := &common.ToolsData{}
	for _, t := range desired {
		d := common.ToolDeclaration{Name: t.Name, Description: t.Description, Schema: t.Parameters}
		old, ok := current[t.Name]
		left, _ := json.Marshal(old)
		right, _ := json.Marshal(d)
		if !ok || string(left) != string(right) {
			delta.Added = append(delta.Added, d)
		}
		delete(current, t.Name)
	}
	for _, t := range c.Tools {
		if _, ok := current[t.Name]; ok {
			delta.Removed = append(delta.Removed, t.Name)
		}
	}
	if c.ToolsInitialized && len(delta.Added) == 0 && len(delta.Removed) == 0 {
		return nil
	}
	_, err := a.record(common.Event{Type: common.ToolsChanged, Tools: delta})
	return err
}

func (a *agent) liveToolCall(callID string) bool { return a.liveCalls[callID] }

type factRequest struct {
	event common.Event
	reply chan error
}

// RecordFact serializes worker-owned skill facts with admission and publication.
// Root may call it before Start while construction is still single-threaded.
func (a *agent) RecordFact(e common.Event) error {
	if !a.started.Load() {
		return a.history.Append(e)
	}
	reply := make(chan error, 1)
	if err := a.box.post(factRequest{e, reply}, true); err != nil {
		return err
	}
	select {
	case err := <-reply:
		return err
	case <-a.done:
		return common.ErrStopped
	}
}
