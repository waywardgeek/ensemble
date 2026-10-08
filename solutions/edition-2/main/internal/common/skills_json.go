package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Required-field and duplicate checks are JSON dispatch, not graph behavior.
func skillObject(data []byte, names ...string) error {
	d := json.NewDecoder(bytes.NewReader(data))
	t, err := d.Token()
	if err != nil || t != json.Delim('{') {
		return fmt.Errorf("invalid skill object")
	}
	allowed := map[string]bool{}
	for _, n := range names {
		allowed[n] = true
	}
	seen := map[string]bool{}
	for d.More() {
		t, err = d.Token()
		if err != nil {
			return fmt.Errorf("invalid skill field")
		}
		n, ok := t.(string)
		if !ok || !allowed[n] || seen[n] {
			return fmt.Errorf("unknown or duplicate skill field")
		}
		seen[n] = true
		var raw json.RawMessage
		if d.Decode(&raw) != nil || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("null or invalid skill field")
		}
	}
	if _, err = d.Token(); err != nil {
		return err
	}
	if _, err = d.Token(); err != io.EOF {
		return fmt.Errorf("extra skill JSON")
	}
	if len(seen) != len(allowed) {
		return fmt.Errorf("missing skill field")
	}
	return nil
}

func (v *SkillState) UnmarshalJSON(data []byte) error {
	if err := skillObject(data, "revision", "primary", "roots", "active", "available", "tools", "retired"); err != nil {
		return err
	}
	type wire SkillState
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*v = SkillState(value)
	return nil
}

func (v *SkillActivation) UnmarshalJSON(data []byte) error {
	if err := skillObject(data, "activation", "name", "type", "body", "sha256", "tools", "dependencies", "offers"); err != nil {
		return err
	}
	type wire SkillActivation
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*v = SkillActivation(value)
	return nil
}

func (v *SkillActive) UnmarshalJSON(data []byte) error {
	if err := skillObject(data, "name", "type", "activation"); err != nil {
		return err
	}
	type wire SkillActive
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*v = SkillActive(value)
	return nil
}

func (v *SkillRetired) UnmarshalJSON(data []byte) error {
	if err := skillObject(data, "name", "activation"); err != nil {
		return err
	}
	type wire SkillRetired
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*v = SkillRetired(value)
	return nil
}

func (v *SkillOffer) UnmarshalJSON(data []byte) error {
	if err := skillObject(data, "name", "description"); err != nil {
		return err
	}
	type wire SkillOffer
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*v = SkillOffer(value)
	return nil
}
func (v *SkillTransition) UnmarshalJSON(data []byte) error {
	type wire SkillTransition
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	names := []string{"action", "name", "state", "activated"}
	if value.Action == "initialize" {
		names = append(names, "ceiling")
	}
	if err := skillObject(data, names...); err != nil {
		return err
	}
	*v = SkillTransition(value)
	return nil
}
