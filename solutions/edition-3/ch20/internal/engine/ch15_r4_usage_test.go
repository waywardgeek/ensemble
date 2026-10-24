package engine

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"ensemble/internal/common"
)

func TestChapter15R4OptionalUsagePlainAndStream(t *testing.T) {
	type fixture struct {
		name   string
		vendor common.Vendor
		usage  string
		want   common.Usage
		bad    bool
	}
	cases := []fixture{
		{"anthropic/omitted-cache", common.Anthropic, `{"input_tokens":100,"output_tokens":7}`, common.Usage{Input: 100, Output: 7}, false},
		{"openai/omitted-details", common.OpenAI, `{"prompt_tokens":100,"completion_tokens":7}`, common.Usage{Input: 100, Output: 7}, false},
		{"openai/empty-details", common.OpenAI, `{"prompt_tokens":100,"completion_tokens":7,"prompt_tokens_details":{}}`, common.Usage{Input: 100, Output: 7}, false},
		{"openai/qualified-write", common.OpenAI, `{"prompt_tokens":100,"completion_tokens":7,"prompt_tokens_details":{"cached_tokens":3,"cache_write_input_tokens":2}}`, common.Usage{Input: 95, Output: 7, CacheRead: 3, CacheWrite: 2}, false},
		{"openai/inconsistent", common.OpenAI, `{"prompt_tokens":4,"completion_tokens":7,"prompt_tokens_details":{"cached_tokens":3,"cache_write_input_tokens":2}}`, common.Usage{Input: -1, Output: 7, CacheRead: 3, CacheWrite: 2}, true},
		{"gemini/omitted-cache", common.Gemini, `{"promptTokenCount":100,"candidatesTokenCount":7}`, common.Usage{Input: 100, Output: 7}, false},
		{"anthropic/missing-output", common.Anthropic, `{"input_tokens":100}`, common.Usage{Input: 100, Output: -1}, true},
		{"openai/missing-input", common.OpenAI, `{"completion_tokens":7,"prompt_tokens_details":{"cached_tokens":3,"cache_write_input_tokens":2}}`, common.Usage{Input: -1, Output: 7, CacheRead: 3, CacheWrite: 2}, true},
		{"gemini/missing-output", common.Gemini, `{"promptTokenCount":100}`, common.Usage{Input: 100, Output: -1}, true},
	}
	for _, vendor := range []common.Vendor{common.Anthropic, common.OpenAI, common.Gemini} {
		for _, whole := range []struct{ name, raw string }{{"missing-whole", ""}, {"null-whole", "null"}, {"invalid-whole", "[]"}} {
			cases = append(cases, fixture{vendor.String() + "/" + whole.name, vendor, whole.raw, common.UnknownUsage(), true})
		}
		cases = append(cases, fixture{vendor.String() + "/empty-usage", vendor, `{}`, common.Usage{Input: -1, Output: -1}, true})
	}
	for _, bad := range []string{`null`, `-1`, `"bad"`, `1.5`} {
		cases = append(cases,
			fixture{"anthropic/invalid-read/" + bad, common.Anthropic, `{"input_tokens":100,"output_tokens":7,"cache_read_input_tokens":` + bad + `}`, common.Usage{Input: 100, Output: 7, CacheRead: -1}, true},
			fixture{"openai/invalid-write/" + bad, common.OpenAI, `{"prompt_tokens":100,"completion_tokens":7,"prompt_tokens_details":{"cached_tokens":3,"cache_write_input_tokens":` + bad + `}}`, common.Usage{Input: -1, Output: 7, CacheRead: 3, CacheWrite: -1}, true},
			fixture{"gemini/invalid-read/" + bad, common.Gemini, `{"promptTokenCount":100,"candidatesTokenCount":7,"cachedContentTokenCount":` + bad + `}`, common.Usage{Input: -1, Output: 7, CacheRead: -1}, true},
		)
	}
	for _, bad := range []string{`null`, `[]`, `7`, `"bad"`} {
		cases = append(cases, fixture{"openai/invalid-details/" + bad, common.OpenAI, `{"prompt_tokens":100,"completion_tokens":7,"prompt_tokens_details":` + bad + `}`, common.Usage{Input: -1, Output: 7, CacheRead: -1, CacheWrite: -1}, true})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			surface := common.ChatCompletions
			switch tc.vendor {
			case common.Anthropic:
				surface = common.Messages
			case common.Gemini:
				surface = common.GenerateContent
			}
			cfg := common.Config{Vendor: tc.vendor, Surface: surface}
			if tc.vendor == common.OpenAI {
				cfg.Model = "gpt-5.6"
			}
			key := "usage"
			if tc.vendor == common.Gemini {
				key = "usageMetadata"
			}
			usage := ""
			if tc.usage != "" {
				usage = `,"` + key + `":` + tc.usage
			}
			var plain, wire string
			switch tc.vendor {
			case common.Anthropic:
				plain = `{"content":[{"type":"text","text":"answer"}]` + usage + `}`
				wire = "data: {\"type\":\"message_start\",\"message\":{\"model\":\"fixture\"" + usage + "}}\n\n" +
					"data: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"answer\"}}\n\n" +
					"data: {\"type\":\"content_block_stop\",\"index\":0}\n\n" +
					"data: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"}}\n\n" +
					"data: {\"type\":\"message_stop\"}\n\n"
			case common.OpenAI:
				plain = `{"choices":[{"message":{"content":"answer"},"finish_reason":"stop"}]` + usage + `}`
				wire = "data: {\"choices\":[{\"delta\":{\"content\":\"answer\"},\"finish_reason\":\"stop\"}]" + usage + "}\n\ndata: [DONE]\n\n"
			case common.Gemini:
				plain = `{"candidates":[{"content":{"parts":[{"text":"answer"}]},"finishReason":"STOP"}]` + usage + `}`
				wire = "data: " + plain + "\n\n"
			}
			for _, stream := range []bool{false, true} {
				var events []common.Event
				var err error
				if stream {
					events, err = parseSSE(strings.NewReader(wire), cfg, common.ParseCallbacks{})
				} else {
					events, err = parse(200, []byte(plain), cfg)
				}
				got := ended(t, events)
				if got.Usage != tc.want || (err != nil) != tc.bad {
					t.Fatalf("stream=%t usage=%+v error=%v want=%+v bad=%t", stream, got.Usage, err, tc.want, tc.bad)
				}
				missing := map[string]bool{}
				for _, key := range got.UsageMissing {
					missing[key] = true
				}
				fields := map[string]int64{"input": tc.want.Input, "output": tc.want.Output, "cache_read": tc.want.CacheRead, "cache_write": tc.want.CacheWrite}
				for name, count := range fields {
					if missing[name] != (count < 0) {
						t.Fatalf("stream=%t missing=%v count %s=%d", stream, got.UsageMissing, name, count)
					}
				}
				if !tc.bad && len(got.UsageMissing) != 0 {
					t.Fatal("omitted optional classified as unmeasured")
				}
			}
		})
	}
}

func TestChapter15R4WholeUsageRecoveryInStream(t *testing.T) {
	cfg := common.Config{Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "gpt-5.6"}
	// Null on an interim chunk must not poison a later valid cumulative report;
	// a supplied invalid final report must not disappear behind older counts.
	valid := `{"prompt_tokens":10,"completion_tokens":2}`
	for _, tc := range []struct {
		first, last string
		want        common.Usage
		bad         bool
	}{
		{`null`, valid, common.Usage{Input: 10, Output: 2}, false},
		{valid, `null`, common.UnknownUsage(), true},
		{valid, `[]`, common.UnknownUsage(), true},
	} {
		wire := fmt.Sprintf("data: {\"choices\":[{\"delta\":{\"content\":\"answer\"}}],\"usage\":%s}\n\ndata: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}],\"usage\":%s}\n\ndata: [DONE]\n\n", tc.first, tc.last)
		events, err := parseSSE(strings.NewReader(wire), cfg, common.ParseCallbacks{})
		if ended(t, events).Usage != tc.want || (err != nil) != tc.bad {
			t.Fatalf("last=%s usage=%+v err=%v", tc.last, ended(t, events).Usage, err)
		}
	}
	// Unsupported implicit models retain zero writes only when the write field
	// is absent; a supplied qualified field is not interpreted on that route.
	cfg.Model = "gpt-5.4"
	for _, present := range []bool{false, true} {
		details := `{"cached_tokens":3}`
		want := common.Usage{Input: 7, Output: 2, CacheRead: 3}
		if present {
			details = `{"cached_tokens":3,"cache_write_input_tokens":2}`
			want.Input = -1
			want.CacheWrite = -1
		}
		raw := json.RawMessage(`{"prompt_tokens":10,"completion_tokens":2,"prompt_tokens_details":` + details + `}`)
		got, _, err := parseUsage(cfg, map[string]json.RawMessage{"usage": raw})
		if got != want || err != nil {
			t.Fatalf("present=%t usage=%+v err=%v", present, got, err)
		}
	}
}
