package grade

// Chapter 19 fixture: an impersonation of OpenAI's Responses API, strict
// enough to grade a student migrating an agent off Chat Completions.
//
// The fixture has two jobs, and they pull in opposite directions. As a
// VALIDATOR it must be pedantic, because the ChatGPT-plan route the chapter
// targets is a narrow doorway: a field that is merely *present* can get the
// request rejected by the real vendor, so the grader has to notice presence,
// not just wrong values. As a SERVER it must be generous, because the student
// needs a stream realistic enough to drive a real renderer — reasoning
// summaries arriving as deltas, a tool call interrupting them, and summaries
// resuming afterwards.
//
// Every streaming event name below was taken from the generated OpenAI SDK
// types (see cr/docs/ch19-research-notes.md), not from memory. Inventing a
// plausible-looking event name would produce a fixture that grades students
// against an API that does not exist, which is the worst failure mode
// available to a grader: silent, confident, and wrong in the student's favour.

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

// ch19RespOptions configures one fixture instance.
//
// Scenarios is consumed in order, one entry per POST /v1/responses, and the
// last entry repeats once exhausted. That "sticky last" rule exists so a test
// can say Scenarios: []string{"summary_around_toolcall"} and have every turn
// of a multi-turn agent loop behave the same way without the test having to
// predict how many turns the student's agent will take.
type ch19RespOptions struct {
	Scenarios []string // per-request script selection; last entry repeats
	FailWith  string   // ChatGPT-plan error code; status is derived, not supplied
	FailAfter int      // successful requests to allow first; 0 = fail immediately
	Secrets   []string // leak canaries; see the secret probe in recordRequest

	// AdmissionStatus is the HTTP status for the "fail_admission" scenario.
	// Direct-admission failures happen before the request reaches the model,
	// so they do NOT carry the standard error object -- the docs warn they
	// come back as {"detail": "..."} instead. Defaults to 503 when zero;
	// 401 and 403 are the other documented values.
	AdmissionStatus int

	// Usage numbers echoed in response.completed. The parent asserts that the
	// student's agent records cache reads, cache WRITES and reasoning tokens,
	// none of which it can do unless these are settable and distinguishable
	// from each other. Cache reads and writes are separate fields on the real
	// API and mean opposite things economically -- a write costs more than an
	// uncached call, a read costs a fraction of one -- so an agent that adds
	// them together, or reports one as the other, is wrong in a way that only
	// shows up on the bill.
	InputTokens      int
	OutputTokens     int
	CachedTokens     int
	CacheWriteTokens int
	ReasoningTokens  int
}

// ch19RespMsgBP is the per-message cache-breakpoint report.
//
// This exists to settle one specific question the parent cannot answer from a
// raw body without reimplementing the walk: did the student attach the
// constitution to a developer-role *content block* (cacheable) or to the
// top-level instructions string (not cacheable, because a string has no block
// to hang a breakpoint on)? BlockCount == 0 with a non-empty message is the
// signature of content sent as a bare string, which is the near-miss version
// of the same mistake.
type ch19RespMsgBP struct {
	Role          string
	BlockCount    int
	HasBreakpoint bool
}

// ch19RespReq is everything the fixture noticed about one POST /v1/responses.
//
// Body and Decoded are both retained deliberately. Decoded answers "is this
// key present?", which a typed struct cannot: encoding/json silently discards
// fields with no matching struct member, and on this route an unexpected field
// is precisely what a banned field looks like. Body is kept because the secret
// probe must search the bytes the student actually sent, including inside
// strings that decoding would restructure.
type ch19RespReq struct {
	Body    []byte
	Decoded map[string]any
	Auth    string
	Roles   []string // roles of input[] elements, in order

	// HasInstructions reports key presence only, not usefulness. The parent
	// reads Decoded["instructions"] when it needs the value; conflating
	// "present" with "non-empty" here would hide a student who sets the field
	// to "" and believes they have sent a constitution.
	HasInstructions bool

	BreakpointCount    int // prompt_cache_breakpoint occurrences, whole tree
	MessageBreakpoints []ch19RespMsgBP

	IncludesEncryptedReasoning bool // include[] contains reasoning.encrypted_content

	// ReasoningSummary is "" both when reasoning is absent and when it is
	// present without a summary. The two are distinguishable via Decoded; they
	// are not worth a second bool here because the parent that cares about the
	// difference is already reading Decoded for the rest of the reasoning block.
	ReasoningSummary string

	// StoreValue and StreamValue carry the decoded value, or nil when the key
	// was absent. nil is distinguishable from JSON false, which decodes to a
	// non-nil any holding bool(false) -- that distinction is the whole reason
	// these are any and not bool.
	StoreValue  any
	StreamValue any

	PromptCacheMode string
}

// ch19RespObs is the deep-copied report handed back by Observations().
type ch19RespObs struct {
	Requests   []ch19RespReq
	Violations []string
	ModelsHits int

	// ModelsAuth records the Authorization header of each GET /v1/models call.
	// It lives here rather than in Requests because Requests means "requests to
	// the Responses endpoint" to every violation test, and quietly widening that
	// meaning would make request-count assertions start failing for reasons
	// unrelated to what they were written to check.
	ModelsAuth []string

	SecretInBody  bool
	LeakedSecrets []string
}

// ch19FakeResponses is the server. Zero value is not usable; call
// ch19NewFakeResponses.
type ch19FakeResponses struct {
	mu   sync.Mutex
	opts ch19RespOptions
	obs  ch19RespObs
	srv  *httptest.Server

	scenarioIdx int // next index into opts.Scenarios
	okCount     int // successful (non-failed) responses served, for FailAfter
	seq         int // SSE sequence_number, monotonic across the process
	respID      int // response id counter
}

// ch19RespBannedFields is the unsupported-field list from OpenAI's first-party
// SIWC preview-limitations documentation, verbatim and in full.
//
// Two corrections are baked in here, and both directions matter. Fields that
// belong to Chat Completions and do not exist on the Responses API at all --
// n, presence_penalty, frequency_penalty, logprobs, stop -- are NOT listed:
// banning them would invent a rule and fail a student for an imaginary
// offence. And it is top_logprobs that is unsupported, not logprobs, which is
// exactly the kind of near-miss a list reconstructed from memory gets wrong.
//
// Note that prompt_cache_retention is unsupported while prompt_cache_options
// is accepted and echoed back; the names are close enough that a student who
// skims will reach for the wrong one.
//
// A slice rather than a map so violation order is deterministic: a grader that
// reports the same problems in a different order on every run teaches students
// to distrust it. These fields all EXIST on the Responses API -- the
// prohibition is plan-level, not API-level.
var ch19RespBannedFields = []string{
	"background",
	"conversation",
	"max_output_tokens",
	"max_tool_calls",
	"metadata",
	"moderation",
	"multi_agent",
	"prompt",
	"prompt_cache_retention",
	"safety_identifier",
	"temperature",
	"top_logprobs",
	"top_p",
	"truncation",
	"user",
}

// ch19RespFailStatus maps each documented ChatGPT-plan error code to the HTTP
// status the docs pair it with. The caller supplies only the code; making them
// supply the status too would let a test assert a pairing the vendor never
// produces, which is how a fixture drifts away from the thing it impersonates.
//
// The split across statuses drives the student's recovery logic, and the
// classes need different handling: 429/503 are "retry with backoff", 401 is
// "refresh the token or re-authenticate", 403 is "stop, this account or route
// will never work". A student who lumps them together either hammers a
// permanent failure or abandons a recoverable one.
var ch19RespFailStatus = map[string]int{
	"subscription_sharing_usage_limit_exceeded":   http.StatusTooManyRequests,    // 429
	"subscription_sharing_usage_unavailable":      http.StatusServiceUnavailable, // 503
	"subscription_sharing_unsupported_capability": http.StatusBadRequest,         // 400, carries error.param
	"subscription_sharing_user_not_eligible":      http.StatusForbidden,          // 403
	"subscription_sharing_route_not_supported":    http.StatusForbidden,          // 403
	"subscription_sharing_invalid_user":           http.StatusUnauthorized,       // 401
	"subscription_sharing_user_unavailable":       http.StatusServiceUnavailable, // 503
	"chatpass_v2_scope_not_authorized":            http.StatusForbidden,          // 403
	"chatpass_v2_invalid_authorization_context":   http.StatusForbidden,          // 403
}

func ch19NewFakeResponses(o ch19RespOptions) *ch19FakeResponses {
	s := &ch19FakeResponses{opts: o}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/responses", s.ch19RespHandleResponses)
	mux.HandleFunc("/v1/models", s.ch19RespHandleModels)
	s.srv = httptest.NewServer(mux)
	return s
}

func (s *ch19FakeResponses) URL() string { return s.srv.URL }

func (s *ch19FakeResponses) Close() { s.srv.Close() }

// Observations returns a deep copy under the lock.
//
// Deep, not shallow: a shallow copy hands the caller the live Decoded maps and
// Body slices, which a concurrently-served request can still be appending to.
// The resulting flake would appear in the parent's assertions, far from here,
// and would look like a student bug rather than a fixture bug.
func (s *ch19FakeResponses) Observations() ch19RespObs {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := ch19RespObs{
		ModelsHits:    s.obs.ModelsHits,
		SecretInBody:  s.obs.SecretInBody,
		Violations:    append([]string(nil), s.obs.Violations...),
		ModelsAuth:    append([]string(nil), s.obs.ModelsAuth...),
		LeakedSecrets: append([]string(nil), s.obs.LeakedSecrets...),
	}
	for _, r := range s.obs.Requests {
		cp := r
		cp.Body = append([]byte(nil), r.Body...)
		cp.Roles = append([]string(nil), r.Roles...)
		cp.MessageBreakpoints = append([]ch19RespMsgBP(nil), r.MessageBreakpoints...)
		if r.Decoded != nil {
			cp.Decoded, _ = ch19RespDeepCopy(r.Decoded).(map[string]any)
		}
		out.Requests = append(out.Requests, cp)
	}
	return out
}

// ch19RespDeepCopy clones a decoded JSON tree. Only the four shapes
// encoding/json can produce into an any are handled; anything else is a scalar
// and is safe to share.
func ch19RespDeepCopy(v any) any {
	switch t := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(t))
		for k, val := range t {
			m[k] = ch19RespDeepCopy(val)
		}
		return m
	case []any:
		a := make([]any, len(t))
		for i, val := range t {
			a[i] = ch19RespDeepCopy(val)
		}
		return a
	default:
		return v
	}
}

// violate appends a violation. Caller holds the lock.
func (s *ch19FakeResponses) violate(format string, args ...any) {
	s.obs.Violations = append(s.obs.Violations, fmt.Sprintf(format, args...))
}

// ---------------------------------------------------------------------------
// POST /v1/responses
// ---------------------------------------------------------------------------

func (s *ch19FakeResponses) ch19RespHandleResponses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.mu.Lock()
		s.violate("method: POST required on /v1/responses, got %s", r.Method)
		s.mu.Unlock()
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.mu.Lock()
		s.violate("body: unreadable: %v", err)
		s.mu.Unlock()
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.ch19RespRecord(body, r.Header.Get("Authorization"))
	scenario := s.ch19RespNextScenario()
	failCode := ""
	if s.opts.FailWith != "" && s.okCount >= s.opts.FailAfter {
		failCode = s.opts.FailWith
	}
	// The response echoes choices the REQUEST made, so they are read back off
	// the record we just filed rather than from Options. A fixture that echoed
	// a configured cache mode instead of the received one would report success
	// to a student who never sent the field.
	last := s.obs.Requests[len(s.obs.Requests)-1]
	cfg := &ch19RespStreamCfg{
		cacheMode:        last.PromptCacheMode,
		includeEncrypted: last.IncludesEncryptedReasoning,
		in:               s.opts.InputTokens,
		out:              s.opts.OutputTokens,
		cached:           s.opts.CachedTokens,
		cacheWrite:       s.opts.CacheWriteTokens,
		reasoningTok:     s.opts.ReasoningTokens,
	}
	// An explicitly scripted failure scenario outranks the FailAfter counter:
	// the caller asked for that shape on this request, so the counter must not
	// quietly convert it into a different kind of failure.
	scripted := scenario == "fail_midstream" || scenario == "fail_admission"
	preStreamFail := failCode != "" && !scripted
	if failCode == "" && !scripted {
		s.okCount++
	}
	admissionStatus := s.opts.AdmissionStatus
	s.mu.Unlock()

	if scenario == "fail_admission" {
		// Rejected before reaching the model: no SSE, no error object.
		s.ch19RespWriteAdmissionFailure(w, admissionStatus)
		return
	}

	// The pre-stream failure path is deliberately taken BEFORE any header is
	// written. A plan-level rejection from the real vendor arrives as an HTTP
	// error with a JSON body, not as an error event inside a 200 stream, and a
	// student whose error handling only inspects SSE frames must fail here.
	if preStreamFail {
		s.ch19RespWriteFailure(w, failCode)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		// Buffering would make incremental delivery untestable while still
		// producing a byte-identical final body, so every streaming assertion
		// would pass against a fixture that does not actually stream. Refusing
		// loudly is the only outcome that cannot be mistaken for success.
		panic("ch19FakeResponses: ResponseWriter does not implement http.Flusher; " +
			"the fixture cannot verify incremental delivery through a buffered writer")
	}
	cfg.w, cfg.f, cfg.id = w, flusher, s.ch19RespNewID()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	switch scenario {
	case "error":
		s.ch19RespStreamError(cfg)
	case "fail_midstream":
		s.ch19RespStreamFailMidstream(cfg, failCode)
	case "text":
		s.ch19RespStreamText(cfg)
	case "summary_around_toolcall":
		s.ch19RespStreamToolcall(cfg)
	case "parallel_toolcalls":
		s.ch19RespStreamParallelToolcalls(cfg)
	case "reasoning_no_summary":
		s.ch19RespStreamReasoningNoSummary(cfg)
	case "multipart_summary":
		s.ch19RespStreamMultipartSummary(cfg)
	default: // "summary_then_text"
		s.ch19RespStreamSummaryThenText(cfg)
	}
}

// ch19RespNextScenario consumes one scenario, sticking on the last. Caller
// holds the lock.
func (s *ch19FakeResponses) ch19RespNextScenario() string {
	if len(s.opts.Scenarios) == 0 {
		return "summary_then_text"
	}
	i := s.scenarioIdx
	if i >= len(s.opts.Scenarios) {
		i = len(s.opts.Scenarios) - 1
	} else {
		s.scenarioIdx++
	}
	return s.opts.Scenarios[i]
}

// ---------------------------------------------------------------------------
// Validation
// ---------------------------------------------------------------------------

// ch19RespRecord validates one request and files both the observations and the
// violations. Caller holds the lock.
func (s *ch19FakeResponses) ch19RespRecord(body []byte, auth string) {
	req := ch19RespReq{Body: append([]byte(nil), body...), Auth: auth}

	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		s.violate("body: not a JSON object: %v", err)
		s.obs.Requests = append(s.obs.Requests, req)
		return
	}
	req.Decoded = decoded

	// store must be present AND exactly false. Absent is its own violation
	// because the ChatGPT plan forbids server-side retention and the API
	// default is store:true -- a student who simply omits the field gets
	// retention they were told not to have, and sees no error from the vendor.
	req.StoreValue = decoded["store"]
	if _, present := decoded["store"]; !present {
		s.violate("store: field absent (ChatGPT plan requires store:false)")
	} else if b, ok := decoded["store"].(bool); !ok || b {
		s.violate("store: must be false, got %v", decoded["store"])
	}

	// stream must be present AND exactly true. Non-streaming works against the
	// real API, so nothing downstream would break -- except the chapter's
	// entire point, which is incremental delivery of reasoning summaries.
	req.StreamValue = decoded["stream"]
	if _, present := decoded["stream"]; !present {
		s.violate("stream: field absent")
	} else if b, ok := decoded["stream"].(bool); !ok || !b {
		s.violate("stream: must be true, got %v", decoded["stream"])
	}

	if m, present := decoded["model"]; !present {
		s.violate("model: field absent")
	} else if str, ok := m.(string); !ok || str == "" {
		s.violate("model: must be a non-empty string")
	}

	// "messages" is called out separately from "input absent" because it is the
	// single most likely migration mistake, and because the two diagnoses lead
	// the student to different fixes: one is a missing field, the other is a
	// field that was renamed between APIs.
	if _, present := decoded["messages"]; present {
		s.violate("messages: Chat Completions key present; the Responses API uses input")
	}

	rawInput, inputPresent := decoded["input"]
	if !inputPresent {
		s.violate("input: field absent")
	} else if arr, ok := rawInput.([]any); !ok {
		s.violate("input: must be an array, got %T", rawInput)
	} else {
		s.ch19RespScanInput(&req, arr)
	}

	for _, f := range ch19RespBannedFields {
		if _, present := decoded[f]; present {
			// Presence, not value: sending temperature:1.0 (the default, a
			// semantic no-op) is still a rejected request on this route.
			s.violate("banned field: %s", f)
		}
	}

	// previous_response_id gets its own violation rather than joining the list
	// above, because the reason differs and so does the fix. The others are
	// simply unsupported; this one is the server-side conversation-threading
	// mechanism, and it cannot work when store:false forbids the server from
	// retaining the previous response in the first place. The student's fix is
	// to carry context forward themselves -- via encrypted reasoning items --
	// not merely to delete a field.
	if _, present := decoded["previous_response_id"]; present {
		s.violate("previous_response_id: must be omitted over HTTP on this route (store:false leaves nothing to reference)")
	}

	if _, present := decoded["instructions"]; present {
		req.HasInstructions = true
	}

	if raw, present := decoded["include"]; present {
		if arr, ok := raw.([]any); ok {
			for _, v := range arr {
				if str, ok := v.(string); ok && str == "reasoning.encrypted_content" {
					// With store:false the vendor keeps nothing, so this is the
					// only way reasoning survives to the next turn. Not a
					// violation to omit -- plenty of agents do not need it --
					// but the parent grades a multi-turn agent on having it.
					req.IncludesEncryptedReasoning = true
				}
			}
		}
	}

	if raw, present := decoded["reasoning"]; present {
		if m, ok := raw.(map[string]any); ok {
			if sum, ok := m["summary"].(string); ok {
				req.ReasoningSummary = sum
			}
		}
	}

	if raw, present := decoded["prompt_cache_options"]; present {
		m, ok := raw.(map[string]any)
		if !ok {
			s.violate("prompt_cache_options: must be an object, got %T", raw)
		} else {
			mode, _ := m["mode"].(string)
			req.PromptCacheMode = mode
			if mode != "implicit" && mode != "explicit" {
				s.violate("prompt_cache_options.mode: must be \"implicit\" or \"explicit\", got %q", mode)
			}
		}
	}

	// The breakpoint walk is recursive and whole-tree rather than a scan of
	// input[].content[], because a student can land a breakpoint in a tool
	// definition or a nested content block and still spend their budget. The
	// vendor counts them wherever they are, so the grader must too.
	req.BreakpointCount = ch19RespCountBreakpoints(decoded)
	if req.BreakpointCount > 4 {
		s.violate("prompt_cache_breakpoint: %d present, documented per-request write limit is 4", req.BreakpointCount)
	}

	// The scheme is split off rather than prefix-matched against "Bearer ".
	// HTTP strips trailing whitespace from header values, so a student whose
	// token variable is empty sends "Bearer " and the server receives exactly
	// "Bearer" -- a prefix test would then report a malformed scheme, sending
	// them to fix the one part of the header that was already right.
	if auth == "" {
		s.violate("authorization: header absent")
	} else if scheme, token, _ := strings.Cut(auth, " "); scheme != "Bearer" {
		s.violate("authorization: header must start with \"Bearer \"")
	} else if strings.TrimSpace(token) == "" {
		s.violate("authorization: bearer token is empty")
	}

	// Secret probe. The body is searched, the Authorization header is not:
	// a token in the header is the credential being used correctly, and a
	// fixture that flagged it would train students to work around the grader
	// instead of fixing real leaks. A token in the body is a token that has
	// escaped into prompt content, which is the leak this chapter is about.
	for _, secret := range s.opts.Secrets {
		if secret == "" {
			continue
		}
		if strings.Contains(string(body), secret) {
			s.obs.SecretInBody = true
			if !ch19RespContains(s.obs.LeakedSecrets, secret) {
				s.obs.LeakedSecrets = append(s.obs.LeakedSecrets, secret)
			}
			// The violation text is redacted even though LeakedSecrets holds
			// the value: violations get printed in grading output and pasted
			// into bug reports, and a reference fixture that spills credentials
			// into a log while teaching students not to leak credentials would
			// be hard to defend.
			s.violate("secret leaked in request body: %s", ch19RespRedact(secret))
		}
	}

	s.obs.Requests = append(s.obs.Requests, req)
}

// ch19RespScanInput records roles and per-message breakpoint placement, and
// rejects system-role messages.
func (s *ch19FakeResponses) ch19RespScanInput(req *ch19RespReq, arr []any) {
	for i, el := range arr {
		m, ok := el.(map[string]any)
		if !ok {
			continue
		}
		role, _ := m["role"].(string)
		if role == "" {
			// Not every input element is a message: function_call_output and
			// reasoning items legitimately have no role, so this is silence
			// rather than a violation.
			continue
		}
		req.Roles = append(req.Roles, role)

		if role == "system" {
			// The Responses API wants developer-role messages or the top-level
			// instructions field. "system" is the Chat Completions habit, and
			// it survives a careless migration because it looks harmless.
			s.violate("input[%d]: role %q is banned; use developer role or top-level instructions", i, role)
		}

		bp := ch19RespMsgBP{Role: role}
		if blocks, ok := m["content"].([]any); ok {
			bp.BlockCount = len(blocks)
			for _, b := range blocks {
				if bm, ok := b.(map[string]any); ok {
					if _, present := bm["prompt_cache_breakpoint"]; present {
						bp.HasBreakpoint = true
					}
				}
			}
		}
		// BlockCount stays 0 when content is a bare string. That is not graded
		// here -- the parent decides whether it matters -- but it is the only
		// signal distinguishing "no blocks" from "blocks without a breakpoint",
		// and a string can never carry a breakpoint at all.
		req.MessageBreakpoints = append(req.MessageBreakpoints, bp)
	}
}

// ch19RespCountBreakpoints walks the whole decoded tree counting
// prompt_cache_breakpoint keys.
func ch19RespCountBreakpoints(v any) int {
	switch t := v.(type) {
	case map[string]any:
		n := 0
		for k, val := range t {
			if k == "prompt_cache_breakpoint" {
				n++
			}
			// Recurse into the value regardless: a breakpoint nested under
			// another breakpoint would be pathological, but skipping the
			// subtree would make the count silently wrong rather than loudly so.
			n += ch19RespCountBreakpoints(val)
		}
		return n
	case []any:
		n := 0
		for _, val := range t {
			n += ch19RespCountBreakpoints(val)
		}
		return n
	default:
		return 0
	}
}

func ch19RespContains(hay []string, needle string) bool {
	for _, h := range hay {
		if h == needle {
			return true
		}
	}
	return false
}

// ch19RespRedact renders a secret safely for log output while keeping it
// identifiable to whoever configured it.
func ch19RespRedact(s string) string {
	if len(s) <= 4 {
		return fmt.Sprintf("<redacted len=%d>", len(s))
	}
	return fmt.Sprintf("%s…<redacted len=%d>", s[:4], len(s))
}

// ---------------------------------------------------------------------------
// SSE emission
//
// Everything in this section is modelled on two live captures taken against
// gpt-6.1-sol (see cr/docs/ch19-research-notes.md and the capture files), not
// on reconstruction from documentation. Where the captures disagreed with the
// written spec the captures won -- notably the usage block, which carries more
// detail than the spec described, and sequence_number, which starts at 0.
// ---------------------------------------------------------------------------

// ch19RespStreamCfg is the per-response context handed to each scenario.
//
// This is a struct rather than a parameter list because the number of knobs
// grew past the point where positional ints are safe: in,out,cached,write,
// reasoning are five adjacent integers, and transposing two of them at a call
// site would produce a fixture that reports plausible-but-wrong token counts.
// Named fields make that class of mistake unrepresentable.
type ch19RespStreamCfg struct {
	w  io.Writer
	f  http.Flusher
	id string

	// cacheMode is echoed back in response.completed when the request carried
	// prompt_cache_options; empty means the request omitted it and the echo is
	// suppressed, which is what the live capture does (the field comes back
	// null rather than defaulted).
	cacheMode string

	// includeEncrypted mirrors the request's include[] choice. With store:false
	// the encrypted reasoning blob is the only way reasoning survives to the
	// next turn, so whether it appears is the agent's decision, not ours.
	includeEncrypted bool

	in, out, cached, cacheWrite, reasoningTok int

	// items accumulates every completed output item so response.completed can
	// carry the real `output` array. Built as we stream rather than written out
	// separately, so the terminal snapshot cannot drift from the deltas that
	// produced it.
	items []any
}

// ch19RespEmit writes one SSE frame as BOTH an event: line and a data: line.
//
// The live captures confirm the real API sends both, and students split
// roughly evenly between dispatching on the event: line and on the "type"
// field inside the JSON. Sending only one would make the grader reject a
// correct agent for choosing the other parsing strategy. The invariant that
// keeps the two honest: data["type"] is always set here, from the same string
// used for the event: line, so they can never disagree.
func (s *ch19FakeResponses) ch19RespEmit(w io.Writer, f http.Flusher, name string, data map[string]any) {
	if data == nil {
		data = map[string]any{}
	}
	data["type"] = name

	// sequence_number is zero-based: the live capture's response.created
	// carries sequence_number 0. An agent that treats the first event as #1
	// will be off by one against the real API for every subsequent frame.
	s.mu.Lock()
	data["sequence_number"] = s.seq
	s.seq++
	s.mu.Unlock()

	payload, err := json.Marshal(data)
	if err != nil {
		// Only reachable if a scenario builds an unmarshalable value, which is
		// a bug in this file rather than in anything a student wrote.
		panic(fmt.Sprintf("ch19FakeResponses: cannot marshal %s: %v", name, err))
	}
	fmt.Fprintf(w, "event: %s\n", name)
	fmt.Fprintf(w, "data: %s\n\n", payload)
	f.Flush() // after EVERY event; batching would defeat the delivery tests
}

func (c *ch19RespStreamCfg) emit(s *ch19FakeResponses, name string, data map[string]any) {
	s.ch19RespEmit(c.w, c.f, name, data)
}

func (s *ch19FakeResponses) ch19RespNewID() string {
	s.mu.Lock()
	s.respID++
	id := fmt.Sprintf("resp_ch19_%d", s.respID)
	s.mu.Unlock()
	return id
}

// ch19RespReasoningItem streams one reasoning item whose summary is delivered
// in the given PARTS, each part being a slice of deltas.
//
// Parts, not a flat delta list, because the live capture at summary:"detailed"
// / effort:"high" produced three separate summary parts across hundreds of
// deltas. summary_index increments per part and is what makes a part boundary
// observable; a renderer that ignores it concatenates three distinct thoughts
// into one run-on paragraph. Passing nil parts emits a reasoning item with NO
// summary events at all, which is a real and correct vendor behaviour -- see
// ch19RespStreamReasoningNoSummary.
func (s *ch19FakeResponses) ch19RespReasoningItem(c *ch19RespStreamCfg, itemID string, outputIndex int, parts [][]string) {
	added := map[string]any{
		"id":      itemID,
		"type":    "reasoning",
		"summary": []any{},
	}
	if c.includeEncrypted {
		added["encrypted_content"] = "ch19-encrypted-reasoning-blob"
	}
	c.emit(s, "response.output_item.added", map[string]any{
		"output_index": outputIndex,
		"item":         added,
	})

	var summaries []any
	for si, phrases := range parts {
		c.emit(s, "response.reasoning_summary_part.added", map[string]any{
			"item_id":       itemID,
			"output_index":  outputIndex,
			"summary_index": si,
			"part":          map[string]any{"type": "summary_text", "text": ""},
		})
		for _, p := range phrases {
			c.emit(s, "response.reasoning_summary_text.delta", map[string]any{
				"item_id":       itemID,
				"output_index":  outputIndex,
				"summary_index": si,
				"delta":         p,
			})
		}
		full := strings.Join(phrases, "")
		c.emit(s, "response.reasoning_summary_text.done", map[string]any{
			"item_id":       itemID,
			"output_index":  outputIndex,
			"summary_index": si,
			"text":          full,
		})
		c.emit(s, "response.reasoning_summary_part.done", map[string]any{
			"item_id":       itemID,
			"output_index":  outputIndex,
			"summary_index": si,
			"part":          map[string]any{"type": "summary_text", "text": full},
		})
		summaries = append(summaries, map[string]any{"type": "summary_text", "text": full})
	}

	done := map[string]any{
		"id":      itemID,
		"type":    "reasoning",
		"summary": summaries,
	}
	if summaries == nil {
		done["summary"] = []any{}
	}
	if c.includeEncrypted {
		done["encrypted_content"] = "ch19-encrypted-reasoning-blob"
	}
	c.items = append(c.items, done)
	c.emit(s, "response.output_item.done", map[string]any{
		"output_index": outputIndex,
		"item":         done,
	})
}

// ch19RespMessageItem streams one complete assistant message item, splitting
// the text into the supplied chunks. The annotations/logprobs empty arrays are
// present because the live capture carries them, and an agent doing strict
// schema validation against the real API must not break against this fixture.
func (s *ch19FakeResponses) ch19RespMessageItem(c *ch19RespStreamCfg, itemID string, outputIndex int, chunks []string) {
	c.emit(s, "response.output_item.added", map[string]any{
		"output_index": outputIndex,
		"item": map[string]any{
			"id":      itemID,
			"type":    "message",
			"role":    "assistant",
			"status":  "in_progress",
			"content": []any{},
		},
	})
	c.emit(s, "response.content_part.added", map[string]any{
		"item_id":       itemID,
		"output_index":  outputIndex,
		"content_index": 0,
		"part": map[string]any{
			"type": "output_text", "text": "",
			"annotations": []any{}, "logprobs": []any{},
		},
	})
	for _, ch := range chunks {
		c.emit(s, "response.output_text.delta", map[string]any{
			"item_id":       itemID,
			"output_index":  outputIndex,
			"content_index": 0,
			"delta":         ch,
			"logprobs":      []any{},
		})
	}
	full := strings.Join(chunks, "")
	c.emit(s, "response.output_text.done", map[string]any{
		"item_id":       itemID,
		"output_index":  outputIndex,
		"content_index": 0,
		"text":          full,
		"logprobs":      []any{},
	})
	c.emit(s, "response.content_part.done", map[string]any{
		"item_id":       itemID,
		"output_index":  outputIndex,
		"content_index": 0,
		"part": map[string]any{
			"type": "output_text", "text": full,
			"annotations": []any{}, "logprobs": []any{},
		},
	})
	done := map[string]any{
		"id":     itemID,
		"type":   "message",
		"role":   "assistant",
		"status": "completed",
		"content": []any{
			map[string]any{
				"type": "output_text", "text": full,
				"annotations": []any{}, "logprobs": []any{},
			},
		},
	}
	c.items = append(c.items, done)
	c.emit(s, "response.output_item.done", map[string]any{
		"output_index": outputIndex,
		"item":         done,
	})
}

// ch19RespFunctionCallItem streams one complete function_call item. The
// argument chunks are deliberately split mid-token so a student who treats
// each delta as standalone JSON fails; arguments are only valid once
// concatenated.
func (s *ch19FakeResponses) ch19RespFunctionCallItem(c *ch19RespStreamCfg, itemID, callID, name string, outputIndex int, argChunks []string) {
	c.emit(s, "response.output_item.added", map[string]any{
		"output_index": outputIndex,
		"item": map[string]any{
			"id":        itemID,
			"type":      "function_call",
			"status":    "in_progress",
			"arguments": "",
			"call_id":   callID,
			"name":      name,
		},
	})
	for _, ch := range argChunks {
		c.emit(s, "response.function_call_arguments.delta", map[string]any{
			"item_id":      itemID,
			"output_index": outputIndex,
			"delta":        ch,
		})
	}
	full := strings.Join(argChunks, "")
	c.emit(s, "response.function_call_arguments.done", map[string]any{
		"item_id":      itemID,
		"output_index": outputIndex,
		"arguments":    full,
	})
	done := map[string]any{
		"id":        itemID,
		"type":      "function_call",
		"status":    "completed",
		"arguments": full,
		"call_id":   callID,
		"name":      name,
	}
	c.items = append(c.items, done)
	c.emit(s, "response.output_item.done", map[string]any{
		"output_index": outputIndex,
		"item":         done,
	})
}

// ch19RespOpen emits the two events every successful stream starts with.
func (s *ch19FakeResponses) ch19RespOpen(c *ch19RespStreamCfg) {
	for _, name := range []string{"response.created", "response.in_progress"} {
		c.emit(s, name, map[string]any{
			"response": map[string]any{
				"id": c.id, "object": "response", "status": "in_progress",
			},
		})
	}
}

// ch19RespCompleted emits the terminal success event.
//
// HTTP 200 is not the success signal on this API -- a stream can open, deliver
// content, and then fail. response.completed is the signal, and an agent that
// treats a clean connection close as success will report fabricated answers.
//
// The usage block is reproduced at the exact depth the live capture uses.
// cached_tokens and cache_write_tokens both live under input_tokens_details,
// and reasoning_tokens under output_tokens_details. Depth is the whole point:
// a student reading any of these from the top level of usage gets a missing
// key that silently becomes zero, then concludes caching or reasoning is not
// working when in fact their accessor is wrong.
func (s *ch19FakeResponses) ch19RespCompleted(c *ch19RespStreamCfg) {
	resp := map[string]any{
		"id":     c.id,
		"object": "response",
		"status": "completed",
		"output": c.items,
		"usage": map[string]any{
			"input_tokens": c.in,
			"input_tokens_details": map[string]any{
				"cached_tokens":      c.cached,
				"cache_write_tokens": c.cacheWrite,
			},
			"output_tokens": c.out,
			"output_tokens_details": map[string]any{
				"reasoning_tokens": c.reasoningTok,
			},
			"total_tokens": c.in + c.out,
		},
	}
	// Echoed back normalised, with the ttl the vendor fills in. Only when the
	// request actually carried the field: the live capture returns null rather
	// than a defaulted object when it was omitted, and inventing a default here
	// would let a student who never sent prompt_cache_options appear to have.
	if c.cacheMode != "" {
		resp["prompt_cache_options"] = map[string]any{
			"mode": c.cacheMode,
			"ttl":  "30m",
		}
	}
	if c.items == nil {
		resp["output"] = []any{}
	}
	c.emit(s, "response.completed", map[string]any{"response": resp})
}

// Scenario (a): plain text, no reasoning item.
func (s *ch19FakeResponses) ch19RespStreamText(c *ch19RespStreamCfg) {
	s.ch19RespOpen(c)
	s.ch19RespMessageItem(c, "msg_ch19_a", 0, ch19RespTextChunks)
	s.ch19RespCompleted(c)
}

// Scenario (b): one reasoning item with a single summary part, then the message.
func (s *ch19FakeResponses) ch19RespStreamSummaryThenText(c *ch19RespStreamCfg) {
	s.ch19RespOpen(c)
	s.ch19RespReasoningItem(c, "rs_ch19_b1", 0, [][]string{ch19RespSummaryPhrasesA})
	s.ch19RespMessageItem(c, "msg_ch19_b", 1, ch19RespTextChunks)
	s.ch19RespCompleted(c)
}

// Scenario (c): reasoning, tool call, reasoning AGAIN, then the message.
//
// The second reasoning item is the entire reason this scenario exists. A
// student who wires summaries up naively will usually get the first batch
// right and then lose the second -- either by tearing the summary renderer
// down when the tool call starts, or by appending the post-tool summary into
// assistant text because by then they believe the "thinking" phase is over.
// Both bugs are invisible in scenario (b) and obvious here.
func (s *ch19FakeResponses) ch19RespStreamToolcall(c *ch19RespStreamCfg) {
	s.ch19RespOpen(c)
	s.ch19RespReasoningItem(c, "rs_ch19_c1", 0, [][]string{ch19RespSummaryPhrasesA})
	s.ch19RespFunctionCallItem(c, "fc_ch19_c", "call_ch19_c1", "read_file", 1, ch19RespArgChunks)
	s.ch19RespReasoningItem(c, "rs_ch19_c2", 2, [][]string{ch19RespSummaryPhrasesB})
	s.ch19RespMessageItem(c, "msg_ch19_c", 3, ch19RespTextChunks)
	s.ch19RespCompleted(c)
}

// Scenario (e): TWO function_call items in one response, back to back.
//
// Taken from a live capture, which showed exactly this shape: each call fully
// opened and closed at its own output_index before the next begins, and no
// message item at all. A student who stores "the" pending tool call in a
// single field instead of a map keyed by call_id loses the first call here,
// and a student who stops reading at the first output_item.done never sees the
// second. Both are silent failures against a single-call fixture.
func (s *ch19FakeResponses) ch19RespStreamParallelToolcalls(c *ch19RespStreamCfg) {
	s.ch19RespOpen(c)
	s.ch19RespFunctionCallItem(c, "fc_ch19_p1", "call_ch19_p1", "read_file", 0, ch19RespArgChunks)
	s.ch19RespFunctionCallItem(c, "fc_ch19_p2", "call_ch19_p2", "list_directory", 1, ch19RespArgChunks2)
	s.ch19RespCompleted(c)
}

// Scenario (f): a reasoning item that emits NO summary events.
//
// This is correct vendor behaviour, not a malfunction, and the live capture
// proves it: a reasoning item arrived with summary:[] and a non-zero
// reasoning_tokens count while not a single reasoning_summary_* event was
// sent. An agent that blocks waiting for a summary part before rendering, or
// that asserts summaries whenever reasoning_tokens > 0, hangs or errors here.
// ReasoningTokens is deliberately defaulted non-zero for this scenario so the
// contradiction is reproduced rather than smoothed over.
func (s *ch19FakeResponses) ch19RespStreamReasoningNoSummary(c *ch19RespStreamCfg) {
	if c.reasoningTok == 0 {
		c.reasoningTok = 18 // the value the live capture reported
	}
	s.ch19RespOpen(c)
	s.ch19RespReasoningItem(c, "rs_ch19_f1", 0, nil)
	s.ch19RespMessageItem(c, "msg_ch19_f", 1, ch19RespTextChunks)
	s.ch19RespCompleted(c)
}

// Scenario (g): one reasoning item whose summary arrives in THREE parts.
//
// Mirrors the measured summary:"detailed" / effort:"high" shape. The part
// boundary is only visible through summary_index, so this is the scenario that
// catches a renderer which concatenates every delta into one paragraph.
func (s *ch19FakeResponses) ch19RespStreamMultipartSummary(c *ch19RespStreamCfg) {
	s.ch19RespOpen(c)
	s.ch19RespReasoningItem(c, "rs_ch19_g1", 0, [][]string{
		ch19RespSummaryPhrasesA,
		ch19RespSummaryPhrasesB,
		ch19RespSummaryPhrasesC,
	})
	s.ch19RespMessageItem(c, "msg_ch19_g", 1, ch19RespTextChunks)
	s.ch19RespCompleted(c)
}

// Scenario (h): a usage limit that strikes MID-STREAM.
//
// The documented behaviour that breaks naive error handling: the request is
// admitted, HTTP 200 is already on the wire, text has already been delivered
// to the user's screen, and only then does a terminal response.failed arrive
// carrying the plan error code. A student who decides success from the status
// line has already committed to a partial answer; one who treats a closed
// stream as completion silently truncates. The only correct reading is that
// response.completed is the success signal and its absence is a failure, which
// is precisely what this scenario forces.
//
// Note it emits real output_text deltas first. A fail-fast variant would let a
// student pass by never rendering anything before the terminal event.
func (s *ch19FakeResponses) ch19RespStreamFailMidstream(c *ch19RespStreamCfg, code string) {
	if code == "" {
		code = "subscription_sharing_usage_limit_exceeded"
	}
	s.ch19RespOpen(c)
	itemID := "msg_ch19_h"
	c.emit(s, "response.output_item.added", map[string]any{
		"output_index": 0,
		"item": map[string]any{
			"id": itemID, "type": "message", "role": "assistant",
			"status": "in_progress", "content": []any{},
		},
	})
	c.emit(s, "response.content_part.added", map[string]any{
		"item_id": itemID, "output_index": 0, "content_index": 0,
		"part": map[string]any{
			"type": "output_text", "text": "",
			"annotations": []any{}, "logprobs": []any{},
		},
	})
	for _, ch := range ch19RespTextChunks[:3] {
		c.emit(s, "response.output_text.delta", map[string]any{
			"item_id": itemID, "output_index": 0, "content_index": 0,
			"delta": ch, "logprobs": []any{},
		})
	}
	// No output_text.done, no content_part.done, no output_item.done and no
	// response.completed: the stream is cut off exactly where a real usage
	// limit would cut it. Anything tidier would be a fixture being polite
	// about a failure the vendor is not polite about.
	c.emit(s, "response.failed", map[string]any{
		"response": map[string]any{
			"id":     c.id,
			"object": "response",
			"status": "failed",
			"error": map[string]any{
				"code":    code,
				"message": "ch19 fixture: scripted mid-stream plan failure (" + code + ")",
			},
		},
	})
}

// Scenario (d), in-stream form: the connection opens cleanly and then fails.

// This is the case that catches an agent treating response.created, or a
// non-empty body, as proof of success.
func (s *ch19FakeResponses) ch19RespStreamError(c *ch19RespStreamCfg) {
	c.emit(s, "response.created", map[string]any{
		"response": map[string]any{"id": c.id, "object": "response", "status": "in_progress"},
	})
	c.emit(s, "response.error", map[string]any{
		"code":    "server_error",
		"message": "ch19 fixture: scripted mid-stream failure",
		"param":   nil,
	})
	c.emit(s, "response.failed", map[string]any{
		"response": map[string]any{
			"id":     c.id,
			"object": "response",
			"status": "failed",
			"error": map[string]any{
				"code":    "server_error",
				"message": "ch19 fixture: scripted mid-stream failure",
			},
		},
	})
}

// ch19RespWriteFailure returns a plan-level rejection as an HTTP error with
// the standard error object, before any streaming has begun.
func (s *ch19FakeResponses) ch19RespWriteFailure(w http.ResponseWriter, code string) {
	status, ok := ch19RespFailStatus[code]
	if !ok {
		// An unrecognised code is still a rejection; defaulting to 400 keeps
		// the fixture usable for codes the chapter adds later without making
		// them silently retryable.
		status = http.StatusBadRequest
	}
	errObj := map[string]any{
		"code":    code,
		"message": "ch19 fixture: scripted plan-level rejection (" + code + ")",
		"type":    "invalid_request_error",
	}
	// Only the unsupported-capability rejection names the offending field.
	// That param is the most actionable thing in the whole error, because it
	// tells the student exactly which request field to delete.
	if code == "subscription_sharing_unsupported_capability" {
		errObj["param"] = "temperature"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": errObj})
}

// ch19RespWriteAdmissionFailure returns a direct-admission failure.
//
// These are rejected before the request ever reaches the model, so they do NOT
// carry the standard error object -- the body is {"detail": "..."} and there
// is no error.code to read. This is the single nastiest shape in the chapter:
// a student whose parser does resp.Error.Code unconditionally panics or
// nil-derefs on precisely the failure it most needs to surface, and because
// the happy path and every scripted error path both have error.code, nothing
// else in the fixture would ever catch it.
func (s *ch19FakeResponses) ch19RespWriteAdmissionFailure(w http.ResponseWriter, status int) {
	if status == 0 {
		status = http.StatusServiceUnavailable
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"detail": "ch19 fixture: scripted direct-admission failure; no error object by design",
	})
}

// Scripted content. Package-level so the tests can assert against the exact
// same values the server sends, instead of duplicating string literals that
// would drift apart.
var (
	// Four chunks, split mid-word, so a student who renders per-delta without
	// buffering produces visibly wrong output rather than accidentally correct
	// output.
	ch19RespTextChunks = []string{"Hel", "lo from ", "the Responses ", "API."}

	ch19RespSummaryPhrasesA = []string{"Reading ", "the request, ", "then deciding ", "what to do."}
	ch19RespSummaryPhrasesB = []string{"Got the file. ", "Now checking ", "what it says."}
	ch19RespSummaryPhrasesC = []string{"One more ", "look before ", "answering."}

	// Each concatenates to valid JSON -- and no single chunk is valid JSON on
	// its own.
	ch19RespArgChunks  = []string{`{"path":`, `"/etc/ho`, `sts"}`}
	ch19RespArgChunks2 = []string{`{"dir`, `":"/tm`, `p"}`}
)

// What each delta sequence concatenates to. Derived rather than written out,
// so a test asserting the joined result cannot disagree with what the server
// actually sent.
var (
	ch19RespTextFull       = strings.Join(ch19RespTextChunks, "")
	ch19RespSummaryAFull   = strings.Join(ch19RespSummaryPhrasesA, "")
	ch19RespSummaryBFull   = strings.Join(ch19RespSummaryPhrasesB, "")
	ch19RespSummaryCFull   = strings.Join(ch19RespSummaryPhrasesC, "")
	ch19RespArgumentsFull  = strings.Join(ch19RespArgChunks, "")
	ch19RespArguments2Full = strings.Join(ch19RespArgChunks2, "")
)

// ---------------------------------------------------------------------------
// GET /v1/models
// ---------------------------------------------------------------------------

// ch19RespHandleModels serves the SIWC model catalogue.
//
// The array is keyed "models", NOT "data". That differs from the familiar
// platform /v1/models shape, and it is the kind of difference that makes a
// student's existing client decode successfully into an empty list and report
// "no models available" rather than failing loudly.
//
// Each entry carries slug, display_name and visibility because the three have
// distinct jobs: slug is what goes in a request, display_name is what a human
// sees, and visibility is the filter. The hidden entry is a trap with a
// purpose -- the docs say to select on visibility == "list", so a student who
// renders the raw catalogue offers their user a model that fails at request
// time. Returning only listed models would make that bug ungradeable.
func (s *ch19FakeResponses) ch19RespHandleModels(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.obs.ModelsHits++
	s.obs.ModelsAuth = append(s.obs.ModelsAuth, r.Header.Get("Authorization"))
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"models": []any{
			map[string]any{
				"id": "gpt-5.6", "object": "model",
				"slug": "gpt-5.6", "display_name": "GPT-5.6", "visibility": "list",
			},
			map[string]any{
				"id": "gpt-5.6-codex", "object": "model",
				"slug": "gpt-5.6-codex", "display_name": "GPT-5.6 Codex", "visibility": "list",
			},
			map[string]any{
				"id": "gpt-5.6-internal-eval", "object": "model",
				"slug": "gpt-5.6-internal-eval", "display_name": "GPT-5.6 Internal Eval",
				"visibility": "hidden",
			},
		},
	})
}
