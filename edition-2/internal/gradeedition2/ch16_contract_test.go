package gradeedition2

import (
	"github.com/waywardgeek/ensemble/internal/grade"
	"path/filepath"
	"strings"
	"testing"
)

func TestChapter16FixtureEnvironment(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "profile")
	skills := filepath.Join(dir, "skills")
	env := subscriptionEnv([]string{"PATH=/bin", "HOME=/foreign", "GOFLAGS=-mod=readonly -p=1", "OTHER=keep", "LLM_API_KEY=foreign", "LLM_MODEL=wrong", "LLM_STREAM_CAPABILITY=disabled", "EN_MCP_SERVERS=foreign", "ENSEMBLE_SAVE=/foreign", "OPENAI_API_KEY=foreign", "OPENAI_OIDC_ISSUER=https://foreign.invalid", "OPENAI_BILLING_MODE=api", "ANTHROPIC_API_KEY=foreign", "GEMINI_API_KEY=foreign", "GOOGLE_API_KEY=foreign", "CH02_LOG=/foreign"}, home, dir, skills, "http://127.0.0.1:1", "http://127.0.0.1:2")
	got := map[string]string{}
	for _, v := range env {
		k, x, _ := strings.Cut(v, "=")
		if _, ok := got[k]; ok {
			t.Fatalf("duplicate %s", k)
		}
		got[k] = x
	}
	for k, v := range map[string]string{"PATH": "/bin", "GOFLAGS": "-mod=readonly -p=1", "OTHER": "keep", "HOME": home, "EN_HOME": home, "OPENAI_OIDC_ISSUER": "http://127.0.0.1:1", "LLM_BASE_URL": "http://127.0.0.1:2", "LLM_VENDOR": "openai", "LLM_MODEL": "gpt-ch19-course", "LLM_RESOLVED_MODEL": "gpt-ch19-course", "LLM_API_KEY": grade.Edition2APIKey, "LLM_STREAM_CAPABILITY": "sse", "OPENAI_REASONING_EFFORT": "medium", "OPENAI_REASONING_SUMMARY": "detailed", "EN_BROWSER_CMD": "true", "EN_SKILLS_DIR": skills, "EN_PRIMARY_SKILL": "base", "EN_DISABLE_STREAMING": "0", "LLM_LOG_PATH": filepath.Join(dir, "events.jsonl"), "LLM_SAVE_PATH": filepath.Join(dir, "save.json"), "LLM_DATA_DIR": filepath.Join(dir, "cr"), "LLM_LOG_MODE": "full", "LLM_TOOLS": "default", "LLM_PAUSE_ON_START": "0", "LLM_SPEAKING": "0", "CH02_LOG": filepath.Join(dir, "events.jsonl")} {
		if got[k] != v {
			t.Errorf("%s=%q want%q", k, got[k], v)
		}
	}
	for _, k := range []string{"EN_MCP_SERVERS", "ENSEMBLE_SAVE", "OPENAI_API_KEY", "OPENAI_BILLING_MODE", "ANTHROPIC_API_KEY", "GEMINI_API_KEY", "GOOGLE_API_KEY"} {
		if _, ok := got[k]; ok {
			t.Errorf("inherited %s", k)
		}
	}
}

func TestChapter16ActualErrorFrames(t *testing.T) {
	for _, tc := range []struct {
		name   string
		frames []map[string]any
		want   string
	}{
		{"actual", []map[string]any{{"type": "error", "data": "quota refused"}, {"type": "turn_ended", "text": ""}}, "quota refused"},
		{"legacy", []map[string]any{{"type": "turn_ended", "error": "quota refused"}}, "quota refused"},
		{"success", []map[string]any{{"type": "part_delta", "kind": "text", "chunk": "error is a word"}, {"type": "turn_ended", "text": "done"}}, ""},
		{"unrelated", []map[string]any{{"type": "trace", "data": "quota refused"}, {"type": "turn_ended"}}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := subscriptionTurnError(tc.frames); got != tc.want {
				t.Fatalf("%q != %q", got, tc.want)
			}
		})
	}
}
