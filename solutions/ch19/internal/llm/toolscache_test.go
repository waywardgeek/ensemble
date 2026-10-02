package llm

import (
	"encoding/json"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// Tests for the first of the four breakpoints: the end of the tool array.
//
// Tools sit at the very front of the cache order (tools, then system, then
// messages), so this marker is the only one that can keep the tool
// declarations warm across a turn that edits the system prompt. Without it,
// changing one word of the system prompt throws the tools away too.

// anthWireTools decodes just the tools array. The breakpoint is read off the
// wire rather than off the struct for the same reason as everywhere else in
// these tests: a marker that exists in Go and not in the JSON buys nothing.
type anthWireTools struct {
	Tools []struct {
		Name         string `json:"name"`
		CacheControl *struct {
			Type string `json:"type"`
		} `json:"cache_control"`
	} `json:"tools"`
}

func renderAnthTools(t *testing.T, c *common.Context, decls []common.ToolDecl) anthWireTools {
	t.Helper()
	cfg := common.Config{
		Vendor:  common.VendorAnthropic,
		Model:   "claude-sonnet-5",
		BaseURL: "https://api.anthropic.com",
		APIKey:  "test-key",
		Tools:   decls,
	}
	req, err := (anthropicSeam{}).Render(c, cfg)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	body, err := common.BodyOf(req)
	if err != nil {
		t.Fatalf("BodyOf: %v", err)
	}
	var wire anthWireTools
	if err := json.Unmarshal(body, &wire); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return wire
}

func decl(name string) common.ToolDecl {
	return common.ToolDecl{
		Name:        name,
		Description: "does " + name,
		Schema:      json.RawMessage(`{"type":"object"}`),
	}
}

// The marker belongs on the LAST tool and nowhere else. On the last one it
// closes a prefix containing every declaration; on an earlier one it would
// leave the remaining tools outside the cached prefix, which is the whole
// thing this breakpoint exists to prevent.
func TestToolsCarryTheFirstBreakpoint(t *testing.T) {
	c := ctxWith(workEntries(1, "alpha"))
	w := renderAnthTools(t, c, []common.ToolDecl{decl("read_file"), decl("run_command"), decl("edit_file")})

	if len(w.Tools) != 3 {
		t.Fatalf("tools on the wire = %d, want 3", len(w.Tools))
	}
	var marked []string
	for _, tl := range w.Tools {
		if tl.CacheControl != nil {
			marked = append(marked, tl.Name)
		}
	}
	if len(marked) != 1 {
		t.Fatalf("marked tools = %v, want exactly one", marked)
	}
	if marked[0] != "edit_file" {
		t.Errorf("marked tool = %q, want the last one (edit_file): a marker on an "+
			"earlier tool leaves the rest of the declarations uncached", marked[0])
	}
}

// An agent can legitimately run with no tools at all, and the marker must not
// invent a block to sit on.
func TestNoToolsLeavesNoToolBreakpoint(t *testing.T) {
	c := ctxWith(workEntries(1, "alpha"))
	w := renderAnthTools(t, c, nil)

	if len(w.Tools) != 0 {
		t.Fatalf("tools on the wire = %d, want 0", len(w.Tools))
	}
}

// The invariant Bill asked for by name: nothing is ever marked at the start of
// a turn.
//
// The rolling marker from the previous round already wrote a cache entry at
// that prefix, so a start-of-turn marker buys nothing. The deeper reason is
// that needing one would be evidence of a bug: it would mean the prefix below
// the newest prompt is being rewritten between rounds, and the cure for that
// is to stop rewriting it, not to spend one of four scarce slots pinning it.
func TestNoBreakpointAtTheStartOfATurn(t *testing.T) {
	// Two stretches of work and no compaction, so the only legitimate marker
	// is the rolling one at the current end. Asserting on POSITION rather than
	// on block text is deliberate: a marker can land on a block whose wire
	// text is empty, such as a tool result, and a text-matching assertion
	// would wave it through.
	c := ctxWith(workEntries(1, "alpha"), workEntries(10, "beta"))
	w := renderAnthWire(t, c)

	last := len(w.Messages) - 1
	for i, m := range w.Messages {
		for _, b := range m.Content {
			if b.CacheControl == nil {
				continue
			}
			if i != last {
				t.Errorf("breakpoint in message %d of %d (role %q, text %q): with no "+
					"compaction the only marker belongs at the current end. That "+
					"prefix is already covered by the previous round's rolling "+
					"marker, and needing one here would mean the prefix is being "+
					"rewritten mid-turn", i, last, m.Role, b.Text)
			}
		}
	}
}
