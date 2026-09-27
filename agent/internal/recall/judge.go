package recall

// The judge, from recall's side of the interface.
//
// Everything here is pure text manipulation: build a prompt, read a reply.
// The model call itself is somebody else's problem — it arrives as a
// common.SnippetJudge — which is what lets this package stay a spoke, and
// what lets the whole judging path be tested without a network.
//
// The judge is a single stateless call, not an agent. That is not a
// simplification of a grander design; it is the design. An agent that could
// recall would trigger its own recall while judging, and that judgment would
// invoke another judge. A function from prompt to text has no context of its
// own to fill, so the failure cannot be expressed. Building a thing that
// cannot exhibit a failure mode beats guarding against it: a guard is one
// refactor from being removed, and a missing capability is not.

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// judgeInstruction is the whole of the judge's job description.
//
// The last line does the heavy lifting. Without it a small model reliably
// returns every candidate, because every candidate does mention the query
// words — that is how BM25 chose them. The judge is only worth its cost if
// it is told that keyword overlap is the thing it is being asked to look
// past.
const judgeInstruction = `You decide which remembered notes an AI coding agent should be shown right now.

Below are the recent conversation, the latest user message, and numbered
candidate snippets retrieved by keyword search.

Pick the snippets the agent would GENUINELY benefit from seeing to answer the
latest message. Keyword mention alone is not relevance: a snippet that merely
contains the same words, but is about a different subject, must be rejected.

Reply with ONLY a JSON array of the indices you picked, most useful first.
Examples: [0, 7, 12] or [3] or []
Pick at most %d. If nothing is genuinely relevant, reply [].`

// judgePrompt assembles the text sent to the judge.
//
// The recent conversation reaches the judge as DATA INSIDE THIS STRING, not
// as history the judge carries between calls. That is the distinction that
// makes the judge stateless: it sees what this prompt shows it and has no
// memory of the last snippet it picked, so every judgment is independent and
// nothing from a previous one can colour the next.
func (r *Recaller) judgePrompt(query string, convo []string, cands []Hit) string {
	var b strings.Builder
	fmt.Fprintf(&b, judgeInstruction, r.cfg.MaxSnippets)
	b.WriteString("\n\n")

	if tail := lastN(convo, r.cfg.ContextMsgs); len(tail) > 0 {
		b.WriteString("RECENT CONVERSATION:\n")
		for _, m := range tail {
			b.WriteString(truncate(strings.TrimSpace(m), 600))
			b.WriteString("\n---\n")
		}
		b.WriteString("\n")
	}

	b.WriteString("LATEST USER MESSAGE:\n")
	b.WriteString(truncate(strings.TrimSpace(query), 2000))
	b.WriteString("\n\nCANDIDATE SNIPPETS:\n")
	for i, c := range cands {
		fmt.Fprintf(&b, "[%d] %s\n%s\n\n", i, attribution(c), truncate(c.Chunk.Content, 900))
	}
	return b.String()
}

// lastN returns the final n elements, or all of them if there are fewer.
func lastN(msgs []string, n int) []string {
	if n <= 0 || len(msgs) == 0 {
		return nil
	}
	if len(msgs) <= n {
		return msgs
	}
	return msgs[len(msgs)-n:]
}

// truncate caps a string at n bytes on a rune boundary.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	cut := n
	for cut > 0 && !utf8Start(s[cut]) {
		cut--
	}
	return s[:cut] + "..."
}

// utf8Start reports whether a byte can begin a UTF-8 rune, so truncation
// never leaves half a character behind.
func utf8Start(b byte) bool { return b&0xC0 != 0x80 }

// parseIndices reads a judge reply into candidate indices.
//
// Small models are unreliable with output format in ways that are individually
// silly and collectively guaranteed: they quote the numbers, they mix quoted
// and unquoted, and they wrap the array in a sentence explaining what they
// did. Each fallback below is one of those, observed rather than imagined.
//
// The second return value separates "the judge said nothing is relevant" from
// "the judge produced something I cannot read". Both yield no indices, and
// they must lead to opposite behaviour: an empty array is a judgment and is
// obeyed, whereas unreadable output is a failure and falls back to BM25.
// Collapsing the two would make a broken judge look like a strict one, and
// recall would silently switch itself off.
func parseIndices(raw string, n int) ([]int, bool) {
	// Models wrap JSON in prose. Take the outermost bracketed span.
	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start < 0 || end < start {
		return nil, false
	}
	body := raw[start : end+1]

	// The common case: a clean array of numbers.
	var nums []int
	if err := json.Unmarshal([]byte(body), &nums); err == nil {
		return clampIndices(nums, n), true
	}

	// Quoted numbers: ["0", "3"].
	var strs []string
	if err := json.Unmarshal([]byte(body), &strs); err == nil {
		out := make([]int, 0, len(strs))
		for _, s := range strs {
			if v, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
				out = append(out, v)
			}
		}
		return clampIndices(out, n), true
	}

	// Mixed types: [0, "3", 7]. Decode element by element so one bad entry
	// costs one index rather than the whole answer.
	var raws []json.RawMessage
	if err := json.Unmarshal([]byte(body), &raws); err == nil {
		out := make([]int, 0, len(raws))
		for _, m := range raws {
			var v int
			if err := json.Unmarshal(m, &v); err == nil {
				out = append(out, v)
				continue
			}
			var s string
			if err := json.Unmarshal(m, &s); err == nil {
				if v, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
					out = append(out, v)
				}
			}
		}
		return clampIndices(out, n), true
	}

	return nil, false
}

// clampIndices drops out-of-range and duplicate indices.
//
// Silently, because a hallucinated index 47 against 12 candidates is not an
// error worth surfacing — it is a small model being a small model, and the
// eleven good indices around it are still good.
func clampIndices(in []int, n int) []int {
	seen := make(map[int]bool, len(in))
	out := make([]int, 0, len(in))
	for _, v := range in {
		if v < 0 || v >= n || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}
