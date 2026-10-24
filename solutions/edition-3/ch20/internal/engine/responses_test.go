package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"ensemble/internal/common"
)

func TestResponsesTerminalAndPrivateParts(t *testing.T) {
	cfg := common.Config{Vendor: common.OpenAI, Surface: common.Responses, Model: "gpt-ch19-course"}
	completed := `{"status":"completed","model":"gpt-ch19-course","output":[{"type":"reasoning","id":"r1","summary":[{"type":"summary_text","text":"Safe summary"}],"encrypted_content":"PRIVATE"},{"type":"message","role":"assistant","content":[{"type":"output_text","text":"Hello"}]}],"usage":{"input_tokens":10,"output_tokens":3,"input_tokens_details":{"cached_tokens":4}}}`
	stream := `data: {"type":"response.reasoning_summary_text.delta","output_index":0,"summary_index":0,"delta":"Safe summary"}` + "\n\n" + `data: {"type":"response.output_text.delta","output_index":1,"content_index":0,"delta":"Hello"}` + "\n\n" + fmt.Sprintf("data: {\"type\":\"response.completed\",\"response\":%s}\n\n", completed)
	var deltas []common.PartDelta
	events, err := parseResponsesSSE(strings.NewReader(stream), cfg, common.ParseCallbacks{OnDelta: func(d common.PartDelta) { deltas = append(deltas, d) }})
	if err != nil {
		t.Fatal(err)
	}
	r := events[1].Response
	if len(r.Parts) != 3 || r.Parts[0].Type != "reasoning_summary" || r.Parts[1].Type != "opaque" || r.Parts[2].Type != "text" || r.Parts[0].ID != deltas[0].PartID || r.Parts[2].ID != deltas[1].PartID || r.Usage.Input != 6 || r.Usage.CacheRead != 4 {
		t.Fatalf("wrong terminal assembly: %+v %+v", r, deltas)
	}
	b, _ := json.Marshal(deltas)
	if bytes.Contains(b, []byte("PRIVATE")) {
		t.Fatal("private delta")
	}
	plain, err := parse(200, []byte(completed), cfg)
	var summaryDelta common.PartDelta
	if err == nil {
		plainDelta(plain[1].Response.Parts[0], common.ParseCallbacks{OnDelta: func(d common.PartDelta) { summaryDelta = d }})
	}
	if err != nil || summaryDelta.Kind != common.DeltaReasoningSummary {
		t.Fatal("plain summary projection", err)
	}
	for _, bad := range []string{
		`data: {"type":"response.function_call_arguments.delta","output_index":0,"delta":"{}"}` + "\n\n",
		"data: [DONE]\n\n",
		`data: {"type":"response.incomplete","response":{"model":"gpt-ch19-course","usage":{"input_tokens":10,"output_tokens":2}}}` + "\n\n",
		`data: {"type":"response.completed","response":{"status":"completed","output":[{"type":"function_call","name":"x","call_id":"c","arguments":"{"}]}}` + "\n\n",
	} {
		events, err := parseResponsesSSE(strings.NewReader(bad), cfg, common.ParseCallbacks{})
		if err == nil {
			t.Fatal("incomplete/invalid stream accepted")
		}
		for _, e := range events {
			if e.Response != nil {
				for _, p := range e.Response.Parts {
					if p.Type == "tool_call" {
						t.Fatal("failed stream retained executable call")
					}
				}
			}
		}
	}
}
func TestResponsesEndpointPolicy(t *testing.T) {
	for _, tc := range []struct {
		base string
		mode common.AuthMode
		want string
	}{{"", common.AuthChatGPT, "https://api.openai.com/v1/responses"}, {"", common.AuthAPI, "https://api.openai.com/v1/responses"}, {"http://127.0.0.1:123/custom/responses", common.AuthChatGPT, "http://127.0.0.1:123/custom/responses"}, {"http://127.0.0.1:123/v1", common.AuthAPI, "http://127.0.0.1:123/v1/responses"}} {
		if got := responsesEndpoint(tc.base, tc.mode); got != tc.want {
			t.Fatalf("got %s want %s", got, tc.want)
		}
	}
}
