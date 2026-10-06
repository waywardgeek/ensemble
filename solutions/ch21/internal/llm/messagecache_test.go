package llm

import (
	"encoding/json"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// anthWireMsgs is the decoded messages array, carrying just enough of each
// content block to locate breakpoints and prove what they are attached to.
type anthWireMsgs struct {
	Messages []struct {
		Role    string `json:"role"`
		Content []struct {
			Type         string `json:"type"`
			Text         string `json:"text"`
			CacheControl *struct {
				Type string `json:"type"`
			} `json:"cache_control"`
		} `json:"content"`
	} `json:"messages"`
}

// renderAnthWire renders a context and decodes the messages array. Breakpoints
// are inspected from the wire, never from the struct: the failure this whole
// file guards against is a marker that exists in Go and not in the JSON.
func renderAnthWire(t *testing.T, ctx *common.Context) anthWireMsgs {
	t.Helper()
	cfg := common.Config{
		Vendor:  common.VendorAnthropic,
		Model:   "claude-sonnet-5",
		BaseURL: "https://api.anthropic.com",
		APIKey:  "test-key",
	}
	req, err := (anthropicSeam{}).Render(ctx, cfg)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	body, err := common.BodyOf(req)
	if err != nil {
		t.Fatalf("BodyOf: %v", err)
	}
	var wire anthWireMsgs
	if err := json.Unmarshal(body, &wire); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return wire
}

// marked returns the text of every block carrying a breakpoint.
func (w anthWireMsgs) marked() []string {
	var out []string
	for _, m := range w.Messages {
		for _, b := range m.Content {
			if b.CacheControl != nil {
				out = append(out, b.Text)
			}
		}
	}
	return out
}

// TestMessagesCarryOneRollingBreakpoint proves the messages array gets exactly
// one breakpoint, at the current end.
//
// An earlier design put a second one at the end of the previous exchange, on
// the theory that it reads a prefix an earlier request already paid to write.
// It is redundant: the rolling marker from the PREVIOUS round wrote an entry
// at that very prefix, and the vendor's look-back finds it. The freed slot
// goes to the tool array, which is a genuinely different boundary and the only
// one that survives a change to the system prompt.
//
// The absent marker is also a bug detector. If one ever looked necessary at
// the start of a turn, the cause would be the prefix below the newest prompt
// being rewritten between rounds, and the cure is to stop rewriting it rather
// than to pin it with one of four scarce slots.
func TestMessagesCarryOneRollingBreakpoint(t *testing.T) {
	ctx := &common.Context{
		Dialogue: []common.Entry{
			{Actor: common.ActorHuman, Parts: common.PartList{common.TextPart{Text: "first"}}},
			{Actor: common.ActorAgent, Parts: common.PartList{common.TextPart{Text: "reply"}}},
			{Actor: common.ActorHuman, Parts: common.PartList{common.TextPart{Text: "second"}}},
		},
	}
	got := renderAnthWire(t, ctx).marked()
	if len(got) != 1 {
		t.Fatalf("want 1 breakpoint in messages, got %d: %q", len(got), got)
	}
	// Asserting WHICH block, not just how many: a marker on the wrong block
	// counts the same and caches nothing.
	if !hasExact(got, "second") {
		t.Errorf("rolling breakpoint not at the newest turn; marked %q", got)
	}
	if hasExact(got, "reply") {
		t.Errorf("breakpoint at the end of the previous exchange; that prefix is "+
			"already covered by the previous round's rolling marker: %q", got)
	}
}

// TestBreakpointNeverInsideEphemera is the ordering guard.
//
// Ephemera are delivered once and then cleared. A breakpoint placed inside
// them caches bytes that will not exist next turn, so the cache misses on
// every single turn while the request still looks correctly marked. That is
// the exact shape of this chapter's bug: a defect whose only symptom is cost.
func TestBreakpointNeverInsideEphemera(t *testing.T) {
	ctx := &common.Context{
		Dialogue: []common.Entry{
			{Actor: common.ActorHuman, Parts: common.PartList{common.TextPart{Text: "stable"}}},
		},
		Ephemera: common.PartList{common.TextPart{Text: "volatile"}},
	}
	wire := renderAnthWire(t, ctx)

	// Prove the fixture actually produced the condition: if the ephemeral text
	// never reached the request, this test would pass while checking nothing.
	if !wireHasText(wire, "volatile") {
		t.Fatal("fixture failed: ephemeral text absent from the request")
	}
	for _, m := range wire.Messages {
		for _, b := range m.Content {
			if b.Text == "volatile" && b.CacheControl != nil {
				t.Error("breakpoint placed inside ephemera: caches bytes that are about to be dropped")
			}
		}
	}
	if got := wire.marked(); !hasExact(got, "stable") {
		t.Errorf("breakpoint did not fall back to the last stable block; marked %q", got)
	}
}

// TestMarkCacheSkipsReplayBlocks guards the trap the brief calls out by name.
//
// anthBlock.MarshalJSON returns Raw verbatim for replay material, so a marker
// set on a replayed block is present in Go and absent on the wire. markCache
// must walk backwards to a block that can carry one rather than set a field
// that will be silently discarded.
func TestMarkCacheSkipsReplayBlocks(t *testing.T) {
	msgs := []anthMsg{{
		Role: "assistant",
		Content: []anthBlock{
			{Type: "text", Text: "carryable"},
			{Raw: json.RawMessage(`{"type":"thinking","thinking":"opaque"}`)},
		},
	}}
	if !markCache(msgs, 0, len(msgs[0].Content)) {
		t.Fatal("markCache found nowhere to place a breakpoint")
	}
	if msgs[0].Content[1].CacheControl != nil {
		t.Error("breakpoint set on replay material: it would never reach the wire")
	}
	if msgs[0].Content[0].CacheControl == nil {
		t.Error("breakpoint did not walk back to the carryable block")
	}
}

// TestMarkCacheOnAllReplayGivesUp: an unmarked prefix is processed uncached,
// never rejected, so finding nowhere to put a marker is a return value and
// not an error.
func TestMarkCacheOnAllReplayGivesUp(t *testing.T) {
	msgs := []anthMsg{{
		Role:    "assistant",
		Content: []anthBlock{{Raw: json.RawMessage(`{"type":"thinking","thinking":"opaque"}`)}},
	}}
	if markCache(msgs, 0, 1) {
		t.Error("markCache claimed to place a breakpoint on replay material")
	}
}

func hasExact(hay []string, needle string) bool {
	for _, s := range hay {
		if s == needle {
			return true
		}
	}
	return false
}

func wireHasText(w anthWireMsgs, text string) bool {
	for _, m := range w.Messages {
		for _, b := range m.Content {
			if b.Text == text {
				return true
			}
		}
	}
	return false
}
