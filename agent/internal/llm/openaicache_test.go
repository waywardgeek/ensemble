package llm

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// oaiWire is the part of a rendered OpenAI request these tests care about.
type oaiWire struct {
	Messages []struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	} `json:"messages"`
	PromptCacheOptions *struct {
		Mode string `json:"mode"`
	} `json:"prompt_cache_options"`
}

func renderOAIWire(t *testing.T, c *common.Context, model string) oaiWire {
	t.Helper()
	req, err := (openAISeam{}).Render(c, common.Config{
		Model:        model,
		SystemPrompt: strings.Repeat("system prompt. ", 40),
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	raw, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var w oaiWire
	if err := json.Unmarshal(raw, &w); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	return w
}

// markedOAI returns the indices of messages carrying a breakpoint.
func markedOAI(w oaiWire) []int {
	var out []int
	for i, m := range w.Messages {
		if strings.Contains(string(m.Content), `"prompt_cache_breakpoint"`) {
			out = append(out, i)
		}
	}
	return out
}

// An explicit-capable model should get markers AND the opt-in, together.
func TestOpenAIExplicitBreakpointsLandAndOptIn(t *testing.T) {
	c := ctxWith(workEntries(1, "ALPHA"), workEntries(2, "BETA"))
	w := renderOAIWire(t, c, "gpt-5.6-course")

	marks := markedOAI(w)
	if len(marks) < 2 {
		t.Fatalf("expected the system prompt and the rolling pair to be marked, got %d: %v", len(marks), marks)
	}
	if marks[0] != 0 {
		t.Errorf("system prompt should carry the first breakpoint, first mark was at %d", marks[0])
	}
	if w.PromptCacheOptions == nil || w.PromptCacheOptions.Mode != "explicit" {
		t.Errorf("markers landed but the request did not opt into explicit mode: %+v", w.PromptCacheOptions)
	}
	if len(marks) > 4 {
		t.Errorf("vendor allows at most 4 cache writes, placed %d: %v", len(marks), marks)
	}
}

// The compaction bound gets its own breakpoint, as on Anthropic.
func TestOpenAIHandoffCarriesItsOwnBreakpoint(t *testing.T) {
	c := ctxWith(workEntries(1, "OLDWORK"))
	applyHandoff(t, c, 50, "HANDOFF-NOTE")
	c.Dialogue = append(c.Dialogue, workEntries(60, "AFTER-ONE")...)
	c.Dialogue = append(c.Dialogue, workEntries(70, "AFTER-TWO")...)

	w := renderOAIWire(t, c, "gpt-5.6-course")

	var handoffIdx = -1
	for i, m := range w.Messages {
		if strings.Contains(string(m.Content), "HANDOFF-NOTE") {
			handoffIdx = i
		}
	}
	if handoffIdx < 0 {
		t.Fatal("handoff note did not appear in the rendered messages")
	}
	if !strings.Contains(string(w.Messages[handoffIdx].Content), `"prompt_cache_breakpoint"`) {
		t.Errorf("the compaction bound carried no breakpoint; marks at %v, handoff at %d",
			markedOAI(w), handoffIdx)
	}
}

// The footgun. An implicit-only model must get neither markers nor the opt-in:
// declaring explicit mode without marking anything disables caching entirely,
// which is strictly worse than saying nothing at all.
func TestOpenAIImplicitModelIsLeftAlone(t *testing.T) {
	c := ctxWith(workEntries(1, "ALPHA"), workEntries(2, "BETA"))
	w := renderOAIWire(t, c, "gpt-5-course")

	if marks := markedOAI(w); len(marks) != 0 {
		t.Errorf("implicit-only model should carry no breakpoints, got %v", marks)
	}
	if w.PromptCacheOptions != nil {
		t.Errorf("implicit-only model must not opt into explicit mode, got %+v", w.PromptCacheOptions)
	}
}

// An unmarked request must never declare explicit mode. This is the invariant
// underneath the test above: the opt-in follows evidence that a marker landed,
// not the intent to place one.
func TestOpenAIOptInRequiresAMarker(t *testing.T) {
	c := ctxWith(workEntries(1, "ALPHA"))
	for _, model := range []string{"gpt-5.6-course", "gpt-5-course", "no-such-model"} {
		w := renderOAIWire(t, c, model)
		if len(markedOAI(w)) == 0 && w.PromptCacheOptions != nil {
			t.Errorf("%s: explicit mode declared with nothing marked", model)
		}
	}
}

// A marker must not be dropped just because it landed on a message with null
// content; it walks back to the nearest message that can carry one.
func TestOpenAIMarkerSkipsNullContent(t *testing.T) {
	msgs := []oaiMsg{
		{Role: "system", Content: strptr("sys")},
		{Role: "assistant", ToolCalls: []oaiToolCall{{ID: "t1"}}}, // null content
	}
	if !markOAICache(msgs, 1) {
		t.Fatal("marker was dropped instead of walking back")
	}
	if !msgs[0].CacheBreak {
		t.Error("marker should have landed on the nearest message with content")
	}
}

// Flipping a message between string and array form must not change its text.
func TestOpenAIMarkedMessageKeepsItsText(t *testing.T) {
	body := "the quick brown fox"
	plainBytes, err := json.Marshal(oaiMsg{Role: "user", Content: strptr(body)})
	if err != nil {
		t.Fatalf("marshal plain: %v", err)
	}
	markedBytes, err := json.Marshal(oaiMsg{Role: "user", Content: strptr(body), CacheBreak: true})
	if err != nil {
		t.Fatalf("marshal marked: %v", err)
	}
	if strings.Contains(string(plainBytes), `"prompt_cache_breakpoint"`) {
		t.Error("an unmarked message carried a breakpoint")
	}
	if !strings.Contains(string(markedBytes), `"prompt_cache_breakpoint"`) {
		t.Error("a marked message carried no breakpoint")
	}
	if !strings.Contains(string(markedBytes), body) {
		t.Errorf("marked form lost the message text: %s", markedBytes)
	}
}
