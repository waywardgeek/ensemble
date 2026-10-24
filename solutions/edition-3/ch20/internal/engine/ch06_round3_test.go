package engine

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"ensemble/internal/common"
)

// This consumer deliberately knows only OnDelta(partID, kind, chunk). Name
// metadata cannot substitute for content a generic append-chunk consumer sees.
func TestCh6R3ToolNameChunks(t *testing.T) {
	cases := []struct {
		name    string
		vendor  common.Vendor
		surface common.Surface
		frames  []string
		plain   string
		names   map[int]string
	}{
		{"openai fragmented", common.OpenAI, common.ChatCompletions, []string{
			`{"model":"actual","choices":[{"delta":{"tool_calls":[{"index":0,"id":"call-1","type":"function","function":{"name":"run_"}}]}}]}`,
			`{"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"name":"command","arguments":"{\"cmd\":"}}]}}]}`,
			`{"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\"printf ok\"}"}}]}}]}`,
			`{"choices":[{"delta":{},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`,
			`[DONE]`,
		}, `{"model":"actual","choices":[{"message":{"tool_calls":[{"id":"call-1","type":"function","function":{"name":"run_command","arguments":"{\"cmd\":\"printf ok\"}"}}]},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`, map[int]string{1: "run_", 2: "command"}},
		{"anthropic whole", common.Anthropic, common.Messages, []string{
			`{"type":"message_start","message":{"model":"actual","usage":{"input_tokens":1}}}`,
			`{"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"call-1","name":"run_command","input":{}}}`,
			`{"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"cmd\":"}}`,
			`{"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"\"printf ok\"}"}}`,
			`{"type":"content_block_stop","index":0}`,
			`{"type":"message_delta","delta":{"stop_reason":"tool_use"},"usage":{"output_tokens":1}}`,
			`{"type":"message_stop"}`,
		}, `{"model":"actual","content":[{"type":"tool_use","id":"call-1","name":"run_command","input":{"cmd":"printf ok"}}],"stop_reason":"tool_use","usage":{"input_tokens":1,"output_tokens":1}}`, map[int]string{2: "run_command"}},
		{"gemini whole signed", common.Gemini, common.GenerateContent, []string{
			`{"modelVersion":"actual","candidates":[{"content":{"parts":[{"functionCall":{"id":"call-1","name":"run_command","args":{"cmd":"printf ok"}},"thoughtSignature":"tool-signature"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1}}`,
		}, `{"modelVersion":"actual","candidates":[{"content":{"parts":[{"functionCall":{"id":"call-1","name":"run_command","args":{"cmd":"printf ok"}},"thoughtSignature":"tool-signature"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1}}`, map[int]string{1: "run_command"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := parser(tc.vendor, tc.surface)
			var wire strings.Builder
			for _, frame := range tc.frames {
				wire.WriteString("data: " + frame + "\n\n")
			}
			frame := 0
			var chunks []string
			arrivals := map[int][]string{}
			appendChunk := func(id uint64, kind common.DeltaKind, chunk string) {
				if kind != common.DeltaToolCall || id != 0 {
					t.Errorf("kind=%s id=%d", kind, id)
				}
				chunks = append(chunks, chunk)
				arrivals[frame] = append(arrivals[frame], chunk)
			}
			streamed, err := parseBody(e, wire.String(), "text/event-stream", common.ParseCallbacks{
				OnFrame: func(common.Frame) { frame++ },
				OnDelta: func(d common.PartDelta) { appendChunk(d.PartID, d.Kind, d.Chunk) },
			})
			if err != nil {
				t.Fatal(err)
			}
			response := ended(t, streamed)
			if len(response.Parts) != 1 || response.Parts[0].Type != "tool_call" {
				t.Fatal(response.Parts)
			}
			part := response.Parts[0]
			if part.ID != 0 || part.Name != "run_command" || string(part.Args) != `{"cmd":"printf ok"}` {
				t.Fatalf("authoritative final changed: %+v", part)
			}
			content := strings.Join(chunks, "")
			if !strings.HasPrefix(content, part.Name) {
				t.Errorf("generic chunks %q omit tool name %q", content, part.Name)
			} else {
				var args any
				if err := json.Unmarshal([]byte(strings.TrimPrefix(content, part.Name)), &args); err != nil {
					t.Errorf("appended args not JSON: %q: %v", content, err)
				}
				if !reflect.DeepEqual(args, map[string]any{"cmd": "printf ok"}) {
					t.Errorf("appended args=%v", args)
				}
			}
			for n, name := range tc.names {
				if !strings.HasPrefix(strings.Join(arrivals[n], ""), name) {
					t.Errorf("frame %d chunks=%q, missing arriving name %q", n, arrivals[n], name)
				}
			}
			var plainChunks []string
			plain, err := parseBody(e, tc.plain, "application/json", common.ParseCallbacks{OnDelta: func(d common.PartDelta) {
				if d.PartID != 0 || d.Kind != common.DeltaToolCall {
					t.Errorf("plain delta %+v", d)
				}
				plainChunks = append(plainChunks, d.Chunk)
			}})
			if err != nil {
				t.Fatal(err)
			}
			if len(plainChunks) != 1 || plainChunks[0] != part.Name+string(part.Args) {
				t.Errorf("plain synthesized chunks=%q", plainChunks)
			}
			if content != strings.Join(plainChunks, "") {
				t.Errorf("stream/plain content differs: %q vs %q", content, plainChunks)
			}
			// Compare full finalized parts/provenance/usage independent of JSON key order.
			canonical := func(v any) any {
				b, err := json.Marshal(v)
				if err != nil {
					t.Fatal(err)
				}
				var out any
				if err := json.Unmarshal(b, &out); err != nil {
					t.Fatal(err)
				}
				return out
			}
			if !reflect.DeepEqual(canonical(response), canonical(ended(t, plain))) {
				t.Fatalf("stream/plain final mismatch: stream=%+v plain=%+v", response, ended(t, plain))
			}
			if tc.vendor == common.Gemini {
				rendered, err := renderGemini(common.Context{Dialogue: []common.Entry{{Actor: common.Assistant, Parts: response.Parts}}}, common.Config{Vendor: common.Gemini, Surface: common.GenerateContent, Model: "actual", ResolvedModel: "actual"}, nil)
				if err != nil || !strings.Contains(string(rendered), "tool-signature") {
					t.Fatalf("signed tool replay lost: %s err=%v", rendered, err)
				}
			}
		})
	}
}
