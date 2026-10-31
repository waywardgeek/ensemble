package history

import (
	"encoding/json"
	"fmt"
	"sort"

	"ensemble/internal/common"
)

// Every fact changes content through this reducer. Known state/event pairs not
// named below leave turn state alone; an unusual order is not a panic.
func apply(c *common.Context, e common.Event) {
	switch e.Type {
	// Classification belongs here, after capture. Replaying a System arrival
	// recovers pending instructions even before the first request is sent.
	case "message_received", "hint_received":
		if e.Message.Actor == common.System || (e.Message.Actor == common.Human && (c.Turn == "in_flight" || c.Turn == "tools_pending")) {
			c.Ephemera = append(c.Ephemera, e.Message.Parts...)
		} else {
			c.Dialogue = append(c.Dialogue, common.Entry{Seq: e.Seq, Kind: common.KindDialogue, Actor: e.Message.Actor, Parts: e.Message.Parts})
		}
		if c.Turn == "idle" || c.Turn == "interrupted" {
			c.Turn = "input_pending"
		}
	case "request_sent":
		c.SentThrough = e.Seq - 1
		if c.StartupTools == nil && e.Request.StartupTools != nil {
			c.StartupTools = e.Request.StartupTools
			c.Tools = append([]common.ToolDeclaration(nil), (*c.StartupTools)...)
		}
		c.Ephemera = nil // consumption is logged, so replay knows it was delivered once.
		if c.Turn == "input_pending" || c.Turn == "idle" {
			c.Turn = "in_flight"
		}
	// Usage and dialogue are derived from the same completed response fact.
	// A failed HTTP exchange records ErrorOccurred and never reaches this arm.
	case "response_ended":
		c.Dialogue = append(c.Dialogue, common.Entry{Seq: e.Seq, Kind: common.KindDialogue, Actor: common.AgentActor, Parts: e.Response.Parts})
		c.LastUsage = e.Response.Usage
		c.Usage.Input += e.Response.Usage.Input
		c.Usage.CacheWrite += e.Response.Usage.CacheWrite
		c.Usage.CacheRead += e.Response.Usage.CacheRead
		c.Usage.Output += e.Response.Usage.Output
		// Stop-reason vocabularies disagree across providers. Actual calls in the
		// normalized parts determine whether this reply requires tool results.
		if c.Turn == "in_flight" {
			c.Turn = "idle"
			for _, p := range e.Response.Parts {
				if p.Type == "tool_call" {
					c.Turn = "tools_pending"
				}
			}
		}
	// The return's own Seq identifies its content for future redaction; CallID
	// supplies correlation without changing authorship to a vendor's user role.
	case "tool_returned":
		p := common.Part{Type: "tool_result", CallID: e.Tool.CallID, Parts: e.Tool.Parts, IsError: e.Tool.IsError}
		// A job address is result metadata, not an attachment. The live result
		// may expose it; later stubs omit it while the captured event retains it.
		if e.Tool.Job != nil {
			p.Ref = e.Tool.Job.Output
		}
		c.Dialogue = append(c.Dialogue, common.Entry{Seq: e.Seq, Kind: common.KindDialogue, Actor: common.Tool, Parts: []common.Part{p}})
		if c.Turn == "tools_pending" && toolsComplete(c.Dialogue) {
			c.Turn = "input_pending"
		}
		if toolsComplete(c.Dialogue) {
			for _, held := range c.Held {
				land(c, held)
			}
			c.Held = nil
		}
	case "skill_loaded":
		survivor(c, common.Entry{Seq: e.Seq, Kind: common.KindSkill, Actor: common.System, Parts: []common.Part{{Type: "text", Text: e.Skill.Body}}})
	case "micro_handoff":
		survivor(c, common.Entry{Seq: e.Seq, Kind: common.KindHandoff, Actor: common.System, Parts: []common.Part{{Type: "text", Text: e.Handoff.Text}}})
	case "tools_changed":
		survivor(c, common.Entry{Seq: e.Seq, Kind: common.KindTools, Actor: common.System, Parts: []common.Part{{Type: "tool_decls", Added: e.Tools.Added, Removed: e.Tools.Removed}}})
	case "compactor_launched":
		c.Compression = &common.PendingCompression{Launch: *e.Compactor}
	case "compaction_outcome":
		c.CompactionOutcome = e.CompactionOutcome
		if c.Compression != nil && c.Compression.Launch.TaskID == e.CompactionOutcome.TaskID {
			c.Compression = nil
		}
	case "band_depopulated":
		if e.Depopulated.Reason == "graduation" && c.Compression != nil {
			c.Compression.Retire = e.Depopulated
		} else {
			depopulate(c, e.Depopulated)
		}
	case "band_populated":
		p := e.Populated
		if pending := c.Compression; pending != nil && pending.Launch.Band == p.Band && pending.Launch.Thru == p.Thru && pending.Launch.TaskID == p.TaskID {
			if pending.Cut != nil {
				redact(c, pending.Cut)
			}
			if pending.Retire != nil {
				depopulate(c, pending.Retire)
			}
			if p.TaskID == "" {
				c.Compression = nil
			} else {
				// Keep the applied receipt until disk finalization is recorded. A
				// restart must finish that work before scanning the active corpus.
				pending.Committed = p
			}
		}
		populate(c, p)
	case "redacted":
		if e.Redact.Reason == "memory compression" && c.Compression != nil {
			c.Compression.Cut = e.Redact
		} else {
			redact(c, e.Redact)
		}
	case "error_occurred", "turn_ended":
		c.Turn = "idle"
	case "turn_interrupted":
		c.Turn = "interrupted"
	}
}

// Only the latest reply defines the outstanding batch. Old results in resent
// history must not make today's request look like a continuation of an old loop.
func toolsComplete(entries []common.Entry) bool {
	// This set lives for one reduction, not for the Agent's lifetime. It cannot
	// grow into an ever-accumulating record of old tool calls or redactions.
	returned := map[string]bool{}
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		for _, p := range entry.Parts {
			if p.Type == "tool_result" {
				returned[p.CallID] = true
			}
			if p.Type == "tool_call" && !returned[p.CallID] {
				return false
			}
		}
		// Stop at the producing reply. Looking farther back would confuse an
		// earlier completed batch with the currently outstanding one.
		if entry.Actor == common.AgentActor {
			return true
		}
	}
	return true
}

// Redaction changes the projection, never the logged parts' backing arrays.
// A summary folds the span once; filters preserve each original entry's Seq.
func redact(c *common.Context, r *common.RedactData) {
	var out []common.Entry
	summarized := false
	for _, entry := range c.Dialogue {
		if entry.Seq < r.From || entry.Seq > r.To || entry.Kind > common.KindDialogue && !(r.Level == "redact_summary" && entry.Kind == common.KindHandoff) {
			out = append(out, entry)
			continue
		}
		// A fold is different from a per-entry filter: insert the replacement
		// once even when the span crosses speakers. System owns that summary.
		if r.Level == "redact_summary" {
			if !summarized && len(r.Replacement) > 0 {
				out = append(out, common.Entry{Seq: r.From, Actor: common.System, Parts: r.Replacement})
				summarized = true
			}
			continue
		}
		var parts []common.Part
		for _, p := range entry.Parts {
			switch r.Level {
			// The call survives result stubbing; its arguments describe how to
			// recover data without trusting an old job-output address.
			case "redact_result":
				if p.Type == "tool_result" {
					encoded, _ := json.Marshal(p.Parts)
					// Stubs deliberately name no output path: job handles may be reused
					// after restart. The original call remains the recovery instruction.
					p.Parts = []common.Part{{Type: "redacted", Stub: fmt.Sprintf("[redacted %d bytes; rerun the tool to recover]", len(encoded))}}
					p.Ref = common.Ref{}
				}
			case "redact_tool":
				if p.Type == "tool_call" || p.Type == "tool_result" {
					continue
				}
			case "redact_dialogue":
				if p.Type == "text" || p.Type == "opaque" {
					continue
				}
			}
			parts = append(parts, p)
		}
		entry.Parts = parts
		if len(parts) > 0 {
			out = append(out, entry)
		}
	}
	c.Dialogue = out
}

// Holding survivors preserves call/result pairing even under serial dispatch:
// another call from this same model reply may still be waiting to start.
func survivor(c *common.Context, entry common.Entry) {
	if !toolsComplete(c.Dialogue) {
		c.Held = append(c.Held, entry)
		return
	}
	land(c, entry)
}
func land(c *common.Context, entry common.Entry) {
	if entry.Kind == common.KindTools {
		for _, part := range entry.Parts {
			names := map[string]bool{}
			for _, name := range part.Removed {
				names[name] = true
			}
			for _, d := range part.Added {
				names[d.Name] = true
			}
			var kept []common.ToolDeclaration
			for _, d := range c.Tools {
				if !names[d.Name] {
					kept = append(kept, d)
				}
			}
			c.Tools = append(kept, part.Added...)
		}
	}
	if entry.Kind == common.KindHandoff {
		redact(c, &common.RedactData{From: 0, To: ^uint64(0), Level: "redact_tool"})
		// The checkpoint clears private replay material too. Copy parts so the
		// event log preserves it while the next request retains visible dialogue.
		var kept []common.Entry
		for _, old := range c.Dialogue {
			var parts []common.Part
			for _, part := range old.Parts {
				if part.Type != "opaque" {
					// Gemini can attach private replay material to visible text.
					part.Opaque = nil
					parts = append(parts, part)
				}
			}
			old.Parts = parts
			if len(parts) > 0 {
				kept = append(kept, old)
			}
		}
		c.Dialogue = kept
	}
	c.Dialogue = append(c.Dialogue, entry)
}

// Band events are data. Stable placement ignores the order in which bands were
// re-enabled, and replacing one identity exposes edits made while it was off.
func populate(c *common.Context, p *common.BandPopulatedData) {
	c.DisabledBands[p.Band] = false
	kind := common.KindSoul + common.EntryKind(p.Band-1)
	var out []common.Entry
	for _, e := range c.Dialogue {
		if e.Kind != kind || (p.Band > common.BandMemory && e.File != p.File) {
			out = append(out, e)
		}
	}
	out = append(out, common.Entry{Kind: kind, File: p.File, Thru: p.Thru, Actor: common.Human, Parts: []common.Part{{Type: "text", Text: p.Text}}})
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		am, bm := a.Kind >= common.KindSoul, b.Kind >= common.KindSoul
		if am != bm {
			return am
		}
		if !am {
			return false
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return fileLess(a.File, b.File)
	})
	c.Dialogue = out
}
func fileLess(a, b common.MemoryFileID) bool {
	return a.Date < b.Date || a.Date == b.Date && a.Num < b.Num
}
func depopulate(c *common.Context, p *common.BandDepopulatedData) {
	if p.Reason == "disabled" {
		c.DisabledBands[p.Band] = true
	}
	kind := common.KindSoul + common.EntryKind(p.Band-1)
	var out []common.Entry
	for _, e := range c.Dialogue {
		if e.Kind != kind || p.Thru != nil && fileLess(*p.Thru, e.File) {
			out = append(out, e)
		}
	}
	c.Dialogue = out
}
