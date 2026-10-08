package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// JSON dispatch is deliberately kept with the value. Presence of all three
// capture fields is required; zero is a real configured policy, not omission.
func (p *TurnPolicy) UnmarshalJSON(data []byte) error {
	d := json.NewDecoder(bytes.NewReader(data))
	t, err := d.Token()
	if err != nil || t != json.Delim('{') {
		return fmt.Errorf("invalid turn policy")
	}
	f := map[string]json.RawMessage{}
	for d.More() {
		t, err = d.Token()
		if err != nil {
			return fmt.Errorf("invalid turn policy")
		}
		k, ok := t.(string)
		if !ok {
			return fmt.Errorf("invalid turn policy")
		}
		if _, exists := f[k]; exists {
			return fmt.Errorf("duplicate turn policy field")
		}
		var raw json.RawMessage
		if d.Decode(&raw) != nil || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("invalid turn policy field")
		}
		f[k] = raw
	}
	if _, err = d.Token(); err != nil {
		return fmt.Errorf("invalid turn policy")
	}
	if _, err = d.Token(); err != io.EOF {
		return fmt.Errorf("invalid turn policy")
	}
	if len(f) != 3 || json.Unmarshal(f["revision"], &p.Revision) != nil || json.Unmarshal(f["max_model_requests"], &p.MaxModelRequests) != nil || json.Unmarshal(f["effective_max_model_requests"], &p.EffectiveMaxModelRequests) != nil {
		return fmt.Errorf("incomplete turn policy")
	}
	effective := p.MaxModelRequests
	if effective == 0 {
		effective = DefaultMaxModelRequests
	}
	if p.MaxModelRequests < 0 || p.MaxModelRequests > 256 || effective != p.EffectiveMaxModelRequests {
		return fmt.Errorf("invalid turn policy limit")
	}
	return nil
}
func (t *TurnEvent) UnmarshalJSON(data []byte) error {
	type wire TurnEvent
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	var f map[string]json.RawMessage
	if err := json.Unmarshal(data, &f); err != nil {
		return err
	}
	if raw, ok := f["policy"]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return fmt.Errorf("turn policy cannot be null")
	}
	*t = TurnEvent(value)
	return nil
}

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

// A missing historical delivery means plain, but an explicit empty or null
// value is invalid rather than a second spelling for that historical omission.
func (r *RequestEvent) UnmarshalJSON(data []byte) error {
	type wire RequestEvent
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if raw, ok := fields["delivery"]; ok {
		var delivery string
		if json.Unmarshal(raw, &delivery) != nil || (delivery != "stream" && delivery != "plain") {
			return fmt.Errorf("invalid request delivery")
		}
	}
	*r = RequestEvent(value)
	return nil
}
