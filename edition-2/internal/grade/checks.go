package grade

import (
	"fmt"
	"strings"

	"github.com/waywardgeek/ensemble/internal/fakeanthropic"
)

// Check is one graded property. Details explain the verdict in the student's
// terms — a failing check that does not say what to fix is a bug in the
// grader, not a lesson.
type Check struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Points  int      `json:"points"`
	Earned  int      `json:"earned"`
	Passed  bool     `json:"passed"`
	Details []string `json:"details,omitempty"`
}

func (c *Check) failf(format string, args ...any) {
	c.Passed = false
	c.Earned = 0
	c.Details = append(c.Details, fmt.Sprintf(format, args...))
}

func (c *Check) notef(format string, args ...any) {
	c.Details = append(c.Details, fmt.Sprintf(format, args...))
}

// Evaluate turns run evidence into the graded checklist.
func Evaluate(res *RunResult) []Check {
	return []Check{
		checkProtocol(res),
		checkWire(res),
		checkCallsPerRound(res),
		checkRepliesParsed(res),
		checkMemory(res),
		checkAppendOnly(res),
		checkUsage(res),
	}
}

// --- 1. the stdio contract -------------------------------------------------

func checkProtocol(res *RunResult) Check {
	c := Check{ID: "protocol", Title: "Speaks the stdio JSON-lines contract", Points: 15, Passed: true, Earned: 15}
	for _, p := range res.Protocol {
		c.failf("%s", p)
	}
	if res.GotRounds != len(Script) {
		c.failf("answered %d of %d rounds", res.GotRounds, len(Script))
	}
	for _, l := range res.Extra {
		c.failf("unexpected stdout line: %q — stdout carries the protocol only; put diagnostics on stderr",
			truncate(l, 160))
	}
	if res.ExitError != "" {
		c.failf("program did not exit cleanly: %s (exit code %d)", res.ExitError, res.ExitCode)
	}
	if c.Passed {
		c.notef("%d rounds, clean exit 0", res.GotRounds)
	}
	return c
}

// --- 2. wire correctness ---------------------------------------------------

func checkWire(res *RunResult) Check {
	c := Check{ID: "wire", Title: "Requests are well-formed Messages API calls", Points: 15, Passed: true, Earned: 15}
	if len(res.Records) == 0 {
		c.failf("the fake API server received no requests at all — is the program reading ANTHROPIC_BASE_URL?")
		return c
	}
	for _, r := range res.Records {
		for _, v := range r.Violations {
			c.failf("request %d: %s", r.Seq, v)
		}
	}
	if c.Passed {
		r := res.Records[0]
		c.notef("%d requests, all well-formed (model=%q, max_tokens=%d, content encoded as %s)",
			len(res.Records), r.Body.Model, r.Body.MaxTokens, contentShapes(r.Body.Messages))
	}
	return c
}

func contentShapes(msgs []fakeanthropic.Message) string {
	seen := map[string]bool{}
	var out []string
	for _, m := range msgs {
		s := m.ContentShape()
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return strings.Join(out, "+")
}

// --- 3. one API call per round ---------------------------------------------

func checkCallsPerRound(res *RunResult) Check {
	c := Check{ID: "calls", Title: "Exactly one API call per round", Points: 10, Passed: true, Earned: 10}
	var overflow int
	for _, r := range res.Records {
		if r.Overflow {
			overflow++
		}
	}
	if len(res.Records) != len(Script) {
		c.failf("expected %d requests, one per round; saw %d", len(Script), len(res.Records))
		if overflow > 0 {
			c.notef("%d request(s) arrived after the %d scripted rounds were used up",
				overflow, len(Script))
		}
		return c
	}
	c.notef("%d rounds, %d requests", len(Script), len(res.Records))
	return c
}

// --- 4. the response was actually parsed ------------------------------------

func checkRepliesParsed(res *RunResult) Check {
	c := Check{ID: "replies", Title: "Answers come from the API response", Points: 10, Passed: true, Earned: 10}
	for i, want := range Replies() {
		if i >= len(res.Answers) {
			c.failf("round %d produced no answer", i+1)
			continue
		}
		if strings.TrimSpace(res.Answers[i]) != want {
			c.failf("round %d: answered %q, but the server returned %q",
				i+1, truncate(res.Answers[i], 120), truncate(want, 120))
		}
	}
	if c.Passed {
		c.notef("every answer matched the text the server returned")
	}
	return c
}

// --- 5. the conversation exists (the point of the exercise) -----------------

func checkMemory(res *RunResult) Check {
	c := Check{
		ID:     "memory",
		Title:  "Round " + fmt.Sprint(MemoryProbeRound) + "'s request still carries the whole history",
		Points: 25, Passed: true, Earned: 25,
	}
	probe := findRequestForRound(res.Records, MemoryProbeRound)
	if probe == nil {
		c.failf("could not find the request carrying round %d's question (%q) as its final user message",
			MemoryProbeRound, truncate(Script[MemoryProbeRound-1].User, 60))
		return c
	}

	var foundInAssistant, foundAnywhere bool
	for _, m := range probe.Body.Messages {
		if strings.Contains(m.Text(), PlantedFact) {
			foundAnywhere = true
			if m.Role == "assistant" {
				foundInAssistant = true
			}
		}
	}
	switch {
	case foundInAssistant:
		c.notef("request %d carried %d messages including the assistant turn containing %q",
			probe.Seq, len(probe.Body.Messages), PlantedFact)
	case foundAnywhere:
		c.failf("%q appears in request %d, but not in an assistant message — the model's own replies "+
			"must be appended to the conversation, not just the user's turns", PlantedFact, probe.Seq)
	default:
		c.failf("request %d does not contain %q anywhere. The server planted that string in its round-1 "+
			"reply; if it is gone, the program is not sending the conversation — it is sending the latest "+
			"question. The API is stateless: the history lives in your process or it lives nowhere.",
			probe.Seq, PlantedFact)
	}

	wantMessages := 2*MemoryProbeRound - 1
	if got := len(probe.Body.Messages); got != wantMessages && c.Passed {
		c.notef("note: expected %d messages by round %d (%d user + %d assistant), saw %d",
			wantMessages, MemoryProbeRound, MemoryProbeRound, MemoryProbeRound-1, got)
	}
	return c
}

// findRequestForRound locates the request whose last message is the given
// round's user text. Matching on content rather than arrival order keeps the
// check honest when a submission makes extra calls.
func findRequestForRound(recs []fakeanthropic.Record, round int) *fakeanthropic.Record {
	want := Script[round-1].User
	for i := range recs {
		msgs := recs[i].Body.Messages
		if len(msgs) == 0 {
			continue
		}
		last := msgs[len(msgs)-1]
		if last.Role == "user" && strings.Contains(last.Text(), want) {
			return &recs[i]
		}
	}
	return nil
}

// --- 6. append-only growth --------------------------------------------------

func checkAppendOnly(res *RunResult) Check {
	c := Check{ID: "growth", Title: "Each request extends the previous one (append-only)", Points: 15, Passed: true, Earned: 15}
	if len(res.Records) < 2 {
		c.failf("need at least two requests to compare; saw %d", len(res.Records))
		return c
	}
	for i := 1; i < len(res.Records); i++ {
		prev := res.Records[i-1].Body.Messages
		cur := res.Records[i].Body.Messages
		if len(cur) < len(prev) {
			c.failf("request %d has %d messages, fewer than request %d's %d — history must only grow",
				i+1, len(cur), i, len(prev))
			continue
		}
		for j := range prev {
			if prev[j].Role != cur[j].Role || prev[j].Text() != cur[j].Text() {
				c.failf("request %d diverges from request %d at messages[%d]: was {%s: %q}, now {%s: %q}",
					i+1, i, j,
					prev[j].Role, truncate(prev[j].Text(), 60),
					cur[j].Role, truncate(cur[j].Text(), 60))
				break
			}
		}
		if grew := len(cur) - len(prev); grew != 2 && c.Passed {
			c.notef("note: request %d grew by %d messages (expected 2: one assistant reply, one new question)",
				i+1, grew)
		}
	}
	if c.Passed {
		c.notef("history grew %d → %d messages, every earlier turn preserved byte-for-byte",
			len(res.Records[0].Body.Messages),
			len(res.Records[len(res.Records)-1].Body.Messages))
	}
	return c
}

// --- 7. usage accounting ----------------------------------------------------

func checkUsage(res *RunResult) Check {
	c := Check{ID: "usage", Title: "Cumulative usage reported and correct", Points: 10, Passed: true, Earned: 10}
	if !res.UsageSeen {
		c.failf("no {\"usage\": {\"input\": .., \"output\": ..}} line after stdin closed")
		return c
	}
	want := res.FakeUsage
	if res.UsageInput == 0 || res.UsageOutput == 0 {
		c.failf("usage reported as input=%d output=%d — both must be nonzero",
			res.UsageInput, res.UsageOutput)
	}
	if res.UsageInput != want.InputTokens || res.UsageOutput != want.OutputTokens {
		c.failf("reported input=%d output=%d, but the server returned totals of input=%d output=%d "+
			"— sum the usage field of every response",
			res.UsageInput, res.UsageOutput, want.InputTokens, want.OutputTokens)
	}
	if c.Passed {
		c.notef("input=%d output=%d, matching the server's totals exactly", res.UsageInput, res.UsageOutput)
		var curve []string
		for _, r := range res.Records {
			curve = append(curve, fmt.Sprint(r.Usage.InputTokens))
		}
		c.notef("input tokens per round: %s — watch it climb; that is the resent history, and it is money",
			strings.Join(curve, " → "))
	}
	return c
}
