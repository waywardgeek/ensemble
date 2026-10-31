package gradeedition2

import (
	"github.com/waywardgeek/ensemble/internal/grade"
	"path/filepath"
	"testing"
)

// TestChapter14Context drives the real optional GUI and Agent through settings,
// completed turns, crashes and reloads; only provider/MCP peers are external fakes.
func TestChapter14Context(t *testing.T) {
	gui, err := filepath.Abs("../../solutions/edition-3/ch14/gui")
	if err != nil {
		t.Fatal(err)
	}
	bin, cleanup, err := grade.Build(gui)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	report, err := Chapter14(bin)
	if err != nil {
		t.Fatal(err)
	}
	for _, check := range report.Checks {
		if !check.Passed {
			t.Errorf("%s: %v", check.ID, check.Details)
		}
	}
}

// This regression retains the independent reviewer’s valid split-message control.
// TestChapter14InlineGroups keeps valid split declarations and invalid post-group user content distinct.
func TestChapter14InlineGroups(t *testing.T) {
	const first = `{"role":"system","content":[{"type":"tool_addition","tool":{"type":"tool_definition","definition":{"name":"a","input_schema":{"type":"object"}}}}]}`
	const second = `{"role":"system","content":[{"type":"tool_removal","tool":{"type":"tool_reference","name":"b"}}]}`
	for _, tc := range []struct {
		name, body string
		valid      bool
	}{
		{"single", `{"messages":[{"role":"user","content":"data"},` + first + `]}`, true},
		{"consecutive_end", `{"messages":[{"role":"user","content":"data"},` + first + `,` + second + `]}`, true},
		{"consecutive_before_assistant", `{"messages":[{"role":"user","content":"data"},` + first + `,` + second + `,{"role":"assistant","content":"ok"}]}`, true},
		{"post_group_user", `{"messages":[{"role":"user","content":"data"},` + first + `,` + second + `,{"role":"user","content":"ephemera"}]}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req, err := contextParse([]byte(tc.body))
			if err != nil {
				t.Fatal(err)
			}
			added, removed, err := contextInline(req)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%t got error %v", tc.valid, err)
			}
			if tc.valid && err == nil && !added["a"] {
				t.Fatal("addition lost")
			}
			if tc.valid && tc.name != "single" && err == nil && !removed["b"] {
				t.Fatal("removal lost")
			}
		})
	}
}
