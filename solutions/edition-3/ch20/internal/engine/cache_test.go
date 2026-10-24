package engine

import (
	"encoding/json"
	"ensemble/internal/common"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestChapter15Usage(t *testing.T) {
	for _, tc := range []struct {
		name string
		v    common.Vendor
		wire string
		want common.Usage
		bad  bool
	}{
		{"anthropic", common.Anthropic, `{"usage":{"input_tokens":10,"output_tokens":4,"cache_read_input_tokens":8,"cache_creation_input_tokens":2}}`, common.Usage{Input: 10, Output: 4, CacheRead: 8, CacheWrite: 2}, false},
		{"anthropic-absent", common.Anthropic, `{"usage":{"input_tokens":10,"output_tokens":4}}`, common.Usage{Input: 10, Output: 4}, false},
		{"anthropic-zero", common.Anthropic, `{"usage":{"input_tokens":0,"output_tokens":0,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}`, common.Usage{}, false},
		{"openai", common.OpenAI, `{"usage":{"prompt_tokens":20,"completion_tokens":4,"prompt_tokens_details":{"cached_tokens":8,"cache_write_tokens":999}}}`, common.Usage{Input: 12, Output: 4, CacheRead: 8}, false},
		{"openai-absent", common.OpenAI, `{"usage":{"prompt_tokens":20,"completion_tokens":4}}`, common.Usage{Input: 20, Output: 4}, false},
		{"openai-zero", common.OpenAI, `{"usage":{"prompt_tokens":20,"completion_tokens":4,"prompt_tokens_details":{"cached_tokens":0}}}`, common.Usage{Input: 20, Output: 4}, false},
		{"gemini", common.Gemini, `{"usageMetadata":{"promptTokenCount":20,"candidatesTokenCount":4,"thoughtsTokenCount":2,"cachedContentTokenCount":8}}`, common.Usage{Input: 12, Output: 6, CacheRead: 8}, false},
		{"gemini-absent", common.Gemini, `{"usageMetadata":{"promptTokenCount":20,"candidatesTokenCount":4}}`, common.Usage{Input: 20, Output: 4}, false},
		{"invalid-cache", common.OpenAI, `{"usage":{"prompt_tokens":20,"completion_tokens":4,"prompt_tokens_details":{"cached_tokens":null}}}`, common.Usage{Input: -1, Output: 4, CacheRead: -1}, true},
		{"inconsistent", common.Gemini, `{"usageMetadata":{"promptTokenCount":2,"candidatesTokenCount":4,"cachedContentTokenCount":8}}`, common.Usage{Input: -1, Output: 4, CacheRead: 8}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var env map[string]json.RawMessage
			_ = json.Unmarshal([]byte(tc.wire), &env)
			got, _, err := parseUsage(common.Config{Vendor: tc.v}, env)
			if got != tc.want || (err != nil) != tc.bad {
				t.Fatalf("got %+v err %v; want %+v bad %v", got, err, tc.want, tc.bad)
			}
		})
	}
}

func TestChapter15LensBoundedAndDetached(t *testing.T) {
	c := NewCache(nil).(*cache)
	cfg := common.Config{Vendor: common.OpenAI, Surface: common.ChatCompletions}
	a := []byte(`{"model":"fixture","messages":[{"role":"user","content":"one"}],"stream":true}`)
	b := []byte(`{"model":"fixture","messages":[{"role":"user","content":"one"},{"role":"assistant","content":"two"}],"stream":true}`)
	c.Observe(cfg, a)
	c.Observe(cfg, b)
	s := c.Status()
	if s.Previous.Digest != digest(a) || s.Current.Digest != digest(b) || s.SharedSegments != 1 || !strings.HasPrefix(s.FirstDifference, "messages byte ") {
		t.Fatalf("%+v", s)
	}
	s.Current.Segments[0].Path = "mutated"
	if c.Status().Current.Segments[0].Path == "mutated" {
		t.Fatal("snapshot aliases owner")
	}
	c.Observe(cfg, a)
	if string(c.previous) != string(b) || string(c.current) != string(a) || c.Status().Previous.Digest != digest(b) {
		t.Fatal("wrong two-request window")
	}
	c.Observe(cfg, a)
	if c.Status().FirstDifference != "" {
		t.Fatal("identical request drifted")
	}
}

func TestChapter15LegalMarkersPreserveContent(t *testing.T) {
	cfg := common.Config{Vendor: common.Anthropic, Surface: common.Messages, Model: "claude-sonnet-4-6"}
	from := common.Provenance{Vendor: cfg.Vendor, Surface: cfg.Surface, Model: cfg.Model}
	c := common.Context{System: "stable system", ActiveSkills: []common.SkillData{{Name: "manual", Body: "MANUAL"}}, Dialogue: []common.Entry{
		{Actor: common.Human, Parts: common.PartList{common.Text("first")}},
		{Actor: common.Assistant, Parts: common.PartList{{Type: "opaque", Data: json.RawMessage(`{"type":"thinking","thinking":"secret","signature":"sig"}`), From: &from}, common.Text("answer")}},
		{Actor: common.Human, Parts: common.PartList{common.Text("second")}},
	}, Ephemera: common.PartList{common.Text("volatile")}}
	defs := []common.ToolDefinition{{Name: "tool", Description: "exact", Parameters: json.RawMessage(`{"type":"object"}`)}}
	marked, err := render(c, cfg, defs)
	if err != nil {
		t.Fatal(err)
	}
	unmarkedCfg := cfg
	unmarkedCfg.Surface = common.ChatCompletions // same renderer, unsupported API gate
	unmarked, err := render(c, unmarkedCfg, defs)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(marked), `"cache_control"`) != 3 || strings.Contains(string(unmarked), "cache_control") {
		t.Fatalf("markers: %s / %s", marked, unmarked)
	}
	// Compare the exact text/order after normalizing only legal string/block forms
	// and removing cache metadata. Native thinking data is left intact.
	normalize := func(body []byte) any {
		var value map[string]any
		_ = json.Unmarshal(body, &value)
		var strip func(any) any
		strip = func(v any) any {
			switch x := v.(type) {
			case map[string]any:
				delete(x, "cache_control")
				for k, z := range x {
					x[k] = strip(z)
				}
			case []any:
				for i, z := range x {
					x[i] = strip(z)
				}
			}
			return v
		}
		strip(value)
		text := func(v any) any {
			blocks, ok := v.([]any)
			if !ok {
				return v
			}
			var out strings.Builder
			for _, b := range blocks {
				m, ok := b.(map[string]any)
				if !ok || m["type"] != "text" {
					return v
				}
				out.WriteString(m["text"].(string))
			}
			return out.String()
		}
		value["system"] = text(value["system"])
		for _, m := range value["messages"].([]any) {
			msg := m.(map[string]any)
			msg["content"] = text(msg["content"])
		}
		return value
	}
	// Native replay is surface-bound, so compare with the same provenance surface.
	c.Dialogue[1].Parts = c.Dialogue[1].Parts[1:]
	marked, _ = render(c, cfg, defs)
	unmarked, _ = render(c, unmarkedCfg, defs)
	if !reflect.DeepEqual(normalize(marked), normalize(unmarked)) {
		t.Fatalf("content changed: %s / %s", marked, unmarked)
	}
	for _, model := range []string{"unknown", "claude-sonnet-5-course", "claude-opus-5-course"} {
		cfg.Model = model
		b, _ := render(c, cfg, defs)
		if strings.Contains(string(b), "cache_control") {
			t.Fatalf("unsupported %s marked", model)
		}
	}
}

func TestChapter15CostsKeepCapturedIdentity(t *testing.T) {
	cfg := common.Config{Vendor: common.Anthropic, Surface: common.Messages, Model: "route"}
	e, err := New(parseAgent{config: cfg})
	if err != nil {
		t.Fatal(err)
	}
	first := common.ResponseData{From: common.Provenance{Model: "claude-sonnet-4-6"}, Usage: common.Usage{Input: 100, Output: 10, CacheRead: 80, CacheWrite: 20}}
	e.account(cfg, &first)
	s := e.Status()
	want := (100*3 + 10*15 + 80*.3 + 20*3.75) / 1e6
	if s.Cost.Status != "known" || math.Abs(s.Cost.USD-want) > 1e-15 || s.LastHitRate == nil || *s.LastHitRate != .4 {
		t.Fatalf("%+v want %g", s, want)
	}
	cfg.Model = "claude-sonnet-4-6"
	e.account(cfg, &common.ResponseData{From: common.Provenance{Model: "unknown-served"}, Usage: common.Usage{Input: 10, Output: 2, CacheRead: -1, CacheWrite: -1}})
	s = e.Status()
	if math.Abs(s.Cost.USD-want) > 1e-15 || s.Cost.Status != "partial" || len(s.Buckets) != 2 || s.Buckets[0].Model != first.From.Model || s.SessionHitRate != nil {
		t.Fatalf("repriced or hid unknown: %+v", s)
	}
	s.Buckets[0].Model = "mutated"
	if e.Status().Buckets[0].Model == "mutated" {
		t.Fatal("bucket aliases owner")
	}
}

func TestChapter15WeightedRatesAndUnknownPrices(t *testing.T) {
	cfg := common.Config{Vendor: common.Anthropic, Surface: common.Messages, Model: "claude-sonnet-4-6"}
	e, err := New(parseAgent{config: cfg})
	if err != nil {
		t.Fatal(err)
	}
	e.account(cfg, &common.ResponseData{Usage: common.Usage{Input: 20, CacheRead: 80}})
	e.account(cfg, &common.ResponseData{Usage: common.Usage{Input: 300}})
	s := e.Status()
	if s.LastHitRate == nil || *s.LastHitRate != 0 || s.SessionHitRate == nil || *s.SessionHitRate != .2 {
		t.Fatalf("not token-weighted: %+v", s)
	}
	if _, ok := common.LookupPrices(common.Anthropic, common.ChatCompletions, cfg.Model); ok {
		t.Fatal("price crossed API surface")
	}
	unknownCfg := common.Config{Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "unpriced"}
	unknown, err := New(parseAgent{config: unknownCfg})
	if err != nil {
		t.Fatal(err)
	}
	unknown.account(unknownCfg, &common.ResponseData{Usage: common.Usage{Input: 10, Output: 2}})
	if unknown.Status().Cost.Status != "unknown" {
		t.Fatal("invented price")
	}
}

func TestChapter18ToolCacheBreakpointEligibility(t *testing.T) {
	cfg := common.Config{Vendor: common.Anthropic, Surface: common.Messages, Model: "claude-sonnet-4-6"}
	c := common.Context{System: "stable system", Dialogue: []common.Entry{{Actor: common.Human, Parts: common.PartList{common.Text("question")}}}}
	for _, tc := range []struct {
		name    string
		enabled []bool
		want    int
	}{
		{"last eligible before deferred tail", []bool{true, true, false}, 1},
		{"eligible final entry", []bool{true, false, true}, 2},
		{"all deferred", []bool{false, false}, -1},
		{"empty catalog", []bool{}, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			catalog := []common.ToolSpec{}
			for i, enabled := range tc.enabled {
				catalog = append(catalog, common.ToolSpec{ToolDefinition: common.ToolDefinition{Name: []string{"first", "second", "third"}[i], Description: "unchanged", Parameters: json.RawMessage(`{"type":"object","properties":{}}`)}, Enabled: enabled})
			}
			before, _ := json.Marshal(catalog)
			body, err := render(c, cfg, nil, anthropicDisclosure{catalog: catalog})
			if err != nil {
				t.Fatal(err)
			}
			var wire struct{ Tools []aDefinition }
			if err := json.Unmarshal(body, &wire); err != nil {
				t.Fatal(err)
			}
			if len(wire.Tools) != len(catalog) {
				t.Fatalf("catalog length changed: %s", body)
			}
			markers := 0
			for i, tool := range wire.Tools {
				if tool.Name != catalog[i].Name || tool.Description != catalog[i].Description || string(tool.InputSchema) != string(catalog[i].Parameters) || tool.DeferLoading == tc.enabled[i] {
					t.Fatalf("catalog changed: %s", body)
				}
				if tool.CacheControl != nil {
					markers++
					if i != tc.want || tool.DeferLoading {
						t.Fatalf("illegal or misplaced tool marker: %s", body)
					}
				}
			}
			wantMarkers := 0
			if tc.want >= 0 {
				wantMarkers = 1
			}
			if markers != wantMarkers || strings.Count(string(body), `"cache_control"`) != 2+wantMarkers {
				t.Fatalf("system/history/tool marker budget changed: %s", body)
			}
			after, _ := json.Marshal(catalog)
			if string(after) != string(before) {
				t.Fatal("render mutated catalog")
			}
		})
	}
}
