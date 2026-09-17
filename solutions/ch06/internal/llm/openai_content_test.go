package llm

import (
	"github.com/waywardgeek/coding-agents-course/solutions/ch06/internal/common"
	"encoding/json"
	"io"
	"testing"
)

// An assistant turn can legitimately end up carrying nothing at all.
//
// A reasoning model that spends its entire max_completion_tokens budget on
// reasoning returns content:"" with finish_reason:"length" — captured live on
// 2026-09-13 against gpt-5-2025-08-07:
//
//	"message": {"role":"assistant","content":"","refusal":null},
//	"finish_reason": "length",
//	"completion_tokens_details": {"reasoning_tokens": 1024}
//
// Parse keeps no parts for that turn, because "" is not worth a text part. But
// the turn is still replayed as history on the next round, and this is where
// it used to go wrong: the renderer represented "no text" as a nil *string,
// which marshals to null, and OpenAI refuses null on a bare assistant message:
//
//	Invalid value for 'content': expected a string, got null.
//
// The vendor never sent us null. We invented it, then sent it back.
//
// The second case below is the other half of the invariant and matters just as
// much: alongside tool_calls, null IS the correct thing to send, so this must
// never be "fixed" by never emitting null at all.
//
// Neither case is reachable from the grader: the fake vendor never returns an
// empty completion, so nothing in `grade -ch 3` can fail if the fix is
// deleted. This test is therefore the only thing standing between the fix and
// a silent regression that would only show up as an intermittent live 400.
func TestOpenAIRenderAssistantContentShape(t *testing.T) {
	cfg := common.Config{
		Vendor:    common.VendorOpenAI,
		Surface:   common.SurfaceChatCompletions,
		Model:     "gpt-5",
		APIKey:    "test-key",
		BaseURL:   "https://api.openai.com/v1",
		MaxTokens: 16,
	}

	contentOf := func(t *testing.T, c *common.Context) string {
		t.Helper()
		req, err := (openAISeam{}).Render(c, cfg)
		if err != nil {
			t.Fatalf("Render: %v", err)
		}
		raw, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var body struct {
			Messages []map[string]json.RawMessage `json:"messages"`
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("unmarshal request body: %v\n%s", err, raw)
		}
		if len(body.Messages) == 0 {
			t.Fatalf("request carried no messages:\n%s", raw)
		}
		return string(body.Messages[len(body.Messages)-1]["content"])
	}

	t.Run("empty agent turn renders as empty string, not null", func(t *testing.T) {
		c := &common.Context{Dialogue: []common.Entry{
			{Seq: 1, Actor: common.ActorHuman, Parts: common.PartList{common.TextPart{Text: "hello"}}},
			{Seq: 2, Actor: common.ActorAgent, Parts: nil},
		}}
		if got := contentOf(t, c); got != `""` {
			t.Errorf("assistant content = %s, want \"\"\n"+
				"null here is rejected live: expected a string, got null", got)
		}
	})

	t.Run("agent turn with tool calls still renders content null", func(t *testing.T) {
		c := &common.Context{Dialogue: []common.Entry{
			{Seq: 1, Actor: common.ActorHuman, Parts: common.PartList{common.TextPart{Text: "what time is it"}}},
			{Seq: 2, Actor: common.ActorAgent, Parts: common.PartList{
				common.ToolCallPart{CallID: "call_1", Name: "clock", Args: json.RawMessage(`{}`)},
			}},
		}}
		if got := contentOf(t, c); got != "null" {
			t.Errorf("assistant content = %s, want null "+
				"(null is correct and expected alongside tool_calls)", got)
		}
	})
}
