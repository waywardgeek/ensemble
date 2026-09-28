package cachelens

import (
	"encoding/json"
	"testing"
)

// sect builds a Sections value in cache order.
func sect(pairs ...string) Sections {
	s := Sections{Raw: map[string]json.RawMessage{}}
	for i := 0; i < len(pairs); i += 2 {
		s.Order = append(s.Order, pairs[i])
		s.Raw[pairs[i]] = json.RawMessage(pairs[i+1])
	}
	return s
}

func find(d Divergence, name string) SectionDiff {
	for _, s := range d.Sections {
		if s.Name == name {
			return s
		}
	}
	return SectionDiff{Name: "(missing)"}
}

func TestIdenticalRequestsReportIdentical(t *testing.T) {
	a := sect("tools", `[{"name":"read"}]`, "system", `"be helpful"`, "messages", `[{"r":"u"}]`)
	d := Compare(a, a)

	if !d.Identical {
		t.Errorf("Identical = false, want true")
	}
	if d.Unstable != "" {
		t.Errorf("Unstable = %q, want empty", d.Unstable)
	}
	if d.CacheableBytes != d.CurrentBytes {
		t.Errorf("CacheableBytes = %d, want %d (all of it)", d.CacheableBytes, d.CurrentBytes)
	}
}

func TestGrowingDialogueIsAnAppendNotAnEdit(t *testing.T) {
	// This is the case that decides whether the instrument is usable.
	//
	// Adding a message to a JSON array changes the byte at the closing
	// bracket: "[{a}]" becomes "[{a},{b}]", so the two strings first differ
	// where one has ']' and the other has ','. Compared naively that reads as
	// an edit — and it happens on every single turn, so the tool would report
	// a prefix bug forever and get muted.
	prior := sect("tools", `[{"name":"read"}]`, "system", `"be helpful"`,
		"messages", `[{"role":"user","content":"hi"}]`)
	current := sect("tools", `[{"name":"read"}]`, "system", `"be helpful"`,
		"messages", `[{"role":"user","content":"hi"},{"role":"assistant","content":"yo"}]`)

	d := Compare(prior, current)

	if got := find(d, "messages").Status; got != StatusAppended {
		t.Errorf("messages status = %q, want %q — a JSON array append was misread",
			got, StatusAppended)
	}
	// And the alarm must stay silent, because nothing above the dialogue moved.
	if d.Unstable != "" {
		t.Errorf("Unstable = %q, want empty — only the dialogue changed", d.Unstable)
	}
	if !d.DialogueChanged {
		t.Error("DialogueChanged = false, want true")
	}
}

func TestChangedToolsAreFlaggedAsUnstable(t *testing.T) {
	// Tools come first in cache order, so a change here voids everything. This
	// is the failure the tool exists to catch: nondeterministic ordering or a
	// re-rendered description silently costing a full cold read every turn.
	prior := sect("tools", `[{"name":"read"},{"name":"write"}]`, "system", `"s"`, "messages", `[]`)
	current := sect("tools", `[{"name":"write"},{"name":"read"}]`, "system", `"s"`, "messages", `[]`)

	d := Compare(prior, current)

	if d.Unstable != "tools" {
		t.Errorf("Unstable = %q, want \"tools\"", d.Unstable)
	}
	if got := find(d, "tools").Status; got != StatusEdited {
		t.Errorf("tools status = %q, want %q", got, StatusEdited)
	}
}

func TestCacheableRunStopsAtFirstChangedSection(t *testing.T) {
	// A later section matching perfectly is worth nothing once an earlier one
	// has moved: it sits past the breakpoint. Counting it would overstate what
	// the provider could actually serve, which is the one number this tool
	// must not inflate.
	prior := sect("tools", `[{"a":1}]`, "system", `"old"`, "messages", `[{"m":1}]`)
	current := sect("tools", `[{"a":1}]`, "system", `"new"`, "messages", `[{"m":1}]`)

	d := Compare(prior, current)

	toolsLen := len(`[{"a":1}]`)
	sysPrefix := find(d, "system").PrefixBytes

	// Only tools (identical) plus the matching head of system may count. The
	// messages section is identical but unreachable.
	want := toolsLen + sysPrefix
	if d.CacheableBytes != want {
		t.Errorf("CacheableBytes = %d, want %d — identical bytes after the first change were counted",
			d.CacheableBytes, want)
	}
	if d.CacheableBytes >= d.CurrentBytes {
		t.Errorf("CacheableBytes = %d, CurrentBytes = %d: a changed request cannot be fully cacheable",
			d.CacheableBytes, d.CurrentBytes)
	}
}

func TestTruncatedDialogueIsNotAnEdit(t *testing.T) {
	// Compaction shortens the dialogue. That is a deliberate act, not drift,
	// and must not be reported as a prefix bug.
	prior := sect("tools", `[]`, "system", `"s"`, "messages", `[{"m":1},{"m":2}]`)
	current := sect("tools", `[]`, "system", `"s"`, "messages", `[{"m":1}]`)

	d := Compare(prior, current)

	if got := find(d, "messages").Status; got != StatusTruncated {
		t.Errorf("messages status = %q, want %q", got, StatusTruncated)
	}
	if d.Unstable != "" {
		t.Errorf("Unstable = %q, want empty", d.Unstable)
	}
}

// TestMovedBreakpointIsNotAnEdit guards the false alarm introduced by the
// rolling breakpoint.
//
// The rolling marker advances every turn, so a block that carried one in the
// previous request carries none in this one. That difference sits deep inside
// the common prefix, where the trailing-append heuristic cannot reach it. A
// byte-exact comparison would therefore classify a perfectly healthy turn as
// an edit and cut the reported cacheable prefix off at the stale marker, on
// every single turn. An instrument that cries wolf every turn is muted within
// a day, and then the real cache break arrives unseen.
func TestMovedBreakpointIsNotAnEdit(t *testing.T) {
	const m = `,"cache_control":{"type":"ephemeral"}`
	prior := sect("tools", `[{"name":"read"}]`, "system", `"be helpful"`,
		"messages", `[{"role":"user","content":[{"type":"text","text":"first"`+m+`}]}]`)
	current := sect("tools", `[{"name":"read"}]`, "system", `"be helpful"`,
		"messages", `[{"role":"user","content":[{"type":"text","text":"first"}]},`+
			`{"role":"user","content":[{"type":"text","text":"second"`+m+`}]}]`)

	d := Compare(prior, current)

	if got := find(d, "messages").Status; got == StatusEdited {
		t.Errorf("messages classified as %v: a moved breakpoint is not a content edit", got)
	}
	if d.Unstable != "" {
		t.Errorf("Unstable = %q, want empty: a rolling breakpoint must not raise an alarm", d.Unstable)
	}
	// The marker is stripped from the comparison, but its presence must still
	// be reported. Otherwise a request that lost every breakpoint would show
	// all sections identical and look better than one that kept them.
	if d.Breakpoints != 1 {
		t.Errorf("Breakpoints = %d, want 1", d.Breakpoints)
	}
}

// TestBreakpointsAreCounted proves the count is reported rather than inferred,
// so losing a marker remains visible even when every section is identical.
func TestBreakpointsAreCounted(t *testing.T) {
	const m = `,"cache_control":{"type":"ephemeral"}`
	unmarked := sect("system", `"be helpful"`, "messages", `[{"role":"user","content":[{"type":"text","text":"hi"}]}]`)
	marked := sect("system", `"be helpful"`+m, "messages", `[{"role":"user","content":[{"type":"text","text":"hi"`+m+`}]}]`)

	if d := Compare(unmarked, unmarked); d.Breakpoints != 0 {
		t.Errorf("Breakpoints = %d on an unmarked request, want 0", d.Breakpoints)
	}
	if d := Compare(unmarked, marked); d.Breakpoints != 2 {
		t.Errorf("Breakpoints = %d, want 2", d.Breakpoints)
	}
}
