package grade

// Tests for the Chapter 19 Responses-API fixture.
//
// These exist because the fixture is graded against before any student code
// exists, and a grader's failure modes are asymmetric. A fixture that reports
// spurious violations fails every submission and looks like a hard chapter; a
// fixture that reports none passes every submission and looks like an easy
// one. Both are indistinguishable from a working grader unless something
// pins down the zero-violation case and the one-violation-per-mistake case
// from opposite directions, which is what TestCh19RespWellFormedRequest and
// TestCh19RespBannedFields do between them.

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

const ch19RespTestAuth = "Bearer ch19-test-token"

// ch19RespGoodBody returns a request that must produce zero violations. It
// returns a fresh map every call so a table-driven test can mutate one copy
// per case without the mutation leaking into the next case.
func ch19RespGoodBody() map[string]any {
	return map[string]any{
		"model":  "gpt-5.6",
		"store":  false,
		"stream": true,
		"input": []any{
			map[string]any{
				"role": "developer",
				"content": []any{
					map[string]any{
						"type":                    "input_text",
						"text":                    "You are a careful agent.",
						"prompt_cache_breakpoint": map[string]any{"type": "prompt_cache_breakpoint"},
					},
				},
			},
			map[string]any{
				"role":    "user",
				"content": []any{map[string]any{"type": "input_text", "text": "hello"}},
			},
		},
		"include":              []any{"reasoning.encrypted_content"},
		"reasoning":            map[string]any{"summary": "auto"},
		"prompt_cache_options": map[string]any{"mode": "explicit"},
	}
}

// ch19RespDo posts a body and returns the status and the full response bytes.
func ch19RespDo(t *testing.T, url string, body map[string]any, auth string) (int, []byte) {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, url+"/v1/responses", bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	return resp.StatusCode, out
}

type ch19RespEvent struct {
	Name string
	Data map[string]any
}

// ch19RespParseSSE parses an event-stream body into ordered frames and, as it
// goes, enforces the fixture's own dual-emission invariant: the event: line and
// the "type" field inside data: must agree. Students are free to dispatch on
// either, so a disagreement here would make the grader's verdict depend on
// which one a given student picked.
func ch19RespParseSSE(t *testing.T, body []byte) []ch19RespEvent {
	t.Helper()
	var events []ch19RespEvent
	var pendingName string

	sc := bufio.NewScanner(bytes.NewReader(body))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "event: "):
			pendingName = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "data: "):
			payload := strings.TrimPrefix(line, "data: ")
			var m map[string]any
			if err := json.Unmarshal([]byte(payload), &m); err != nil {
				t.Fatalf("event %q: data is not valid JSON: %v\n%s", pendingName, err, payload)
			}
			typ, _ := m["type"].(string)
			if typ != pendingName {
				t.Fatalf("event: line %q disagrees with data type %q", pendingName, typ)
			}
			events = append(events, ch19RespEvent{Name: pendingName, Data: m})
			pendingName = ""
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan SSE: %v", err)
	}
	return events
}

func ch19RespEventNames(events []ch19RespEvent) []string {
	names := make([]string, len(events))
	for i, e := range events {
		names[i] = e.Name
	}
	return names
}

// 1. The case that matters most: a correct request must be silent.
func TestCh19RespWellFormedRequestHasNoViolations(t *testing.T) {
	s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"text"}})
	defer s.Close()

	status, body := ch19RespDo(t, s.URL(), ch19RespGoodBody(), ch19RespTestAuth)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", status, body)
	}

	obs := s.Observations()
	if len(obs.Violations) != 0 {
		t.Fatalf("well-formed request produced %d violations: %q", len(obs.Violations), obs.Violations)
	}
	if len(obs.Requests) != 1 {
		t.Fatalf("Requests = %d, want 1", len(obs.Requests))
	}

	r := obs.Requests[0]
	if r.StoreValue != false {
		t.Errorf("StoreValue = %#v, want false", r.StoreValue)
	}
	if r.StreamValue != true {
		t.Errorf("StreamValue = %#v, want true", r.StreamValue)
	}
	if got, want := strings.Join(r.Roles, ","), "developer,user"; got != want {
		t.Errorf("Roles = %q, want %q", got, want)
	}
	if r.HasInstructions {
		t.Errorf("HasInstructions = true, but the body has no instructions field")
	}
	if !r.IncludesEncryptedReasoning {
		t.Errorf("IncludesEncryptedReasoning = false, want true")
	}
	if r.ReasoningSummary != "auto" {
		t.Errorf("ReasoningSummary = %q, want %q", r.ReasoningSummary, "auto")
	}
	if r.PromptCacheMode != "explicit" {
		t.Errorf("PromptCacheMode = %q, want %q", r.PromptCacheMode, "explicit")
	}
	if r.BreakpointCount != 1 {
		t.Errorf("BreakpointCount = %d, want 1", r.BreakpointCount)
	}
	if r.Auth != ch19RespTestAuth {
		t.Errorf("Auth = %q, want %q", r.Auth, ch19RespTestAuth)
	}
	if len(r.Body) == 0 || r.Decoded == nil {
		t.Errorf("raw Body and Decoded must both be retained")
	}

	// The per-message breakpoint report is what lets the parent tell a
	// cacheable developer content block from an uncacheable instructions
	// string, so it has to be populated even on the clean path.
	if len(r.MessageBreakpoints) != 2 {
		t.Fatalf("MessageBreakpoints = %d, want 2", len(r.MessageBreakpoints))
	}
	dev := r.MessageBreakpoints[0]
	if dev.Role != "developer" || dev.BlockCount != 1 || !dev.HasBreakpoint {
		t.Errorf("developer message report = %+v, want role=developer blocks=1 breakpoint=true", dev)
	}
	if usr := r.MessageBreakpoints[1]; usr.Role != "user" || usr.HasBreakpoint {
		t.Errorf("user message report = %+v, want role=user breakpoint=false", usr)
	}
}

// 2. Every banned field, one at a time, must produce exactly one violation
// naming it. "Exactly one" is the real assertion: a field that trips two
// checks would double-count a single mistake in the student's score.
//
// The table is keyed off ch19RespBannedFields itself and fails loudly on any
// field it has no value for, so adding a field to the fixture without adding
// a case here cannot silently leave it untested.
func TestCh19RespBannedFields(t *testing.T) {
	values := map[string]any{
		"background":             true,
		"conversation":           "conv_abc",
		"max_output_tokens":      512,
		"max_tool_calls":         4,
		"metadata":               map[string]any{"k": "v"},
		"moderation":             "auto",
		"multi_agent":            true,
		"prompt":                 map[string]any{"id": "pmpt_abc"},
		"prompt_cache_retention": "24h",
		"safety_identifier":      "user-hash-abc",
		"temperature":            0.7,
		"top_logprobs":           5,
		"top_p":                  0.9,
		"truncation":             "auto",
		"user":                   "user-abc",
	}

	if len(values) != len(ch19RespBannedFields) {
		t.Fatalf("test table has %d entries but ch19RespBannedFields has %d",
			len(values), len(ch19RespBannedFields))
	}

	for _, field := range ch19RespBannedFields {
		field := field
		t.Run(field, func(t *testing.T) {
			val, ok := values[field]
			if !ok {
				t.Fatalf("test table has no value for banned field %q", field)
			}
			s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"text"}})
			defer s.Close()

			body := ch19RespGoodBody()
			body[field] = val
			ch19RespDo(t, s.URL(), body, ch19RespTestAuth)

			got := s.Observations().Violations
			if len(got) != 1 {
				t.Fatalf("violations = %q, want exactly 1 naming %q", got, field)
			}
			if want := "banned field: " + field; got[0] != want {
				t.Fatalf("violation = %q, want %q", got[0], want)
			}
		})
	}

	// Chat Completions fields that do not exist on the Responses API must NOT
	// be banned. Banning them invents a rule, and a student who has correctly
	// migrated would be failed for an offence the vendor never defined. This
	// is the direction of error a grader is least likely to notice, because
	// the violation text still reads perfectly plausibly.
	t.Run("chat_completions_fields_are_not_banned", func(t *testing.T) {
		for _, f := range []string{"n", "presence_penalty", "frequency_penalty", "logprobs", "stop"} {
			if ch19RespContains(ch19RespBannedFields, f) {
				t.Errorf("%q must not be in the banned list: it is a Chat Completions "+
					"field that does not exist on the Responses API", f)
			}
		}
	})

	// The near-miss pair. top_logprobs is unsupported; logprobs is not, and
	// prompt_cache_retention is unsupported while prompt_cache_options is
	// accepted and echoed back.
	t.Run("near_miss_names", func(t *testing.T) {
		if !ch19RespContains(ch19RespBannedFields, "top_logprobs") {
			t.Error("top_logprobs must be banned")
		}
		if ch19RespContains(ch19RespBannedFields, "prompt_cache_options") {
			t.Error("prompt_cache_options is accepted, not banned")
		}
		if !ch19RespContains(ch19RespBannedFields, "prompt_cache_retention") {
			t.Error("prompt_cache_retention must be banned")
		}
	})

	// Presence, not value: the default temperature is semantically a no-op but
	// is still a rejected request on this route, and a student who "cleans up"
	// by sending defaults instead of omitting fields must still fail.
	t.Run("default_valued_field_still_banned", func(t *testing.T) {
		s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"text"}})
		defer s.Close()
		body := ch19RespGoodBody()
		body["temperature"] = 1.0
		ch19RespDo(t, s.URL(), body, ch19RespTestAuth)
		if got := s.Observations().Violations; len(got) != 1 {
			t.Fatalf("violations = %q, want exactly 1", got)
		}
	})

	// previous_response_id is its own rule with its own diagnosis, so it must
	// not be folded into the generic banned-field list.
	t.Run("previous_response_id_is_its_own_violation", func(t *testing.T) {
		if ch19RespContains(ch19RespBannedFields, "previous_response_id") {
			t.Error("previous_response_id must not be in the generic banned list")
		}
		s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"text"}})
		defer s.Close()
		body := ch19RespGoodBody()
		body["previous_response_id"] = "resp_abc"
		ch19RespDo(t, s.URL(), body, ch19RespTestAuth)

		got := s.Observations().Violations
		if len(got) != 1 {
			t.Fatalf("violations = %q, want exactly 1", got)
		}
		if !strings.Contains(got[0], "previous_response_id") {
			t.Errorf("violation = %q, want one naming previous_response_id", got[0])
		}
	})
}

// Top-level instructions is EXPLICITLY PERMITTED by the docs ("Use
// instructions or developer messages"). The fixture records its presence and
// says nothing about it; whether a constitution belongs there rather than in a
// cacheable developer block is a caching judgement that belongs to the
// parent's check, not to the wire-format validator. Flagging it here would
// fail a conforming student.
func TestCh19RespInstructionsIsPermitted(t *testing.T) {
	s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"text"}})
	defer s.Close()

	b := ch19RespGoodBody()
	b["instructions"] = "You are a careful agent."
	ch19RespDo(t, s.URL(), b, ch19RespTestAuth)

	obs := s.Observations()
	if len(obs.Violations) != 0 {
		t.Fatalf("instructions produced violations %q; the docs permit the field", obs.Violations)
	}
	if !obs.Requests[0].HasInstructions {
		t.Error("HasInstructions = false, want true")
	}
}

// 3. The two highest-probability migration mistakes.
func TestCh19RespMigrationMistakes(t *testing.T) {
	t.Run("messages_instead_of_input", func(t *testing.T) {
		s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"text"}})
		defer s.Close()

		body := ch19RespGoodBody()
		body["messages"] = body["input"]
		delete(body, "input")
		ch19RespDo(t, s.URL(), body, ch19RespTestAuth)

		got := s.Observations().Violations
		if !ch19RespAnyContains(got, "messages:") {
			t.Errorf("violations = %q, want one naming the messages key", got)
		}
		// Both diagnoses are wanted, because they point at different fixes:
		// the key was renamed AND the required key is now missing.
		if !ch19RespAnyContains(got, "input: field absent") {
			t.Errorf("violations = %q, want one reporting input absent", got)
		}
	})

	t.Run("system_role", func(t *testing.T) {
		s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"text"}})
		defer s.Close()

		body := ch19RespGoodBody()
		in := body["input"].([]any)
		in[0].(map[string]any)["role"] = "system"
		ch19RespDo(t, s.URL(), body, ch19RespTestAuth)

		obs := s.Observations()
		if !ch19RespAnyContains(obs.Violations, `role "system" is banned`) {
			t.Fatalf("violations = %q, want one naming the system role", obs.Violations)
		}
		// Observed roles are recorded regardless of the verdict, so the parent
		// can report what was sent rather than only that it was wrong.
		if got := strings.Join(obs.Requests[0].Roles, ","); got != "system,user" {
			t.Errorf("Roles = %q, want %q", got, "system,user")
		}
	})

	t.Run("input_not_an_array", func(t *testing.T) {
		s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"text"}})
		defer s.Close()
		body := ch19RespGoodBody()
		body["input"] = "just a string"
		ch19RespDo(t, s.URL(), body, ch19RespTestAuth)
		if got := s.Observations().Violations; !ch19RespAnyContains(got, "input: must be an array") {
			t.Fatalf("violations = %q, want one reporting input is not an array", got)
		}
	})

	t.Run("store_absent_and_store_true_both_caught", func(t *testing.T) {
		for _, tc := range []struct {
			name   string
			mutate func(map[string]any)
			want   string
		}{
			{"absent", func(b map[string]any) { delete(b, "store") }, "store: field absent"},
			{"true", func(b map[string]any) { b["store"] = true }, "store: must be false"},
		} {
			s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"text"}})
			body := ch19RespGoodBody()
			tc.mutate(body)
			ch19RespDo(t, s.URL(), body, ch19RespTestAuth)
			if got := s.Observations().Violations; !ch19RespAnyContains(got, tc.want) {
				t.Errorf("%s: violations = %q, want one containing %q", tc.name, got, tc.want)
			}
			s.Close()
		}
	})

	t.Run("bad_authorization", func(t *testing.T) {
		for _, tc := range []struct{ auth, want string }{
			{"", "authorization: header absent"},
			{"Token abc", `must start with "Bearer "`},
			// Sent as "Bearer " with an empty token; HTTP strips the trailing
			// space, so the fixture must recognise a bare scheme as an empty
			// token rather than as a malformed one. Writing "Bearer " here
			// would be a test that cannot fail for the stated reason.
			{"Bearer", "bearer token is empty"},
		} {
			s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"text"}})
			ch19RespDo(t, s.URL(), ch19RespGoodBody(), tc.auth)
			if got := s.Observations().Violations; !ch19RespAnyContains(got, tc.want) {
				t.Errorf("auth %q: violations = %q, want one containing %q", tc.auth, got, tc.want)
			}
			s.Close()
		}
	})
}

// 4. The ordered event sequence for the scenario the chapter is built around.
//
// This asserts the full list rather than spot-checking, because the bug it
// guards against is structural: an item opened and never closed, or closed in
// the wrong order, still produces every event type the spot-check would look
// for.
func TestCh19RespToolcallEventSequence(t *testing.T) {
	s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"summary_around_toolcall"}})
	defer s.Close()

	status, body := ch19RespDo(t, s.URL(), ch19RespGoodBody(), ch19RespTestAuth)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	events := ch19RespParseSSE(t, body)
	got := ch19RespEventNames(events)

	want := []string{
		"response.created",
		"response.in_progress",

		// reasoning item #1
		"response.output_item.added",
		"response.reasoning_summary_part.added",
		"response.reasoning_summary_text.delta",
		"response.reasoning_summary_text.delta",
		"response.reasoning_summary_text.delta",
		"response.reasoning_summary_text.delta",
		"response.reasoning_summary_text.done",
		"response.reasoning_summary_part.done",
		"response.output_item.done",

		// the tool call that interrupts it
		"response.output_item.added",
		"response.function_call_arguments.delta",
		"response.function_call_arguments.delta",
		"response.function_call_arguments.delta",
		"response.function_call_arguments.done",
		"response.output_item.done",

		// reasoning item #2 -- the whole point of the scenario
		"response.output_item.added",
		"response.reasoning_summary_part.added",
		"response.reasoning_summary_text.delta",
		"response.reasoning_summary_text.delta",
		"response.reasoning_summary_text.delta",
		"response.reasoning_summary_text.done",
		"response.reasoning_summary_part.done",
		"response.output_item.done",

		// the visible answer
		"response.output_item.added",
		"response.content_part.added",
		"response.output_text.delta",
		"response.output_text.delta",
		"response.output_text.delta",
		"response.output_text.delta",
		"response.output_text.done",
		"response.content_part.done",
		"response.output_item.done",

		"response.completed",
	}

	if len(got) != len(want) {
		t.Fatalf("got %d events, want %d\ngot:  %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("event %d = %q, want %q\nfull: %v", i, got[i], want[i], got)
		}
	}

	// Independently of the literal list above, prove the property the scenario
	// exists to test: summary deltas on BOTH sides of the function_call item.
	fcIdx := -1
	for i, e := range events {
		if e.Name != "response.output_item.added" {
			continue
		}
		if item, ok := e.Data["item"].(map[string]any); ok && item["type"] == "function_call" {
			fcIdx = i
			break
		}
	}
	if fcIdx < 0 {
		t.Fatal("no output_item.added carrying an item of type function_call")
	}
	before, after := 0, 0
	for i, e := range events {
		if e.Name != "response.reasoning_summary_text.delta" {
			continue
		}
		if i < fcIdx {
			before++
		} else {
			after++
		}
	}
	if before < 3 {
		t.Errorf("summary deltas before the tool call = %d, want >= 3", before)
	}
	if after < 3 {
		t.Errorf("summary deltas AFTER the tool call = %d, want >= 3; "+
			"without these a student who drops summaries post-tool-call is undetectable", after)
	}

	// The arguments must only be valid JSON once concatenated, so that a
	// student parsing each delta independently cannot accidentally succeed.
	var args strings.Builder
	for _, e := range events {
		if e.Name == "response.function_call_arguments.delta" {
			d, _ := e.Data["delta"].(string)
			args.WriteString(d)
		}
	}
	if args.String() != ch19RespArgumentsFull {
		t.Errorf("concatenated arguments = %q, want %q", args.String(), ch19RespArgumentsFull)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(args.String()), &parsed); err != nil {
		t.Errorf("concatenated arguments are not valid JSON: %v", err)
	}
	for _, e := range events {
		if e.Name != "response.function_call_arguments.delta" {
			continue
		}
		d, _ := e.Data["delta"].(string)
		var scratch any
		if json.Unmarshal([]byte(d), &scratch) == nil {
			t.Errorf("argument delta %q parses as standalone JSON; "+
				"the split must force students to accumulate", d)
		}
	}

	// function_call items are useless without the identity needed to answer
	// them, so the fixture must supply both.
	item := events[fcIdx].Data["item"].(map[string]any)
	if item["name"] == "" || item["call_id"] == "" {
		t.Errorf("function_call item missing name/call_id: %+v", item)
	}

	// response.completed carries usage at the real nesting depth.
	last := events[len(events)-1]
	resp, ok := last.Data["response"].(map[string]any)
	if !ok {
		t.Fatalf("response.completed has no response object: %+v", last.Data)
	}
	if resp["status"] != "completed" {
		t.Errorf("completed status = %v, want \"completed\"", resp["status"])
	}
	if _, ok := resp["usage"].(map[string]any); !ok {
		t.Fatalf("response.completed has no usage block")
	}
}

// Usage numbers must be settable, and each must sit at the depth the live
// capture uses. A student reading any of them from the top level of usage
// silently gets zero and concludes the feature is broken, so depth is part of
// the contract and is asserted negatively as well as positively.
func TestCh19RespCompletedUsageShape(t *testing.T) {
	s := ch19NewFakeResponses(ch19RespOptions{
		Scenarios:        []string{"text"},
		InputTokens:      1200,
		OutputTokens:     34,
		CachedTokens:     1024,
		CacheWriteTokens: 128,
		ReasoningTokens:  52,
	})
	defer s.Close()

	_, body := ch19RespDo(t, s.URL(), ch19RespGoodBody(), ch19RespTestAuth)
	events := ch19RespParseSSE(t, body)
	last := events[len(events)-1]
	if last.Name != "response.completed" {
		t.Fatalf("last event = %q, want response.completed", last.Name)
	}
	usage := last.Data["response"].(map[string]any)["usage"].(map[string]any)

	if got := usage["input_tokens"].(float64); got != 1200 {
		t.Errorf("input_tokens = %v, want 1200", got)
	}
	if got := usage["output_tokens"].(float64); got != 34 {
		t.Errorf("output_tokens = %v, want 34", got)
	}
	if got := usage["total_tokens"].(float64); got != 1234 {
		t.Errorf("total_tokens = %v, want 1234", got)
	}

	inDetails, ok := usage["input_tokens_details"].(map[string]any)
	if !ok {
		t.Fatalf("usage has no input_tokens_details: %+v", usage)
	}
	if got := inDetails["cached_tokens"].(float64); got != 1024 {
		t.Errorf("cached_tokens = %v, want 1024", got)
	}
	// Cache reads and cache writes are separate fields that mean opposite
	// things on the bill; an agent that conflates them reports a saving where
	// it incurred a premium.
	if got, ok := inDetails["cache_write_tokens"].(float64); !ok || got != 128 {
		t.Errorf("cache_write_tokens = %v (present=%v), want 128", got, ok)
	}

	outDetails, ok := usage["output_tokens_details"].(map[string]any)
	if !ok {
		t.Fatalf("usage has no output_tokens_details: %+v", usage)
	}
	if got := outDetails["reasoning_tokens"].(float64); got != 52 {
		t.Errorf("reasoning_tokens = %v, want 52", got)
	}

	for _, wrong := range []string{"cached_tokens", "cache_write_tokens", "reasoning_tokens"} {
		if _, found := usage[wrong]; found {
			t.Errorf("%s must not also appear at the top level of usage", wrong)
		}
	}
}

// sequence_number is zero-based in the live capture. An agent that assumes the
// first event is #1 is off by one against the real API on every frame, which
// breaks gap detection and any resume-from-sequence logic.
func TestCh19RespSequenceNumbersAreZeroBased(t *testing.T) {
	s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"summary_then_text"}})
	defer s.Close()

	_, body := ch19RespDo(t, s.URL(), ch19RespGoodBody(), ch19RespTestAuth)
	events := ch19RespParseSSE(t, body)

	for i, e := range events {
		got, ok := e.Data["sequence_number"].(float64)
		if !ok {
			t.Fatalf("event %d (%s) has no sequence_number", i, e.Name)
		}
		if int(got) != i {
			t.Fatalf("event %d (%s) sequence_number = %v, want %d", i, e.Name, got, i)
		}
	}
}

// prompt_cache_options comes back normalised, with the ttl the vendor fills
// in -- but only when the request actually carried the field. Echoing a
// default would tell a student who never sent it that caching was configured.
func TestCh19RespPromptCacheOptionsEcho(t *testing.T) {
	t.Run("echoed_with_ttl_when_sent", func(t *testing.T) {
		s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"text"}})
		defer s.Close()

		_, body := ch19RespDo(t, s.URL(), ch19RespGoodBody(), ch19RespTestAuth)
		events := ch19RespParseSSE(t, body)
		resp := events[len(events)-1].Data["response"].(map[string]any)

		echo, ok := resp["prompt_cache_options"].(map[string]any)
		if !ok {
			t.Fatalf("response.completed has no prompt_cache_options echo: %+v", resp)
		}
		if echo["mode"] != "explicit" {
			t.Errorf("echoed mode = %v, want \"explicit\"", echo["mode"])
		}
		if echo["ttl"] != "30m" {
			t.Errorf("echoed ttl = %v, want \"30m\"", echo["ttl"])
		}
	})

	t.Run("absent_when_not_sent", func(t *testing.T) {
		s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"text"}})
		defer s.Close()

		b := ch19RespGoodBody()
		delete(b, "prompt_cache_options")
		_, body := ch19RespDo(t, s.URL(), b, ch19RespTestAuth)
		events := ch19RespParseSSE(t, body)
		resp := events[len(events)-1].Data["response"].(map[string]any)

		if _, found := resp["prompt_cache_options"]; found {
			t.Errorf("prompt_cache_options echoed for a request that never sent it")
		}
	})
}

// Two function_call items in one response, taken from a live capture. A
// student who keeps "the" pending call in one field instead of a map keyed by
// call_id loses the first; one who stops at the first output_item.done never
// sees the second. Both pass against a single-call fixture.
func TestCh19RespParallelToolcalls(t *testing.T) {
	s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"parallel_toolcalls"}})
	defer s.Close()

	_, body := ch19RespDo(t, s.URL(), ch19RespGoodBody(), ch19RespTestAuth)
	events := ch19RespParseSSE(t, body)

	type call struct {
		name, callID, args string
		outputIndex        int
	}
	var calls []call
	argsBy := map[string]string{}
	for _, e := range events {
		switch e.Name {
		case "response.function_call_arguments.delta":
			id, _ := e.Data["item_id"].(string)
			d, _ := e.Data["delta"].(string)
			argsBy[id] += d
		case "response.output_item.done":
			item, _ := e.Data["item"].(map[string]any)
			if item == nil || item["type"] != "function_call" {
				continue
			}
			idx, _ := e.Data["output_index"].(float64)
			calls = append(calls, call{
				name:        item["name"].(string),
				callID:      item["call_id"].(string),
				args:        item["arguments"].(string),
				outputIndex: int(idx),
			})
		}
	}

	if len(calls) != 2 {
		t.Fatalf("got %d function_call items, want 2 (events=%v)", len(calls), ch19RespEventNames(events))
	}
	if calls[0].callID == calls[1].callID {
		t.Errorf("both calls share call_id %q; they must be distinguishable", calls[0].callID)
	}
	if calls[0].outputIndex != 0 || calls[1].outputIndex != 1 {
		t.Errorf("output_index = %d,%d, want 0,1", calls[0].outputIndex, calls[1].outputIndex)
	}
	if calls[0].args != ch19RespArgumentsFull || calls[1].args != ch19RespArguments2Full {
		t.Errorf("arguments = %q,%q, want %q,%q",
			calls[0].args, calls[1].args, ch19RespArgumentsFull, ch19RespArguments2Full)
	}
	for _, c := range calls {
		var scratch map[string]any
		if err := json.Unmarshal([]byte(c.args), &scratch); err != nil {
			t.Errorf("call %q arguments are not valid JSON: %v", c.callID, err)
		}
	}

	// Each call must be fully closed before the next opens: the live capture
	// interleaves nothing, and an agent that assumes interleaving is possible
	// builds a far more complex accumulator than it needs.
	names := ch19RespEventNames(events)
	firstDone, secondAdded := -1, -1
	seenAdded := 0
	for i, n := range names {
		if n == "response.output_item.added" {
			seenAdded++
			if seenAdded == 2 {
				secondAdded = i
			}
		}
		if n == "response.output_item.done" && firstDone < 0 {
			firstDone = i
		}
	}
	if firstDone < 0 || secondAdded < 0 || firstDone > secondAdded {
		t.Errorf("first call did not close before the second opened (firstDone=%d secondAdded=%d)",
			firstDone, secondAdded)
	}
}

// A reasoning item with NO summary events, and a non-zero reasoning_tokens
// count. This is correct vendor behaviour reproduced from a live capture, not
// a malfunction: an agent that waits for a summary part before rendering, or
// that asserts summaries whenever reasoning_tokens > 0, breaks here.
func TestCh19RespReasoningWithoutSummary(t *testing.T) {
	s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"reasoning_no_summary"}})
	defer s.Close()

	_, body := ch19RespDo(t, s.URL(), ch19RespGoodBody(), ch19RespTestAuth)
	events := ch19RespParseSSE(t, body)

	for _, e := range events {
		if strings.HasPrefix(e.Name, "response.reasoning_summary") {
			t.Fatalf("unexpected summary event %q in the no-summary scenario", e.Name)
		}
	}

	sawReasoningItem := false
	for _, e := range events {
		if e.Name != "response.output_item.done" {
			continue
		}
		item, _ := e.Data["item"].(map[string]any)
		if item == nil || item["type"] != "reasoning" {
			continue
		}
		sawReasoningItem = true
		sum, ok := item["summary"].([]any)
		if !ok || len(sum) != 0 {
			t.Errorf("reasoning item summary = %#v, want an empty array", item["summary"])
		}
	}
	if !sawReasoningItem {
		t.Fatal("no reasoning item in the no-summary scenario")
	}

	// The contradiction is the point: tokens were spent on reasoning and no
	// summary was produced.
	usage := events[len(events)-1].Data["response"].(map[string]any)["usage"].(map[string]any)
	rt := usage["output_tokens_details"].(map[string]any)["reasoning_tokens"].(float64)
	if rt == 0 {
		t.Error("reasoning_tokens = 0; the scenario must reproduce spend-without-summary")
	}
}

// Summaries arrive in multiple PARTS. The boundary is visible only through
// summary_index, so this is what catches a renderer that concatenates three
// distinct thoughts into one run-on paragraph.
func TestCh19RespMultipartSummary(t *testing.T) {
	s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"multipart_summary"}})
	defer s.Close()

	_, body := ch19RespDo(t, s.URL(), ch19RespGoodBody(), ch19RespTestAuth)
	events := ch19RespParseSSE(t, body)

	byPart := map[int][]string{}
	var partOrder []int
	for _, e := range events {
		if e.Name != "response.reasoning_summary_text.delta" {
			continue
		}
		idx, ok := e.Data["summary_index"].(float64)
		if !ok {
			t.Fatalf("summary delta has no summary_index: %+v", e.Data)
		}
		if _, seen := byPart[int(idx)]; !seen {
			partOrder = append(partOrder, int(idx))
		}
		d, _ := e.Data["delta"].(string)
		byPart[int(idx)] = append(byPart[int(idx)], d)
	}

	if len(byPart) != 3 {
		t.Fatalf("got %d summary parts, want 3", len(byPart))
	}
	// Indices must be 0,1,2 in order; a renderer keying on them needs them
	// dense and ascending.
	for i, p := range partOrder {
		if p != i {
			t.Errorf("part %d has summary_index %d, want %d", i, p, i)
		}
	}
	want := []string{ch19RespSummaryAFull, ch19RespSummaryBFull, ch19RespSummaryCFull}
	for i, w := range want {
		if got := strings.Join(byPart[i], ""); got != w {
			t.Errorf("part %d = %q, want %q", i, got, w)
		}
	}

	// Each part must be opened and closed in its own right, or a consumer
	// cannot tell where one thought ends and the next begins.
	added, done := 0, 0
	for _, e := range events {
		switch e.Name {
		case "response.reasoning_summary_part.added":
			added++
		case "response.reasoning_summary_part.done":
			done++
		}
	}
	if added != 3 || done != 3 {
		t.Errorf("summary part added/done = %d/%d, want 3/3", added, done)
	}
}

// The encrypted reasoning blob is the only way reasoning survives a turn when
// store:false, so it must appear when the request asked for it and must not
// be invented when it did not.
func TestCh19RespEncryptedReasoningEcho(t *testing.T) {
	for _, tc := range []struct {
		name    string
		include bool
	}{{"requested", true}, {"not_requested", false}} {
		t.Run(tc.name, func(t *testing.T) {
			s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"summary_then_text"}})
			defer s.Close()

			b := ch19RespGoodBody()
			if !tc.include {
				delete(b, "include")
			}
			_, body := ch19RespDo(t, s.URL(), b, ch19RespTestAuth)

			found := false
			for _, e := range ch19RespParseSSE(t, body) {
				if e.Name != "response.output_item.done" {
					continue
				}
				item, _ := e.Data["item"].(map[string]any)
				if item == nil || item["type"] != "reasoning" {
					continue
				}
				if _, ok := item["encrypted_content"]; ok {
					found = true
				}
			}
			if found != tc.include {
				t.Errorf("encrypted_content present = %v, want %v", found, tc.include)
			}
		})
	}
}

// 5. Summaries arrive as separate delta events and concatenate to the phrase.
// One event carrying the whole paragraph would satisfy any assertion about the
// final text while destroying the incremental delivery the chapter is about,
// so the count is asserted alongside the content.
func TestCh19RespSummaryDeltasAreIncremental(t *testing.T) {
	s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"summary_then_text"}})
	defer s.Close()

	_, body := ch19RespDo(t, s.URL(), ch19RespGoodBody(), ch19RespTestAuth)
	events := ch19RespParseSSE(t, body)

	var deltas []string
	for _, e := range events {
		if e.Name == "response.reasoning_summary_text.delta" {
			d, ok := e.Data["delta"].(string)
			if !ok {
				t.Fatalf("summary delta event has no string delta: %+v", e.Data)
			}
			deltas = append(deltas, d)
		}
	}
	if len(deltas) < 4 {
		t.Fatalf("got %d summary delta events, want >= 4", len(deltas))
	}
	if joined := strings.Join(deltas, ""); joined != ch19RespSummaryAFull {
		t.Errorf("concatenated summary = %q, want %q", joined, ch19RespSummaryAFull)
	}

	// The .done event must report the same full text the deltas built, or a
	// renderer that trusts .done and one that accumulates deltas disagree.
	for _, e := range events {
		if e.Name == "response.reasoning_summary_text.done" {
			if e.Data["text"] != ch19RespSummaryAFull {
				t.Errorf("summary .done text = %v, want %q", e.Data["text"], ch19RespSummaryAFull)
			}
		}
	}

	// Summary text must never be delivered as output_text: that is exactly the
	// bug where a student concatenates reasoning into the assistant's answer.
	var outText strings.Builder
	for _, e := range events {
		if e.Name == "response.output_text.delta" {
			d, _ := e.Data["delta"].(string)
			outText.WriteString(d)
		}
	}
	if outText.String() != ch19RespTextFull {
		t.Errorf("output text = %q, want %q", outText.String(), ch19RespTextFull)
	}
	if strings.Contains(outText.String(), ch19RespSummaryAFull) {
		t.Errorf("summary text leaked into output_text deltas")
	}

	// The raw-reasoning event family is a different thing from the summary
	// family and this fixture does not emit it; if it ever starts, the
	// renderer contract changes and this test should be the one to say so.
	for _, e := range events {
		if strings.HasPrefix(e.Name, "response.reasoning_text.") {
			t.Errorf("unexpected raw reasoning event %q", e.Name)
		}
	}
}

// 6. The documented per-request write limit is four breakpoints. Four is fine;
// five is a violation. Testing both sides pins the boundary -- an off-by-one
// here would either waste a student's budget or let them overspend silently.
func TestCh19RespBreakpointLimit(t *testing.T) {
	bodyWithBreakpoints := func(n int) map[string]any {
		b := ch19RespGoodBody()
		input := []any{}
		for i := 0; i < n; i++ {
			input = append(input, map[string]any{
				"role": "developer",
				"content": []any{
					map[string]any{
						"type":                    "input_text",
						"text":                    "block",
						"prompt_cache_breakpoint": map[string]any{"type": "prompt_cache_breakpoint"},
					},
				},
			})
		}
		b["input"] = input
		return b
	}

	for _, tc := range []struct {
		n       int
		violate bool
	}{{4, false}, {5, true}} {
		s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"text"}})
		ch19RespDo(t, s.URL(), bodyWithBreakpoints(tc.n), ch19RespTestAuth)
		obs := s.Observations()
		s.Close()

		if got := obs.Requests[0].BreakpointCount; got != tc.n {
			t.Errorf("n=%d: BreakpointCount = %d, want %d", tc.n, got, tc.n)
		}
		hit := ch19RespAnyContains(obs.Violations, "prompt_cache_breakpoint:")
		if hit != tc.violate {
			t.Errorf("n=%d: breakpoint violation = %v, want %v (violations=%q)",
				tc.n, hit, tc.violate, obs.Violations)
		}
	}

	// The count is a whole-tree walk, so a breakpoint buried somewhere other
	// than input[].content[] still spends budget and still counts.
	t.Run("nested_anywhere", func(t *testing.T) {
		s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"text"}})
		defer s.Close()
		b := ch19RespGoodBody()
		b["tools"] = []any{map[string]any{
			"type": "function", "name": "t",
			"deeply": map[string]any{"nested": []any{
				map[string]any{"prompt_cache_breakpoint": map[string]any{}},
			}},
		}}
		ch19RespDo(t, s.URL(), b, ch19RespTestAuth)
		// one from the good body's developer block, one from the tool tree
		if got := s.Observations().Requests[0].BreakpointCount; got != 2 {
			t.Errorf("BreakpointCount = %d, want 2", got)
		}
	})

	t.Run("bad_cache_mode", func(t *testing.T) {
		s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"text"}})
		defer s.Close()
		b := ch19RespGoodBody()
		b["prompt_cache_options"] = map[string]any{"mode": "aggressive"}
		ch19RespDo(t, s.URL(), b, ch19RespTestAuth)
		if got := s.Observations().Violations; !ch19RespAnyContains(got, "prompt_cache_options.mode") {
			t.Fatalf("violations = %q, want one naming prompt_cache_options.mode", got)
		}
	})
}

// 7. A secret in the body is a leak. The same secret in the Authorization
// header is the credential being used correctly, and must not be reported --
// otherwise the only way to pass is to stop authenticating.
func TestCh19RespSecretLeakProbe(t *testing.T) {
	const secret = "sk-ch19-super-secret-value"

	t.Run("in_body_is_detected", func(t *testing.T) {
		s := ch19NewFakeResponses(ch19RespOptions{
			Scenarios: []string{"text"},
			Secrets:   []string{secret},
		})
		defer s.Close()

		b := ch19RespGoodBody()
		b["input"].([]any)[1].(map[string]any)["content"] =
			[]any{map[string]any{"type": "input_text", "text": "my token is " + secret}}
		ch19RespDo(t, s.URL(), b, ch19RespTestAuth)

		obs := s.Observations()
		if !obs.SecretInBody {
			t.Fatal("SecretInBody = false, want true")
		}
		if len(obs.LeakedSecrets) != 1 || obs.LeakedSecrets[0] != secret {
			t.Fatalf("LeakedSecrets = %q, want [%q]", obs.LeakedSecrets, secret)
		}
		if !ch19RespAnyContains(obs.Violations, "secret leaked in request body") {
			t.Fatalf("violations = %q, want one reporting a body leak", obs.Violations)
		}
		// The violation text is grading output; it must not reprint the secret.
		for _, v := range obs.Violations {
			if strings.Contains(v, secret) {
				t.Errorf("violation %q reprints the secret verbatim", v)
			}
		}
	})

	t.Run("in_auth_header_only_is_not_a_leak", func(t *testing.T) {
		s := ch19NewFakeResponses(ch19RespOptions{
			Scenarios: []string{"text"},
			Secrets:   []string{secret},
		})
		defer s.Close()

		ch19RespDo(t, s.URL(), ch19RespGoodBody(), "Bearer "+secret)

		obs := s.Observations()
		if obs.SecretInBody {
			t.Error("SecretInBody = true for a secret that appeared only in the Authorization header")
		}
		if len(obs.LeakedSecrets) != 0 {
			t.Errorf("LeakedSecrets = %q, want empty", obs.LeakedSecrets)
		}
		if len(obs.Violations) != 0 {
			t.Errorf("violations = %q, want none", obs.Violations)
		}
	})
}

// 8. The model list contains a hidden entry on purpose.
func TestCh19RespModelsEndpoint(t *testing.T) {
	s := ch19NewFakeResponses(ch19RespOptions{})
	defer s.Close()

	req, err := http.NewRequest(http.MethodGet, s.URL()+"/v1/models", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Authorization", ch19RespTestAuth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var payload struct {
		Models []struct {
			ID          string `json:"id"`
			Object      string `json:"object"`
			Slug        string `json:"slug"`
			DisplayName string `json:"display_name"`
			Visibility  string `json:"visibility"`
		} `json:"models"`
		Data []json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// The SIWC catalogue is keyed "models", not the platform API's "data".
	// A client built for the wrong key decodes cleanly into an empty slice and
	// reports "no models available" instead of failing, so assert both that
	// the right key is populated and that the wrong one is absent.
	if len(payload.Models) == 0 {
		t.Fatal("no \"models\" array; the SIWC catalogue is keyed \"models\", not \"data\"")
	}
	if payload.Data != nil {
		t.Error("response carries a \"data\" key; this route uses \"models\"")
	}

	listed, hidden := 0, 0
	for _, m := range payload.Models {
		if m.ID == "" || m.Object != "model" || m.Slug == "" || m.DisplayName == "" {
			t.Errorf("incomplete model entry: %+v", m)
		}
		switch m.Visibility {
		case "list":
			listed++
		case "hidden":
			hidden++
		default:
			t.Errorf("unexpected visibility %q", m.Visibility)
		}
	}
	if listed < 1 {
		t.Error("no model with visibility \"list\"")
	}
	if hidden < 1 {
		t.Error("no model with visibility \"hidden\"; a student who forgets to filter would be undetectable")
	}

	obs := s.Observations()
	if obs.ModelsHits != 1 {
		t.Errorf("ModelsHits = %d, want 1", obs.ModelsHits)
	}
	if len(obs.ModelsAuth) != 1 || obs.ModelsAuth[0] != ch19RespTestAuth {
		t.Errorf("ModelsAuth = %q, want [%q]", obs.ModelsAuth, ch19RespTestAuth)
	}
	// Model listing is not a Responses call and must not be filed as one.
	if len(obs.Requests) != 0 {
		t.Errorf("Requests = %d, want 0; /v1/models must not pollute the Responses log", len(obs.Requests))
	}
}

// 9. Error paths: a plan-level rejection arrives as a non-2xx HTTP response
// with a JSON code, before any streaming. The in-stream variant is separate,
// because an agent can handle one and not the other.
func TestCh19RespErrorScenarios(t *testing.T) {
	for _, tc := range []struct {
		code       string
		wantStatus int
	}{
		{"subscription_sharing_usage_limit_exceeded", http.StatusTooManyRequests},
		{"subscription_sharing_usage_unavailable", http.StatusServiceUnavailable},
		{"subscription_sharing_unsupported_capability", http.StatusBadRequest},
		{"subscription_sharing_user_not_eligible", http.StatusForbidden},
		{"subscription_sharing_route_not_supported", http.StatusForbidden},
		{"subscription_sharing_invalid_user", http.StatusUnauthorized},
		{"subscription_sharing_user_unavailable", http.StatusServiceUnavailable},
		{"chatpass_v2_scope_not_authorized", http.StatusForbidden},
		{"chatpass_v2_invalid_authorization_context", http.StatusForbidden},
	} {
		t.Run(tc.code, func(t *testing.T) {
			s := ch19NewFakeResponses(ch19RespOptions{
				Scenarios: []string{"text"},
				FailWith:  tc.code,
			})
			defer s.Close()

			status, body := ch19RespDo(t, s.URL(), ch19RespGoodBody(), ch19RespTestAuth)
			if status != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", status, tc.wantStatus, body)
			}
			if status >= 200 && status < 300 {
				t.Fatalf("status %d is 2xx; a rejection must not look like success", status)
			}
			var payload struct {
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
					Type    string `json:"type"`
				} `json:"error"`
			}
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("error body is not JSON: %v\n%s", err, body)
			}
			if payload.Error.Code != tc.code {
				t.Errorf("error.code = %q, want %q", payload.Error.Code, tc.code)
			}
			if payload.Error.Message == "" || payload.Error.Type == "" {
				t.Errorf("error body missing message/type: %+v", payload.Error)
			}
			// Validation still ran: a rejected request is still observed.
			if len(s.Observations().Requests) != 1 {
				t.Errorf("a rejected request must still be recorded")
			}
		})
	}

	// FailAfter lets the parent exercise a mid-session failure, which is where
	// token-refresh and retry bugs actually live.
	t.Run("fail_after", func(t *testing.T) {
		s := ch19NewFakeResponses(ch19RespOptions{
			Scenarios: []string{"text"},
			FailWith:  "subscription_sharing_usage_limit_exceeded",
			FailAfter: 1,
		})
		defer s.Close()

		if status, _ := ch19RespDo(t, s.URL(), ch19RespGoodBody(), ch19RespTestAuth); status != http.StatusOK {
			t.Fatalf("first request status = %d, want 200", status)
		}
		if status, _ := ch19RespDo(t, s.URL(), ch19RespGoodBody(), ch19RespTestAuth); status != http.StatusTooManyRequests {
			t.Fatalf("second request status = %d, want 429", status)
		}
	})

	// The in-stream variant: HTTP 200, a clean open, then failure. An agent
	// that treats response.created or a non-empty body as success reports a
	// fabricated answer to its user here.
	t.Run("mid_stream_error", func(t *testing.T) {
		s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"error"}})
		defer s.Close()

		status, body := ch19RespDo(t, s.URL(), ch19RespGoodBody(), ch19RespTestAuth)
		if status != http.StatusOK {
			t.Fatalf("status = %d, want 200", status)
		}
		names := ch19RespEventNames(ch19RespParseSSE(t, body))
		want := []string{"response.created", "response.error", "response.failed"}
		if strings.Join(names, ",") != strings.Join(want, ",") {
			t.Fatalf("events = %v, want %v", names, want)
		}
	})
}

// Scenario selection is per-request and sticks on the last entry, so a test can
// script a short prefix and let an agent loop as long as it needs to.
func TestCh19RespScenarioSequencing(t *testing.T) {
	s := ch19NewFakeResponses(ch19RespOptions{
		Scenarios: []string{"text", "summary_then_text"},
	})
	defer s.Close()

	wantHasSummary := []bool{false, true, true, true}
	for i, want := range wantHasSummary {
		_, body := ch19RespDo(t, s.URL(), ch19RespGoodBody(), ch19RespTestAuth)
		names := ch19RespEventNames(ch19RespParseSSE(t, body))
		got := false
		for _, n := range names {
			if n == "response.reasoning_summary_text.delta" {
				got = true
			}
		}
		if got != want {
			t.Errorf("request %d: hasSummary = %v, want %v (events=%v)", i, got, want, names)
		}
	}

	// The default, with no Scenarios configured at all, must still produce
	// summaries -- that is the chapter's subject, and a silent default of
	// "text" would make the interesting path opt-in.
	d := ch19NewFakeResponses(ch19RespOptions{})
	defer d.Close()
	_, body := ch19RespDo(t, d.URL(), ch19RespGoodBody(), ch19RespTestAuth)
	if !strings.Contains(string(body), "response.reasoning_summary_text.delta") {
		t.Error("default scenario produced no reasoning summary deltas")
	}
}

// Observations must hand back a deep copy. A caller that mutates the returned
// tree, or a request that lands afterwards, must not disturb what an earlier
// caller is holding.
func TestCh19RespObservationsAreDeepCopied(t *testing.T) {
	s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"text"}})
	defer s.Close()

	ch19RespDo(t, s.URL(), ch19RespGoodBody(), ch19RespTestAuth)

	first := s.Observations()
	first.Requests[0].Decoded["model"] = "mutated-by-caller"
	first.Requests[0].Body[0] = 'X'
	first.Violations = append(first.Violations, "injected")

	second := s.Observations()
	if second.Requests[0].Decoded["model"] != "gpt-5.6" {
		t.Errorf("Decoded was shared with the caller: model = %v",
			second.Requests[0].Decoded["model"])
	}
	if second.Requests[0].Body[0] != '{' {
		t.Errorf("Body was shared with the caller: first byte = %q", second.Requests[0].Body[0])
	}
	if len(second.Violations) != 0 {
		t.Errorf("Violations was shared with the caller: %q", second.Violations)
	}
}

func ch19RespAnyContains(hay []string, needle string) bool {
	for _, h := range hay {
		if strings.Contains(h, needle) {
			return true
		}
	}
	return false
}

// The 400 rejection names the offending field in error.param. That param is
// the most actionable part of the error -- it tells the student which field to
// delete -- and only this code carries it.
func TestCh19RespUnsupportedCapabilityCarriesParam(t *testing.T) {
	s := ch19NewFakeResponses(ch19RespOptions{
		Scenarios: []string{"text"},
		FailWith:  "subscription_sharing_unsupported_capability",
	})
	defer s.Close()

	status, body := ch19RespDo(t, s.URL(), ch19RespGoodBody(), ch19RespTestAuth)
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", status)
	}
	var payload struct {
		Error struct {
			Code  string `json:"code"`
			Param string `json:"param"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload.Error.Param == "" {
		t.Error("unsupported_capability rejection has no error.param naming the offending field")
	}
}

// Status is derived from the code, so the fixture cannot be asked to produce a
// code/status pairing the vendor never emits. This also guards the table
// against drift: every documented code must map to a non-zero status.
func TestCh19RespFailStatusTableIsComplete(t *testing.T) {
	want := map[string]int{
		"subscription_sharing_usage_limit_exceeded":   429,
		"subscription_sharing_usage_unavailable":      503,
		"subscription_sharing_unsupported_capability": 400,
		"subscription_sharing_user_not_eligible":      403,
		"subscription_sharing_route_not_supported":    403,
		"subscription_sharing_invalid_user":           401,
		"subscription_sharing_user_unavailable":       503,
		"chatpass_v2_scope_not_authorized":            403,
		"chatpass_v2_invalid_authorization_context":   403,
	}
	if len(ch19RespFailStatus) != len(want) {
		t.Errorf("ch19RespFailStatus has %d codes, want %d", len(ch19RespFailStatus), len(want))
	}
	for code, status := range want {
		got, ok := ch19RespFailStatus[code]
		if !ok {
			t.Errorf("missing documented code %q", code)
			continue
		}
		if got != status {
			t.Errorf("%s maps to %d, want %d", code, got, status)
		}
	}
}

// A usage limit that strikes after HTTP 200 and after text has reached the
// screen. The status line already said success, so the only correct reading is
// that response.completed is the success signal and its absence is a failure.
func TestCh19RespFailMidstream(t *testing.T) {
	const code = "subscription_sharing_usage_limit_exceeded"
	s := ch19NewFakeResponses(ch19RespOptions{
		Scenarios: []string{"fail_midstream"},
		FailWith:  code,
	})
	defer s.Close()

	status, body := ch19RespDo(t, s.URL(), ch19RespGoodBody(), ch19RespTestAuth)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200; the whole point is that the status lies", status)
	}

	events := ch19RespParseSSE(t, body)
	names := ch19RespEventNames(events)

	// Text really was delivered first: a fail-fast variant would let a student
	// pass by never rendering anything before the terminal event.
	deltas := 0
	for _, n := range names {
		if n == "response.output_text.delta" {
			deltas++
		}
	}
	if deltas < 3 {
		t.Errorf("got %d output_text.delta events before the failure, want >= 3", deltas)
	}

	for _, n := range names {
		if n == "response.completed" {
			t.Fatal("response.completed must not appear in a failed stream")
		}
	}

	last := events[len(events)-1]
	if last.Name != "response.failed" {
		t.Fatalf("terminal event = %q, want response.failed", last.Name)
	}
	resp, ok := last.Data["response"].(map[string]any)
	if !ok {
		t.Fatalf("response.failed has no response object: %+v", last.Data)
	}
	if resp["status"] != "failed" {
		t.Errorf("status = %v, want \"failed\"", resp["status"])
	}
	errObj, ok := resp["error"].(map[string]any)
	if !ok {
		t.Fatalf("response.failed carries no error object: %+v", resp)
	}
	if errObj["code"] != code {
		t.Errorf("error.code = %v, want %q", errObj["code"], code)
	}
}

// Direct-admission failures carry {"detail": ...} and NO error object. A
// parser that reaches for error.code unconditionally breaks on exactly the
// failure it most needs to report, and no other path in this fixture would
// catch that, because every other error shape does have error.code.
func TestCh19RespFailAdmission(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusServiceUnavailable} {
		s := ch19NewFakeResponses(ch19RespOptions{
			Scenarios:       []string{"fail_admission"},
			AdmissionStatus: status,
		})

		got, body := ch19RespDo(t, s.URL(), ch19RespGoodBody(), ch19RespTestAuth)
		s.Close()

		if got != status {
			t.Errorf("status = %d, want %d", got, status)
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("admission body is not JSON: %v\n%s", err, body)
		}
		if _, ok := payload["detail"].(string); !ok {
			t.Errorf("admission body has no \"detail\" string: %s", body)
		}
		if _, ok := payload["error"]; ok {
			t.Errorf("admission failure must NOT carry an error object: %s", body)
		}
		if bytes.Contains(body, []byte("event:")) {
			t.Errorf("admission failure must not stream any SSE: %s", body)
		}
	}

	// Zero means "use the documented default" rather than "send HTTP 0".
	t.Run("defaults_to_503", func(t *testing.T) {
		s := ch19NewFakeResponses(ch19RespOptions{Scenarios: []string{"fail_admission"}})
		defer s.Close()
		got, _ := ch19RespDo(t, s.URL(), ch19RespGoodBody(), ch19RespTestAuth)
		if got != http.StatusServiceUnavailable {
			t.Errorf("status = %d, want 503", got)
		}
	})
}
