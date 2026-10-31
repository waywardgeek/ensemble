package gradeedition2

import (
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// TestChapter2AcceptedSolution exercises the public command path against the
// unchanged accepted solution.
// Targeted source mutations are recorded separately in the enhancement notes.
func TestChapter2AcceptedSolution(t *testing.T) {
	bin, cleanup, err := grade.Build("../../solutions/ch02")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	report, err := Chapter2(bin)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Passed || report.Score != 100 || report.MaxScore != 100 {
		t.Fatalf("corrected Chapter 2: %+v", report)
	}
}

// TestChapter2InterleavedContent checks the allowed string form without
// weakening result ordering, correlation, or payload assertions.
func TestChapter2InterleavedContent(t *testing.T) {
	const body = `{"messages":[{"role":"user","content":"read both"},{"role":"assistant","content":[{"type":"tool_use","id":"one"}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"one","content":"first result"},{"type":"tool_result","tool_use_id":"two","content":"second result"},{"type":"text","text":"keep this instruction"}]}]}`
	for _, tc := range []struct {
		name, body string
		pass       bool
	}{
		{"string text", body, true},
		{"block text", strings.Replace(body, `"content":"read both"`, `"content":[{"type":"text","text":"read both"}]`, 1), true},
		{"wrong id", strings.Replace(body, `"tool_use_id":"two"`, `"tool_use_id":"wrong"`, 1), false},
		{"lost output", strings.Replace(body, "second result", "lost", 1), false},
		{"wrong order", strings.Replace(body, `{"type":"tool_result","tool_use_id":"two","content":"second result"},{"type":"text","text":"keep this instruction"}`, `{"type":"text","text":"keep this instruction"},{"type":"tool_result","tool_use_id":"two","content":"second result"}`, 1), false},
		{"invalid content", strings.Replace(body, `"content":"read both"`, `"content":42`, 1), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := grade.Check{Passed: true, Points: 15, Earned: 15}
			checkInterleaved(&c, tc.body)
			if c.Passed != tc.pass {
				t.Fatalf("passed=%v want=%v details=%v", c.Passed, tc.pass, c.Details)
			}
		})
	}
}
