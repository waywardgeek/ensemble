package llm

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"example.com/ensemble/internal/common"
)

// The unchanged provider body exposed a Chapter 6 assembly bug during Chapter 9.
// These maintenance cases vary the published argument/terminal boundaries only.
func TestMessagesZeroByteArguments(t *testing.T) {
	body, err := os.ReadFile("testdata/messages-empty-argument.sse")
	if err != nil {
		t.Fatal(err)
	}
	if len(body) != 2668 || fmt.Sprintf("%x", sha256.Sum256(body)) != "e6f7b7446b3dab2fe5ef6466111a6790c31ac52d45eccc8296a416ea9a4971ef" {
		t.Fatal("original provider fixture changed")
	}
	cases := []struct {
		name, start, want string
		deltas            []any
		terminal, valid   bool
	}{
		{"recorded", "", `{}`, nil, true, true},
		{"no-deltas", `{"path":"."}`, `{"path":"."}`, nil, true, true},
		{"one-empty", `{"path":"."}`, `{"path":"."}`, []any{""}, true, true},
		{"two-empty", `{"path":"."}`, `{"path":"."}`, []any{"", ""}, true, true},
		{"replacement", `{"path":"old","discard":true}`, `{"path":"new"}`, []any{"", `{"path":`, "", `"new"}`, ""}, true, true},
		{"whitespace", `{}`, "", []any{" "}, true, false},
		{"malformed", `{}`, "", []any{"", "{"}, true, false},
		{"null-replacement", `{}`, "", []any{"null"}, true, false},
		{"array-replacement", `{}`, "", []any{"[]"}, true, false},
		{"null-start", `null`, "", []any{""}, true, false},
		{"missing-start", `missing`, "", []any{""}, true, false},
		{"wrong-delta-type", `{}`, "", []any{nil}, true, false},
		{"missing-terminal", `{}`, "", []any{""}, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire := string(body)
			if tc.name != "recorded" {
				var transformed strings.Builder
				for _, line := range strings.Split(string(body), "\n") {
					if !strings.HasPrefix(line, "data: ") {
						continue
					}
					var event map[string]any
					if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &event); err != nil {
						t.Fatal(err)
					}
					emit := func() {
						data, err := json.Marshal(event)
						if err != nil {
							t.Fatal(err)
						}
						fmt.Fprintf(&transformed, "data: %s\n\n", data)
					}
					if block, ok := event["content_block"].(map[string]any); ok && block["type"] == "tool_use" {
						if tc.start == "missing" {
							delete(block, "input")
						} else {
							var input any
							if err := json.Unmarshal([]byte(tc.start), &input); err != nil {
								t.Fatal(err)
							}
							block["input"] = input
						}
					}
					if delta, ok := event["delta"].(map[string]any); ok && delta["type"] == "input_json_delta" {
						for _, value := range tc.deltas {
							delta["partial_json"] = value
							emit()
						}
						continue
					}
					if event["type"] != "message_stop" || tc.terminal {
						emit()
					}
				}
				wire = transformed.String()
			}
			op, owner := testOperation()
			parsed, err := parseStream(context.Background(), op, common.Config{Vendor: "anthropic", Model: "claude-sonnet-4-6"}, &splitReader{data: []byte(wire)})
			if !tc.valid {
				if err == nil {
					t.Fatal("accepted invalid argument or incomplete response")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var calls []common.Part
			for _, part := range parsed.Response.Parts {
				if part.Type == "tool_call" {
					calls = append(calls, part)
				}
			}
			if len(calls) != 1 || calls[0].Name != "list_directory" {
				t.Fatalf("calls: %#v", calls)
			}
			var want, actual, displayed any
			json.Unmarshal([]byte(tc.want), &want)
			if err := json.Unmarshal(calls[0].Args, &actual); err != nil {
				t.Fatal(err)
			}
			var fragments strings.Builder
			for _, fragment := range owner.fragments {
				if fragment.Channel == "tool_args" {
					if fragment.Text == "" {
						t.Fatal("empty display fragment")
					}
					fragments.WriteString(fragment.Text)
				}
			}
			if err := json.Unmarshal([]byte(fragments.String()), &displayed); err != nil {
				t.Fatal("arguments displayed more than once or malformed:", err)
			}
			if !reflect.DeepEqual(want, actual) || !reflect.DeepEqual(want, displayed) {
				t.Fatalf("want %#v, accepted %#v, displayed %#v", want, actual, displayed)
			}
			if parsed.Response.Usage.Input != 2016 || parsed.Response.Usage.Output != 51 {
				t.Fatalf("usage changed: %#v", parsed.Response.Usage)
			}
		})
	}
}
