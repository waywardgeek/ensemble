package llm

import "example.com/ensemble/internal/common"

func sessionCollections(owner common.Engine, c common.Context, e common.Event) error {
	if c.Session == nil {
		return nil
	}
	entries, parts := 0, 0
	seenHints := map[uint64]bool{}
	var countParts func([]common.Part)
	countParts = func(list []common.Part) {
		parts += len(list)
		for _, p := range list {
			countParts(p.Parts)
		}
	}
	for _, list := range [][]common.Entry{c.Entries, c.Instructions, c.Ephemera, c.Hints, c.PendingSkills, c.DeferredSkills} {
		for _, entry := range list {
			if entry.Purpose == "hint" {
				if seenHints[entry.Seq] {
					continue
				}
				seenHints[entry.Seq] = true
			}
			entries++
			countParts(entry.Parts)
		}
	}
	if c.Pending != nil {
		entries++
		countParts(c.Pending.Parts)
	}
	calls := len(c.Calls)
	turns := len(c.TurnIDs)
	switch e.Type {
	case "message_received":
		if e.Message != nil {
			entries++
			countParts(e.Message.Parts)
		}
	case "hint_received":
		entries++
		parts++
	case "response_ended":
		if e.Response != nil {
			entries++
			countParts(e.Response.Parts)
			for _, p := range e.Response.Parts {
				if p.Type == "tool_call" {
					calls++
				}
			}
		}
	case "tool_returned":
		if e.Tool != nil {
			entries++
			parts++
			countParts(e.Tool.Parts)
		}
	case "turn_started":
		turns++
	case "skills_initialized", "skills_changed":
		if e.Skills != nil {
			for _, m := range e.Skills.Activated {
				if m.Type != "primary" {
					entries++
					if m.Body != "" {
						parts++
					}
				}
			}
		}
	}
	if entries > 1000000 || parts > 1000000 || calls > 1000000 || turns > 1000000 {
		return &common.SessionError{Code: "session_limit", Detail: "semantic entry, part or seen-identity collection limit"}
	}
	return nil
}
