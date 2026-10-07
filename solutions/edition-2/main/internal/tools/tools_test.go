package tools

import (
	"encoding/json"
	"example.com/ensemble/internal/common"
	"fmt"
	"os"
	"strings"
	"testing"
)

type testRoot struct{}

func (testRoot) Logf(string, ...any)          {}
func (testRoot) Publish(string, common.Event) {}

type testAgent struct{ config common.Config }

func (a testAgent) Config() common.Config     { return a.config }
func (a testAgent) Workspace() string         { return a.config.Workspace }
func (a testAgent) Ensemble() common.Ensemble { return testRoot{} }
func registry(t *testing.T) *Registry {
	t.Helper()
	r, err := New(testAgent{common.Config{Workspace: t.TempDir()}}, []string{"read_file", "list_directory", "search_files", "write_file", "edit_file", "run_command"})
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func invoke(t *testing.T, r *Registry, name, args string, failed bool) string {
	t.Helper()
	v := r.Execute(common.Part{Type: "tool_call", CallID: "id", Name: name, Args: json.RawMessage(args)})
	if v.IsError != failed {
		t.Fatalf("%s %s: error=%v: %s", name, args, v.IsError, *v.Parts[0].Text)
	}
	return *v.Parts[0].Text
}
func plant(t *testing.T, r *Registry, path, text string) {
	t.Helper()
	if err := os.WriteFile(r.path(path), []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
}
func TestReadRangesLimitsAndEmpty(t *testing.T) {
	r := registry(t)
	plant(t, r, "notes", "alpha\nbeta\ngamma\ndelta\nepsilon\n")
	plant(t, r, "empty", "")
	if got := invoke(t, r, "read_file", `{"path":"notes","start_line":3,"end_line":4}`, false); got != "gamma\ndelta\n" {
		t.Fatal(got)
	}
	for _, a := range []string{`{"path":"notes","start_line":6}`, `{"path":"notes","end_line":2,"start_line":3}`, `{"path":"empty","start_line":1}`, `{"path":"empty","end_line":1}`, `{"path":"missing"}`} {
		invoke(t, r, "read_file", a, true)
	}
	if got := invoke(t, r, "read_file", `{"path":"empty","end_line":0}`, false); got != "" {
		t.Fatal(got)
	}
	if got := invoke(t, r, "read_file", `{"path":"empty"}`, false); got != "" {
		t.Fatal(got)
	}
	if got := invoke(t, r, "read_file", `{"path":"notes","max_bytes":6,"end_line":1}`, false); got != "alpha\n" {
		t.Fatal(got)
	}
	if got := invoke(t, r, "read_file", `{"path":"notes","max_bytes":5}`, false); !strings.Contains(got, "truncated: max_bytes") {
		t.Fatal(got)
	}
}
func TestSearchWindowsAndLimits(t *testing.T) {
	r := registry(t)
	plant(t, r, "notes.md", "alpha line one\nbeta line two\ngamma line three\ndelta line four\nepsilon line five\n")
	plant(t, r, "binary", "gamma\x00\n")
	os.Mkdir(r.path(".git"), 0700)
	plant(t, r, ".git/hidden", "gamma\n")
	os.Symlink(r.path("notes.md"), r.path("linked"))
	for _, test := range []struct{ args, want string }{{`{"pattern":"gamma","context_lines":1}`, "notes.md-2-beta line two\nnotes.md:3:gamma line three\nnotes.md-4-delta line four\n"}, {`{"pattern":"beta|delta","context_lines":1}`, "notes.md-1-alpha line one\nnotes.md:2:beta line two\nnotes.md-3-gamma line three\nnotes.md:4:delta line four\nnotes.md-5-epsilon line five\n"}, {`{"pattern":"alpha|epsilon","context_lines":1}`, "notes.md:1:alpha line one\nnotes.md-2-beta line two\n--\nnotes.md-4-delta line four\nnotes.md:5:epsilon line five\n"}, {`{"pattern":"none"}`, "No matches.\n"}, {`{"pattern":"gamma","file_pattern":"*.go"}`, "No matches.\n"}} {
		if got := invoke(t, r, "search_files", test.args, false); got != test.want {
			t.Fatalf("%s: %q", test.args, got)
		}
	}
	for _, a := range []string{`{"pattern":"["}`, `{"pattern":"gamma","file_pattern":"["}`} {
		invoke(t, r, "search_files", a, true)
	}
	if got := invoke(t, r, "search_files", `{"pattern":"gamma","max_matches":1}`, false); strings.Contains(got, "truncated") {
		t.Fatal(got)
	}
	if got := invoke(t, r, "search_files", `{"pattern":"alpha|epsilon","max_matches":1}`, false); !strings.Contains(got, "truncated: max_matches") {
		t.Fatal(got)
	}
	if got := invoke(t, r, "search_files", `{"pattern":"gamma","max_bytes":4}`, false); !strings.Contains(got, "truncated: max_bytes") {
		t.Fatal(got)
	}
}
func TestWritesAndExactEdits(t *testing.T) {
	r := registry(t)
	invoke(t, r, "write_file", `{"path":"nested/new","content":""}`, false)
	invoke(t, r, "write_file", `{"path":"nested/new","content":"x"}`, true)
	invoke(t, r, "write_file", `{"path":"nested/new","content":"first anchor\nsecond anchor\n","overwrite":true}`, false)
	before, _ := os.ReadFile(r.path("nested/new"))
	for _, a := range []string{`{"path":"nested/new","old_text":"anchor","new_text":"bad"}`, `{"path":"nested/new","old_text":"absent","new_text":"bad"}`, `{"path":"nested/new","old_text":"","new_text":"bad"}`} {
		invoke(t, r, "edit_file", a, true)
		after, _ := os.ReadFile(r.path("nested/new"))
		if string(after) != string(before) {
			t.Fatal("refusal changed bytes")
		}
	}
	invoke(t, r, "edit_file", `{"path":"nested/new","old_text":"first anchor","new_text":"first edit"}`, false)
	invoke(t, r, "edit_file", `{"path":"nested/new","old_text":"first edit","new_text":""}`, false)
	if got := invoke(t, r, "edit_file", `{"path":"nested/new","old_text":"second anchor","new_text":"second anchor"}`, false); !strings.Contains(got, "no-op") {
		t.Fatal(got)
	}
	invoke(t, r, "write_file", `{"path":"nested/new","content":"tail","append":true}`, false)
	after, _ := os.ReadFile(r.path("nested/new"))
	if string(after) != "\nsecond anchor\ntail" {
		t.Fatal(string(after))
	}
}
func TestValidationSelectionAndSnapshots(t *testing.T) {
	r := registry(t)
	for _, a := range []string{`{"path":42}`, `{"path":null}`, `{}`, `{"path":""}`, `{"path":"x","unknown":1}`, `{"path":"x","max_bytes":0}`, `{"path":"x","max_bytes":1.5}`, `{"path":"x","start_line":-1}`} {
		invoke(t, r, "read_file", a, true)
	}
	other, err := New(r.parent, []string{"read_file"})
	if err != nil {
		t.Fatal(err)
	}
	invoke(t, other, "write_file", `{"path":"forbidden","content":"x"}`, true)
	if _, err = os.Stat(r.path("forbidden")); !os.IsNotExist(err) {
		t.Fatal("disabled tool executed")
	}
	defs := r.Declarations()
	defs[0].Schema[0] = '!'
	if !json.Valid(r.Declarations()[0].Schema) {
		t.Fatal("declarations alias registry")
	}
	if len(other.Declarations()) != 1 {
		t.Fatal("sets leaked")
	}
}
func TestListingAndCommandDraining(t *testing.T) {
	r := registry(t)
	plant(t, r, "b", "x")
	plant(t, r, "a", "x")
	os.Mkdir(r.path("z"), 0700)
	if got := invoke(t, r, "list_directory", `{}`, false); got != "a\nb\nz/\n" {
		t.Fatal(got)
	}
	if got := invoke(t, r, "list_directory", `{"max_entries":3,"max_bytes":7}`, false); strings.Contains(got, "truncated") {
		t.Fatal(got)
	}
	if got := invoke(t, r, "list_directory", `{"max_entries":1}`, false); !strings.Contains(got, "max_entries") {
		t.Fatal(got)
	}

}

func TestUTF8ByteCapsSurviveJSON(t *testing.T) {
	r := registry(t)
	plant(t, r, "unicode", "éX")
	for limit, want := range map[int]string{1: "", 2: "é", 3: "éX"} {
		args := fmt.Sprintf(`{"path":"unicode","max_bytes":%d}`, limit)
		got := invoke(t, r, "read_file", args, false)
		encoded, _ := json.Marshal(got)
		var decoded string
		json.Unmarshal(encoded, &decoded)
		expected := want
		if limit < 3 {
			expected += "\n[truncated: max_bytes limit reached]\n"
		}
		if decoded != expected || strings.Contains(decoded, "\uFFFD") {
			t.Fatalf("read cap%d: %q", limit, decoded)
		}

	}
	// The file name makes the listing/search prefix itself multibyte.
	other := registry(t)
	plant(t, other, "éX", "match")
	for _, tool := range []string{"list_directory", "search_files"} {
		args := `{"max_bytes":1}`
		if tool == "search_files" {
			args = `{"pattern":"match","max_bytes":1}`
		}
		got := invoke(t, other, tool, args, false)
		if got != "\n[truncated: max_bytes limit reached]\n" {
			t.Fatal(tool, got)
		}
	}
}
