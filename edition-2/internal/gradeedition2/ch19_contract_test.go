package gradeedition2

import (
	"encoding/json"
	"reflect"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
	"path/filepath"
	"strings"
	"testing"
)

func TestSandboxCurrentEnvironment(t *testing.T) {
	f := sandboxFixture{root: t.TempDir(), skills: "/fixture/skills", credential: "sk-course-only"}
	env := sandboxEnv([]string{"PATH=/usr/bin:/bin", "HOME=/host", "TMPDIR=/host/tmp", "LLM_MODEL=host", "LLM_BASE_URL=https://host.invalid", "OPENAI_BILLING_MODE=chatgpt", "OPENAI_REASONING_EFFORT=high", "OPENAI_REASONING_SUMMARY=detailed", "LLM_REQUIRE_TOOL_EXPLANATION=true", "LLM_SURFACE=responses", "OPENAI_API_KEY=host-secret", "SERVICE_TOKEN=host-secret", "EN_PRIMARY_SKILL=host", "CH07_AGENT_LOG=/host/log", "CH02_LOG=/host/log", "E3_TRACE=/host/trace", "EDITION2_WEB_TOKEN=host-secret"}, f, "http://127.0.0.1:1")
	got := map[string]string{}
	for _, item := range env {
		k, v, _ := strings.Cut(item, "=")
		if _, duplicate := got[k]; duplicate {
			t.Fatalf("duplicate setting %s", k)
		}
		got[k] = v
		if strings.Contains(item, "host-secret") {
			t.Fatal("inherited secret survived")
		}
	}
	for _, k := range []string{"OPENAI_BILLING_MODE", "OPENAI_REASONING_EFFORT", "OPENAI_REASONING_SUMMARY", "LLM_REQUIRE_TOOL_EXPLANATION", "LLM_SURFACE", "CH07_AGENT_LOG", "CH02_LOG", "E3_TRACE", "EDITION2_WEB_TOKEN"} {
		if _, ok := got[k]; ok {
			t.Errorf("inherited %s survived", k)
		}
	}
	for k, want := range map[string]string{"HOME": f.root, "TMPDIR": f.root, "PATH": "/usr/bin:/bin", "LLM_VENDOR": "anthropic", "LLM_MODEL": "claude-sonnet-4-6", "LLM_BASE_URL": "http://127.0.0.1:1", "EN_SKILLS_DIR": f.skills, "EN_PRIMARY_SKILL": "base", "ENSEMBLE_SAVE": filepath.Join(f.root, "save.json"), "LLM_LOG_PATH": filepath.Join(f.root, "events.jsonl"), "EN_DISABLE_STREAMING": "1", "SERVICE_TOKEN": f.credential, "LLM_API_KEY": f.credential, "OPENAI_API_KEY": f.credential, "ANTHROPIC_API_KEY": f.credential, "GEMINI_API_KEY": f.credential} {
		if got[k] != want {
			t.Errorf("%s=%q want%q", k, got[k], want)
		}
	}
}

func TestSandboxLogProbeContract(t *testing.T) {
	current := sandboxLogProbes(true)
	want := []sandboxProbe{
		{"log-limit", "tool_limits", map[string]any{"max_output_bytes": 65536}},
		{"log", "read_file", map[string]any{"path": "api.log", "max_bytes": 65536}},
	}
	if !reflect.DeepEqual(current, want) {
		t.Fatalf("current probe contract: %#v", current)
	}
	legacy := []sandboxProbe{{"log", "read_file", map[string]any{"path": "api.log", "max_bytes": 65536, "max_output_bytes": 65536}}}
	if !reflect.DeepEqual(sandboxLogProbes(false), legacy) {
		t.Fatal("historical probes changed")
	}
}

func TestSandboxSafeRefusalContract(t *testing.T) {
	for _, tc := range []struct {
		name, text, id         string
		isError, current, want bool
	}{
		{"disabled", "run_command is disabled in safe mode", "run_command", true, true, true},
		{"unknown", "unknown tool run_command", "run_command", true, true, true},
		{"success", "run_command is disabled in safe mode", "run_command", false, true, false},
		{"empty", " ", "run_command", true, true, false},
		{"uncorrelated", "unknown tool run_command", "other", true, true, false},
		{"legacy-unknown", "unknown tool run_command", "run_command", true, false, true},
		{"legacy-disabled", "run_command is disabled in safe mode", "run_command", true, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]any{"messages": []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "tool_result", "tool_use_id": tc.id, "is_error": tc.isError, "content": tc.text}}}}})
			if err != nil {
				t.Fatal(err)
			}
			c := grade.Check{Passed: true}
			sandboxSafeRefusal(&c, fakevendor.Recorded{Body: body}, "run_command", tc.current)
			if c.Passed != tc.want {
				t.Fatalf("passed=%v want=%v details=%v", c.Passed, tc.want, c.Details)
			}
		})
	}
}
