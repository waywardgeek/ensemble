package engine

import (
	"encoding/json"
	"ensemble/internal/common"
	"strings"
	"testing"
)

func geminiFrames(parts ...string) string {
	var b strings.Builder
	for i, p := range parts {
		finish := ""
		if i == len(parts)-1 {
			finish = `,"finishReason":"STOP"`
		}
		b.WriteString(`data: {"modelVersion":"actual","candidates":[{"content":{"parts":` + p + `}` + finish + `}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1}}` + "\n\n")
	}
	return b.String()
}
func TestCh6R1GeminiNativeParts(t *testing.T) {
	cases := []struct {
		name   string
		frames []string
		final  string
	}{
		{"explicit text parts", []string{`[{"text":"a"},{"text":"b"}]`, `[{"text":"c"},{"text":"d"}]`}, `[{"text":"a"},{"text":"bc"},{"text":"d"}]`},
		{"signed closes text", []string{`[{"text":"a","thoughtSignature":"sig-a"}]`, `[{"text":"b"}]`}, `[{"text":"a","thoughtSignature":"sig-a"},{"text":"b"}]`},
		{"incoming signed is explicit", []string{`[{"text":"a"}]`, `[{"text":"b","thoughtSignature":"sig-b"}]`}, `[{"text":"a"},{"text":"b","thoughtSignature":"sig-b"}]`},
		{"trailing signature exact binding", []string{`[{"text":"a"},{"text":"b"}]`, `[{"thoughtSignature":"sig-b"}]`, `[{"text":"c"}]`}, `[{"text":"a"},{"text":"b","thoughtSignature":"sig-b"},{"text":"c"}]`},
		{"no signature overwrite", []string{`[{"text":"a","thoughtSignature":"sig-a"}]`, `[{"thoughtSignature":"standalone"}]`}, `[{"text":"a","thoughtSignature":"sig-a"},{"thoughtSignature":"standalone"}]`},
		{"standalone opaque", []string{`[{"thoughtSignature":"standalone"},{"provider_extension":{"n":3}}]`}, `[{"thoughtSignature":"standalone"},{"provider_extension":{"n":3}}]`},
		{"thought classification first", []string{`[{"thought":true,"functionCall":{"name":"not-a-call","args":{}},"text":"hidden","thoughtSignature":"sig"}]`}, `[{"thought":true,"functionCall":{"name":"not-a-call","args":{}},"text":"hidden","thoughtSignature":"sig"}]`},
		{"thought boundaries", []string{`[{"thought":true,"text":"a"},{"thought":true,"text":"b"}]`, `[{"thought":true,"text":"c","thoughtSignature":"sig-c"}]`}, `[{"thought":true,"text":"a"},{"thought":true,"text":"b"},{"thought":true,"text":"c","thoughtSignature":"sig-c"}]`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := parser(common.Gemini, common.GenerateContent)
			var deltas []common.PartDelta
			stream, err := parseBody(e, geminiFrames(tc.frames...), "text/event-stream", common.ParseCallbacks{OnDelta: func(d common.PartDelta) { deltas = append(deltas, d) }})
			if err != nil {
				t.Fatal(err)
			}
			plainBody := `{"modelVersion":"actual","candidates":[{"content":{"parts":` + tc.final + `},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1}}`
			plain, err := parseBody(e, plainBody, "application/json", common.ParseCallbacks{})
			if err != nil {
				t.Fatal(err)
			}
			got, want := ended(t, stream), ended(t, plain)
			a, _ := json.Marshal(got)
			b, _ := json.Marshal(want)
			var av, bv any
			json.Unmarshal(a, &av)
			json.Unmarshal(b, &bv)
			a, _ = json.Marshal(av)
			b, _ = json.Marshal(bv)
			if string(a) != string(b) {
				t.Fatalf("stream %s\nplain  %s", a, b)
			}
			for _, d := range deltas {
				if d.PartID >= uint64(len(got.Parts)) {
					t.Fatalf("unbound delta %+v", d)
				}
			}
			if tc.name == "explicit text parts" {
				wantIDs := []uint64{0, 1, 1, 2}
				if len(deltas) != len(wantIDs) {
					t.Fatal(deltas)
				}
				for i, id := range wantIDs {
					if deltas[i].PartID != id {
						t.Fatal(deltas)
					}
				}
			}
			if tc.name == "thought classification first" && got.Parts[0].Type != "opaque" {
				t.Fatal(got.Parts)
			}
			context := common.Context{Dialogue: []common.Entry{{Actor: common.Assistant, Parts: got.Parts}}}
			cfg := common.Config{Vendor: common.Gemini, Surface: common.GenerateContent, Model: "route", ResolvedModel: "actual"}
			same, err := renderGemini(context, cfg, nil)
			if err != nil {
				t.Fatal(err)
			}
			var rendered struct {
				Contents []struct {
					Parts []json.RawMessage `json:"parts"`
				} `json:"contents"`
			}
			if err := json.Unmarshal(same, &rendered); err != nil {
				t.Fatal(err)
			}
			var expected []json.RawMessage
			json.Unmarshal([]byte(tc.final), &expected)
			if len(rendered.Contents) != 1 || len(rendered.Contents[0].Parts) != len(expected) {
				t.Fatalf("same-route lost layout: %s", same)
			}
			for i, raw := range expected {
				var x, y any
				json.Unmarshal(raw, &x)
				json.Unmarshal(rendered.Contents[0].Parts[i], &y)
				xx, _ := json.Marshal(x)
				yy, _ := json.Marshal(y)
				if string(xx) != string(yy) {
					t.Fatalf("part %d replay %s want %s", i, yy, xx)
				}
			}
			for _, target := range []common.Config{
				{Vendor: common.Gemini, Surface: common.GenerateContent, ResolvedModel: "other"},
				{Vendor: common.Gemini, Surface: common.ChatCompletions, ResolvedModel: "actual"},
				{Vendor: common.OpenAI, Surface: common.ChatCompletions, ResolvedModel: "actual"},
				{Vendor: common.Anthropic, Surface: common.Messages, ResolvedModel: "actual"},
			} {
				target.Model = "route"
				cross, err := render(context, target, nil)
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(string(cross), "thoughtSignature") || strings.Contains(string(cross), "provider_extension") || strings.Contains(string(cross), "hidden") {
					t.Fatalf("opaque crossed target %+v: %s", target, cross)
				}
				if tc.name == "signed closes text" {
					// Ordinary text remains portable even where providers join it.
					body := string(cross)
					if target.Vendor == common.Gemini {
						if !strings.Contains(body, `"text":"a"`) || !strings.Contains(body, `"text":"b"`) {
							t.Fatalf("text lost: %s", body)
						}
					} else if !strings.Contains(body, `"content":"ab"`) {
						t.Fatalf("text lost: %s", body)
					}
				}
			}
		})
	}
}
func TestCh6R1OpenAIReasoningIdentity(t *testing.T) {
	wire := "data: " + `{"model":"actual","choices":[{"delta":{"reasoning":"first"}}]}` + "\n\ndata: " + `{"choices":[{"delta":{"reasoning_content":"second"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}` + "\n\ndata: [DONE]\n\n"
	var deltas []common.PartDelta
	events, err := parseBody(parser(common.OpenAI, common.ChatCompletions), wire, "text/event-stream", common.ParseCallbacks{OnDelta: func(d common.PartDelta) { deltas = append(deltas, d) }})
	if err != nil {
		t.Fatal(err)
	}
	if len(deltas) != 2 || deltas[0].PartID != 0 || deltas[1].PartID != 1 {
		t.Fatal(deltas)
	}
	parts := ended(t, events).Parts
	if len(parts) != 2 {
		t.Fatal(parts)
	}
	for _, p := range parts {
		var data map[string]string
		json.Unmarshal(p.Data, &data)
		if data["reasoning"] == "first" && p.ID != 0 {
			t.Fatalf("reasoning ID=%d", p.ID)
		}
		if data["reasoning_content"] == "second" && p.ID != 1 {
			t.Fatalf("reasoning_content ID=%d", p.ID)
		}
	}
}
func TestCh6R1OpenAIRefusalParity(t *testing.T) {
	for _, content := range []string{"", "visible"} {
		t.Run("content="+content, func(t *testing.T) {
			prefix := ""
			msg := map[string]any{"refusal": "cannot comply"}
			if content != "" {
				prefix = "data: " + `{"choices":[{"delta":{"content":"visible"}}]}` + "\n\n"
				msg["content"] = content
			}
			wire := prefix + "data: " + `{"model":"actual","choices":[{"delta":{"refusal":"cannot "}}]}` + "\n\ndata: " + `{"choices":[{"delta":{"refusal":"comply"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}` + "\n\ndata: [DONE]\n\n"
			var deltas []common.PartDelta
			e := parser(common.OpenAI, common.ChatCompletions)
			stream, err := parseBody(e, wire, "text/event-stream", common.ParseCallbacks{OnDelta: func(d common.PartDelta) { deltas = append(deltas, d) }})
			if err != nil {
				t.Fatal(err)
			}
			raw, _ := json.Marshal(map[string]any{"model": "actual", "choices": []any{map[string]any{"message": msg, "finish_reason": "stop"}}, "usage": map[string]int{"prompt_tokens": 1, "completion_tokens": 1}})
			plain, err := parseBody(e, string(raw), "application/json", common.ParseCallbacks{})
			if err != nil {
				t.Fatal(err)
			}
			a, _ := json.Marshal(ended(t, stream))
			b, _ := json.Marshal(ended(t, plain))
			if string(a) != string(b) {
				t.Fatalf("stream %s\nplain %s", a, b)
			}
			text := ""
			for _, d := range deltas {
				text += d.Chunk
			}
			if text != content+"cannot comply" {
				t.Fatal(text)
			}
		})
	}
}
func TestCh6R1GeminiToolResultEnvelope(t *testing.T) {
	for _, tc := range []struct {
		name, payload string
		bad           bool
	}{
		{"object success", `{"value":3}`, false}, {"object error", `{"value":3}`, true}, {"false collision", `{"is_error":false,"value":3}`, true}, {"true collision", `{"is_error":true,"value":3}`, false}, {"string error", "failed", true}, {"string success", "done", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := common.Context{Dialogue: []common.Entry{{Actor: common.Assistant, Parts: common.PartList{{Type: "tool_call", CallID: "call", Name: "work", Args: json.RawMessage(`{}`)}}}, {Actor: common.Tool, Parts: common.PartList{{Type: "tool_result", CallID: "call", IsError: tc.bad, Parts: common.PartList{common.Text(tc.payload)}}}}}}
			wire, err := renderGemini(c, common.Config{Model: "route"}, nil)
			if err != nil {
				t.Fatal(err)
			}
			var decoded struct {
				Contents []struct {
					Parts []struct {
						Result *gResult `json:"functionResponse"`
					} `json:"parts"`
				} `json:"contents"`
			}
			json.Unmarshal(wire, &decoded)
			result := decoded.Contents[1].Parts[0].Result
			if result.ID != "call" || result.Name != "work" {
				t.Fatal(result)
			}
			var envelope map[string]json.RawMessage
			json.Unmarshal(result.Response, &envelope)
			if tc.bad || strings.Contains(tc.payload, "is_error") {
				var flag bool
				if json.Unmarshal(envelope["is_error"], &flag) != nil || flag != tc.bad {
					t.Fatalf("lost actual status: %s", result.Response)
				}
				var value any
				if json.Unmarshal([]byte(tc.payload), &value) == nil {
					actual, _ := json.Marshal(value)
					if string(envelope["result"]) != string(actual) {
						t.Fatalf("lost object payload: %s", result.Response)
					}
				}
			}
			if tc.name == "object success" && string(result.Response) != tc.payload {
				t.Fatalf("changed success object: %s", result.Response)
			}
		})
	}
}

func TestCh6R1OpenAIEmptyReasoningParity(t *testing.T) {
	e := parser(common.OpenAI, common.ChatCompletions)
	frames := []string{`{"model":"actual","choices":[{"delta":{"reasoning":""}}]}`, `{"choices":[{"delta":{"reasoning_content":"","content":"answer"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`, `[DONE]`}
	var stream strings.Builder
	for _, f := range frames {
		stream.WriteString("data: " + f + "\n\n")
	}
	streamed, err := parseBody(e, stream.String(), "text/event-stream", common.ParseCallbacks{})
	if err != nil {
		t.Fatal(err)
	}
	plain := `{"model":"actual","choices":[{"message":{"content":"answer","reasoning_content":"","reasoning":""},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`
	events, err := parseBody(e, plain, "application/json", common.ParseCallbacks{})
	if err != nil {
		t.Fatal(err)
	}
	actual, expected := ended(t, streamed), ended(t, events)
	if len(actual.Parts) != 3 || len(expected.Parts) != 3 || actual.Parts[0].ID != 1 || actual.Parts[1].ID != 2 || actual.Parts[2].ID != 0 {
		t.Fatalf("empty identities: %+v", actual.Parts)
	}
	for i := range actual.Parts {
		if string(actual.Parts[i].Data) != string(expected.Parts[i].Data) {
			t.Fatalf("parity: %+v %+v", actual.Parts, expected.Parts)
		}
	}
}
