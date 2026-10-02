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
