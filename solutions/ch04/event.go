package main

// The event log. Append-only, ordered by Seq, never edited in place.
//
// Everything in this file is vendor-independent. That is not a style
// preference: it is the chapter's central claim, and the fact that this file
// compiles without importing anything vendor-shaped is the proof.

import (
	"encoding/json"
	"fmt"
	"time"
)

// Seq orders the log. Ordering is primary; Time is metadata and may be wrong,
// duplicated, or non-monotonic across machines.
type Seq uint64

type EventType uint8

// The taxonomy grows by appending. Chapter 4 added JobKilled; Chapter 5 adds
// Interrupted. Additive, always: a number once assigned is never reused.
const (
	MessageReceived EventType = iota + 1
	RequestSent
	ResponseStarted
	ResponseEnded
	ToolCalled
	ToolReturned
	Redacted
	ErrorOccurred
	// JobKilled records that a running job was ended by kill_job or by
	// Shutdown. A job that finishes on its own needs no event of its own: its
	// ToolReturned (or the wait_for_job that observed it) carries the status.
	JobKilled
)

var eventTypeNames = map[EventType]string{
	MessageReceived: "message_received",
	RequestSent:     "request_sent",
	ResponseStarted: "response_started",
	ResponseEnded:   "response_ended",
	ToolCalled:      "tool_called",
	ToolReturned:    "tool_returned",
	Redacted:        "redacted",
	ErrorOccurred:   "error_occurred",
	JobKilled:       "job_killed",
}

func (t EventType) String() string {
	if s, ok := eventTypeNames[t]; ok {
		return s
	}
	return fmt.Sprintf("event(%d)", uint8(t))
}

func (t EventType) MarshalJSON() ([]byte, error) {
	s, ok := eventTypeNames[t]
	if !ok {
		return nil, fmt.Errorf("refusing to marshal unknown event type %d", uint8(t))
	}
	return json.Marshal(s)
}

// UnmarshalJSON refuses unknown event types loudly. Skipping one silently
// produces a context that is wrong in a way nothing downstream can detect.
func (t *EventType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	for k, v := range eventTypeNames {
		if normalizeName(v) == normalizeName(s) {
			*t = k
			return nil
		}
	}
	return fmt.Errorf("unknown event type %q: refusing to load this log", s)
}

// Actor says who produced content. There is deliberately no To field:
// addressing is a property of the room, not of the message.
type Actor uint8

const (
	ActorHuman Actor = iota + 1
	ActorAgent
	ActorSystem
	ActorTool
)

var actorNames = map[Actor]string{
	ActorHuman:  "human",
	ActorAgent:  "agent",
	ActorSystem: "system",
	ActorTool:   "tool",
}

func (a Actor) String() string { return actorNames[a] }

func (a Actor) MarshalJSON() ([]byte, error) {
	s, ok := actorNames[a]
	if !ok {
		return nil, fmt.Errorf("refusing to marshal invalid actor %d", uint8(a))
	}
	return json.Marshal(s)
}

func (a *Actor) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	for k, v := range actorNames {
		if normalizeName(v) == normalizeName(s) {
			*a = k
			return nil
		}
	}
	return fmt.Errorf("unknown actor %q: refusing to load this log", s)
}

// Event is the unit of the log. Exactly one payload pointer is non-nil,
// selected by Type. Verbose on purpose: it round-trips as JSON with no
// registry, and it makes the reducer's switch exhaustive by construction.
type Event struct {
	Seq  Seq       `json:"seq"`
	Type EventType `json:"type"`
	Time time.Time `json:"time"`

	Message  *MessageData  `json:"message,omitempty"`
	Request  *RequestData  `json:"request,omitempty"`
	Response *ResponseData `json:"response,omitempty"`
	Tool     *ToolData     `json:"tool,omitempty"`
	Redact   *RedactData   `json:"redact,omitempty"`
	Error    *ErrorData    `json:"error,omitempty"`
	Job      *JobData      `json:"job,omitempty"`
}

type MessageData struct {
	Actor Actor    `json:"actor"`
	Parts PartList `json:"parts"`
}

// RequestData records that we sent a request, and to whom. It deliberately
// does not store the request body: the body is derived output, reproducible
// by replaying the log through a renderer.
type RequestData struct {
	To Provenance `json:"to"`
}

type ResponseData struct {
	Parts PartList   `json:"parts"`
	Usage Usage      `json:"usage"`
	From  Provenance `json:"from"`
}

// ToolData covers both the dispatch of a call and the return of its result.
type ToolData struct {
	CallID  string          `json:"call_id"`
	Name    string          `json:"name,omitempty"`
	Args    json.RawMessage `json:"args,omitempty"`
	Parts   PartList        `json:"parts,omitempty"`
	IsError bool            `json:"is_error,omitempty"`
	// Job is the supervision record for calls that are jobs: which handle
	// the dispatcher issued, where the full output lives, and how far the job
	// had got when this event was written. Absent on the calls that are not
	// jobs — the job verbs and tool_limits — which is how the log shows the
	// difference between "did this tool get a handle" and "should it have".
	Job *JobData `json:"job,omitempty"`
}

// JobData is the job as the log sees it. Output is a Ref of kind handle,
// which is the first use of Chapter 2's Ref for what it was declared for:
// "framework-managed output, cr/io/<handle>". It is never rendered into the
// context as a part; the context gets a stub with the path in it, and the
// path is what the model reads a range of when it wants more.
type JobData struct {
	Handle   int       `json:"handle"`
	Status   JobStatus `json:"status"`
	Output   Ref       `json:"output"`
	Bytes    int       `json:"bytes"`
	ExitCode *int      `json:"exit_code,omitempty"`
	// Reason is set on JobKilled events only: "kill_job" or "shutdown".
	Reason string `json:"reason,omitempty"`
	// Cwd is the directory a run_command job actually ran in, recorded when
	// the process started and only when it was not the working directory.
	// Recorded, never inferred: the log states where the command ran rather
	// than leaving a reader to reconstruct it from the arguments.
	Cwd string `json:"cwd,omitempty"`
}

// RedactData names a SPAN and a LEVEL. The span says where, the level says
// what. A target-plus-flag design cannot express "every tool result older
// than the last save_memory", which is the compaction that matters most.
type RedactData struct {
	From        Seq       `json:"from"`
	To          Seq       `json:"to"`
	Level       Redaction `json:"level"`
	Replacement PartList  `json:"replacement,omitempty"` // RedactSummary only
	Reason      string    `json:"reason,omitempty"`
}

// ErrorData is infrastructure failure. A tool that ran and failed is ordinary
// tool content (ToolResultPart.IsError), not this.
type ErrorData struct {
	Message string `json:"message"`
	Status  int    `json:"status,omitempty"`
}

// Redaction levels, weakest first. Chapter 2 exercises only RedactResult.
type Redaction uint8

const (
	RedactResult   Redaction = iota + 1 // result content -> stub; the call survives
	RedactTool                          // call and result both go
	RedactDialogue                      // prose and reasoning go
	RedactSummary                       // span replaced by compressed prose
)

var redactionNames = map[Redaction]string{
	RedactResult:   "redact_result",
	RedactTool:     "redact_tool",
	RedactDialogue: "redact_dialogue",
	RedactSummary:  "redact_summary",
}

func (r Redaction) String() string { return redactionNames[r] }

func (r Redaction) MarshalJSON() ([]byte, error) {
	s, ok := redactionNames[r]
	if !ok {
		return nil, fmt.Errorf("refusing to marshal invalid redaction level %d", uint8(r))
	}
	return json.Marshal(s)
}

func (r *Redaction) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	for k, v := range redactionNames {
		if normalizeName(v) == normalizeName(s) {
			*r = k
			return nil
		}
	}
	return fmt.Errorf("unknown redaction level %q: refusing to load this log", s)
}

// Usage counts tokens and never money. Prices change; counts are history.
// A dollar amount in the log is wrong the moment a vendor reprices, and it
// destroys the ability to re-cost historical sessions under new rates.
//
// All four fields are DISJOINT and sum to the billable total. Vendors
// disagree about whether that is true of their own reporting. Making it true
// is the parser's job, and it is the purest seam bug in the chapter: nothing
// crashes, no test fails, the number is simply not the number.
type Usage struct {
	Input      int `json:"input"`       // neither read from nor written to cache
	CacheWrite int `json:"cache_write"` // typically costs MORE than plain input
	CacheRead  int `json:"cache_read"`  // typically an order of magnitude LESS
	Output     int `json:"output"`
}

func (u *Usage) Add(o Usage) {
	u.Input += o.Input
	u.CacheWrite += o.CacheWrite
	u.CacheRead += o.CacheRead
	u.Output += o.Output
}
