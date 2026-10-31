package gradeedition2

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

func sandboxJobBody(id, text string, bad bool) []byte {
	b, _ := json.Marshal(map[string]any{"messages": []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "tool_result", "tool_use_id": id, "content": text, "is_error": bad}}}}})
	return b
}
func sandboxJobScript() *sandboxScript {
	return &sandboxScript{probes: []sandboxProbe{{"git", "run_command", map[string]any{"command": "owned-command"}}, {"file", "read_file", map[string]any{"path": "owned.txt"}}}, firstText: "CREDENTIAL_ECHO_fake", final: "COMPLETE", observed: map[string]sandboxObservation{}}
}

func TestSandboxScriptWaitsForActualJob(t *testing.T) {
	s := sandboxJobScript()
	first := s.route(nil)
	if first.ToolID != "git" || first.Text != s.firstText {
		t.Fatal(first)
	}
	body := sandboxJobBody("git", "job 41 running; 12 bytes\nEARLY_OUTPUT", false)
	saved := append([]byte(nil), body...)
	wait := s.route(body)
	var args map[string]any
	if err := json.Unmarshal([]byte(wait.ToolArgs), &args); err != nil {
		t.Fatal(err)
	}
	if wait.ToolName != "wait_for_job" || wait.ToolID == "git" || args["handle"] != float64(41) || args["ai_callback_delay"] != float64(5) {
		t.Fatalf("bad wait: %#v %v", wait, args)
	}
	// Further running observations get new call IDs, never rerun the command.
	nextWait := s.route(sandboxJobBody(wait.ToolID, "job 41 running; 12 bytes", false))
	if nextWait.ToolName != "wait_for_job" || nextWait.ToolID == wait.ToolID {
		t.Fatal(nextWait)
	}
	next := s.route(sandboxJobBody(nextWait.ToolID, "LATE_OUTPUT\nexit_code: 0", false))
	if next.ToolID != "file" {
		t.Fatal(next)
	}
	final := s.route(sandboxJobBody("file", "FILE_CONTENT", false))
	if final.Text != s.final || s.err != nil || s.totalWaits != 2 || s.next != 2 || !s.finished {
		t.Fatalf("state %#v", s)
	}
	text, bad, err := s.observed["git"].result()
	if err != nil || bad || !strings.Contains(text, "EARLY_OUTPUT") || !strings.Contains(text, "LATE_OUTPUT") || !strings.Contains(text, "exit_code: 0") {
		t.Fatalf("%q %v %v", text, bad, err)
	}
	if !bytes.Equal(s.observed["git"][0].request.Body, saved) {
		t.Fatal("recorded request rewritten")
	}
	body[0] = '!'
	if !bytes.Equal(s.observed["git"][0].request.Body, saved) {
		t.Fatal("observation aliases caller bytes")
	}
	s.route(nil)
	if s.err == nil {
		t.Fatal("extra post-final request accepted")
	}
}

func TestSandboxScriptRejectsUnpairedOrChangedJob(t *testing.T) {
	t.Run("wrong-id", func(t *testing.T) {
		s := sandboxJobScript()
		s.route(nil)
		s.route(sandboxJobBody("other", "exit_code: 0", false))
		if s.err == nil || s.next != 0 {
			t.Fatal("uncorrelated result accepted")
		}
	})
	t.Run("changed-handle", func(t *testing.T) {
		s := sandboxJobScript()
		s.route(nil)
		w := s.route(sandboxJobBody("git", "job 41 running; pending", false))
		s.route(sandboxJobBody(w.ToolID, "job 42 running; pending", false))
		if s.err == nil || s.next != 0 {
			t.Fatal("changed handle accepted")
		}
	})
	t.Run("bounded", func(t *testing.T) {
		s := sandboxJobScript()
		r := s.route(nil)
		for n := 0; n < 9; n++ {
			r = s.route(sandboxJobBody(r.ToolID, "job 41 running; pending", false))
		}
		if s.err == nil || s.totalWaits != 8 || s.next != 0 {
			t.Fatalf("unbounded %#v", s)
		}
	})
	t.Run("refusal-is-not-running", func(t *testing.T) {
		s := sandboxJobScript()
		s.route(nil)
		r := s.route(sandboxJobBody("git", "command denied", true))
		if s.totalWaits != 0 || r.ToolID != "file" {
			t.Fatal(r)
		}
		_, bad, e := s.observed["git"].result()
		if e != nil || !bad {
			t.Fatal("refusal erased")
		}
	})
}

func TestSandboxExplicitOutputBudget(t *testing.T) {
	p := &mcpCommand{lines: make(chan string, 1)}
	go func() { time.Sleep(20 * time.Millisecond); p.lines <- "terminal" }()
	if line, err := p.readWithin(time.Second); err != nil || line != "terminal" {
		t.Fatalf("%q %v", line, err)
	}
	if _, err := p.readWithin(time.Millisecond); err == nil || !strings.Contains(err.Error(), "deadline") {
		t.Fatal("unbounded empty read", err)
	}
	p.lines <- "legacy"
	if line, err := p.read(); err != nil || line != "legacy" {
		t.Fatalf("default read changed: %q %v", line, err)
	}
}

func TestSandboxCommandTerminalEvidence(t *testing.T) {
	for _, tc := range []struct {
		name, text, id     string
		bad, success, want bool
	}{
		{"completed", "COMMAND_OK\nexit_code: 0", "call", false, true, true},
		{"running", "job 41 running; pending", "call", false, true, false},
		{"missing-exit", "COMMAND_OK", "call", false, true, false},
		{"nonzero-positive", "COMMAND_OK\nexit_code: 1", "call", false, true, false},
		{"wrong-id", "exit_code: 0", "other", false, true, false},
		{"negative-refusal", "denied", "call", true, false, true},
		{"negative-nonzero", "exit_code: 1", "call", false, false, true},
		{"negative-success", "exit_code: 0", "call", false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := grade.Check{Passed: true}
			o := sandboxObservation{{fakevendor.Recorded{Body: sandboxJobBody(tc.id, tc.text, tc.bad)}, "call"}}
			sandboxCommand(&c, o, "probe", tc.success)
			if c.Passed != tc.want {
				t.Fatalf("passed=%v want=%v %v", c.Passed, tc.want, c.Details)
			}
		})
	}
	c := grade.Check{Passed: true}
	o := sandboxObservation{{fakevendor.Recorded{Body: sandboxJobBody("start", "job 41 running; output says exit_code: 0", false)}, "start"}, {fakevendor.Recorded{Body: sandboxJobBody("wait", "done, missing terminal exit", false)}, "wait"}}
	sandboxCommand(&c, o, "probe", true)
	if c.Passed {
		t.Fatal("earlier output substituted for terminal exit")
	}
}

func TestSandboxGitCommandContract(t *testing.T) {
	const legacy = "git init -q repo && cd repo && printf git-control > note && git add note && git commit -qm git-control && git log -1 --format=%s"
	if got := sandboxGitCommand(false); got != legacy {
		t.Fatalf("legacy changed: %s", got)
	}
	want := strings.Replace(legacy, "git log -1", "git --no-pager log -1", 1)
	if got := sandboxGitCommand(true); got != want {
		t.Fatalf("current changed beyond pager control: %s", got)
	}
}
