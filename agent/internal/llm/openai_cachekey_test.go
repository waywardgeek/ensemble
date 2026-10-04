package llm

import (
	"encoding/json"
	"strings"
	"testing"
)

// The cache key exists to steer successive requests of one conversation to the
// same cache machine. Everything worth asserting about it follows from a single
// question: does the key hold still while the conversation grows?

func toolsFixture(desc string) []respTool {
	return []respTool{{
		Type:        "function",
		Name:        "read_file",
		Description: desc,
		Parameters:  json.RawMessage(`{"type":"object"}`),
	}}
}

func systemItem(text string) respItem {
	return respItem{
		Type:    "message",
		Role:    "developer",
		Content: []respContent{{Type: "input_text", Text: text}},
	}
}

func userItem(text string) respItem {
	return respItem{
		Type:    "message",
		Role:    "user",
		Content: []respContent{{Type: "input_text", Text: text}},
	}
}

func TestCacheKeyIsStableForTheSamePrefix(t *testing.T) {
	tools := toolsFixture("read a file")
	items := []respItem{systemItem("you are an agent"), userItem("hello")}

	a := respCacheKey("gpt-6-astra", tools, items)
	b := respCacheKey("gpt-6-astra", tools, items)
	if a != b {
		t.Fatalf("key is not deterministic: %q then %q", a, b)
	}
	if !strings.HasPrefix(a, "en-") {
		t.Errorf("key %q should carry the en- prefix that identifies its producer", a)
	}
}

// This is the one that matters. A key derived from the whole request would
// change on every turn, sending each request to a different machine and
// producing exactly the miss the key is meant to prevent. The failure would be
// invisible from the outside: the field is present, the request succeeds, and
// the cache simply never hits.
func TestCacheKeySurvivesAGrowingConversation(t *testing.T) {
	tools := toolsFixture("read a file")
	sys := systemItem("you are an agent")

	short := respCacheKey("gpt-6-astra", tools, []respItem{sys, userItem("hello")})

	grown := []respItem{sys}
	for i := 0; i < 40; i++ {
		grown = append(grown, userItem(strings.Repeat("more conversation ", 20)))
	}
	long := respCacheKey("gpt-6-astra", tools, grown)

	if short != long {
		t.Fatalf("key changed as the conversation grew:\n  short %q\n  long  %q\nA key that moves every turn defeats its own purpose.", short, long)
	}
}

func TestCacheKeyChangesWhenTheFrozenPrefixChanges(t *testing.T) {
	base := respCacheKey("gpt-6-astra", toolsFixture("read a file"), []respItem{systemItem("you are an agent")})

	cases := []struct {
		name string
		key  string
	}{
		{"different model", respCacheKey("gpt-5.6-sol", toolsFixture("read a file"), []respItem{systemItem("you are an agent")})},
		{"different tools", respCacheKey("gpt-6-astra", toolsFixture("read a file, carefully"), []respItem{systemItem("you are an agent")})},
		{"different system block", respCacheKey("gpt-6-astra", toolsFixture("read a file"), []respItem{systemItem("you are a different agent")})},
	}
	for _, c := range cases {
		if c.key == base {
			t.Errorf("%s: key did not change, so a stale machine would be reused for a prefix it has never seen", c.name)
		}
	}
}

func TestCacheKeyHandlesAnEmptyConversation(t *testing.T) {
	// Rendering can legitimately produce no items; the key must not panic and
	// must still distinguish one tool set from another.
	a := respCacheKey("gpt-6-astra", toolsFixture("read a file"), nil)
	b := respCacheKey("gpt-6-astra", toolsFixture("write a file"), nil)
	if a == "" || b == "" {
		t.Fatal("empty key")
	}
	if a == b {
		t.Error("tool declarations are part of the frozen prefix and must affect the key")
	}
}
