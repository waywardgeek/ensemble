package llm

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"example.com/ensemble/internal/common"
)

func TestEscapedOpaqueResponseBound(t *testing.T) {
	for _, kind := range []string{"thinking", "signature", "refusal"} {
		for _, limit := range []string{"small", "exact", "one-over"} {
			t.Run(kind+"/"+limit, func(t *testing.T) {
				vendor := "anthropic"
				prefix := `data: {"type":"message_start","message":{"usage":{"input_tokens":1,"output_tokens":0}}}` + "\n\n" + `data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":"V"}}` + "\n\n" + `data: {"type":"content_block_stop","index":0}` + "\n\n" + `data: {"type":"content_block_start","index":1,"content_block":{"type":"thinking","thinking":""}}` + "\n\n"
				template := `{"type":"thinking","thinking":""}`
				if kind == "signature" {
					template = `{"type":"thinking","thinking":"","signature":""}`
				}
				suffix := `data: {"type":"content_block_stop","index":1}` + "\n\n" + `data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":1}}` + "\n\n" + `data: {"type":"message_stop"}` + "\n\n"
				if kind == "refusal" {
					vendor = "openai"
					template = `{"refusal":""}`
					prefix = `data: {"choices":[{"index":0,"delta":{"content":"V"}}]}` + "\n\n"
					suffix = `data: {"choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}` + "\n\n" + `data: {"choices":[],"usage":{"prompt_tokens":1,"completion_tokens":1}}` + "\n\ndata: [DONE]\n\n"
				}
				encoded := 64 << 10
				if limit != "small" {
					encoded = responseLimit - 1 - len(template)
				}
				if limit == "one-over" {
					encoded++
				}
				text := strings.Repeat("\x00", encoded/6) + strings.Repeat("x", encoded%6)
				readers := []io.Reader{strings.NewReader(prefix)}
				for len(text) > 0 {
					n := len(text)
					if n > 64<<10 {
						n = 64 << 10
					}
					fragment := text[:n]
					text = text[n:]
					var payload string
					if kind == "refusal" {
						payload = fmt.Sprintf(`{"choices":[{"index":0,"delta":{"refusal":%s}}]}`, rawValue(fragment))
					} else {
						payload = fmt.Sprintf(`{"type":"content_block_delta","index":1,"delta":{"type":%q,%q:%s}}`, kind+"_delta", kind, rawValue(fragment))
					}
					frame := "data: " + payload + "\n\n"
					if len(frame) > pendingLimit {
						t.Fatal("fixture exceeds wire limit")
					}
					readers = append(readers, strings.NewReader(frame))
				}
				readers = append(readers, strings.NewReader(suffix))
				op, _ := testOperation()
				parsed, err := parseStream(context.Background(), op, common.Config{Vendor: vendor, Model: "fixture"}, io.MultiReader(readers...))
				if limit == "one-over" {
					if err == nil || !strings.Contains(err.Error(), "assembled response exceeds") {
						t.Fatalf("wrong overflow result: %v", err)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if len(parsed.Response.Parts) != 2 || parsed.Response.Parts[0].Text == nil || *parsed.Response.Parts[0].Text != "V" {
					t.Fatalf("parts: %d", len(parsed.Response.Parts))
				}
				actual := 1 + len(parsed.Response.Parts[1].Data)
				expected := 1 + len(template) + encoded
				if actual != expected {
					t.Fatalf("retained cost %d want %d", actual, expected)
				}
			})
		}
	}
}
