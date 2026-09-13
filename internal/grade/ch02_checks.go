package grade

// Chapter 2 checks. Sum = 100.
//
//	session      0   stdio protocol honoured; directives acked; request census
//	ch1parity   20   all seven Chapter 1 checks still pass, unchanged
//	logdump      5   log round-trips: dump -> render in a fresh process
//	replay      10   two renders of one log are byte-identical
//	redaction   10   a Redacted event names its target; content absent later
//	ephemera    10   delivered exactly once, then absent
//	usage       10   four token categories normalized from all three vendors
//	seam-render 15   one log renders correctly to all three request shapes
//	seam-parse  15   three responses -> contexts identical apart from provenance
//
//	ref-roundtrip  \
//	ref-oldformat   >  ch2RefSerialization   a blob's location is a Ref, and a
//	ref-zerokind   /                         Ref never filled in is refused
//	ref-render         ch2RefRender          RefURI -> the vendor's remote form
//	ref-redaction      ch2RefRedaction       a stub keeps the superseded Ref
//
// The five ref-* IDs are itemized for DIAGNOSIS, not for weight. They cover
// three skills, and points are split only where the skills separate: a student
// who cannot serialize a Ref has one bug, not three, and should be told which
// of the three symptoms fired without being charged three times for it. The
// three serialization IDs therefore SHARE one budget.

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Point budget for the Ref amendment.
//
// PENDING AUTHOR RULING. The instruction was "take the 15 from ch1parity
// (25->20)", which is two different numbers: 25->20 frees 5, not 15. Both
// readings are encoded here, because the total is machine-checked against 100
// by TestCh2ReferenceSolutionScores100 and cannot be split.
//
// Currently set to the 15-point reading. To take the 5-point reading instead,
// change these four constants to 3 / 1 / 1 / 20 — nothing else moves.
const (
	ch2RefSerialization = 7 // shared by ref-roundtrip, ref-oldformat, ref-zerokind
	ch2RefRender        = 4
	ch2RefRedaction     = 4
	ch2ParityPoints     = 10
)

// The three serialization IDs SHARE one budget: losing the skill entirely costs
// ch2RefSerialization, not three times it. The budget is split across them
// rather than parked on one, because a zero-point check cannot satisfy course
// policy P9 — deleting the behavior it guards would fire the check without
// moving the score, which is precisely the green dashboard P9 exists to catch.
const (
	ch2RefRoundTrip = 3
	ch2RefOldFormat = 2
	ch2RefZeroKind  = 2
)

func Ch2Evaluate(r *Ch2Result) []Check {
	return []Check{
		ch2Session(r),
		ch2Parity(r),
		ch2LogDump(r),
		ch2Replay(r),
		ch2Redaction(r),
		ch2Ephemera(r),
		ch2UsageCheck(r),
		ch2SeamRender(r),
		ch2SeamParse(r),
		ch2RefRoundTripCheck(r),
		ch2RefOldFormatCheck(r),
		ch2RefZeroKindCheck(r),
		ch2RefRenderCheck(r),
		ch2RefRedactionCheck(r),
	}
}

// ch2Session is worth zero and can still sink a submission. Without it, one
// unacknowledged directive fails four checks at once and the student gets four
// mysteries instead of one cause.
func ch2Session(r *Ch2Result) Check {
	c := Check{ID: "session", Title: "stdio protocol, directives, request census", Points: 0, Passed: true}

	census := []string{}
	for _, v := range Ch2Vendors {
		s := r.Session[v]
		if s == nil {
			c.failf("no session was recorded for %s", v)
			continue
		}
		census = append(census, fmt.Sprintf("%s:%d requests, %d answers", v, len(s.Requests), len(s.Answers)))
		for _, p := range s.Protocol {
			if strings.HasPrefix(p, "ack:") {
				continue
			}
			c.failf("%s: %s", v, p)
		}
		for _, e := range s.Extra {
			c.failf("%s: unrecognized stdout line %.60q", v, e)
		}
	}
	if e := r.Ephemera; e != nil {
		census = append(census, fmt.Sprintf("ephemera:%d requests", len(e.Requests)))
		acked := false
		for _, p := range e.Protocol {
			if p == "ack:ephemeral" {
				acked = true
			}
		}
		if !acked {
			c.failf("the {\"ephemeral\":...} directive was never acknowledged with {\"ack\":\"ephemeral\"}")
		}
	}
	c.Details = append(c.Details, "census: "+strings.Join(census, "; "))
	return c
}

func ch2Parity(r *Ch2Result) Check {
	c := Check{ID: "ch1parity", Title: "all seven Chapter 1 checks still pass", Points: ch2ParityPoints, Passed: true, Earned: ch2ParityPoints}
	if r.Ch1Err != "" {
		c.failf("Chapter 1 harness could not run: %s", r.Ch1Err)
		return c
	}
	if len(r.Ch1) == 0 {
		c.failf("Chapter 1 harness produced no checks")
		return c
	}
	for _, sub := range r.Ch1 {
		if !sub.Passed {
			c.failf("Chapter 1 check %q now fails: %s", sub.ID, strings.Join(sub.Details, "; "))
		}
	}
	if c.Passed {
		c.Details = append(c.Details, fmt.Sprintf("%d Chapter 1 checks still pass", len(r.Ch1)))
	}
	return c
}

func ch2LogDump(r *Ch2Result) Check {
	c := Check{ID: "logdump", Title: "log round-trips: dump -> render in a fresh process", Points: 5, Passed: true, Earned: 5}
	s := r.Session["anthropic"]
	if s == nil {
		c.failf("no anthropic session")
		return c
	}
	if strings.TrimSpace(s.DumpOut) == "" {
		c.failf("`ch02 dump` produced nothing on stdout (stderr: %.200s)", s.DumpErr)
		return c
	}
	if s.LogErr != "" {
		c.failf("dumped log is not valid JSON-lines: %s", s.LogErr)
		return c
	}
	if len(s.Log) == 0 {
		c.failf("dumped log contained no events")
		return c
	}
	// Seq must be present and ascending: ordering is primary.
	prev := -1
	for _, l := range s.Log {
		if l.Seq <= prev {
			c.failf("dumped log is not in ascending seq order (saw %d after %d)", l.Seq, prev)
			break
		}
		prev = l.Seq
	}
	if strings.TrimSpace(r.RoundTripOut) == "" {
		c.failf("rendering the dumped log in a fresh process produced nothing (stderr: %.200s)", r.RoundTripErr)
	}
	if c.Passed {
		c.Details = append(c.Details, fmt.Sprintf("%d events dumped and re-rendered", len(s.Log)))
	}
	return c
}

func ch2Replay(r *Ch2Result) Check {
	c := Check{ID: "replay", Title: "two renders of one log are byte-identical", Points: 10, Passed: true, Earned: 10}
	for _, v := range Ch2Vendors {
		a, b := r.Render[v], r.RenderTwice[v]
		if strings.TrimSpace(a) == "" {
			c.failf("%s: `render` produced nothing (stderr: %.200s)", v, r.RenderErr[v])
			continue
		}
		if a != b {
			c.failf("%s: two renders of the same log differ. Look for the clock, a random id, "+
				"Go's randomized map iteration order, or iteration over a set.", v)
		}
	}
	if c.Passed {
		c.Details = append(c.Details, "all three vendors render deterministically")
	}
	return c
}

func ch2Redaction(r *Ch2Result) Check {
	c := Check{ID: "redaction", Title: "redacted content is absent from later renders", Points: 10, Passed: true, Earned: 10}
	for _, v := range Ch2Vendors {
		plain, red := r.Render[v], r.Redacted[v]
		if strings.TrimSpace(red) == "" {
			c.failf("%s: rendering the redacted log produced nothing (stderr: %.200s)", v, r.RedactedErr[v])
			continue
		}
		// Negative control: without the Redacted event the secret MUST be
		// present. Otherwise a submission that simply never renders tool
		// results would pass this check for the wrong reason.
		if !strings.Contains(plain, RedactedSecret) {
			c.failf("%s: the un-redacted render does not contain the tool output at all, "+
				"so this check cannot show that redaction did anything", v)
			continue
		}
		if strings.Contains(red, RedactedSecret) {
			c.failf("%s: redacted tool output (%q) still appears in the rendered request", v, RedactedSecret)
		}
		// The CALL must survive. RedactResult stubs the result and keeps the
		// call, so the model can still see what it asked for and why.
		if !strings.Contains(red, "read_file") {
			c.failf("%s: the tool CALL disappeared too. RedactResult stubs the result and keeps the call.", v)
		}
	}
	if c.Passed {
		c.Details = append(c.Details, "tool output stubbed, tool call preserved, on all three vendors")
	}
	return c
}

func ch2Ephemera(r *Ch2Result) Check {
	c := Check{ID: "ephemera", Title: "delivered exactly once, then absent", Points: 10, Passed: true, Earned: 10}
	e := r.Ephemera
	if e == nil {
		c.failf("no ephemera session was recorded")
		return c
	}
	const marker = "CURRENT_TIME=2026-09-12T00:00:00Z"
	if len(e.Requests) < 2 {
		c.failf("expected at least 2 requests in the ephemera session, saw %d", len(e.Requests))
		return c
	}
	var carrying []int
	for i, req := range e.Requests {
		if strings.Contains(string(req.Body), marker) {
			carrying = append(carrying, i+1)
		}
	}
	switch len(carrying) {
	case 0:
		c.failf("the ephemeral part was never delivered in any request")
	case 1:
		if carrying[0] != 1 {
			c.failf("the ephemeral part was delivered in request %d; it should ride the FIRST request after it arrives", carrying[0])
		}
	default:
		c.failf("the ephemeral part was delivered in %d requests (%v); it must be delivered exactly once. "+
			"A stale timestamp is not stale data, it is a lie.", len(carrying), carrying)
	}
	if c.Passed {
		c.Details = append(c.Details, fmt.Sprintf("delivered once, in request %d of %d", carrying[0], len(e.Requests)))
	}
	return c
}

// ch2UsageCheck is the silent bug made loud. Nothing crashes, no test fails,
// the number is simply not the number — so it is graded directly.
func ch2UsageCheck(r *Ch2Result) Check {
	c := Check{ID: "usage", Title: "four token categories, normalized, disjoint", Points: 10, Passed: true, Earned: 10}

	// Main session: round 1 {100,0,50,30} + round 2 {12,0,200,8}.
	// Anthropic reports this DISJOINT, OpenAI and Gemini as a SUBSET of their
	// prompt totals. A parser that sums naively gets 362 input on two of the
	// three vendors instead of 112.
	want := Ch2Usage{Input: 112, CacheWrite: 0, CacheRead: 250, Output: 38}
	for _, v := range Ch2Vendors {
		s := r.Session[v]
		if s == nil {
			c.failf("no session for %s", v)
			continue
		}
		if s.Usage == nil {
			c.failf("%s: no usage line was printed at end of session", v)
			continue
		}
		if *s.Usage != want {
			c.failf("%s: usage %+v, want %+v. Cached tokens are a SUBSET of the prompt total on "+
				"OpenAI and Gemini, and DISJOINT from it on Anthropic; and Gemini reports thinking "+
				"tokens separately from candidate tokens.", v, *s.Usage, want)
		}
	}

	// Cache-write probe: only Anthropic and OpenAI report a cache-write token
	// count at all. Gemini reports none anywhere in usageMetadata, so the
	// honest canonical answer for Gemini is zero — the asymmetry is real and
	// must not be papered over with an invented number.
	for _, v := range Ch2Vendors {
		s := r.UsageProbe[v]
		if s == nil || s.Usage == nil {
			c.failf("%s: no usage line from the cache-write probe", v)
			continue
		}
		w := Ch2Usage{Input: 7, CacheWrite: 300, CacheRead: 0, Output: 11}
		if v == "gemini" {
			w = Ch2Usage{Input: 7, CacheWrite: 0, CacheRead: 0, Output: 11}
		}
		if *s.Usage != w {
			c.failf("%s cache-write probe: usage %+v, want %+v", v, *s.Usage, w)
		}
	}
	if c.Passed {
		c.Details = append(c.Details, "input/cache_write/cache_read/output normalized on all three vendors")
	}
	return c
}

// --- seam-render -----------------------------------------------------------

func ch2SeamRender(r *Ch2Result) Check {
	c := Check{ID: "seam-render", Title: "one log, three vendor request shapes", Points: 15, Passed: true, Earned: 15}

	// Vendor field names are graded EXACTLY. `tool_use_id` is Anthropic's
	// spelling, not the student's, and misspelling it is a real bug rather
	// than a naming preference. Only the student's OWN names are normalized.
	checkAnthropicRequest(&c, r.Render["anthropic"])
	checkOpenAIRequest(&c, r.Render["openai"])
	checkGeminiRequest(&c, r.Render["gemini"])
	checkGeminiThoughtReplay(&c, r.ThoughtReplay, r.ThoughtReplayErr)
	if c.Passed {
		c.Details = append(c.Details, "Exhibit A (three authorships) and Exhibit B (the merged message) both correct")
		c.Details = append(c.Details, "per-call replay material survives a Gemini round trip")
	}
	return c
}

// checkGeminiThoughtReplay grades ToolCallPart.Opaque on the render side.
//
// Gemini's thought signature is not a standalone thinking block: it arrives as
// a SIBLING KEY of functionCall, bound to that one call. Replay the call to
// Gemini 3.x without it and the API answers 400, which is why §2.4a carries a
// field for per-call material at all. A student who stores opaque material
// only as a standalone part has nowhere to put this, and their agent breaks in
// Chapter 3 rather than here — so it is graded here.
func checkGeminiThoughtReplay(c *Check, body, stderr string) {
	if strings.TrimSpace(body) == "" {
		c.failf("gemini: `render` produced nothing for the thought-replay log (stderr: %.200s)",
			stderr)
		return
	}
	m := decode(c, "gemini(thought-replay)", body)
	if m == nil {
		return
	}
	contents, _ := m["contents"].([]any)
	if len(contents) == 0 {
		c.failf("gemini: no `contents` array in the thought-replay render")
		return
	}
	found := false
	for _, raw := range contents {
		turn, _ := raw.(map[string]any)
		parts, _ := turn["parts"].([]any)
		for _, p := range parts {
			part, _ := p.(map[string]any)
			if _, isCall := part["functionCall"]; !isCall {
				continue
			}
			found = true
			sig, ok := part["thoughtSignature"]
			if !ok {
				c.failf("gemini: the replayed functionCall carries no `thoughtSignature`. " +
					"The log's tool call has opaque material recorded against this exact " +
					"model, so it must be replayed as a sibling key of functionCall. " +
					"Gemini 3.x answers 400 when a replayed functionCall is missing its " +
					"signature — this is the failure ToolCallPart.Opaque exists to prevent.")
				continue
			}
			if s, _ := sig.(string); s != "sig-bound-to-this-call" {
				c.failf("gemini: the replayed functionCall carries thoughtSignature %v, but the "+
					"log recorded %q against this call. Opaque material is carried "+
					"verbatim and never reconstructed.", sig, "sig-bound-to-this-call")
			}
		}
	}
	if !found {
		c.failf("gemini: the thought-replay render contains no `functionCall` part at all; " +
			"the assistant turn that issued the tool call was dropped from the request.")
	}
}

func decode(c *Check, vendor, body string) map[string]any {
	if strings.TrimSpace(body) == "" {
		c.failf("%s: `render` produced nothing on stdout", vendor)
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		c.failf("%s: `render` did not print a single JSON object (%v). stdout must carry the "+
			"request body and nothing else.", vendor, err)
		return nil
	}
	return m
}

func checkAnthropicRequest(c *Check, body string) {
	m := decode(c, "anthropic", body)
	if m == nil {
		return
	}
	if _, ok := m["system"]; !ok {
		c.failf("anthropic: no top-level `system` parameter. Anthropic takes the system prompt " +
			"as a request parameter, not as a message.")
	}
	msgs, _ := m["messages"].([]any)
	if len(msgs) == 0 {
		c.failf("anthropic: no `messages` array")
		return
	}
	var toolUseID, resultID string
	sawMergedText := false
	for _, raw := range msgs {
		msg, _ := raw.(map[string]any)
		role, _ := msg["role"].(string)
		blocks, _ := msg["content"].([]any)
		seenResult := false
		for _, b := range blocks {
			blk, _ := b.(map[string]any)
			switch blk["type"] {
			case "tool_use":
				if role != "assistant" {
					c.failf("anthropic: a tool_use block appeared in a %q message; it belongs to the assistant", role)
				}
				toolUseID, _ = blk["id"].(string)
			case "tool_result":
				// EXHIBIT A: Anthropic says the HUMAN said it. That is false,
				// and it is the tidiest available lie under this format.
				if role != "user" {
					c.failf("anthropic: tool_result must be carried in a \"user\" message, not %q", role)
				}
				if _, ok := blk["tool_use_id"]; !ok {
					c.failf("anthropic: tool_result block has no `tool_use_id`")
				}
				resultID, _ = blk["tool_use_id"].(string)
				seenResult = true
			case "text":
				// EXHIBIT B: the human's next instruction merged into the same
				// user message. Required because a tool_result must immediately
				// follow its tool_use, and because tool_result blocks must come
				// FIRST in the content array — text before a tool_result is a 400.
				if role == "user" && seenResult {
					sawMergedText = true
				}
				if role == "user" && !seenResult {
					for _, b2 := range blocks {
						blk2, _ := b2.(map[string]any)
						if blk2["type"] == "tool_result" {
							c.failf("anthropic: a text block precedes a tool_result in the same user " +
								"message. tool_result blocks must come FIRST; text after them.")
							break
						}
					}
				}
			}
		}
	}
	if toolUseID == "" {
		c.failf("anthropic: no tool_use block found for the supplied tool call")
	}
	if resultID == "" {
		c.failf("anthropic: no tool_result block found for the supplied tool result")
	}
	if toolUseID != "" && resultID != "" && toolUseID != resultID {
		c.failf("anthropic: tool_use id %q and tool_result tool_use_id %q do not match", toolUseID, resultID)
	}
	if !sawMergedText {
		c.failf("anthropic: the tool result and the human's following instruction were not merged " +
			"into one user message with the result first")
	}
	// Opaque replay material was issued by this exact model, so it must be
	// handed back.
	if !strings.Contains(body, "sig-exhibit-1") {
		c.failf("anthropic: the thinking block issued by this same model was not replayed. " +
			"Opaque material is carried, never interpreted, and handed back to the model that issued it.")
	}
}

func checkOpenAIRequest(c *Check, body string) {
	m := decode(c, "openai", body)
	if m == nil {
		return
	}
	msgs, _ := m["messages"].([]any)
	if len(msgs) == 0 {
		c.failf("openai: no `messages` array")
		return
	}
	sawSystem, sawToolRole := false, false
	var callID, resultID string
	for _, raw := range msgs {
		msg, _ := raw.(map[string]any)
		role, _ := msg["role"].(string)
		switch role {
		case "system", "developer":
			sawSystem = true
		case "tool":
			// EXHIBIT A: OpenAI invents a role for it.
			sawToolRole = true
			if _, ok := msg["tool_call_id"]; !ok {
				c.failf("openai: a tool message has no `tool_call_id`")
			}
			resultID, _ = msg["tool_call_id"].(string)
		case "assistant":
			if tcs, ok := msg["tool_calls"].([]any); ok && len(tcs) > 0 {
				tc, _ := tcs[0].(map[string]any)
				callID, _ = tc["id"].(string)
				fn, _ := tc["function"].(map[string]any)
				if fn == nil {
					c.failf("openai: tool_calls entry has no `function` object")
					break
				}
				// arguments is a JSON-ENCODED STRING here, not an object.
				if _, isString := fn["arguments"].(string); !isString {
					c.failf("openai: tool_calls[0].function.arguments must be a JSON-encoded STRING, " +
						"not an object. This is the vendor's encoding decision, and it is why the " +
						"context stores the decoded object instead.")
				}
			}
		}
	}
	if !sawSystem {
		c.failf("openai: no system (or developer) message. OpenAI takes the system prompt as a " +
			"message inside the array — the same fact Anthropic puts in a parameter.")
	}
	if !sawToolRole {
		c.failf("openai: no message with role \"tool\" for the supplied tool result")
	}
	if callID != "" && resultID != "" && callID != resultID {
		c.failf("openai: tool_calls id %q and tool_call_id %q do not match", callID, resultID)
	}
	if callID == "" {
		c.failf("openai: the assistant's tool call was not rendered")
	}
	if strings.Contains(body, "sig-exhibit-1") {
		c.failf("openai: replay material issued by an Anthropic model was sent to OpenAI. " +
			"Opaque material goes back only to the exact model that issued it.")
	}
}

func checkGeminiRequest(c *Check, body string) {
	m := decode(c, "gemini", body)
	if m == nil {
		return
	}
	if _, ok := m["systemInstruction"]; !ok {
		c.failf("gemini: no top-level `systemInstruction`. Gemini hoists the system prompt clean " +
			"out of the message list — one fact, a third placement.")
	}
	if _, ok := m["messages"]; ok {
		c.failf("gemini: the request has a `messages` array. Gemini calls it `contents`.")
	}
	contents, _ := m["contents"].([]any)
	if len(contents) == 0 {
		c.failf("gemini: no `contents` array")
		return
	}
	sawModelRole, sawCall, sawResponse := false, false, false
	for _, raw := range contents {
		turn, _ := raw.(map[string]any)
		role, _ := turn["role"].(string)
		if role == "assistant" {
			c.failf("gemini: role \"assistant\" is not a Gemini role; the assistant is called \"model\"")
		}
		if role == "model" {
			sawModelRole = true
		}
		parts, _ := turn["parts"].([]any)
		if parts == nil {
			c.failf("gemini: a turn has no `parts` array")
		}
		for _, p := range parts {
			part, _ := p.(map[string]any)
			if fc, ok := part["functionCall"].(map[string]any); ok {
				sawCall = true
				if _, isObj := fc["args"].(map[string]any); !isObj {
					c.failf("gemini: functionCall.args must be a JSON object, not a string")
				}
			}
			if fr, ok := part["functionResponse"].(map[string]any); ok {
				// EXHIBIT A: Gemini splits the difference — a functionResponse
				// part inside a user turn.
				sawResponse = true
				if role != "user" {
					c.failf("gemini: functionResponse appeared in a %q turn; it belongs in a user turn", role)
				}
				if name, _ := fr["name"].(string); name == "" {
					c.failf("gemini: functionResponse.name is required and is empty. The context " +
						"stores only the call id, so the renderer must resolve the name from the " +
						"matching tool call.")
				}
				if _, isObj := fr["response"].(map[string]any); !isObj {
					c.failf("gemini: functionResponse.response must be a JSON OBJECT. A bare string " +
						"is a 400 — a scalar every other vendor accepts has to be wrapped here.")
				}
			}
		}
	}
	if !sawModelRole {
		c.failf("gemini: no turn with role \"model\"")
	}
	if !sawCall {
		c.failf("gemini: the assistant's tool call was not rendered as a functionCall part")
	}
	if !sawResponse {
		c.failf("gemini: the tool result was not rendered as a functionResponse part")
	}
	if strings.Contains(body, "sig-exhibit-1") {
		c.failf("gemini: replay material issued by an Anthropic model was sent to Gemini")
	}
}

// --- seam-parse ------------------------------------------------------------

func ch2SeamParse(r *Ch2Result) Check {
	c := Check{ID: "seam-parse", Title: "three responses, one context, provenance preserved", Points: 15, Passed: true, Earned: 15}

	proj := map[string][]string{}
	for _, v := range Ch2Vendors {
		s := r.Session[v]
		if s == nil || len(s.Log) == 0 {
			c.failf("%s: no dumped log to compare", v)
			return c
		}
		proj[v] = SaidProjection(s.Log)
	}
	base := proj["anthropic"]
	for _, v := range Ch2Vendors[1:] {
		if !equalStrings(base, proj[v]) {
			c.failf("%s: the parsed context differs from anthropic's beyond provenance.\n  anthropic: %v\n  %s: %v\n"+
				"  Everything the model SAID must normalize; only the record of who said it survives.",
				v, base, v, proj[v])
		}
	}

	// Provenance must be PRESERVED, not normalized away. A submission whose
	// three contexts are fully identical has thrown it away, and will be
	// unable to render a valid Gemini request after a tool call.
	seen := map[string]string{}
	for _, v := range Ch2Vendors {
		vendor, model, surface := provenanceOf(r.Session[v].Log)
		if vendor == "" || model == "" {
			c.failf("%s: response events carry no provenance (vendor=%q model=%q). Provenance is "+
				"recorded at write time and can never be reconstructed later.", v, vendor, model)
			continue
		}
		if surface == "" {
			c.failf("%s: provenance has no surface. Signature validity is scoped to "+
				"(vendor, model, surface), not to vendor.", v)
		}
		key := vendor + "/" + model
		if prev, dup := seen[key]; dup {
			c.failf("%s: provenance %q is identical to %s's. The three contexts must differ here "+
				"and nowhere else.", v, key, prev)
		}
		seen[key] = v
	}
	if c.Passed {
		c.Details = append(c.Details, fmt.Sprintf("three vendors, one context; provenance distinct: %v", keysOf(seen)))
	}
	return c
}

func provenanceOf(lines []Ch2LogLine) (vendor, model, surface string) {
	for _, l := range lines {
		if l.Type != normName("response_ended") {
			continue
		}
		resp := getMap(l.Data, "response")
		if resp == nil {
			continue
		}
		from := getMap(resp, "from", "provenance")
		if from == nil {
			continue
		}
		return getStr(from, "vendor"), getStr(from, "model"), getStr(from, "surface")
	}
	return "", "", ""
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func keysOf(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
