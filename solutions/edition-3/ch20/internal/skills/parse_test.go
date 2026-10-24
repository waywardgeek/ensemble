package skills

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFlexibleListsAndLiteralVariables(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	text := "---\nname: sample\ntype: loadable\ndescription: sample skill\ntools: read_file write_file\ndepends:\n - common\n - other\nloadable-skills: next later\n---\nUse $TOOLS; keep $UNKNOWN and ${UNKNOWN}.\n"
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	skill, err := parseSkill(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(skill.Tools, []string{"read_file", "write_file"}) || !reflect.DeepEqual(skill.Depends, []string{"common", "other"}) || !reflect.DeepEqual(skill.Offers, []string{"next", "later"}) {
		t.Fatalf("metadata: %+v", skill)
	}
	if got := expand(skill.Instructions, map[string]string{"TOOLS": "read_file"}); got != "Use read_file; keep $UNKNOWN and ${UNKNOWN}.\n" {
		t.Fatal(got)
	}
}
func TestMalformedMetadata(t *testing.T) {
	cases := []struct{ name, text string }{
		{"opening boundary", "name: x\ntype: primary\n---\nbody"},
		{"closing boundary", "---\nname: x\ntype: primary\nbody"},
		{"missing name", "---\ntype: primary\n---\nbody"},
		{"blank name", "---\nname: \ntype: primary\n---\nbody"},
		{"missing type", "---\nname: x\n---\nbody"},
		{"invalid type", "---\nname: x\ntype: wrong\n---\nbody"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "SKILL.md")
			if err := os.WriteFile(path, []byte(tc.text), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := parseSkill(path); err == nil {
				t.Fatalf("accepted %q", tc.text)
			}
		})
	}
}
func TestMinimalMetadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "SKILL.md")
	if err := os.WriteFile(path, []byte("---\nname: x\ntype: primary\n---\nbody"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := parseSkill(path); err != nil {
		t.Fatal(err)
	}
}
