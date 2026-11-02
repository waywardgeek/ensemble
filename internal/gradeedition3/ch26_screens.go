package gradeedition3

import (
	"encoding/json"
	"fmt"
)

// screenshotSelectionSchema is the declaration oracle for the pending host-screen
// driver. It awards no points and cannot establish native capture, default-display
// behavior, authority, cleanup, or delayed image delivery by itself.
func screenshotSelectionSchema(raw json.RawMessage) error {
	var schema struct {
		Type       string                     `json:"type"`
		Properties map[string]json.RawMessage `json:"properties"`
		Required   []string                   `json:"required"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		return err
	}
	if schema.Type != "object" {
		return fmt.Errorf("screenshot arguments must be an object")
	}
	for name := range schema.Properties {
		switch name {
		case "display", "ai_callback_delay", "ai_callback_pattern", "max_output_bytes":
		default:
			return fmt.Errorf("screenshot advertises unsupported selector %q", name)
		}
	}
	var display struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(schema.Properties["display"], &display); err != nil {
		return fmt.Errorf("missing or invalid display declaration: %w", err)
	}
	if display.Type != "integer" {
		return fmt.Errorf("display must be an integer selector")
	}
	for _, name := range schema.Required {
		if name == "display" {
			return fmt.Errorf("display must be optional; omission selects display zero")
		}
	}
	return nil
}
