package llm

import (
	"context"
	"example.com/ensemble/internal/common"
	"fmt"
	"strings"
	"testing"
)

// Fixed 128-byte fragments distinguish incremental assembly from rebuilding the
// complete prefix. Report measurements; do not gate on machine-specific time.
func BenchmarkStreamAssembly(b *testing.B) {
	for _, vendor := range []string{"anthropic", "openai", "gemini"} {
		for _, size := range []int{64 << 10, 128 << 10} {
			b.Run(fmt.Sprintf("%s/%dKiB", vendor, size>>10), func(b *testing.B) {
				text := strings.Repeat("x", 128)
				var prefix, fragment, suffix string
				switch vendor {
				case "anthropic":
					prefix = "data: {\"type\":\"message_start\",\"message\":{\"usage\":{\"input_tokens\":1,\"output_tokens\":0}}}\n\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n"
					fragment = "data: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"" + text + "\"}}\n\n"
					suffix = "data: {\"type\":\"content_block_stop\",\"index\":0}\n\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":1}}\n\ndata: {\"type\":\"message_stop\"}\n\n"
				case "openai":
					fragment = "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"" + text + "\"}}]}\n\n"
					suffix = "data: {\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\"}]}\n\ndata: {\"choices\":[],\"usage\":{\"prompt_tokens\":1,\"completion_tokens\":1}}\n\ndata: [DONE]\n\n"
				case "gemini":
					fragment = "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"" + text + "\"}]}}]}\n\n"
					suffix = "data: {\"candidates\":[{\"finishReason\":\"STOP\"}],\"usageMetadata\":{\"promptTokenCount\":1,\"candidatesTokenCount\":1}}\n\n"
				}
				body := prefix + strings.Repeat(fragment, size/128) + suffix
				b.ReportAllocs()
				b.SetBytes(int64(size))
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					op, _ := testOperation()
					if _, err := parseStream(context.Background(), op, common.Config{Vendor: vendor, Model: "fixture"}, strings.NewReader(body)); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
