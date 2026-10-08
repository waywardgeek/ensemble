package gui

import (
	"encoding/json"
	"example.com/ensemble"
)

// ProjectPart traverses typed part children only. Argument keys are user data.
func ProjectPart(owner ServerOwner, p ensemble.Part) map[string]any {
	if p.Type == "opaque" {
		return map[string]any{"type": "opaque", "placeholder": true}
	}
	p.Opaque = nil
	children := p.Parts
	p.Parts = nil
	raw, _ := json.Marshal(p)
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	if p.Type == "tool_result" {
		out["is_error"] = p.IsError
	}
	if children != nil {
		parts := make([]any, 0, len(children))
		for _, child := range children {
			parts = append(parts, ProjectPart(owner, child))
		}
		out["parts"] = parts
	}
	return out
}
func projectParts(owner ServerOwner, parts []ensemble.Part) []any {
	out := make([]any, 0, len(parts))
	for _, p := range parts {
		out = append(out, ProjectPart(owner, p))
	}
	return out
}
func ProjectEvent(owner ServerOwner, e ensemble.Event) map[string]any {
	raw, _ := json.Marshal(e)
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	if e.Response != nil {
		out["response"].(map[string]any)["parts"] = projectParts(owner, e.Response.Parts)
	}
	if e.Message != nil {
		out["message"].(map[string]any)["parts"] = projectParts(owner, e.Message.Parts)
	}
	if e.Type == "tool_returned" && e.Tool != nil {
		out["tool"].(map[string]any)["is_error"] = e.Tool.IsError
	}
	if e.Tool != nil && e.Tool.Parts != nil {
		out["tool"].(map[string]any)["parts"] = projectParts(owner, e.Tool.Parts)
	}
	return out
}
func projectObservation(owner ServerOwner, o ensemble.Observation) map[string]any {
	raw, _ := json.Marshal(o)
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	if o.Kind == "part_delta" {
		out["text"] = o.Text
	}
	if o.Part != nil {
		out["part"] = ProjectPart(owner, *o.Part)
	}
	if o.Event.Type != "" {
		out["event"] = ProjectEvent(owner, o.Event)
	} else {
		delete(out, "event")
	}
	return out
}
