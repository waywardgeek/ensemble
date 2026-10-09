package common

import "testing"

// opus5 is a real price sheet, used so the arithmetic in these tests is the
// arithmetic that will actually bill Bill.
var opus5 = Pricing{Input: 5, CacheWrite: 6.25, CacheRead: 0.5, Output: 25}

func TestCostUSDSumsAllFourCategories(t *testing.T) {
	u := Usage{Input: 1000, CacheWrite: 2000, CacheRead: 100000, Output: 500}

	// 1000*5 + 2000*6.25 + 100000*0.5 + 500*25
	//  = 5000 +     12500 +      50000 +   12500 = 80000 per million.
	const want = 0.08

	got := CostUSD(u, opus5)
	if got != want {
		t.Fatalf("CostUSD = %v, want %v", got, want)
	}
}

// TestCostUSDChargesCacheReadAtItsOwnRate states the economic claim the whole
// caching feature rests on: the same token count costs an order of magnitude
// less when it arrives from cache than when it arrives fresh. If these two
// numbers are ever equal, either the rates have been conflated or caching has
// stopped being worth doing.
func TestCostUSDChargesCacheReadAtItsOwnRate(t *testing.T) {
	const tokens = 1000000

	fresh := CostUSD(Usage{Input: tokens}, opus5)
	cached := CostUSD(Usage{CacheRead: tokens}, opus5)

	if fresh != 5.0 {
		t.Errorf("one million fresh input tokens = %v, want 5.0", fresh)
	}
	if cached != 0.5 {
		t.Errorf("one million cached input tokens = %v, want 0.5", cached)
	}
	if cached >= fresh {
		t.Fatalf("cached reads (%v) are not cheaper than fresh input (%v)", cached, fresh)
	}
}

// TestCacheHitRateExcludesOutput pins a design decision that is easy to
// quietly undo. Output tokens are not an input category and no cache could
// ever have supplied them, so they must not sit in the denominator. If they
// did, the hit rate would sag whenever the model simply said more, and a
// verbose turn would look like a caching regression.
func TestCacheHitRateExcludesOutput(t *testing.T) {
	terse := Usage{Input: 1000, CacheWrite: 1000, CacheRead: 8000, Output: 100}
	verbose := Usage{Input: 1000, CacheWrite: 1000, CacheRead: 8000, Output: 100000}

	if got := CacheHitRate(terse); got != 0.8 {
		t.Errorf("terse turn hit rate = %v, want 0.8", got)
	}
	if CacheHitRate(terse) != CacheHitRate(verbose) {
		t.Fatalf("hit rate changed with output length: %v vs %v (output must not be in the denominator)",
			CacheHitRate(terse), CacheHitRate(verbose))
	}
}

// TestCacheHitRateEmptySessionIsZero guards the display. A fresh session has
// sent nothing, and a naive ratio would be NaN, which renders as "NaN%" in the
// status bar the moment the GUI connects.
func TestCacheHitRateEmptySessionIsZero(t *testing.T) {
	if got := CacheHitRate(Usage{}); got != 0 {
		t.Fatalf("CacheHitRate of an empty session = %v, want 0", got)
	}
}

func TestZeroPriceSheetIsUnpriced(t *testing.T) {
	if (Pricing{}).Priced() {
		t.Error("the zero sheet reports itself as priced")
	}
	if !opus5.Priced() {
		t.Error("a real sheet reports itself as unpriced")
	}
}

// TestEveryOfferedModelIsPriced is the guard that matters when someone adds a
// model later. ModelList is exactly the set the GUI offers in its dropdown, so
// an unpriced entry there is a model a user can select and then be shown a
// confident, wrong "$0.00" for.
func TestEveryOfferedModelIsPriced(t *testing.T) {
	for _, m := range ModelList() {
		if !m.Features.Price.Priced() {
			t.Errorf("model %q is offered in the GUI but has no price sheet", m.ID)
		}
	}
}

// Per-model pricing. These pin the bug Chapter 22 fixes: the GUI used to
// price the whole session at the currently selected model's rate, so
// switching models retroactively re-priced every token already spent.
//
// Assertions are written as relationships between CostUSD calls rather than
// as hardcoded dollar figures, so they state the property rather than a
// units convention, and they do not break when a price in the model table is
// updated.

// closeEnough compares dollar amounts. CostByModel sums over a map, so the
// order of addition is not fixed between runs, and float addition is not
// associative -- an exact comparison here would flake rather than fail.
func closeEnough(a, b float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < 1e-9
}

func TestCostByModelSumsEachModelAtItsOwnRate(t *testing.T) {
	const cheap, dear = "gemini-3.8-flash", "gpt-6-astra"

	cheapF, ok := LookupModel(cheap)
	if !ok {
		t.Fatalf("%s missing from the model table", cheap)
	}
	dearF, ok := LookupModel(dear)
	if !ok {
		t.Fatalf("%s missing from the model table", dear)
	}

	cheapUse := Usage{Input: 1_000_000, Output: 200_000}
	dearUse := Usage{Input: 50_000, Output: 10_000}

	got, complete := CostByModel(map[string]Usage{cheap: cheapUse, dear: dearUse})
	if !complete {
		t.Fatal("both models are priced, so the tally should be complete")
	}

	want := CostUSD(cheapUse, cheapF.Price) + CostUSD(dearUse, dearF.Price)
	if !closeEnough(got, want) {
		t.Errorf("cost = %v, want sum of per-model costs %v", got, want)
	}

	// The bug, stated as an assertion. The old code multiplied the session
	// total by the current model's price. The great majority of these tokens
	// were spent on the cheap model, so billing them all at the dear model's
	// rate inflates the figure -- and it would have changed the moment the
	// operator switched models, with no request having been sent.
	session := Usage{
		Input:  cheapUse.Input + dearUse.Input,
		Output: cheapUse.Output + dearUse.Output,
	}
	buggy := CostUSD(session, dearF.Price)
	if buggy == got {
		t.Fatal("fixture is useless: the two pricing methods agree, so this " +
			"test could not have caught the bug it exists for")
	}
	if buggy < got {
		t.Errorf("expected the old method to over-bill here, got %v vs %v", buggy, got)
	}
}

// Switching models must not change the price of tokens already spent. This is
// the user-visible symptom: the session cost jumped when you picked a
// different model, without sending anything.
func TestCostByModelIsStableAcrossAModelSwitch(t *testing.T) {
	spent := map[string]Usage{
		"gemini-3.8-flash": {Input: 500_000, Output: 100_000},
	}
	before, _ := CostByModel(spent)

	// The operator switches to an expensive model and sends nothing. The
	// per-model tally is untouched, so the cost must be untouched.
	after, _ := CostByModel(spent)
	if before != after {
		t.Errorf("cost changed from %v to %v without any tokens being spent", before, after)
	}

	// Now one request goes to the expensive model. Cost rises by exactly that
	// request's cost at that model's rate, and not by a re-pricing of history.
	dearF, ok := LookupModel("gpt-6-astra")
	if !ok {
		t.Fatal("gpt-6-astra missing from the model table")
	}
	dearUse := Usage{Input: 1_000, Output: 500}
	spent["gpt-6-astra"] = dearUse

	grown, _ := CostByModel(spent)
	if delta, want := grown-before, CostUSD(dearUse, dearF.Price); !closeEnough(delta, want) {
		t.Errorf("cost grew by %v, want exactly the new request's cost %v", delta, want)
	}
}

// A missing price is not a price of zero. The distinction matters precisely
// when a model has been added and nobody has filled in its row yet: silently
// billing it at nothing reads as a working meter.
func TestCostByModelReportsIncompleteForUnpricedModels(t *testing.T) {
	if f, ok := LookupModel("gpt-ch19-course"); !ok {
		t.Fatal("gpt-ch19-course missing from the model table")
	} else if f.Price.Priced() {
		t.Skip("gpt-ch19-course has gained a price; pick another unpriced model")
	}

	_, complete := CostByModel(map[string]Usage{
		"gpt-ch19-course": {Input: 1_000, Output: 100},
	})
	if complete {
		t.Error("a model with no price must mark the tally incomplete")
	}

	// A model that was selected but never sent to contributes nothing and
	// must not make the tally incomplete -- otherwise merely opening the
	// dropdown would blank the meter.
	if _, complete := CostByModel(map[string]Usage{"gpt-ch19-course": {}}); !complete {
		t.Error("a model with zero usage must not mark the tally incomplete")
	}
}

// The counter must attribute each request to the model that served it, which
// is the write half of the same property.
func TestUsageCounterSplitsByModel(t *testing.T) {
	var c UsageCounter
	c.RecordUsage("model-a", Usage{Input: 10, Output: 1})
	c.RecordUsage("model-b", Usage{Input: 200, Output: 20})
	c.RecordUsage("model-a", Usage{Input: 5, Output: 2})

	by := c.UsageByModel()
	if got, want := (by["model-a"]), (Usage{Input: 15, Output: 3}); got != want {
		t.Errorf("model-a = %+v, want %+v", got, want)
	}
	if got, want := (by["model-b"]), (Usage{Input: 200, Output: 20}); got != want {
		t.Errorf("model-b = %+v, want %+v", got, want)
	}
	if got, want := c.SessionUsage(), (Usage{Input: 215, Output: 23}); got != want {
		t.Errorf("session = %+v, want %+v", got, want)
	}
	if got, want := c.LastUsage(), (Usage{Input: 5, Output: 2}); got != want {
		t.Errorf("last = %+v, want %+v", got, want)
	}

	// The returned map is a copy: pricing walks it while the engine may still
	// be recording into it.
	by["model-a"] = Usage{Input: 9999}
	if c.UsageByModel()["model-a"].Input != 15 {
		t.Error("UsageByModel handed out a live reference to internal state")
	}
}
