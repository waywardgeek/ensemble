package main

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestPackageBoundaries(t *testing.T) {
	// Fixtures use unrelated identifiers: the policy is about placement/imports.
	base := map[string]string{
		"go.mod": "module example.test/course\n\ngo 1.23\n",
		"api.go": `package course
import "example.test/course/internal/provider"
func Start() { provider.Run() }`,
		"internal/provider/run.go": `package provider
import "example.test/course/internal/common"
func Run() { _ = common.Record{} }`,
		"internal/common/types.go": `package common
type Record struct{ Text string }
func (r Record) String() string { return describe(r) }
func describe(r Record) string { return r.Text }`,
		"cmd/run.go": `package main
import app "example.test/course"
func main() { app.Start() }`,
	}
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{"control", nil, nil},
		{"cli-private-import", map[string]string{
			"cmd/private.go": `package main; import _ "example.test/course/internal/provider"`,
		}, []string{"cli-public"}},
		{"new-sibling-spoke", map[string]string{
			"internal/newspoke/new.go":      `package newspoke; const Version = 1`,
			"internal/provider/sideways.go": `package provider; import _ "example.test/course/internal/newspoke"`,
		}, []string{"star-import"}},
		{"common-implementation", map[string]string{
			"internal/common/policy.go": `package common; func ChooseModel() string { return "policy" }`,
		}, []string{"common-behavior"}},
		{"common-import", map[string]string{
			"internal/newspoke/new.go":    `package newspoke; const Version = 1`,
			"internal/common/sideways.go": `package common; import _ "example.test/course/internal/newspoke"`,
		}, []string{"common-import"}},
		{"common-extra-method", map[string]string{
			"internal/common/policy.go": `package common; func (r Record) Retry() {}`,
		}, []string{"common-behavior"}},
		{"false-interface-method", map[string]string{
			"internal/common/types.go": `package common
type Record struct{ Text string }
func (r Record) String(x int) int { return x }`,
		}, []string{"common-behavior"}},
		{"shadowed-helper-name", map[string]string{
			"internal/common/types.go": `package common
type Record struct{ Text string }
func (r Record) String() string { describe := "local"; return describe }
func describe(r Record) string { return r.Text }`,
		}, []string{"common-behavior"}},
		{"equivalent-interface-type-alias", map[string]string{
			"internal/common/types.go": `package common
type Text = string
type Record struct{ Text string }
func (r Record) String() Text { return r.Text }`,
		}, nil},
		{"nested-module-independent", map[string]string{
			"consumer/go.mod":  "module example.test/consumer\n",
			"consumer/main.go": `package main; func main() {}`,
		}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			for _, files := range []map[string]string{base, tc.files} {
				for path, contents := range files {
					full := filepath.Join(root, path)
					if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(full, []byte(contents), 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			findings, err := inspect(root)
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, f := range findings {
				got = append(got, f.ID)
			}
			sort.Strings(got)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v, want %v: %+v", got, tc.want, findings)
			}
		})
	}
}

func TestOptionalModulePublicVocabulary(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"core/go.mod":                  "module example.test/core\n\ngo 1.23\n",
		"core/api.go":                  "package core; type Owner interface { Logf(string,...any) }",
		"gui/go.mod":                   "module example.test/gui\n\ngo 1.23\nrequire example.test/core v0.0.0\nreplace example.test/core => ../core\n",
		"gui/internal/common/types.go": `package common; import "example.test/core"; type Server interface { Application() core.Owner }`,
	}
	for name, data := range files {
		p := filepath.Join(root, name)
		if e := os.MkdirAll(filepath.Dir(p), 0700); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(p, []byte(data), 0600); e != nil {
			t.Fatal(e)
		}
	}
	gui := filepath.Join(root, "gui")
	if got, e := inspect(gui); e != nil || len(got) != 0 {
		t.Fatalf("valid public boundary: %v %v", got, e)
	}
	p := filepath.Join(gui, "internal/common/behavior.go")
	if e := os.WriteFile(p, []byte("package common; func ChooseModel()string{return \"forbidden\"}"), 0600); e != nil {
		t.Fatal(e)
	}
	got, e := inspect(gui)
	if e != nil || len(got) != 1 || got[0].ID != "common-behavior" {
		t.Fatalf("module-aware analysis lost behavior refusal: %v %v", got, e)
	}
}
