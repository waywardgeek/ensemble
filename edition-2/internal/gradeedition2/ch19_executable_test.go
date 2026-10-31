package gradeedition2

import (
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

func TestSandboxExecutableExplanations(t *testing.T) {
	s := sandboxJobScript()
	s.explained = true
	r := s.route(nil)
	if r.Text != s.firstText {
		t.Fatal("lost credential echo")
	}
	r = s.route(sandboxJobBody(r.ToolID, "job 41 running; pending", false))
	if r.ToolName != "wait_for_job" || strings.TrimSpace(r.Text) == "" {
		t.Fatal("unexplained wait", r)
	}
	r = s.route(sandboxJobBody(r.ToolID, "exit_code: 0", false))
	if r.ToolID != "file" || strings.TrimSpace(r.Text) == "" {
		t.Fatal("unexplained next tool", r)
	}
}

func TestSandboxExecutableRefusal(t *testing.T) {
	for _, tc := range []struct {
		id, text  string
		bad, want bool
	}{
		{"call", "tool is not enabled: run_command", true, true},
		{"call", "tool is not enabled: run_command", false, false},
		{"wrong", "tool is not enabled: run_command", true, false},
		{"call", "unknown job handle", true, false},
	} {
		c := grade.Check{Passed: true}
		sandboxExecutableRefusal(&c, fakevendor.Recorded{Body: sandboxJobBody(tc.id, tc.text, tc.bad)}, "call")
		if c.Passed != tc.want {
			t.Fatalf("%+v: %+v", tc, c)
		}
	}
	if strings.Contains(sandboxExecutableGitCommand, "commit") || !strings.Contains(sandboxExecutableGitCommand, "git show :note") {
		t.Fatal("Git control must inspect actual staged content without committing")
	}
}
