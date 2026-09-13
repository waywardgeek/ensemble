// Package fakeanthropic implements a deterministic stand-in for the Anthropic
// Messages API, used by the course auto-grader.
//
// Design rules, in order of importance:
//
//  1. It records everything. Every inbound request is captured verbatim
//     (headers, raw body, parsed body) so that checks run *after* the student
//     program has exited, against evidence rather than against live state.
//
//  2. It is forgiving at the transport layer and strict in the record. A
//     malformed request still gets a well-formed 200 response, with the
//     violation written to the record. A grader that returns 400 on the first
//     mistake teaches the student one bug per run; this one teaches them all
//     the bugs in a single run.
//
//  3. It is deterministic. Replies are scripted by request index, and token
//     counts are a pure function of the payload. No model, no network, no
//     API key, no cost, identical output on every machine.
package fakeanthropic

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
)

// Usage mirrors the `usage` object of a real Messages API response.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Message is one entry of the request's `messages` array. Content is kept raw
// because the API accepts both a bare string and an array of content blocks,
// and a naive Chapter 1 program is expected to send the string form.
type Message struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// Text flattens the content field to plain text for comparison purposes.
// It accepts the string form, the block-array form, or anything else (in
// which case the raw JSON is returned so a diff is still legible).
func (m Message) Text() string {
	var s string
	if err := json.Unmarshal(m.Content, &s); err == nil {
		return s
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(m.Content, &blocks); err == nil {
		var b strings.Builder
		for _, blk := range blocks {
			if blk.Type == "text" || blk.Text != "" {
				b.WriteString(blk.Text)
			}
		}
		return b.String()
	}
	return string(m.Content)
}

// ContentShape reports how the content field was encoded: "string", "blocks",
// "empty" or "unknown".
func (m Message) ContentShape() string {
	if len(m.Content) == 0 {
		return "empty"
	}
	var s string
	if err := json.Unmarshal(m.Content, &s); err == nil {
		return "string"
	}
	var blocks []json.RawMessage
	if err := json.Unmarshal(m.Content, &blocks); err == nil {
		return "blocks"
	}
	return "unknown"
}

// MessagesRequest is the subset of the request body Chapter 1 cares about.
type MessagesRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    json.RawMessage `json:"system"`
	Messages  []Message       `json:"messages"`
	Stream    bool            `json:"stream"`
}

// SystemText flattens the system field (string or block array) to plain text.
func (r MessagesRequest) SystemText() string {
	if len(r.System) == 0 {
		return ""
	}
	return Message{Content: r.System}.Text()
}

// Record is everything the fake saw and did for one HTTP request.
type Record struct {
	Seq        int             // 1-based arrival order
	Method     string          //
	Path       string          //
	Header     http.Header     // copy of inbound headers
	RawBody    []byte          // exact bytes received
	Body       MessagesRequest // parsed body (zero value if ParseErr != "")
	ParseErr   string          // non-empty if the body was not valid JSON
	Violations []string        // wire-format problems found at receive time
	Overflow   bool            // arrived after the scripted rounds were exhausted
	Reply      string          // assistant text served back
	Usage      Usage           // usage reported in the response
}

// Server is the fake API. The zero value is not usable; call Start.
type Server struct {
	// Script supplies the assistant reply for the nth request (0-based). If
	// the program makes more requests than there are scripted replies, the
	// overflow reply is served and a violation is recorded.
	Script []string

	mu       sync.Mutex
	records  []Record
	httpSrv  *http.Server
	listener net.Listener
	baseURL  string
}

// Start binds to an ephemeral localhost port and begins serving. The returned
// base URL is what belongs in ANTHROPIC_BASE_URL.
func (s *Server) Start() (string, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	s.listener = ln
	s.baseURL = "http://" + ln.Addr().String()

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handle)
	s.httpSrv = &http.Server{Handler: mux}
	go func() { _ = s.httpSrv.Serve(ln) }()
	return s.baseURL, nil
}

// Close shuts the server down.
func (s *Server) Close() {
	if s.httpSrv != nil {
		_ = s.httpSrv.Close()
	}
}

// Records returns a snapshot of everything received, in arrival order.
func (s *Server) Records() []Record {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Record, len(s.records))
	copy(out, s.records)
	return out
}

// countTokens is the fake's deterministic stand-in for a tokenizer: one token
// per four characters, rounded up, minimum one. The exact rule does not matter
// to the student — what matters is that the grader can predict the totals
// exactly and therefore catch a program that invents its usage report.
func countTokens(s string) int {
	n := (len(s) + 3) / 4
	if n < 1 {
		return 1
	}
	return n
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	rec := Record{
		Method: r.Method,
		Path:   r.URL.Path,
		Header: r.Header.Clone(),
	}

	buf := make([]byte, 0, 64*1024)
	tmp := make([]byte, 32*1024)
	for {
		n, err := r.Body.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if err != nil {
			break
		}
	}
	rec.RawBody = buf

	// --- wire checks, recorded not enforced -----------------------------
	if r.Method != http.MethodPost {
		rec.Violations = append(rec.Violations,
			fmt.Sprintf("method is %s, the Messages API requires POST", r.Method))
	}
	if !strings.HasSuffix(r.URL.Path, "/messages") {
		rec.Violations = append(rec.Violations,
			fmt.Sprintf("path %q does not end in /messages (expected /v1/messages)", r.URL.Path))
	}
	if got := r.Header.Get("x-api-key"); got == "" {
		rec.Violations = append(rec.Violations, "missing x-api-key header")
	} else if got != ExpectedAPIKey {
		rec.Violations = append(rec.Violations, fmt.Sprintf(
			"x-api-key was %q, but the grader put %q in ANTHROPIC_API_KEY "+
				"— read the key from the environment instead of hardcoding one",
			truncateKey(got), ExpectedAPIKey))
	}
	if r.Header.Get("anthropic-version") == "" {
		rec.Violations = append(rec.Violations, "missing anthropic-version header")
	}
	if ct := r.Header.Get("content-type"); !strings.HasPrefix(ct, "application/json") {
		rec.Violations = append(rec.Violations,
			fmt.Sprintf("content-type is %q, expected application/json", ct))
	}

	if err := json.Unmarshal(buf, &rec.Body); err != nil {
		rec.ParseErr = err.Error()
		rec.Violations = append(rec.Violations, "request body is not valid JSON: "+err.Error())
	} else {
		if rec.Body.Model == "" {
			rec.Violations = append(rec.Violations, "model field is empty")
		} else if rec.Body.Model != ExpectedModel {
			rec.Violations = append(rec.Violations,
				fmt.Sprintf("model is %q, want %q — the model must come from ANTHROPIC_MODEL, not a hardcoded string",
					rec.Body.Model, ExpectedModel))
		}
		if strings.TrimSpace(rec.Body.SystemText()) == "" {
			rec.Violations = append(rec.Violations,
				"system field is empty — the agent must send a system prompt")
		}
		if rec.Body.MaxTokens <= 0 {
			rec.Violations = append(rec.Violations,
				"max_tokens is missing or non-positive (the API will not guess)")
		}
		if rec.Body.Stream {
			rec.Violations = append(rec.Violations,
				"stream:true — Chapter 1 is non-streaming")
		}
		if len(rec.Body.Messages) == 0 {
			rec.Violations = append(rec.Violations, "messages array is empty")
		}
		for i, m := range rec.Body.Messages {
			switch m.Role {
			case "user", "assistant":
			default:
				rec.Violations = append(rec.Violations,
					fmt.Sprintf("messages[%d].role is %q, expected user or assistant", i, m.Role))
			}
			if strings.TrimSpace(m.Text()) == "" {
				rec.Violations = append(rec.Violations,
					fmt.Sprintf("messages[%d] has empty content", i))
			}
			if i > 0 && m.Role == rec.Body.Messages[i-1].Role {
				rec.Violations = append(rec.Violations,
					fmt.Sprintf("messages[%d] and [%d] are both %q — roles must strictly alternate",
						i-1, i, m.Role))
			}
		}
		if n := len(rec.Body.Messages); n > 0 {
			if rec.Body.Messages[0].Role != "user" {
				rec.Violations = append(rec.Violations, "conversation must begin with a user message")
			}
			if rec.Body.Messages[n-1].Role != "user" {
				rec.Violations = append(rec.Violations,
					"the final message must be the user's — otherwise there is nothing to answer")
			}
		}
	}

	// --- reply ----------------------------------------------------------
	s.mu.Lock()
	rec.Seq = len(s.records) + 1
	idx := rec.Seq - 1
	var reply string
	if idx < len(s.Script) {
		reply = s.Script[idx]
	} else {
		reply = "(the fake server has run out of scripted replies)"
		// Not a wire-format fault — the request may be perfectly well-formed.
		// Request *count* is the caller's check to make.
		rec.Overflow = true
	}
	rec.Reply = reply

	in := countTokens(rec.Body.SystemText())
	for _, m := range rec.Body.Messages {
		in += countTokens(m.Text())
	}
	rec.Usage = Usage{InputTokens: in, OutputTokens: countTokens(reply)}
	s.records = append(s.records, rec)
	s.mu.Unlock()

	resp := map[string]any{
		"id":            fmt.Sprintf("msg_fake_%03d", rec.Seq),
		"type":          "message",
		"role":          "assistant",
		"model":         "claude-fake-course-1",
		"content":       splitIntoBlocks(reply),
		"stop_reason":   "end_turn",
		"stop_sequence": nil,
		"usage":         rec.Usage,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// TotalUsage sums the usage the fake reported across every request served.
func (s *Server) TotalUsage() Usage {
	s.mu.Lock()
	defer s.mu.Unlock()
	var t Usage
	for _, r := range s.records {
		t.InputTokens += r.Usage.InputTokens
		t.OutputTokens += r.Usage.OutputTokens
	}
	return t
}

// ExpectedAPIKey is the credential the grader puts in ANTHROPIC_API_KEY. The
// fake requires this exact value.
//
// It used to accept any non-empty string, which graded §1.2's "read all three,
// hardcode none" vacuously: deleting the os.Getenv call and hardcoding a
// plausible-looking key still scored 100. Comparing against the value the
// harness actually supplied is the whole fix, and it costs one equality.
const ExpectedAPIKey = "sk-ant-course-grader-fake"

// ExpectedModel is the model the grader puts in ANTHROPIC_MODEL. The fake
// requires the request body to name this exact model.
//
// Same disease as ExpectedAPIKey, same cure. A non-empty check graded §1.2's
// "read the model from the environment" vacuously: hardcoding "claude-3-5-
// sonnet" and never reading ANTHROPIC_MODEL still scored 100. The harness sets
// the variable FROM this constant, so there is one source of truth and a
// student who echoes the environment back to us cannot be told apart from one
// who guessed — because guessing this string is not a thing anyone does.
const ExpectedModel = "claude-fake-course-1"

// truncateKey keeps a credential out of the transcript while still showing the
// student enough to recognise their own hardcoded string.
func truncateKey(k string) string {
	if len(k) <= 12 {
		return k
	}
	return k[:12] + "..."
}

// splitIntoBlocks renders one reply as SEVERAL text blocks.
//
// §1.2 tells the student that content is "always a list of typed blocks. Walk
// it and concatenate the text blocks", and calls reading content[0].text the
// classic day-one stumble. A fixture that only ever served ONE block graded
// that instruction vacuously — the stumble the chapter warns about loudest
// scored full marks. Serving two blocks makes the walk load-bearing.
//
// The blocks concatenate to exactly `reply`, so every expected answer, and the
// output token count derived from it, are unchanged.
func splitIntoBlocks(reply string) []map[string]string {
	one := []map[string]string{{"type": "text", "text": reply}}
	third := len(reply) / 3
	if third == 0 {
		return one
	}
	i := strings.IndexByte(reply[third:], ' ')
	if i < 0 {
		return one
	}
	cut := third + i + 1 // the space stays on the first block
	if cut <= 0 || cut >= len(reply) {
		return one
	}
	return []map[string]string{
		{"type": "text", "text": reply[:cut]},
		{"type": "text", "text": reply[cut:]},
	}
}
