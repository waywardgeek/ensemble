// Package fakevendor serves fake Anthropic, OpenAI and Gemini endpoints from
// one HTTP server, routed by path.
//
// Why it exists: Chapter 2 asks the student to build three renderers and three
// parsers. Nobody should have to pay three subscriptions to finish a chapter,
// and a student with one API key — or none — must be able to score 100.
//
// What it does NOT prove: a fake is a MODEL of a vendor, and a model is wrong
// in exactly the places you did not think to model. Passing against this
// server is not a claim about production. Draft 3 of this chapter asserted a
// vendor capability that our own fakes happily accepted and that turned out to
// be false the moment it was run against the real API.
package fakevendor

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
)

// dumpPath, when set via FAKEVENDOR_DUMP, makes the server append every
// request body it receives to that file, verbatim. Off by default and inert
// when unset: it adds a file write, never a change to what is served.
//
// It exists so that a claim like "this refactor did not change what we send"
// can be settled by diffing bytes instead of by comparing two scores of 100.
// A score is a summary; the request body is the artifact.
var dumpPath = os.Getenv("FAKEVENDOR_DUMP")

func dumpRequest(path string, body []byte) {
	if dumpPath == "" {
		return
	}
	dumpMu.Lock()
	defer dumpMu.Unlock()
	f, err := os.OpenFile(dumpPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "===== REQUEST %s =====\n%s\n", path, body)
}

var dumpMu sync.Mutex

// Canonical is usage as the book defines it: four disjoint categories summing
// to the billable total. Each vendor below re-expresses these numbers in its
// own convention, which is the whole point — the student's parser has to
// convert them back.
type Canonical struct{ Input, CacheWrite, CacheRead, Output int }

// ToolCall is one requested call inside a reply.
//
// Chapter 3 needs a single assistant message to carry MORE THAN ONE call, so
// that matching results to calls by id is forced rather than optional: with
// one call outstanding, keying results by position, by name or by id are the
// same program, and the grader cannot tell them apart.
type ToolCall struct {
	ID   string
	Name string
	Args string // raw JSON object, e.g. `{"path":"config.json"}`
}

// Reply is one scripted model response.
type Reply struct {
	Text     string
	ToolName string
	ToolArgs string // raw JSON object, e.g. `{"path":"config.json"}`
	ToolID   string
	Usage    Canonical

	// Thinking is reasoning content. Empty in every Chapter 2 fixture, so
	// those responses stay byte-identical; set it to exercise the reasoning
	// path, where the vendors differ most. Anthropic streams it as a
	// thinking block whose SIGNATURE arrives last, Gemini as thought parts,
	// and OpenAI not at all.
	Thinking string

	// Tools carries two or more calls. When it is empty the singular
	// ToolName/ToolArgs/ToolID fields are used instead, so every Chapter 2
	// fixture keeps producing byte-identical wire output.
	Tools []ToolCall

	// GeminiThoughts splits Output so that the Gemini response reports
	// thoughtsTokenCount separately from candidatesTokenCount. They are
	// DISJOINT on the wire, so a parser that assumes thoughts are included in
	// candidates undercounts billed output. Verified 2026-09-12.
	GeminiThoughts int

	// Status and ErrBody override a normal reply to exercise error parsing.
	Status  int
	ErrBody string
}

// calls normalizes the two ways a reply can request tools into one list.
func (r Reply) calls() []ToolCall {
	if len(r.Tools) > 0 {
		return r.Tools
	}
	if r.ToolName != "" {
		return []ToolCall{{ID: r.ToolID, Name: r.ToolName, Args: r.ToolArgs}}
	}
	return nil
}

// Recorded is one request as the server received it.
type Recorded struct {
	Vendor string
	Path   string
	Body   []byte
}

// JSON parses the recorded body, returning nil if it was not valid JSON.
// A malformed request is still answered with 200: the grader records
// everything and judges afterwards, so one run surfaces every bug rather than
// one bug per run.
func (r Recorded) JSON() map[string]any {
	var m map[string]any
	if json.Unmarshal(r.Body, &m) != nil {
		return nil
	}
	return m
}

// Models reported by each fake, so that provenance has something real to
// capture. They differ on purpose: seam-parse requires the three contexts to
// be identical APART FROM provenance, and a submission whose contexts are
// fully identical has thrown provenance away.
var Models = map[string]string{
	"anthropic": "claude-sonnet-5-fake",
	"openai":    "gpt-5-fake",
	"gemini":    "gemini-3.5-flash-fake",
}

type Server struct {
	mu       sync.Mutex
	replies  []Reply
	n        int
	requests []Recorded
	opt      Options

	ts *httptest.Server
}

// Options tune the fake for being driven by hand (cmd/fakevendor). The grader
// uses none of them: New() is the grader's constructor and its behaviour is
// unchanged — a script that runs out repeats its last reply.
type Options struct {
	// Cycle restarts the script from the top when it runs out, instead of
	// repeating the last reply. A REPL session sends an unbounded number of
	// requests; without this every turn after the script ends gets the same
	// answer, and if that answer is a tool call the loop never terminates.
	Cycle bool
	// Trace, if set, gets one line per request: which vendor's endpoint it hit,
	// whether it carried tool results, and which reply was served. This is
	// how a person WATCHES the loop rather than only being scored by it.
	Trace io.Writer
	// Addr, if set, is a fixed listen address such as "127.0.0.1:8089".
	// Empty picks a free port, as httptest does.
	Addr string
}

func New(replies []Reply) *Server {
	return NewWithOptions(replies, Options{})
}

func NewWithOptions(replies []Reply, opt Options) *Server {
	s := &Server{replies: replies, opt: opt}
	if opt.Addr == "" {
		s.ts = httptest.NewServer(http.HandlerFunc(s.handle))
		return s
	}
	ln, err := net.Listen("tcp", opt.Addr)
	if err != nil {
		panic("fakevendor: listen " + opt.Addr + ": " + err.Error())
	}
	s.ts = httptest.NewUnstartedServer(http.HandlerFunc(s.handle))
	s.ts.Listener = ln
	s.ts.Start()
	return s
}

func (s *Server) URL() string { return s.ts.URL }
func (s *Server) Close()      { s.ts.Close() }

func (s *Server) Requests() []Recorded {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Recorded{}, s.requests...)
}

// RequestsFor returns only the requests that arrived at one vendor's endpoint.
func (s *Server) RequestsFor(vendor string) []Recorded {
	var out []Recorded
	for _, r := range s.Requests() {
		if r.Vendor == vendor {
			out = append(out, r)
		}
	}
	return out
}

func vendorFor(path string) string {
	switch {
	case strings.Contains(path, "/v1/messages"):
		return "anthropic"
	case strings.Contains(path, "/chat/completions"), strings.Contains(path, "/v1/responses"):
		return "openai"
	case strings.Contains(path, ":generateContent"), strings.Contains(path, "/v1beta/"):
		return "gemini"
	}
	return ""
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	vendor := vendorFor(r.URL.Path)
	dumpRequest(r.URL.Path, body)

	s.mu.Lock()
	s.requests = append(s.requests, Recorded{Vendor: vendor, Path: r.URL.Path, Body: body})
	seq := s.n + 1
	idx := s.n
	if s.opt.Cycle && len(s.replies) > 0 {
		idx = s.n % len(s.replies)
	}
	var reply Reply
	if idx < len(s.replies) {
		reply = s.replies[idx]
	} else if len(s.replies) > 0 {
		idx = len(s.replies) - 1
		reply = s.replies[idx]
	}
	s.n++
	s.mu.Unlock()

	if s.opt.Cycle {
		reply = uniquifyIDs(reply, seq)
	}
	if s.opt.Trace != nil {
		fmt.Fprintf(s.opt.Trace, "fake: #%d %-9s %s  %s  -> reply %d/%d: %s\n",
			seq, vendorOrUnknown(vendor), r.URL.Path, describeRequest(body), idx+1, len(s.replies), describeReply(reply))
	}

	if vendor == "" {
		// An unknown path is a student bug worth reporting precisely, but it
		// still gets a 200 so the rest of the session continues and the census
		// in the `session` check can name it.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"error":{"message":"fake: unrecognized path %s"}}`, r.URL.Path)
		return
	}

	// An error is JSON even when the request asked to stream, which is why
	// the parsers decide by the response's content type rather than by what
	// they asked for.
	if reply.Status != 0 && reply.Status != http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(reply.Status)
		io.WriteString(w, reply.ErrBody)
		return
	}

	if wantsStream(r, body, vendor) {
		switch vendor {
		case "anthropic":
			anthropicStream(w, reply)
		case "openai":
			openAIStream(w, reply)
		case "gemini":
			geminiStream(w, reply)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	switch vendor {
	case "anthropic":
		io.WriteString(w, anthropicBody(reply))
	case "openai":
		io.WriteString(w, openAIBody(reply))
	case "gemini":
		io.WriteString(w, geminiBody(reply))
	}
}

func jsonStr(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// anthropicBody reports usage DISJOINT: input_tokens already excludes anything
// read from or written to cache. Verified against platform.claude.com
// 2026-09-12, which documents
// total = cache_read + cache_creation + input.
func anthropicBody(r Reply) string {
	var blocks []string
	if r.Thinking != "" {
		blocks = append(blocks, fmt.Sprintf(`{"type":"thinking","thinking":%s,"signature":%s}`,
			jsonStr(r.Thinking), jsonStr("sig-fake-thinking")))
	}
	if r.Text != "" {
		blocks = append(blocks, fmt.Sprintf(`{"type":"text","text":%s}`, jsonStr(r.Text)))
	}
	if r.ToolName != "" || len(r.Tools) > 0 {
		for _, c := range r.calls() {
			blocks = append(blocks, fmt.Sprintf(`{"type":"tool_use","id":%s,"name":%s,"input":%s}`,
				jsonStr(c.ID), jsonStr(c.Name), c.Args))
		}
	}
	stop := "end_turn"
	if len(r.calls()) > 0 {
		stop = "tool_use"
	}
	return fmt.Sprintf(`{"id":"msg_fake","type":"message","role":"assistant","model":%s,
"content":[%s],"stop_reason":%s,
"usage":{"input_tokens":%d,"cache_creation_input_tokens":%d,"cache_read_input_tokens":%d,"output_tokens":%d}}`,
		jsonStr(Models["anthropic"]), strings.Join(blocks, ","), jsonStr(stop),
		r.Usage.Input, r.Usage.CacheWrite, r.Usage.CacheRead, r.Usage.Output)
}

// openAIBody reports usage SUBSET: prompt_tokens is the grand total, with
// cached_tokens and cache_write_tokens as buckets INSIDE it. Verified live
// 2026-09-12 — an identical repeated request reported the same prompt_tokens
// on the cache miss and the cache hit.
//
// Note also: arguments is a JSON-encoded STRING, and content is null (not "")
// when tool calls are present.
func openAIBody(r Reply) string {
	content := "null"
	if r.Text != "" {
		content = jsonStr(r.Text)
	}
	tools := ""
	finish := "stop"
	if calls := r.calls(); len(calls) > 0 {
		var items []string
		for _, c := range calls {
			items = append(items, fmt.Sprintf(`{"id":%s,"type":"function","function":{"name":%s,"arguments":%s}}`,
				jsonStr(c.ID), jsonStr(c.Name), jsonStr(c.Args)))
		}
		tools = fmt.Sprintf(`,"tool_calls":[%s]`, strings.Join(items, ","))
		finish = "tool_calls"
	}
	prompt := r.Usage.Input + r.Usage.CacheWrite + r.Usage.CacheRead
	return fmt.Sprintf(`{"id":"chatcmpl_fake","object":"chat.completion","model":%s,
"choices":[{"index":0,"message":{"role":"assistant","content":%s%s},"finish_reason":%s}],
"usage":{"prompt_tokens":%d,"completion_tokens":%d,"total_tokens":%d,
"prompt_tokens_details":{"cached_tokens":%d,"cache_write_tokens":%d}}}`,
		jsonStr(Models["openai"]), content, tools, jsonStr(finish),
		prompt, r.Usage.Output, prompt+r.Usage.Output,
		r.Usage.CacheRead, r.Usage.CacheWrite)
}

// geminiBody is the one that disagrees with itself: cachedContentTokenCount is
// INSIDE promptTokenCount, but thoughtsTokenCount is NOT inside
// candidatesTokenCount. Verified 2026-09-12 by arithmetic on live samples.
//
// Also note finishReason: "STOP" even when a functionCall is present. There is
// no tool-call value in the enum, so a parser keyed on the stop signal — which
// works on both other vendors — silently never calls a tool.
func geminiBody(r Reply) string {
	var parts []string
	if r.Thinking != "" {
		parts = append(parts, fmt.Sprintf(`{"text":%s,"thought":true}`, jsonStr(r.Thinking)))
	}
	if r.Text != "" {
		parts = append(parts, fmt.Sprintf(`{"text":%s}`, jsonStr(r.Text)))
	}
	for _, c := range r.calls() {
		parts = append(parts, fmt.Sprintf(`{"functionCall":{"id":%s,"name":%s,"args":%s},"thoughtSignature":%s}`,
			jsonStr(c.ID), jsonStr(c.Name), c.Args, jsonStr("sig-fake-"+c.ID)))
	}
	thoughts := r.GeminiThoughts
	if thoughts > r.Usage.Output {
		thoughts = 0
	}
	candidates := r.Usage.Output - thoughts
	prompt := r.Usage.Input + r.Usage.CacheRead
	return fmt.Sprintf(`{"modelVersion":%s,
"candidates":[{"content":{"role":"model","parts":[%s]},"finishReason":"STOP"}],
"usageMetadata":{"promptTokenCount":%d,"candidatesTokenCount":%d,"cachedContentTokenCount":%d,"thoughtsTokenCount":%d,"totalTokenCount":%d}}`,
		jsonStr(Models["gemini"]), strings.Join(parts, ","),
		prompt, candidates, r.Usage.CacheRead, thoughts, prompt+candidates+thoughts)
}

// --- trace helpers (hand-run only) -----------------------------------------

func vendorOrUnknown(v string) string {
	if v == "" {
		return "unknown"
	}
	return v
}

// describeRequest says, in a few words, what a request body carried. It
// counts the three vendors' tool-result markers rather than parsing the body,
// because the point of the trace is to show the loop turning, not to grade
// it — the grader does the parsing.
func describeRequest(body []byte) string {
	n := strings.Count(string(body), `"tool_result"`) + // anthropic
		strings.Count(string(body), `"role":"tool"`) + strings.Count(string(body), `"role": "tool"`) + // openai
		strings.Count(string(body), `"functionResponse"`) // gemini
	tools := strings.Contains(string(body), `"tools"`)
	var parts []string
	parts = append(parts, fmt.Sprintf("%d bytes", len(body)))
	if tools {
		parts = append(parts, "declares tools")
	}
	if n > 0 {
		parts = append(parts, fmt.Sprintf("carries %d tool result(s)", n))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func describeReply(r Reply) string {
	switch {
	case r.Status != 0:
		return fmt.Sprintf("HTTP %d", r.Status)
	case len(r.Tools) > 1:
		var names []string
		for _, c := range r.Tools {
			names = append(names, c.Name)
		}
		return fmt.Sprintf("%d tool_use blocks: %s", len(r.Tools), strings.Join(names, ", "))
	case r.ToolName != "":
		return fmt.Sprintf("tool_use %s %s", r.ToolName, r.ToolArgs)
	case r.Text != "":
		return fmt.Sprintf("text %q", r.Text)
	default:
		return "(empty)"
	}
}

// uniquifyIDs gives a cycled reply's tool calls IDs no earlier request saw.
// A real vendor never reuses a tool call id, and the reference engine relies
// on that: it computes outstanding calls from the WHOLE dialogue, so a reused
// id looks already answered and the tool is silently never run. Measured: the
// second REPL turn against a cycling script stopped after one request until
// this was added.
func uniquifyIDs(r Reply, seq int) Reply {
	if r.ToolID != "" {
		r.ToolID = fmt.Sprintf("%s_%d", r.ToolID, seq)
	}
	if len(r.Tools) > 0 {
		calls := make([]ToolCall, len(r.Tools))
		for i, c := range r.Tools {
			c.ID = fmt.Sprintf("%s_%d", c.ID, seq)
			calls[i] = c
		}
		r.Tools = calls
	}
	return r
}
