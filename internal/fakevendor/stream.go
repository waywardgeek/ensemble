package fakevendor

// SSE mode.
//
// The same scripted replies, delivered as Server-Sent Events instead of one
// JSON document. Nothing about the script changes; only the framing does.
// That is the property worth testing, because it is the property the seam
// claims: a response is the same response whether it trickles or lands.
//
// Two rules are load-bearing here and both cost real debugging to find.
//
// ONE: every `data:` payload must be a SINGLE LINE. SSE joins consecutive
// data lines with a newline, so a pretty-printed JSON body does not arrive as
// JSON with newlines in it — it arrives as several data fields joined back
// together, and whether that reparses is luck. The non-streaming bodies in
// fake.go contain literal newlines for readability, which is exactly why the
// builders here are separate rather than reused.
//
// TWO: chunk on RUNE boundaries. Splitting a multi-byte character in half
// produces two invalid UTF-8 fragments that concatenate back to the right
// string, so a test with ASCII fixtures passes and the bug surfaces on the
// first non-English response.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// chunkSize is how many runes the fake puts in one delta.
//
// Small on purpose. A fake that sends one chunk per response cannot tell a
// working incremental parser from one that only handles the whole-document
// case, which is the single most likely way for a streaming parser to be
// wrong while appearing to work.
const chunkSize = 3

// wantsStream reports whether a request asked for SSE.
//
// Anthropic and OpenAI ask in the request BODY; Gemini asks in the URL. A
// fake that checks only the body serves Gemini a JSON document forever and
// the student's Gemini stream parser is never once exercised.
func wantsStream(r *http.Request, body []byte, vendor string) bool {
	if vendor == "gemini" {
		return strings.Contains(r.URL.RawQuery, "alt=sse") ||
			strings.Contains(r.URL.Path, "streamGenerateContent")
	}
	var m map[string]any
	if json.Unmarshal(body, &m) != nil {
		return false
	}
	on, _ := m["stream"].(bool)
	return on
}

// runeChunks splits s into pieces of at most n runes.
func runeChunks(s string, n int) []string {
	if s == "" {
		return nil
	}
	r := []rune(s)
	var out []string
	for i := 0; i < len(r); i += n {
		j := i + n
		if j > len(r) {
			j = len(r)
		}
		out = append(out, string(r[i:j]))
	}
	return out
}

type sseWriter struct {
	w http.ResponseWriter
	f http.Flusher
}

func newSSEWriter(w http.ResponseWriter) *sseWriter {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	f, _ := w.(http.Flusher)
	s := &sseWriter{w: w, f: f}
	s.flush()
	return s
}

func (s *sseWriter) flush() {
	if s.f != nil {
		s.f.Flush()
	}
}

// send writes one event. eventType may be empty, which is how OpenAI and
// Gemini frame everything they send.
func (s *sseWriter) send(eventType, data string) {
	if strings.ContainsAny(data, "\n\r") {
		// Refuse rather than emit a frame that will be silently rejoined into
		// something that is not JSON. A fake that corrupts its own output
		// teaches the student to debug the fake.
		panic("fakevendor: SSE data must be a single line: " + data)
	}
	if eventType != "" {
		fmt.Fprintf(s.w, "event: %s\n", eventType)
	}
	fmt.Fprintf(s.w, "data: %s\n\n", data)
	s.flush()
}

// --- anthropic -------------------------------------------------------------

func anthropicStream(w http.ResponseWriter, r Reply) {
	s := newSSEWriter(w)
	idx := 0

	s.send("message_start", fmt.Sprintf(
		`{"type":"message_start","message":{"id":"msg_fake","type":"message","role":"assistant","model":%s,"content":[],"stop_reason":null,"usage":{"input_tokens":%d,"cache_creation_input_tokens":%d,"cache_read_input_tokens":%d,"output_tokens":0}}}`,
		jsonStr(Models["anthropic"]), r.Usage.Input, r.Usage.CacheWrite, r.Usage.CacheRead))

	// Reasoning first, as the real API sends it. The signature arrives in its
	// own delta at the END of the block, which is what forces a student's
	// parser to rebuild the thinking block rather than to pass through a
	// content_block_start it already has.
	if r.Thinking != "" {
		s.send("content_block_start", fmt.Sprintf(
			`{"type":"content_block_start","index":%d,"content_block":{"type":"thinking","thinking":"","signature":""}}`, idx))
		for _, c := range runeChunks(r.Thinking, chunkSize) {
			s.send("content_block_delta", fmt.Sprintf(
				`{"type":"content_block_delta","index":%d,"delta":{"type":"thinking_delta","thinking":%s}}`, idx, jsonStr(c)))
		}
		s.send("content_block_delta", fmt.Sprintf(
			`{"type":"content_block_delta","index":%d,"delta":{"type":"signature_delta","signature":%s}}`, idx, jsonStr("sig-fake-thinking")))
		s.send("content_block_stop", fmt.Sprintf(`{"type":"content_block_stop","index":%d}`, idx))
		idx++
	}

	if r.Text != "" {
		s.send("content_block_start", fmt.Sprintf(
			`{"type":"content_block_start","index":%d,"content_block":{"type":"text","text":""}}`, idx))
		for _, c := range runeChunks(r.Text, chunkSize) {
			s.send("content_block_delta", fmt.Sprintf(
				`{"type":"content_block_delta","index":%d,"delta":{"type":"text_delta","text":%s}}`, idx, jsonStr(c)))
		}
		s.send("content_block_stop", fmt.Sprintf(`{"type":"content_block_stop","index":%d}`, idx))
		idx++
	}

	for _, call := range r.calls() {
		// The NAME arrives whole in the header; only the arguments stream.
		s.send("content_block_start", fmt.Sprintf(
			`{"type":"content_block_start","index":%d,"content_block":{"type":"tool_use","id":%s,"name":%s,"input":{}}}`,
			idx, jsonStr(call.ID), jsonStr(call.Name)))
		for _, c := range runeChunks(call.Args, chunkSize) {
			s.send("content_block_delta", fmt.Sprintf(
				`{"type":"content_block_delta","index":%d,"delta":{"type":"input_json_delta","partial_json":%s}}`, idx, jsonStr(c)))
		}
		s.send("content_block_stop", fmt.Sprintf(`{"type":"content_block_stop","index":%d}`, idx))
		idx++
	}

	stop := "end_turn"
	if len(r.calls()) > 0 {
		stop = "tool_use"
	}
	// Output tokens are only knowable at the end, so message_start reported
	// zero and this corrects it. A parser that reads usage once, at the
	// start, reports every turn as costing nothing.
	s.send("message_delta", fmt.Sprintf(
		`{"type":"message_delta","delta":{"stop_reason":%s},"usage":{"output_tokens":%d}}`, jsonStr(stop), r.Usage.Output))
	s.send("message_stop", `{"type":"message_stop"}`)
}

// --- openai ----------------------------------------------------------------

func openAIStream(w http.ResponseWriter, r Reply) {
	s := newSSEWriter(w)
	model := jsonStr(Models["openai"])

	chunk := func(delta, finish string) string {
		fin := "null"
		if finish != "" {
			fin = jsonStr(finish)
		}
		return fmt.Sprintf(
			`{"id":"chatcmpl_fake","object":"chat.completion.chunk","model":%s,"choices":[{"index":0,"delta":%s,"finish_reason":%s}]}`,
			model, delta, fin)
	}

	s.send("", chunk(`{"role":"assistant","content":""}`, ""))

	for _, c := range runeChunks(r.Text, chunkSize) {
		s.send("", chunk(fmt.Sprintf(`{"content":%s}`, jsonStr(c)), ""))
	}

	for i, call := range r.calls() {
		// First fragment carries id and name; later ones carry ONLY argument
		// text. A parser that overwrites name on every fragment erases it.
		s.send("", chunk(fmt.Sprintf(
			`{"tool_calls":[{"index":%d,"id":%s,"type":"function","function":{"name":%s,"arguments":""}}]}`,
			i, jsonStr(call.ID), jsonStr(call.Name)), ""))
		for _, c := range runeChunks(call.Args, chunkSize) {
			s.send("", chunk(fmt.Sprintf(
				`{"tool_calls":[{"index":%d,"function":{"arguments":%s}}]}`, i, jsonStr(c)), ""))
		}
	}

	finish := "stop"
	if len(r.calls()) > 0 {
		finish = "tool_calls"
	}
	s.send("", chunk(`{}`, finish))

	// The usage chunk has an EMPTY choices array, and arrives only because
	// the request set stream_options.include_usage. A parser that indexes
	// choices[0] without checking the length panics exactly here.
	prompt := r.Usage.Input + r.Usage.CacheWrite + r.Usage.CacheRead
	s.send("", fmt.Sprintf(
		`{"id":"chatcmpl_fake","object":"chat.completion.chunk","model":%s,"choices":[],"usage":{"prompt_tokens":%d,"completion_tokens":%d,"total_tokens":%d,"prompt_tokens_details":{"cached_tokens":%d,"cache_write_tokens":%d}}}`,
		model, prompt, r.Usage.Output, prompt+r.Usage.Output, r.Usage.CacheRead, r.Usage.CacheWrite))

	s.send("", "[DONE]")
}

// --- gemini ----------------------------------------------------------------

func geminiStream(w http.ResponseWriter, r Reply) {
	s := newSSEWriter(w)
	model := jsonStr(Models["gemini"])

	frame := func(parts, finish, usage string) string {
		fin := "null"
		if finish != "" {
			fin = jsonStr(finish)
		}
		out := fmt.Sprintf(
			`{"modelVersion":%s,"candidates":[{"content":{"role":"model","parts":[%s]},"finishReason":%s}]`,
			model, parts, fin)
		if usage != "" {
			out += `,"usageMetadata":` + usage
		}
		return out + "}"
	}

	for _, c := range runeChunks(r.Thinking, chunkSize) {
		s.send("", frame(fmt.Sprintf(`{"text":%s,"thought":true}`, jsonStr(c)), "", ""))
	}
	if r.Thinking != "" {
		s.send("", frame(fmt.Sprintf(`{"text":"","thought":true,"thoughtSignature":%s}`,
			jsonStr("sig-fake-thought")), "", ""))
	}

	for _, c := range runeChunks(r.Text, chunkSize) {
		s.send("", frame(fmt.Sprintf(`{"text":%s}`, jsonStr(c)), "", ""))
	}

	// A functionCall arrives COMPLETE, in one frame, arguments and all.
	// Gemini does not stream function arguments, which is the fact that made
	// ModelFeatures.Stream a bitmask instead of a bool. The fake models the
	// limitation rather than papering over it: a fake that streams arguments
	// Gemini would never stream teaches a parser to expect them.
	for _, call := range r.calls() {
		s.send("", frame(fmt.Sprintf(`{"functionCall":{"id":%s,"name":%s,"args":%s},"thoughtSignature":%s}`,
			jsonStr(call.ID), jsonStr(call.Name), call.Args, jsonStr("sig-fake-"+call.ID)), "", ""))
	}

	thoughts := r.GeminiThoughts
	if thoughts > r.Usage.Output {
		thoughts = 0
	}
	candidates := r.Usage.Output - thoughts
	prompt := r.Usage.Input + r.Usage.CacheRead
	usage := fmt.Sprintf(
		`{"promptTokenCount":%d,"candidatesTokenCount":%d,"cachedContentTokenCount":%d,"thoughtsTokenCount":%d,"totalTokenCount":%d}`,
		prompt, candidates, r.Usage.CacheRead, thoughts, prompt+candidates+thoughts)

	// Final frame: no parts, the stop reason, and the token counts.
	s.send("", frame("", "STOP", usage))
}
