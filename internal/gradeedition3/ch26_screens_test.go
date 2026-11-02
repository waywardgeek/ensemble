package gradeedition3

import (
	"encoding/json"
	"testing"
)

func TestScreenshotSelectionSchema(t *testing.T) {
	valid := `{"type":"object","properties":{"display":{"type":"integer","default":0},"ai_callback_delay":{"type":"number"}},"required":[]}`
	if err := screenshotSelectionSchema(json.RawMessage(valid)); err != nil {
		t.Fatal(err)
	}
	negatives := []string{
		`null`, `[]`, `{}`, `{"type":"object","properties":{}}`,
		`{"type":"object","properties":{"display":{"type":"string"}}}`,
		`{"type":"object","properties":{"display":{"type":"integer"}},"required":["display"]}`,
	}
	for _, name := range []string{"x", "y", "width", "height", "rectangle", "window"} {
		negatives = append(negatives, `{"type":"object","properties":{"display":{"type":"integer"},"`+name+`":{"type":"integer"}}}`)
	}
	for _, raw := range negatives {
		if err := screenshotSelectionSchema(json.RawMessage(raw)); err == nil {
			t.Errorf("accepted invalid screenshot schema: %s", raw)
		}
	}
}
