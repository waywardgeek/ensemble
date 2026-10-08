package gui

import (
	"bytes"
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
	out := projectObject(owner, p)
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
	out := projectObject(owner, e)
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
	out := projectObject(owner, o)
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

// Preserve numeric wire tokens while projecting typed values. A float64
// intermediate would round valid uint64 revisions before the browser sees them.
// The owner remains available for diagnostics throughout this conversion.
func projectObject(owner ServerOwner, value any) map[string]any {
	raw, _ := json.Marshal(value)
	out := map[string]any{}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	_ = decoder.Decode(&out)
	return out
}
