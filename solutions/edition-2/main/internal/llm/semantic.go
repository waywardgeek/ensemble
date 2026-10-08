package llm

import (
	"encoding/json"
	"example.com/ensemble/internal/common"
	"reflect"
	"sort"
)

func cloneValue(v reflect.Value, collections bool) reflect.Value {
	if !v.IsValid() {
		return v
	}
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		out := reflect.New(v.Type().Elem())
		out.Elem().Set(cloneValue(v.Elem(), collections))
		return out
	case reflect.Slice:
		if v.IsNil() {
			if collections && v.Type() != reflect.TypeOf(json.RawMessage{}) {
				return reflect.MakeSlice(v.Type(), 0, 0)
			}
			return reflect.Zero(v.Type())
		}
		out := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		for i := 0; i < v.Len(); i++ {
			out.Index(i).Set(cloneValue(v.Index(i), collections))
		}
		return out
	case reflect.Map:
		if v.IsNil() {
			if collections {
				return reflect.MakeMap(v.Type())
			}
			return reflect.Zero(v.Type())
		}
		out := reflect.MakeMap(v.Type())
		it := v.MapRange()
		for it.Next() {
			out.SetMapIndex(it.Key(), cloneValue(it.Value(), collections))
		}
		return out
	case reflect.Struct:
		out := reflect.New(v.Type()).Elem()
		out.Set(v)
		for i := 0; i < v.NumField(); i++ {
			if out.Field(i).CanSet() && v.Type().Field(i).IsExported() {
				out.Field(i).Set(cloneValue(v.Field(i), collections))
			}
		}
		return out
	case reflect.Interface:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		out := reflect.New(v.Type()).Elem()
		out.Set(cloneValue(v.Elem(), collections))
		return out
	}
	return v
}
func guidanceIndex(c *common.Context, seq uint64) int {
	return sort.Search(len(c.Guidance), func(i int) bool { return c.Guidance[i].Seq >= seq })
}
func recordSemantic(owner common.Engine, c *common.Context, e common.Event) {
	switch e.Type {
	case "session_initialized":
		c.Session = e.Session
	case "turn_started":
		if c.Turns == nil {
			c.Turns = map[string]common.TurnFact{}
		}
		c.Turns[e.Turn.RequestID] = common.TurnFact{Index: e.Turn.RequestIndex, Start: e.Seq, Policy: e.Turn.Policy}
		if e.Turn.RequestIndex > c.RequestCursor {
			c.RequestCursor = e.Turn.RequestIndex
		}
	case "turn_ended":
		t := c.Turns[e.Turn.RequestID]
		t.End = e.Seq
		t.Outcome = e.Turn.Outcome
		c.Turns[e.Turn.RequestID] = t
	case "response_ended":
		r := e.Response
		c.Responses = append(c.Responses, common.ResponseFact{Seq: e.Seq, From: r.From, Requested: r.Requested, ModelReported: r.ModelReported, Usage: *r.Usage, RawUsage: r.RawUsage, StopReason: r.StopReason})
		if c.SkillMode {
			batch := false
			for _, p := range r.Parts {
				if p.Type == "tool_call" {
					batch = true
				}
			}
			if batch {
				for _, h := range c.Hints {
					if h.Anchor == 0 {
						i := guidanceIndex(c, h.Seq)
						if i < len(c.Guidance) {
							c.Guidance[i].Anchor = e.Seq
						}
					}
				}
			}
		}
	case "hint_received":
		c.Guidance = append(c.Guidance, common.GuidanceFact{Seq: e.Seq, Kind: "hint", Anchor: c.SkillBatch})
	case "message_received":
		if e.Message.Purpose == "ephemeral" {
			c.Guidance = append(c.Guidance, common.GuidanceFact{Seq: e.Seq, Kind: "ephemeral"})
		}
	case "request_sent":
		c.RequestSeqs = append(c.RequestSeqs, e.Seq)
		seqs := append(append([]uint64{}, e.Request.Hints...), e.Request.Ephemera...)
		for _, seq := range seqs {
			i := guidanceIndex(c, seq)
			if i < len(c.Guidance) && c.Guidance[i].Seq == seq {
				c.Guidance[i].ConsumedAt = e.Seq
			}
		}
	case "redacted":
		r := e.Redact
		c.Redactions = append(c.Redactions, common.RedactionFact{Seq: e.Seq, From: r.From, To: r.To, Level: r.Level, Reason: r.Reason})
	case "tool_limits_set", "tool_limits_consumed":
		kind := "set"
		if e.Type == "tool_limits_consumed" {
			kind = "consumed"
		}
		l := e.Limits
		c.LimitFacts = append(c.LimitFacts, common.LimitFact{Seq: e.Seq, Kind: kind, CallID: l.CallID, Name: l.Name, Overrides: l.Overrides})
	}
}
func Settled(owner common.Engine, c common.Context) bool {
	return c.TurnID == "" && c.Pending == nil && !c.Active && !unresolved(owner, c) && c.SkillBatch == 0 && len(c.DeferredSkills) == 0 && len(c.PendingSkills) == 0
}

// ValidateSemantic checks reduced relationships directly. It does not reconstruct
// an event archive or require the raw log prefix that a snapshot import omits.
func ValidateSemantic(owner common.Engine, c common.Context) error {
	ReindexContext(owner, &c)
	if c.Index.Entries > 1000000 || c.Index.Parts > 1000000 {
		return &common.SessionError{Code: "session_corrupt", Detail: "semantic entry or part limit"}
	}
	bad := func(detail string) error { return &common.SessionError{Code: "session_corrupt", Detail: detail} }
	if c.LastSeq == 0 || c.Session == nil || c.Session.Identity == nil {
		return bad("semantic session metadata missing")
	}
	if c.Session.OriginAsOf != 0 || c.Session.OriginSHA256 != "" || c.Session.HighWatermarks != nil {
		return bad("semantic creation fact has anchor fields")
	}
	calls := map[string]common.Part{}
	callSeq := map[string]uint64{}
	results := map[string]bool{}
	resultSeq := map[string]uint64{}
	responses := map[uint64]bool{}
	entries := append(append(append(append([]common.Entry{}, c.Entries...), c.Instructions...), c.Ephemera...), c.Hints...)
	if len(entries) > 1000000 || len(c.Calls) > 1000000 || len(c.TurnIDs) > 1000000 {
		return bad("semantic collection limit")
	}
	parts := 0
	for _, e := range entries {
		if e.Seq == 0 || e.Seq > c.LastSeq || e.Anchor > c.LastSeq {
			return bad("invalid entry coordinate")
		}
		if e.Purpose != "skill" && (e.Activation != 0 || e.SkillName != "") {
			return bad("misplaced skill coordinate")
		}
		if e.Purpose != "skill" && e.Purpose != "hint" && e.Anchor != 0 {
			return bad("misplaced batch anchor")
		}
		switch e.Purpose {
		case "dialogue":
			if e.Actor != "human" && e.Actor != "agent" && e.Actor != "tool" {
				return bad("invalid dialogue actor")
			}
		case "instruction", "ephemeral", "skill":
			if e.Actor != "system" {
				return bad("invalid system entry")
			}
		case "hint":
			if e.Actor != "human" {
				return bad("invalid hint")
			}
		default:
			return bad("invalid entry purpose")
		}
		if e.Actor == "agent" {
			responses[e.Seq] = true
		}
		for _, p := range e.Parts {
			parts++
			if parts > 1000000 {
				return bad("semantic part limit")
			}
			if reason := semanticPart(owner, p, false); reason != "" {
				return bad(reason)
			}
			switch p.Type {
			case "tool_call":
				if _, ok := calls[p.CallID]; ok || e.Actor != "agent" {
					return bad("duplicate/misplaced call")
				}
				calls[p.CallID] = p
				callSeq[p.CallID] = e.Seq
			case "tool_result":
				if results[p.CallID] || e.Actor != "tool" {
					return bad("duplicate/misplaced result")
				}
				results[p.CallID] = true
				resultSeq[p.CallID] = e.Seq
				parts += len(p.Parts)
			}
		}
	}
	if len(calls) != len(c.Calls) {
		return bad("seen call set mismatch")
	}
	for id, state := range c.Calls {
		p, ok := calls[id]
		if !ok || !reflect.DeepEqual(p, state.Part) || state.Part.CallID != id || state.Returned != results[id] || state.Returned && !state.Dispatched {
			return bad("call pairing mismatch")
		}
		if state.Dispatched != (state.CalledAt != 0) || state.Returned != (state.ReturnedAt != 0) || state.CalledAt > c.LastSeq || state.ReturnedAt > c.LastSeq || state.Dispatched && state.CalledAt <= callSeq[id] || state.Returned && (state.ReturnedAt <= state.CalledAt || state.ReturnedAt != resultSeq[id]) {
			return bad("call dispatch/result coordinate mismatch")
		}
		if state.JobHandle != 0 {
			if _, ok := c.Jobs[state.JobHandle]; !ok {
				return bad("call job reference missing")
			}
		}
	}
	for id := range results {
		if _, ok := calls[id]; !ok {
			return bad("result without call")
		}
	}
	batchSeq := map[uint64]bool{}
	for _, seq := range callSeq {
		batchSeq[seq] = true
	}
	for _, e := range entries {
		if e.Anchor != 0 && (!c.SkillMode || !batchSeq[e.Anchor]) {
			return bad("invalid batch anchor reference")
		}
	}
	if len(c.Turns) != len(c.TurnIDs) {
		return bad("turn identity set mismatch")
	}
	turns := make([]common.TurnFact, 0, len(c.Turns))
	for id, t := range c.Turns {
		if !c.TurnIDs[id] || id == "" || t.Start == 0 || t.Start > c.LastSeq || t.End > c.LastSeq || t.End != 0 && t.End <= t.Start {
			return bad("invalid turn coordinates")
		}
		turns = append(turns, t)
	}
	sort.Slice(turns, func(i, j int) bool { return turns[i].Start < turns[j].Start })
	var last, priorEnd uint64
	for _, t := range turns {
		if t.Index == 0 || t.Index <= last || t.Index > c.RequestCursor || last != 0 && (priorEnd == 0 || t.Start <= priorEnd) {
			return bad("invalid request cursor")
		}
		last = t.Index
		priorEnd = t.End
		if t.Policy != nil {
			effective := t.Policy.MaxModelRequests
			if effective == 0 {
				effective = common.DefaultMaxModelRequests
			}
			if t.Policy.MaxModelRequests < 0 || t.Policy.MaxModelRequests > 256 || effective != t.Policy.EffectiveMaxModelRequests {
				return bad("invalid saved turn policy")
			}
		}
		switch t.Outcome {
		case "success", "interrupted", "canceled", "error", "round_limit", "stopped":
			if t.End == 0 {
				return bad("missing terminal turn boundary")
			}
		case "":
			if t.End != 0 {
				return bad("missing turn outcome")
			}
		default:
			return bad("invalid turn outcome")
		}
	}
	var seq uint64
	for _, r := range c.Responses {
		if r.Seq <= seq || !responses[r.Seq] || !validProvenance(owner, &r.From) || r.Requested != nil && !validProvenance(owner, r.Requested) || r.Usage.Input < 0 || r.Usage.CacheWrite < 0 || r.Usage.CacheRead < 0 || r.Usage.Output < 0 {
			return bad("invalid accepted response fact")
		}
		seq = r.Seq
		delete(responses, r.Seq)
	}
	if len(responses) != 0 {
		return bad("missing accepted response usage fact")
	}
	for h, j := range c.Jobs {
		if h != j.Handle {
			return bad("job identity mismatch")
		}
		if reason := jobProblem(owner, j, nil); reason != "" {
			return bad(reason)
		}
	}
	for _, r := range c.Redactions {
		if r.Seq > c.LastSeq || r.From == 0 || r.From > r.To || r.To >= r.Seq || r.Level != "redact_result" || r.Reason == "" {
			return bad("invalid redaction span")
		}
		for _, e := range c.Entries {
			if e.Seq < r.From || e.Seq > r.To {
				continue
			}
			for _, p := range e.Parts {
				if p.Type == "tool_result" {
					for _, child := range p.Parts {
						if child.Type != "redacted" || child.Stub != "[redacted]" {
							return bad("redaction projection mismatch")
						}
					}
				}
			}
		}
	}
	pending := map[uint64]common.Entry{}
	requests := map[uint64]bool{}
	seq = 0
	for _, s := range c.RequestSeqs {
		if s <= seq || s > c.LastSeq {
			return bad("invalid request-send coordinates")
		}
		seq = s
		requests[s] = true
	}
	for _, e := range append(append([]common.Entry{}, c.Hints...), c.Ephemera...) {
		pending[e.Seq] = e
	}
	seq = 0
	for _, g := range c.Guidance {
		if g.Seq <= seq || g.Seq > c.LastSeq || g.ConsumedAt > c.LastSeq || g.ConsumedAt != 0 && g.ConsumedAt <= g.Seq || g.Kind != "hint" && g.Kind != "ephemeral" {
			return bad("invalid guidance state")
		}
		if g.ConsumedAt != 0 && (!requests[g.ConsumedAt] || g.Anchor >= g.ConsumedAt) {
			return bad("guidance consumption lacks a later request")
		}
		if g.Anchor != 0 && (!c.SkillMode || !batchSeq[g.Anchor]) {
			return bad("guidance batch reference missing")
		}
		seq = g.Seq
		e, ok := pending[g.Seq]
		if (g.ConsumedAt == 0) != ok || ok && e.Anchor != g.Anchor {
			return bad("guidance consumption mismatch")
		}
		delete(pending, g.Seq)
	}
	if len(pending) != 0 {
		return bad("missing guidance receipt")
	}
	return nil
}

// A semantic Part has a closed variant, unlike a forward-compatible outer event.
func semanticPart(owner common.Engine, p common.Part, result bool) string {
	if reason := partProblem(owner, &p, result); reason != "" {
		return reason
	}
	allowed := common.Part{Type: p.Type}
	if len(p.Parts) == 0 {
		allowed.Parts = p.Parts
	}
	switch p.Type {
	case "text":
		allowed.Text = p.Text
		allowed.From = p.From
		allowed.Opaque = p.Opaque
	case "tool_call":
		allowed.CallID = p.CallID
		allowed.Name = p.Name
		allowed.From = p.From
		allowed.Args = p.Args
		allowed.ArgumentsText = p.ArgumentsText
		allowed.Opaque = p.Opaque
	case "tool_result":
		allowed.CallID = p.CallID
		allowed.Parts = p.Parts
		allowed.IsError = p.IsError
		for _, child := range p.Parts {
			if reason := semanticPart(owner, child, true); reason != "" {
				return reason
			}
		}
	case "opaque":
		allowed.From = p.From
		allowed.Data = p.Data
	case "blob":
		allowed.MIME = p.MIME
		allowed.Ref = p.Ref
	case "redacted":
		allowed.Stub = p.Stub
		allowed.Ref = p.Ref
	}
	if !reflect.DeepEqual(p, allowed) {
		return "inactive part field is populated"
	}
	return ""
}
