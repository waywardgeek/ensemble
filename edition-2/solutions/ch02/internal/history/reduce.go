package history

import (
	"encoding/json"
	"fmt"

	"ensemble/internal/common"
)

// Every fact changes content through this reducer. Known state/event pairs not
// named below leave turn state alone; an unusual order is not a panic.
func apply(c *common.Context, e common.Event) {
	switch e.Type {
	// Classification belongs here, after capture. Replaying a System arrival
	// recovers pending instructions even before the first request is sent.
	case "message_received":
		if e.Message.Actor == common.System {
			c.Ephemera = append(c.Ephemera, e.Message.Parts...)
		} else {
			c.Dialogue = append(c.Dialogue, common.Entry{Seq: e.Seq, Actor: e.Message.Actor, Parts: e.Message.Parts})
		}
		if c.Turn == "idle" {
			c.Turn = "input_pending"
		}
	case "request_sent":
		c.Ephemera = nil // consumption is logged, so replay knows it was delivered once.
		if c.Turn == "input_pending" {
			c.Turn = "in_flight"
		}
	// Usage and dialogue are derived from the same completed response fact.
	// A failed HTTP exchange records ErrorOccurred and never reaches this arm.
	case "response_ended":
		c.Dialogue = append(c.Dialogue, common.Entry{Seq: e.Seq, Actor: common.AgentActor, Parts: e.Response.Parts})
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
		c.Dialogue = append(c.Dialogue, common.Entry{Seq: e.Seq, Actor: common.Tool, Parts: []common.Part{p}})
		if c.Turn == "tools_pending" && toolsComplete(c.Dialogue) {
			c.Turn = "input_pending"
		}
	case "redacted":
		redact(c, e.Redact)
	case "error_occurred":
		if c.Turn == "in_flight" {
			c.Turn = "idle"
		}
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
		if entry.Seq < r.From || entry.Seq > r.To {
			out = append(out, entry)
			continue
		}
		// A fold is different from a per-entry filter: insert the replacement
		// once even when the span crosses speakers. System owns that summary.
		if r.Level == "redact_summary" {
			if !summarized {
				out = append(out, common.Entry{Seq: r.From, Actor: common.System, Parts: r.Replacement})
				summarized = true
			}
			continue
		}
		var parts []common.Part
		for _, p := range entry.Parts {
			switch r.Level {
			// The call survives result stubbing, and each removed part retains
			// its own Ref rather than losing addresses in one combined string.
			case "redact_result":
				if p.Type == "tool_result" {
					var stubs []common.Part
					for _, old := range p.Parts {
						// Count serialized part bytes, not guessed tokens. Replaying the
						// same event then produces the same useful deletion marker.
						encoded, _ := json.Marshal(old)
						stubs = append(stubs, common.Part{Type: "redacted", Stub: fmt.Sprintf("[redacted %d bytes]", len(encoded)), Ref: old.Ref})
					}
					p.Parts = stubs
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
		out = append(out, entry)
	}
	c.Dialogue = out
}
