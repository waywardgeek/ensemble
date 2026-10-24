package engine

import (
	"encoding/json"
	"ensemble/internal/common"
	"strings"
	"testing"
)

func TestChapter15R1MarkerPositions(t *testing.T) {
	for _, vendor := range []common.Vendor{common.Anthropic, common.OpenAI} {
		cfg := common.Config{Vendor: vendor, Surface: common.Messages, Model: "claude-sonnet-4-6"}
		marker := "cache_control"
		want := 4
		if vendor == common.OpenAI {
			cfg.Surface = common.ChatCompletions
			cfg.Model = "gpt-5.6"
			marker = "prompt_cache_breakpoint"
			want = 3
		}
		c := common.Context{System: "SYSTEM manual", Handoff: "surviving handoff", Dialogue: []common.Entry{{Actor: common.Human, Parts: common.PartList{common.Text("old history")}}, {Actor: common.Assistant, Parts: common.PartList{common.Text("stable end")}}}, Ephemera: common.PartList{common.Text("transient")}}
		defs := []common.ToolDefinition{{Name: "tool", Parameters: json.RawMessage(`{"type":"object"}`)}}
		body, err := render(c, cfg, defs)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(body), `"`+marker+`"`) != want {
			t.Fatalf("%s markers: %s", vendor, body)
		}
		var wire map[string]json.RawMessage
		_ = json.Unmarshal(body, &wire)
		if vendor == common.OpenAI && !strings.Contains(string(wire["prompt_cache_options"]), `"explicit"`) {
			t.Fatalf("missing request mode: %s", body)
		}
		if strings.Contains(string(wire["tools"]), "prompt_cache_breakpoint") {
			t.Fatal("OpenAI tool marker")
		}
		var messages []struct{ Content json.RawMessage }
		_ = json.Unmarshal(wire["messages"], &messages)
		var marked []string
		for _, m := range messages {
			var blocks []map[string]json.RawMessage
			if json.Unmarshal(m.Content, &blocks) != nil {
				continue
			}
			for _, b := range blocks {
				if _, ok := b[marker]; ok {
					var text string
					_ = json.Unmarshal(b["text"], &text)
					marked = append(marked, text)
				}
			}
		}
		joined := strings.Join(marked, "|")
		if !strings.Contains(joined, "surviving handoff") || !strings.Contains(joined, "stable end") || strings.Contains(joined, "transient") || strings.Contains(joined, "old history") {
			t.Fatalf("wrong boundaries %q; %s", joined, body)
		}
		c.Handoff = ""
		body, _ = render(c, cfg, defs)
		if strings.Count(string(body), `"`+marker+`"`) != want-1 {
			t.Fatalf("missing handoff filled with older history: %s", body)
		}
		cfg.Model = "unknown-future"
		cfg.ResolvedModel = ""
		body, _ = render(c, cfg, defs)
		if strings.Contains(string(body), `"`+marker+`"`) || strings.Contains(string(body), "prompt_cache_options") {
			t.Fatalf("unknown model opted in: %s", body)
		}
	}
}

func TestChapter15R1LensByteSensitivity(t *testing.T) {
	c := NewCache(nil)
	cfg := common.Config{Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "gpt-5.6"}
	a := []byte(`{"model":"gpt-5.6","messages":[{"role":"user", "content":[{"type":"text","text":"literal prompt_cache_breakpoint and cache_control", "prompt_cache_breakpoint":{"mode":"explicit"}}]}]}`)
	b := []byte(`{"model":"gpt-5.6","messages":[{"role":"user", "content":[{"type":"text","text":"literal prompt_cache_breakpoint and cache_control" }]}]}`)
	c.Observe(cfg, a)
	c.Observe(cfg, b)
	if s := c.Status(); s.FirstDifference != "" {
		t.Fatalf("marker-only false edit: %+v", s)
	}
	changed := []byte(strings.Replace(string(b), `"role":"user", `, `"role":"user",  `, 1))
	c.Observe(cfg, changed)
	if c.Status().FirstDifference == "" {
		t.Fatal("canonical JSON hid whitespace change")
	}
}

func TestChapter15R1QualifiedOpenAIWrites(t *testing.T) {
	cfg := common.Config{Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "gpt-5.6"}
	body := []byte(`{"model":"gpt-5.6","choices":[{"message":{"content":"ok"}}],"usage":{"prompt_tokens":100,"completion_tokens":7,"prompt_tokens_details":{"cached_tokens":30,"cache_write_input_tokens":20}}}`)
	events, err := parse(200, body, cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range events {
		if e.Type == common.ResponseEnded && e.Response.Usage != (common.Usage{Input: 50, Output: 7, CacheRead: 30, CacheWrite: 20}) {
			t.Fatalf("wrong disjoint facts: %+v", e.Response.Usage)
		}
	}
}

func TestChapter15R1ExactMarkerStripping(t *testing.T) {
	cfg := common.Config{Vendor: common.Anthropic, Surface: common.Messages}
	body := []byte("{\n \"tools\":[{\"input_schema\":{\"cache_control\":{\"type\":\"ephemeral\"}}, \"cache_control\" : {\"type\":\"ephemeral\"} }],\n \"messages\":[{\"role\":\"assistant\",\"content\":[{\"type\":\"thinking\",\"cache_control\":\"protected-native-property\"},{\"type\":\"tool_use\",\"input\":{\"prompt_cache_breakpoint\":{\"mode\":\"explicit\"}},\"cache_control\":{\"type\":\"ephemeral\"}},{\"type\":\"text\",\"text\":\"literal \\\"cache_control\\\":{\\\"type\\\":\\\"ephemeral\\\"}\"}]}]\n}")
	want := []byte("{\n \"tools\":[{\"input_schema\":{\"cache_control\":{\"type\":\"ephemeral\"}}  }],\n \"messages\":[{\"role\":\"assistant\",\"content\":[{\"type\":\"thinking\",\"cache_control\":\"protected-native-property\"},{\"type\":\"tool_use\",\"input\":{\"prompt_cache_breakpoint\":{\"mode\":\"explicit\"}}},{\"type\":\"text\",\"text\":\"literal \\\"cache_control\\\":{\\\"type\\\":\\\"ephemeral\\\"}\"}]}]\n}")
	if got := stripCacheMarkers(cfg, body); string(got) != string(want) {
		t.Fatalf("non-marker bytes changed:\ngot %s\nwant %s", got, want)
	}
}

func TestChapter15R1MovingMarkerStablePrefix(t *testing.T) {
	for _, cfg := range []common.Config{{Vendor: common.Anthropic, Surface: common.Messages, Model: "claude-sonnet-4-6"}, {Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "gpt-5.6"}} {
		c := common.Context{System: "manual", Dialogue: []common.Entry{{Actor: common.Human, Parts: common.PartList{common.Text("old")}}}}
		first, err := render(c, cfg, nil)
		if err != nil {
			t.Fatal(err)
		}
		c.Dialogue = append(c.Dialogue, common.Entry{Actor: common.Assistant, Parts: common.PartList{common.Text("new")}})
		second, err := render(c, cfg, nil)
		if err != nil {
			t.Fatal(err)
		}
		lens := NewCache(nil)
		lens.Observe(cfg, first)
		lens.Observe(cfg, second)
		found := false
		for _, s := range lens.Status().Sections {
			if s.Name == "messages" {
				found = true
				if s.Change != "appended" {
					t.Fatalf("%s moving marker changed old bytes: %+v\n%s\n%s", cfg.Vendor, lens.Status(), first, second)
				}
			}
		}
		if !found {
			t.Fatal("no messages section")
		}
	}
}

func TestChapter15R1PromptOptionsAndRoute(t *testing.T) {
	cfg := common.Config{Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "gpt-5.6", BaseURL: "https://example.invalid/v1"}
	a := []byte(`{"model":"gpt-5.6","tools":[{"function":{"name":"tool"}}],"messages":[{"role":"system","content":"manual"},{"role":"user","content":"literal cache_control"}],"stream":false}`)
	b := []byte(strings.Replace(string(a), `"stream":false`, `"stream":true`, 1))
	lens := NewCache(nil)
	first := lens.Observe(cfg, a)
	if first.Fraction != nil {
		t.Fatal("first send prediction invented")
	}
	p := lens.Observe(cfg, b)
	s := lens.Status()
	if p.Fraction == nil || *p.Fraction != 1 || s.OptionsChange != "changed" || s.FirstDifference != "" {
		t.Fatalf("options contaminated prompt %+v", s)
	}
	if len(s.Sections) != 3 || s.Sections[0].Name != "tools" || s.Sections[1].Name != "system" || s.Sections[2].Name != "messages" {
		t.Fatalf("cache order %+v", s.Sections)
	}
	for _, sec := range s.Sections {
		if sec.Change != "unchanged" {
			t.Fatalf("options changed section %+v", sec)
		}
	}
	cfg.BaseURL = "https://other.invalid/v1"
	p = lens.Observe(cfg, b)
	if p.Fraction != nil || !strings.Contains(p.Reason, "reset") {
		t.Fatalf("cross-route prediction %+v", p)
	}
	cfg.Model = "other"
	b = []byte(strings.Replace(string(b), "gpt-5.6", "other", 1))
	p = lens.Observe(cfg, b)
	if p.Fraction != nil {
		t.Fatal("cross-model prediction")
	}
}

func TestChapter15R1WriteDependenciesAndSSE(t *testing.T) {
	cfg := common.Config{Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "gpt-5.6"}
	for _, tc := range []struct {
		name, details string
		want          common.Usage
		bad           bool
	}{
		{"reported", `{"cached_tokens":30,"cache_write_input_tokens":20}`, common.Usage{Input: 50, Output: 7, CacheRead: 30, CacheWrite: 20}, false},
		{"zero", `{"cached_tokens":0,"cache_write_input_tokens":0}`, common.Usage{Input: 100, Output: 7}, false},
		{"missing-write", `{"cached_tokens":30}`, common.Usage{Input: 70, Output: 7, CacheRead: 30}, false},
		{"missing-read", `{"cache_write_input_tokens":20}`, common.Usage{Input: 80, Output: 7, CacheWrite: 20}, false},
		{"null-write", `{"cached_tokens":30,"cache_write_input_tokens":null}`, common.Usage{Input: -1, Output: 7, CacheRead: 30, CacheWrite: -1}, true},
		{"negative-read", `{"cached_tokens":-1,"cache_write_input_tokens":20}`, common.Usage{Input: -1, Output: 7, CacheRead: -1, CacheWrite: 20}, true},
		{"fraction-write", `{"cached_tokens":30,"cache_write_input_tokens":2.5}`, common.Usage{Input: -1, Output: 7, CacheRead: 30, CacheWrite: -1}, true},
		{"inconsistent", `{"cached_tokens":90,"cache_write_input_tokens":20}`, common.Usage{Input: -1, Output: 7, CacheRead: 90, CacheWrite: 20}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var env map[string]json.RawMessage
			_ = json.Unmarshal([]byte(`{"usage":{"prompt_tokens":100,"completion_tokens":7,"prompt_tokens_details":`+tc.details+`}}`), &env)
			got, _, err := parseUsage(cfg, env)
			if got != tc.want || (err != nil) != tc.bad {
				t.Fatalf("%+v %v; want %+v bad=%v", got, err, tc.want, tc.bad)
			}
		})
	}
	final := `{"prompt_tokens":9007199254741000,"completion_tokens":7,"prompt_tokens_details":{"cached_tokens":30,"cache_write_input_tokens":20}}`
	plain, err := parse(200, []byte(`{"model":"gpt-5.6","choices":[{"message":{"content":"ok"}}],"usage":`+final+`}`), cfg)
	if err != nil {
		t.Fatal(err)
	}
	stream := "data: {\"model\":\"gpt-5.6\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"ok\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":50,\"completion_tokens\":2,\"prompt_tokens_details\":{\"cached_tokens\":10,\"cache_write_input_tokens\":1}}}\n\ndata: {\"choices\":[],\"usage\":" + final + "}\n\ndata: [DONE]\n\n"
	streamed, err := parseSSE(strings.NewReader(stream), cfg, common.ParseCallbacks{})
	if err != nil {
		t.Fatal(err)
	}
	usage := func(events []common.Event) common.Usage {
		for _, e := range events {
			if e.Type == common.ResponseEnded {
				return e.Response.Usage
			}
		}
		t.Fatal("missing response")
		return common.Usage{}
	}
	want := common.Usage{Input: 9007199254740950, Output: 7, CacheRead: 30, CacheWrite: 20}
	if usage(plain) != want || usage(streamed) != want {
		t.Fatalf("plain %+v SSE %+v", usage(plain), usage(streamed))
	}
}

func TestChapter15R1ActualHandoffAndFallbackPositions(t *testing.T) {
	for _, cfg := range []common.Config{{Vendor: common.Anthropic, Surface: common.Messages, Model: "claude-sonnet-4-6", ResolvedModel: "claude-sonnet-4-6"}, {Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "gpt-5.6", ResolvedModel: "gpt-5.6"}} {
		c := common.Context{System: "manual stays in SYSTEM", Handoff: "typed surviving handoff", Ephemera: common.PartList{common.Text("transient tail")}, Dialogue: []common.Entry{{Actor: common.Human, Parts: common.PartList{common.Text("older history")}}, {Actor: common.Assistant, Parts: common.PartList{common.Text("last retained history")}}}}
		if cfg.Vendor == common.Anthropic {
			c.Dialogue = append(c.Dialogue, common.Entry{Actor: common.Assistant, Parts: common.PartList{{Type: "opaque", From: &common.Provenance{Vendor: cfg.Vendor, Surface: cfg.Surface, Model: cfg.Model}, Data: json.RawMessage(" \n {\"type\":\"thinking\",\"thinking\":\"protected\",\"signature\":\"signed\"}")}}})
		}
		b, err := render(c, cfg, nil)
		if err != nil {
			t.Fatal(err)
		}
		var request struct {
			Messages []struct{ Content []map[string]json.RawMessage }
		}
		if err = json.Unmarshal(b, &request); err != nil {
			t.Fatal(err)
		}
		seen := map[string]bool{}
		marked := map[string]bool{}
		for _, m := range request.Messages {
			for _, part := range m.Content {
				var text, kind string
				_ = json.Unmarshal(part["text"], &text)
				_ = json.Unmarshal(part["type"], &kind)
				if text != "" {
					seen[text] = true
					marked[text] = len(part["cache_control"])+len(part["prompt_cache_breakpoint"]) > 0
				}
				if kind == "thinking" {
					if string(part["signature"]) != `"signed"` || len(part["cache_control"]) != 0 {
						t.Fatalf("protected native block modified: %s", b)
					}
				}
			}
		}
		for text, want := range map[string]bool{"older history": false, "last retained history": true, "typed surviving handoff": true, "transient tail": false} {
			if !seen[text] || marked[text] != want {
				t.Fatalf("%s boundary %q marker=%v want=%v: %s", cfg.Vendor, text, marked[text], want, b)
			}
		}
	}
	for _, cfg := range []common.Config{{Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "gpt-5.4"}, {Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "unknown"}, {Vendor: common.OpenAI, Surface: common.Messages, Model: "gpt-5.6"}, {Vendor: common.Anthropic, Surface: common.Messages, Model: "unknown"}} {
		b, err := render(common.Context{System: "manual", Handoff: "handoff", Dialogue: []common.Entry{{Actor: common.Human, Parts: common.PartList{common.Text("prompt")}}}}, cfg, nil)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(b), `"prompt_cache_options"`) || strings.Contains(string(b), `"prompt_cache_breakpoint"`) || strings.Contains(string(b), `"cache_control"`) {
			t.Fatalf("unsupported route opted in: %+v %s", cfg, b)
		}
	}
	if markOpenAIMessage(&oMessage{Role: "user", Content: []any{map[string]any{"type": "image_url"}, map[string]any{"type": "text", "text": nil}}}) {
		t.Fatal("illegal content target marked")
	}
	if _, known := common.LookupPrices(common.OpenAI, common.ChatCompletions, "gpt-5.6"); known {
		t.Fatal("invented GPT-5.6 prices")
	}
}
