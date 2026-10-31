package gradeedition2

import (
	"encoding/json"
	"testing"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

func TestChapter9DeferredDeclarations(t *testing.T) {
	for _, tc := range []struct {
		name           string
		tools, content string
		want           []string
		pass           bool
	}{
		{"startup", `[{"name":"load_skill"},{"name":"write_file","defer_loading":true}]`, `[]`, []string{"load_skill"}, true},
		{"disclosed", `[{"name":"load_skill"},{"name":"write_file","defer_loading":true}]`, `[{"type":"tool_reference","tool_name":"write_file"}]`, []string{"load_skill", "write_file"}, true},
		{"mixed content", `[{"name":"load_skill"},{"name":"write_file","defer_loading":true}]`, `[{"type":"text","text":"Enabled skills: code-tools"},{"type":"tool_reference","tool_name":"write_file"}]`, []string{"load_skill", "write_file"}, false},
		{"repeated reference", `[{"name":"load_skill"},{"name":"write_file","defer_loading":true}]`, `[{"type":"tool_reference","tool_name":"write_file"},{"type":"tool_reference","tool_name":"write_file"}]`, []string{"load_skill", "write_file"}, true},
		{"ordinary declarations", `[{"name":"load_skill"},{"name":"write_file"}]`, `[]`, []string{"load_skill", "write_file"}, true},
		{"deferred is not exposed", `[{"name":"load_skill"},{"name":"write_file","defer_loading":true}]`, `[]`, []string{"load_skill", "write_file"}, false},
		{"missing schema", `[{"name":"load_skill"}]`, `[{"type":"tool_reference","tool_name":"write_file"}]`, []string{"load_skill", "write_file"}, false},
		{"duplicate schema", `[{"name":"load_skill"},{"name":"load_skill"}]`, `[]`, []string{"load_skill"}, false},
		{"text is not disclosure", `[{"name":"load_skill"},{"name":"write_file","defer_loading":true}]`, `[{"type":"text","text":"write_file"}]`, []string{"load_skill", "write_file"}, false},
		{"orphan reference", `[{"name":"load_skill"},{"name":"write_file","defer_loading":true}]`, `[{"type":"tool_reference","tool_name":"write_file"}]`, []string{"load_skill", "write_file"}, false},
		{"wrong-role reference", `[{"name":"load_skill"},{"name":"write_file","defer_loading":true}]`, `[{"type":"tool_reference","tool_name":"write_file"}]`, []string{"load_skill", "write_file"}, false},
		{"extra capability", `[{"name":"load_skill"},{"name":"write_file"}]`, `[]`, []string{"load_skill"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			call := `{"role":"assistant","content":[{"type":"tool_use","id":"load","name":"load_skill","input":{"name":"code-tools"}}]},`
			role := "user"
			if tc.name == "orphan reference" {
				call = ""
			}
			if tc.name == "wrong-role reference" {
				role = "assistant"
			}
			body := []byte(`{"tools":` + tc.tools + `,"messages":[` + call + `{"role":"` + role + `","content":[{"type":"tool_result","tool_use_id":"load","content":` + tc.content + `}]}]}`)
			if !json.Valid(body) {
				t.Fatal("invalid test JSON")
			}
			c := grade.Check{Passed: true, Points: 10, Earned: 10}
			skillDeclarations(&c, fakevendor.Recorded{Body: body}, tc.want)
			if c.Passed != tc.pass {
				t.Fatalf("passed=%t want=%t: %v", c.Passed, tc.pass, c.Details)
			}
		})
	}
}

func TestChapter9SkillResultContent(t *testing.T) {
	for _, tc := range []struct {
		name, tool, content string
		call, bad, pass     bool
	}{
		{"reference-only load", "load_skill", `[{"type":"tool_reference","tool_name":"write_file"}]`, true, false, true},
		{"ordinary result", "write_file", `"written"`, true, false, true},
		{"empty ordinary result", "write_file", `[]`, true, false, false},
		{"references from wrong tool", "write_file", `[{"type":"tool_reference","tool_name":"write_file"}]`, true, false, false},
		{"orphan references", "load_skill", `[{"type":"tool_reference","tool_name":"write_file"}]`, false, false, false},
		{"failed load", "load_skill", `[{"type":"tool_reference","tool_name":"write_file"}]`, true, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			messages := []any{}
			if tc.call {
				messages = append(messages, map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "tool_use", "id": "load", "name": tc.tool, "input": map[string]any{}}}})
			}
			messages = append(messages, map[string]any{"role": "user", "content": []any{map[string]any{"type": "tool_result", "tool_use_id": "load", "content": json.RawMessage(tc.content), "is_error": tc.bad}}})
			body, err := json.Marshal(map[string]any{"messages": messages})
			if err != nil {
				t.Fatal(err)
			}
			c := grade.Check{Passed: true, Points: 10, Earned: 10}
			skillResult(&c, fakevendor.Recorded{Body: body}, "load", false)
			if c.Passed != tc.pass {
				t.Fatalf("passed=%t want=%t: %v", c.Passed, tc.pass, c.Details)
			}
		})
	}
}
