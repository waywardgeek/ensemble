package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"example.com/ensemble/internal/common"
)

// Real Messages stream in ch09/live-anthropic-g/responses/002.body supplied
// input:{} followed by partial_json:"". Zero argument bytes cannot replace
// the complete start object; nonempty partial JSON still must validate.
func TestMessagesEmptyArgumentDeltaKeepsStartObject(t *testing.T) {
	for _, c := range []struct {
		name, initial, want string
		deltas              []string
		valid               bool
	}{
		{"none", `{}`, `{}`, nil, true},
		{"empty", `{}`, `{}`, []string{""}, true},
		{"empty-with-fields", `{"path":"."}`, `{"path":"."}`, []string{"", ""}, true},
		{"partial-replaces", `{}`, `{"path":"."}`, []string{"", `{"path":`, "", `"."}`, ""}, true},
		{"malformed-is-not-fallback", `{}`, "", []string{"", "{"}, false},
		{"whitespace-is-not-empty", `{}`, "", []string{" "}, false},
		{"missing-start-object", `null`, "", []string{""}, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			var wire strings.Builder
			wire.WriteString("data: {\"type\":\"message_start\",\"message\":{\"model\":\"fixture\",\"content\":[],\"usage\":{\"input_tokens\":2}}}\n\n")
			fmt.Fprintf(&wire, "data: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"tool_use\",\"id\":\"c1\",\"name\":\"list_directory\",\"input\":%s}}\n\n", c.initial)
			for _, delta := range c.deltas {
				b, _ := json.Marshal(delta)
				fmt.Fprintf(&wire, "data: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"input_json_delta\",\"partial_json\":%s}}\n\n", b)
			}
			wire.WriteString("data: {\"type\":\"content_block_stop\",\"index\":0}\n\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"tool_use\"},\"usage\":{\"output_tokens\":3}}\n\ndata: {\"type\":\"message_stop\"}\n\n")
			op, owner := testOperation()
			parsed, err := parseStream(context.Background(), op, common.Config{Vendor: "anthropic", Model: "fixture"}, strings.NewReader(wire.String()))
			if !c.valid {
				if err == nil {
					t.Fatal("invalid arguments accepted by falling back")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(parsed.Response.Parts) != 1 || string(parsed.Response.Parts[0].Args) != c.want {
				t.Fatalf("wrong complete call: %+v", parsed.Response.Parts)
			}
			args := ""
			for _, f := range owner.fragments {
				if f.Channel == "tool_args" {
					args += f.Text
				}
			}
			if args != c.want {
				t.Fatalf("argument display %q; want %q", args, c.want)
			}
		})
	}
}
