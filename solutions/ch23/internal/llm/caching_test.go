package llm

import (
	"encoding/json"
	"testing"
)

// The system prompt is the largest stable span in a request, so it carries the
// cache breakpoint that pays for itself first. These tests lock the wire shape:
// the vendor accepts a bare string too, but a string cannot carry
// cache_control, and a silently uncached request looks exactly like a cached
// one until the bill arrives.

func TestSystemPromptCarriesACacheBreakpoint(t *testing.T) {
	blocks := systemBlocks("you are a helpful agent")
	if len(blocks) != 1 {
		t.Fatalf("want exactly one system block, got %d", len(blocks))
	}
	b := blocks[0]
	if b.Type != "text" {
		t.Errorf("block type = %q, want %q", b.Type, "text")
	}
	if b.Text != "you are a helpful agent" {
		t.Errorf("block text = %q, want the prompt verbatim", b.Text)
	}
	if b.CacheControl == nil {
		t.Fatal("no cache_control on the system block: the prefix will never be cached")
	}
	if b.CacheControl.Type != "ephemeral" {
		t.Errorf("cache_control type = %q, want %q", b.CacheControl.Type, "ephemeral")
	}
}

// An empty prompt must produce nil rather than an empty slice, so omitempty
// drops the key and a promptless request keeps the bytes it always had.
func TestEmptySystemPromptIsOmittedEntirely(t *testing.T) {
	if got := systemBlocks(""); got != nil {
		t.Fatalf("empty prompt produced %#v, want nil so omitempty can drop the field", got)
	}

	raw, err := json.Marshal(anthRequest{Model: "m", MaxTokens: 1})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, present := m["system"]; present {
		t.Errorf("system key present for an empty prompt: %s", raw)
	}
}

// The shape below is copied from a real request that achieves a high cache hit
// rate in production, not invented here. A regression in this test means the
// breakpoint stopped being sent.
func TestSystemSectionMarshalsAsABlockArray(t *testing.T) {
	raw, err := json.Marshal(anthRequest{
		Model:     "m",
		MaxTokens: 1,
		System:    systemBlocks("prompt"),
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	sys, present := m["system"]
	if !present {
		t.Fatalf("no system key in %s", raw)
	}

	// Must be an array, not a string: json.Unmarshal into []… proves the shape.
	var blocks []struct {
		Type         string `json:"type"`
		Text         string `json:"text"`
		CacheControl *struct {
			Type string `json:"type"`
		} `json:"cache_control"`
	}
	if err := json.Unmarshal(sys, &blocks); err != nil {
		t.Fatalf("system is not a content-block array (%v): %s", err, sys)
	}
	if len(blocks) != 1 {
		t.Fatalf("want 1 system block, got %d: %s", len(blocks), sys)
	}
	if blocks[0].CacheControl == nil || blocks[0].CacheControl.Type != "ephemeral" {
		t.Errorf("system block lost its cache breakpoint on the wire: %s", sys)
	}
}
