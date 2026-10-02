package common

import (
	"encoding/json"
	"testing"
)

// Reasoning summaries are a new capability, and the risk with a new capability
// is not that it fails to work. It is that it gets switched on for models that
// do not have it, because switching it on everywhere is one character shorter
// than switching it on precisely.

// TestStreamAllExcludesReasoningSummary guards a deliberate omission.
//
// StreamAll is used by roughly twenty rows. Adding StreamReasoningSummary to
// it would be a one-word edit that silently asserted every Anthropic and
// Gemini model emits reasoning summaries, which none of them do. The claim
// would never fail loudly: nothing would emit the delta for those vendors, so
// the table would simply be wrong and no test would notice. This test is the
// thing that notices.
func TestStreamAllExcludesReasoningSummary(t *testing.T) {
	if StreamAll.Has(StreamReasoningSummary) {
		t.Error("StreamAll claims reasoning-summary streaming; it is deliberately excluded " +
			"because no Anthropic or Gemini row was measured emitting summaries")
	}
	for _, s := range []Stream{StreamText, StreamThinking, StreamToolArgs} {
		if !StreamAll.Has(s) {
			t.Errorf("StreamAll lost a kind it used to carry: %v", s)
		}
	}
}

// TestOnlyMeasuredRowsClaimReasoningSummary keeps the table honest.
//
// The table records measurements. A row claims summary streaming only if
// someone watched it stream a summary. That began as the chapter 19 default
// and its grader twin; gpt-6-astra and gpt-5.6-sol joined them after both were
// run against the live ChatGPT-plan route carrying thinking, tools and
// streamed summaries together.
//
// Adding a row here is a claim about the world, so it belongs in the same
// commit as the measurement that justifies it and nowhere else.
func TestOnlyMeasuredRowsClaimReasoningSummary(t *testing.T) {
	want := map[string]bool{
		"gpt-6.1-sol":     true,
		"gpt-ch19-course": true,
		"gpt-6-astra":     true,
		"gpt-5.6-sol":     true,
	}
	for id, f := range models() {
		claims := f.Stream.Has(StreamReasoningSummary)
		if claims && !want[id] {
			t.Errorf("model %q claims StreamReasoningSummary but was never measured doing it", id)
		}
		if !claims && want[id] {
			t.Errorf("model %q was measured streaming summaries but does not claim it", id)
		}
	}
}

// TestResponsesCacheWriteBudget records the documented per-request limit.
//
// Four is not a round number someone liked. It is the number of explicit
// breakpoints OpenAI will write per request, and it is exactly the number of
// breakpoints chapter 18's architecture places.
func TestResponsesCacheWriteBudget(t *testing.T) {
	for _, id := range []string{"gpt-6.1-sol", "gpt-ch19-course"} {
		f, ok := LookupModel(id)
		if !ok {
			t.Fatalf("model %q missing from the table", id)
		}
		if f.MaxCacheWrites != 4 {
			t.Errorf("%s MaxCacheWrites = %d, want 4", id, f.MaxCacheWrites)
		}
		if f.Caching != CacheExplicit {
			t.Errorf("%s Caching = %v, want CacheExplicit: a cache-write budget is "+
				"meaningless unless breakpoints are placed explicitly", id, f.Caching)
		}
	}
}

// TestDefaultModelIsPriced catches the failure where a new default lands with
// an empty pricing row and the cost meter quietly reports nothing.
func TestDefaultModelIsPriced(t *testing.T) {
	f, ok := LookupModel("gpt-6.1-sol")
	if !ok {
		t.Fatal("gpt-6.1-sol missing from the table")
	}
	if f.Price.Input == 0 || f.Price.Output == 0 {
		t.Fatal("gpt-6.1-sol has no price; the usage meter would report every session as unpriced")
	}
	// The two ratios chapter 18 reasoned about, now first-party confirmed:
	// a cached read costs 5% of an uncached one, a cache write 1.25x.
	if got, want := f.Price.CacheRead/f.Price.Input, 0.05; got != want {
		t.Errorf("cache read is %.3fx input, want %.3fx", got, want)
	}
	if got, want := f.Price.CacheWrite/f.Price.Input, 1.25; got != want {
		t.Errorf("cache write is %.3fx input, want %.3fx", got, want)
	}
}

// TestReasoningSummaryKindRoundTrips checks the enum discipline the codebase
// already applies to every other kind: a value survives JSON, and the zero
// value is refused rather than silently meaning "thinking".
func TestReasoningSummaryKindRoundTrips(t *testing.T) {
	b, err := json.Marshal(DeltaReasoningSummary)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(b) != `"reasoning_summary"` {
		t.Errorf("wire spelling = %s, want \"reasoning_summary\"", b)
	}
	var back DeltaKind
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back != DeltaReasoningSummary {
		t.Errorf("round trip produced %v, want DeltaReasoningSummary", back)
	}
	// Appending rather than renumbering matters: these values are written into
	// event logs that a later run replays.
	if DeltaReasoningSummary <= DeltaToolCall {
		t.Error("DeltaReasoningSummary was inserted before an existing kind, " +
			"which renumbers values already written to event logs")
	}
}

// TestKindOKGatesReasoningSummary proves the new bit is actually consulted. A
// delta kind with no gate is a delta kind that leaks from every model.
func TestKindOKGatesReasoningSummary(t *testing.T) {
	if (StreamText | StreamToolArgs).KindOK(DeltaReasoningSummary) {
		t.Error("a model without StreamReasoningSummary accepted a summary delta")
	}
	if !(StreamText | StreamReasoningSummary).KindOK(DeltaReasoningSummary) {
		t.Error("a model with StreamReasoningSummary rejected a summary delta")
	}
}
