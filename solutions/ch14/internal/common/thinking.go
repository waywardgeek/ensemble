package common

// Thinking effort — vendor-neutral vocabulary for reasoning configuration.
//
// The same shape as Stream/StreamingFor, and deliberately so: a capability
// that is data about a model, resolved against a Config, with a safe default
// for an unknown model.

// ThinkingEffort controls how much reasoning budget to request.
//
// iota+1 so the zero value is INVALID/unset, per the house rule. An unset
// ThinkingEffort resolves to ThinkingHigh in ThinkingFor, because that is
// the default Bill wants — and a hand-built Config{} must get it. Same
// trap DisableStreaming was designed around.
type ThinkingEffort uint8

const (
	ThinkingOff    ThinkingEffort = iota + 1 // 1
	ThinkingLow                              // 2
	ThinkingMedium                           // 3
	ThinkingHigh                             // 4
)

var thinkingEffortNames = map[ThinkingEffort]string{
	ThinkingOff:    "off",
	ThinkingLow:    "low",
	ThinkingMedium: "medium",
	ThinkingHigh:   "high",
}

func (t ThinkingEffort) String() string {
	if s, ok := thinkingEffortNames[t]; ok {
		return s
	}
	return "invalid"
}

// ThinkingFor resolves the thinking configuration for a request.
//
// Returns the effective effort and the budget in tokens. Zero budget means
// "do not include a thinking block in the request".
//
// AN UNKNOWN MODEL GETS THINKING OFF, AND THAT IS NOT AN ERROR. Consistent
// with StreamingFor and for the same open-set reason: model names are an
// open set, new ones appear weekly, and a framework that refuses to work
// until its table has heard of your model is unusable. Not streaming and not
// thinking are both safe: the same response arrives, just without extras.
//
// Budget per effort, floored at Anthropic's 1024 minimum:
//
//	high   = ceiling
//	medium = max(1024, ceiling/4)
//	low    = 1024
func ThinkingFor(cfg Config) (ThinkingEffort, int) {
	effort := cfg.Thinking
	if effort == 0 {
		// Zero (unset) resolves to HIGH. The zero value picks the default.
		effort = ThinkingHigh
	}
	if effort == ThinkingOff {
		return ThinkingOff, 0
	}

	features, ok := LookupModel(cfg.Model)
	if !ok {
		// Unknown model: safe fallback, same as StreamingFor.
		return ThinkingOff, 0
	}
	ceiling := features.MaxThinkingTokens
	if ceiling == 0 {
		// Model is known but has no thinking budget listed: it does not
		// support thinking.
		return ThinkingOff, 0
	}

	var budget int
	switch effort {
	case ThinkingHigh:
		budget = ceiling
	case ThinkingMedium:
		budget = ceiling / 4
		if budget < 1024 {
			budget = 1024
		}
	case ThinkingLow:
		budget = 1024
	}

	return effort, budget
}

// EnsureMaxTokens enforces the invariant that max_tokens must strictly
// exceed budget_tokens. Anthropic rejects the request otherwise.
//
// Rather than erroring, raise max_tokens to budget + a reply allowance.
// This is called ONCE in the render path, not per-vendor — one place to
// enforce, one place to break.
const thinkingReplyAllowance = 16384

func EnsureMaxTokens(maxTokens, budgetTokens int) int {
	if budgetTokens <= 0 {
		return maxTokens
	}
	floor := budgetTokens + thinkingReplyAllowance
	if maxTokens < floor {
		return floor
	}
	return maxTokens
}
