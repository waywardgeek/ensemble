package gradeedition2

import (
	"encoding/json"
	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"io"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestChapter15FixtureEnvironment(t *testing.T) {
	dir := t.TempDir()
	opts := invoiceOpts{dir: dir, vendor: "anthropic", model: "claude-sonnet-5"}
	env := invoiceEnv([]string{"PATH=/owned/path", "HOME=/owned/home", "GOFLAGS=-mod=readonly -p=1",
		"LLM_MODEL=wrong", "LLM_TOOLS=none", "LLM_PAUSE_ON_START=true", "LLM_SPEAKING=true",
		"ENSEMBLE_SAVE=/foreign/save", "ENSEMBLE_LOG_MODE=dialogue_only", "LLM_LOG_PATH=/foreign/log",
		"EN_MCP_SERVERS=foreign", "EN_SKILLS_DIR=/foreign/skills", "ANTHROPIC_API_KEY=not-a-real-key",
		"OPENAI_BASE_URL=https://foreign.invalid", "GEMINI_API_KEY=not-a-real-key"}, "/owned/skills", "http://127.0.0.1:1", opts)
	values := map[string]string{}
	for _, item := range env {
		k, v, _ := strings.Cut(item, "=")
		if _, ok := values[k]; ok {
			t.Fatalf("duplicate key %s", k)
		}
		values[k] = v
	}
	for key, want := range map[string]string{
		"PATH": "/owned/path", "HOME": "/owned/home", "GOFLAGS": "-mod=readonly -p=1",
		"ENSEMBLE_SAVE": filepath.Join(dir, "save.json"), "ENSEMBLE_LOG_MODE": "full",
		"LLM_LOG_PATH": filepath.Join(dir, "events.jsonl"), "LLM_MODEL": opts.model, "LLM_RESOLVED_MODEL": opts.model,
		"LLM_TOOLS": "default", "LLM_PAUSE_ON_START": "false", "LLM_SPEAKING": "false",
		"EN_SKILLS_DIR": "/owned/skills", "EN_PRIMARY_SKILL": "base", "EN_DISABLE_STREAMING": "1", "LLM_STREAM_CAPABILITY": "plain",
		"LLM_BASE_URL": "http://127.0.0.1:1", "LLM_API_KEY": "fake",
	} {
		if values[key] != want {
			t.Errorf("%s=%q want %q", key, values[key], want)
		}
	}
	if _, ok := values["EN_MCP_SERVERS"]; ok {
		t.Fatal("inherited MCP process remained enabled")
	}
}

// Exercise the exact successful body through the real fake HTTP server; a
// predicate over an unused map would not verify the fixture sent to a client.
func TestChapter15OpenAIUsageFixture(t *testing.T) {
	u := fakevendor.Canonical{Input: 11, CacheWrite: 22, CacheRead: 33, Output: 44}
	body := invoiceOpenAIUsage("gpt-5.6", u)
	s := fakevendor.New([]fakevendor.Reply{{JSONBody: body}})
	defer s.Close()
	r, err := http.Post(s.URL()+"/v1/chat/completions", "application/json", strings.NewReader(`{"model":"gpt-5.6","stream":false}`))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
	}
	if r.StatusCode != 200 || string(b) != body {
		t.Fatalf("status=%d body=%s", r.StatusCode, b)
	}
	var got struct {
		Model string
		Usage struct {
			Prompt  int            `json:"prompt_tokens"`
			Output  int            `json:"completion_tokens"`
			Total   int            `json:"total_tokens"`
			Details map[string]int `json:"prompt_tokens_details"`
		}
	}
	if err = json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.Model != "gpt-5.6" || got.Usage.Prompt != 66 || got.Usage.Output != 44 || got.Usage.Total != 110 || got.Usage.Details["cached_tokens"] != 33 || got.Usage.Details["cache_write_input_tokens"] != 22 {
		t.Fatalf("wrong wire usage: %+v", got)
	}
	if _, old := got.Usage.Details["cache_write_tokens"]; old {
		t.Fatal("legacy field in documented Chapter15 packet")
	}
	if got.Usage.Prompt-got.Usage.Details["cached_tokens"]-got.Usage.Details["cache_write_input_tokens"] != u.Input {
		t.Fatal("disjoint input not recoverable")
	}
}

func TestChapter15SnapshotFrames(t *testing.T) {
	usage := map[string]any{"type": "usage", "input": float64(11), "cache_read": float64(22),
		"total": map[string]any{"input": float64(101), "cache_read": float64(202)}}
	settings := map[string]any{"type": "current_settings", "settings": map[string]any{"theme": "dark"}}
	replay := map[string]any{"type": "replay", "data": map[string]any{"events": []any{
		map[string]any{"type": "text", "text": "not usage"}, usage,
	}}}
	for name, frames := range map[string][]map[string]any{
		"replay-before-settings":     {replay, settings},
		"replay-after-settings":      {settings, replay},
		"standalone-before-settings": {usage, settings},
		"standalone-after-settings":  {settings, usage},
	} {
		t.Run(name, func(t *testing.T) {
			p := &guiPeer{messages: make(chan map[string]any, len(frames)), errors: make(chan error)}
			for _, frame := range frames {
				p.messages <- frame
			}
			got, err := invoiceReadSnapshot(p)
			if err != nil || !reflect.DeepEqual(got, usage) {
				t.Fatalf("got %#v, %v", got, err)
			}
			// The observer returns the actual map, not a normalized copy.
			got["identity_probe"] = name
			if usage["identity_probe"] != name {
				t.Fatal("usage map was replaced")
			}
			delete(usage, "identity_probe")
		})
	}
	for name, frames := range map[string][]map[string]any{
		"settings-only":          {settings},
		"auxiliary-is-not-usage": {settings, {"type": "meter", "data": usage}},
		"unrelated-nested-usage": {settings, {"type": "other", "data": map[string]any{"events": []any{usage}}}},
		"usage-without-settings": {usage},
		"malformed-settings":     {usage, {"type": "current_settings", "settings": "invalid"}},
	} {
		t.Run(name, func(t *testing.T) {
			var seen invoiceSnapshotFrames
			for _, frame := range frames {
				seen.observe(frame)
			}
			if seen.settings && seen.usage != nil && seen.err == nil {
				t.Fatal("incomplete/invalid snapshot accepted")
			}
		})
	}
	p := &guiPeer{messages: make(chan map[string]any, 2), errors: make(chan error)}
	p.messages <- usage
	p.messages <- map[string]any{"type": "current_settings"}
	if _, err := invoiceReadSnapshot(p); err == nil || !strings.Contains(err.Error(), "settings object") {
		t.Fatalf("malformed settings not refused: %v", err)
	}
}

func TestChapter15AnthropicIdentityFixture(t *testing.T) {
	for _, model := range []string{"claude-sonnet-4-6", "claude-sonnet-5-course"} {
		t.Run(model, func(t *testing.T) {
			u := fakevendor.Canonical{Input: 11, CacheWrite: 22, CacheRead: 33, Output: 44}
			text := "Preserve \"quoted\" text\nand its newline."
			body := invoiceAnthropicUsage(model, text, u)
			s := fakevendor.New([]fakevendor.Reply{{JSONBody: body}})
			defer s.Close()
			r, err := http.Post(s.URL()+"/v1/messages", "application/json", strings.NewReader(`{"model":"`+model+`","stream":false}`))
			if err != nil {
				t.Fatal(err)
			}
			defer r.Body.Close()
			b, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			if r.StatusCode != 200 || string(b) != body {
				t.Fatalf("status=%d body=%s", r.StatusCode, b)
			}
			var got struct {
				Model, Type, Role string
				Stop              string `json:"stop_reason"`
				Content           []struct{ Type, Text string }
				Usage             map[string]int
			}
			if err := json.Unmarshal(b, &got); err != nil {
				t.Fatal(err)
			}
			want := map[string]int{"input_tokens": 11, "cache_creation_input_tokens": 22, "cache_read_input_tokens": 33, "output_tokens": 44}
			if got.Model != model || got.Type != "message" || got.Role != "assistant" || got.Stop != "end_turn" || len(got.Content) != 1 || got.Content[0].Type != "text" || got.Content[0].Text != text || !reflect.DeepEqual(got.Usage, want) {
				t.Fatalf("wrong Messages response: %+v", got)
			}
		})
	}
}
