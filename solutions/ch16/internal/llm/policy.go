package llm

// Chapter 15's context policy: decide, before each request, what leaves the
// window, and record every decision as an ordinary Redacted event.
//
// Nothing here edits the context. It emits events through Record, the
// reducer applies them, and a replay applies them again. That is what makes
// the cuts reproducible after the settings that chose them have changed: the
// event stores the Seq numbers, and nothing ever recomputes them.
//
// It runs in Turn, which both dispatchers (Engine.Execute for the
// synchronous path, Actor.dispatchTool for the actor) go through, so there is
// one policy, not two.

import (
	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// Reasons written into the events. For a human reading the log; nothing
// branches on them.
const (
	reasonRoundTrip = "round trip: not kept"
	reasonLadder    = "ladder"
)

func (e *Engine) target() int {
	if e.Target == nil {
		return 0
	}
	return e.Target()
}

// curate records this request's cuts. Rule 6 first (it only touches the
// newest batch), then rule 7 (which walks the whole window).
func (e *Engine) curate() error {
	feats, _ := common.LookupModel(e.Cfg.Model)
	b := common.BudgetsFor(e.target())
	lastReq, prevReq := e.lastTwoRequests()
	if feats.StubsToolResults {
		if err := e.stubUnkeptBatch(b, lastReq, prevReq); err != nil {
			return err
		}
	}
	return e.ladder(b, lastReq)
}

// lastTwoRequests returns the log indices of the two newest RequestSent
// events, -1 where there is none.
func (e *Engine) lastTwoRequests() (last, prev int) {
	last, prev = -1, -1
	ev := e.Log.Events
	for i := len(ev) - 1; i >= 0; i-- {
		if ev[i].Type != common.RequestSent {
			continue
		}
		if last < 0 {
			last = i
			continue
		}
		prev = i
		break
	}
	return last, prev
}

// stubUnkeptBatch is rule 6. The batch is the tool results that arrived
// between the previous two requests, which is to say the results the newest
// request carried for the first time. The model has now seen them once. If
// its response to that request called keep_tool_results, they stay;
// otherwise every one above the threshold becomes a stub in the request about
// to be built. keep_tool_results' own result is never stubbed: it is the
// acknowledgement of a decision, and it is judged by nobody.
func (e *Engine) stubUnkeptBatch(b common.Budgets, lastReq, prevReq int) error {
	if lastReq < 0 {
		return nil
	}
	ev := e.Log.Events
	for _, x := range ev[lastReq+1:] {
		if x.Type == common.ResponseEnded && x.Response != nil {
			for _, p := range x.Response.Parts {
				if c, ok := p.(common.ToolCallPart); ok && c.Name == common.KeepToolResults {
					return nil
				}
			}
			break
		}
	}
	var seqs []common.Seq
	for _, x := range ev[prevReq+1 : lastReq] {
		if x.Type == common.ToolReturned && x.Tool != nil && x.Tool.Name != common.KeepToolResults {
			seqs = append(seqs, x.Seq)
		}
	}
	for _, s := range seqs {
		if liveResultBytes(e.entryAt(s)) <= b.Threshold {
			continue
		}
		if err := e.Record(common.Event{Type: common.Redacted, Redact: &common.RedactData{
			From: s, To: s, Level: common.RedactResult, Reason: reasonRoundTrip,
		}}); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) entryAt(s common.Seq) *common.Entry {
	for i := range e.Ctx.Dialogue {
		d := &e.Ctx.Dialogue[i]
		if d.Seq == s && d.Kind == common.KindDialogue {
			return d
		}
	}
	return nil
}

// ladder is rule 7. Two bands above the live tail:
//
//	[ dialogue only | calls, results stubbed | full tool traffic ]
//	     RedactTool        RedactResult          (the tail)
//
// A band is left alone until it passes twice its budget, then one event cuts
// it back to one budget. Steps, not a trickle: every cut re-sends everything
// after it uncached, so one larger cut now and then costs less than a small
// one on every request.
//
// Entries the model has not yet been sent are never cut. A result that
// arrived after the last request goes out whole at least once.
func (e *Engine) ladder(b common.Budgets, lastReq int) error {
	if lastReq < 0 {
		return nil
	}
	seen := e.Log.Events[lastReq].Seq

	// Results band.
	d := e.Ctx.Dialogue
	sizes := make([]int, len(d))
	total := 0
	for i := range d {
		sizes[i] = liveResultBytes(&d[i])
		total += sizes[i]
	}
	if total > 2*b.Results {
		if cut := tailCut(d, sizes, b.Results, seen); cut >= 0 {
			if from, ok := firstWith(d, sizes, cut); ok {
				if err := e.Record(common.Event{Type: common.Redacted, Redact: &common.RedactData{
					From: from, To: d[cut].Seq, Level: common.RedactResult, Reason: reasonLadder,
				}}); err != nil {
					return err
				}
			}
		}
	}

	// Calls band, measured after the results cut has landed.
	d = e.Ctx.Dialogue
	sizes = callBandBytes(d)
	total = 0
	for _, n := range sizes {
		total += n
	}
	if total <= 2*b.Calls {
		return nil
	}
	cut := tailCut(d, sizes, b.Calls, seen)
	if cut < 0 {
		return nil
	}
	cut = balancedAtOrBelow(d, cut)
	if cut < 0 {
		return nil
	}
	from, ok := firstWith(d, toolPartCounts(d), cut)
	if !ok {
		return nil
	}
	return e.Record(common.Event{Type: common.Redacted, Redact: &common.RedactData{
		From: from, To: d[cut].Seq, Level: common.RedactTool, Reason: reasonLadder,
	}})
}

// tailCut walks from the newest entry, keeping entries until the kept bytes
// would pass budget, and returns the index of the newest entry to cut: it and
// everything older goes. -1 if nothing need go. Survivors and entries the
// model has not seen are never the cut.
func tailCut(d []common.Entry, sizes []int, budget int, seen common.Seq) int {
	kept := 0
	for i := len(d) - 1; i >= 0; i-- {
		if d[i].Kind != common.KindDialogue {
			continue
		}
		if kept+sizes[i] > budget && d[i].Seq <= seen {
			return i
		}
		kept += sizes[i]
	}
	return -1
}

// firstWith returns the Seq of the oldest dialogue entry at or below cut that
// has something to cut, so the span starts where the band does.
func firstWith(d []common.Entry, sizes []int, cut int) (common.Seq, bool) {
	for i := 0; i <= cut; i++ {
		if d[i].Kind == common.KindDialogue && sizes[i] > 0 {
			return d[i].Seq, true
		}
	}
	return 0, false
}

// balancedAtOrBelow moves a RedactTool cut down until no call at or below it
// has its result above it. RedactTool removes calls and results together; a
// span that ended between a call and its result would leave the result
// without its call, which every vendor refuses.
func balancedAtOrBelow(d []common.Entry, cut int) int {
	resultAt := map[string]int{}
	for i, e := range d {
		for _, p := range e.Parts {
			if r, ok := p.(common.ToolResultPart); ok {
				resultAt[r.CallID] = i
			}
		}
	}
	// reach[i] = the furthest index any call in d[0..i] needs.
	reach := -1
	ok := make([]bool, len(d))
	for i, e := range d {
		for _, p := range e.Parts {
			if c, isCall := p.(common.ToolCallPart); isCall {
				at, answered := resultAt[c.CallID]
				if !answered {
					at = len(d) // never balanced past an open call
				}
				if at > reach {
					reach = at
				}
			}
		}
		ok[i] = reach <= i
	}
	for i := cut; i >= 0; i-- {
		if ok[i] && d[i].Kind == common.KindDialogue {
			return i
		}
	}
	return -1
}

// liveResultBytes counts the bytes of tool results in an entry that have not
// been stubbed.
func liveResultBytes(e *common.Entry) int {
	if e == nil || e.Kind != common.KindDialogue {
		return 0
	}
	n := 0
	for _, p := range e.Parts {
		r, ok := p.(common.ToolResultPart)
		if !ok || common.IsStubbed(r) {
			continue
		}
		for _, q := range r.Parts {
			if t, ok := q.(common.TextPart); ok {
				n += len(t.Text)
			}
		}
	}
	return n
}

// callBandBytes measures the middle band per entry: the arguments of calls
// whose results are stubs, plus the stubs themselves.
func callBandBytes(d []common.Entry) []int {
	stubbed := map[string]bool{}
	for _, e := range d {
		for _, p := range e.Parts {
			if r, ok := p.(common.ToolResultPart); ok && common.IsStubbed(r) {
				stubbed[r.CallID] = true
			}
		}
	}
	out := make([]int, len(d))
	for i, e := range d {
		if e.Kind != common.KindDialogue {
			continue
		}
		for _, p := range e.Parts {
			switch v := p.(type) {
			case common.ToolCallPart:
				if stubbed[v.CallID] {
					out[i] += len(v.Args)
				}
			case common.ToolResultPart:
				if common.IsStubbed(v) {
					out[i] += len(v.Parts[0].(common.RedactedPart).Stub)
				}
			}
		}
	}
	return out
}

// toolPartCounts marks entries that hold any tool call or result.
func toolPartCounts(d []common.Entry) []int {
	out := make([]int, len(d))
	for i, e := range d {
		if e.Kind != common.KindDialogue {
			continue
		}
		for _, p := range e.Parts {
			switch p.(type) {
			case common.ToolCallPart, common.ToolResultPart:
				out[i]++
			}
		}
	}
	return out
}
