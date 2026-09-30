package llm

// Cache bounds — where a breakpoint can safely sit.
//
// Placing a marker is easy. Proving that every byte before it will be
// reproduced exactly by later requests is the whole problem, and the answer is
// a property of how compaction works rather than of any vendor's wire format.
// That is why this lives here and not in claude.go or openai.go.

import "github.com/waywardgeek/ensemble/agent/internal/common"

// newestHandoffIndex returns the index in dialogue of the newest handoff
// entry, or -1 when the conversation has never been compacted.
//
// WHY THIS IS A SAFE CACHE BOUND, AND WHY THE ANSWER IS NOT OBVIOUS.
//
// A micro_handoff rewrites the history behind it, so the request that performs
// one always misses. This is about every request after that one. The question
// is whether the rewritten region stays put, and here it does, for two reasons
// that are both easy to lose in a refactor:
//
//   - The strip is TOTAL. land() removes every tool call and every tool result
//     from every dialogue entry, along with every recall entry. Nothing is held
//     back, so nothing is left for a later pass to remove.
//   - The strip is applied ONCE, by the reducer, and frozen into the
//     projection. It is not recomputed from the tail on every render.
//
// Together those put the region below the newest handoff in the TERMINAL state
// of the redaction ladder in policy.go: dialogue only. No later cut can move
// it, so a breakpoint there survives arbitrarily many further compactions.
//
// Neither property follows from the idea of a handoff. An agent that strips at
// render time, and keeps a few recent tool pairs behind the boundary as
// evidence that tools exist, has neither: the kept pairs are counted back from
// a boundary that moves to the newest handoff on every render, so the prefix
// ending at an older handoff is rewritten the moment a newer one appears and
// its entry dies. Such an agent must use the first surviving tool call as its
// bound instead. Same rule, different answer, because the surrounding design
// differs.
//
// If a later change ever keeps tool traffic behind a handoff, this bound
// becomes wrong and the only symptom is the invoice. The test asserts the
// property rather than the placement, so it fails first.
func newestHandoffIndex(dialogue []common.Entry) int {
	for i := len(dialogue) - 1; i >= 0; i-- {
		if dialogue[i].Kind == common.KindHandoff {
			return i
		}
	}
	return -1
}
