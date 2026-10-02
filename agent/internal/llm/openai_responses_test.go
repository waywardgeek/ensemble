package llm

// Tests for the Responses surface.
//
// The stream fixture in testdata/responses_stream.sse is a VERBATIM capture
// from the live endpoint, not a hand-written approximation. That matters:
// a fixture invented from the same understanding as the parser will agree
// with the parser even when both are wrong about the format. This one carries
// reasoning summaries, assistant text, a function call and an encrypted
// reasoning blob in a single response, which is the combination that exercises
// every branch at once.

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

type respRecorder struct {
	next   uint64
	deltas map[common.DeltaKind]int
	text   strings.Builder
	sum    strings.Builder
	args   strings.Builder
	finals []common.Part
	events []common.Event
}

func newRespRecorder() *respRecorder {
	return &respRecorder{deltas: map[common.DeltaKind]int{}}
}

func (r *respRecorder) callbacks() common.StreamCallbacks {
	return common.StreamCallbacks{
		AllocPartID: func() uint64 { r.next++; return r.next },
		OnDelta: func(_ uint64, kind common.DeltaKind, chunk string) {
			r.deltas[kind]++
			switch kind {
			case common.DeltaText:
				r.text.WriteString(chunk)
			case common.DeltaReasoningSummary:
				r.sum.WriteString(chunk)
			case common.DeltaToolCall:
				r.args.WriteString(chunk)
			}
		},
		OnPartFinal: func(_ uint64, p common.Part) { r.finals = append(r.finals, p) },
		OnFrame:     func(string, []byte) {},
		OnEvent:     func(e common.Event) { r.events = append(r.events, e) },
	}
}

func sseResponse(t *testing.T, path string) *http.Response {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	return &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(string(b))),
	}
}

// The headline capability of the whole chapter: reasoning summaries arrive as
// incremental deltas rather than as a token count after the fact.
func TestCh19ResponsesStreamsReasoningSummaries(t *testing.T) {
	rec := newRespRecorder()
	if err := (responsesSeam{}).Parse(sseResponse(t, "testdata/responses_stream.sse"), rec.callbacks()); err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if got := rec.deltas[common.DeltaReasoningSummary]; got == 0 {
		t.Fatal("no reasoning summary deltas: the accessibility payload of this " +
			"chapter never reached the consumer")
	}
	if rec.sum.Len() == 0 {
		t.Fatal("summary deltas carried no text")
	}

	// Summaries must NOT be reported as ordinary assistant text. A consumer
	// that cannot tell them apart reads the model's thinking aloud as if it
	// were the answer.
	if strings.Contains(rec.text.String(), rec.sum.String()) && rec.sum.Len() > 0 {
		t.Error("summary text leaked into the DeltaText stream")
	}
	if rec.deltas[common.DeltaText] == 0 {
		t.Error("expected assistant text deltas as well as summaries")
	}
}

// Reasoning must come back as replayable material, or the next turn starts
// cold and the second reason for the migration is lost.
func TestCh19ResponsesCapturesEncryptedReasoning(t *testing.T) {
	rec := newRespRecorder()
	if err := (responsesSeam{}).Parse(sseResponse(t, "testdata/responses_stream.sse"), rec.callbacks()); err != nil {
		t.Fatalf("Parse: %v", err)
	}

	var ended *common.ResponseData
	for _, e := range rec.events {
		if e.Type == common.ResponseEnded {
			ended = e.Response
		}
	}
	if ended == nil {
		t.Fatal("no ResponseEnded event")
	}

	var opaque int
	for _, p := range ended.Parts {
		op, ok := p.(common.OpaquePart)
		if !ok {
			continue
		}
		opaque++
		// The bytes must be the vendor's, verbatim: an AEAD blob that has
		// been through a decode/encode round trip is a rejected request.
		if !strings.Contains(string(op.Data), "encrypted_content") {
			t.Error("opaque part does not carry encrypted_content")
		}
		if op.From.Surface != common.SurfaceResponses {
			t.Errorf("replay material tagged with surface %v, want SurfaceResponses", op.From.Surface)
		}
		if op.From.Vendor != common.VendorOpenAI {
			t.Errorf("replay material tagged with vendor %v", op.From.Vendor)
		}
		if got := respSummaryText(op); got == "" {
			t.Error("reasoning item has no readable summary text")
		}
	}
	if opaque != 1 {
		t.Errorf("got %d opaque reasoning parts, want 1", opaque)
	}
}

func TestCh19ResponsesParsesTextAndToolCall(t *testing.T) {
	rec := newRespRecorder()
	if err := (responsesSeam{}).Parse(sseResponse(t, "testdata/responses_stream.sse"), rec.callbacks()); err != nil {
		t.Fatalf("Parse: %v", err)
	}

	var ended *common.ResponseData
	for _, e := range rec.events {
		if e.Type == common.ResponseEnded {
			ended = e.Response
		}
	}
	if ended == nil {
		t.Fatal("no ResponseEnded event")
	}

	var texts, calls int
	for _, p := range ended.Parts {
		switch v := p.(type) {
		case common.TextPart:
			texts++
		case common.ToolCallPart:
			calls++
			if v.Name == "" {
				t.Error("tool call has no name")
			}
			if v.CallID == "" {
				t.Error("tool call has no call_id: the result cannot be correlated back")
			}
		}
	}
	if texts == 0 {
		t.Error("no text part recovered")
	}
	if calls != 1 {
		t.Errorf("got %d tool calls, want 1", calls)
	}

	// Usage must land in the four disjoint buckets, and the reasoning tokens
	// that were billed must not be double counted into Output.
	if ended.Usage.Output == 0 {
		t.Error("no output tokens recorded")
	}
	if ended.Usage.Input == 0 {
		t.Error("no input tokens recorded")
	}
}

// A truncated stream must not be recorded as a whole turn.
func TestCh19ResponsesIncompleteStreamIsAnError(t *testing.T) {
	body, err := os.ReadFile("testdata/responses_stream.sse")
	if err != nil {
		t.Fatal(err)
	}
	cut := string(body)
	if i := strings.Index(cut, "event: response.completed"); i > 0 {
		cut = cut[:i]
	} else {
		t.Fatal("fixture has no response.completed to cut at")
	}

	rec := newRespRecorder()
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(cut)),
	}
	if err := (responsesSeam{}).Parse(resp, rec.callbacks()); err == nil {
		t.Fatal("a stream that ended without response.completed was reported as success")
	}
}

// A mid-stream failure arrives with HTTP 200, because the status line was
// written before the vendor knew anything was wrong. This is the shape a
// plan-usage limit takes, and a parser that only checks the status code
// reports a silently empty answer as success.
func TestCh19ResponsesMidStreamFailureIsAnError(t *testing.T) {
	const stream = "event: response.created\n" +
		`data: {"type":"response.created","response":{"id":"resp_1"}}` + "\n\n" +
		"event: response.failed\n" +
		`data: {"type":"response.failed","response":{"error":{"type":"usage_limit_reached",` +
		`"code":"usage_limit_reached","message":"ChatGPT plan usage limit reached"}}}` + "\n\n"

	rec := newRespRecorder()
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(stream)),
	}
	err := (responsesSeam{}).Parse(resp, rec.callbacks())
	if err == nil {
		t.Fatal("response.failed on a 200 stream was reported as success")
	}
	if !strings.Contains(err.Error(), "usage_limit_reached") {
		t.Errorf("error does not name the vendor's code: %v", err)
	}
}

// --- render ---------------------------------------------------------------

func renderResponses(t *testing.T, c *common.Context, cfg common.Config) map[string]any {
	t.Helper()
	req, err := (responsesSeam{}).Render(c, cfg)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	b, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("request body is not JSON: %v", err)
	}
	if !strings.HasSuffix(req.URL.Path, "/v1/responses") {
		t.Errorf("posted to %s, want /v1/responses", req.URL.Path)
	}
	return got
}

// The constitution must be a developer-role content block, not the top-level
// "instructions" string. "instructions" is legal; it just cannot carry a cache
// breakpoint, and losing the boundary on the largest and most stable span of
// the request is the most expensive cache mistake available.
func TestCh19ResponsesSystemPromptIsACacheableBlock(t *testing.T) {
	cfg := common.Config{
		Model:        "gpt-ch19-course",
		BaseURL:      "http://example.invalid",
		APIKey:       "k",
		SystemPrompt: "You are a careful assistant.",
	}
	c := ctxWith([]common.Entry{{Seq: 1, Actor: common.ActorHuman, Parts: common.PartList{common.TextPart{Text: "hi"}}}})

	got := renderResponses(t, c, cfg)

	if _, bad := got["instructions"]; bad {
		t.Error("system prompt sent as top-level instructions: it cannot carry a cache breakpoint there")
	}

	input, _ := got["input"].([]any)
	if len(input) == 0 {
		t.Fatal("empty input array")
	}
	first, _ := input[0].(map[string]any)
	if first["role"] != "developer" {
		t.Errorf("first input item has role %v, want developer", first["role"])
	}
	content, _ := first["content"].([]any)
	if len(content) == 0 {
		t.Fatal("developer message has no content blocks")
	}
	block, _ := content[0].(map[string]any)
	if _, ok := block["prompt_cache_breakpoint"]; !ok {
		t.Error("no cache breakpoint on the system prompt block")
	}
	if got["prompt_cache_options"] == nil {
		t.Error("explicit cache mode not declared even though a breakpoint attached")
	}
}

// Declaring explicit mode with nothing marked turns caching OFF entirely, so
// the option must follow evidence that a marker landed, never intent.
func TestCh19ResponsesNoBreakpointsMeansNoExplicitMode(t *testing.T) {
	cfg := common.Config{
		Model:   "gpt-5.6-sol-course", // implicit caching model
		BaseURL: "http://example.invalid",
		APIKey:  "k",
	}
	c := ctxWith([]common.Entry{{Seq: 1, Actor: common.ActorHuman, Parts: common.PartList{common.TextPart{Text: "hi"}}}})

	got := renderResponses(t, c, cfg)
	if got["prompt_cache_options"] != nil {
		t.Error("explicit mode declared on a model that does not use explicit breakpoints: " +
			"this disables caching rather than enabling it")
	}
}

// store:false means the vendor keeps no copy, so reasoning we do not ask to
// have returned is reasoning we can never replay.
func TestCh19ResponsesIsStatelessAndAsksForReasoning(t *testing.T) {
	cfg := common.Config{
		Model:    "gpt-ch19-course",
		BaseURL:  "http://example.invalid",
		APIKey:   "k",
		Thinking: common.ThinkingMedium,
	}
	c := ctxWith([]common.Entry{{Seq: 1, Actor: common.ActorHuman, Parts: common.PartList{common.TextPart{Text: "hi"}}}})

	got := renderResponses(t, c, cfg)

	if got["store"] != false {
		t.Errorf("store = %v, want false", got["store"])
	}
	if _, bad := got["previous_response_id"]; bad {
		t.Error("previous_response_id sent with store:false: there is no server copy to continue")
	}
	inc, _ := got["include"].([]any)
	var asked bool
	for _, v := range inc {
		if v == "reasoning.encrypted_content" {
			asked = true
		}
	}
	if !asked {
		t.Error("did not ask for reasoning.encrypted_content, so reasoning cannot survive the turn")
	}
}

// The Chat Completions workaround must NOT be imported here. That flag records
// a restriction on the other endpoint, whose own error message names this one
// as the fix; honouring it here would silence thinking on every tool-bearing
// turn, which is every turn an agent takes.
func TestCh19ResponsesKeepsThinkingWithTools(t *testing.T) {
	f, ok := common.LookupModel("gpt-5.6-sol")
	if !ok {
		t.Skip("model row absent")
	}
	if !f.NoThinkingWithTools {
		t.Skip("gpt-5.6-sol no longer carries the restriction; test is moot")
	}

	cfg := common.Config{
		Model:    "gpt-5.6-sol",
		BaseURL:  "http://example.invalid",
		APIKey:   "k",
		Thinking: common.ThinkingMedium,
		Tools: []common.ToolDecl{{
			Name:        "get_weather",
			Description: "weather",
			Schema:      json.RawMessage(`{"type":"object"}`),
		}},
	}
	c := ctxWith([]common.Entry{{Seq: 1, Actor: common.ActorHuman, Parts: common.PartList{common.TextPart{Text: "hi"}}}})

	got := renderResponses(t, c, cfg)
	tools, _ := got["tools"].([]any)
	if len(tools) == 0 {
		t.Fatal("tools were dropped")
	}
	reasoning, _ := got["reasoning"].(map[string]any)
	if reasoning == nil {
		t.Fatal("reasoning dropped on a tool-bearing turn: the Chat Completions " +
			"workaround was imported onto a surface that does not need it")
	}
	if reasoning["effort"] == "none" {
		t.Error("reasoning effort forced to none on the Responses surface")
	}
}

// Tool declarations are flat here. Chat Completions nests the real declaration
// under a "function" object; sending that shape to this surface declares a
// tool with no name.
func TestCh19ResponsesToolsAreFlat(t *testing.T) {
	cfg := common.Config{
		Model:   "gpt-ch19-course",
		BaseURL: "http://example.invalid",
		APIKey:  "k",
		Tools: []common.ToolDecl{{
			Name:        "get_weather",
			Description: "weather",
			Schema:      json.RawMessage(`{"type":"object"}`),
		}},
	}
	c := ctxWith([]common.Entry{{Seq: 1, Actor: common.ActorHuman, Parts: common.PartList{common.TextPart{Text: "hi"}}}})

	got := renderResponses(t, c, cfg)
	tools, _ := got["tools"].([]any)
	if len(tools) != 1 {
		t.Fatalf("got %d tools, want 1", len(tools))
	}
	tool, _ := tools[0].(map[string]any)
	if _, nested := tool["function"]; nested {
		t.Error("tool declared in the nested Chat Completions shape")
	}
	if tool["name"] != "get_weather" {
		t.Errorf("tool name = %v, want get_weather at the top level", tool["name"])
	}
}
