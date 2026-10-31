package gradeedition2

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/gorilla/websocket"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

func TestChapter18FixtureEnvironment(t *testing.T) {
	work := t.TempDir()
	env := webStatusEnv([]string{"PATH=/safe/bin", "HOME=/owned/home", "GOCACHE=/cache", "GOPROXY=off", "LLM_MODEL=host", "LLM_MODEL=duplicate", "LLM_BASE_URL=https://external.invalid", "EN_MCP_JSON=host", "ENSEMBLE_SAVE=/host/save", "ANTHROPIC_API_KEY=real", "OPENAI_PROFILE=host", "GEMINI_API_KEY=real", "CH07_AGENT_LOG=/host/gui", "CH02_LOG=/host/cli", "E3_TRACE=/host/trace", "EDITION2_WEB_TOKEN=host"}, work, "/owned/skills", "http://127.0.0.1:12345", "claude-sonnet-4-6", "ensemble")
	got := map[string]string{}
	for _, s := range env {
		k, v, _ := strings.Cut(s, "=")
		if _, exists := got[k]; exists {
			t.Fatalf("duplicate %s", k)
		}
		got[k] = v
	}
	for k, want := range map[string]string{"PATH": "/safe/bin", "HOME": "/owned/home", "GOCACHE": "/cache", "GOPROXY": "off", "LLM_MODEL": "claude-sonnet-4-6", "LLM_VENDOR": "anthropic", "LLM_RESOLVED_MODEL": "", "LLM_BASE_URL": "http://127.0.0.1:12345", "EN_PRIMARY_SKILL": "ensemble", "EN_SKILLS_DIR": "/owned/skills", "ENSEMBLE_SAVE": filepath.Join(work, "save.json"), "LLM_LOG_PATH": filepath.Join(work, "events.jsonl"), "ENSEMBLE_LOG_MODE": "full", "LLM_PAUSE_ON_START": "false", "LLM_SPEAKING": "false", "LLM_STREAM_CAPABILITY": "plain", "LLM_TOOLS": "default", "EN_DISABLE_STREAMING": "1", "ANTHROPIC_API_KEY": "fake", "OPENAI_API_KEY": "fake", "GEMINI_API_KEY": "fake"} {
		v, exists := got[k]
		if !exists || v != want {
			t.Errorf("%s=%q present%v want%q", k, v, exists, want)
		}
	}
	for _, k := range []string{"CH07_AGENT_LOG", "CH02_LOG", "E3_TRACE", "EDITION2_WEB_TOKEN", "EN_MCP_JSON", "OPENAI_PROFILE"} {
		if _, exists := got[k]; exists {
			t.Errorf("inherited %s", k)
		}
	}
}

func TestChapter18StatusFixturePackets(t *testing.T) {
	for _, model := range []string{"claude-sonnet-4-6", "claude-opus-4-6", "claude-course-unpriced"} {
		t.Run(model, func(t *testing.T) {
			u := fakevendor.Canonical{Input: 11, Output: 22, CacheRead: 33, CacheWrite: 44}
			replies := []fakevendor.Reply{statusReply(model, 2, u), {JSONBody: invoiceAnthropicUsage(model, "DONE", fakevendor.Canonical{})}}
			srv := fakevendor.New(replies)
			defer srv.Close()
			for i := range replies {
				resp, err := http.Post(srv.URL()+"/v1/messages", "application/json", strings.NewReader(`{"model":"`+model+`","messages":[]}`))
				if err != nil {
					t.Fatal(err)
				}
				var body map[string]any
				err = json.NewDecoder(resp.Body).Decode(&body)
				resp.Body.Close()
				if err != nil || resp.StatusCode != 200 || body["model"] != model {
					t.Fatalf("response %d %v status%d %v", i, err, resp.StatusCode, body)
				}
				usage := object(body["usage"])
				want := fakevendor.Canonical{}
				if i == 0 {
					want = u
				}
				for k, n := range map[string]int{"input_tokens": want.Input, "output_tokens": want.Output, "cache_read_input_tokens": want.CacheRead, "cache_creation_input_tokens": want.CacheWrite} {
					if usage[k] != float64(n) {
						t.Errorf("%s=%v want%d", k, usage[k], n)
					}
				}
				block := object(list(body["content"])[0])
				if i == 0 && (block["type"] != "tool_use" || block["id"] != "status-2" || block["name"] != "agent_status" || body["stop_reason"] != "tool_use") {
					t.Fatalf("tool packet %v", body)
				}
				if i == 1 && (block["type"] != "text" || block["text"] != "DONE" || body["stop_reason"] != "end_turn") {
					t.Fatalf("final packet %v", body)
				}
			}
			if len(srv.Requests()) != 2 {
				t.Fatal("request capture lost")
			}
		})
	}
}

func TestChapter18ImportBoundary(t *testing.T) {
	const base = "course.invalid/core"
	for _, name := range []string{"tool", "tools"} {
		pkgs := []listedPackage{{ImportPath: base + "/internal/common"}, {ImportPath: base + "/internal/" + name, Imports: []string{base + "/internal/common", "fmt"}}}
		if c := statusImportBoundary(pkgs, base); !c.Passed {
			t.Fatalf("%s: %+v", name, c)
		}
		pkgs[1].Imports = append(pkgs[1].Imports, base+"/internal/engine")
		if c := statusImportBoundary(pkgs, base); c.Passed {
			t.Fatalf("%s accepted direct engine import", name)
		}
	}
	if c := statusImportBoundary([]listedPackage{{ImportPath: base + "/internal/common"}}, base); c.Passed {
		t.Fatal("missing tools boundary accepted")
	}
}

func TestChapter18ModelControl(t *testing.T) {
	for _, tc := range []struct {
		name      string
		live      bool
		frames    []map[string]any
		wantError bool
	}{
		{"live", true, []map[string]any{
			{"type": "usage", "agent_id": "other", "meter": map[string]any{"selected_model": "new"}},
			{"type": "usage", "agent_id": "owned", "meter": map[string]any{"selected_model": "old"}},
			{"type": "usage", "agent_id": "owned", "meter": map[string]any{"selected_model": "new"}}}, false},
		{"wrong-agent", true, []map[string]any{{"type": "usage", "agent_id": "other", "meter": map[string]any{"selected_model": "new"}}}, true},
		{"wrong-model", true, []map[string]any{{"type": "usage", "agent_id": "owned", "meter": map[string]any{"selected_model": "old"}}}, true},
		{"refusal", true, []map[string]any{{"type": "error", "agent_id": "owned", "error": "busy"}}, true},
		{"legacy", false, []map[string]any{{"type": "settings_changed", "settings": map[string]any{"model": "new"}}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			commands := make(chan map[string]any, 1)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
				if err != nil {
					return
				}
				defer c.Close()
				var cmd map[string]any
				if c.ReadJSON(&cmd) != nil {
					return
				}
				commands <- cmd
				for _, frame := range tc.frames {
					if c.WriteJSON(frame) != nil {
						return
					}
				}
				// Keep the peer alive until the selector finishes, including its
				// unchanged deadline for wrong-model/Agent observations.
				c.ReadMessage()
			}))
			defer srv.Close()
			p, err := openGUIPeer("ws" + strings.TrimPrefix(srv.URL, "http"))
			if err != nil {
				t.Fatal(err)
			}
			defer p.close()
			err = selectStatusModel(p, "owned", "new", tc.live)
			if (err != nil) != tc.wantError {
				t.Fatalf("error %v wantError%v", err, tc.wantError)
			}
			want := map[string]any{"type": "set_model", "agent_id": "owned", "text": "new"}
			if !tc.live {
				want = map[string]any{"type": "update_settings", "settings": map[string]any{"model": "new"}}
			}
			if got := <-commands; !reflect.DeepEqual(got, want) {
				t.Fatalf("command %v want%v", got, want)
			}
		})
	}
}

func TestChapter18StructuredStatus(t *testing.T) {
	u := fakevendor.Canonical{Input: 100, Output: 7, CacheRead: 80, CacheWrite: 20}
	base := func() map[string]any {
		return map[string]any{
			"selected_model": "fixture-model", "last": map[string]any{"input": 100, "output": 7, "cache_read": 80, "cache_write": 20},
			"session":          map[string]any{"input": 100, "output": 7, "cache_read": 80, "cache_write": 20},
			"session_hit_rate": .4, "cost": map[string]any{"status": "known", "usd": .01234},
			"unrelated_diagnostic": "unknown diagnostic is not unknown pricing"}
	}
	for _, tc := range []struct {
		name                    string
		change                  func(map[string]any)
		known, reportOK, costOK bool
	}{
		{"known", nil, true, true, true},
		{"partial", func(m map[string]any) { object(m["cost"])["status"] = "partial" }, false, true, true},
		{"unknown", func(m map[string]any) { m["cost"] = map[string]any{"status": "unknown"} }, false, true, true},
		{"missing-model", func(m map[string]any) { delete(m, "selected_model") }, true, false, true},
		{"wrong-model", func(m map[string]any) { m["selected_model"] = "other" }, true, false, true},
		{"missing-count", func(m map[string]any) { delete(object(m["last"]), "cache_write") }, true, false, true},
		{"fractional-count", func(m map[string]any) { object(m["session"])["input"] = 100.0001 }, true, false, true},
		{"string-count", func(m map[string]any) { object(m["session"])["input"] = "100" }, true, false, true},
		{"missing-rate", func(m map[string]any) { delete(m, "session_hit_rate") }, true, false, true},
		{"wrong-denominator", func(m map[string]any) { m["session_hit_rate"] = 80.0 / 180 }, true, false, true},
		{"missing-cost", func(m map[string]any) { delete(m, "cost") }, true, true, false},
		{"null-usd", func(m map[string]any) { object(m["cost"])["usd"] = nil }, true, true, false},
		{"string-usd", func(m map[string]any) { object(m["cost"])["usd"] = ".01234" }, true, true, false},
		{"wrong-usd", func(m map[string]any) { object(m["cost"])["usd"] = .2 }, true, true, false},
		{"unpriced-claimed-known", nil, false, true, false},
		{"priced-claimed-partial", func(m map[string]any) { object(m["cost"])["status"] = "partial" }, true, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := base()
			if tc.change != nil {
				tc.change(m)
			}
			body, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			r := grade.Check{Passed: true}
			statusReport(&r, string(body), "fixture-model", u, u)
			if r.Passed != tc.reportOK {
				t.Fatalf("report %v", r.Details)
			}
			c := grade.Check{Passed: true}
			statusCost(&c, string(body), .01234, tc.known)
			if c.Passed != tc.costOK {
				t.Fatalf("cost %v", c.Details)
			}
		})
	}
	for _, text := range []string{`{"selected_model":`, `{"cost":{"status":"known","usd":1e999}}`, `{}`} {
		r := grade.Check{Passed: true}
		statusReport(&r, text, "fixture-model", u, u)
		c := grade.Check{Passed: true}
		statusCost(&c, text, .01234, true)
		if r.Passed || c.Passed {
			t.Fatalf("malformed/missing accepted: %s", text)
		}
	}
	// Keep the original representation accepted, not reserialized into JSON.
	text := "Model: fixture-model\nLast: 100 in / 7 out / 80 cache read / 20 cache write\nSession total: 100 in / 7 out / 80 cache read / 20 cache write\nCache hit rate: 40.0%\nCost: $0.01234"
	r := grade.Check{Passed: true}
	statusReport(&r, text, "fixture-model", u, u)
	c := grade.Check{Passed: true}
	statusCost(&c, text, .01234, true)
	if !r.Passed || !c.Passed {
		t.Fatalf("legacy report %v cost %v", r.Details, c.Details)
	}
}
