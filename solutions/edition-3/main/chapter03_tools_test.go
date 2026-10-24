package ensemble_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ensemble"
)

func c3Agent(t *testing.T, config ensemble.Config) ensemble.Agent {
	t.Helper()
	config.Model = "local-fixture-model"
	a, err := ensemble.New(nil).NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}
	return a
}
func c3Call(t *testing.T, a ensemble.Agent, name string, args any) (string, bool) {
	t.Helper()
	raw, err := json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}
	parts, failed := a.Tools().Dispatch(context.Background(), name, raw)
	if len(parts) != 1 || parts[0].Type != "text" {
		t.Fatalf("invalid tool result: %+v", parts)
	}
	return parts[0].Text, failed
}
func c3Write(t *testing.T, path, text string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
}
func c3Contents(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestChapter3Round1Registry(t *testing.T) {
	a := c3Agent(t, ensemble.Config{})
	if a.Tools().Parent() != a {
		t.Fatal("registry lost its actual owner")
	}
	names := map[string]bool{}
	for _, d := range a.Tools().Definitions() {
		if names[d.Name] || !json.Valid(d.Parameters) || d.Description == "" {
			t.Fatalf("invalid definition: %+v", d)
		}
		names[d.Name] = true
	}
	for _, name := range []string{"read_file", "list_directory", "search_files", "write_file", "edit_file", "run_command", "tool_limits", "wait_for_job", "send_input", "kill_job", "compact_context", "micro_handoff", "keep_tool_results", "agent_status", "spawn_sub_agent", "join_agents", "agents_status", "shutdown_agent", "send_message_to_parent", "respond_to_parent_message", "submit"} {
		if !names[name] {
			t.Fatalf("missing %s", name)
		}
	}
	if len(names) != 21 {
		t.Fatalf("definitions: %v", names)
	}
	for _, vendor := range []ensemble.Vendor{ensemble.Anthropic, ensemble.OpenAI, ensemble.Gemini} {
		empty := c3Agent(t, ensemble.Config{Vendor: vendor, Tools: []string{}})
		if empty.Tools() == a.Tools() || len(empty.Tools().Definitions()) != 0 {
			t.Fatal("empty registry was shared or defaulted")
		}
		wire, err := empty.Render()
		if err != nil {
			t.Fatal(err)
		}
		var request map[string]json.RawMessage
		if err := json.Unmarshal(wire, &request); err != nil {
			t.Fatal(err)
		}
		if _, present := request["tools"]; present {
			t.Fatalf("empty registry advertised tools: %s", wire)
		}
		if _, failed := c3Call(t, empty, "read_file", map[string]any{"path": "not-read"}); !failed {
			t.Fatal("empty registry dispatched a tool")
		}
	}
	selected := []string{"read_file"}
	b := c3Agent(t, ensemble.Config{Tools: selected})
	selected[0] = "write_file"
	copiedConfig := b.Config()
	copiedConfig.Tools[0] = "edit_file"
	defs := b.Tools().Definitions()
	defs[0].Name = "changed"
	defs[0].Parameters[0] = '!'
	if got := b.Tools().Definitions(); len(got) != 1 || got[0].Name != "read_file" || !json.Valid(got[0].Parameters) {
		t.Fatal("caller mutated registry")
	}
	target := filepath.Join(t.TempDir(), "untouched")
	if _, failed := c3Call(t, b, "write_file", map[string]any{"path": target, "content": "no"}); !failed {
		t.Fatal("unselected write dispatched")
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatal("disabled write had an effect")
	}
	c3Write(t, target, "owned bytes")
	if text, failed := c3Call(t, b, "read_file", map[string]any{"path": target}); failed || text != "owned bytes" {
		t.Fatalf("selected entry not dispatchable: %q failed=%t", text, failed)
	}
	if len(a.Tools().Definitions()) != 21 {
		t.Fatal("another Agent changed defaults")
	}
	for _, selected := range [][]string{{"read_file", "read_file"}, {"unknown"}} {
		if _, err := ensemble.New(nil).NewAgent(ensemble.Config{Model: "fixture", Tools: selected}); err == nil {
			t.Fatal("invalid registry selection accepted")
		}
	}
}

// Chapter 4 keeps explicit read bounds, but observes every ordinary call through
// Jobs. Assert the retained bytes separately from the bounded inline report.
func c3JobOutput(t *testing.T, a ensemble.Agent) (ensemble.JobData, string) {
	t.Helper()
	var latest ensemble.Job
	for _, job := range a.Parent().JobTable() {
		if job.Parent().Parent() == a && (latest == nil || job.Data().Handle > latest.Data().Handle) {
			latest = job
		}
	}
	if latest == nil {
		t.Fatal("ordinary call did not allocate a job")
	}
	data := latest.Data()
	if data.Status != ensemble.JobDone || data.Output.Kind != ensemble.RefHandle {
		t.Fatalf("completed job metadata: %+v", data)
	}
	path := data.Output.Locator
	if !filepath.IsAbs(path) {
		path = filepath.Join(a.Config().BaseDir, path)
	}
	disk := c3Contents(t, path)
	if data.Bytes != int64(len(disk)) {
		t.Fatalf("job bytes=%d disk=%d", data.Bytes, len(disk))
	}
	return data, disk
}

func TestChapter3Round1ReadBoundsAndCap(t *testing.T) {
	a := c3Agent(t, ensemble.Config{})
	path := filepath.Join(t.TempDir(), "lines.txt")
	full := strings.Repeat("p", 70000) + "\nalpha\nbeta\nlast"
	c3Write(t, path, full)
	cases := []struct {
		name string
		args map[string]any
		want string
	}{
		{"range before cap", map[string]any{"start_line": 2, "end_line": 3, "max_bytes": 8}, "alpha\nbe\n[truncated at max_bytes=8]\n"},
		{"exact requested range", map[string]any{"start_line": 2, "end_line": 2, "max_bytes": 6}, "alpha\n"},
		{"inclusive EOF", map[string]any{"start_line": 3, "end_line": 99}, "beta\nlast"},
		{"zero end means EOF", map[string]any{"start_line": 3, "end_line": 0}, "beta\nlast"},
		{"past EOF", map[string]any{"start_line": 99}, ""},
		{"explicit larger cap", map[string]any{"max_bytes": 40000}, strings.Repeat("p", 40000) + "\n[truncated at max_bytes=40000]\n"},
		{"default full disk output", map[string]any{}, full},
		{"zero defaults", map[string]any{"start_line": 0, "end_line": 0, "max_bytes": 0}, full},
	}
	const inlineCap = 16384
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.args["path"] = path
			got, failed := c3Call(t, a, "read_file", tc.args)
			if failed {
				t.Fatalf("read failed: %s", got)
			}
			data, disk := c3JobOutput(t, a)
			if disk != tc.want {
				t.Fatalf("retained read mismatch: got length=%d want length=%d", len(disk), len(tc.want))
			}
			if len(tc.want) <= inlineCap {
				if got != tc.want {
					t.Fatalf("small completed output changed: %q", got)
				}
			} else {
				header, payload, found := strings.Cut(got, "\n")
				wantHeader := fmt.Sprintf("job %d done; %d bytes total; output: %s", data.Handle, data.Bytes, data.Output.Locator)
				if !found || header != wantHeader {
					t.Fatalf("incorrect truncated-report header: got %q want %q", header, wantHeader)
				}
				wantPayload := fmt.Sprintf("%s\n[output truncated; %d bytes omitted; %d bytes total; full output: %s]\n%s", tc.want[:inlineCap/2], len(tc.want)-inlineCap, len(tc.want), data.Output.Locator, tc.want[len(tc.want)-inlineCap/2:])
				if len(got) > inlineCap+1024 || payload != wantPayload {
					t.Fatalf("incorrect head/tail/cap/omitted-count/locator payload: inline length=%d disk length=%d payload=%q", len(got), len(disk), payload)
				}
			}
		})
	}
	for _, args := range []map[string]any{
		{"max_bytes": -1}, {"max_bytes": 1.5}, {"max_bytes": "8"}, {"max_bytes": nil}, {"max_bytes": true},
		{"start_line": -1}, {"start_line": nil}, {"start_line": "2"}, {"end_line": -1}, {"end_line": nil},
		{"start_line": 3, "end_line": 2}, {"start_line": 1.5}, {"end_line": true}, {"unknown": 1},
	} {
		args["path"] = path
		if _, failed := c3Call(t, a, "read_file", args); !failed {
			t.Fatalf("invalid arguments accepted: %v", args)
		}
	}
	if _, failed := c3Call(t, a, "read_file", map[string]any{"path": path + "-missing"}); !failed {
		t.Fatal("missing file was not an error")
	}
	c3Write(t, path, strings.Repeat("x", inlineCap))
	if text, failed := c3Call(t, a, "read_file", map[string]any{"path": path}); failed || text != strings.Repeat("x", inlineCap) {
		t.Fatal("exact observation cap falsely truncated")
	}
	c3Write(t, path, "")
	if text, failed := c3Call(t, a, "read_file", map[string]any{"path": path}); failed || text != "" {
		t.Fatal("empty read failed")
	}
}

func TestChapter3Round1ShellEnvironmentAndResults(t *testing.T) {
	// Synthetic values only. Inspect the actual shell environment, never print a
	// credential. PTY output is merged in Chapter 4, not the old separated JSON.
	for _, name := range []string{"LLM_API_KEY", "ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY"} {
		t.Setenv(name, "synthetic-test-only")
	}
	t.Setenv("E3_BUILD_SETTING", "retained")
	t.Setenv("GOFLAGS", "-mod=readonly -p=1")
	a := c3Agent(t, ensemble.Config{})
	command := `if [ "${LLM_API_KEY+x}${ANTHROPIC_API_KEY+x}${OPENAI_API_KEY+x}${GEMINI_API_KEY+x}" != "" ]; then printf 'credential inherited' >&2; exit 91; fi
printf '%s|%s' "$E3_BUILD_SETTING" "$GOFLAGS"
printf 'separate stderr' >&2
exit 7`
	text, failed := c3Call(t, a, "run_command", map[string]any{"command": command, "cwd": t.TempDir()})
	if failed {
		t.Fatalf("normal exit became tool error: %s", text)
	}
	data, disk := c3JobOutput(t, a)
	if disk != "retained|-mod=readonly -p=1separate stderr\nexit_code: 7\n" || data.ExitCode == nil || *data.ExitCode != 7 || !strings.Contains(text, disk) {
		t.Fatalf("wrong merged output/exit: disk=%q job=%+v report=%q", disk, data, text)
	}
	for _, code := range []int{0, 127} {
		text, failed = c3Call(t, a, "run_command", map[string]any{"command": fmt.Sprintf("printf normal; exit %d", code)})
		data, disk = c3JobOutput(t, a)
		if failed || data.ExitCode == nil || *data.ExitCode != code || disk != fmt.Sprintf("normal\nexit_code: %d\n", code) || !strings.Contains(text, disk) {
			t.Fatalf("ordinary exit %d became startup failure: failed=%v job=%+v output=%q", code, failed, data, text)
		}
	}
	text, failed = c3Call(t, a, "run_command", map[string]any{"command": "printf should-not-run", "cwd": filepath.Join(t.TempDir(), "missing")})
	data, disk = c3JobOutput(t, a)
	if !failed || data.ExitCode != nil || strings.Contains(disk, "should-not-run") || !strings.Contains(disk, "starting command:") || !strings.Contains(text, disk) {
		t.Fatalf("startup failure not truthful: failed=%v job=%+v output=%q", failed, data, text)
	}
}

func TestChapter3Round1SearchGroups(t *testing.T) {
	a := c3Agent(t, ensemble.Config{})
	dir := t.TempDir()
	first := filepath.Join(dir, "a.txt")
	second := filepath.Join(dir, "b.txt")
	c3Write(t, first, "a1\nhit-a2\na3\nhit-a4\na5\na6\na7\nhit-a8\na9\n")
	c3Write(t, second, "b1\nhit-b2\nb3\n")
	text, failed := c3Call(t, a, "search_files", map[string]any{"path": dir, "pattern": "hit-", "context_lines": 1})
	want := first + "-1-a1\n" + first + ":2:hit-a2\n" + first + "-3-a3\n" + first + ":4:hit-a4\n" + first + "-5-a5\n--\n" + first + "-7-a7\n" + first + ":8:hit-a8\n" + first + "-9-a9\n--\n" + second + "-1-b1\n" + second + ":2:hit-b2\n" + second + "-3-b3\n"
	if failed || text != want {
		t.Fatalf("context groups mismatch: failed=%t\ngot %q\nwant %q", failed, text, want)
	}
	zero := first + ":2:hit-a2\n" + first + ":4:hit-a4\n" + first + ":8:hit-a8\n" + second + ":2:hit-b2\n"
	for _, explicit := range []bool{false, true} {
		args := map[string]any{"path": dir, "pattern": "hit-"}
		if explicit {
			args["context_lines"] = 0
		}
		if text, failed := c3Call(t, a, "search_files", args); failed || text != zero {
			t.Fatalf("zero-context mismatch: %q", text)
		}
	}
	// Windows touching at consecutive lines form one group, not two.
	c3Write(t, first, "a1\nhit-a2\na3\na4\nhit-a5\na6\n")
	text, failed = c3Call(t, a, "search_files", map[string]any{"path": first, "pattern": "hit-", "context_lines": 1})
	if failed || strings.Contains(text, "--\n") || strings.Count(text, "\n") != 6 {
		t.Fatalf("adjacent windows not merged: %q", text)
	}
}

func TestChapter3Round1WriteRefusalAndRegisteredHandlers(t *testing.T) {
	a := c3Agent(t, ensemble.Config{})
	dir := t.TempDir()
	target := filepath.Join(dir, "child", "file")
	if text, failed := c3Call(t, a, "write_file", map[string]any{"path": target, "content": "original"}); failed {
		t.Fatal(text)
	}
	text, failed := c3Call(t, a, "write_file", map[string]any{"path": target, "content": "destroy"})
	if !failed || !strings.Contains(text, "8 bytes") || !strings.Contains(text, "overwrite:true") || !strings.Contains(text, "append:true") || c3Contents(t, target) != "original" {
		t.Fatalf("unhelpful or destructive refusal: %q", text)
	}
	if text, failed := c3Call(t, a, "write_file", map[string]any{"path": target, "content": "+", "append": true}); failed || c3Contents(t, target) != "original+" {
		t.Fatalf("append failed: %s", text)
	}
	for _, args := range []map[string]any{
		{"overwrite": "true"}, {"append": nil}, {"overwrite": true, "append": true}, {"overwrite": nil},
	} {
		args["path"] = target
		args["content"] = "bad"
		if _, failed := c3Call(t, a, "write_file", args); !failed || c3Contents(t, target) != "original+" {
			t.Fatal("invalid write changed bytes")
		}
	}
	if text, failed := c3Call(t, a, "write_file", map[string]any{"path": target, "content": "replacement", "overwrite": true}); failed || c3Contents(t, target) != "replacement" {
		t.Fatalf("overwrite failed: %s", text)
	}
	if text, failed := c3Call(t, a, "list_directory", map[string]any{"path": dir}); failed || text != "child/\n" {
		t.Fatalf("list entry failed: %q", text)
	}
	if text, failed := c3Call(t, a, "edit_file", map[string]any{"path": target, "old_text": "replacement", "new_text": "edited"}); failed || c3Contents(t, target) != "edited" {
		t.Fatalf("edit entry failed: %q", text)
	}
}

func TestChapter3EditDiagnostics(t *testing.T) {
	a := c3Agent(t, ensemble.Config{})
	diagnosticBody := func(text, path, anchor, content string) (string, error) {
		// Only the required leading identifier is exempt from content checks. Never
		// remove path occurrences elsewhere: an echoed path in the body may leak aa.
		body, ok := strings.CutPrefix(text, fmt.Sprintf("file %q:", path))
		if !ok {
			return "", fmt.Errorf("refusal did not start with the requested file identifier: %q", text)
		}
		if strings.Contains(body, anchor) {
			return "", fmt.Errorf("diagnostic body leaked an anchor: %q", body)
		}
		if excerpt := strings.TrimSpace(content); excerpt != "" && strings.Contains(body, excerpt) {
			return "", fmt.Errorf("diagnostic body leaked a file excerpt: %q", body)
		}
		return body, nil
	}
	cases := []struct {
		name, content, anchor string
		count                 int
		locations             string
		omitted               int
	}{
		{"zero", "keep these bytes\n", "missing", 0, "", 0},
		{"multiple same line", "site site site", "site", 3, "[1 1 1]", 0},
		{"overlapping", "aaaa", "aa", 3, "[1 1 1]", 0},
		{"separate lines", "before\nsite\nother\nsite\n", "site", 2, "[2 4]", 0},
		{"multiline overlapping", "before\nx\nx\nx\n", "x\nx", 2, "[2 3]", 0},
		{"newline start", "a\nb\n", "\n", 2, "[1 2]", 0},
		{"bounded locations", strings.Repeat("site\n", 11), "site", 11, "[1 2 3 4 5 6 7 8]", 3},
		{"overlapping bounded locations", strings.Repeat("a", 12), "aa", 11, "[1 1 1 1 1 1 1 1]", 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Keep the path/anchor collision deterministic, independent of TMPDIR.
			path := filepath.Join(t.TempDir(), "aa-file")
			c3Write(t, path, tc.content)
			text, failed := c3Call(t, a, "edit_file", map[string]any{"path": path, "old_text": tc.anchor, "new_text": "must-not-appear"})
			if !failed {
				t.Fatalf("nonunique edit succeeded: %q", text)
			}
			if c3Contents(t, path) != tc.content {
				t.Fatal("refusal changed file bytes")
			}
			body, err := diagnosticBody(text, path, tc.anchor, tc.content)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(body, fmt.Sprintf("%d matches", tc.count)) || !strings.Contains(body, "file unchanged") || !strings.Contains(body, "reread the file") {
				t.Fatalf("incomplete refusal: %q", text)
			}
			if tc.count > 1 && (!strings.Contains(body, "match-start lines: "+tc.locations) || !strings.Contains(body, "larger exact anchor")) {
				t.Fatalf("wrong ambiguity locations/guidance: %q", text)
			}
			if tc.omitted > 0 {
				if !strings.Contains(body, fmt.Sprintf("location list truncated: %d more match locations omitted", tc.omitted)) {
					t.Fatalf("missing truncation/count: %q", text)
				}
			} else if strings.Contains(body, "truncated") || strings.Contains(body, "omitted") {
				t.Fatalf("untruncated locations reported as truncated: %q", text)
			}
			// The same checker must detect real body leaks, not merely accept
			// the legitimate collision in the required leading identifier.
			for _, leak := range []string{tc.anchor, tc.content} {
				if _, err := diagnosticBody(text+"; leaked: "+leak, path, tc.anchor, tc.content); err == nil {
					t.Fatal("diagnostic checker accepted a body anchor/excerpt")
				}
			}
			if tc.anchor == "aa" {
				if _, err := diagnosticBody(text+"; repeated: "+fmt.Sprintf("file %q:", path), path, tc.anchor, tc.content); err == nil {
					t.Fatal("diagnostic checker removed a path occurrence from the body")
				}
			}
		})
	}
}

func TestChapter3EditUniqueAndDeletion(t *testing.T) {
	a := c3Agent(t, ensemble.Config{})
	for _, tc := range []struct{ name, content, anchor, replacement, want string }{
		{"unique multiline", "before\nanchor\ncontinued\nafter", "anchor\ncontinued", "changed", "before\nchanged\nafter"},
		{"unique deletion", "before\nremove\nafter", "remove\n", "", "before\nafter"},
		{"whole-file deletion", "delete me", "delete me", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "file")
			c3Write(t, path, tc.content)
			text, failed := c3Call(t, a, "edit_file", map[string]any{"path": path, "old_text": tc.anchor, "new_text": tc.replacement})
			if failed || c3Contents(t, path) != tc.want {
				t.Fatalf("unique edit mismatch: failed=%t result=%q", failed, text)
			}
		})
	}
}
