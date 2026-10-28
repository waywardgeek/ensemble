package gradeedition2

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

func TestChapter18ExecutableStatusCorrelation(t *testing.T) {
	message := map[string]any{"role": "tool", "tool_call_id": "status-0", "content": "model: fixture"}
	for _, tc := range []struct {
		name     string
		messages []any
		pass     bool
	}{
		{"paired", []any{message}, true},
		{"duplicate", []any{message, message}, false},
		{"wrong id", []any{map[string]any{"role": "tool", "tool_call_id": "other", "content": "model: fixture"}}, false},
		{"assistant text", []any{map[string]any{"role": "assistant", "content": "model: fixture"}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(map[string]any{"messages": tc.messages})
			c := grade.Check{Passed: true, Points: 20, Earned: 20}
			text := statusExecutableResult(&c, fakevendor.Recorded{Body: body}, "status-0")
			if c.Passed != tc.pass || (tc.pass && text != "model: fixture") {
				t.Fatalf("%+v text=%q", c, text)
			}
		})
	}
}

func TestChapter18ExecutableFixtures(t *testing.T) {
	reply := statusExecutableReply("gpt-4o-mini", 2, fakevendor.Canonical{Input: 100, CacheRead: 20, Output: 3})
	var body map[string]any
	if err := json.Unmarshal([]byte(reply.JSONBody), &body); err != nil {
		t.Fatal(err)
	}
	usage := object(body["usage"])
	if body["model"] != "gpt-4o-mini" || body["service_tier"] != "default" || usage["prompt_tokens"] != float64(120) || object(usage["prompt_tokens_details"])["cached_tokens"] != float64(20) || usage["cache_write_input_tokens"] != nil {
		t.Fatal(body)
	}
	message := object(object(list(body["choices"])[0])["message"])
	call := object(list(message["tool_calls"])[0])
	if call["id"] != "status-2" || object(call["function"])["name"] != "agent_status" || object(call["function"])["arguments"] != "{}" {
		t.Fatal(call)
	}
	source := t.TempDir()
	if err := os.MkdirAll(filepath.Join(source, "web-search"), 0700); err != nil {
		t.Fatal(err)
	}
	manual := "---\nname: web-search\ndescription: preserve this shipped manual\ntype: loadable\ntools: firecrawl_search firecrawl_scrape\nmcp_servers:\n  - name: web\n    transport: http\n    url: https://external.invalid/mcp\n    auth-env: ORIGINAL_TOKEN\n---\nKeep the original teaching text.\n"
	if err := os.WriteFile(filepath.Join(source, "web-search", "SKILL.md"), []byte(manual), 0600); err != nil {
		t.Fatal(err)
	}
	for _, executable := range []bool{false, true} {
		dir, err := webSkillsFor(t.TempDir(), "http://127.0.0.1:12345", executable, source)
		if err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(dir, "web-search", "SKILL.md"))
		if err != nil {
			t.Fatal(err)
		}
		if executable && (!strings.Contains(string(data), "Keep the original teaching text.") || strings.Contains(string(data), "external.invalid") || strings.Contains(string(data), "ORIGINAL_TOKEN")) {
			t.Fatal(string(data))
		}
		transport := "url"
		if executable {
			transport = "http"
		}
		if !strings.Contains(string(data), "transport: "+transport+"\n") || !strings.Contains(string(data), "auth-env: EDITION2_WEB_TOKEN") {
			t.Fatal(string(data))
		}
	}
}

func TestChapter18ExecutableAuthorityRefusal(t *testing.T) {
	for _, tc := range []struct{ executable, isError, pass bool }{
		{true, true, true}, {false, true, false}, {true, false, false},
	} {
		body, _ := json.Marshal(map[string]any{"messages": []any{map[string]any{"role": "user", "content": []any{map[string]any{
			"type": "tool_result", "tool_use_id": "before", "is_error": tc.isError, "content": "tool is not enabled: firecrawl_search",
		}}}}})
		c := grade.Check{Passed: true, Points: 15, Earned: 15}
		webUnknownFor(&c, fakevendor.Recorded{Body: body}, "before", tc.executable)
		if c.Passed != tc.pass {
			t.Fatalf("%+v: %+v", tc, c)
		}
	}
}
