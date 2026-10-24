package history

import (
	"ensemble/internal/common"
	"fmt"
)

// Shape/order validation belongs to the archive envelope. Known payloads are
// checked against the current projection before a reducer mutates anything.
func semantic(e common.Event, projection common.Context) error {
	c := &projection
	if err := payloadValid(e); err != nil {
		return err
	}
	if e.Redact != nil {
		if e.Redact.To >= e.Seq {
			return fmt.Errorf("redaction exceeds admitted history")
		}
		found := false
		for _, entry := range c.Dialogue {
			if entry.Seq >= e.Redact.From && entry.Seq <= e.Redact.To && dialogue(entry) {
				for _, index := range e.Redact.Parts {
					if index < 0 || index >= len(entry.Parts) {
						return fmt.Errorf("redaction part is absent")
					}
				}
				found = true
			}
		}
		if !found {
			return fmt.Errorf("redaction has no projected dialogue")
		}
	}
	if e.Skill != nil && e.Skill.ActiveSeq > e.Seq {
		return fmt.Errorf("skill activation is in the future")
	}
	return nil
}
func dialogue(entry common.Entry) bool { return entry.Kind == "" || entry.Kind == common.KindDialogue }
func outstanding(c *common.Context) bool {
	pending := map[string]bool{}
	for _, entry := range c.Dialogue {
		if dialogue(entry) {
			for _, p := range entry.Parts {
				if p.Type == "tool_call" {
					pending[p.CallID] = true
				}
				if p.Type == "tool_result" {
					delete(pending, p.CallID)
				}
			}
		}
	}
	return len(pending) != 0
}
func flushSurvivors(c *common.Context) {
	if outstanding(c) {
		return
	}
	for _, entry := range c.Held {
		if entry.Kind == common.KindSkill {
			skill := *entry.Skill
			found := false
			for i := range c.ActiveSkills {
				if c.ActiveSkills[i].Name == skill.Name {
					c.ActiveSkills[i] = skill
					found = true
					break
				}
			}
			if !found {
				c.ActiveSkills = append(c.ActiveSkills, skill)
			}
		}
		c.Dialogue = append(c.Dialogue, entry)
	}
	c.Held = nil
	if c.PendingHandoff == "" {
		return
	}
	c.Handoff, c.HandoffSeq = c.PendingHandoff, c.PendingHandoffSeq
	out := make([]common.Entry, 0, len(c.Dialogue)+1)
	for _, entry := range c.Dialogue {
		if entry.Kind == common.KindHandoff {
			continue
		}
		if dialogue(entry) {
			entry.Parts = clearToolParts(entry.Parts)
			if len(entry.Parts) == 0 {
				continue
			}
		}
		out = append(out, entry)
	}
	// Held survivors keep their event Seq even when their delivery position is
	// after later receipts. Validation must not mistake this for a broken log.
	c.Dialogue = append(out, common.Entry{Kind: common.KindHandoff, Seq: c.HandoffSeq, Actor: common.Human, Handoff: &common.HandoffData{Text: c.Handoff}, Parts: common.PartList{common.Text(c.Handoff)}})
	c.Ephemera = nil
	c.PendingHandoff = ""
	c.PendingHandoffSeq = 0
}
func clearToolParts(parts common.PartList) common.PartList {
	out := common.PartList{}
	for _, p := range parts {
		if p.Type == "tool_call" || p.Type == "tool_result" || p.Type == "thinking" || p.Type == "reasoning" || p.Type == "opaque" {
			continue
		}
		p.Opaque = nil
		if len(p.Parts) > 0 {
			p.Parts = clearToolParts(p.Parts)
		}
		out = append(out, p)
	}
	return out
}
func removeToolTraffic(c *common.Context, r common.RedactData) {
	selected := map[string]bool{}
	for _, entry := range c.Dialogue {
		if dialogue(entry) && entry.Seq >= r.From && entry.Seq <= r.To {
			for _, p := range entry.Parts {
				if p.Type == "tool_call" || p.Type == "tool_result" {
					selected[p.CallID] = true
				}
			}
		}
	}
	out := make([]common.Entry, 0, len(c.Dialogue))
	for _, entry := range c.Dialogue {
		if dialogue(entry) {
			parts := common.PartList{}
			for _, p := range entry.Parts {
				if selected[p.CallID] && (p.Type == "tool_call" || p.Type == "tool_result") {
					continue
				}
				parts = append(parts, p)
			}
			entry.Parts = parts
			if len(parts) == 0 {
				continue
			}
		}
		out = append(out, entry)
	}
	c.Dialogue = out
}
