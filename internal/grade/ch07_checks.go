package grade

// Chapter 7 checks — 100 points.
//
// The chapter's claim is that content can be watched as it arrives without
// the seam growing a second shape for it. These checks grade that claim from
// OUTSIDE the module, through the exercise binary's observation stream, so
// nothing here depends on how a student spelled their internals.

import (
	"fmt"
	"sort"
	"strings"
)

// Ch7Evaluate scores a ch7 submission.
func Ch7Evaluate(r *Ch7Result) []Check {
	return []Check{
		ch7StreamDeltas(r),
		ch7DeltasMatchFinal(r),
		ch7ThinkingStreamed(r),
		ch7ToolParamsStreamed(r),
		ch7DeliveryNotContent(r),
		ch7Parity(r),
	}
}

// ch7MinTextChunks is the floor for "actually incremental".
//
// The scripted reply is 49 runes and the fake chunks at 3 runes, so a working
// parser reports about 17 deltas for it. A parser that only handles a whole
// document reports exactly 1. Eight is far enough above 1 to be unambiguous
// and far enough below 17 to survive a student choosing a different chunk
// boundary, which is not the thing being graded.
const ch7MinTextChunks = 8

func ch7Ready(c *Check, r *Ch7Result) bool {
	if !r.ExBuildOK {
		c.failf("exercise did not build: %s", r.ExBuildErr)
		return false
	}
	if !r.Stream.Ran {
		c.failf("exercise did not run")
		return false
	}
	if r.Stream.ExitErr != "" {
		c.failf("exercise exited with an error: %s", r.Stream.ExitErr)
		if s := r.Stream.Stderr; s != "" {
			c.notef("stderr: %s", firstLines(s, 5))
		}
		return false
	}
	return true
}

// stream-deltas: content arrives in pieces, not in one lump.
func ch7StreamDeltas(r *Ch7Result) Check {
	c := Check{ID: "stream-deltas", Title: "content streams as many deltas", Points: 20, Passed: true, Earned: 20}
	if !ch7Ready(&c, r) {
		return c
	}

	n := r.Stream.KindCount("text")
	if n < ch7MinTextChunks {
		c.failf("saw %d text deltas, want at least %d", n, ch7MinTextChunks)
		c.notef("a response delivered as one chunk reports exactly 1; that is the failure this check exists to catch")
		if n == 0 {
			c.notef("zero text deltas usually means the request never asked to stream, or the response content type was not checked")
		}
		return c
	}
	c.notef("saw %d text deltas across %d parts", n, len(r.Stream.TextByPart()))
	return c
}

// deltas-match-final: the pieces reassemble to the authoritative part.
//
// Grouped BY PART, which is the whole reason a delta carries a part id. A
// grader that concatenated every text delta in the turn and compared it to
// the last assistant message would pass while the ids were pure noise, and
// would keep passing if a student minted a fresh id per chunk.
func ch7DeltasMatchFinal(r *Ch7Result) Check {
	c := Check{ID: "deltas-match-final", Title: "deltas reassemble into their finalized part", Points: 20, Passed: true, Earned: 20}
	if !ch7Ready(&c, r) {
		return c
	}

	byPart := r.Stream.TextByPart()
	finals := r.Stream.FinalText
	if len(finals) == 0 {
		c.failf("no text part_final observations were reported")
		c.notef("a stream needs a final: deltas are for watching, the finalized part is the authority a client re-renders from")
		return c
	}

	ids := make([]uint64, 0, len(finals))
	for id := range finals {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	matched, multi := 0, 0
	for _, id := range ids {
		want := finals[id]
		got, ok := byPart[id]
		if !ok {
			c.failf("part %d finalized with text but no delta ever carried part_id %d", id, id)
			c.notef("the id on a delta must be the id on its final; if they are allocated separately they cannot correlate")
			return c
		}
		if got != want {
			c.failf("part %d: deltas concatenate to %q but the final says %q", id, got, want)
			return c
		}
		matched++
		if countDeltasFor(r.Stream, id) > 1 {
			multi++
		}
	}

	if multi == 0 {
		c.failf("every text part arrived in a single delta, so reassembly was never exercised")
		return c
	}
	c.notef("%d text parts reassembled exactly, %d of them from multiple deltas", matched, multi)
	return c
}

func countDeltasFor(run Ch7Exec, id uint64) int {
	n := 0
	for _, d := range run.Deltas {
		if d.PartID == id && d.Kind == "text" {
			n++
		}
	}
	return n
}

// thinking-streamed: reasoning is visible, and distinguishable from the reply.
func ch7ThinkingStreamed(r *Ch7Result) Check {
	c := Check{ID: "thinking-streamed", Title: "reasoning streams under its own kind", Points: 15, Passed: true, Earned: 15}
	if !ch7Ready(&c, r) {
		return c
	}

	n := r.Stream.KindCount("thinking")
	if n < 2 {
		c.failf("saw %d deltas of kind \"thinking\", want at least 2", n)
		if n == 0 {
			c.notef("the scripted response contains a reasoning block; reporting none means thinking deltas are dropped or labelled as text")
			c.notef("labelling reasoning as text is worse than dropping it: it lands in the answer")
		} else {
			c.notef("one delta means the block was reported whole rather than as it arrived")
		}
		return c
	}

	// Reasoning must not leak into the reply text.
	//
	// Compared as ONE window of the reassembled block, never chunk by chunk.
	// A 3-rune chunk is a common word: the first version of this check
	// compared every chunk against the reply and failed the reference
	// solution because "The" appears in both. A check that cannot pass a
	// correct implementation is worse than no check.
	const leakWindow = 24
	joined := ""
	for _, d := range r.Stream.Deltas {
		if d.Kind == "thinking" {
			joined += d.Chunk
		}
	}
	if len([]rune(joined)) >= leakWindow {
		window := string([]rune(joined)[:leakWindow])
		for id, final := range r.Stream.FinalText {
			if strings.Contains(final, window) {
				c.failf("reasoning leaked into finalized reply text of part %d: %q", id, window)
				c.notef("labelling reasoning as text is worse than dropping it, because it lands in the answer")
				return c
			}
		}
	}
	c.notef("saw %d thinking deltas, none leaking into reply text", n)
	return c
}

// tool-params-streamed: a tool call is spelled out as it is decided.
func ch7ToolParamsStreamed(r *Ch7Result) Check {
	c := Check{ID: "tool-params-streamed", Title: "tool calls stream name and arguments", Points: 15, Passed: true, Earned: 15}
	if !ch7Ready(&c, r) {
		return c
	}

	n := r.Stream.KindCount("tool_call")
	if n < 3 {
		c.failf("saw %d deltas of kind \"tool_call\", want at least 3 (a name and several argument fragments)", n)
		if n == 0 {
			c.notef("the scripted response calls a tool; reporting no tool_call deltas means the arguments were never surfaced")
		}
		return c
	}
	if len(r.Stream.FinalTool) == 0 {
		c.failf("tool_call deltas were reported but no tool part was ever finalized")
		c.notef("arguments are incomplete JSON until the part finalizes, so the final is the only thing safe to act on")
		return c
	}

	var joined string
	for _, d := range r.Stream.Deltas {
		if d.Kind == "tool_call" {
			joined += d.Chunk
		}
	}
	for _, name := range r.Stream.FinalTool {
		if !strings.Contains(joined, name) {
			c.failf("finalized tool %q never appeared in the tool_call deltas", name)
			return c
		}
	}
	c.notef("saw %d tool_call deltas reassembling to %q", n, truncate(joined, 80))
	return c
}

// delivery-not-content: streaming changes the chunk count and nothing else.
//
// This is the chapter's thesis stated as a check. A scripted reply exercised
// twice — once streamed, once not — must produce identical finalized content
// and a different number of deltas. "The response is the same response
// whether it trickles or lands."
func ch7DeliveryNotContent(r *Ch7Result) Check {
	c := Check{ID: "delivery-not-content", Title: "streaming changes delivery, not content", Points: 10, Passed: true, Earned: 10}
	if !ch7Ready(&c, r) {
		return c
	}

	// The plain run must have completed.
	if !r.Plain.Ran {
		c.failf("plain (non-streaming) run did not execute")
		return c
	}
	if r.Plain.ExitErr != "" {
		c.failf("plain run exited with an error: %s", r.Plain.ExitErr)
		return c
	}

	// 1. Finalized reply text must be identical, byte for byte.
	if r.Stream.Assistant != r.Plain.Assistant {
		c.failf("assistant text differs between streaming and plain runs")
		c.notef("streaming: %s", truncate(r.Stream.Assistant, 120))
		c.notef("plain:     %s", truncate(r.Plain.Assistant, 120))
		return c
	}
	if r.Stream.Assistant == "" {
		c.failf("neither run produced an assistant line")
		return c
	}

	// 2. Same number of finalized parts, same kinds in order.
	if len(r.Stream.Finals) != len(r.Plain.Finals) {
		c.failf("streaming produced %d part finals, plain produced %d", len(r.Stream.Finals), len(r.Plain.Finals))
		return c
	}
	for i := range r.Stream.Finals {
		sk := r.Stream.Finals[i].Kind
		pk := r.Plain.Finals[i].Kind
		if sk != pk {
			c.failf("part final %d: streaming kind %q, plain kind %q", i, sk, pk)
			return c
		}
	}

	// 3. Delta counts must differ — that is what streaming is FOR.
	sn := len(r.Stream.Deltas)
	pn := len(r.Plain.Deltas)
	if sn == pn {
		c.failf("streaming and plain runs both produced %d deltas; the flag changed nothing", sn)
		return c
	}

	// Also verify the plain run produced sensible deltas (moved from stream-deltas).
	pt := r.Plain.KindCount("text")
	st := r.Stream.KindCount("text")
	if pt == 0 {
		c.failf("with streaming disabled the agent emitted NO text deltas")
		c.notef("a non-streamed response is a stream of length one; it still reports one delta per part, or every observer breaks when streaming is turned off")
		return c
	}
	if pt >= st {
		c.failf("streaming disabled produced %d text deltas, streaming produced %d: the flag changed nothing", pt, st)
		return c
	}

	c.notef("same text (%d bytes), same %d part kinds, different delta counts (%d streaming vs %d plain)",
		len(r.Stream.Assistant), len(r.Stream.Finals), sn, pn)
	return c
}

// ch6-parity: everything chapter 6 promised still holds.
func ch7Parity(r *Ch7Result) Check {
	c := Check{ID: "ch6-parity", Title: "chapter 6 behavior is unchanged", Points: 20, Passed: true, Earned: 20}
	if r.Ch6Err != "" {
		c.failf("ch6 harness did not run: %s", r.Ch6Err)
		return c
	}
	if r.Ch6Result == nil {
		c.failf("ch6 harness produced no result")
		return c
	}
	var failed []string
	for _, sub := range Ch6Evaluate(r.Ch6Result) {
		if !sub.Passed {
			failed = append(failed, fmt.Sprintf("%s (%s)", sub.ID, firstOr(sub.Details, "no detail")))
		}
	}
	if len(failed) > 0 {
		c.failf("%d ch6 check(s) now fail: %v", len(failed), failed)
		return c
	}
	c.notef("all ch6 checks still pass")
	return c
}

func firstOr(list []string, def string) string {
	if len(list) == 0 {
		return def
	}
	return list[0]
}

func firstLines(s string, n int) string {
	var kept []string
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if len(kept) >= n {
			break
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, " | ")
}
