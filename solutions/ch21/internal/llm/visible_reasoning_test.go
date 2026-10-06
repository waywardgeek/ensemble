package llm

import (
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// Some models reason internally and then act in silence. That is unworkable
// when the supervisor reviews the work by listening to it as it happens: a run
// nobody could follow cannot be reviewed afterwards except by reading every
// diff, which in practice means discarding the whole run.
//
// These tests pin the rule that such a model must narrate before it acts, and
// pin it as a property of the MODEL rather than of the agent, so a model that
// narrates willingly is never slowed down by it.

func vrActor(t *testing.T, model string) *Actor {
	t.Helper()
	e := lostCallEngine(t)
	e.Cfg.Model = model
	return &Actor{eng: e}
}

func vrCalls(ids ...string) []common.ToolCallPart {
	var out []common.ToolCallPart
	for _, id := range ids {
		out = append(out, common.ToolCallPart{CallID: id, Name: "edit_file"})
	}
	return out
}

// errorResults collects every tool result in the context that is flagged as an
// error, keyed by call id.
func errorResults(c *common.Context) map[string]string {
	out := map[string]string{}
	for _, e := range c.Dialogue {
		for _, p := range e.Parts {
			res, ok := p.(common.ToolResultPart)
			if !ok || !res.IsError {
				continue
			}
			var sb strings.Builder
			for _, rp := range res.Parts {
				if tp, ok := rp.(common.TextPart); ok {
					sb.WriteString(tp.Text)
				}
			}
			out[res.CallID] = sb.String()
		}
	}
	return out
}

func TestVisibleReasoningRefusesASilentToolCall(t *testing.T) {
	a := vrActor(t, "gpt-5.6-sol")
	if !a.enforceVisibleReasoning("", vrCalls("c1")) {
		t.Fatal("a tool call with no narration was allowed through")
	}
	got := errorResults(a.eng.Ctx)
	if len(got) != 1 {
		t.Fatalf("want 1 error result, got %d", len(got))
	}
	if !strings.Contains(got["c1"], "VISIBLE REASONING REQUIRED") {
		t.Errorf("result does not state the rule: %.80s", got["c1"])
	}
	// The model must be told plainly that it did not get away with it, or it
	// will carry on as though the edit landed.
	if !strings.Contains(got["c1"], "Nothing was executed") {
		t.Errorf("result does not say the call did not run: %.80s", got["c1"])
	}
}

// Whitespace is not narration. A model that emits a stray newline before acting
// would otherwise satisfy the check while telling the supervisor nothing.
func TestVisibleReasoningTreatsWhitespaceAsSilence(t *testing.T) {
	a := vrActor(t, "gpt-5.6-sol")
	if !a.enforceVisibleReasoning("  \n\t ", vrCalls("c1")) {
		t.Fatal("whitespace-only narration was accepted as narration")
	}
}

func TestVisibleReasoningAllowsANarratedToolCall(t *testing.T) {
	a := vrActor(t, "gpt-5.6-sol")
	narration := "I'll read actor.go to find the dispatch loop."
	if a.enforceVisibleReasoning(narration, vrCalls("c1")) {
		t.Fatal("a narrated tool call was refused")
	}
	if n := len(errorResults(a.eng.Ctx)); n != 0 {
		t.Fatalf("narrated call produced %d error results, want 0", n)
	}
}

// The rule is a property of the model. A model that narrates without being
// forced must not pay for this, so the check must not fire for it even when a
// turn legitimately has no text.
func TestVisibleReasoningIsNotEnforcedForOtherModels(t *testing.T) {
	a := vrActor(t, "claude-sonnet-5")
	if a.enforceVisibleReasoning("", vrCalls("c1")) {
		t.Fatal("enforcement fired for a model that does not require it")
	}
	if f, ok := common.LookupModel("claude-sonnet-5"); !ok || f.RequiresVisibleReasoning {
		t.Fatal("claude-sonnet-5 should exist and should not require enforcement")
	}
}

// An unknown model must not be enforced against either. Enforcing on a
// zero-value feature row would turn "I have never heard of this model" into a
// behaviour change, which is how a table lookup quietly becomes policy.
func TestVisibleReasoningIsNotEnforcedForUnknownModels(t *testing.T) {
	a := vrActor(t, "no-such-model")
	if a.enforceVisibleReasoning("", vrCalls("c1")) {
		t.Fatal("enforcement fired for an unknown model")
	}
}

// Every call in the batch is answered. A vendor requires a result for each
// call it made, so refusing some and ignoring the rest would trade this fault
// for a malformed request.
func TestVisibleReasoningRefusesEveryCallInTheBatch(t *testing.T) {
	a := vrActor(t, "gpt-5.6-sol")
	if !a.enforceVisibleReasoning("", vrCalls("c1", "c2", "c3")) {
		t.Fatal("silent batch was allowed through")
	}
	got := errorResults(a.eng.Ctx)
	for _, id := range []string{"c1", "c2", "c3"} {
		if _, ok := got[id]; !ok {
			t.Errorf("call %s got no result; a vendor requires one per call", id)
		}
	}
	if len(got) != 3 {
		t.Fatalf("want 3 error results, got %d", len(got))
	}
}

// The message has a job beyond stopping the model: it has to recruit it. These
// are the clauses that make the agent improvable, and a shortened message that
// dropped them would still pass every test above.
func TestVisibleReasoningMessageAsksForSurprisesToBeReported(t *testing.T) {
	// The constant is hard-wrapped prose, so a phrase that reads as one line
	// may be split across two. Compare on collapsed whitespace: the assertion
	// is about what the message SAYS, and it must not break the next time
	// someone reflows a paragraph.
	flat := strings.Join(strings.Fields(visibleReasoningRequired), " ")
	for _, want := range []string{
		"what you expected",
		"what you got instead",
		"still being built",
	} {
		if !strings.Contains(flat, want) {
			t.Errorf("message does not ask the model to report surprises: missing %q", want)
		}
	}
	// It must also say that internal reasoning is not a substitute, which is
	// the specific confusion that makes a model think it already complied.
	if !strings.Contains(flat, "INTERNAL REASONING DOES NOT COUNT") {
		t.Error("message does not rule out internal reasoning as narration")
	}
}

func TestGPT56SolRequiresVisibleReasoning(t *testing.T) {
	f, ok := common.LookupModel("gpt-5.6-sol")
	if !ok {
		t.Fatal("gpt-5.6-sol missing from the model table")
	}
	if !f.RequiresVisibleReasoning {
		t.Error("gpt-5.6-sol should require visible reasoning: it reasons internally and acts silently")
	}
}
