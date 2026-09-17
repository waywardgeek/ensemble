package main

import (
	"github.com/waywardgeek/coding-agents-course/solutions/ch06/internal/common"
	"github.com/waywardgeek/coding-agents-course/solutions/ch06/internal/jobs"
	"github.com/waywardgeek/coding-agents-course/solutions/ch06/internal/llm"
	"github.com/waywardgeek/coding-agents-course/solutions/ch06/internal/tools"
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
		return tools.ToolEditFile(nil, args)
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

// write_file REFUSES to replace a file that exists unless asked by name, and
// the refusal says what would have been lost. Creating a new file and
// appending are untouched: only the destructive shape is gated. This is the
// answer to "write_file has no dry run and no diff": the refusal is the dry
// run, priced at one round trip.
func TestWriteFileRefusesSilentOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	const original = "one\ntwo\nthree\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	call := func(m map[string]any) (string, error) {
		args, _ := json.Marshal(m)
		return tools.ToolWriteFile(nil, args)
	}

	out, err := call(map[string]any{"path": path, "content": "X"})
	if err == nil {
		t.Fatalf("write_file replaced an existing file without overwrite:true and returned %q", out)
	}
	for _, want := range []string{"refused", "exists", "14 bytes", "3 lines", "overwrite:true", "edit_file"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal lacks %q: %v", want, err)
		}
	}
	if got, _ := os.ReadFile(path); string(got) != original {
		t.Errorf("file was modified by a refused write:\n%s", got)
	}

	// Asked by name, it replaces, and the reply says what it replaced.
	out, err = call(map[string]any{"path": path, "content": "X", "overwrite": true})
	if err != nil {
		t.Fatalf("overwrite:true refused: %v", err)
	}
	if !strings.Contains(out, "replaced 14 bytes, 3 lines") {
		t.Errorf("overwrite reply does not say what it replaced: %q", out)
	}
	if got, _ := os.ReadFile(path); string(got) != "X" {
		t.Errorf("overwrite not applied: %q", got)
	}

	// Controls: a new file needs no flag and reports nothing replaced;
	// append to an existing file is never refused.
	fresh := filepath.Join(dir, "new.txt")
	if out, err = call(map[string]any{"path": fresh, "content": "hi"}); err != nil || out != "wrote 2 bytes to "+fresh {
		t.Errorf("new file: out=%q err=%v", out, err)
	}
	if _, err = call(map[string]any{"path": path, "content": "Y", "append": true}); err != nil {
		t.Errorf("append refused: %v", err)
	}
	if got, _ := os.ReadFile(path); string(got) != "XY" {
		t.Errorf("append not applied: %q", got)
	}
}

// search_files renders context in grep's format when asked, and exactly the
// old path:line:text lines when not. Windows that overlap merge; groups that
// do not touch are separated by "--", across files too.
func TestSearchFilesContextLines(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(a, []byte("l1\nl2 hit\nl3\nl4\nl5\nl6\nl7 hit\nl8 hit\nl9\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("m1 hit\nm2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	call := func(ctx int) string {
		args, _ := json.Marshal(map[string]any{"pattern": "hit", "path": dir, "context_lines": ctx})
		out, err := tools.ToolSearchFiles(nil, args)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}

	if got, want := call(0), strings.Join([]string{
		a + ":2:l2 hit", a + ":7:l7 hit", a + ":8:l8 hit", b + ":1:m1 hit",
	}, "\n"); got != want {
		t.Errorf("context_lines 0:\n got: %q\nwant: %q", got, want)
	}

	if got, want := call(1), strings.Join([]string{
		a + "-1-l1", a + ":2:l2 hit", a + "-3-l3",
		"--",
		a + "-6-l6", a + ":7:l7 hit", a + ":8:l8 hit", a + "-9-l9",
		"--",
		b + ":1:m1 hit", b + "-2-m2",
	}, "\n"); got != want {
		t.Errorf("context_lines 1:\n got: %q\nwant: %q", got, want)
	}
}

// A pending tool_limits lands on the very next call, whatever it is, and
// that call's report says so. The one-shot rule is the design; this test is
// about the footgun in it being LOUD: a burn shows up in the result it
// caused, not as an unexplained truncation two calls later. The third call
// proves the note is one-shot too.
func TestToolLimitsConsumptionIsVisible(t *testing.T) {
	t.Chdir(t.TempDir())
	e := llm.NewEngine(common.Config{}, "log.jsonl", jobs.NewJobs(cliHost{}), tools.NewRegistry(), cliHost{})
	run := func(id, name, args string) string {
		if err := e.Execute(common.ToolCallPart{CallID: id, Name: name, Args: json.RawMessage(args)}); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		last := e.Log.Events[len(e.Log.Events)-1]
		if last.Type != common.ToolReturned || last.Tool == nil || len(last.Tool.Parts) != 1 {
			t.Fatalf("%s: last event is not a one-part common.ToolReturned: %+v", name, last)
		}
		if last.Tool.IsError {
			t.Fatalf("%s returned an error: %s", name, last.Tool.Parts[0].(common.TextPart).Text)
		}
		return last.Tool.Parts[0].(common.TextPart).Text
	}

	if out := run("c1", "tool_limits", `{"max_output_bytes":4096}`); !strings.Contains(out, "whichever tool that is") {
		t.Errorf("tool_limits reply does not warn where the limits land: %q", out)
	}
	out := run("c2", "list_directory", `{}`)
	if !strings.HasPrefix(out, "[tool_limits consumed by this list_directory call: ") ||
		!strings.Contains(out, "max_output_bytes 4096]\n") {
		t.Errorf("consuming call does not report the limits it took: %q", out)
	}
	if out := run("c3", "list_directory", `{}`); strings.Contains(out, "tool_limits") {
		t.Errorf("note repeated on a call that consumed nothing: %q", out)
	}

	// Two tool_limits in a row: the second is "the next call" for the first,
	// and the wasted set is reported like any other consumption.
	run("c4", "tool_limits", `{"ai_callback_delay":7}`)
	if out := run("c5", "tool_limits", `{}`); !strings.HasPrefix(out, "[tool_limits consumed by this tool_limits call: ai_callback_delay 7s") {
		t.Errorf("a tool_limits that clobbers a pending set does not say so: %q", out)
	}
}
