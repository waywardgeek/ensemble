package gradeedition3

import (
	"os"
	"testing"
)

func TestHintCLI(t *testing.T) {
	bin := os.Getenv("EDITION3_HINT_CLI")
	if bin == "" {
		t.Skip("set EDITION3_HINT_CLI to an isolated CLI wrapper")
	}
	for _, name := range []string{"early", "aged", "interrupt-now", "interrupt-after_tool"} {
		t.Run(name, func(t *testing.T) {
			if err := runHintBatch(bin, name); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestHintResultPairing(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		good       bool
	}{
		{"text", `{"messages":[{"content":[{"type":"tool_result","tool_use_id":"call","content":"actual"}]}]}`, true},
		{"parts", `{"messages":[{"content":[{"type":"tool_result","tool_use_id":"call","content":[{"type":"text","text":"actual"}]}]}]}`, true},
		{"arguments are not results", `{"messages":[{"content":[{"type":"tool_use","id":"call","input":{"text":"actual"}}]}]}`, false},
		{"wrong ID", `{"messages":[{"content":[{"type":"tool_result","tool_use_id":"other","content":"actual"}]}]}`, false},
		{"skip refusal", `{"messages":[{"content":[{"type":"tool_result","tool_use_id":"call","is_error":true,"content":"interrupted; not executed"}]}]}`, false},
		{"duplicate", `{"messages":[{"content":[{"type":"tool_result","tool_use_id":"call","content":"actual"},{"type":"tool_result","tool_use_id":"call","content":"actual"}]}]}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			text, err := hintToolText([]byte(tc.body), "call")
			if tc.good {
				if err != nil || text != "actual" {
					t.Fatalf("%q, %v", text, err)
				}
			} else if err == nil {
				t.Fatal("accepted invalid result")
			}
		})
	}
}
