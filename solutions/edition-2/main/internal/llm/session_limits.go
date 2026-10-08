package llm

import "example.com/ensemble/internal/common"

func partCount(owner common.Engine, list []common.Part) int {
	n := len(list)
	for _, p := range list {
		n += partCount(owner, p.Parts)
	}
	return n
}

// ReindexContext derives admission counts once when installing semantic state.
// This cache is never serialized or accepted from a checkpoint.
func ReindexContext(owner common.Engine, c *common.Context) {
	d := common.ContextIndex{Ready: true}
	seenHints := map[uint64]bool{}
	for _, list := range [][]common.Entry{c.Entries, c.Instructions, c.Ephemera, c.Hints, c.PendingSkills, c.DeferredSkills} {
		for _, e := range list {
			if e.Purpose == "hint" {
				if seenHints[e.Seq] {
					continue
				}
				seenHints[e.Seq] = true
			}
			d.Entries++
			d.Parts += partCount(owner, e.Parts)
		}
	}
	if c.Pending != nil {
		d.Entries++
		d.Parts += partCount(owner, c.Pending.Parts)
	}
	for _, call := range c.Calls {
		if !call.Returned {
			d.Unresolved++
		}
	}
	c.Index = d
}
func nextIndex(owner common.Engine, c common.Context, e common.Event) common.ContextIndex {
	if !c.Index.Ready {
		ReindexContext(owner, &c)
	}
	d := c.Index
	add := func(parts []common.Part) { d.Entries++; d.Parts += partCount(owner, parts) }
	switch e.Type {
	case "message_received":
		if e.Message != nil {
			add(e.Message.Parts)
		}
	case "hint_received":
		d.Entries++
		d.Parts++
	case "response_ended":
		if e.Response != nil {
			add(e.Response.Parts)
			for _, p := range e.Response.Parts {
				if p.Type == "tool_call" {
					d.Unresolved++
				}
			}
		}
	case "tool_returned":
		if e.Tool != nil {
			add(e.Tool.Parts)
			d.Parts++
			d.Unresolved--
		}
	case "error_occurred", "turn_ended":
		if c.Pending != nil {
			d.Entries--
			d.Parts -= partCount(owner, c.Pending.Parts)
		}
	case "request_sent":
		for _, h := range c.Hints {
			d.Entries--
			d.Parts -= partCount(owner, h.Parts)
		}
		if e.Request != nil {
			consumed := map[uint64]bool{}
			for _, seq := range e.Request.Ephemera {
				consumed[seq] = true
			}
			for _, entry := range c.Ephemera {
				if consumed[entry.Seq] {
					d.Entries--
					d.Parts -= partCount(owner, entry.Parts)
				}
			}
		}
	case "skills_initialized", "skills_changed":
		if e.Skills != nil {
			for _, m := range e.Skills.Activated {
				if m.Type != "primary" {
					d.Entries++
					if m.Body != "" {
						d.Parts++
					}
				}
			}
		}
	}
	return d
}
func ValidateSessionCollections(owner common.Engine, c common.Context, e common.Event) error {
	if c.Session == nil {
		return nil
	}
	d := nextIndex(owner, c, e)
	calls, turns := len(c.Calls), len(c.TurnIDs)
	if e.Type == "response_ended" && e.Response != nil {
		for _, p := range e.Response.Parts {
			if p.Type == "tool_call" {
				calls++
			}
		}
	}
	if e.Type == "turn_started" {
		turns++
	}
	if d.Entries > 1000000 || d.Parts > 1000000 || calls > 1000000 || turns > 1000000 {
		return &common.SessionError{Code: "session_limit", Detail: "semantic entry, part or seen-identity collection limit"}
	}
	return nil
}
