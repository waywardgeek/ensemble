package gradeedition2

import "testing"

func TestPersistenceSavePublicFields(t *testing.T) {
	for _, tc := range []struct {
		name, raw string
		valid     bool
	}{
		{"original", `{"config":{},"as_of":0,"context":{},"log":[]}`, true},
		{"retention", `{"config":{},"as_of":0,"context":{},"log":[],"log_mode":"dialogue_only","dialogue_log":[]}`, true},
		{"missing-log", `{"config":{},"as_of":0,"context":{}}`, false},
		{"unknown", `{"config":{},"as_of":0,"context":{},"log":[],"unexpected":true}`, false},
		{"null", `null`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := readPersistenceSave([]byte(tc.raw))
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%t error=%v", tc.valid, err)
			}
		})
	}
}
