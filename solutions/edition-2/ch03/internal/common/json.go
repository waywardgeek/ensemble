package common

import (
	"encoding/json"
	"fmt"
)

// Empty result arrays are meaningful; omitempty would make a valid result look
// as though its required parts field had never been supplied.
func (p Part) MarshalJSON() ([]byte, error) {
	type wire Part
	if p.Type == "tool_result" {
		return json.Marshal(struct {
			wire
			Parts []Part `json:"parts"`
		}{wire(p), p.Parts})
	}
	return json.Marshal(wire(p))
}
func (t ToolEvent) MarshalJSON() ([]byte, error) {
	type wire ToolEvent
	if t.Parts != nil {
		return json.Marshal(struct {
			wire
			Parts []Part `json:"parts"`
		}{wire(t), t.Parts})
	}
	return json.Marshal(wire(t))
}
func (p *Part) UnmarshalJSON(data []byte) error {
	type wire Part
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if raw, ok := fields["is_error"]; ok && string(raw) == "null" {
		return fmt.Errorf("is_error must be Boolean")
	}
	*p = Part(value)
	return nil
}
func (t *ToolEvent) UnmarshalJSON(data []byte) error {
	type wire ToolEvent
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if raw, ok := fields["is_error"]; ok && string(raw) == "null" {
		return fmt.Errorf("is_error must be Boolean")
	}
	*t = ToolEvent(value)
	return nil
}
