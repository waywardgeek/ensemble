package llm

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// oaiEffort renders an OpenAI request and reports what reasoning_effort, if
// any, reached the body.
func oaiEffort(t *testing.T, model string, tools []common.ToolDecl) (string, bool) {
	t.Helper()
	ctx := &common.Context{
		Dialogue: []common.Entry{
			{Actor: common.ActorHuman, Parts: common.PartList{common.TextPart{Text: "hello"}}},
		},
	}
	cfg := common.Config{
		Model:     model,
		MaxTokens: 256,
		Surface:   common.SurfaceChatCompletions,
		Tools:     tools,
	}
	req, err := (openAISeam{}).Render(ctx, cfg)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(body, &top); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	raw, present := top["reasoning_effort"]
	if !present {
		return "", false
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatalf("reasoning_effort is not a string: %s", raw)
	}
	return s, true
}

// TestNoThinkingWithToolsSendsNoneNotNothing pins down a live-only failure.
//
// Some models refuse a reasoning budget in the same request as function tools,
// and refuse it with a 400 rather than a quieter answer:
//
//	Function tools with reasoning_effort are not supported for gpt-5.6-sol
//	in /v1/chat/completions. To use function tools, use /v1/responses or set
//	reasoning_effort to 'none'.
//
// Which is fatal rather than cosmetic, because an agent always has tools. Every
// turn failed. The fake vendor could not have caught this: it accepts whatever
// we send, so the request was well formed by our lights and refused by theirs.
//
// The subtle half, and the reason this test asserts a VALUE rather than an
// absence: omitting the field does not mean "no reasoning". It means "your
// default", and the default here IS a reasoning effort, so the request is
// refused with the identical error while the field appears nowhere in the body.
// That is a genuinely confusing thing to debug, because the error names a field
// you can grep for and not find. It has to be sent, as "none".
func TestNoThinkingWithToolsSendsNoneNotNothing(t *testing.T) {
	tools := []common.ToolDecl{
		{Name: "read_file", Description: "read a file", Schema: json.RawMessage(`{"type":"object"}`)},
	}

	// The affected model, with tools: explicitly none.
	if f, ok := common.LookupModel("gpt-5.6-course"); !ok || !f.NoThinkingWithTools {
		t.Fatal("gpt-5.6-sol should be marked NoThinkingWithTools; the fixture for this test is gone")
	}
	got, present := oaiEffort(t, "gpt-5.6-course", tools)
	if !present {
		t.Error("reasoning_effort absent; it must be SENT as \"none\", because absent means the vendor's default and the vendor's default is what gets the request refused")
	} else if got != "none" {
		t.Errorf("reasoning_effort %q, want \"none\"", got)
	}

	// The same model with no tools has no conflict to resolve, so the
	// configured effort should survive. This is the half that stops the fix
	// from quietly becoming "this model never reasons".
	got, present = oaiEffort(t, "gpt-5.6-course", nil)
	if present && got == "none" {
		t.Error("reasoning_effort forced to \"none\" with no tools declared; the constraint is about tools and reasoning TOGETHER, so with no tools there is nothing to give up")
	}

	// A model without the flag must be untouched, or the fix has stopped being
	// a fact about one model and become a branch that applies to everyone.
	if f, ok := common.LookupModel("gpt-5-course"); ok && !f.NoThinkingWithTools {
		got, present = oaiEffort(t, "gpt-5-course", tools)
		if present && got == "none" {
			t.Error("gpt-6-astra is not marked NoThinkingWithTools, yet its effort was forced to \"none\"")
		}
	}
}
