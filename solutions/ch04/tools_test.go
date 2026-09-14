package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The reference REFUSES an anchor it cannot resolve to exactly one place —
// both when it matches nothing and when it matches more than once — and in
// neither case touches the file. §3.5 states this as part of the declined
// decision; this test is what makes it a fact about the code rather than a
// sentence about it. The grader deliberately does not require refusing (any
// coherent answer is accepted), so the grader cannot be what proves it.
func TestEditFileRefusesZeroAndManyMatches(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	const original = "one\ntwo\ntwo\nthree\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	call := func(old string) (string, error) {
		args, _ := json.Marshal(map[string]string{"path": path, "old_text": old, "new_text": "X"})
		return toolEditFile(nil, args)
	}
	cases := []struct{ name, old, wantErr string }{
		{"zero matches", "seven", "not found"},
		{"many matches", "two", "appears 2 times"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := call(tc.old)
			if err == nil {
				t.Fatalf("edit_file applied an ambiguous edit and returned %q; want a refusal", out)
			}
			if !strings.Contains(err.Error(), "refused") || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("refusal does not say what it saw: %v", err)
			}
			got, _ := os.ReadFile(path)
			if string(got) != original {
				t.Errorf("file was modified by a refused edit:\n%s", got)
			}
		})
	}

	// Control: exactly one match is applied, so the refusals above are not
	// just "edit_file never works".
	if _, err := call("three"); err != nil {
		t.Fatalf("unique anchor refused: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "one\ntwo\ntwo\nX\n" {
		t.Errorf("unique edit not applied:\n%s", got)
	}
}
