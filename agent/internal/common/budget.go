package common

// Chapter 15 rule 10: one knob.
//
// The user chooses a target context size in bytes. Everything the context
// policy needs is derived from it by one formula, so the settings tab has
// one number to explain and the chapter has one line to print:
//
//	stub threshold = T/100    results band = T/8    calls band = T/16
//
// A band is cut back to its budget when it passes twice its budget, so at
// its worst the tool bytes in the window are 2(T/8 + T/16) = 3T/8, and the
// remaining five-eighths are the prefix, the survivors and the dialogue,
// which is what the chapter says to keep.

// DefaultContextTarget is used when settings say nothing: 400,000 bytes, about
// 100K tokens at four bytes a token.
const DefaultContextTarget = 400_000

// MinContextTarget stops a typo from stubbing every result: below this the
// stub threshold would fall under the size of a directory listing.
const MinContextTarget = 20_000

// KeepToolResults is the tool the policy looks for in the model's response.
const KeepToolResults = "keep_tool_results"

// Budgets are the derived numbers.
type Budgets struct {
	Target    int // T, bytes
	Threshold int // a result bigger than this is stubbed per round trip
	Results   int // full tool-result bytes kept at the tail
	Calls     int // call-argument and stub bytes kept above the results band
}

// BudgetsFor derives the budgets from a target. Zero means the default.
func BudgetsFor(target int) Budgets {
	if target <= 0 {
		target = DefaultContextTarget
	}
	if target < MinContextTarget {
		target = MinContextTarget
	}
	return Budgets{Target: target, Threshold: target / 100, Results: target / 8, Calls: target / 16}
}
