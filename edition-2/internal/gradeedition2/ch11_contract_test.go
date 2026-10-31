package gradeedition2

import (
	"encoding/json"
	"testing"
)

func TestMCPNormalizedSchema(t *testing.T) {
	original := mcpTool("remote_echo", "")["inputSchema"]
	for _, tc := range []struct {
		name, schema string
		want         bool
	}{
		{"unchanged", `{"type":"object","properties":{"text":{"type":"string"}}}`, true},
		{"normalized-controls", `{"type":"object","properties":{"text":{"type":"string"},"ai_callback_delay":{"type":"number"},"ai_callback_pattern":{"type":"string"},"max_output_bytes":{"type":"integer"}},"required":[],"additionalProperties":false}`, true},
		{"missing-remote", `{"type":"object","properties":{}}`, false},
		{"changed-remote", `{"type":"object","properties":{"text":{"type":"number"}}}`, false},
		{"unexpected-property", `{"type":"object","properties":{"text":{"type":"string"},"escape":{"type":"string"}}}`, false},
		{"wrong-control", `{"type":"object","properties":{"text":{"type":"string"},"ai_callback_delay":{"type":"string"}}}`, false},
		{"added-required", `{"type":"object","properties":{"text":{"type":"string"}},"required":["text"]}`, false},
		{"required-control", `{"type":"object","properties":{"text":{"type":"string"},"ai_callback_delay":{"type":"number"}},"required":["ai_callback_delay"]}`, false},
		{"invalid-required", `{"type":"object","properties":{"text":{"type":"string"}},"required":"text"}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got any
			if err := json.Unmarshal([]byte(tc.schema), &got); err != nil {
				t.Fatal(err)
			}
			if result := mcpNormalizedSchema(got, original); result != tc.want {
				t.Fatalf("accepted=%t want %t", result, tc.want)
			}
		})
	}
}

func TestMCPFailureDiagnostic(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		want       bool
	}{
		{"omitted", "normal context", true},
		{"unavailable", "[automatic observation unavailable: AUTO_FAILURE_ONLY]", true},
		{"error", "error collecting automatic tool: AUTO_FAILURE_ONLY", true},
		{"failed", "observation failed: AUTO_FAILURE_ONLY", true},
		{"bare", "AUTO_FAILURE_ONLY", false},
		{"successful-label", "snapshot: AUTO_FAILURE_ONLY", false},
		{"unrelated-error", "unrelated error\nsnapshot: AUTO_FAILURE_ONLY", false},
		{"accumulated", "error: AUTO_FAILURE_ONLY\nerror: AUTO_FAILURE_ONLY", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			messages := []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "text", "text": tc.text}}}}
			if got := mcpFailureDiagnostic(messages); got != tc.want {
				t.Fatalf("accepted=%t want %t", got, tc.want)
			}
		})
	}
}
